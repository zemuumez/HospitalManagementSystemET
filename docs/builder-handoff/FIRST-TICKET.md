# First complete vertical slice: packages and insurance

## Why this is next

The source has these aggregates and frontend forms, while the audit found missing endpoints and local-success behavior in ServicesWorkspace. Fixing them unlocks real admission references. Recheck current HEAD first in case another builder has already started this ticket.

## Original package behavior

- Package has unique required name, description, required integer discount and total amount. Do not invent a package status column merely because a current UI type has one.
- Service lines contain service ID, quantity, rate and amount. The source picker uses active services; each line is validated against the child model rules.
- Repository recalculates line amount = quantity × rate. Subtotal is sum of line amounts; total = subtotal − subtotal × discount / 100.
- Store/update use a transaction around parent and children. Update submits line IDs; preserve correct ownership of those IDs.
- Show includes service lines and services. Export exists. Deletion is blocked if patient admissions reference the package.

## Original insurance behavior

- Required name, insurance number/code, service tax, hospital rate and integer discount. Name is application-unique; the dump index is not necessarily a unique constraint. Discount nullability in SQL differs from create validation.
- Insurance has remark, total, active/inactive status and optional currency symbol in the dump. Disease lines contain name and charge. Source update replaces disease lines in a transaction. Source UI leaves at least one required disease row.
- **The original JavaScript adds service tax as a monetary amount, not a percentage:** base = service tax + hospital rate + sum(disease charges); total = base − base × discount / 100.
- That JavaScript uses integer parsing for charges and can truncate fractions; the source repository trusts the submitted insurance total. These are historical weaknesses, not recommended target behavior. Calculate exact money on the server and record the intentional decimal/rounding correction.
- Active/deactivate and export exist. Deletion is blocked when referenced by admissions.

## Implementation steps

1. Read the corresponding rows in TABLES.md, MODEL-RULES.md, LEGACY-ACTIONS.md and mapping register. Record all field sizes/nullability, child rules, actions and role groups in the new contract. Original shared show routes include Admin/Doctor/Case Manager/Patient/Receptionist; administration routes include receptionist. Do not expose unrelated records because the original route group is broad.
2. Decide endpoint shapes consistent with the existing API, for example package/insurance collections, detail and status action. These are proposed endpoints, not a claim that they exist. Add search, paging and typed error responses.
3. Add parent/line storage and needed admission FKs. Preserve immutable applied price snapshots when a package is used. Do not cascade-delete historical lines when a service is archived.
4. Put normalization, validation, exact totals and permitted transitions in Go application/domain logic. Enforce bounds and non-negative prices/quantities; choose whether fractional quantities are supported based on source semantics and document the decision. Reject invalid discount ranges and child IDs owned by another parent.
5. Execute parent, child replacement/update, audit and outbox effects atomically. Lock/version as needed to prevent lost concurrent edits. Handle in-use deletion conflicts as readable errors.
6. Connect forms and lookups, use authoritative totals, keep invalid/failed forms open, block duplicate submits, remove preview rows/local-success fallbacks, and reload persisted records. Preserve original field ordering and navigation.
7. Update OpenAPI, module contract, field parity rows, acceptance register and repair evidence. Commit only the verified vertical slice; do not mark the entire Services workspace complete.

## Required tests

- Create package with two active services, verify exact totals, detail lines and browser reload; update one/remove one/add one and verify persisted ownership.
- Reject duplicate names, invalid references, invalid discount/quantity/rate and cross-parent line IDs with no partial writes.
- Force a child-write failure and verify parent/lines/audit roll back.
- Create insurance with fractional charges and additive service tax; verify selected rounding and total, update disease rows, toggle status and reload.
- Referenced package/insurance deletion returns conflict and preserves admission history; unused deletion/archive follows the recorded policy.
- Probe all nine roles and anonymous users for list/detail/create/update/delete/export/status; include patient ownership and guessed IDs.
- Test concurrent edit policy, duplicate submit/network retry, server/network failures and a new browser context.
- Verify export/print content and totals, no secret/patient disclosure, and empty-list/error states.
