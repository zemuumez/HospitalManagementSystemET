-- Migration 047: Add ward_name to hospital_bed for multi-ward inpatient visualization
-- and patient_advance_payment for advance billing ledger tracking.

ALTER TABLE hospital_bed
  ADD COLUMN IF NOT EXISTS ward_name text NOT NULL DEFAULT 'General Ward' CHECK (length(ward_name) BETWEEN 1 AND 80);

CREATE INDEX IF NOT EXISTS idx_hospital_bed_ward ON hospital_bed(ward_name);

CREATE TABLE IF NOT EXISTS patient_advance_payment (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  receipt_number bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
  patient_id uuid NOT NULL REFERENCES patient(id),
  amount_minor bigint NOT NULL CHECK (amount_minor > 0),
  payment_date date NOT NULL DEFAULT CURRENT_DATE,
  payment_method text NOT NULL DEFAULT 'cash' CHECK (payment_method IN ('cash', 'bank', 'telebirr', 'card')),
  notes text NOT NULL DEFAULT '' CHECK (length(notes) <= 500),
  received_by text NOT NULL REFERENCES "user"(id),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_advance_payment_patient ON patient_advance_payment(patient_id);
