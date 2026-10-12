# Bed Management final correction review

Reviewed commit: `8681265`; prior review: `5eb8257`; main baseline: `adcecd2`.

Decision: **Changes required.** Current-operation INSERT ... RETURNING replaces the prior wall-clock replay pattern, but deployment coverage, provenance semantics and failing verification still prevent acceptance.

## H1 — High: already-applied 065 databases skip the fixes

The correction edits migrations 063, 064 and 065 in place and adds no later migration. A database that applied 065 from `fefe73e` has that filename in schema_migration, so the normal runner skips the new CTE and provenance-enrichment code. The added historical test covers pending 065 after older 064, not already-applied 065.

Independent isolated reproduction extended the historical-064 fixture by executing the actual 065 SQL from fefe73e and recording its filename before invoking the current runner. After correcting the fixture's discharge constraint violation (H2), the runner skipped 065 and the assertion failed: `expected original_notes enriched, got: ''`. Package duration 3.523 seconds; test duration 2.23 seconds.

Required: add a new corrective migration (next available number), preserving historical audit assertions and applying safe enrichment/reconciliation to already-recorded 065 states. Test actual historical 063/064/065 variants with populated assignments/audit rows and filename-based skips, as well as fresh installs and pending-065 upgrades. Do not rely on changing a previously recorded migration to repair those databases.

## H2 — High: submitted PostgreSQL suite still fails before claimed acceptance checks

Independent database-enabled execution of `go test ./internal/adapters/postgres -run TestBed_ -count=1 -v` returned exit code 1: **14 passed, 2 failed**, package duration 38.319 seconds.

- `TestBed_ReviewR4_AlreadyApplied064Runner` fails updating encounter 1 to discharged (line 1654). It omits discharge_summary and sets discharged_at three hours before the fixture's current-time admitted_at. Migration 003's encounter_check1 requires a nonempty discharge summary and discharged_at >= admitted_at. The test never reaches the new 065 checks.
- `TestBed_F4_MultipleEncountersPerPatient` fails inserting the historical discharged encounter (line 1759) because discharge_summary is absent. This also prevents the claimed history/active-selector verification.

The historical 063/064 constants in the first test are reduced SQL excerpts rather than exact complete historical migrations; for example the 063 duplicate reconciliation and later 064 paths are omitted. They are useful focused fixtures but should not be reported as coverage of every actual historical migration path.

Required: construct valid clinical fixtures without weakening production constraints, check all writes, run complete historical SQL where claiming upgrade lifecycle coverage, and rerun the database-enabled suite to a clean exit. Update verification claims to the results actually produced.

## H3 — Medium: provenance enrichment still substitutes or mislabels historical facts

Migration 065 explicitly sets original_created_at = original_assigned_from when original_created_at is unknown. Assignment start time and database record creation time are different facts; the schema permits them to differ. A missing source row does not justify declaring the start time to be its original creation time.

The enrichment also writes today's surviving notes/JSON into fields labeled original_notes/original_record on an earlier audit event, without a separate observation timestamp/source marker. That row may have changed after the earlier audit, as the submitted replay fixture itself demonstrates for assigned_to. Preserve that evidence, but identify it as a later surviving-state snapshot rather than implying it was captured by the earlier event.

Required: retain unknown creation timestamps as unknown, preserve original audit assertions, and store recovered surviving-state facts with explicit source and observation time (separate snapshot fields or an appended enrichment event). Only populate original facts where they are actually known. Test differing assigned_from/created_at, a deleted source row, and a source row changed after the original audit.

## Evidence boundaries

- Independently ran the full submitted 16-test Bed PostgreSQL suite; results above.
- Independent already-applied-065 regression reproduced skipped enrichment. Initial reviewer fixture hit the same discharge check failure; only the temporary fixture was corrected before reaching the target upgrade scenario. Temporary test removed afterward.
- Source reviewed for migration changes. Browser QA, frontend typecheck/build and contrast were not independently rerun in this pass. No acceptance claim is made for those gates.
- Documentation-only review; application code and main unchanged. No live clinical-data mutation.

Resolve H1–H3 and provide a clean database-enabled run plus populated historical-upgrade evidence before acceptance.
