package httpapi

import (
	"context"
	"fmt"
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
