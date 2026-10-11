# Odontogram workspace review — 2026-10-11

Reviewed local commit: `9eaa114`. Decision: changes requested. The feature branch was not available from origin during review; the exact submitted commit was imported from the local HospitalManagementSystemET repository.

## R1 — Critical: shared clinical authorization now bypasses doctor assignment

In `postgres/clinical_record_scope.go`, the doctor branch adds `purpose IN ('clinical','odontogram_write','odontogram_read')` as an OR alternative to patient/encounter assignment. Any doctor consequently passes for any existing patient for those purposes. Known chart IDs permit unrelated doctors to read, update or delete charts, and the shared helper also weakens other clinical endpoints. Restore assignment/care-team checks; a purpose identifies an operation, not a grant of patient access. Align list and detail policies. Test two unrelated doctors against list, detail, PDF, create, update, delete and existing non-dental clinical endpoints.

## R2 — High: omission of version bypasses optimistic concurrency

`SaveOdontogramChartInput.Validate` allows zero, and `UpdateOdontogramChart` checks equality only when `input.Version > 0`. An update omitting version overwrites the current chart regardless of its intervening edits. Require a positive expected version for updates and reject missing/zero/stale values. Creation may retain a separate version contract.

## R3 — High: compatibility writes target ordinary examinations and invert lock order

`SetToothCondition` without chartId selects `ORDER BY is_legacy DESC, created_at DESC LIMIT 1` without requiring `is_legacy=true`. If only modern charts exist, a legacy caller silently edits the latest examination. Target only an explicit chart or a designated legacy chart, creating the latter safely if absent. Concurrent first calls also need serialization/uniqueness to prevent duplicate legacy charts.

The compatibility path writes the child row before updating the parent, while aggregate updates lock the parent before replacing children. Use a consistent parent-first lock protocol to avoid deadlocks between these paths. Verify simultaneous compatibility and aggregate writes, including JSON/child agreement and version behavior.

## R4 — High: editing migrated charts destroys preserved tooth provenance

Migration 060 initially retains the old tooth rows, but its parent description selects only one tooth's notes and its doctor selects one diagnosis author. `GetOdontogramChart` exposes only the condition map and parent description. Saving that chart deletes every original child row and recreates them with the same parent description, current actor and current timestamps. Distinct original tooth notes, authors, encounters and dates are therefore lost on the first ordinary edit.

Preserve legacy provenance in an immutable snapshot or retain/edit the complete per-tooth data deliberately. Clearly distinguish aggregate legacy time bounds from a real examination event. Add a migration-and-edit test with multiple original notes, authors, encounters and timestamps; verifying only the immediate migration is insufficient.

## R5 — High: custom legends are not durable clinical definitions

`odontogram-register.tsx` continues loading/saving custom legend names and colors solely in sessionStorage. The database preserves `IMP-Z` but a fresh session loses its meaning and color; the submitted fresh-session screenshot itself no longer includes that legend in the palette. Persist the definitions or chart-specific legend snapshots and reload them across sessions. Verify code, meaning and color in editor, view, print and PDF, including after legend edits/deletion.

## R6 — Medium: PDF validity check does not establish usable output

`generateOdontogramPDF` emits one page with unwrapped notes and 16-point downward advances per tooth. A valid 52-tooth chart runs below the page boundary; long clinical notes run off the right edge. Raw UTF-8 strings with the built-in Helvetica font do not provide Amharic text support. Checking `%PDF-1.4` and byte count misses all these issues.

Use wrapping, pagination and a Unicode-capable embedded font; render and inspect exported reports with long notes, Amharic patient names/custom codes and all 52 supported teeth. Assert that every finding remains visible. Include the chart/legend visualization if required by the legacy print contract.

## R7 — Medium: patient selection silently stops at the first page

`loadClinicalMetadata` fetches only `patients?page=1&size=100`. The selector cannot choose patients outside that result page, even though the chart register is paginated. Implement scoped server-backed patient search/pagination and load the selected patient independently when editing. Test selection and editing beyond the first 100 patients.

## Evidence and boundaries

Independently executed `go test ./internal/adapters/postgres -run TestOdontogram_ -count=1` with PostgreSQL enabled: pass, exit 0, 21.857s. The findings above are source-review findings for cases beyond that suite. Browser/frontend gates were not independently rerun. Screenshots alone do not demonstrate dark-mode, Amharic, PDF rendering or durable legend parity. Main was not modified. Resolve the findings as one batch with focused regression cases and preserve already-applied migration compatibility where corrective schema changes are needed.
