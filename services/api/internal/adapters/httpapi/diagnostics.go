package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) diagnostics(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, _ = strconv.Atoi(raw)
	}
	var out any
	var e error
	status := 200
	switch {
	case r.URL.Path == "/v1/diagnostic-tests" && r.Method == "GET":
		var data []domain.DiagnosticTest
		data, e = s.Diagnostics.Tests(r.Context(), a, r.URL.Query().Get("search"), page)
		out = map[string]any{"tests": data}
	case r.URL.Path == "/v1/diagnostic-tests" && r.Method == "POST":
		var i domain.DiagnosticTestInput
		if !decode(w, r, &i) {
			return true
		}
		out, e = s.Diagnostics.CreateTest(r.Context(), a, i)
		status = 201
	case r.URL.Path == "/v1/diagnostic-orders" && r.Method == "GET":
		var data []domain.DiagnosticOrder
		data, e = s.Diagnostics.Orders(r.Context(), a, r.URL.Query().Get("encounterId"), page)
		out = map[string]any{"orders": data}
	case r.URL.Path == "/v1/diagnostic-orders" && r.Method == "POST":
		var i domain.DiagnosticOrderInput
		if !decode(w, r, &i) {
			return true
		}
		out, e = s.Diagnostics.Order(r.Context(), a, i, r.Header.Get("Idempotency-Key"))
		status = 201
	case strings.HasPrefix(r.URL.Path, "/v1/diagnostic-orders/"):
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/diagnostic-orders/"), "/")
		if len(parts) != 2 {
			return false
		}
		switch {
		case parts[1] == "actions" && r.Method == "POST":
			var i domain.DiagnosticAction
			if !decode(w, r, &i) {
				return true
			}
			out, e = s.Diagnostics.Transition(r.Context(), a, parts[0], i)
		case parts[1] == "results" && r.Method == "POST":
			var i domain.DiagnosticResultInput
			if !decode(w, r, &i) {
				return true
			}
			out, e = s.Diagnostics.Submit(r.Context(), a, parts[0], i)
			status = 201
		case parts[1] == "results" && r.Method == "GET":
			var data []domain.DiagnosticResult
			data, e = s.Diagnostics.Results(r.Context(), a, parts[0], page)
			out = map[string]any{"results": data}
		default:
			return false
		}
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
