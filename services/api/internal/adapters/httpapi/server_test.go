package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type actors struct{}

func (actors) Actor(_ context.Context, id string) (domain.Actor, error) {
	return domain.Actor{ID: id, Name: "Test", Role: "patient"}, nil
}
func TestIdentityCannotBeForgedByHeader(t *testing.T) {
	s := Server{Actors: actors{}, Origin: "https://hospital.test"}
	r := httptest.NewRequest("GET", "/v1/me", nil)
	r.Header.Set("X-User-ID", "admin")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("wanted 401 got %d", w.Code)
	}
}
func TestOriginCheckedBeforeAuthentication(t *testing.T) {
	s := Server{Actors: actors{}, Origin: "https://hospital.test"}
	r := httptest.NewRequest("POST", "/v1/patients", nil)
	r.Header.Set("Origin", "https://evil.test")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
}
func TestIntrospectionRejectsExpiredAndMismatchedSessions(t *testing.T) {
	for _, tc := range []struct {
		user   string
		expiry time.Time
		want   int
	}{{"u1", time.Now().Add(time.Hour), 200}, {"u1", time.Now().Add(-time.Hour), 401}, {"u2", time.Now().Add(time.Hour), 401}} {
		auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{"user":{"id":"u1"},"session":{"userId":%q,"expiresAt":%q}}`, tc.user, tc.expiry.Format(time.RFC3339))
		}))
		s := Server{Actors: actors{}, AuthURL: auth.URL, Client: auth.Client()}
		r := httptest.NewRequest("GET", "/v1/me", nil)
		r.Header.Set("Cookie", "test=session")
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		auth.Close()
		if w.Code != tc.want {
			t.Fatalf("want %d got %d", tc.want, w.Code)
		}
	}
}

type mockRepo struct {
	overviewFn func(context.Context, domain.Actor) (domain.Overview, error)
}

func (m mockRepo) PatientProfile(context.Context, domain.Actor, string) (domain.PatientProfile, error) {
	return domain.PatientProfile{}, nil
}
func (m mockRepo) UpdatePatientProfile(context.Context, domain.Actor, string, domain.PatientProfileInput) (domain.PatientProfile, error) {
	return domain.PatientProfile{}, nil
}
func (m mockRepo) PatientRevisions(context.Context, domain.Actor, string, int) ([]domain.ProfileRevision, error) {
	return nil, nil
}
func (m mockRepo) LinkPatient(context.Context, domain.Actor, string, domain.PatientAccess) error {
	return nil
}
func (m mockRepo) Patients(context.Context, domain.Actor, string, int) ([]domain.Patient, error) {
	return nil, nil
}
func (m mockRepo) RegisterPatient(context.Context, domain.Actor, domain.PatientInput) (domain.Patient, error) {
	return domain.Patient{}, nil
}
func (m mockRepo) Overview(ctx context.Context, a domain.Actor) (domain.Overview, error) {
	if m.overviewFn != nil {
		return m.overviewFn(ctx, a)
	}
	return domain.Overview{}, nil
}
func (m mockRepo) Enqueue(context.Context, domain.Actor, domain.MessageInput, string) (domain.Message, error) {
	return domain.Message{}, nil
}
func (m mockRepo) Messages(context.Context) ([]domain.Message, error) {
	return nil, nil
}

type roleActors struct {
	role string
}

func (r roleActors) Actor(_ context.Context, id string) (domain.Actor, error) {
	return domain.Actor{ID: id, Name: "Test User", Role: r.role}, nil
}

func TestOverviewHTTPAuthorizationAndScoping(t *testing.T) {
	tests := []struct {
		role       string
		wantCode   int
		expectBody bool
	}{
		{"admin", 200, true},
		{"doctor", 200, true},
		{"patient", 200, true},
		{"receptionist", 200, true},
		{"accountant", 403, false},
		{"nurse", 403, false},
		{"pharmacist", 403, false},
		{"lab_technician", 403, false},
		{"case_manager", 403, false},
		{"unknown", 403, false},
	}

	for _, tc := range tests {
		t.Run("overview HTTP "+tc.role, func(t *testing.T) {
			repo := mockRepo{
				overviewFn: func(ctx context.Context, a domain.Actor) (domain.Overview, error) {
					if a.Role == "admin" {
						return domain.Overview{
							PatientCount:  10,
							InvoicesMinor: 50000,
							Doctors:       5,
							Patients:      10,
						}, nil
					}
					return domain.Overview{
						PatientCount: 3,
						Patients:     3,
					}, nil
				},
			}

			auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{"user":{"id":"u1"},"session":{"userId":"u1","expiresAt":%q}}`,
					time.Now().Add(time.Hour).Format(time.RFC3339))
			}))
			defer auth.Close()

			s := Server{
				Actors:  roleActors{role: tc.role},
				App:     application.Hospital{Store: repo},
				AuthURL: auth.URL,
				Client:  auth.Client(),
			}

			r := httptest.NewRequest("GET", "/v1/overview", nil)
			r.Header.Set("Cookie", "test=session")
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, r)

			if w.Code != tc.wantCode {
				t.Fatalf("role %s: want status %d got %d (body: %s)", tc.role, tc.wantCode, w.Code, w.Body.String())
			}

			if tc.wantCode == 200 {
				var ov domain.Overview
				if err := json.Unmarshal(w.Body.Bytes(), &ov); err != nil {
					t.Fatalf("failed to decode response for %s: %v", tc.role, err)
				}
				if tc.role != "admin" && (ov.InvoicesMinor != 0 || ov.Doctors != 0) {
					t.Fatalf("role %s leaked administrator totals: %+v", tc.role, ov)
				}
			}
		})
	}
}
