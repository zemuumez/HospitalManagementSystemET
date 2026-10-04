package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"hms.local/api/internal/domain"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadinessAndRedactedRequestLogs(t *testing.T) {
	var logs bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	defer slog.SetDefault(old)
	for _, tc := range []struct {
		ready func(context.Context) error
		want  int
	}{{nil, 503}, {func(context.Context) error { return errors.New("private database credentials") }, 503}, {func(context.Context) error { return nil }, 200}} {
		req := httptest.NewRequest("GET", "/readyz?patient=private-patient", nil)
		req.Header.Set("Cookie", "private-session")
		req.Header.Set("X-Request-ID", "forged-id")
		w := httptest.NewRecorder()
		Server{Ready: tc.ready}.Handler().ServeHTTP(w, req)
		if w.Code != tc.want || len(w.Header().Get("X-Request-ID")) != 32 {
			t.Fatal(w.Code, w.Header())
		}
	}
	for _, secret := range []string{"private-patient", "private-session", "private database credentials", "forged-id"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatal("request log leaked sensitive input", secret)
		}
	}
}
func TestStableErrorCodes(t *testing.T) {
	for _, tc := range []struct {
		err    error
		code   string
		status int
	}{{domain.ErrForbidden, "FORBIDDEN", 403}, {domain.ErrValidation, "VALIDATION_FAILED", 422}, {domain.ErrConflict, "IDEMPOTENCY_CONFLICT", 409}, {domain.ErrNotFound, "NOT_FOUND", 404}, {domain.ErrStale, "STATE_CONFLICT", 409}, {errors.New("private database text"), "INTERNAL_ERROR", 500}} {
		w := httptest.NewRecorder()
		fail(w, tc.err)
		var body map[string]string
		if e := json.Unmarshal(w.Body.Bytes(), &body); e != nil {
			t.Fatal(e)
		}
		if w.Code != tc.status || body["code"] != tc.code || strings.Contains(w.Body.String(), "private database") {
			t.Fatal(w.Code, w.Body.String())
		}
	}
}
