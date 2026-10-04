package httpapi

import (
	"net/http"
	"strings"

	"hms.local/api/internal/domain"
)

func (s Server) clinicalCare(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	switch {
	// ─── Bed Occupancy & Bed Assignments ─────────────────────────────────────────
	case r.URL.Path == "/v1/bed-occupancy/report" && r.Method == "GET":
		rep, err := s.ClinicalCare.BedOccupancyReport(r.Context(), a)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, rep)
		return true

	case r.URL.Path == "/v1/bed-assignments" && r.Method == "POST":
		var input domain.AssignBedInput
		if !decode(w, r, &input) {
			return true
		}
		assign, err := s.ClinicalCare.AssignBed(r.Context(), a, input)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 201, assign)
		return true

	// ─── Care Team Members ───────────────────────────────────────────────────────
	case strings.HasPrefix(r.URL.Path, "/v1/care-team/") && r.Method == "DELETE":
		memberID := strings.TrimPrefix(r.URL.Path, "/v1/care-team/")
		if err := s.ClinicalCare.RevokeCareTeamMember(r.Context(), a, memberID); err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]bool{"ok": true})
		return true

	// ─── Encounter Sub-resources ─────────────────────────────────────────────────
	case strings.HasPrefix(r.URL.Path, "/v1/encounters/"):
		rest := strings.TrimPrefix(r.URL.Path, "/v1/encounters/")
		parts := strings.Split(rest, "/")
		if len(parts) >= 2 {
			encounterID := parts[0]
			sub := parts[1]

			switch sub {
			case "bed-assignments":
				if r.Method == "GET" {
					list, err := s.ClinicalCare.ListBedAssignments(r.Context(), a, encounterID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]any{"assignments": list})
					return true
				}

			case "care-team":
				if r.Method == "GET" {
					team, err := s.ClinicalCare.CareTeam(r.Context(), a, encounterID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]any{"careTeam": team})
					return true
				}
				if r.Method == "POST" {
					var input domain.AddCareTeamMemberInput
					if !decode(w, r, &input) {
						return true
					}
					m, err := s.ClinicalCare.AddCareTeamMember(r.Context(), a, encounterID, input)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 201, m)
					return true
				}

			case "diagnoses":
				if r.Method == "GET" {
					diag, err := s.ClinicalCare.Diagnoses(r.Context(), a, encounterID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]any{"diagnoses": diag})
					return true
				}
				if r.Method == "POST" {
					var input domain.AddDiagnosisInput
					if !decode(w, r, &input) {
						return true
					}
					d, err := s.ClinicalCare.AddDiagnosis(r.Context(), a, encounterID, input)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 201, d)
					return true
				}

			case "procedures":
				if r.Method == "GET" {
					procs, err := s.ClinicalCare.Procedures(r.Context(), a, encounterID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]any{"procedures": procs})
					return true
				}
				if r.Method == "POST" {
					var input domain.AddProcedureInput
					if !decode(w, r, &input) {
						return true
					}
					p, err := s.ClinicalCare.AddProcedure(r.Context(), a, encounterID, input)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 201, p)
					return true
				}

			case "attachments":
				if r.Method == "GET" {
					atts, err := s.ClinicalCare.Attachments(r.Context(), a, encounterID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]any{"attachments": atts})
					return true
				}
				if r.Method == "POST" {
					var input domain.AddAttachmentInput
					if !decode(w, r, &input) {
						return true
					}
					att, err := s.ClinicalCare.AddAttachment(r.Context(), a, encounterID, input)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 201, att)
					return true
				}

			case "admission-details":
				if r.Method == "GET" {
					details, err := s.ClinicalCare.IPDAdmissionDetails(r.Context(), a, encounterID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, details)
					return true
				}
				if r.Method == "PUT" || r.Method == "POST" {
					var details domain.IPDAdmissionDetails
					if !decode(w, r, &details) {
						return true
					}
					details.EncounterID = encounterID
					if err := s.ClinicalCare.SaveIPDAdmissionDetails(r.Context(), a, details); err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]bool{"ok": true})
					return true
				}

			case "billing":
				if len(parts) == 2 && r.Method == "GET" {
					b, err := s.ClinicalCare.EncounterBilling(r.Context(), a, encounterID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, b)
					return true
				}
				if len(parts) == 2 && (r.Method == "PUT" || r.Method == "POST") {
					var input domain.UpdateEncounterBillingInput
					if !decode(w, r, &input) {
						return true
					}
					b, err := s.ClinicalCare.UpdateEncounterBilling(r.Context(), a, encounterID, input)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, b)
					return true
				}
				if len(parts) == 3 && parts[2] == "clearance" && r.Method == "POST" {
					var body struct {
						WaiverReason string `json:"waiverReason"`
					}
					if !decode(w, r, &body) {
						return true
					}
					b, err := s.ClinicalCare.GrantFinancialClearance(r.Context(), a, encounterID, body.WaiverReason)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, b)
					return true
				}
				if len(parts) == 3 && parts[2] == "invoice" && r.Method == "POST" {
					var body struct {
						InvoiceID string `json:"invoiceId"`
					}
					if !decode(w, r, &body) {
						return true
					}
					b, err := s.ClinicalCare.LinkEncounterInvoice(r.Context(), a, encounterID, body.InvoiceID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, b)
					return true
				}

			case "discharge-summary":
				if r.Method == "GET" {
					ds, err := s.ClinicalCare.DischargeSummary(r.Context(), a, encounterID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, ds)
					return true
				}
				if r.Method == "PUT" || r.Method == "POST" {
					var ds domain.DischargeSummary
					if !decode(w, r, &ds) {
						return true
					}
					ds.EncounterID = encounterID
					if err := s.ClinicalCare.SaveDischargeSummary(r.Context(), a, ds); err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]bool{"ok": true})
					return true
				}
			}
		}

	// ─── Patient OPD Follow-ups, Referrals, Odontogram ───────────────────────────
	case strings.HasPrefix(r.URL.Path, "/v1/patients/"):
		rest := strings.TrimPrefix(r.URL.Path, "/v1/patients/")
		parts := strings.Split(rest, "/")
		if len(parts) >= 2 {
			patientID := parts[0]
			sub := parts[1]

			switch sub {
			case "follow-ups":
				if r.Method == "GET" {
					list, err := s.ClinicalCare.OPDFollowUps(r.Context(), a, patientID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]any{"followUps": list})
					return true
				}
				if r.Method == "POST" {
					var input domain.CreateOPDFollowUpInput
					if !decode(w, r, &input) {
						return true
					}
					input.PatientID = patientID
					f, err := s.ClinicalCare.CreateOPDFollowUp(r.Context(), a, input)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 201, f)
					return true
				}

			case "referrals":
				if r.Method == "GET" {
					list, err := s.ClinicalCare.PatientReferrals(r.Context(), a, patientID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]any{"referrals": list})
					return true
				}
				if r.Method == "POST" {
					var input domain.CreatePatientReferralInput
					if !decode(w, r, &input) {
						return true
					}
					input.PatientID = patientID
					ref, err := s.ClinicalCare.CreatePatientReferral(r.Context(), a, input)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 201, ref)
					return true
				}

			case "odontogram":
				if r.Method == "GET" {
					list, err := s.ClinicalCare.Odontogram(r.Context(), a, patientID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]any{"teeth": list})
					return true
				}
				if r.Method == "PUT" || r.Method == "POST" {
					var input domain.SetToothConditionInput
					if !decode(w, r, &input) {
						return true
					}
					e, err := s.ClinicalCare.SetToothCondition(r.Context(), a, patientID, input)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, e)
					return true
				}
			}
		}
	}

	return false
}
