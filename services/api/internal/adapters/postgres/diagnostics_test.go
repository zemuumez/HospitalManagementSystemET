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

func testDiagnostics(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, cases []domain.Case) {
	t.Helper()
	ctx := context.Background()
	admin, doctor, other, patient := actors[0], actors[1], actors[2], actors[3]
	lab := domain.Actor{ID: "lab", Role: "lab_technician"}
	if _, e := db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES('lab','Lab','lab@example.test');INSERT INTO staff_access(user_id,role) VALUES('lab','lab_technician')`); e != nil {
		t.Fatal(e)
	}
	d := application.Diagnostics{Store: store}
	clinic := application.Clinical{Store: store, Now: time.Now}
	enc, e := clinic.Admit(ctx, admin, domain.EncounterInput{Kind: "opd", CaseID: cases[0].ID, AdmittedAt: time.Now().Add(-time.Minute)}, "diagnostic-encounter-01")
	if e != nil {
		t.Fatal(e)
	}
	test, e := d.CreateTest(ctx, lab, domain.DiagnosticTestInput{Kind: "pathology", Name: "Synthetic assay", ShortName: "SYN", Category: "Test", ChargeMinor: 100, Parameters: []domain.DiagnosticParameter{{Name: "Synthetic parameter", Unit: "test-unit", ReferenceRange: "Hospital-defined", ValueType: "number"}}})
	if e != nil {
		t.Fatal(e)
	}
	input := domain.DiagnosticOrderInput{EncounterID: enc.ID, TestID: test.ID, Indication: "Synthetic investigation"}
	if _, e = d.Order(ctx, other, input, "diagnostic-order-001"); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("wrong doctor ordered test", e)
	}
	order, e := d.Order(ctx, doctor, input, "diagnostic-order-001")
	if e != nil {
		t.Fatal(e)
	}
	retry, e := d.Order(ctx, doctor, input, "diagnostic-order-001")
	if e != nil || retry.ID != order.ID {
		t.Fatal("diagnostic retry", e)
	}
	if _, e = db.Exec(ctx, `INSERT INTO diagnostic_parameter(test_id,position,name,unit,value_type) VALUES($1,2,'Late','','text')`, test.ID); e == nil {
		t.Fatal("ordered parameters changed")
	}
	if rows, e := d.Orders(ctx, patient, enc.ID, 1); e != nil || len(rows) != 0 {
		t.Fatal("unreleased order disclosed", e)
	}
	if _, e = d.Results(ctx, patient, order.ID, 1); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("unreleased results disclosed", e)
	}
	if _, e = d.Transition(ctx, lab, order.ID, domain.DiagnosticAction{Version: 1, Action: "process"}); !errors.Is(e, domain.ErrStale) {
		t.Fatal("sample collection skipped", e)
	}
	order, e = d.Transition(ctx, lab, order.ID, domain.DiagnosticAction{Version: 1, Action: "collect", SampleReference: "SAMPLE-1"})
	if e != nil {
		t.Fatal(e)
	}
	order, e = d.Transition(ctx, lab, order.ID, domain.DiagnosticAction{Version: 2, Action: "process"})
	if e != nil {
		t.Fatal(e)
	}
	bad := domain.DiagnosticResultInput{Version: 3, Summary: "Synthetic result", Values: []domain.DiagnosticValue{{Position: 1, Value: "NaN"}}}
	if _, e = d.Submit(ctx, lab, order.ID, bad); !errors.Is(e, domain.ErrValidation) {
		t.Fatal("nonfinite numeric result", e)
	}
	bad.Values[0].Value = "12.3"
	var results [2]domain.DiagnosticResult
	var errs [2]error
	var wg sync.WaitGroup
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) { defer wg.Done(); results[n], errs[n] = d.Submit(ctx, lab, order.ID, bad) }(n)
	}
	wg.Wait()
	winner := 0
	if errs[0] != nil {
		winner = 1
	}
	if errs[winner] != nil || !errors.Is(errs[1-winner], domain.ErrStale) {
		t.Fatal("concurrent result submission", errs)
	}
	result := results[winner]
	if _, e = db.Exec(ctx, `UPDATE diagnostic_value SET value='15' WHERE result_id=$1`, result.ID); e == nil {
		t.Fatal("result overwritten")
	}
	if _, e = db.Exec(ctx, `INSERT INTO diagnostic_value(result_id,position,value) VALUES($1,2,'late')`, result.ID); e == nil {
		t.Fatal("sealed result extended")
	}
	if _, e = d.Transition(ctx, lab, order.ID, domain.DiagnosticAction{Version: 4, Action: "sign"}); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("lab signed clinical review", e)
	}
	if _, e = d.Transition(ctx, other, order.ID, domain.DiagnosticAction{Version: 4, Action: "sign"}); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("wrong doctor signed", e)
	}
	order, e = d.Transition(ctx, doctor, order.ID, domain.DiagnosticAction{Version: 4, Action: "sign"})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = d.Results(ctx, patient, order.ID, 1); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("signed but unreleased result disclosed", e)
	}
	order, e = d.Transition(ctx, doctor, order.ID, domain.DiagnosticAction{Version: 5, Action: "release"})
	if e != nil {
		t.Fatal(e)
	}
	published, e := d.Results(ctx, patient, order.ID, 1)
	if e != nil || len(published) != 1 || !published[0].Released || len(published[0].Parameters) != 1 || published[0].Parameters[0].Unit != "test-unit" {
		t.Fatal("released report", e, published)
	}
	amendment := domain.DiagnosticResultInput{Version: 6, Summary: "Corrected synthetic result", Values: []domain.DiagnosticValue{{Position: 1, Value: "13.2"}}}
	if _, e = d.Submit(ctx, lab, order.ID, amendment); !errors.Is(e, domain.ErrValidation) {
		t.Fatal("amendment without reason", e)
	}
	amendment.AmendmentReason = "Corrected transcription"
	amended, e := d.Submit(ctx, lab, order.ID, amendment)
	if e != nil || amended.Revision != 2 {
		t.Fatal("amendment", e)
	}
	published, e = d.Results(ctx, patient, order.ID, 1)
	if e != nil || len(published) != 1 || published[0].Values[0].Value != "12.3" {
		t.Fatal("pending amendment leaked or old report lost", e, published)
	}
	order, e = d.Transition(ctx, doctor, order.ID, domain.DiagnosticAction{Version: 7, Action: "sign"})
	if e != nil {
		t.Fatal(e)
	}
	order, e = d.Transition(ctx, doctor, order.ID, domain.DiagnosticAction{Version: 8, Action: "release"})
	if e != nil {
		t.Fatal(e)
	}
	published, e = d.Results(ctx, patient, order.ID, 1)
	if e != nil || len(published) != 2 || published[0].Values[0].Value != "13.2" {
		t.Fatal("released amendment history", e)
	}
	if _, e = d.Results(ctx, other, order.ID, 1); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("cross-doctor report disclosure", e)
	}
	if _, e = d.Results(ctx, actors[4], order.ID, 1); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("reception report disclosure", e)
	}
	if _, e = db.Exec(ctx, `DELETE FROM diagnostic_review WHERE result_id=$1`, result.ID); e == nil {
		t.Fatal("signed review removed")
	}
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who := strings.TrimPrefix(r.Header.Get("Cookie"), "session=")
		fmt.Fprintf(w, `{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, who, who, time.Now().Add(time.Hour).Format(time.RFC3339))
	}))
	defer auth.Close()
	handler := httpapi.Server{Diagnostics: d, Actors: store, AuthURL: auth.URL, Origin: "http://hospital.test", Client: auth.Client()}.Handler()
	for _, tc := range []struct {
		actor, method, path, body, origin string
		want                              int
	}{
		{"patient", "GET", "/v1/diagnostic-orders/" + order.ID + "/results", "", "", 200},
		{"reception", "GET", "/v1/diagnostic-orders/" + order.ID + "/results", "", "", 403},
		{"lab", "POST", "/v1/diagnostic-tests", `{"extra":true}`, "http://hospital.test", 400},
		{"lab", "POST", "/v1/diagnostic-tests", `{}`, "http://evil.test", 403},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Cookie", "session="+tc.actor)
		req.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatalf("diagnostic HTTP wanted %d got %d: %s", tc.want, w.Code, w.Body.String())
		}
	}
}
