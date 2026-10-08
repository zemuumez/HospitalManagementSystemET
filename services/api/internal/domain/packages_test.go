package domain

import (
	"encoding/json"
	"testing"
)

func TestPackageValidation(t *testing.T) {
	// Valid input
	valid := PackageInput{
		Name:        "General Checkup Package",
		Description: "Standard medical package",
		Discount:    10,
		Services: []PackageServiceLineInput{
			{ServiceID: "svc-1", Quantity: 1, RateMinor: 25000},
			{ServiceID: "svc-2", Quantity: 2, RateMinor: 10000},
		},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid package, got %v", err)
	}

	// Empty name
	p := valid
	p.Name = "   "
	if err := p.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for empty name, got %v", err)
	}

	// Name too long (> 160 chars)
	p = valid
	p.Name = string(make([]byte, 161))
	if err := p.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for long name, got %v", err)
	}

	// Negative discount
	p = valid
	p.Discount = -1
	if err := p.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for negative discount, got %v", err)
	}

	// Discount > 100
	p = valid
	p.Discount = 101
	if err := p.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for discount > 100, got %v", err)
	}

	// Empty services
	p = valid
	p.Services = []PackageServiceLineInput{}
	if err := p.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for empty services, got %v", err)
	}

	// Zero quantity
	p = valid
	p.Services = []PackageServiceLineInput{{ServiceID: "svc-1", Quantity: 0, RateMinor: 1000}}
	if err := p.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for quantity 0, got %v", err)
	}

	// Negative rate
	p = valid
	p.Services = []PackageServiceLineInput{{ServiceID: "svc-1", Quantity: 1, RateMinor: -500}}
	if err := p.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for negative rate, got %v", err)
	}

	// Duplicate service ID
	p = valid
	p.Services = []PackageServiceLineInput{
		{ServiceID: "svc-1", Quantity: 1, RateMinor: 1000},
		{ServiceID: "svc-1", Quantity: 2, RateMinor: 2000},
	}
	if err := p.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for duplicate service, got %v", err)
	}
}

func TestCalculatePackageTotals(t *testing.T) {
	// Subtotal = 1*25000 + 2*15000 = 55000. Discount = 15%.
	// 55000 * 15 = 825000. (825000 + 50) / 100 = 8250.
	// Total = 55000 - 8250 = 46750.
	services := []PackageServiceLineInput{
		{ServiceID: "s1", Quantity: 1, RateMinor: 25000},
		{ServiceID: "s2", Quantity: 2, RateMinor: 15000},
	}
	subtotal, discountAmt, total, lines, err := CalculatePackageTotals(15, services)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subtotal != 55000 {
		t.Fatalf("expected subtotal 55000, got %d", subtotal)
	}
	if discountAmt != 8250 {
		t.Fatalf("expected discountAmt 8250, got %d", discountAmt)
	}
	if total != 46750 {
		t.Fatalf("expected total 46750, got %d", total)
	}
	if len(lines) != 2 || lines[0] != 25000 || lines[1] != 30000 {
		t.Fatalf("unexpected line amounts: %v", lines)
	}

	// Zero discount
	subtotal, discountAmt, total, _, err = CalculatePackageTotals(0, services)
	if err != nil || subtotal != 55000 || discountAmt != 0 || total != 55000 {
		t.Fatalf("unexpected total with 0%% discount: sub=%d, disc=%d, tot=%d, err=%v", subtotal, discountAmt, total, err)
	}

	// 100% discount
	subtotal, discountAmt, total, _, err = CalculatePackageTotals(100, services)
	if err != nil || total != 0 || discountAmt != 55000 {
		t.Fatalf("unexpected total with 100%% discount: tot=%d, disc=%d, err=%v", total, discountAmt, err)
	}

	// Fractional rounding test:
	// Subtotal = 333 cents, discount = 10%.
	// (333 * 10 + 50) / 100 = 3380 / 100 = 33 cents discount.
	// Total = 333 - 33 = 300 cents.
	fractionalServices := []PackageServiceLineInput{
		{ServiceID: "s1", Quantity: 1, RateMinor: 333},
	}
	subtotal, discountAmt, total, _, err = CalculatePackageTotals(10, fractionalServices)
	if err != nil || subtotal != 333 || discountAmt != 33 || total != 300 {
		t.Fatalf("unexpected rounding: sub=%d, disc=%d, tot=%d, err=%v", subtotal, discountAmt, total, err)
	}
}

func TestPackageInput_DuplicateChildID_RegressionR1(t *testing.T) {
	in := PackageInput{
		Name:     "Duplicate Child ID Package",
		Discount: 0,
		Services: []PackageServiceLineInput{
			{ID: "child-line-1", ServiceID: "svc-1", Quantity: 1, RateMinor: 100},
			{ID: "child-line-1", ServiceID: "svc-2", Quantity: 1, RateMinor: 200},
		},
	}
	if err := in.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for duplicate child line ID, got %v", err)
	}
}

func TestPackageTotals_Overflow_RegressionR2(t *testing.T) {
	// Probe: base = 100000000000000000 with 100% discount
	// Previously overflowed int64 producing total 184467440737095515
	services := []PackageServiceLineInput{
		{ServiceID: "s1", Quantity: 1, RateMinor: 100000000000000000},
	}
	_, _, total, _, err := CalculatePackageTotals(100, services)
	if err == nil && total != 0 {
		t.Fatalf("expected error or total 0 for 100%% discount, got total=%d err=%v", total, err)
	}

	in := PackageInput{
		Name:     "Overflow Package",
		Discount: 100,
		Services: services,
	}
	if err := in.Validate(); err != ErrValidation {
		t.Fatalf("expected ErrValidation for amount exceeding MaxMoneyMinor, got %v", err)
	}

	// Maximum boundary within limits
	maxSafeServices := []PackageServiceLineInput{
		{ServiceID: "s1", Quantity: 1, RateMinor: MaxMoneyMinor},
	}
	subtotal, disc, tot, _, err := CalculatePackageTotals(100, maxSafeServices)
	if err != nil {
		t.Fatalf("expected MaxMoneyMinor calculation to succeed without overflow, got %v", err)
	}
	if subtotal != MaxMoneyMinor || disc != MaxMoneyMinor || tot != 0 {
		t.Fatalf("unexpected totals at boundary: sub=%d disc=%d tot=%d", subtotal, disc, tot)
	}
}

func TestPackageServiceLineInput_ExplicitZeroRate_RegressionR3(t *testing.T) {
	data := []byte(`{"service_id": "svc-1", "quantity": 1, "rate_minor": 0}`)
	var line PackageServiceLineInput
	if err := json.Unmarshal(data, &line); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !line.HasRate {
		t.Fatalf("expected HasRate to be true for explicit rate_minor 0")
	}
	if line.RateMinor != 0 {
		t.Fatalf("expected RateMinor 0, got %d", line.RateMinor)
	}

	omittedData := []byte(`{"service_id": "svc-1", "quantity": 1}`)
	var omittedLine PackageServiceLineInput
	if err := json.Unmarshal(omittedData, &omittedLine); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if omittedLine.HasRate {
		t.Fatalf("expected HasRate to be false when rate is omitted")
	}
}
