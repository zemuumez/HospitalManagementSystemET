-- Migration 063: Bed Management Workspace Parity
-- Enforces concurrency protection on bed assignments, adds description to hospital_bed,
-- losslessly archives duplicate open assignments in bed_assignment_migration_audit,
-- and ensures strict unique constraints for active bed occupancy without fabricating clinical timestamps.

-- 1. Ensure hospital_bed has description column
ALTER TABLE hospital_bed
  ADD COLUMN IF NOT EXISTS description text NOT NULL DEFAULT '' CHECK (length(description) <= 2000);

-- 2. Performance indexes on hospital_bed
CREATE INDEX IF NOT EXISTS idx_hospital_bed_active_state ON hospital_bed(active, state);
CREATE INDEX IF NOT EXISTS idx_hospital_bed_type ON hospital_bed(type_id, active);

-- 3. Dedicated audit table for historical bed assignment provenance and reconciliation
CREATE TABLE IF NOT EXISTS bed_assignment_migration_audit (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  bed_assignment_id uuid NOT NULL,
  encounter_id uuid NOT NULL,
  bed_id uuid NOT NULL,
  patient_id uuid NOT NULL,
  original_assigned_from timestamptz NOT NULL,
  original_assigned_to timestamptz,
  original_notes text NOT NULL DEFAULT '',
  original_assigned_by text NOT NULL,
  original_created_at timestamptz NOT NULL,
  original_record jsonb NOT NULL,
  reconciled_assigned_to timestamptz,
  administrative_quarantine_at timestamptz,
  resolution_action text NOT NULL,
  reason text NOT NULL,
  reconciled_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX IF NOT EXISTS idx_ba_audit_assignment ON bed_assignment_migration_audit(bed_assignment_id);
CREATE INDEX IF NOT EXISTS idx_ba_audit_encounter ON bed_assignment_migration_audit(encounter_id);
CREATE INDEX IF NOT EXISTS idx_ba_audit_bed ON bed_assignment_migration_audit(bed_id);

-- 4. Reconcile duplicate open assignments losslessly before creating unique partial indexes.
-- Conservative policy:
-- In any conflicting open group (multiple rows with assigned_to IS NULL for the same physical bed or encounter):
-- The legitimate active keeper (rn = 1) is preserved as open. Priority: active encounter currently assigned to this bed,
-- then most recent assigned_from, created_at, id.
-- All other duplicate rows (rn > 1) are archived with full evidence (to_jsonb(ba.*), notes, timestamps, attribution)
-- into bed_assignment_migration_audit. Unknown clinical end times are preserved as NULL without fabricating durations.
-- Administrative quarantine time is recorded separately.

WITH duplicate_bed_open AS (
  SELECT ba.id, ba.encounter_id, ba.bed_id, ba.patient_id, ba.assigned_from, ba.assigned_to,
         ba.notes, ba.assigned_by, ba.created_at, to_jsonb(ba.*) AS full_record,
         ROW_NUMBER() OVER (
           PARTITION BY ba.bed_id 
           ORDER BY 
             CASE WHEN e.status = 'active' AND e.bed_id = ba.bed_id THEN 0 ELSE 1 END,
             ba.assigned_from DESC,
             ba.created_at DESC, 
             ba.id ASC
         ) AS rn
  FROM bed_assignment ba
  LEFT JOIN encounter e ON e.id = ba.encounter_id
  WHERE ba.assigned_to IS NULL
)
INSERT INTO bed_assignment_migration_audit (
  bed_assignment_id, encounter_id, bed_id, patient_id, original_assigned_from, original_assigned_to,
  original_notes, original_assigned_by, original_created_at, original_record,
  reconciled_assigned_to, administrative_quarantine_at, resolution_action, reason
)
SELECT id, encounter_id, bed_id, patient_id, assigned_from, assigned_to,
       notes, assigned_by, created_at, full_record,
       NULL, clock_timestamp(), 'quarantined_duplicate',
       'Duplicate open assignment for physical bed quarantined during initial parity migration'
FROM duplicate_bed_open
WHERE rn > 1;

DELETE FROM bed_assignment
WHERE id IN (
  SELECT bed_assignment_id
  FROM bed_assignment_migration_audit
  WHERE resolution_action = 'quarantined_duplicate'
);

WITH duplicate_enc_open AS (
  SELECT ba.id, ba.encounter_id, ba.bed_id, ba.patient_id, ba.assigned_from, ba.assigned_to,
         ba.notes, ba.assigned_by, ba.created_at, to_jsonb(ba.*) AS full_record,
         ROW_NUMBER() OVER (
           PARTITION BY ba.encounter_id 
           ORDER BY 
             CASE WHEN e.status = 'active' AND e.bed_id = ba.bed_id THEN 0 ELSE 1 END,
             ba.assigned_from DESC,
             ba.created_at DESC, 
             ba.id ASC
         ) AS rn
  FROM bed_assignment ba
  LEFT JOIN encounter e ON e.id = ba.encounter_id
  WHERE ba.assigned_to IS NULL
)
INSERT INTO bed_assignment_migration_audit (
  bed_assignment_id, encounter_id, bed_id, patient_id, original_assigned_from, original_assigned_to,
  original_notes, original_assigned_by, original_created_at, original_record,
  reconciled_assigned_to, administrative_quarantine_at, resolution_action, reason
)
SELECT id, encounter_id, bed_id, patient_id, assigned_from, assigned_to,
       notes, assigned_by, created_at, full_record,
       NULL, clock_timestamp(), 'quarantined_duplicate',
       'Duplicate open assignment for encounter quarantined during initial parity migration'
FROM duplicate_enc_open
WHERE rn > 1
  AND NOT EXISTS (
    SELECT 1 FROM bed_assignment_migration_audit a WHERE a.bed_assignment_id = duplicate_enc_open.id
  );

DELETE FROM bed_assignment
WHERE id IN (
  SELECT bed_assignment_id
  FROM bed_assignment_migration_audit
  WHERE resolution_action = 'quarantined_duplicate'
);

-- 5. Unique partial indexes enforcing concurrency protection
-- Exactly one active assignment per physical bed
CREATE UNIQUE INDEX IF NOT EXISTS one_active_bed_assignment
  ON bed_assignment(bed_id)
  WHERE assigned_to IS NULL;

-- Exactly one active bed assignment per encounter
CREATE UNIQUE INDEX IF NOT EXISTS one_active_encounter_bed_assignment
  ON bed_assignment(encounter_id)
  WHERE assigned_to IS NULL;

-- 6. Performance index on bed_type
CREATE INDEX IF NOT EXISTS idx_bed_type_active ON bed_type(active, name);

-- 7. Allow IPD encounters to release beds on discharge and await assignment
ALTER TABLE encounter DROP CONSTRAINT IF EXISTS encounter_check;
ALTER TABLE encounter ADD CONSTRAINT encounter_check CHECK((kind='ipd') OR (kind='opd' AND bed_id IS NULL));
