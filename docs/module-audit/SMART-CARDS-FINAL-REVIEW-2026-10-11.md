# Smart Cards correction review — 2026-10-11

Submitted commit: `b65cc98`. Decision: changes requested.

The correction includes patient scope in both list queries, a stale-card deletion guard, conservative optional-field disclosure for template-less cards, independent QR decoding, and localized canvas labels. Three issues remain in the new idempotency implementation.

## F1 — High: already-applied migration 057 never creates the request table

`db/migrations/057_patient_smart_cards_parity.sql:39` adds `smart_card_generation_request` to an existing migration. `apps/web/scripts/migrate.ts:18` skips filenames already recorded in `schema_migration`. A database that ran the previous Smart Cards version therefore never gets this table. The updated frontend always supplies a key, causing generation to query a nonexistent relation and fail.

Deliver a new numbered migration that creates the table/index safely on both already-applied-057 databases and fresh installations. Test an upgrade from the actual prior 057 contents with its migration marker present, then generate and replay a card request using the upgraded application. Do not rely exclusively on fresh-schema tests.

## F2 — High: replay can undo a deliberate credential revocation

`postgres/patient_extensions.go:478–490` returns a cached response only while every cached card is still active. Otherwise it falls through to generation, revokes current cards and issues replacements; line 696 overwrites the original request's response.

Reproduction: generate A using key K; delete A's assignment; replay the identical request with K. The replay issues a new active credential and restores the assignment. Similarly, replaying an older key after a deliberate reassignment can supersede the newer credential. Completed requests must never become new issuance commands because their original cards were revoked.

Independent PostgreSQL evidence: a temporary test using the repository's isolated-schema fixture executed that sequence and failed with `replaying completed request after deletion issued a NEW active credential` (test duration 2.90s, process exit 1). The temporary test was removed after recording the result.

Keep completed request records immutable. Return the original issuance outcome or an explicit stale/revoked conflict without mutating assignments or credentials. Test replay after assignment deletion, direct revocation, reassignment and template deletion. Keep enough durable request identity to reject reuse after template deletion rather than cascading away the entire idempotency history.

## F3 — Medium: frontend retries discard the original request key

`smart-cards-workspace.tsx:316` generates a new UUID inside each invocation of the submit handler. A network failure after the server commits leaves the dialog open; clicking submit again creates a different key. The server therefore treats the retry as a new issuance and revokes the first card, despite the new server-side idempotency table.

Retain the key and payload for one logical submission across uncertain responses and user retries. Allocate a new key for a deliberate new operation, not merely another transport attempt. Add a browser test that allows the first request to commit, drops its response, retries from the same dialog, and confirms the same card/token and one issuance record.

## Verification limits and next gate

The existing PostgreSQL Smart Cards suites independently passed (`go test ./internal/adapters/postgres -run TestSmartCards_ -count=1`, database enabled, exit 0, 20.525s). The focused failing PostgreSQL regression above demonstrates a case those suites omit. Source review confirms F1 and F3. Browser, TypeScript and formatting results in the delivery report were not independently rerun for this review. Resolve F1–F3 together and include the upgrade, revocation/replay and lost-response browser regressions before acceptance. Main and Odontogram were not modified.
