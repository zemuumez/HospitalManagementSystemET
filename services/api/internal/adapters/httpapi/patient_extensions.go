package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"hms.local/api/internal/domain"
)

func (s Server) publicPatientExtensions(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path == "/v1/smart-cards/verify" && r.Method == "GET" {
		token := r.URL.Query().Get("token")
		res, err := s.PatientExt.VerifySmartCardQR(r.Context(), token)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, res)
		return true
	}
	return false
}

func (s Server) patientExtensions(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	switch {
	// ─── ID Card Templates ────────────────────────────────────────────────────────
	case r.URL.Path == "/v1/smart-cards/templates" && r.Method == "GET":
		list, err := s.PatientExt.ListTemplates(r.Context(), a)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"templates": list})
		return true

	case r.URL.Path == "/v1/smart-cards/templates" && r.Method == "POST":
		var input domain.CreatePatientIDCardTemplateInput
		if !decode(w, r, &input) {
			return true
		}
		tmpl, err := s.PatientExt.CreateTemplate(r.Context(), a, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, map[string]any{"template": tmpl})
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/smart-cards/templates/"):
		tmplID := strings.TrimPrefix(r.URL.Path, "/v1/smart-cards/templates/")
		if r.Method == "GET" {
			tmpl, err := s.PatientExt.GetTemplate(r.Context(), a, tmplID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"template": tmpl})
			return true
		}
		if r.Method == "PUT" {
			var input domain.UpdatePatientIDCardTemplateInput
			if !decode(w, r, &input) {
				return true
			}
			tmpl, err := s.PatientExt.UpdateTemplate(r.Context(), a, tmplID, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"template": tmpl})
			return true
		}
		if r.Method == "DELETE" {
			if err := s.PatientExt.DeleteTemplate(r.Context(), a, tmplID); err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]bool{"ok": true})
			return true
		}

	// ─── Smart Cards Generation & Management ─────────────────────────────────────
	case r.URL.Path == "/v1/smart-cards/generate" && r.Method == "POST":
		var input domain.GenerateSmartCardsInput
		if !decode(w, r, &input) {
			return true
		}
		res, err := s.PatientExt.GenerateSmartCards(r.Context(), a, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, res)
		return true

	case r.URL.Path == "/v1/smart-cards/cards" && r.Method == "GET":
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
		filter := domain.ListIssuedCardsFilter{
			Search:     r.URL.Query().Get("search"),
			TemplateID: r.URL.Query().Get("templateId"),
			Page:       page,
			PageSize:   pageSize,
		}
		res, err := s.PatientExt.ListIssuedCards(r.Context(), a, filter)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, res)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/smart-cards/cards/") && r.Method == "DELETE":
		cardID := strings.TrimPrefix(r.URL.Path, "/v1/smart-cards/cards/")
		if err := s.PatientExt.DeleteCardAssignment(r.Context(), a, cardID); err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]bool{"ok": true})
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/patients/") && strings.HasSuffix(r.URL.Path, "/identities") && r.Method == "GET":
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/patients/"), "/identities")
		items, err := s.PatientExt.Identities(r.Context(), a, id)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"identities": items})
		return true

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
