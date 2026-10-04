package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type metricsCollector struct {
	mu           sync.Mutex
	requestCount map[string]int64
	durationSecs map[string]float64
}

var globalMetrics = &metricsCollector{
	requestCount: make(map[string]int64),
	durationSecs: make(map[string]float64),
}

func (m *metricsCollector) record(method string, status int, duration time.Duration) {
	key := method + "_" + strconv.Itoa(status)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestCount[key]++
	m.durationSecs[method] += duration.Seconds()
}

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
			elapsed := time.Since(started)
			globalMetrics.record(r.Method, status, elapsed)
			slog.Info("http_request", "request_id", id, "method", r.Method, "status", status, "duration_ms", elapsed.Milliseconds())
		}()
		// Deliberately exclude URL/query, headers, cookies, body, names and identifiers.
		next.ServeHTTP(tracked, r)
	})
}

func (s Server) metrics(w http.ResponseWriter, r *http.Request) {
	globalMetrics.mu.Lock()
	counts := make(map[string]int64, len(globalMetrics.requestCount))
	for k, v := range globalMetrics.requestCount {
		counts[k] = v
	}
	durations := make(map[string]float64, len(globalMetrics.durationSecs))
	for k, v := range globalMetrics.durationSecs {
		durations[k] = v
	}
	globalMetrics.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, "# HELP hms_http_requests_total Total number of HTTP requests processed")
	fmt.Fprintln(w, "# TYPE hms_http_requests_total counter")
	for k, count := range counts {
		var method, status string
		if n := len(k); n > 4 && k[n-4] == '_' {
			method = k[:n-4]
			status = k[n-3:]
		} else {
			method = k
			status = "200"
		}
		fmt.Fprintf(w, "hms_http_requests_total{method=\"%s\",status=\"%s\"} %d\n", method, status, count)
	}

	fmt.Fprintln(w, "# HELP hms_http_request_duration_seconds_total Total duration of HTTP requests in seconds")
	fmt.Fprintln(w, "# TYPE hms_http_request_duration_seconds_total counter")
	for method, dur := range durations {
		fmt.Fprintf(w, "hms_http_request_duration_seconds_total{method=\"%s\"} %.6f\n", method, dur)
	}

	fmt.Fprintln(w, "# HELP go_goroutines Number of goroutines currently existing")
	fmt.Fprintln(w, "# TYPE go_goroutines gauge")
	fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())

	fmt.Fprintln(w, "# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use")
	fmt.Fprintln(w, "# TYPE go_memstats_alloc_bytes gauge")
	fmt.Fprintf(w, "go_memstats_alloc_bytes %d\n", memStats.Alloc)

	fmt.Fprintln(w, "# HELP hms_app_info Application build and deployment metadata")
	fmt.Fprintln(w, "# TYPE hms_app_info gauge")
	fmt.Fprintln(w, "hms_app_info{version=\"1.0.0\",service=\"hms-api\"} 1")
}
