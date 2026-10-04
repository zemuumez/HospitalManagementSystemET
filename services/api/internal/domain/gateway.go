package domain

import (
	"errors"
	"time"
)

var ErrUnavailable = errors.New("dependency unavailable")

type Checkout struct {
	ID          string    `json:"id"`
	InvoiceID   string    `json:"invoiceId"`
	ActorID     string    `json:"-"`
	AmountMinor int64     `json:"amountMinor"`
	ProviderID  string    `json:"providerId"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"createdAt"`
}
type CheckoutResult struct {
	Checkout
	ClientSecret string `json:"clientSecret"`
}
type GatewayIntent struct {
	ID, ClientSecret string
	AmountMinor      int64
	Currency         string
	Live             bool
}
type GatewayEvent struct {
	ID, Kind, IntentID, CheckoutID, InvoiceID, Currency, Hash string
	AmountMinor                                               int64
	Live                                                      bool
}
type PaymentReview struct {
	EventID    string    `json:"eventId"`
	IntentID   string    `json:"intentId"`
	State      string    `json:"state"`
	Reason     string    `json:"reason"`
	ReceivedAt time.Time `json:"receivedAt"`
}
