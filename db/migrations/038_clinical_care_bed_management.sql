-- Migration 038: Clinical Care and Bed Management
-- Bed assignments & occupancy, care-team delegation, diagnoses, procedures,
-- attachments, admission details/packages/insurance, encounter billing & financial clearance,
-- discharge summaries, follow-ups, referrals, and odontogram dental chart.

-- 1. Bed assignment history and occupancy tracking
CREATE TABLE bed_assignment (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  bed_id       uuid NOT NULL REFERENCES hospital_bed(id),
  encounter_id uuid NOT NULL REFERENCES encounter(id) ON DELETE CASCADE,
  patient_id   uuid NOT NULL REFERENCES patient(id),
  assigned_from timestamptz NOT NULL DEFAULT clock_timestamp(),
  assigned_to   timestamptz,
  notes        text NOT NULL DEFAULT '' CHECK (length(notes) <= 1000),
  assigned_by  text NOT NULL REFERENCES "user"(id),
  created_at   timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_bed_assignment_encounter ON bed_assignment(encounter_id);
CREATE INDEX idx_bed_assignment_bed ON bed_assignment(bed_id, assigned_from);
CREATE INDEX idx_bed_assignment_patient ON bed_assignment(patient_id);

-- 2. Care team delegation & observation visibility
CREATE TABLE encounter_care_team (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  encounter_id uuid NOT NULL REFERENCES encounter(id) ON DELETE CASCADE,
  staff_id     text NOT NULL REFERENCES "user"(id),
  role_title   varchar(80) NOT NULL CHECK (length(trim(role_title)) >= 1),
  assigned_at  timestamptz NOT NULL DEFAULT clock_timestamp(),
  revoked_at   timestamptz,
  assigned_by  text NOT NULL REFERENCES "user"(id),
  notes        text NOT NULL DEFAULT '' CHECK (length(notes) <= 500)
);
CREATE INDEX idx_encounter_care_team ON encounter_care_team(encounter_id, staff_id);

ALTER TABLE clinical_note ADD COLUMN IF NOT EXISTS patient_visible boolean NOT NULL DEFAULT false;

-- 3. Consultation registers: Diagnoses, procedures, attachments
CREATE TABLE encounter_diagnosis (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  encounter_id uuid NOT NULL REFERENCES encounter(id) ON DELETE CASCADE,
  icd10_code   varchar(32) NOT NULL DEFAULT '',
  description  text NOT NULL CHECK (length(trim(description)) >= 1 AND length(description) <= 2000),
  category     varchar(32) NOT NULL DEFAULT 'provisional' CHECK (category IN ('provisional', 'final', 'differential')),
  status       varchar(32) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'resolved', 'ruled_out')),
  diagnosed_by text NOT NULL REFERENCES "user"(id),
  diagnosed_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  created_at   timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_encounter_diagnosis ON encounter_diagnosis(encounter_id);

CREATE TABLE encounter_procedure (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  encounter_id     uuid NOT NULL REFERENCES encounter(id) ON DELETE CASCADE,
  name             varchar(160) NOT NULL CHECK (length(trim(name)) >= 1),
  description      text NOT NULL DEFAULT '' CHECK (length(description) <= 2000),
  performed_by     text NOT NULL REFERENCES "user"(id),
  performed_at     timestamptz NOT NULL DEFAULT clock_timestamp(),
  anesthesia_type  text NOT NULL DEFAULT '' CHECK (length(anesthesia_type) <= 100),
  findings         text NOT NULL DEFAULT '' CHECK (length(findings) <= 4000),
  complications    text NOT NULL DEFAULT '' CHECK (length(complications) <= 2000),
  created_at       timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_encounter_procedure ON encounter_procedure(encounter_id);

CREATE TABLE encounter_attachment (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  encounter_id    uuid NOT NULL REFERENCES encounter(id) ON DELETE CASCADE,
  title           varchar(160) NOT NULL CHECK (length(trim(title)) >= 1),
  file_url        text NOT NULL CHECK (length(trim(file_url)) >= 1 AND length(file_url) <= 1000),
  file_type       varchar(64) NOT NULL DEFAULT 'document',
  file_size_bytes bigint NOT NULL DEFAULT 0 CHECK (file_size_bytes >= 0),
  uploaded_by     text NOT NULL REFERENCES "user"(id),
  uploaded_at     timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_encounter_attachment ON encounter_attachment(encounter_id);

-- 4. IPD admission packages, insurance, and guardians
CREATE TABLE ipd_admission_details (
  encounter_id             uuid PRIMARY KEY REFERENCES encounter(id) ON DELETE CASCADE,
  package_name             text NOT NULL DEFAULT '' CHECK (length(package_name) <= 160),
  package_charge_minor     bigint NOT NULL DEFAULT 0 CHECK (package_charge_minor >= 0),
  insurance_policy_number  text NOT NULL DEFAULT '' CHECK (length(insurance_policy_number) <= 100),
  insurance_provider       text NOT NULL DEFAULT '' CHECK (length(insurance_provider) <= 160),
  guardian_name            text NOT NULL DEFAULT '' CHECK (length(guardian_name) <= 150),
  guardian_relation        text NOT NULL DEFAULT '' CHECK (length(guardian_relation) <= 100),
  guardian_phone           text NOT NULL DEFAULT '' CHECK (length(guardian_phone) <= 64),
  guardian_address         text NOT NULL DEFAULT '' CHECK (length(guardian_address) <= 500),
  created_at               timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at               timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- 5. Encounter billing, invoice linkage, and financial clearance
ALTER TABLE service_invoice_link DROP CONSTRAINT IF EXISTS service_invoice_link_source_type_check;
ALTER TABLE service_invoice_link ADD CONSTRAINT service_invoice_link_source_type_check
  CHECK(source_type IN ('service', 'operation', 'ambulance', 'blood_issue', 'appointment', 'encounter'));

CREATE TABLE encounter_billing (
  encounter_id         uuid PRIMARY KEY REFERENCES encounter(id) ON DELETE CASCADE,
  bed_days             integer NOT NULL DEFAULT 1 CHECK (bed_days >= 0),
  bed_total_minor      bigint NOT NULL DEFAULT 0 CHECK (bed_total_minor >= 0),
  doctor_fee_minor     bigint NOT NULL DEFAULT 0 CHECK (doctor_fee_minor >= 0),
  procedure_fee_minor  bigint NOT NULL DEFAULT 0 CHECK (procedure_fee_minor >= 0),
  other_charges_minor  bigint NOT NULL DEFAULT 0 CHECK (other_charges_minor >= 0),
  total_minor          bigint NOT NULL DEFAULT 0 CHECK (total_minor >= 0),
  invoice_id           uuid REFERENCES invoice(id) ON DELETE SET NULL,
  financial_clearance  boolean NOT NULL DEFAULT false,
  cleared_by           text REFERENCES "user"(id),
  cleared_at           timestamptz,
  waiver_reason        text NOT NULL DEFAULT '' CHECK (length(waiver_reason) <= 500),
  created_at           timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at           timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_encounter_billing_invoice ON encounter_billing(invoice_id);

-- 6. Structured discharge summaries
CREATE TABLE discharge_summary (
  encounter_id           uuid PRIMARY KEY REFERENCES encounter(id) ON DELETE CASCADE,
  admission_diagnosis    text NOT NULL DEFAULT '' CHECK (length(admission_diagnosis) <= 2000),
  discharge_diagnosis    text NOT NULL DEFAULT '' CHECK (length(discharge_diagnosis) <= 2000),
  condition_at_discharge varchar(64) NOT NULL DEFAULT 'improved'
    CHECK (condition_at_discharge IN ('recovered', 'improved', 'unchanged', 'referred', 'deceased')),
  hospital_course        text NOT NULL DEFAULT '' CHECK (length(hospital_course) <= 10000),
  surgical_procedures    text NOT NULL DEFAULT '' CHECK (length(surgical_procedures) <= 4000),
  discharge_medications  text NOT NULL DEFAULT '' CHECK (length(discharge_medications) <= 4000),
  follow_up_advice       text NOT NULL DEFAULT '' CHECK (length(follow_up_advice) <= 2000),
  follow_up_date         date,
  signed_by              text NOT NULL REFERENCES "user"(id),
  signed_at              timestamptz NOT NULL DEFAULT clock_timestamp(),
  created_at             timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at             timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- 7. OPD follow-ups and patient referrals
CREATE TABLE opd_follow_up (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  encounter_id   uuid NOT NULL REFERENCES encounter(id) ON DELETE CASCADE,
  patient_id     uuid NOT NULL REFERENCES patient(id),
  doctor_id      text NOT NULL REFERENCES "user"(id),
  follow_up_date date NOT NULL,
  notes          text NOT NULL DEFAULT '' CHECK (length(notes) <= 2000),
  status         varchar(32) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'attended', 'cancelled')),
  created_at     timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_opd_follow_up_patient ON opd_follow_up(patient_id, follow_up_date);
CREATE INDEX idx_opd_follow_up_doctor ON opd_follow_up(doctor_id, follow_up_date);

CREATE TABLE patient_referral (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  encounter_id      uuid NOT NULL REFERENCES encounter(id) ON DELETE CASCADE,
  patient_id        uuid NOT NULL REFERENCES patient(id),
  referral_type     varchar(32) NOT NULL DEFAULT 'outward' CHECK (referral_type IN ('inward', 'outward')),
  external_facility text NOT NULL CHECK (length(trim(external_facility)) >= 1 AND length(external_facility) <= 200),
  department        text NOT NULL DEFAULT '' CHECK (length(department) <= 100),
  reason            text NOT NULL CHECK (length(trim(reason)) >= 1 AND length(reason) <= 2000),
  referred_by       text NOT NULL REFERENCES "user"(id),
  referred_at       timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_patient_referral ON patient_referral(patient_id, referred_at);

-- 8. Odontogram (Dental Chart)
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
