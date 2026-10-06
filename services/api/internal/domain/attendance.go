package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Shift struct {
	StaffCount           int       `json:"staffCount"`
	Code                 string    `json:"code"`
	IsDefault            bool      `json:"isDefault"`
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	StartTime            string    `json:"startTime"`
	EndTime              string    `json:"endTime"`
	GracePeriodMinutes   int       `json:"gracePeriodMinutes"`
	BreakDurationMinutes int       `json:"breakDurationMinutes"`
	HalfDayMinutes       int       `json:"halfDayMinutes"`
	FullDayMinutes       int       `json:"fullDayMinutes"`
	IsOvernight          bool      `json:"isOvernight"`
	Active               bool      `json:"active"`
	Version              int       `json:"version"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type ShiftInput struct {
	Code                 string `json:"code"`
	IsDefault            bool   `json:"isDefault"`
	Name                 string `json:"name"`
	StartTime            string `json:"startTime"`
	EndTime              string `json:"endTime"`
	GracePeriodMinutes   int    `json:"gracePeriodMinutes"`
	BreakDurationMinutes int    `json:"breakDurationMinutes"`
	HalfDayMinutes       int    `json:"halfDayMinutes"`
	FullDayMinutes       int    `json:"fullDayMinutes"`
	Active               bool   `json:"active"`
}

func parseTimeOfDay(s string) (int, int, error) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, 0, ErrValidation
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, 0, ErrValidation
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, 0, ErrValidation
	}
	return h, m, nil
}

func (s *ShiftInput) Validate() error {
	s.Code = strings.ToUpper(strings.TrimSpace(s.Code))
	if len(s.Code) > 32 || (s.IsDefault && !s.Active) {
		return ErrValidation
	}
	for i, c := range s.Code {
		if !(c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || i > 0 && (c == '_' || c == '-')) {
			return ErrValidation
		}
	}
	s.Name = strings.TrimSpace(s.Name)
	if len([]rune(s.Name)) < 1 || len([]rune(s.Name)) > 100 {
		return ErrValidation
	}
	startH, startM, err := parseTimeOfDay(s.StartTime)
	if err != nil {
		return ErrValidation
	}
	s.StartTime = fmt.Sprintf("%02d:%02d:00", startH, startM)
	endH, endM, err := parseTimeOfDay(s.EndTime)
	if err != nil {
		return ErrValidation
	}
	s.EndTime = fmt.Sprintf("%02d:%02d:00", endH, endM)

	if s.GracePeriodMinutes < 0 || s.GracePeriodMinutes > 120 {
		return ErrValidation
	}
	if s.BreakDurationMinutes < 0 || s.BreakDurationMinutes > 300 {
		return ErrValidation
	}
	if s.HalfDayMinutes <= 0 || s.HalfDayMinutes > 1440 {
		return ErrValidation
	}
	if s.FullDayMinutes < s.HalfDayMinutes || s.FullDayMinutes > 1440 {
		return ErrValidation
	}
	return nil
}

type ShiftAssignment struct {
	ID            string    `json:"id"`
	StaffID       string    `json:"staffId"`
	StaffName     string    `json:"staffName"`
	ShiftID       string    `json:"shiftId"`
	ShiftName     string    `json:"shiftName"`
	EffectiveFrom string    `json:"effectiveFrom"`
	EffectiveTo   *string   `json:"effectiveTo"`
	Note          string    `json:"note"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ShiftAssignmentInput struct {
	StaffID       string  `json:"staffId"`
	ShiftID       string  `json:"shiftId"`
	EffectiveFrom string  `json:"effectiveFrom"`
	EffectiveTo   *string `json:"effectiveTo"`
	Note          string  `json:"note"`
	Active        *bool   `json:"active,omitempty"`
}

func (a *ShiftAssignmentInput) Validate() error {
	a.StaffID = strings.TrimSpace(a.StaffID)
	if a.StaffID == "" || len(a.StaffID) > 128 {
		return ErrValidation
	}
	if !UUIDPattern.MatchString(a.ShiftID) {
		return ErrValidation
	}
	from, err := time.ParseInLocation("2006-01-02", a.EffectiveFrom, HospitalLocation)
	if err != nil {
		return ErrValidation
	}
	if a.EffectiveTo != nil && *a.EffectiveTo != "" {
		to, err := time.ParseInLocation("2006-01-02", *a.EffectiveTo, HospitalLocation)
		if err != nil || to.Before(from) {
			return ErrValidation
		}
	}
	if len([]rune(a.Note)) > 500 {
		return ErrValidation
	}
	return nil
}

type LeaveRequest struct {
	ID            string     `json:"id"`
	StaffID       string     `json:"staffId"`
	StaffName     string     `json:"staffName"`
	FromDate      string     `json:"fromDate"`
	ToDate        string     `json:"toDate"`
	Days          int        `json:"days"`
	LeaveType     string     `json:"leaveType"`
	Reason        string     `json:"reason"`
	Status        string     `json:"status"`
	ApproverID    *string    `json:"approverId"`
	ApproverName  *string    `json:"approverName"`
	ApproverNotes string     `json:"approverNotes"`
	ActionedAt    *time.Time `json:"actionedAt"`
	Version       int        `json:"version"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type LeaveRequestInput struct {
	StaffID   string `json:"staffId"`
	FromDate  string `json:"fromDate"`
	ToDate    string `json:"toDate"`
	LeaveType string `json:"leaveType"`
	Reason    string `json:"reason"`
}

func (l *LeaveRequestInput) Validate() error {
	l.StaffID = strings.TrimSpace(l.StaffID)
	if len(l.StaffID) > 128 {
		return ErrValidation
	}
	from, err := time.ParseInLocation("2006-01-02", l.FromDate, HospitalLocation)
	if err != nil {
		return ErrValidation
	}
	to, err := time.ParseInLocation("2006-01-02", l.ToDate, HospitalLocation)
	if err != nil || to.Before(from) {
		return ErrValidation
	}
	switch l.LeaveType {
	case "casual", "annual", "emergency", "sick", "other":
	default:
		return ErrValidation
	}
	l.Reason = strings.TrimSpace(l.Reason)
	if len([]rune(l.Reason)) < 1 || len([]rune(l.Reason)) > 1000 {
		return ErrValidation
	}
	return nil
}

type LeaveApprovalInput struct {
	Status        string `json:"status"`
	ApproverNotes string `json:"approverNotes"`
	Version       int    `json:"version"`
}

func (a *LeaveApprovalInput) Validate() error {
	switch a.Status {
	case "approved", "rejected", "cancelled":
	default:
		return ErrValidation
	}
	if a.Version < 1 {
		return ErrValidation
	}
	a.ApproverNotes = strings.TrimSpace(a.ApproverNotes)
	if len([]rune(a.ApproverNotes)) > 1000 {
		return ErrValidation
	}
	return nil
}

type LeaveFilter struct {
	StaffID  string
	Status   string
	Page     int
	PageSize int
}

type StaffMember struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type BreakRecord struct {
	ID                 string     `json:"id"`
	AttendanceRecordID string     `json:"attendanceRecordId"`
	StartAt            time.Time  `json:"startAt"`
	EndAt              *time.Time `json:"endAt"`
	DurationMinutes    *int       `json:"durationMinutes"`
	Reason             string     `json:"reason"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type AttendanceRecord struct {
	ID                string        `json:"id"`
	StaffID           string        `json:"staffId"`
	StaffName         string        `json:"staffName"`
	WorkDate          string        `json:"workDate"`
	ShiftID           string        `json:"shiftId"`
	ShiftName         string        `json:"shiftName"`
	CheckInAt         time.Time     `json:"checkInAt"`
	CheckOutAt        *time.Time    `json:"checkOutAt"`
	Status            string        `json:"status"`
	ApprovalStatus    string        `json:"approvalStatus"`
	TotalBreakMinutes int           `json:"totalBreakMinutes"`
	WorkedMinutes     int           `json:"workedMinutes"`
	LateMinutes       int           `json:"lateMinutes"`
	EarlyOutMinutes   int           `json:"earlyOutMinutes"`
	OvertimeMinutes   int           `json:"overtimeMinutes"`
	Source            string        `json:"source"`
	AdminNotes        string        `json:"adminNotes"`
	Breaks            []BreakRecord `json:"breaks"`
	Version           int           `json:"version"`
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

type AttendanceCorrection struct {
	ID                 string         `json:"id"`
	AttendanceRecordID string         `json:"attendanceRecordId"`
	ActorID            string         `json:"actorId"`
	ActorName          string         `json:"actorName"`
	Reason             string         `json:"reason"`
	BeforeSnapshot     map[string]any `json:"beforeSnapshot"`
	AfterSnapshot      map[string]any `json:"afterSnapshot"`
	CreatedAt          time.Time      `json:"createdAt"`
}

type AttendanceApprovalEvent struct {
	ID                 string    `json:"id"`
	AttendanceRecordID string    `json:"attendanceRecordId"`
	ActorID            string    `json:"actorId"`
	ActorName          string    `json:"actorName"`
	FromStatus         string    `json:"fromStatus"`
	ToStatus           string    `json:"toStatus"`
	Reason             string    `json:"reason"`
	CreatedAt          time.Time `json:"createdAt"`
}

type AttendanceHistory struct {
	Record         AttendanceRecord          `json:"record"`
	Corrections    []AttendanceCorrection    `json:"corrections"`
	ApprovalEvents []AttendanceApprovalEvent `json:"approvalEvents"`
}

type AttendanceSummary struct {
	WorkDate             string `json:"workDate"`
	TotalCount           int    `json:"totalCount"`
	PresentCount         int    `json:"presentCount"`
	LateCount            int    `json:"lateCount"`
	HalfDayCount         int    `json:"halfDayCount"`
	AbsentCount          int    `json:"absentCount"`
	TotalWorkedMinutes   int    `json:"totalWorkedMinutes"`
	TotalOvertimeMinutes int    `json:"totalOvertimeMinutes"`
	PendingApprovalCount int    `json:"pendingApprovalCount"`
}

type AttendanceFilter struct {
	StaffID        string
	WorkDate       string
	FromDate       string
	ToDate         string
	Status         string
	ApprovalStatus string
	Page           int
	PageSize       int
}

type AdminAttendanceInput struct {
	StaffID           string     `json:"staffId"`
	WorkDate          string     `json:"workDate"`
	ShiftID           string     `json:"shiftId"`
	CheckInAt         time.Time  `json:"checkInAt"`
	CheckOutAt        *time.Time `json:"checkOutAt"`
	TotalBreakMinutes int        `json:"totalBreakMinutes"`
	AdminNotes        string     `json:"adminNotes"`
	Reason            string     `json:"reason"`
}

func (i *AdminAttendanceInput) Validate() error {
	i.StaffID = strings.TrimSpace(i.StaffID)
	if i.StaffID == "" || len(i.StaffID) > 128 {
		return ErrValidation
	}
	if !UUIDPattern.MatchString(i.ShiftID) {
		return ErrValidation
	}
	if _, err := time.ParseInLocation("2006-01-02", i.WorkDate, HospitalLocation); err != nil {
		return ErrValidation
	}
	if i.CheckInAt.IsZero() {
		return ErrValidation
	}
	if i.CheckOutAt != nil && i.CheckOutAt.Before(i.CheckInAt) {
		return ErrValidation
	}
	if i.TotalBreakMinutes < 0 || i.TotalBreakMinutes > 1440 {
		return ErrValidation
	}
	if len([]rune(i.AdminNotes)) > 1000 || len([]rune(i.Reason)) > 1000 {
		return ErrValidation
	}
	return nil
}

type AdminCorrectionInput struct {
	Version           int        `json:"version"`
	ShiftID           string     `json:"shiftId"`
	CheckInAt         time.Time  `json:"checkInAt"`
	CheckOutAt        *time.Time `json:"checkOutAt"`
	TotalBreakMinutes int        `json:"totalBreakMinutes"`
	AdminNotes        string     `json:"adminNotes"`
	Reason            string     `json:"reason"`
}

func (i *AdminCorrectionInput) Validate() error {
	if i.Version < 1 {
		return ErrValidation
	}
	if !UUIDPattern.MatchString(i.ShiftID) {
		return ErrValidation
	}
	if i.CheckInAt.IsZero() {
		return ErrValidation
	}
	if i.CheckOutAt != nil && i.CheckOutAt.Before(i.CheckInAt) {
		return ErrValidation
	}
	if i.TotalBreakMinutes < 0 || i.TotalBreakMinutes > 1440 {
		return ErrValidation
	}
	i.Reason = strings.TrimSpace(i.Reason)
	if len([]rune(i.Reason)) < 1 || len([]rune(i.Reason)) > 1000 {
		return ErrValidation
	}
	if len([]rune(i.AdminNotes)) > 1000 {
		return ErrValidation
	}
	return nil
}

type ApprovalInput struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func (i *ApprovalInput) Validate() error {
	switch i.Status {
	case "submitted", "approved", "rejected":
	default:
		return ErrValidation
	}
	i.Reason = strings.TrimSpace(i.Reason)
	if len([]rune(i.Reason)) > 1000 {
		return ErrValidation
	}
	return nil
}

func CalculateAttendance(shift Shift, workDate string, checkInAt time.Time, checkOutAt *time.Time, totalBreakMinutes int) (status string, lateMinutes, earlyOutMinutes, workedMinutes, overtimeMinutes int, err error) {
	workDay, err := time.ParseInLocation("2006-01-02", workDate, HospitalLocation)
	if err != nil {
		return "", 0, 0, 0, 0, ErrValidation
	}

	startH, startM, err := parseTimeOfDay(shift.StartTime)
	if err != nil {
		return "", 0, 0, 0, 0, ErrValidation
	}
	scheduledStart := time.Date(workDay.Year(), workDay.Month(), workDay.Day(), startH, startM, 0, 0, HospitalLocation)

	endH, endM, err := parseTimeOfDay(shift.EndTime)
	if err != nil {
		return "", 0, 0, 0, 0, ErrValidation
	}
	endDay := workDay
	if shift.IsOvernight || (endH < startH) || (endH == startH && endM <= startM) {
		endDay = endDay.AddDate(0, 0, 1)
	}
	scheduledEnd := time.Date(endDay.Year(), endDay.Month(), endDay.Day(), endH, endM, 0, 0, HospitalLocation)

	graceThreshold := scheduledStart.Add(time.Duration(shift.GracePeriodMinutes) * time.Minute)
	if checkInAt.In(HospitalLocation).After(graceThreshold) {
		lateMinutes = int(checkInAt.In(HospitalLocation).Sub(scheduledStart).Minutes())
	} else {
		lateMinutes = 0
	}

	if checkOutAt == nil {
		status = "checked_in"
		earlyOutMinutes = 0
		workedMinutes = 0
		overtimeMinutes = 0
		return status, lateMinutes, earlyOutMinutes, workedMinutes, overtimeMinutes, nil
	}

	checkOutLocal := checkOutAt.In(HospitalLocation)
	checkInLocal := checkInAt.In(HospitalLocation)
	if checkOutLocal.Before(checkInLocal) {
		return "", 0, 0, 0, 0, ErrValidation
	}

	if checkOutLocal.Before(scheduledEnd) {
		earlyOutMinutes = int(scheduledEnd.Sub(checkOutLocal).Minutes())
	} else {
		earlyOutMinutes = 0
	}

	elapsed := int(checkOutLocal.Sub(checkInLocal).Minutes())
	if totalBreakMinutes < 0 || totalBreakMinutes > elapsed {
		return "", 0, 0, 0, 0, ErrValidation
	}
	workedMinutes = elapsed - totalBreakMinutes

	if workedMinutes > shift.FullDayMinutes {
		overtimeMinutes = workedMinutes - shift.FullDayMinutes
	} else {
		overtimeMinutes = 0
	}

	if workedMinutes >= shift.FullDayMinutes {
		if lateMinutes > 0 {
			status = "late"
		} else {
			status = "present"
		}
	} else if workedMinutes >= shift.HalfDayMinutes {
		status = "half_day"
	} else {
		status = "absent"
	}

	return status, lateMinutes, earlyOutMinutes, workedMinutes, overtimeMinutes, nil
}
