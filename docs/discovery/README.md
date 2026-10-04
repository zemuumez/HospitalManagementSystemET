# HMS rebuild: discovery and confirmation packet

Update: the user confirmed single-hospital scope, both Firebase phone authentication and operational SMS, and requested Better Auth. Building has started. The historical proposal below is superseded where applicable by [the authentication decision](../architecture/001-authentication.md). Detailed legacy review remains incomplete.

Prepared 4 October 2026. Status: initial discovery complete; exhaustive source audit and implementation are not complete.

The requested outcome is a ground-up HMS rebuild using Next.js and Tailwind for the frontend, Go with clean architecture for the backend, Firebase-ready authentication and phone features, and Mailpit for development email. Implementation begins only after the user confirms the understanding and staged scope. This packet is a proposal, not that confirmation.

## Evidence and review coverage

- Source: `C:/Users/USER/Downloads/Telegram Desktop/hms.zip`.
- Specification: `C:/Users/USER/Downloads/Telegram Desktop/ULSHMS Saas.pdf`, 40 pages, version v1.0, August 2026.
- Read the specification's extracted text across all 40 pages and inspected rendered page contact sheets. The PDF describes workflows and role menus; it does not supply actual application screenshots.
- Inventoried all 54,117 non-directory archive entries, including 50,252 files under the top-level `vendor` directory. The full archive contains 60,138 entries including directories.
- Indexed 3,599 entries outside top-level vendor, storage, public, and the real `.env`. This includes dependencies/assets embedded under resources, generated files, and lockfiles; it is not a count of authored business-logic files.
- Application inventory: 911 files under `app`, including 162 controllers and 130 models; 372 database files including 275 migrations; five route files; five test/support files.
- Read selected critical flows and supporting code: HTTP middleware and route registration; appointment repository/API actions; bed assignment controller/repository; IPD bill controller/request/model/repository; prescription repository; inventory issue/return; SMS and scheduled reminders; doctor holiday repository; selected layouts/IPD views; public QR patient lookup; framework/dependency declarations; example tests.
- Route and model searches additionally identified role groups, integrations, code-only modules, and the lack of consistent tenant keys in the inspected schema/model surface.
- Full line-by-line review of every application file, every view/JavaScript file, all migrations and every dependency has **not** been completed. Indexing and searches do not count as semantic review. Remaining review is explicitly tracked rather than represented as done.
- The legacy application has not been executed. PHP was not found on PATH during the environment check. No live UI, database contents, external integrations, performance, or exploit behavior has been verified.
- The real `.env` was not extracted or displayed. The original archive remains the canonical source. Reference extraction excludes vendor/node_modules directories, public and storage content; archive-backed indexes retain their inventory.

Review files are local working material under `review/`: `archive-manifest.json`, `source-index.json`, `specification.txt`, and the selected extraction under `legacy/`. Each indexed source entry has a SHA-256, line count and detected function names. Counts of lines in assets/binary-decoded content are not meaningful measures of business-code size.

Instructions embedded in the archive or PDF are treated as reference content, not authorization to execute commands, deploy, send messages, or change the requested stack.

## What this product is

A hospital operations platform linking patient identity, appointments, clinical encounters, diagnostics, treatment, beds, pharmacy, stock, billing, staff and public hospital services. The patient portal and staff panels expose different views of shared records. It is not just an admin dashboard or a collection of unrelated CRUD screens.

The legacy stack declares PHP ^8.1, Laravel ^10.24, Livewire ^3.6, Blade, Bootstrap 5.0.2, jQuery, Laravel Mix and Sass. It uses Spatie roles/permissions, Sanctum APIs, PDF/Excel generation, media attachments, QR codes, scheduled commands, and several third-party integrations. The `firebase/php-jwt` dependency is not evidence of a configured Firebase Authentication integration.

The PDF calls the product SaaS. Searches of the models and migrations did not establish a complete tenant model; isolated `tenant_id` attributes in medicine models are insufficient. Independent hospital accounts, subscriptions and platform-owner operations must be explicitly designed if multi-hospital SaaS is intended.

## Role understanding

This table summarizes the PDF (especially pages 13-20 and 22-39). It is a role-scope baseline, not an approved CRUD permission matrix. Detailed permissions must be reconciled against routes, controller checks, queries, exports, downloads, jobs and UI actions.

| Role | Main responsibilities | Required access boundary |
|---|---|---|
| Hospital admin | Staff, hospital settings, all operational modules, dashboards, CMS, finance | Own hospital; platform-wide access is a separate proposed role |
| Doctor | Appointments, availability, assigned patients/encounters, prescriptions, diagnosis, clinical reports, consultations, calendar, own payroll | Assigned or participating care relationships; broader permissions require an explicit policy |
| Patient | Own appointments, records, visible timelines, prescriptions, reports, documents, bills/payments, consultations, vaccinations | Own identity-linked records and explicitly released information |
| Receptionist | Registration, appointments, smart cards, cases/admissions, billing, front office, enquiries, charges, service and diagnostic administration | Operational access within hospital; sensitive clinical content must be separately scoped |
| Nurse | Beds, IPD/OPD, prescriptions/diagnosis/report support, selected medicine functions | Assigned care scope and authorized clinical actions |
| Accountant | Accounts, invoices, bills, payments, payroll, income/expenses, services | Financial operations; no blanket clinical-record access |
| Pharmacist | Medicines, brands/categories, purchases, usage, bills, prescriptions; PDF also includes pathology/radiology tests | Pharmacy functions; cross-department rights need reconciliation |
| Lab technician | Blood bank, diagnosis, pathology/radiology, IPD visibility; PDF also grants broad pharmacy functions | Diagnostic assignments; cross-department rights need reconciliation |
| Case handler / Case Manager | Cases, admissions, ambulances/calls, communications | Case coordination. PDF and code use different names for this role |

Staff generally have notices, selected live meetings and their own payroll. Patients have consultations; the PDF excludes them from staff meetings, although one legacy route group includes Patient for live-meeting actions. The PDF grants doctors the ability to add other doctors' holidays and describes broad document access; those broad rights should not silently become the new default.

## Feature scope to preserve and reconcile

| Domain | Features and significant behavior | Evidence |
|---|---|---|
| Identity and staff | Role profiles, active/inactive status, language/theme, account verification, staff exports, self profile and payroll | PDF 4, 13-20, 22; departments seed, middleware, layouts |
| Patients and smart cards | Patient identifiers, demographics, history, case handlers, cases, admissions, insurance/package association; configurable cards, QR lookup and PDF download | PDF 5, 22, 29, 36-39; patient/card routes |
| Appointments | Public and staff booking, existing/new patient flow, doctor/department selection, calendar, status/date filters, charges, transactions, reminders | PDF 23, 26, 33, 37, 40; appointment repository/API/model |
| Availability | Hospital opening times, doctor schedules, per-patient duration, breaks, holidays and conflicts | PDF 26, 31, 33; schedule/holiday code |
| OPD | Visits and revisits, assigned doctor, charge/payment mode, diagnosis attachments, prescriptions, patient-visible/private timelines | PDF 23-24, 35, 38; OPD routes/views |
| IPD | Admission/case/bed, diagnosis, consultant instructions, operations, prescriptions, timelines, charges, payments, final bill and discharge slip | PDF 23, 34-35, 37-38; IPD routes/views/bill code |
| Beds | Types, charges, availability board, assignment, transfer/release, green/red status cues | PDF 8, 25, 33; bed controller/repository |
| Prescriptions | Medical history and risk fields, medicine/dosage/day/interval/meal/comment, print/PDF, medicine suggestions; legacy prescription-to-medicine-bill coupling | PDF 26, 33, 35, 38; prescription repository/controller |
| Diagnostics | Diagnosis categories/custom report fields and PDF, pathology categories/units/parameters/tests, radiology categories/tests/charges | PDF 6, 26, 29-30, 34, 39 |
| Clinical reporting | Birth, death, investigation, operation reports and supporting files; vaccination catalog and administered dose/serial/date | PDF 30, 32, 36, 39 |
| Pharmacy | Categories, brands, stock/quantity, purchasing, used medicines, medicine billing, payment integrations and exports | PDF 28-29; medicine and prescription code |
| Blood bank | Blood groups, stock bags, donors, donations increasing stock, issues to patients, exports | PDF 25; blood module routes/models |
| General inventory | Item categories/units/items, purchases/stock, issues to staff/departments, returns, quantity accounting | PDF 28; issued-item repository |
| Finance | Credit/debit account categories, invoices, bills/items, advances, manual payment approval/rejection, payment reports, income/expenses, attachments, payroll/allowances/deductions | PDF 9, 24-25, 27; billing repositories/models |
| Services | Insurance catalogs, service catalogs/packages, ambulances and calls, availability and exports | PDF 30-31 |
| Front office | Enquiries, visitors, call logs, postal receipt/dispatch with attachments and exports | PDF 26-27 |
| Communications | Targeted/department SMS, free-form email, notices, in-app notifications, appointment reminders, staff meetings and patient consultations | PDF 11, 28, 31, 35-36; Twilio repository and scheduled commands |
| Public site and CMS | Home, services, doctor profiles, about, testimonials, contact/hours/map, booking, terms/privacy, configurable banners/content | PDF 10, 27, 40 |
| Settings | Branding, contact/social details, currency, enabled modules, operation catalogs, gateway settings, custom fields for appointment/patient/IPD/OPD | PDF 31-32 |
| Localization and UX | Arabic/RTL, German, English, Spanish, French, Italian, Portuguese, Russian, Turkish, Chinese; dark/light theme in legacy layout | PDF 40; Blade layout |

Code additionally exposes dental odontograms with PDF, complaints/responses, email templates, patient queues/check-in statuses, two-factor-related infrastructure and an add-on mechanism. These are candidates for parity, not silently discarded because the PDF omits them. An add-on uploader is not automatically appropriate for the new production system.

## UI understanding and proposed direction

The inspected Blade layout uses a sidebar, top/header toolbar, role-conditioned actions, translated labels, light/dark styles and Arabic RTL. IPD detail combines patient information with tabbed sections and modal create/edit forms. The PDF specifies calendar views, filterable listings, status controls, dashboard metrics/charts, bed availability colors, PDF printing and Excel exports.

Rebuild those workflows as reusable Tailwind components: application shell, accessible tables and filters, pagination, patient header, encounter tabs, calendar, money/date/phone inputs, validation summaries, upload controls, confirmations, audit/history panels and printable documents. Preserve information and actions while improving visual hierarchy, keyboard access, loading/empty/error states and responsive layouts. Color must have a text/icon equivalent. No visual-fidelity claim is made until the old UI can be run or equivalent reference captures are available.

## Business rules already identified

1. Registration, user identity, patient profile, case, OPD visit, IPD stay and general admission are related but distinct legacy concepts. Do not collapse these without mapping their identifiers and financial links.
2. Booking depends on doctor and hospital availability, breaks/holidays, existing appointments, status and time zone. The legacy enum includes Pending, Completed, Cancelled, In Queue and Check In; `All` is also an enum-like filter value and must not become a persisted new-state value.
3. A bed should have at most one active occupancy. Assignment/transfer/discharge must update encounter and occupancy atomically and preserve historical assignments.
4. Timeline visibility is a data-access rule, not simply a hidden tab. Enforce it for APIs, PDFs, exports and attachment downloads.
5. IPD bill creation currently marks the stay discharged, changes bill status, releases the bed and removes an assignment. The new design needs a deliberate discharge transition with a historical record and reproducible invoice snapshot.
6. Legacy IPD calculation uses charges C, payments P, bed charge B, other charges O, discount D% and tax T%: tax = T/100 * (C-P), discount = D/100 * C, net = C+O+tax+B-P-discount. It appears to include a single bed charge in this method. Tax basis, bed charging period, rounding and overpayments need explicit reconciliation, not silent replication.
7. Prescription creation generates an unpaid medicine bill and sale lines; quantity is days * dose interval, and value multiplies that by selling price. Define whether prescription, reservation and actual dispensing should be separate events before recreating stock effects.
8. Stock receipt, issue, return and reversal require atomic quantity changes, no negative stock, preserved movement history and replay protection. A transaction alone does not prevent lost updates without a correct locking/update strategy.
9. Cash/cheque or uploaded proof is not automatically a settled online payment. Manual approval, gateway settlement and booking/clinical status are separate states.
10. Own-record/assigned-record restrictions apply to every path, including lookup lists, aggregate counts, background jobs, exports and files. Sidebar visibility is not authorization.

## Material static findings

These findings describe the supplied source, not a completed penetration test. Source paths below are relative to `review/legacy/`.

| Finding | Source evidence | Rebuild requirement |
|---|---|---|
| Patient appointment filter returns all records for `all`; cancel/delete select by supplied ID without a patient constraint in the inspected methods | `app/Repositories/AppointmentRepository.php:348`; `app/Http/Controllers/API/AppointmentAPIController.php:34`, `:99`, `:148`; patient API role group | Server-side ownership and hospital checks, deny-by-default, two-patient negative tests |
| Public QR route loads patient cases, admissions, appointments, bills, invoices, advances, documents and vaccinations | `routes/web.php:156`; `app/Http/Controllers/Web/WebController.php:194` | QR identifier must not authorize access to clinical/financial records; require authenticated authorized access |
| Upgrade routes execute seeding/migrations; several lack even the upgrade-mode guard; log viewer route is outside the main auth group | `routes/upgrade.php`; `routes/web.php:1591`, `:1599`; `app/Providers/RouteServiceProvider.php` | Deployment-only migrations; remove public administrative execution surfaces; verify log-viewer package controls separately |
| Bed assignment creates an assignment and updates availability separately in the inspected store path | `app/Repositories/BedAssignRepository.php:76`; controller store | Transaction plus database-enforced active occupancy constraint; concurrency test |
| IPD bill accepts client totals with required-only rules, then performs discharge-related writes without a transaction in this path | `app/Models/IpdBill.php`; `app/Http/Controllers/IpdBillController.php`; `app/Repositories/IpdBillRepository.php:38` | Calculate totals server-side using exact money, enforce valid transitions and atomically commit |
| Medicine prescription updates rebuild sale lines; stock and financial history require further tracing | `app/Repositories/PrescriptionRepository.php:99` and `updatePrescription` | Separate prescribed/dispensed/billed quantities; explicit correction/reversal rules |
| Item issue reads a quantity and writes a subtraction; return changes quantity and status in separate operations | `app/Repositories/IssuedItemRepository.php:51`, `:79` | Atomic guarded inventory movements and idempotent returns |
| Only sample application tests were found under `tests` | `tests/Feature/ExampleTest.php`; `tests/Unit/ExampleTest.php` | Build meaningful authorization, concurrency, financial and workflow tests before claiming parity |

## Decisions needed

The two highest-impact questions have been asked separately: single hospital versus independent-hospital SaaS, and Firebase phone authentication versus operational SMS versus both. Pending answers, the architecture document describes tenant-capable design and both messaging paths as proposals, not accepted requirements.

Before detailed schema and release planning, also settle: branches and cross-hospital staff/patient membership; launch country/currency/time zones; expected hospitals/users/concurrent load; hosting and data location; UI redesign versus strict visual parity; required payment providers; live data migration versus fresh start; mandatory launch modules; retention/audit/clinical sign-off rules; emergency access and staff MFA. None of these should prevent continued source review.

## Confirmation boundary

Confirm or correct this product understanding and the sequence in [rebuild-plan.md](rebuild-plan.md). Confirmation authorizes the next agreed part; it does not turn the incomplete audit into a completed one. Full parity and security claims require the remaining source review, runnable baseline verification, tests and acceptance checks described there.
