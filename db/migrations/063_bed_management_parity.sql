-- Migration 063: Bed Management Workspace Parity
-- Enforces concurrency protection on bed assignments, adds description to hospital_bed,
-- and ensures strict unique constraints for active bed occupancy.

-- 1. Ensure hospital_bed has description column
ALTER TABLE hospital_bed
  ADD COLUMN IF NOT EXISTS description text NOT NULL DEFAULT '' CHECK (length(description) <= 2000);

-- 2. Performance indexes on hospital_bed
CREATE INDEX IF NOT EXISTS idx_hospital_bed_active_state ON hospital_bed(active, state);
CREATE INDEX IF NOT EXISTS idx_hospital_bed_type ON hospital_bed(type_id, active);

-- 3. Reconcile any prior duplicate open assignments safely before creating unique constraints
WITH duplicate_bed_open AS (
  SELECT id, ROW_NUMBER() OVER (PARTITION BY bed_id ORDER BY created_at DESC, id) AS rn
  FROM bed_assignment
  WHERE assigned_to IS NULL
)
UPDATE bed_assignment
SET assigned_to = created_at
WHERE id IN (SELECT id FROM duplicate_bed_open WHERE rn > 1);

WITH duplicate_enc_open AS (
  SELECT id, ROW_NUMBER() OVER (PARTITION BY encounter_id ORDER BY created_at DESC, id) AS rn
  FROM bed_assignment
  WHERE assigned_to IS NULL
)
UPDATE bed_assignment
SET assigned_to = created_at
WHERE id IN (SELECT id FROM duplicate_enc_open WHERE rn > 1);

-- 4. Unique partial indexes enforcing concurrency protection
-- Exactly one active assignment per physical bed
CREATE UNIQUE INDEX IF NOT EXISTS one_active_bed_assignment
  ON bed_assignment(bed_id)
  WHERE assigned_to IS NULL;

-- Exactly one active bed assignment per encounter
CREATE UNIQUE INDEX IF NOT EXISTS one_active_encounter_bed_assignment
  ON bed_assignment(encounter_id)
  WHERE assigned_to IS NULL;

-- 5. Performance index on bed_type
CREATE INDEX IF NOT EXISTS idx_bed_type_active ON bed_type(active, name);

-- 6. Allow IPD encounters to release beds on discharge and await assignment
ALTER TABLE encounter DROP CONSTRAINT IF EXISTS encounter_check;
ALTER TABLE encounter ADD CONSTRAINT encounter_check CHECK((kind='ipd') OR (kind='opd' AND bed_id IS NULL));
