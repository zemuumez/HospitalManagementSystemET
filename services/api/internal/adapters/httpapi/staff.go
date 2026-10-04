package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strings"
)

func (s Server) staff(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 && len(parts) != 4 {
		return false
	}
	if parts[0] != "v1" || parts[1] != "staff-profiles" {
		return false
	}
	id := parts[2]
	if len(parts) == 4 && parts[3] == "role" && r.Method == "PATCH" {
		var i domain.StaffRoleInput
		if !decode(w, r, &i) {
			return true
		}
		if e := s.Staff.ChangeRole(r.Context(), a, id, i); e != nil {
			fail(w, e)
		} else {
			write(w, 200, map[string]bool{"saved": true})
		}
		return true
	}
	if len(parts) != 3 {
		return false
	}
	switch r.Method {
	case "GET":
		out, e := s.Staff.Profile(r.Context(), a, id)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, out)
		}
	case "PATCH":
		var i domain.StaffProfileInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.Staff.Update(r.Context(), a, id, i)
		if e != nil {
			fail(w, e)
		} else {
			write(w, 200, out)
		}
	default:
		w.Header().Set("Allow", "GET, PATCH")
		write(w, 405, map[string]string{"code": "METHOD_NOT_ALLOWED", "error": "Method not allowed"})
	}
	return true
}
