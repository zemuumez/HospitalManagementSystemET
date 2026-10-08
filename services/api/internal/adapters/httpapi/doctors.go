package httpapi

import (
	"net/http"
	"strings"

	"hms.local/api/internal/domain"
)

type createDoctorRequest struct {
	UserID            string               `json:"userId,omitempty"`
	ID                string               `json:"id,omitempty"`
	Name              string               `json:"name,omitempty"`
	Email             string               `json:"email,omitempty"`
	DepartmentID      string               `json:"departmentId,omitempty"`
	DepartmentIdSnake string               `json:"department_id,omitempty"`
	Department        string               `json:"department,omitempty"`
	Specialist        string               `json:"specialist,omitempty"`
	Designation       string               `json:"designation,omitempty"`
	Qualification     string               `json:"qualification,omitempty"`
	Phone             string               `json:"phone,omitempty"`
	Gender            string               `json:"gender,omitempty"`
	DateOfBirth       string               `json:"dateOfBirth,omitempty"`
	DOB               string               `json:"dob,omitempty"`
	BloodGroup        string               `json:"bloodGroup,omitempty"`
	BloodGroupSnake   string               `json:"blood_group,omitempty"`
	Address1          string               `json:"address1,omitempty"`
	Address2          string               `json:"address2,omitempty"`
	City              string               `json:"city,omitempty"`
	Zip               string               `json:"zip,omitempty"`
	Description       string               `json:"description,omitempty"`
	PhotoURL          string               `json:"photoUrl,omitempty"`
	PhotoUrlSnake     string               `json:"photo_url,omitempty"`
	AppointmentCharge *float64             `json:"appointmentCharge,omitempty"`
	OpdCharge         *float64             `json:"opdCharge,omitempty"`
	SlotMinutes       int                  `json:"slotMinutes,omitempty"`
	Hours             []domain.DoctorHours `json:"hours,omitempty"`
	Version           int                  `json:"version,omitempty"`
}

type updateDoctorRequest struct {
	Name              *string              `json:"name,omitempty"`
	Email             *string              `json:"email,omitempty"`
	DepartmentID      *string              `json:"departmentId,omitempty"`
	DepartmentIdSnake *string              `json:"department_id,omitempty"`
	Specialist        *string              `json:"specialist,omitempty"`
	Designation       *string              `json:"designation,omitempty"`
	Qualification     *string              `json:"qualification,omitempty"`
	Phone             *string              `json:"phone,omitempty"`
	Gender            *string              `json:"gender,omitempty"`
	DateOfBirth       *string              `json:"dateOfBirth,omitempty"`
	DOB               *string              `json:"dob,omitempty"`
	BloodGroup        *string              `json:"bloodGroup,omitempty"`
	BloodGroupSnake   *string              `json:"blood_group,omitempty"`
	Address1          *string              `json:"address1,omitempty"`
	Address2          *string              `json:"address2,omitempty"`
	City              *string              `json:"city,omitempty"`
	Zip               *string              `json:"zip,omitempty"`
	Description       *string              `json:"description,omitempty"`
	PhotoURL          *string              `json:"photoUrl,omitempty"`
	PhotoUrlSnake     *string              `json:"photo_url,omitempty"`
	AppointmentCharge *float64             `json:"appointmentCharge,omitempty"`
	OpdCharge         *float64             `json:"opdCharge,omitempty"`
	SlotMinutes       *int                 `json:"slotMinutes,omitempty"`
	Hours             []domain.DoctorHours `json:"hours,omitempty"`
	Version           int                  `json:"version"`
}

type doctorStatusRequest struct {
	Active *bool `json:"active,omitempty"`
	Status *int  `json:"status,omitempty"`
}

func (s Server) doctors(w http.ResponseWriter, r *http.Request, a domain.Actor) bool {
	if strings.HasSuffix(r.URL.Path, "/ext") {
		return false
	}
	if r.URL.Path != "/v1/doctors" && !strings.HasPrefix(r.URL.Path, "/v1/doctors/") {
		return false
	}

	switch {
	case r.URL.Path == "/v1/doctors" && r.Method == "GET":
		status := r.URL.Query().Get("status")
		deptID := r.URL.Query().Get("departmentId")
		if deptID == "" {
			deptID = r.URL.Query().Get("department_id")
		}
		search := r.URL.Query().Get("search")
		out, err := s.Scheduling.ListDoctors(r.Context(), a, status, deptID, search)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"doctors": out})
		return true

	case r.URL.Path == "/v1/doctors" && r.Method == "POST":
		var req createDoctorRequest
		if !decode(w, r, &req) {
			return true
		}
		userID := req.UserID
		if userID == "" {
			userID = req.ID
		}
		deptID := req.DepartmentID
		if deptID == "" {
			deptID = req.DepartmentIdSnake
		}

		if req.Specialist != "" || deptID != "" {
			dob := req.DateOfBirth
			if dob == "" {
				dob = req.DOB
			}
			bg := req.BloodGroup
			if bg == "" {
				bg = req.BloodGroupSnake
			}
			photo := req.PhotoURL
			if photo == "" {
				photo = req.PhotoUrlSnake
			}
			var apptCharge, opdCharge float64
			if req.AppointmentCharge != nil {
				apptCharge = *req.AppointmentCharge
			}
			if req.OpdCharge != nil {
				opdCharge = *req.OpdCharge
			}
			in := domain.CreateDoctorInput{
				UserID:            userID,
				DepartmentID:      deptID,
				Specialist:        req.Specialist,
				Designation:       req.Designation,
				Qualification:     req.Qualification,
				Phone:             req.Phone,
				Gender:            req.Gender,
				DateOfBirth:       dob,
				BloodGroup:        bg,
				Address1:          req.Address1,
				Address2:          req.Address2,
				City:              req.City,
				Zip:               req.Zip,
				Description:       req.Description,
				PhotoURL:          photo,
				AppointmentCharge: apptCharge,
				OpdCharge:         opdCharge,
				SlotMinutes:       req.SlotMinutes,
				Hours:             req.Hours,
			}
			out, err := s.Scheduling.CreateDoctorProfile(r.Context(), a, in)
			if err != nil {
				fail(w, err)
				return true
			}
			write(w, 201, out)
			return true
		}

		// Backward-compatible fallback for legacy SaveDoctor
		d := domain.Doctor{
			ID:          req.ID,
			Department:  req.Department,
			SlotMinutes: req.SlotMinutes,
			Hours:       req.Hours,
			Version:     req.Version,
		}
		out, err := s.Scheduling.SaveDoctor(r.Context(), a, d)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, out)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/doctors/") && strings.HasSuffix(r.URL.Path, "/status") && r.Method == "PATCH":
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/doctors/"), "/status")
		if id == "" || strings.Contains(id, "/") {
			return false
		}
		var req doctorStatusRequest
		if !decode(w, r, &req) {
			return true
		}
		if req.Active == nil && req.Status == nil {
			fail(w, domain.ErrValidation)
			return true
		}
		if req.Status != nil && (*req.Status != 0 && *req.Status != 1) {
			fail(w, domain.ErrValidation)
			return true
		}
		if req.Active != nil && req.Status != nil {
			statusActive := (*req.Status == 1)
			if *req.Active != statusActive {
				fail(w, domain.ErrValidation)
				return true
			}
		}
		var active bool
		if req.Active != nil {
			active = *req.Active
		} else {
			active = (*req.Status == 1)
		}
		err := s.Scheduling.SetDoctorStatus(r.Context(), a, id, active)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, map[string]any{"id": id, "active": active, "saved": true})
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/doctors/") && r.Method == "GET":
		id := strings.TrimPrefix(r.URL.Path, "/v1/doctors/")
		if id == "" || strings.Contains(id, "/") {
			return false
		}
		doc, err := s.Scheduling.Doctor(r.Context(), a, id)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, doc)
		return true

	case strings.HasPrefix(r.URL.Path, "/v1/doctors/") && r.Method == "PUT":
		id := strings.TrimPrefix(r.URL.Path, "/v1/doctors/")
		if id == "" || strings.Contains(id, "/") {
			return false
		}
		var req updateDoctorRequest
		if !decode(w, r, &req) {
			return true
		}
		deptID := req.DepartmentID
		if deptID == nil {
			deptID = req.DepartmentIdSnake
		}
		dob := req.DateOfBirth
		if dob == nil {
			dob = req.DOB
		}
		bg := req.BloodGroup
		if bg == nil {
			bg = req.BloodGroupSnake
		}
		photo := req.PhotoURL
		if photo == nil {
			photo = req.PhotoUrlSnake
		}
		in := domain.UpdateDoctorInput{
			Name:              req.Name,
			Email:             req.Email,
			DepartmentID:      deptID,
			Specialist:        req.Specialist,
			Designation:       req.Designation,
			Qualification:     req.Qualification,
			Phone:             req.Phone,
			Gender:            req.Gender,
			DateOfBirth:       dob,
			BloodGroup:        bg,
			Address1:          req.Address1,
			Address2:          req.Address2,
			City:              req.City,
			Zip:               req.Zip,
			Description:       req.Description,
			PhotoURL:          photo,
			AppointmentCharge: req.AppointmentCharge,
			OpdCharge:         req.OpdCharge,
			SlotMinutes:       req.SlotMinutes,
			Hours:             req.Hours,
			Version:           req.Version,
		}
		out, err := s.Scheduling.UpdateDoctorProfile(r.Context(), a, id, in)
		if err != nil {
			fail(w, err)
			return true
		}
		write(w, 200, out)
		return true
	}

	return false
}
