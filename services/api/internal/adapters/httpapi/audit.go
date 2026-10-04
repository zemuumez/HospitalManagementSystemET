package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
)

func (s Server) audit(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	if r.URL.Path != "/v1/audit-events" {
		return false
	}
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		write(w, 405, map[string]string{"code": "METHOD_NOT_ALLOWED", "error": "Method not allowed"})
		return true
	}
	f := domain.AuditFilter{ActorID: r.URL.Query().Get("actorId"), Action: r.URL.Query().Get("action"), ResourceID: r.URL.Query().Get("resourceId")}
	if raw := r.URL.Query().Get("before"); raw != "" {
		var e error
		f.Before, e = strconv.ParseInt(raw, 10, 64)
		if e != nil {
			fail(w, domain.ErrValidation)
			return true
		}
	}
	out, e := s.Audit.Events(r.Context(), a, f)
	if e != nil {
		fail(w, e)
		return true
	}
	var next int64
	if len(out) == 50 {
		next = out[len(out)-1].ID
	}
	write(w, 200, map[string]any{"events": out, "nextBefore": next, "pageSize": 50})
	return true
}
