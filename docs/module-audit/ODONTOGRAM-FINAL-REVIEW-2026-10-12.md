# Odontogram final correction review — 2026-10-12

Reviewed local commit: `796bfc4`. Decision: changes requested.

The positive-version and legacy-only compatibility fixes remain present. The new migration restores established legend definitions, and PDF generation now supplies font widths and measured wrapping. Three issues remain.

## G1 — High: duplicate-chart reconciliation deletes clinical history before snapshotting it

Migration 061 moves non-conflicting teeth, deletes remaining conflicting teeth, deletes duplicate chart headers, and only afterward constructs `legacy_provenance`. If two legacy charts contain tooth 1 with different conditions or notes, one record is irretrievably lost. Duplicate header descriptions, authors and dates are also discarded. Existing non-NULL provenance on a keeper is not merged with duplicate provenance.

Independent PostgreSQL reproduction: adapted the supplied F1 fixture so both legacy charts contain tooth 1, retaining their distinct notes. Migration succeeds but leaves one child. A second assertion allowing one current child still fails because provenance contains only the surviving note. Focused regression exit 1, 2.60s test duration. This confirms loss rather than merely a change in current-row count.

Capture complete original chart headers, tooth identities/content and existing provenance before any merge/delete. Keep all conflicting observations in a durable archive/snapshot even if one is selected as current. Do not guess clinical precedence from chart timestamps. Test same-tooth conflicts, different authors/encounters, existing snapshots and deterministic reconciliation. Account explicitly for already-applied 061; deleted history cannot be reconstructed without a prior data source.

## G2 — High: reference checks do not coordinate with chart writers

`UpdateOdontogramLegend` and `DeleteOdontogramLegend` lock the legend row, but `CreateOdontogramChart`, `UpdateOdontogramChart` and `SetToothCondition` never acquire that lock or otherwise enforce a reference. A chart transaction can insert an uncommitted reference; the legend mutation sees no committed reference and renames/deletes the definition. Both commit, leaving an unmatched stored code.

Implement a shared binding/mutation protocol or immutable per-chart legend snapshots. Test both commit orders using deterministic barriers, not just sequential in-use checks. Also fix the aggregate reference expression: `odontogram ? $1` checks tooth-number keys, whereas codes are JSON values. Replace provenance substring LIKE matching with exact condition-value matching to avoid accidental matches and wildcard behavior.

## G3 — Medium: intermediate migration 060 schema is not reconciled

The `d89af48` version of 060 already created `odontogram_legend` with `code varchar(16)` and `name varchar(100)`. Migration 061's `CREATE TABLE IF NOT EXISTS` leaves those types unchanged, while its fresh-install schema permits 64-character codes and 128-character names. Upgraded and fresh databases therefore retain different contracts. A longer code accepted by current validation can fail in PostgreSQL only on the upgraded database.

Explicitly alter existing column types/constraints in a corrective migration. Test the actual intermediate-060 schema as well as original-060 and fresh installs, including boundary-length custom codes/names. Preserve existing user edits: `is_default=true` alone does not distinguish untouched wrong seeds from default rows customized through the prior editor, so do not blanket-overwrite those without a deliberate reconciliation policy.

## Verification and boundaries

Independently ran the submitted PostgreSQL suites with `go test ./internal/adapters/postgres -run TestOdontogram_ -count=1`: pass, exit 0, 42.891s. The additional overlapping-tooth regression failed as described; its temporary source was removed after recording results. G2 and G3 are source-review findings. Browser/build and rendered-PDF checks were not independently rerun. Main remains unchanged. Deliver one correction batch covering G1–G3 with fresh and upgrade regressions.
