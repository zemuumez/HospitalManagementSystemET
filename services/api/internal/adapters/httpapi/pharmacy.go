package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) pharmacy(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, _ = strconv.Atoi(raw)
	}
	var out any
	var e error
	status := 200
	switch {
	case r.URL.Path == "/v1/medicines" && r.Method == "GET":
		var data []domain.Medicine
		data, e = s.Pharmacy.Medicines(r.Context(), a, r.URL.Query().Get("search"), page)
		out = map[string]any{"medicines": data}
	case r.URL.Path == "/v1/medicines" && r.Method == "POST":
		var i domain.MedicineInput
		if !decode(w, r, &i) {
			return true
		}
		out, e = s.Pharmacy.CreateMedicine(r.Context(), a, i)
		status = 201
	case r.URL.Path == "/v1/medicine-batches" && r.Method == "GET":
		var data []domain.MedicineBatch
		data, e = s.Pharmacy.Batches(r.Context(), a, r.URL.Query().Get("medicineId"), page)
		out = map[string]any{"batches": data}
	case r.URL.Path == "/v1/medicine-batches" && r.Method == "POST":
		var i domain.BatchInput
		if !decode(w, r, &i) {
			return true
		}
		out, e = s.Pharmacy.Receive(r.Context(), a, i, r.Header.Get("Idempotency-Key"))
		status = 201
	case r.URL.Path == "/v1/medication-orders" && r.Method == "GET":
		var data []domain.MedicationOrder
		data, e = s.Pharmacy.Orders(r.Context(), a, r.URL.Query().Get("encounterId"), page)
		out = map[string]any{"orders": data}
	case r.URL.Path == "/v1/medication-orders" && r.Method == "POST":
		var i domain.MedicationInput
		if !decode(w, r, &i) {
			return true
		}
		out, e = s.Pharmacy.Sign(r.Context(), a, i, r.Header.Get("Idempotency-Key"))
		status = 201
	case strings.HasPrefix(r.URL.Path, "/v1/medication-orders/") && strings.HasSuffix(r.URL.Path, "/cancel") && r.Method == "POST":
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/medication-orders/"), "/")
		if len(parts) != 2 {
			return false
		}
		var i struct {
			Reason string `json:"reason"`
		}
		if !decode(w, r, &i) {
			return true
		}
		e = s.Pharmacy.Cancel(r.Context(), a, parts[0], i.Reason)
		out = map[string]bool{"cancelled": true}
	case r.URL.Path == "/v1/pharmacy-movements" && r.Method == "GET":
		var data []domain.StockMovement
		data, e = s.Pharmacy.Movements(r.Context(), a, r.URL.Query().Get("batchId"), page)
		out = map[string]any{"movements": data}
	case r.URL.Path == "/v1/pharmacy-movements" && r.Method == "POST":
		var i domain.StockInput
		if !decode(w, r, &i) {
			return true
		}
		out, e = s.Pharmacy.Move(r.Context(), a, i, r.Header.Get("Idempotency-Key"))
		status = 201
	default:
		return false
	}
	if e != nil {
		fail(w, e)
	} else {
		write(w, status, out)
	}
	return true
}
