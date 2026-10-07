# Workspace-by-workspace implementation reference

This expands the source-audit acceptance matrix into individual work packets. The audit observations are historical; the repair notes below supersede only their named findings. Every workspace also requires the shared action/role/ownership/failure tests in SECURITY-AND-ACCEPTANCE.md. A generic list is not completion.

## Identity and users

**Original obligation / workflow:** User plus role-owner row; role middleware and request authorization; active/inactive state; password/profile/MFA/language.

**Audit checkpoint:** Existing auth/HTTP/staff tests; nine-role authenticated probes; CMS permission advertisement discrepancy.

**Required implementation and acceptance:** Nine populated workspaces; editable role/permission policy; create/update/deactivate user; stale sessions; invitation/reset/MFA UI.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Dashboard

**Original obligation / workflow:** Role-specific aggregates and owned records; correct billing sources/date intervals.

**Audit checkpoint:** Overview scope/ready-bed regression passed; page probe.

**Required implementation and acceptance:** Reconcile every metric/chart against its original business meaning; no demo fallback.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Smart-card templates/cards

**Original obligation / workflow:** Unique template name/color; six toggles; all/single/unassigned generation; template delete clears references.

**Audit checkpoint:** Browser proves session-only persistence; separate current token-card API.

**Required implementation and acceptance:** Full template CRUD/generation plus DB and cross-context reload/print/QR.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Appointments/calendar

**Original obligation / workflow:** Patient/doctor/department/date required; doctor timetable, available slots, booking status and payment/custom fields.

**Audit checkpoint:** Existing scheduling/idempotency/overlap tests; catalog page probes.

**Required implementation and acceptance:** Exact form fields, doctor charge, calendar actions, cancellation/status rules and patient self-booking journey.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Queues

**Original obligation / workflow:** Appointment FK, queue number/display order and doctor/date context.

**Audit checkpoint:** Existing queue regression; page probe.

**Required implementation and acceptance:** Operator transitions, public display privacy, original theme/config and concurrent numbering.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Attendance

**Original obligation / workflow:** Not in ZIP; reference is newer demo with shifts/assignments/reports/leaves/corrections.

**Audit checkpoint:** Existing attendance regression/default-shift tests; staff-vs-patient shift-list probes.

**Required implementation and acceptance:** Live-reference field contracts; every non-shift tab; leave/correction approvals; actual frontend persistence.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Patients/cases

**Original obligation / workflow:** User vs patient ID; case number/date/doctor/status; patient demographics/custom fields.

**Audit checkpoint:** Patient create/reload and scoped tests passed.

**Required implementation and acceptance:** Original full patient form; patient/case edit/archive; duplicates/merge and all contact fields.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Admissions

**Original obligation / workflow:** Required patient/doctor/admission date; optional policy; package/insurance/agent/guardian/bed.

**Audit checkpoint:** Existing clinical/bed tests; target admission shape differs.

**Required implementation and acceptance:** Distinct original admission workflow, references and standalone bill link.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## IPD register

**Original obligation / workflow:** Required case/doctor/admission/bed/type; active case checks in controller; bed availability.

**Audit checkpoint:** Existing admission/bed transaction tests; role collection probes.

**Required implementation and acceptance:** Original create/edit/delete/archive rules, unique active case, nurses/lab workspace decision, custom fields.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## OPD register/visits

**Original obligation / workflow:** Required patient/doctor/date/charge/payment mode; case optional; revisit creates visit context.

**Audit checkpoint:** Existing encounter tests; separate IPD/OPD collection probes.

**Required implementation and acceptance:** Nullable case contract, per-doctor prices, total visits/revisit history, payment and custom fields.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## IPD/OPD diagnosis

**Original obligation / workflow:** Report type/date/description/media rather than only ICD fields.

**Audit checkpoint:** Routes currently reuse encounter register.

**Required implementation and acceptance:** CRUD, uploaded report, scoped download and patient-visible behavior.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Consultant instruction

**Original obligation / workflow:** Doctor, applied date, instruction date/text.

**Audit checkpoint:** Generic care team does not represent instruction contract.

**Required implementation and acceptance:** Distinct add/edit/delete/read and doctor choices.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Encounter operations

**Original obligation / workflow:** Operation/category/reference/date, clinician and OT/anesthesia/assistant details.

**Audit checkpoint:** Generic procedure fields differ.

**Required implementation and acceptance:** Original field round trip, operation catalog selection and print/history.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Encounter prescriptions

**Original obligation / workflow:** Header/footer plus category/medicine/dose/interval/day/instruction/time lines.

**Audit checkpoint:** Source captured; existing medication tests cover a different layer.

**Required implementation and acceptance:** Grouped prescription CRUD and print; stock dispensing bridge; validate every line.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Encounter timelines

**Original obligation / workflow:** Title/date/description/media and visible_to_person.

**Audit checkpoint:** Generic register/note path lacks original workflow.

**Required implementation and acceptance:** CRUD and visibility authorization, media replacement/removal.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## IPD charges/payments/bills

**Original obligation / workflow:** Applied-price snapshots; payment before bill; discount/tax/other charges; final bill/discharge/bed release.

**Audit checkpoint:** Existing invoice/clinical transaction tests; source formula differs.

**Required implementation and acceptance:** Reconcile authoritative formula; receipt/bill print; correction/void policy; atomic discharge.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Odontograms

**Original obligation / workflow:** Required chart JSON/patient/doctor/description; multiple chart IDs; chart PDF.

**Audit checkpoint:** Real upsert proves history mismatch; code conversion loses four labels.

**Required implementation and acceptance:** Chart aggregate, lossless codes, clear/edit/delete, patient/doctor scoping and original print.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## General prescriptions

**Original obligation / workflow:** History/vitals/advice/next visit; medication lines and state.

**Audit checkpoint:** Existing pharmacy/prescription tests; role read probes.

**Required implementation and acceptance:** Original line/header fields, diagnosis link, status transitions and PDF.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Doctors/departments

**Original obligation / workflow:** Distinct clinical departments vs role departments; profile, fee, schedule, holidays/breaks.

**Audit checkpoint:** Existing master/staff/scheduling tests; page audit records failures.

**Required implementation and acceptance:** Every doctor field and lookup; in-use department delete; recurring break/holiday semantics.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Beds/types/assignments

**Original obligation / workflow:** Bed business ID/type/rate, availability, assign/discharge and old admission/IPD relation.

**Audit checkpoint:** Existing concurrent occupancy/transfer tests.

**Required implementation and acceptance:** Exact forms and bed-status UI; maintenance vs vacancy; assignment history/print.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Billing accounts/invoices

**Original obligation / workflow:** Account type, line quantities/prices, discounts, identifiers/status and patient scope.

**Audit checkpoint:** Existing invoice protection/payment tests and role probes.

**Required implementation and acceptance:** Draft/edit/issue/void semantics against original; print and per-account totals.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Bills/outgoing payments

**Original obligation / workflow:** Standalone bill/admission link distinct from invoices; payments pay_to/account/date/amount.

**Audit checkpoint:** Original families are not interchangeable; generic billing workspace.

**Required implementation and acceptance:** Document-kind mapping; CRUD/corrections and reports; avoid turning outgoing payments into patient receipts.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Advance payments

**Original obligation / workflow:** Required patient/receipt string/amount/date; notifications.

**Audit checkpoint:** Original CRUD exists; current table mostly overview integration.

**Required implementation and acceptance:** Receipt CRUD or reversal policy, allocation/balance, notification and print.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Finance/payroll

**Original obligation / workflow:** Head enums, amount/date/reference/media; polymorphic payroll owner and salary components/status.

**Audit checkpoint:** Existing finance/payroll tests; expenses role probes.

**Required implementation and acceptance:** Original exports/slips, posted immutability and own-payroll view.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Packages/services/insurance

**Original obligation / workflow:** Package name unique, integer discount, required total; valid service/quantity/rate lines; insurance required identifiers/rate/tax/disease rows; in-use delete blocks.

**Audit checkpoint:** Runtime packages/insurances 404; frontend false local save; original controllers use transactions.

**Required implementation and acceptance:** Real grouped CRUD, line totals, admission linkage, rollback and truthful errors.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Pharmacy masters/purchases/sales

**Original obligation / workflow:** Category/brand/catalog; purchase header+lot/expiry/quantity/amount; grouped sale tax/discount/payment; use/return.

**Audit checkpoint:** Existing batch/ledger tests; dedicated purchase aggregate gap.

**Required implementation and acceptance:** Original full purchase/bill CRUD, effective stock, expiry selection, payments and print.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Inventory

**Original obligation / workflow:** Required department/user/issuer/date/category/item/quantity; return status/date; stock receipts.

**Audit checkpoint:** Existing inventory transaction tests; original issue decrements stock transactionally.

**Required implementation and acceptance:** Exact receipt supplier/store/price fields; issue/return balances, low stock and reports.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Blood bank

**Original obligation / workflow:** Donor/group/bag balances, donation/issue, patient/doctor/cost and issue history.

**Audit checkpoint:** Existing blood tests; aggregate stock denies patients while own issues remain separate.

**Required implementation and acceptance:** Original create/update/delete transitions, stock races, report and role ownership.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Pathology/radiology

**Original obligation / workflow:** Catalog/modality/category/unit/parameters; report days and patient_result association.

**Audit checkpoint:** Existing diagnostic catalog/result/review tests.

**Required implementation and acceptance:** Source-compatible test/parameter CRUD, order/result separation and report layout.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## General diagnosis reports

**Original obligation / workflow:** Patient/doctor/category/report number with named property values.

**Audit checkpoint:** Template alone does not represent report.

**Required implementation and acceptance:** CRUD/validation/print of all original report properties.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Vaccinations/vital reports

**Original obligation / workflow:** Vaccine master/administered doses; patient/case/doctor/date in birth/death/operation/investigation reports.

**Audit checkpoint:** Existing domain/scope tests.

**Required implementation and acceptance:** Original forms, corrections, report attachments and authorized print.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Documents/types/media

**Original obligation / workflow:** Patient/uploader/type/title plus polymorphic media collection.

**Audit checkpoint:** Existing secure attachment tests; original document register not equivalent.

**Required implementation and acceptance:** Metadata CRUD, version/replacement, type lookup and restricted download.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Ambulances/calls

**Original obligation / workflow:** Vehicle identity/driver/type; patient/date/charge/availability and billing.

**Audit checkpoint:** Existing ambulance transaction tests; access matrix probes.

**Required implementation and acceptance:** Original vehicle/call forms, edit/in-use delete policy and invoice link.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Front office

**Original obligation / workflow:** Visitor purpose/identity/contact/times; call type/follow-up; incoming/outgoing postal and media.

**Audit checkpoint:** Existing front-office tests; page probes.

**Required implementation and acceptance:** Conditional original fields, CRUD, attachments and receptionist scope.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## CMS/enquiries/notices/reviews/complaints

**Original obligation / workflow:** Public vs internal content, publication/read state, complaint user-FK/resolution and media.

**Audit checkpoint:** Existing CMS/front-office tests; some UI storage/preview signals.

**Required implementation and acceptance:** CMS editor → public reload, real enquiry submission, notice scope and complaint resolution.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Settings/custom fields

**Original obligation / workflow:** Original keys/branding/currency/language/schedules/modules; typed dynamic field definitions and values.

**Audit checkpoint:** General-settings proxy mismatch; several browser-only settings.

**Required implementation and acceptance:** Persist full forms; safely expose allowed values; original ordering; Amharic; do not expose secrets.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Notifications/mail/SMS

**Original obligation / workflow:** Recipient/read/meta vs delivery history; templates, consent and retry/delivery state.

**Audit checkpoint:** Existing outbox tests; inbox/template structural gaps.

**Required implementation and acceptance:** Inbox read/unread, template editing, real dev Mailpit/SMS capture and retry; Firebase remains separate.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Live consultations/meetings/calendar

**Original obligation / workflow:** Meeting ownership/participants/provider/status; per-user OAuth and event IDs.

**Audit checkpoint:** Existing live-consultation tests; role probes; calendar schema gaps.

**Required implementation and acceptance:** Authorized join/start/cancel, local provider fakes, provider credential lifecycle and original calendar sync scope.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Add-ons

**Original obligation / workflow:** Original package/upload/license/activation UI; code execution is not a required security behavior.

**Audit checkpoint:** Source routes indexed; target scope unresolved.

**Required implementation and acceptance:** Explicit supported metadata/activation behavior and safe replacement for arbitrary code upload.

**Action coverage:** inventory every list/detail/create/edit/delete or archive/status/print/export/download action present in the route and field evidence. Map each to its permitted actors, record scope, state rules, API, transaction and browser test; do not assume all CRUD verbs apply to every record.

## Corrections after the historical matrix

General-settings proxy and transactional bulk updates, CMS permission advertising, admin aggregate doctor absences and inventory movements are repaired. Inventory category/item create and receive/issue persistence are repaired. See ../module-audit/REPAIRS.md and README.md. Full forms, remaining inventory actions and all-module parity remain open.

Attendance is demo-derived, not ZIP-derived. Add-on behavior and unsafe historical role/delete behavior require explicit design decisions. Public CMS, administration, personal staff and patient views must remain distinct even when they use the same underlying entity.
