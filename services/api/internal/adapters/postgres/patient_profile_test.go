package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func testPatientProfiles(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient) {
	t.Helper()
	ctx := context.Background()
	h := application.Hospital{Store: store, Now: time.Now}
	admin, doctor, other, patient := actors[0], actors[1], actors[2], actors[3]
	profile, e := h.PatientProfile(ctx, admin, patients[0].ID)
	if e != nil || profile.Version != 1 || !profile.Active || profile.Gender != "unknown" {
		t.Fatal("profile initialization", e, profile)
	}
	if _, e = h.PatientProfile(ctx, patient, patients[1].ID); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("patient profile leaked", e)
	}
	if _, e = h.PatientProfile(ctx, other, patients[0].ID); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("other doctor profile leaked", e)
	}
	if _, e = h.PatientProfile(ctx, doctor, patients[0].ID); e != nil {
		t.Fatal("assigned doctor cannot read profile", e)
	}
	input := profile.PatientProfileInput
	input.Email = "synthetic@example.test"
	input.Gender = "female"
	input.BloodGroup = "AB+"
	input.Address1 = "Synthetic address"
	input.City = "Addis Ababa"
	input.PostalCode = "00100"
	input.EmergencyName = "Test contact"
	input.EmergencyPhone = "+251911000001"
	input.Reason = "Correct demographics"
	if _, e = h.UpdatePatientProfile(ctx, patient, profile.ID, input); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("patient altered demographics", e)
	}
	if _, e = h.UpdatePatientProfile(ctx, doctor, profile.ID, input); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("doctor altered demographics", e)
	}
	var results [2]domain.PatientProfile
	var errs [2]error
	var wg sync.WaitGroup
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			results[n], errs[n] = h.UpdatePatientProfile(ctx, admin, profile.ID, input)
		}(n)
	}
	wg.Wait()
	winner := 0
	if errs[0] != nil {
		winner = 1
	}
	if errs[winner] != nil || !errors.Is(errs[1-winner], domain.ErrStale) || results[winner].Version != 2 {
		t.Fatal("lost update protection", errs, results)
	}
	own, e := h.PatientProfile(ctx, patient, profile.ID)
	if e != nil || own.Email != input.Email || own.PostalCode != "00100" {
		t.Fatal("profile persistence", e, own)
	}
	revisions, e := h.PatientRevisions(ctx, admin, profile.ID, 1)
	if e != nil || len(revisions) != 1 || revisions[0].Before.Email != "" || revisions[0].After.Email != input.Email || revisions[0].Reason != input.Reason {
		t.Fatal("profile revision", e, revisions)
	}
	if _, e = h.PatientRevisions(ctx, patient, profile.ID, 1); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("patient accessed privileged revision log", e)
	}
	if _, e = db.Exec(ctx, `UPDATE patient_profile_revision SET reason='changed' WHERE patient_id=$1`, profile.ID); e == nil {
		t.Fatal("revision changed")
	}
	if _, e = db.Exec(ctx, `DELETE FROM patient WHERE id=$1`, profile.ID); e == nil {
		t.Fatal("patient with retained history deleted")
	}
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who := strings.TrimPrefix(r.Header.Get("Cookie"), "session=")
		fmt.Fprintf(w, `{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, who, who, time.Now().Add(time.Hour).Format(time.RFC3339))
	}))
	defer auth.Close()
	handler := httpapi.Server{App: h, Actors: store, AuthURL: auth.URL, Origin: "http://hospital.test", Client: auth.Client()}.Handler()
	for _, tc := range []struct {
		actor, method, path, body, origin string
		want                              int
	}{
		{"patient", "GET", "/v1/patients/" + profile.ID, "", "", 200},
		{"patient", "GET", "/v1/patients/" + patients[1].ID, "", "", 404},
		{"admin", "PATCH", "/v1/patients/" + profile.ID + "/profile", `{"role":"admin"}`, "http://hospital.test", 400},
		{"admin", "PATCH", "/v1/patients/" + profile.ID + "/profile", `{}`, "http://evil.test", 403},
		{"patient", "GET", "/v1/patients/" + profile.ID + "/revisions", "", "", 403},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Cookie", "session="+tc.actor)
		req.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatalf("profile HTTP wanted %d got %d: %s", tc.want, w.Code, w.Body.String())
		}
	}
}
