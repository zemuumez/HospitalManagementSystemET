package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hms.local/api/internal/domain"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client does not log provider responses or client secrets. BaseURL is injectable for local tests.
type Client struct {
	Secret, WebhookSecret, BaseURL string
	Live                           bool
	HTTP                           *http.Client
}

func (c Client) Enabled() bool {
	return c.Secret != "" && c.WebhookSecret != "" && ((c.Live && strings.HasPrefix(c.Secret, "sk_live_")) || (!c.Live && strings.HasPrefix(c.Secret, "sk_test_")))
}
func (c Client) Intent(ctx context.Context, checkout domain.Checkout) (domain.GatewayIntent, error) {
	var out domain.GatewayIntent
	if !c.Enabled() {
		return out, domain.ErrUnavailable
	}
	base := c.BaseURL
	if base == "" {
		base = "https://api.stripe.com"
	}
	method := http.MethodPost
	path := "/v1/payment_intents"
	form := url.Values{"amount": {strconv.FormatInt(checkout.AmountMinor, 10)}, "currency": {"etb"}, "payment_method_types[]": {"card"}, "metadata[hms_checkout_id]": {checkout.ID}, "metadata[hms_invoice_id]": {checkout.InvoiceID}}
	if checkout.ProviderID != "" {
		method = http.MethodGet
		path += "/" + url.PathEscape(checkout.ProviderID)
		form = nil
	}
	req, e := http.NewRequestWithContext(ctx, method, base+path, strings.NewReader(form.Encode()))
	if e != nil {
		return out, domain.ErrUnavailable
	}
	req.SetBasicAuth(c.Secret, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if method == http.MethodPost {
		req.Header.Set("Idempotency-Key", "hms-checkout-"+checkout.ID)
	}
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	response, e := client.Do(req)
	if e != nil {
		return out, domain.ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return out, domain.ErrUnavailable
	}
	var v struct {
		ID       string            `json:"id"`
		Secret   string            `json:"client_secret"`
		Amount   int64             `json:"amount"`
		Currency string            `json:"currency"`
		Live     bool              `json:"livemode"`
		Metadata map[string]string `json:"metadata"`
	}
	if e = json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&v); e != nil {
		return out, domain.ErrUnavailable
	}
	if !strings.HasPrefix(v.ID, "pi_") || len(v.ID) > 100 || v.Secret == "" || v.Amount != checkout.AmountMinor || v.Currency != "etb" || v.Live != c.Live || v.Metadata["hms_checkout_id"] != checkout.ID || v.Metadata["hms_invoice_id"] != checkout.InvoiceID || (checkout.ProviderID != "" && checkout.ProviderID != v.ID) {
		return out, domain.ErrUnavailable
	}
	return domain.GatewayIntent{ID: v.ID, ClientSecret: v.Secret, AmountMinor: v.Amount, Currency: v.Currency, Live: v.Live}, nil
}
func (c Client) Verify(body []byte, header string, now time.Time) (domain.GatewayEvent, error) {
	var out domain.GatewayEvent
	if !c.Enabled() {
		return out, domain.ErrUnavailable
	}
	if len(body) > 64<<10 || len(header) > 4096 {
		return out, domain.ErrForbidden
	}
	var timestamp string
	signatures := []string{}
	for _, part := range strings.Split(header, ",") {
		pair := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(pair) != 2 {
			continue
		}
		if pair[0] == "t" {
			if timestamp != "" {
				return out, domain.ErrForbidden
			}
			timestamp = pair[1]
		}
		if pair[0] == "v1" {
			signatures = append(signatures, pair[1])
		}
	}
	stamp, e := strconv.ParseInt(timestamp, 10, 64)
	if e != nil || stamp < now.Unix()-300 || stamp > now.Unix()+300 {
		return out, domain.ErrForbidden
	}
	mac := hmac.New(sha256.New, []byte(c.WebhookSecret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(body)
	wanted := mac.Sum(nil)
	valid := false
	for _, signature := range signatures {
		raw, e := hex.DecodeString(signature)
		if e == nil && hmac.Equal(raw, wanted) {
			valid = true
		}
	}
	if !valid {
		return out, domain.ErrForbidden
	}
	var event struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Live bool   `json:"livemode"`
		Data struct {
			Object struct {
				ID       string            `json:"id"`
				Status   string            `json:"status"`
				Amount   int64             `json:"amount_received"`
				Currency string            `json:"currency"`
				Live     bool              `json:"livemode"`
				Metadata map[string]string `json:"metadata"`
			} `json:"object"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &event) != nil || !strings.HasPrefix(event.ID, "evt_") || len(event.ID) > 100 || len(event.Type) > 200 || event.Live != c.Live {
		return out, domain.ErrForbidden
	}
	digest := sha256.Sum256(body)
	out = domain.GatewayEvent{ID: event.ID, Kind: event.Type, Hash: hex.EncodeToString(digest[:]), Live: event.Live}
	if event.Type != "payment_intent.succeeded" {
		return out, nil
	}
	v := event.Data.Object
	if !strings.HasPrefix(v.ID, "pi_") || len(v.ID) > 100 || v.Status != "succeeded" || v.Live != c.Live || v.Amount < 1 {
		return out, domain.ErrValidation
	}
	out.IntentID = v.ID
	out.AmountMinor = v.Amount
	out.Currency = v.Currency
	out.CheckoutID = v.Metadata["hms_checkout_id"]
	out.InvoiceID = v.Metadata["hms_invoice_id"]
	return out, nil
}
