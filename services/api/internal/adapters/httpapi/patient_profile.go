package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

func (s Server) patientProfile(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	if !strings.HasPrefix(r.URL.Path, "/v1/patients/") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/patients/"), "/")
	var out any
	var e error
	switch {
	case len(parts) == 1 && r.Method == "GET":
		out, e = s.App.PatientProfile(r.Context(), a, parts[0])
	case len(parts) == 2 && parts[1] == "profile" && r.Method == "PATCH":
		var i domain.PatientProfileInput
		if !decode(w, r, &i) {
			return true
		}
		out, e = s.App.UpdatePatientProfile(r.Context(), a, parts[0], i)
	case len(parts) == 2 && parts[1] == "revisions" && r.Method == "GET":
		page := 1
		if raw := r.URL.Query().Get("page"); raw != "" {
			page, _ = strconv.Atoi(raw)
		}
		var revisions []domain.ProfileRevision
		revisions, e = s.App.PatientRevisions(r.Context(), a, parts[0], page)
		out = map[string]any{"revisions": revisions}
	default:
		return false
	}
	if e != nil {
		fail(w, e)
	} else {
		write(w, 200, out)
	}
	return true
}
