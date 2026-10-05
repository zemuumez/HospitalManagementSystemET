package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"hms.local/api/internal/domain"
)

func (s Server) diagnosticReports(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	switch {
	// ─── Categories & Units ───────────────────────────────────────────────────────
	case r.URL.Path == "/v1/diagnostic-categories":
		if r.Method == "GET" {
			kind := r.URL.Query().Get("kind")
			list, err := s.DiagnosticReports.DiagnosticCategories(r.Context(), a, kind)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"categories": list})
			return true
		}
		if r.Method == "POST" {
			var input domain.DiagnosticCategoryInput
			if !decode(w, r, &input) {
				return true
			}
			c, err := s.DiagnosticReports.CreateDiagnosticCategory(r.Context(), a, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, c)
			return true
		}

	case r.URL.Path == "/v1/diagnostic-units":
		if r.Method == "GET" {
			list, err := s.DiagnosticReports.DiagnosticUnits(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"units": list})
			return true
		}
		if r.Method == "POST" {
			var input domain.DiagnosticUnitInput
			if !decode(w, r, &input) {
				return true
			}
			u, err := s.DiagnosticReports.CreateDiagnosticUnit(r.Context(), a, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, u)
			return true
		}

	// ─── Diagnostic Report Files & Portal Release ─────────────────────────────────
	case strings.HasPrefix(r.URL.Path, "/v1/diagnostic-orders/") && strings.HasSuffix(r.URL.Path, "/files"):
		orderID := strings.TrimPrefix(r.URL.Path, "/v1/diagnostic-orders/")
		orderID = strings.TrimSuffix(orderID, "/files")
		if r.Method == "GET" {
			files, err := s.DiagnosticReports.DiagnosticReportFiles(r.Context(), a, orderID)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"files": files})
			return true
		}
		if r.Method == "POST" {
			var input domain.UploadDiagnosticReportFileInput
			if !decode(w, r, &input) {
				return true
			}
			f, err := s.DiagnosticReports.UploadDiagnosticReportFile(r.Context(), a, orderID, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, f)
			return true
		}

	case strings.HasPrefix(r.URL.Path, "/v1/diagnostic-report-files/") && strings.HasSuffix(r.URL.Path, "/release") && r.Method == "POST":
		fileID := strings.TrimPrefix(r.URL.Path, "/v1/diagnostic-report-files/")
		fileID = strings.TrimSuffix(fileID, "/release")
		if err := s.DiagnosticReports.ReleaseReportFileToPortal(r.Context(), a, fileID); err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]bool{"ok": true})
		return true

	// ─── Diagnosis Templates ──────────────────────────────────────────────────────
	case r.URL.Path == "/v1/diagnosis-templates":
		if r.Method == "GET" {
			templates, err := s.DiagnosticReports.DiagnosisTemplates(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"templates": templates})
			return true
		}
		if r.Method == "POST" {
			var input domain.DiagnosisTemplateInput
			if !decode(w, r, &input) {
				return true
			}
			t, err := s.DiagnosticReports.CreateDiagnosisTemplate(r.Context(), a, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, t)
			return true
		}

	// ─── Vaccines ─────────────────────────────────────────────────────────────────
	case r.URL.Path == "/v1/vaccines":
		if r.Method == "GET" {
			list, err := s.DiagnosticReports.Vaccines(r.Context(), a)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"vaccines": list})
			return true
		}
		if r.Method == "POST" {
			var input domain.VaccineInput
			if !decode(w, r, &input) {
				return true
			}
			v, err := s.DiagnosticReports.CreateVaccine(r.Context(), a, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, v)
			return true
		}

	// ─── Vital Reports ────────────────────────────────────────────────────────────
	case r.URL.Path == "/v1/vital-reports/birth":
		if r.Method == "GET" {
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			reports, err := s.DiagnosticReports.BirthReports(r.Context(), a, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"reports": reports})
			return true
		}
		if r.Method == "POST" {
			var input domain.CreateBirthReportInput
			if !decode(w, r, &input) {
				return true
			}
			br, err := s.DiagnosticReports.CreateBirthReport(r.Context(), a, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, br)
			return true
		}

	case r.URL.Path == "/v1/vital-reports/death":
		if r.Method == "GET" {
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			reports, err := s.DiagnosticReports.DeathReports(r.Context(), a, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"reports": reports})
			return true
		}
		if r.Method == "POST" {
			var input domain.CreateDeathReportInput
			if !decode(w, r, &input) {
				return true
			}
			dr, err := s.DiagnosticReports.CreateDeathReport(r.Context(), a, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, dr)
			return true
		}

	case r.URL.Path == "/v1/vital-reports/operation":
		if r.Method == "GET" {
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			reports, err := s.DiagnosticReports.OperationReports(r.Context(), a, page)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 200, map[string]any{"reports": reports})
			return true
		}
		if r.Method == "POST" {
			var input domain.CreateOperationReportInput
			if !decode(w, r, &input) {
				return true
			}
			opr, err := s.DiagnosticReports.CreateOperationReport(r.Context(), a, input)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, opr)
			return true
		}

	// ─── Patient Vaccinations & Investigations ────────────────────────────────────
	case strings.HasPrefix(r.URL.Path, "/v1/patients/"):
		rest := strings.TrimPrefix(r.URL.Path, "/v1/patients/")
		parts := strings.Split(rest, "/")
		if len(parts) >= 2 {
			patientID := parts[0]
			sub := parts[1]

			switch sub {
			case "vaccinations":
				if r.Method == "GET" {
					list, err := s.DiagnosticReports.PatientVaccinations(r.Context(), a, patientID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]any{"vaccinations": list})
					return true
				}
				if r.Method == "POST" {
					var input domain.AdministerVaccineInput
					if !decode(w, r, &input) {
						return true
					}
					input.PatientID = patientID
					pv, err := s.DiagnosticReports.AdministerVaccine(r.Context(), a, input)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 201, pv)
					return true
				}

			case "investigations":
				if r.Method == "GET" {
					list, err := s.DiagnosticReports.InvestigationReports(r.Context(), a, patientID)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 200, map[string]any{"investigations": list})
					return true
				}
				if r.Method == "POST" {
					var input domain.CreateInvestigationReportInput
					if !decode(w, r, &input) {
						return true
					}
					input.PatientID = patientID
					ir, err := s.DiagnosticReports.CreateInvestigationReport(r.Context(), a, input)
					if err != nil {
						fail(w, err)
						return true
					}
					write(w, 201, ir)
					return true
				}
			}
		}
	}

	return false
}
