# Backend operations and error contract

The Go API emits an independently generated 128-bit X-Request-ID for correlation. Structured request logs include only that ID, HTTP method, response status and elapsed milliseconds. They exclude URL/query, bodies, cookies, authorization headers and patient identifiers. Tests assert that supplied sensitive strings and forged request IDs are absent from logs.

GET /healthz is process liveness. GET /readyz checks PostgreSQL under a one-second context deadline and returns 200 ready or 503 unavailable without connection details. Route the readiness endpoint only through your infrastructure monitoring as appropriate; the API itself remains private.

Errors retain the existing `error` display string and now add stable `code` values: INVALID_JSON (400), UNAUTHENTICATED (401), ORIGIN_DENIED/FORBIDDEN (403), NOT_FOUND (404), IDEMPOTENCY_CONFLICT/STATE_CONFLICT (409), VALIDATION_FAILED (422), INTERNAL_ERROR (500). Consumers should use codes instead of matching English sentences. Unexpected database/internal error text is not returned to clients. Complete OpenAPI coverage and generated frontend types remain pending.

Go tests and vet passed, including readiness failure/success, error-code mappings and redaction. Metrics/alert rules, tracing, proxy/TLS deployment, secret rotation and least-privilege database roles remain separate pending operational tasks. Request IDs do not claim an independent security assessment.

## Message worker ownership and recovery

Stop old workers before applying migration 012, then start the new worker binary. The old worker does not understand leases and cannot claim rows under the new constraint. Migrations convert old processing rows to uncertain; check provider evidence before any future reconciliation. Pending rows remain queued.

The new worker atomically claims one pending row with a random token and 45-second lease, increments attempts, and renews ownership every 10 seconds while sending. Delivery has a 20-second context; SMTP socket operations also have explicit deadlines. Finalization has its own five-second timeout. Only an unexpired matching token may finish a row. On restart, workers reclaim no uncertain messages: up to 100 expired rows per iteration become uncertain. A delayed worker cannot change those rows. HTTP 5xx SMS outcomes are conservatively uncertain. Provider acceptance is still not handset delivery confirmation.

To inspect without exposing message bodies or recipients, group `message_outbox` by status and count; inspect expired processing counts and attempts. Alerting deployment and operator reconciliation/retry APIs remain pending. Do not manually reset uncertain messages to pending without confirming non-delivery with the provider. Blank provider settings are covered by the provider setup worksheet; development SMS capture is refused in production.
