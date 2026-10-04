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
- [ ] Successful Firebase OTP/link/sign-in end-to-end verification. Configuration/emulator success-path testing remains open.

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
| Patients | `/v1/patients`, `/v1/patients/{id}` access linking | Patients / live register | Full demographics, detail/edit, files, care team |
| Scheduling | `/v1/doctors`, `/v1/slots`, `/v1/appointments`, appointment PATCH | Schedules, Appointments, portal appointments | Calendars, queues, absence UI, fees, reminders |
| Clinical | `/v1/beds`, `/v1/cases`, `/v1/encounters` | Beds, Cases, OPD/IPD and portal lists | Original submodules and fields |
| Notes/discharge | Encounter `/notes` and `/discharge` | Encounter details | Attachments, amendments, templates, print/PDF |
| Messaging | `/v1/messages` | Communications | Preferences, callbacks, scheduled delivery, reconciliation |
| Billing | `/v1/charge-accounts`, `/v1/invoices`, invoice detail/payments, billing patient index | Accounts, Invoices, portal invoices | Gateways, taxes, voiding/corrections, print, reconciliation, source charges |
| Other original modules | No completed API yet | Marked frontend previews | See module coverage appendix |

The browser uses `/api/hms/*`; Go's `/v1/*` endpoints remain private. The complete original-screen inventory is in [backend-module-coverage.md](backend-module-coverage.md).

## 3. Full backend implementation checklist

### A. Source parity and domain contracts

- [ ] Complete authored Laravel controllers, requests, repositories, policies, models, migrations, jobs, routes and views review for every module.
- [ ] Reconcile PDF role descriptions with implemented permissions and source behavior; document contradictions.
- [ ] Revisit the live reference and exercise every role when it is accessible. Earlier attempts timed out.
- [ ] Record field types, nullability, required rules, relationships, lifecycle states, side effects, exports and error behavior per workflow.
- [ ] Define API contracts and stable error codes; publish OpenAPI and generate/check frontend types.
- [ ] Document intentional changes for a single hospital; exclude SaaS tenant/subscription mechanics unless explicitly requested.
- [ ] Identify inactive/commented legacy routes separately from active product features.

### B. Identity, authorization and security

- [x] Better Auth credentials/sessions and active database roles.
- [x] Administrator provisioning and disablement.
- [ ] Invitation acceptance with expiring single-use links, first-login password policy and identity verification.
- [ ] Staff profile edits, audited role changes and safe reassignment/deprovisioning of active clinical work.
- [ ] Real staff MFA/step-up authentication; replace or remove the security-preview demonstration before production.
- [ ] Password recovery completion, reset expiry/replay tests, session management and revoke-all UI.
- [ ] Firebase emulator success-path tests: OTP, link, login, revoked token, replay, disabled account and concurrent linking.
- [ ] Firebase production project configuration, authorized domains and OTP anti-abuse limits.
- [ ] Audited phone-change/unlink/recovery process that cannot claim someone else's account.
- [ ] Full role/action/record matrix for admin, doctor, patient, nurse, receptionist, pharmacist, accountant, case manager and lab technician.
- [ ] Apply field-level privacy, team assignment, export authorization and minimum necessary record visibility.
- [ ] CSRF/origin/proxy tests for every mutation; public endpoint abuse limits.
- [ ] Production CSP/security headers, TLS, secure cookies, secret rotation and least-privilege database roles.
- [ ] Immutable/retained audit policy, redacted application logs and privileged audit review UI.
- [ ] Independent security assessment, dependency scanning and remediation before production.

### C. Patient and staff master data

- [x] Patient creation, generated MRN, scoped list/search and explicit portal linkage.
- [ ] Full patient demographics: email, sex/gender, blood group, addresses, contact details, status and relevant original fields.
- [ ] Patient detail/edit with optimistic concurrency and audit history; no destructive clinical-record deletion.
- [ ] Duplicate detection, merge/correction policy and imported identifier preservation.
- [ ] Doctor departments, qualifications, specialties, fees, profiles, photos and availability management.
- [ ] Nurse, receptionist, accountant, pharmacist, case-manager and lab-technician profile parity.
- [ ] Emergency/guardian contacts, consent/preferences and staff/team assignments.
- [ ] Smart-card templates, issued identifiers, signed/authorized QR lookup, download and revocation.

### D. Scheduling and front desk

- [x] Core weekly schedules, slots, atomic booking and status transitions.
- [ ] Hospital opening hours, holiday exceptions and doctor absence/break CRUD.
- [ ] Safe rescheduling and cancellation reasons/history; preserve notification/payment associations.
- [ ] Appointment fees, transaction records, payment states and refunds.
- [ ] Calendar views and patient queues backed by server data; concurrent queue token allocation.
- [ ] Scheduled reminders with durable jobs, deduplication, preferences and timezone handling.
- [ ] Public appointment requests, spam protection, verification and staff confirmation workflow.
- [ ] Notices, enquiries, visitor records, postal dispatch/receive and call logs.

### E. Clinical care and bed management

- [x] Case relationships and basic OPD/IPD admission/discharge.
- [x] Atomic bed occupancy and immutable signed notes.
- [ ] Bed-type master data, maintenance/unavailable states, transfers and occupancy history.
- [ ] Nursing/team access, vitals and observations with units/ranges and timestamps.
- [ ] Consultation registers, diagnoses, procedures/operations, clinical timelines and attachments.
- [ ] Prescriptions with medicine references, dose, route, frequency, duration, instructions and signed versions.
- [ ] Encounter charges/payments and invoice linkage; discharge financial policy must be explicit.
- [ ] Admission packages/insurance/guardians and full original admission fields.
- [ ] Discharge summary templates, signing, print/PDF and post-discharge correction/addendum workflow.
- [ ] OPD repeat visits, follow-ups, referral handling and patient-visible summaries.
- [ ] Odontogram patient/tooth/procedure history and image/print/export persistence.

### F. Diagnostics and treatment records

- [ ] Pathology/radiology categories, units, parameters, reference ranges and test catalogs.
- [ ] Orders, samples/accessions, collection, processing, result review/sign-off and amendments.
- [ ] Authorized report files and patient portal report release.
- [ ] Diagnosis templates/tests/results and linkage to cases/encounters.
- [ ] Vaccination catalog, administered doses, lot/expiry records and schedules.
- [ ] Birth/death/operation/investigation reports with authorized edits and printable outputs.

### G. Pharmacy, blood bank and inventory

- [ ] Medicines, categories, brands, units, suppliers, purchases and batches/expiry.
- [ ] Append-only stock movements; transactional no-negative-stock checks and concurrent dispensing tests.
- [ ] Prescription-to-dispensation linkage; partial dispensing, returns and cancellation rules.
- [ ] Medicine bills, payments, discounts and reversals linked to the financial ledger.
- [ ] Blood groups, donors, donations, components, screening, inventory and issued units.
- [ ] Blood expiry/compatibility workflow and traceability; do not infer clinical decisions from UI labels.
- [ ] General inventory items/categories, stock receipt, issue/return, low-stock alerts and reconciliation.

### H. Billing, finance and payroll

- [ ] Charge accounts and immutable issued invoice/line snapshots.
- [x] Core invoice integer totals, server-calculated percentage discounts and documented rounding.
- [ ] Tax rules, tax calculation, tax reports and wider workflow pricing.
- [ ] Bills/invoices, itemized services, quantities, printable receipts and original print-template parity.
- [x] Core invoice partial/full payments, idempotent posting, overpayment prevention and append-only refunds.
- [ ] Expanded reversal/voiding approvals and reconciliation workflows.
- [ ] Payment provider integrations, signed webhooks, event replay protection and settlement reconciliation.
- [ ] IPD/OPD/pharmacy/ambulance/lab charges linked to their source workflows without duplicate billing.
- [ ] Expenses, income, account transfers and daily/monthly financial reports.
- [ ] Insurance, packages, policy details, claims and patient responsibility.
- [ ] Employee payroll, allowances/deductions, approval, payout and payroll slips.
- [ ] Currency configuration/migration rules; current new financial contracts use ETB only.
- [ ] Tax/compliance requirements must be established before asserting accounting/legal compliance.

### I. Operational modules, content and communications

- [ ] Attendance/check-in/out, shifts, breaks, overtime, corrections and approval history.
- [ ] Ambulances, assignment/calls, tariffs and billing.
- [ ] Services, charge categories, operations, custom fields and validated module settings.
- [ ] CMS home/about/services/doctors/testimonials/contact/terms/map content persisted and published safely.
- [ ] Hospital general settings, logo/favicon, schedules, language and queue theme persisted.
- [ ] Complaints, notices, testimonials moderation and front-office enquiry lifecycle.
- [ ] Live consultations/meetings, provider tokens/permissions and visit linkage.
- [x] Operational email/SMS outbox and safe development transports.
- [ ] Delivery callbacks, preferences/consent, templates, localization, scheduled jobs and retry/reconciliation UI.
- [ ] Production SMTP/SMS setup and controlled delivery tests using approved recipients.

### J. Files, data, operations and release

- [ ] Private attachment storage, size/type limits, malware checks, authorized downloads and retention.
- [ ] No public patient-upload buckets or predictable unauthenticated file URLs.
- [ ] Migration/import mapping from Laravel, trial runs, reconciliation reports and rollback plan.
- [ ] Pagination/filtering/search/export for every list; avoid unbounded option loaders.
- [ ] Background workers with leases/heartbeats, crash recovery, retries and observability.
- [ ] Structured redacted logs, request IDs, metrics, health/readiness checks and alerting.
- [ ] Database indexes/query plans, connection limits, load tests and capacity targets.
- [ ] Automated backups, encryption, retention and a demonstrated restore drill.
- [ ] CI gates for migrations, unit/integration/browser tests, build and dependencies.
- [ ] Staging deployment, environment separation, secret management and production runbooks.
- [ ] User acceptance sign-off, migration cutover, rollback rehearsal and post-release monitoring.

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
- [ ] Firebase success-path/emulator tests and recovery/phone-binding races.
- [ ] Password reset completion, expiry, reuse, session revocation and abuse throttling.
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
4. Connect the UI while preserving the original design and label unfinished behavior honestly.
5. Run meaningful unit/database/API/browser checks and inspect the result.
6. Update this checklist, the module inventory and known limitations with actual evidence.
7. Commit the completed step and push to `origin/main`; never include local secrets, database files or legacy archives.

Next implementation: prescriptions, pharmacy/inventory movements and their billing linkage, followed by diagnostics and the remaining domain areas above. No new approval is needed for the already authorized development work.
