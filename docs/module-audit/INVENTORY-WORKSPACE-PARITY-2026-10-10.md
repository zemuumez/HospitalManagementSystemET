# Inventory Workspace Parity, Contracts, and Verification

**Date:** 2026-10-10  
**Branch:** `feat/inventory-workspace-parity`  
**Reference:** `docs/builder-handoff/IMPLEMENTATION-PLAN.md` (Phase 1, Item 4), `docs/module-audit/REPAIRS.md`, `docs/module-audit/INVENTORY-WORKSPACE-REVIEW-2026-10-10.md`

---

## 1. Executive Summary

This document verifies the complete parity, review remediation (Findings I1–I5), and hardened integrity of the **Inventory Workspace** (`/modules/items`, `/modules/item-categories`, `/modules/item-stocks`, and `/modules/issued-items`). All synthetic mocks, client-fabricated balances, local preview saves, and fallback success messages have been eradicated.

The workspace is connected to PostgreSQL through Go backend services (`/v1/inventory/*`) and Next.js authenticated proxy routes (`/api/hms/inventory/*`). Concurrency control, transactional balance tracking, strict non-negative balance enforcement, return-quantity boundaries, non-restocked write-off ledger accuracy, referenced-record deletion protection (`409 Conflict`), and bilingual English/Amharic UI support have been fully verified across 23 browser journeys.

---

## 2. Source-to-Implementation Checklist

| Area / Entity | Legacy Source Route / Blade | PostgreSQL Table | API Endpoint | Parity Status | Verification Notes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Item Categories** | `item-categories.index`<br>`item_categories/index.blade.php` | `inventory_category` | `GET/POST /v1/inventory/categories`<br>`DELETE /v1/inventory/categories/{id}` | **COMPLETE** | Real DB persistence, name/description fields, duplicate rejection, and referenced deletion protection (`RECORD_IN_USE` / `409 Conflict`). |
| **Items** | `items.index`<br>`items/index.blade.php` | `inventory_item` | `GET/POST /v1/inventory/items`<br>`PATCH /v1/inventory/items/{id}`<br>`DELETE /v1/inventory/items/{id}` | **COMPLETE** | Category relation, unit of measure, live milligram/minor-unit stock balance, exact fractional reorder thresholds (1.5 units = 1500 milli) and zero threshold support without truncation, referenced deletion protection. |
| **Item Stocks (Receipts)** | `item-stocks.index`<br>`item_stocks/index.blade.php` | `inventory_movement`<br>(`kind = 'receive'`) | `POST /v1/inventory/movements`<br>`GET /v1/inventory/movements` | **COMPLETE** | Supplier, store name (`store_name`), reference number, unit price, quantity, attachment URL (`attachment_url`), audit trail. Audited Void Receipt via server-owned `void_receipt` movement linked to `original_id` (`delta_milli = -quantity_milli`), protected by partial unique index, minimum running balance check, and UI voided badge/disabled state. |
| **Issued Items** | `issued-items.index`<br>`issued_items/index.blade.php` | `inventory_movement`<br>(`kind = 'issue'`) | `POST /v1/inventory/movements`<br>`GET /v1/inventory/movements` | **COMPLETE** | Staff recipient linkage (`recipient_id`), resolved recipient display name ("Nurse Genet Lemma"), department (`department`), issued date (`issued_date`), return due date (`return_due_date`), issued by (`issued_by`). Decreases `balance_milli` with strict check preventing negative balance. Audited Void Issue via restocked reversal movement (`kind = 'return'`). Server-side `returnStatus` filtering (`returnable` vs `returned`) before pagination. |
| **Item Returns** | `issued-items.return`<br>`issued_items/return_modal.blade.php` | `inventory_movement`<br>(`kind = 'return'`) | `POST /v1/inventory/movements` | **COMPLETE** | Partial and full returns. Links `original_id`. Rejects over-return with `409 Conflict` (`STATE_CONFLICT`). Supports `restock: false` for damaged/expired items without balance inflation. Interactive partial return badge opens return modal directly. |
| **Suppliers** | Supplier metadata in stock receipts | `inventory_movement.supplier` | `/v1/inventory/movements` | **COMPLETE** | Captured and persisted on receipt transactions; filterable and displayed in item stocks register. |
| **Stores** | Store metadata in receipts | `inventory_movement.store_name` | `/v1/inventory/movements` | **COMPLETE** | Persisted across receipts for multi-store audit tracking. |
| **Receipt Attachments** | Invoice / receipt document link | `inventory_movement.attachment_url` | `/v1/attachments`<br>`/v1/inventory/movements` | **COMPLETE** | Private operational upload lifecycle (`is_public = false`, `patient_id = nil`) via `/api/hms/attachments`. Token validated and locked (`FOR UPDATE`) on receipt creation. Referenced tokens protected from direct retirement (`409 Conflict`), Settings displacement, and abandoned cleanup. |

---

## 3. Stock Integrity & Concurrency Guarantees

1. **Transactional Operations (`inventory_movement` & `inventory_item`)**:
   - Stock movements (`receive`, `issue`, `return`, `writeoff`) execute inside an atomic database transaction.
   - Movements insert an immutable ledger entry with an idempotent movement key (`Idempotency-Key` header between 16 and 80 characters).
   - If any step fails (e.g. invalid reference, duplicate key, or balance violation), the entire transaction aborts with 0 balance drift.

2. **Insufficient Stock Prevention**:
   - For `issue` movements: `inventory_item.balance_milli` is queried using `SELECT balance_milli FROM inventory_item WHERE id = $1 FOR UPDATE`.
   - If `balance_milli < requested_milli`, the transaction returns `domain.ErrStale` (`409 Conflict`, `STATE_CONFLICT`), preventing over-spending.

3. **Concurrency Serialization**:
   - The PostgreSQL row-level lock (`FOR UPDATE`) serializes parallel issuance requests on the same item.
   - Verified via parallel worker promises executing simultaneous issue requests: exactly one request succeeds and the competing request is cleanly rejected with `409 Conflict`.

4. **Return Bounds & Non-Restocked Returns**:
   - Returns reference the original issue movement via `original_id`.
   - The backend computes cumulative returns on the issued movement (`SELECT sum(quantity_milli) FROM inventory_movement WHERE original_id = $1 AND kind = 'return'`).
   - If `already_returned + return_quantity > issued_quantity`, the operation returns `domain.ErrStale` (`409 Conflict`).
   - When returning damaged or expired goods with `restock: false`, the movement is recorded for audit tracking (`delta_milli = 0`), and `inventory_item.balance_milli` remains unchanged, preserving physical inventory accuracy.

5. **Durable Void Receipts & Running Balance Protection**:
   - Receipt reversals use server-owned `kind = 'void_receipt'` linked to the receipt's `original_id`.
   - Migration `054_inventory_parity.sql` establishes a unique partial index:
     `CREATE UNIQUE INDEX inventory_movement_void_receipt ON inventory_movement(original_id) WHERE kind='void_receipt';`
   - Reversals acquire `SELECT ... FOR UPDATE` on both the original receipt and the item row.
   - Verified running balance protection: Reversals compute the minimum running balance of the item from the time of the receipt to the present:
     `SELECT COALESCE(MIN(running_bal), current_bal) FROM (SELECT SUM(delta_milli) OVER (ORDER BY created_at ASC, id ASC) AS running_bal ...)`
     If `minRunningBal < originalQuantity`, the void is rejected with `domain.ErrConflict` (`409 Conflict`), preventing reversals of stock that was consumed prior to subsequent replenishments.
   - Repeated void attempts across distinct request keys are rejected by the database unique partial index (`409 Conflict`).

6. **Referenced Record Deletion & Attachment Protection**:
   - `DELETE /v1/inventory/categories/{id}` checks `SELECT EXISTS(SELECT 1 FROM inventory_item WHERE category_id = $1)`. If items exist, returns `domain.ErrInUse` (`409 Conflict`, `RECORD_IN_USE`).
   - `DELETE /v1/inventory/items/{id}` checks `SELECT EXISTS(SELECT 1 FROM inventory_movement WHERE item_id = $1)`. If movements exist, returns `domain.ErrInUse` (`409 Conflict`, `RECORD_IN_USE`).
   - Receipt attachments bind private tokens (`is_public = false`, no patient/encounter association). Binding locks the attachment row (`FOR UPDATE`) and validates operational ownership.
   - Direct retirement (`DELETE /v1/attachments/{token}`) checks `SELECT EXISTS(SELECT 1 FROM inventory_movement WHERE attachment_url LIKE '%' || token || '%')` and returns `domain.ErrInUse` (`409 Conflict`).
   - Displaced Settings token cleanup and aged abandoned attachment cleanup in `adapters/postgres/cms_settings.go` include candidate pre-filtering and locked rechecks protecting references in `inventory_movement.attachment_url`.

---

## 4. Independent Review Remediation (R1–R4)

Following the independent review (`INVENTORY-CORRECTIONS-REVIEW-2026-10-10.md` on `origin/codex/inventory-corrections-review`), the following hardened behaviors were implemented and verified:

- **R1 (Receipt Attachment Reference Protection)**:
  - Frontend uploads receipts with private operational policy (`isPublic: false`).
  - Binding endpoint locks `secure_attachment` via `FOR UPDATE`, rejects clinical files (`patient_id != nil || encounter_id != nil`), and validates token existence.
  - Multi-tier deletion protection prevents direct deletion (`DELETE /v1/attachments/{token}` returns 409 Conflict if bound), preserves shared attachments when displaced in Settings, and excludes bound receipt attachments from background cleanup jobs.
  - Validated via real byte upload/download, simulated cleanup, and conflict assertion on deletion attempts.

- **R2 (Durable Linked Void Receipt & Intervening Consumption Protection)**:
  - Replaced unlinked write-offs with server-owned `kind: 'void_receipt'` linked to `original_id`.
  - Database schema enforces `original_id IS NOT NULL`, `delta_milli = -quantity_milli`, and unique reversal per receipt via partial unique index.
  - Validates minimum historical running balance (`minRunningBal >= origReceipt.QuantityMilli`) to prevent voiding consumed stock even after subsequent replenishments.
  - UI displays durable "Voided" badge and disables the "Void Receipt" button. Repeated attempts across fresh request keys return `409 Conflict`.
  - Independent receipts for the same item remain individually voidable without cross-interference.

- **R3 (Complete Catalog Pagination & Server-Side Issue Status Filtering)**:
  - Dialog dropdown selectors implement automatic pagination loops (`fetchCompleteCategories` and `fetchCompleteItems`), retrieving all pages whenever catalog totals exceed 500 items (tested with 505 items).
  - Server-side `returnStatus` query parameter (`returnable` vs `returned`) filters issued movements in PostgreSQL prior to pagination and calculates exact filtered totals.
  - Changing the issue status filter resets table pagination to page 1.
  - Movement counts separated into distinct per-kind counters (`totalReceipts` and `totalIssues`).

- **R4 (Browser Gate & Scoped Contrast Assertion)**:
  - Scoped Journey 19 selector to `.legacy-workspace nav button[data-tab="item-categories"]` so contrast evaluates actual high-contrast slate text within the workspace.
  - Journeys 20–23 execute through to completion with exit code 0.

---

## 5. Final Review Remediation (F1–F3)

Following the final review (`INVENTORY-FINAL-REVIEW-2026-10-11.md` on `origin/codex/inventory-final-review`), three remaining integrity blockers were remediated and verified:

- **F1 (Database Upgrades & Dedicated Migration 055)**:
  - Reverted `054_inventory_parity.sql` to its exact original state so that existing databases that already executed 054 are not skipped.
  - Added dedicated migration `055_inventory_void_and_ledger_order.sql` to add `void_receipt` kind check constraints, the unique partial index `inventory_movement_void_receipt`, `ledger_seq`, and `balance_after_milli` columns.
  - Existing database movements are backfilled deterministically during migration with trigger protection (`DISABLE TRIGGER ... ENABLE TRIGGER`), ensuring full forward and upgrade compatibility.

- **F2 (Deterministic Per-Item Ledger Order & Consumption Protection)**:
  - Replaced start-time `created_at` window calculations with deterministic monotonic per-item `ledger_seq` and serialized `balance_after_milli`, assigned under the item's `SELECT ... FOR UPDATE` lock.
  - Historical consumption check calculates `MIN(balance_after_milli) WHERE item_id = $1 AND ledger_seq >= origSeq`.
  - Verified against the witness sequence: Transaction B begins at t1, Issue C begins at t2 (t2 > t1) and consumes stock, and B replenishes stock. Stored start-time order would appear unconsumed (100), but actual serialized ledger minimum is 50. Voiding Receipt A is cleanly rejected with `409 Conflict`.
  - Handled timestamp ties deterministically (receipts precede issues if timestamps tie).

- **F3 (Aged Private Operational Upload Lifecycle & Boundaries)**:
  - Updated `CleanupAbandonedAttachments` in `cms_settings.go` to include eligible non-clinical private operational attachments (`patient_id IS NULL AND encounter_id IS NULL`).
  - Fixed sub-second duration truncation in `CleanupAbandonedAttachments` so non-zero test intervals do not fall back to 24 hours.
  - Verified lifecycle boundaries: aged unreferenced private uploads (48h old) are deleted; recent pending uploads (<24h old) are retained; aged referenced private receipts (48h old) are retained; and clinical records (`patient_id != nil`) are strictly preserved.

---

## 6. Ledger & Upgrade Review Remediation (G1–G2)

Following the ledger review (`INVENTORY-LEDGER-REVIEW-2026-10-11.md` on `origin/codex/inventory-ledger-review`), two critical audit findings were resolved and verified:

- **G1 (Migrated Legacy History Conservative Policy)**:
  - Inferred legacy balances cannot be trusted as true serialized history because historical lock-acquisition order cannot be reconstructed from `created_at` timestamps or delta signs under clock skew.
  - Removed the speculative timestamp backfill from migration `055_inventory_void_and_ledger_order.sql`. Pre-existing legacy movements remain strictly with `ledger_seq = 0, balance_after_milli = 0`.
  - Enforced an explicit conservative policy in `services/api/internal/adapters/postgres/inventory.go`: automated `void_receipt` reversals of unverified legacy receipts (`ledger_seq <= 0`) are safely rejected with `domain.ErrConflict` (`409 Conflict`).
  - Completely removed the sequence-zero timestamp reverse-window fallback query, avoiding false proofs of historical ordering.
  - New post-cutover receipts recorded under the item lock receive strictly monotonic positive `ledger_seq > 0` and accurate running balances, allowing genuine consumption checking via `MIN(balance_after_milli) WHERE ledger_seq >= origSeq`.
  - Added a populated prior-054 upgrade regression test in `F1` featuring out-of-order timestamps and intervening consumption, proving legacy receipts cannot bypass consumption protection and verifying that fresh post-upgrade receipts void cleanly.

- **G2 (Database-Enabled Go Suite Fixture Repairs)**:
  - `R1` (`inventory_test.go`): Replaced hardcoded attachment tokens with distinct cryptographically random 32-character hex tokens and 64-character hashes for each test fixture, eliminating unique constraint collisions (`secure_attachment_token_key`).
  - `F2` & `F1` (`inventory_test.go`): Lengthened all test idempotency keys to `>= 16` characters (`key-order-rec-a-0001`, `key-order-rec-b-0001`, `key-tie-rec-00000001`, `key-tie-issue-0000001`, `key-upg-rec-00000001`, `key-upg-void-00000001`), satisfying application validation constraints.
  - `S4` (`cms_settings_test.go`): Linked clinical fixture 4 to a real test patient so it is correctly identified as clinical, and added an explicit aged private nonclinical fixture (Fixture 5) verifying that non-clinical private uploads are properly cleaned up.
  - Browser test fixture runner (`verify-inventory-workspace.mjs`): Updated `insertMovementFixture` to compute and store `ledger_seq` and `balance_after_milli`, ensuring test database fixtures match post-cutover application ledger sequencing.

---

## 7. UI & Localization Parity

- **Unified 4-Tab Workspace (`inventory-workspace.tsx`)**:
  - `items`: Item register, live balance badges (`In Stock`, `Low Stock`, `Out of Stock`), category filtering, search, view modal with movement history, edit modal, delete protection.
  - `item-categories`: Category register, item count, search, create modal, delete action with conflict error banner.
  - `item-stocks`: Stock receipts register, item selector with live balance preview, supplier, store name, reference, unit price, quantity, receipt attachment upload/link, void receipt action, voided status badge.
  - `issued-items`: Issued items register, recipient display name resolution, department, business dates, interactive status badges (`Issued`, `Partial Return (returned/issued)`, `Returned`), server-side return status filter, return modal, void issue action.
- **Review Findings (I1–I5) Implemented**:
  - **I1 (Cutoff Prevention)**: Server-backed pagination controls across all registers; movements search filtering (`search` query param); full catalog retrieval loop for dialog selectors so catalog items (>500 records) are never hidden.
  - **I2 (Empty-Category Guard)**: Category selection change strictly clears child item selection; empty category disables child item selector and submission; no fallback to `items[0]`.
  - **I3 (Source Workflows & Fields)**: Receipt attachments, store name, department, issued-by, business dates, resolved staff display names, audited void receipt with running balance protection, and audited void issue. Accessible `role="alert"` on all dialog error banners.
  - **I4 (Exact Precision & Zero Threshold)**: Exact milli-unit fractional thresholds (1.5 units = 1500 milli) preserved through edit forms without integer rounding; zero threshold (`min={0}`) supported.
  - **I5 (Authentic Dark Mode & Complete Amharic Localization)**: Shell `.legacy-dark` theme toggle integration with computed card background (`#12151f`, `rgb(18, 21, 31)`); 100% Amharic translation for all tabs, action buttons, table headers, search placeholders, and dialog controls.

---

## 8. Verification Results

### 8.1 Quality Gates

| Suite / Gate | Command | Result |
| :--- | :--- | :--- |
| **Go Isolated Database Tests** | `go test ./...` in `services/api` | **PASSED** (100% passing across adapters/postgres, adapters/httpapi, delivery, domain, stripe, privatefiles, including F1, F2, F3, G1, G2 regressions) |
| **Frontend Typecheck** | `npm run typecheck` | **PASSED** (0 errors) |
| **Code Formatting** | `npm run format:check` | **PASSED** (100% clean Prettier) |
| **Frontend Unit Tests** | `npm test` | **PASSED** (8/8 unit tests passed, including translation completeness) |
| **Isolated Inventory Browser Suite** | `npm run test:inventory` | **PASSED** (23/23 journeys passed, 0 failures, exit code 0) |

### 8.2 Browser Suite Journeys (23/23 Passed)

1. `Journey 1`: Admin lands on Inventory workspace with clean 4-tab interface.
2. `Journey 2`: Empty database state verified without hardcoded mocks.
3. `Journey 3`: Category created and persisted in PostgreSQL across reloads.
4. `Journey 4`: Category editing updates database and UI table.
5. `Journey 5`: Item created under category with 0 opening balance.
6. `Journey 6`: Item details modal displays item specifications and 0 balance.
7. `Journey 7`: Stock receipt (+100 units) with financials updates item balance.
8. `Journey 8`: Stock details modal displays receipt metadata and unit cost.
9. `Journey 9`: Stock issue (-40 units) reduces balance to 60 units.
10. `Journey 10`: Over-issue attempt (100 units against 60 balance) rejected with `409 Conflict`.
11. `Journey 11`: Partial return of 15 units (with restock) increases balance to 75 units; badge updates to `Partial Return (15/40)`.
12. `Journey 12`: Over-return attempt (30 units against 25 remaining) rejected with `409 Conflict`.
13. `Journey 13`: Remaining return of 25 units (with restock) restores balance to 100 units; badge updates to `Returned`.
14. `Journey 14`: Non-restocked return (damaged items) records audit movement with delta 0 without inflating physical stock balance.
15. `Journey 15`: Row-level locking serializes concurrent requests and prevents overspending.
16. `Journey 16`: Referenced deletion blocked with `409 Conflict`; unreferenced category and item deleted cleanly.
17. `Journey 17`: Unauthorized doctor role blocked from mutations with `403 Forbidden`.
18. `Journey 18`: Fresh browser session verifies 100% persisted state from PostgreSQL.
19. `Journey 19`: Theme & Amharic Localization parity (Finding I5 regression) verifying real `.legacy-dark` computed card styling, high-contrast tab navigation, and complete Amharic controls.
20. `Journey 20`: Findings I1 and R3 regression — full catalog pagination across 505 items in modal selectors; server-backed returnStatus filtering before pagination with page reset; discoverability of older outstanding issues behind returned issues.
21. `Journey 21`: Finding I2 regression — empty category disables submission and clears child selection; prevents unintended stock mutations.
22. `Journey 22`: Findings I3, R1, R2, F3, and G1 regression — real file attachment upload/download bytes; direct attachment retirement conflict (409); aged cleanup preservation of bound receipts while deleting aged abandoned private uploads; shared settings asset displacement preservation; rejection of clinical/missing tokens; durable linked `void_receipt` reversal with `original_id`; UI voided badge and disabled buttons; repeated void rejection (409); independent voiding of multiple receipts on the same item; and intervening consumption protection across subsequent replenishment (409).
23. `Journey 23`: Finding I4 regression — fractional (1500 milli = 1.5 units) and zero (0 milli = 0 units) reorder thresholds preserved without rounding drift.

---

## 9. Visual Evidence

The following authentic screenshots were captured during isolated browser execution:

- `01_inventory_admin_landing.png`: Admin landing on inventory workspace.
- `02_inventory_empty_state.png`: Clean empty states across tabs before data creation.
- `03_inventory_category_created.png`: Category created and listed in table.
- `04_inventory_item_created.png`: Item registered with category linkage and zero balance.
- `05_inventory_item_details.png`: Item details modal showing attributes and history.
- `06_inventory_stock_received.png`: Stock received (+100 units) reflecting live balance.
- `07_inventory_item_issued.png`: Stock issued (-40 units) reflecting remaining 60 units.
- `08_inventory_over_issue_rejected.png`: Insufficient stock rejection banner preventing balance drift.
- `09_inventory_partial_return.png`: Partial return modal and updated 75 balance.
- `10_inventory_remaining_return_completed.png`: Complete return updating issue status badge to Returned.
- `11_inventory_record_protection.png`: Modal rejection preventing deletion of referenced item/category.
- `12_inventory_dark_mode.png`: Authentic dark mode showing real `.legacy-shell.legacy-dark` with high-contrast row text, navigation tabs, pill counters, and vibrant WCAG-compliant status badges.
- `13_inventory_amharic_localization.png`: Authentic Amharic localization covering all tabs, header subtitles ("የሕክምና ቁሳቁሶች ክምችት ሚዛን እና የድጋሚ ማዘዣ መጠንን ይከታተሉ"), action buttons, table headers, search placeholders, and dialog controls.

---

## 10. Remaining Limitations & Exclusions

- **Standalone Purchase Orders Aggregate**: Basic inventory tracks stock receipts (`item-stocks`) and issues (`issued-items`). Multi-step procurement approval hierarchies and purchase order contracts belong to future procurement modules and are not part of the inventory workspace contract.
- **Lot Number & Supplier-Contact Tracking**: The inspected movement contract does not include separate lot expiration registers or supplier contact books; supplier name and store name are captured as transaction metadata.
- **Smart Cards & Odontogram**: Kept strictly out of scope for this batch as instructed. Main branch remains unchanged at `5cb7f81`.
