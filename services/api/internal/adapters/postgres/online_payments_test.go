package postgres

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/adapters/stripe"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type testGateway struct{}

func (testGateway) Enabled() bool { return true }
func (testGateway) Intent(ctx context.Context, c domain.Checkout) (domain.GatewayIntent, error) {
	return domain.GatewayIntent{ID: "pi_" + c.ID, ClientSecret: "local_fake_secret", AmountMinor: c.AmountMinor, Currency: "etb"}, nil
}
func (testGateway) Verify([]byte, string, time.Time) (domain.GatewayEvent, error) {
	return domain.GatewayEvent{}, domain.ErrForbidden
}
func testOnlinePayments(t *testing.T, db *pgxpool.Pool, store Store, actors []domain.Actor, patients []domain.Patient) {
	t.Helper()
	ctx := context.Background()
	b := application.Billing{Store: store, Now: time.Now}
	p := application.OnlinePayments{Store: store, Provider: testGateway{}, Now: time.Now}
	account, e := b.CreateAccount(ctx, actors[0], "Online payments test")
	if e != nil {
		t.Fatal(e)
	}
	makeInvoice := func(key string) domain.Invoice {
		t.Helper()
		v, e := b.CreateInvoice(ctx, actors[0], domain.InvoiceInput{PatientID: patients[0].ID, InvoiceDate: time.Now().In(domain.HospitalLocation).Format("2006-01-02"), Lines: []domain.InvoiceLine{{AccountID: account.ID, Quantity: 1, UnitPriceMinor: 10000}}}, key)
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	invoice := makeInvoice("online-invoice-test")
	if _, e = p.Checkout(ctx, actors[1], invoice.ID); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("doctor online finance", e)
	}
	otherPatient := domain.Actor{ID: "other-patient", Role: "patient"}
	if _, e = p.Checkout(ctx, otherPatient, invoice.ID); !errors.Is(e, domain.ErrNotFound) {
		t.Fatal("cross patient checkout", e)
	}
	var wg sync.WaitGroup
	var checkouts [2]domain.CheckoutResult
	var errs [2]error
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) { defer wg.Done(); checkouts[n], errs[n] = p.Checkout(ctx, actors[3], invoice.ID) }(n)
	}
	wg.Wait()
	if errs[0] != nil || errs[1] != nil || checkouts[0].ID != checkouts[1].ID {
		t.Fatal("duplicate checkout", errs)
	}
	c := checkouts[0]
	event := domain.GatewayEvent{ID: "evt_local_payment", Kind: "payment_intent.succeeded", IntentID: c.ProviderID, CheckoutID: c.ID, InvoiceID: invoice.ID, AmountMinor: 10000, Currency: "etb", Hash: "synthetic-hash"}
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) { defer wg.Done(); errs[n] = store.ApplyPaymentEvent(ctx, event) }(n)
	}
	wg.Wait()
	if errs[0] != nil || errs[1] != nil {
		t.Fatal("event replay race", errs)
	}
	event.ID = "evt_second_same_intent"
	event.Hash = "another-hash"
	if e = store.ApplyPaymentEvent(ctx, event); e != nil {
		t.Fatal(e)
	}
	saved, e := b.Invoice(ctx, actors[3], invoice.ID)
	if e != nil || saved.PaidMinor != 10000 {
		t.Fatal("online ledger amount", saved, e)
	}
	var count int
	if e = db.QueryRow(ctx, `SELECT count(*) FROM invoice_payment WHERE invoice_id=$1`, invoice.ID).Scan(&count); e != nil || count != 1 {
		t.Fatal("duplicate provider payment", count, e)
	}
	event.Hash = "changed"
	if e = store.ApplyPaymentEvent(ctx, event); !errors.Is(e, domain.ErrConflict) {
		t.Fatal("changed event replay", e)
	}
	if _, e = db.Exec(ctx, `DELETE FROM payment_event WHERE event_id=$1`, event.ID); e == nil {
		t.Fatal("provider evidence removed")
	}
	// Provider success after a manual payment must go to reconciliation, never overpay the ledger.
	invoice2 := makeInvoice("online-review-invoice")
	c2, e := p.Checkout(ctx, actors[0], invoice2.ID)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = b.RecordPayment(ctx, actors[0], invoice2.ID, domain.PaymentInput{AmountMinor: 5000, Method: "cash"}, "online-manual-partial"); e != nil {
		t.Fatal(e)
	}
	event = domain.GatewayEvent{ID: "evt_overpayment_review", Kind: "payment_intent.succeeded", IntentID: c2.ProviderID, CheckoutID: c2.ID, InvoiceID: invoice2.ID, AmountMinor: 10000, Currency: "etb", Hash: "review-hash"}
	if e = store.ApplyPaymentEvent(ctx, event); e != nil {
		t.Fatal(e)
	}
	saved, e = b.Invoice(ctx, actors[0], invoice2.ID)
	if e != nil || saved.PaidMinor != 5000 {
		t.Fatal("provider overpayment posted", e)
	}
	reviews, e := p.Reviews(ctx, actors[0], 1)
	if e != nil || len(reviews) != 1 {
		t.Fatal("missing reconciliation event", reviews, e)
	}
	if _, e = p.Reviews(ctx, actors[3], 1); !errors.Is(e, domain.ErrForbidden) {
		t.Fatal("patient read reconciliation", e)
	}
	disabled := application.OnlinePayments{Store: store, Now: time.Now}
	if _, e = disabled.Checkout(ctx, actors[0], invoice.ID); !errors.Is(e, domain.ErrUnavailable) {
		t.Fatal("blank gateway enabled", e)
	}

	invoice3 := makeInvoice("online-http-invoice")
	c3, e := p.Checkout(ctx, actors[0], invoice3.ID)
	if e != nil {
		t.Fatal(e)
	}
	body, _ := json.Marshal(map[string]any{"id": "evt_http_local", "type": "payment_intent.succeeded", "livemode": false, "data": map[string]any{"object": map[string]any{"id": c3.ProviderID, "status": "succeeded", "amount_received": 10000, "currency": "etb", "livemode": false, "metadata": map[string]string{"hms_checkout_id": c3.ID, "hms_invoice_id": invoice3.ID}}}})
	secret := "whsec_local_only"
	stamp := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stamp + "."))
	mac.Write(body)
	signature := "t=" + stamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))
	handler := httpapi.Server{OnlinePayments: application.OnlinePayments{Store: store, Provider: stripe.Client{Secret: "sk_test_local_only", WebhookSecret: secret}, Now: time.Now}}.Handler()
	for _, tc := range []struct {
		signature, body string
		want            int
	}{{"", string(body), 403}, {signature, string(body), 200}, {signature, string(body), 200}, {signature, strings.Repeat("x", 65537), 413}} {
		req := httptest.NewRequest("POST", "/v1/webhooks/stripe", strings.NewReader(tc.body))
		req.Header.Set("Stripe-Signature", tc.signature)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Fatalf("webhook HTTP wanted %d got %d: %s", tc.want, w.Code, w.Body.String())
		}
	}
	saved, e = b.Invoice(ctx, actors[0], invoice3.ID)
	if e != nil || saved.PaidMinor != 10000 {
		t.Fatal("signed webhook did not post", e)
	}
}
