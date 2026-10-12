# Odontogram migration sequence review — 2026-10-12

Reviewed local commit: `1648ae9`. Decision: changes requested for the remaining history-preservation blocker.

## J1 — High: 061 deletes the data that 062 is intended to preserve

This commit restores the destructive deduplication body in migration 061. On an original-060 database with duplicate legacy charts, the normal runner executes 061 before 062. Migration 061 deletes duplicate headers, their existing provenance and conflicting teeth before 062 can archive them. Adding a later archival migration does not protect this upgrade path.

The submitted H2 fixture applies 060, seeds the data, then directly executes 062. It bypasses 061 and therefore does not test the production sequence for that populated state.

Independent PostgreSQL reproduction: copied the H2 fixture into a temporary review test and inserted execution of current 061 immediately before 062. Result: exit 1, test duration 2.67s, package duration 4.626s. Assertions failed with:

- `H2 LOSS: Standalone empty duplicate chart header was omitted from provenance!`
- `H2 LOSS: Prior snapshot tooth 19 was omitted!`

The temporary test was removed after recording the result. This reproduces loss from the current submitted migrations, not merely unrecoverable loss on databases that previously ran an older release.

Required correction: ensure the earliest pending migration that merges/deletes records first archives the complete headers, children and existing snapshots. For databases still at 060, make pending 061 non-destructive or fully preserving before 062 runs; a later migration alone cannot recover data deleted moments earlier. Keep 062 as the corrective path for databases that already recorded 061, documenting the limits of historical recovery.

Also remove the blanket exclusion of prior `chart_header` and `tooth_entry` records from 062's `existing_prov` collection. Existing typed records may describe deleted source rows and cannot safely be regenerated from current tables. Preserve all existing snapshot records, deduplicating exact records if needed rather than filtering by type.

Acceptance regression: start with actual 060 and its recorded migration state; seed overlapping teeth, an empty duplicate with unique notes, and both typed and untyped prior provenance. Run every pending migration in filename order through the production runner. Verify that all original header/child IDs, observations and prior snapshots remain recoverable. Retain a separate already-applied-061 test. Do not skip 061 in the populated preservation scenario.

Main was not modified. This review independently executed the focused failing PostgreSQL regression; the full submitted suites and browser/build gates were not rerun on this commit.
