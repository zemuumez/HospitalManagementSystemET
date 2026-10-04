package domain

import (
	"testing"
	"time"
)

func TestAttendanceCalculations(t *testing.T) {
	dayShift := Shift{
		ID:                   "11111111-1111-1111-1111-111111111111",
		Name:                 "Day Shift",
		StartTime:            "08:00:00",
		EndTime:              "17:00:00",
		GracePeriodMinutes:   15,
		BreakDurationMinutes: 60,
		HalfDayMinutes:       240,
		FullDayMinutes:       480,
		IsOvernight:          false,
		Active:               true,
	}

	nightShift := Shift{
		ID:                   "22222222-2222-2222-2222-222222222222",
		Name:                 "Night Shift",
		StartTime:            "20:00:00",
		EndTime:              "05:00:00",
		GracePeriodMinutes:   15,
		BreakDurationMinutes: 60,
		HalfDayMinutes:       240,
		FullDayMinutes:       480,
		IsOvernight:          true,
		Active:               true,
	}

	workDate := "2026-10-05"

	t.Run("on_time_full_day", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 8, 0, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 5, 17, 0, 0, 0, HospitalLocation)
		status, late, early, worked, ot, err := CalculateAttendance(dayShift, workDate, in, &out, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "present" || late != 0 || early != 0 || worked != 480 || ot != 0 {
			t.Fatalf("unexpected results: status=%s, late=%d, early=%d, worked=%d, ot=%d", status, late, early, worked, ot)
		}
	})

	t.Run("within_grace_period", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 8, 15, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 5, 17, 15, 0, 0, HospitalLocation)
		status, late, early, worked, ot, err := CalculateAttendance(dayShift, workDate, in, &out, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "present" || late != 0 || early != 0 || worked != 480 || ot != 0 {
			t.Fatalf("grace period should not be marked late: status=%s, late=%d", status, late)
		}
	})

	t.Run("late_check_in", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 8, 30, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 5, 17, 30, 0, 0, HospitalLocation)
		status, late, early, worked, ot, err := CalculateAttendance(dayShift, workDate, in, &out, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "late" || late != 30 || early != 0 || worked != 480 || ot != 0 {
			t.Fatalf("should be marked late: status=%s, late=%d", status, late)
		}
	})

	t.Run("early_departure", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 8, 0, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 5, 16, 0, 0, 0, HospitalLocation)
		status, late, early, worked, ot, err := CalculateAttendance(dayShift, workDate, in, &out, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "half_day" || late != 0 || early != 60 || worked != 420 || ot != 0 {
			t.Fatalf("early departure check: status=%s, early=%d, worked=%d", status, early, worked)
		}
	})

	t.Run("overtime", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 8, 0, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 5, 19, 0, 0, 0, HospitalLocation)
		status, late, early, worked, ot, err := CalculateAttendance(dayShift, workDate, in, &out, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "present" || late != 0 || early != 0 || worked != 600 || ot != 120 {
			t.Fatalf("overtime check: status=%s, worked=%d, ot=%d", status, worked, ot)
		}
	})

	t.Run("half_day_worked", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 8, 0, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 5, 13, 0, 0, 0, HospitalLocation)
		status, _, _, worked, _, err := CalculateAttendance(dayShift, workDate, in, &out, 30)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "half_day" || worked != 270 {
			t.Fatalf("half day check: status=%s, worked=%d", status, worked)
		}
	})

	t.Run("absent_under_half_day", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 8, 0, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 5, 10, 0, 0, 0, HospitalLocation)
		status, _, _, worked, _, err := CalculateAttendance(dayShift, workDate, in, &out, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "absent" || worked != 120 {
			t.Fatalf("absent check: status=%s, worked=%d", status, worked)
		}
	})

	t.Run("checked_in_only", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 8, 5, 0, 0, HospitalLocation)
		status, late, early, worked, ot, err := CalculateAttendance(dayShift, workDate, in, nil, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "checked_in" || late != 0 || early != 0 || worked != 0 || ot != 0 {
			t.Fatalf("open check-in check: status=%s", status)
		}
	})

	t.Run("overnight_shift_normal", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 20, 0, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 6, 5, 0, 0, 0, HospitalLocation)
		status, late, early, worked, ot, err := CalculateAttendance(nightShift, workDate, in, &out, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "present" || late != 0 || early != 0 || worked != 480 || ot != 0 {
			t.Fatalf("overnight shift check: status=%s, late=%d, early=%d, worked=%d", status, late, early, worked)
		}
	})

	t.Run("overnight_shift_late_and_early_out", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 20, 30, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 6, 4, 30, 0, 0, HospitalLocation)
		status, late, early, worked, _, err := CalculateAttendance(nightShift, workDate, in, &out, 60)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != "half_day" || late != 30 || early != 30 || worked != 420 {
			t.Fatalf("overnight late/early check: status=%s, late=%d, early=%d, worked=%d", status, late, early, worked)
		}
	})

	t.Run("invalid_time_sequence", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 17, 0, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 5, 8, 0, 0, 0, HospitalLocation)
		_, _, _, _, _, err := CalculateAttendance(dayShift, workDate, in, &out, 0)
		if err != ErrValidation {
			t.Fatalf("expected ErrValidation for checkout before checkin, got: %v", err)
		}
	})

	t.Run("invalid_break_duration", func(t *testing.T) {
		in := time.Date(2026, 10, 5, 8, 0, 0, 0, HospitalLocation)
		out := time.Date(2026, 10, 5, 17, 0, 0, 0, HospitalLocation)
		_, _, _, _, _, err := CalculateAttendance(dayShift, workDate, in, &out, 600)
		if err != ErrValidation {
			t.Fatalf("expected ErrValidation for break longer than elapsed time, got: %v", err)
		}
	})
}

func TestAttendanceValidation(t *testing.T) {
	shiftInput := ShiftInput{
		Name:                 "Special Shift",
		StartTime:            "07:30",
		EndTime:              "15:30",
		GracePeriodMinutes:   15,
		BreakDurationMinutes: 45,
		HalfDayMinutes:       240,
		FullDayMinutes:       480,
		Active:               true,
	}
	if err := shiftInput.Validate(); err != nil {
		t.Fatalf("valid shift should pass: %v", err)
	}

	badShift := shiftInput
	badShift.HalfDayMinutes = 600
	badShift.FullDayMinutes = 480
	if err := badShift.Validate(); err != ErrValidation {
		t.Fatalf("half day > full day must fail")
	}

	badTime := shiftInput
	badTime.StartTime = "25:00"
	if err := badTime.Validate(); err != ErrValidation {
		t.Fatalf("invalid time format must fail")
	}

	assignInput := ShiftAssignmentInput{
		StaffID:       "user-1",
		ShiftID:       "11111111-1111-1111-1111-111111111111",
		EffectiveFrom: "2026-10-01",
	}
	if err := assignInput.Validate(); err != nil {
		t.Fatalf("valid assignment should pass: %v", err)
	}

	badAssign := assignInput
	to := "2026-09-01"
	badAssign.EffectiveTo = &to
	if err := badAssign.Validate(); err != ErrValidation {
		t.Fatalf("effectiveTo before effectiveFrom must fail")
	}

	adminInput := AdminAttendanceInput{
		StaffID:           "user-1",
		WorkDate:          "2026-10-05",
		ShiftID:           "11111111-1111-1111-1111-111111111111",
		CheckInAt:         time.Now(),
		TotalBreakMinutes: 60,
	}
	if err := adminInput.Validate(); err != nil {
		t.Fatalf("valid admin input should pass: %v", err)
	}

	correctionInput := AdminCorrectionInput{
		Version:           1,
		ShiftID:           "11111111-1111-1111-1111-111111111111",
		CheckInAt:         time.Now(),
		TotalBreakMinutes: 60,
		Reason:            "Badge reader malfunction",
	}
	if err := correctionInput.Validate(); err != nil {
		t.Fatalf("valid correction input should pass: %v", err)
	}

	badCorrection := correctionInput
	badCorrection.Reason = ""
	if err := badCorrection.Validate(); err != ErrValidation {
		t.Fatalf("correction without reason must fail")
	}

	approvalInput := ApprovalInput{
		Status: "approved",
		Reason: "Approved by medical director",
	}
	if err := approvalInput.Validate(); err != nil {
		t.Fatalf("valid approval input should pass: %v", err)
	}

	badApproval := ApprovalInput{
		Status: "invalid_status",
	}
	if err := badApproval.Validate(); err != ErrValidation {
		t.Fatalf("invalid approval status must fail")
	}
}
