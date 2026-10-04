package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) schedulingChanges(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, _ = strconv.Atoi(raw)
	}
	switch {
	case r.URL.Path == "/v1/doctor-absences" && r.Method == "GET":
		out, e := s.Scheduling.Absences(r.Context(), a, r.URL.Query().Get("doctorId"), page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"absences": out, "page": page, "pageSize": 25})
		}
	case r.URL.Path == "/v1/doctor-absences" && r.Method == "POST":
		var i domain.AbsenceInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Scheduling.CreateAbsence(r.Context(), a, i)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 201, out)
		}
	case strings.HasPrefix(r.URL.Path, "/v1/doctor-absences/") && r.Method == "PATCH":
		var i domain.AbsenceCancel
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Scheduling.CancelAbsence(r.Context(), a, strings.TrimPrefix(r.URL.Path, "/v1/doctor-absences/"), i)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, out)
		}
	case strings.HasPrefix(r.URL.Path, "/v1/appointments/"):
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/appointments/"), "/")
		if len(parts) != 2 {
			return false
		}
		if parts[1] == "reschedule" && r.Method == "POST" {
			var i domain.RescheduleInput
			if !decode(w, r, &i) {
				return true
			}
			out, e := s.Scheduling.Reschedule(r.Context(), a, parts[0], i)
			if e != nil {
				fail(w, e)
			} else {
				write(w, 200, out)
			}
		} else if parts[1] == "status-history" && r.Method == "GET" {
			out, e := s.Scheduling.StatusHistory(r.Context(), a, parts[0], page)
			if e != nil {
				fail(w, e)
			} else {
				write(w, 200, map[string]any{"events": out, "page": page, "pageSize": 25})
			}
		} else if parts[1] == "reschedule-history" && r.Method == "GET" {
			out, e := s.Scheduling.RescheduleHistory(r.Context(), a, parts[0], page)
			if e != nil {
				fail(w, e)
			} else {
				write(w, 200, map[string]any{"events": out, "page": page, "pageSize": 25})
			}
		} else {
			return false
		}
	default:
		return false
	}
	return true
}
