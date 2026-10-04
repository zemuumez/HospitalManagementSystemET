package httpapi

import (
	"hms.local/api/internal/domain"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) paymentWebhook(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != "/v1/webhooks/stripe" || r.Method != "POST" {
		return false
	}
	body, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<10))
	if e != nil {
		write(w, 413, map[string]string{"error": "Webhook body too large", "code": "INVALID_BODY"})
		return true
	}
	if e = s.OnlinePayments.Webhook(r.Context(), body, r.Header.Get("Stripe-Signature")); e != nil {
		fail(w, e)
	} else {
		write(w, 200, map[string]bool{"received": true})
	}
	return true
}
func (s Server) onlinePayments(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	if r.URL.Path == "/v1/payment-reviews" && r.Method == "GET" {
		page := 1
		if raw := r.URL.Query().Get("page"); raw != "" {
			page, _ = strconv.Atoi(raw)
		}
		out, e := s.OnlinePayments.Reviews(r.Context(), a, page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"events": out, "page": page, "pageSize": 25})
		}
		return true
	}
	if !strings.HasPrefix(r.URL.Path, "/v1/invoices/") || !strings.HasSuffix(r.URL.Path, "/checkout") || r.Method != "POST" {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/invoices/"), "/")
	if len(parts) != 2 {
		return false
	}
	var input struct{}
	if !decode(w, r, &input) {
		return true
	}
	out, e := s.OnlinePayments.Checkout(r.Context(), a, parts[0])
	if e != nil {
		fail(w, e)
	} else {
		write(w, 200, out)
	}
	return true
}
