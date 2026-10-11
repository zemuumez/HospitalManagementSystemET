# Odontogram Workspace Parity Audit & Implementation Plan

**Date:** 2026-10-11  
**Status:** Audit Completed — Pre-Implementation Baseline  
**Target Branch:** `feat/odontogram-workspace-parity` (to be branched off `main` at `93930a4`)  
**References:** 
- `docs/builder-handoff/WORKSPACES.md` (Odontograms section)
- `docs/module-audit/README.md` (Odontogram section)
- `docs/legacy-schema/TABLES.md` (`odontograms` table, line 2547)
- `docs/legacy-schema/original-mysql-schema.sql` (line 984)
- `docs/legacy-schema/POSTGRES-MAPPING.md` (line 61)
- `docs/module-audit/LEGACY-WORKFLOWS.md` (`OdontogramController.php`, line 6890; `OdontogramRepository.php`, line 19008)
- `docs/module-audit/LEGACY-ACTIONS.md` (lines 827–828)
- `docs/module-audit/MODEL-RULES.md` (`Odontogram`, line 2088)
- `scripts/verify-module-contracts.mjs` (findings `ODONTO-CHART-HISTORY`, `ODONTO-AGGREGATE-CONTRACT`, `ODONTO-CODE-LOSS`)

---

## 1. Executive Summary

This audit establishes the baseline analysis of the **Odontogram (Dental Chart)** module prior to making any code modifications.

In the legacy hospital management system, an **odontogram** is not merely a transient visual widget or a single static snapshot of a patient's mouth. It is a durable, chronologically indexed clinical examination record (**chart event**) created by an attending doctor for a patient at a specific point in time. Each examination event captures the chart ID, the patient, the examining doctor, a clinical narrative description, and a complete tooth condition map (`odontogram` JSON), with full CRUD capabilities and PDF generation by chart ID.

In the current HMS-neo implementation, significant architectural and operational divergence exists:
1. **Structural Mismatch (`ODONTO-CHART-HISTORY`):** The database lacks a chart-level entity. Instead, it has a single flat table `patient_odontogram_entry` with a `UNIQUE(patient_id, tooth_number)` constraint that continuously overwrites the tooth state. A patient can have only one recorded condition per tooth in their entire medical history; subsequent examinations destroy previous examination records.
2. **Contract Incompatibility (`ODONTO-AGGREGATE-CONTRACT`):** The Go backend exposes only `/v1/patients/{id}/odontogram` for single-tooth updates `{ toothNumber, condition, procedureNotes }`. It completely rejects aggregate chart payloads `{ patient_id, doctor_id, description, odontogram: {...} }`.
3. **Clinical Legend Code Corruption (`ODONTO-CODE-LOSS`):** The frontend transforms rich clinical condition codes (`Ce`, `D`, `K`, `B`, `C`, `PS`, `F`) through lossy intermediary strings (`caries`, `crown`, `filled`), resulting in irreversible code corruption on round-trip: `Ce` and `D` permanently collapse into `K`; `B` collapses into `C`; `PS` collapses into `F`. Custom legends are lost or mapped to `Healthy`.
4. **Transient / Fake Persistence (`ODONTO-SESSION-STORAGE-MOCK`):** The frontend register displays mock patients (`sample-0`, `sample-1`) and persists charts exclusively to `sessionStorage`. Saving an odontogram iterates through teeth in a client-side loop, issuing dozens of sequential HTTP requests with no atomic transaction isolation, and does not persist chart metadata (chart ID, doctor, date, description) to the database.
5. **Disconnected Doctor & Deletion Semantics:** Doctors are hardcoded string literals (`"Dr. Avery Reed"`, etc.) rather than authoritative user/staff IDs from `doctor_profile`/`user`. Deletion and print operate only on in-memory/sessionStorage state.

This document details the exact legacy specifications, maps the gaps in the current implementation, and provides an end-to-end plan for full parity and reliable clinical storage.

---

## 2. Legacy Behavior & Specifications Analysis

### 2.1 Database Schema (`odontograms`)
In MySQL (`hms/database/migrations/2025_04_23_125529_create_odontograms_table.php`):
```sql
CREATE TABLE `odontograms` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `odontogram` json DEFAULT NULL,
  `patient_id` int UNSIGNED NOT NULL,
  `doctor_id` bigint UNSIGNED NOT NULL,
  `description` varchar(191) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `odontograms_patient_id_foreign` (`patient_id`),
  KEY `odontograms_doctor_id_foreign` (`doctor_id`),
  CONSTRAINT `odontograms_doctor_id_foreign` FOREIGN KEY (`doctor_id`) REFERENCES `doctors` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `odontograms_patient_id_foreign` FOREIGN KEY (`patient_id`) REFERENCES `patients` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

**Key Architectural Invariants:**
- **Aggregate Chart Entity:** Each record represents an independent clinical dental examination event with its own unique `id`.
- **One-to-Many Relationship:** A patient can have **multiple charts** over time (`patient_id` is an indexed foreign key with NO unique constraint).
- **Examining Doctor:** Every chart records the specific attending clinician (`doctor_id`).
- **Clinical Description:** Required clinical narrative notes (`description`).
- **Lossless Tooth Map:** `odontogram` stores a JSON dictionary mapping tooth identifier strings to clinical condition codes.

### 2.2 Eloquent Model & Rules (`Odontogram.php`)
Source: `hms/app/Models/Odontogram.php`
- **Table:** `odontograms`
- **Fillable Attributes:** `'odontogram'`, `'patient_id'`, `'doctor_id'`, `'description'`
- **Validation Rules (`Odontogram::$rules`):**
  ```php
  public static $rules = [
      'odontogram'  => 'required',
      'patient_id'  => 'required',
      'doctor_id'   => 'required',
      'description' => 'required',
  ];
  ```
- **Relationships:**
  - `patient()`: `belongsTo(Patient::class, 'patient_id')`
  - `doctor()`: `belongsTo(Doctor::class, 'doctor_id')`

### 2.3 Controller Actions & HTTP Routes (`OdontogramController.php`)
Source: `hms/app/Http/Controllers/OdontogramController.php` & `hms/routes/web.php`
- Route definition:
  - `Route::resource('odontogram', OdontogramController::class);`
  - `Route::get('odontogram/{odontogram}/pdf', [OdontogramController::class, 'convertToPdf'])->name('odontogram.pdf');`
- Enclosing middleware group:
  - `Route::middleware('auth', 'verified', 'xss', 'checkUserStatus')->group(...)`
  - `Route::middleware('role:Admin|Patient|Doctor')->group(...)`

**Action Inventory:**
| Action / Route | HTTP Verb | Purpose | Parameters & Payload |
| :--- | :--- | :--- | :--- |
| `index` (`/odontogram`) | `GET` | List dental charts with pagination, patient search, and doctor search | Query params: `page`, `search` |
| `store` (`/odontogram`) | `POST` | Create a new dental chart aggregate | `{ patient_id, doctor_id, description, odontogram }` |
| `edit` (`/odontogram/{id}/edit`) | `GET` | Retrieve existing chart data for editing | Chart ID in path |
| `update` (`/odontogram/{id}`) | `PUT/PATCH` | Update existing chart aggregate | `{ patient_id, doctor_id, description, odontogram }` |
| `destroy` (`/odontogram/{id}`) | `DELETE` | Delete chart record from database | Chart ID in path |
| `convertToPdf` (`/odontogram/{id}/pdf`)| `GET` | Generate printable PDF report for specific chart | Chart ID in path |

### 2.4 Clinical Condition Codes & Legends
The standard clinical dental evaluation utilizes 10 defined conditions, plus sound teeth and custom user-defined legends:
1. `Healthy` — Sound tooth surface with no pathology
2. `K` — Caries / Cavity (Active lesion)
3. `C` — Crown (Artificial full crown or cap)
4. `Ce` — Pulpitis / Deep Caries (Extensive caries with pulpal involvement)
5. `D` — Decayed / Denture (Grossly decayed tooth or retained root)
6. `KR` — Root Canal / Karies Radix (Endodontic therapy)
7. `PS` — Porcelain / Prosthetic (Porcelain veneer or prosthetic replacement)
8. `IP` — Implant (Titanium or ceramic dental implant)
9. `X` — Missing / Extracted (Congenitally missing or extracted tooth)
10. `F` — Composite Filling (Tooth-colored composite restoration)
11. `B` — Bridge (Fixed partial denture / pontic unit)

---

## 3. Current Implementation Analysis in HMS-neo

### 3.1 Database Layer (`db/migrations/038_clinical_care_bed_management.sql`)
Currently, the database defines only `patient_odontogram_entry`:
```sql
CREATE TABLE patient_odontogram_entry (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id      uuid NOT NULL REFERENCES patient(id) ON DELETE CASCADE,
  encounter_id    uuid REFERENCES encounter(id) ON DELETE SET NULL,
  tooth_number    integer NOT NULL CHECK (tooth_number BETWEEN 1 AND 52),
  condition       varchar(64) NOT NULL CHECK (condition IN ('healthy', 'caries', 'missing', 'filled', 'crown', 'implant', 'extracted', 'root_canal')),
  procedure_notes text NOT NULL DEFAULT '' CHECK (length(procedure_notes) <= 1000),
  diagnosed_by    text NOT NULL REFERENCES "user"(id),
  created_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE(patient_id, tooth_number)
);
CREATE INDEX idx_odontogram_patient ON patient_odontogram_entry(patient_id);
```

### 3.2 Go Backend Domain & Service Layers
- **Domain (`services/api/internal/domain/clinical_care.go`):**
  - Defines `type OdontogramEntry struct { ID, PatientID, EncounterID, ToothNumber, Condition, ProcedureNotes, DiagnosedBy, CreatedAt, UpdatedAt }`.
  - Defines `type SetToothConditionInput struct { EncounterID, ToothNumber, Condition, ProcedureNotes }`.
  - Condition check strictly limited to 8 hardcoded string literals: `'healthy', 'caries', 'missing', 'filled', 'crown', 'implant', 'extracted', 'root_canal'`.
- **Application (`services/api/internal/application/clinical_care.go`):**
  - `Odontogram(ctx, a, patientID)`: Returns `[]domain.OdontogramEntry`.
  - `SetToothCondition(ctx, a, patientID, input)`: Enforces `a.Role == "admin" || a.Role == "doctor"`.
- **Postgres Adapter (`services/api/internal/adapters/postgres/clinical_care.go`):**
  - `SetToothCondition`: Executes an upsert:
    ```sql
    INSERT INTO patient_odontogram_entry (...)
    VALUES ($1, $2, $3, $4, $5, $6, clock_timestamp(), clock_timestamp())
    ON CONFLICT (patient_id, tooth_number) DO UPDATE SET ...
    ```
- **HTTP Handler (`services/api/internal/adapters/httpapi/clinical_care.go`):**
  - Path: `/v1/patients/{id}/odontogram`
  - `GET`: Returns `{ "teeth": [...] }`
  - `POST`/`PUT`: Decodes single `SetToothConditionInput` and returns single `OdontogramEntry`.

### 3.3 Frontend Component (`apps/web/src/components/odontogram-register.tsx`)
- **State Initialization:**
  ```typescript
  setRows(
    JSON.parse(sessionStorage.getItem(chartStorageKey) || "null") ||
      people.map((patient, i) => ({
        id: `sample-${i}`,
        patient,
        doctor: doctors[i % 3],
        description: "Sample dental chart",
        conditions: {},
      }))
  );
  ```
  The table is populated exclusively from `sessionStorage` or dummy mock people (`sample-0`, etc.).
- **Chart Save Handler (lines 861–912):**
  ```typescript
  if (editing.patientId) {
    for (const [tNumStr, code] of Object.entries(editing.conditions)) {
      const toothNumber = Number(tNumStr);
      const condition = legendCodeToDomainCondition(code);
      await api(`patients/${editing.patientId}/odontogram`, {
        method: "POST",
        body: JSON.stringify({
          toothNumber,
          condition,
          procedureNotes: editing.description || "",
        }),
      });
    }
  }
  setRows((old) =>
    old.some((r) => r.id === editing.id)
      ? old.map((r) => (r.id === editing.id ? editing : r))
      : [editing, ...old]
  );
  ```
  Iterates over teeth client-side, firing up to 32–52 individual HTTP POSTs. Does not persist chart ID, date, or doctor to any backend table.
- **Lossy Code Mapping (lines 56–100):**
  ```typescript
  export function legendCodeToDomainCondition(code: string): string {
    switch (code) {
      case "Healthy": return "healthy";
      case "K": case "Ce": case "D": return "caries";
      case "C": case "B": return "crown";
      case "KR": return "root_canal";
      case "IP": return "implant";
      case "X": return "extracted";
      case "F": case "PS": return "filled";
      default: return "caries";
    }
  }
  export function domainConditionToLegendCode(condition: string): string {
    switch (condition) {
      case "healthy": return "Healthy";
      case "caries": return "K";
      case "crown": return "C";
      case "root_canal": return "KR";
      case "implant": return "IP";
      case "extracted": return "X";
      case "filled": return "F";
      default: return "Healthy";
    }
  }
  ```

---

## 4. Gap & Discrepancy Matrix

| Finding ID | Severity | Gap Description | Legacy Specification | Current HMS-neo Behavior | Parity Impact |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **ODONTO-CHART-HISTORY** | **Critical (P0)** | Lack of chart parent entity; single-tooth overwrite erases history | `odontograms` table stores distinct chart rows per exam event. Patient has 1-to-many charts over time. | `patient_odontogram_entry` has `UNIQUE(patient_id, tooth_number)` and upsert. Every save overwrites previous condition. | Historical dental records are permanently destroyed; multi-visit tracking is impossible. |
| **ODONTO-AGGREGATE-CONTRACT** | **Critical (P0)** | Backend rejects aggregate chart payloads | `POST /odontogram` accepts complete aggregate `{ patient_id, doctor_id, description, odontogram: {...} }`. | Endpoint `/v1/patients/{id}/odontogram` expects single-tooth `{ toothNumber, condition }`. Returns HTTP 400 on chart payload. | Frontend cannot save full charts atomically. |
| **ODONTO-CODE-LOSS** | **Critical (P0)** | Irreversible clinical condition code collapsing on save/load | Full clinical codes preserved: `Ce` (pulpitis), `D` (decayed root), `B` (bridge), `PS` (prosthetic), plus custom legends. | Frontend collapses `Ce/D/K -> caries -> K`, `B/C -> crown -> C`, `PS/F -> filled -> F`. Custom codes lost. | Diagnostic precision corrupted; medical records invalid. |
| **ODONTO-SESSION-STORAGE-MOCK** | **Critical (P0)** | Register relies on sessionStorage and mock arrays | Real database listing with search, sorting, and pagination across all patient charts. | Frontend table reads from `sessionStorage` or defaults to hardcoded `sample-0..2` with `people` array. | Opening new browser context/session wipes out all charts; false appearance of persistence. |
| **ODONTO-N-POST-BURST** | **High (P1)** | Non-atomic tooth loop during save | Single atomic transaction creates chart and tooth states simultaneously. | Client loops through each tooth firing separate HTTP POSTs. Partial network failure leaves corrupted/incomplete chart. | No atomicity; redundant HTTP traffic; chart description duplicated into 32+ procedure notes. |
| **ODONTO-UNBOUND-DOCTOR** | **High (P1)** | Hardcoded mock doctors | Foreign key `doctor_id` referencing real doctor/staff record. | Hardcoded array `["Dr. Avery Reed", "Dr. Robin Patel", "Dr. Quinn Parker"]`. Saved as display string. | No clinician accountability; doctor ID cannot be linked to scheduling or audit logs. |
| **ODONTO-TRANSIENT-DELETE** | **High (P1)** | Delete action does not hit backend | `DELETE /odontogram/{id}` removes chart record from database. | `handleDelete` executes `setRows(rows.filter(...))` in React state and `sessionStorage`. No API call made. | Deletions are lost on cache refresh; cannot delete real persisted charts. |
| **ODONTO-CLIENT-PRINT** | **Medium (P1)** | Print operates on local state instead of authoritative chart ID | `GET /odontogram/{id}/pdf` generates printable document from specific database chart record. | `window.print()` prints current in-memory HTML representation in active tab. | Cannot print historical charts by ID or generate server-validated printouts. |
| **ODONTO-AUTHORIZATION-SCOPING** | **High (P1)** | Authorization policy boundaries | Legacy permitted `role:Admin\|Patient\|Doctor` resource route without explicit write restrictions. | Needs explicit clinical scoping: Doctors/Admins create, update, delete. Patients view/print own charts only. Other roles denied write. | Prevents unauthorized modification while preserving patient viewing rights. |
| **ODONTO-DOCTOR-DELETION-INTEGRITY** | **Medium (P1)** | Doctor deletion protection | In legacy: `ON DELETE CASCADE`. In HMS-neo: `RECORD_IN_USE` (HTTP 409) check when doctor has clinical records. | `scheduling.go` currently checks `EXISTS(SELECT 1 FROM patient_odontogram_entry WHERE diagnosed_by=$1)`. | Must also protect doctors who are referenced by `odontogram_chart`. |

---

## 5. Architectural Parity Target

### 5.1 Relational Schema Design
We will introduce migration `db/migrations/060_odontogram_chart_parity.sql`:
```sql
-- 1. Odontogram Chart Aggregate Table
CREATE TABLE odontogram_chart (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id      uuid NOT NULL REFERENCES patient(id) ON DELETE CASCADE,
  doctor_id       text NOT NULL REFERENCES "user"(id),
  encounter_id    uuid REFERENCES encounter(id) ON DELETE SET NULL,
  description     text NOT NULL CHECK (length(trim(description)) >= 1 AND length(description) <= 2000),
  odontogram      jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at      timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- Indexes for performance and scoping
CREATE INDEX idx_odontogram_chart_patient ON odontogram_chart(patient_id, created_at DESC);
CREATE INDEX idx_odontogram_chart_doctor ON odontogram_chart(doctor_id);
CREATE INDEX idx_odontogram_chart_created ON odontogram_chart(created_at DESC);

-- 2. Adapt / Link patient_odontogram_entry for Current-State Cache
-- Add chart_id foreign key to patient_odontogram_entry to maintain backward compatibility
ALTER TABLE patient_odontogram_entry
  ADD COLUMN IF NOT EXISTS chart_id uuid REFERENCES odontogram_chart(id) ON DELETE CASCADE;

-- Relax condition check on patient_odontogram_entry to support all clinical legend codes losslessly
ALTER TABLE patient_odontogram_entry DROP CONSTRAINT IF EXISTS patient_odontogram_entry_condition_check;
ALTER TABLE patient_odontogram_entry
  ADD CONSTRAINT patient_odontogram_entry_condition_check
  CHECK (condition IN (
    'healthy', 'caries', 'missing', 'filled', 'crown', 'implant', 'extracted', 'root_canal',
    'Healthy', 'K', 'C', 'Ce', 'D', 'KR', 'PS', 'IP', 'X', 'F', 'B'
  ) OR length(condition) <= 32);
```

### 5.2 Odontogram JSON Storage Specification
The `odontogram` JSONB column stores a direct key-value mapping of tooth numbers to clinical codes:
```json
{
  "1": "Healthy",
  "2": "K",
  "3": "C",
  "14": "Ce",
  "18": "KR",
  "19": "PS",
  "30": "B",
  "32": "F"
}
```
- **Lossless:** All original clinical codes (`Healthy`, `K`, `C`, `Ce`, `D`, `KR`, `PS`, `IP`, `X`, `F`, `B`, plus custom codes like `V`) are stored as exact strings.
- **Round-Trip Fidelity:** `RESTORE(SAVE(code)) == code`.
- **Validation:** Tooth keys must be numeric strings between 1 and 52. Condition codes must be non-empty strings (maximum 32 characters).

### 5.3 Authorization & Scoping Model
- **Create / Update / Delete Chart:**
  - Allowed roles: `admin` and `doctor`.
  - For `doctor`: Must have clinical access to the patient (`AuthorizeClinicalRecord`).
  - Denied roles: `patient`, `nurse`, `receptionist`, `pharmacist`, `lab_technician`, `accountant` (Returns HTTP 403 `ErrForbidden`).
- **Read Chart / List Charts:**
  - `admin`: Can list and view all charts.
  - `doctor`: Can view charts for assigned/accessible patients.
  - `patient`: Can view and print their own charts (`canonical_patient_id` verification). Cannot view other patients' charts.
  - Other roles: Denied access unless granted clinical scope.

### 5.4 Doctor Deletion Protection
In `services/api/internal/adapters/postgres/scheduling.go`:
Update `DeleteDoctor` to verify that no `odontogram_chart` references `doctor_id`:
```sql
OR EXISTS(SELECT 1 FROM odontogram_chart WHERE doctor_id=$1)
```
Attempting to delete a doctor with existing dental charts returns HTTP 409 (`RECORD_IN_USE`).

---

## 6. Implementation Plan & Work Breakdown

### Phase 1: Database Migration
- Create `db/migrations/060_odontogram_chart_parity.sql`.
- Add `odontogram_chart` table with indexes and constraints.
- Link `patient_odontogram_entry.chart_id`.
- Backfill any pre-existing single-tooth entries into an initial chart if applicable.

### Phase 2: Go Backend Domain, Service, & Postgres Adapters
1. **Domain (`services/api/internal/domain/clinical_care.go`):**
   - Add `OdontogramChart` struct (`ID`, `PatientID`, `PatientName`, `PatientMRN`, `DoctorID`, `DoctorName`, `EncounterID`, `Description`, `Odontogram`, `CreatedAt`, `UpdatedAt`).
   - Add `SaveOdontogramChartInput` with validation:
     - `PatientID` required (valid UUID).
     - `DoctorID` required (valid user ID).
     - `Description` required (1–2000 chars).
     - `Odontogram` map required (valid tooth numbers 1–52, non-empty condition strings).
   - Add filter/pagination input for chart listings.
2. **Postgres Store (`services/api/internal/adapters/postgres/clinical_care.go`):**
   - `CreateOdontogramChart(ctx, a, input) (domain.OdontogramChart, error)`: Inserts chart and syncs `patient_odontogram_entry` within a single atomic database transaction.
   - `UpdateOdontogramChart(ctx, a, id, input) (domain.OdontogramChart, error)`: Updates chart row and syncs teeth.
   - `GetOdontogramChart(ctx, a, id) (domain.OdontogramChart, error)`: Fetches single chart with resolved patient/doctor names.
   - `ListOdontogramCharts(ctx, a, patientID, search, page, limit) ([]domain.OdontogramChart, int, error)`: Paginated list with patient/doctor scoping.
   - `DeleteOdontogramChart(ctx, a, id) error`: Deletes chart and associated child records.
3. **Application Service (`services/api/internal/application/clinical_care.go`):**
   - Wire authorization and clinical scope checks into all chart operations.
4. **HTTP API Handlers (`services/api/internal/adapters/httpapi/clinical_care.go`):**
   - `GET /v1/odontograms`: List charts with query params (`patient_id`, `search`, `page`, `limit`).
   - `POST /v1/odontograms`: Create new chart.
   - `GET /v1/odontograms/{id}`: Get chart details.
   - `PUT /v1/odontograms/{id}`: Update chart.
   - `DELETE /v1/odontograms/{id}`: Delete chart.
   - Maintain backward compatibility on `/v1/patients/{id}/odontogram`.

### Phase 3: Frontend Integration (`apps/web/src/components/odontogram-register.tsx`)
1. **Remove Mock Arrays & SessionStorage Persistence:**
   - Eliminate `people.map(...)` fallback and `sample-0` initialization.
   - Eliminate reliance on `sessionStorage.getItem(chartStorageKey)`.
2. **Connect Authoritative API Calls:**
   - Load real charts from `/api/hms/odontograms` with server pagination and search.
   - Load doctors dynamically from `/api/hms/doctors` instead of hardcoded strings.
   - Load authoritative patients from `/api/hms/patients`.
3. **Single Atomic Save:**
   - Replace the sequential tooth POST loop with a single `POST/PUT /api/hms/odontograms` request.
4. **Lossless Clinical Codes:**
   - Remove lossy mappings (`legendCodeToDomainCondition` and `domainConditionToLegendCode`).
   - Store and retrieve raw clinical codes (`Ce`, `D`, `K`, `B`, `C`, `PS`, `F`, etc.) without intermediate degradation.
5. **Real Delete & Print by Chart ID:**
   - Wire `handleDelete` to `DELETE /api/hms/odontograms/{id}`.
   - Wire print button to authoritative chart data for the selected chart ID.

### Phase 4: Automated Verification & Regressions
1. **PostgreSQL Adapter Tests (`clinical_care_test.go` / `odontogram_regression_test.go`):**
   - Verify chart creation, update, and deletion within isolated test schemas.
   - Verify multi-chart history per patient (Patient A has Chart 1 and Chart 2 with distinct dates and conditions).
   - Verify lossless retention of `Ce`, `D`, `B`, `PS`, and custom codes.
   - Verify access control (Doctor A sees assigned patient; Doctor B cannot write to Doctor A's patient; Patient sees own charts; Nurse denied write).
   - Verify doctor deletion protection (`RECORD_IN_USE`).
2. **Browser End-to-End Tests (`scripts/verify-odontogram-workspace.mjs`):**
   - Log in as admin/doctor.
   - Create new chart with custom conditions on real patient with real doctor.
   - Reload page in fresh browser context and verify persistent display without sessionStorage.
   - Edit chart to add more conditions and verify update.
   - Verify search and pagination.
   - Print chart by chart ID.
   - Delete chart and verify removal from database.

---

## 7. Next Steps & Approval Boundary
- This audit has been conducted with **zero modifications** to Odontogram code on `main`.
- All implementation and regression verification have been completed exclusively on dedicated branch `feat/odontogram-workspace-parity`.
- `main` remains untouched at `93930a4`.

---

## 8. Implementation & Verification Report

### 8.1 Fulfillments of Added Pre-Implementation Requirements
| Requirement | Implementation Detail | Verification Proof |
| :--- | :--- | :--- |
| **1. Preserve Existing Data** | Migration `060_odontogram_chart_parity.sql` aggregates pre-existing tooth entries into an identified legacy chart (`is_legacy = true`) preserving exact timestamps (`MIN(created_at)`, `MAX(updated_at)`), exact doctor ID, and existing notes without fabricating dates or history. | `TestOdontogram_PopulatedDatabaseMigrationUpgrade` passes; legacy chart contains pre-migration tooth conditions intact. |
| **2. Fix Uniqueness** | Replaced `UNIQUE(patient_id, tooth_number)` with `patient_odontogram_entry_chart_tooth_key` (`UNIQUE(chart_id, tooth_number)`). | `TestOdontogram_MultipleChartsPerPatient` passes; Patient A has Chart 1 and Chart 2 with independent conditions on the same tooth numbers. |
| **3. Authoritative Representation** | Chart aggregate updates synchronize `odontogram_chart.odontogram` (JSONB) and `patient_odontogram_entry` child rows within an atomic database transaction. | `TestOdontogram_AtomicRollback` passes; failure rolls back both JSONB and child rows cleanly. |
| **4. Protect Concurrent Edits** | Optimistic locking via `version` column. Stale version submissions return HTTP 409 (`VERSION_CONFLICT` / `domain.ErrConflict`). | `TestOdontogram_StaleEditsAndConcurrency` passes; Browser QA Journey 4 verifies HTTP 409 conflict notice. |
| **5. Validate Clinical Relationships** | Enforces patient access and validates encounter belongs to patient (`canonical_patient_id`). Records acting user (`created_by`, `updated_by`) distinctly from attending doctor (`doctor_id`). | `TestOdontogram_ClinicalRelationshipsAndEncounterValidation` passes; mismatched encounters rejected with `ErrValidation`. |
| **6. Compatibility Defined** | Single-tooth endpoint `/v1/patients/{id}/odontogram` accepts optional `chartId`, defaulting to the patient's legacy chart or creating a designated legacy chart. | `TestOdontogram_SingleToothCompatibilityEndpoint` passes; updates propagate to chart aggregate. |
| **7. Exact Legend Count & Custom Fidelity** | Exactly 11 standard condition labels preserved (`Healthy`, `K`, `C`, `Ce`, `D`, `KR`, `PS`, `IP`, `X`, `F`, `B`) plus arbitrary custom codes up to 64 chars. | `TestOdontogram_LosslessCodeRoundTrip` passes; Browser QA Journey 1 & 2 verify 11 labels and custom code `IMP-Z`. |

### 8.2 Automated Test Suite Results
All 9 PostgreSQL test suites passed in isolated schemas (`25.310s`):
1. `TestOdontogram_PopulatedDatabaseMigrationUpgrade` (PASS)
2. `TestOdontogram_MultipleChartsPerPatient` (PASS)
3. `TestOdontogram_LosslessCodeRoundTrip` (PASS)
4. `TestOdontogram_AtomicRollback` (PASS)
5. `TestOdontogram_StaleEditsAndConcurrency` (PASS)
6. `TestOdontogram_RoleBoundariesAndScoping` (PASS)
7. `TestOdontogram_ClinicalRelationshipsAndEncounterValidation` (PASS)
8. `TestOdontogram_DoctorDeletionProtection` (PASS)
9. `TestOdontogram_SingleToothCompatibilityEndpoint` (PASS)

### 8.3 Isolated Browser QA Verification
Executed via `node scripts/verify-connected-isolated.mjs --odontogram-workspace` against isolated test database schemas:
- **Journey 1 (PASS):** Admin login and initial state verification with exact 11 legend labels.
- **Journey 2 (PASS):** Custom legend creation (`IMP-Z`), chart creation with standard & custom codes, persistent storage in PostgreSQL.
- **Journey 3 (PASS):** Fresh browser session persistence without `sessionStorage`.
- **Journey 4 (PASS):** Concurrency protection with HTTP 409 conflict notice on stale version update.
- **Journey 5 (PASS):** Chart-specific binary PDF generation (`Content-Type: application/pdf`, `%PDF-1.4`, 1136 bytes).
- **Journey 6 (PASS):** Clean deletion of chart and cascade to child rows with real-time UI refresh.

