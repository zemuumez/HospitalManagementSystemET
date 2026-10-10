# Inventory remediation independent review

Reviewed submission: `0c5927c` on `feat/inventory-workspace-parity`.
Verdict: **changes required before acceptance**. Main is untouched.

## Improvements confirmed

Receipt binding now locks attachment rows and rejects missing/clinical tokens. Direct retirement, Settings displacement, and cleanup reference queries include Inventory. Receipt voids now have an original receipt link, a unique index, and visible voided state. Catalog retrieval traverses subsequent pages, issue return-status filtering occurs before server pagination, and the theme assertion targets the Inventory tab. Preserve these changes.

## F1 — Existing installations do not receive the void schema upgrade (high)

`db/migrations/054_inventory_parity.sql` was changed in place to add `void_receipt` constraints and its unique index. The earlier submitted `95e1f3b` already contained migration 054. `apps/web/scripts/migrate.ts:18` skips a filename present in `schema_migration`, so these new statements never run on a database that applied the previous 054.

An isolated upgrade probe applied the previous 054, recorded migration filenames, then used the current runner's skip-by-name behavior. Result: zero migrations applied; the installed movement-kind constraint still disallows `void_receipt`. The probe schema was removed. Fresh-schema tests do not cover this upgrade path.

Required: add a new numbered migration for the constraint/index changes, retaining compatibility with both prior-054 and fresh installations. Do not fix this by deleting migration history or requiring a database reset. Verify an upgrade with existing stock/movement data and exercise a receipt void afterwards.

## F2 — Consumption protection orders transactions by start time, not movement order (high)

The minimum-running-balance query in `services/api/internal/adapters/postgres/inventory.go` orders by `created_at DESC, id DESC`. Migration 010 defines `created_at DEFAULT now()`: PostgreSQL transaction start time. An operation can begin earlier but acquire the item lock later than another operation. Random UUID order also does not resolve equal timestamps into ledger order.

Concrete sequence: receipt A adds 100; replenishment transaction B begins but waits before obtaining the item lock; issue C begins later and applies -50; B then applies +100. Actual serialized balances are 100, 50, 150. Stored start-time order is A, B, C. The current reverse-window query reconstructs a minimum of 100, so voiding A is allowed even though actual balance fell to 50 before replenishment. An isolated SQL probe of the exact window calculation returned 100 for this witness. This is a query-level reproduction, not a claim that an end-to-end concurrent test was run.

Required: record a deterministic per-item ledger order under the item lock (or persist sufficient ordered balance information), and calculate historical minima in that order. Include a deterministic regression with controlled transaction start/lock acquisition order, plus timestamp ties. Preserve the unique original-receipt reversal protection and idempotency. Decide explicitly how preexisting ambiguous history is handled; do not infer guaranteed ordering from UUIDs.

## F3 — Abandoned private receipt uploads are excluded from cleanup (medium)

Receipt uploads are now correctly private and nonclinical. However, `CleanupAbandonedAttachments` in `services/api/internal/adapters/postgres/cms_settings.go:515` still requires `sa.is_public = true`, and its per-candidate guard at line 561 skips `!c.isPublic`. An upload abandoned by cancelling the receipt form or replacing its selected file can therefore never be reclaimed by the operational cleanup worker, regardless of age.

The new Go preservation test does not exercise the age boundary: it supplies `time.Nanosecond`, which truncates to zero seconds and falls back to 24 hours, while the fixture is newly created. It also uses a private fixture that is excluded before the new Inventory reference checks. Passing that assertion does not establish aged private-upload lifecycle behavior.

Required: include eligible nonclinical private operational uploads in bounded cleanup, while preserving patient/encounter files and all referenced historical assets. Use explicit attachment purpose metadata if needed to distinguish other private file types. Test aged unreferenced private uploads being removed, recent pending uploads retained, aged referenced private receipts retained, clinical records retained, and a binding/cleanup race under the shared lock protocol.

## Independent gates

Full Go/PostgreSQL suite (`go test ./... -count=1` with the test database configured), TypeScript typecheck, all 8 frontend unit tests, and root formatting check passed. All 23 isolated browser journeys passed with exit code 0; schema cleanup succeeded. Local log: `.local/inventory-final-review.log`. Upgrade/window probes ran in isolation and removed their temporary schema; implementation files were not modified. The passing gates do not cover the upgrade and boundary cases above.

## Next correction batch

Address F1–F3 together with the boundary regressions above. Retain the accepted fixes and accurate evidence distinctions. Return the final HEAD and upgrade/concurrency/cleanup results. Keep main, Smart Cards, and Odontogram unchanged until independent acceptance.
