package delivery

import (
	"context"
	"testing"
)

func TestDevelopmentTransportGuards(t *testing.T) {
	t.Setenv("SMS_PROVIDER", "capture")
	t.Setenv("APP_ENV", "development")
	r, e := Send(context.Background(), "sms", "+251911111111", "", "Synthetic")
	if e != nil || r.Status != "captured" {
		t.Fatal(r, e)
	}
	t.Setenv("APP_ENV", "production")
	r, e = Send(context.Background(), "sms", "+251911111111", "", "Synthetic")
	if e == nil || r.Status != "failed" {
		t.Fatal("production capture accepted", r, e)
	}
	t.Setenv("SMS_PROVIDER", "twilio")
	t.Setenv("TWILIO_ACCOUNT_SID", "")
	t.Setenv("TWILIO_AUTH_TOKEN", "")
	t.Setenv("TWILIO_FROM", "")
	r, e = Send(context.Background(), "sms", "+251911111111", "", "Synthetic")
	if e == nil || r.Status != "failed" {
		t.Fatal("blank credentials accepted", r, e)
	}
	r, e = Send(context.Background(), "unknown", "example@example.test", "", "Synthetic")
	if e == nil || r.Status != "failed" {
		t.Fatal("unknown channel accepted", r, e)
	}
}
