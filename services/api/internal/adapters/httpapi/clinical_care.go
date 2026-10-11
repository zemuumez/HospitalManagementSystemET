package httpapi

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"hms.local/api/internal/domain"
)

//go:embed fonts/ebrima.ttf
var ebrimaFontData []byte

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

	// ─── Odontogram Legends ───────────────────────────────────────────────────────
	case r.URL.Path == "/v1/odontograms/legends":
		if r.Method == "GET" {
			legends, err := s.ClinicalCare.ListOdontogramLegends(r.Context())
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"legends": legends})
			return true
		}
		if r.Method == "POST" {
			var input domain.SaveOdontogramLegendInput
			if !decode(w, r, &input) {
				return true
			}
			legend, err := s.ClinicalCare.CreateOdontogramLegend(r.Context(), a, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, legend)
			return true
		}

	case strings.HasPrefix(r.URL.Path, "/v1/odontograms/legends/"):
		legendID := strings.TrimPrefix(r.URL.Path, "/v1/odontograms/legends/")
		if r.Method == "PUT" || r.Method == "PATCH" {
			var input domain.SaveOdontogramLegendInput
			if !decode(w, r, &input) {
				return true
			}
			legend, err := s.ClinicalCare.UpdateOdontogramLegend(r.Context(), a, legendID, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, legend)
			return true
		}
		if r.Method == "DELETE" {
			if err := s.ClinicalCare.DeleteOdontogramLegend(r.Context(), a, legendID); err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]bool{"deleted": true})
			return true
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
		if strings.HasSuffix(chartID, "/pdf") || strings.HasSuffix(chartID, "/export.pdf") {
			chartID = strings.TrimSuffix(strings.TrimSuffix(chartID, "/export.pdf"), "/pdf")
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

type ttfCMap struct {
	segCount      int
	endCode       []uint16
	startCode     []uint16
	idDelta       []int16
	idRangeOffset []uint16
	glyphIdArray  []byte
}

func parseFormat4CMap(buf []byte) (*ttfCMap, error) {
	if len(buf) < 12 {
		return nil, fmt.Errorf("ttf data too short")
	}
	numTables := int(binary.BigEndian.Uint16(buf[4:6]))
	var cmapOffset int
	for i := 0; i < numTables; i++ {
		off := 12 + i*16
		if off+16 > len(buf) {
			break
		}
		if string(buf[off:off+4]) == "cmap" {
			cmapOffset = int(binary.BigEndian.Uint32(buf[off+8 : off+12]))
			break
		}
	}
	if cmapOffset == 0 || cmapOffset+4 > len(buf) {
		return nil, fmt.Errorf("cmap table not found")
	}

	numSubtables := int(binary.BigEndian.Uint16(buf[cmapOffset+2 : cmapOffset+4]))
	for i := 0; i < numSubtables; i++ {
		subHeader := cmapOffset + 4 + i*8
		if subHeader+8 > len(buf) {
			break
		}
		platformID := binary.BigEndian.Uint16(buf[subHeader : subHeader+2])
		subOffset := cmapOffset + int(binary.BigEndian.Uint32(buf[subHeader+4:subHeader+8]))
		if subOffset+6 > len(buf) {
			continue
		}
		format := binary.BigEndian.Uint16(buf[subOffset : subOffset+2])
		if (platformID == 3 || platformID == 0) && format == 4 {
			length := int(binary.BigEndian.Uint16(buf[subOffset+2 : subOffset+4]))
			if subOffset+length > len(buf) {
				continue
			}
			segCount := int(binary.BigEndian.Uint16(buf[subOffset+6:subOffset+8])) / 2
			endCountOffset := subOffset + 14
			startCountOffset := endCountOffset + segCount*2 + 2
			idDeltaOffset := startCountOffset + segCount*2
			idRangeOffset := idDeltaOffset + segCount*2

			cmap := &ttfCMap{
				segCount:      segCount,
				endCode:       make([]uint16, segCount),
				startCode:     make([]uint16, segCount),
				idDelta:       make([]int16, segCount),
				idRangeOffset: make([]uint16, segCount),
			}
			for s := 0; s < segCount; s++ {
				cmap.endCode[s] = binary.BigEndian.Uint16(buf[endCountOffset+s*2 : endCountOffset+s*2+2])
				cmap.startCode[s] = binary.BigEndian.Uint16(buf[startCountOffset+s*2 : startCountOffset+s*2+2])
				cmap.idDelta[s] = int16(binary.BigEndian.Uint16(buf[idDeltaOffset+s*2 : idDeltaOffset+s*2+2]))
				cmap.idRangeOffset[s] = binary.BigEndian.Uint16(buf[idRangeOffset+s*2 : idRangeOffset+s*2+2])
			}
			cmap.glyphIdArray = buf[idRangeOffset:]
			return cmap, nil
		}
	}
	return nil, fmt.Errorf("format 4 cmap not found")
}

func (c *ttfCMap) lookup(r rune) uint16 {
	if r > 0xFFFF {
		return 0
	}
	code := uint16(r)
	for s := 0; s < c.segCount; s++ {
		if code <= c.endCode[s] {
			if code < c.startCode[s] {
				return 0
			}
			if c.idRangeOffset[s] == 0 {
				return uint16((int(code) + int(c.idDelta[s])) & 0xFFFF)
			}
			offset := int(c.idRangeOffset[s]) + (s * 2) + int(code-c.startCode[s])*2
			if offset+2 <= len(c.glyphIdArray) {
				return binary.BigEndian.Uint16(c.glyphIdArray[offset : offset+2])
			}
			return 0
		}
	}
	return 0
}

var (
	ebrimaOnce sync.Once
	ebrimaCMap *ttfCMap
)

func getEbrimaCMap() *ttfCMap {
	ebrimaOnce.Do(func() {
		if len(ebrimaFontData) > 0 {
			cmap, _ := parseFormat4CMap(ebrimaFontData)
			ebrimaCMap = cmap
		}
	})
	return ebrimaCMap
}

func wrapPDFText(text string, maxRunes int) []string {
	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
	var current strings.Builder
	currentLen := 0
	for _, w := range words {
		wLen := utf8.RuneCountInString(w)
		if currentLen == 0 {
			current.WriteString(w)
			currentLen = wLen
		} else if currentLen+1+wLen <= maxRunes {
			current.WriteString(" ")
			current.WriteString(w)
			currentLen += 1 + wLen
		} else {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(w)
			currentLen = wLen
		}
	}
	if currentLen > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

func generateOdontogramPDF(chart domain.OdontogramChart) []byte {
	cmap := getEbrimaCMap()
	if cmap == nil || len(ebrimaFontData) == 0 {
		return []byte("%PDF-1.4\n% Odontogram Report\n")
	}

	type pdfLine struct {
		x, y int
		size int
		text string
	}

	var pages [][]pdfLine
	var currentPage []pdfLine
	curY := 740

	pushLine := func(x int, size int, text string) {
		if curY < 70 {
			pages = append(pages, currentPage)
			currentPage = nil
			curY = 740
		}
		currentPage = append(currentPage, pdfLine{x: x, y: curY, size: size, text: text})
		curY -= (size + 5)
	}

	// Header
	pushLine(50, 16, "Hospital Odontogram Examination Report / የጥርስ ምርመራ ሪፖርት")
	patName := chart.PatientName
	if patName == "" {
		patName = "Patient " + chart.PatientID
	}
	mrn := chart.PatientMRN
	if mrn == "" {
		mrn = "N/A"
	}
	pushLine(50, 10, fmt.Sprintf("Patient / ታካሚ: %s  |  MRN: %s", patName, mrn))
	docName := chart.DoctorName
	if docName == "" {
		docName = "Doctor"
	}
	pushLine(50, 10, fmt.Sprintf("Attending Doctor: %s  |  Date: %s", docName, chart.CreatedAt.Format("2006-01-02 15:04:05")))
	pushLine(50, 10, fmt.Sprintf("Chart ID: %s  |  Version: %d", chart.ID, chart.Version))
	curY -= 8

	pushLine(50, 12, "Clinical Notes / ክሊኒካዊ ማስታወሻ:")
	desc := chart.Description
	if desc == "" {
		desc = "Dental Examination Chart"
	}
	wrappedNotes := wrapPDFText(desc, 65)
	for _, l := range wrappedNotes {
		pushLine(65, 10, l)
	}
	curY -= 8

	pushLine(50, 12, "Dental Condition Findings:")

	var toothNums []int
	for tNum := range chart.Odontogram {
		toothNums = append(toothNums, tNum)
	}
	sort.Ints(toothNums)

	if len(toothNums) == 0 {
		pushLine(65, 10, "All teeth healthy / No active conditions charted")
	} else {
		for _, tNum := range toothNums {
			cond := chart.Odontogram[tNum]
			pushLine(65, 10, fmt.Sprintf("Tooth #%d: Condition Code [%s]", tNum, cond))
		}
	}

	if len(currentPage) > 0 {
		pages = append(pages, currentPage)
	}
	numPages := len(pages)
	if numPages == 0 {
		numPages = 1
		pages = [][]pdfLine{{{x: 50, y: 740, size: 12, text: "Hospital Odontogram Report"}}}
	}

	usedGIDs := make(map[uint16]rune)
	usedGIDs[cmap.lookup(' ')] = ' '

	encodeHex := func(s string) string {
		var b bytes.Buffer
		for _, r := range s {
			gid := cmap.lookup(r)
			usedGIDs[gid] = r
			fmt.Fprintf(&b, "%04X", gid)
		}
		return b.String()
	}

	contentBytes := make([][]byte, numPages)
	for pIdx, p := range pages {
		var st bytes.Buffer
		for _, line := range p {
			hexStr := encodeHex(line.text)
			st.WriteString("BT\n")
			st.WriteString(fmt.Sprintf("/F1 %d Tf\n", line.size))
			st.WriteString(fmt.Sprintf("%d %d Td\n", line.x, line.y))
			st.WriteString(fmt.Sprintf("<%s> Tj\n", hexStr))
			st.WriteString("ET\n")
		}
		// Footer
		st.WriteString("BT\n/F1 9 Tf\n50 40 Td\n")
		footerHex := encodeHex(fmt.Sprintf("Page %d of %d  •  Official Hospital Record  •  HospitalManagementSystemET", pIdx+1, numPages))
		st.WriteString(fmt.Sprintf("<%s> Tj\nET\n", footerHex))
		contentBytes[pIdx] = st.Bytes()
	}

	var doc bytes.Buffer
	doc.WriteString("%PDF-1.4\n")

	catalogObj := 1
	pagesObj := 2
	fontType0Obj := 3 + 2*numPages
	cidFontObj := fontType0Obj + 1
	fontDescObj := fontType0Obj + 2
	fontFileObj := fontType0Obj + 3
	toUnicodeObj := fontType0Obj + 4
	totalObjs := toUnicodeObj

	offsets := make([]int, totalObjs+1)

	// 1. Catalog
	offsets[catalogObj] = doc.Len()
	doc.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Catalog /Pages %d 0 R >>\nendobj\n", catalogObj, pagesObj))

	// 2. Pages
	offsets[pagesObj] = doc.Len()
	var kids bytes.Buffer
	for i := 0; i < numPages; i++ {
		fmt.Fprintf(&kids, "%d 0 R ", 3+i)
	}
	doc.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Pages /Kids [%s] /Count %d >>\nendobj\n", pagesObj, kids.String(), numPages))

	// Page objects: 3 to 2+numPages
	for i := 0; i < numPages; i++ {
		pageID := 3 + i
		contentID := 3 + numPages + i
		offsets[pageID] = doc.Len()
		doc.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Page /Parent %d 0 R /MediaBox [0 0 612 792] /Contents %d 0 R /Resources << /Font << /F1 %d 0 R >> >> >>\nendobj\n", pageID, pagesObj, contentID, fontType0Obj))
	}

	// Content streams: 3+numPages to 2+2*numPages
	for i := 0; i < numPages; i++ {
		contentID := 3 + numPages + i
		offsets[contentID] = doc.Len()
		doc.WriteString(fmt.Sprintf("%d 0 obj\n<< /Length %d >>\nstream\n", contentID, len(contentBytes[i])))
		doc.Write(contentBytes[i])
		doc.WriteString("\nendstream\nendobj\n")
	}

	// Font Type0
	offsets[fontType0Obj] = doc.Len()
	doc.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Font /Subtype /Type0 /BaseFont /Ebrima /Encoding /Identity-H /DescendantFonts [%d 0 R] /ToUnicode %d 0 R >>\nendobj\n", fontType0Obj, cidFontObj, toUnicodeObj))

	// CIDFont
	offsets[cidFontObj] = doc.Len()
	doc.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Font /Subtype /CIDFontType2 /BaseFont /Ebrima /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /FontDescriptor %d 0 R /CIDToGIDMap /Identity >>\nendobj\n", cidFontObj, fontDescObj))

	// FontDescriptor
	offsets[fontDescObj] = doc.Len()
	doc.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /FontDescriptor /FontName /Ebrima /Flags 32 /FontBBox [-500 -300 1200 1000] /ItalicAngle 0 /Ascent 800 /Descent -200 /CapHeight 700 /StemV 80 /FontFile2 %d 0 R >>\nendobj\n", fontDescObj, fontFileObj))

	// FontFile2 stream (embedded TTF)
	offsets[fontFileObj] = doc.Len()
	doc.WriteString(fmt.Sprintf("%d 0 obj\n<< /Length %d /Length1 %d >>\nstream\n", fontFileObj, len(ebrimaFontData), len(ebrimaFontData)))
	doc.Write(ebrimaFontData)
	doc.WriteString("\nendstream\nendobj\n")

	// ToUnicode CMap
	var cmapBuf bytes.Buffer
	cmapBuf.WriteString("/CIDInit /ProcSet findresource begin\n12 dict begin\nbegincmap\n/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n/CMapName /Custom-ToUnicode def\n/CMapType 2 def\n1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n")

	var sortedGIDs []int
	for gid := range usedGIDs {
		sortedGIDs = append(sortedGIDs, int(gid))
	}
	sort.Ints(sortedGIDs)

	for i := 0; i < len(sortedGIDs); i += 100 {
		end := i + 100
		if end > len(sortedGIDs) {
			end = len(sortedGIDs)
		}
		fmt.Fprintf(&cmapBuf, "%d beginbfchar\n", end-i)
		for _, gid := range sortedGIDs[i:end] {
			r := usedGIDs[uint16(gid)]
			fmt.Fprintf(&cmapBuf, "<%04X> <%04X>\n", gid, r)
		}
		cmapBuf.WriteString("endbfchar\n")
	}
	cmapBuf.WriteString("endcmap\nCMapName currentdict /CMap defineresource pop\nend\nend\n")
	cmapBytes := cmapBuf.Bytes()

	offsets[toUnicodeObj] = doc.Len()
	doc.WriteString(fmt.Sprintf("%d 0 obj\n<< /Length %d >>\nstream\n", toUnicodeObj, len(cmapBytes)))
	doc.Write(cmapBytes)
	doc.WriteString("\nendstream\nendobj\n")

	// XREF Table
	xrefOffset := doc.Len()
	doc.WriteString(fmt.Sprintf("xref\n0 %d\n", totalObjs+1))
	doc.WriteString("0000000000 65535 f \n")
	for i := 1; i <= totalObjs; i++ {
		doc.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}

	doc.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root %d 0 R >>\nstartxref\n%d\n%%%%EOF\n", totalObjs+1, catalogObj, xrefOffset))

	return doc.Bytes()
}

func GenerateOdontogramPDF(chart domain.OdontogramChart) []byte {
	return generateOdontogramPDF(chart)
}
