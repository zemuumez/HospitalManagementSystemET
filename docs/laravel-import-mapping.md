# Legacy Laravel (Smart Hospital) to HMS ET PostgreSQL Migration & Reconciliation Plan

## 1. Overview & Architecture

This document defines the field-level data migration mapping, transformation rules, reconciliation validation, trial-run strategy, and rollback plan for migrating legacy MySQL/Laravel Smart Hospital data into the production Go/PostgreSQL Hospital Management System (HMS ET).

### Migration Pipeline Principles
1. **Zero Schema Mutation During App Execution**: All target schemas are strictly governed by migrations `001` through `040`.
2. **Idempotent Trial Runs**: Data is staged in temporary import schemas (`staging_legacy_*`) before atomic insertion into production tables within transactional boundaries.
3. **Immutable Audit Trails**: Preserves legacy IDs in an audit map table (`legacy_id_mapping`) so every imported record is fully traceable.
4. **Data Cleansing & Normalization**: Strips invalid characters, standardizes phone numbers to Ethiopian format (`+2519...` / `+2517...`), converts dates to UTC `timestamptz`, and translates monetary values to integer cents.

---

## 2. Table & Field Mapping Specification

### 2.1 Patients & Demographics
| Legacy MySQL (`patients`) | HMS ET PostgreSQL (`patient`, `patient_contact_consent`) | Transformation / Business Rules |
| :--- | :--- | :--- |
| `id` (int auto-increment) | `legacy_id_mapping.legacy_id` -> `patient.id` (UUID) | Generate deterministic or random UUID; map in `legacy_id_mapping` |
| `patient_name` | `patient.full_name` | Trim whitespace; title case |
| `mobileno` | `patient.phone` | Normalize to `+251...` E.164 standard |
| `email` | `patient.email` | Lowercase; null if empty or invalid format |
| `gender` | `patient.gender` | Standardize to `'male'`, `'female'`, `'other'` |
| `dob` | `patient.date_of_birth` | Parse `YYYY-MM-DD` |
| `blood_group` | `patient.blood_group` | Standardize: `'A+'`, `'A-'`, `'B+'`, `'B-'`, `'AB+'`, `'AB-'`, `'O+'`, `'O-'` |
| `guardian_name` | `patient_contact_consent.guardian_name` | Trimmed string |
| `guardian_relation` | `patient_contact_consent.guardian_relation` | Mapped to kinship relation |
| `guardian_phone` | `patient_contact_consent.guardian_phone` | Normalized phone |
| `address` | `patient.address` | Text normalization |

### 2.2 Staff & Users
| Legacy MySQL (`staff`, `users`) | HMS ET PostgreSQL (`"user"`, `staff_profile`) | Transformation / Business Rules |
| :--- | :--- | :--- |
| `id` | `"user".id` | Generate UUID / map ID |
| `name`, `surname` | `"user".name` | Concatenate `name + " " + surname` |
| `email` | `"user".email` | Lowercase unique |
| `role` | `"user".role` | Map: `Admin` -> `admin`, `Doctor` -> `doctor`, `Nurse` -> `nurse`, `Pharmacist` -> `pharmacist`, `Pathologist`/`Radiologist` -> `lab_technician`, `Receptionist` -> `receptionist`, `Accountant` -> `accountant` |
| `contact_no` | `staff_profile.phone` | Normalized phone |
| `designation`, `qualification` | `staff_profile.designation`, `qualification` | Text |

### 2.3 Appointments & Visits
| Legacy MySQL (`appointment`, `visit_details`) | HMS ET PostgreSQL (`appointment`, `encounter`) | Transformation / Business Rules |
| :--- | :--- | :--- |
| `id` | `appointment.id` | UUID mapping |
| `patient_id` | `appointment.patient_id` | Lookup target UUID from `legacy_id_mapping` |
| `doctor` (staff_id) | `appointment.doctor_id` | Lookup target UUID |
| `date`, `time` | `appointment.scheduled_time` | Combine and parse as UTC `timestamptz` |
| `appointment_status` | `appointment.status` | Map: `approved` -> `'confirmed'`, `pending` -> `'scheduled'`, `cancel` -> `'cancelled'` |

### 2.4 Pharmacy & Inventory
| Legacy MySQL (`pharmacy`, `medicine_bad_stock`) | HMS ET PostgreSQL (`pharmacy_inventory`, `pharmacy_batch`) | Transformation / Business Rules |
| :--- | :--- | :--- |
| `medicine_name` | `pharmacy_item.name` | Unique item catalog |
| `medicine_category_id` | `pharmacy_item.category_id` | Master data mapping |
| `batch_no` | `pharmacy_batch.batch_number` | Exact string |
| `expiry_date` | `pharmacy_batch.expiry_date` | Date format |
| `available_quantity` | `pharmacy_batch.current_quantity` | Integer count >= 0 |
| `sale_price` | `pharmacy_batch.unit_price_cents` | Convert to integer cents (`round(price * 100)`) |

### 2.5 Billing, Charges & Payments
| Legacy MySQL (`patient_charges`, `payment`) | HMS ET PostgreSQL (`invoice`, `invoice_item`, `payment`) | Transformation / Business Rules |
| :--- | :--- | :--- |
| `charge_id`, `amount` | `invoice_item.unit_price_cents`, `total_cents` | Decimal to integer cents |
| `paid_amount` | `payment.amount_cents` | Decimal to integer cents |
| `payment_mode` | `payment.payment_method` | Map: `Cash` -> `'cash'`, `Cheque` -> `'bank_transfer'`, `Online` -> `'telebirr'` / `'cbe_birr'` |
| `date` | `payment.created_at` | Parse UTC timestamp |

---

## 3. Data Integrity & Reconciliation Verification

During migration execution, the automated reconciliation script (`scripts/reconcile_import.go`) validates:
1. **Row Count Parity**: Source count minus filtered invalid/test rows equals target count exactly.
2. **Financial Balance Sums**: Total debit and credit cents in source matches target down to the cent:
   $$\sum \text{legacy\_charges} = \sum \text{target\_invoice\_items}$$
   $$\sum \text{legacy\_payments} = \sum \text{target\_payments}$$
3. **Orphan Check**: Zero target records referencing unmapped patient IDs, doctor IDs, or encounter IDs.
4. **Checksum Sampling**: 10% random sample hashing of core clinical records to ensure byte-level field fidelity.

---

## 4. Rollback & Backout Strategy

If reconciliation fails or business verification encounters critical blockers:
1. **Database Rollback**:
   - The entire migration script executes inside an explicit transaction (`BEGIN ... COMMIT`). Any error triggers an immediate `ROLLBACK`.
   - If cutover was already initiated, restore from the pre-migration snapshot taken immediately prior to maintenance window start:
     ```powershell
     .\scripts\restore.ps1 -BackupFile "backups/pre_migration_snapshot_encrypted.bin" -TargetDb "hms_prod"
     ```
2. **DNS & Service Rollback**:
   - Point DNS/reverse-proxy traffic back to the legacy server instance.
   - Verify legacy Laravel system operations and notify clinical coordinators.
