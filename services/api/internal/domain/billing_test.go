package domain

import (
	"testing"
	"time"
)

func TestInvoiceExactTotals(t *testing.T) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	i := InvoiceInput{PatientID: "00000000-0000-4000-8000-000000000001", InvoiceDate: "2026-10-04", DiscountBasisPoints: 1000, Lines: []InvoiceLine{{AccountID: "00000000-0000-4000-8000-000000000002", Quantity: 3, UnitPriceMinor: 333}}}
	sub, total, e := i.Totals(now)
	if e != nil || sub != 999 || total != 899 {
		t.Fatalf("round once half up: %d %d %v", sub, total, e)
	}
	i.DiscountBasisPoints = 10000
	if _, _, e = i.Totals(now); e == nil {
		t.Fatal("zero invoice accepted")
	}
	i.DiscountBasisPoints = 0
	i.Lines[0].Quantity = 10001
	if _, _, e = i.Totals(now); e == nil {
		t.Fatal("quantity limit missing")
	}
}
