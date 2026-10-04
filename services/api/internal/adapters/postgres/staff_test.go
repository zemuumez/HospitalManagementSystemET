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

func testStaff(t *testing.T, db *pgxpool.Pool, s Store, actors []domain.Actor) {
	t.Helper()
	ctx := context.Background()
	app := application.Staff{Store: s, Now: time.Now}
	p, e := app.Profile(ctx, actors[0], "doctor")
	if e != nil || p.Version != 1 {
		t.Fatal(p, e)
	}
	input := domain.StaffProfileInput{Version: 1, Reason: "Verify staff record", Details: domain.StaffDetails{GivenName: "Synthetic", FamilyName: "Doctor", Gender: "unknown", DateOfBirth: "1980-02-03", Phone: "+251911111111", Designation: "Consultant", Qualification: "Test qualification", Specialty: "Test specialty", PostalCode: "00100"}}
	p, e = app.Update(ctx, actors[0], "doctor", input)
	if e != nil || p.Version != 2 || p.Name != "Synthetic Doctor" || p.Details.PostalCode != "00100" {
		t.Fatal(p, e)
	}
	if _, e = app.Update(ctx, actors[0], "doctor", input); !errors.Is(e, domain.ErrStale) {
		t.Fatal("lost update", e)
	}
	input.Version = 2
	var wg sync.WaitGroup
	outcomes := make(chan error, 2)
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := app.Update(ctx, actors[0], "doctor", input); outcomes <- err }()
	}
	wg.Wait()
	close(outcomes)
	success := 0
	stale := 0
	for err := range outcomes {
		if err == nil {
			success++
		} else if errors.Is(err, domain.ErrStale) {
			stale++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || stale != 1 {
		t.Fatal(success, stale)
	}
	bad := input
	bad.Details.DateOfBirth = "3000-01-01"
	if _, e = app.Update(ctx, actors[0], "doctor", bad); !errors.Is(e, domain.ErrValidation) {
		t.Fatal(e)
	}
	for _, role := range []string{"doctor", "nurse", "patient", "receptionist", "pharmacist", "accountant", "case_manager", "lab_technician"} {
		a := domain.Actor{ID: "admin", Role: role}
		if _, e = app.Profile(ctx, a, "doctor"); !errors.Is(e, domain.ErrForbidden) {
			t.Fatal(role, e)
		}
		if _, e = app.Update(ctx, a, "doctor", input); !errors.Is(e, domain.ErrForbidden) {
			t.Fatal(role, e)
		}
	}
	change := domain.StaffRoleInput{PreviousRole: "admin", Role: "nurse", Reason: "New staff assignment"}
	if e = app.ChangeRole(ctx, actors[0], "admin", change); !errors.Is(e, domain.ErrStale) {
		t.Fatal("self demotion", e)
	}
	change.PreviousRole = "doctor"
	if e = app.ChangeRole(ctx, actors[0], "doctor", change); !errors.Is(e, domain.ErrStale) {
		t.Fatal("active clinical role changed", e)
	}
	change.PreviousRole = "patient"
	if e = app.ChangeRole(ctx, actors[0], "patient", change); !errors.Is(e, domain.ErrStale) {
		t.Fatal("portal owner role changed", e)
	}
	if _, e = db.Exec(ctx, `INSERT INTO "user"(id,name,email) VALUES('role-test','Role Test','role-test@example.test'); INSERT INTO staff_access(user_id,role) VALUES('role-test','receptionist'); INSERT INTO session(id,token,"userId","expiresAt") VALUES('role-session','role-token','role-test',now()+interval '1 hour')`); e != nil {
		t.Fatal(e)
	}
	change.PreviousRole = "receptionist"
	if e = app.ChangeRole(ctx, actors[0], "role-test", change); e != nil {
		t.Fatal(e)
	}
	var sessions int
	if e = db.QueryRow(ctx, `SELECT count(*) FROM session WHERE "userId"='role-test'`).Scan(&sessions); e != nil || sessions != 0 {
		t.Fatal(sessions, e)
	}
	if e = app.ChangeRole(ctx, actors[0], "role-test", change); !errors.Is(e, domain.ErrStale) {
		t.Fatal("stale role change", e)
	}
	for _, sql := range []string{`DELETE FROM staff_role_event`, `UPDATE staff_profile_revision SET reason='tampered'`} {
		if _, e = db.Exec(ctx, sql); e == nil {
			t.Fatal("history mutable")
		}
	}
	auth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.Header.Get("Cookie"), "session=")
		fmt.Fprintf(w, `{"user":{"id":%q},"session":{"userId":%q,"expiresAt":%q}}`, id, id, time.Now().Add(time.Hour).Format(time.RFC3339))
	}))
	defer auth.Close()
	handler := httpapi.Server{Staff: app, Actors: s, AuthURL: auth.URL, Client: auth.Client(), Origin: "http://hospital.test"}.Handler()
	for _, tc := range []struct {
		method, path, actor, body, origin string
		status                            int
	}{{"GET", "/v1/staff-profiles/doctor", "admin", "", "", 200}, {"GET", "/v1/staff-profiles/doctor", "patient", "", "", 403}, {"PATCH", "/v1/staff-profiles/doctor", "admin", "{}", "https://evil.test", 403}, {"PATCH", "/v1/staff-profiles/doctor", "admin", `{"unexpected":true}`, "http://hospital.test", 400}, {"PATCH", "/v1/staff-profiles/role-test/role", "admin", `{"previousRole":"nurse","role":"pharmacist","reason":"New duties"}`, "http://hospital.test", 200}} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Cookie", "session="+tc.actor)
		req.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Fatal(tc, w.Code, w.Body.String())
		}
	}
}
