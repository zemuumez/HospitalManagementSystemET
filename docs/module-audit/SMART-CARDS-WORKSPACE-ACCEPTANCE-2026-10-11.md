# Smart Cards review acceptance — 2026-10-11

Implementation reviewed: `21e9d29` on `feat/smart-cards-workspace-parity`.

Decision: accept the reviewed Smart Cards correction batch. No outstanding findings remain from R1–R5, F1–F3 and U1 in this review sequence.

## Final blocker closure

Migration 059 explicitly removes existing template-id foreign keys from `smart_card_generation_request`. It repairs databases that already applied the intermediate 057 and original 058, rather than relying on a changed CREATE TABLE statement. Request rows and their primary/unique constraints are preserved. The updated 058 also reconciles that constraint when it runs for the first time.

Inspected regression coverage verifies the original-057, intermediate-057, already-applied-058 and fresh-installation paths. Completed request history survives template deletion; identical replay returns the original revoked credential without issuance, and changed-payload reuse conflicts.

## Independent verification

Executed with PostgreSQL enabled, using isolated test schemas:

`go test ./internal/adapters/postgres -run TestSmartCards_ -count=1 -v`

Result: all 12 suites passed, exit 0, package duration 26.952s. Coverage includes template integrity, generation modes, access scope, stale deletion, legacy disclosure, idempotency, upgrade paths and replay after revocation/deletion.

The final commit changes migrations, database regression tests and documentation only. Earlier source review checked the retained browser submission key, immutable replay behavior, independent QR decoding and localized canvas output.

Evidence boundary: browser journeys, TypeScript and formatting passes for this delivery are implementer-reported; they were not independently rerun in this acceptance worktree. This acceptance records closure of the reviewed findings, not a claim that every repository or deployment gate was independently executed.

## Integration boundary

Only this acceptance document was added by the reviewer. Main and Odontogram were not modified. The feature branch is ready for the normal integration workflow. Apply all outstanding migrations through 059 to the target database and restart the Go backend with the integrated binary when deploying; an old running binary does not acquire routes through frontend hot reload.
