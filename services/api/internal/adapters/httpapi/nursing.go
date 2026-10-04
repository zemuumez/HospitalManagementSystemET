package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) nursing(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	page := 1
	if raw := r.URL.Query().Get("page"); raw != "" {
		page, _ = strconv.Atoi(raw)
	}
	if r.URL.Path == "/v1/nursing-encounters" && r.Method == "GET" {
		out, e := s.Clinical.NursingEncounters(r.Context(), a, page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"encounters": out, "page": page, "pageSize": 25})
		}
		return true
	}
	if !strings.HasPrefix(r.URL.Path, "/v1/encounters/") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/encounters/"), "/")
	if len(parts) != 2 {
		return false
	}
	id := parts[0]
	switch {
	case parts[1] == "nurses" && r.Method == "GET":
		out, e := s.Clinical.Nurses(r.Context(), a, id)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"nurses": out})
		}
	case parts[1] == "nurses" && r.Method == "POST":
		var i domain.NurseAssignment
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Clinical.AssignNurse(r.Context(), a, id, i)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, out)
		}
	case parts[1] == "vitals" && r.Method == "GET":
		out, e := s.Clinical.Vitals(r.Context(), a, id, page)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]any{"vitals": out, "page": page, "pageSize": 25})
		}
	case parts[1] == "vitals" && r.Method == "POST":
		var i domain.VitalsInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Clinical.RecordVitals(r.Context(), a, id, i, r.Header.Get("Idempotency-Key"))
		if e != nil {
			fail(w, e)
		} else {
			write(w, 201, out)
		}
	default:
		return false
	}
	return true
}
