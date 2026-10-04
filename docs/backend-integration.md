# Backend integration progress

## Verified step 1: identities and scheduling

The Go API now owns doctor schedules and appointments. Additive migration 002 stores typed doctor hours, appointments, optimistic versions, and absence intervals. All hospital times currently use Africa/Addis_Ababa; split working periods represent breaks. No sample records are copied into the database.

Authenticated, active actors are resolved through Better Auth and database roles. Patient appointment lists and mutations are restricted to their linked patient record; doctor lists are restricted to their own appointments. Staff provisioning remains in the authentication service so password hashing uses Better Auth. Only active administrators can create accounts or disable access; disablement revokes sessions. Creating an account does not claim a verified email or associate an existing patient automatically.

Bookings serialize on doctor and patient records, enforce working hours and overlap checks, and support idempotent retries. Status transitions and schedule edits use optimistic versions. Schedule changes are conservatively rejected while future bookings remain. Optional booking SMS is queued in the booking transaction and contains no clinical problem text. Live SMS credentials are not configured; development uses capture delivery. Existing Mailpit email and Firebase proof verification are retained.

Validation: Go tests/vet, TypeScript, and the running Next/Go/PostgreSQL integration suite. Tests exercise access denial, scoped lists, duplicate retries, concurrent competing bookings, invalid lifecycle changes, staff access changes, and existing authentication/message behavior. Synthetic records are removed afterward.

## Still required for full backend parity

- Integrate the frontend forms, patient portal links, profile editing and staff invitation lifecycle.
- Doctor holidays/absence management, rescheduling, hospital opening hours, appointment fees/payments and timed reminders.
- OPD/IPD, case/admission/discharge workflows, beds and clinical record signing.
- Prescriptions, pharmacy/inventory movements, laboratory/radiology, blood bank, vaccines and odontogram persistence.
- Exact-decimal billing, invoices/payments/refunds, payroll and financial reconciliation.
- Private file storage, reports, attendance, CMS/public booking, queues and other operational modules.
- Provider credential setup and end-to-end Firebase/live SMS verification, deployment hardening, backups and restore checks.

The frontend preview modules are not evidence of backend completion. This document records implemented behavior and remaining work separately.
