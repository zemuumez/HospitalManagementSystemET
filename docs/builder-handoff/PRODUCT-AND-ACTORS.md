# Original product and actor sequences

## Product model

The original application coordinates hospital administration, patient registration, outpatient visits, inpatient stays, clinical records, diagnostics, pharmacy, stock, billing, staff and public communications. It has a public website and authenticated role workspaces sharing patient and staff identities. It is not one flat collection of unrelated CRUD tables: a clinical action can reserve a bed, create a charge, produce a prescription, consume stock and send a notification.

The target is one hospital. The source dump's SaaS-looking name and Cashier subscription tables do not require a new multi-tenant subscription product. The demo's “Super Admin” button is a demo-login shortcut, not evidence that target hospital staff need unrestricted platform-administrator privileges.

## Shared patient journey

Public visitor submits an enquiry or appointment request → receptionist confirms or registers patient → patient identity/profile is linked to a case or appointment → doctor sees the authorized patient/visit → doctor records diagnosis, instructions, orders and prescription → laboratory/pharmacy staff carry out their permitted work → accountant/receptionist records authorized charges and receipts → patient views permitted records and receives notifications.

An inpatient path adds active case selection, doctor/admission and bed reservation → inpatient child records and charges/payments → bill reconciliation → discharge and bed release. An outpatient revisit is a new visit context linked to the patient/history; it must not overwrite the previous visit. A separate legacy patient-admission record can carry package/insurance/guardian/policy details and a standalone bill. Do not merge these three workflows just because they all mention an admission or patient.

## Administrator

1. Sign in, establish a server-verified session and load an authorized aggregate dashboard.
2. Configure hospital branding, currency, language, schedule, clinical departments, catalogs, modules and staff access.
3. Create/invite staff and maintain patient/staff active state without deleting historical ownership.
4. Manage scheduling, beds, service/insurance/package catalogs, inventory and permitted administrative registers.
5. Review attendance/leave/corrections, billing reports, notices and delivery failures.
6. Inspect audit history and operational health; perform controlled corrections rather than rewriting posted ledgers.

Limits: admin actions still need validation, audit, transactional consistency, safe file access and secret redaction. A browser-supplied role is never authorization. Destructive history removal is not automatically justified by legacy delete buttons.

## Doctor

1. Sign in and see scoped appointments, assigned patients and relevant totals.
2. Open an authorized patient/case/visit and review permitted history.
3. Create/update clinical observations, diagnoses, instructions, prescriptions, operations and visible timelines as permitted by the clinical contract.
4. Request/review diagnostics; maintain consultation schedule and own absence where supported.
5. Revisit or close the relevant encounter through allowed state transitions; print authorized clinical records.

Limits: own/assigned scope must be checked per record, not inferred from list filtering. Current absence API requires a doctor's own ID; admin can use an aggregate register. Doctor access to every original shared show route is not permission to disclose all patient records. Financial administration and stock write access require explicit policy.

## Nurse

1. Sign in and receive the permitted ward/assigned-patient context.
2. Review authorized inpatient details and care instructions.
3. Record nursing observations/administration in supported target workflows and communicate clinical changes.
4. Use own attendance/leave and permitted payroll view.

Limits: nursing chart access and edit boundaries must be reconciled explicitly. Current clinical-read policy differs from some broad original IPD routes. Do not solve this by granting all clinical or billing operations. Nursing-specific target contracts may extend the legacy UI; label those extensions.

## Receptionist

1. Find an existing patient before creating another identity; capture required demographics/contact/custom fields.
2. Book or reschedule an appointment using department, doctor, working hours, holidays, breaks and slot availability.
3. Register case/admission/OPD/IPD using the correct flow and available resources.
4. Manage visitor/call/postal registers and enquiries; guide queue transitions.
5. Perform explicitly authorized payment/receipt and service-catalog operations.

Limits: registration access does not imply all clinical notes or full financial administration. Original package/insurance routes include receptionist access, which must be reconciled with current named permissions. Never permit arbitrary user creation with an escalated role through patient registration.

## Accountant

1. Open scoped finance totals and ledgers.
2. Maintain authorized accounts, invoices, bills, incoming/advance receipts, outgoing payments, expenses/income and payroll.
3. Validate line quantities/prices, issue documents, reconcile payments and produce authorized slips/reports.
4. Correct or reverse posted entries through recorded transitions; reconcile balances and reports.

Limits: outgoing `payments` with account/payee are distinct from patient receipts and provider transactions. Do not authorize clinical edits from finance access. All totals are server-calculated exact money; cross-currency aggregation requires an explicit policy.

## Pharmacist

1. Maintain permitted medicine category/brand/catalog and supplier/batch purchase information.
2. Review authorized prescription or medicine-sale requests.
3. Select usable batches, validate expiry/availability, dispense and record the sale/use atomically.
4. Handle authorized returns/corrections and reconcile stock and medicine bills.

Limits: a generic inventory ledger is not the full pharmacy purchase/billing workflow. No negative stock, duplicate dispensing or silent substitution. Do not assume current general-inventory admin-only policy should be widened simply because pharmacist medicine access exists.

## Lab technician

1. Open assigned/authorized diagnostic work.
2. Resolve pathology/radiology test, category, unit and parameter definitions.
3. Capture sample/result information and patient-specific values; validate units and required fields.
4. Submit/review/release results according to the target state contract; allow authorized report viewing/printing.

Limits: test catalog, order, result, general diagnosis property report and imaging attachment are different entities. Do not grant access to unrelated patient history or permit unreviewed results to appear as final. Broad legacy route access needs an explicit scoped decision.

## Case manager / case handler

1. Find/register authorized patient cases and review case status/doctor assignment.
2. Coordinate admission details, guardian/contact and package/insurance references where permitted.
3. Track the related encounter and administrative follow-up.

Limits: case management does not imply prescribing, administering medications or posting unrestricted payments. Original admission rights and current target policy differ; add action-specific tests before enabling the workspace.

## Patient

1. Sign in or complete an approved account/phone linking flow.
2. View own profile, appointments and permitted clinical/financial records.
3. Request an appointment, access authorized reports/prescriptions and participate in own consultations.
4. View own bills/receipts and initiate supported payment flows; use own card where enabled.

Limits: ownership must hold for list/detail/export/download and mutation routes, including guessed IDs. Patients cannot inspect aggregate blood stock merely because they can read their own blood issue. The broad original odontogram route group is not evidence that patient chart editing is safe or required; record that decision.

## Public visitor and automated actors

Public visitors see only published CMS content, public doctor/service information and safe appointment/enquiry forms. Public queue screens and QR cards must reveal only explicitly approved data. Workers deliver outbox messages and handle provider callbacks with bounded privileges, retry/idempotency and audit context. They do not impersonate an unrestricted administrator.

## Cross-actor sequence acceptance

For each journey create two patients and two staff members, not one fixture. Prove the intended handoffs work and the unrelated actor cannot list, open, change, print or download the other record. Repeat after reassignment, deactivation and session revocation. Verify a reload and a new browser context show committed state. Role buttons and hidden navigation are usability features, never the security boundary.
