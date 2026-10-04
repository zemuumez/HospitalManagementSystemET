# Diagnostics backend contract

Section 3 backend implementation; no new diagnostic frontend integration.

## Source review and model

Reviewed legacy PathologyTest, RadiologyTest, PathologyParameter, PathologyParameterItem and PathologyTestRepository. The source combines patient-linked test definitions and result rows, and replaces parameter results on edits. This implementation separates reusable definitions, encounter orders, sample references, result revisions and review/release events.

Migration 009 supports pathology/radiology definitions with name, short name, category label, method, report days, exact charge and ordered parameters (name, unit label, reference-range text, number/text value type). Definitions/parameters are retained; ordered definitions cannot receive additional parameters. Dedicated category/unit masters, editable draft catalog versions, original charge-category linkage, subcategory fields and full export/print parity remain open. Reference ranges are configured by clinical staff; no ranges, abnormality interpretation or treatment recommendation are inferred by the software.

## Routes

| Route | Behavior |
|---|---|
| GET /v1/diagnostic-tests?search=&page=1 | Catalog with parameter definitions; admin, doctor, lab technician |
| POST /v1/diagnostic-tests | Admin/lab creates a typed definition and parameters |
| GET /v1/diagnostic-orders?encounterId=&page=1 | Scoped worklist; optional encounter filter |
| POST /v1/diagnostic-orders | Assigned doctor orders test for an active encounter; indication and Idempotency-Key required |
| POST /v1/diagnostic-orders/{id}/actions | Versioned collect/process/sign/release/reject/cancel |
| POST /v1/diagnostic-orders/{id}/results | Lab submits complete typed values and summary using current order version |
| GET /v1/diagnostic-orders/{id}/results?page=1 | Scoped result revisions with parameter names, units and reference text |

Lists are bounded to 25 rows (page 1..1000); a test/result has 1..50 parameters. Existing session, Origin, 32 KiB JSON and unknown-field checks apply. These private Go routes are not yet on the frontend proxy. Catalog/ordering JSON names and exact bounds are defined in `internal/domain/diagnostics.go`.

## Lifecycle and privacy

Ordered -> collected (unique sample/image reference) -> processing -> review -> signed -> released. Only lab technicians collect/process/submit; only the assigned doctor signs/releases/rejects/cancels. Admin can inspect but cannot substitute for clinical signing. Rejection returns an unapproved result to processing while retaining it. A replacement or post-release amendment requires a reason and creates a new revision.

Patients see only their own orders with a released report and only released result revisions. While an amendment awaits review, the old released result remains visible; internal pending status/version is hidden. Doctors are scoped by encounter assignment. Reception, nursing, pharmacy and finance roles receive no diagnostic results through these routes. Lab access is limited to this workflow, not unrestricted patient profile/clinical-note access.

Concurrent transitions and submissions lock the order and compare versions; stale retries return 409 rather than duplicate results. Order creation supports same-key replay and rejects changed payloads. Result values are validated against the definition, including finite numeric values. Content is sealed at commit; updates, deletions and late value insertion are rejected. Review/release and lifecycle events are append-only and audited.

## Evidence and remaining work

Go/database/HTTP tests cover wrong-doctor ordering, order replay, skipped collection, duplicate/invalid parameters and nonfinite results, simultaneous submissions, signed-content tampering, unauthorized signing, unreleased report privacy, amendment reasons and retained release history, result field labels, HTTP Origin and strict JSON. All tests and vet passed with the disposable-schema database suite enabled.

Still required: complete category/unit/catalog CRUD and source fields, sample rejection/recollection, reviewer specialty/team assignment policy, attachments and authorized report PDFs, diagnosis templates, billing source linkage, vaccines and birth/death/operation/investigation workflows. These are not marked complete by this increment.
