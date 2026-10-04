package domain_test

import (
	"testing"

	"hms.local/api/internal/domain"
)

func TestAssignBedInput_Validate(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		in := domain.AssignBedInput{
			BedID:       "d3b07384-d113-4a7b-a621-000000000001",
			EncounterID: "d3b07384-d113-4a7b-a621-000000000002",
			PatientID:   "d3b07384-d113-4a7b-a621-000000000003",
			Notes:       "Post-op recovery bed",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("invalid bed id", func(t *testing.T) {
		in := domain.AssignBedInput{
			BedID:       "invalid",
			EncounterID: "d3b07384-d113-4a7b-a621-000000000002",
			PatientID:   "d3b07384-d113-4a7b-a621-000000000003",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestCareTeamMember_Validate(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		in := domain.AddCareTeamMemberInput{
			StaffID:   "nurse-101",
			RoleTitle: "Primary Ward Nurse",
			Notes:     "Night shift lead",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("empty role title", func(t *testing.T) {
		in := domain.AddCareTeamMemberInput{
			StaffID:   "nurse-101",
			RoleTitle: "  ",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestEncounterDiagnosis_Validate(t *testing.T) {
	t.Run("valid diagnosis", func(t *testing.T) {
		in := domain.AddDiagnosisInput{
			ICD10Code:   "J18.9",
			Description: "Pneumonia, unspecified organism",
			Category:    "final",
			Status:      "active",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		in := domain.AddDiagnosisInput{
			Description: "   ",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestEncounterProcedure_Validate(t *testing.T) {
	t.Run("valid procedure", func(t *testing.T) {
		in := domain.AddProcedureInput{
			Name:           "Appendectomy",
			Description:    "Laparoscopic appendectomy",
			AnesthesiaType: "General",
			Findings:       "Acute inflamed appendix without perforation",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}

func TestDischargeSummary_Validate(t *testing.T) {
	t.Run("valid discharge summary", func(t *testing.T) {
		followDate := "2026-10-20"
		ds := domain.DischargeSummary{
			EncounterID:          "d3b07384-d113-4a7b-a621-000000000001",
			AdmissionDiagnosis:   "Acute Appendicitis",
			DischargeDiagnosis:   "Post-appendectomy recovered",
			ConditionAtDischarge: "recovered",
			HospitalCourse:       "Patient underwent surgery smoothly. Normal vitals on discharge.",
			FollowUpDate:         &followDate,
		}
		if err := ds.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("invalid follow up date", func(t *testing.T) {
		badDate := "not-a-date"
		ds := domain.DischargeSummary{
			EncounterID:  "d3b07384-d113-4a7b-a621-000000000001",
			FollowUpDate: &badDate,
		}
		if err := ds.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestOdontogram_Validate(t *testing.T) {
	t.Run("valid tooth condition", func(t *testing.T) {
		in := domain.SetToothConditionInput{
			ToothNumber:    18,
			Condition:      "caries",
			ProcedureNotes: "Needs composite filling",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("invalid tooth number", func(t *testing.T) {
		in := domain.SetToothConditionInput{
			ToothNumber: 99,
			Condition:   "healthy",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid condition", func(t *testing.T) {
		in := domain.SetToothConditionInput{
			ToothNumber: 1,
			Condition:   "unknown_condition",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}
