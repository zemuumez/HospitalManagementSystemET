package domain_test

import (
	"hms.local/api/internal/domain"
	"testing"
	"time"
)

func TestAmbulanceValidation(t *testing.T) {
	valid := domain.AmbulanceInput{
		VehicleNumber: "ETH-AMB-01",
		VehicleModel:  "Toyota HiAce",
		YearMade:      "2023",
		DriverName:    "Abebe Kebede",
		DriverLicense: "DL-12345",
		DriverContact: "+251911223344",
		VehicleType:   2,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid ambulance, got %v", err)
	}

	invalidYear := valid
	invalidYear.YearMade = "23"
	if err := invalidYear.Validate(); err == nil {
		t.Fatalf("expected error for invalid year made")
	}

	invalidType := valid
	invalidType.VehicleType = 3
	if err := invalidType.Validate(); err == nil {
		t.Fatalf("expected error for invalid vehicle type")
	}

	emptyNum := valid
	emptyNum.VehicleNumber = ""
	if err := emptyNum.Validate(); err == nil {
		t.Fatalf("expected error for empty vehicle number")
	}
}

func TestAmbulanceCallValidation(t *testing.T) {
	valid := domain.AmbulanceCallInput{
		AmbulanceID: "11111111-1111-1111-1111-111111111111",
		PatientID:   "22222222-2222-2222-2222-222222222222",
		DriverName:  "Abebe Kebede",
		CallDate:    time.Now(),
		AmountMinor: 150000,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid call input, got %v", err)
	}

	invalidAmount := valid
	invalidAmount.AmountMinor = -10
	if err := invalidAmount.Validate(); err == nil {
		t.Fatalf("expected error for negative amount")
	}

	invalidAmbulance := valid
	invalidAmbulance.AmbulanceID = "not-a-uuid"
	if err := invalidAmbulance.Validate(); err == nil {
		t.Fatalf("expected error for invalid ambulance ID")
	}

	updateValid := domain.AmbulanceCallUpdateInput{
		AmbulanceID: "11111111-1111-1111-1111-111111111111",
		CallDate:    time.Now(),
		AmountMinor: 150000,
		Status:      "completed",
		Version:     1,
	}
	if err := updateValid.Validate(); err != nil {
		t.Fatalf("expected valid update input, got %v", err)
	}

	invalidStatus := updateValid
	invalidStatus.Status = "flying"
	if err := invalidStatus.Validate(); err == nil {
		t.Fatalf("expected error for invalid status")
	}
}
