package domain_test

import (
	"testing"

	"hms.local/api/internal/domain"
)

func TestEnqueuePatientInput_Validate(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		in := domain.EnqueuePatientInput{
			DoctorID:  "doc-123",
			PatientID: "d3b07384-d113-4a7b-a621-000000000001",
			QueueDate: "2026-10-05",
			Notes:     "Regular checkup",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		in := domain.EnqueuePatientInput{
			DoctorID:  "doc-123",
			PatientID: "d3b07384-d113-4a7b-a621-000000000001",
			QueueDate: "invalid-date",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestUpdateQueueStatusInput_Validate(t *testing.T) {
	t.Run("valid status", func(t *testing.T) {
		for _, s := range []string{"in_consultation", "completed", "skipped"} {
			in := domain.UpdateQueueStatusInput{
				Status: s,
			}
			if err := in.Validate(); err != nil {
				t.Fatalf("expected nil error for status %s, got %v", s, err)
			}
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		in := domain.UpdateQueueStatusInput{
			Status: "invalid_status",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestCreatePublicAppointmentRequestInput_Validate(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		in := domain.CreatePublicAppointmentRequestInput{
			PatientName:   "Abebe Bikila",
			PatientEmail:  "abebe@example.com",
			PatientPhone:  "+251911002233",
			DoctorID:      "doc-456",
			PreferredDate: "2026-10-10",
			Problem:       "Persistent cough and mild fever",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		in := domain.CreatePublicAppointmentRequestInput{
			PatientName:   "  ",
			PatientPhone:  "+251911002233",
			DoctorID:      "doc-456",
			PreferredDate: "2026-10-10",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("short phone", func(t *testing.T) {
		in := domain.CreatePublicAppointmentRequestInput{
			PatientName:   "Abebe Bikila",
			PatientPhone:  "123",
			DoctorID:      "doc-456",
			PreferredDate: "2026-10-10",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}
