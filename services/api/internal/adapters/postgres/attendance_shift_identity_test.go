package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
)

func testShiftIdentity(t *testing.T, store Store, actor domain.Actor) {
	t.Helper()
	ctx := context.Background()
	app := application.Attendance{Store: store, Now: time.Now}
	input := domain.ShiftInput{Name: "Configured default", Code: " cfg ", IsDefault: true, Active: true, StartTime: "08:00", EndTime: "17:00", HalfDayMinutes: 240, FullDayMinutes: 480}
	sh, err := app.CreateShift(ctx, actor, input)
	if err != nil || sh.Code != "CFG" || !sh.IsDefault {
		t.Fatalf("code/default not persisted: %+v %v", sh, err)
	}
	resolved, err := store.resolveShift(ctx, nil, "unassigned", "2030-01-01", "")
	if err != nil || resolved.ID != sh.ID {
		t.Fatalf("default still depends on name: %+v %v", resolved, err)
	}
	if _, err = store.resolveShift(ctx, nil, "unassigned", "2030-01-01", "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Fatal("invalid requested shift silently fell back")
	}
	input.Name = "Duplicate code"
	if _, err = app.CreateShift(ctx, actor, input); err == nil {
		t.Fatal("duplicate code accepted")
	}
	resolved, err = store.resolveShift(ctx, nil, "unassigned", "2030-01-01", "")
	if err != nil || resolved.ID != sh.ID {
		t.Fatal("failed create cleared default", err)
	}
	input.Name = "Renamed default"
	updated, err := app.UpdateShift(ctx, actor, sh.ID, input, sh.Version)
	if err != nil || !updated.IsDefault || updated.Code != "CFG" {
		t.Fatal("rename lost default/code", err)
	}
	if _, err = app.UpdateShift(ctx, actor, sh.ID, input, sh.Version); !errors.Is(err, domain.ErrStale) {
		t.Fatal("stale version accepted", err)
	}
	input.Active = false
	if _, err = app.UpdateShift(ctx, actor, sh.ID, input, updated.Version); !errors.Is(err, domain.ErrValidation) {
		t.Fatal("inactive default accepted", err)
	}
	// Concurrent administrators cannot leave two default shifts.
	var wg sync.WaitGroup
	for _, code := range []string{"CONCURRENT-A", "CONCURRENT-B"} {
		wg.Add(1)
		go func(code string) {
			defer wg.Done()
			i := input
			i.Code, i.Name, i.Active = code, code, true
			if _, e := app.CreateShift(ctx, actor, i); e != nil {
				t.Error(e)
			}
		}(code)
	}
	wg.Wait()
	var defaults int
	if err := store.DB.QueryRow(ctx, `SELECT count(*) FROM attendance_shift WHERE is_default`).Scan(&defaults); err != nil || defaults != 1 {
		t.Fatalf("expected exactly one default: %d %v", defaults, err)
	}
	shifts, err := app.ListShifts(ctx, actor)
	if err != nil {
		t.Fatal(err)
	}
	assigned := 0
	for _, shift := range shifts {
		assigned += shift.StaffCount
	}
	var staff int
	if err := store.DB.QueryRow(ctx, `SELECT count(*) FROM staff_access WHERE active AND role<>'patient'`).Scan(&staff); err != nil || assigned != staff {
		t.Fatalf("each active staff member must count under one resolved shift: %d want %d, %v", assigned, staff, err)
	}
}
