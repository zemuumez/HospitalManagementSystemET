package httpapi

import (
	"net/http"
	"strings"

	"hms.local/api/internal/domain"
)

func (s Server) attachments(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	switch {
	case r.URL.Path == "/v1/attachments" && r.Method == "POST":
		var input domain.CreateSecureAttachmentInput
		if !decode(w, r, &input) {
			return true
		}
		att, err := s.Attachments.CreateAttachment(r.Context(), a, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, http.StatusCreated, att)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/attachments/") && r.Method == "GET":
		token := strings.TrimPrefix(r.URL.Path, "/v1/attachments/")
		att, err := s.Attachments.GetAttachmentByToken(r.Context(), a, token)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, http.StatusOK, att)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/patients/") && strings.HasSuffix(r.URL.Path, "/attachments") && r.Method == "GET":
		patientID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/patients/"), "/attachments")
		list, err := s.Attachments.ListPatientAttachments(r.Context(), a, patientID)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, http.StatusOK, list)
		return true

	default:
		return false
	}
}
