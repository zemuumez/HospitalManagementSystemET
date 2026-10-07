package domain

import (
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
	subtotal, discountAmt, total, lines := CalculatePackageTotals(15, services)
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
	subtotal, discountAmt, total, _ = CalculatePackageTotals(0, services)
	if subtotal != 55000 || discountAmt != 0 || total != 55000 {
		t.Fatalf("unexpected total with 0%% discount: sub=%d, disc=%d, tot=%d", subtotal, discountAmt, total)
	}

	// 100% discount
	subtotal, discountAmt, total, _ = CalculatePackageTotals(100, services)
	if total != 0 || discountAmt != 55000 {
		t.Fatalf("unexpected total with 100%% discount: tot=%d, disc=%d", total, discountAmt)
	}

	// Fractional rounding test:
	// Subtotal = 333 cents, discount = 10%.
	// (333 * 10 + 50) / 100 = 3380 / 100 = 33 cents discount.
	// Total = 333 - 33 = 300 cents.
	fractionalServices := []PackageServiceLineInput{
		{ServiceID: "s1", Quantity: 1, RateMinor: 333},
	}
	subtotal, discountAmt, total, _ = CalculatePackageTotals(10, fractionalServices)
	if subtotal != 333 || discountAmt != 33 || total != 300 {
		t.Fatalf("unexpected rounding: sub=%d, disc=%d, tot=%d", subtotal, discountAmt, total)
	}
}
