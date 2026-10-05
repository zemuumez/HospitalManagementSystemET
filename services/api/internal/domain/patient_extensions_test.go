package domain_test

import (
	"testing"

	"hms.local/api/internal/domain"
)

func TestPatientContactConsent_Validate(t *testing.T) {
	t.Run("valid consent", func(t *testing.T) {
		c := domain.PatientContactConsent{
			PatientID:          "d3b07384-d113-4a7b-a621-000000000001",
			GuardianName:       "Jane Doe",
			GuardianRelation:   "Mother",
			GuardianPhone:      "+251911223344",
			GuardianEmail:      "jane@example.com",
			SmsConsent:         true,
			EmailConsent:       true,
			DataSharingConsent: true,
		}
		if err := c.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("invalid patient id", func(t *testing.T) {
		c := domain.PatientContactConsent{
			PatientID: "not-a-uuid",
		}
		if err := c.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestIssueSmartCardInput_Validate(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		in := domain.IssueSmartCardInput{
			PatientID:    "d3b07384-d113-4a7b-a621-000000000001",
			TemplateName: "standard",
			ValidityDays: 365,
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("invalid patient id", func(t *testing.T) {
		in := domain.IssueSmartCardInput{
			PatientID: "bad-id",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestMergePatientInput_Validate(t *testing.T) {
	t.Run("valid merge", func(t *testing.T) {
		in := domain.MergePatientInput{
			PrimaryPatientID: "d3b07384-d113-4a7b-a621-000000000001",
			MergedPatientID:  "d3b07384-d113-4a7b-a621-000000000002",
			Reason:           "Duplicate account created by front desk in error",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("same patient id", func(t *testing.T) {
		in := domain.MergePatientInput{
			PrimaryPatientID: "d3b07384-d113-4a7b-a621-000000000001",
			MergedPatientID:  "d3b07384-d113-4a7b-a621-000000000001",
			Reason:           "Duplicate",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("empty reason", func(t *testing.T) {
		in := domain.MergePatientInput{
			PrimaryPatientID: "d3b07384-d113-4a7b-a621-000000000001",
			MergedPatientID:  "d3b07384-d113-4a7b-a621-000000000002",
			Reason:           "",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}
