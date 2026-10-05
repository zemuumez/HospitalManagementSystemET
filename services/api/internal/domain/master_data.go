package domain

import (
	"strings"
	"time"
)

// ─── Doctor departments ───────────────────────────────────────────────────────

type DoctorDepartment struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Archived    bool      `json:"archived"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type DoctorDepartmentInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     int    `json:"version"` // 0 = create, >0 = update
	Reason      string `json:"reason"`
}

func (i *DoctorDepartmentInput) Validate(isCreate bool) error {
	i.Title = strings.TrimSpace(i.Title)
	i.Description = strings.TrimSpace(i.Description)
	i.Reason = strings.TrimSpace(i.Reason)
	if len([]rune(i.Title)) < 1 || len([]rune(i.Title)) > 160 {
		return ErrValidation
	}
	if len(i.Description) > 2000 {
		return ErrValidation
	}
	if !isCreate {
		if i.Version < 1 || i.Version > 1000000000 {
			return ErrValidation
		}
		if !validText(&i.Reason, 1, 500) {
			return ErrValidation
		}
	}
	return nil
}

type DoctorDepartmentRevision struct {
	Version    int              `json:"version"`
	ActorID    string           `json:"actorId"`
	Reason     string           `json:"reason"`
	Before     DoctorDepartment `json:"before"`
	After      DoctorDepartment `json:"after"`
	CreatedAt  time.Time        `json:"createdAt"`
}

// ─── Doctor profile extensions ────────────────────────────────────────────────

type DoctorExtInput struct {
	DepartmentID      string  `json:"departmentId"`
	Description       string  `json:"description"`
	PhotoURL          string  `json:"photoUrl"`
	OpdCharge         float64 `json:"opdCharge"`
	AppointmentCharge float64 `json:"appointmentCharge"`
	Version           int     `json:"version"`
	Reason            string  `json:"reason"`
}

func (i *DoctorExtInput) Validate() error {
	i.Description = strings.TrimSpace(i.Description)
	i.PhotoURL = strings.TrimSpace(i.PhotoURL)
	i.Reason = strings.TrimSpace(i.Reason)
	if len(i.Description) > 2000 || len(i.PhotoURL) > 500 {
		return ErrValidation
	}
	if i.OpdCharge < 0 || i.OpdCharge > 1_000_000 || i.AppointmentCharge < 0 || i.AppointmentCharge > 1_000_000 {
		return ErrValidation
	}
	if i.Version < 1 || i.Version > 1000000000 {
		return ErrValidation
	}
	if !validText(&i.Reason, 1, 500) {
		return ErrValidation
	}
	if i.DepartmentID != "" && !UUIDPattern.MatchString(i.DepartmentID) {
		return ErrValidation
	}
	return nil
}

type DoctorExt struct {
	DoctorID          string  `json:"doctorId"`
	DepartmentID      string  `json:"departmentId"`
	DepartmentTitle   string  `json:"departmentTitle"`
	Description       string  `json:"description"`
	PhotoURL          string  `json:"photoUrl"`
	OpdCharge         float64 `json:"opdCharge"`
	AppointmentCharge float64 `json:"appointmentCharge"`
	Version           int     `json:"version"`
}

// ─── Hospital opening hours ───────────────────────────────────────────────────

type HospitalHoursEntry struct {
	Weekday     int `json:"weekday"`
	OpenMinute  int `json:"openMinute"`
	CloseMinute int `json:"closeMinute"`
}

func (h HospitalHoursEntry) valid() bool {
	return h.Weekday >= 0 && h.Weekday <= 6 &&
		h.OpenMinute >= 0 && h.OpenMinute <= 1439 &&
		h.CloseMinute > h.OpenMinute && h.CloseMinute <= 1440
}

type HospitalHoursInput struct {
	Hours []HospitalHoursEntry `json:"hours"`
}

func (i *HospitalHoursInput) Validate() error {
	if len(i.Hours) > 7 {
		return ErrValidation
	}
	seen := map[int]bool{}
	for _, h := range i.Hours {
		if !h.valid() || seen[h.Weekday] {
			return ErrValidation
		}
		seen[h.Weekday] = true
	}
	return nil
}

// ─── Hospital date overrides ──────────────────────────────────────────────────

type HospitalDateOverride struct {
	ID           string  `json:"id"`
	OverrideDate string  `json:"overrideDate"` // YYYY-MM-DD
	Closed       bool    `json:"closed"`
	OpenMinute   *int    `json:"openMinute,omitempty"`
	CloseMinute  *int    `json:"closeMinute,omitempty"`
	Label        string  `json:"label"`
	ActorID      string  `json:"actorId"`
	CreatedAt    time.Time `json:"createdAt"`
}

type HospitalDateOverrideInput struct {
	OverrideDate string `json:"overrideDate"`
	Closed       bool   `json:"closed"`
	OpenMinute   *int   `json:"openMinute,omitempty"`
	CloseMinute  *int   `json:"closeMinute,omitempty"`
	Label        string `json:"label"`
}

func (i *HospitalDateOverrideInput) Validate() error {
	i.Label = strings.TrimSpace(i.Label)
	if len(i.Label) > 200 {
		return ErrValidation
	}
	_, err := time.Parse("2006-01-02", i.OverrideDate)
	if err != nil {
		return ErrValidation
	}
	if i.Closed {
		if i.OpenMinute != nil || i.CloseMinute != nil {
			return ErrValidation
		}
	} else {
		if i.OpenMinute == nil || i.CloseMinute == nil {
			return ErrValidation
		}
		if *i.OpenMinute < 0 || *i.OpenMinute > 1439 || *i.CloseMinute <= *i.OpenMinute || *i.CloseMinute > 1440 {
			return ErrValidation
		}
	}
	return nil
}

// ─── Patient profile additional fields (embedded in PatientProfileInput) ──────
// PatientProfileExtInput holds the additional source-form fields added in migration 036.
// It is expected to be embedded into the patient_profile table via an ALTER TABLE.
// The existing PatientProfileInput is extended via the PatientProfileExt struct below.

type PatientProfileExt struct {
	FatherName     string `json:"fatherName"`
	Religion       string `json:"religion"`
	ReferralSource string `json:"referralSource"`
	Notes          string `json:"notes"`
}

func (e *PatientProfileExt) Validate() error {
	e.FatherName = strings.TrimSpace(e.FatherName)
	e.Religion = strings.TrimSpace(e.Religion)
	e.ReferralSource = strings.TrimSpace(e.ReferralSource)
	e.Notes = strings.TrimSpace(e.Notes)
	if len([]rune(e.FatherName)) > 150 || len([]rune(e.Religion)) > 100 ||
		len([]rune(e.ReferralSource)) > 200 || len([]rune(e.Notes)) > 4000 {
		return ErrValidation
	}
	return nil
}
