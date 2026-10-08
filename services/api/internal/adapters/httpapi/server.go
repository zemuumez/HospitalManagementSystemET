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
	Staff              application.Staff
	Audit              application.Audit
	OnlinePayments     application.OnlinePayments
	Inventory          application.Inventory
	Attendance         application.Attendance
	Ambulance          application.AmbulanceService
	Packages           application.PackagesService
	Insurances         application.InsurancesService
	ServicesOperations application.ServicesOperationsService
	CMSSettings        application.CMSSettingsService
	FrontOffice        application.FrontOfficeService
	LiveConsultation   application.LiveConsultationService
	PharmacyBloodBank  application.PharmacyBloodBankService
	FinancePayroll     application.FinancePayrollService
	MasterData         application.MasterDataService
	PatientExt         application.PatientExtensionsService
	AppointmentOps     application.AppointmentOpsService
	ClinicalCare       application.ClinicalCareService
	DiagnosticReports  application.DiagnosticReportsService
	Attachments        application.AttachmentsService
	App                application.Hospital
	Scheduling         application.Scheduling
	Clinical           application.Clinical
	Billing            application.Billing
	Pharmacy           application.Pharmacy
	Diagnostics        application.Diagnostics
	Ready              func(context.Context) error
	Actors             ActorStore
	AuthURL            string
	Origin             string
	Client             *http.Client
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
	code := "INTERNAL_ERROR"
	if errors.Is(err, domain.ErrUnavailable) {
		code = "DEPENDENCY_UNAVAILABLE"
		status = 503
		message = "This service is not configured or is temporarily unavailable"
	}
	if errors.Is(err, domain.ErrForbidden) {
		code = "FORBIDDEN"
		status = 403
		message = "You do not have permission for this action"
	}
	if errors.Is(err, domain.ErrValidation) {
		code = "VALIDATION_FAILED"
		status = 422
		message = "Check the supplied fields and try again"
	}
	if errors.Is(err, domain.ErrExportLimitExceeded) {
		code = "EXPORT_LIMIT_EXCEEDED"
		status = 422
		message = "Export exceeds maximum limit of 5,000 records. Please refine search filters to reduce the result set."
	}
	if errors.Is(err, domain.ErrConflict) {
		code = "IDEMPOTENCY_CONFLICT"
		status = 409
		message = "This request key was already used for different data"
	}
	if errors.Is(err, domain.ErrNotFound) {
		code = "NOT_FOUND"
		status = 404
		message = "Record not found"
	}
	if errors.Is(err, domain.ErrStale) {
		code = "STATE_CONFLICT"
		status = 409
		message = "The record changed or the requested action is no longer available. Refresh and try again."
	}
	if errors.Is(err, domain.ErrInUse) {
		code = "RECORD_IN_USE"
		status = 409
		message = "Record is in use by patient admissions and cannot be deleted"
	}
	write(w, status, map[string]string{"error": message, "code": code})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(v) != nil || d.Decode(&struct{}{}) != io.EOF {
		write(w, 400, map[string]string{"error": "Invalid JSON request", "code": "INVALID_JSON"})
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
	limiter := &publicLimiter{}
	return observe(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/readyz" && r.Method == "GET" {
			ctx, cancel := context.WithTimeout(r.Context(), time.Second)
			defer cancel()
			if s.Ready == nil || s.Ready(ctx) != nil {
				write(w, 503, map[string]string{"status": "unavailable"})
				return
			}
			write(w, 200, map[string]string{"status": "ready"})
			return
		}
		if r.URL.Path == "/healthz" && r.Method == "GET" {
			write(w, 200, map[string]string{"status": "ok"})
			return
		}
		if r.URL.Path == "/metrics" && r.Method == "GET" {
			s.metrics(w, r)
			return
		}
		if s.paymentWebhook(w, r) {
			return
		}
		if !limiter.guard(w, r, s.Origin) {
			return
		}
		if s.publicPatientExtensions(w, r) {
			return
		}
		if s.publicAppointmentOps(w, r) {
			return
		}
		if r.Method != "GET" && r.Header.Get("Origin") != s.Origin {
			write(w, 403, map[string]string{"error": "Request origin is not allowed", "code": "ORIGIN_DENIED"})
			return
		}
		a, err := s.identify(r)
		if err != nil {
			write(w, 401, map[string]string{"error": "Sign in to continue", "code": "UNAUTHENTICATED"})
			return
		}
		if s.staff(w, r, a) {
			return
		}
		if s.attendance(w, r, a) {
			return
		}
		if s.ambulance(w, r, a) {
			return
		}
		if s.packages(w, r, a) {
			return
		}
		if s.insurances(w, r, a) {
			return
		}
		if s.servicesOperations(w, r, a) {
			return
		}
		if s.cmsSettings(w, r, a) {
			return
		}
		if s.frontOffice(w, r, a) {
			return
		}
		if s.liveConsultations(w, r, a) {
			return
		}
		if s.pharmacyBloodBank(w, r, a) {
			return
		}
		if s.financePayroll(w, r, a) {
			return
		}
		if s.masterData(w, r, a) {
			return
		}
		if s.patientExtensions(w, r, a) {
			return
		}
		if s.appointmentOps(w, r, a) {
			return
		}
		if s.clinicalCare(w, r, a) {
			return
		}
		if s.diagnosticReports(w, r, a) {
			return
		}
		if s.attachments(w, r, a) {
			return
		}
		if s.audit(w, r, a) {
			return
		}
		if s.inventory(w, r, a) {
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
		if s.onlinePayments(w, r, a) {
			return
		}
		if s.billing(w, r, a) {
			return
		}
		if s.schedulingChanges(w, r, a) {
			return
		}
		if s.nursing(w, r, a) {
			return
		}
		if s.clinical(w, r, a) {
			return
		}
		if s.doctors(w, r, a) {
			return
		}
		if s.doctorSchedules(w, r, a) {
			return
		}
		// Step 4: s.doctorOPDCharges
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
			write(w, 404, map[string]string{"error": "Endpoint not found", "code": "NOT_FOUND"})
		}
	}))
}
