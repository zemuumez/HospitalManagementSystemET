package domain_test

import (
	"testing"

	"hms.local/api/internal/domain"
)

func TestDoctorDepartmentInput_Validate(t *testing.T) {
	t.Run("valid create", func(t *testing.T) {
		in := domain.DoctorDepartmentInput{
			Title:       "Cardiology",
			Description: "Heart care",
		}
		if err := in.Validate(true); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		in := domain.DoctorDepartmentInput{
			Title: "   ",
		}
		if err := in.Validate(true); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("valid update", func(t *testing.T) {
		in := domain.DoctorDepartmentInput{
			Title:   "Neurology",
			Version: 1,
			Reason:  "Name update",
		}
		if err := in.Validate(false); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("update missing version or reason", func(t *testing.T) {
		in := domain.DoctorDepartmentInput{
			Title:   "Neurology",
			Version: 0,
			Reason:  "Name update",
		}
		if err := in.Validate(false); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}

		in2 := domain.DoctorDepartmentInput{
			Title:   "Neurology",
			Version: 1,
			Reason:  "",
		}
		if err := in2.Validate(false); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestDoctorExtInput_Validate(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		in := domain.DoctorExtInput{
			DepartmentID:      "d3b07384-d113-4a7b-a621-000000000001",
			Description:       "Senior Consultant",
			PhotoURL:          "https://example.com/photo.jpg",
			OpdCharge:         500.0,
			AppointmentCharge: 350.0,
			Version:           1,
			Reason:            "Initial setup",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("negative charge", func(t *testing.T) {
		in := domain.DoctorExtInput{
			OpdCharge: -10,
			Version:   1,
			Reason:    "Update",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestHospitalHoursInput_Validate(t *testing.T) {
	t.Run("valid schedule", func(t *testing.T) {
		in := domain.HospitalHoursInput{
			Hours: []domain.HospitalHoursEntry{
				{Weekday: 1, OpenMinute: 480, CloseMinute: 1020}, // Mon 08:00 - 17:00
				{Weekday: 2, OpenMinute: 480, CloseMinute: 1020},
			},
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("duplicate weekday", func(t *testing.T) {
		in := domain.HospitalHoursInput{
			Hours: []domain.HospitalHoursEntry{
				{Weekday: 1, OpenMinute: 480, CloseMinute: 1020},
				{Weekday: 1, OpenMinute: 600, CloseMinute: 1100},
			},
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid times", func(t *testing.T) {
		in := domain.HospitalHoursInput{
			Hours: []domain.HospitalHoursEntry{
				{Weekday: 1, OpenMinute: 1000, CloseMinute: 500},
			},
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestHospitalDateOverrideInput_Validate(t *testing.T) {
	t.Run("valid closed override", func(t *testing.T) {
		in := domain.HospitalDateOverrideInput{
			OverrideDate: "2026-12-25",
			Closed:       true,
			Label:        "Christmas Day",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("valid open override with times", func(t *testing.T) {
		open := 540
		close := 720
		in := domain.HospitalDateOverrideInput{
			OverrideDate: "2026-12-26",
			Closed:       false,
			OpenMinute:   &open,
			CloseMinute:  &close,
			Label:        "Half Day",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("closed with times set fails", func(t *testing.T) {
		open := 540
		in := domain.HospitalDateOverrideInput{
			OverrideDate: "2026-12-25",
			Closed:       true,
			OpenMinute:   &open,
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestPatientProfileExt_Validate(t *testing.T) {
	t.Run("valid ext", func(t *testing.T) {
		ext := domain.PatientProfileExt{
			FatherName:     "John Doe Sr.",
			Religion:       "Christianity",
			ReferralSource: "General Practitioner",
			Notes:          "Known allergy to penicillin",
		}
		if err := ext.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}
