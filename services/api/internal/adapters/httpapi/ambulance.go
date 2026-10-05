package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) ambulance(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	if !strings.HasPrefix(r.URL.Path, "/v1/ambulances") && !strings.HasPrefix(r.URL.Path, "/v1/ambulance-calls") {
		return false
	}

	// /v1/ambulances routes
	if strings.HasPrefix(r.URL.Path, "/v1/ambulances") {
		switch {
		case r.URL.Path == "/v1/ambulances" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			if page < 1 {
				page = 1
			}
			var avail *bool
			if qAvail := r.URL.Query().Get("available"); qAvail != "" {
				b := qAvail == "true" || qAvail == "1"
				avail = &b
			}
			list, total, err := s.Ambulance.Ambulances(r.Context(), a, page, avail)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"ambulances": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/ambulances" && r.Method == "POST":
			var in domain.AmbulanceInput
			if !decode(w, r, &in) {
				return true
			}
			amb, err := s.Ambulance.CreateAmbulance(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, amb)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/ambulances/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/ambulances/")
			amb, err := s.Ambulance.Ambulance(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, amb)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/ambulances/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/ambulances/")
			var req struct {
				domain.AmbulanceInput
				Version int `json:"version"`
			}
			if !decode(w, r, &req) {
				return true
			}
			amb, err := s.Ambulance.UpdateAmbulance(r.Context(), a, id, req.AmbulanceInput, req.Version)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, amb)
			return true
		}
	}

	// /v1/ambulance-calls routes
	if strings.HasPrefix(r.URL.Path, "/v1/ambulance-calls") {
		switch {
		case r.URL.Path == "/v1/ambulance-calls" && r.Method == "GET":
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			if page < 1 {
				page = 1
			}
			status := r.URL.Query().Get("status")
			patientID := r.URL.Query().Get("patient_id")
			ambulanceID := r.URL.Query().Get("ambulance_id")

			list, total, err := s.Ambulance.Calls(r.Context(), a, page, status, patientID, ambulanceID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"ambulance_calls": list, "total": total, "page": page})
			return true

		case r.URL.Path == "/v1/ambulance-calls" && r.Method == "POST":
			var in domain.AmbulanceCallInput
			if !decode(w, r, &in) {
				return true
			}
			key := r.Header.Get("Idempotency-Key")
			call, err := s.Ambulance.CreateCall(r.Context(), a, in, key)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, call)
			return true

		case strings.HasSuffix(r.URL.Path, "/bill") && r.Method == "POST":
			callID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/ambulance-calls/"), "/bill")
			var in domain.SourceInvoiceInput
			if !decode(w, r, &in) {
				return true
			}
			key := r.Header.Get("Idempotency-Key")
			inv, err := s.Ambulance.BillCall(r.Context(), a, callID, in, key, "")
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, inv)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/ambulance-calls/") && r.Method == "GET":
			id := strings.TrimPrefix(r.URL.Path, "/v1/ambulance-calls/")
			call, err := s.Ambulance.Call(r.Context(), a, id)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, call)
			return true

		case strings.HasPrefix(r.URL.Path, "/v1/ambulance-calls/") && (r.Method == "PUT" || r.Method == "PATCH"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/ambulance-calls/")
			var in domain.AmbulanceCallUpdateInput
			if !decode(w, r, &in) {
				return true
			}
			call, err := s.Ambulance.UpdateCall(r.Context(), a, id, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, call)
			return true
		}
	}

	return false
}
