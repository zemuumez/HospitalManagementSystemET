package delivery

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"time"
)

type Result struct {
	Status     string
	ProviderID string
}

func Send(ctx context.Context, channel, to, subject, body string) (Result, error) {
	if channel == "sms" {
		provider := os.Getenv("SMS_PROVIDER")
		if provider == "" || provider == "capture" {
			if os.Getenv("APP_ENV") == "production" {
				return Result{Status: "failed"}, errors.New("capture SMS prohibited in production")
			}
			return Result{Status: "captured"}, nil
		}
		if provider != "twilio" {
			return Result{Status: "failed"}, errors.New("unknown SMS provider")
		}
		sid, token, from := os.Getenv("TWILIO_ACCOUNT_SID"), os.Getenv("TWILIO_AUTH_TOKEN"), os.Getenv("TWILIO_FROM")
		if !strings.HasPrefix(sid, "AC") || len(sid) != 34 || token == "" || from == "" {
			return Result{Status: "failed"}, errors.New("missing SMS configuration")
		}
		form := url.Values{"To": {to}, "From": {from}, "Body": {body}}
		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.twilio.com/2010-04-01/Accounts/"+url.PathEscape(sid)+"/Messages.json", strings.NewReader(form.Encode()))
		if err != nil {
			return Result{Status: "failed"}, err
		}
		req.SetBasicAuth(sid, token)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := (&http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}).Do(req)
		if err != nil {
			return Result{Status: "uncertain"}, errors.New("SMS provider outcome uncertain")
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return Result{Status: "failed"}, errors.New("SMS provider rejected request")
		}
		var response struct {
			SID string `json:"sid"`
		}
		if json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&response) != nil || response.SID == "" {
			return Result{Status: "uncertain"}, errors.New("SMS provider outcome uncertain")
		}
		// 'sent' means accepted by the provider, not confirmed handset delivery.
		return Result{Status: "sent", ProviderID: response.SID}, nil
	}
	host, port := os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT")
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "1025"
	}
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = "noreply@hms.local"
	}
	if strings.ContainsAny(from+to+subject, "\r\n") {
		return Result{Status: "failed"}, errors.New("invalid mail headers")
	}
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return Result{Status: "failed"}, errors.New("mail transport unavailable")
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return Result{Status: "failed"}, err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return Result{Status: "failed"}, err
		}
	} else if os.Getenv("APP_ENV") == "production" {
		return Result{Status: "failed"}, errors.New("production SMTP requires TLS")
	}
	if user := os.Getenv("SMTP_USER"); user != "" {
		if err = client.Auth(smtp.PlainAuth("", user, os.Getenv("SMTP_PASSWORD"), host)); err != nil {
			return Result{Status: "failed"}, err
		}
	}
	if err = client.Mail(from); err != nil {
		return Result{Status: "failed"}, err
	}
	if err = client.Rcpt(to); err != nil {
		return Result{Status: "failed"}, err
	}
	writer, err := client.Data()
	if err != nil {
		return Result{Status: "failed"}, err
	}
	_, err = fmt.Fprintf(writer, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", from, to, subject, body)
	if err != nil {
		return Result{Status: "uncertain"}, err
	}
	if err = writer.Close(); err != nil {
		return Result{Status: "uncertain"}, err
	}
	_ = client.Quit()
	return Result{Status: "sent"}, nil
}
