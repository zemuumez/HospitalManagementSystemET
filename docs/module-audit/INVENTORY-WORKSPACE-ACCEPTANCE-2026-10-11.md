# Inventory workspace acceptance

Accepted implementation: `4926f9a978da4a5690525a68be7fe42188c41012` on `feat/inventory-workspace-parity`.
Verdict: **accepted for the reviewed Inventory workspace scope**. No outstanding blocking findings from the Inventory review sequence remain. This is not acceptance of the entire HMS or of deferred modules.

## Independent verification

All checks ran against the accepted implementation:

| Gate | Result |
| --- | --- |
| Full Go suite with PostgreSQL enabled, `go test ./... -count=1` | Passed, exit 0; PostgreSQL package 55.311 seconds |
| TypeScript typecheck | Passed |
| Frontend unit tests | 8/8 passed |
| Repository formatting check | Passed |
| Isolated Inventory browser suite | 23/23 journeys passed, exit 0; schema cleanup succeeded |

Local verification logs: `.local/inventory-acceptance-go.log` and `.local/inventory-acceptance-review.log`. These are local evidence files, not required deployment artifacts.

## Review closure

- I1/I2/I4: register pagination, complete category/item selector retrieval, server-side issue-status filtering, empty-category safeguards, and exact fractional/zero reorder editing are covered by the delivered changes and regressions.
- I3 and attachment reviews: source business fields, recipient display, receipt uploads and reference protection, private operational cleanup, and linked receipt reversal behavior are implemented. Item locking, idempotency, audit writes, stock reconciliation, and cumulative return protection are preserved.
- I5/R4: the real shell theme and Amharic checks pass, including the scoped Inventory navigation assertion.
- F1/F2/G1/H1: dedicated migration 055 introduces the ledger schema; corrective migration 056 also repairs installations that already applied the original timestamp-backfilling 055. The populated upgrade regression reproduces the old inferred history, records migration history, applies 056, verifies retained stock totals and movement rows, checks cutover provenance and legacy flags, rejects legacy receipt voids, and successfully voids a new post-repair receipt.
- G2: the database-enabled full suite now passes with repaired attachment tokens, valid idempotency keys, and correctly linked clinical fixtures.

## Intentional operational contract

Migration 056 conservatively marks **all movements existing at cutover** as legacy/unverified and resets their derived sequence/balance-history metadata. It preserves item stock totals, movement identities, quantities, financial amounts, and audit rows. Automatic voids of those legacy receipts return 409 Conflict. The application does not infer their historical ordering or silently bypass this restriction. Fresh post-cutover movements receive trusted sequence/balance metadata under the item lock and retain normal validated reversal behavior.

Apply the complete migration chain through 056 using the normal migration runner before starting the matching application version. Do not manually rerun 056 on an already migrated live database: the runner's recorded migration state is the one-time cutover boundary.

This acceptance covers the reviewed inventory receipt/issue/return and correction workflows. It does not certify standalone procurement, lot/expiry tracking, or supplier address-book modules, nor exact pixel parity across every HMS screen.

## Integration boundary

This review changes documentation only. Main remains at the observed `5cb7f81a65e4061ef6d853bb30a09347761f4aa8`; no merge or deployment was performed. Inventory is ready for integration. Smart Cards and Odontogram remain untouched; Smart Cards is the next planned workspace after Inventory integration.
