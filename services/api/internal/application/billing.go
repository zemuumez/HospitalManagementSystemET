package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
	"time"
)

type BillingRepository interface {
	ChargeAccounts(context.Context, int) ([]domain.ChargeAccount, error)
	CreateChargeAccount(context.Context, domain.Actor, string) (domain.ChargeAccount, error)
	Invoices(context.Context, domain.Actor, int) ([]domain.Invoice, error)
	Invoice(context.Context, domain.Actor, string) (domain.Invoice, error)
	CreateInvoice(context.Context, domain.Actor, domain.InvoiceInput, int64, int64, string) (domain.Invoice, error)
	Payments(context.Context, domain.Actor, string) ([]domain.Payment, error)
	RecordPayment(context.Context, domain.Actor, string, domain.PaymentInput, string) (domain.Payment, error)
	BillingPatients(context.Context, domain.Actor, string, int) ([]domain.Patient, error)
}
type Billing struct {
	Store BillingRepository
	Now   func() time.Time
}

func (b Billing) Accounts(ctx context.Context, a domain.Actor, page int) ([]domain.ChargeAccount, error) {
	if !a.Can("billing.manage") {
		return nil, domain.ErrForbidden
	}
	if page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return b.Store.ChargeAccounts(ctx, page)
}
func (b Billing) CreateAccount(ctx context.Context, a domain.Actor, name string) (domain.ChargeAccount, error) {
	if !a.Can("billing.manage") {
		return domain.ChargeAccount{}, domain.ErrForbidden
	}
	name = strings.TrimSpace(name)
	if len([]rune(name)) < 1 || len([]rune(name)) > 100 {
		return domain.ChargeAccount{}, domain.ErrValidation
	}
	return b.Store.CreateChargeAccount(ctx, a, name)
}
func (b Billing) Invoices(ctx context.Context, a domain.Actor, page int) ([]domain.Invoice, error) {
	if !a.Can("billing.read") {
		return nil, domain.ErrForbidden
	}
	if page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return b.Store.Invoices(ctx, a, page)
}
func (b Billing) Invoice(ctx context.Context, a domain.Actor, id string) (domain.Invoice, error) {
	if !a.Can("billing.read") {
		return domain.Invoice{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.Invoice{}, domain.ErrValidation
	}
	return b.Store.Invoice(ctx, a, id)
}
func (b Billing) CreateInvoice(ctx context.Context, a domain.Actor, i domain.InvoiceInput, key string) (domain.Invoice, error) {
	if !a.Can("billing.manage") {
		return domain.Invoice{}, domain.ErrForbidden
	}
	if len(key) < 16 || len(key) > 80 {
		return domain.Invoice{}, domain.ErrValidation
	}
	i.Lines = append([]domain.InvoiceLine(nil), i.Lines...)
	subtotal, total, e := i.Totals(b.Now())
	if e != nil {
		return domain.Invoice{}, e
	}
	return b.Store.CreateInvoice(ctx, a, i, subtotal, total, key)
}
func (b Billing) Payments(ctx context.Context, a domain.Actor, id string) ([]domain.Payment, error) {
	if !a.Can("billing.read") {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return nil, domain.ErrValidation
	}
	return b.Store.Payments(ctx, a, id)
}
func (b Billing) RecordPayment(ctx context.Context, a domain.Actor, id string, p domain.PaymentInput, key string) (domain.Payment, error) {
	if !a.Can("billing.manage") {
		return domain.Payment{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || len(key) < 16 || len(key) > 80 {
		return domain.Payment{}, domain.ErrValidation
	}
	if e := p.Validate(); e != nil {
		return domain.Payment{}, e
	}
	return b.Store.RecordPayment(ctx, a, id, p, key)
}
func (b Billing) Patients(ctx context.Context, a domain.Actor, search string, page int) ([]domain.Patient, error) {
	if !a.Can("billing.manage") {
		return nil, domain.ErrForbidden
	}
	if len(search) > 80 || page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return b.Store.BillingPatients(ctx, a, search, page)
}
