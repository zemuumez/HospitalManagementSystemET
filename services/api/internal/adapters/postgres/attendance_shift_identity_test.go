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

func testAttendanceLeavesAndAssignments(t *testing.T, store Store, actors []domain.Actor) {
	t.Helper()
	ctx := context.Background()
	app := application.Attendance{Store: store, Now: time.Now}

	admin := actors[0]
	doctor := actors[1]
	patient := actors[3]

	// 1. Enhanced duty assignment with note and active flag
	sh, err := app.CreateShift(ctx, admin, domain.ShiftInput{
		Name: "Special Ward Shift",
		Code: "WARD-SPEC",
		StartTime: "14:00",
		EndTime: "22:00",
		HalfDayMinutes: 240,
		FullDayMinutes: 480,
		Active: true,
	})
	if err != nil {
		t.Fatalf("failed to create shift: %v", err)
	}

	trueVal := true
	as, err := app.AssignShift(ctx, admin, domain.ShiftAssignmentInput{
		StaffID: doctor.ID,
		ShiftID: sh.ID,
		EffectiveFrom: "2030-01-01",
		Note: "Covering emergency ward",
		Active: &trueVal,
	})
	if err != nil {
		t.Fatalf("failed to assign shift: %v", err)
	}
	if as.Note != "Covering emergency ward" || !as.Active {
		t.Fatalf("expected note and active=true in assignment, got note=%q active=%v", as.Note, as.Active)
	}

	assignments, err := app.ListShiftAssignments(ctx, admin, doctor.ID)
	if err != nil || len(assignments) == 0 {
		t.Fatalf("failed to list assignments: %v", err)
	}
	found := false
	for _, a := range assignments {
		if a.ID == as.ID && a.Note == "Covering emergency ward" && a.Active {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created assignment with note not found in list")
	}

	// 2. Inactive assignment should not resolve
	sh2, err := app.CreateShift(ctx, admin, domain.ShiftInput{
		Name: "Inactive Shift Assignment",
		Code: "INACT-SPEC",
		StartTime: "10:00",
		EndTime: "18:00",
		HalfDayMinutes: 240,
		FullDayMinutes: 480,
		Active: true,
	})
	if err != nil {
		t.Fatalf("failed to create second shift: %v", err)
	}

	falseVal := false
	_, err = app.AssignShift(ctx, admin, domain.ShiftAssignmentInput{
		StaffID: doctor.ID,
		ShiftID: sh2.ID,
		EffectiveFrom: "2030-01-02",
		Note: "Inactive assignment",
		Active: &falseVal,
	})
	if err != nil {
		t.Fatalf("failed to assign inactive shift: %v", err)
	}

	resolved, err := store.resolveShift(ctx, nil, doctor.ID, "2030-01-03", "")
	if err != nil {
		t.Fatalf("failed to resolve shift: %v", err)
	}
	if resolved.ID == sh2.ID {
		t.Fatalf("inactive assignment was unexpectedly resolved: %v", resolved)
	}

	// 3. Leave Requests Lifecycle
	// Patient cannot access leave requests
	if _, _, err := app.ListLeaveRequests(ctx, patient, domain.LeaveFilter{}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden for patient listing leaves, got %v", err)
	}
	if _, err := app.CreateLeaveRequest(ctx, patient, domain.LeaveRequestInput{
		FromDate: "2030-05-10",
		ToDate: "2030-05-12",
		LeaveType: "casual",
		Reason: "Sick",
	}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden for patient creating leave, got %v", err)
	}

	// Doctor submits leave request (3 days)
	lr, err := app.CreateLeaveRequest(ctx, doctor, domain.LeaveRequestInput{
		FromDate: "2030-05-10",
		ToDate: "2030-05-12",
		LeaveType: "casual",
		Reason: "Family event attendance",
	})
	if err != nil {
		t.Fatalf("failed to create leave request: %v", err)
	}
	if lr.Days != 3 || lr.Status != "pending" || lr.Version != 1 || lr.StaffID != doctor.ID {
		t.Fatalf("invalid leave request data: %+v", lr)
	}

	// Doctor sees own leave request
	docLeaves, total, err := app.ListLeaveRequests(ctx, doctor, domain.LeaveFilter{})
	if err != nil || total < 1 || len(docLeaves) == 0 {
		t.Fatalf("doctor cannot list own leaves: %v total=%d", err, total)
	}

	// Doctor cannot approve own request
	if _, err := app.UpdateLeaveStatus(ctx, doctor, lr.ID, domain.LeaveApprovalInput{
		Status: "approved",
		ApproverNotes: "Self approval attempt",
		Version: 1,
	}); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden for non-admin approving leave, got %v", err)
	}

	// Admin approves request
	approved, err := app.UpdateLeaveStatus(ctx, admin, lr.ID, domain.LeaveApprovalInput{
		Status: "approved",
		ApproverNotes: "Approved by administration",
		Version: 1,
	})
	if err != nil {
		t.Fatalf("admin approval failed: %v", err)
	}
	if approved.Status != "approved" || approved.Version != 2 || approved.ApproverID == nil || *approved.ApproverID != admin.ID {
		t.Fatalf("approval state mismatch: %+v", approved)
	}

	// Stale version rejection
	if _, err := app.UpdateLeaveStatus(ctx, admin, lr.ID, domain.LeaveApprovalInput{
		Status: "rejected",
		ApproverNotes: "Stale attempt",
		Version: 1,
	}); !errors.Is(err, domain.ErrStale) {
		t.Fatalf("expected ErrStale on version 1, got %v", err)
	}

	// Doctor submits another leave request and cancels it
	lr2, err := app.CreateLeaveRequest(ctx, doctor, domain.LeaveRequestInput{
		FromDate: "2030-06-01",
		ToDate: "2030-06-02",
		LeaveType: "sick",
		Reason: "Medical recovery",
	})
	if err != nil {
		t.Fatalf("failed to create leave request 2: %v", err)
	}

	cancelled, err := app.UpdateLeaveStatus(ctx, doctor, lr2.ID, domain.LeaveApprovalInput{
		Status: "cancelled",
		ApproverNotes: "Feeling better, cancelling",
		Version: 1,
	})
	if err != nil {
		t.Fatalf("doctor cancel failed: %v", err)
	}
	if cancelled.Status != "cancelled" || cancelled.Version != 2 {
		t.Fatalf("expected cancelled status with version 2: %+v", cancelled)
	}

	// Cannot transition a finalized/cancelled leave request
	if _, err := app.UpdateLeaveStatus(ctx, admin, lr2.ID, domain.LeaveApprovalInput{
		Status: "approved",
		ApproverNotes: "Attempting to approve cancelled",
		Version: 2,
	}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation when modifying cancelled request, got %v", err)
	}
}

