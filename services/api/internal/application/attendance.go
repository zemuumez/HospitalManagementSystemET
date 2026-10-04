package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
	"time"
)

type AttendanceStore interface {
	ListShifts(ctx context.Context, actor domain.Actor) ([]domain.Shift, error)
	CreateShift(ctx context.Context, actor domain.Actor, input domain.ShiftInput) (domain.Shift, error)
	UpdateShift(ctx context.Context, actor domain.Actor, id string, input domain.ShiftInput, version int) (domain.Shift, error)

	ListShiftAssignments(ctx context.Context, actor domain.Actor, staffID string) ([]domain.ShiftAssignment, error)
	AssignShift(ctx context.Context, actor domain.Actor, input domain.ShiftAssignmentInput) (domain.ShiftAssignment, error)

	GetTodayRecord(ctx context.Context, actor domain.Actor, staffID string, workDate string) (*domain.AttendanceRecord, error)
	ClockIn(ctx context.Context, actor domain.Actor, staffID string, shiftID string, now time.Time, key string) (domain.AttendanceRecord, error)
	ClockOut(ctx context.Context, actor domain.Actor, staffID string, now time.Time, key string) (domain.AttendanceRecord, error)
	StartBreak(ctx context.Context, actor domain.Actor, staffID string, now time.Time, reason string, key string) (domain.BreakRecord, error)
	EndBreak(ctx context.Context, actor domain.Actor, staffID string, now time.Time, key string) (domain.BreakRecord, error)

	ListAttendanceRecords(ctx context.Context, actor domain.Actor, filter domain.AttendanceFilter) ([]domain.AttendanceRecord, int, error)
	GetAttendanceRecord(ctx context.Context, actor domain.Actor, id string) (domain.AttendanceRecord, error)
	GetAttendanceHistory(ctx context.Context, actor domain.Actor, id string) (domain.AttendanceHistory, error)

	AdminCreateRecord(ctx context.Context, actor domain.Actor, input domain.AdminAttendanceInput, key string) (domain.AttendanceRecord, error)
	AdminCorrectRecord(ctx context.Context, actor domain.Actor, id string, input domain.AdminCorrectionInput) (domain.AttendanceRecord, error)
	AdminUpdateApproval(ctx context.Context, actor domain.Actor, id string, input domain.ApprovalInput) (domain.AttendanceRecord, error)

	GetAttendanceSummary(ctx context.Context, actor domain.Actor, date string) (domain.AttendanceSummary, error)
}

type Attendance struct {
	Store AttendanceStore
	Now   func() time.Time
}

func (a Attendance) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

func (a Attendance) ListShifts(ctx context.Context, actor domain.Actor) ([]domain.Shift, error) {
	if !actor.Can("attendance.read_own") && !actor.Can("attendance.manage") {
		return nil, domain.ErrForbidden
	}
	return a.Store.ListShifts(ctx, actor)
}

func (a Attendance) CreateShift(ctx context.Context, actor domain.Actor, input domain.ShiftInput) (domain.Shift, error) {
	if !actor.Can("attendance.manage") {
		return domain.Shift{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.Shift{}, err
	}
	return a.Store.CreateShift(ctx, actor, input)
}

func (a Attendance) UpdateShift(ctx context.Context, actor domain.Actor, id string, input domain.ShiftInput, version int) (domain.Shift, error) {
	if !actor.Can("attendance.manage") {
		return domain.Shift{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || version < 1 {
		return domain.Shift{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.Shift{}, err
	}
	return a.Store.UpdateShift(ctx, actor, id, input, version)
}

func (a Attendance) ListShiftAssignments(ctx context.Context, actor domain.Actor, staffID string) ([]domain.ShiftAssignment, error) {
	staffID = strings.TrimSpace(staffID)
	if staffID != "" && staffID == actor.ID {
		if !actor.Can("attendance.read_own") {
			return nil, domain.ErrForbidden
		}
	} else {
		if !actor.Can("attendance.manage") {
			return nil, domain.ErrForbidden
		}
	}
	return a.Store.ListShiftAssignments(ctx, actor, staffID)
}

func (a Attendance) AssignShift(ctx context.Context, actor domain.Actor, input domain.ShiftAssignmentInput) (domain.ShiftAssignment, error) {
	if !actor.Can("attendance.manage") {
		return domain.ShiftAssignment{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.ShiftAssignment{}, err
	}
	return a.Store.AssignShift(ctx, actor, input)
}

func (a Attendance) Today(ctx context.Context, actor domain.Actor) (*domain.AttendanceRecord, error) {
	if !actor.Can("attendance.read_own") {
		return nil, domain.ErrForbidden
	}
	workDate := a.now().In(domain.HospitalLocation).Format("2006-01-02")
	return a.Store.GetTodayRecord(ctx, actor, actor.ID, workDate)
}

func (a Attendance) ClockIn(ctx context.Context, actor domain.Actor, shiftID string, key string) (domain.AttendanceRecord, error) {
	if !actor.Can("attendance.clock") {
		return domain.AttendanceRecord{}, domain.ErrForbidden
	}
	if shiftID != "" && !domain.UUIDPattern.MatchString(shiftID) {
		return domain.AttendanceRecord{}, domain.ErrValidation
	}
	return a.Store.ClockIn(ctx, actor, actor.ID, shiftID, a.now(), key)
}

func (a Attendance) ClockOut(ctx context.Context, actor domain.Actor, key string) (domain.AttendanceRecord, error) {
	if !actor.Can("attendance.clock") {
		return domain.AttendanceRecord{}, domain.ErrForbidden
	}
	return a.Store.ClockOut(ctx, actor, actor.ID, a.now(), key)
}

func (a Attendance) StartBreak(ctx context.Context, actor domain.Actor, reason string, key string) (domain.BreakRecord, error) {
	if !actor.Can("attendance.clock") {
		return domain.BreakRecord{}, domain.ErrForbidden
	}
	if len([]rune(reason)) > 500 {
		return domain.BreakRecord{}, domain.ErrValidation
	}
	return a.Store.StartBreak(ctx, actor, actor.ID, a.now(), reason, key)
}

func (a Attendance) EndBreak(ctx context.Context, actor domain.Actor, key string) (domain.BreakRecord, error) {
	if !actor.Can("attendance.clock") {
		return domain.BreakRecord{}, domain.ErrForbidden
	}
	return a.Store.EndBreak(ctx, actor, actor.ID, a.now(), key)
}

func (a Attendance) ListRecords(ctx context.Context, actor domain.Actor, filter domain.AttendanceFilter) ([]domain.AttendanceRecord, int, error) {
	if actor.Can("attendance.manage") {
		// Admin can view any or filter by specific staff
	} else if actor.Can("attendance.read_own") {
		// Non-admin can only view their own records
		filter.StaffID = actor.ID
	} else {
		return nil, 0, domain.ErrForbidden
	}

	if filter.WorkDate != "" {
		if _, err := time.ParseInLocation("2006-01-02", filter.WorkDate, domain.HospitalLocation); err != nil {
			return nil, 0, domain.ErrValidation
		}
	}
	if filter.FromDate != "" {
		if _, err := time.ParseInLocation("2006-01-02", filter.FromDate, domain.HospitalLocation); err != nil {
			return nil, 0, domain.ErrValidation
		}
	}
	if filter.ToDate != "" {
		if _, err := time.ParseInLocation("2006-01-02", filter.ToDate, domain.HospitalLocation); err != nil {
			return nil, 0, domain.ErrValidation
		}
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 25
	}
	return a.Store.ListAttendanceRecords(ctx, actor, filter)
}

func (a Attendance) GetRecord(ctx context.Context, actor domain.Actor, id string) (domain.AttendanceRecord, error) {
	if !domain.UUIDPattern.MatchString(id) {
		return domain.AttendanceRecord{}, domain.ErrValidation
	}
	rec, err := a.Store.GetAttendanceRecord(ctx, actor, id)
	if err != nil {
		return domain.AttendanceRecord{}, err
	}
	if !actor.Can("attendance.manage") && rec.StaffID != actor.ID {
		return domain.AttendanceRecord{}, domain.ErrForbidden
	}
	return rec, nil
}

func (a Attendance) GetHistory(ctx context.Context, actor domain.Actor, id string) (domain.AttendanceHistory, error) {
	if !domain.UUIDPattern.MatchString(id) {
		return domain.AttendanceHistory{}, domain.ErrValidation
	}
	hist, err := a.Store.GetAttendanceHistory(ctx, actor, id)
	if err != nil {
		return domain.AttendanceHistory{}, err
	}
	if !actor.Can("attendance.manage") && hist.Record.StaffID != actor.ID {
		return domain.AttendanceHistory{}, domain.ErrForbidden
	}
	return hist, nil
}

func (a Attendance) AdminCreate(ctx context.Context, actor domain.Actor, input domain.AdminAttendanceInput, key string) (domain.AttendanceRecord, error) {
	if !actor.Can("attendance.manage") {
		return domain.AttendanceRecord{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.AttendanceRecord{}, err
	}
	return a.Store.AdminCreateRecord(ctx, actor, input, key)
}

func (a Attendance) AdminCorrect(ctx context.Context, actor domain.Actor, id string, input domain.AdminCorrectionInput) (domain.AttendanceRecord, error) {
	if !actor.Can("attendance.manage") {
		return domain.AttendanceRecord{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.AttendanceRecord{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.AttendanceRecord{}, err
	}
	return a.Store.AdminCorrectRecord(ctx, actor, id, input)
}

func (a Attendance) AdminUpdateApproval(ctx context.Context, actor domain.Actor, id string, input domain.ApprovalInput) (domain.AttendanceRecord, error) {
	if !actor.Can("attendance.manage") {
		return domain.AttendanceRecord{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.AttendanceRecord{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.AttendanceRecord{}, err
	}
	return a.Store.AdminUpdateApproval(ctx, actor, id, input)
}

func (a Attendance) Summary(ctx context.Context, actor domain.Actor, date string) (domain.AttendanceSummary, error) {
	if !actor.Can("attendance.manage") {
		return domain.AttendanceSummary{}, domain.ErrForbidden
	}
	if date == "" {
		date = a.now().In(domain.HospitalLocation).Format("2006-01-02")
	} else if _, err := time.ParseInLocation("2006-01-02", date, domain.HospitalLocation); err != nil {
		return domain.AttendanceSummary{}, domain.ErrValidation
	}
	return a.Store.GetAttendanceSummary(ctx, actor, date)
}
