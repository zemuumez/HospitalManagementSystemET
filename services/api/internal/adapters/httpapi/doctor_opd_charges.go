package httpapi

import (
	"net/http"
	"strings"

	"hms.local/api/internal/domain"
)

type saveOPDChargeRequest struct {
	DoctorID            string   `json:"doctorId,omitempty"`
	DoctorIdSnake       string   `json:"doctor_id,omitempty"`
	StandardCharge      *float64 `json:"standardCharge,omitempty"`
	StandardChargeSnake *float64 `json:"standard_charge,omitempty"`
	CurrencySymbol      string   `json:"currencySymbol,omitempty"`
}

func (s Server) doctorOPDCharges(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	if r.URL.Path != "/v1/doctor-opd-charges" && !strings.HasPrefix(r.URL.Path, "/v1/doctor-opd-charges/") {
		return false
	}

	switch {
	case r.URL.Path == "/v1/doctor-opd-charges" && r.Method == "GET":
		search := r.URL.Query().Get("search")
		out, err := s.Scheduling.DoctorOPDCharges(r.Context(), a, search)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"charges": out})
		return true

	case r.URL.Path == "/v1/doctor-opd-charges" && r.Method == "POST":
		var req saveOPDChargeRequest
		if !decode(w, r, &req) {
			return true
		}
		docID := req.DoctorID
		if docID == "" {
			docID = req.DoctorIdSnake
		}
		var charge float64
		if req.StandardCharge != nil {
			charge = *req.StandardCharge
		} else if req.StandardChargeSnake != nil {
			charge = *req.StandardChargeSnake
		}
		in := domain.SaveDoctorOPDChargeInput{
			DoctorID:       docID,
			StandardCharge: charge,
			CurrencySymbol: req.CurrencySymbol,
		}
		out, err := s.Scheduling.SaveDoctorOPDCharge(r.Context(), a, in)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, out)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/doctor-opd-charges/") && r.Method == "GET":
		id := strings.TrimPrefix(r.URL.Path, "/v1/doctor-opd-charges/")
		if id == "" || strings.Contains(id, "/") {
			return false
		}
		out, err := s.Scheduling.DoctorOPDCharge(r.Context(), a, id)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, out)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/doctor-opd-charges/") && r.Method == "DELETE":
		id := strings.TrimPrefix(r.URL.Path, "/v1/doctor-opd-charges/")
		if id == "" || strings.Contains(id, "/") {
			return false
		}
		err := s.Scheduling.DeleteDoctorOPDCharge(r.Context(), a, id)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"deleted": true, "doctorId": id})
		return true
	}

	return false
}
