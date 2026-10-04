package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) billing(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, _ = strconv.Atoi(raw)
	}
	switch {
	case strings.HasPrefix(r.URL.Path, "/v1/diagnostic-orders/") && strings.HasSuffix(r.URL.Path, "/invoice") && r.Method == "POST":
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/diagnostic-orders/"), "/")
		if len(parts) != 2 {
			return false
		}
		var i domain.PharmacyInvoiceInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Billing.BillDiagnostic(r.Context(), a, parts[0], i, r.Header.Get("Idempotency-Key"))
		if e != nil {
			fail(w, e)
		} else {
			write(w, 201, out)
		}
	case strings.HasPrefix(r.URL.Path, "/v1/pharmacy-movements/") && strings.HasSuffix(r.URL.Path, "/invoice") && r.Method == "POST":
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/pharmacy-movements/"), "/")
		if len(parts) != 2 {
			return false
		}
		var i domain.PharmacyInvoiceInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Billing.BillDispensing(r.Context(), a, parts[0], i, r.Header.Get("Idempotency-Key"))
		if e != nil {
			fail(w, e)
		} else {
			write(w, 201, out)
		}
	case r.URL.Path == "/v1/billing-patients" && r.Method == "GET":
		out, e := s.Billing.Patients(r.Context(), a, r.URL.Query().Get("search"), page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"patients": out})
		}
	case r.URL.Path == "/v1/charge-accounts" && r.Method == "GET":
		out, e := s.Billing.Accounts(r.Context(), a, page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"accounts": out})
		}
	case r.URL.Path == "/v1/charge-accounts" && r.Method == "POST":
		var i struct {
			Name string `json:"name"`
		}
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Billing.CreateAccount(r.Context(), a, i.Name)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 201, out)
		}
	case r.URL.Path == "/v1/invoices" && r.Method == "GET":
		out, e := s.Billing.Invoices(r.Context(), a, page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"invoices": out})
		}
	case r.URL.Path == "/v1/invoices" && r.Method == "POST":
		var i domain.InvoiceInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Billing.CreateInvoice(r.Context(), a, i, r.Header.Get("Idempotency-Key"))
		if e != nil {
			fail(w, e)
		} else {
			write(w, 201, out)
		}
	case strings.HasPrefix(r.URL.Path, "/v1/invoices/"):
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/invoices/"), "/")
		id := parts[0]
		switch {
		case len(parts) == 1 && r.Method == "GET":
			out, e := s.Billing.Invoice(r.Context(), a, id)
			if e != nil {
				fail(w, e)
			} else {
				write(w, 200, out)
			}
		case len(parts) == 2 && parts[1] == "payments" && r.Method == "GET":
			out, e := s.Billing.Payments(r.Context(), a, id)
			if e != nil {
				fail(w, e)
			} else {
				write(w, 200, map[string]any{"payments": out})
			}
		case len(parts) == 2 && parts[1] == "payments" && r.Method == "POST":
			var i domain.PaymentInput
			if !decode(w, r, &i) {
				return true
			}
			out, e := s.Billing.RecordPayment(r.Context(), a, id, i, r.Header.Get("Idempotency-Key"))
			if e != nil {
				fail(w, e)
			} else {
				write(w, 201, out)
			}
		default:
			return false
		}
	default:
		return false
	}
	return true
}
