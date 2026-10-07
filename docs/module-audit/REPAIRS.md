# Verified repairs after the audit

Historical observations in test-results.json and PAGE-RESULTS.md remain unchanged. This register records subsequent fixes; it does not mark whole modules complete.

## 2026-10-07: endpoint and permission contracts

- Doctor absence register: admin can list all absences, with optional doctor filter and 25-row pagination. Doctors still require their own ID; other roles cannot read this register. Repository access enforces the same boundary.
- Inventory movement register: admin can list all movements or filter by item UUID, with pagination. All other roles remain denied, including direct repository calls. Existing per-item behavior remains tested.
- General settings: enabled the existing authenticated Go endpoint through the Next proxy. Bulk saves validate all entries before writing, reject normalized duplicate keys, and persist all fields/audit events in one transaction. Tests force a database failure to prove rollback.
- CMS permissions are now advertised consistently; removed duplicate messages permission advertising.
- Verification: full Go suite passed with real isolated PostgreSQL; targeted real Better Auth/Next/Go/browser run passed 49 assertions across nine roles, anonymous access, settings persistence/validation, and six affected pages. Temporary schemas/services were removed. No public schema migration or external credentials changed.

These repairs do not establish package/insurance endpoints, complete settings UI integration, or original all-role CRUD parity. Continue using ACCEPTANCE-MATRIX.md for the unverified workflows.
