package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
	"time"
)

func (s Store) ListShifts(ctx context.Context, actor domain.Actor) ([]domain.Shift, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, name, to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'), grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active, version, created_at, updated_at FROM attendance_shift ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []domain.Shift
	for rows.Next() {
		var sh domain.Shift
		if err = rows.Scan(&sh.ID, &sh.Name, &sh.StartTime, &sh.EndTime, &sh.GracePeriodMinutes, &sh.BreakDurationMinutes, &sh.HalfDayMinutes, &sh.FullDayMinutes, &sh.IsOvernight, &sh.Active, &sh.Version, &sh.CreatedAt, &sh.UpdatedAt); err != nil {
			return nil, err
		}
		shifts = append(shifts, sh)
	}
	return shifts, rows.Err()
}

func (s Store) CreateShift(ctx context.Context, actor domain.Actor, input domain.ShiftInput) (domain.Shift, error) {
	var sh domain.Shift
	startH, startM, _ := parseTimeOfDayHelper(input.StartTime)
	endH, endM, _ := parseTimeOfDayHelper(input.EndTime)
	isOvernight := (endH < startH) || (endH == startH && endM <= startM)

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return sh, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `INSERT INTO attendance_shift (name, start_time, end_time, grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active) VALUES ($1, $2::time, $3::time, $4, $5, $6, $7, $8, $9) RETURNING id, name, to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'), grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active, version, created_at, updated_at`,
		input.Name, input.StartTime, input.EndTime, input.GracePeriodMinutes, input.BreakDurationMinutes, input.HalfDayMinutes, input.FullDayMinutes, isOvernight, input.Active).
		Scan(&sh.ID, &sh.Name, &sh.StartTime, &sh.EndTime, &sh.GracePeriodMinutes, &sh.BreakDurationMinutes, &sh.HalfDayMinutes, &sh.FullDayMinutes, &sh.IsOvernight, &sh.Active, &sh.Version, &sh.CreatedAt, &sh.UpdatedAt)
	if err != nil {
		return sh, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'attendance.shift_created', $2)`, actor.ID, sh.ID); err != nil {
		return sh, err
	}
	return sh, tx.Commit(ctx)
}

func (s Store) UpdateShift(ctx context.Context, actor domain.Actor, id string, input domain.ShiftInput, version int) (domain.Shift, error) {
	var sh domain.Shift
	startH, startM, _ := parseTimeOfDayHelper(input.StartTime)
	endH, endM, _ := parseTimeOfDayHelper(input.EndTime)
	isOvernight := (endH < startH) || (endH == startH && endM <= startM)

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return sh, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `UPDATE attendance_shift SET name=$1, start_time=$2::time, end_time=$3::time, grace_period_minutes=$4, break_duration_minutes=$5, half_day_minutes=$6, full_day_minutes=$7, is_overnight=$8, active=$9, version=version+1, updated_at=now() WHERE id=$10 AND version=$11 RETURNING id, name, to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'), grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active, version, created_at, updated_at`,
		input.Name, input.StartTime, input.EndTime, input.GracePeriodMinutes, input.BreakDurationMinutes, input.HalfDayMinutes, input.FullDayMinutes, isOvernight, input.Active, id, version).
		Scan(&sh.ID, &sh.Name, &sh.StartTime, &sh.EndTime, &sh.GracePeriodMinutes, &sh.BreakDurationMinutes, &sh.HalfDayMinutes, &sh.FullDayMinutes, &sh.IsOvernight, &sh.Active, &sh.Version, &sh.CreatedAt, &sh.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		var currentVer int
		if e := tx.QueryRow(ctx, `SELECT version FROM attendance_shift WHERE id=$1`, id).Scan(&currentVer); e == nil {
			return sh, domain.ErrStale
		}
		return sh, domain.ErrNotFound
	}
	if err != nil {
		return sh, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'attendance.shift_updated', $2)`, actor.ID, sh.ID); err != nil {
		return sh, err
	}
	return sh, tx.Commit(ctx)
}

func parseTimeOfDayHelper(s string) (int, int, error) {
	var h, m, sec int
	_, err := fmt.Sscanf(s, "%02d:%02d:%02d", &h, &m, &sec)
	if err != nil {
		_, err = fmt.Sscanf(s, "%02d:%02d", &h, &m)
	}
	return h, m, err
}

func (s Store) ListShiftAssignments(ctx context.Context, actor domain.Actor, staffID string) ([]domain.ShiftAssignment, error) {
	rows, err := s.DB.Query(ctx, `SELECT a.id, a.staff_id, u.name, a.shift_id, s.name, a.effective_from::text, a.effective_to::text, a.created_at FROM attendance_shift_assignment a JOIN "user" u ON u.id=a.staff_id JOIN attendance_shift s ON s.id=a.shift_id WHERE ($1 = '' OR a.staff_id = $1) ORDER BY a.effective_from DESC, a.created_at DESC`, staffID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []domain.ShiftAssignment
	for rows.Next() {
		var sa domain.ShiftAssignment
		var effectiveTo *string
		if err = rows.Scan(&sa.ID, &sa.StaffID, &sa.StaffName, &sa.ShiftID, &sa.ShiftName, &sa.EffectiveFrom, &effectiveTo, &sa.CreatedAt); err != nil {
			return nil, err
		}
		sa.EffectiveTo = effectiveTo
		assignments = append(assignments, sa)
	}
	return assignments, rows.Err()
}

func (s Store) AssignShift(ctx context.Context, actor domain.Actor, input domain.ShiftAssignmentInput) (domain.ShiftAssignment, error) {
	var sa domain.ShiftAssignment
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return sa, err
	}
	defer tx.Rollback(ctx)

	var staffActive bool
	if err = tx.QueryRow(ctx, `SELECT active FROM staff_access WHERE user_id=$1`, input.StaffID).Scan(&staffActive); err != nil || !staffActive {
		return sa, domain.ErrValidation
	}

	var shiftActive bool
	if err = tx.QueryRow(ctx, `SELECT active FROM attendance_shift WHERE id=$1`, input.ShiftID).Scan(&shiftActive); err != nil || !shiftActive {
		return sa, domain.ErrValidation
	}

	err = tx.QueryRow(ctx, `INSERT INTO attendance_shift_assignment (staff_id, shift_id, effective_from, effective_to) VALUES ($1, $2, $3::date, NULLIF($4,'')::date) RETURNING id, staff_id, shift_id, effective_from::text, effective_to::text, created_at`,
		input.StaffID, input.ShiftID, input.EffectiveFrom, pointerString(input.EffectiveTo)).
		Scan(&sa.ID, &sa.StaffID, &sa.ShiftID, &sa.EffectiveFrom, &sa.EffectiveTo, &sa.CreatedAt)
	if err != nil {
		return sa, err
	}

	_ = tx.QueryRow(ctx, `SELECT name FROM "user" WHERE id=$1`, sa.StaffID).Scan(&sa.StaffName)
	_ = tx.QueryRow(ctx, `SELECT name FROM attendance_shift WHERE id=$1`, sa.ShiftID).Scan(&sa.ShiftName)

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'attendance.shift_assigned', $2)`, actor.ID, sa.ID); err != nil {
		return sa, err
	}
	return sa, tx.Commit(ctx)
}

func pointerString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (s Store) resolveShift(ctx context.Context, q interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}, staffID string, workDate string, requestedShiftID string) (domain.Shift, error) {
	if q == nil {
		q = s.DB
	}
	var sh domain.Shift
	var row pgx.Row
	if requestedShiftID != "" {
		row = q.QueryRow(ctx, `SELECT id, name, to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'), grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active, version, created_at, updated_at FROM attendance_shift WHERE id=$1 AND active=true`, requestedShiftID)
	} else {
		// Check explicit assignment
		row = q.QueryRow(ctx, `SELECT s.id, s.name, to_char(s.start_time, 'HH24:MI:SS'), to_char(s.end_time, 'HH24:MI:SS'), s.grace_period_minutes, s.break_duration_minutes, s.half_day_minutes, s.full_day_minutes, s.is_overnight, s.active, s.version, s.created_at, s.updated_at FROM attendance_shift_assignment a JOIN attendance_shift s ON s.id=a.shift_id WHERE a.staff_id=$1 AND a.effective_from <= $2::date AND (a.effective_to IS NULL OR a.effective_to >= $2::date) AND s.active=true ORDER BY a.effective_from DESC, a.created_at DESC LIMIT 1`, staffID, workDate)
	}

	err := row.Scan(&sh.ID, &sh.Name, &sh.StartTime, &sh.EndTime, &sh.GracePeriodMinutes, &sh.BreakDurationMinutes, &sh.HalfDayMinutes, &sh.FullDayMinutes, &sh.IsOvernight, &sh.Active, &sh.Version, &sh.CreatedAt, &sh.UpdatedAt)
	if err == nil {
		return sh, nil
	}

	// Fallback to default Day Shift
	err = q.QueryRow(ctx, `SELECT id, name, to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'), grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active, version, created_at, updated_at FROM attendance_shift WHERE name='Day Shift' AND active=true`).
		Scan(&sh.ID, &sh.Name, &sh.StartTime, &sh.EndTime, &sh.GracePeriodMinutes, &sh.BreakDurationMinutes, &sh.HalfDayMinutes, &sh.FullDayMinutes, &sh.IsOvernight, &sh.Active, &sh.Version, &sh.CreatedAt, &sh.UpdatedAt)
	return sh, err
}

func (s Store) GetTodayRecord(ctx context.Context, actor domain.Actor, staffID string, workDate string) (*domain.AttendanceRecord, error) {
	row := s.DB.QueryRow(ctx, `SELECT r.id, r.staff_id, u.name, r.work_date::text, r.shift_id, s.name, r.check_in_at, r.check_out_at, r.status, r.approval_status, r.total_break_minutes, r.worked_minutes, r.late_minutes, r.early_out_minutes, r.overtime_minutes, r.source, r.admin_notes, r.version, r.created_at, r.updated_at FROM attendance_record r JOIN "user" u ON u.id=r.staff_id JOIN attendance_shift s ON s.id=r.shift_id WHERE r.staff_id=$1 AND (r.work_date=$2::date OR r.check_out_at IS NULL) ORDER BY (r.check_out_at IS NULL) DESC, r.check_in_at DESC LIMIT 1`, staffID, workDate)

	var rec domain.AttendanceRecord
	err := row.Scan(&rec.ID, &rec.StaffID, &rec.StaffName, &rec.WorkDate, &rec.ShiftID, &rec.ShiftName, &rec.CheckInAt, &rec.CheckOutAt, &rec.Status, &rec.ApprovalStatus, &rec.TotalBreakMinutes, &rec.WorkedMinutes, &rec.LateMinutes, &rec.EarlyOutMinutes, &rec.OvertimeMinutes, &rec.Source, &rec.AdminNotes, &rec.Version, &rec.CreatedAt, &rec.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	breaks, err := s.loadBreaks(ctx, s.DB, rec.ID)
	if err != nil {
		return nil, err
	}
	rec.Breaks = breaks
	return &rec, nil
}

func (s Store) loadBreaks(ctx context.Context, q interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}, recordID string) ([]domain.BreakRecord, error) {
	rows, err := q.Query(ctx, `SELECT id, attendance_record_id, start_at, end_at, duration_minutes, reason, created_at FROM attendance_break WHERE attendance_record_id=$1 ORDER BY start_at ASC`, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var breaks []domain.BreakRecord
	for rows.Next() {
		var b domain.BreakRecord
		if err = rows.Scan(&b.ID, &b.AttendanceRecordID, &b.StartAt, &b.EndAt, &b.DurationMinutes, &b.Reason, &b.CreatedAt); err != nil {
			return nil, err
		}
		breaks = append(breaks, b)
	}
	return breaks, rows.Err()
}

func (s Store) ClockIn(ctx context.Context, actor domain.Actor, staffID string, shiftID string, now time.Time, key string) (domain.AttendanceRecord, error) {
	var rec domain.AttendanceRecord
	workDate := now.In(domain.HospitalLocation).Format("2006-01-02")

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return rec, err
	}
	defer tx.Rollback(ctx)
	if replay, e := attendanceReplay(ctx, tx, actor.ID, "clock-in", key, []string{staffID, shiftID}, &rec); e != nil || replay {
		return rec, e
	}

	// Lock on staff to prevent concurrent duplicate check-ins
	var lockedID string
	if err = tx.QueryRow(ctx, `SELECT user_id FROM staff_access WHERE user_id=$1 AND active=true FOR UPDATE`, staffID).Scan(&lockedID); err != nil {
		return rec, domain.ErrForbidden
	}

	var existingID string
	if err = tx.QueryRow(ctx, `SELECT id FROM attendance_record WHERE staff_id=$1 AND (work_date=$2::date OR check_out_at IS NULL)`, staffID, workDate).Scan(&existingID); err == nil {
		return rec, domain.ErrConflict
	}

	shift, err := s.resolveShift(ctx, tx, staffID, workDate, shiftID)
	if err != nil {
		return rec, err
	}

	status, late, early, worked, ot, err := domain.CalculateAttendance(shift, workDate, now, nil, 0)
	if err != nil {
		return rec, err
	}

	err = tx.QueryRow(ctx, `INSERT INTO attendance_record (staff_id, work_date, shift_id, check_in_at, status, approval_status, late_minutes, early_out_minutes, worked_minutes, overtime_minutes, source) VALUES ($1, $2::date, $3, $4, $5, 'draft', $6, $7, $8, $9, 'self_service') RETURNING id, staff_id, work_date::text, shift_id, check_in_at, check_out_at, status, approval_status, total_break_minutes, worked_minutes, late_minutes, early_out_minutes, overtime_minutes, source, admin_notes, version, created_at, updated_at`,
		staffID, workDate, shift.ID, now, status, late, early, worked, ot).
		Scan(&rec.ID, &rec.StaffID, &rec.WorkDate, &rec.ShiftID, &rec.CheckInAt, &rec.CheckOutAt, &rec.Status, &rec.ApprovalStatus, &rec.TotalBreakMinutes, &rec.WorkedMinutes, &rec.LateMinutes, &rec.EarlyOutMinutes, &rec.OvertimeMinutes, &rec.Source, &rec.AdminNotes, &rec.Version, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		return rec, err
	}

	rec.ShiftName = shift.Name
	_ = tx.QueryRow(ctx, `SELECT name FROM "user" WHERE id=$1`, staffID).Scan(&rec.StaffName)

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'attendance.clocked_in', $2)`, actor.ID, rec.ID); err != nil {
		return rec, err
	}

	if err = attendanceRemember(ctx, tx, actor.ID, "clock-in", key, []string{staffID, shiftID}, rec); err != nil {
		return rec, err
	}
	return rec, tx.Commit(ctx)
}

func (s Store) ClockOut(ctx context.Context, actor domain.Actor, staffID string, now time.Time, key string) (domain.AttendanceRecord, error) {
	var rec domain.AttendanceRecord
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return rec, err
	}
	defer tx.Rollback(ctx)
	if replay, e := attendanceReplay(ctx, tx, actor.ID, "clock-out", key, staffID, &rec); e != nil || replay {
		return rec, e
	}

	// Find the latest open record for this staff member
	var recordID string
	var workDate string
	var shiftID string
	var checkInAt time.Time
	var version int
	err = tx.QueryRow(ctx, `SELECT id, work_date::text, shift_id, check_in_at, version FROM attendance_record WHERE staff_id=$1 AND check_out_at IS NULL ORDER BY check_in_at DESC LIMIT 1 FOR UPDATE`, staffID).
		Scan(&recordID, &workDate, &shiftID, &checkInAt, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return rec, domain.ErrNotFound
	}
	if err != nil {
		return rec, err
	}

	// Auto-close any active break
	var openBreakID string
	var breakStartAt time.Time
	if err = tx.QueryRow(ctx, `SELECT id, start_at FROM attendance_break WHERE attendance_record_id=$1 AND end_at IS NULL ORDER BY start_at DESC LIMIT 1 FOR UPDATE`, recordID).Scan(&openBreakID, &breakStartAt); err == nil {
		dur := int(now.Sub(breakStartAt).Minutes())
		if dur < 0 {
			dur = 0
		}
		if _, err = tx.Exec(ctx, `UPDATE attendance_break SET end_at=$1, duration_minutes=$2 WHERE id=$3`, now, dur, openBreakID); err != nil {
			return rec, err
		}
	}

	// Calculate total break minutes
	var totalBreaks int
	_ = tx.QueryRow(ctx, `SELECT COALESCE(SUM(duration_minutes), 0) FROM attendance_break WHERE attendance_record_id=$1 AND end_at IS NOT NULL`, recordID).Scan(&totalBreaks)

	// Load shift
	var shift domain.Shift
	err = tx.QueryRow(ctx, `SELECT id, name, to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'), grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active, version, created_at, updated_at FROM attendance_shift WHERE id=$1`, shiftID).
		Scan(&shift.ID, &shift.Name, &shift.StartTime, &shift.EndTime, &shift.GracePeriodMinutes, &shift.BreakDurationMinutes, &shift.HalfDayMinutes, &shift.FullDayMinutes, &shift.IsOvernight, &shift.Active, &shift.Version, &shift.CreatedAt, &shift.UpdatedAt)
	if err != nil {
		return rec, err
	}

	status, late, early, worked, ot, err := domain.CalculateAttendance(shift, workDate, checkInAt, &now, totalBreaks)
	if err != nil {
		return rec, err
	}

	err = tx.QueryRow(ctx, `UPDATE attendance_record SET check_out_at=$1, status=$2, total_break_minutes=$3, worked_minutes=$4, late_minutes=$5, early_out_minutes=$6, overtime_minutes=$7, version=version+1, updated_at=now() WHERE id=$8 AND version=$9 RETURNING id, staff_id, work_date::text, shift_id, check_in_at, check_out_at, status, approval_status, total_break_minutes, worked_minutes, late_minutes, early_out_minutes, overtime_minutes, source, admin_notes, version, created_at, updated_at`,
		now, status, totalBreaks, worked, late, early, ot, recordID, version).
		Scan(&rec.ID, &rec.StaffID, &rec.WorkDate, &rec.ShiftID, &rec.CheckInAt, &rec.CheckOutAt, &rec.Status, &rec.ApprovalStatus, &rec.TotalBreakMinutes, &rec.WorkedMinutes, &rec.LateMinutes, &rec.EarlyOutMinutes, &rec.OvertimeMinutes, &rec.Source, &rec.AdminNotes, &rec.Version, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		return rec, err
	}

	rec.ShiftName = shift.Name
	_ = tx.QueryRow(ctx, `SELECT name FROM "user" WHERE id=$1`, staffID).Scan(&rec.StaffName)

	breaks, err := s.loadBreaks(ctx, tx, rec.ID)
	if err != nil {
		return rec, err
	}
	rec.Breaks = breaks

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'attendance.clocked_out', $2)`, actor.ID, rec.ID); err != nil {
		return rec, err
	}

	if err = attendanceRemember(ctx, tx, actor.ID, "clock-out", key, staffID, rec); err != nil {
		return rec, err
	}
	return rec, tx.Commit(ctx)
}

func (s Store) StartBreak(ctx context.Context, actor domain.Actor, staffID string, now time.Time, reason string, key string) (domain.BreakRecord, error) {
	var b domain.BreakRecord
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return b, err
	}
	defer tx.Rollback(ctx)
	if replay, e := attendanceReplay(ctx, tx, actor.ID, "break-start", key, []string{staffID, reason}, &b); e != nil || replay {
		return b, e
	}

	var recordID string
	err = tx.QueryRow(ctx, `SELECT id FROM attendance_record WHERE staff_id=$1 AND check_out_at IS NULL ORDER BY check_in_at DESC LIMIT 1 FOR UPDATE`, staffID).Scan(&recordID)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, domain.ErrNotFound
	}
	if err != nil {
		return b, err
	}

	// Check if already on break
	var activeBreakID string
	if err = tx.QueryRow(ctx, `SELECT id FROM attendance_break WHERE attendance_record_id=$1 AND end_at IS NULL LIMIT 1`, recordID).Scan(&activeBreakID); err == nil {
		return b, domain.ErrConflict
	}

	err = tx.QueryRow(ctx, `INSERT INTO attendance_break (attendance_record_id, start_at, reason) VALUES ($1, $2, $3) RETURNING id, attendance_record_id, start_at, end_at, duration_minutes, reason, created_at`,
		recordID, now, reason).
		Scan(&b.ID, &b.AttendanceRecordID, &b.StartAt, &b.EndAt, &b.DurationMinutes, &b.Reason, &b.CreatedAt)
	if err != nil {
		return b, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'attendance.break_started', $2)`, actor.ID, b.ID); err != nil {
		return b, err
	}
	if err = attendanceRemember(ctx, tx, actor.ID, "break-start", key, []string{staffID, reason}, b); err != nil {
		return b, err
	}
	return b, tx.Commit(ctx)
}

func (s Store) EndBreak(ctx context.Context, actor domain.Actor, staffID string, now time.Time, key string) (domain.BreakRecord, error) {
	var b domain.BreakRecord
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return b, err
	}
	defer tx.Rollback(ctx)
	if replay, e := attendanceReplay(ctx, tx, actor.ID, "break-end", key, staffID, &b); e != nil || replay {
		return b, e
	}

	var recordID string
	err = tx.QueryRow(ctx, `SELECT id FROM attendance_record WHERE staff_id=$1 AND check_out_at IS NULL ORDER BY check_in_at DESC LIMIT 1 FOR UPDATE`, staffID).Scan(&recordID)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, domain.ErrNotFound
	}
	if err != nil {
		return b, err
	}

	var breakID string
	var startAt time.Time
	err = tx.QueryRow(ctx, `SELECT id, start_at FROM attendance_break WHERE attendance_record_id=$1 AND end_at IS NULL ORDER BY start_at DESC LIMIT 1 FOR UPDATE`, recordID).Scan(&breakID, &startAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, domain.ErrNotFound
	}
	if err != nil {
		return b, err
	}

	dur := int(now.Sub(startAt).Minutes())
	if dur < 0 {
		return b, domain.ErrValidation
	}

	err = tx.QueryRow(ctx, `UPDATE attendance_break SET end_at=$1, duration_minutes=$2 WHERE id=$3 RETURNING id, attendance_record_id, start_at, end_at, duration_minutes, reason, created_at`,
		now, dur, breakID).
		Scan(&b.ID, &b.AttendanceRecordID, &b.StartAt, &b.EndAt, &b.DurationMinutes, &b.Reason, &b.CreatedAt)
	if err != nil {
		return b, err
	}

	// Update record's total_break_minutes
	if _, err = tx.Exec(ctx, `UPDATE attendance_record SET total_break_minutes=(SELECT COALESCE(SUM(duration_minutes),0) FROM attendance_break WHERE attendance_record_id=$1 AND end_at IS NOT NULL), updated_at=now() WHERE id=$1`, recordID); err != nil {
		return b, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'attendance.break_ended', $2)`, actor.ID, b.ID); err != nil {
		return b, err
	}
	if err = attendanceRemember(ctx, tx, actor.ID, "break-end", key, staffID, b); err != nil {
		return b, err
	}
	return b, tx.Commit(ctx)
}

func (s Store) ListAttendanceRecords(ctx context.Context, actor domain.Actor, filter domain.AttendanceFilter) ([]domain.AttendanceRecord, int, error) {
	where := `WHERE 1=1`
	args := []any{}
	idx := 1

	if filter.StaffID != "" {
		where += fmt.Sprintf(` AND r.staff_id = $%d`, idx)
		args = append(args, filter.StaffID)
		idx++
	}
	if filter.WorkDate != "" {
		where += fmt.Sprintf(` AND r.work_date = $%d::date`, idx)
		args = append(args, filter.WorkDate)
		idx++
	}
	if filter.FromDate != "" {
		where += fmt.Sprintf(` AND r.work_date >= $%d::date`, idx)
		args = append(args, filter.FromDate)
		idx++
	}
	if filter.ToDate != "" {
		where += fmt.Sprintf(` AND r.work_date <= $%d::date`, idx)
		args = append(args, filter.ToDate)
		idx++
	}
	if filter.Status != "" {
		where += fmt.Sprintf(` AND r.status = $%d`, idx)
		args = append(args, filter.Status)
		idx++
	}
	if filter.ApprovalStatus != "" {
		where += fmt.Sprintf(` AND r.approval_status = $%d`, idx)
		args = append(args, filter.ApprovalStatus)
		idx++
	}

	var total int
	countQuery := fmt.Sprintf(`SELECT count(*) FROM attendance_record r %s`, where)
	if err := s.DB.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	query := fmt.Sprintf(`SELECT r.id, r.staff_id, u.name, r.work_date::text, r.shift_id, s.name, r.check_in_at, r.check_out_at, r.status, r.approval_status, r.total_break_minutes, r.worked_minutes, r.late_minutes, r.early_out_minutes, r.overtime_minutes, r.source, r.admin_notes, r.version, r.created_at, r.updated_at FROM attendance_record r JOIN "user" u ON u.id=r.staff_id JOIN attendance_shift s ON s.id=r.shift_id %s ORDER BY r.work_date DESC, r.check_in_at DESC LIMIT $%d OFFSET $%d`, where, idx, idx+1)
	args = append(args, filter.PageSize, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []domain.AttendanceRecord
	for rows.Next() {
		var rec domain.AttendanceRecord
		if err = rows.Scan(&rec.ID, &rec.StaffID, &rec.StaffName, &rec.WorkDate, &rec.ShiftID, &rec.ShiftName, &rec.CheckInAt, &rec.CheckOutAt, &rec.Status, &rec.ApprovalStatus, &rec.TotalBreakMinutes, &rec.WorkedMinutes, &rec.LateMinutes, &rec.EarlyOutMinutes, &rec.OvertimeMinutes, &rec.Source, &rec.AdminNotes, &rec.Version, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, 0, err
		}
		records = append(records, rec)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	for i := range records {
		breaks, err := s.loadBreaks(ctx, s.DB, records[i].ID)
		if err != nil {
			return nil, 0, err
		}
		records[i].Breaks = breaks
	}

	return records, total, nil
}

func (s Store) GetAttendanceRecord(ctx context.Context, actor domain.Actor, id string) (domain.AttendanceRecord, error) {
	var rec domain.AttendanceRecord
	row := s.DB.QueryRow(ctx, `SELECT r.id, r.staff_id, u.name, r.work_date::text, r.shift_id, s.name, r.check_in_at, r.check_out_at, r.status, r.approval_status, r.total_break_minutes, r.worked_minutes, r.late_minutes, r.early_out_minutes, r.overtime_minutes, r.source, r.admin_notes, r.version, r.created_at, r.updated_at FROM attendance_record r JOIN "user" u ON u.id=r.staff_id JOIN attendance_shift s ON s.id=r.shift_id WHERE r.id=$1`, id)

	err := row.Scan(&rec.ID, &rec.StaffID, &rec.StaffName, &rec.WorkDate, &rec.ShiftID, &rec.ShiftName, &rec.CheckInAt, &rec.CheckOutAt, &rec.Status, &rec.ApprovalStatus, &rec.TotalBreakMinutes, &rec.WorkedMinutes, &rec.LateMinutes, &rec.EarlyOutMinutes, &rec.OvertimeMinutes, &rec.Source, &rec.AdminNotes, &rec.Version, &rec.CreatedAt, &rec.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return rec, domain.ErrNotFound
	}
	if err != nil {
		return rec, err
	}

	breaks, err := s.loadBreaks(ctx, s.DB, rec.ID)
	if err != nil {
		return rec, err
	}
	rec.Breaks = breaks
	return rec, nil
}

func (s Store) GetAttendanceHistory(ctx context.Context, actor domain.Actor, id string) (domain.AttendanceHistory, error) {
	var hist domain.AttendanceHistory
	rec, err := s.GetAttendanceRecord(ctx, actor, id)
	if err != nil {
		return hist, err
	}
	hist.Record = rec

	// Corrections
	corrRows, err := s.DB.Query(ctx, `SELECT c.id, c.attendance_record_id, c.actor_id, COALESCE(u.name, c.actor_id), c.reason, c.before_snapshot, c.after_snapshot, c.created_at FROM attendance_correction c LEFT JOIN "user" u ON u.id=c.actor_id WHERE c.attendance_record_id=$1 ORDER BY c.created_at DESC`, id)
	if err != nil {
		return hist, err
	}
	defer corrRows.Close()

	for corrRows.Next() {
		var c domain.AttendanceCorrection
		var beforeJSON, afterJSON []byte
		if err = corrRows.Scan(&c.ID, &c.AttendanceRecordID, &c.ActorID, &c.ActorName, &c.Reason, &beforeJSON, &afterJSON, &c.CreatedAt); err != nil {
			return hist, err
		}
		_ = json.Unmarshal(beforeJSON, &c.BeforeSnapshot)
		_ = json.Unmarshal(afterJSON, &c.AfterSnapshot)
		hist.Corrections = append(hist.Corrections, c)
	}

	// Approval events
	appRows, err := s.DB.Query(ctx, `SELECT a.id, a.attendance_record_id, a.actor_id, COALESCE(u.name, a.actor_id), a.from_status, a.to_status, a.reason, a.created_at FROM attendance_approval_history a LEFT JOIN "user" u ON u.id=a.actor_id WHERE a.attendance_record_id=$1 ORDER BY a.created_at DESC`, id)
	if err != nil {
		return hist, err
	}
	defer appRows.Close()

	for appRows.Next() {
		var ev domain.AttendanceApprovalEvent
		if err = appRows.Scan(&ev.ID, &ev.AttendanceRecordID, &ev.ActorID, &ev.ActorName, &ev.FromStatus, &ev.ToStatus, &ev.Reason, &ev.CreatedAt); err != nil {
			return hist, err
		}
		hist.ApprovalEvents = append(hist.ApprovalEvents, ev)
	}

	return hist, nil
}

func (s Store) AdminCreateRecord(ctx context.Context, actor domain.Actor, input domain.AdminAttendanceInput, key string) (domain.AttendanceRecord, error) {
	var rec domain.AttendanceRecord
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return rec, err
	}
	defer tx.Rollback(ctx)
	if replay, e := attendanceReplay(ctx, tx, actor.ID, "admin-create", key, input, &rec); e != nil || replay {
		return rec, e
	}

	var staffActive bool
	if err = tx.QueryRow(ctx, `SELECT active FROM staff_access WHERE user_id=$1 FOR UPDATE`, input.StaffID).Scan(&staffActive); err != nil || !staffActive {
		return rec, domain.ErrValidation
	}

	var existingID string
	if err = tx.QueryRow(ctx, `SELECT id FROM attendance_record WHERE staff_id=$1 AND (work_date=$2::date OR ($3 AND check_out_at IS NULL))`, input.StaffID, input.WorkDate, input.CheckOutAt == nil).Scan(&existingID); err == nil {
		return rec, domain.ErrConflict
	}

	var shift domain.Shift
	err = tx.QueryRow(ctx, `SELECT id, name, to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'), grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active, version, created_at, updated_at FROM attendance_shift WHERE id=$1`, input.ShiftID).
		Scan(&shift.ID, &shift.Name, &shift.StartTime, &shift.EndTime, &shift.GracePeriodMinutes, &shift.BreakDurationMinutes, &shift.HalfDayMinutes, &shift.FullDayMinutes, &shift.IsOvernight, &shift.Active, &shift.Version, &shift.CreatedAt, &shift.UpdatedAt)
	if err != nil {
		return rec, err
	}

	status, late, early, worked, ot, err := domain.CalculateAttendance(shift, input.WorkDate, input.CheckInAt, input.CheckOutAt, input.TotalBreakMinutes)
	if err != nil {
		return rec, err
	}

	err = tx.QueryRow(ctx, `INSERT INTO attendance_record (staff_id, work_date, shift_id, check_in_at, check_out_at, status, approval_status, total_break_minutes, worked_minutes, late_minutes, early_out_minutes, overtime_minutes, source, admin_notes) VALUES ($1, $2::date, $3, $4, $5, $6, 'draft', $7, $8, $9, $10, $11, 'administrator', $12) RETURNING id, staff_id, work_date::text, shift_id, check_in_at, check_out_at, status, approval_status, total_break_minutes, worked_minutes, late_minutes, early_out_minutes, overtime_minutes, source, admin_notes, version, created_at, updated_at`,
		input.StaffID, input.WorkDate, shift.ID, input.CheckInAt, input.CheckOutAt, status, input.TotalBreakMinutes, worked, late, early, ot, input.AdminNotes).
		Scan(&rec.ID, &rec.StaffID, &rec.WorkDate, &rec.ShiftID, &rec.CheckInAt, &rec.CheckOutAt, &rec.Status, &rec.ApprovalStatus, &rec.TotalBreakMinutes, &rec.WorkedMinutes, &rec.LateMinutes, &rec.EarlyOutMinutes, &rec.OvertimeMinutes, &rec.Source, &rec.AdminNotes, &rec.Version, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		return rec, err
	}

	rec.ShiftName = shift.Name
	_ = tx.QueryRow(ctx, `SELECT name FROM "user" WHERE id=$1`, input.StaffID).Scan(&rec.StaffName)

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'attendance.admin_created', $2)`, actor.ID, rec.ID); err != nil {
		return rec, err
	}

	if err = attendanceRemember(ctx, tx, actor.ID, "admin-create", key, input, rec); err != nil {
		return rec, err
	}
	return rec, tx.Commit(ctx)
}

func (s Store) AdminCorrectRecord(ctx context.Context, actor domain.Actor, id string, input domain.AdminCorrectionInput) (domain.AttendanceRecord, error) {
	var rec domain.AttendanceRecord
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return rec, err
	}
	defer tx.Rollback(ctx)

	var before domain.AttendanceRecord
	row := tx.QueryRow(ctx, `SELECT r.id, r.staff_id, u.name, r.work_date::text, r.shift_id, s.name, r.check_in_at, r.check_out_at, r.status, r.approval_status, r.total_break_minutes, r.worked_minutes, r.late_minutes, r.early_out_minutes, r.overtime_minutes, r.source, r.admin_notes, r.version, r.created_at, r.updated_at FROM attendance_record r JOIN "user" u ON u.id=r.staff_id JOIN attendance_shift s ON s.id=r.shift_id WHERE r.id=$1 FOR UPDATE`, id)
	err = row.Scan(&before.ID, &before.StaffID, &before.StaffName, &before.WorkDate, &before.ShiftID, &before.ShiftName, &before.CheckInAt, &before.CheckOutAt, &before.Status, &before.ApprovalStatus, &before.TotalBreakMinutes, &before.WorkedMinutes, &before.LateMinutes, &before.EarlyOutMinutes, &before.OvertimeMinutes, &before.Source, &before.AdminNotes, &before.Version, &before.CreatedAt, &before.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return rec, domain.ErrNotFound
	}
	if err != nil {
		return rec, err
	}

	if before.Version != input.Version {
		return rec, domain.ErrStale
	}

	if (before.ApprovalStatus == "approved" || before.ApprovalStatus == "submitted") && input.CheckOutAt == nil {
		return rec, domain.ErrConflict
	}

	// Load shift
	var shift domain.Shift
	err = tx.QueryRow(ctx, `SELECT id, name, to_char(start_time, 'HH24:MI:SS'), to_char(end_time, 'HH24:MI:SS'), grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active, version, created_at, updated_at FROM attendance_shift WHERE id=$1`, input.ShiftID).
		Scan(&shift.ID, &shift.Name, &shift.StartTime, &shift.EndTime, &shift.GracePeriodMinutes, &shift.BreakDurationMinutes, &shift.HalfDayMinutes, &shift.FullDayMinutes, &shift.IsOvernight, &shift.Active, &shift.Version, &shift.CreatedAt, &shift.UpdatedAt)
	if err != nil {
		return rec, err
	}

	status, late, early, worked, ot, err := domain.CalculateAttendance(shift, before.WorkDate, input.CheckInAt, input.CheckOutAt, input.TotalBreakMinutes)
	if err != nil {
		return rec, err
	}

	// Revert approval to submitted if approved record was modified
	newApproval := before.ApprovalStatus
	if before.ApprovalStatus == "approved" {
		newApproval = "submitted"
	}

	err = tx.QueryRow(ctx, `UPDATE attendance_record SET shift_id=$1, check_in_at=$2, check_out_at=$3, total_break_minutes=$4, worked_minutes=$5, late_minutes=$6, early_out_minutes=$7, overtime_minutes=$8, status=$9, approval_status=$10, admin_notes=$11, version=version+1, updated_at=now() WHERE id=$12 AND version=$13 RETURNING id, staff_id, work_date::text, shift_id, check_in_at, check_out_at, status, approval_status, total_break_minutes, worked_minutes, late_minutes, early_out_minutes, overtime_minutes, source, admin_notes, version, created_at, updated_at`,
		shift.ID, input.CheckInAt, input.CheckOutAt, input.TotalBreakMinutes, worked, late, early, ot, status, newApproval, input.AdminNotes, id, before.Version).
		Scan(&rec.ID, &rec.StaffID, &rec.WorkDate, &rec.ShiftID, &rec.CheckInAt, &rec.CheckOutAt, &rec.Status, &rec.ApprovalStatus, &rec.TotalBreakMinutes, &rec.WorkedMinutes, &rec.LateMinutes, &rec.EarlyOutMinutes, &rec.OvertimeMinutes, &rec.Source, &rec.AdminNotes, &rec.Version, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		return rec, err
	}

	rec.ShiftName = shift.Name
	rec.StaffName = before.StaffName

	beforeBytes, _ := json.Marshal(before)
	afterBytes, _ := json.Marshal(rec)

	if _, err = tx.Exec(ctx, `INSERT INTO attendance_correction (attendance_record_id, actor_id, reason, before_snapshot, after_snapshot) VALUES ($1, $2, $3, $4::jsonb, $5::jsonb)`,
		id, actor.ID, input.Reason, beforeBytes, afterBytes); err != nil {
		return rec, err
	}

	if before.ApprovalStatus == "approved" && newApproval == "submitted" {
		if _, err = tx.Exec(ctx, `INSERT INTO attendance_approval_history (attendance_record_id, actor_id, from_status, to_status, reason) VALUES ($1, $2, 'approved', 'submitted', 'Reset to submitted due to administrative correction: ' || $3)`,
			id, actor.ID, input.Reason); err != nil {
			return rec, err
		}
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'attendance.corrected', $2)`, actor.ID, rec.ID); err != nil {
		return rec, err
	}

	return rec, tx.Commit(ctx)
}

func (s Store) AdminUpdateApproval(ctx context.Context, actor domain.Actor, id string, input domain.ApprovalInput) (domain.AttendanceRecord, error) {
	var rec domain.AttendanceRecord
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return rec, err
	}
	defer tx.Rollback(ctx)

	var before domain.AttendanceRecord
	row := tx.QueryRow(ctx, `SELECT r.id, r.staff_id, u.name, r.work_date::text, r.shift_id, s.name, r.check_in_at, r.check_out_at, r.status, r.approval_status, r.total_break_minutes, r.worked_minutes, r.late_minutes, r.early_out_minutes, r.overtime_minutes, r.source, r.admin_notes, r.version, r.created_at, r.updated_at FROM attendance_record r JOIN "user" u ON u.id=r.staff_id JOIN attendance_shift s ON s.id=r.shift_id WHERE r.id=$1 FOR UPDATE`, id)
	err = row.Scan(&before.ID, &before.StaffID, &before.StaffName, &before.WorkDate, &before.ShiftID, &before.ShiftName, &before.CheckInAt, &before.CheckOutAt, &before.Status, &before.ApprovalStatus, &before.TotalBreakMinutes, &before.WorkedMinutes, &before.LateMinutes, &before.EarlyOutMinutes, &before.OvertimeMinutes, &before.Source, &before.AdminNotes, &before.Version, &before.CreatedAt, &before.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return rec, domain.ErrNotFound
	}
	if err != nil {
		return rec, err
	}

	if (input.Status == "submitted" || input.Status == "approved") && before.CheckOutAt == nil {
		return rec, domain.ErrConflict
	}
	var openBreak bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM attendance_break WHERE attendance_record_id=$1 AND end_at IS NULL)`, id).Scan(&openBreak); err != nil {
		return rec, err
	}
	if openBreak && (input.Status == "submitted" || input.Status == "approved") {
		return rec, domain.ErrConflict
	}

	if before.ApprovalStatus == input.Status {
		return before, nil
	}

	err = tx.QueryRow(ctx, `UPDATE attendance_record SET approval_status=$1, version=version+1, updated_at=now() WHERE id=$2 RETURNING id, staff_id, work_date::text, shift_id, check_in_at, check_out_at, status, approval_status, total_break_minutes, worked_minutes, late_minutes, early_out_minutes, overtime_minutes, source, admin_notes, version, created_at, updated_at`,
		input.Status, id).
		Scan(&rec.ID, &rec.StaffID, &rec.WorkDate, &rec.ShiftID, &rec.CheckInAt, &rec.CheckOutAt, &rec.Status, &rec.ApprovalStatus, &rec.TotalBreakMinutes, &rec.WorkedMinutes, &rec.LateMinutes, &rec.EarlyOutMinutes, &rec.OvertimeMinutes, &rec.Source, &rec.AdminNotes, &rec.Version, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		return rec, err
	}

	rec.ShiftName = before.ShiftName
	rec.StaffName = before.StaffName

	if _, err = tx.Exec(ctx, `INSERT INTO attendance_approval_history (attendance_record_id, actor_id, from_status, to_status, reason) VALUES ($1, $2, $3, $4, $5)`,
		id, actor.ID, before.ApprovalStatus, input.Status, input.Reason); err != nil {
		return rec, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'attendance.approval_updated', $2)`, actor.ID, rec.ID); err != nil {
		return rec, err
	}

	return rec, tx.Commit(ctx)
}

func (s Store) GetAttendanceSummary(ctx context.Context, actor domain.Actor, date string) (domain.AttendanceSummary, error) {
	var sum domain.AttendanceSummary
	sum.WorkDate = date

	err := s.DB.QueryRow(ctx, `SELECT count(*), count(*) FILTER(WHERE status='present'), count(*) FILTER(WHERE status='late'), count(*) FILTER(WHERE status='half_day'), count(*) FILTER(WHERE status='absent'), COALESCE(SUM(worked_minutes), 0), COALESCE(SUM(overtime_minutes), 0), count(*) FILTER(WHERE approval_status='submitted') FROM attendance_record WHERE work_date=$1::date`, date).
		Scan(&sum.TotalCount, &sum.PresentCount, &sum.LateCount, &sum.HalfDayCount, &sum.AbsentCount, &sum.TotalWorkedMinutes, &sum.TotalOvertimeMinutes, &sum.PendingApprovalCount)
	return sum, err
}
