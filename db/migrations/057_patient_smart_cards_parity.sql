-- Migration 057: Patient Smart Cards & ID Card Templates Parity
-- Authoritative issuance, transactional generation, persistent display identity,
-- secure verification exposure controls, and template management.

CREATE TABLE IF NOT EXISTS patient_id_card_template (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name              text NOT NULL UNIQUE CHECK (length(trim(name)) >= 1 AND length(name) <= 100),
  color             text NOT NULL CHECK (length(trim(color)) >= 1 AND length(color) <= 30),
  email             boolean NOT NULL DEFAULT true,
  phone             boolean NOT NULL DEFAULT true,
  dob               boolean NOT NULL DEFAULT true,
  blood_group       boolean NOT NULL DEFAULT true,
  address           boolean NOT NULL DEFAULT true,
  patient_unique_id boolean NOT NULL DEFAULT true,
  created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at        timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX IF NOT EXISTS idx_patient_id_card_template_name ON patient_id_card_template(lower(name));

CREATE SEQUENCE IF NOT EXISTS patient_unique_id_seq START WITH 1001;

ALTER TABLE patient ADD COLUMN IF NOT EXISTS patient_unique_id text UNIQUE CHECK (patient_unique_id IS NULL OR length(trim(patient_unique_id)) >= 3);
ALTER TABLE patient ADD COLUMN IF NOT EXISTS smart_card_template_id uuid REFERENCES patient_id_card_template(id) ON DELETE SET NULL;

UPDATE patient
SET patient_unique_id = 'P-' || lpad(nextval('patient_unique_id_seq')::text, 5, '0')
WHERE patient_unique_id IS NULL;

ALTER TABLE patient_smart_card ADD COLUMN IF NOT EXISTS template_id uuid REFERENCES patient_id_card_template(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_patient_smart_card_template ON patient_smart_card(template_id);

INSERT INTO patient_id_card_template (name, color, email, phone, dob, blood_group, address, patient_unique_id)
VALUES
  ('Standard', '#f4bf16', true, true, true, true, true, true),
  ('VIP', '#6571ff', true, true, false, true, false, true)
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS smart_card_generation_request (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id        text NOT NULL REFERENCES "user"(id),
  idempotency_key text NOT NULL,
  template_id     uuid NOT NULL REFERENCES patient_id_card_template(id) ON DELETE CASCADE,
  mode            text NOT NULL CHECK (mode IN ('all', 'single', 'unassigned')),
  patient_id      uuid REFERENCES patient(id) ON DELETE SET NULL,
  response_json   jsonb NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE(actor_id, idempotency_key)
);
CREATE INDEX IF NOT EXISTS idx_smart_card_gen_req_actor_key ON smart_card_generation_request(actor_id, idempotency_key);
