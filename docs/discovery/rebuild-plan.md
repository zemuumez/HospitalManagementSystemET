# Proposed architecture and staged rebuild

Update: single-hospital scope and building are now authorized. Better Auth owns sessions, Firebase supplies phone identity proof, and Go owns authorization. See [the accepted authentication decision](../architecture/001-authentication.md) and [current delivery status](../progress.md). The original proposal below is retained as discovery history; its multi-hospital/session proposals and confirmation gate no longer govern the build.

Status: design proposal awaiting the user's confirmation. No Next.js/Go application or infrastructure has been implemented.

## Architecture

Use a modular Go backend with explicit domain boundaries and an independently deployable worker. Start with a modular monolith so appointments, occupancy, stock and financial transitions can use straightforward database transactions. Split services only where measured scale or ownership justifies the operational complexity.

```text
Next.js + TypeScript + Tailwind
  -> same-origin /api routing
  -> Go HTTP adapters -> application use cases -> domain rules
                              -> repository/provider interfaces
                              -> PostgreSQL / private object storage
                              -> transactional outbox -> workers
                              -> Firebase / SMTP / SMS / payment adapters
```

Go owns business rules and authorization. Next.js owns rendering and interaction; server-rendered pages and actions use the same authorized Go interfaces. Never implement a second conflicting set of billing or access rules in the frontend. Shared OpenAPI contracts generate the typed client; database entities are not the public API.

Suggested repository structure:

```text
apps/web/                         Next.js app, feature UI, Tailwind design tokens
services/api/cmd/api/             API entry point
services/api/cmd/worker/          scheduled and queued work
services/api/internal/<domain>/domain/
services/api/internal/<domain>/application/
services/api/internal/<domain>/adapters/http/
services/api/internal/<domain>/adapters/postgres/
services/api/internal/platform/   configuration, auth, telemetry, transactions
contracts/openapi/               API specification and generation configuration
db/migrations/                   forward schema changes and rollback guidance
infra/development/               local services and emulator configuration
docs/                            decisions, traceability, runbooks and acceptance
```

Domain code imports no HTTP framework, Firebase SDK or database implementation. Use cases depend on narrow interfaces for repositories, transaction boundaries, identity verification, mail, SMS, payment, clock and ID generation. Keep abstractions driven by actual workflows; do not replace every simple function with an interface.

Proposed domains: identity/access; hospitals/configuration; staff; patients/cases; scheduling; clinical encounters; beds; diagnostics/reports; pharmacy; blood bank; inventory; billing/payments/payroll; communications/consultations; public CMS/front office. Platform subscriptions are a separate domain if independent-hospital SaaS is confirmed.

PostgreSQL is the proposed system of record. Use exact decimal amounts or integer minor units with currency metadata, explicit foreign keys, constrained state transitions and optimistic versions where useful. Document/file bytes belong in private object storage with metadata in SQL. Redis is optional for distributed limits/cache; it is not the authoritative bed, stock or payment ledger.

## Hospital isolation

If SaaS is confirmed, model platform operator, hospital, branch, user identity and hospital membership separately. One verified Firebase UID may map to several memberships; the active hospital must be authorized server-side. Every scoped query, job, file key, export, cache key and audit event carries validated hospital context. Never trust an arbitrary request header or client-selected hospital ID.

Use tenant-aware foreign/unique keys and a transaction-scoped database isolation strategy, potentially PostgreSQL row-level security with a restricted application role. Prove isolation with hospital A/B tests and pooled-connection tests. Platform support access should be explicit, scoped and audited. Hospital admins do not automatically read other hospitals. Clinical records do not become globally shared merely because a patient uses one identity.

## Firebase authentication and phone features

Proposed browser flow: Firebase sign-in -> short-lived ID token -> Go verifies identity and recent authentication -> CSRF-protected session exchange -> Secure, HttpOnly, SameSite session cookie -> Go verifies session and active local membership on subsequent requests. Bound session duration, revocation, account disablement, logout and sensitive-action reauthentication are part of the design. Choose a consistent same-origin deployment so cookies and CSRF controls are straightforward.

Keep role assignments, clinical relationships and hospital membership authoritative in the application database. Firebase establishes identity; it does not enforce Go/PostgreSQL row authorization. Minimize identity data sent to Firebase. Do not place medical information in custom claims, tokens or SMS. Treat Firebase App Check as an additional abuse signal if adopted, not a replacement for identity/authorization.

Phone sign-in/verification and hospital SMS are distinct interfaces:

- Firebase phone authentication handles OTP verification, web reCAPTCHA, allowed domains, account linking and recovery. Never merge patient identities merely because a client submits a matching phone/email. A verified phone number alone should not grant staff privileges.
- Operational SMS handles appointment reminders and approved messages via a replaceable provider adapter. Twilio matches the legacy implementation, but provider selection remains open. Include normalized numbers, delivery status, retries, idempotency, rate/cost limits and notification preferences.
- Development uses Firebase Auth Emulator and test phone numbers/codes, plus a fake operational SMS provider. Production must refuse emulator settings or fake auth modes. Test revocation/disabled identities and expired/wrong-project tokens.

Firebase's official documentation covers [ID-token verification](https://firebase.google.com/docs/auth/admin/verify-id-tokens), [session cookies](https://firebase.google.com/docs/auth/admin/manage-cookies), [web phone authentication](https://firebase.google.com/docs/auth/web/phone-auth), and [Auth Emulator behavior](https://firebase.google.com/docs/emulator-suite/connect_auth). In particular, emulator configuration accepts unsigned emulator credentials; that configuration must not reach production.

## Development email with Mailpit

After confirmation, provide a local Compose profile with PostgreSQL, Mailpit and any chosen emulator/worker services. In containers the Go mail adapter uses `mailpit:1025`; host processes use `127.0.0.1:1025`. Bind the Mailpit web UI to `127.0.0.1:8025`, keep it development-only and leave external relay disabled. Production uses a separate configured mail transport.

Mailpit captures application SMTP mail, not SMS. Firebase's hosted default email delivery does not automatically pass through Mailpit. For email verification/password reset flows owned by this application, generate supported action links server-side and send the templates through the mail adapter, or exercise the emulator's local action flow. Verify the chosen flow end to end before describing all authentication mail as captured.

Ports and SMTP behavior are documented by [Mailpit configuration](https://mailpit.axllent.org/docs/configuration/) and [SMTP documentation](https://mailpit.axllent.org/docs/configuration/smtp/). Test recipient, subject, action link, rendering and absence of real delivery.

## Security and reliability acceptance

Translate the requested high security standard into evidence:

- Deny-by-default permissions combining action, hospital, role and record relationship; test every API category with unauthorized users. Separate read, create, update, sign, void, approve, export and manage-settings rights.
- Authenticated private file access, content/size validation, malware scanning/quarantine where deployed, short-lived access URLs and no public patient QR disclosure.
- CSRF protection for cookie-authenticated mutations; strict origin/CORS policy; parameterized SQL; input limits; output encoding; content security policy; no state-changing GET endpoints.
- Audit sensitive reads, writes, approvals, exports, permission changes and emergency access. Redact medical content/tokens from ordinary logs; define audit retention and restricted access.
- Secrets supplied at deployment, dependency locking/scanning, least-privilege service credentials, encrypted transport/storage/backups, tested restore and incident procedures.
- Database constraints and concurrent tests for appointment conflicts, active bed occupancy and nonnegative stock. Idempotent payment webhooks, verified signatures, amount/currency/reference reconciliation and immutable posted financial records with adjustments.
- Patient-safe clinical amendments, clinician sign-off and provenance. Medicine suggestions remain an explicitly reviewed assistive feature with controlled data disclosure; they cannot silently finalize prescriptions.
- Bounded pagination, indexed access paths, request deadlines/body limits, connection pools, outbox-based durable jobs, retry/backoff/dead-letter handling and graceful shutdown. Do not promise capacity without a workload and measured results.

Security is not deferred to the final phase. Each delivery part has negative authorization and integrity tests. Final release also needs a threat-model review, end-to-end workflow acceptance, performance and recovery exercises. No compliance certification or absolute security claim is implied by choosing these tools.

Next.js authorization guidance reinforces checking access at the data boundary: [authentication guide](https://nextjs.org/docs/app/guides/authentication). Tailwind's [Next.js setup guide](https://tailwindcss.com/docs/installation/framework-guides/nextjs) will be used when scaffolding. Pin compatible supported versions during implementation rather than copying old package versions from the archive.

## Delivery parts and exit criteria

| Part | Work | Reviewable exit condition |
|---|---|---|
| 0A: Discovery baseline | PDF review, archive/source inventory, initial workflow/security findings, scope questions | This packet; user confirms or corrects understanding |
| 0B: Detailed parity audit | Complete authored-source review by domain: routes -> requests -> controllers -> repositories/helpers/models -> schema -> UI/JS -> jobs/exports/files; classify bundled dependencies/generated assets separately; establish runnable legacy baseline if feasible | Completed coverage ledger, permission/action matrix, schema/field map, workflow/state diagrams, screen inventory and discrepancy decisions; no claim of full review while entries remain pending |
| 1: Platform foundation | Next.js/Tailwind shell, Go layering, PostgreSQL migrations, configuration/CI, local services, Firebase identity/session adapter, development mail/SMS, hospital isolation if confirmed | Reproducible local startup; Firebase emulator sign-in; protected page/API; rejected unauthorized/cross-hospital requests; captured Mailpit email; no real SMS |
| 2: Staff, patients and design system | Role-aware navigation, staff/memberships, patient records, cases, documents, smart cards, audit and reusable UI | Admin creates staff/patient; role restrictions proven; private document/card access; accessible reusable screens |
| 3: Scheduling and front desk | Hospital/doctor schedules, holidays/breaks, appointments, queue/check-in, reminders, public booking | Race-safe booking; correct local-time display; reschedule/cancel/refund decisions; reminder retry does not duplicate delivery |
| 4: Clinical care | OPD/revisits, IPD, beds, diagnoses, instructions, operations, prescriptions, visibility and discharge | Complete patient encounter flow; parallel bed assignment rejected; patient sees only released records; history retained |
| 5: Diagnostics and clinical reports | Pathology, radiology, diagnosis templates, birth/death/investigation/operation reports, vaccinations, dental parity decision | Authorized result creation/release/amendment and correct PDF output |
| 6: Pharmacy, blood and inventory | Catalogs, purchases, stock movements, dispensing, medicine bills, donations/issues, inventory returns | No negative stock or duplicated movements under concurrency/retries; traceable usage and reversals |
| 7: Finance and services | Charges/packages/insurance, bills/invoices/advances, approvals, payment adapters/reconciliation, payroll, income/expenses, ambulances | Server-computed exact totals; duplicate webhook safe; ledger reconciles; hospital-approved calculation examples pass |
| 8: Communications, portals and CMS | Remaining role dashboards, consultations/calendar, notices/SMS/mail/templates, enquiries/front office, CMS, localization/RTL, complaints | Role-specific acceptance flows and configured integration sandbox tests; public content contains no private data |
| 9: Migration and release | Legacy ID/data/attachment import, identity migration, dry runs, security/performance/recovery checks, observability, release/rollback | Signed parity checklist, reconciled migration totals, successful restore, operational runbooks and explicit launch decision |

Dependencies cross parts: minimal billing contracts and encounter charges are designed in Part 0B and supported during clinical work; full finance interfaces arrive in Part 7. Patient views and localization are built alongside their modules, then completed in Part 8. Each part produces a working vertical slice and a demo/checklist before proceeding to the next agreed part.

Migration must preserve legacy identifiers, ownership, dates/time zones, monetary precision and files. Do not assume existing Laravel passwords can be imported into Firebase without verifying hash/import support; plan account linking or a controlled reset. Reconcile financial/stock/bed data before cutover. Avoid dual writes without a defined reconciliation strategy.

## Next confirmation

Confirm the product understanding and whether Part 0B is the next agreed step. Detailed audit remains necessary to satisfy the requested every-line review; beginning scaffolding is a separate decision after that baseline and its discrepancies are understood.
