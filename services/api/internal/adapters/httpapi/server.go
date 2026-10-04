package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ActorStore interface {
	Actor(context.Context, string) (domain.Actor, error)
}
type Server struct {
	App         application.Hospital
	Scheduling  application.Scheduling
	Clinical    application.Clinical
	Billing     application.Billing
	Pharmacy    application.Pharmacy
	Diagnostics application.Diagnostics
	Actors      ActorStore
	AuthURL     string
	Origin      string
	Client      *http.Client
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "Unable to complete the request"
	if errors.Is(err, domain.ErrForbidden) {
		status = 403
		message = "You do not have permission for this action"
	}
	if errors.Is(err, domain.ErrValidation) {
		status = 422
		message = "Check the supplied fields and try again"
	}
	if errors.Is(err, domain.ErrConflict) {
		status = 409
		message = "This request key was already used for different data"
	}
	if errors.Is(err, domain.ErrNotFound) {
		status = 404
		message = "Record not found"
	}
	if errors.Is(err, domain.ErrStale) {
		status = 409
		message = "The record changed or the selected time is unavailable. Refresh and try again."
	}
	write(w, status, map[string]string{"error": message})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(v) != nil || d.Decode(&struct{}{}) != io.EOF {
		write(w, 400, map[string]string{"error": "Invalid JSON request"})
		return false
	}
	return true
}
func (s Server) identify(r *http.Request) (domain.Actor, error) {
	if r.Header.Get("Cookie") == "" {
		return domain.Actor{}, domain.ErrForbidden
	}
	req, err := http.NewRequestWithContext(r.Context(), "GET", s.AuthURL+"/api/auth/get-session", nil)
	if err != nil {
		return domain.Actor{}, err
	}
	req.Header.Set("Cookie", r.Header.Get("Cookie"))
	req.Header.Set("Accept", "application/json")
	resp, err := s.Client.Do(req)
	if err != nil {
		return domain.Actor{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return domain.Actor{}, domain.ErrForbidden
	}
	var result struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
		Session struct {
			UserID    string    `json:"userId"`
			ExpiresAt time.Time `json:"expiresAt"`
		} `json:"session"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&result) != nil || result.User.ID == "" || result.Session.UserID != result.User.ID || !result.Session.ExpiresAt.After(time.Now()) {
		return domain.Actor{}, domain.ErrForbidden
	}
	return s.Actors.Actor(r.Context(), result.User.ID)
}
func (s Server) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" && r.Method == "GET" {
			write(w, 200, map[string]string{"status": "ok"})
			return
		}
		if r.Method != "GET" && r.Header.Get("Origin") != s.Origin {
			write(w, 403, map[string]string{"error": "Request origin is not allowed"})
			return
		}
		a, err := s.identify(r)
		if err != nil {
			write(w, 401, map[string]string{"error": "Sign in to continue"})
			return
		}
		if s.diagnostics(w, r, a) {
			return
		}
		if s.patientProfile(w, r, a) {
			return
		}
		if s.pharmacy(w, r, a) {
			return
		}
		if s.billing(w, r, a) {
			return
		}
		if s.clinical(w, r, a) {
			return
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/v1/patients/") && r.Method == "PATCH":
			var access domain.PatientAccess
			if !decode(w, r, &access) {
				return
			}
			if e := s.App.LinkPatient(r.Context(), a, strings.TrimPrefix(r.URL.Path, "/v1/patients/"), access); e != nil {
				fail(w, e)
				return
			}
			write(w, 200, map[string]bool{"saved": true})
		case r.URL.Path == "/v1/doctors" && r.Method == "GET":
			out, e := s.Scheduling.Doctors(r.Context(), a)
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, map[string]any{"doctors": out})
		case r.URL.Path == "/v1/doctors" && r.Method == "POST":
			var d domain.Doctor
			if !decode(w, r, &d) {
				return
			}
			out, e := s.Scheduling.SaveDoctor(r.Context(), a, d)
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, out)
		case r.URL.Path == "/v1/slots" && r.Method == "GET":
			out, e := s.Scheduling.Slots(r.Context(), a, r.URL.Query().Get("doctorId"), r.URL.Query().Get("date"))
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, map[string]any{"slots": out, "timezone": "Africa/Addis_Ababa"})
		case r.URL.Path == "/v1/appointments" && r.Method == "GET":
			page := 1
			if raw := r.URL.Query().Get("page"); raw != "" {
				page, _ = strconv.Atoi(raw)
			}
			out, e := s.Scheduling.Appointments(r.Context(), a, page)
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, map[string]any{"appointments": out, "page": page, "pageSize": 25})
		case r.URL.Path == "/v1/appointments" && r.Method == "POST":
			var i domain.AppointmentInput
			if !decode(w, r, &i) {
				return
			}
			out, e := s.Scheduling.Book(r.Context(), a, i, r.Header.Get("Idempotency-Key"))
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 201, out)
		case strings.HasPrefix(r.URL.Path, "/v1/appointments/") && r.Method == "PATCH":
			var c domain.AppointmentChange
			if !decode(w, r, &c) {
				return
			}
			out, e := s.Scheduling.Change(r.Context(), a, strings.TrimPrefix(r.URL.Path, "/v1/appointments/"), c)
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, out)
		case r.URL.Path == "/v1/me" && r.Method == "GET":
			write(w, 200, map[string]any{"user": a, "permissions": a.Permissions()})
		case r.URL.Path == "/v1/overview" && r.Method == "GET":
			o, e := s.App.Overview(r.Context(), a)
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, o)
		case r.URL.Path == "/v1/patients" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			if page == 0 {
				page = 1
			}
			p, e := s.App.Patients(r.Context(), a, r.URL.Query().Get("search"), page)
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, map[string]any{"patients": p, "page": page, "pageSize": 25})
		case r.URL.Path == "/v1/patients" && r.Method == "POST":
			var p domain.PatientInput
			if !decode(w, r, &p) {
				return
			}
			out, e := s.App.Register(r.Context(), a, p)
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 201, out)
		case r.URL.Path == "/v1/messages" && r.Method == "GET":
			out, e := s.App.Messages(r.Context(), a)
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 200, map[string]any{"messages": out})
		case r.URL.Path == "/v1/messages" && r.Method == "POST":
			var m domain.MessageInput
			if !decode(w, r, &m) {
				return
			}
			out, e := s.App.Enqueue(r.Context(), a, m, r.Header.Get("Idempotency-Key"))
			if e != nil {
				fail(w, e)
				return
			}
			write(w, 202, out)
		default:
			write(w, 404, map[string]string{"error": "Endpoint not found"})
		}
	})
}
