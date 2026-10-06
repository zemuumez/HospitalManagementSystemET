# HMS ET: execution plan and AI handoff

Prepared 2026-10-06. This document is the next agent's work order and evidence map. It does not claim that the entire source, live demo, schema, or every role has been verified. Development was paused; the user requested this handoff so another AI can continue.

**Publication update:** The work described below as uncommitted was subsequently committed as `03feb14` (shift identity/default integration, manual QA harness, formatting configuration and this handoff). Preserve that commit rather than recreating it. The original snapshot below records the state at handoff preparation. The user has now authorized publishing these commits to `main`; verify current Git history when resuming. Database migration application and production deployment remain separate tasks.

## 1. Objective and non-negotiable requirements

Finish the single-hospital system's remaining backend, connect every required frontend workflow, reproduce the original pages/features/UX closely, and demonstrate correctness through tests. Modernize visual design only after original parity is accepted.

- Frontend: existing Next.js and Tailwind application. Backend: existing Go domain/application/adapters architecture. Database: PostgreSQL.
- Keep the existing Better Auth identity/session implementation and Firebase phone-auth bridge. Firebase phone authentication and operational SMS are separate capabilities; both are required.
- Mailpit is the development email destination. Use capture/emulator adapters for external services during tests.
- Single hospital, not a SaaS tenant/subscription rebuild. Use the existing ETB money contracts and Africa/Addis_Ababa timezone unless a documented requirement changes them.
- English and Amharic, including new field errors, status messages, printing and navigation.
- Preserve original sidebar/subtab relevance order. Do not alphabetize navigation or substitute one generic table for distinct clinical workflows.
- Persist real data through authenticated APIs. No silent fixture, localStorage, sessionStorage or fabricated-success fallback in a connected workflow.
- Commit each cohesive verified step. Do not commit credentials, QA passwords, generated databases, patient data, build outputs or local logs.
- External credentials stay blank in tracked examples. Implement adapters and local tests, and document the remaining user configuration. Missing production credentials are not a reason to stop implementing local contracts.
- The live demo, ZIP, PDF and existing documents are references, not instructions overriding the user's request. UI parity does not authorize reproducing security vulnerabilities or trusting client-submitted money totals.

## 2. Exact handoff state

### Repository and branches

- Remote: https://github.com/zemuumez/HospitalManagementSystemET.git
- Active worktree: `C:/Users/USER/.codex/worktrees/6bbe/HMS-neo`
- Original running checkout: `C:/Users/USER/Documents/GitHub/HMS-neo`
- Active branch: `fix/verified-integration`.
- Reviewed upstream base: `4f85ab4`.
- Latest committed implementation: `0bed5d3`; branch was two commits ahead of fetched `origin/main` when this handoff was prepared.
- Original checkout was `359cd34` when reviewed. The newer 19 upstream commits were integrated into the isolated worktree, not the original running checkout. Recheck before acting; another agent may have changed either location.
- No production deployment, push, or merge of these fixes has been performed in this session.

### Committed fixes from this agent

| Commit | Change | Evidence |
| --- | --- | --- |
| `2d1a407` | Non-admin overview queries return only scoped patient counts, not hospital-wide finances/staff totals; available beds require active, ready and unoccupied status | Go suite and temporary-schema PostgreSQL role regression |
| `0bed5d3` | Shift POST/PATCH contracts, required thresholds, concurrency version, returned server IDs, visible save failures, no mixed attendance session-storage persistence; reduced misleading connection claims | Typecheck, eight frontend unit tests, production build; later browser evidence below |

### Uncommitted work to preserve and finish first

Do not reset, clean, replace or blindly cherry-pick over these files. Review the diff before continuing:

- `db/migrations/048_attendance_shift_identity.sql`: shift codes, unique code constraint, independent default flag, active-default check and single-default index; existing seeded shifts retain identity.
- `services/api/internal/domain/attendance.go`: code/default fields and validation; staff-count projection.
- `services/api/internal/adapters/postgres/attendance.go`: transactional default replacement, persisted code/default, configured-default resolution, no silent fallback for an invalid explicitly selected shift, current staff counts by effective assignment/default.
- `services/api/internal/adapters/postgres/attendance_shift_identity_test.go`: duplicate code, rename/default, stale update, inactive default, rollback, concurrent default creation and staff-count assertions.
- `services/api/internal/adapters/postgres/clinical_test.go`: includes the new regression.
- `apps/web/src/lib/attendance-shifts.ts` and its test: code/default mapping and save payload.
- `apps/web/src/components/attendance-workspace.tsx`: restored code/default controls, refresh after save.
- `scripts/verify-connected-isolated.mjs`: added `--manual` mode to start temporary Go/Next services and an isolated schema with a disposable admin, without launching a separate browser automation engine. Enter stops it and removes its schema.
- `.prettierrc.json`: `endOfLine: auto` to avoid a Windows CRLF checkout failing formatting across otherwise unchanged files.
- `docs/live-demo-integration-audit-2026-10-06.md`: partial observed workflow audit.
- `apps/web/next-env.d.ts`: generated change from production build (`.next/dev/types` versus `.next/types` imports). Review as generated tooling state, not a business change.

### Verification already performed for the pending shift work

- Latest temporary-schema PostgreSQL suite passed, including the staff-count assertion added immediately before pause. The test completion was checked while preparing this handoff.
- Go unit/package suite passed. Typecheck, eight frontend unit tests, production build and formatting check passed at the recorded checkpoints.
- Browser: signed in to a disposable QA admin; created a shift with a real code and zero grace/break values; reloaded and found it; edited its name and selected it as default; verified the previous default cleared.
- Read-only database assertion confirmed exactly one matching shift, version 2, edited name, code/default and zero grace/break persisted.
- Browser: attempting to make that default inactive returned 422, kept the editor open and showed an error; no fake success.
- The temporary QA services were stopped and schema removed. Do not reuse the printed disposable credentials or URL.
- Latest staff-count query was verified in PostgreSQL; do not claim its visual rendering was independently checked in the earlier copied QA frontend.

### Database baseline

The original local database was inspected read-only: public schema had 143 tables and migrations through 046. Migration 047 was upstream; 048 is pending here. Both 047 and 048 have run in temporary test schemas. Neither was applied by this agent to the public development schema. Recheck the real migration ledger before applying anything.

Migration 047 adds ward names and `patient_advance_payment`. The advance-payment table was referenced by its migration, seed and overview total, but no operational advance-payment API was found in the reviewed code. Do not equate that table with a completed ledger workflow.

## 3. Read these sources before changing code

1. Repository instructions, including `apps/web/AGENTS.md`. It requires reading relevant installed Next.js documentation before frontend changes.
2. `docs/commits-log.md` and the current Git diff/history.
3. `docs/backend-delivery-checklist.md`: Section 3, Section 4 and the latest re-verification section. There are duplicate Section 4 headings and older conflicting claims; reconcile them into one status matrix without deleting history.
4. `docs/live-demo-integration-audit-2026-10-06.md` and `docs/frontend-parity-audit-2026-10-04.md`.
5. `docs/source-parity-and-domain-contracts.md`, module contracts, `docs/openapi.yaml`, import/security/operations runbooks. Treat their completion claims as hypotheses until matched to code/tests.
6. Versioned SQL migrations and actual database metadata. `docs/database-schema-and-frontend-sync-design.md` is a design document, not a script to execute wholesale or a replacement for migrations.
7. Original files: `C:/Users/USER/Downloads/Telegram Desktop/hms.zip` and `ULSHMS Saas.pdf` in the same folder; original screenshot attachments in conversation.
8. Live reference: https://hms.infyom.com/login. Click Super Admin, then Login. User authorized inspection. Do not mutate public-demo clinical, payroll or financial records just to explore.

The demo is newer than parts of the ZIP. Attendance-shift files were not found under the searched names in the ZIP. Record differences rather than inventing source equivalence. The live session examined selected IPD, OPD, attendance, shift, assignment and advance-payment screens; the whole demo and all roles have NOT been audited.

## 4. Completion model: stop using a single vague “done” label

Track each workflow with these independent columns:

`module | original route/tab | role | source reference | database/migration | domain rule | API/permission | frontend route | persistence test | negative test | visual evidence | commit | status`

Allowed statuses: not inspected, missing, backend partial, backend verified, frontend connected, end-to-end verified, external configuration pending, user acceptance pending.

For every workflow, capture:

- Inputs, required/nullability rules, types, lengths, units, default values and searchable foreign-key options.
- State transitions, who can trigger them, permitted edits after finalization, archive/reversal behavior and retained history.
- Side effects: bed occupancy, stock, charges, payment balance, notifications, audit events and generated documents.
- Request/response schemas, pagination, filters/sorts, error codes, idempotency and concurrency behavior.
- Exact list, create/edit, detail, modal, tab, export/print, loading, empty and failed states.
- Tests and concrete evidence. A page loading, count printed to console, screenshot existing, or seed containing rows does not prove functionality.

Do not simply check all old boxes. Some unchecked items already have partial or complete code; verify and reference it. Some checked items have only UI/demo coverage; reopen them with the specific missing evidence.

## 5. Execution order and checkpoints

### Phase 0 — recover this work safely and establish baseline

1. Read status, diff, branch, remote and worktree list. Preserve pending code above. Check for another agent's changes before fetching or merging.
2. Review migration 048 and its application/store/query changes. Confirm all shift projections scan the added fields; a mismatch was found and fixed during tests.
3. Review compatibility for existing shift callers, duplicate code errors, default deactivation and concurrency. Give users actionable validation messages rather than generic server errors.
4. Run focused tests only where the diff or uncertainty warrants them; reuse recorded evidence for unchanged behavior. Re-run the shift regression before final acceptance if code changed.
5. Commit the verified pending shift slice, QA harness and evidence in cohesive commits. Keep generated environment changes out unless needed by the repository's tooling policy.
6. Establish baseline build, typecheck, tests and formatting. Record pre-existing failures separately from new regressions.
7. Create the workflow matrix above and reconcile the contradictory checklist. Do not restart the entire system.

Exit: clean, understood branch; preserved and committed pending work; runnable isolated QA; clear remaining work.

### Phase 1 — finish source/demo inventory and API contracts

Work module by module, immediately implementing a verified slice rather than spending all time producing another large design document.

- Inspect every sidebar destination, top tab, create/edit form, detail page, row action, filter and print/export. Capture field/action names without copying demo patient data into fixtures.
- Inspect real roles: admin, doctor, nurse, receptionist, accountant, pharmacist, lab technician, case manager and patient. The demo's Super Admin button maps to the intended hospital administration experience; do not introduce SaaS super-admin tenancy.
- Compare Laravel controllers, request validators, repositories, models, migrations, jobs, policies, Blade views and PDF role descriptions. Inventory authored application code; vendor/generated files do not need line-by-line review.
- Document demo/ZIP/PDF conflicts and deliberate security/integrity differences. Do not silently remove a feature based on an older audit document.
- Update OpenAPI against actual handlers. Use one frontend contract/type source and check drift in CI.
- Define shared error shape with stable code, safe message and field errors. Distinguish validation, forbidden, missing, stale version, conflicting idempotency key, unavailable provider and uncertain delivery.

Exit: each implementation slice has an observed/reference contract and explicit acceptance criteria before it is called complete.

### Phase 2 — remove remaining correctness and authorization hazards

- Verify the overview fix through the HTTP/proxy path under real non-admin sessions, not only the repository. Keep role-specific dashboard cards aligned with returned permissions/data.
- Replace active bed-rate sums masquerading as bill totals with authoritative issued-bill totals. Define gross receipts versus net of refunds and reconcile every dashboard/report aggregate to the ledger.
- Review all `catch`, fixture defaults, hardcoded counts, browser storage, success banners and `isLive` flags. Remove automatic fallback from connected workflows. Do not merely hide a misleading banner while still showing fabricated records.
- Loading is not “unavailable”; show separate loading, empty, failure and successful states. On partial dependency failure identify the unavailable section.
- Every visible action must call the actual authorized operation or be clearly unavailable with the reason. Never perform only a local row update for a clinical/financial success.
- Test list/detail/count/export/download scopes independently. Patient/doctor restrictions must cover subqueries and related records, not only the outer query.
- Enforce foreign-key ownership, active-user status and role checks inside the mutation transaction where races matter.

Exit: no unauthorized aggregate disclosure and no connected action reporting success without authoritative persistence.

### Phase 3 — complete attendance as the first full end-to-end module

Backend work:

- Finish shift identity/default work in Phase 0. Validate code uniqueness, times, overnight shifts, minute boundaries and active/default compatibility. Existing half/full-day thresholds are backend requirements; expose them coherently without losing original controls.
- Count current staff by effective assignment/default, excluding patient accounts; agree date/time boundary behavior.
- Duty assignments: real staff/shift IDs, effective-from/to, note and active state. Inspect existing schema/API support before adding migrations. Handle overlapping periods deterministically; preserve historical attendance snapshots.
- Implement delete/archive/end-date behavior compatible with original UX and retained attendance history. Do not hard-delete referenced shifts or assignments.
- Daily report: real staff roles, assigned shifts, absence/leave rules, check-in/out, late/early, worked/overtime/break values, leave type/reason, filters and pagination. Avoid presenting absence simply because a filtered page lacks a record.
- Leave: date range, partial days, type/reason, request/approve/reject/cancel, approver identity, versioning, overlap rules and audit history.
- Attendance corrections: submitted check-out/correction requests, review/approval, reason, original values and calculated durations. Never overwrite history silently.
- Manual attendance: create/correct with staff ID, shift, date/times, break details, remarks and audit actor. Resolve overnight timestamps in hospital timezone.

Frontend work:

- Match original order: Attendance, Daily Report, Shifts, Duty Assignments, Leave Requests, Attendance Requests. Manage Attendance remains a separate sidebar entry. Remove duplicate conflicting subtab rows.
- Match original full-page shift and assignment forms, including required markers, Save/Cancel/back flow. Keep the working persistence while improving layout parity.
- Replace all sample people, sample leaves/requests and fabricated metric counts. Use searchable paginated staff/shift selectors.
- Connect every create/edit/status/review/archive action. Keep the form open on rejection; prevent duplicate submit; preserve the request key when retrying the same operation.
- Link dashboard metrics to appropriately filtered reports as the original does.

Acceptance: admin creates/edits/defaults a shift, assigns real staff, staff clocks in/out with breaks, leave and correction requests are reviewed, admin corrects attendance with retained history, reports and metrics reconcile, reload preserves every result. Unauthorized roles cannot mutate. Empty database contains no sample rows.

### Phase 4 — patient/staff identity, scheduling and front desk

- Complete original demographic/profile fields, guardian/emergency contact, photos/documents and history. Enforce duplicate detection and canonical identity consistently across readers/writers.
- Review existing merge implementation and guard status; test retained IDs, portal ownership conflicts, clinical/financial references, concurrent writes and audit trail before enabling any merge action.
- Staff role-specific profiles, department/specialty, activation/deactivation, care reassignment, invitation acceptance, session revocation and history UI.
- Smart cards: templates, visibility toggles, color, generated unique card IDs, actual patient fields, authorized/signed QR resolution, preview/download/print and revocation. QR must not expose unrestricted patient records.
- Doctor schedules/absences, appointment availability, create/reschedule/cancel with required reason, queues, public booking requests and confirmation. Concurrent requests cannot double-book a slot.
- Reminders and notifications use durable jobs/outbox with deduplication and preferences; do not send live SMS/email in tests.

Acceptance: register/profile/edit/upload/card/appointment flows persist; doctor and patient access are scoped; merging and scheduling do not orphan or duplicate records; cancellation/rescheduling has correct audit, notification and financial effects.

### Phase 5 — IPD/OPD encounter workspaces and discharge

Replace the current aliases of `ipd-diagnosis`, consultant registers, prescriptions, charges, payments, bills and timelines to `EncounterRegister`. A register is the entry point, not every detail tab.

IPD detail tabs, in observed order:

1. Overview: patient/admission/case/bed/doctor identity, symptoms, status and real summary sections.
2. Diagnosis: required report type/date, description, authorized document upload/download, edits/history.
3. Consultant Instruction: consultant/doctor, applied/instruction dates, text and attribution.
4. Operations: operation/category/date, clinicians/technician, notes and appropriate billing linkage.
5. Charges: type/category/code, standard/applied amounts, date, source linkage, authorized corrections.
6. Prescriptions: grouped medication lines, sign/version/replace, document output and dispensing linkage.
7. Timelines: date/title/description, attachment, patient visibility and history.
8. Payments: receipt, amount/date/method/note/document; durable ledger, idempotency, reversal policy.
9. Bills: itemized charges/payments, bed charge, discount/tax/other charges, paid/net amounts, print and Generate Bill & Discharge Patient.

OPD detail tabs: Overview, Visits, Diagnosis, Timelines, Prescriptions. Visits needs Revisit with existing-patient context and preserved previous encounters. Summary totals must come from scoped APIs.

Additional backend requirements:

- Full admission intake, vitals, case/doctor and bed-type/bed dependencies, custom fields, guardian/insurance/package details where required.
- Atomic admission/transfer/bed occupancy, maintenance exclusions, bed history and care-team authorization.
- Signed/closed clinical records retain revisions/addenda; do not allow unrestricted overwrite or delete.
- Discharge validation, financial clearance/authorized exceptions, discharge summary, bed release and bill finalization must agree transactionally. Generate documents from committed data.
- Original Laravel `IpdBillRepository.php` confirms combined bill/discharge side effects but also trusts submitted totals and conflates states. Preserve the flow, not those integrity flaws.
- Explicitly reconcile legacy tax/discount bases with the new ledger contract. Do not invent legal/tax policy or quietly calculate tax on a different base.

Acceptance: a complete admission through transfer, notes/diagnosis, orders, charges, payment, bill, discharge and print; an OPD visit through revisit and prescription; patient portal release checks; concurrent bed allocation and duplicate discharge protection; rollback leaves no partial financial/bed state.

### Phase 6 — financial ledger and advance payments

- Complete the 047 advance-payment API: list/detail/create with patient ownership, generated receipt, amount/date/method, notes, actor and audit. Use integer minor units and idempotency.
- Define allocation to invoices, remaining advance balance, cancellation/refund/reversal and immutable history. Lock balances to prevent concurrent double allocation/refund. Do not add payments twice to dashboard totals when allocated.
- Complete accounts, bills/invoices and itemized sources, payroll, payments, payment reports, advances and manual billing payments, preserving the original eight-tab order.
- Reconcile source charges from appointments, encounters, diagnostics, medicine dispensing, ambulance/services and operations. Same source cannot be billed twice; patient/source mismatches fail.
- Define issued/draft/void/paid/part-paid/refunded states, correction documents, permissions and approval rules. Never silently alter an issued invoice's historical totals.
- Implement exact discount/tax rounding, receipt/slip/PDF formats and report reconciliation. Add overflow/negative/boundary tests.
- Complete income/expense heads and records, transfers, payroll allowances/deductions/approval/payout/slips, insurance/packages/claims and patient responsibility to the agreed scope.
- Providers: verified payment initiation, signed callbacks, duplicate/out-of-order event handling, uncertain outcomes, settlement/refund reconciliation and operator resolution. Leave real credentials blank.

Acceptance: ledger sums match invoices, receipts, advances, source charges, reversals and reports; duplicate requests cannot move money twice; patient access is restricted; local provider tests pass. Live provider acceptance remains separately pending until configured.

### Phase 7 — diagnostics, prescriptions, pharmacy, blood and inventory

- Pathology/radiology catalogs: categories, units, parameters/reference ranges, charges and required original fields; order, sample, result, review/release, amendments, attachments and print.
- Prescriptions: medication/dose/route/frequency/duration and original document fields, grouped lines, signing/replacement, refill/cancellation rules, doctor attribution and patient release.
- Pharmacy: brands/categories/medicines, purchase/batches/expiry, dispense against valid orders, atomic stock movement, returns/recalls, medicine bills/payments and reversals.
- Blood bank: donor/donation/component/screening/expiry/stock/issue/return traceability. Explicit clinician-approved compatibility rules; do not infer treatment decisions from UI colors or labels.
- Inventory: item/category/unit/supplier, stock receipt/issue/return/adjustment/count approval, department/staff/date/attachment fields, low-stock jobs and financial links.
- Odontogram: register and selected patient/doctor, tooth/procedure/legend state, custom legends/colors, revisions, print/export. Preserve original tooth geometry/interaction and persist actual selections; legends are not a substitute for clinical history.

Acceptance: released diagnostics visible only to permitted roles; signed prescriptions drive authorized dispensing; concurrent dispensing/issuing cannot make stock negative; stock/billing reversals reconcile; odontogram survives reload and prints correctly.

### Phase 8 — remaining operational/public modules

Audit and close each of these explicitly; none is complete merely because a component exists:

| Module | Required completion checks |
| --- | --- |
| Documents | Types, upload metadata, owner/record links, private download, versions/archive |
| Front office | Visitors, calls, postal dispatch/receive, dates/contacts, filters, attachments and permissions |
| Enquiries/complaints/reviews/notices | Public versus staff visibility, moderation/status/history and notification effects |
| Front CMS/landing page | Home/about/services/appointment/terms/map, images, testimonials, notice boards; publish/save consistency with public pages |
| Hospital charges | Doctor OPD charges, categories/types/services, pricing snapshots and existing-bill stability |
| Services | Ambulances/calls, service packages/insurance and operations; dispatch lifecycle and billing sources |
| Live consultations/meetings | Patient/visit linkage, role permissions, provider credentials/tokens, safe join links, status and failure states |
| SMS/mail/templates | Template variables, localization, recipients, consent/preferences, outbox, callbacks, retries/uncertain state and delivery history |
| Reports/vaccinations | Birth/death/operation/investigation reports; vaccine catalog/doses/lot/expiry/schedules; authorized correction and print |
| Settings | General profile/logo/favicon, hospital schedule, currency, modules, operations, payment settings, custom fields and queue theme |
| AddOn | Reconcile original user-visible function with safe supported extension policy; do not implement arbitrary uploaded executable code |

Acceptance: every required sidebar/tab has real scoped data and working lifecycle actions, public changes publish correctly, and prints/exports honor permissions and localization.

### Phase 9 — identity/security completion and provider setup

Do this alongside modules, not only after all features are written:

- Real-role permission matrix for all nine roles, including record scope, field visibility, exports and private files. Treat preview role selectors as visual aids only.
- Invitation acceptance, recovery, revoke-all/session management, MFA enrollment/challenge, privileged step-up and care reassignment/deprovisioning UI. Remove production access to security demonstration flows.
- Firebase phone proof expiry, issuer/audience/provider checks, replay prevention, binding to existing accounts, phone change/unlink/recovery, revoked/inactive sessions and takeover races.
- CSRF/origin checking on every mutation, bounded bodies, upload size/content validation, safe filenames/paths, malware quarantine and authorized downloads.
- Security headers/CSP, TLS/secure cookies, secret rotation, least-privilege DB/service accounts, audit access/retention and sensitive-log redaction.
- Dependency/security scanning and remediation; independent assessment is a distinct release prerequisite, not an AI certification claim.

Provider documentation must be a practical user runbook: create account/project, enable service, obtain each setting, exact environment-variable name, file/location, allowed domains/callback URL, development versus production value, local test, controlled live test, expected result, rotation and troubleshooting. Map the actual code's variables; do not invent names.

Cover Firebase, operational SMS provider, SMTP/Mailpit, payment provider(s), meeting provider if used, private storage/scanner, and deployment secrets. No real credential values in the runbook or commits. Add a configuration validator that reports missing keys without exposing values. Real SMS/payment tests need designated recipients/accounts and remain separate from local completion.

### Phase 10 — data migration, performance and operations

- Test all migrations from an empty DB and upgrades from a populated 046 baseline; preserve IDs, clinical history and balances. Test 047/048 backfills specifically.
- Use new migrations for changes to shared schema; do not rewrite old migrations already applied outside disposable tests.
- Trial Laravel import in a disposable database: explicit mapping, counts, identifiers, duplicate handling, dates/timezones, currencies, files, ownership, clinical states and financial/stock reconciliation. Produce an exceptions report.
- Do not use synthetic seeds to conceal missing APIs. Seed scripts must be explicitly development-only, bounded and safe to repeat or clearly single-use.
- Establish user-approved capacity targets rather than inventing “world-class” numbers: concurrent users, patient/visit volumes, list latency, job throughput, availability and recovery objectives.
- Inspect query plans/indexes, pagination/options/export bounds, connection pools and contention. Load-test real workflows and permissions with representative synthetic data.
- Verify durable workers under retries, crashes, duplicate callbacks and provider timeouts. Monitor backlog, failures, dead letters and uncertain outcomes.
- Demonstrate backup and restore into a clean environment, including private files and reconciliation; document encryption/retention/access and recovery timings.
- Stage deployment with separate secrets/environment, migrations, health/readiness, logs/metrics/alerts, rollback and cutover rehearsal. Existing Docker/runbook files alone are not execution evidence.

### Phase 11 — full acceptance and release gate

- Execute the per-role and per-module matrix below. Attach evidence and commit hashes to the checklist.
- Compare original and rebuilt list/form/detail/tab/print at matching viewport, light/dark theme and both languages. Fix functional parity before aesthetic modernization.
- Complete user acceptance using realistic synthetic hospital journeys; explicitly list any deferred feature or external setup. Do not bury exclusions in a “100% done” summary.
- Merge the reviewed branch only after checks pass, no other agent's work is lost and migrations/backups are ready. Update the running local checkout deliberately; do not assume a worktree commit changes the app at port 3000.
- Production remains gated on actual provider configuration, deployment verification, restore drill, security review and user sign-off.

## 6. Mandatory QA matrix

### Every module/API

1. Authorized create/read/update/status/archive path, then reload/new session and assert database state.
2. All denied roles and cross-record ID substitution. Verify data is absent from JSON, counts, exports and downloads as well as hidden in UI.
3. Required/malformed/extra/oversized/Unicode/boundary inputs. Safe error messages and preserved form values.
4. Empty data, slow response, network loss, 401/403/404/409/422/5xx and partial dependency failure. No false success or demo fallback.
5. Same idempotency key/same payload returns same result; same key/different payload conflicts; retry after lost response cannot duplicate work.
6. Concurrent edits return a stale conflict rather than losing changes. Finalized/signed records enforce their lifecycle.
7. Search, filters, sort and pagination operate over the entire dataset; changing filters resets pages appropriately; total count uses identical scope.
8. Audit attribution, old/new state or history linkage, timestamps and notifications agree with committed state.
9. English/Amharic, keyboard/focus/dialog handling, readable responsive tables, dates/units/money and print output.

### Critical transaction tests

| Workflow | Required invariant |
| --- | --- |
| Booking | Two simultaneous requests cannot occupy the same constrained slot |
| Admission/transfer | One active occupant per available bed; rollback restores occupancy/history |
| Discharge | No partial bill/discharge/bed-release result after error or retry |
| Shift defaults | At most one active default; stale/failed change cannot clear the valid default |
| Attendance | Overnight and break totals consistent; corrections retain originals; approvals authorized |
| Dispensing/blood/inventory | Stock never negative; movement/order/financial linkage atomic |
| Invoice/payment/advance/refund | Exact minor-unit totals, no double posting/allocation/refund, immutable issued history |
| Patient identity | Merge/reassignment does not leak portal ownership or orphan clinical/financial references |
| Phone auth | Expired/replayed proof, binding races, revoked users and takeover attempts fail |
| Provider jobs | Crash/timeout/duplicate webhook does not silently resend uncertain operations |

### Nine real-role journeys

- Admin: configure/provision, patient/card, appointment/admission, staffing/attendance, financial review, audit.
- Receptionist: permitted patient registration/scheduling/admission/queue/front desk; denied hospital-wide finance/staff administration.
- Doctor: own schedule/assigned patients, OPD/IPD care, prescriptions/diagnostics/discharge as allowed; denied unrelated records.
- Nurse: assigned care/observations and own attendance; no unauthorized prescription signing or financial actions.
- Accountant: authorized invoices/payments/advances/refunds/reports/payroll; no clinical editing or unrelated clinical detail.
- Pharmacist: catalog/stock/dispensing/returns with signed order context; denied unauthorized clinical and HR data.
- Lab technician: permitted diagnostics and release workflow; denied unrelated clinical/financial administration.
- Case manager: permitted cases/operational/ambulance activities; verify actual contract, do not inherit admin access.
- Patient: only own released records, appointments, invoices, cards/downloads and supported requests; ID substitution denied.

### Evidence standard

For each accepted slice record: reference route, implemented route/API, fixture/schema, test command/scenario, expected versus actual result, relevant screenshot or assertion, commit and remaining limitations. Screenshot presence and a printed row count are smoke evidence only. Strengthen `verify-all-operational-workspaces.mjs` with actual assertions before relying on its success summary.

## 7. Practical commands and environment notes

Run from the active worktree, not blindly from the original checkout:

```powershell
git status --short --branch
git diff --stat
git log -8 --oneline
npm ci
npm run typecheck
npm test
npm run format:check
```

Go commands from `services/api`:

```powershell
go test ./...
go vet ./...
```

`go test ./...` alone skips PostgreSQL integration if `HMS_TEST_DATABASE_URL` is absent. The existing `TestClinicalTransactions` creates a random isolated schema, applies migrations and removes it. Use the existing loopback connection securely as `HMS_TEST_DATABASE_URL`; never print it or point these tests at production.

The normal connected scripts expect `apps/web/.env.local`; this worktree deliberately has no copied credentials. Load an existing approved local environment into the process, or create an ignored worktree-local file with user-supplied values. Do not copy real values to `.env.example`.

Useful existing commands: `npm run test:connected`, `test:integration`, `test:invitations`, `test:recovery`, `test:firebase`, `test:operational`. Inspect each runner's actual assertions and fixture boundaries before using it. Some need Mailpit or emulator prerequisites. Follow the next agent's available browser-tool policy rather than assuming unrestricted Playwright launch is permitted.

Manual isolated browser harness: `node scripts/verify-connected-isolated.mjs --manual` with an approved local environment already loaded. It prints only newly generated disposable QA login details, starts temporary services, and waits for Enter to clean up. Never leave it running or promote that account to the development database.

For production build, use `npm run build` with required environment values loaded. Passing Node `--env-file` directly to the Next executable caused worker `NODE_OPTIONS` errors here; loading the environment in a parent process and spawning Next without that argument worked. The build also regenerates `next-env.d.ts`.

On this Windows host, Git metadata and shared Go/npm caches may need tool approval. Use the environment's normal permission mechanism, not an alternate path to bypass it. Elevated Git may require a narrowly scoped `safe.directory` setting for this worktree. Do not globally disable ownership checks.

## 8. Commit and handoff discipline

- Before each slice: inspect status and existing instructions, define scope and acceptance, confirm current reference/contract.
- Implement migration/domain/store/application/HTTP/proxy/types/UI together where required. Reuse existing services rather than building a parallel generic backend.
- Run focused tests, then broader checks warranted by the change. Do not repeat expensive unchanged tests solely to accumulate pass counts.
- Commit only intended files. Suggested sequence: pending shift identity/default; attendance assignments; attendance reports/reviews; patient/staff gaps; encounter details; clinical tabs; discharge; advances/ledger; diagnostics/pharmacy; remaining modules; security/provider docs; final QA/release evidence.
- In each commit/log record behavior changed, why, test evidence and remaining gaps. Update checklist statuses only to the level proven.
- Do not claim a branch is merged, pushed, deployed or visible at port 3000 unless it actually is.
- Preserve clinical/financial history; do not “fix” a test by weakening authorization, dropping constraints, removing assertions or silently seeding expected values.

## 9. Definition of finished

**Backend complete:** every agreed Section 3 workflow has implemented and verified schema/domain/API/authorization/lifecycle behavior; test evidence covers integrity and failure paths; unresolved policy decisions and real-provider configuration are clearly separated. A migration or contract document alone is not completion.

**Integration complete:** every required original page/tab/action uses authoritative APIs and correct IDs; persists after reload; handles errors/conflicts; preserves original workflow/layout and English/Amharic; no operational preview remains disguised as live.

**QA complete:** the module/role matrix passes, original parity evidence exists, migrations/import/restore and operational failure cases are exercised, and all failures are either fixed or explicitly accepted by the user. No blanket completion claim while required items are missing.

**Ready for production:** the above plus actual external configuration, controlled provider tests, staging/security/backup/cutover verification and user acceptance. The other AI can complete local implementation without the user's credentials, but must not claim live delivery or deployment verification without them.

## 10. First instruction to the next AI

Read this handoff, Git status/diff and repository instructions. Preserve and review the pending shift/default work first, commit the verified slice, reconcile the checklist, then finish attendance assignments/report/review flows and the original encounter-specific IPD/OPD workspaces. Follow the phases and evidence gates above through the remaining backend, integration and QA. Do not restart the project, discard pending work, redesign the UI, fabricate demo data as production data, or mark tasks complete based only on page loading. Commit every verified step and report remaining blockers precisely.
