# Live demo integration audit — 2026-10-06

Reference: https://hms.infyom.com/login, Super Admin shortcut followed by Login. Observations below are read-only; no demo records were created, edited, deleted, billed or discharged. This is a partial audit, not a claim of exhaustive coverage.

## Confirmed flows

| Area | Observed original behavior | Integration requirement |
| --- | --- | --- |
| IPD register | Search, filter, new admission; number opens detail, separate edit/delete actions | Encounter ID must select the detail workspace; tabs cannot all display the register |
| IPD details | Overview, Diagnosis, Consultant Instruction, Operations, Charges, Prescriptions, Timelines, Payments, Bills | Keep this ordering and bind each tab to the selected admission |
| IPD diagnosis | Modal requires report type/date; description and document optional | Persist report metadata and private attachment; reload and access-scope checks |
| IPD payments | New Payment; date, amount, payment mode, document, note columns | Durable ledger with idempotency and correct admission ownership |
| IPD bills | Charge/payment tables; bed charge, total charges, gross total, discount percentage, tax percentage, other charges, paid amount and net payable; print bill/discharge slip; Generate Bill & Discharge Patient | Atomic finalization with bed release and printable outputs. Do not treat a displayed total as a posted bill |
| OPD details | Patient summary with total cases/admissions/appointments; Overview, Visits, Diagnosis, Timelines, Prescriptions | Separate encounter detail and patient visit history |
| OPD visits | Revisit opens create form with revisit context; visit list includes doctor/date/charge/payment/symptoms/notes/edit/delete | New encounter tied to existing patient; preserve prior visit history |
| Attendance dashboard | Metrics link to filtered reports/lists; shift summary and recent leave requests | Real counts and filter navigation, not hardcoded zero/sample staff |
| Attendance tabs | Attendance, Daily Report, Shifts, Duty Assignments, Leave Requests, Attendance Requests; Manage Attendance is a separate sidebar item | Restore original order and avoid duplicate/reordered tabs |
| Shift create | Full-page form: required name/code/start/end, grace/break minutes, default flag, active flag, Save/Cancel | Persist code/default separately from name/active; uniqueness and atomic default switch |
| Duty assignment create | Required staff/shift/effective-from; optional effective-to/note; active flag | Use staff IDs and shift IDs; persist note/active state and effective dates |
| Advance payments | Receipt number, patient, date, amount columns; modal requires patient, generated receipt, amount/date | Implement real advance-payment API; table/dashboard sum alone is insufficient |

## Original source cross-check

Read `hms/app/Repositories/IpdBillRepository.php` from the supplied ZIP. `saveBill` changes the discharge/bill state and releases the bed. `getBillList` combines admission charges, payments, bed charge, percentage discount/tax and other charges. The old code accepts submitted totals and conflates bill state with discharge; preserve the visible flow while enforcing server-calculated amounts and transactional integrity in Go. Tax/discount bases must be explicitly reconciled and covered by tests before claiming financial parity.

No attendance-shift implementation was found under the searched names in the supplied ZIP; the newer deployed demo remains the observed reference for these controls.

## Remaining audit

- [ ] All IPD/OPD tab forms, editing, attachments and print layouts.
- [ ] Assignment lifecycle, leave approval/rejection, attendance corrections and history.
- [ ] Every remaining sidebar module, subtab, modal, filter, action and role-specific restriction.
- [ ] Local browser comparison at matching viewport, light/dark modes, English/Amharic and narrow layouts.
- [ ] Persistence after reload, validation failures, authorization denials and concurrent edits for each integrated flow.

The public demo can establish visible behavior, but its UI alone does not establish hidden business rules or security. Cross-check source and backend constraints before implementation.
