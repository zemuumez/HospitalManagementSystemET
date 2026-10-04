package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) cmsSettings(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	// Hospital General Settings
	if strings.HasPrefix(r.URL.Path, "/v1/general-settings") {
		switch {
		case r.URL.Path == "/v1/general-settings" && r.Method == "GET":
			settings, err := s.CMSSettings.GeneralSettings(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"general_settings": settings})
			return true

		case r.URL.Path == "/v1/general-settings" && (r.Method == "POST" || r.Method == "PUT"):
			var body map[string]string
			if !decode(w, r, &body) {
				return true
			}
			if err := s.CMSSettings.UpdateGeneralSettings(r.Context(), a, body); err != nil {
				fail(w, err)
				return true
			}
			updated, err := s.CMSSettings.GeneralSettings(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"saved": true, "general_settings": updated})
			return true
		}
	}

	// Hospital Schedules
	if strings.HasPrefix(r.URL.Path, "/v1/hospital-schedules") {
		switch {
		case r.URL.Path == "/v1/hospital-schedules" && r.Method == "GET":
			schedules, err := s.CMSSettings.HospitalSchedules(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"hospital_schedules": schedules})
			return true

		case r.URL.Path == "/v1/hospital-schedules" && (r.Method == "POST" || r.Method == "PUT"):
			var in domain.HospitalScheduleDayInput
			if !decode(w, r, &in) {
				return true
			}
			day, err := s.CMSSettings.UpdateHospitalSchedule(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, day)
			return true
		}
	}

	// Front CMS Settings
	if strings.HasPrefix(r.URL.Path, "/v1/front-cms-settings") {
		switch {
		case r.URL.Path == "/v1/front-cms-settings" && r.Method == "GET":
			typeFilter := r.URL.Query().Get("type")
			settings, err := s.CMSSettings.FrontCMSSettings(r.Context(), a, typeFilter)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"front_cms_settings": settings})
			return true

		case r.URL.Path == "/v1/front-cms-settings" && (r.Method == "POST" || r.Method == "PUT"):
			var in domain.FrontCMSSettingInput
			if !decode(w, r, &in) {
				return true
			}
			setting, err := s.CMSSettings.UpdateFrontCMSSetting(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, setting)
			return true
		}
	}

	// Testimonials
	if strings.HasPrefix(r.URL.Path, "/v1/testimonials") {
		switch {
		case r.URL.Path == "/v1/testimonials" && r.Method == "GET":
			var statusFilter *int
			if raw := r.URL.Query().Get("status"); raw != "" {
				if sVal, err := strconv.Atoi(raw); err == nil {
					statusFilter = &sVal
				}
			}
			list, err := s.CMSSettings.Testimonials(r.Context(), a, statusFilter)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"testimonials": list})
			return true

		case r.URL.Path == "/v1/testimonials" && r.Method == "POST":
			var in domain.CMSTestimonialInput
			if !decode(w, r, &in) {
				return true
			}
			t, err := s.CMSSettings.CreateTestimonial(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, t)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/testimonials/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/testimonials/")
			t, err := s.CMSSettings.Testimonial(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, t)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/testimonials/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/testimonials/")
			var in domain.CMSTestimonialInput
			if !decode(w, r, &in) {
				return true
			}
			t, err := s.CMSSettings.UpdateTestimonial(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, t)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/testimonials/") && r.Method == "DELETE":
			id := strings.TrimPrefix(r.URL.Path, "/v1/testimonials/")
			if err := s.CMSSettings.DeleteTestimonial(r.Context(), a, id); err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]bool{"deleted": true})
			return true
		}
	}

	return false
}
