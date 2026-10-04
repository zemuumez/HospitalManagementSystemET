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

func testNursing(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, cases []domain.Case) {
	t.Helper()
	ctx := context.Background()
	clinic := application.Clinical{Store: store, Now: time.Now}
	for _, id := range []string{"nurse", "other-nurse"} {
		if _, e := db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES($1,$1,$1||'@example.test')`, id); e != nil {
			t.Fatal(e)
		}
		if _, e := db.Exec(ctx, `INSERT INTO staff_access(user_id,role) VALUES($1,'nurse')`, id); e != nil {
			t.Fatal(e)
		}
	}
	nurse := domain.Actor{ID: "nurse", Role: "nurse"}
	other := domain.Actor{ID: "other-nurse", Role: "nurse"}
	enc, e := clinic.Admit(ctx, actors[0], domain.EncounterInput{Kind: "opd", CaseID: cases[0].ID, AdmittedAt: time.Now().Add(-time.Hour)}, "nursing-encounter-key")
	if e != nil {
		t.Fatal(e)
	}
	pulse := 80.0
	input := domain.VitalsInput{ObservedAt: time.Now().Add(-time.Minute), Measurements: domain.Measurements{PulseBpm: &pulse}}
	if _, e = clinic.RecordVitals(ctx, nurse, enc.ID, input, "nursing-vitals-key"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("unassigned nurse wrote", e)
	}
	assignment, e := clinic.AssignNurse(ctx, actors[1], enc.ID, domain.NurseAssignment{NurseID: nurse.ID, Active: true, Reason: "Shift assignment"})
	if e != nil || assignment.Version != 1 {
		t.Fatal(e)
	}
	if _, e = clinic.AssignNurse(ctx, actors[2], enc.ID, domain.NurseAssignment{NurseID: other.ID, Active: true, Reason: "Forged"}); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("other doctor assigned nurse", e)
	}
	list, e := clinic.NursingEncounters(ctx, nurse, 1)
	if e != nil || len(list) != 1 || list[0].ID != enc.ID {
		t.Fatal("nurse encounter scope", list, e)
	}
	list, e = clinic.NursingEncounters(ctx, other, 1)
	if e != nil || len(list) != 0 {
		t.Fatal("other nurse encounter scope", list, e)
	}
	vital, e := clinic.RecordVitals(ctx, nurse, enc.ID, input, "nursing-vitals-key")
	if e != nil {
		t.Fatal(e)
	}
	retry, e := clinic.RecordVitals(ctx, nurse, enc.ID, input, "nursing-vitals-key")
	if e != nil || retry.ID != vital.ID {
		t.Fatal("vitals retry", e)
	}
	altered := input
	altered.Note = "Changed"
	if _, e = clinic.RecordVitals(ctx, nurse, enc.ID, altered, "nursing-vitals-key"); !errors.Is(e, domain.ErrConflict) {
		t.Fatal("vitals changed-key replay", e)
	}
	if _, e = clinic.Vitals(ctx, other, enc.ID, 1); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("other nurse vitals access", e)
	}
	if _, e = clinic.Vitals(ctx, actors[3], enc.ID, 1); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("unreleased patient vitals", e)
	}
	correction := input
	correction.CorrectionOf = vital.ID
	correction.CorrectionReason = "Transcription correction"
	newPulse := 81.0
	correction.Measurements.PulseBpm = &newPulse
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, errs[n] = clinic.RecordVitals(ctx, actors[1], enc.ID, correction, fmt.Sprintf("vitals-correction-key-%d", n))
		}(n)
	}
	wg.Wait()
	if !((errs[0] == nil && errors.Is(errs[1], domain.ErrStale)) || (errs[1] == nil && errors.Is(errs[0], domain.ErrStale))) {
		t.Fatal("correction fork", errs)
	}
	values, e := clinic.Vitals(ctx, actors[1], enc.ID, 1)
	if e != nil || len(values) != 2 {
		t.Fatal("retained vitals", values, e)
	}
	if _, e = db.Exec(ctx, `UPDATE clinical_vitals SET note='changed' WHERE id=$1`, vital.ID); e == nil {
		t.Fatal("signed vitals changed")
	}
	if _, e = db.Exec(ctx, `DELETE FROM clinical_vitals WHERE id=$1`, vital.ID); e == nil {
		t.Fatal("signed vitals deleted")
	}
	if _, e = clinic.AssignNurse(ctx, actors[0], enc.ID, domain.NurseAssignment{NurseID: nurse.ID, Active: false, Version: 1, Reason: "Shift ended"}); e != nil {
		t.Fatal(e)
	}
	if _, e = clinic.Vitals(ctx, nurse, enc.ID, 1); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("revoked nurse access", e)
	}
	if _, e = clinic.RecordVitals(ctx, nurse, enc.ID, input, "nursing-vitals-key"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("revoked nurse replay access", e)
	}
	if _, e = clinic.Discharge(ctx, actors[1], enc.ID, domain.Discharge{Version: 1, Summary: "Completed"}); e != nil {
		t.Fatal(e)
	}
	if _, e = clinic.RecordVitals(ctx, actors[1], enc.ID, input, "closed-vitals-key"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("closed encounter accepted vitals", e)
	}
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who := strings.TrimPrefix(r.Header.Get("Cookie"), "session=")
		fmt.Fprintf(w, `{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, who, who, time.Now().Add(time.Hour).Format(time.RFC3339))
	}))
	defer auth.Close()
	handler := httpapi.Server{Clinical: clinic, Actors: store, AuthURL: auth.URL, Origin: "http://hospital.test", Client: auth.Client()}.Handler()
	for _, tc := range []struct {
		method, path, actor, body string
		want                      int
	}{
		{"GET", "/v1/nursing-encounters", "nurse", "", 200},
		{"GET", "/v1/encounters/" + enc.ID + "/vitals", "nurse", "", 404},
		{"GET", "/v1/encounters/" + enc.ID + "/vitals", "doctor", "", 200},
		{"GET", "/v1/encounters/" + enc.ID + "/vitals", "reception", "", 403},
		{"POST", "/v1/encounters/" + enc.ID + "/vitals", "doctor", `{"unexpected":true}`, 400},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Cookie", "session="+tc.actor)
		req.Header.Set("Origin", "http://hospital.test")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatalf("nursing HTTP wanted %d got %d: %s", tc.want, w.Code, w.Body.String())
		}
	}
}
