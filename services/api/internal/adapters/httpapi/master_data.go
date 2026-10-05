package httpapi

import (
	"hms.local/api/internal/domain"
	"net/http"
	"strconv"
	"strings"
)

// masterData dispatches routes for doctor departments, doctor ext profile,
// hospital hours, hospital date overrides and patient profile extensions.
func (s Server) masterData(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	switch {
	// ── Doctor departments ────────────────────────────────────────────────────
	case r.URL.Path == "/v1/doctor-departments" && r.Method == "GET":
		archived := r.URL.Query().Get("archived") == "true"
		out, e := s.MasterData.DoctorDepartments(r.Context(), a, archived)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, map[string]any{"departments": out})
		return true

	case r.URL.Path == "/v1/doctor-departments" && r.Method == "POST":
		var i domain.DoctorDepartmentInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.MasterData.CreateDoctorDepartment(r.Context(), a, i)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 201, out)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/doctor-departments/") &&
		!strings.Contains(strings.TrimPrefix(r.URL.Path, "/v1/doctor-departments/"), "/") &&
		r.Method == "GET":
		id := strings.TrimPrefix(r.URL.Path, "/v1/doctor-departments/")
		out, e := s.MasterData.DoctorDepartment(r.Context(), a, id)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, out)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/doctor-departments/") &&
		!strings.Contains(strings.TrimPrefix(r.URL.Path, "/v1/doctor-departments/"), "/") &&
		r.Method == "PATCH":
		id := strings.TrimPrefix(r.URL.Path, "/v1/doctor-departments/")
		var i domain.DoctorDepartmentInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.MasterData.UpdateDoctorDepartment(r.Context(), a, id, i)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, out)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/doctor-departments/") &&
		strings.HasSuffix(r.URL.Path, "/archive") &&
		r.Method == "POST":
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/doctor-departments/"), "/archive")
		var body struct {
			Version int    `json:"version"`
			Reason  string `json:"reason"`
		}
		if !decode(w, r, &body) {
			return true
		}
		out, e := s.MasterData.ArchiveDoctorDepartment(r.Context(), a, id, body.Version, body.Reason)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, out)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/doctor-departments/") &&
		strings.HasSuffix(r.URL.Path, "/revisions") &&
		r.Method == "GET":
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/doctor-departments/"), "/revisions")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		out, e := s.MasterData.DoctorDepartmentRevisions(r.Context(), a, id, page)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, map[string]any{"revisions": out, "page": page, "pageSize": 25})
		return true

	// ── Doctor profile extensions ─────────────────────────────────────────────
	case strings.HasPrefix(r.URL.Path, "/v1/doctors/") &&
		strings.HasSuffix(r.URL.Path, "/ext") &&
		r.Method == "GET":
		doctorID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/doctors/"), "/ext")
		out, e := s.MasterData.DoctorExt(r.Context(), a, doctorID)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, out)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/doctors/") &&
		strings.HasSuffix(r.URL.Path, "/ext") &&
		r.Method == "PUT":
		doctorID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/doctors/"), "/ext")
		var i domain.DoctorExtInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.MasterData.SaveDoctorExt(r.Context(), a, doctorID, i)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, out)
		return true

	// ── Hospital hours ────────────────────────────────────────────────────────
	case r.URL.Path == "/v1/hospital-hours" && r.Method == "GET":
		out, e := s.MasterData.HospitalHours(r.Context(), a)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, map[string]any{"hours": out})
		return true

	case r.URL.Path == "/v1/hospital-hours" && r.Method == "PUT":
		var i domain.HospitalHoursInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.MasterData.SaveHospitalHours(r.Context(), a, i)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, map[string]any{"hours": out})
		return true

	// ── Hospital date overrides ───────────────────────────────────────────────
	case r.URL.Path == "/v1/hospital-date-overrides" && r.Method == "GET":
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		out, e := s.MasterData.HospitalDateOverrides(r.Context(), a, page)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, map[string]any{"overrides": out, "page": page, "pageSize": 25})
		return true

	case r.URL.Path == "/v1/hospital-date-overrides" && r.Method == "POST":
		var i domain.HospitalDateOverrideInput
		if !decode(w, r, &i) {
			return true
		}
		out, e := s.MasterData.CreateHospitalDateOverride(r.Context(), a, i)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 201, out)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/hospital-date-overrides/") && r.Method == "DELETE":
		id := strings.TrimPrefix(r.URL.Path, "/v1/hospital-date-overrides/")
		if e := s.MasterData.DeleteHospitalDateOverride(r.Context(), a, id); e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, map[string]bool{"deleted": true})
		return true

	// ── Patient profile extensions ────────────────────────────────────────────
	case strings.HasPrefix(r.URL.Path, "/v1/patients/") &&
		strings.HasSuffix(r.URL.Path, "/ext") &&
		r.Method == "GET":
		patientID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/patients/"), "/ext")
		out, e := s.MasterData.PatientProfileExt(r.Context(), a, patientID)
		if e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, out)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/patients/") &&
		strings.HasSuffix(r.URL.Path, "/ext") &&
		r.Method == "PUT":
		patientID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/patients/"), "/ext")
		var i domain.PatientProfileExt
		if !decode(w, r, &i) {
			return true
		}
		if e := s.MasterData.SavePatientProfileExt(r.Context(), a, patientID, i); e != nil {
			fail(w, e)
			return true
		}
		write(w, 200, map[string]bool{"saved": true})
		return true
	}
	return false
}
