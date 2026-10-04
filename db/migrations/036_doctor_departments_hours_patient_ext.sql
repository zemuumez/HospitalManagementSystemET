-- Migration 036: Doctor departments (versioned), doctor profile extensions,
-- patient additional source fields, hospital opening hours and date overrides.

-- ─── Doctor departments ───────────────────────────────────────────────────────
CREATE TABLE doctor_department (
 id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 title     text NOT NULL CHECK (length(trim(title)) BETWEEN 1 AND 160),
 description text NOT NULL DEFAULT '' CHECK (length(description) <= 2000),
 archived  boolean NOT NULL DEFAULT false,
 version   integer NOT NULL DEFAULT 1 CHECK (version > 0),
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX doctor_department_title ON doctor_department(lower(trim(title))) WHERE NOT archived;

-- Revision history for department renames/archive actions
CREATE TABLE doctor_department_revision (
 department_id uuid NOT NULL REFERENCES doctor_department(id),
 version       integer NOT NULL,
 actor_id      text NOT NULL REFERENCES "user"(id),
 reason        text NOT NULL CHECK (length(reason) BETWEEN 1 AND 500),
 before_snapshot jsonb NOT NULL,
 after_snapshot  jsonb NOT NULL,
 created_at    timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY (department_id, version)
);
CREATE TRIGGER immutable_dept_revision BEFORE UPDATE OR DELETE ON doctor_department_revision
 FOR EACH ROW EXECUTE FUNCTION protect_pharmacy_history();

-- ─── Doctor profile extensions ────────────────────────────────────────────────
-- Link doctor profiles to departments and add optional fee / bio fields.
ALTER TABLE doctor_profile
 ADD COLUMN IF NOT EXISTS department_id   uuid REFERENCES doctor_department(id),
 ADD COLUMN IF NOT EXISTS description     text NOT NULL DEFAULT '' CHECK (length(description) <= 2000),
 ADD COLUMN IF NOT EXISTS photo_url       text NOT NULL DEFAULT '' CHECK (length(photo_url) <= 500),
 ADD COLUMN IF NOT EXISTS opd_charge      numeric(12,2) NOT NULL DEFAULT 0 CHECK (opd_charge >= 0),
 ADD COLUMN IF NOT EXISTS appointment_charge numeric(12,2) NOT NULL DEFAULT 0 CHECK (appointment_charge >= 0);

-- ─── Patient additional source fields ─────────────────────────────────────────
ALTER TABLE patient_profile
 ADD COLUMN IF NOT EXISTS father_name     text NOT NULL DEFAULT '' CHECK (length(father_name) <= 150),
 ADD COLUMN IF NOT EXISTS religion        text NOT NULL DEFAULT '' CHECK (length(religion) <= 100),
 ADD COLUMN IF NOT EXISTS referral_source text NOT NULL DEFAULT '' CHECK (length(referral_source) <= 200),
 ADD COLUMN IF NOT EXISTS notes          text NOT NULL DEFAULT '' CHECK (length(notes) <= 4000);

-- ─── Hospital opening hours ───────────────────────────────────────────────────
-- Global weekly schedule: which weekdays the hospital is open and from/to.
CREATE TABLE hospital_hours (
 weekday      integer PRIMARY KEY CHECK (weekday BETWEEN 0 AND 6),
 open_minute  integer NOT NULL CHECK (open_minute BETWEEN 0 AND 1439),
 close_minute integer NOT NULL CHECK (close_minute BETWEEN 1 AND 1440 AND close_minute > open_minute),
 updated_at   timestamptz NOT NULL DEFAULT now()
);

-- Date-specific overrides: close on a public holiday, or open on a special date.
CREATE TABLE hospital_date_override (
 id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 override_date date NOT NULL UNIQUE,
 closed      boolean NOT NULL DEFAULT true,   -- true = closed all day
 open_minute  integer CHECK (open_minute BETWEEN 0 AND 1439),
 close_minute integer CHECK (close_minute BETWEEN 1 AND 1440 AND close_minute > open_minute),
 label       text NOT NULL DEFAULT '' CHECK (length(label) <= 200),
 actor_id    text NOT NULL REFERENCES "user"(id),
 created_at  timestamptz NOT NULL DEFAULT now(),
 -- when closed=true, open/close must be null; when closed=false, they must be set
 CHECK (
   (closed = true AND open_minute IS NULL AND close_minute IS NULL) OR
   (closed = false AND open_minute IS NOT NULL AND close_minute IS NOT NULL)
 )
);
CREATE INDEX hospital_date_override_date ON hospital_date_override(override_date);
