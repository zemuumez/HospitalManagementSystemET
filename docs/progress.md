# Current backend checkpoint

Section 3 remains in progress; Section 4 new frontend integration has not started. All verified increments through `74b76fd` are committed and pushed. Local migrations 001–037 are applied and the local Go API has been rebuilt/restarted with those changes; readiness returned 200. Real Firebase, SMS and payment credentials remain blank in the production worksheet.

## Implemented and verified backend work

| Area | Working backend scope | Main remaining scope |
|---|---|---|
| Identity | Better Auth sessions, admin provisioning/disablement, phone linking/login, real Firebase Auth-emulator tests, password recovery, authenticator MFA, invitations, and comprehensive 9-role action/record matrix (`docs/role-permission-matrix.md`) | Phone-change/recovery administration, privileged-action step-up, production setup, new auth UI |
| Staff and audit | Shared demographic/qualification profiles, doctor departments with versioned revisions, doctor profile extensions (fees, bio, photo, department link), retained revisions, guarded role changes, session revocation, immutable audit rows and admin review API | Full role-specific fields/files, care reassignment, history/review screens, production retention/archival |
| Patients | MRN registration, explicit portal links, scoped search/detail, versioned demographics/contact/consent, additional source fields (father name, religion, referral source, notes), guardian contact & consent preferences, duplicate detection candidate search, atomic merge with immutable audit trail (`patient_merge_event`), smart cards with cryptographic QR tokens and public verification | Custom photos and complete legacy form parity |
| Scheduling | Weekly schedules, slots, atomic booking, doctor absences, same-doctor rescheduling, cancellation/status history, hospital opening hours and holiday/exception date overrides, patient queues with daily sequential token allocation, public appointment requests and review workflow, and appointment billing/invoice links | Scheduled reminders with durable jobs |
| Clinical | Cases, IPD/OPD, bed types/states/transfers/history, bed assignment history, bed occupancy reports, nursing assignment/vitals/corrections, observation release, care-team delegation, consultation diagnoses, procedures, attachments, IPD packages/insurance/guardians, encounter billing & financial clearance, anti-double-billing invoice links, structured prescriptions, discharge summaries, OPD follow-ups, referrals, odontogram, signed notes/discharge and addenda (migrations 003, 011, 014, 015, 021, 034, 038) | Clinical range alert engine, PDF document rendering |
| Diagnostics | Versioned/retired definitions, typed parameters, scoped orders, sample rejection/recollection, processing/results, sign/release/amendments, source billing, category/unit masters, report file uploads, patient portal report release, diagnosis templates, vaccine catalog, patient vaccinations, birth reports, death reports, operation reports, and investigation reports (migrations 009, 024, 026, 039) | Specialty verification rules, external lab HL7/FHIR sync |
| Pharmacy/inventory | Batch stock, immutable movements, signed medication orders, bounded dispensing/returns/disposal, source invoices, categories, brands, blood bank inventory across 8 groups, donors, donations, blood issues and structured prescriptions; general inventory ledger | Full bills/credits, recalls, stock approvals/notifications |
| Finance | Immutable invoice snapshots, exact ETB totals, manual payments/refunds, pharmacy/diagnostic links, optional verified Stripe checkout, conflict review queue, expense heads/expenses ledger, income heads/incomes ledger, employee payroll with staff slips, anti-double-billing service/ambulance/blood/appointment invoice links (`service_invoice_link`), and period financial summary reports (migrations 005, 017, 035, 037; contracts) | Taxes, voids, reconciliation resolution, provider refunds/settlement, insurance |
| Communications | Durable outbox, Mailpit/captured SMS, optional Twilio, leases/heartbeats, uncertain-crash recovery and stale-worker protection | Callbacks, consent enforcement across workflows, scheduling, retries/reconciliation and live provider tests |
| Operations | Request IDs/redacted logs, readiness, transactional migrations, attendance, ambulances and calls billing, services, charges, operations, custom fields, module settings, CMS and hospital general settings, schedules, testimonials, complaints, notices, enquiries, visitors, call logs, postals, live consultations, meetings, provider settings, hosted CI, secure attachments with cryptographic access tokens and SHA-256 integrity (migration 040), Prometheus `/metrics` exposition, automated AES-256 encrypted backup/restore scripts (`backup.ps1`, `backup.sh`, `restore.ps1`, `restore.sh`), legacy Laravel migration mapping and automated reconciliation verification (`scripts/reconcile_import.go`), production Caddy/Docker Compose stack, and operational runbooks (`backup-restore-runbook.md`, `database-capacity-and-query-plans.md`, `observability-and-alerting.md`, `production-deployment-runbook.md`, `cutover-and-release-playbook.md`) | Section 4 frontend integration |

The detailed acceptance list is [Section 3 of the delivery checklist](backend-delivery-checklist.md). [Module coverage](backend-module-coverage.md) retains every original screen as open until its full fields, business workflow and integration are verified. Core support in the table above is not a claim of complete original-module parity.

## Verification evidence

Go unit/database/HTTP/concurrency tests and vet pass against disposable schemas through migration 035. The web tests, formatting, typecheck and production build passed after adding invitations and operational modules. Real Firebase Auth-emulator tests include MFA enforcement; Mailpit suites cover reset and invitation completion. Existing connected browser journeys run in a separate copied app/database schema. The older API suite was moved into that same isolation model after audit retention correctly blocked its old row-deletion cleanup.

Hosted CI passed commit `ee7ffed` ([run 37231554242](https://github.com/zemuumez/HospitalManagementSystemET/actions/runs/37231554242)), scheduling commit `f561304` ([run 37231893895](https://github.com/zemuumez/HospitalManagementSystemET/actions/runs/37231893895)) and invitation commit `1ddde6d` ([run 37232198746](https://github.com/zemuumez/HospitalManagementSystemET/actions/runs/37232198746)). Later runs must be checked separately; a pushed commit is not automatically a passing CI result.

Provider setup steps are in [provider-setup.md](provider-setup.md); [deployment/.env.example](../deployment/.env.example) intentionally contains blank real credential fields. No real SMS, external email, payment charge or production deployment was performed.

---

The following sections retain earlier frontend/foundation checkpoints as historical context; their pending lists are superseded by the current checklist above.

# Latest frontend review

The four source-based frontend checkpoints are documented in [the review report](frontend-four-step-review.md). Role menus/personal records, doctor details, dependent clinical forms, source-field Amharic coverage and frontend regression checks have been added. The live demo remains unreachable, and exact PDF/clinical parity is not claimed. Backend integration has not been extended in this pass.

# Delivery status

## Screenshot-driven parity pass

Public landing/pages, English/Amharic switching, explicit original navigation order, separate IPD/OPD entries, attendance previews, CMS sub-tabs, expanded settings and smart-card QR/PNG previews are added. See the [independent parity audit](frontend-parity-audit-2026-10-04.md) for evidence, tests and remaining gaps. The public reference demo timed out, so deployed role-by-role verification is still open.

## Frontend-first direction

The original-style frontend preview now supersedes the initial teal UI. It includes 111 module routes, source-derived forms, module/role navigation and specialized clinical, calendar, bed, billing and dental views. These use marked sample data. Existing live patients remain at `/live-patients`. See [preview scope, verification and outstanding parity work](frontend-preview.md). This is not a claim that every legacy feature is complete.

## Working foundation

- Next.js/TypeScript/Tailwind responsive login, dashboard, patient directory, registration dialog, communications and account screens.
- Better Auth database sessions, email/password login, logout and password reset via SMTP. Registration is administrator-provisioned; no public staff signup.
- Firebase phone identity bridge: recent verified phone proof, explicit authenticated linking, UID binding, replay checks and normal Better Auth session issuance. Real Firebase project credentials are not configured. Successful Firebase OTP/link/sign-in still needs an emulator or configured-project end-to-end check; invalid proof and local policy tests pass.
- Go domain/application/adapter boundaries, private HTTP API, live session introspection, active-role lookup and scoped queries. Nine legacy role identifiers are represented; only the permissions for this slice are implemented.
- PostgreSQL patient registration and scoped/paginated/searchable directory. Patient/doctor scoping is enforced for both rows and counts. Registration/read audit events are stored.
- Durable communication outbox with atomic claims, request deduplication, captured development SMS, SMTP delivery and configurable Twilio adapter. No real SMS was sent.
- Transactional migrations, initial-admin command, local environment examples, optional PostgreSQL/Mailpit Compose stack, Firebase emulator configuration and setup instructions.

## Verification completed

- TypeScript type check and optimized Next.js production build.
- Firebase proof policy tests: rejects wrong provider, old/future authentication and malformed phone.
- Go tests and `go vet`: validation, deny-by-default role policy, forged identity header denial, mutation origin denial, expired/mismatched session denial.
- Live PostgreSQL/Better Auth/Go integration checks: valid sign-in, anonymous denial, CSRF denial, patient create denial, invalid DOB rejection, patient and assigned-doctor row/count isolation, immediate disabled-account denial, invalid Firebase proof denial, simultaneous message deduplication, operational email in Mailpit, password-reset email in Mailpit, captured SMS and revoked session after logout. Synthetic fixture users/patients/outbox rows are removed after the run.
- Headless Chrome: actual email login, dashboard, registration dialog and Escape, mobile navigation, no horizontal overflow at 390px, and no page errors. Desktop/mobile screenshots were visually inspected.
- npm dependency audit: zero reported vulnerabilities after patched gRPC/UUID overrides. This is not an independent security audit.

## Remaining delivery

This is the first working vertical slice, not a full HMS replacement. Remaining modules include staff administration/invitations, permission editor, patient detail/edit/care-team/portal linking, appointments/queue, OPD/IPD/beds, diagnostics/reports, pharmacy/blood/inventory, billing/payroll/payment providers, consultations, CMS, localization, data migration and production operations/security hardening. Firebase success-path integration, SMS delivery callbacks, scheduled reminders and provider reconciliation/retry tools remain open.

The exhaustive legacy review remains tracked in the discovery ledger. Complete each affected workflow audit while building its corresponding slice; do not infer full parity from this foundation.

## Local review

Open the running application at http://127.0.0.1:3000 and Mailpit at http://127.0.0.1:8025. The local admin's generated credentials are in ignored `.local/admin-login.txt`. Secrets and the legacy extraction are excluded from version control. No deployment, system service installation or live SMS connection was made.
