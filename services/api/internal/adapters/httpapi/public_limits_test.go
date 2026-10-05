package httpapi

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublicLimits(t *testing.T) {
	l := &publicLimiter{}
	now := time.Now()
	for i := 0; i < 20; i++ {
		if !l.allow("one", now) {
			t.Fatal("early limit")
		}
	}
	if l.allow("one", now) {
		t.Fatal("limit bypass")
	}
	if !l.allow("two", now) || !l.allow("one", now.Add(time.Minute)) {
		t.Fatal("quota reset")
	}
	r := httptest.NewRequest("POST", "/v1/public/appointment-requests", nil)
	r.Header.Set("Origin", "https://evil.test")
	w := httptest.NewRecorder()
	if l.guard(w, r, "https://hospital.test") || w.Code != 403 {
		t.Fatal("cross origin public write")
	}
}
