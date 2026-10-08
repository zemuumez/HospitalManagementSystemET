package domain

import (
	"testing"
)

func TestCalculateInsuranceTotals(t *testing.T) {
	// Example: service tax 50.00 ETB (5000 minor), hospital rate 150.00 ETB (15000 minor)
	// Disease 1: 45.00 ETB (4500 minor), Disease 2: 55.00 ETB (5500 minor)
	// Base = 5000 + 15000 + 4500 + 5500 = 30000 minor (300.00 ETB)
	// Discount = 10% -> 3000 minor (30.00 ETB)
	// Total = 27000 minor (270.00 ETB)
	diseases := []InsuranceDiseaseLineInput{
		{DiseaseName: "Malaria", DiseaseChargeMinor: 4500},
		{DiseaseName: "Typhoid", DiseaseChargeMinor: 5500},
	}
	base, discountAmt, total, err := CalculateInsuranceTotals(5000, 15000, 10, diseases)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base != 30000 {
		t.Fatalf("expected base 30000, got %d", base)
	}
	if discountAmt != 3000 {
		t.Fatalf("expected discount amount 3000, got %d", discountAmt)
	}
	if total != 27000 {
		t.Fatalf("expected total 27000, got %d", total)
	}

	// Test half-up rounding:
	// Base = 125 minor (1.25 ETB), Discount = 15%
	// Exact discount = 18.75 minor -> rounds to 19 minor
	// Total = 125 - 19 = 106 minor
	base, discountAmt, total, err = CalculateInsuranceTotals(125, 0, 15, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base != 125 {
		t.Fatalf("expected base 125, got %d", base)
	}
	if discountAmt != 19 {
		t.Fatalf("expected rounded discount 19, got %d", discountAmt)
	}
	if total != 106 {
		t.Fatalf("expected total 106, got %d", total)
	}

	// Test 0% discount
	base, discountAmt, total, err = CalculateInsuranceTotals(5000, 10000, 0, diseases)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if discountAmt != 0 {
		t.Fatalf("expected 0 discount, got %d", discountAmt)
	}
	if total != base {
		t.Fatalf("expected total %d == base %d, got %d", base, base, total)
	}
}

func TestInsuranceInputNormalizeAndValidate(t *testing.T) {
	taxFloat := 45.50
	rateFloat := 120.25
	chargeFloat := 85.75

	in := InsuranceInput{
		Name:             "  Ethiopian Health Insurance Agency  ",
		AltInsuranceNo:   "EHIA-9921",
		AltInsuranceCode: "EHIA-ETH",
		AltServiceTax:    &taxFloat,
		AltHospitalRate:  &rateFloat,
		Discount:         12,
		Diseases: []InsuranceDiseaseLineInput{
			{
				AltDiseaseName:   "Pneumonia",
				AltDiseaseCharge: &chargeFloat,
			},
		},
	}

	if err := in.Validate(); err != nil {
		t.Fatalf("expected valid normalized input, got error: %v", err)
	}

	if in.Name != "Ethiopian Health Insurance Agency" {
		t.Fatalf("expected trimmed name, got %q", in.Name)
	}
	if in.InsuranceNo != "EHIA-9921" {
		t.Fatalf("expected insurance no EHIA-9921, got %q", in.InsuranceNo)
	}
	if in.InsuranceCode != "EHIA-ETH" {
		t.Fatalf("expected insurance code EHIA-ETH, got %q", in.InsuranceCode)
	}
	if in.ServiceTaxMinor != 4550 {
		t.Fatalf("expected service tax minor 4550, got %d", in.ServiceTaxMinor)
	}
	if in.HospitalRateMinor != 12025 {
		t.Fatalf("expected hospital rate minor 12025, got %d", in.HospitalRateMinor)
	}
	if len(in.Diseases) != 1 || in.Diseases[0].DiseaseName != "Pneumonia" || in.Diseases[0].DiseaseChargeMinor != 8575 {
		t.Fatalf("expected normalized disease line with 8575 minor charge, got %+v", in.Diseases)
	}

	// Rejections
	emptyName := in
	emptyName.Name = ""
	if emptyName.Validate() == nil {
		t.Fatal("expected error on empty name")
	}

	badDiscount := in
	badDiscount.Discount = 101
	if badDiscount.Validate() == nil {
		t.Fatal("expected error on discount > 100")
	}

	noDiseases := in
	noDiseases.Diseases = nil
	if noDiseases.Validate() == nil {
		t.Fatal("expected error on empty diseases")
	}
}

func TestInsuranceTotals_Overflow_RegressionR2(t *testing.T) {
	// Probe: base = 100000000000000000 with 100% discount
	// Previously overflowed int64 producing total 184467440737095515
	diseases := []InsuranceDiseaseLineInput{
		{DiseaseName: "Malaria", DiseaseChargeMinor: 100000000000000000},
	}
	_, _, total, err := CalculateInsuranceTotals(0, 0, 100, diseases)
	if err == nil && total != 0 {
		t.Fatalf("expected error or total 0 for 100%% discount, got total=%d err=%v", total, err)
	}

	in := InsuranceInput{
		Name:          "Overflow Insurance",
		InsuranceNo:   "INS-OVF",
		InsuranceCode: "OVF",
		Discount:      100,
		Diseases:      diseases,
	}
	if err := in.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for overflow amount, got %v", err)
	}

	// Maximum boundary within limits
	maxSafeDiseases := []InsuranceDiseaseLineInput{
		{DiseaseName: "D1", DiseaseChargeMinor: MaxMoneyMinor},
	}
	base, disc, tot, err := CalculateInsuranceTotals(0, 0, 100, maxSafeDiseases)
	if err != nil {
		t.Fatalf("expected MaxMoneyMinor calculation to succeed without overflow, got %v", err)
	}
	if base != MaxMoneyMinor || disc != MaxMoneyMinor || tot != 0 {
		t.Fatalf("unexpected totals at boundary: base=%d disc=%d tot=%d", base, disc, tot)
	}
}

func TestInsuranceInput_InvalidStatusAndDuplicateID(t *testing.T) {
	statusBad := 2
	in := InsuranceInput{
		Name:          "Bad Status Insurance",
		InsuranceNo:   "INS-BAD",
		InsuranceCode: "BAD",
		Status:        &statusBad,
		Diseases: []InsuranceDiseaseLineInput{
			{DiseaseName: "D1", DiseaseChargeMinor: 1000},
		},
	}
	if err := in.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for status %d, got %v", statusBad, err)
	}

	statusOK := 1
	in.Status = &statusOK
	in.Diseases = []InsuranceDiseaseLineInput{
		{ID: "line-1", DiseaseName: "D1", DiseaseChargeMinor: 1000},
		{ID: "line-1", DiseaseName: "D2", DiseaseChargeMinor: 2000},
	}
	if err := in.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for duplicate disease line ID, got %v", err)
	}
}
