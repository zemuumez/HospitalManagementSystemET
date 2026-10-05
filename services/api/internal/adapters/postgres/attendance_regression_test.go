package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"sync"
	"testing"
	"time"
)

func testAttendanceRetry(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor) {
	t.Helper()
	ctx := context.Background()
	now := time.Date(2030, 1, 1, 8, 0, 0, 0, domain.HospitalLocation)
	app := application.Attendance{Store: store, Now: func() time.Time { return now }}
	actor := actors[1]
	first, err := app.ClockIn(ctx, actor, "", "retry-clock-in")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Minute)
	replay, err := app.ClockIn(ctx, actor, "", "retry-clock-in")
	if err != nil || replay.ID != first.ID || !replay.CheckInAt.Equal(first.CheckInAt) {
		t.Fatalf("clock-in replay: %v, %v", replay, err)
	}
	var wg sync.WaitGroup
	for n := 0; n < 8; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := app.ClockIn(ctx, actor, "", "retry-clock-in")
			if e != nil || r.ID != first.ID {
				t.Errorf("concurrent replay: %v", e)
			}
		}()
	}
	wg.Wait()
	if _, err = app.ClockIn(ctx, actor, "22222222-2222-2222-2222-222222222222", "retry-clock-in"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("changed payload must conflict: %v", err)
	}
	start, err := app.StartBreak(ctx, actor, "meal", "retry-break-start")
	if err != nil {
		t.Fatal(err)
	}
	again, err := app.StartBreak(ctx, actor, "meal", "retry-break-start")
	if err != nil || again.ID != start.ID {
		t.Fatalf("break start replay: %v", err)
	}
	now = now.Add(30 * time.Minute)
	end, err := app.EndBreak(ctx, actor, "retry-break-end")
	if err != nil {
		t.Fatal(err)
	}
	endAgain, err := app.EndBreak(ctx, actor, "retry-break-end")
	if err != nil || endAgain.ID != end.ID {
		t.Fatalf("break end replay: %v", err)
	}
	now = now.Add(8 * time.Hour)
	out, err := app.ClockOut(ctx, actor, "retry-clock-out")
	if err != nil {
		t.Fatal(err)
	}
	outAgain, err := app.ClockOut(ctx, actor, "retry-clock-out")
	if err != nil || outAgain.ID != out.ID || outAgain.WorkedMinutes != out.WorkedMinutes {
		t.Fatalf("clock-out replay: %v", err)
	}
	input := domain.AdminAttendanceInput{StaffID: actor.ID, WorkDate: "2029-12-30", ShiftID: first.ShiftID, CheckInAt: first.CheckInAt.AddDate(0, 0, -2), Reason: "test manual entry"}
	manual, err := app.AdminCreate(ctx, actors[0], input, "retry-admin")
	if err != nil {
		t.Fatal(err)
	}
	manualAgain, err := app.AdminCreate(ctx, actors[0], input, "retry-admin")
	if err != nil || manualAgain.ID != manual.ID {
		t.Fatalf("manual entry replay: %v", err)
	}
}

func testAttendanceLifecycle(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor) {
	ctx := context.Background()
	actor := actors[2]
	now := time.Date(2040, 1, 1, 20, 0, 0, 0, domain.HospitalLocation)
	app := application.Attendance{Store: store, Now: func() time.Time { return now }}
	first, err := app.ClockIn(ctx, actor, "22222222-2222-2222-2222-222222222222", "overnight-first")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(5 * time.Hour)
	t.Run("reject overlapping overnight record", func(t *testing.T) {
		if _, err := app.ClockIn(ctx, actor, "", "overnight-second"); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("second open attendance accepted: %v", err)
		}
	})
	t.Run("today retains previous night open record", func(t *testing.T) {
		today, err := app.Today(ctx, actor)
		if err != nil || today == nil || today.ID != first.ID {
			t.Fatalf("overnight record unavailable: %v %v", today, err)
		}
	})
	t.Run("reject approval of unfinished attendance", func(t *testing.T) {
		if _, err := app.AdminUpdateApproval(ctx, actors[0], first.ID, domain.ApprovalInput{Status: "approved"}); !errors.Is(err, domain.ErrConflict) {
			t.Fatalf("open record approved: %v", err)
		}
	})
	now = now.Add(4 * time.Hour)
	closed, err := app.ClockOut(ctx, actor, "overnight-out")
	if err != nil {
		t.Fatal(err)
	}
	approved, err := app.AdminUpdateApproval(ctx, actors[0], first.ID, domain.ApprovalInput{Status: "approved"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = app.ClockOut(ctx, actor, "new-out-after-approval"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("approved record changed by clock-out: %v", err)
	}
	current, err := app.GetRecord(ctx, actor, first.ID)
	if err != nil || current.Version != approved.Version || current.WorkedMinutes != closed.WorkedMinutes {
		t.Fatalf("approved totals changed: %v", err)
	}
}
