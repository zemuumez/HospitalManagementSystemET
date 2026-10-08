package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"hms.local/api/internal/domain"
)

func (s Server) insurances(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	// Export endpoint
	if r.URL.Path == "/v1/insurances-export" && r.Method == "GET" {
		search := r.URL.Query().Get("search")
		allowTruncated := r.URL.Query().Get("allow_truncated") == "true" || r.URL.Query().Get("truncate") == "1"
		list, total, truncated, err := s.Insurances.ExportInsurancesBounded(r.Context(), a, search, allowTruncated)
		if err != nil {
			if errors.Is(err, domain.ErrExportLimitExceeded) {
				write(w, 422, map[string]any{
					"error":     "Export exceeds maximum limit of 5,000 records. Please refine search filters.",
					"code":      "EXPORT_LIMIT_EXCEEDED",
					"total":     total,
					"max_limit": 5000,
				})
				return true
			}
			fail(w, err)
			return true
		}
		res := map[string]any{
			"insurances": list,
			"total":      total,
			"truncated":  truncated,
			"max_limit":  5000,
		}
		if truncated {
			res["warning"] = "Export truncated to 5,000 records. Please refine search filters."
		}
		write(w, 200, res)
		return true
	}

	if !strings.HasPrefix(r.URL.Path, "/v1/insurances") {
		return false
	}

	switch {
	case r.URL.Path == "/v1/insurances" && r.Method == "GET":
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 25
		}
		search := r.URL.Query().Get("search")
		list, total, err := s.Insurances.Insurances(r.Context(), a, page, limit, search)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"insurances": list, "total": total, "page": page, "limit": limit})
		return true

	case r.URL.Path == "/v1/insurances" && r.Method == "POST":
		var in domain.InsuranceInput
		if !decode(w, r, &in) {
			return true
		}
		ins, err := s.Insurances.CreateInsurance(r.Context(), a, in)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, ins)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/insurances/") && strings.HasSuffix(r.URL.Path, "/status") && r.Method == "PATCH":
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/insurances/"), "/status")
		var body struct {
			Status *int `json:"status"`
		}
		if !decode(w, r, &body) {
			return true
		}
		if body.Status == nil || (*body.Status != 0 && *body.Status != 1) {
			fail(w, domain.ErrValidation)
			return true
		}
		ins, err := s.Insurances.SetStatus(r.Context(), a, id, *body.Status)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, ins)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/insurances/") && r.Method == "GET":
		id := strings.TrimPrefix(r.URL.Path, "/v1/insurances/")
		ins, err := s.Insurances.Insurance(r.Context(), a, id)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, ins)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/insurances/") && (r.Method == "PUT" || r.Method == "PATCH"):
		id := strings.TrimPrefix(r.URL.Path, "/v1/insurances/")
		var in domain.InsuranceInput
		if !decode(w, r, &in) {
			return true
		}
		ins, err := s.Insurances.UpdateInsurance(r.Context(), a, id, in)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, ins)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/insurances/") && r.Method == "DELETE":
		id := strings.TrimPrefix(r.URL.Path, "/v1/insurances/")
		if err := s.Insurances.DeleteInsurance(r.Context(), a, id); err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]bool{"deleted": true})
		return true
	}

	return false
}
