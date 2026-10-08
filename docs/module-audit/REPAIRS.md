# Verified repairs after the audit

Historical observations in test-results.json and PAGE-RESULTS.md remain unchanged. This register records subsequent fixes; it does not mark whole modules complete.

## 2026-10-07: endpoint and permission contracts

- Doctor absence register: admin can list all absences, with optional doctor filter and 25-row pagination. Doctors still require their own ID; other roles cannot read this register. Repository access enforces the same boundary.
- Inventory movement register: admin can list all movements or filter by item UUID, with pagination. All other roles remain denied, including direct repository calls. Existing per-item behavior remains tested.
- General settings: enabled the existing authenticated Go endpoint through the Next proxy. Bulk saves validate all entries before writing, reject normalized duplicate keys, and persist all fields/audit events in one transaction. Tests force a database failure to prove rollback.
- CMS permissions are now advertised consistently; removed duplicate messages permission advertising.
- Verification: full Go suite passed with real isolated PostgreSQL; targeted real Better Auth/Next/Go/browser run passed 49 assertions across nine roles, anonymous access, settings persistence/validation, and six affected pages. Temporary schemas/services were removed. No public schema migration or external credentials changed.

These repairs do not establish package/insurance endpoints, complete settings UI integration, or original all-role CRUD parity. Continue using ACCEPTANCE-MATRIX.md for the unverified workflows.

## 2026-10-07: inventory mutation integration

- Removed seeded inventory records, local success rows, and client-invented stock balance adjustments. Empty responses now produce empty lists; failed loads show an error.
- Category/item creation and receive/issue forms use persisted API responses and reload authoritative lists. Errors remain visible in the open form; repeated submit is blocked while saving. Stock request keys are retained for an unchanged failed request.
- Replaced free-text department recipients with a searchable selection of active staff IDs. Added accessible form names and Amharic messages for the new status/error controls.
- Verification: typecheck and eight frontend tests passed. Combined real-login/API/browser run passed 55 checks, including six inventory persistence/failure scenarios: category reload, item opening balance, receive/issue reconciliation, over-issue rejection, HTTP failure, and network failure. See inventory-repairs.json. Public development records were not used or changed.
- Remaining inventory parity: original supplier/store catalogs, full pagination/search, edit/archive UI, returns/write-offs UI, original field/layout comparison and complete role workflows. These are not marked complete by the create/receive/issue checks.

## 2026-10-08: package backend and contract parity (Step 1)

- Migration `049_packages_catalog.sql`: implemented `package` and `package_service` tables with strict constraints, minor-unit money types, and foreign key linkage to `ipd_admission_details(package_id)` with `ON DELETE RESTRICT`.
- Authoritative calculation: server recalculates line amounts, subtotal, explicit half-up percentage discount rounding, and total amount. Client totals are never trusted.
- Transactional mutations & ownership: transactional create and update enforce that submitted line IDs belong strictly to the target package. Cross-parent child line IDs are rejected with conflict. Forced child-write failure rolls back all parent and line rows atomically.
- Admission in-use protection: deletion is strictly rejected if referenced by patient admissions (`ipd_admission_details`), returning `409 Conflict` (`RECORD_IN_USE`).
- Scoped permissions & endpoints: `packages.manage` for `admin` and `receptionist`; `packages.read` for `admin`, `receptionist`, `doctor`, `case_manager`, `patient`. Anonymous access returns 401; unauthorized roles (`nurse`, `accountant`, etc.) receive 403.
- Verification: isolated-schema PostgreSQL test suite (`TestClinicalTransactions/testPackages`) verified multi-line creation, line update/add/remove, duplicate name rejection, invalid service references, cross-parent line ID rejection, atomic rollback, admission in-use protection, role authorization, concurrency, search, and pagination.

## 2026-10-08: package frontend integration (Step 2)

- Connected `ServicesWorkspace` to live `/v1/packages` API: create, edit, delete, and list operations persist authoritatively in PostgreSQL.
- Removed synthetic mock arrays and optimistic local saves: table reflects server-calculated totals (102.00 ETB), child service line breakdowns, and discounts.
- Live modal editing & deletion: package edit modal supports modifying service lines, quantities, rates, and discounts with live recalculation preview. Deletion of in-use packages displays server conflict message (`409 Conflict`), while unreferenced packages delete cleanly (`200 OK`).
- Multi-context verification: freshly authenticated browser contexts verify persisted records and child lines surviving full page reloads.

## 2026-10-08: insurance backend & frontend integration (Steps 3 & 4)

- Migration `050_insurance_catalog.sql`: implemented `insurance` and `insurance_disease` tables with minor-unit monetary precision (`service_tax_minor`, `hospital_rate_minor`, `disease_charge_minor`, `total_minor`), foreign key linkage to `ipd_admission_details(insurance_id)` with `ON DELETE RESTRICT`, and atomic disease replacement transactions.
- Authoritative calculation: server recalculates base sum, disease charges, discount amounts with half-up rounding, and net total.
- Real `/v1/insurances` API integration: full-page creation form, dynamic disease rows, details modal, edit modal with disease line updates, and active/inactive status toggle.
- In-use deletion protection: deleting an insurance linked to patient admissions is blocked with `409 Conflict` (`RECORD_IN_USE`).

## 2026-10-08: review corrections & hardening (R1–R7)

- **R1 (Child Line Deduplication)**: Validator and repository strictly reject duplicate child-line IDs before any mutations. Persisted child line amounts reconcile exactly to parent totals.
- **R2 (Checked Monetary Arithmetic)**: Replaced unchecked int64 calculations with checked arithmetic across `packages` and `insurances`, preventing integer overflow even under 100% discounts or extreme values.
- **R3 (Explicit Zero Rate Preservation)**: Package creation and editing distinguish between omitted rates and explicitly submitted `0` rates; catalog defaults are only applied when rates are genuinely omitted.
- **R4 (Complete Exports)**: Implemented batched streaming exports (`ExportPackages` and `ExportInsurances`) via `/v1/packages-export` and `/v1/insurances-export` supporting up to 5,000 records without silent truncation to 25 rows. Tested with 35+ records in Go and Playwright.
- **R5 (Server Pagination & Catalog Lookup)**: Wired server-side `page`, `pageSize`, `search` (300ms debounce), and server-reported `total` across Packages and Insurances tabs. Package modal service selector performs bounded multi-page active catalog loading (`status=1`, limit=100 per page up to 5,000 items) ensuring active services beyond the first 100 are selectable. Existing selected archived/inactive service references are preserved for display in edit forms without permitting new invalid selections. Includes in-modal error alert banners with retry recovery.
- **R6 (Failure Isolation & Retry States)**: Separated per-module loading, error, and connection states. Failed requests display visible error alerts with retry triggers without masquerading as empty catalogs.
- **R7 (Idempotent Status Updates)**: `PATCH /v1/insurances/{id}/status` accepts `{ "status": 0 | 1 }` payload and updates status idempotently; invalid statuses return `422 Unprocessable Entity`; UI disables toggles during in-flight requests and avoids duplicate toggles.

## Retained Coverage & Documentation Gaps

- **Admission UI vs SQL Fixtures**: In-use deletion tests insert `ipd_admission_details` links directly via SQL (including OPD fixtures), verifying database FK restrict behavior. Standalone original patient admission workflows selecting and persisting package/insurance IDs remain an unfinished separate module.
- **Catalog Visibility Policy**: `packages.read` and `insurances.read` permissions intentionally expose catalog visibility to `admin`, `receptionist`, `doctor`, `case_manager`, and `patient` roles, while modification permissions (`packages.manage`, `insurances.manage`) remain strictly confined to `admin` and `receptionist`.
- **Concurrency & Version Policy**: Database row locks serialize concurrent mutations under a last-write-wins policy with `clock_timestamp()` updates rather than optimistic concurrency control (no version tags).
- **Public Development Isolation**: All integration checks ran against transient QA schemas or test databases without altering public development records.

## Next implementation order

1. Doctor workspace: remove remaining preview rows/saves and align doctor account creation, department, and absence forms with their actual contracts.
2. Settings frontend: connect the original field form to the now-available atomic endpoint, including the appropriate image storage contract.


