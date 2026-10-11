package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
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

	// ─── Odontogram Charts ───────────────────────────────────────────────────────
	case r.URL.Path == "/v1/odontograms":
		if r.Method == "GET" {
			page := 1
			if raw := r.URL.Query().Get("page"); raw != "" {
				page, _ = strconv.Atoi(raw)
			}
			limit := 25
			if raw := r.URL.Query().Get("limit"); raw != "" {
				if val, err := strconv.Atoi(raw); err == nil && val > 0 && val <= 100 {
					limit = val
				}
			}
			charts, total, err := s.ClinicalCare.ListOdontogramCharts(r.Context(), a, r.URL.Query().Get("patientId"), r.URL.Query().Get("search"), page, limit)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{
				"charts": charts,
				"total":  total,
				"page":   page,
				"limit":  limit,
			})
			return true
		}
		if r.Method == "POST" {
			var input domain.SaveOdontogramChartInput
			if !decode(w, r, &input) {
				return true
			}
			chart, err := s.ClinicalCare.CreateOdontogramChart(r.Context(), a, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, chart)
			return true
		}

	case strings.HasPrefix(r.URL.Path, "/v1/odontograms/"):
		chartID := strings.TrimPrefix(r.URL.Path, "/v1/odontograms/")
		if strings.HasSuffix(chartID, "/pdf") {
			chartID = strings.TrimSuffix(chartID, "/pdf")
			if r.Method == "GET" {
				chart, err := s.ClinicalCare.GetOdontogramChart(r.Context(), a, chartID)
				if err != nil {
					fail(w, err)
					return true
				}
				pdfBytes := generateOdontogramPDF(chart)
				w.Header().Set("Content-Type", "application/pdf")
				w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"odontogram-%s.pdf\"", chart.ID))
				w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(pdfBytes)
				return true
			}
		} else {
			if r.Method == "GET" {
				chart, err := s.ClinicalCare.GetOdontogramChart(r.Context(), a, chartID)
				if err != nil {
					fail(w, err)
					return true
				}
				write(w, 200, chart)
				return true
			}
			if r.Method == "PUT" || r.Method == "PATCH" {
				var input domain.SaveOdontogramChartInput
				if !decode(w, r, &input) {
					return true
				}
				chart, err := s.ClinicalCare.UpdateOdontogramChart(r.Context(), a, chartID, input)
				if err != nil {
					if errors.Is(err, domain.ErrConflict) {
						write(w, 409, map[string]string{
							"error": "Concurrent edit conflict: chart version is stale. Refresh and try again.",
							"code":  "VERSION_CONFLICT",
						})
						return true
					}
					fail(w, err)
					return true
				}
				write(w, 200, chart)
				return true
			}
			if r.Method == "DELETE" {
				if err := s.ClinicalCare.DeleteOdontogramChart(r.Context(), a, chartID); err != nil {
					fail(w, err)
					return true
				}
				write(w, 200, map[string]bool{"deleted": true})
				return true
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
					bodyBytes, err := io.ReadAll(r.Body)
					if err != nil {
						fail(w, domain.ErrValidation)
						return true
					}
					// Support both chart aggregate payload and single tooth condition
					var chartInput domain.SaveOdontogramChartInput
					if err := json.Unmarshal(bodyBytes, &chartInput); err == nil && len(chartInput.Odontogram) > 0 {
						chartInput.PatientID = patientID
						chart, err := s.ClinicalCare.CreateOdontogramChart(r.Context(), a, chartInput)
						if err != nil {
							fail(w, err)
							return true
						}
						write(w, 201, chart)
						return true
					}

					var input domain.SetToothConditionInput
					if err := json.Unmarshal(bodyBytes, &input); err != nil {
						fail(w, domain.ErrValidation)
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

func escapePDFText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

func generateOdontogramPDF(chart domain.OdontogramChart) []byte {
	var stream bytes.Buffer
	stream.WriteString("BT\n")
	stream.WriteString("/F1 18 Tf\n")
	stream.WriteString("50 740 Td\n")
	stream.WriteString("(Hospital Odontogram Examination Report) Tj\n")
	stream.WriteString("/F1 11 Tf\n")
	stream.WriteString("0 -28 Td\n")
	stream.WriteString(fmt.Sprintf("(Patient: %s) Tj\n", escapePDFText(chart.PatientName)))
	stream.WriteString("0 -18 Td\n")
	stream.WriteString(fmt.Sprintf("(Patient MRN: %s) Tj\n", escapePDFText(chart.PatientMRN)))
	stream.WriteString("0 -18 Td\n")
	stream.WriteString(fmt.Sprintf("(Attending Doctor: %s) Tj\n", escapePDFText(chart.DoctorName)))
	stream.WriteString("0 -18 Td\n")
	stream.WriteString(fmt.Sprintf("(Chart ID: %s    Version: %d) Tj\n", escapePDFText(chart.ID), chart.Version))
	stream.WriteString("0 -18 Td\n")
	stream.WriteString(fmt.Sprintf("(Date: %s) Tj\n", escapePDFText(chart.CreatedAt.Format("2006-01-02 15:04:05"))))
	stream.WriteString("0 -24 Td\n")
	stream.WriteString(fmt.Sprintf("(Clinical Notes: %s) Tj\n", escapePDFText(chart.Description)))
	stream.WriteString("0 -28 Td\n")
	stream.WriteString("/F1 13 Tf\n")
	stream.WriteString("(Dental Condition Findings:) Tj\n")
	stream.WriteString("/F1 10 Tf\n")

	var toothNums []int
	for tNum := range chart.Odontogram {
		toothNums = append(toothNums, tNum)
	}
	sort.Ints(toothNums)

	if len(toothNums) == 0 {
		stream.WriteString("0 -18 Td\n")
		stream.WriteString("(All 32 teeth healthy / No active conditions charted) Tj\n")
	} else {
		for _, tNum := range toothNums {
			cond := chart.Odontogram[tNum]
			stream.WriteString("0 -16 Td\n")
			stream.WriteString(fmt.Sprintf("(  * Tooth #%d: Condition Code [%s]) Tj\n", tNum, escapePDFText(cond)))
		}
	}
	stream.WriteString("\nET\n")

	streamBytes := stream.Bytes()
	streamLen := len(streamBytes)

	var doc bytes.Buffer
	doc.WriteString("%PDF-1.4\n")

	offsets := make([]int, 6)

	// Object 1: Catalog
	offsets[1] = doc.Len()
	doc.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Object 2: Pages
	offsets[2] = doc.Len()
	doc.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// Object 3: Page
	offsets[3] = doc.Len()
	doc.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n")

	// Object 4: Stream Content
	offsets[4] = doc.Len()
	doc.WriteString(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n", streamLen))
	doc.Write(streamBytes)
	doc.WriteString("\nendstream\nendobj\n")

	// Object 5: Font
	offsets[5] = doc.Len()
	doc.WriteString("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	// XREF
	xrefOffset := doc.Len()
	doc.WriteString("xref\n0 6\n")
	doc.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		doc.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}

	doc.WriteString(fmt.Sprintf("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset))

	return doc.Bytes()
}
