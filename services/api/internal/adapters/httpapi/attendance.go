package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) attendance(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	if !strings.HasPrefix(r.URL.Path, "/v1/attendance") {
		return false
	}

	switch {
	case r.URL.Path == "/v1/attendance/shifts" && r.Method == "GET":
		shifts, err := s.Attendance.ListShifts(r.Context(), a)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"shifts": shifts})
		return true

	case r.URL.Path == "/v1/attendance/shifts" && r.Method == "POST":
		var input domain.ShiftInput
		if !decode(w, r, &input) {
			return true
		}
		sh, err := s.Attendance.CreateShift(r.Context(), a, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, sh)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/attendance/shifts/") && r.Method == "PATCH":
		id := strings.TrimPrefix(r.URL.Path, "/v1/attendance/shifts/")
		var req struct {
			domain.ShiftInput
			Version int `json:"version"`
		}
		if !decode(w, r, &req) {
			return true
		}
		sh, err := s.Attendance.UpdateShift(r.Context(), a, id, req.ShiftInput, req.Version)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, sh)
		return true

	case r.URL.Path == "/v1/attendance/assignments" && r.Method == "GET":
		staffID := r.URL.Query().Get("staffId")
		assignments, err := s.Attendance.ListShiftAssignments(r.Context(), a, staffID)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"assignments": assignments})
		return true

	case r.URL.Path == "/v1/attendance/assignments" && r.Method == "POST":
		var input domain.ShiftAssignmentInput
		if !decode(w, r, &input) {
			return true
		}
		as, err := s.Attendance.AssignShift(r.Context(), a, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, as)
		return true

	case r.URL.Path == "/v1/attendance/today" && r.Method == "GET":
		rec, err := s.Attendance.Today(r.Context(), a)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"record": rec})
		return true

	case r.URL.Path == "/v1/attendance/clock-in" && r.Method == "POST":
		var req struct {
			ShiftID string `json:"shiftId"`
		}
		if r.Body != nil && r.ContentLength > 0 {
			if !decode(w, r, &req) {
				return true
			}
		}
		rec, err := s.Attendance.ClockIn(r.Context(), a, req.ShiftID, r.Header.Get("Idempotency-Key"))
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, rec)
		return true

	case r.URL.Path == "/v1/attendance/clock-out" && r.Method == "POST":
		rec, err := s.Attendance.ClockOut(r.Context(), a, r.Header.Get("Idempotency-Key"))
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, rec)
		return true

	case r.URL.Path == "/v1/attendance/breaks/start" && r.Method == "POST":
		var req struct {
			Reason string `json:"reason"`
		}
		if r.Body != nil && r.ContentLength > 0 {
			if !decode(w, r, &req) {
				return true
			}
		}
		brk, err := s.Attendance.StartBreak(r.Context(), a, req.Reason, r.Header.Get("Idempotency-Key"))
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, brk)
		return true

	case r.URL.Path == "/v1/attendance/breaks/end" && r.Method == "POST":
		brk, err := s.Attendance.EndBreak(r.Context(), a, r.Header.Get("Idempotency-Key"))
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, brk)
		return true

	case r.URL.Path == "/v1/attendance/records" && r.Method == "GET":
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		filter := domain.AttendanceFilter{
			StaffID:        r.URL.Query().Get("staffId"),
			WorkDate:       r.URL.Query().Get("workDate"),
			FromDate:       r.URL.Query().Get("fromDate"),
			ToDate:         r.URL.Query().Get("toDate"),
			Status:         r.URL.Query().Get("status"),
			ApprovalStatus: r.URL.Query().Get("approvalStatus"),
			Page:           page,
			PageSize:       pageSize,
		}
		records, total, err := s.Attendance.ListRecords(r.Context(), a, filter)
		if err != nil {
			fail(w, err)
			return true
		}
		if records == nil {
			records = []domain.AttendanceRecord{}
		}
		write(w, 200, map[string]any{"records": records, "total": total, "page": filter.Page, "pageSize": filter.PageSize})
		return true

	case r.URL.Path == "/v1/attendance/records" && r.Method == "POST":
		var input domain.AdminAttendanceInput
		if !decode(w, r, &input) {
			return true
		}
		rec, err := s.Attendance.AdminCreate(r.Context(), a, input, r.Header.Get("Idempotency-Key"))
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, rec)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/attendance/records/") && strings.HasSuffix(r.URL.Path, "/history") && r.Method == "GET":
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/attendance/records/"), "/history")
		hist, err := s.Attendance.GetHistory(r.Context(), a, id)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, hist)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/attendance/records/") && strings.HasSuffix(r.URL.Path, "/approval") && r.Method == "POST":
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/attendance/records/"), "/approval")
		var input domain.ApprovalInput
		if !decode(w, r, &input) {
			return true
		}
		rec, err := s.Attendance.AdminUpdateApproval(r.Context(), a, id, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, rec)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/attendance/records/") && r.Method == "GET":
		id := strings.TrimPrefix(r.URL.Path, "/v1/attendance/records/")
		rec, err := s.Attendance.GetRecord(r.Context(), a, id)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, rec)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/attendance/records/") && r.Method == "PATCH":
		id := strings.TrimPrefix(r.URL.Path, "/v1/attendance/records/")
		var input domain.AdminCorrectionInput
		if !decode(w, r, &input) {
			return true
		}
		rec, err := s.Attendance.AdminCorrect(r.Context(), a, id, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, rec)
		return true

	case r.URL.Path == "/v1/attendance/leaves" && r.Method == "GET":
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		filter := domain.LeaveFilter{
			StaffID:  r.URL.Query().Get("staffId"),
			Status:   r.URL.Query().Get("status"),
			Page:     page,
			PageSize: pageSize,
		}
		leaves, total, err := s.Attendance.ListLeaveRequests(r.Context(), a, filter)
		if err != nil {
			fail(w, err)
			return true
		}
		if leaves == nil {
			leaves = []domain.LeaveRequest{}
		}
		write(w, 200, map[string]any{"leaves": leaves, "total": total, "page": filter.Page, "pageSize": filter.PageSize})
		return true

	case r.URL.Path == "/v1/attendance/leaves" && r.Method == "POST":
		var input domain.LeaveRequestInput
		if !decode(w, r, &input) {
			return true
		}
		lr, err := s.Attendance.CreateLeaveRequest(r.Context(), a, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, lr)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/attendance/leaves/") && strings.HasSuffix(r.URL.Path, "/status") && r.Method == "POST":
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/attendance/leaves/"), "/status")
		var input domain.LeaveApprovalInput
		if !decode(w, r, &input) {
			return true
		}
		lr, err := s.Attendance.UpdateLeaveStatus(r.Context(), a, id, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, lr)
		return true

	case r.URL.Path == "/v1/attendance/summary" && r.Method == "GET":
		date := r.URL.Query().Get("date")
		sum, err := s.Attendance.Summary(r.Context(), a, date)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, sum)
		return true

	case r.URL.Path == "/v1/attendance/staff" && r.Method == "GET":
		staff, err := s.Attendance.ListStaff(r.Context(), a)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"staff": staff})
		return true

	default:
		write(w, 404, map[string]string{"error": "Attendance endpoint not found", "code": "NOT_FOUND"})
		return true
	}
}
