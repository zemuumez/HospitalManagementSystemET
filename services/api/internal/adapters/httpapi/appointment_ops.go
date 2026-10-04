package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"hms.local/api/internal/domain"
)

func (s Server) publicAppointmentOps(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path == "/v1/public/appointment-requests" && r.Method == "POST" {
		var input domain.CreatePublicAppointmentRequestInput
		if !decode(w, r, &input) {
			return true
		}
		req, err := s.AppointmentOps.CreatePublicRequest(r.Context(), input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, req)
		return true
	}
	return false
}

func (s Server) appointmentOps(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	switch {
	// ─── Public Appointment Requests (Staff Management) ─────────────────────────
	case r.URL.Path == "/v1/appointment-requests" && r.Method == "GET":
		status := r.URL.Query().Get("status")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		list, err := s.AppointmentOps.ListPublicRequests(r.Context(), a, status, page)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"requests": list})
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/appointment-requests/") && strings.HasSuffix(r.URL.Path, "/review") && r.Method == "POST":
		id := strings.TrimPrefix(r.URL.Path, "/v1/appointment-requests/")
		id = strings.TrimSuffix(id, "/review")
		var input domain.ReviewPublicAppointmentRequestInput
		if !decode(w, r, &input) {
			return true
		}
		req, err := s.AppointmentOps.ReviewPublicRequest(r.Context(), a, id, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, req)
		return true

	// ─── Patient Queues ──────────────────────────────────────────────────────────
	case r.URL.Path == "/v1/patient-queues" && r.Method == "GET":
		docID := r.URL.Query().Get("doctorId")
		date := r.URL.Query().Get("date")
		items, err := s.AppointmentOps.DoctorQueue(r.Context(), a, docID, date)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"queue": items})
		return true

	case r.URL.Path == "/v1/patient-queues" && r.Method == "POST":
		var input domain.EnqueuePatientInput
		if !decode(w, r, &input) {
			return true
		}
		item, err := s.AppointmentOps.EnqueuePatient(r.Context(), a, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, item)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/patient-queues/") && r.Method == "PATCH":
		id := strings.TrimPrefix(r.URL.Path, "/v1/patient-queues/")
		var input domain.UpdateQueueStatusInput
		if !decode(w, r, &input) {
			return true
		}
		item, err := s.AppointmentOps.UpdateQueueStatus(r.Context(), a, id, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, item)
		return true

	// ─── Appointment Billing ─────────────────────────────────────────────────────
	case strings.HasPrefix(r.URL.Path, "/v1/appointments/"):
		rest := strings.TrimPrefix(r.URL.Path, "/v1/appointments/")
		parts := strings.Split(rest, "/")
		if len(parts) >= 2 && parts[1] == "billing" {
			apptID := parts[0]
			if len(parts) == 2 && r.Method == "GET" {
				rec, err := s.AppointmentOps.GetAppointmentBilling(r.Context(), a, apptID)
				if err != nil {
					fail(w, err)
					return true
				}
				write(w, 200, rec)
				return true
			}
			if len(parts) == 3 && parts[2] == "fee" && r.Method == "POST" {
				var body struct {
					FeeMinor int64 `json:"feeMinor"`
				}
				if !decode(w, r, &body) {
					return true
				}
				rec, err := s.AppointmentOps.SetAppointmentFee(r.Context(), a, apptID, body.FeeMinor)
				if err != nil {
					fail(w, err)
					return true
				}
				write(w, 200, rec)
				return true
			}
			if len(parts) == 3 && parts[2] == "invoice" && r.Method == "POST" {
				var body struct {
					InvoiceID string `json:"invoiceId"`
				}
				if !decode(w, r, &body) {
					return true
				}
				rec, err := s.AppointmentOps.LinkAppointmentInvoice(r.Context(), a, apptID, body.InvoiceID)
				if err != nil {
					fail(w, err)
					return true
				}
				write(w, 200, rec)
				return true
			}
		}
	}

	return false
}
