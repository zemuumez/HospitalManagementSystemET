# Backend operations and error contract

The Go API emits an independently generated 128-bit X-Request-ID for correlation. Structured request logs include only that ID, HTTP method, response status and elapsed milliseconds. They exclude URL/query, bodies, cookies, authorization headers and patient identifiers. Tests assert that supplied sensitive strings and forged request IDs are absent from logs.

GET /healthz is process liveness. GET /readyz checks PostgreSQL under a one-second context deadline and returns 200 ready or 503 unavailable without connection details. Route the readiness endpoint only through your infrastructure monitoring as appropriate; the API itself remains private.

Errors retain the existing `error` display string and now add stable `code` values: INVALID_JSON (400), UNAUTHENTICATED (401), ORIGIN_DENIED/FORBIDDEN (403), NOT_FOUND (404), IDEMPOTENCY_CONFLICT/STATE_CONFLICT (409), VALIDATION_FAILED (422), INTERNAL_ERROR (500). Consumers should use codes instead of matching English sentences. Unexpected database/internal error text is not returned to clients. Complete OpenAPI coverage and generated frontend types remain pending.

Go tests and vet passed, including readiness failure/success, error-code mappings and redaction. Metrics/alert rules, tracing, proxy/TLS deployment, secret rotation and least-privilege database roles remain separate pending operational tasks. Request IDs do not claim an independent security assessment.
