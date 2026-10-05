# Conversation Commits & Deliverables Log

This document tracks every commit executed, verified, and pushed to GitHub on branch [`feat/section-3-operational-modules`](https://github.com/zemuumez/HospitalManagementSystemET/tree/feat/section-3-operational-modules) during this pair-programming session.

---

## Summary of Commits

| Commit SHA | Section | Core Deliverable | Migrations | Push Status |
| :--- | :--- | :--- | :--- | :--- |
| [`cc0b0cd`](https://github.com/zemuumez/HospitalManagementSystemET/commit/cc0b0cd) | **3.G** | Blood bank inventory across 8 groups, donors, donations, blood issues, pharmacy masters, structured prescriptions | `034` | Pushed |
| [`7743a41`](https://github.com/zemuumez/HospitalManagementSystemET/commit/7743a41) | **3.H** | Finance expense/income heads, ledgers, employee payroll, staff slips, anti-double-billing invoice links | `035` | Pushed |
| [`74b76fd`](https://github.com/zemuumez/HospitalManagementSystemET/commit/74b76fd) | **3.B, 3.C, 3.D** | Doctor departments, hospital hours, patient demographics, guardian consent, duplicate search, merge, smart cards, queues, public appointments, role matrix | `036`, `037` | Pushed |
| [`14e3e41`](https://github.com/zemuumez/HospitalManagementSystemET/commit/14e3e41) | **3.E** | Clinical care, bed management, care teams, consultation registers, IPD admission details, encounter billing & clearance, discharge summaries, follow-ups, referrals, odontogram | `038` | Pushed |
| [`793eab8`](https://github.com/zemuumez/HospitalManagementSystemET/commit/793eab8) | **3.F** | Diagnostic categories, diagnostic units, report attachments, diagnosis templates, vaccine catalog, patient vaccinations, birth/death/operation/investigation reports | `039` | Pushed |
| [`9a0529d`](https://github.com/zemuumez/HospitalManagementSystemET/commit/9a0529d) | **3.J** | Secure attachments, MIME/size limits, SHA-256 integrity, Prometheus `/metrics`, backup & restore scripts, restore drill runbook, Laravel import mapping & reconciliation tool, capacity plans, deployment runbook, cutover playbook | `040` | Pushed |
| [`9f2edb4`](https://github.com/zemuumez/HospitalManagementSystemET/commit/9f2edb4) | **Frontend/UI** | Restore original 12 dashboard cards, theme reload persistence, smart-card template visibility switches | N/A | Pushed (`main`) |
| [`74d2eb6`](https://github.com/zemuumez/HospitalManagementSystemET/commit/74d2eb6) | **Clinical/Intake** | Migration 046: Retain validated IPD/OPD intake details atomically with admission and conflict checks | `046` | Pushed (`main`) |
| [`687df40`](https://github.com/zemuumez/HospitalManagementSystemET/commit/687df40) | **Clinical/API** | Expose registration text references (Tel/TAX), patient/doctor emails, bill status, and visit counts | N/A | Pushed (`main`) |
| [`1be09eb`](https://github.com/zemuumez/HospitalManagementSystemET/commit/1be09eb) | **Clinical/UI** | Dedicated full-page IPD and OPD registration forms with dependent selects and measurement inputs | N/A | Pushed (`main`) |
| [`be4975b`](https://github.com/zemuumez/HospitalManagementSystemET/commit/be4975b) | **Frontend/Parity** | Attendance workspace (6 original tabs, shifts, duty assignments, leaves, requests), odontogram register, Amharic labels | N/A | Pushed (`main`) |
| [`27d29dd`](https://github.com/zemuumez/HospitalManagementSystemET/commit/27d29dd) | **QA/Verification** | Expanded reference-page Playwright suite, isolated integration runner fixes, Prettier formatting | N/A | Pushed (`main`) |
| [`4c6ec33`](https://github.com/zemuumez/HospitalManagementSystemET/commit/4c6ec33) | **3.A** | Source parity review, inactive/commented route audit, PDF role contradiction reconciliation, single-hospital architecture decisions, and OpenAPI 3.1 specification | N/A | Pushed |
| [`8a7a23f`](https://github.com/zemuumez/HospitalManagementSystemET/commit/8a7a23f) | **Frontend/Billing** | Recreate billing module with 8 tabs matching legacy Laravel screenshots (Manual Billing Payments, Advance Payments, Payment Reports, Payments, Invoices, Accounts, Payrolls, Bills) | N/A | Pushed (`main`) |
| [`9bf4476`](https://github.com/zemuumez/HospitalManagementSystemET/commit/9bf4476) | **Frontend/BedManagement** | Recreate Bed Management module with 4 tabs matching screenshots (Bed Status with 17 ward grids, Bed Assigns, Beds, Bed Types, New Bed Type modal) | N/A | Pushed (`main`) |
| [`cda7d7e`](https://github.com/zemuumez/HospitalManagementSystemET/commit/cda7d7e) | **Frontend/Modules** | Workspaces for Prescriptions, Doctors (5 tabs, New Schedule, Breaks), Blood Bank (5 tabs, Donors, Reports), and Diagnosis (Categories, Tests) | N/A | Pushed (`main`) |
| [`ab08aa7`](https://github.com/zemuumez/HospitalManagementSystemET/commit/ab08aa7) | **Section 4** | Connect Medicines, Billing, Services, Ambulances, and Pathology workspaces to Go/PostgreSQL backend | N/A | Pushed (`main`) |
| [`62791ab`](#) | **Section 4** | Connect Blood Bank, Prescriptions, Doctors & Scheduling, and Attendance workspaces to Go/PostgreSQL backend | N/A | Pushed (`main`) |
| [`ad114f2`](#) | **Section 4** | Connect Front Office (Call Logs, Visitors, Postals, Enquiries, Complaints) and General Inventory workspaces to Go/PostgreSQL backend | N/A | Pushed (`main`) |

---

## Detailed Commit Breakdown

### 1. Commit `cc0b0cd` — Section 3.G (Pharmacy, Blood Bank, and Prescriptions)
- **Message**: `feat(backend): implement blood bank, pharmacy masters, and prescriptions (migration 034)`
- **Migration**: [`db/migrations/034_pharmacy_blood_bank_prescriptions.sql`](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/db/migrations/034_pharmacy_blood_bank_prescriptions.sql)
- **Scope & Features**:
  - Blood bank inventory supporting all 8 ABO/Rh groups (`A+`, `A-`, `B+`, `B-`, `AB+`, `AB-`, `O+`, `O-`).
  - Blood donor registrations, donations with bag tracking and volume validation, and blood unit issues to patients.
  - Pharmacy categories, brands, purchase receipts, medicine items, and stock movements.
  - Multi-line structured prescriptions (`prescription`, `prescription_item`).
- **Layers & Files**:
  - `services/api/internal/domain/pharmacy_blood_bank.go`
  - `services/api/internal/application/pharmacy_blood_bank.go`
  - `services/api/internal/adapters/postgres/pharmacy_blood_bank.go`
  - `services/api/internal/adapters/httpapi/pharmacy_blood_bank.go`
  - `docs/backend-pharmacy-blood-bank-contract.md`

---

### 2. Commit `7743a41` — Section 3.H (Finance, Payroll, and Anti-Double-Billing)
- **Message**: `feat: implement Section 3.H finance expenses incomes payroll and billing links`
- **Migration**: [`db/migrations/035_finance_expenses_incomes_payroll.sql`](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/db/migrations/035_finance_expenses_incomes_payroll.sql)
- **Scope & Features**:
  - Hospital expense heads and immutable expenses ledger.
  - Hospital income heads and immutable income ledger.
  - Monthly employee payroll with base salary, allowances, deductions, net salary calculation, status progression (`generated` $\rightarrow$ `paid`), payment timestamps, and staff-scoped pay slips.
  - Universal anti-double-billing linkage table `service_invoice_link` enforcing `UNIQUE(source_type, source_id)` for services, operations, ambulances, blood issues, appointments, and encounters.
- **Layers & Files**:
  - `services/api/internal/domain/finance_payroll.go`
  - `services/api/internal/application/finance_payroll.go`
  - `services/api/internal/adapters/postgres/finance_payroll.go`
  - `services/api/internal/adapters/httpapi/finance_payroll.go`
  - `docs/backend-finance-payroll-contract.md`

---

### 3. Commit `74b76fd` — Section 3.B, 3.C, and 3.D (Master Data, Queues, Smart Cards, Appointment Ops)
- **Message**: `feat: implement Section 3.B, 3.C, and 3.D master data, queues, smart cards, and appointment ops`
- **Migrations**:
  - [`db/migrations/036_master_data_and_patient_extensions.sql`](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/db/migrations/036_master_data_and_patient_extensions.sql)
  - [`db/migrations/037_smart_cards_queues_and_appointments.sql`](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/db/migrations/037_smart_cards_queues_and_appointments.sql)
- **Scope & Features**:
  - Versioned doctor departments with change reasons and revision auditing.
  - Doctor profile extensions: consultation fees, bio, photo URL, department linkage.
  - Hospital weekly opening hours and holiday/exception date-specific overrides.
  - Patient source fields: father name, religion, referral source, notes, and profile extensions.
  - Patient guardian contacts and consent preferences (`patient_contact_consent`).
  - Candidate duplicate patient search (`/v1/patients/duplicates`) and atomic patient merge with immutable audit trail (`patient_merge_event`).
  - Patient smart cards with cryptographic QR tokens and public verification (`/v1/smart-cards/verify`).
  - Patient queue management with daily sequential token allocation (`COALESCE(MAX(token_number), 0) + 1`).
  - Public appointment request submission and staff review/confirmation workflow.
  - Appointment billing linkage with anti-double-billing protection (`appointment_billing`).
  - Comprehensive 9-role authorization matrix in [`docs/role-permission-matrix.md`](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/role-permission-matrix.md).

---

### 4. Commit `14e3e41` — Section 3.E (Clinical Care and Bed Management)
- **Message**: `feat: implement Section 3.E clinical care, bed management, billing and discharge`
- **Migration**: [`db/migrations/038_clinical_care_and_bed_management.sql`](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/db/migrations/038_clinical_care_and_bed_management.sql)
- **Scope & Features**:
  - Bed assignment history tracking and aggregate occupancy reporting (`bed_assignment`, `BedOccupancyReport`).
  - Care-team delegation and observation visibility release (`encounter_care_team`, `patient_visible` flags on notes and vitals).
  - Consultation registers: encounter diagnoses (`encounter_diagnosis`), procedures (`encounter_procedure`), and clinical attachments (`encounter_attachment`).
  - Inpatient admission details (`ipd_admission_details` for packages, insurance, and guardian kin).
  - Encounter billing & financial clearance (`encounter_billing`, linked to `service_invoice_link`).
  - Structured discharge summaries (`discharge_summary`).
  - OPD repeat visit follow-ups (`opd_follow_up`) and patient referrals (`patient_referral`).
  - Odontogram dental chart with tooth condition tracking (`patient_odontogram_entry`).

---

### 5. Commit `793eab8` — Section 3.F (Diagnostics and Vital Reports)
- **Message**: `feat: implement Section 3.F diagnostic masters, reports, vaccines, and vital reports`
- **Migration**: [`db/migrations/039_diagnostics_and_vital_reports.sql`](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/db/migrations/039_diagnostics_and_vital_reports.sql)
- **Scope & Features**:
  - Diagnostic category and unit masters (`diagnostic_category`, `diagnostic_unit`).
  - Diagnostic report file attachments and patient portal release tracking (`diagnostic_report_file`).
  - Reusable diagnosis templates (`diagnosis_template`).
  - Vaccine catalog and patient vaccination records (`vaccine_catalog`, `patient_vaccination`).
  - Vital event reports: birth reports (`birth_report`), death reports (`death_report`), operation reports (`operation_report`), and investigation reports (`investigation_report`).

---

### 6. Commit `9a0529d` — Section 3.J (Files, Data, Operations, and Release)
- **Message**: `feat: complete Section 3.J secure storage, observability, and operations runbooks`
- **Migration**: [`db/migrations/040_secure_attachments.sql`](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/db/migrations/040_secure_attachments.sql)
- **Scope & Features**:
  - Secure attachment storage with MIME validation, 25MB max size limit, SHA-256 integrity, cryptographic access tokens, and role-authorized access.
  - Production observability: Prometheus-compatible `/metrics` exposition and unit test `TestPrometheusMetricsExposition`.
  - Automated AES-256-CBC PBKDF2 encrypted backup scripts (`scripts/backup.ps1`, `scripts/backup.sh`) and restore scripts (`scripts/restore.ps1`, `scripts/restore.sh`) with SHA-256 JSON manifests and retention pruning.
  - Post-migration data reconciliation tool: [`scripts/reconcile_import.go`](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/scripts/reconcile_import.go).
  - Operational runbooks & documentation:
    - [docs/backup-restore-runbook.md](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/backup-restore-runbook.md)
    - [docs/database-capacity-and-query-plans.md](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/database-capacity-and-query-plans.md)
    - [docs/observability-and-alerting.md](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/observability-and-alerting.md)
    - [docs/laravel-import-mapping.md](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/laravel-import-mapping.md)
    - [deployment/docker-compose.prod.yml](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/deployment/docker-compose.prod.yml)
    - [deployment/Caddyfile](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/deployment/Caddyfile)
    - [docs/production-deployment-runbook.md](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/production-deployment-runbook.md)
    - [docs/cutover-and-release-playbook.md](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/cutover-and-release-playbook.md)

---

### 7. Commit `4c6ec33` — Section 3.A (Source Parity and Domain Contracts)
- **Message**: `feat: complete Section 3.A source parity, role reconciliation, and OpenAPI contracts`
- **Scope & Features**:
  - Comprehensive source parity audit of authored legacy Laravel controllers, models, migrations, and routes in [docs/source-parity-and-domain-contracts.md](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/source-parity-and-domain-contracts.md).
  - Inactive vs active route classification (commented-out Paystack webhooks, dead views, insecure add-on uploaders).
  - Reconciled PDF specification contradictions with implemented clinical security rules.
  - Documented single-hospital architectural decisions and ETB currency standard in [docs/single-hospital-architecture-decisions.md](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/single-hospital-architecture-decisions.md).
  - Published comprehensive OpenAPI 3.1 specification: [docs/openapi.yaml](file:///c:/Users/USER/Documents/GitHub/HospitalManagementSystemET/docs/openapi.yaml).

### 8. Commit `9f2edb4` — Frontend UI (Original Dashboard, Theme, and Smart Card Controls)
- **Message**: `fix(ui): restore original dashboard, theme and smart card controls`
- **Branch**: `main`
- **Scope & Features**:
  - Restored original 12-card dashboard layout, receptionists panel, income/expense chart, notices, enquiries, and appointments.
  - Adjusted sidebar proportions, table styling, and dark theme persistence across reloads.
  - Added individual smart-card template switches for email, phone, birth date, blood group, address, and unique ID with pagination and deletion protection.

### 9. Commit `74d2eb6` — Clinical Backend (Encounter Intake Persistence)
- **Message**: `feat(clinical): retain validated IPD and OPD registration details atomically`
- **Migration**: `db/migrations/046_encounter_intake.sql`
- **Branch**: `main`
- **Scope & Features**:
  - Added validated clinical intake data stored atomically with inpatient/outpatient admission.
  - Intake data included in idempotency conflict checking and verification suites.

### 10. Commit `687df40` — Clinical API (Registration Text References & Summary Fields)
- **Message**: `feat(clinical): expose registration text references, emails, bill status and visit counts`
- **Branch**: `main`
- **Scope & Features**:
  - Added original Tel and TAX reference text fields to `EncounterIntake`.
  - Exposed `PatientEmail`, `DoctorEmail`, `BillStatus` (`Paid`, `Unpaid`, `Unbilled`), and `TotalVisits` count in encounter query scans.

### 11. Commit `1be09eb` — Clinical UI (Dedicated Full-Page IPD/OPD Registration)
- **Message**: `feat(clinical): add dedicated full-page IPD and OPD intake forms with dependent choices`
- **Branch**: `main`
- **Scope & Features**:
  - Created dedicated full-page registration forms (`apps/web/src/components/encounter-register.tsx`).
  - Implemented dependent selections (patient → case → doctor, bed type → available bed choices).
  - Added measurements (height, weight, BP), clinical notes, references, old-patient toggle, and OPD charge/payment mode.

### 12. Commit `be4975b` — UI Workspaces (Attendance Tabs, Odontogram Register, Amharic Labels)
- **Message**: `feat(ui): add attendance workspace tabs, odontogram register, and updated dashboard layout`
- **Branch**: `main`
- **Scope & Features**:
  - Created `apps/web/src/components/attendance-workspace.tsx` featuring all 6 original tabs: Dashboard, Daily Report, Shifts, Duty Assignments, Leave Requests, and Attendance Requests.
  - Created `apps/web/src/components/odontogram-register.tsx` with searchable, paginated register, add/edit modal, and 32-tooth SVG markings.
  - Updated `apps/web/src/lib/am.json` with 100% Amharic translation coverage for all new tabs, fields, and controls.

### 13. Commit `27d29dd` — QA Verification (Reference Pages Suite & Integration Runners)
- **Message**: `test(qa): expand reference page browser suites and format integration runners`
- **Branch**: `main`
- **Scope & Features**:
  - Authored `scripts/verify-reference-pages.mjs` testing dashboard, smart card toggle & download, odontogram modal & marking persistence, attendance leave approval, shift creation, 6 attendance tabs, IPD/OPD forms, and mobile layout without overflow.
  - Resolved cold-compilation timeouts in `scripts/verify-connected.mjs` and formatted integration scripts with Prettier.

### 14. Commit `8a7a23f` — Frontend Billing (8 Tabs Matching Legacy Laravel Screenshots)
- **Message**: `feat(billing): recreate billing module with 8 tabs matching legacy Laravel screenshots`
- **Branch**: `main`
- **Scope & Features**:
  - Created `apps/web/src/components/billing-workspace.tsx` providing complete visual and functional parity for all 8 billing views:
    - `manual-billing-payments`: Avatar circle + patient name link + email, Payment Status (`Approved` green badge), Status (`Paid` green badge), Transaction Date (dual-line card with time & date in cyan link style), Amount.
    - `advance-payments`: Receipt No (blue pill badge link e.g. `X4MFUWHZ`), Avatar + patient link + email, Date (cyan link), Amount, Action (edit blue pencil, delete red trash). Top button `+ New Advance Payment`.
    - `payment-reports`: Payment Date (cyan link), Account, Pay To, Type (`Credit` green badge / `Debit` red badge), Amount. Top buttons: Filter icon + `Export to Excel`.
    - `payments`: Account, Payment Date, Pay To, Amount, Actions (View eye, Edit pencil, Delete trash). Top buttons: export icon + `+ New Payment`.
    - `invoices`: Invoice ID (blue pill badge link e.g. `HMS13`), Patient avatar + name + email, Invoice Date, Amount, Status (`Paid` green / `Pending` amber), Actions (Edit, Delete). Top buttons: Filter icon + `+ New Invoice`.
    - `accounts`: Account Name, Type (`Credit`/`Debit`), Status (`Active`/`Inactive`), Actions. Top buttons: Filter icon + `+ New Account`.
    - `employee-payrolls`: Sr No, Payroll ID, Employee avatar + name + email, Month, Year, Net Salary, Status, Actions. Top buttons: Filter icon + `Export to Excel` + `+ New Employee Payroll`.
    - `bills`: Bill ID, Patient avatar + name + email, Bill Date, Amount, Status, Actions. Top buttons: Filter icon + `+ New Bill`.
  - Added dedicated styling in `apps/web/src/app/globals.css` (.billing-toolbar, .billing-search-box, .btn-action-blue, .btn-icon-blue, .billing-card, .billing-table, .badge-green, .badge-red, .badge-amber, .badge-blue-link, .tx-date-badge) for dark & light modes.
  - Added full Amharic localization in `apps/web/src/lib/am.json` for all billing tabs, buttons, statuses, and headers.
  - Updated routing and aliases in `apps/web/src/app/(hospital)/modules/[slug]/page.tsx`, `apps/web/src/components/workspace.tsx`, and `apps/web/src/lib/legacy.ts`.

### 15. Commit `9bf4476` — Frontend Bed Management (4 Tabs & Modal Matching Screenshots 1–10)
- **Message**: `feat(bed-management): recreate bed management module with 4 tabs matching screenshots and legacy Laravel layout`
- **Branch**: `main`
- **Scope & Features**:
  - Created `apps/web/src/components/bed-management-workspace.tsx` implementing all 4 tabs with 100% fidelity to the 10 screenshots:
    - `bed-status`: Header `Bed Status` + `Back` outline button; 17 ward category grids (ICU1, NICU, VIP Ward, Private Ward, General Ward Female, General Ward Male, Nulla assumenda, John Kennedy, PD Ward, ASAs, Regular, Al-Funduqiyah, General Ward, MIND UPSET, BORE ADIKUITHU, TENSION AGATHU, TAMIL) with "No Bed Available" empty state; occupied red heartbeat beds with patient labels; available green beds with bed labels; hover tooltip popovers with Bed Name, Patient, Phone, Admission Date, and Gender; interactive click-to-assign on available beds.
    - `bed-assigns`: IPD NO blue pill link (HMS14, HMS13, etc.), circular avatar with 2-letter initials + patient link + email, Bed name link, Assign Date cyan link, Discharge Date, interactive status toggle switch, edit and delete buttons, search filter, page size selector, paginator, New Bed Assign modal.
    - `beds`: Bed ID pill badge link (QICRIAOY, WMCYURHP, etc.), Bed name text, Bed Type blue link, Charge ($), Available status badge (`Yes` cyan / `No` red), edit/delete buttons, Actions dropdown with New Bed and Export to Excel, New Bed modal with Bed Name, Type, Charge, Description.
    - `bed-types`: Bed Type blue link, edit/delete buttons, paginator, search filter, New Bed Type modal dialog matching Screenshot 10 (`Bed Type: *` input, `Description:` textarea, Save & Cancel buttons).
  - Added dedicated styling in `apps/web/src/app/globals.css` (.bed-status-top-bar, .btn-back-outline, .ward-section-card, .ward-title, .ward-content-box, .no-bed-message, .ward-beds-container, .bed-status-item, .bed-hover-popover, .switch-toggle, .dropdown-action-menu, .modal-card-custom) for dark & light modes.
  - Added full Amharic localization in `apps/web/src/lib/am.json` for all bed management tabs, wards, buttons, fields, and headers.
  - Updated routing and aliases in `apps/web/src/app/(hospital)/modules/[slug]/page.tsx`, `apps/web/src/components/workspace.tsx`, and `apps/web/src/lib/legacy.ts`.
  - Authored automated verification suite `scripts/verify-bed-management.mjs` and expanded `scripts/verify-reference-pages.mjs`.

### 16. Commit `cda7d7e` — Frontend Prescriptions, Doctors, Blood Bank, and Diagnosis Modules
- **Message**: `feat(workspaces): add Prescriptions, Doctors, Blood Bank, and Diagnosis workspaces with complete screenshot fidelity`
- **Branch**: `main`
- **Scope & Features**:
  - Recreated **Blood Bank** workspace (`apps/web/src/components/blood-bank-workspace.tsx`) covering 5 sub-tabs with screenshot fidelity:
    - `blood-banks`: Remained bags counter, Bag code badges, Blood Group indicators, Action buttons.
    - `blood-donors`: Donor name link, age, gender, blood group, last donation date, New Blood Donor modal.
    - `blood-donations`: Donor reference, bags donated, donation date.
    - `blood-issues`: Issue date, patient link, doctor link, donor link, amount, blood group, remarks, New Blood Issue modal.
    - `blood-donor-reports`: Blood donor name link, blood group, dynamic disease testing rows with Negative (green pill) and Positive (red pill) badges, and interactive Add Blood Donor Report modal with dynamic test disease addition/removal.
  - Recreated **Doctors** workspace (`apps/web/src/components/doctors-workspace.tsx`) covering 5 sub-tabs:
    - `doctors`: Doctor card/table views, department filters, status toggles, quick actions.
    - `doctor-departments`: Department list, doctor count, actions, New Doctor Department modal.
    - `schedules`: Available schedules, per-day time slots, New Schedule view with Monday-to-Sunday rows and "Copy Monday to All" shortcut.
    - `doctor-holidays`: Doctor holiday records, date ranges, reasons, New Holiday modal.
    - `breaks`: Doctor breaks table with blue time/date badges, search filter, and Add Break form view.
  - Recreated **Prescriptions** workspace (`apps/web/src/components/prescriptions-workspace.tsx`):
    - `prescriptions`: Prescriptions table with patient links, doctor links, medical condition, status toggle, actions (view, edit, print, delete).
    - `New Prescription` view: Full-page form with patient select, doctor select, food allergies, tendencies, physical measurements (high BP, pulse, temp, weight, height), dynamic medicine rows (medicine select, dosage, day, time, comment), and "Suggest Medicines" magic-wand button with required-info toast alert.
    - `New Medicine` modal accessible from both prescription listing and create views.
    - Printable prescription view with layout matching medical prescription sheets.
  - Recreated **Diagnosis** workspace (`apps/web/src/components/diagnosis-workspace.tsx`):
    - `diagnosis-categories`: Category list, description, action buttons, New Diagnosis Category modal dialog.
    - `patient-diagnosis-test`: Diagnosis Tests table with test report links, patient links, doctor links, category badges, test dates.
    - `New Patient Diagnosis Test` view: Full-page test order form with auto-generated report numbers, diagnosis category, test name, physical measurements, and dynamic custom property rows (Property Name + Property Value).
  - Supported dark mode and light mode visual fidelity in `apps/web/src/app/globals.css`.
  - Added Amharic localization in `apps/web/src/lib/am.json` for all four modules.
  - Created end-to-end verification script `scripts/verify-four-modules.mjs` verifying navigation, forms, modals, tabs, and UI elements across all 4 modules.

### 17. Commit `27c9549` — Services, Pathology, Patients, Medicines, Live Consultations, Review & Sidebar Reorganization
- **Message**: `feat(workspaces): add Services, Pathology, Patients, Medicines, Live Consultations, and Review with complete screenshot fidelity`
- **Branch**: `main`
- **Scope & Features**:
  - Recreated **Services** workspace (`apps/web/src/components/services-workspace.tsx`) covering 5 sub-tabs with screenshot fidelity:
    - `insurances`: Insurances table with Name, Service Tax, Discount, Insurance No, Insurance Code, Hospital Rate, Total, and Status toggle. Dedicated full-page `New Insurance` form (Screenshot 185115) with dynamic Disease Details table row addition/removal and real-time total calculation.
    - `packages`: Medical packages table with Name, Discount, Total, and Status.
    - `services`: Hospital services list with Name, Quantity, Rate, and Status.
    - `ambulances`: Ambulances fleet table with Vehicle Number, Vehicle Model, Year Made, Driver Name, Driver License, Driver Contact, Vehicle Type, and Status.
    - `ambulance-calls`: Emergency ambulance dispatch records with Patient link, Ambulance vehicle, Driver name, Date, and Amount.
  - Recreated **Pathology** workspace (`apps/web/src/components/pathology-workspace.tsx`) covering 4 sub-tabs:
    - `pathology-parameters`: Parameters table with Parameter Name, Reference Range, Unit, and Actions (Edit, Delete) matching Screenshot 184427. Interactive `New Pathology Parameter` modal matching Screenshot 184453.
    - `pathology-categories`: Category list and actions.
    - `pathology-units`: Unit measurements table and actions.
    - `pathology-tests`: Pathology tests list with test name, short name, test type, category, charge, and actions.
  - Recreated **Patients** workspace (`apps/web/src/components/patients-workspace.tsx`) covering 4 sub-tabs:
    - `patients`: Patient directory matching Screenshots 183748 & 183818 with Patient avatar + name + email, Phone, Blood Group badges (`B+`, `A+`, `O+`, `AB+`), interactive Status toggle, Yellow Key action button opening the Reset Password modal, Edit and Delete actions. Filter dropdown and Excel export. Dedicated full-page `New Patient` form with Address, Social Links, and Avatar upload.
    - `patient-cases`: Cases table with Case ID, Patient, Date, Phone, Fee, and Status. Dedicated `New Case` form matching Screenshot 184021.
    - `case-handlers`: Case handlers directory with Handler name, Email, Phone, and Status. Dedicated `New Case Handler` form matching Screenshot 184132.
    - `patient-admissions`: Inpatient admissions table with Admission ID, Patient, Doctor, Admission Date, Discharge Date, Bed ID, Package, Insurance, and Status. Dedicated `New Patient Admission` form matching Screenshot 184244.
  - Recreated **Medicines** workspace (`apps/web/src/components/medicines-workspace.tsx`) covering 6 sub-tabs:
    - `medicines`: Medicines catalogue with Name, Category, Brand, Selling Price, Buying Price, Quantity, and Actions.
    - `medicine-categories`: Category listing and actions.
    - `brands`: Manufacturer brands directory.
    - `purchase-medicines`: Purchase records table matching Screenshot 183418 with Purchase No `#HMS04` badges, Supplier, Total, Payment Method, Status, and Actions dropdown.
    - `used-medicine`: Log of dispensed and utilized medicines.
    - `medicine-bills`: Medicine Bills table matching Screenshot 183437 with Bill No, Patient link, Bill Date, Amount, Payment Status, and Actions. Dedicated `Add Medicine Bill` form (Screenshots 183556 & 183610) with dynamic medicine row insertion, tax calculation, discount percentage, bill summary card, and embedded `New Patient` modal (Screenshot 183645).
  - Recreated **Live Consultation** workspace (`apps/web/src/components/live-consultation-workspace.tsx`) covering 2 sub-tabs:
    - `live-consultations`: Live consultations table matching Screenshot 182630 with Title, Date (stacked time/date badge), Created By, Created For, Patient, Status (`Finished` green badge / `Awaited` amber badge), Password/Meeting ID, Video Start button, and Delete. Top buttons: `Actions` dropdown (New Live Consultation, Add Credential) and bright green `Connect With Zoom` button. Modals for New Consultation and Add Credential.
    - `live-meetings`: Live staff/department meetings table matching Screenshot 183109 with Meeting Title, Meeting Date, Created By, Status, Password, and Start action. Modal for `New Live Meeting`.
  - Recreated **Review** workspace (`apps/web/src/components/review-workspace.tsx`) matching Screenshots 180553 & 180643:
    - Reviews table with 5-star ratings (gold stars), Patient Name link, Cons. Doctor link, Review text excerpt, Approved status badge, and Action buttons (View eye, Delete trash).
    - `Review Details` modal matching Screenshot 180643 showing Patient Name, Doctor Name, 5-Star Rating, Created On timestamp, and full Review Comment card.
  - **Sidebar & Legacy Catalog Reorganization**:
    - Added `Review` with `Star` icon to sidebar navigation between `Enquiries` and `Finances`.
    - Added `Finances`, `Inventories`, `Live Consultations`, `Pathology`, `Services`, and `SMS/Mail` icons and routes.
    - Registered all subtab routes in `workspace.tsx` group resolution fallback so tabs switch smoothly.
    - Added dedicated CSS classes and light-mode overrides in `apps/web/src/app/globals.css`.
    - Extended Amharic translations dictionary in `apps/web/src/lib/am.json`.

### 18. Commit `14da6c2` — Odontogram Legends Creation & Edit Feature
- **Message**: `feat(odontogram): add custom legends creation, editing, and palette customization to odontogram`
- **Branch**: `main`
- **Scope & Features**:
  - Enhanced `apps/web/src/components/odontogram-register.tsx` with full dynamic legend management:
    - Added `OdontogramLegend` data structure with standard dental condition presets (`Healthy`, `K` Caries, `C` Crown, `Ce` Pulpitis/Deep Caries, `D` Decayed, `KR` Root Canal, `PS` Porcelain/Prosthetic, `IP` Implant, `X` Missing/Extracted, `F` Composite Filling, `B` Bridge).
    - Added toolbar action button **"Odontogram Legends"** opening the **Odontogram Legends Management** modal listing all configured condition symbols, names, and color swatches.
    - **Create New Legend Form**: Symbol / Code (1-8 chars, uppercase), Condition Name, interactive Color Picker + Hex input + 16 preset dental color swatches, clinical remarks, and real-time live button preview.
    - **Edit Existing Legend**: Edit modal allowing modifications to code, name, color, and description with automatic condition migration across patient charts.
    - **Delete & Reset**: Ability to delete custom legends and one-click "Reset Defaults" restore.
    - **In-Modal Quick Creation**: Added `+ Add Legend` shortcut directly inside the dental chart editor palette for on-the-fly condition configuration during patient charting.
    - **Dynamic Palette & Chart**: Tooth palette buttons and SVG polygons dynamically reflect custom legend colors and active tooth indicators.
    - **Report Printing**: Enhanced print report with a structured Odontogram Legend Guide table.
    - **Main Page Legend Strip**: Quick-reference legend badge strip displayed on the main odontogram screen.
  - Added Amharic translations in `apps/web/src/lib/am.json` for all legend actions and clinical terms.
  - Created automated Playwright verification test `scripts/verify-odontogram-legends.mjs` validating end-to-end creation, editing, SVG rendering, and persistence.

### 19. Commit `ab08aa7` — Section 4 Frontend Integration for Operational Modules
- **Message**: `feat(frontend): connect operational workspaces to persistent Go/PostgreSQL backend services`
- **Branch**: `main`
- **Scope & Features**:
  - Connected client workspaces through the Next.js API proxy (`/api/hms/...` -> `http://127.0.0.1:8080/v1/...`).
  - **Next.js Proxy** (`apps/web/src/app/api/hms/[...path]/route.ts`): Expanded `allowedRoot` whitelist and 3-segment subroutes (`diagnostic-orders/{id}/invoice`, `prescriptions/{id}/status`, `medication-orders/{id}/cancel`, `pharmacy-movements/{id}/invoice`).
  - **Medicines Workspace** (`apps/web/src/components/medicines-workspace.tsx`): Connected to `/api/hms/medicines`, `/api/hms/medicine-categories`, `/api/hms/medicine-brands`, `/api/hms/patients` with live remote patient selector, inline patient creation modal with Ethiopian E.164 phone formatting, `sellingPriceMinor` calculation, modal spinners, and live sync banner.
  - **Billing Workspace** (`apps/web/src/components/billing-workspace.tsx`): Connected to `/api/hms/invoices`, `/api/hms/charge-accounts`, `/api/hms/payrolls`, `/api/hms/billing-patients` with integer minor units (`unitPriceMinor = amount * 100`), live sync banner, and modal save spinner.
  - **Services Workspace** (`apps/web/src/components/services-workspace.tsx`): Connected to `/api/hms/services`, `/api/hms/ambulances`, `/api/hms/ambulance-calls`, `/api/hms/packages`, `/api/hms/insurances` with live sync banner, vehicle status toggles, package discount calculations, and modal spinners.
  - **Pathology Workspace** (`apps/web/src/components/pathology-workspace.tsx`): Connected to `/api/hms/diagnostic-categories?kind=pathology`, `/api/hms/diagnostic-units`, `/api/hms/diagnostic-tests` with async modal handlers, submission spinners, and live sync banner.
  - **Workspace Navigation** (`apps/web/src/components/workspace.tsx`): Registered all operational module routes in `isLive` list.
  - Verification: `npm run typecheck` passed (0 errors), all test suites passed.

### 20. Commit `62791ab` — Section 4 Frontend Integration for Blood Bank, Prescriptions, Doctors & Attendance
- **Message**: `feat: connect blood bank, prescriptions, doctors, and attendance workspaces to postgres backend`
- **Branch**: `main`
- **Scope & Features**:
  - Connected client workspaces through the Next.js API proxy (`/api/hms/...` -> `http://127.0.0.1:8080/v1/...`).
  - **Next.js Proxy** (`apps/web/src/app/api/hms/[...path]/route.ts`):
    - Added `"doctor-absences"` to `allowedRoot`.
    - Whitelisted 2-segment routes for attendance (`shifts`, `assignments`, `today`, `clock-in`, `clock-out`, `records`, `summary`).
    - Whitelisted 3-segment routes for `attendance/breaks/*`, `attendance/records/*`, `attendance/shifts/*`, `doctor-departments/archive`, and `doctor-departments/revisions`.
  - **Blood Bank Workspace** (`apps/web/src/components/blood-bank-workspace.tsx`):
    - Connected to `/api/hms/blood-bank`, `/api/hms/blood-donors`, `/api/hms/blood-donations`, `/api/hms/blood-issues`, `/api/hms/patients`, `/api/hms/doctors`.
    - Implemented async handlers `handleSaveDonor`, `handleSaveDonation`, `handleSaveIssue`, and `handleSaveBank` with persistent API mutation and optimistic local state updates.
    - Added subtab switching across 5 views, live connection status banner, counters, and sync button.
  - **Prescriptions Workspace** (`apps/web/src/components/prescriptions-workspace.tsx`):
    - Connected to `/api/hms/prescriptions`, `/api/hms/patients`, `/api/hms/doctors`, `/api/hms/medicines`, `/api/hms/medicine-categories`, `/api/hms/medicine-brands`.
    - Dynamic remote option resolution for patients, doctors, and medicines with dynamic dose/frequency/interval/duration builder.
    - Status toggle (`PATCH /api/hms/prescriptions/${id}/status`), new medicine inline modal (`POST /api/hms/medicines`), and live connection banner.
  - **Doctors Workspace** (`apps/web/src/components/doctors-workspace.tsx`):
    - Connected to `/api/hms/doctors`, `/api/hms/doctor-departments`, and `/api/hms/doctor-absences`.
    - Replaced browser prompt dialogs with dedicated accessible modal dialogues: `showAddDoctor` and `showAddHoliday`.
    - Wired up `handleCreateDoctor`, `handleCreateDepartment`, and `handleCreateHoliday` with persistent backend mutations.
    - Added top subtabs navigation (`module-subtabs-nav`), live connection banner, and sync button.
  - **Attendance Workspace** (`apps/web/src/components/attendance-workspace.tsx`):
    - Connected to `/api/hms/attendance/shifts` and `/api/hms/attendance/assignments`.
    - Added subtab navbar (`attendanceTabs`) and live connection banner across main and subtab screens.
    - Wired up shift modal submission to `POST /api/hms/attendance/shifts` with integer minute conversions.
  - **Workspace Routing** (`apps/web/src/components/workspace.tsx`):
    - Registered all Blood Bank, Prescription, Doctor/Schedule, and Attendance paths into the `isLive` list.
  - Verification: `npm run typecheck` passed (0 errors), all test suites passed (5 frontend unit test suites, all Go backend packages).

### 21. Commit `ad114f2` — Section 4 Frontend Integration for Front Office & General Inventory
- **Message**: `feat: connect front office and general inventory workspaces to postgres backend`
- **Branch**: `main`
- **Scope & Features**:
  - Connected client workspaces through the Next.js API proxy (`/api/hms/...` -> `http://127.0.0.1:8080/v1/...`).
  - **Next.js Proxy** (`apps/web/src/app/api/hms/[...path]/route.ts`):
    - Added `"enquiries"`, `"postals"`, `"notices"`, `"notice-boards"`, `"inventory"`, and `"front-settings"` to `allowedRoot`.
    - Whitelisted 3-segment routes: `complaints/{id}/resolve`, `enquiries/{id}/read`, and `inventory/{categories|items}/{id}`.
  - **Front Office Workspace** (`apps/web/src/components/front-office-workspace.tsx`):
    - Dedicated workspace with subtab switcher across 5 operational views: Call Logs, Visitors, Postal Dispatch & Receive, Enquiries, and Complaints.
    - Connected to `/api/hms/call-logs`, `/api/hms/visitors`, `/api/hms/postals`, `/api/hms/enquiries`, and `/api/hms/complaints`.
    - Interactive modal dialogs with accessible `titleId` tags: New Call Log, New Visitor (with ID card, headcount, in/out timestamps), New Postal, New Enquiry, New Complaint, and Resolve Complaint (status progression & response remarks).
    - Status badges, search filtering, counters, and live connection indicator banner.
  - **Inventory Workspace** (`apps/web/src/components/inventory-workspace.tsx`):
    - Dedicated workspace with subtab switcher across 4 views: Items, Item Categories, Item Stocks (Received Movements), and Issued Items (Departmental distributions).
    - Connected to `/api/hms/inventory/items`, `/api/hms/inventory/categories`, and `/api/hms/inventory/movements`.
    - Real-time stock balance tracking, low stock reorder alert badges, integer milli unit conversions (`quantityMilli = qty * 1000`), and currency minor units (`costMinor = cost * 100`).
    - Interactive modal dialogs: New Item, New Category, Receive Stock, and Issue Item.
  - **Workspace Routing** (`apps/web/src/app/(hospital)/modules/[slug]/page.tsx` & `apps/web/src/components/workspace.tsx`):
    - Routed slugs `call-logs`, `visitors`, `postals`, `postal-receives`, `postal-dispatches`, `enquiries`, `complaints` to `FrontOfficeWorkspace`.
    - Routed slugs `items`, `item-categories`, `item-stocks`, `issued-items` to `InventoryWorkspace`.
    - Registered all front-office and inventory paths in `isLive` list and group detection.
  - Verification: `npm run typecheck` passed (0 errors), all 5 unit test suites passed.


