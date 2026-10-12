# Bed Management independent review

Reviewed implementation: `b5ebedb37e2e99df391ca77dd01da41e57f4f391`.
Accepted baseline: `adcecd2034b6c9a644d4f9bc48cc7f01f426ad15`.

Decision: **Changes required; not accepted for integration.** Removing fixture arrays is useful, but authorization, lifecycle consistency, concurrency, migration history, retrieval, and verification gaps remain.

## R1 — High: new reads and discharge bypass clinical scope

`services/api/internal/application/clinical_care.go:105` permits any doctor or nurse to list all assignments. The repository takes no actor and returns patient identity/contact information without encounter scoping. `DischargeBedAssignment` at line 115 checks role and UUID only; its repository implementation also has no patient-access check. This differs from `AssignBed` and encounter-specific `ListBedAssignments`, which call `AuthorizeClinicalRecord`.

`Clinical.SearchBeds` also forwards no actor to the new search implementation, which now exposes patient names, phone numbers, gender, admission dates and encounter IDs under the broad beds.read permission. Capacity access must not implicitly grant patient-detail access.

Independent isolated regression: seeded nurse has no care-team/encounter assignment and `AuthorizeClinicalRecord` returns ErrForbidden. Nevertheless, the application service returns the patient's assignment and successfully discharges it (nil error).

Required: scope assignment reads/totals and discharge to the canonical encounter-access policy; authorize patient-bearing bed details or redact them for capacity-only viewers. Add service/HTTP tests for unrelated and assigned staff, receptionist policy, patient users, and admins, including unchanged records on denied writes.

## R2 — High: existing transfer/discharge paths diverge from assignment state

`services/api/internal/adapters/postgres/bed_management.go` `transferBedTx` changes encounter.bed_id and writes bed_event but never closes/creates bed_assignment records. `services/api/internal/adapters/postgres/clinical.go` encounter discharge similarly does not close open assignments. New occupancy queries consider both encounters and assignments, so the old bed can remain unavailable after transfer or clinical discharge.

Independent isolated regression: assign bed A using the new service, then transfer to B using Store.TransferBed with the correct encounter version. Transfer succeeds, but A still has one open assignment and B has zero. The submitted transfer test calls AssignBed twice and therefore misses the existing transfer API.

The new discharge routine locks assignment then encounter; AssignBed locks encounter then updates the old assignment. Concurrent transfer/discharge can deadlock through this reversed order.

Required: share one transactional occupancy lifecycle across admission, both transfer interfaces, bed release, and clinical discharge. Use a consistent lock order, preserve events/history, distinguish bed release from clinical encounter discharge, and verify both interfaces in cross-path and concurrent regressions.

## R3 — High: advertised optimistic locking is optional and omitted by UI

`AssignBedInput.Validate` does not require a positive version. Store.AssignBed checks versions only when input.Version > 0. `apps/web/src/components/bed-management-workspace.tsx:627` sends no version at all. Thus the normal UI bypasses the stale-encounter safeguard; two outdated users can sequentially move the same encounter to different beds without an explicit stale rejection.

Independent isolated application-service regression: an assignment request with omitted version succeeds. The submitted assignment tests also omit versions. Bed discharge has no expected-version contract, and bed-type deletion likewise makes version checking conditional.

Required: define and enforce expected-version contracts for affected mutations, send actual versions from the UI, and test omitted/zero/negative and stale versions. Preserve form data and show conflicts inside the active dialog.

## R4 — High: migration fabricates historical assignment end dates

`db/migrations/063_bed_management_parity.sql:14` ranks duplicate open assignments by created_at and sets assigned_to = created_at on all but the selected row, first by bed and then by encounter. This invents clinical end times and chooses an authoritative assignment without checking encounter.bed_id or archiving the original open-state evidence. It can leave encounter occupancy and the remaining assignment inconsistent. Retaining bed_event alone does not preserve the modified assignment facts.

The upgrade test seeds only one non-conflicting assignment; it does not exercise either reconciliation UPDATE.

Required: use a documented conservative reconciliation policy. Preserve full original records/provenance, do not infer clinical discharge timestamps from insertion order, and reconcile or explicitly quarantine ambiguous occupancy. Cover conflicting bed/encounter mappings and differing timestamps with the normal migration runner. Since 063 is reported already applied, supply any required new corrective migration and document which overwritten facts require backup/audit recovery; editing 063 alone cannot repair deployed databases.

## R5 — Medium: first-page limits still hide selectable patients and capacity

`bed-management-workspace.tsx:230,350,420,426` fetches only page 1 of types, beds, patients and encounters. The board and bed/type selectors stop at 100; the encounter endpoint still returns a fixed 25 per page regardless of the requested pageSize=100 (`postgres/clinical.go:250`, `httpapi/clinical.go:170`). Older active encounters therefore cannot be selected, while global occupancy totals may describe beds absent from the board. These auxiliary fetch errors are swallowed before reporting connected status.

Required: use server-backed searchable selectors and complete/paginated board retrieval, with independent totals and explicit load failures. Test more than 100 beds/types/patients and more than 25 encounters, selecting an older active encounter and a bed beyond page one.

## R6 — High: browser QA mutates shared clinical data and embeds a credential

`scripts/verify-bed-management-workspace.mjs` defaults to the shared local HMS database, includes a literal database password in its fallback URL, selects an arbitrary existing IPD encounter/patient, updates patient_profile.gender, then uses that encounter for assignment/discharge tests. Its finally block closes connections/browser but does not remove fixtures or restore clinical changes. This is not an isolated browser suite. Screenshots containing multiple previous QA types are consistent with retained test data.

Required: remove the credential from tracked source and rotate it if still valid. Use the project's isolated browser runner with a unique schema, newly built isolated API, dedicated synthetic patients/encounters/accounts, and guaranteed cleanup. Fail closed without explicit test configuration. Do not automatically delete existing QA records that may reference real clinical records; first inventory and review affected data. Add lifecycle, role-boundary, pagination, stale-version, empty-state and failed-save assertions rather than relying only on screenshots.

## R7 — Medium: localization, currency and truthful display remain incomplete

The supplied Amharic screenshot still shows the English technical service banner and Sync Backend. The source emits dollar signs in bed charges/CSV despite the requested ETB minor-unit contract. Assignment rows still manufacture an @example.com email when absent (`bed-management-workspace.tsx:330`), and the board treats any gender other than female as Male. The board also maps every unavailable bed to occupied, losing the distinction between maintenance and actual occupancy.

The browser language/theme journey captures screenshots, force-adds a dark class and reports PASS without asserting translations or contrast. The supplied dark screenshot has visibly dim description text; no measured contrast result was independently obtained.

Required: use translated user-facing status text, the canonical theme toggle, ETB formatting, honest missing-value placeholders, and separate ready/occupied/maintenance states. Verify all tabs/dialogs, errors, status labels and exports; add actual localization and contrast assertions. User-entered names/descriptions need not be translated.

## Verification and boundaries

- Independently ran `go test ./internal/adapters/postgres -run TestBed_ -count=1 -v` with PostgreSQL enabled: all seven submitted suites passed, 18.783 seconds.
- Added temporary isolated reviewer tests: unrelated-nurse listing/discharge and canonical-transfer assignment consistency failed (3.619 seconds); omitted-version validation failed separately (3.504 seconds). Temporary test files were removed after capturing results; no application implementation changes remain.
- Did not execute the browser script because source inspection showed shared-database mutation and no fixture cleanup. Browser, frontend unit and typecheck results in the implementation summary remain implementer-reported.
- Inspected supplied screenshots and source. No independent measured dark-mode contrast or full browser acceptance is claimed.
- This review changes documentation only. Main and accepted modules were not modified.

Acceptance requires R1–R7 corrections, dedicated regressions, safe isolated browser verification, and a fresh independent review of the resulting commit.
