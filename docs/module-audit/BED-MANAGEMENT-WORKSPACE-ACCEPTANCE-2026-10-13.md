# Bed Management workspace acceptance

Accepted implementation: `7221ba4` on `feat/bed-management-workspace-parity`.
Baseline main: `adcecd2`.
Final prior review: `3debe15` (H1–H3).

Decision: **Accepted for integration within the reviewed Bed Management parity scope.** No outstanding blocker from the final H1–H3 correction batch was found in this pass. This acceptance does not claim independent execution of every browser or deployment gate.

## Final findings resolved

- **H1:** Dedicated migration `066_bed_assignment_provenance_and_audit_repair.sql` reaches databases that already recorded 065. It adds surviving-state snapshot columns and repairs the earlier original_created_at constraints/default. Audit mutations use current-operation INSERT ... RETURNING IDs rather than historical audit entries selected by wall-clock recency.
- **H2:** The submitted database-enabled suite now completes successfully. Historical discharged fixtures satisfy discharge summary/timestamp constraints, and patient fixtures respect the existing one-active-IPD-stay invariant.
- **H3:** Current pending 065 no longer substitutes assignment start time for unknown creation time or copies later observations into original fields. Migration 066 stores surviving facts in separate surviving_* columns with source and observation time; missing-source fixtures retain unknown values. Existing historical end-time assertions remain distinct from later observations.

## Independently executed verification

1. `go test ./internal/adapters/postgres -run TestBed_ -count=1 -v`, with PostgreSQL enabled and isolated test schemas: **17/17 suites passed**, exit 0, package duration **41.235 seconds**. Includes role scoping/redaction, occupancy concurrency, transfer/discharge consistency, positive versions, populated upgrades, duplicate handling, already-applied migrations and historical/current encounter selection.
2. `node --test apps/web/src/lib/bed_management.test.ts`: **9/9 passed**, exit 0.
3. Additional reviewer upgrade regression replaced the submitted historical SQL excerpts with the complete SQL retrieved from `392edb6` (063 and 064) and `fefe73e` (065). With populated assignment/audit fixtures and recorded filenames, the pending 066 upgrade **passed**, exit 0, package duration **3.775 seconds**. It verified surviving-state snapshots, differing assignment/creation times, deleted source rows, preserved occupancy and skipped prior migration filenames. Temporary reviewer test removed afterward.

## Scope and evidence limits

Acceptance covers the submitted Bed Status, Assignments, Beds and Bed Types implementation and the successive reviewed corrections. Prior source review established the revised encounter access predicates, selector traversal/search, version contracts and isolated browser runner integration. This final pass concentrated on H1–H3 and the regression suite.

Frontend typecheck, Go build, browser QA and measured contrast results supplied by the implementer were not independently rerun in this final pass. The focused backend suite is not a claim that every package's tests were rerun. Test skipping without database configuration is not database verification; the independent run above was database-enabled.

Historical facts already overwritten by earlier deployed migrations cannot be recovered merely by applying the new migration. Recovery may require retained backups or external audit records. Surviving-state snapshots must not be interpreted as proof of unavailable original clinical facts.

## Integration handoff

Integrate the accepted implementation and this acceptance record through the established review workflow. Apply pending migrations through 066 using the normal filename-tracking runner, rebuild/restart the API, and smoke-check assignment, transfer, release, capacity updates and authorized visibility against the deployed build. Preserve backups before a populated database cutover.

This review made documentation-only changes, left main unchanged, and did not alter live clinical records.
