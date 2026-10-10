# Inventory corrections independent review

Reviewed HEAD: `95e1f3b` on `feat/inventory-workspace-parity`, including `6d0ce5e`.

Verdict: **changes required**. The empty-category safeguards (I2), fractional/zero reorder editing (I4), and visible theme/localization corrections (I5) are materially improved. Three remaining implementation issues prevent acceptance of I1/I3; the browser gate also fails as described below. Main was not modified.

## Verification

Independent full Go/PostgreSQL suite (`go test ./... -count=1` with the test database configured), TypeScript typecheck, all 8 frontend tests, and repository-root formatting check passed. The isolated browser suite passed journeys 1–18, then failed Journey 19 with exit code 1. Isolated-schema cleanup succeeded. Journeys 20–23 did not execute. Local run log: `.local/inventory-corrections-review.log`. R1–R3 below are code-inspection findings; R4 records the observed test failure.

## R1 — Receipt attachments are not integrated into reference protection (high)

`apps/web/src/components/inventory-workspace.tsx:528` uploads receipts with `isPublic=true`, then stores the returned URL on the movement. `services/api/internal/adapters/postgres/inventory.go:282` simply persists that string. There is no attachment existence/type/ownership validation or attachment-row locking when binding it.

All retirement paths still recognize only `hospital_general_setting` and `front_cms_setting`: direct retirement in `attachments.go`, displaced-token retirement in `cms_settings.go:154`, and abandoned cleanup in `cms_settings.go:493`. They do not check `inventory_movement.attachment_url`. Consequently, an aged uploaded receipt referenced only by Inventory qualifies for cleanup; direct retirement also succeeds despite the receipt reference. A settings asset shared with a receipt can be retired when displaced from Settings. Binding can also race retirement and persist a broken URL.

Journey 22 fills Attachment URL with `https://files.ulshms.local/receipts/batch-invoice-001.pdf` and checks the saved string/link. It does not upload a file, download its bytes, or run cleanup against a bound receipt.

Required: use a validated attachment reference and coordinated locking for binding and every retirement path; include Inventory references in candidate pre-filtering and locked rechecks. Establish the intended receipt access policy explicitly rather than inheriting public-branding semantics. Preserve attachment access for historical/voided receipts. Test real upload/download, missing/retired and inappropriate clinical binding, direct retirement conflict, aged cleanup preservation, shared Settings/Inventory references, and binding-versus-retirement concurrency. Preserve the accepted Settings lifecycle behavior.

## R2 — Void Receipt is an unlinked repeatable write-off (high)

`inventory-workspace.tsx:680` posts a normal `kind=writeoff` containing item, quantity, store, and a free-text reason. It sends no original receipt ID. The domain explicitly forbids `originalId` for write-offs (`services/api/internal/domain/inventory.go`), and `MoveInventory` only checks the current item balance for this kind. No record marks the original receipt reversed, and no uniqueness constraint prevents another reversal.

Concrete consequence: receive A=100 and B=100 for the same item; void A twice using fresh request keys. Both write-offs satisfy the balance check, reducing 200 to 0 although B remains a valid receipt. Existing per-request idempotency does not prevent a second logical void with a new key. Likewise, consuming A and later replenishing stock permits its purported consumed-receipt void because only aggregate availability is checked. The report's claim that consumed receipts are always protected is therefore unsupported.

Required: implement a server-owned reversal operation linked to the original movement, with transactional validation, a durable reversed state, and protection against repeated/concurrent voids across different request keys. Define and enforce the consumption/correction rule, retain history, and expose the resulting state in registers/details. Distinguish voids from ordinary issue returns if the UI calls them voids. Provide a usable audited amendment/replacement path for source receipt corrections. Test two receipts of one item, repeated and concurrent void attempts, intervening consumption/replenishment, and partial-return issue corrections. Do not rely on client reason text as the reversal identity.

## R3 — Pagination still leaves selectors and issue filters incomplete (medium)

Register pagination is now server-backed, but `fetchInventoryData` at `inventory-workspace.tsx:270` fetches selector catalogs once at `page=1&limit=500` and ignores catalog totals. Category/item 501 is still silently unavailable to the forms; related category/unit lookups also depend on that truncated catalog. The comment claiming dropdowns are never truncated is incorrect. The new 25-plus fixtures do not cover this boundary.

The issued-item status filter at component lines 1551–1559 still filters only the current server page. It is not sent to the API, does not reset pagination, and totals remain unfiltered. A page full of returned issues can therefore show no returnable rows while older outstanding issues exist on later pages. Tab counters reuse the current movement result total rather than separate receipt/issue totals.

Required: paginated/searchable selectors that reach the complete catalog, independent related-record display data, and server-side return-status filtering before pagination with matching totals. Use per-kind totals or accurately label counts. Test at least 501 catalog entries and an outstanding issue behind a page of returned issues, including filter changes from later pages. Keep previously corrected empty-category safeguards intact.

## R4 — Independent browser gate fails at Journey 19

At `scripts/verify-inventory-workspace.mjs:1488`, the assertion reports `Expected high-contrast slate text, got oklab(0.546413 -0.00859972 -0.0364768)`. The card background and status badge assertions preceding it passed (`rgb(18, 21, 31)` and `rgb(52, 211, 153)`).

The navigation selector is global: `nav button:not([class*="bg-primary"])`. It is not scoped to the Inventory tab navigation and may select another shell control. This run establishes a failing gate, not that the supplied Inventory screenshot is false. Identify the actual selected element, scope the assertion to a stable Inventory navigation target, wait for the applied theme styles, and verify the intended text/background contrast. Fix CSS if that intended target fails; do not simply remove or weaken the assertion. Rerun all 23 journeys on the final committed HEAD.

## Delivery boundary

Address R1–R4 as one correction batch, with focused regressions and an accurate source/evidence matrix. Do not claim uploaded-attachment lifecycle verification from an external URL string test, or consumed-receipt protection from a single insufficient-balance example. Rerun the existing quality gates and the added boundary tests. Keep main, Smart Cards, and Odontogram unchanged pending acceptance.
