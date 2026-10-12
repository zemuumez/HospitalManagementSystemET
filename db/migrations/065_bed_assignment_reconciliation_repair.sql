-- Migration 065: Bed Assignment Reconciliation Repair
-- Corrective repair migration for bed management parity (F1, F2).
-- Safely applies schema corrections and audit reconciliation to databases where migration 064
-- was already recorded in schema_migration prior to this repair.
--
-- NOTE ON HISTORICAL DATA RECOVERY (F2):
-- In environments where original migration 063 (commit b5ebedb) was executed prior to this repair,
-- duplicate open assignments whose end timestamps were overwritten with created_at or other heuristics
-- cannot be reconstructed from database tables alone without restoring from a pre-upgrade database backup
-- or external audit log. Migration 065 ensures all surviving records are archived losslessly,
-- captures complete row facts (to_jsonb(ba.*), notes, timestamps, attribution), prevents any further data loss,
-- and ensures unique partial indexes and audit tables remain fully consistent without fabricating clinical timestamps.

-- 1. Ensure bed_assignment_migration_audit exists with full provenance schema
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
  original_record jsonb NOT NULL DEFAULT '{}'::jsonb,
  reconciled_assigned_to timestamptz,
  administrative_quarantine_at timestamptz,
  resolution_action text NOT NULL,
  reason text NOT NULL,
  reconciled_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

ALTER TABLE bed_assignment_migration_audit
  ADD COLUMN IF NOT EXISTS original_notes text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS original_assigned_by text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS original_created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  ADD COLUMN IF NOT EXISTS original_record jsonb NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS administrative_quarantine_at timestamptz DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_ba_audit_assignment ON bed_assignment_migration_audit(bed_assignment_id);
CREATE INDEX IF NOT EXISTS idx_ba_audit_encounter ON bed_assignment_migration_audit(encounter_id);
CREATE INDEX IF NOT EXISTS idx_ba_audit_bed ON bed_assignment_migration_audit(bed_id);

-- 2. Comprehensive Reconciliation & Provenance Archiving for Conflicting Groups:
-- Conservative policy:
-- For any physical bed or encounter with multiple open assignments (assigned_to IS NULL):
-- Preserve the legitimate active keeper (rn = 1). Priority: active encounter assigned to this bed, latest assigned_from/created_at/id.
-- For all other duplicate open rows (rn > 1): archive complete facts into bed_assignment_migration_audit and remove from active bed occupancy.

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
       'Duplicate open assignment for physical bed quarantined during repair migration'
FROM duplicate_bed_open
WHERE rn > 1
  AND NOT EXISTS (
    SELECT 1 FROM bed_assignment_migration_audit a WHERE a.bed_assignment_id = duplicate_bed_open.id
  );

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
       'Duplicate open assignment for encounter quarantined during repair migration'
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

-- 3. Safely restore legitimate active assignments where bed is vacant AND encounter currently has no open assignment
WITH vacant_restoration_candidates AS (
  SELECT ba.id AS ba_id, ba.encounter_id, ba.bed_id, ba.patient_id, ba.assigned_from, ba.assigned_to,
         ba.notes, ba.assigned_by, ba.created_at, to_jsonb(ba.*) AS full_record,
         ROW_NUMBER() OVER (
           PARTITION BY ba.bed_id, ba.encounter_id
           ORDER BY ba.assigned_from DESC, ba.created_at DESC, ba.id ASC
         ) AS rn
  FROM bed_assignment ba
  JOIN encounter e ON e.id = ba.encounter_id
  WHERE e.status = 'active'
    AND e.bed_id = ba.bed_id
    AND ba.assigned_to IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM bed_assignment open_bed
      WHERE open_bed.bed_id = ba.bed_id AND open_bed.assigned_to IS NULL
    )
    AND NOT EXISTS (
      SELECT 1 FROM bed_assignment open_enc
      WHERE open_enc.encounter_id = ba.encounter_id AND open_enc.assigned_to IS NULL
    )
    AND NOT EXISTS (
      SELECT 1 FROM bed_assignment_migration_audit a
      WHERE a.bed_assignment_id = ba.id
    )
)
INSERT INTO bed_assignment_migration_audit (
  bed_assignment_id, encounter_id, bed_id, patient_id, original_assigned_from, original_assigned_to,
  original_notes, original_assigned_by, original_created_at, original_record,
  reconciled_assigned_to, administrative_quarantine_at, resolution_action, reason
)
SELECT ba_id, encounter_id, bed_id, patient_id, assigned_from, assigned_to,
       notes, assigned_by, created_at, full_record,
       NULL, NULL, 'restored_active',
       'Assignment matched active encounter bed_id with vacant slot; restored to active occupancy during repair'
FROM vacant_restoration_candidates
WHERE rn = 1;

UPDATE bed_assignment
SET assigned_to = NULL
WHERE id IN (
  SELECT bed_assignment_id
  FROM bed_assignment_migration_audit
  WHERE resolution_action = 'restored_active'
    AND reconciled_at >= clock_timestamp() - interval '1 minute'
);

-- 4. Ensure unique partial indexes exist without conflict
CREATE UNIQUE INDEX IF NOT EXISTS one_active_bed_assignment
  ON bed_assignment(bed_id)
  WHERE assigned_to IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS one_active_encounter_bed_assignment
  ON bed_assignment(encounter_id)
  WHERE assigned_to IS NULL;
