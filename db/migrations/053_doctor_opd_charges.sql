-- Migration 053: Doctor OPD charges master table with currency
-- Provides dedicated relational storage matching Laravel DoctorOPDCharge model.

CREATE TABLE IF NOT EXISTS doctor_opd_charge (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  doctor_id       text NOT NULL UNIQUE REFERENCES doctor_profile(user_id) ON DELETE CASCADE,
  standard_charge numeric(12,2) NOT NULL CHECK (standard_charge >= 0),
  currency_symbol varchar(10) NOT NULL DEFAULT 'ETB' CHECK (length(currency_symbol) <= 10),
  created_by      text NOT NULL REFERENCES "user"(id),
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_doctor_opd_charge_doc ON doctor_opd_charge(doctor_id);
