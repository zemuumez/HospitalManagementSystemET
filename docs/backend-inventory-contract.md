# General inventory backend

Section 3 backend only. Reviewed legacy Item, ItemStock, IssuedItem, inventory routes and the PDF inventory description. Medicines remain in their separate pharmacy workflow.

## Contracts

Administrator-only routes under `/v1/inventory`:

- GET/POST `categories`, PATCH `categories/{id}`: name, description, active and version.
- GET/POST `items`, PATCH `items/{id}`: categoryId, name, unit, description, reorderMilli, active and version. GET supports search, lowStock=true and page.
- GET `movements?itemId=...&page=1`: retained stock history.
- POST `movements`: itemId, kind, quantityMilli, recipientId, originalId, supplier, storeName, reference, costMinor, restock and reason; Idempotency-Key required.

Quantities use integer thousandths of the named unit: 1250 means 1.250 kg if unit is kg. Cost is the total receipt cost in ETB minor units. No implicit unit conversions or floating-point balance calculations occur. Lists are paginated at 25 rows. PATCH requires the current version. No destructive delete API is exposed.

`receive` requires a purchase reference and records supplier/store/cost. `issue` requires an active non-patient staff recipient. `return` derives recipient from its original issue; callers cannot substitute one. Total returns cannot exceed issued quantity. Only an explicit administrator restock decision increases balance; damaged returns can be recorded without restocking. `writeoff` decreases stock with a mandatory reason. All movements require a reason. Original issue dates are server timestamps; backdated issue/return dates, department masters, attachments and full original export layouts remain open.

## Integrity

Item locks serialize receipts/issues/returns. Stock cannot fall below zero or exceed the defined limit. Deferred database checks reconcile balance against immutable movement deltas. Reusing a request key returns the original movement only for matching input. Duplicate item/supplier/purchase-reference receipts conflict. Historical units cannot be changed; nonempty items cannot be archived. Inactive categories/items reject new receipts/issues. Returns can still be received against historic issues after archival, with explicit assessment.

Low-stock filtering is server-backed and respects each item's reorder threshold. It is a query, not yet a scheduled email/SMS alert. Authorization starts administrator-only, consistent with a conservative source-role mapping; delegated inventory roles and department custody need explicit policy before wider access.

## Verification

The disposable PostgreSQL suite verifies fractional-unit receipt persistence, key replay/conflict, concurrent stock depletion, bounded reusable/damaged returns, original recipient derivation, low-stock results, versioned edits, unit-change and archive restrictions, balance tampering, ledger immutability, role denial, HTTP strict JSON and Origin checks. Go tests and vet passed. Full uploads, purchase document/finance integration, department/date field parity, inventory count approval and scheduled alerts remain pending; the frontend is unchanged.
