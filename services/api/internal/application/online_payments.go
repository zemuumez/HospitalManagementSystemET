package application

import (
	"context"
	"hms.local/api/internal/domain"
	"time"
)

type PaymentProvider interface {
	Enabled() bool
	Intent(context.Context, domain.Checkout) (domain.GatewayIntent, error)
	Verify([]byte, string, time.Time) (domain.GatewayEvent, error)
}
type PaymentRepository interface {
	Checkout(context.Context, domain.Actor, string) (domain.Checkout, error)
	BindCheckout(context.Context, domain.Checkout, domain.GatewayIntent) error
	ApplyPaymentEvent(context.Context, domain.GatewayEvent) error
	PaymentReviews(context.Context, int) ([]domain.PaymentReview, error)
}
type OnlinePayments struct {
	Store    PaymentRepository
	Provider PaymentProvider
	Now      func() time.Time
}

func (p OnlinePayments) Checkout(ctx context.Context, a domain.Actor, id string) (domain.CheckoutResult, error) {
	var out domain.CheckoutResult
	if !a.Can("billing.read") {
		return out, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return out, domain.ErrValidation
	}
	if p.Provider == nil || !p.Provider.Enabled() {
		return out, domain.ErrUnavailable
	}
	c, e := p.Store.Checkout(ctx, a, id)
	if e != nil {
		return out, e
	}
	if c.State != "pending" && c.State != "ready" {
		return out, domain.ErrStale
	}
	if c.ProviderID == "" && p.Now().Sub(c.CreatedAt) > 23*time.Hour {
		return out, domain.ErrStale
	}
	intent, e := p.Provider.Intent(ctx, c)
	if e != nil {
		return out, e
	}
	if e = p.Store.BindCheckout(ctx, c, intent); e != nil {
		return out, e
	}
	c.ProviderID = intent.ID
	c.State = "ready"
	return domain.CheckoutResult{Checkout: c, ClientSecret: intent.ClientSecret}, nil
}
func (p OnlinePayments) Webhook(ctx context.Context, body []byte, signature string) error {
	if p.Provider == nil || !p.Provider.Enabled() {
		return domain.ErrUnavailable
	}
	event, e := p.Provider.Verify(body, signature, p.Now())
	if e != nil {
		return e
	}
	return p.Store.ApplyPaymentEvent(ctx, event)
}
func (p OnlinePayments) Reviews(ctx context.Context, a domain.Actor, page int) ([]domain.PaymentReview, error) {
	if !a.Can("billing.manage") {
		return nil, domain.ErrForbidden
	}
	if page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return p.Store.PaymentReviews(ctx, page)
}
