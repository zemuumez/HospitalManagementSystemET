package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hms.local/api/internal/domain"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func signed(body []byte, secret string, at time.Time) string {
	stamp := strconv.FormatInt(at.Unix(), 10)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stamp + "."))
	h.Write(body)
	return "t=" + stamp + ",v1=" + hex.EncodeToString(h.Sum(nil))
}
func TestWebhookVerification(t *testing.T) {
	now := time.Now()
	c := Client{Secret: "sk_test_local_only", WebhookSecret: "whsec_local_test_only"}
	body := []byte(`{"id":"evt_local","type":"payment_intent.succeeded","livemode":false,"data":{"object":{"id":"pi_local","status":"succeeded","amount_received":100,"currency":"etb","livemode":false,"metadata":{"hms_checkout_id":"00000000-0000-0000-0000-000000000001","hms_invoice_id":"00000000-0000-0000-0000-000000000002"}}}}`)
	event, e := c.Verify(body, signed(body, c.WebhookSecret, now), now)
	if e != nil || event.AmountMinor != 100 || event.IntentID != "pi_local" {
		t.Fatal(event, e)
	}
	for _, header := range []string{"", signed(body, "wrong", now), signed(body, c.WebhookSecret, now.Add(-6*time.Minute)), signed(body, c.WebhookSecret, now.Add(6*time.Minute)), signed(body, c.WebhookSecret, now) + ",t=1"} {
		if _, e = c.Verify(body, header, now); e == nil {
			t.Fatal("invalid signature accepted")
		}
	}
	altered := append(append([]byte{}, body...), byte(' '))
	if _, e = c.Verify(altered, signed(body, c.WebhookSecret, now), now); e == nil {
		t.Fatal("body mutation accepted")
	}
	live := []byte(strings.ReplaceAll(string(body), `"livemode":false`, `"livemode":true`))
	if _, e = c.Verify(live, signed(live, c.WebhookSecret, now), now); e == nil {
		t.Fatal("wrong mode accepted")
	}
	if (Client{}).Enabled() {
		t.Fatal("blank provider enabled")
	}
}
func TestIntentRequest(t *testing.T) {
	checkout := domain.Checkout{ID: "00000000-0000-0000-0000-000000000001", InvoiceID: "00000000-0000-0000-0000-000000000002", AmountMinor: 12345}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _, ok := r.BasicAuth()
		if !ok || user != "sk_test_local_only" {
			t.Error("missing auth")
		}
		if r.Method == "POST" {
			if e := r.ParseForm(); e != nil {
				t.Error(e)
			}
			if r.Form.Get("amount") != "12345" || r.Form.Get("currency") != "etb" || r.Form.Get("confirm") != "" || r.Header.Get("Idempotency-Key") != "hms-checkout-"+checkout.ID {
				t.Error("unsafe provider request")
			}
		}
		if r.Method == "GET" && r.URL.Path != "/v1/payment_intents/pi_local" {
			t.Error("invalid retrieve path")
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "pi_local", "client_secret": "synthetic_client_secret", "amount": 12345, "currency": "etb", "livemode": false, "metadata": map[string]string{"hms_checkout_id": checkout.ID, "hms_invoice_id": checkout.InvoiceID}})
	}))
	defer server.Close()
	c := Client{Secret: "sk_test_local_only", WebhookSecret: "whsec_local_test_only", BaseURL: server.URL, HTTP: server.Client()}
	result, e := c.Intent(context.Background(), checkout)
	if e != nil || result.ID != "pi_local" {
		t.Fatal(result, e)
	}
	checkout.ProviderID = result.ID
	if _, e = c.Intent(context.Background(), checkout); e != nil {
		t.Fatal(e)
	}
}
