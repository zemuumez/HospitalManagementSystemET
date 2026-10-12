# Bed Management upgrade review

Reviewed implementation: `fefe73e`; prior review: `b577746`.
Accepted main baseline: `adcecd2`.

Decision: **Changes required; not accepted for integration.** Encounter read predicates now remove the patient-level clinician bypass, and bed-type/encounter traversal plus server-backed patient search address the prior first-page implementation gaps. The migration repair and claimed verification still have blockers.

## G1 — High: restoration UPDATE replays earlier audit actions without checking current occupancy

Migration 065 selects restoration candidates using vacancy checks, but its subsequent UPDATE is not restricted to the rows selected by that operation. It updates every audit entry with resolution_action = 'restored_active' and reconciled_at within the last minute. Migration 064 uses the same pattern.

An earlier 064 restoration can be followed by a legitimate release and new assignment before 065 runs. Although 065's candidate query correctly excludes the occupied slot, the UPDATE replays the old audit action and sets the earlier assignment back to open. This violates one_active_bed_assignment. Without a conflicting row, replay could silently undo the subsequent closure. Elapsed wall-clock time is not a reliable mutation identifier; a sufficiently slow migration can also exclude its own newly inserted audit entries.

Independent isolated reproduction used the actual 063/064 SQL from commit 392edb6, recorded those filenames, and seeded/restored a closed assignment. It then closed that assignment, inserted a new open assignment, and ran the pending repair with accurate earlier migration markers. Migration 065 failed with SQLSTATE 23505 / one_active_bed_assignment (2.37-second test).

Required: couple mutations to the exact current-operation IDs, for example INSERT ... RETURNING through a CTE or a transaction-local candidate set. Do not replay historical audit rows based on recency. Recheck/lock current occupancy as appropriate. Cover existing audit rows followed by later release/reassignment, old audit timestamps, and the actual historical migration versions with the production filename-skip behavior.

## G2 — High: actual original-064 audit rows are not upgraded to full surviving provenance

The original 064 table from 392edb6 did not contain original_record, original_notes, original_assigned_by, original_created_at, or administrative_quarantine_at. Migration 065 adds default values, but never fills full surviving assignment facts for those already-audited rows. Candidate queries explicitly exclude assignments already present in the audit table. Thus original_record remains {}, original_notes/original_assigned_by remain empty, and original_created_at is populated with migration time rather than the actual surviving creation time.

065 also omits 064's superseded-closed/historical reconciliation paths. Adding columns and handling open duplicates/vacant restorations does not establish the report's claim that all surviving affected history on already-applied 064 databases is losslessly captured.

Independent isolated reproduction applied actual 392edb6 migrations 063/064 to a populated fixture, recorded their filenames, then applied only 065. The source assignment with its notes and author remained available, but its audit entry still lacked full provenance (2.40-second test). This is recoverable surviving evidence, distinct from overwritten clinical end times that genuinely require a backup.

Required: preserve prior audit assertions unchanged and append a clearly identified surviving-state snapshot, or explicitly enrich separate fields with known surviving evidence and provenance. Do not label repair time as original creation time or overwrite historical original_assigned_to with a later observed value. Test populated historical schemas and audit actions, not just current DDL with old filenames marked applied. Document what is known, what is only a post-repair observation, and what cannot be reconstructed.

## G3 — High: submitted database suite fails and upgrade fixtures do not prove reported lifecycles

Independent PostgreSQL-enabled command:

`go test ./internal/adapters/postgres -run TestBed_ -count=1 -v`

Result: **11 passed, 5 failed**, package duration 40.184 seconds, exit code 1. Failures:

1. `TestBed_ReviewR1_RoleAccessScopeAndRedaction`: searches for Derartu, while the shared fixture seeds Abebe Bikila; expected authorized match is absent. This failure does not establish a new authorization implementation defect.
2. `TestBed_ReviewR4_FreshInstallAndProvenanceArchiving`: runner attempts to reapply 002_scheduling.sql and fails because doctor_profile already exists.
3. `TestBed_ReviewR4_VacantSlotRestoration`: same setup/runner failure.
4. `TestBed_ReviewR4_AlreadyApplied064Runner`: same setup/runner failure. It also reads the current revised 063/064 files instead of executing the historically deployed SQL and contains no assignment/audit rows to validate their repair.
5. `TestBed_F4_MultipleEncountersPerPatient`: the fixture ignores insertion errors while trying to create multiple active IPD stays for one patient, contrary to the established one_active_inpatient_stay index in migration 003. Both requested new IDs are absent.

Required: check every fixture write error; record all migrations actually installed by setup; use real historical DDL/content for historical lifecycle tests. Keep the established one-active-IPD-stay invariant unless a separately approved contract change requires otherwise. Test multiple historical encounters with the valid active encounter selectable instead of silently violating that invariant. Rerun the full database-enabled suite to a clean exit and update the report with actual results.

## Verification and review boundaries

- Independently ran the submitted 16-test PostgreSQL suite with the failures above.
- Independently ran `node --test apps/web/src/lib/bed_management.test.ts`: 9/9 passed.
- Temporary reviewer tests initially hit the same missing migration-marker setup issue. After recording migrations 001–062 accurately, they reached actual historical 063/064 plus current 065 and reproduced G1/G2; combined package duration 6.039 seconds. Temporary files were removed after recording results.
- Source inspection confirms the patient-level clinician predicate was removed from encounter-bearing reads and the revised selectors issue additional requests. This review does not claim independent full browser, contrast, typecheck, build or large-catalog UI verification.
- Documentation-only review; no application edits or live clinical-data mutations. Main unchanged.

Resolve G1–G3, verify populated historical upgrades and all database tests, then request independent acceptance again.
