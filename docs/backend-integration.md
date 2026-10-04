# Backend integration progress

## Verified step 1: identities and scheduling

The Go API now owns doctor schedules and appointments. Additive migration 002 stores typed doctor hours, appointments, optimistic versions, and absence intervals. All hospital times currently use Africa/Addis_Ababa; split working periods represent breaks. No sample records are copied into the database.

Authenticated, active actors are resolved through Better Auth and database roles. Patient appointment lists and mutations are restricted to their linked patient record; doctor lists are restricted to their own appointments. Staff provisioning remains in the authentication service so password hashing uses Better Auth. Only active administrators can create accounts or disable access; disablement revokes sessions. Creating an account does not claim a verified email or associate an existing patient automatically.

Bookings serialize on doctor and patient records, enforce working hours and overlap checks, and support idempotent retries. Status transitions and schedule edits use optimistic versions. Schedule changes are conservatively rejected while future bookings remain. Optional booking SMS is queued in the booking transaction and contains no clinical problem text. Live SMS credentials are not configured; development uses capture delivery. Existing Mailpit email and Firebase proof verification are retained.

Validation: Go tests/vet, TypeScript, and the running Next/Go/PostgreSQL integration suite. Tests exercise access denial, scoped lists, duplicate retries, concurrent competing bookings, invalid lifecycle changes, staff access changes, and existing authentication/message behavior. Synthetic records are removed afterward.

## Still required for full backend parity

- Complete profile editing and the staff invitation lifecycle; preserve the remaining original patient demographics and ancillary form fields when integrating them.
- Doctor holidays/absence management, rescheduling, hospital opening hours, appointment fees/payments and timed reminders.
- Complete OPD/IPD submodules, admission demographics/vitals, nursing assignment, transfers, amendments and encounter billing. Core cases/admissions/beds/signing are implemented below.
- Prescriptions, pharmacy/inventory movements, laboratory/radiology, blood bank, vaccines and odontogram persistence.
- Exact-decimal billing, invoices/payments/refunds, payroll and financial reconciliation.
- Private file storage, reports, attendance, CMS/public booking, queues and other operational modules.
- Provider credential setup and end-to-end Firebase/live SMS verification, deployment hardening, backups and restore checks.

The frontend preview modules are not evidence of backend completion. This document records implemented behavior and remaining work separately.

## Verified step 2: connected screens and patient access

Users, Schedules, Appointments, patient registration and portal appointments now use persistent records. Administrator-only patient account/care-team linking validates account roles and prevents reassignment of existing portal ownership. The preview role selector never grants API permissions. Unintegrated modules retain their preview notice.

The connected browser test creates a synthetic administrator, provisions a doctor through the UI, saves a schedule, registers a patient, books an appointment and verifies it after reload. Desktop/mobile checks passed without browser errors; all fixtures are removed. Run `node --env-file=apps/web/.env.local scripts/verify-connected.mjs` against local development services. API integration tests also cover ownership linkage and duplicate account-link refusal. The older preview regression script predates these connected workflows and must not be used to submit connected forms.

## Verified step 3: clinical encounters

Migration 003 adds beds, patient cases, OPD/IPD encounters and signed clinical notes. Database indexes prevent two active occupants of a bed and simultaneous active inpatient stays for one patient. Encounters retain the admission-time bed rate in integer minor units. Case/patient/doctor consistency uses a composite foreign key. Admissions support idempotent retries; timestamps are normalized to PostgreSQL precision.

Only the assigned doctor can sign notes or discharge a patient. Signed notes reject database updates/deletes; corrections require another entry while the encounter remains active. Discharge retains the summary and releases bed availability. Patients see only their own encounters; receptionists cannot read clinical notes. Nursing care-team assignment is pending and nursing access to clinical notes remains denied.

Beds, cases, OPD/IPD lists/details and patient portal encounter lists are connected. The browser test verifies bed/case creation, admission, doctor sign-in, discharge after reload and released availability. `HMS_TEST_DATABASE_URL` enables Go PostgreSQL tests in a newly created isolated schema that is dropped afterward; these test bed races, idempotency, ownership, immutable notes and discharge. They never create signed test notes in the real development schema.
