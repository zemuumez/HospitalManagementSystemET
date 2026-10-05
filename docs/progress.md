# Current backend checkpoint

The large operational branch is reviewed, merged and pushed to `main` in `d618b0e`. Development migrations 001–045 are applied, the rebuilt `api-operational.exe` is running on port 8080, readiness returned 200, and post-migration database reconciliation passed. An encrypted pre-migration development backup is retained under ignored `.local/pre-operational-backup`; its local restore key is not committed. Private development attachments use ignored `.local/private-attachments`.

Section 3 is still incomplete; Section 4 new frontend integration has not started. The corrected [delivery checklist](backend-delivery-checklist.md) is authoritative. Integrating the submitted branch does not turn partial fields, API scaffolding, print previews or runbooks into completed original workflows.

## Integrated scope

| Area | Verified working scope | Remaining scope |
|---|---|---|
| Identity | Better Auth sessions, active roles, invitations, recovery, Firebase phone/MFA emulator flows | Production credentials, phone recovery/change, privileged step-up and new UI |
| Patients and clinical records | Retained identity merges, portal ownership/consent conflict checks, original history preservation, scoped reads, care-team grant/revoke, guarded bed transfers and signed summaries | Complete source form/print parity, configurable alerts, remaining clinical release policies |
| Scheduling/front desk | Atomic booking, queue tokens and transitions, appointment attribution, public request review, front-office records | Contact verification, durable reminders and new module frontend integration |
| Finance | Immutable invoices/payments/refunds, verified delivered-service links, payment-derived appointment status, guarded clearance, expenses/income and payroll payout | Tax/insurance/claim workflows, account transfers, independent payroll approval, provider settlements and print parity |
| Pharmacy/blood/inventory | Existing stock ledger and dispensing; catalogue masters, structured prescriptions, aggregate blood donations/issues | Full prescription-to-dispensing integration, blood unit screening/expiry/compatibility/traceability and remaining source parity |
| Files and operations | Private stored uploads/downloads, clinician release, diagnostic result release checks, production scan gate, request quotas, reconciliation and authenticated backup/restore tools | Real scanner acceptance, coordinated file/database recovery, retention/import/capacity/staging and independent security/UAT |
| Other operational modules | Attendance, ambulance calls/billing, CMS/settings, services/operations, consultations/meetings and report records | Real meeting-provider provisioning, full business workflow/role/field parity and frontend integration |

## Verification

The merged code passed uncached Go/PostgreSQL transaction, concurrency, authorization and HTTP tests through migration 045, `go vet`, web unit tests, formatting, build and TypeScript checks. Connected API and browser suites passed, as did Firebase/MFA, recovery and invitations. The synthetic backup/restore drill passed exact data reconciliation, wrong-key/tampering rejection and nonempty-target protection. The production npm dependency audit found zero vulnerabilities.

[Hosted CI for merge d618b0e passed](https://github.com/zemuumez/HospitalManagementSystemET/actions/runs/37254122802). See [the review evidence](backend-review-checkpoint-2026-10-05.md) for individual commits and limitations.

Real Firebase, SMS, payment and production email credentials remain blank in [deployment/.env.example](../deployment/.env.example); setup instructions are in [provider-setup.md](provider-setup.md). No real SMS, external email, payment charge or production deployment was performed.

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
