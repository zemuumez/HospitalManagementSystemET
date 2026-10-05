package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
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
