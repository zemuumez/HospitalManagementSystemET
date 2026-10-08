# Doctor Deletion Protection & Reference Auditing (Step 2 Acceptance)

**Date**: 2026-10-08  
**Branch**: `feat/doctor-deletion-parity`  
**Reference**: [doctor-workspace-parity-plan.md](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/doctor-workspace-parity-plan.md) (Step 2)  
**Specification Sources**:
- Legacy Controller: `review/legacy/app/Http/Controllers/DoctorController.php` (`destroy` method)
- Legacy Dependencies: `review/legacy/app/Models/` (`PatientCase.php`, `PatientAdmission.php`, `Appointment.php`, `BirthReport.php`, `DeathReport.php`, `InvestigationReport.php`, `OperationReport.php`, `Prescription.php`, `IpdPatientDepartment.php`, `EmployeePayroll.php`)
- Target Endpoints:
  - Go Backend: `DELETE /v1/doctors/{id}` in [doctors.go](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/services/api/internal/adapters/httpapi/doctors.go)
  - Next.js Staff API: `DELETE /api/staff` in [route.ts](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/apps/web/src/app/api/staff/route.ts)
  - Next.js Proxy: `DELETE /api/hms/doctors/{id}` via [route.ts](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/apps/web/src/app/api/hms/%5B...path%5D/route.ts)

---

## 1. Source-to-Target Dependency Mapping

In the original Laravel HMS, `DoctorController@destroy` protects against deleting doctors referenced by nine clinical models plus employee payroll. In the target PostgreSQL database, these map directly to the active clinical schema:

| # | Original Laravel Model | Original Relation / Foreign Key | Target PostgreSQL Table & Column | Conflict Behavior |
|---|---|---|---|---|
| 1 | `PatientCase` | `patient_cases.doctor_id` | `patient_case.doctor_id` | HTTP 409 (`RECORD_IN_USE`) |
| 2 | `PatientAdmission` | `patient_admissions.doctor_id` | `encounter.doctor_id` (`kind = 'ipd'`) | HTTP 409 (`RECORD_IN_USE`) |
| 3 | `Appointment` | `appointments.doctor_id` | `appointment.doctor_id` | HTTP 409 (`RECORD_IN_USE`) |
| 4 | `BirthReport` | `birth_reports.doctor_id` | `birth_report.delivered_by` | HTTP 409 (`RECORD_IN_USE`) |
| 5 | `DeathReport` | `death_reports.doctor_id` | `death_report.certified_by` | HTTP 409 (`RECORD_IN_USE`) |
| 6 | `InvestigationReport` | `investigation_reports.doctor_id` | `investigation_report.investigated_by` | HTTP 409 (`RECORD_IN_USE`) |
| 7 | `OperationReport` | `operation_reports.doctor_id` | `operation_report.surgeon_id` | HTTP 409 (`RECORD_IN_USE`) |
| 8 | `Prescription` | `prescriptions.doctor_id` | `prescription.doctor_id` | HTTP 409 (`RECORD_IN_USE`) |
| 9 | `IpdPatientDepartment` | `ipd_patient_departments.doctor_id` | `ipd_admission_details` joined via `encounter.doctor_id` | HTTP 409 (`RECORD_IN_USE`) |
| 10 | `EmployeePayroll` | `employee_payrolls.owner_id` (polymorphic `Doctor`) | `employee_payroll.user_id` | HTTP 409 (`RECORD_IN_USE`) |
| + | Clinical extensions | Follow-ups, Queues, Live Consultations, Odontograms, Notes, Patient clinician, Outbox, Historical audits | `opd_follow_up`, `patient_queue`, `public_appointment_request`, `live_consultation`, `patient_referral`, `patient_odontogram_entry`, `clinical_note`, `patient(user_id, clinician_user_id)`, `audit_event(actor_id)`, `message_outbox(actor_id)` | HTTP 409 (`RECORD_IN_USE`) |

---

## 2. Deletion Protection Architecture & Invariants

1. **Strict Role Authorization**:
   - Only authenticated actors with `admin` role can initiate doctor deletion.
   - Non-admin callers (Doctor, Nurse, Receptionist, Patient, or unauthenticated requests) are strictly rejected with HTTP 403 / 401 without any database modifications.
2. **Self-Deletion Prevention**:
   - Admins attempting to delete their own account receive HTTP 409 Conflict.
3. **Reference Integrity & Historical Record Preservation**:
   - Before any deletion, all 10 clinical and payroll dependencies plus additional clinical tables are queried.
   - If any referencing record exists, the transaction immediately rolls back and returns HTTP 409 Conflict with error code `RECORD_IN_USE`. All clinical history and doctor records remain untouched.
4. **Complete Atomic Cleanup for Unreferenced Doctors**:
   - In a single database transaction, unreferenced doctors have all operational, profile, authentication, and session records removed:
     - `doctor_hours`
     - `doctor_absence`
     - `doctor_profile`
     - `staff_profile_revision`
     - `staff_role_event`
     - `staff_profile`
     - `staff_access`
     - `session`
     - `verification` (2FA tokens / devices)
     - `account` (auth credentials)
     - `audit_event` (logs immutable `doctor.deleted` record with `actor_id = admin.id` and `resource_id = doctor.id`)
     - `"user"` (base user record)
5. **No Bypass Across Existing Deletion Paths**:
   - `DELETE /v1/doctors/{id}` (Go HTTP API) enforces the full validation, lock, conflict check, and transactional cleanup.
   - `DELETE /api/hms/doctors/{id}` (Next.js proxy route) proxies directly to Go backend `DELETE /v1/doctors/{id}`.
   - `DELETE /api/staff` (Next.js staff route) enforces identical administrator authorization, conflict checks against all 10 dependencies, and transactional cleanup.
6. **Transaction Rollback Guarantee**:
   - If any step fails during deletion (e.g. injected trigger failure on `"user"` deletion), the transaction rolls back completely. Zero partial deletions occur, and all doctor records survive.
   - Retrying the request after resolving the failure succeeds cleanly.

---

## 3. Automated Test Evidence

### A. PostgreSQL Adapter & Go Unit Tests (`TestClinicalTransactions`)
Added Section 14 to [doctors_test.go](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/services/api/internal/adapters/postgres/doctors_test.go):
- **Unauthorized requests**:
  - Unauthenticated `DELETE /v1/doctors/{id}` -> HTTP 401
  - Doctor session `DELETE /v1/doctors/{id}` -> HTTP 403
  - Nurse session `DELETE /v1/doctors/{id}` -> HTTP 403
  - Receptionist session `DELETE /v1/doctors/{id}` -> HTTP 403
  - Patient session `DELETE /v1/doctors/{id}` -> HTTP 403
- **Self-deletion**:
  - Admin `DELETE /v1/doctors/{admin.id}` -> HTTP 409
- **Non-existent doctor**:
  - Admin `DELETE /v1/doctors/{random}` -> HTTP 404
- **10 Separate Dependency Tests**:
  - `patient_case`: Doctor referenced -> HTTP 409, doctor preserved, case preserved
  - `encounter`: Doctor referenced -> HTTP 409, doctor preserved, encounter preserved
  - `appointment`: Doctor referenced -> HTTP 409, doctor preserved, appointment preserved
  - `birth_report`: Doctor referenced -> HTTP 409, doctor preserved, birth report preserved
  - `death_report`: Doctor referenced -> HTTP 409, doctor preserved, death report preserved
  - `investigation_report`: Doctor referenced -> HTTP 409, doctor preserved, investigation report preserved
  - `operation_report`: Doctor referenced as surgeon -> HTTP 409, doctor preserved, operation report preserved
  - `prescription`: Doctor referenced -> HTTP 409, doctor preserved, prescription preserved
  - `ipd_admission_details`: Doctor referenced via encounter -> HTTP 409, doctor preserved, admission preserved
  - `employee_payroll`: Doctor referenced -> HTTP 409, doctor preserved, payroll preserved
- **Successful unreferenced deletion**:
  - Unreferenced doctor deleted -> HTTP 200 `{"deleted": true, "id": "..."}`
  - Zero rows remaining across `user`, `staff_access`, `staff_profile`, `doctor_profile`, `doctor_hours`, `session`, `account`
  - Exactly 1 audit event logged with `action = 'doctor.deleted'`
- **Transaction rollback verification**:
  - Injected trigger on `"user"` deletion causes failure -> HTTP 500
  - All records survive intact (counts unchanged for user, staff_access, doctor_profile, doctor_hours, session, account; zero deleted audit events)
  - Dropping trigger and retrying deletion succeeds with HTTP 200.

Result: `PASS` (7.97s, all tests passing).

### B. Full Go Test Suite
```powershell
go test ./... -count=1
```
Result:
```
ok   hms.local/api/internal/adapters/delivery   1.168s
ok   hms.local/api/internal/adapters/httpapi    1.259s
ok   hms.local/api/internal/adapters/postgres   8.667s
ok   hms.local/api/internal/adapters/privatefiles 0.557s
ok   hms.local/api/internal/adapters/stripe     1.436s
ok   hms.local/api/internal/domain              0.562s
```

### C. Connected Isolated Integration Test Suite (`scripts/integration.mjs`)
Added doctor deletion verification section:
- Verified unauthorized deletion attempts (HTTP 401 / 403)
- Verified self-deletion prevention (HTTP 409)
- Verified non-existent doctor (HTTP 404)
- Verified in-use protection on referenced doctor returning HTTP 409 with `code: 'RECORD_IN_USE'`
- Verified unreferenced doctor deletion through Go proxy (`DELETE /api/hms/doctors/{id}`) with complete database removal and audit event verification
- Verified unreferenced doctor deletion through Next.js staff endpoint (`DELETE /api/staff`) with complete database removal and audit event verification

Result:
```
PASS: admin-only staff provisioning and disablement; scheduling validation, ownership, concurrent booking, idempotency, schedule conflict, lifecycle and released slots.
PASS: doctor deletion parity: authorized admin only, in-use conflict on clinical dependencies, unreferenced transactional cleanup, audit events, across both Next.js and Go proxy paths.
PASS: session auth, CSRF, role denial, patient/doctor record scope and aggregates, validation, disablement, invalid Firebase proof, concurrent message deduplication, captured SMS, Mailpit email/reset, logout revocation.
Connected integration verification passed in isolated schema.
```

### D. Formatting and Types
- `npm run format:check`: PASS (All matched files use Prettier code style)
- `npm run typecheck`: PASS (TypeScript clean)
- `npm test`: PASS (8/8 tests pass)
