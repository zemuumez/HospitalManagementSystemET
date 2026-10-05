# Production Observability, Metrics, and Alerting Strategy

## 1. Metrics Exposition & Architecture

The HMS ET Go backend exposes standard Prometheus-compatible telemetry at `/metrics`.

### 1.1 Core Application Metrics
- **`hms_http_requests_total{method, status}`**: Monotonically increasing counter of processed HTTP requests grouped by method and HTTP status code.
- **`hms_http_request_duration_seconds_total{method}`**: Aggregate request execution time in seconds.
- **`go_goroutines`**: Active Go runtime coroutine count.
- **`go_memstats_alloc_bytes`**: Current heap allocation in bytes.
- **`hms_app_info{version, service}`**: Constant gauge with application release metadata.

---

## 2. Prometheus Alerting Rules Specification

```yaml
groups:
  - name: hms_api_alerts
    rules:
      - alert: HighHttp5xxErrorRate
        expr: |
          sum(rate(hms_http_requests_total{status=~"5.."}[5m]))
          /
          sum(rate(hms_http_requests_total[5m])) > 0.01
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "API 5xx error rate exceeds 1%"
          description: "Over 1% of requests returned HTTP 500/503 in the last 5 minutes."

      - alert: HighRequestLatency
        expr: |
          rate(hms_http_request_duration_seconds_total[5m])
          /
          rate(hms_http_requests_total[5m]) > 0.200
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Average HTTP request latency exceeds 200ms"
          description: "Database queries or external integrations may be bottlenecking."

      - alert: DatabaseReadinessFailed
        expr: probe_success{instance="http://hms-api:8080/readyz"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "API database readiness check failed"
          description: "The API service cannot ping the primary PostgreSQL instance."

      - alert: OutboxMessageBacklog
        expr: hms_outbox_pending_messages > 50
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Outbox delivery worker lag detected"
          description: "More than 50 SMS/Email notifications are queued for over 10 minutes."
```

---

## 3. Structured Logging & Redaction Compliance

1. **Request Tracing**: Every inbound request receives a cryptographically secure 128-bit `X-Request-ID` emitted in HTTP headers and bound to log entries.
2. **Zero Sensitive Data Logging**:
   - URLs, query parameters, request bodies, cookies, passwords, patient national IDs, and card numbers are strictly omitted from standard access logs (`observability.go`).
   - Log format is machine-parsable JSON:
     ```json
     {"time":"2026-10-05T02:30:00Z","level":"INFO","msg":"http_request","request_id":"c6f3...","method":"GET","status":200,"duration_ms":12}
     ```
