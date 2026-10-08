-- Migration 051: Doctor profile parity additions
-- Adds specialist, designation and qualification columns directly to doctor_profile
-- to align with legacy Doctor model fields and ensure fast relational queries.

ALTER TABLE doctor_profile
  ADD COLUMN IF NOT EXISTS specialist text NOT NULL DEFAULT '' CHECK (length(specialist) <= 191),
  ADD COLUMN IF NOT EXISTS designation text NOT NULL DEFAULT '' CHECK (length(designation) <= 191),
  ADD COLUMN IF NOT EXISTS qualification text NOT NULL DEFAULT '' CHECK (length(qualification) <= 191);

CREATE INDEX IF NOT EXISTS idx_doctor_profile_dept ON doctor_profile(department_id);
CREATE INDEX IF NOT EXISTS idx_doctor_profile_specialist ON doctor_profile(lower(trim(specialist)));
