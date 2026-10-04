# Diagnostic catalog revisions

Migration 024 preserves the immutable test definitions introduced in migration 009 while allowing staff to retire and replace them. Existing orders keep their exact test ID, parameter names/units/reference ranges and tariff. Changing a catalog entry never changes an already ordered test or its source invoice.

- POST `/v1/diagnostic-tests/{id}/revisions`: administrator or lab technician sends the complete existing test-input fields plus current `version` and a required `reason` (up to 1,000 characters). The transaction locks the active definition, retires it and creates a new active definition with new ID and incremented revision. The kind must remain pathology or radiology as originally defined. A conflicting name rolls back the whole revision, including retirement.
- POST `/v1/diagnostic-tests/{id}/archive`: `{version,reason}` retires an active definition. Archived definitions cannot be reactivated or used for new orders.
- GET `/v1/diagnostic-tests/{id}/revisions?page=1`: roles with catalog access can retrieve 25 versions at a time, newest first. Any member ID identifies the series. Fields include active, version, revision, rootId and supersedes, plus the retained definition and parameters.
- GET `/v1/diagnostic-tests` returns active definitions for selection. Existing order/result reads still resolve archived definitions. A retry of an existing order remains valid after retirement.

Definition updates/deletion remain blocked at the database; the only allowed update is a one-way active-to-retired transition with version increment. Revision/archive reasons and actors are retained separately and operations are audited. No historical reference range or price is overwritten. Category/unit master tables and remaining legacy catalog fields remain open.

Isolated PostgreSQL tests verify concurrent revisions, old parameter/tariff preservation, role denial, active-only lists, archived-order rejection, original order retry, archive version conflicts, immutable definitions/events and the history HTTP endpoint. Full Go tests and vet pass.
