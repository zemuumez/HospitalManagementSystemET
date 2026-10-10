# Inventory ledger and upgrade review

Reviewed commit: `8cc47e7` on `feat/inventory-workspace-parity`.
Verdict: **changes required**. Main was not modified.

The dedicated 055 migration resolves skipped-054 schema delivery; 054 matches its earlier submitted version. New application writes assign ledger sequence and balance under the item lock. Cleanup now includes private nonclinical uploads. The remaining implementation blocker concerns pre-migration history, followed by failing test fixtures.

## G1 — Backfilled history is incorrectly trusted as actual serialized history (high)

Migration `055_inventory_void_and_ledger_order.sql:33` derives both sequence and balance from `created_at`, delta sign, and UUID. These produce a deterministic ordering, but cannot recover historical lock-acquisition order. Every old row receives a positive sequence, so the void check in `services/api/internal/adapters/postgres/inventory.go` trusts these inferred balances exactly like newly recorded balances.

Independent reproduction used a fresh isolated schema with migrations through 054, inserted valid ledger movements in actual order A(+100), C(-50), B(+100), and supplied timestamps in transaction-start order A, B, C. Each movement committed with the matching item balance. Applying the real migration 055 produced:

| Order | Actual balances before upgrade | Backfilled balances |
| --- | --- | --- |
| 1 | A: 100 | A: 100 |
| 2 | C: 50 | B: 200 |
| 3 | B: 150 | C: 150 |

The new void minimum query returned **100** for receipt A (quantity 100), hiding the actual minimum of **50**. Thus the consumption check still permits an unsafe void for migrated history. The temporary schema was removed. This reproduction exercised the real migration and void decision query; it did not invoke the HTTP void endpoint.

Required: distinguish inferred legacy history from trusted sequences recorded under the item lock. A conservative approach is to reject automatic voids of legacy receipts with unverified history and provide an explicit reconciliation/correction path. Preserve existing stock balances and audit records. New receipts created after cutover may use the trusted sequence algorithm. Do not present timestamp/sign/UUID sorting as proof of historical order, and remove or safely reject the sequence-zero timestamp fallback.

Test a populated prior-054 upgrade with reversed transaction-start/commit order and equal timestamps. Verify migrated ambiguous receipts cannot bypass consumption protection, and verify fresh post-upgrade receipts can still be voided correctly. The current F1 test checks an already freshly migrated schema; it does not run this populated upgrade.

## G2 — Full PostgreSQL test gate fails; boundary tests do not reach their assertions

Independent `go test ./... -count=1` with `HMS_TEST_DATABASE_URL` configured exited **1**. Failures:

- `R1: receipt attachment lifecycle and reference protection`, `inventory_test.go:302`: duplicate token violates `secure_attachment_token_key`. The new aged fixture reuses an earlier token; the recent fixture also reuses the receipt token. Use distinct generated tokens for all fixtures.
- `F2: transaction-start vs lock-acquisition ordering and timestamp ties`, `inventory_test.go:527`: validation failed. `key-order-rec-a` is below the application's 16-character idempotency-key minimum. Audit all new fixture keys, including the tie cases.
- `F1: database upgrade schema validation and void operation`, `inventory_test.go:651`: validation failed. `key-upg-rec-001` is also too short.
- `S4: attachment retirement on settings update & 404 on retired token`, `cms_settings_test.go:1136`: the fixture labelled clinical has `is_public=false` but no patient/encounter link. Under the explicit new operational policy it is eligible for cleanup. Link this clinical fixture to a real test patient/encounter and separately verify aged private nonclinical cleanup; do not restore the old blanket exclusion of private files.

Keep these tests enabled, check fixture-creation errors, and run the full suite with the database environment configured. A pass with the PostgreSQL test skipped is not evidence for this gate. Retain the accepted Settings worker isolation fix.

## Verification and delivery

TypeScript typecheck, all 8 frontend tests, and formatting passed. All 23 isolated browser journeys passed with exit code 0 and successful schema cleanup (local log: `.local/inventory-ledger-review.log`). The full Go suite failed as recorded above. No implementation files were changed during review.

Address G1 and G2 in one bounded correction batch. Preserve the already resolved pagination, empty-category, attachment-reference, and new-write sequencing behavior. Return a populated-upgrade regression result and the full database-enabled test results. Keep main, Smart Cards, and Odontogram unchanged pending acceptance.
