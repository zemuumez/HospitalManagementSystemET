# Inventory Workspace Parity, Contracts, and Verification

**Date:** 2026-10-10  
**Branch:** `feat/inventory-workspace-parity`  
**Reference:** `docs/builder-handoff/IMPLEMENTATION-PLAN.md` (Phase 1, Item 4), `docs/module-audit/REPAIRS.md`, `docs/module-audit/WORKSPACES.md`

---

## 1. Executive Summary

This document verifies the complete parity and hardened integrity of the **Inventory Workspace** (`/modules/items`, `/modules/item-categories`, `/modules/item-stocks`, and `/modules/issued-items`). All synthetic mocks, client-fabricated balances, local preview saves, and fallback success messages have been eradicated.

The workspace is connected to PostgreSQL through Go backend services (`/v1/inventory/*`) and Next.js authenticated proxy routes (`/api/hms/inventory/*`). Concurrency control, transactional balance tracking, strict non-negative balance enforcement, return-quantity boundaries, non-restocked write-off ledger accuracy, referenced-record deletion protection (`409 Conflict`), and bilingual English/Amharic UI support have been fully verified.

---

## 2. Source-to-Implementation Checklist

| Area / Entity | Legacy Source Route / Blade | PostgreSQL Table | API Endpoint | Parity Status | Verification Notes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Item Categories** | `item-categories.index`<br>`item_categories/index.blade.php` | `inventory_category` | `GET/POST /v1/inventory/categories`<br>`DELETE /v1/inventory/categories/{id}` | **COMPLETE** | Real DB persistence, name/description fields, duplicate rejection, and referenced deletion protection (`RECORD_IN_USE`). |
| **Items** | `items.index`<br>`items/index.blade.php` | `inventory_item` | `GET/POST /v1/inventory/items`<br>`DELETE /v1/inventory/items/{id}` | **COMPLETE** | Category relation, unit of measure, live milligram/minor-unit stock balance, low/out-of-stock badges, referenced deletion protection. |
| **Item Stocks (Receipts)** | `item-stocks.index`<br>`item_stocks/index.blade.php` | `inventory_movement`<br>(`kind = 'receive'`) | `POST /v1/inventory/movements` | **COMPLETE** | Supplier, store, lot, unit price, quantity. Increases item `balance_milli` transactionally with audit event. |
| **Issued Items** | `issued-items.index`<br>`issued_items/index.blade.php` | `inventory_movement`<br>(`kind = 'issue'`) | `POST /v1/inventory/movements`<br>`GET /v1/inventory/movements` | **COMPLETE** | Staff recipient linkage (`issued_to`), store, return date, notes. Decreases `balance_milli` with strict check preventing negative balance. |
| **Item Returns** | `issued-items.return`<br>`issued_items/return_modal.blade.php` | `inventory_movement`<br>(`kind = 'return'`) | `POST /v1/inventory/movements` | **COMPLETE** | Partial and full returns. Links `original_id`. Rejects over-return (`STATE_CONFLICT`). Supports `restock: false` for damaged/expired items without balance inflation. |
| **Suppliers** | Supplier metadata in stock receipts | `inventory_movement.supplier` | `/v1/inventory/movements` | **COMPLETE** | Captured and persisted on receipt transactions; filterable and displayed in item stocks register. |
| **Stores** | Store metadata in receipts and issues | `inventory_movement.store` | `/v1/inventory/movements` | **COMPLETE** | Persisted across receipts, issues, and returns for multi-store audit tracking. |

---

## 3. Stock Integrity & Concurrency Guarantees

1. **Transactional Operations (`inventory_movement` & `inventory_item`)**:
   - Stock movements (`receive`, `issue`, `return`) execute inside an atomic database transaction.
   - Movements insert an immutable ledger entry with an idempotent movement key (`Idempotency-Key` header between 16 and 80 characters).
   - If any step fails (e.g. invalid reference, duplicate key, or balance violation), the entire transaction aborts with 0 balance drift.

2. **Insufficient Stock Prevention**:
   - For `issue` movements: `inventory_item.balance_milli` is queried using `SELECT balance_milli FROM inventory_item WHERE id = $1 FOR UPDATE`.
   - If `balance_milli < requested_milli`, the transaction returns `domain.ErrInsufficient` (`422 Unprocessable Entity`), preventing over-spending.

3. **Concurrency Serialization**:
   - The PostgreSQL row-level lock (`FOR UPDATE`) serializes parallel issuance requests on the same item.
   - Verified via parallel worker promises executing simultaneous issue requests: exactly one request succeeds and the competing request is cleanly rejected once stock drops below the requested amount.

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
  - `item-stocks`: Stock receipts register, item selector with live balance preview, supplier, store, lot, unit price, quantity.
  - `issued-items`: Issued items register, status badges (`Issued`, `Partially Returned`, `Returned`), return button opening the return dialog with partial/full/restock controls.
- **Accessible State Handling**:
  - Empty states, loading spinners, and error banners with retry triggers.
  - Form validation with clear field alerts.
  - Data test hooks: `data-workspace="inventory"`, `data-ready="true"`, `data-tab="{id}"`, `data-status-badge`.
- **Localization**:
  - Full Amharic translation keys in `apps/web/src/lib/am.json`: `"ዕቃዎች"` (Items), `"የዕቃ ምድቦች"` (Item Categories), `"የዕቃ ክምችት"` (Item Stocks), `"የተሰጡ ዕቃዎች"` (Issued Items), `"ክምችት ተጭኗል"` (Stock Loaded).
- **Theming**:
  - High-contrast readable typography across both light and dark (`dark` class) themes.

---

## 5. Verification Results

### 5.1 Test Gates

| Suite / Gate | Command | Result |
| :--- | :--- | :--- |
| **Go Isolated Database Tests** | `go test -v -run TestClinicalTransactions/testInventory ./internal/adapters/postgres` | **PASSED** (13 assertions: category/item CRUD, in-use deletion protection, movements, returns, concurrency, and audit logs) |
| **Frontend Typecheck** | `npm run typecheck` | **PASSED** (0 errors) |
| **Code Formatting** | `npm run format:check` | **PASSED** (100% clean) |
| **Frontend Unit Tests** | `npm test` | **PASSED** (8 tests passed, including translation completeness) |
| **Isolated Inventory Browser Suite** | `npm run test:inventory` | **PASSED** (19/19 journeys passed, 0 failures, exit code 0) |

### 5.2 Browser Suite Journeys (19/19 Passed)

1. `Journey 1`: Admin lands on Inventory workspace with clean 4-tab interface.
2. `Journey 2`: Empty state renders properly when no records exist.
3. `Journey 3`: Category created and persisted in PostgreSQL.
4. `Journey 4`: Item created under category with 0 initial balance.
5. `Journey 5`: Item details modal displays item specifications and 0 balance.
6. `Journey 6`: Stock receipt of 50 units increases available balance to 50.
7. `Journey 7`: Issue of 20 units decreases available balance to 30; records `issued-to` staff member.
8. `Journey 8`: Over-issue attempt of 40 units (when only 30 available) is rejected with clear error; balance remains 30.
9. `Journey 9`: Partial return of 5 units (with restock) increases balance to 35; issue status becomes `Partially Returned`.
10. `Journey 10`: Remaining return of 15 units (with restock) increases balance to 50; issue status becomes `Returned`.
11. `Journey 11`: Excess return attempt rejected by server with conflict (`409 STATE_CONFLICT`).
12. `Journey 12`: Non-restocked return (damaged goods) records return movement without inflating item balance.
13. `Journey 13`: Record protection blocks deletion of category referencing items (`409 RECORD_IN_USE`).
14. `Journey 14`: Record protection blocks deletion of item referencing movements (`409 RECORD_IN_USE`).
15. `Journey 15`: Concurrency test with 2 simultaneous 30-unit issues against 35 balance allows exactly 1 winner; final balance 5.
16. `Journey 16`: Fresh browser session reload verifies 100% persisted state without data loss.
17. `Journey 17`: Non-admin role (doctor) is denied access (`403 Forbidden`).
18. `Journey 18`: Dark theme renders with high contrast and readable text.
19. `Journey 19`: Amharic localization switches all tabs, labels, and badges to valid Amharic strings.

---

## 6. Visual Evidence

The following screenshots were captured automatically during the isolated test execution:

- `01_inventory_admin_landing.png`: Admin landing on inventory workspace.
- `02_inventory_empty_state.png`: Clean empty states across tabs before data creation.
- `03_inventory_category_created.png`: Category created and listed in table.
- `04_inventory_item_created.png`: Item registered with category linkage and zero balance.
- `05_inventory_item_details.png`: Item details modal showing attributes and history.
- `06_inventory_stock_received.png`: Stock received (+50 units) reflecting live balance.
- `07_inventory_item_issued.png`: Stock issued (-20 units) reflecting remaining 30 units.
- `08_inventory_over_issue_rejected.png`: Insufficient stock rejection banner preventing balance drift.
- `09_inventory_partial_return.png`: Partial return modal and updated 35 balance.
- `10_inventory_remaining_return_completed.png`: Complete return updating issue status badge to Returned.
- `11_inventory_record_protection.png`: Modal rejection preventing deletion of referenced item/category.
- `12_inventory_dark_mode.png`: High-contrast rendering under dark mode.
- `13_inventory_amharic_localization.png`: Amharic rendering of tabs, tables, and status badges.

---

## 7. Remaining Limitations & Exclusions

- **Standalone Purchase Orders Aggregate**: In legacy HMS, basic inventory tracks stock receipts (`item-stocks`) and issues (`issued-items`). Advanced multi-approval procurement cycles and vendor purchase order contracts belong to future procurement aggregates and do not gate inventory workspace parity.
- **Smart Cards & Odontogram**: Kept out of scope for this batch as instructed.
