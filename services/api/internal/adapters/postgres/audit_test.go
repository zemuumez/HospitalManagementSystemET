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
	"testing"
	"time"
)

func testAudit(t *testing.T, db *pgxpool.Pool, s Store, actors []domain.Actor) {
	t.Helper()
	ctx := context.Background()
	app := application.Audit{Store: s}
	for i := 0; i < 55; i++ {
		if _, e := db.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES('admin','test.audit','synthetic')`); e != nil {
			t.Fatal(e)
		}
	}
	first, e := app.Events(ctx, actors[0], domain.AuditFilter{Action: "test.audit"})
	if e != nil || len(first) != 50 {
		t.Fatal(len(first), e)
	}
	second, e := app.Events(ctx, actors[0], domain.AuditFilter{Action: "test.audit", Before: first[49].ID})
	if e != nil || len(second) != 5 || second[0].ID >= first[49].ID {
		t.Fatal(second, e)
	}
	absent, e := app.Events(ctx, actors[0], domain.AuditFilter{ActorID: "doctor", Action: "test.audit"})
	if e != nil || len(absent) != 0 {
		t.Fatal(absent, e)
	}
	for _, role := range []string{"doctor", "nurse", "patient", "receptionist", "pharmacist", "accountant", "case_manager", "lab_technician"} {
		if _, e = app.Events(ctx, domain.Actor{ID: "admin", Role: role}, domain.AuditFilter{}); !errors.Is(e, domain.ErrForbidden) {
			t.Fatal(role, e)
		}
	}
	if _, e = app.Events(ctx, actors[0], domain.AuditFilter{Before: -1}); !errors.Is(e, domain.ErrValidation) {
		t.Fatal(e)
	}
	for _, sql := range []string{`UPDATE audit_event SET action='tampered' WHERE action='test.audit'`, `DELETE FROM audit_event WHERE action='test.audit'`, `TRUNCATE audit_event`} {
		if _, e = db.Exec(ctx, sql); e == nil {
			t.Fatal("audit mutation allowed", sql)
		}
	}
	var reviewed int
	if e = db.QueryRow(ctx, `SELECT count(*) FROM audit_event WHERE action='audit.reviewed'`).Scan(&reviewed); e != nil || reviewed != 3 {
		t.Fatal(reviewed, e)
	}
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.Header.Get("Cookie"), "session=")
		fmt.Fprintf(w, `{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, id, id, time.Now().Add(time.Hour).Format(time.RFC3339))
	}))
	defer auth.Close()
	handler := httpapi.Server{Audit: app, Actors: s, AuthURL: auth.URL, Client: auth.Client(), Origin: "http://hospital.test"}.Handler()
	for _, tc := range []struct {
		path, actor string
		status      int
	}{{"/v1/audit-events?action=test.audit", "admin", 200}, {"/v1/audit-events", "patient", 403}, {"/v1/audit-events?before=bad", "admin", 422}, {"/v1/audit-events?before=-1", "admin", 422}} {
		req := httptest.NewRequest("GET", tc.path, nil)
		req.Header.Set("Cookie", "session="+tc.actor)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Fatal(tc, w.Code, w.Body.String())
		}
	}
}
