# Inventory cutover review

Reviewed submission: `cc43d1a` on `feat/inventory-workspace-parity`.
Verdict: **one remaining upgrade blocker**. Main was not modified.

## Confirmed resolutions

The database-enabled full Go suite now passes, including the repaired fixtures and populated prior-054 regression. Fresh application writes record sequence and balance under the item lock. Applying the current migration 055 to a prior-054 database leaves old history at sequence zero, and the application rejects automatic voids of those receipts. The unsafe sequence-zero timestamp fallback is removed.

## H1 — Already-applied original migration 055 is not repaired (high)

The fix edits migration 055 in place. `apps/web/scripts/migrate.ts` skips migration filenames already recorded in `schema_migration`. A database that applied 055 from `8cc47e7` retains its timestamp-derived positive `ledger_seq` and `balance_after_milli` values. Removing the backfill from the file does not undo those stored values, and the new `origSeq <= 0` rejection does not apply to them.

Independent isolated reproduction:

1. Apply migrations through 054 and create the valid historical witness with actual balances 100, 50, 150 and transaction-start order A, B, C rather than applied order A, C, B.
2. Apply the real 055 from `8cc47e7`, and record its filename as applied.
3. Apply the current runner's skip-by-name logic: the edited 055 is skipped.
4. Observe that A/B/C still have sequences 1/2/3 and reconstructed balances 100/200/150. The current minimum query returns 100 for receipt A (quantity 100), concealing the actual dip to 50.

The probe used a temporary schema that was removed afterwards. It exercised migration SQL and the void decision query, not the HTTP endpoint. The new prior-054 test does not cover this already-applied-055 path.

Required bounded correction:

- Add a new numbered corrective migration for databases that have already applied the earlier 055; do not rely on another edit of an applied migration or deleting migration history.
- Explicitly mark potentially inferred history unverified, or conservatively disallow automated voids where trustworthy provenance cannot be established. Protect genuinely new writes after the repair with a clear cutover rule. Do not guess which rows were inferred from timestamps, UUIDs, or current balance alone.
- Preserve stock totals, original movement identities, and audit history. Record any repair metadata deliberately, without silently rewriting historical financial amounts.
- Add a populated upgrade regression that applies the original 055, records it in migration history, then runs the corrective migration and proves legacy void rejection plus successful trusted post-repair receipt voiding. Retain the existing fresh and prior-054 tests.
- Correct the stale migration 055 header claiming it backfills history.

This is a continuation of the legacy-history protection requirement, not additional Inventory feature scope. No other implementation finding is added in this review.

## Independent verification

Full Go/PostgreSQL suite (`go test ./... -count=1`, database configured), TypeScript typecheck, all 8 frontend tests, and formatting passed. All 23 isolated browser journeys passed with exit code 0 and successful schema cleanup. Local logs: `.local/inventory-cutover-go.log` and `.local/inventory-cutover-review.log`. Implementation files were not modified.

Keep main, Smart Cards, and Odontogram unchanged until this upgrade path is covered and acceptance is recorded.
