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
| **Item Stocks (Receipts)** | `item-stocks.index`<br>`item_stocks/index.blade.php` | `inventory_movement`<br>(`kind = 'receive'`) | `POST /v1/inventory/movements`<br>`GET /v1/inventory/movements` | **COMPLETE** | Supplier, store name (`store_name`), reference number, unit price, quantity, attachment URL (`attachment_url`), audit trail. Audited Void Receipt via stock write-off movement (`kind = 'writeoff'`), blocked with `409 Conflict` if stock has already been consumed. |
| **Issued Items** | `issued-items.index`<br>`issued_items/index.blade.php` | `inventory_movement`<br>(`kind = 'issue'`) | `POST /v1/inventory/movements`<br>`GET /v1/inventory/movements` | **COMPLETE** | Staff recipient linkage (`recipient_id`), resolved recipient display name ("Nurse Genet Lemma"), department (`department`), issued date (`issued_date`), return due date (`return_due_date`), issued by (`issued_by`). Decreases `balance_milli` with strict check preventing negative balance. Audited Void Issue via restocked reversal movement (`kind = 'return'`). |
| **Item Returns** | `issued-items.return`<br>`issued_items/return_modal.blade.php` | `inventory_movement`<br>(`kind = 'return'`) | `POST /v1/inventory/movements` | **COMPLETE** | Partial and full returns. Links `original_id`. Rejects over-return with `409 Conflict` (`STATE_CONFLICT`). Supports `restock: false` for damaged/expired items without balance inflation. Interactive partial return badge opens return modal directly. |
| **Suppliers** | Supplier metadata in stock receipts | `inventory_movement.supplier` | `/v1/inventory/movements` | **COMPLETE** | Captured and persisted on receipt transactions; filterable and displayed in item stocks register. |
| **Stores** | Store metadata in receipts | `inventory_movement.store_name` | `/v1/inventory/movements` | **COMPLETE** | Persisted across receipts for multi-store audit tracking. |
| **Receipt Attachments** | Invoice / receipt document link | `inventory_movement.attachment_url` | `/v1/attachments`<br>`/v1/inventory/movements` | **COMPLETE** | Authorized upload lifecycle via `/api/hms/attachments` and direct link in table/details modal. |

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

5. **Referenced Record Deletion Protection**:
   - `DELETE /v1/inventory/categories/{id}` checks `SELECT EXISTS(SELECT 1 FROM inventory_item WHERE category_id = $1)`. If items exist, returns `domain.ErrInUse` (`409 Conflict`, `RECORD_IN_USE`).
   - `DELETE /v1/inventory/items/{id}` checks `SELECT EXISTS(SELECT 1 FROM inventory_movement WHERE item_id = $1)`. If movements exist, returns `domain.ErrInUse` (`409 Conflict`, `RECORD_IN_USE`).
   - Unreferenced categories and items delete cleanly and record `inventory_category.deleted` / `inventory_item.deleted` audit events.

---

## 4. UI & Localization Parity

- **Unified 4-Tab Workspace (`inventory-workspace.tsx`)**:
  - `items`: Item register, live balance badges (`In Stock`, `Low Stock`, `Out of Stock`), category filtering, search, view modal with movement history, edit modal, delete protection.
  - `item-categories`: Category register, item count, search, create modal, delete action with conflict error banner.
  - `item-stocks`: Stock receipts register, item selector with live balance preview, supplier, store name, reference, unit price, quantity, receipt attachment upload/link, void receipt action.
  - `issued-items`: Issued items register, recipient display name resolution, department, business dates, interactive status badges (`Issued`, `Partial Return (returned/issued)`, `Returned`), return modal, void issue action.
- **Review Findings (I1–I5) Implemented**:
  - **I1 (Cutoff Prevention)**: Server-backed pagination controls across all registers; movements search filtering (`search` query param); large catalog retrieval (`limit=500`) for dialog selectors so older items/categories/staff are never hidden.
  - **I2 (Empty-Category Guard)**: Category selection change strictly clears child item selection; empty category disables child item selector and submission; no fallback to `items[0]`.
  - **I3 (Source Workflows & Fields)**: Receipt attachments, store name, department, issued-by, business dates, resolved staff display names, audited void receipt (with 409 consumption check), and audited void issue. Accessible `role="alert"` on all dialog error banners.
  - **I4 (Exact Precision & Zero Threshold)**: Exact milli-unit fractional thresholds (1.5 units = 1500 milli) preserved through edit forms without integer rounding; zero threshold (`min={0}`) supported.
  - **I5 (Authentic Dark Mode & Complete Amharic Localization)**: Shell `.legacy-dark` theme toggle integration with computed card background (`#12151f`, `rgb(18, 21, 31)`); 100% Amharic translation for all tabs, action buttons, table headers, search placeholders, and dialog controls.

---

## 5. Verification Results

### 5.1 Quality Gates

| Suite / Gate | Command | Result |
| :--- | :--- | :--- |
| **Go Isolated Database Tests** | `go test ./...` in `services/api` | **PASSED** (100% passing across adapters/postgres, adapters/httpapi, delivery, domain, stripe, privatefiles) |
| **Frontend Typecheck** | `npm run typecheck` | **PASSED** (0 errors) |
| **Code Formatting** | `npm run format:check` | **PASSED** (100% clean Prettier) |
| **Frontend Unit Tests** | `npm test` | **PASSED** (8/8 unit tests passed, including translation completeness) |
| **Isolated Inventory Browser Suite** | `npm run test:inventory` | **PASSED** (23/23 journeys passed, 0 failures, exit code 0) |

### 5.2 Browser Suite Journeys (23/23 Passed)

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
19. `Journey 19`: Theme & Amharic Localization parity (Finding I5 regression) verifying real `.legacy-dark` computed card styling and complete Amharic controls.
20. `Journey 20`: Finding I1 regression — pagination across categories, items, and movements; server-backed search; older outstanding issue discoverability and return workflow (>25 records).
21. `Journey 21`: Finding I2 regression — empty category disables submission and clears child selection; prevents unintended stock mutations.
22. `Journey 22`: Finding I3 regression — source fields (`store_name`, `department`, `issued_date`, `return_due_date`, `issued_by`), receipt attachments, recipient name resolution, and void workflows (rejection after consumption with `409 Conflict`, unconsumed receipt void write-off, and issue void reversal).
23. `Journey 23`: Finding I4 regression — fractional (1500 milli = 1.5 units) and zero (0 milli = 0 units) reorder thresholds preserved without rounding drift.

---

## 6. Visual Evidence

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
- `12_inventory_dark_mode.png`: Authentic dark mode showing real `.legacy-shell.legacy-dark` and `#12151f` card background.
- `13_inventory_amharic_localization.png`: Authentic Amharic localization covering all tabs, action buttons, table headers, search placeholders, and dialog controls.

---

## 7. Remaining Limitations & Exclusions

- **Standalone Purchase Orders Aggregate**: Basic inventory tracks stock receipts (`item-stocks`) and issues (`issued-items`). Multi-step procurement approval hierarchies and purchase order contracts belong to future procurement modules and are not part of the inventory workspace contract.
- **Lot Number & Supplier-Contact Tracking**: The inspected movement contract does not include separate lot expiration registers or supplier contact books; supplier name and store name are captured as transaction metadata.
- **Smart Cards & Odontogram**: Kept strictly out of scope for this batch as instructed. Main branch remains unchanged at `5cb7f81`.
