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
	base, discountAmt, total := CalculateInsuranceTotals(5000, 15000, 10, diseases)
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
	base, discountAmt, total = CalculateInsuranceTotals(125, 0, 15, nil)
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
	base, discountAmt, total = CalculateInsuranceTotals(5000, 10000, 0, diseases)
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
