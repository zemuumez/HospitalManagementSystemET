# Independent doctor backend review

Reviewed `ead3b4805794ccfff206874913f1f718fa77d330` and audit commit `50e1c03`. **Verdict: changes requested; do not merge yet.** This is a faithful-port review, not a request for new product features.

## Verification actually performed

- Full Go suite `go test ./... -count=1 -v` with loopback HMS_TEST_DATABASE_URL enabled: passed, including the isolated PostgreSQL transaction suite and doctor tests.
- Typecheck, eight frontend tests and formatting: passed.
- Additional isolated PostgreSQL application/repository probe reproduced the privacy and default/validation discrepancies below; see [observations](doctor-review-probes.json).
- Direct read of original ZIP DoctorController.php, DoctorRepository.php, Doctor.php and schedule helper to resolve original behavior. No Laravel execution or shared development record changes.
- No browser doctor-workflow acceptance claimed: this submission is backend-only and its frontend coordination is not implemented.

## D1 — high: doctor list bypasses private-profile access restrictions

Locations: `services/api/internal/application/scheduling.go` ListDoctors/Doctor; `services/api/internal/adapters/postgres/scheduling.go` doctor/ListDoctors; `services/api/internal/domain/scheduling.go` Doctor response.

Detail denies patients and denies doctors opening another doctor's record. ListDoctors instead permits appointments.read (including patient) and returns the same expanded Doctor objects for every active doctor. Newly added DOB, blood group, home address and other profile fields are populated and serialized. A patient probe received DOB `1990-01-01`, blood group `AB+` and synthetic private address while the same doctor's detail operation returned forbidden. A doctor can likewise bypass the own-record restriction through the list.

Fix: retain an appointment-picker/directory response containing only intended public/booking fields; keep private profile fields behind scoped administrative/detail authorization. Do not simply break patient appointment selection by forbidding all doctor lookup. Test list AND detail serialized output for all relevant roles, using two doctors with populated private fields. Also inspect callers of Scheduling.Doctors, since they use the expanded helper too.

## D2 — medium: default scheduling contradicts original behavior

Locations: `services/api/internal/domain/scheduling.go` CreateDoctorInput.Validate; `services/api/internal/adapters/postgres/scheduling.go` CreateDoctorProfile default hours.

Original source facts:

- DoctorRepository::store creates a Schedule with per_patient_time `01:00:00` (60 minutes).
- DoctorController::store loops Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday and creates each ScheduleDay with available_from `10:00:00`, available_to `19:30:00`.
- getDoctorSchedule helper uses the same original defaults.

Current defaults are 30 minutes and Monday–Friday 09:00–17:00. These are neither a stack requirement nor an accepted intentional deviation. The submitted test proves the new defaults rather than parity.

Fix: port the original default schedule and slot duration, preserving explicitly supplied valid schedules. Add a regression that asserts all seven days and exact original times/duration. Correct the audit evidence and do not automatically overwrite existing explicitly configured doctor schedules.

## D3 — medium: required doctor fields accepted empty

Location: `services/api/internal/domain/scheduling.go` CreateDoctorInput/UpdateDoctorInput validation.

Original Doctor::$rules requires designation, qualification and gender, in addition to specialist and identity fields. The new validator only bounds designation/qualification lengths and checks gender if nonempty. The independent probe successfully created a profile with all three omitted. Update also allows blank designation/qualification. This changes the original form contract without a documented necessity.

Fix: enforce the original required fields on the complete create flow; on partial updates validate the merged effective record or reject explicit blank required values as appropriate. Preserve omitted fields. Add missing/blank/whitespace tests, phone/address/blood-group validation according to evidence, and compare every source-required field in the mapping. Do not tighten unrelated fields based on guesses.

## D4 — high for step completion: account/profile creation remains disconnected

The assigned work and the new plan's Step 1 explicitly include coordinated account and doctor-profile creation with rollback/compensation. No Next staff route or provisioning flow changed. POST /v1/doctors requires a pre-existing user already assigned role doctor, while /api/staff still creates only identity/account/staff_access and commits separately. The Go tests insert identities directly; they do not test the original “New Doctor” account-plus-profile operation or failure cleanup.

Fix: finish the planned orchestration using the existing Better Auth-compatible mechanism. A successful create must deliver the linked identity, role, complete doctor/staff profile and original schedule; a failure must not leave an unexplained active orphan login. Test duplicate email, invalid department, failure after identity creation, retry and correct linkage through the real coordinating endpoint. Include original name/email editing or explicitly identify the remaining part instead of calling account/profile parity complete. Keep credential implementation and security policies already established; do not copy Laravel password hashes.

## D5 — medium: invalid status numbers deactivate doctors

Location: `services/api/internal/adapters/httpapi/doctors.go` doctorStatusRequest branch.

The handler translates status using `active = (*req.Status == 1)` without validating the numeric enum. Thus status=2, -1 or 99 becomes false and can deactivate the doctor and revoke sessions instead of returning validation error. This is source-inspection evidence; the submitted tests cover only valid status changes.

Fix: accept only explicit supported status values (0/1), reject missing/invalid/conflicting active/status fields with no state/audit/session mutation, and test those cases over HTTP. Keep valid desired-status requests idempotent.

## Correction order and scope

1. Fix D1 and D5 with HTTP/serialization and no-mutation regressions; commit.
2. Fix D2 and D3 to match the original source, and update the audit's source-to-target table; commit.
3. Complete D4's existing assigned account/profile creation slice and failure tests; commit.
4. Run Go tests with PostgreSQL enabled, typecheck, frontend tests and formatting; run focused real-auth integration for the coordinating create endpoint. Return commits and stop for review.

Do not expand into doctor deletion, new scheduling features, full frontend modernization or unrelated modules. The optimistic version check, department validation, session revocation and history retention already implemented should be preserved. Record remaining work honestly; passing implementation tests alone does not establish source parity.
