# Odontogram upgrade review — 2026-10-12

Reviewed local commit: `bfdc268`. Decision: changes requested.

Shared sorted legend locks and exact condition-value reference checks are now present. Snapshotting conflicting tooth observations before deduplication improves G1. Three upgrade/history issues remain.

## H1 — High: already-applied 061 databases skip the corrections

The only migration changed by this commit is 061. The production runner skips filenames recorded in `schema_migration`, so installations that applied 061 from `796bfc4` will not execute the new provenance capture, column widening or seed reconciliation. In particular, intermediate-060 databases retain narrow legend columns after the earlier 061.

Add a new corrective migration for already-applied 061 and test the actual prior migration contents with their recorded markers using the normal runner. Cover original 060, intermediate 060, already-applied 061 and fresh installation. Preserve all surviving data. Explicitly document that observations already deleted by old 061 cannot be recovered from the remaining tables; recovery requires a pre-upgrade backup or another retained source. Do not represent a later snapshot as recovered lost history.

## H2 — High: complete chart headers are still not archived

The provenance CTE unions existing provenance, child entries and JSON conditions. It never independently serializes each original chart header. An empty legacy chart with a distinct description but no child rows, JSON conditions or prior snapshot contributes zero elements and is then deleted if it is not the keeper. Even nonempty charts omit original header `created_by`, `updated_by`, `version`, chart timestamps and some encounter information when provenance is derived from child rows. Child timestamps are not substitutes for header timestamps.

Archive every source header as its own typed record before mutation, including its identity, full JSON, notes, attribution, encounter, timestamps, version and pre-existing provenance. Include child row identities as well as contents. Test an empty duplicate with unique notes, differing header versus child timestamps/authors, and existing snapshots; assert all header and tooth records remain recoverable after consolidation. Avoid treating this as a reason to rerun the old destructive migration on live data.

## H3 — Medium: seed reconciliation still overwrites customized fields

For KR/Ce/D/PS, the UPDATE predicates match an old name alone, or `is_default` plus one color; each UPDATE replaces name, description and color together. A user who retained the old name but customized the description/color still matches and loses those edits. Some alternate color matches can also overwrite a customized name. Testing only a customized B legend does not establish protection for the four reconciled codes.

Identify untouched seeds using the complete known old fingerprint, or perform field-aware reconciliation that preserves customized fields and records ambiguous cases for review. Add cases with customized name, color and description for each affected code. Preserve custom rows even when they resemble an old default.

## Evidence boundaries

Independently executed `go test ./internal/adapters/postgres -run TestOdontogram_ -count=1` with PostgreSQL enabled: pass, exit 0, 46.746s. The findings above are source-inspection cases not covered by that result, based on migration 061 and the established filename-based migration runner. Browser/build and rendered PDF checks were not independently rerun. Main was not modified. Resolve H1–H3 together before acceptance.
