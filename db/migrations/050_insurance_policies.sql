-- Migration 050: Insurance policies, disease charge lines, and admission linkage
CREATE TABLE insurance (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(160) NOT NULL UNIQUE,
  service_tax_minor bigint NOT NULL DEFAULT 0 CHECK (service_tax_minor >= 0),
  discount integer NOT NULL DEFAULT 0 CHECK (discount >= 0 AND discount <= 100),
  remark text NOT NULL DEFAULT '',
  insurance_no varchar(191) NOT NULL,
  insurance_code varchar(191) NOT NULL,
  hospital_rate_minor bigint NOT NULL DEFAULT 0 CHECK (hospital_rate_minor >= 0),
  total_minor bigint NOT NULL DEFAULT 0 CHECK (total_minor >= 0),
  status smallint NOT NULL DEFAULT 1 CHECK (status IN (0, 1)),
  currency_symbol varchar(100) NOT NULL DEFAULT 'ETB',
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_insurance_name ON insurance(name);
CREATE INDEX idx_insurance_no ON insurance(insurance_no);

CREATE TABLE insurance_disease (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  insurance_id uuid NOT NULL REFERENCES insurance(id) ON DELETE CASCADE,
  disease_name varchar(191) NOT NULL,
  disease_charge_minor bigint NOT NULL DEFAULT 0 CHECK (disease_charge_minor >= 0),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_insurance_disease_insurance ON insurance_disease(insurance_id);

-- Link to IPD admission details for admission references and deletion protection
ALTER TABLE ipd_admission_details
  ADD COLUMN IF NOT EXISTS insurance_id uuid REFERENCES insurance(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_ipd_admission_details_insurance ON ipd_admission_details(insurance_id);
