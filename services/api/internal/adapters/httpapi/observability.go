package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

type observedResponse struct {
	http.ResponseWriter
	status int
}

func (w *observedResponse) WriteHeader(code int) {
	if w.status != 0 {
		return
	}
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
func (w *observedResponse) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}
func (w *observedResponse) Unwrap() http.ResponseWriter { return w.ResponseWriter }
func observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		random := make([]byte, 16)
		if _, e := rand.Read(random); e != nil {
			write(w, 500, map[string]string{"error": "Unable to complete the request", "code": "INTERNAL_ERROR"})
			return
		}
		id := hex.EncodeToString(random)
		w.Header().Set("X-Request-ID", id)
		tracked := &observedResponse{ResponseWriter: w}
		defer func() {
			status := tracked.status
			if status == 0 {
				status = 200
			}
			slog.Info("http_request", "request_id", id, "method", r.Method, "status", status, "duration_ms", time.Since(started).Milliseconds())
		}()
		// Deliberately exclude URL/query, headers, cookies, body, names and identifiers.
		next.ServeHTTP(tracked, r)
	})
}
