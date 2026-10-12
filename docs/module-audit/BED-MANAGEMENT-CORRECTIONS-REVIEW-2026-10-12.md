# Bed Management corrections review

Reviewed commit: `392edb6` on `feat/bed-management-workspace-parity`.
Prior review: `09ef974`; accepted main baseline: `adcecd2`.

Decision: **Changes required.** The new positive-version contracts, assignment synchronization, unrelated-nurse restrictions, display improvements and isolated-runner integration improve the implementation. Four remaining findings prevent acceptance.

## F1 — High: corrective migration 064 fails on a surviving open duplicate

In `db/migrations/064_bed_assignment_reconciliation_audit.sql`, step 1 restores every closed row matching an active encounter and bed by setting assigned_to to NULL. It does this before resolving another open assignment and while migration 063's unique partial indexes remain active.

A normal original-063 outcome is two assignments for the same encounter/bed: the older row has assigned_to = created_at, and the newer row remains open. Migration 064 attempts to reopen the older row and violates one_active_bed_assignment (or the encounter index). The transaction fails rather than repairing the database.

Independent reproduction: extended the submitted R4 fixture with a surviving open assignment for the same encounter/bed. Applying 063 then 064 failed with PostgreSQL SQLSTATE 23505, constraint one_active_bed_assignment (test duration 2.28 seconds, package duration 4.002 seconds). The submitted fixture has only one artificially closed row and cannot exercise this collision.

Required: reconcile the entire conflicting group under a documented conservative policy before enforcing one open assignment. Cover original already-applied 063, current 063, fresh installation and duplicate encounter/bed combinations using filename-based migration tracking. Preserve all original records/provenance. Since 064 is now reported deployed, a repair must account for already-recorded migrations, not only edits to old files.

## F2 — High: migrations still invent clinical timestamps and lose original open-state provenance

Both current 063 and 064 use `COALESCE(e.discharged_at, ba.assigned_from + interval '1 hour', ba.created_at)`. The one-hour fallback is fabricated, not an authoritative clinical end time. Moreover, encounter discharge is not generally evidence of when an earlier bed assignment ended during a transfer. Step 3 of 064 writes clock_timestamp() into assigned_to for some conflicts; labeling this quarantine does not separate administrative quarantine time from a clinical end time.

Current 063 modifies duplicate rows before any audit snapshot exists. Migration 064's timestamp-repair scan only selects assigned_to = created_at, so it misses the new one-hour values written by 063. An isolated sequential test with two initially open assignments confirmed that the modified older assignment receives zero migration audit records (test duration 2.26 seconds, package duration 3.570 seconds).

The audit table also does not capture complete original rows (for example notes, assigned_by and created_at). A timestamp-equality heuristic cannot prove that a row was modified by original 063 rather than legitimately ended with equal timestamps.

Required: archive complete original assignment evidence before mutation, retain unknown clinical end times as unknown, and represent quarantine/reconciliation separately. Do not manufacture duration or conflate whole-encounter discharge with an earlier bed transfer. Document unrecoverable prior changes and backup/audit recovery requirements. Test original open-state preservation through the full 063-to-064 sequence, including already-applied versions, rather than only a fabricated singleton fixture.

## F3 — High: duplicated read-scope predicates remain broader than canonical encounter access

`ListAllBedAssignments` in `services/api/internal/adapters/postgres/clinical_care.go` and `SearchBeds` in `clinical.go` authorize a doctor whenever patient.clinician_user_id matches, even if the requested assignment belongs to another doctor's encounter. `AuthorizeClinicalRecord` permits that patient-level relationship only when no encounter ID is specified; for a specific encounter it requires attending/care-team access.

Independent isolated reproduction: an active IPD encounter belongs to doctor A; patient.clinician_user_id is doctor B, who is not on that encounter's team. Canonical admission authorization denies B, but assignment listing returns the row and bed search returns its patient name. Both reviewer assertions failed (test duration 2.23 seconds, package duration 3.508 seconds). The new unrelated-nurse test does not cover this relationship.

Required: reuse or exactly match canonical encounter authorization for patient-bearing search predicates, fields, list rows and totals. Preserve capacity-only visibility through redaction. Add service/HTTP regressions for a patient-level clinician who lacks access to a different clinician's encounter, alongside assigned staff and unrelated staff.

## F4 — Medium: selectors are still bounded local filters, not server-backed search

`apps/web/src/components/bed-management-workspace.tsx` still fetches only `bed-types?page=1&pageSize=100` and page 1 of patients/encounters with pageSize=200. `patientSearch` and `encounterSearch` are used by local useMemo filters, not debounced API queries or subsequent-page retrieval. Increasing the encounter limit from 25 to 200 moves the cutoff; it does not remove it. An older active encounter beyond the fetched page remains unselectable, and bed types beyond the first 100 are unavailable in selectors.

The new test seeds 31 encounters and requests 50, so it proves a larger backend page works but not that the UI can find an older record beyond its first page. The all-bed board loop is an improvement and is distinct from the still-bounded patient/type/encounter selectors.

Required: implement actual server-backed scoped search/pagination with request cancellation or stale-response protection, or complete bounded-page traversal where appropriate. Verify selecting patients, types and active encounters beyond current cutoffs through browser interactions, including when one patient has multiple encounters.

## Verification and evidence boundaries

- Independently executed PostgreSQL-enabled `go test ./internal/adapters/postgres -run TestBed_ -count=1 -v`: all 12 submitted suites pass; package duration 31.436 seconds.
- Independently executed `node --test apps/web/src/lib/bed_management.test.ts`: all nine tests pass.
- Added temporary isolated reviewer regressions for F1, F2 and F3; all reproduced the described failures. Removed the temporary test file after recording evidence. An initial F3 fixture attempt violated the existing encounter/case FK; the corrected fixture preserved that relationship and reproduced the access discrepancy.
- Inspected isolated-browser-runner integration and source. Did not independently rerun browser QA, TypeScript, full Go suite, or measured contrast in this review; those results remain implementer-reported.
- No production database changes or application implementation edits were made by this review. Documentation-only branch; main untouched.

Resolve F1–F4 and supply focused regressions plus safe isolated browser verification before requesting acceptance again.
