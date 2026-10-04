package httpapi

import (
	"net/http"
	"strings"

	"hms.local/api/internal/domain"
)

func (s Server) publicPatientExtensions(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path == "/v1/smart-cards/verify" && r.Method == "GET" {
		token := r.URL.Query().Get("token")
		card, err := s.PatientExt.VerifySmartCardQR(r.Context(), token)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, card)
		return true
	}
	return false
}

func (s Server) patientExtensions(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	switch {
	case r.URL.Path == "/v1/patients/duplicates" && r.Method == "GET":
		q := r.URL.Query().Get("q")
		dupes, err := s.PatientExt.FindDuplicates(r.Context(), a, q)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"duplicates": dupes})
		return true

	case r.URL.Path == "/v1/patients/merge" && r.Method == "POST":
		var input domain.MergePatientInput
		if !decode(w, r, &input) {
			return true
		}
		if err := s.PatientExt.MergePatients(r.Context(), a, input); err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]bool{"ok": true})
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/smart-cards/") && strings.HasSuffix(r.URL.Path, "/revoke") && r.Method == "POST":
		cardID := strings.TrimPrefix(r.URL.Path, "/v1/smart-cards/")
		cardID = strings.TrimSuffix(cardID, "/revoke")
		var body struct {
			Reason string `json:"reason"`
		}
		if !decode(w, r, &body) {
			return true
		}
		if err := s.PatientExt.RevokeSmartCard(r.Context(), a, cardID, body.Reason); err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]bool{"ok": true})
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/patients/") && strings.HasSuffix(r.URL.Path, "/consent"):
		patientID := strings.TrimPrefix(r.URL.Path, "/v1/patients/")
		patientID = strings.TrimSuffix(patientID, "/consent")
		if r.Method == "GET" {
			c, err := s.PatientExt.GetConsent(r.Context(), a, patientID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, c)
			return true
		}
		if r.Method == "PUT" || r.Method == "POST" {
			var c domain.PatientContactConsent
			if !decode(w, r, &c) {
				return true
			}
			c.PatientID = patientID
			if err := s.PatientExt.SaveConsent(r.Context(), a, c); err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]bool{"ok": true})
			return true
		}

	case strings.HasPrefix(r.URL.Path, "/v1/patients/") && strings.HasSuffix(r.URL.Path, "/smart-cards"):
		patientID := strings.TrimPrefix(r.URL.Path, "/v1/patients/")
		patientID = strings.TrimSuffix(patientID, "/smart-cards")
		if r.Method == "GET" {
			cards, err := s.PatientExt.ListSmartCards(r.Context(), a, patientID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"smartCards": cards})
			return true
		}
		if r.Method == "POST" {
			var input domain.IssueSmartCardInput
			if !decode(w, r, &input) {
				return true
			}
			input.PatientID = patientID
			card, err := s.PatientExt.IssueSmartCard(r.Context(), a, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, card)
			return true
		}
	}

	return false
}
