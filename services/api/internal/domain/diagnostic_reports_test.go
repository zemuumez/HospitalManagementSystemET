package domain_test

import (
	"testing"
	"time"

	"hms.local/api/internal/domain"
)

func TestDiagnosticCategoryInput_Validate(t *testing.T) {
	t.Run("valid category", func(t *testing.T) {
		in := domain.DiagnosticCategoryInput{
			Name:        "Molecular Biology",
			Kind:        "pathology",
			Description: "PCR and genetic testing",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("invalid kind", func(t *testing.T) {
		in := domain.DiagnosticCategoryInput{
			Name: "Cardiology",
			Kind: "other",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestDiagnosticUnitInput_Validate(t *testing.T) {
	t.Run("valid unit", func(t *testing.T) {
		in := domain.DiagnosticUnitInput{
			Name:        "ng/mL",
			Description: "Nanograms per milliliter",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		in := domain.DiagnosticUnitInput{
			Name: "   ",
		}
		if err := in.Validate(); err != domain.ErrValidation {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	})
}

func TestVaccineAndAdminister_Validate(t *testing.T) {
	t.Run("valid vaccine", func(t *testing.T) {
		v := domain.VaccineInput{
			Name:             "COVID-19 mRNA",
			TargetDisease:    "SARS-CoV-2",
			RecommendedDoses: 2,
			MinAgeMonths:     6,
			Instructions:     "Intramuscular injection",
		}
		if err := v.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("valid administration", func(t *testing.T) {
		exp := "2027-12-31"
		in := domain.AdministerVaccineInput{
			PatientID:  "d3b07384-d113-4a7b-a621-000000000001",
			VaccineID:  "d3b07384-d113-4a7b-a621-000000000002",
			DoseNumber: 1,
			LotNumber:  "LOT-9988",
			ExpiryDate: &exp,
			Notes:      "Well tolerated",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}

func TestVitalReports_Validate(t *testing.T) {
	t.Run("valid birth report", func(t *testing.T) {
		in := domain.CreateBirthReportInput{
			ChildName:   "Abebech Tadesse",
			Gender:      "female",
			BirthDate:   time.Now().UTC(),
			WeightKg:    3.25,
			MotherName:  "Sara Bekele",
			FatherName:  "Tadesse Alemu",
			DeliveredBy: "doc-101",
			Notes:       "Normal spontaneous vaginal delivery",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("valid death report", func(t *testing.T) {
		in := domain.CreateDeathReportInput{
			PatientID:            "d3b07384-d113-4a7b-a621-000000000001",
			DeathDate:            time.Now().UTC(),
			CauseOfDeath:         "Cardiopulmonary arrest secondary to acute myocardial infarction",
			CertifiedBy:          "doc-101",
			GuardianAcknowledged: "Brother Kebede",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("valid operation report", func(t *testing.T) {
		in := domain.CreateOperationReportInput{
			EncounterID:            "d3b07384-d113-4a7b-a621-000000000001",
			PatientID:              "d3b07384-d113-4a7b-a621-000000000002",
			OperationName:          "Open Cholecystectomy",
			SurgeonID:              "doc-202",
			AssistantSurgeon:       "doc-303",
			Anesthetist:            "doc-404",
			AnesthesiaType:         "General",
			OperationDate:          time.Now().UTC(),
			PreOperativeDiagnosis:  "Symptomatic Cholelithiasis",
			PostOperativeDiagnosis: "Chronic Cholecystitis with Cholelithiasis",
			ProcedureTechnique:     "Standard subcostal incision, cystic duct and artery ligated",
			Findings:               "Distended gallbladder with multiple gallstones",
		}
		if err := in.Validate(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}
