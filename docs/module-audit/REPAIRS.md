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

## Next implementation order

1. Packages and insurances: implement source-backed parent/line aggregates with authoritative totals, transactional writes, in-use deletion checks and role tests. Remove remaining false local success in ServicesWorkspace as part of that integration.
2. Doctor workspace: remove remaining preview rows/saves and align doctor account creation, department and absence forms with their actual contracts.
3. Settings frontend: connect the original field form to the now-available atomic endpoint, including the appropriate image storage contract.
4. Smart-card templates/generation and chart-level odontograms: implement missing persistence/history without losing original codes.
5. IPD/OPD detail workflows and remaining ACCEPTANCE-MATRIX families: compare fields/rules, then implement and test each complete workflow. All-role CRUD parity is still open.
