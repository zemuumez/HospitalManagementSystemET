-- Migration 065: Bed Assignment Reconciliation Repair
-- Corrective repair migration for bed management parity (G1, G2).
-- Safely applies schema corrections, provenance enrichment, and audit reconciliation
-- to databases where migration 064 was already recorded in schema_migration prior to this repair.
--
-- NOTE ON HISTORICAL DATA RECOVERY (G2):
-- In environments where original migration 063 or 064 was executed prior to this repair:
-- 1. Surviving assignments in bed_assignment are enriched into bed_assignment_migration_audit
--    with their true creation time, notes, author, and full JSON record.
-- 2. Where source rows were previously deleted or timestamps were overwritten in-place,
--    original clinical end times cannot be reconstructed from database tables alone without
--    restoring from a pre-upgrade database backup.
-- 3. Migration 065 ensures all surviving records are archived losslessly, enriches historical audit entries,
--    strictly binds mutations to current-operation candidate IDs (eliminating replay bugs),
--    and ensures unique partial indexes and audit tables remain fully consistent without fabricating clinical timestamps.

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
  original_assigned_by text NOT NULL DEFAULT '',
  original_created_at timestamptz DEFAULT NULL,
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
  ADD COLUMN IF NOT EXISTS original_created_at timestamptz DEFAULT NULL,
  ADD COLUMN IF NOT EXISTS original_record jsonb NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS administrative_quarantine_at timestamptz DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_ba_audit_assignment ON bed_assignment_migration_audit(bed_assignment_id);
CREATE INDEX IF NOT EXISTS idx_ba_audit_encounter ON bed_assignment_migration_audit(encounter_id);
CREATE INDEX IF NOT EXISTS idx_ba_audit_bed ON bed_assignment_migration_audit(bed_id);

-- 2. Provenance enrichment for existing audit rows from original 064 (G2):
-- Backfill surviving facts from bed_assignment into audit records that lacked them
UPDATE bed_assignment_migration_audit a
SET original_notes = ba.notes,
    original_assigned_by = ba.assigned_by,
    original_created_at = ba.created_at,
    original_record = to_jsonb(ba.*)
FROM bed_assignment ba
WHERE a.bed_assignment_id = ba.id
  AND (a.original_record = '{}'::jsonb OR a.original_record IS NULL);

-- For any audit rows where the source assignment was previously deleted,
-- retain original_assigned_from as original_created_at (authoritative historical fact)
-- rather than labeling repair execution time as creation time.
UPDATE bed_assignment_migration_audit
SET original_created_at = original_assigned_from
WHERE original_created_at IS NULL;

-- 3. Comprehensive Reconciliation & Provenance Archiving for Conflicting Groups:
-- Conservative policy:
-- For any physical bed or encounter with multiple open assignments (assigned_to IS NULL):
-- Preserve the legitimate active keeper (rn = 1). Priority: active encounter assigned to this bed, latest assigned_from/created_at/id.
-- For all other duplicate open rows (rn > 1): archive complete facts and remove from active occupancy using RETURNING.

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
),
inserted_bed_duplicates AS (
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
    )
  RETURNING bed_assignment_id
)
DELETE FROM bed_assignment
WHERE id IN (SELECT bed_assignment_id FROM inserted_bed_duplicates);

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
),
inserted_enc_duplicates AS (
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
    )
  RETURNING bed_assignment_id
)
DELETE FROM bed_assignment
WHERE id IN (SELECT bed_assignment_id FROM inserted_enc_duplicates);

-- 4. Superseded closed duplicate assignments:
-- For closed rows where a surviving open assignment ALREADY occupies the unique slot:
-- Do NOT reopen the closed row; archive it as superseded duplicate so it never violates unique constraints.
WITH superseded_closed AS (
  SELECT ba.id AS ba_id, ba.encounter_id, ba.bed_id, ba.patient_id, ba.assigned_from, ba.assigned_to,
         ba.notes, ba.assigned_by, ba.created_at, to_jsonb(ba.*) AS full_record
  FROM bed_assignment ba
  JOIN encounter e ON e.id = ba.encounter_id
  WHERE e.status = 'active'
    AND e.bed_id = ba.bed_id
    AND ba.assigned_to IS NOT NULL
    AND (
      EXISTS (
        SELECT 1 FROM bed_assignment open_bed
        WHERE open_bed.bed_id = ba.bed_id AND open_bed.assigned_to IS NULL AND open_bed.id != ba.id
      )
      OR EXISTS (
        SELECT 1 FROM bed_assignment open_enc
        WHERE open_enc.encounter_id = ba.encounter_id AND open_enc.assigned_to IS NULL AND open_enc.id != ba.id
      )
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
       assigned_to, clock_timestamp(), 'quarantined_superseded_duplicate',
       'Superseded duplicate assignment closed prior to migration; active slot occupied by surviving assignment'
FROM superseded_closed;

-- 5. Reconcile historical closed assignments with zero duration (assigned_to = created_at):
WITH historical_zero_duration AS (
  SELECT ba.id AS ba_id, ba.encounter_id, ba.bed_id, ba.patient_id, ba.assigned_from, ba.assigned_to,
         ba.notes, ba.assigned_by, ba.created_at, to_jsonb(ba.*) AS full_record,
         CASE 
           WHEN e.status = 'discharged' AND e.discharged_at IS NOT NULL AND NOT EXISTS (
             SELECT 1 FROM bed_assignment later
             WHERE later.encounter_id = ba.encounter_id AND later.assigned_from > ba.assigned_from
           ) THEN e.discharged_at
           ELSE NULL
         END AS resolved_end
  FROM bed_assignment ba
  JOIN encounter e ON e.id = ba.encounter_id
  WHERE ba.assigned_to = ba.created_at
    AND NOT EXISTS (
      SELECT 1 FROM bed_assignment_migration_audit a WHERE a.bed_assignment_id = ba.id
    )
),
inserted_zero_duration AS (
  INSERT INTO bed_assignment_migration_audit (
    bed_assignment_id, encounter_id, bed_id, patient_id, original_assigned_from, original_assigned_to,
    original_notes, original_assigned_by, original_created_at, original_record,
    reconciled_assigned_to, administrative_quarantine_at, resolution_action, reason
  )
  SELECT ba_id, encounter_id, bed_id, patient_id, assigned_from, assigned_to,
         notes, assigned_by, created_at, full_record,
         resolved_end, 
         CASE WHEN resolved_end IS NULL THEN clock_timestamp() ELSE NULL END,
         CASE WHEN resolved_end IS NOT NULL THEN 'reconciled_to_encounter_timeline' ELSE 'quarantined_historical_zero_duration' END,
         CASE WHEN resolved_end IS NOT NULL 
              THEN 'Reconciled zero-duration assignment using authoritative final encounter discharge'
              ELSE 'Historical zero-duration assignment preserved with unknown clinical end time' END
  FROM historical_zero_duration
  RETURNING bed_assignment_id, reconciled_assigned_to, resolution_action
)
UPDATE bed_assignment ba
SET assigned_to = izd.reconciled_assigned_to
FROM inserted_zero_duration izd
WHERE ba.id = izd.bed_assignment_id
  AND izd.resolution_action = 'reconciled_to_encounter_timeline'
  AND izd.reconciled_assigned_to IS NOT NULL;

-- 6. Safely restore legitimate active assignments where bed is vacant AND encounter currently has no open assignment (G1)
-- Crucially binds the restoration UPDATE strictly to the newly inserted audit rows via RETURNING bed_assignment_id,
-- completely preventing replay of historical audit actions!
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
),
inserted_restorations AS (
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
  WHERE rn = 1
  RETURNING bed_assignment_id
)
UPDATE bed_assignment
SET assigned_to = NULL
WHERE id IN (SELECT bed_assignment_id FROM inserted_restorations);

-- 7. Quarantine any surviving conflicting open assignments:
WITH conflicting_surviving_open AS (
  SELECT ba.id AS ba_id, ba.encounter_id, ba.bed_id, ba.patient_id, ba.assigned_from, ba.assigned_to,
         ba.notes, ba.assigned_by, ba.created_at, to_jsonb(ba.*) AS full_record
  FROM bed_assignment ba
  JOIN encounter e ON e.id = ba.encounter_id
  WHERE ba.assigned_to IS NULL
    AND (
      e.status != 'active'
      OR (e.bed_id IS NOT NULL AND e.bed_id != ba.bed_id)
      OR EXISTS (
        SELECT 1 FROM bed_assignment other
        WHERE other.bed_id = ba.bed_id
          AND other.assigned_to IS NULL
          AND other.id != ba.id
          AND (other.assigned_from > ba.assigned_from OR (other.assigned_from = ba.assigned_from AND other.created_at > ba.created_at))
      )
      OR EXISTS (
        SELECT 1 FROM bed_assignment other
        WHERE other.encounter_id = ba.encounter_id
          AND other.assigned_to IS NULL
          AND other.id != ba.id
          AND (other.assigned_from > ba.assigned_from OR (other.assigned_from = ba.assigned_from AND other.created_at > ba.created_at))
      )
    )
    AND NOT EXISTS (
      SELECT 1 FROM bed_assignment_migration_audit a WHERE a.bed_assignment_id = ba.id
    )
),
inserted_conflicting_open AS (
  INSERT INTO bed_assignment_migration_audit (
    bed_assignment_id, encounter_id, bed_id, patient_id, original_assigned_from, original_assigned_to,
    original_notes, original_assigned_by, original_created_at, original_record,
    reconciled_assigned_to, administrative_quarantine_at, resolution_action, reason
  )
  SELECT ba_id, encounter_id, bed_id, patient_id, assigned_from, assigned_to,
         notes, assigned_by, created_at, full_record,
         NULL, clock_timestamp(), 'quarantined_conflicting_assignment',
         'Conflicting open assignment archived and quarantined during repair without fabricating clinical end time'
  FROM conflicting_surviving_open
  RETURNING bed_assignment_id
)
DELETE FROM bed_assignment
WHERE id IN (SELECT bed_assignment_id FROM inserted_conflicting_open);

-- 8. Ensure unique partial indexes exist without conflict
CREATE UNIQUE INDEX IF NOT EXISTS one_active_bed_assignment
  ON bed_assignment(bed_id)
  WHERE assigned_to IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS one_active_encounter_bed_assignment
  ON bed_assignment(encounter_id)
  WHERE assigned_to IS NULL;
