# Clinical Encounters & IPD/OPD Detail Registers Audit Report

**Date:** 2026-10-13  
**Baseline Commit:** `053b8cf` on `main`  
**Working Branch:** `feat/clinical-encounters-workspace-parity`  
**Scope:** Complete Inpatient (IPD) & Outpatient (OPD) Encounter Lifecycle, Detail Registers, Patient-Scoped Access, Ownership, Concurrency Protection, and Bed Management Integration.

---

## 1. Executive Summary

Following the full acceptance, migration (066), and verified live deployment of the **Bed Management Workspace** (`053b8cf`), this audit evaluates the current state of **Clinical Encounters and IPD/OPD Detail Registers** across legacy contracts, PostgreSQL schemas, Go HTTP/store services, and Next.js frontend interfaces.

In the legacy HMS architecture, inpatient admissions (`ipd_patient_departments`) and outpatient visits (`opd_patient_departments`) anchor all clinical workflows. Beyond basic admission and appointment scheduling, each encounter maintains rich child registers:
1. **Admission & Demographics:** Vitals intake (height, weight, BP), guardian contact, medical package, insurance carrier, and policy details.
2. **Bed Management Integration:** Real-time bed assignment, ward location, standard room rates, bed transfers, and atomic release on discharge.
3. **Consultant Register & Care Team:** Multi-disciplinary clinical staffing, delegated doctors and nurses, role titles, and assignment timestamps.
4. **Clinical Diagnoses:** ICD-10 codification, provisional vs. differential vs. final categorization, and status progression.
5. **Procedures & Operations:** Clinical/surgical interventions, anesthesia type, clinical findings, complications, and operating clinician.
6. **Clinical Notes & Timelines:** Immutable, cryptographically attributed clinical progress notes, patient visibility toggles, and document attachments.
7. **Billing & Financial Clearance:** Inpatient stay bed-day aggregates, doctor consultation charges, procedure fee totals, invoice linkage, and pre-discharge financial clearance.
8. **Discharge Lifecycle & Summaries:** Condition at discharge, hospital course summary, discharge medications, follow-up advice, and atomic vacancy release.
9. **OPD Follow-Up Rostering:** Scheduled revisit dates, outpatient doctor pricing (`doctor_opd_charge`), and attendance status.

### Current Audit Verdict
While core database tables exist across migrations `001`, `003`, `010`, `011`, `012`, `038`, and `063–066`, the current user-facing application exhibits significant gaps:
- **Routing Collapse:** Ten distinct sidebar subtab routes (`ipd-patient-departments`, `ipd-diagnosis`, `ipd-consultant-registers`, `ipd-charges`, `ipd-payments`, `ipd-bills`, `ipd-timelines`, `opd-patient-departments`, `opd-diagnosis`, `opd-timelines`) collapse to a single generic `<EncounterRegister>` listing, ignoring subtab parameters.
- **Rudimentary Encounter Detail View:** `<EncounterDetails>` in `connected-clinical.tsx` only renders symptoms, raw intake key-value pairs, a basic clinical note box, and a discharge button. It completely lacks interactive child registers for diagnoses, care teams, procedures, admission details, billing clearance, bed history/transfers, and discharge summaries.
- **API Proxy Blocking:** The Next.js API proxy (`/api/hms/[...path]/route.ts`) rejects valid subresource paths such as `/encounters/{id}/diagnoses`, `/encounters/{id}/procedures`, `/encounters/{id}/attachments`, `/encounters/{id}/admission-details`, `/encounters/{id}/billing`, and length-4 sub-paths like `/billing/clearance`.
- **Missing Single Encounter Query Endpoint:** The Go backend lacks `GET /v1/encounters/{id}`, preventing direct link navigation, detail refresh, or deep-linking into individual admissions.
- **Discharge State Discrepancy & Bed Release Omission:** In `clinical.go`, the `Discharge` endpoint discharges the encounter and closes open `bed_assignment` rows, but omits updating `hospital_bed.state = 'ready'` in the bed catalog, leaving discharged beds incorrectly marked as occupied on the live bed status board. Furthermore, `Discharge` enforces `e.doctor_id = a.ID`, preventing administrative medical supervisors (`role = 'admin'`) from authorizing clinical discharges.

---

## 2. Legacy Contract & Route Specification Analysis

The legacy application (`LEGACY-ACTIONS.md`, `LEGACY-WORKFLOWS.md`, `MODEL-RULES.md`, `source-evidence.json`) establishes explicit behavioral contracts across IPD and OPD registers:

### 2.1 Inpatient Care (IPD)
- **Controller/Repository:** `IpdPatientDepartmentController`, `IpdPatientDepartmentRepository`.
- **Fields:** `patient_id` (required), `doctor_id` (required), `case_id` (required), `admission_date` (required), `bed_type_id` (required), `bed_id` (required), `weight` (nullable numeric <= 200), `height` (nullable numeric), `bp` (nullable numeric <= 200), `notes`, `symptoms`, `discharge` (0: active, 1: discharged).
- **Sub-Actions:**
  - `ipd-consultant-register`: Store, edit, update, delete consultant instruction registers.
  - `ipd-diagnosis`: Store, edit, update, delete diagnoses with optional report attachments.
  - `ipd-charges`: Store, edit, update, delete itemized hospital charges (`charge_type_id`, `charge_category_id`, `applied_charge`).
  - `ipd-timelines`: Store, edit, update, delete chronological events with document upload.
  - `ipd-payments`: Store and verify payments across modes (`Cash`, `Cheque`, `Stripe`, etc.).
  - `ipd-bills`: Bill calculation totaling bed charges, doctor fees, operations, and charges.
  - `ipd-discharge-patient/{id}/pdf`: Generate clinical discharge summary slip.

### 2.2 Outpatient Care (OPD)
- **Controller/Repository:** `OpdPatientDepartmentController`, `OpdPatientDepartmentRepository`.
- **Fields:** `patient_id` (required), `doctor_id` (required), `case_id` (optional/required depending on first visit), `appointment_date` (required), `standard_charge` (required, in ETB minor currency), `payment_mode` (1: Cash, 2: Cheque), `weight`, `height`, `bp`, `symptoms`, `notes`, `is_old_patient` (boolean).
- **Sub-Actions:**
  - `get-doctor-opd-charge`: Live retrieval of doctor's standard OPD charge.
  - `opd-diagnosis`: Store, edit, update, delete outpatient diagnosis records.
  - `opd-timelines`: Store, edit, update, delete outpatient clinical timelines.
  - `opd_follow_up`: Schedule follow-up revisit dates, physician attribution, and attendance tracking.

---

## 3. Database Schema & Migration Review

### 3.1 Encounter Master (`encounter` table)
- **Table Definition:** Migration `001_baseline_schema.sql` and `003_clinical.sql`.
- **Columns:**
  - `id uuid PRIMARY KEY DEFAULT gen_random_uuid()`
  - `number bigint NOT NULL` (sequential admission/visit identifier)
  - `kind text NOT NULL CHECK (kind IN ('ipd', 'opd'))`
  - `case_id uuid NOT NULL REFERENCES patient_case(id)`
  - `patient_id uuid NOT NULL REFERENCES patient(id)`
  - `doctor_id text NOT NULL REFERENCES "user"(id)`
  - `bed_id uuid REFERENCES hospital_bed(id)`
  - `admitted_at timestamptz NOT NULL`
  - `status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'discharged'))`
  - `discharged_at timestamptz`
  - `discharge_summary text NOT NULL DEFAULT ''`
  - `symptoms text NOT NULL DEFAULT ''`
  - `bed_charge_minor bigint NOT NULL DEFAULT 0`
  - `version integer NOT NULL DEFAULT 1`
  - `created_by text NOT NULL REFERENCES "user"(id)`
  - `request_key text NOT NULL`
  - `created_at timestamptz NOT NULL DEFAULT now()`
  - `admission_bed_id uuid`
  - `intake jsonb NOT NULL DEFAULT '{}'`
- **Key Constraints:**
  - `encounter_check`: `CHECK ((kind = 'ipd') OR (kind = 'opd' AND bed_id IS NULL))` (as repaired in migration `063`).
  - `encounter_check1`: `CHECK (((status = 'active' AND discharged_at IS NULL) OR (status = 'discharged' AND discharged_at >= admitted_at AND length(discharge_summary) > 0)))`.
  - `encounter_created_by_request_key_key`: `UNIQUE (created_by, request_key)`.
  - Partial Unique Index `one_active_inpatient_stay`: `UNIQUE (patient_id) WHERE kind = 'ipd' AND status = 'active'`.

### 3.2 Child Register Schemas
1. **`encounter_care_team` (Consultants / Staff):**
   - Columns: `id`, `encounter_id`, `staff_id`, `role_title`, `assigned_at`, `revoked_at`, `assigned_by`, `notes`.
   - Index: `idx_encounter_care_team(encounter_id)`.
2. **`encounter_diagnosis` (Diagnoses):**
   - Columns: `id`, `encounter_id`, `icd10_code`, `description`, `category` (`provisional`, `final`, `differential`), `status` (`active`, `resolved`, `ruled_out`), `diagnosed_by`, `diagnosed_at`, `created_at`.
   - Index: `idx_encounter_diagnosis(encounter_id)`.
3. **`encounter_procedure` (Operations / Procedures):**
   - Columns: `id`, `encounter_id`, `name`, `description`, `performed_by`, `performed_at`, `anesthesia_type`, `findings`, `complications`, `created_at`.
   - Index: `idx_encounter_procedure(encounter_id)`.
4. **`encounter_attachment` (Timelines & Documents):**
   - Columns: `id`, `encounter_id`, `title`, `file_url`, `file_type`, `file_size_bytes`, `uploaded_by`, `uploaded_at`.
   - Index: `idx_encounter_attachment(encounter_id)`.
5. **`ipd_admission_details` (Admission Package, Insurance, Guardian):**
   - Columns: `encounter_id PRIMARY KEY`, `package_name`, `package_charge_minor`, `insurance_policy_number`, `insurance_provider`, `guardian_name`, `guardian_relation`, `guardian_phone`, `guardian_address`, `created_at`, `updated_at`.
6. **`encounter_billing` (Stay Charges & Clearance):**
   - Columns: `encounter_id PRIMARY KEY`, `bed_days`, `bed_total_minor`, `doctor_fee_minor`, `procedure_fee_minor`, `other_charges_minor`, `total_minor`, `invoice_id REFERENCES invoice(id)`, `financial_clearance`, `cleared_by`, `cleared_at`, `waiver_reason`, `created_at`, `updated_at`.
7. **`discharge_summary` (Structured Discharge Record):**
   - Columns: `encounter_id PRIMARY KEY`, `admission_diagnosis`, `discharge_diagnosis`, `condition_at_discharge`, `hospital_course`, `surgical_procedures`, `discharge_medications`, `follow_up_advice`, `follow_up_date`, `signed_by`, `signed_at`, `created_at`, `updated_at`.
8. **`opd_follow_up` (Outpatient Revisits):**
   - Columns: `id`, `encounter_id`, `patient_id`, `doctor_id`, `follow_up_date`, `notes`, `status` (`scheduled`, `attended`, `cancelled`), `created_at`.

---

## 4. API Endpoints & Backend Implementation Audit

### 4.1 Implemented & Functional Backend Endpoints
- `GET /v1/encounters?kind=ipd|opd&page=X&pageSize=Y`: Scoped listing returning encounters with joined patient name, doctor name, and bed name.
- `POST /v1/encounters`: Admission handler verifying doctor active status, patient case ownership, locking bed (`SELECT FOR UPDATE`), creating admission `bed_event`, and recording audit logs.
- `POST /v1/encounters/{id}/notes` & `GET /v1/encounters/{id}/notes`: Cryptographic signing and retrieval of clinical progress notes.
- `POST /v1/encounters/{id}/transfer`: Bed transfer releasing source bed, allocating target bed, and updating version.
- `GET /v1/encounters/{id}/diagnoses` & `POST /v1/encounters/{id}/diagnoses`: Diagnoses CRUD.
- `GET /v1/encounters/{id}/care-team` & `POST /v1/encounters/{id}/care-team`: Care team member assignment.
- `GET /v1/encounters/{id}/procedures` & `POST /v1/encounters/{id}/procedures`: Procedure logs.
- `GET /v1/encounters/{id}/admission-details` & `PUT /v1/encounters/{id}/admission-details`: Admission package & insurance details.
- `GET /v1/encounters/{id}/billing` & `PUT /v1/encounters/{id}/billing`: Encounter billing totals.
- `POST /v1/encounters/{id}/billing/clearance`: Financial clearance approval.
- `GET /v1/patients/{patientId}/follow-ups` & `POST /v1/patients/{patientId}/follow-ups`: OPD follow-up management.

### 4.2 Backend Deficiencies & Security Gaps
1. **Missing `GET /v1/encounters/{id}`:**
   - There is no single-encounter retrieval route in `clinical.go` or `clinical_care.go`.
   - **Impact:** Direct page refresh, deep links to `/modules/ipd-patient-departments?id=...`, or modal detail reloads cannot fetch fresh encounter state from the server.
2. **Missing `PATCH/PUT /v1/encounters/{id}`:**
   - There is no endpoint to edit an active encounter's symptoms, intake vitals, notes, or doctor assignment.
3. **Discharge Bed State Mutation Omission:**
   - In `clinical.go:371–382`:
     ```go
     if out.BedID != "" {
         var dummy string
         _ = tx.QueryRow(ctx, `SELECT id FROM hospital_bed WHERE id=$1 FOR UPDATE`, out.BedID).Scan(&dummy)
         if _, err = tx.Exec(ctx, `UPDATE bed_assignment SET assigned_to=$2 WHERE encounter_id=$1 AND assigned_to IS NULL`, id, now); err != nil {
             return out, err
         }
     }
     ```
     The query closes `bed_assignment`, but **fails to update `hospital_bed.state = 'ready'`**.
   - Consequently, when an encounter is discharged via `/v1/encounters/{id}/discharge`, the physical bed remains stuck in `occupied` state in the bed catalog!
4. **Discharge Role Access Limitation:**
   - In `clinical.go:361`:
     ```go
     scanEncounter(tx.QueryRow(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE e.id=$1 AND e.doctor_id=$2 FOR UPDATE OF e`, id, a.ID))
     ```
     This strictly forbids administrators (`a.Role == "admin"`) from executing clinical discharges, even when supervising care or performing administrative cutovers.
5. **Care Team Member Revocation Missing from HTTP Layer:**
   - Store has `RevokeCareTeamMember`, but HTTP router in `clinical_care.go` has no `DELETE /v1/encounters/{id}/care-team/{memberId}` route.

---

## 5. Frontend UI Deficiencies & Mock Fallbacks

### 5.1 Route Mapping Flattening
In `apps/web/src/app/(hospital)/modules/[slug]/page.tsx`:
```tsx
if (
  [
    "ipd-patient-departments",
    "ipd-diagnosis",
    "ipd-consultant-registers",
    "ipd-prescriptions",
    "ipd-charges",
    "ipd-payments",
    "ipd-bills",
    "ipd-timelines",
  ].includes(slug)
)
  return <EncounterRegister key={slug} kind="ipd" />;
if (
  ["opd-patient-departments", "opd-diagnosis", "opd-timelines"].includes(slug)
)
  return <EncounterRegister key={slug} kind="opd" />;
```
- Clicking **IPD Diagnoses** or **Consultant Register** in the sidebar loads the standard encounter list rather than contextualizing or opening the selected child register.

### 5.2 Disconnected / Rudimentary Encounter Details
- `<EncounterDetails>` in `connected-clinical.tsx` renders only basic header metadata, clinical notes, and a discharge summary form.
- The 7 core clinical registers already supported in the database are completely unrendered in the UI:
  1. **Diagnoses Tab:** Empty / unrendered.
  2. **Consultant Register Tab:** Empty / unrendered.
  3. **Procedures & Operations Tab:** Empty / unrendered.
  4. **Bed History & Transfers Tab:** Empty / unrendered.
  5. **Admission Packages & Insurance Tab:** Empty / unrendered.
  6. **Billing & Financial Clearance Tab:** Empty / unrendered.
  7. **OPD Follow-Up Visits Tab:** Empty / unrendered.

### 5.3 Next.js API Proxy Blocklist
In `apps/web/src/app/api/hms/[...path]/route.ts`:
- `isAllowedPath` allows only:
  `["notes", "discharge", "care-team", "consultations", "ipd-details", "billing-summary", "clearance", "transfer", "bed-assignments", "bed-history"]`.
- The following live Go endpoints are blocked with HTTP 404:
  - `encounters/{id}/diagnoses`
  - `encounters/{id}/procedures`
  - `encounters/{id}/attachments`
  - `encounters/{id}/admission-details`
  - `encounters/{id}/billing`
  - `encounters/{id}/discharge-summary`
  - All 4-segment subroutes: `encounters/{id}/billing/clearance`, `encounters/{id}/billing/invoice`, `encounters/{id}/care-team/{memberId}`.

---

## 6. Implementation Scope & Plan

To achieve full parity, eliminate all mock fallbacks, and satisfy the user acceptance criteria:

### 6.1 Database Schema Corrections (Migration 067)
- Author migration `067_clinical_encounter_parity_and_discharge_repair.sql`:
  1. Add missing index `idx_encounter_kind_status` on `encounter(kind, status, admitted_at)`.
  2. Ensure `encounter_diagnosis`, `encounter_procedure`, `encounter_care_team`, `ipd_admission_details`, `encounter_billing`, `discharge_summary`, and `opd_follow_up` have consistent foreign keys and timestamps.
  3. Repair any historical discharged IPD encounters where `hospital_bed.state` remained marked `occupied`, releasing beds whose active assignment has ended.
  4. Verify that historical clinical records remain immutable and append-only.

### 6.2 Backend Go API Hardening
1. **Add `GET /v1/encounters/{id}`:**
   - Retrieve a single encounter by ID with full joined patient, doctor, and bed data.
   - Enforce canonical actor scoping (`encounterScope`: admin, receptionist, assigned doctor, care team doctor/nurse, or portal patient).
2. **Add `PATCH /v1/encounters/{id}`:**
   - Update encounter symptoms, notes, intake vitals, and doctor reassignment with positive `version` validation (returns HTTP 409 on stale edits).
3. **Atomic Bed Release in Encounter Discharge:**
   - In `Store.Discharge`, update `hospital_bed.state = 'ready'` and `available = true` atomically in the transaction alongside `encounter.bed_id = NULL` and `bed_assignment.assigned_to = now`.
   - Broaden discharge authorization to allow administrators (`a.Role == "admin" || e.doctor_id == a.ID`).
4. **Care Team Member Revocation:**
   - Wire `DELETE /v1/encounters/{id}/care-team/{memberId}` to `Store.RevokeCareTeamMember`.
5. **Encounter Diagnostics & Procedures:**
   - Ensure `DELETE` or `PATCH` on diagnoses and procedures are supported where appropriate.

### 6.3 Next.js API Proxy (`route.ts`) Expansion
- Expand `isAllowedPath` to permit:
  - `encounters/{id}` (GET, PATCH)
  - `encounters/{id}/diagnoses` (GET, POST, DELETE)
  - `encounters/{id}/procedures` (GET, POST)
  - `encounters/{id}/attachments` (GET, POST)
  - `encounters/{id}/admission-details` (GET, PUT)
  - `encounters/{id}/billing` (GET, PUT)
  - `encounters/{id}/billing/clearance` (POST)
  - `encounters/{id}/billing/invoice` (POST)
  - `encounters/{id}/discharge-summary` (GET, PUT)
  - `encounters/{id}/care-team/{memberId}` (DELETE)

### 6.4 Comprehensive Encounter Workspace UI (`encounter-workspace.tsx`)
1. **Dynamic Navigation & Subtab Context:**
   - Support sidebar slug routing so navigating to `/modules/ipd-diagnosis` opens the encounter workspace pre-focused on Diagnoses, while `/modules/ipd-consultant-registers` pre-focuses on Care Team / Consultants.
2. **Server-Backed Encounter Lists:**
   - Full pagination, search filtering across number/patient/doctor, and status filters (`All`, `Active`, `Discharged`).
3. **Tabbed Clinical Detail Register:**
   - **Tab 1: Overview & Vitals:** Patient demographics, admission date, doctor, bed info, height, weight, BP, notes.
   - **Tab 2: Diagnoses Register:** Live list of ICD-10 diagnoses with provisional/final categories, active/resolved status, and modal form to record new diagnoses.
   - **Tab 3: Consultant Register (Care Team):** Live list of assigned doctors and nurses, role titles, assignment dates, and modal to delegate care team members.
   - **Tab 4: Bed History & Transfers (IPD only):** Live bed assignment details, historical bed transitions, standard room charges in ETB, and interactive Bed Transfer modal (integrated with Bed Management!).
   - **Tab 5: Procedures & Operations:** Operative/clinical procedure records, anesthesia type, findings, complications, and new procedure modal.
   - **Tab 6: Clinical Notes & Timelines:** Immutable doctor signed notes, document attachments, and upload support.
   - **Tab 7: Admission & Insurance:** Inpatient admission package, insurance policy number, provider, and guardian details.
   - **Tab 8: Billing & Financial Clearance:** Itemized charges, total balance, financial clearance status badge, and clearance waiver action.
   - **Tab 9: Discharge Management:** Structured discharge summary form (condition at discharge, instructions, medications), and atomic discharge button releasing the bed.
   - **Tab 10: Follow-Ups (OPD only):** Outpatient revisit schedule and doctor consultation fees.
4. **UX & Accessibility Standards:**
   - Full Amharic localization across all tabs, badges, and modals.
   - Dark mode with WCAG AA contrast compliance (>= 4.5:1).
   - Zero mock fallbacks or hardcoded fake arrays.

---

## 7. Acceptance Criteria Matrix

| Verification Gate | Test Scenario | Expected Outcome |
| :--- | :--- | :--- |
| **Schema & Migrations** | Fresh install & upgrade through migration runner (066 -> 067) | Migration 067 applies cleanly, records in `schema_migration`, no syntax/constraint errors. |
| **Single Encounter API** | `GET /v1/encounters/{id}` with valid session | Returns full encounter JSON with patient, doctor, case, and bed data. |
| **Unauthorized Access Rejection** | Doctor B requests Doctor A's encounter details | Returns HTTP 403 Forbidden; audit event logged. |
| **Patient-Level Access Scope** | Clinician Doctor B requests encounter where Doctor A is primary | Forbidden unless Doctor B is added to `encounter_care_team`. |
| **Concurrency & Stale Edits** | Simultaneous update or discharge with stale `version` | Returns HTTP 409 Conflict / Stale version rejection. |
| **Atomic Bed Release** | Discharge an active IPD patient with assigned bed | Encounter status -> `discharged`, `bed_id` -> `NULL`, `bed_assignment.assigned_to` -> timestamp, `hospital_bed.state` -> `'ready'`, `available` -> `true`. |
| **Bed Transfer Integration** | Transfer IPD patient from Bed 1 to Bed 2 | Old bed returns to available; new bed becomes occupied; `bed_event` recorded. |
| **Diagnoses Persistence** | Add provisional ICD-10 diagnosis via UI; refresh session | Record persists in PostgreSQL `encounter_diagnosis` and re-renders in fresh session. |
| **Care Team Delegation** | Add nurse/doctor to care team; verify delegation | Record persists in `encounter_care_team`; new member gains read access to encounter. |
| **Amharic & Dark Mode** | Switch to Amharic (`am`) and toggle dark theme | All register tabs, table headers, and badges translated; contrast meets WCAG AA. |

---

## 8. Verification Plan & Test Strategy

1. **Go Unit & PostgreSQL Regression Tests:**
   - Expand `services/api/internal/adapters/postgres/clinical_test.go` and `clinical_care_test.go` with dedicated test suites covering:
     - `TestEncounter_GetByID_SecurityScope`
     - `TestEncounter_Update_ConcurrencyProtection`
     - `TestEncounter_Discharge_AtomicBedRelease`
     - `TestEncounter_Diagnoses_Lifecycle`
     - `TestEncounter_CareTeam_DelegationAndRevocation`
     - `TestEncounter_Procedures_Persistence`
     - `TestEncounter_AdmissionDetails_And_Billing`
2. **Browser QA Journeys (`scripts/verify-clinical-encounters-workspace.mjs`):**
   - **Journey 1:** IPD Patient Admission with live Bed selection, verifying occupied bed board update.
   - **Journey 2:** Detail View Navigation across all tabs (Diagnoses, Care Team, Procedures, Admission Details, Billing).
   - **Journey 3:** Diagnoses and Care Team mutations with fresh-session persistence verification.
   - **Journey 4:** Bed Transfer from Bed A to Bed B with atomic state verification.
   - **Journey 5:** Discharge Patient with atomic bed release back to Available state.
   - **Journey 6:** OPD Visit registration, doctor fee rendering, and follow-up scheduling.
   - **Journey 7:** Security scope rejection (unauthorized doctor denied access).
   - **Journey 8:** Amharic localization and dark mode WCAG AA contrast verification.
3. **Isolated Integration Runner:**
   - Wire `--clinical-encounters-workspace` into `scripts/verify-connected-isolated.mjs` for fail-closed isolated schema testing.
