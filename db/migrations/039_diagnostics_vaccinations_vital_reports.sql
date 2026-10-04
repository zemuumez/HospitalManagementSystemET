-- Migration 039: Diagnostics Masters & Reports, Vaccinations, and Vital Reports
-- Category/Unit masters, test extensions, report attachments & portal release,
-- diagnosis templates, vaccination catalog & administration, birth, death, operation,
-- and investigation reports.

-- 1. Diagnostic category and unit masters
CREATE TABLE diagnostic_category (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        varchar(100) NOT NULL,
  kind        varchar(32) NOT NULL CHECK (kind IN ('pathology', 'radiology')),
  description text NOT NULL DEFAULT '' CHECK (length(description) <= 1000),
  created_at  timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE(kind, name)
);

CREATE TABLE diagnostic_unit (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name        varchar(50) NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '' CHECK (length(description) <= 500),
  created_at  timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- Seed initial common categories and units
INSERT INTO diagnostic_category (name, kind, description) VALUES
  ('Clinical Pathology', 'pathology', 'Routine blood, urine, and body fluid tests'),
  ('Hematology', 'pathology', 'Complete blood counts, coagulation, and blood smears'),
  ('Biochemistry', 'pathology', 'Metabolic panels, renal, liver and cardiac enzymes'),
  ('Microbiology', 'pathology', 'Cultures, gram stains, and antibiotic susceptibility'),
  ('X-Ray', 'radiology', 'General projection radiography'),
  ('Ultrasound', 'radiology', 'Abdominal, pelvic, and vascular sonography'),
  ('CT Scan', 'radiology', 'Computed tomography imaging'),
  ('MRI', 'radiology', 'Magnetic resonance imaging')
ON CONFLICT (kind, name) DO NOTHING;

INSERT INTO diagnostic_unit (name, description) VALUES
  ('mg/dL', 'Milligrams per deciliter'),
  ('g/dL', 'Grams per deciliter'),
  ('mmol/L', 'Millimoles per liter'),
  ('uIU/mL', 'Micro international units per milliliter'),
  ('cells/mcL', 'Cells per microliter'),
  ('U/L', 'Units per liter'),
  ('%', 'Percentage'),
  ('sec', 'Seconds')
ON CONFLICT (name) DO NOTHING;

-- Extend diagnostic_test
ALTER TABLE diagnostic_test
  ADD COLUMN IF NOT EXISTS sub_category text NOT NULL DEFAULT '' CHECK (length(sub_category) <= 100),
  ADD COLUMN IF NOT EXISTS preparation_instructions text NOT NULL DEFAULT '' CHECK (length(preparation_instructions) <= 2000);

-- 2. Diagnostic report files and portal release
CREATE TABLE diagnostic_report_file (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id         uuid NOT NULL REFERENCES diagnostic_order(id) ON DELETE CASCADE,
  file_name        varchar(255) NOT NULL CHECK (length(trim(file_name)) >= 1),
  file_url         text NOT NULL CHECK (length(trim(file_url)) >= 1),
  file_size_bytes  bigint NOT NULL DEFAULT 0 CHECK (file_size_bytes >= 0),
  mime_type        varchar(100) NOT NULL DEFAULT 'application/pdf',
  patient_released boolean NOT NULL DEFAULT false,
  released_at      timestamptz,
  released_by      text REFERENCES "user"(id),
  uploaded_by      text NOT NULL REFERENCES "user"(id),
  uploaded_at      timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_diag_report_order ON diagnostic_report_file(order_id);
CREATE INDEX idx_diag_report_portal ON diagnostic_report_file(order_id) WHERE patient_released = true;

-- 3. Diagnosis templates
CREATE TABLE diagnosis_template (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title            varchar(160) NOT NULL UNIQUE CHECK (length(trim(title)) >= 1),
  category         varchar(100) NOT NULL DEFAULT '',
  template_content text NOT NULL CHECK (length(trim(template_content)) >= 1),
  created_by       text NOT NULL REFERENCES "user"(id),
  created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at       timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- 4. Vaccination catalog and administration records
CREATE TABLE vaccine_catalog (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name               varchar(160) NOT NULL UNIQUE CHECK (length(trim(name)) >= 1),
  target_disease     text NOT NULL DEFAULT '' CHECK (length(target_disease) <= 500),
  recommended_doses  integer NOT NULL DEFAULT 1 CHECK (recommended_doses > 0),
  min_age_months     integer NOT NULL DEFAULT 0 CHECK (min_age_months >= 0),
  instructions       text NOT NULL DEFAULT '' CHECK (length(instructions) <= 2000),
  created_at         timestamptz NOT NULL DEFAULT clock_timestamp()
);

INSERT INTO vaccine_catalog (name, target_disease, recommended_doses, min_age_months, instructions) VALUES
  ('BCG', 'Tuberculosis', 1, 0, 'Administer at birth intradermally'),
  ('OPV (Oral Polio)', 'Poliomyelitis', 4, 0, 'Birth, 6, 10, 14 weeks oral drops'),
  ('Pentavalent (DTP-HepB-Hib)', 'Diphtheria, Tetanus, Pertussis, Hepatitis B, Haemophilus influenzae b', 3, 2, 'Intramuscular at 6, 10, 14 weeks'),
  ('Pneumococcal (PCV)', 'Streptococcus pneumoniae diseases', 3, 2, 'Intramuscular at 6, 10, 14 weeks'),
  ('Rotavirus', 'Rotavirus gastroenteritis', 2, 2, 'Oral at 6 and 10 weeks'),
  ('Measles', 'Measles', 2, 9, 'Subcutaneous at 9 and 15 months'),
  ('Tetanus Toxoid (TT)', 'Tetanus prevention in pregnancy / adults', 5, 180, 'Intramuscular scheduled protection')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE patient_vaccination (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id       uuid NOT NULL REFERENCES patient(id) ON DELETE CASCADE,
  vaccine_id       uuid NOT NULL REFERENCES vaccine_catalog(id),
  dose_number      integer NOT NULL CHECK (dose_number > 0),
  lot_number       varchar(100) NOT NULL DEFAULT '' CHECK (length(lot_number) <= 100),
  expiry_date      date,
  administered_at  timestamptz NOT NULL DEFAULT clock_timestamp(),
  administered_by  text NOT NULL REFERENCES "user"(id),
  next_due_date    date,
  notes            text NOT NULL DEFAULT '' CHECK (length(notes) <= 1000),
  created_at       timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_patient_vaccination ON patient_vaccination(patient_id, administered_at DESC);

-- 5. Birth, death, operation, and investigation vital reports
CREATE TABLE birth_report (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  report_number    varchar(64) NOT NULL UNIQUE,
  child_name       varchar(160) NOT NULL CHECK (length(trim(child_name)) >= 1),
  gender           varchar(16) NOT NULL CHECK (gender IN ('male', 'female', 'other')),
  birth_date       timestamptz NOT NULL,
  weight_kg        numeric(5,2) NOT NULL CHECK (weight_kg > 0 AND weight_kg <= 15),
  mother_patient_id uuid REFERENCES patient(id) ON DELETE SET NULL,
  mother_name      varchar(160) NOT NULL CHECK (length(trim(mother_name)) >= 1),
  father_name      varchar(160) NOT NULL DEFAULT '',
  delivered_by     text NOT NULL REFERENCES "user"(id),
  notes            text NOT NULL DEFAULT '' CHECK (length(notes) <= 2000),
  created_by       text NOT NULL REFERENCES "user"(id),
  created_at       timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at       timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_birth_report_date ON birth_report(birth_date DESC);

CREATE TABLE death_report (
  id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  report_number          varchar(64) NOT NULL UNIQUE,
  patient_id             uuid NOT NULL REFERENCES patient(id) ON DELETE CASCADE,
  death_date             timestamptz NOT NULL,
  cause_of_death         text NOT NULL CHECK (length(trim(cause_of_death)) >= 1 AND length(cause_of_death) <= 2000),
  certified_by           text NOT NULL REFERENCES "user"(id),
  guardian_acknowledged  varchar(160) NOT NULL DEFAULT '',
  notes                  text NOT NULL DEFAULT '' CHECK (length(notes) <= 2000),
  created_by             text NOT NULL REFERENCES "user"(id),
  created_at             timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at             timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_death_report_patient ON death_report(patient_id);

CREATE TABLE operation_report (
  id                       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  report_number            varchar(64) NOT NULL UNIQUE,
  encounter_id             uuid NOT NULL REFERENCES encounter(id) ON DELETE CASCADE,
  patient_id               uuid NOT NULL REFERENCES patient(id) ON DELETE CASCADE,
  operation_name           varchar(160) NOT NULL CHECK (length(trim(operation_name)) >= 1),
  surgeon_id               text NOT NULL REFERENCES "user"(id),
  assistant_surgeon        text NOT NULL DEFAULT '' CHECK (length(assistant_surgeon) <= 160),
  anesthetist              text NOT NULL DEFAULT '' CHECK (length(anesthetist) <= 160),
  anesthesia_type          text NOT NULL DEFAULT '' CHECK (length(anesthesia_type) <= 100),
  operation_date           timestamptz NOT NULL,
  pre_operative_diagnosis  text NOT NULL DEFAULT '' CHECK (length(pre_operative_diagnosis) <= 2000),
  post_operative_diagnosis text NOT NULL DEFAULT '' CHECK (length(post_operative_diagnosis) <= 2000),
  procedure_technique      text NOT NULL DEFAULT '' CHECK (length(procedure_technique) <= 5000),
  findings                 text NOT NULL DEFAULT '' CHECK (length(findings) <= 4000),
  complications            text NOT NULL DEFAULT '' CHECK (length(complications) <= 2000),
  created_by               text NOT NULL REFERENCES "user"(id),
  created_at               timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at               timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_op_report_patient ON operation_report(patient_id);
CREATE INDEX idx_op_report_encounter ON operation_report(encounter_id);

CREATE TABLE investigation_report (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  report_number      varchar(64) NOT NULL UNIQUE,
  patient_id         uuid NOT NULL REFERENCES patient(id) ON DELETE CASCADE,
  encounter_id       uuid REFERENCES encounter(id) ON DELETE SET NULL,
  title              varchar(160) NOT NULL CHECK (length(trim(title)) >= 1),
  investigation_type varchar(64) NOT NULL CHECK (length(trim(investigation_type)) >= 1),
  clinical_notes     text NOT NULL DEFAULT '' CHECK (length(clinical_notes) <= 4000),
  conclusion         text NOT NULL DEFAULT '' CHECK (length(conclusion) <= 2000),
  investigated_by    text NOT NULL REFERENCES "user"(id),
  created_at         timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_inv_report_patient ON investigation_report(patient_id);
