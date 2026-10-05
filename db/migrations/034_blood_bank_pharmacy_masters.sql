-- 034_blood_bank_pharmacy_masters.sql
-- Medicine Categories, Brands, Blood Bank Inventory, Donors, Issues, and Structured Prescriptions

-- 1. Medicine Categories & Brands
CREATE TABLE medicine_category (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(160) NOT NULL UNIQUE,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE medicine_brand (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(160) NOT NULL UNIQUE,
  email varchar(191) NOT NULL DEFAULT '',
  phone varchar(64) NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- 2. Blood Bank
CREATE TABLE blood_bank (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  blood_group varchar(16) NOT NULL UNIQUE CHECK(blood_group IN ('A+', 'A-', 'B+', 'B-', 'AB+', 'AB-', 'O+', 'O-')),
  remained_bags integer NOT NULL DEFAULT 0 CHECK(remained_bags >= 0),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- Pre-seed the 8 standard blood groups
INSERT INTO blood_bank (blood_group, remained_bags) VALUES
  ('A+', 0),
  ('A-', 0),
  ('B+', 0),
  ('B-', 0),
  ('AB+', 0),
  ('AB-', 0),
  ('O+', 0),
  ('O-', 0)
ON CONFLICT (blood_group) DO NOTHING;

CREATE TABLE blood_donor (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(191) NOT NULL,
  age integer NOT NULL CHECK(age BETWEEN 16 AND 80),
  gender smallint NOT NULL CHECK(gender IN (0, 1)),
  blood_group varchar(16) NOT NULL CHECK(blood_group IN ('A+', 'A-', 'B+', 'B-', 'AB+', 'AB-', 'O+', 'O-')),
  last_donate_date timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_blood_donor_group ON blood_donor(blood_group);

CREATE TABLE blood_donation (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  donor_id uuid NOT NULL REFERENCES blood_donor(id) ON DELETE CASCADE,
  bags integer NOT NULL DEFAULT 1 CHECK(bags > 0),
  donation_date timestamptz NOT NULL DEFAULT clock_timestamp(),
  recorded_by text NOT NULL REFERENCES "user"(id),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE blood_issue (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  issue_date timestamptz NOT NULL,
  doctor_id text NOT NULL,
  donor_id uuid REFERENCES blood_donor(id) ON DELETE SET NULL,
  patient_id text NOT NULL,
  blood_group varchar(16) NOT NULL CHECK(blood_group IN ('A+', 'A-', 'B+', 'B-', 'AB+', 'AB-', 'O+', 'O-')),
  bags integer NOT NULL DEFAULT 1 CHECK(bags > 0),
  amount_minor bigint NOT NULL DEFAULT 0 CHECK(amount_minor >= 0),
  remarks text NOT NULL DEFAULT '',
  issued_by text NOT NULL REFERENCES "user"(id),
  invoice_id uuid REFERENCES invoice(id),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_blood_issue_patient ON blood_issue(patient_id);
CREATE INDEX idx_blood_issue_date ON blood_issue(issue_date);

-- 3. Prescriptions
CREATE TABLE prescription (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id text NOT NULL,
  doctor_id text NOT NULL,
  encounter_id uuid REFERENCES encounter(id),
  food_allergies text NOT NULL DEFAULT '',
  tendency_bleed text NOT NULL DEFAULT '',
  heart_disease text NOT NULL DEFAULT '',
  high_blood_pressure text NOT NULL DEFAULT '',
  diabetic text NOT NULL DEFAULT '',
  surgery text NOT NULL DEFAULT '',
  accident text NOT NULL DEFAULT '',
  others text NOT NULL DEFAULT '',
  medical_history text NOT NULL DEFAULT '',
  current_medication text NOT NULL DEFAULT '',
  female_pregnancy text NOT NULL DEFAULT '',
  breast_feeding text NOT NULL DEFAULT '',
  health_insurance text NOT NULL DEFAULT '',
  low_income text NOT NULL DEFAULT '',
  reference text NOT NULL DEFAULT '',
  status smallint NOT NULL DEFAULT 0 CHECK(status IN (0, 1)),
  plus_rate text NOT NULL DEFAULT '',
  temperature text NOT NULL DEFAULT '',
  problem_description text NOT NULL DEFAULT '',
  test text NOT NULL DEFAULT '',
  advice text NOT NULL DEFAULT '',
  next_visit_qty text NOT NULL DEFAULT '',
  next_visit_time text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_prescription_patient ON prescription(patient_id);
CREATE INDEX idx_prescription_doctor ON prescription(doctor_id);

CREATE TABLE prescription_medicine (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  prescription_id uuid NOT NULL REFERENCES prescription(id) ON DELETE CASCADE,
  medicine_id uuid REFERENCES medicine(id),
  medicine_name text NOT NULL,
  dosage text NOT NULL DEFAULT '',
  day text NOT NULL DEFAULT '',
  time text NOT NULL DEFAULT '',
  comment text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
