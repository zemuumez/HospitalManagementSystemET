package httpapi

import (
	"net/http"
	"strings"

	"hms.local/api/internal/domain"
)

type saveScheduleRequest struct {
	DoctorID       string                  `json:"doctorId,omitempty"`
	DoctorIdSnake  string                  `json:"doctor_id,omitempty"`
	PerPatientTime string                  `json:"perPatientTime,omitempty"`
	SlotMinutes    int                     `json:"slotMinutes,omitempty"`
	Days           []domain.ScheduleDayRow `json:"days,omitempty"`
}

type createHolidayRequest struct {
	DoctorID      string `json:"doctorId,omitempty"`
	DoctorIdSnake string `json:"doctor_id,omitempty"`
	Date          string `json:"date"`
	Reason        string `json:"reason,omitempty"`
	Name          string `json:"name,omitempty"`
}

type createBreakRequest struct {
	DoctorID       string `json:"doctorId,omitempty"`
	DoctorIdSnake  string `json:"doctor_id,omitempty"`
	BreakFrom      string `json:"breakFrom,omitempty"`
	BreakFromSnake string `json:"break_from,omitempty"`
	BreakTo        string `json:"breakTo,omitempty"`
	BreakToSnake   string `json:"break_to,omitempty"`
	EveryDay       *bool  `json:"everyDay,omitempty"`
	EveryDaySnake  *bool  `json:"every_day,omitempty"`
	Date           string `json:"date,omitempty"`
}

func (s Server) doctorSchedules(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	// 1. Doctor schedules
	if r.URL.Path == "/v1/doctor-schedules" || strings.HasPrefix(r.URL.Path, "/v1/doctor-schedules/") {
		switch {
		case r.URL.Path == "/v1/doctor-schedules" && r.Method == "GET":
			docID := r.URL.Query().Get("doctorId")
			if docID == "" {
				docID = r.URL.Query().Get("doctor_id")
			}
			out, err := s.Scheduling.DoctorSchedules(r.Context(), a, docID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"schedules": out})
			return true

		case r.URL.Path == "/v1/doctor-schedules" && r.Method == "POST":
			var req saveScheduleRequest
			if !decode(w, r, &req) {
				return true
			}
			docID := req.DoctorID
			if docID == "" {
				docID = req.DoctorIdSnake
			}
			in := domain.SaveDoctorScheduleInput{
				DoctorID:       docID,
				PerPatientTime: req.PerPatientTime,
				SlotMinutes:    req.SlotMinutes,
				Days:           req.Days,
			}
			out, err := s.Scheduling.SaveDoctorSchedule(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, out)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/doctor-schedules/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/doctor-schedules/")
			if id == "" || strings.Contains(id, "/") {
				return false
			}
			out, err := s.Scheduling.DoctorSchedule(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, out)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/doctor-schedules/") && r.Method == "DELETE":
			id := strings.TrimPrefix(r.URL.Path, "/v1/doctor-schedules/")
			if id == "" || strings.Contains(id, "/") {
				return false
			}
			err := s.Scheduling.DeleteDoctorSchedule(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"deleted": true, "doctorId": id})
			return true
		}
	}

	// 2. Doctor holidays
	if r.URL.Path == "/v1/doctor-holidays" || strings.HasPrefix(r.URL.Path, "/v1/doctor-holidays/") {
		switch {
		case r.URL.Path == "/v1/doctor-holidays" && r.Method == "GET":
			docID := r.URL.Query().Get("doctorId")
			if docID == "" {
				docID = r.URL.Query().Get("doctor_id")
			}
			out, err := s.Scheduling.DoctorHolidays(r.Context(), a, docID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"holidays": out})
			return true

		case r.URL.Path == "/v1/doctor-holidays" && r.Method == "POST":
			var req createHolidayRequest
			if !decode(w, r, &req) {
				return true
			}
			docID := req.DoctorID
			if docID == "" {
				docID = req.DoctorIdSnake
			}
			reason := req.Reason
			if reason == "" {
				reason = req.Name
			}
			in := domain.CreateDoctorHolidayInput{
				DoctorID: docID,
				Date:     req.Date,
				Reason:   reason,
			}
			out, err := s.Scheduling.CreateDoctorHoliday(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, out)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/doctor-holidays/") && r.Method == "DELETE":
			id := strings.TrimPrefix(r.URL.Path, "/v1/doctor-holidays/")
			if id == "" || strings.Contains(id, "/") {
				return false
			}
			err := s.Scheduling.DeleteDoctorHoliday(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"deleted": true, "id": id})
			return true
		}
	}

	// 3. Doctor lunch breaks
	if r.URL.Path == "/v1/doctor-breaks" || strings.HasPrefix(r.URL.Path, "/v1/doctor-breaks/") {
		switch {
		case r.URL.Path == "/v1/doctor-breaks" && r.Method == "GET":
			docID := r.URL.Query().Get("doctorId")
			if docID == "" {
				docID = r.URL.Query().Get("doctor_id")
			}
			out, err := s.Scheduling.DoctorBreaks(r.Context(), a, docID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"breaks": out})
			return true

		case r.URL.Path == "/v1/doctor-breaks" && r.Method == "POST":
			var req createBreakRequest
			if !decode(w, r, &req) {
				return true
			}
			docID := req.DoctorID
			if docID == "" {
				docID = req.DoctorIdSnake
			}
			bFrom := req.BreakFrom
			if bFrom == "" {
				bFrom = req.BreakFromSnake
			}
			bTo := req.BreakTo
			if bTo == "" {
				bTo = req.BreakToSnake
			}
			everyDay := true
			if req.EveryDay != nil {
				everyDay = *req.EveryDay
			} else if req.EveryDaySnake != nil {
				everyDay = *req.EveryDaySnake
			} else if req.Date != "" {
				everyDay = false
			}
			in := domain.CreateDoctorBreakInput{
				DoctorID:  docID,
				BreakFrom: bFrom,
				BreakTo:   bTo,
				EveryDay:  everyDay,
				Date:      req.Date,
			}
			out, err := s.Scheduling.CreateDoctorBreak(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, out)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/doctor-breaks/") && r.Method == "DELETE":
			id := strings.TrimPrefix(r.URL.Path, "/v1/doctor-breaks/")
			if id == "" || strings.Contains(id, "/") {
				return false
			}
			err := s.Scheduling.DeleteDoctorBreak(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"deleted": true, "id": id})
			return true
		}
	}

	return false
}
