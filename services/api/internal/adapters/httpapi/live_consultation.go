package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) liveConsultations(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	// Provider settings
	if r.URL.Path == "/v1/live-consultation-settings" || r.URL.Path == "/v1/live-consultations/provider-settings" {
		switch {
		case r.Method == "GET":
			setting, err := s.LiveConsultation.ProviderSetting(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, setting)
			return true

		case r.Method == "POST" || r.Method == "PUT":
			var in domain.LiveProviderSettingInput
			if !decode(w, r, &in) {
				return true
			}
			setting, err := s.LiveConsultation.UpdateProviderSetting(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, setting)
			return true
		}
	}

	// Live Consultations
	if strings.HasPrefix(r.URL.Path, "/v1/live-consultations") {
		switch {
		case r.URL.Path == "/v1/live-consultations" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			var statusFilter *int
			if raw := r.URL.Query().Get("status"); raw != "" {
				if sVal, err := strconv.Atoi(raw); err == nil {
					statusFilter = &sVal
				}
			}
			list, total, err := s.LiveConsultation.LiveConsultations(r.Context(), a, page, statusFilter)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"live_consultations": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/live-consultations" && r.Method == "POST":
			var in domain.LiveConsultationInput
			if !decode(w, r, &in) {
				return true
			}
			c, err := s.LiveConsultation.CreateLiveConsultation(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, c)
			return true

		case strings.HasSuffix(r.URL.Path, "/status") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/live-consultations/")
			id = strings.TrimSuffix(id, "/status")
			var in struct {
				Status int `json:"status"`
			}
			if !decode(w, r, &in) {
				return true
			}
			c, err := s.LiveConsultation.UpdateLiveConsultationStatus(r.Context(), a, id, in.Status)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, c)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/live-consultations/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/live-consultations/")
			c, err := s.LiveConsultation.LiveConsultation(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, c)
			return true
		}
	}

	// Live Meetings
	if strings.HasPrefix(r.URL.Path, "/v1/live-meetings") {
		switch {
		case r.URL.Path == "/v1/live-meetings" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			var statusFilter *int
			if raw := r.URL.Query().Get("status"); raw != "" {
				if sVal, err := strconv.Atoi(raw); err == nil {
					statusFilter = &sVal
				}
			}
			list, total, err := s.LiveConsultation.LiveMeetings(r.Context(), a, page, statusFilter)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"live_meetings": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/live-meetings" && r.Method == "POST":
			var in domain.LiveMeetingInput
			if !decode(w, r, &in) {
				return true
			}
			m, err := s.LiveConsultation.CreateLiveMeeting(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, m)
			return true

		case strings.HasSuffix(r.URL.Path, "/status") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/live-meetings/")
			id = strings.TrimSuffix(id, "/status")
			var in struct {
				Status int `json:"status"`
			}
			if !decode(w, r, &in) {
				return true
			}
			m, err := s.LiveConsultation.UpdateLiveMeetingStatus(r.Context(), a, id, in.Status)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, m)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/live-meetings/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/live-meetings/")
			m, err := s.LiveConsultation.LiveMeeting(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, m)
			return true
		}
	}

	return false
}
