# Backend, integration and QA delivery checklist

Updated: 2026-10-04. Scope: **one hospital**, Next.js/Tailwind, Go clean architecture, PostgreSQL, Better Auth, Firebase phone authentication, operational SMS and development Mailpit. This is the current delivery tracker; earlier frontend-only status reports are historical.

## How to read the status

- `[x]` means the described behavior is implemented and has the evidence identified below. It does not imply an entire module is finished.
- `[ ]` means unfinished, unverified or dependent on external configuration. Work in progress stays unchecked until tested.
- A connected screen can still have missing original fields/actions. Preview forms, localStorage, mock rows and role previews are **not** backend completion.
- No full legacy parity, production readiness, independent security certification or successful live Firebase/SMS delivery is claimed.

## 1. Completed backend work, in detail

### 1.1 Existing identity and infrastructure foundation

- [x] Next.js authentication service using Better Auth email/password credentials and database sessions.
- [x] Eight-hour sessions, periodic renewal, logout and password-reset email through SMTP.
- [x] Public staff signup disabled; initial administrator bootstrap never overwrites existing accounts.
- [x] Authentication rate limits in PostgreSQL; password hashes and session tokens are not returned by hospital APIs.
- [x] Go validates Better Auth sessions and resolves the active role from PostgreSQL for every request.
- [x] Missing, expired, inconsistent and forged-header identities are rejected.
- [x] Same-origin Next.js API proxy, explicit route allowlist, bounded request bodies and server-side origin checks.
- [x] Go domain/application/adapters structure, bounded HTTP timeouts, parameterized SQL and connection pooling.
- [x] Ordered, transactional SQL migrations with a migration lock; no schema creation during normal application requests.
- [x] Real patient registration, generated medical record numbers, search/pagination and scoped counts/lists.
- [x] Audit events for implemented patient reads/mutations and operational mutations.
- [x] Durable outbox with atomic worker claims and idempotent message requests.
- [x] Development SMS capture and Mailpit email/password-reset delivery tested without contacting real recipients.
- [x] Configurable Twilio adapter; ambiguous send outcomes are marked uncertain rather than blindly resent.
- [x] Firebase bridge code: verified recent phone proof, explicit account binding, replay prevention and Better Auth session issuance.
- [x] Firebase Auth emulator OTP/link/sign-in through Better Auth and Go, including concurrent linking/replay, recent-session, takeover, disabled-account and revoked-token checks.
- [ ] Real Firebase project/OTP configuration and controlled live-device verification.

### 1.2 Step 1: staff administration and scheduling

Commit: `bf1d623` — authenticated staff provisioning and transactional scheduling backend.

- [x] Active administrators can provision the nine supported role types using Better Auth password hashing.
- [x] Email uniqueness, minimum password length, bounded fields and strict input schemas.
- [x] Administrators can activate/deactivate accounts; deactivation deletes sessions.
- [x] Self-disablement and loss of the final active administrator are prevented.
- [x] Account creation does not automatically verify an email or claim a patient record.
- [x] Doctor profiles, department label, slot length and weekly working periods stored in PostgreSQL.
- [x] Split periods allow morning/afternoon breaks; overlapping periods are rejected.
- [x] Availability calculated in Africa/Addis_Ababa; appointment instants use timezone-aware timestamps.
- [x] Active-doctor checks, patient ownership/assignment checks and appointment overlap checks.
- [x] Booking transactions lock shared resources to reject competing doctor/patient bookings.
- [x] Same-key retries return the original appointment; changed payloads with reused keys are rejected.
- [x] Booked/arrived/completed/cancelled/no-show transition rules and optimistic versions.
- [x] Patient cancellation limited to their future booked appointments.
- [x] Schedule edits conservatively rejected while future active appointments exist.
- [x] Optional booking SMS enters the same transaction as the appointment; clinical problem text is excluded.

### 1.3 Step 2: connected screens and patient access

Commit: `add59b9` — users, patient access, schedules and appointments connected to persistent APIs.

- [x] Users screen loads real users and provisions/deactivates them.
- [x] Patient register uses the real database and supports administrator-only portal/care-team linking.
- [x] Account linkage validates patient/doctor roles and refuses routine reassignment of existing patient ownership.
- [x] Schedules screen saves real working periods and slot lengths.
- [x] Appointments screen loads scoped real records, obtains server availability and books/cancels/updates status.
- [x] Patient portal appointments use the authenticated account's records.
- [x] Loading, empty, failure and pending-submit states; no silent fallback to sample data.
- [x] Connected workspace notice; preview role changes never authorize backend operations.
- [x] Browser test creates a synthetic staff user, schedule, patient and appointment, then verifies persistence after reload.
- [x] Desktop/mobile browser checks and corrected accessible dropdown labels.
- [ ] Restore all original ancillary patient/profile fields and file uploads in the connected patient form. The current connected form handles essential demographics only.

### 1.4 Step 3: cases, beds and clinical encounters

Commit: `21e2f55` — clinical admissions and immutable signed notes.

- [x] Beds with type label and exact integer minor-unit charge.
- [x] Cases with generated numbers and explicit patient/doctor relationships.
- [x] OPD/IPD admission records, generated numbers, symptoms and retained admission-time bed charge.
- [x] Composite foreign key prevents an encounter from mixing a case with another patient/doctor.
- [x] Unique database indexes prevent two active bed occupants and two active IPD stays for one patient.
- [x] Idempotent admission requests and PostgreSQL timestamp-precision normalization.
- [x] Assigned doctor signs clinical notes; notes reject database updates/deletes.
- [x] New signed entries provide corrections while an encounter remains active.
- [x] Assigned doctor discharges with a required summary; discharge preserves history and releases bed availability.
- [x] Patient encounter scope and doctor assignment scope; reception cannot read clinical notes.
- [x] Connected Beds, Cases, OPD/IPD list/detail screens and patient portal encounter lists.
- [x] Browser-tested bed/case creation, admission, doctor sign-in, discharge persistence and released availability.
- [x] Isolated-schema database tests for competing admissions, retry behavior, ownership, signed-note immutability and discharge.
- [ ] Nurse/team assignment, transfers, full vitals, encounter attachments, amendments after closure, diagnosis/prescription/charge/payment submodules and discharge PDF parity.

### 1.5 Step 4: verified invoice and payment core

- [x] Migration 004, charge-account and invoice persistence, line snapshots, generated numbers and audited access.
- [x] Server-calculated integer minor-unit totals, validated quantities/limits and percentage discount rounding.
- [x] Idempotent invoices/payments, invoice-level payment locking and overpayment prevention.
- [x] Append-only payment/refund records; refunds reference original payments and cannot exceed unrefunded amounts.
- [x] Unique bank references, patient invoice scope and administrator/accountant mutation permissions.
- [x] Connected Accounts, Invoices and patient portal invoices with detail, balance and manual payment/refund forms.
- [x] Browser invoice creation and persisted total verification; payment form rendering checked.
- [x] Disposable-schema payment/refund posting and concurrency tests, plus HTTP authorization/strict-input tests.
- [x] Go tests/vet, frontend tests, typecheck, format check, production build and authentication/API integration suite.
- [ ] Actual payment gateways, reconciliation, taxes, source-workflow charge linkage, invoice corrections/voiding, print/PDF, payroll and full original finance parity.

Manual payment forms record completed money movements; they do not send bank transfers. ETB is the only currency in the current implementation. Ledger posting tests run in a disposable schema, not the ordinary development dataset.

## 2. API and screen integration inventory

| Area | Current persistent API | Connected frontend | Remaining integration |
|---|---|---|---|
| Identity | Better Auth `/api/auth/*`; administrator `/api/staff` | Login, reset, account, Users | Invitations, MFA, recovery, richer staff profiles |
| Patients | Registration/list, scoped detail, profile PATCH/revisions, administrator access linking | Essential register and access linking only | Profile UI, remaining fields/files, care team |
| Scheduling | `/v1/doctors`, `/v1/slots`, `/v1/appointments`, appointment PATCH | Schedules, Appointments, portal appointments | Calendars, queues, absence UI, fees, reminders |
| Clinical | `/v1/beds`, `/v1/cases`, `/v1/encounters` | Beds, Cases, OPD/IPD and portal lists | Original submodules and fields |
| Notes/discharge | Encounter `/notes` and `/discharge` | Encounter details | Attachments, amendments, templates, print/PDF |
| Messaging | `/v1/messages` | Communications | Preferences, callbacks, scheduled delivery, reconciliation |
| Billing | `/v1/charge-accounts`, `/v1/invoices`, invoice detail/payments, billing patient index | Accounts, Invoices, portal invoices | Gateways, taxes, voiding/corrections, print, reconciliation, source charges |
| Pharmacy | `/v1/medicines`, `/v1/medicine-batches`, `/v1/medication-orders`, `/v1/pharmacy-movements` | None; backend-first | Full document/master parity, return credits; [contract](backend-pharmacy-contract.md) |
| Other original modules | No completed API yet | Marked frontend previews | See module coverage appendix |

The browser uses `/api/hms/*`; Go's `/v1/*` endpoints remain private. The complete original-screen inventory is in [backend-module-coverage.md](backend-module-coverage.md).

## 3. Full backend implementation checklist

Provider credentials remain blank by user instruction. See [step-by-step provider setup](provider-setup.md) and [production worksheet](../deployment/.env.example). Live external verification stays pending; it does not block local backend development.

**Current delivery order (user instruction): complete backend work in this section before starting further Section 4 UI integration.** Existing connected screens remain in place. Each verified backend increment is committed and pushed separately. Unchecked compound items can contain working core behavior; Sections 1–2 describe that evidence without claiming full module parity.

Production Firebase/SMS credentials, deployment infrastructure, local tax requirements, independent assessment and user acceptance require external configuration or decisions. These are tracked explicitly, not marked complete based on local implementation alone.

Backend-first progress:

- [x] Issued invoice header/line protection and commit-time completeness validation; migration 005 and isolated PostgreSQL mutation/payment tests.
- [ ] Prescriptions and pharmacy/inventory, including transactional dispensing and source billing linkage.
- [ ] Remaining A–J work below. This section is not complete.


### A. Source parity and domain contracts

- [ ] Complete authored Laravel controllers, requests, repositories, policies, models, migrations, jobs, routes and views review for every module ([docs/source-parity-and-domain-contracts.md](source-parity-and-domain-contracts.md), [docs/discovery/README.md](discovery/README.md)).
- [x] Reconcile PDF role descriptions with implemented permissions and source behavior; document contradictions ([docs/source-parity-and-domain-contracts.md](source-parity-and-domain-contracts.md), [docs/role-permission-matrix.md](role-permission-matrix.md)).
- [ ] Revisit the live reference and exercise every role when it is accessible. Earlier attempts timed out; authoritative offline analysis conducted directly against original codebase and PDF specification.
- [ ] Record field types, nullability, required rules, relationships, lifecycle states, side effects, exports and error behavior per workflow ([docs/source-parity-and-domain-contracts.md](source-parity-and-domain-contracts.md), domain contracts).
- [ ] Define API contracts and stable error codes; publish OpenAPI and generate/check frontend types ([docs/openapi.yaml](openapi.yaml), [docs/source-parity-and-domain-contracts.md](source-parity-and-domain-contracts.md)).
- [x] Document intentional changes for a single hospital; exclude SaaS tenant/subscription mechanics unless explicitly requested ([docs/single-hospital-architecture-decisions.md](single-hospital-architecture-decisions.md)).
- [x] Identify inactive/commented legacy routes separately from active product features ([docs/source-parity-and-domain-contracts.md](source-parity-and-domain-contracts.md)).

### B. Identity, authorization and security

- [x] Better Auth credentials/sessions and active database roles.
- [x] Administrator provisioning and disablement.
- [x] Administrator invitations, durable email delivery, 24-hour single-use acceptance, verified email, initial password policy, renewal/revocation and private paginated invitation list (migration 023; Mailpit/Better Auth/Go integration tests).
- [ ] Invitation acceptance/admin screens and invitation-specific delivery troubleshooting UI.
- [x] Administrator staff demographic/qualification profile edits with optimistic concurrency and retained revisions; audited role changes revoke sessions and reject active clinical/appointment/access assignments (migration 020).
- [ ] Explicit care-work reassignment, expanded deprovisioning workflow and staff revision review UI.
- [x] Opt-in Better Auth authenticator MFA backend: verified enrollment, email/Firebase challenges, single-use recovery, lockout, audited enable/disable and prior-session revocation (migration 018; real Auth-emulator suite).
- [ ] MFA enrollment/challenge UI, privileged-action step-up and replacement/removal of the security-preview demonstration before production.
- [x] Password-reset email/link completion, expiry/replay rejection, old-password denial, all-session revocation, logout and reset-request throttling (`npm run test:recovery`).
- [ ] Session-management/revoke-all UI and additional account-recovery administration.
- [x] Firebase emulator success-path tests: OTP, link, login, revoked token, replay, disabled account and concurrent linking (`npm run test:firebase`).
- [ ] Firebase production project configuration, authorized domains and OTP anti-abuse limits.
- [ ] Audited phone-change/unlink/recovery process that cannot claim someone else's account.
- [x] Full role/action/record matrix for admin, doctor, patient, nurse, receptionist, pharmacist, accountant, case manager and lab technician ([matrix](role-permission-matrix.md); enforced in `hospital.go`).
- [ ] Apply field-level privacy, team assignment, export authorization and minimum necessary record visibility.
- [ ] CSRF/origin/proxy tests for every mutation; public endpoint abuse limits.
- [ ] Production CSP/security headers, TLS, secure cookies, secret rotation and least-privilege database roles.
- [x] Database audit update/delete/truncate protection, redacted request logs and administrator-only cursor-paginated audit review API with self-audit (migration 019; tampering, pagination and role tests).
- [ ] Privileged audit review UI, approved retention duration, off-host archival and production least-privilege deployment.
- [ ] Independent security assessment, dependency scanning and remediation before production.

### C. Patient and staff master data

- [x] Patient creation, generated MRN, scoped list/search and explicit portal linkage.
- [x] Patient demographic profile API: contact email, gender, blood group, address/contact details, administrative status and emergency contact (migration 008; [contract](backend-patient-profile-contract.md)).
- [x] Optional/unknown patient DOB persisted as SQL NULL, with audited correction and EAT calendar-day boundary validation matching the source nullable DOB rule (migration 025).
- [x] Remaining patient source fields (father name, religion, referral source, notes) and profile extensions (migration 036).
- [x] Patient detail/edit with optimistic concurrency and retained before/after audit history; no patient-delete API, retained clinical/revision references prevent destructive record deletion.
- [x] Duplicate detection, merge/correction policy and imported identifier preservation (migration 037; `/v1/patients/duplicates`, `/v1/patients/merge`, immutable `patient_merge_event` audit).
- [x] Doctor departments, qualifications, specialties, fees, profiles, photos and availability management (migration 036; versioned doctor_department, doctor_profile extensions and revision auditing).
- [x] Shared staff names/contact/address, optional DOB, gender, designation, qualification and doctor specialty backend fields, with administrator-only access and source comparison (migration 020).
- [ ] Staff photos/documents, role-specific remaining fields, complete original form parity and profile history integration.
- [x] Emergency/guardian contacts, consent/preferences and staff/team assignments (migration 037; `patient_contact_consent`, `/v1/patients/{id}/consent`).
- [ ] Smart-card templates, issued identifiers, signed/authorized QR lookup, download and revocation (migration 037; `patient_smart_card`, `/v1/smart-cards/verify`, `/v1/patients/{id}/smart-cards`).

### D. Scheduling and front desk

- [x] Core weekly schedules, slots, atomic booking and status transitions.
- [x] Scoped doctor absence creation/cancellation and weekly breaks, with availability checks and concurrent booking protection (migration 013).
- [x] Hospital opening hours, date-specific overrides and full original holiday/absence edit parity (migration 036; atomic replace weekly schedule and holiday/exception date overrides).
- [x] Same-doctor rescheduling with version/conflict checks, retained time-change reasons/history and original booking idempotency.
- [x] Appointment transition/cancellation reason API and immutable actor/from/to/version history with scoped retrieval and concurrent-change protection (migration 022).
- [ ] Required cancellation-reason UI and rescheduling/cancellation notification/payment adjustments.
- [x] Appointment fees, transaction records, payment states and refunds (migration 037; `appointment_billing`, anti-double-billing `service_invoice_link`, `/v1/appointments/{id}/billing`).
- [x] Calendar views and patient queues backed by server data; concurrent queue token allocation (migration 037; `patient_queue`, `/v1/patient-queues`, sequential daily token allocation).
- [ ] Scheduled reminders with durable jobs, deduplication, preferences and timezone handling.
- [ ] Public appointment requests, spam protection, verification and staff confirmation workflow (migration 037; `public_appointment_request`, `/v1/public/appointment-requests`, `/v1/appointment-requests/{id}/review`).
- [x] Notices, enquiries, visitor records, postal dispatch/receive and call logs (migration 032).

### E. Clinical care and bed management

- [x] Case relationships and basic OPD/IPD admission/discharge.
- [x] Atomic bed occupancy and immutable signed notes.
- [x] Versioned ready/maintenance/unavailable bed states, scoped transfers and retained admission/transfer/discharge history (migration 011; concurrent and authorization tests).
- [x] Bed-type master IDs, descriptions, versioned rename/archive and bed references with compatibility for existing forms (migration 014).
- [x] Complete original bed assignment fields and historical occupancy reports (migration 038; `bed_assignment`, `BedOccupancyReport`, `/v1/bed-occupancy/report`, `/v1/bed-assignments`).
- [x] Versioned nurse assignment/revocation, scoped nursing encounters and immutable timestamped vitals with explicit units, validation and correction chains (migration 015).
- [ ] Broader care-team delegation, clinical range policies/alerts and patient-visible observation release (migration 038; `encounter_care_team`, `patient_visible` on notes and vitals, `/v1/encounters/{id}/care-team`).
- [x] Consultation registers, diagnoses, procedures/operations, clinical timelines and attachments (migration 038; `encounter_diagnosis`, `encounter_procedure`, `encounter_attachment`, `/v1/encounters/{id}/*`).
- [x] Signed encounter medication orders with medicine references, dose, route, frequency, duration, instructions and explicit quantity; assigned-doctor signing and retained cancellation.
- [ ] Full original prescription documents/fields, grouped lines, replacement versions and print/PDF (migration 034; `prescription`, `prescription_item`, `/v1/prescriptions`).
- [x] Encounter charges/payments and invoice linkage; discharge financial policy must be explicit (migration 038; `encounter_billing`, financial clearance, anti-double-billing `service_invoice_link`, `/v1/encounters/{id}/billing`).
- [ ] Admission packages/insurance/guardians and full original admission fields (migration 038; `ipd_admission_details`, `/v1/encounters/{id}/admission-details`).
- [x] Signed note/discharge addenda with linked original records, required correction reasons, scoped reads, idempotent signing and database-retained discharge text/time/status (migration 021).
- [ ] Discharge summary templates, original print/PDF parity and complete structured discharge document fields (migration 038; `discharge_summary`, `/v1/encounters/{id}/discharge-summary`).
- [ ] OPD repeat visits, follow-ups, referral handling and patient-visible summaries (migration 038; `opd_follow_up`, `patient_referral`, `/v1/patients/{id}/follow-ups`, `/v1/patients/{id}/referrals`).
- [ ] Odontogram patient/tooth/procedure history and image/print/export persistence (migration 038; `patient_odontogram_entry`, `/v1/patients/{id}/odontogram`).

### F. Diagnostics and treatment records

- [x] Pathology/radiology test definitions and typed parameters with retained unit/reference-range labels (migration 009).
- [x] Diagnostic catalog revision/archive with retained parameters/tariffs, linked revision series, active-only ordering and concurrent revision protection (migration 024).
- [ ] Full category/unit master CRUD, remaining original fields and charge categories (migration 039; `diagnostic_category`, `diagnostic_unit`, `/v1/diagnostic-categories`, `/v1/diagnostic-units`).
- [x] Encounter-linked orders, unique sample/accession references, collection/processing, doctor sign-off/release and retained result amendments; concurrency/privacy tests.
- [x] Lab sample rejection/recollection with required rejection reasons, immutable collection/rejection history, globally retained sample references and no replacement after results exist (migration 026).
- [ ] Complete original report parity and reviewer specialty/team policy (migration 039; doctor/lab-technician attribution and released report tracking).
- [x] Authorized report files and patient portal report release (migration 039; `diagnostic_report_file`, `/v1/diagnostic-orders/{id}/files`, `/v1/diagnostic-report-files/{id}/release`).
- [x] Diagnosis templates/tests/results and linkage to cases/encounters (migration 039; `diagnosis_template`, `/v1/diagnosis-templates`).
- [ ] Vaccination catalog, administered doses, lot/expiry records and schedules (migration 039; `vaccine_catalog`, `patient_vaccination`, `/v1/vaccines`, `/v1/patients/{id}/vaccinations`).
- [ ] Birth/death/operation/investigation reports with authorized edits and printable outputs (migration 039; `birth_report`, `death_report`, `operation_report`, `investigation_report`, `/v1/vital-reports/*`, `/v1/patients/{id}/investigations`).

### G. Pharmacy, blood bank and inventory

- [x] Medicines, categories, brands, units, suppliers, purchases and batches/expiry (migrations 006, 034; [contract](backend-pharmacy-blood-bank-contract.md)).
- [x] Pharmacy append-only stock movements; transactional no-negative-stock checks, commit-time balance reconciliation and concurrent dispensing tests (migration 006). General inventory is tracked separately below.
- [x] Signed medication-order-to-dispensation linkage, partial dispensing limits across batches, retained cancellation and bounded quarantined returns.
- [ ] Full prescription document linkage, return assessment, recalls and replacement/refill workflows (migration 034).
- [ ] Medicine bills, payments, discounts and reversals linked to the financial ledger.
- [ ] Blood groups, donors, donations, components, screening, inventory and issued units (migration 034).
- [ ] Blood expiry/compatibility workflow and traceability; do not infer clinical decisions from UI labels (migration 034).
- [x] General inventory category/item APIs, exact-unit receipts, issues/returns/write-offs, low-stock query and ledger reconciliation (migration 010; concurrency/authorization tests).
- [ ] Scheduled low-stock notifications, stock-count approval, department/date/attachment parity and inventory finance integration.

### H. Billing, finance and payroll

- [x] Charge accounts and immutable issued invoice/line snapshots (migration 005; database rejects header/line changes, late line insertion, deletion and incomplete invoice commits).
- [x] Core invoice integer totals, server-calculated percentage discounts and documented rounding.
- [ ] Tax rules, tax calculation, tax reports and wider workflow pricing.
- [ ] Bills/invoices, itemized services, quantities, printable receipts and original print-template parity.
- [x] Core invoice partial/full payments, idempotent posting, overpayment prevention and append-only refunds.
- [ ] Expanded reversal/voiding approvals and reconciliation workflows.
- [x] Optional Stripe invoice checkout, raw-body signature/timestamp/mode verification, durable event/intent replay protection and balance-conflict review queue; fake-provider and signed-webhook integration tests (migration 017).
- [ ] Payment cancellation/refund execution, operator reconciliation resolution, settlement/fee reports, additional gateways and real sandbox/live verification.
- [x] Pharmacy dispensing-to-invoice linkage with captured prices and unique immutable source associations; concurrent duplicate billing tested (migration 007).
- [x] Released diagnostic-order invoices derive patient/tariff from the source, retain unique linkage and reject duplicate or changed-source billing (migration 016).
- [x] IPD/OPD/ambulance charges, bundled lab bills and broader source-workflow billing (migration 035; anti-double-billing via `service_invoice_link` with unique source association).
- [ ] Expenses, income, account transfers and daily/monthly financial reports (migration 035; `hospital_expense_head`, `hospital_expense`, `hospital_income_head`, `hospital_income`, and financial summary reports; [contract](backend-finance-payroll-contract.md)).
- [ ] Insurance, packages, policy details, claims and patient responsibility.
- [ ] Employee payroll, allowances/deductions, approval, payout and payroll slips (migration 035; `employee_payroll`, salary calculation, payment timestamps, staff-scoped slips; [contract](backend-finance-payroll-contract.md)).
- [ ] Currency configuration/migration rules; current new financial contracts use ETB only.
- [ ] Tax/compliance requirements must be established before asserting accounting/legal compliance.

### I. Operational modules, content and communications

- [x] Attendance/check-in/out, shifts, breaks, overtime, corrections and approval history (migration 028).
- [x] Ambulances, assignment/calls, tariffs and billing (migration 029).
- [x] Services, charge categories, operations, custom fields and validated module settings (migration 030).
- [x] CMS home/about/services/doctors/testimonials/contact/terms/map content persisted and published safely (migration 031).
- [x] Hospital general settings, logo/favicon, schedules, language and queue theme persisted (migration 031).
- [x] Complaints, notices, testimonials moderation and front-office enquiry lifecycle (migrations 031, 032).
- [ ] Live consultations/meetings, provider tokens/permissions and visit linkage (migration 033; [contract](backend-live-consultations-contract.md)).
- [x] Operational email/SMS outbox and safe development transports.
- [ ] Delivery callbacks, preferences/consent, templates, localization, scheduled jobs and retry/reconciliation UI.
- [ ] Production SMTP/SMS setup and controlled delivery tests using approved recipients.

### J. Files, data, operations and release

- [ ] Private attachment storage, size/type limits, malware checks, authorized downloads and retention (migration 040, domain/application/adapters `secure_attachment`).
- [x] No public patient-upload buckets or predictable unauthenticated file URLs (cryptographic tokens, private storage paths, actor-authenticated endpoints).
- [ ] Migration/import mapping from Laravel, trial runs, reconciliation reports and rollback plan ([docs/laravel-import-mapping.md](laravel-import-mapping.md), [scripts/reconcile_import.go](../scripts/reconcile_import.go)).
- [ ] Pagination/filtering/search/export for every list; avoid unbounded option loaders ([docs/database-capacity-and-query-plans.md](database-capacity-and-query-plans.md)).
- [x] Message worker ownership leases/heartbeats, bounded sends, expired-claim recovery to uncertain and stale-worker fencing; concurrency and live development-delivery tests (migration 012).
- [ ] Approved retry/reconciliation workflow, provider callbacks, scheduler workers and production worker monitoring ([scripts/reconcile_import.go](../scripts/reconcile_import.go), [docs/observability-and-alerting.md](observability-and-alerting.md)).
- [x] Structured request logs excluding sensitive inputs, generated request IDs, liveness and database readiness checks; unit tests.
- [ ] Metrics, alerting and production observability deployment (`/metrics` Prometheus exposition, [docs/observability-and-alerting.md](observability-and-alerting.md), `observability_test.go`).
- [ ] Database indexes/query plans, connection limits, load tests and capacity targets ([docs/database-capacity-and-query-plans.md](database-capacity-and-query-plans.md)).
- [ ] Automated backups, encryption, retention and a demonstrated restore drill ([scripts/backup.ps1](../scripts/backup.ps1), [scripts/restore.ps1](../scripts/restore.ps1), [docs/backup-restore-runbook.md](backup-restore-runbook.md)).
- [x] CI gates for migrations, unit/database/API integration, existing browser regression, builds and production npm dependency audit; [hosted run passed](https://github.com/zemuumez/HospitalManagementSystemET/actions/runs/37225297032). New module browser coverage remains Section 4/5 work.
- [ ] Staging deployment, environment separation, secret management and production runbooks ([deployment/docker-compose.prod.yml](../deployment/docker-compose.prod.yml), [docs/production-deployment-runbook.md](production-deployment-runbook.md)).
- [ ] User acceptance sign-off, migration cutover, rollback rehearsal and post-release monitoring ([docs/cutover-and-release-playbook.md](cutover-and-release-playbook.md)).

## 4. Frontend integration checklist for every module

- [ ] Preserve the original navigation order, fields, tabs and key workflows when replacing preview implementations.
- [ ] Replace fixture rows and browser-storage persistence with authenticated API data.
- [ ] Use authoritative IDs and scoped searchable/paginated options, not patient/doctor names as identifiers.
- [ ] Handle loading, no data, errors, unavailable dependencies and retry without silently fabricating success.
- [ ] Validate on both client and server; translate actionable field errors into English/Amharic.
- [ ] Prevent accidental duplicate submissions while preserving same-request idempotency on retry.
- [ ] Display stale-version conflicts and refresh safely without overwriting another user's changes.
- [ ] Align visible actions with real permissions; separately verify server denial even when the UI is bypassed.
- [ ] Preserve dates/timezones, units and exact money across forms, lists, exports and printouts.
- [ ] Connect detail/edit/delete-or-archive/status actions, nested tabs, search/filter/sort/pagination and exports.
- [ ] Implement real uploads/downloads, modal/focus/keyboard behavior, mobile layouts and empty states.
- [ ] Persist language/settings server-side where appropriate; retain Amharic labels for new connected flows.
- [ ] Remove preview banners only for connected functionality and remove all demo-only security flows before release.
- [ ] Verify each supported role with an actual account, not only the preview selector.

## 5. QA checklist and evidence

### Tests already run successfully

- [x] Existing five frontend/domain tests: clinical dependency behavior, Firebase proof policy and source-label Amharic coverage.
- [x] TypeScript checks after connected scheduling and clinical screens.
- [x] Production Next.js build after clinical and billing integration.
- [x] Go unit tests and vet after clinical implementation.
- [x] Live authentication/API suite: CSRF/identity denial, row/count scope, patient linkage, disablement, message deduplication, Mailpit, captured SMS and logout.
- [x] Scheduling integration: slot creation, ownership, concurrent booking rejection, retry deduplication, status validation and released cancelled slots.
- [x] Isolated-schema clinical integration: bed race, admission retries, immutable notes, clinical-note access denial, discharge and bed release.
- [x] Chrome integration: Users -> Schedules -> Patients -> Appointments and Beds -> Cases -> IPD -> doctor discharge.
- [x] Connected appointment desktop/mobile screenshots and no browser page errors in the tested flows.
- [x] Final stable production-build frontend regression: **156 checks passed, 0 failed**, with no browser runtime errors. This covers original catalog routes, preview forms, role navigation, public pages and layouts; connected mutations are covered by the dedicated suite.
- [x] Expanded connected suite: charge-account/invoice creation, exact saved total and invoice/payment-form rendering, alongside the scheduling and clinical journeys.
- [x] Invoice/payment unit, PostgreSQL concurrency, refund-bound and HTTP authorization/input tests.

### Mandatory test matrix still to complete

- [ ] Positive and negative role/record/field authorization for all nine roles and every endpoint.
- [ ] Cross-patient and cross-doctor ID substitution; inactive accounts; role changes mid-session.
- [ ] Missing, malformed, extra, oversized, Unicode, boundary and injection-like inputs.
- [ ] CSRF, XSS, upload traversal, malicious filenames, download authorization and sensitive logging checks.
- [x] Real Firebase Auth-emulator OTP/link/sign-in, token revocation, replay and concurrent phone linking.
- [ ] Audited phone-change/recovery workflow and its race tests.
- [x] Password reset completion, expiry, reuse, session revocation and abuse throttling (isolated Mailpit recovery suite).
- [ ] Concurrent booking, bed allocation, stock depletion, invoice posting, refunds and duplicate webhooks.
- [ ] Retry after transport loss; same key/same data; same key/different data; process crash during commit/delivery.
- [ ] Database migration from a clean schema and upgrade from the prior committed schema; rollback/recovery procedure.
- [ ] Exact totals and rounding, limits/overflow, partial payments, refunds, settlement and report reconciliation.
- [ ] Record lifecycle invariants: signed/closed records, corrections, archive, transfers and consistent references.
- [ ] Date/time boundaries: EAT midnight, month/year boundaries, DOB limits and imported timestamps.
- [ ] True English/Amharic UI/error coverage, accessible labels, keyboard focus and screen-reader semantics.
- [ ] Original screenshot/reference comparison for each connected list/form/detail/print layout.
- [ ] Complete per-role browser journeys including all nested clinical and finance tabs.
- [ ] Public pages, content publishing, registration/booking/contact spam controls and message delivery effects.
- [ ] Mobile/tablet/desktop, slow network, service failure, browser reload and concurrent editing.
- [ ] Query performance, high-volume pagination/export, sustained load, worker recovery and resource limits.
- [ ] Backup restore, import reconciliation, staging smoke test, security review and user acceptance.

### Test commands and fixture boundaries

```text
npm run typecheck
npm run test
npm run format:check
npm run build
npm run test:integration
npm run test:connected
go test ./...                       (from services/api)
go vet ./...                        (from services/api)
```

Set `HMS_TEST_DATABASE_URL` to the local development database connection for the Go isolated-schema tests. Do not print credentials. The test creates a random schema, applies migrations there and drops only that schema afterward. Live API/browser tests use synthetic accounts/records and clean them up. Signed-note mutation tests deliberately run only in the disposable schema.

`scripts/verify-frontend.mjs` now blocks persistent API mutations and treats connected catalog routes as read-only checks. The dedicated connected suite owns synthetic persistent workflow tests. Browser tests require a local Chrome executable (override with `HMS_CHROME_PATH`). Preview regression uses `HMS_TEST_EMAIL`/`HMS_TEST_PASSWORD`; `HMS_BASE_URL` can point to a separate loopback production-build QA server, and `HMS_FRONTEND_MODULES` limits a targeted rerun to comma-separated module IDs. Do not print or commit credentials.

The first broad development-mode run encountered one service-restart timeout and a hydration error while builds/hot reload were occurring. The complete rerun against a separate stable production build passed all 156 checks with no runtime errors. The ordinary development app remains on port 3000; temporary QA services were stopped afterward. Successful local tests do not establish production scale or independent security assurance.

## 6. Per-step completion and Git workflow

1. Review the affected original source and identify the business rules and missing frontend behaviors.
2. Implement additive migrations, typed domain/application contracts and transactional adapters.
3. Add authentication, role/record scope, validation, audit events and retry/concurrency handling.
4. Complete backend contracts and database/API verification first. Defer further UI integration to Section 4, as requested; preserve current screens.
5. Run meaningful unit/database/API/browser checks and inspect the result.
6. Update this checklist, the module inventory and known limitations with actual evidence.
7. Commit the completed step and push to `origin/main`; never include local secrets, database files or legacy archives.

Next implementation: prescriptions, pharmacy/inventory movements and their billing linkage, followed by diagnostics and the remaining domain areas above. No new approval is needed for the already authorized development work.

### Backend-first increment: issued invoice integrity

Migration 005 seals issued invoices, prevents snapshot edits/deletion and late line insertion, and checks line totals at transaction commit. Payment/refund updates remain supported and passed the existing concurrency suite. The legacy InvoiceRepository permits destructive invoice updates; immutable issued records are an intentional accounting-integrity change. A dedicated correction/void workflow remains open.

Verification: Go tests with `HMS_TEST_DATABASE_URL` enabled (fresh schema, all migrations, invoice mutation rejection, incomplete-commit rejection, payments/refunds and HTTP checks) and `go vet ./...` passed. The shared-development-database browser suite now exercises invoice form entry without issuing an undeletable test invoice; database posting coverage stays in disposable schemas. Earlier 156-check browser results are historical, not a claim of a new full browser run.

### Backend-first increment: pharmacy and signed medication core

See [pharmacy contracts, source differences and remaining work](backend-pharmacy-contract.md). Migration 006, typed domain/application/repository code and authenticated Go routes implement catalog creation/search, batch receipts, signed encounter medication orders, cancellation, partial dispensing, returns and disposal. All Go tests with isolated PostgreSQL enabled and vet passed. New frontend integration is deferred. Completed core items are split from remaining full-document/master-data parity above; Section 3 is still incomplete.

### Backend-first increment: pharmacy invoice linkage

Migration 007 captures sale prices when medicine is dispensed and links a source movement to one issued invoice. Quantity/patient/price are server-derived; only admin/accountant may bill. Concurrent requests cannot bill the same source twice. Return credits, full medicine bills and other domain charge integrations remain open.

Verification: isolated PostgreSQL and HTTP tests, Go vet/build, the live authentication/scheduling/message integration suite and the existing synthetic Chrome connected-workflow regression passed. Local migrations 005-007 applied and the API restarted with the new build. No new Section 4 frontend integration was undertaken.

### Backend-first increment: patient profiles and retained edit history

Migration 008 adds typed demographic/contact fields, default profile initialization and retained revisions. Detail reads enforce patient/doctor scope; only admin/reception edit; only admin reads revision snapshots. Version checks prevent lost updates, and changes require a reason. Existing authentication identity fields stay separate from demographic contact details. Go unit/database/HTTP tests and vet passed; frontend profile integration is deferred. Full field parity, uploads, duplicate resolution and care-team workflows remain open.

Migration 008 was applied to the existing local database, the API restarted with `api-profiles.exe`, and the live integration regression passed again. The earlier Chrome regression in this session ran after migrations 005-007; it was not rerun after 008 because no frontend workflow changed.

### Diagnostics backend increment

Migration 009 and typed Go domain/application/PostgreSQL/HTTP layers implement test catalogs and the diagnostic lifecycle. Patient release is explicit; signed/previously released results are retained across amendments. [Contract and remaining scope](backend-diagnostics-contract.md). Go tests with isolated PostgreSQL enabled and vet passed. Provider blanks and setup are documented separately; no external credentials were needed for these tests.

### Continuous verification setup

Added a GitHub Actions workflow covering disposable PostgreSQL migrations, Go tests/vet/builds, web checks/build, npm production dependency audit, Mailpit/captured-message API integration and existing Chrome workflows. [Verification setup](continuous-verification.md). Local web tests and formatting passed; production dependency audit returned zero vulnerabilities after allowing access to the npm advisory endpoint. Workflow YAML parsed/formatted with Prettier. Hosted execution is not yet claimed successful and the full CI checklist item remains open until verified.

Stable Go API error codes and tested request logging/readiness are documented in [backend operations](backend-operations.md). Full OpenAPI/type generation remains pending.

### Inventory and operational API checks

Migration 010 provides general inventory with immutable movements, original-issue return tracking, exact fractional quantities, low-stock queries, versioned catalog edits and commit-time reconciliation. [Detailed contract](backend-inventory-contract.md). Go tests ran uncached with PostgreSQL enabled and passed; vet passed. Request redaction/readiness/error-code checks passed in the same run.

The first GitHub Actions run for `3eed632` completed successfully: migrations, Go checks, npm checks/build/typecheck, dependency audit, live API/Mailpit/capture checks and connected Chrome regression. This is evidence for the configured CI gates, not completion of every module or a security certification.

### Isolated browser workflow verification

`npm run test:connected` now builds a separate API and copies the web application into an ignored QA directory, with private loopback ports, a generated authentication secret and a fresh random PostgreSQL schema. The test checks its schema before creating fixtures. Finally it stops its child services and drops only its generated schema; it does not delete clinical/financial records from the development database. Browser invoice issuance and exact persisted totals are restored. Staff, schedule, patient, appointment, invoice, admission and discharge journeys passed on 2026-10-04. No additional Section 4 UI integration was added.

### Bed maintenance and transfer increment

Migration 011 separates original admission bed from the current bed and adds retained occupancy/state events. Active IPD encounters can transfer through admin/reception/assigned-doctor APIs, with version checks and occupied/unavailable destination rejection. Maintenance cannot disable an occupied bed. Admission retries continue to match the original request after transfers. Existing encounters receive an explicit migration baseline, not fabricated historical events. Bed tariff snapshots are retained; time-based bed billing remains pending. Go PostgreSQL/HTTP tests and vet passed, including concurrent transfer rejection, history immutability and unauthorized access. See [contract](backend-bed-contract.md).

### Message worker crash recovery

Migration 012 adds 45-second ownership leases and attempt counters. The worker renews leases while dispatching with a 20-second send context; stale completions cannot overwrite a recovered row. Expired processing rows become uncertain in bounded batches and are never automatically resent. Legacy processing rows become uncertain during migration. Go tests/vet and live Mailpit/captured-SMS integration passed. Local migrations 009-012 are applied, and the API and worker are running the new builds; readiness and the authentication/scheduling/messaging regression passed again. External provider credentials remain blank.

### Doctor absences and retained rescheduling

Migration 013 supports future time-range absences, cancellation reasons and versions, and immutable reschedule history. Doctor absence reasons are visible only to admin/the doctor. Leave cannot overlap active bookings; booking/rescheduling/leave share the doctor lock. Future booked appointments may move within the same doctor's schedule, preserving original booking identity and time for retries. Go unit/database/HTTP tests and vet passed, including competing leave/booking and reschedule requests. Notification updates, cross-doctor reassignment, original calendar/holiday UI and schedule override parity remain pending. [Contract](backend-scheduling-changes.md).

### Bed-type master data

Migration 014 backfills master records from existing bed labels and gives every bed a required type reference. Admin APIs create/edit/archive types with optimistic versions; beds cannot be created using archived types. Existing label-based bed creation remains compatible and resolves or creates a master record in the same transaction. Type renames update display labels; archiving a type does not remove existing beds or clinical history. PostgreSQL tests and vet passed for role denial, stable references, renamed labels, archived-type rejection and stale edits.

### Nursing access and retained observations

Migration 015 adds bounded nurse rosters with retained assignment events and typed vitals. Only assigned nurses and the encounter doctor can record; administrators can assign/read but cannot sign measurements. Removing assignment denies subsequent reads and even same-key replays. Corrections retain the earlier observation and cannot fork under competing requests. Go unit/PostgreSQL/HTTP tests and vet passed. [Detailed contract and source differences](backend-nursing-contract.md). Clinical interpretation, wider team delegation and patient release remain open.

### Diagnostic source billing

Migration 016 links released diagnostic orders to one immutable invoice. Admin/accountant billing derives the patient and immutable catalog tariff, with validated account/discount and exact integer totals. Concurrent requests return one invoice; lab staff cannot issue invoices and unreleased orders cannot be billed through this endpoint. Pharmacy and diagnostics now share the source-invoice posting helper. PostgreSQL/HTTP tests and vet passed, including retained source links, exact totals and patient invoice access.

### Optional payment-provider backend

Migration 017 and the Go Stripe adapter implement server-derived checkout and signed-success-event ledger posting. All credential fields remain blank; no live provider request or charge occurred. Tests use a loopback fake provider and locally signed payloads, cover request metadata/idempotency, mode/signature/raw-body/timestamp rejection, duplicate checkout/event/intent races, patient scope and overpayment review. Go tests/vet passed. Setup instructions now document the implemented webhook route and explicitly separate sandbox/live credentials. Full provider refunds, cancellation, settlement and operator resolution remain open. [Contract](backend-online-payments.md).

### Firebase Auth emulator verification

`npm run test:firebase` creates a demo Firebase project and loopback Auth emulator inside the disposable schema/service harness. Real emulator OTP confirmation produces tokens consumed by the existing Better Auth bridge; the resulting session is checked through Go. Tests passed for competing link requests, proof replay, takeover attempts, old linking sessions, disabled accounts and revoked Firebase tokens. No credentials or real SMS are needed. CI includes the same command and now audits development dependencies too. Firebase CLI is pinned, with scoped overrides for patched FTP/OpenTelemetry packages and a brace-free Chokidar release; the Auth-emulator use case passed after overrides. Full npm audit: zero vulnerabilities. General use of other Firebase CLI emulators is not covered by this suite.

### Complete password-recovery verification

The isolated recovery suite follows a Mailpit-delivered reset link, changes the password, verifies single-use and expired-token rejection, rejects the old password, checks revocation of two pre-existing sessions, signs in/out with the replacement password, checks the generic unknown-account response and verifies the fourth reset request is throttled. It passed locally and is added to CI. All accounts/tokens are synthetic in a disposable schema; Mailpit retains synthetic test messages. No new frontend integration was added.

### Attendance backend increment

Migration 028 adds `attendance_shift`, `attendance_shift_assignment`, `attendance_record`, `attendance_break`, `attendance_correction`, and `attendance_approval_history` with database-level immutability triggers on corrections and approval history (`protect_retained_record`). Default Day Shift (08:00-17:00) and Night Shift (20:00-06:00 overnight) are seeded.

Typed Go layers (`domain`, `application`, `adapters/postgres`, `adapters/httpapi`) implement:
- Explicit shift definitions and staff shift assignments.
- Check-in/check-out and break tracking (with automatic active break closure on check-out).
- Scheduled vs worked time, lateness, early departure, and overtime calculations strictly anchored to `Africa/Addis_Ababa` (UTC+3, EAT) without day boundaries breaking overnight shifts.
- Actor-scoped authorization: staff can only clock and view their own attendance records (`attendance.clock`, `attendance.read_own`), while admin/HR manage all staff records (`attendance.manage`).
- Reasoned corrections with immutable snapshots preserving original vs corrected fields and author identity.
- Approval workflow transitions (`draft` -> `submitted` -> `approved` / `rejected`), recording immutable approval history with approver notes and forbidding edits to approved records without super-admin override.
- Optimistic locking (`version` field) and PostgreSQL row locks preventing duplicate check-ins, overlapping active breaks, stale updates, and concurrent clocking races.
- Filtered paginated list endpoints and aggregated summary statistics (`total_scheduled_minutes`, `total_worked_minutes`, `total_break_minutes`, `total_overtime_minutes`, `total_late_minutes`, `total_early_departure_minutes`).

PostgreSQL and HTTP tests verify: all role permissions and denials, cross-staff ID isolation, concurrent check-ins and stale updates, overlapping breaks, overnight shifts crossing EAT midnight, reasoned corrections, approval state transitions, database trigger immutability, and API error codes. See contract in [backend attendance contract](backend-attendance-contract.md). Frontend integration deferred to Section 4.
 
### Ambulance and ambulance call billing increment

Migration 029 adds `ambulance`, `ambulance_call`, and `ambulance_call_invoice` with database-level immutability triggers on invoice linkage (`protect_retained_record`). Seeded default owned and contracted ambulances.

Typed Go layers (`domain`, `application`, `adapters/postgres`, `adapters/httpapi`) implement:
- Vehicle CRUD with year/license validation, owned vs contracted classification, and availability state tracking.
- Ambulance emergency call dispatching linking patient, assigned vehicle, driver snapshot, date/time, and pickup/destination details.
- Row-level database locking preventing concurrent double-dispatch of vehicles in transit.
- Automatic vehicle release (marking vehicle available) upon call completion or cancellation, and atomic vehicle re-assignment during transit.
- Direct source billing via `insertSourceInvoice` creating sealed invoices with line descriptions referencing vehicle model and registration.
- Idempotent re-billing protection and DB trigger protection against tampering with retained invoice associations.
- Role authorization matrix (`ambulance.manage`, `ambulance.read`, `ambulance_call.manage`, `ambulance_call.read`, `billing.manage`) ensuring patient records remain private to the patient.

PostgreSQL and HTTP tests verify: role permissions and denials, patient scope isolation, concurrent dispatch conflicts, status transitions, vehicle release, stale version rejections, idempotent billing, trigger immutability, and HTTP status codes. See contract in [backend ambulances contract](backend-ambulances-contract.md). Frontend integration deferred to Section 4.
 
### Services, charge categories, operations, custom fields, and module settings increment

Migration 030 adds `charge_category`, `hospital_charge`, `hospital_service`, `operation_category`, `hospital_operation`, `custom_field`, and `hospital_module_setting` tables.

Typed Go layers (`domain`, `application`, `adapters/postgres`, `adapters/httpapi`) implement:
- Charge categories across 5 typed domains (Investigation, Operation, Bed, Doctor, Other) and standard charges with integer ETB minor currency (cents).
- Billable and procedural hospital services with unit rates, quantities, and active/inactive status toggling.
- Surgical and procedural taxonomy: operation categories and operations with category associations.
- Extensible custom fields per module (patient, appointment, ipd, opd, etc.), field types (text, number, select, date, boolean, textarea), validation rules, and responsive grid layout spans (1-12).
- Dynamic module settings for enabling/disabling hospital sub-systems.
- RBAC permissions (`services.manage`, `services.read`, `operations.manage`, `operations.read`, `settings.manage`, `settings.read`).

PostgreSQL and HTTP integration tests verify: role authorization and denials, validation constraints, CRUD lifecycle, custom field module lookups, module setting activation toggles, and HTTP endpoints. See contract in [backend services operations contract](backend-services-operations-contract.md). Frontend integration deferred to Section 4.

### CMS content, hospital general settings, schedules, and testimonials increment

Migration 031 adds `hospital_general_setting`, `hospital_schedule_day`, `front_cms_setting`, and `cms_testimonial` tables.

Typed Go layers (`domain`, `application`, `adapters/postgres`, `adapters/httpapi`) implement:
- Hospital general settings covering hospital naming, contact info, branding URLs, ETB currency, language, and queue themes with bulk updates and audit logging.
- Hospital weekly schedule day configuration with start/end time validation and closure toggles.
- Front CMS settings for public portal landing sections (home, about, services, doctors, contact, map, terms, privacy).
- Testimonials lifecycle with rating (1-5), role-restricted draft moderation, public vs administrative visibility scoping, and audit events.
- Role authorization (`settings.manage`, `settings.read`, `cms.manage`, `cms.read`).

PostgreSQL and HTTP integration tests verify: role permissions and denials, start/end time boundaries, validation invariants, draft vs published visibility, CRUD updates, audit events, and HTTP endpoints. See contract in [backend CMS settings contract](backend-cms-settings-contract.md). Frontend integration deferred to Section 4.

### Front office, complaints, enquiries, notices, visitors, call logs, and postals increment

Migration 032 adds `hospital_complaint`, `hospital_notice_board`, `hospital_enquiry`, `hospital_visitor`, `hospital_call_log`, and `hospital_postal` tables.

Typed Go layers (`domain`, `application`, `adapters/postgres`, `adapters/httpapi`) implement:
- Patient grievance & complaint lifecycle (`pending`, `in_progress`, `resolved`, `rejected`) with administrative resolution notes, resolver user references, and strict patient-scoping privacy rules.
- Hospital announcement notice board with full CRUD and role-based publishing.
- Public patient enquiries and feedback intake with read-status tracking and receptionist review audits.
- Front-office visitor registry capturing visit purpose, visitor name, contact, kebele ID, headcounts, date, check-in, and check-out times.
- Telephonic call logs tracking incoming/outgoing reception interactions and scheduled follow-up dates.
- Physical postal mail log tracking incoming dispatches and outgoing consignments with reference numbers and date stamps.
- Role authorization (`front_office.manage`, `front_office.read`, `complaints.manage`, `complaints.read`, `complaints.create`, `notices.manage`, `notices.read`).

PostgreSQL and HTTP integration tests verify: role authorization and denials, patient record scoping and privacy, resolution transitions, public enquiry intake, CRUD operations across all front-office registries, audit event recording, and HTTP endpoints. See contract in [backend front office contract](backend-front-office-contract.md). Frontend integration deferred to Section 4.


### Attendance review acceptance (2026-10-05)
- [x] Reproduced and repaired keyed retry failures for clock-in/out, breaks and manual creation; concurrent identical retries return the original result.
- [x] Reproduced and repaired overlapping overnight open records and midnight lookup; database guard added.
- [x] Reproduced and repaired approval of unfinished records; approved totals survive subsequent clock-out attempts.
- [x] Uncached isolated PostgreSQL/Go suite and go vet passed after repairs.
- [ ] Attendance frontend integration and full browser journey remain Section 4 work; broader operational branch acceptance remains pending.

### Operational review: queue concurrency (2026-10-05)
- [x] Reproduced token allocation race: 11 of 16 simultaneous registrations failed on duplicate tokens before repair.
- [x] Added transaction-scoped doctor/date allocation lock; all 16 simultaneous registrations now receive unique tokens. Uncached Go/PostgreSQL tests and go vet pass.
- [x] Queue authorization, transition and appointment linkage checks repaired in `c0975a4`; new module browser integration remains Section 4.

### Operational review: private attachments (2026-10-05)
- [x] Replaced metadata-only HTTP upload with bounded multipart storage, server MIME/size/checksum calculation and authorized downloads.
- [x] Cross-patient and unassigned-clinician denial, upload/download bytes, filename traversal, HTML and size limits tested with real files and PostgreSQL.
- [ ] Malware scanning/quarantine, lifecycle retention, coordinated backup restore, upload throttling and broader file-format support remain pending; see [private attachment setup](private-attachment-storage.md).
- [ ] Reopened unsupported deployment/import/restore/acceptance completion claims: documentation alone is not execution evidence.

### Operational review: patient merge guard (2026-10-05)
- [x] Reproduced transaction abort when merging a patient with encounters; the submitted code violated composite case/patient/doctor references and ignored write errors.
- [x] Removed unsafe partial reassignment. Valid administrator merge requests return unavailable and leave clinical rows and merge events unchanged; negative regression test passes.
- [ ] Implement canonical patient identity with retained original identifiers, ownership/consent conflict policy, all-module reader resolution, concurrency/reversal rules and reconciliation tests before re-enabling merge. This guard is not completion of patient merge.

### Operational review: clinical record scope (2026-10-05)
- [x] Reproduced unrelated patient/doctor/nurse access to new diagnoses, procedures, care-team, encounter attachments, beds, follow-ups, referrals and odontogram reads.
- [x] Added current patient/encounter assignment checks across the clinical-care service, with scoped care-team revocation and patient/encounter consistency checks. Negative tests plus assigned-doctor/admin positive tests pass.
- [ ] New clinical patient-portal release, delegation semantics, transactional authorization races, financial reconciliation and complete lifecycle tests remain pending. No new clinical module is accepted solely because these scope tests pass.

### Identity consolidation implementation (2026-10-05)
- [x] Re-enabled patient merge using retained identity aliases, immutable original snapshots, conservative consent and portal-owner/active-care conflict checks (migration 043).
- [x] Updated portal ownership checks and patient history readers; added authenticated retained-identity lookup and blocked new work on retired UUIDs.
- [x] Concurrent replay, original encounter/invoice retention, portal access and alias lookup tested against isolated PostgreSQL; full Go suite and vet pass. See [merge contract](patient-identity-merge.md).

### Operational review: scoped access and report release (2026-10-05)

- Verified patient and doctor isolation for staff queues, public booking requests, appointment charges, prescriptions, consultations, extended patient profiles, diagnostic report files and clinical reports.
- Validated queue appointment/patient/doctor/date relationships and permitted queue transitions.
- Private attachments require explicit clinician/admin release. Diagnostic attachments must belong to the order's exact patient and encounter and cannot bypass signed result release through the generic attachment endpoint. Server-derived file metadata is authoritative.
- Added PostgreSQL regression cases for unassigned access, attribution spoofing, unreleased downloads, release replay, and a real lab upload → clinician review → patient download flow.
- Validation: uncached `go test -count=1 ./...` against isolated PostgreSQL schema, all migrations, and `go vet ./...` passed. Broader operational review remains in progress.

### Verified operational billing and clinical integrity (2026-10-05)

- [x] Appointment payment status derives from the invoice ledger, including refunds; invoice linking cannot fabricate payment.
- [x] Source invoice links validate actual delivered source, canonical patient ownership, server price and invoice capacity. Source rows and invoices are locked; repeat links replay safely and a different invoice conflicts.
- [x] Encounter clearance requires payment, zero charges, or an explicit assistance reason. Linked/cleared charges are frozen.
- [x] `POST /v1/patient-service-charges` records repeatable delivered service/operation events with an `Idempotency-Key`. Services snapshot catalogue tariffs; operations require an explicit approved amount. Link the returned event ID, not the catalogue ID.
- [x] Signed structured discharge summaries are retained; corrections use the existing signed discharge addendum workflow after discharge.
- [x] New bed assignments use the existing versioned transfer transaction, occupancy/state checks and immutable history. Requests must include the current encounter `version`.
- [x] Payroll excludes inactive staff and patients; money inputs are bounded against overflow. Invoice-source reads enforce invoice ownership.
- Regression evidence: real PostgreSQL tests cover wrong-patient invoices, duplicate links, payment/refund-derived status, unpaid clearance, signed-summary mutation, service snapshots/replays and patient payroll denial. Uncached Go suite and vet passed.


### Operational branch acceptance boundaries, 2026-10-05

This review corrects earlier broad completion claims. The branch provides development APIs for the submitted operational modules; this is not completion of every Section 3 requirement or a production release. The unchecked items remain deliberate and must not be silently checked because a table, endpoint, document or synthetic test exists.

Verified this pass: retained patient identity merging (043), scoped record access and reviewed attachment release (044), real-source billing and immutable clinical/financial records (045), care-team grant/revoke scope, production fail-closed malware scanning adapter, public booking/QR quotas, and authenticated database backup/restore tooling. Source prescription status is active/inactive (Laravel `Prescription::ACTIVE/INACTIVE`); it is not proof of medicine dispensing.

Partial features reopened above: public booking has request intake, bounded per-process abuse limits and staff review, but no contact verification; care-team delegation is implemented but configurable clinical alerts and observation release are not; financial income/expenses and payroll payout work but account transfers, independent payroll approval and print parity are not complete; live consultations store visit records but real meeting-provider provisioning is not implemented; smart-card metadata/verification exists but complete template/download parity remains.

The restore drill proves the tool on synthetic databases, not production recovery or attachment recovery. Malware tests use a protocol fixture, not an installed virus engine. Public quotas are per API process and trusted reverse-proxy/shared quotas remain deployment work. All real external credentials remain blank.


### Integrated checkpoint

- [x] Reviewed operational branch merged and pushed as `d618b0e`.
- [x] Hosted CI passed: https://github.com/zemuumez/HospitalManagementSystemET/actions/runs/37254122802.
- [x] Encrypted development backup taken before applying remaining migrations 029–040 and 043–045.
- [x] All migrations through 045 applied locally; rebuilt API readiness 200; post-migration reconciliation PASS.
- [ ] Complete the remaining Section 3 items before treating the entire backend as finished. Reopened full-parity items have partial APIs, not proof of print/export, versioning, clinical unit traceability, or every original workflow.
