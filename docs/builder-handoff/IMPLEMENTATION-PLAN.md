# Ordered implementation and integration plan

## Working agreement

Work against the current repository, not a fresh rewrite. Keep existing migrations and verified protections. Build complete vertical slices: schema → domain/use case → repository transaction → HTTP authorization/validation → Next proxy → original form/list/detail/print → real persistence tests. Never finish all generic tables first and assume the UI will fit later.

For every ticket create a short evidence record with original fields/actions, current gaps, explicit policy decisions, migrations/endpoints, permitted roles and record scope, tests executed/results, remaining gaps and commit hash. Use [WORKSPACES.md](WORKSPACES.md) as the module backlog and the 1,202-row field register for column-level coverage. A checkbox means its acceptance evidence exists, not that code was generated.

## Phase 0 — establish the exact baseline

- Inspect status, branch, remote and latest commits; do not overwrite another builder's changes.
- Read README and repair evidence. Confirm the inventory and settings repairs remain present.
- Inspect migration ledger, environment examples, scripts and current contracts. Keep provider secrets blank. Do not reset the shared database.
- Start an isolated database/schema and local services using repository scripts. Establish real Better Auth login for all nine roles with representative linked records.
- Reproduce only the next ticket's known gap first. Avoid spending credits rerunning the entire browser crawl after every small change.

Exit: reproducible baseline and one bounded ticket selected, with source/target discrepancies written down.

## Phase 1 — close concrete missing endpoints and false saves

1. Packages/insurance: execute FIRST-TICKET.md completely.
2. Doctors: align user/profile creation, department and fee fields, active state, schedules, holidays/breaks and absence scope. Remove any preview record/save fallback. Verify doctor versus user IDs and partial-creation rollback.
3. Settings: connect the full original General form to the atomic endpoint; add authorized image upload/reference lifecycle; persist currency/language/schedule/module/custom-field settings. Do not return provider secrets to the browser.
4. Inventory completion: supplier/store catalogs and receipt fields, edit/archive, pagination/search, return/write-off transitions, stock reconciliation and original issue form. Retain current idempotency/error protections.

Exit: each named screen supports real actions and reload, with denied roles and failure behavior tested. Do not claim unrelated modules complete.

## Phase 2 — restore aggregates that generic CRUD cannot represent

5. Smart cards: template CRUD, unique name/color, six field toggles, all/single/unassigned generation, patient assignment, safe QR/print/download, template/card deletion semantics. Separate template presentation from secure token authorization.
6. Odontogram: migrate to chart-level history, preserve all original codes without lossy conversion, required patient/doctor/description, add/edit/delete/print and scoped history. Preserve existing tooth data through an explicit migration rather than discarding it.
7. Patients/cases/admissions: full forms/custom fields, duplicate detection/merge policy, distinct standalone admission and package/insurance linkage. Verify identity maps and active-case rules.
8. OPD: optional source case, doctor prices, payment mode, visit count/history, revisit, original form fields and state rules.
9. IPD: admission/edit, valid active case, bed selection/occupancy, separate detail tabs and concurrency protections.

Exit: original aggregates round-trip with their fields and history, not generic notes or a repeated encounter register.

## Phase 3 — complete clinical workspaces

10. Implement IPD tabs in original order: Overview, Diagnosis, Consultant Instruction, Operations, Charges, Prescriptions, Timelines, Payments, Bills. Implement OPD Overview, Visits, Diagnosis, Timelines, Prescriptions. Each tab needs distinct child CRUD, attachments, validation, role/record scope and print where applicable.
11. General prescriptions: complete header/vitals/advice/follow-up and medication lines; bridge dispensing without making a prescription itself consume stock twice.
12. Diagnostics: pathology/radiology masters, units/parameters, orders/results, general diagnosis property reports, review/release and report layouts. Preserve patient-specific results separately from catalog definitions.
13. Nursing/vaccination/birth/death/operation/investigation records: reconcile original and target-extension fields, corrections, ownership, attachments and print.
14. Documents/media: metadata/type CRUD, private upload/download, replacement/versioning and deletion policy across every parent module.

Exit: a multi-actor clinical journey completes with unrelated-patient access denied; history remains readable after master-data updates.

## Phase 4 — finance, pharmacy and stock end to end

15. Reconcile invoice, standalone bill, IPD bill, outgoing payment, advance receipt and online transaction meanings. Record formulas before implementing them. The legacy IPD formula is not interchangeable with invoice tax logic; decide tax/discount/payment/bed-charge treatment explicitly.
16. Implement document-specific create/edit/issue/void/reversal/allocation and receipt/print/report flows. Make discharge, bill finalization and bed release atomic where required. Preserve history instead of deleting old bed assignments.
17. Pharmacy purchase header/lines/batches/expiry, sales/bills/payment, dispensing/use/return and authoritative stock. Test concurrent last-unit dispensing.
18. Blood donors/donations/issues and ambulances/calls: source fields, availability/stock transitions, costs and linked patient/billing records.
19. Finance/payroll: accounts, heads, salary components, posting/slips/export and own-payroll access. Reconcile dashboard totals with these actual sources.

Exit: money and stock reconcile after retries, reversals and concurrent requests; printed and dashboard totals match committed records.

## Phase 5 — remaining operational and public journeys

20. Attendance: shifts/default assignment, duty periods, daily report, clock/break records, leave/correction approval and manual management. ZIP has no attendance implementation; use documented demo contracts and record all inferred rules, including overnight shifts/time zones/partial-day leave. Test approver versus requester and conflicting approvals.
21. Appointment/calendar/queue: booking/rescheduling/cancellation, availability, fees, patient self-service, public queue privacy and concurrent numbering.
22. Front office: visitors/calls/postal and media, conditional fields and follow-up.
23. CMS: editing → publication → public reload for home/about/services/appointment/terms/map, plus enquiries/notices/testimonials/reviews/complaints and resolution states.
24. Messaging: inbox/read state, templates, recipients, email/SMS outbox/retry/delivery events. Verify Mailpit locally. Firebase phone proof remains separate from operational SMS.
25. Live consultation/meeting and calendar provider flows: participant authorization, join/start/cancel, stored provider references, callback validation and local provider fakes. Document credentials without requiring them for local tests.
26. Add-ons: record supported metadata/activation scope and a safe replacement for arbitrary executable uploads. Do not implement unrestricted code upload to imitate a demo button.

Exit: all families in WORKSPACES.md have explicit implemented/accepted/excluded/blocked status; no silent “not applicable” gaps.

## Phase 6 — complete parity and release evidence

- Run populated CRUD/role/ownership matrices for every module and its child actions, exports and downloads.
- Reconcile every source field and action to target implementation or a documented intentional exclusion. Review conditional form requirements and enum/state behavior, not just labels.
- Perform desktop/mobile, light/dark, English/Amharic, keyboard/error/empty/loading checks against the captured original. Preserve relevance-based sidebar/tab order.
- Execute multi-actor journeys, concurrency/idempotency and security tests; verify query plans for major registers and bounded pagination.
- Rehearse schema upgrade/import/backup/restore and documented deployment/rollback. Separate credential-dependent provider smoke tests from local acceptance.
- Update the delivery checklist with exact evidence and remaining external prerequisites. Prepare a user acceptance demo using synthetic data.

## Commands and efficient verification

Inspect package.json and scripts before running; commands below reflect the checkpoint, not a guarantee that later revisions retain them:

```text
npm run typecheck
npm test
npm run format:check
node --env-file=apps/web/.env.local scripts/verify-connected-isolated.mjs --endpoint-repairs
npm run test:integration
```

Go tests must run from the actual Go module with the isolated PostgreSQL configuration expected by the repository. Find the current go.mod and test harness rather than inventing a database URL. Read continuous-verification.md and the isolated runner before using shared ports/schemas. The endpoint-repairs mode is the regression for the recent 55-check repair run; add module tests rather than treating that suite as universal coverage.

For a slice: run focused domain/repository tests → relevant HTTP/role tests → typecheck/frontend tests → real-login browser round-trip. Run the broader suite at meaningful checkpoints or after shared-layer changes. Save concise evidence, update checklists and commit. Push verified commits to main and sync the user's clean Desktop checkout without force or destructive reset.

## Definition of complete

A module is complete only when original fields/actions and deliberate deviations are documented; schema constraints and transactions are correct; permissions cover role and record ownership; every displayed action has real persistence or an explicit disabled explanation; reload/print/export work; failure and retry do not fabricate success; acceptance tests pass; documentation and checklist evidence are committed. Provider credentials being blank may block live delivery certification, but not local integration and provider-failure testing.
