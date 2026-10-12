# Comprehensive HMS Modules Audit & Parity Roadmap (2026-10-12)

**Audit Date:** 2026-10-12  
**Baseline Commit:** `adcecd2` on `main`  
**Working Branch:** `audit/remaining-module-parity`  
**Boundary Enforcement:** Strictly non-mutating audit. No application changes, database migrations, or live-data modifications were executed. Preserves verified and accepted parity records across Doctors, Packages & Insurance, Settings & CMS, Inventory, Smart Cards, and Odontograms.

---

## 1. Executive Summary

Following the full acceptance and production integration of the **Odontogram Workspace** at `adcecd2`, an exhaustive system audit was performed across all remaining Hospital Management System (HMS) modules. The audit compared legacy Laravel route declarations (`LEGACY-ACTIONS.md`), controllers/repositories (`LEGACY-WORKFLOWS.md`), validation models (`MODEL-RULES.md`), and current Next.js/Go implementations.

### Key Audit Findings
1. **Proliferation of Disconnected Mock UI Workspaces:** While the Go backend and PostgreSQL schemas contain extensive table declarations (migrations 001–062), numerous frontend workspaces remain purely client-side React components populated with hardcoded static fixtures (`useState`). These workspaces simulate successful saves, filter queries, and deletions in component memory, creating a false impression of working functionality that is wiped on browser refresh.
2. **Clinical Detail Routing Collapses:** Routes for IPD detail registers (diagnoses, consultant registers, prescriptions, charges, payments, bills, timelines, and operations) are routed generically to `EncounterRegister kind="ipd"`. Dedicated legacy child entities (e.g. instruction dates, itemized charge snapshots, operation clinician rosters, patient-visible flags) are absent.
3. **High-Impact Architectural Prerequisite:** **Bed Management** (`bed-status`, `bed-assigns`, `beds`, `bed-types`) represents the foundational physical constraint of all inpatient care (IPD). Currently, `BedManagementWorkspace` is 100% disconnected mock data with hardcoded wards and simulated beds, despite existing PostgreSQL tables and Go services.

---

## 2. Module Inventory and Classification

Each module family across the HMS system is classified into one of four verified states:
- **Accepted:** Formal acceptance review published, regression suite verified against PostgreSQL, browser/build verified, and integrated into `main`.
- **Partial:** Endpoints or database tables exist, but significant functional gaps, disconnected UI workspaces, mock data, or missing persistence mechanisms remain.
- **Missing:** Workflow or child models lack persistence models, or UI relies entirely on generic in-memory fallback screens (`LegacyScreen`).
- **Not verified:** Backend services exist, but lack dedicated regression test coverage and end-to-end browser verification.

| Module Family | Status | Primary Code Paths | Observed Gaps & Audit Evidence |
| :--- | :--- | :--- | :--- |
| **Doctor Profiles & Schedules** | **Accepted** | `doctors-workspace.tsx`<br>`services/api/.../doctors.go` | Accepted in `DOCTOR-WORKSPACE-ACCEPTANCE-2026-10-09.md` (`be4bfa9`). Full schedule, holiday, break, and charge parity. |
| **Medical Packages & Insurance** | **Accepted** | `services-workspace.tsx`<br>`services/api/.../packages.go` | Accepted in `PACKAGES-INSURANCE-ACCEPTANCE-2026-10-08.md` (`997195c`). Transactional line aggregates, in-use protection. |
| **Settings, CMS & Custom Fields** | **Accepted** | `settings-workspace.tsx`<br>`services/api/.../cms_settings.go` | Accepted in `SETTINGS-WORKSPACE-ACCEPTANCE-2026-10-10.md` (`d6978ce`). Hospital schedule, currency, branding, custom fields. |
| **Inventory Management** | **Accepted** | `inventory-workspace.tsx`<br>`services/api/.../inventory.go` | Accepted in `INVENTORY-WORKSPACE-ACCEPTANCE-2026-10-11.md` (`cf3a0e1`). Ledger sequencing, negative-stock barriers, voiding. |
| **Patient Smart Cards** | **Accepted** | `smart-cards-workspace.tsx`<br>`services/api/.../patient_extensions.go` | Accepted in `SMART-CARDS-WORKSPACE-ACCEPTANCE-2026-10-11.md` (`3810abc`). Multi-mode generation, QR decode, DB persistence. |
| **Odontograms (Dental Charts)** | **Accepted** | `odontogram-register.tsx`<br>`services/api/.../clinical_care.go` | Accepted in `ODONTOGRAM-WORKSPACE-ACCEPTANCE-2026-10-12.md` (`adcecd2`). 11 conditions, custom legends, 409 concurrency, PDF. |
| **Bed Management** | **Partial** | `bed-management-workspace.tsx`<br>`services/api/.../bed_management.go` | **Incomplete Integration & Mock Fallbacks:** UI calls 9 live endpoints (`beds`, `bed-types`, `bed-occupancy/report`, etc.) but masks empty responses with static fixtures (`initialWards`, `initialBedAssigns`, `initialBeds`, `initialBedTypes`), executes local-only mutations after API errors ("Preserved locally"), and lacks delete persistence. |
| **Inpatient (IPD) Care** | **Missing / Partial** | `encounter-register.tsx`<br>`connected-clinical.tsx` | All 8 IPD subtabs collapse to generic encounter table. No child models for diagnoses, consultant instructions, operations, charges, or bills. |
| **Outpatient (OPD) Care** | **Partial** | `encounter-register.tsx`<br>`connected-clinical.tsx` | OPD revisit history, doctor pricing, and outpatient diagnoses/timelines collapse into generic encounter notes. |
| **Prescriptions Workspace** | **Partial** | `prescriptions-workspace.tsx`<br>`services/api/.../clinical_care.go` | `prescriptions-workspace.tsx` uses 100% hardcoded mock prescriptions (`INITIAL_PRESCRIPTIONS`). Rich clinical headers and medicine lines not saved. |
| **Pharmacy & Medicines** | **Partial** | `medicines-workspace.tsx`<br>`services/api/.../pharmacy.go` | `medicines-workspace.tsx` has 6 mock tabs (`useState<PurchaseMedicineRow>`). Zero API calls. Go backend masters exist but disconnected from UI. |
| **Billing & Finance** | **Partial** | `billing-workspace.tsx`<br>`services/api/.../billing.go` | Accounts/invoices work in `connected-billing.tsx`, but `billing-workspace.tsx` (bills, advance payments, manual payments, payroll) is 100% mock data. |
| **Attendance & Rostering** | **Partial** | `attendance-workspace.tsx`<br>`services/api/.../attendance.go` | Shifts & assignments connected, but leave requests, correction requests, reports, and manual attendance remain client-side previews (`initial`). |
| **Blood Bank** | **Partial** | `blood-bank-workspace.tsx`<br>`services/api/.../pharmacy_blood_bank.go` | 100% hardcoded mock data in UI (`INITIAL_BLOOD_BANKS`, `INITIAL_DONORS`). Go `/v1/blood-bank` backend routes exist but disconnected. |
| **Diagnostics (Pathology/Radiology)** | **Partial** | `pathology-workspace.tsx`<br>`diagnosis-workspace.tsx` | UI uses mock arrays (`ParameterItem`, `INITIAL_CATEGORIES`). Radiology falls back to `LegacyScreen`. Report parameters and orders unpersisted. |
| **Patient Demographics & Admissions** | **Partial** | `patients-workspace.tsx`<br>`services/api/.../patient_profile.go` | `/modules/patients` renders mock `PatientsWorkspace`. Real API only wired on `/live-patients`. Admissions tab disconnected from `ipd_admission_details`. |
| **Appointments & Queues** | **Not verified** | `connected-scheduling.tsx`<br>`services/api/.../scheduling.go` | Basic booking functional, but queue displays, calendar actions, payment linkages, and patient self-service lack full parity verification. |
| **Front Office** | **Partial** | `front-office-workspace.tsx`<br>`services/api/.../front_office.go` | `front-office-workspace.tsx` uses `seedFallback()` with mock data. Zero API calls despite existing Go postgres/httpapi implementations. |
| **Live Consultations** | **Partial** | `live-consultation-workspace.tsx`<br>`services/api/.../live_consultation.go` | UI has hardcoded mock data (`ConsultationRow[]`). Backend routes exist but disconnected from frontend. |
| **Vital Reports & Documents** | **Missing** | `legacy-screen.tsx`<br>`services/api/.../diagnostic_reports.go` | Birth, death, investigation, and operation reports fall back to `LegacyScreen` with local React state and no DB persistence. |
| **Ambulances & Calls** | **Not verified** | `services-workspace.tsx`<br>`services/api/.../ambulance.go` | Connected to `/api/hms/ambulances`, but lacks module-level regression suite, access matrix checks, and lifecycle verification. |

---

## 3. Deep-Dive Gap Analysis Across Modules

### 3.1 Bed Management (`bed-status`, `bed-assigns`, `beds`, `bed-types`)
- **Legacy Contract:**
  - `BedController`, `BedTypeController`, `BedAssignController`, `BedStatusController`.
  - Bed types: unique name, description, active status.
  - Beds: unique business identifier/name, bed type association, standard charge (minor currency), ward/room assignment, maintenance flag, real-time availability.
  - Bed assignments: patient admission link, bed link, assign date, discharge date, notes, active status. Assigning marks bed unavailable; discharging releases bed.
- **Current State:**
  - `apps/web/src/components/bed-management-workspace.tsx` is 2,922 lines long and interacts with 9 existing backend endpoints (`beds?page=1`, `bed-types?page=1`, `bed-occupancy/report`, `patients?page=1`, `encounters?kind=ipd&page=1`, `POST bed-types`, `PATCH bed-types/{id}`, `POST beds`, `POST bed-assignments`).
  - **Incomplete Integration & Mock Fallback Defects:**
    1. *Mock Fallbacks Masking Database State:* Static arrays (`initialWards`, `initialBedAssigns`, `initialBeds`, `initialBedTypes`) initialize state. When API queries return empty sets (`length === 0`), the UI falls back to hardcoded fixtures instead of displaying empty states.
    2. *Local-Only Mutation on Error:* In `handleSaveBed` and `handleSaveBedAssign`, API errors are caught, but handlers silently append fabricated records (`BED...`, `ba-...`) to local React state ("Preserved locally") and dismiss modals.
    3. *Name-Based Matching & Fabricated Identifiers:* Patient assignment relies on string comparison (`givenName + ' ' + familyName`), encounter matching scans by `patientName`, and assigns generated `"HMS..."` numbers without persistent encounter validation.
    4. *Unbacked Delete Operations:* `handleDeleteConfirm` executes entirely in-memory against React state without calling backend deletion endpoints.
    5. *Missing Assignment Listing Query:* `bed-assigns` tab never queries top-level bed assignments from the API on mount, leaving historical assignments purely mock.
- **Backend Endpoint Mapping & Discrepancies:**
  - Existing Endpoints:
    - `GET /v1/bed-types` & `POST /v1/bed-types` / `PATCH /v1/bed-types/{id}` (`clinical.go`, `bed_types.go`).
    - `GET /v1/beds` & `POST /v1/beds` / `PATCH /v1/beds/{id}` (state toggle only in `bed_management.go`).
    - `GET /v1/bed-occupancy/report` (`clinical_care.go`).
    - `POST /v1/bed-assignments` (`clinical_care.go`, limited to transfers).
    - `POST /v1/encounters/{id}/transfer` & `GET /v1/encounters/{id}/bed-history` (`bed_management.go`).
  - Missing Endpoints & Guards:
    - `DELETE /v1/bed-types/{id}` with in-use protection (prevent deletion if beds reference type).
    - `DELETE /v1/beds/{id}` with in-use protection (prevent deletion if bed is occupied).
    - `GET /v1/bed-assignments` (top-level listing with patient/encounter joins, search, and pagination).
    - Full `PUT/PATCH /v1/beds/{id}` (updating name, type, charge, ward, version check).
    - Assignment discharge endpoint to release occupied beds and record discharge timestamps.
  - **Verdict:** High-impact functional disconnect. Ready for immediate cutover, endpoint completion, and hardening.

---

### 3.2 Clinical Encounters & Inpatient/Outpatient Care (IPD / OPD)
- **Legacy Contract:**
  - IPD requires: Case number, admission date, doctor, bed assignment, package/insurance references, and admission notes.
  - Sub-resources:
    - Diagnoses: Report type, report date, clinical notes, document attachment.
    - Consultant Register: Applied date, instruction date, consultant doctor, detailed clinical orders.
    - Prescriptions: Grouped header/footer, line-item medications with dosage, frequency, days, instructions.
    - Charges: Charge category, code, standard rate, quantity, applied rate snapshot.
    - Payments: Immediate patient payments against inpatient ledger.
    - Timelines: Events, descriptions, visibility flag to patient (`visible_to_person`).
    - Final Bill / Discharge: Bed duration computation, charges sum, payment credit reconciliation, tax/discount, atomic discharge releasing bed.
- **Current State:**
  - In `apps/web/src/app/(hospital)/modules/[slug]/page.tsx`, 8 IPD slugs and 3 OPD slugs are routed to `<EncounterRegister kind="ipd" />` or `<EncounterRegister kind="opd" />`.
  - `connected-clinical.tsx` renders a single generic note editor (`/api/hms/encounters/{id}/notes`) and a simple discharge summary text field.
  - Child detail records are completely unmodeled in PostgreSQL and absent from the UI.
- **Verdict:** Largest clinical architectural gap, but heavily dependent on Bed Management, Diagnostics, and Prescriptions.

---

### 3.3 Medicines & Pharmacy
- **Legacy Contract:**
  - Medicine categories and brands catalog.
  - Medicines master with generic composition, side effects, buying price, selling price, and min stock alert.
  - Purchases: Supplier, purchase number, invoice date, batch number, expiry date, purchase price, MRP, quantity, tax, payment status. Transactionally increments batch inventory.
  - Medicine Bills: Outpatient and walk-in sales, barcode scanning, batch expiry validation, stock decrement, receipt printing.
- **Current State:**
  - `apps/web/src/components/medicines-workspace.tsx` has 6 subtabs, all operating over hardcoded fixtures (`useState<PurchaseMedicineRow[]>`, `useState<MedicineBillRow[]>`).
  - No API integration exists in the component.
- **Verdict:** High operational and financial impact; requires transactional batch ledger implementation.

---

### 3.4 Billing, Finance & Payroll
- **Legacy Contract:**
  - Accounts and patient invoices.
  - Standalone IPD bills distinct from general invoices.
  - Manual bill payments and advance payment receipts (numbered deposit receipts with refund/allocation policy).
  - Payroll: Employee salary structures, basic salary, allowances, deductions, net salary calculation, status (paid/unpaid), pay slip generation.
  - Expenses & Incomes: Expense/income heads, receipts, payment modes.
- **Current State:**
  - Basic invoices work in `connected-billing.tsx`.
  - However, `/modules/bills`, `/modules/advance-payments`, `/modules/employee-payrolls`, `/modules/expenses`, and `/modules/incomes` map to `BillingWorkspace`, which relies entirely on canonical mock data (`canonicalManualPayments`, `canonicalAdvancePayments`, etc.).
- **Verdict:** Financial integrity gap.

---

### 3.5 Attendance & Staff Duty Rostering
- **Legacy / Reference Contract:**
  - Shift definitions: Name, code, start/end times, grace minutes, break minutes, default shift.
  - Duty assignments: Staff assignment to shifts with effective date ranges.
  - Attendance punches / records: Check-in, check-out, late in, early out, overtime, break duration.
  - Leave requests: Casual/annual/sick/emergency leaves, date ranges, approval lifecycle.
  - Correction requests: Missing punch adjustment requests and supervisor approvals.
- **Current State:**
  - Shifts and duty assignments are wired to Go API (`/api/hms/attendance/shifts`, `/api/hms/attendance/assignments`).
  - Leave requests, attendance corrections, daily reports, and manual attendance entries remain mock previews in `attendance-workspace.tsx`.
- **Verdict:** Mid-tier operational module; core shift engine is functional.

---

### 3.6 Diagnostics: Pathology, Radiology & Patient Diagnosis Tests
- **Legacy Contract:**
  - Test categories, parameter units, reference ranges by age/gender.
  - Pathology & radiology tests with standard pricing.
  - Patient diagnosis orders and multi-parameter test report generation.
- **Current State:**
  - `pathology-workspace.tsx` and `diagnosis-workspace.tsx` use hardcoded initial state.
  - Radiology routes fall back to `LegacyScreen`.
  - Backend tables in migration `009`, `016`, `024`, `039` exist but lack complete UI bindings.
- **Verdict:** Moderate clinical complexity.

---

### 3.7 Front Office
- **Legacy Contract:**
  - Call logs, visitor logs, postal dispatches/receives, website enquiries, and patient complaints.
- **Current State:**
  - `front-office-workspace.tsx` defines form modals and state, but populates tables via `seedFallback()` mock arrays.
  - Backend postgres/httpapi implementations exist in `front_office.go` and migration `032`.
- **Verdict:** Low clinical risk; administrative workflow.

---

## 4. Risk and Dependency Ranking

To determine the optimal implementation order, all unresolved module families are evaluated across three dimensions:
1. **Clinical & Operational Risk:** Impact on patient care, safety, and physical facility operations.
2. **Workflow Impact:** Frequency of daily usage by hospital medical and administrative staff.
3. **Architectural Dependency:** Whether this module is a prerequisite for other modules.

| Rank | Module Family | Clinical Risk | Workflow Impact | Dependency Status | Priority Justification |
| :---: | :--- | :---: | :---: | :---: | :--- |
| **1** | **Bed Management** | **High** | **Critical** | **Foundation for IPD** | **Immediate Next Target.** Physical hospital capacity constraint. Prerequisite for inpatient care, room allocation, nursing, and stay billing. Backend tables exist; UI is 100% mock. |
| **2** | **Clinical Encounters (IPD / OPD Details)** | **Critical** | **Critical** | Depends on Beds & Diagnosis | Central clinical medical record. Massive scope; requires Bed Management to be live and verified first. |
| **3** | **Medicines & Pharmacy** | **High** | **High** | Standalone / Clinical link | Dispensation safety and inventory value. Large scope requiring batch ledger transactions. |
| **4** | **General Prescriptions Workspace** | **High** | **High** | Depends on Pharmacy & Doctors | Clinical medication safety. Relies on medicines catalog for drug selection. |
| **5** | **Billing, Advances & Payroll** | **High** | **High** | Depends on IPD & Services | Financial reconciliation. Core invoices work; advance payments and bills need attention. |
| **6** | **Attendance (Leaves & Daily Punches)** | **Medium** | **Medium** | Standalone | Shifts are done; remaining work covers staff leave approvals and daily time tracking. |
| **7** | **Diagnostics (Pathology & Radiology)** | **Medium** | **Medium** | Depends on Patients/Doctors | Laboratory test catalog and patient results. Backend mostly ready. |
| **8** | **Patient Demographics & Admissions** | **Medium** | **Medium** | Foundation | Core patient API works; requires UI consolidation for `/modules/patients`. |
| **9** | **Blood Bank** | **Medium** | **Medium** | Standalone | Blood donor/issue balances. Backend ready; UI requires wiring. |
| **10** | **Front Office** | **Low** | **Low** | Standalone | Administrative logs (visitors, calls, postals). Backend exists. |

---

## 5. Recommended Next Module: Bed Management Workspace

### 5.1 Rationale
1. **Foundation of Inpatient Hospital Systems:** A hospital cannot admit inpatients, perform surgeries, or calculate inpatient stay charges without an accurate, real-time record of physical beds, wards, and occupancy.
2. **Eliminates Deceptive Mock Fallbacks & Local-Only Mutations:** While `BedManagementWorkspace` currently makes 9 API calls, it masks empty database tables with hardcoded fixtures (`initialWards`, `initialBedAssigns`, `initialBeds`, `initialBedTypes`), updates local React state on failed saves ("Preserved locally"), matches patients by names instead of persistent IDs, and executes deletions purely in memory. Replacing these fallbacks with true server-backed state brings genuine operational parity.
3. **Backend Infrastructure Already Built:** PostgreSQL tables (`bed`, `bed_type`, `bed_transfer`) and Go store implementations (`bed_management.go`, `bed_types.go`) already exist. Connecting and hardening this module delivers immediate, tangible parity with high confidence.
4. **Unblocks the Roadmap:** Stabilizing Bed Management is the mandatory stepping stone before tackling the massive Inpatient Clinical Care (IPD) module.

---

### 5.2 Bounded Implementation Scope

The Bed Management parity task shall encompass the following deliverables:

#### Scope Item 1: Bed Types Management (`/modules/bed-types`)
- Full CRUD against `/api/hms/bed-types`.
- Unique bed type name validation, description, active status toggle.
- In-use deletion protection: Prevent deletion of a bed type if active beds are assigned to it.
- Optimistic concurrency protection using `version`.

#### Scope Item 2: Beds Catalog & Maintenance (`/modules/beds`)
- Full CRUD against `/api/hms/beds`.
- Bed name, bed type selection, ward/room assignment, standard charge in minor currency (ETB cents).
- Maintenance state toggle (`is_maintenance`): Maintenance beds cannot be assigned to patients.
- In-use deletion protection: Beds with active patient occupancy cannot be deleted.

#### Scope Item 3: Bed Assignments & Transfers (`/modules/bed-assigns`)
- Assign available bed to an active admission/encounter.
- Atomic state update: Assigning a bed marks `available = false` in a database transaction.
- Discharging or unassigning a bed records discharge timestamp and marks `available = true`.
- Bed transfer support: Atomically moves patient from source bed to target bed, releasing source bed and reserving target bed.

#### Scope Item 4: Real-Time Bed Status Board (`/modules/bed-status`)
- Grouped ward visualization rendering live bed tiles by ward (`ICU`, `NICU`, `VIP`, `General`, `Private`).
- Real-time indicator for Available, Occupied, and Maintenance beds.
- Occupied bed card displays admitted patient name, gender, phone, and admission timestamp loaded from live database.
- Bed Occupancy aggregate summary: Total beds, occupied count, available count, and occupancy percentage loaded from `/api/hms/bed-occupancy/report`.

---

### 5.3 Acceptance Criteria & Verification Matrix

| Verification Gate | Scenario | Assertions |
| :--- | :--- | :--- |
| **Database Parity (Go)** | `TestBed_CRUDAndUniqueConstraints` | Create bed type, create bed with charge; verify unique name constraints and schema versioning. |
| **Concurrency & Integrity (Go)** | `TestBed_ConcurrentAssignmentRace` | Two simultaneous requests attempting to assign the same available bed: exactly one succeeds; the second receives HTTP 409 or 422 conflict. |
| **In-Use Protection (Go)** | `TestBed_DeleteProtectionWhenOccupied` | Attempt to delete an occupied bed fails with descriptive error; attempt to delete a bed type with existing beds fails. |
| **UI Live Persistence (Playwright)** | `Journey 1: Bed Type & Bed Creation` | Create new bed type and bed via UI; reload page in a brand new browser session; confirm database records persist. |
| **UI Status Board (Playwright)** | `Journey 2: Bed Status Board` | Verify ward grouping renders live database beds; verify tile status color changes from green (Available) to red (Occupied) upon assignment. |
| **UI Transfer & Release (Playwright)** | `Journey 3: Bed Transfer & Discharge` | Perform bed transfer; verify previous bed immediately returns to Available and new bed becomes Occupied. |

---

## 6. Audit Sign-Off

- **Audit Completion:** Completed and verified on branch `audit/remaining-module-parity`.
- **Target Branch State:** Clean working directory based on `main` at `adcecd2`.
- **Next Milestone:** Formulate execution plan for **Bed Management Workspace Parity** and proceed with implementation.
