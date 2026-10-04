package domain

import (
	"strings"
	"time"
)

type ChargeAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type InvoiceLine struct {
	AccountID      string `json:"accountId"`
	AccountName    string `json:"accountName"`
	Description    string `json:"description"`
	Quantity       int64  `json:"quantity"`
	UnitPriceMinor int64  `json:"unitPriceMinor"`
}
type InvoiceInput struct {
	PatientID           string        `json:"patientId"`
	InvoiceDate         string        `json:"invoiceDate"`
	DiscountBasisPoints int64         `json:"discountBasisPoints"`
	Lines               []InvoiceLine `json:"lines"`
}

func (i *InvoiceInput) Totals(now time.Time) (int64, int64, error) {
	date, e := time.ParseInLocation("2006-01-02", i.InvoiceDate, HospitalLocation)
	if e != nil || date.After(now) || date.Before(now.AddDate(-1, 0, 0)) || !UUIDPattern.MatchString(i.PatientID) || i.DiscountBasisPoints < 0 || i.DiscountBasisPoints > 10000 || len(i.Lines) < 1 || len(i.Lines) > 100 {
		return 0, 0, ErrValidation
	}
	var subtotal int64
	for n := range i.Lines {
		l := &i.Lines[n]
		l.Description = strings.TrimSpace(l.Description)
		l.AccountName = ""
		if !UUIDPattern.MatchString(l.AccountID) || len([]rune(l.Description)) > 500 || l.Quantity < 1 || l.Quantity > 10000 || l.UnitPriceMinor < 0 || l.UnitPriceMinor > 1000000000 {
			return 0, 0, ErrValidation
		}
		subtotal += l.Quantity * l.UnitPriceMinor
		if subtotal > 1000000000000 {
			return 0, 0, ErrValidation
		}
	}
	// Round the percentage discount once to the nearest minor unit, half up.
	discount := (subtotal*i.DiscountBasisPoints + 5000) / 10000
	total := subtotal - discount
	if total < 1 {
		return 0, 0, ErrValidation
	}
	return subtotal, total, nil
}

type Invoice struct {
	ID                  string        `json:"id"`
	Number              int64         `json:"number"`
	PatientID           string        `json:"patientId"`
	PatientName         string        `json:"patientName"`
	InvoiceDate         string        `json:"invoiceDate"`
	SubtotalMinor       int64         `json:"subtotalMinor"`
	DiscountBasisPoints int64         `json:"discountBasisPoints"`
	TotalMinor          int64         `json:"totalMinor"`
	PaidMinor           int64         `json:"paidMinor"`
	Version             int           `json:"version"`
	Lines               []InvoiceLine `json:"lines"`
}
type PaymentInput struct {
	AmountMinor       int64  `json:"amountMinor"`
	Method            string `json:"method"`
	Reference         string `json:"reference"`
	OriginalPaymentID string `json:"originalPaymentId"`
	Reason            string `json:"reason"`
}

func (p *PaymentInput) Validate() error {
	p.Reference = strings.TrimSpace(p.Reference)
	p.Reason = strings.TrimSpace(p.Reason)
	if p.AmountMinor < 1 || p.AmountMinor > 1000000000000 || (p.Method != "cash" && p.Method != "bank") || len([]rune(p.Reference)) > 100 || (p.Method == "bank" && p.Reference == "") || len([]rune(p.Reason)) > 500 || (p.OriginalPaymentID != "" && (!UUIDPattern.MatchString(p.OriginalPaymentID) || p.Reason == "")) {
		return ErrValidation
	}
	return nil
}

type Payment struct {
	PaymentInput
	ID        string    `json:"id"`
	Direction string    `json:"direction"`
	CreatedAt time.Time `json:"createdAt"`
}

// A pharmacy invoice derives patient, quantity and price from a real dispensing.
type SourceInvoiceInput struct {
	AccountID           string `json:"accountId"`
	DiscountBasisPoints int64  `json:"discountBasisPoints"`
}

// PharmacyInvoiceInput retains compatibility with the existing dispensing contract.
type PharmacyInvoiceInput = SourceInvoiceInput
