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

	case r.URL.Path == "/v1/bed-assignments" && r.Method == "GET":
		page := 1
		if raw := r.URL.Query().Get("page"); raw != "" {
			page, _ = strconv.Atoi(raw)
		}
		pageSize := 25
		if raw := r.URL.Query().Get("pageSize"); raw != "" {
			if ps, err := strconv.Atoi(raw); err == nil && ps > 0 {
				pageSize = ps
			}
		}
		search := r.URL.Query().Get("search")
		status := r.URL.Query().Get("status")
		list, total, err := s.ClinicalCare.ListAllBedAssignments(r.Context(), a, search, status, page, pageSize)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"assignments": list, "total": total, "page": page, "pageSize": pageSize})
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

	case strings.HasPrefix(r.URL.Path, "/v1/bed-assignments/") && strings.HasSuffix(r.URL.Path, "/discharge") && r.Method == "POST":
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/bed-assignments/"), "/discharge")
		var input domain.DischargeBedAssignmentInput
		_ = json.NewDecoder(io.LimitReader(r.Body, 16<<10)).Decode(&input)
		assign, err := s.ClinicalCare.DischargeBedAssignment(r.Context(), a, id, input.DischargeNotes)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, assign)
		return true

	case r.URL.Path == "/v1/bed-transfers" && r.Method == "POST":
		var input struct {
			EncounterID string `json:"encounterId"`
			PatientID   string `json:"patientId"`
			BedID       string `json:"bedId"`
			Notes       string `json:"notes"`
			Version     int    `json:"version"`
		}
		if !decode(w, r, &input) {
			return true
		}
		assign, err := s.ClinicalCare.AssignBed(r.Context(), a, domain.AssignBedInput{
			EncounterID: input.EncounterID,
			PatientID:   input.PatientID,
			BedID:       input.BedID,
			Notes:       input.Notes,
			Version:     input.Version,
		})
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, assign)
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

type ttfFontInfo struct {
	cmap        *ttfCMap
	unitsPerEm  uint16
	numHMetrics uint16
	hmtxData    []byte
}

func parseFormat4CMapFromOffset(buf []byte, cmapOffset int) (*ttfCMap, error) {
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

func parseTTFFont(buf []byte) (*ttfFontInfo, error) {
	if len(buf) < 12 {
		return nil, fmt.Errorf("ttf data too short")
	}
	numTables := int(binary.BigEndian.Uint16(buf[4:6]))
	var cmapOffset int
	var headOffset, headLen int
	var hheaOffset, hheaLen int
	var hmtxOffset, hmtxLen int

	for i := 0; i < numTables; i++ {
		off := 12 + i*16
		if off+16 > len(buf) {
			break
		}
		tag := string(buf[off : off+4])
		offset := int(binary.BigEndian.Uint32(buf[off+8 : off+12]))
		length := int(binary.BigEndian.Uint32(buf[off+12 : off+16]))
		switch tag {
		case "cmap":
			cmapOffset = offset
		case "head":
			headOffset = offset
			headLen = length
		case "hhea":
			hheaOffset = offset
			hheaLen = length
		case "hmtx":
			hmtxOffset = offset
			hmtxLen = length
		}
	}

	cmap, err := parseFormat4CMapFromOffset(buf, cmapOffset)
	if err != nil {
		return nil, err
	}

	info := &ttfFontInfo{
		cmap:        cmap,
		unitsPerEm:  2048,
		numHMetrics: 0,
	}

	if headOffset > 0 && headOffset+headLen <= len(buf) && headLen >= 20 {
		info.unitsPerEm = binary.BigEndian.Uint16(buf[headOffset+18 : headOffset+20])
	}
	if hheaOffset > 0 && hheaOffset+hheaLen <= len(buf) && hheaLen >= 36 {
		info.numHMetrics = binary.BigEndian.Uint16(buf[hheaOffset+34 : hheaOffset+36])
	}
	if hmtxOffset > 0 && hmtxOffset+hmtxLen <= len(buf) {
		info.hmtxData = buf[hmtxOffset : hmtxOffset+hmtxLen]
	}

	return info, nil
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

func (f *ttfFontInfo) glyphWidth(gid uint16) uint16 {
	if f == nil || len(f.hmtxData) < 4 || f.unitsPerEm == 0 {
		return 600
	}
	var adv uint16
	if gid < f.numHMetrics {
		off := int(gid) * 4
		if off+2 <= len(f.hmtxData) {
			adv = binary.BigEndian.Uint16(f.hmtxData[off : off+2])
		}
	} else if f.numHMetrics > 0 {
		off := int(f.numHMetrics-1) * 4
		if off+2 <= len(f.hmtxData) {
			adv = binary.BigEndian.Uint16(f.hmtxData[off : off+2])
		}
	}
	if adv == 0 {
		return 500
	}
	w := uint16((uint32(adv) * 1000) / uint32(f.unitsPerEm))
	if w == 0 {
		return 500
	}
	return w
}

func (f *ttfFontInfo) textWidthPt(text string, size int) float64 {
	if f == nil || f.cmap == nil {
		return float64(utf8.RuneCountInString(text)) * float64(size) * 0.6
	}
	var totalUnits float64
	for _, r := range text {
		gid := f.cmap.lookup(r)
		w := f.glyphWidth(gid)
		totalUnits += float64(w)
	}
	return (totalUnits * float64(size)) / 1000.0
}

var (
	ebrimaOnce sync.Once
	ebrimaFont *ttfFontInfo
)

func getEbrimaFont() *ttfFontInfo {
	ebrimaOnce.Do(func() {
		if len(ebrimaFontData) > 0 {
			font, _ := parseTTFFont(ebrimaFontData)
			ebrimaFont = font
		}
	})
	return ebrimaFont
}

func wrapPDFTextWithWidth(text string, size int, maxWidth float64, font *ttfFontInfo) []string {
	if text == "" {
		return nil
	}
	paragraphs := strings.Split(text, "\n")
	var result []string

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		words := strings.Fields(para)
		if len(words) == 0 {
			continue
		}

		var curLine strings.Builder
		curWidth := 0.0

		for _, word := range words {
			wordWidth := font.textWidthPt(word, size)
			if wordWidth > maxWidth {
				// Token exceeds maxWidth on its own; split safely into sub-chunks of runes
				runes := []rune(word)
				var chunk strings.Builder
				chunkWidth := 0.0
				for _, r := range runes {
					rW := font.textWidthPt(string(r), size)
					if chunk.Len() > 0 && chunkWidth+rW > maxWidth {
						if curLine.Len() > 0 {
							result = append(result, curLine.String())
							curLine.Reset()
							curWidth = 0.0
						}
						result = append(result, chunk.String())
						chunk.Reset()
						chunkWidth = 0.0
					}
					chunk.WriteRune(r)
					chunkWidth += rW
				}
				if chunk.Len() > 0 {
					subWord := chunk.String()
					subWidth := chunkWidth
					if curLine.Len() == 0 {
						curLine.WriteString(subWord)
						curWidth = subWidth
					} else {
						spaceWidth := font.textWidthPt(" ", size)
						if curWidth+spaceWidth+subWidth <= maxWidth {
							curLine.WriteString(" ")
							curLine.WriteString(subWord)
							curWidth += spaceWidth + subWidth
						} else {
							result = append(result, curLine.String())
							curLine.Reset()
							curLine.WriteString(subWord)
							curWidth = subWidth
						}
					}
				}
				continue
			}

			if curLine.Len() == 0 {
				curLine.WriteString(word)
				curWidth = wordWidth
			} else {
				spaceWidth := font.textWidthPt(" ", size)
				if curWidth+spaceWidth+wordWidth <= maxWidth {
					curLine.WriteString(" ")
					curLine.WriteString(word)
					curWidth += spaceWidth + wordWidth
				} else {
					result = append(result, curLine.String())
					curLine.Reset()
					curLine.WriteString(word)
					curWidth = wordWidth
				}
			}
		}

		if curLine.Len() > 0 {
			result = append(result, curLine.String())
		}
	}

	if len(result) == 0 {
		return []string{""}
	}
	return result
}

func generateOdontogramPDF(chart domain.OdontogramChart) []byte {
	font := getEbrimaFont()
	if font == nil || font.cmap == nil || len(ebrimaFontData) == 0 {
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
		if curY < 65 {
			pages = append(pages, currentPage)
			currentPage = nil
			curY = 740
		}
		currentPage = append(currentPage, pdfLine{x: x, y: curY, size: size, text: text})
		curY -= (size + 5)
	}

	pushWrapped := func(x int, size int, text string) {
		maxWidth := 562.0 - float64(x)
		lines := wrapPDFTextWithWidth(text, size, maxWidth, font)
		for _, l := range lines {
			pushLine(x, size, l)
		}
	}

	// 1. Header (wrapped to never overflow page margin)
	pushWrapped(50, 13, "Hospital Odontogram Examination Report / የጥርስ ምርመራ ሪፖርት")
	curY -= 3

	// 2. Patient Information
	patName := chart.PatientName
	if patName == "" {
		patName = "Patient " + chart.PatientID
	}
	mrn := chart.PatientMRN
	if mrn == "" {
		mrn = "N/A"
	}
	pushWrapped(50, 10, fmt.Sprintf("Patient / ታካሚ: %s  |  MRN: %s", patName, mrn))

	// 3. Attending Doctor & Date
	docName := chart.DoctorName
	if docName == "" {
		docName = "Doctor"
	}
	pushWrapped(50, 10, fmt.Sprintf("Attending Doctor: %s  |  Date: %s", docName, chart.CreatedAt.Format("2006-01-02 15:04:05")))

	// 4. Chart ID & Version
	pushWrapped(50, 10, fmt.Sprintf("Chart ID: %s  |  Version: %d", chart.ID, chart.Version))
	curY -= 6

	// 5. Clinical Examination Notes
	pushWrapped(50, 11, "Clinical Examination Notes / ክሊኒካዊ ማስታወሻ:")
	desc := chart.Description
	if desc == "" {
		desc = "Dental Examination Chart"
	}
	pushWrapped(65, 9, desc)
	curY -= 6

	// 6. Dental Condition Findings
	pushWrapped(50, 11, "Dental Condition Findings / የጥርስ ምርመራ ውጤቶች:")

	var toothNums []int
	for tNum := range chart.Odontogram {
		toothNums = append(toothNums, tNum)
	}
	sort.Ints(toothNums)

	if len(toothNums) == 0 {
		pushWrapped(65, 9, "All teeth healthy / No active pathological conditions charted")
	} else {
		for _, tNum := range toothNums {
			cond := chart.Odontogram[tNum]
			pushWrapped(65, 9, fmt.Sprintf("Tooth #%d: Condition Code [%s]", tNum, cond))
		}
	}
	curY -= 6

	// 7. Clinical Legend Definitions (F3 & F4: Displays complete meaning for all used codes, including custom definitions)
	pushWrapped(50, 11, "Clinical Legend Definitions / የሕክምና ምልክቶች መፍቻ:")

	legendDescMap := map[string]string{
		"Healthy": "Healthy Tooth - Sound tooth surface with no pathology",
		"K":       "Caries / Cavity - Active dental caries lesion",
		"C":       "Crown - Artificial full crown or cap",
		"Ce":      "Pulpitis / Deep Caries - Extensive caries with pulpal involvement",
		"D":       "Decayed / Denture - Grossly decayed or retained root",
		"KR":      "Root Canal (Karies Radix) - Endodontic therapy / Root canal treated",
		"PS":      "Porcelain / Prosthetic - Porcelain veneer or prosthetic replacement",
		"IP":      "Implant - Titanium or ceramic dental implant",
		"X":       "Missing / Extracted - Congenitally missing or extracted tooth",
		"F":       "Composite Filling - Tooth-colored composite restoration",
		"B":       "Bridge - Fixed partial denture / pontic unit",
	}

	var usedCodes []string
	seenCodes := make(map[string]bool)
	for _, tNum := range toothNums {
		c := chart.Odontogram[tNum]
		if !seenCodes[c] {
			seenCodes[c] = true
			usedCodes = append(usedCodes, c)
		}
	}
	sort.Strings(usedCodes)

	for _, c := range usedCodes {
		infoStr, exists := legendDescMap[c]
		if !exists {
			infoStr = fmt.Sprintf("Custom Clinical Condition Code [%s]", c)
		}
		pushWrapped(65, 9, fmt.Sprintf("[%s] %s", c, infoStr))
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
	usedGIDs[font.cmap.lookup(' ')] = ' '

	encodeHex := func(s string) string {
		var b bytes.Buffer
		for _, r := range s {
			gid := font.cmap.lookup(r)
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

	// Collect used GIDs and build /DW and /W array for CIDFont (F4)
	var sortedGIDs []int
	for gid := range usedGIDs {
		sortedGIDs = append(sortedGIDs, int(gid))
	}
	sort.Ints(sortedGIDs)

	var wBuf bytes.Buffer
	wBuf.WriteString("/DW 600\n  /W [\n")
	for _, gid := range sortedGIDs {
		w := font.glyphWidth(uint16(gid))
		fmt.Fprintf(&wBuf, "    %d [%d]\n", gid, w)
	}
	wBuf.WriteString("  ]")

	// CIDFont
	offsets[cidFontObj] = doc.Len()
	doc.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Font /Subtype /CIDFontType2 /BaseFont /Ebrima /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /FontDescriptor %d 0 R /CIDToGIDMap /Identity\n  %s\n>>\nendobj\n", cidFontObj, fontDescObj, wBuf.String()))

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
