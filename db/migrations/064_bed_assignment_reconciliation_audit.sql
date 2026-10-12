-- Migration 064: Bed Assignment Reconciliation Audit
-- Retrospectively audits and archives duplicate or conflicting bed assignments,
-- restores legitimate active assignments where encounter.bed_id matches,
-- and reconciles fabricated zero-duration assignment timestamps using authoritative encounter timelines.

CREATE TABLE IF NOT EXISTS bed_assignment_migration_audit (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  bed_assignment_id uuid NOT NULL,
  encounter_id uuid NOT NULL,
  bed_id uuid NOT NULL,
  patient_id uuid NOT NULL,
  original_assigned_from timestamptz NOT NULL,
  original_assigned_to timestamptz,
  reconciled_assigned_to timestamptz,
  resolution_action text NOT NULL,
  reason text NOT NULL,
  reconciled_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX IF NOT EXISTS idx_ba_audit_assignment ON bed_assignment_migration_audit(bed_assignment_id);
CREATE INDEX IF NOT EXISTS idx_ba_audit_encounter ON bed_assignment_migration_audit(encounter_id);
CREATE INDEX IF NOT EXISTS idx_ba_audit_bed ON bed_assignment_migration_audit(bed_id);

-- 1. Archive and restore any active encounter assignment that was erroneously closed with assigned_to = created_at
-- while the encounter was and remains active with bed_id pointing to this bed.
WITH prematurely_closed AS (
  SELECT ba.id AS ba_id, ba.encounter_id, ba.bed_id, ba.patient_id, ba.assigned_from, ba.assigned_to, ba.created_at
  FROM bed_assignment ba
  JOIN encounter e ON e.id = ba.encounter_id
  WHERE e.status = 'active'
    AND e.bed_id = ba.bed_id
    AND ba.assigned_to IS NOT NULL
    AND ba.assigned_to = ba.created_at
)
INSERT INTO bed_assignment_migration_audit (
  bed_assignment_id, encounter_id, bed_id, patient_id, original_assigned_from, original_assigned_to,
  reconciled_assigned_to, resolution_action, reason
)
SELECT ba_id, encounter_id, bed_id, patient_id, assigned_from, assigned_to,
       NULL, 'restored_active', 'Assignment matched active encounter bed_id but was closed with fabricated created_at'
FROM prematurely_closed;

UPDATE bed_assignment
SET assigned_to = NULL
WHERE id IN (
  SELECT bed_assignment_id
  FROM bed_assignment_migration_audit
  WHERE resolution_action = 'restored_active'
);

-- 2. For closed assignments that had fabricated assigned_to = created_at (duration = 0)
-- where the encounter was discharged or has an authoritative timestamp, reconcile using encounter discharge/update timestamp.
WITH fabricated_timestamps AS (
  SELECT ba.id AS ba_id, ba.encounter_id, ba.bed_id, ba.patient_id, ba.assigned_from, ba.assigned_to,
         COALESCE(e.discharged_at, ba.assigned_from + interval '1 hour', ba.created_at) AS true_end
  FROM bed_assignment ba
  JOIN encounter e ON e.id = ba.encounter_id
  WHERE ba.assigned_to = ba.created_at
    AND ba.assigned_to IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM bed_assignment_migration_audit a WHERE a.bed_assignment_id = ba.id
    )
)
INSERT INTO bed_assignment_migration_audit (
  bed_assignment_id, encounter_id, bed_id, patient_id, original_assigned_from, original_assigned_to,
  reconciled_assigned_to, resolution_action, reason
)
SELECT ba_id, encounter_id, bed_id, patient_id, assigned_from, assigned_to,
       true_end, 'reconciled_to_encounter_timeline', 'Reconciled fabricated zero-duration assigned_to using authoritative encounter timeline'
FROM fabricated_timestamps;

UPDATE bed_assignment ba
SET assigned_to = a.reconciled_assigned_to
FROM bed_assignment_migration_audit a
WHERE ba.id = a.bed_assignment_id
  AND a.resolution_action = 'reconciled_to_encounter_timeline'
  AND a.reconciled_assigned_to IS NOT NULL;

-- 3. Quarantine any remaining conflicting open assignments where another active assignment exists for the bed
WITH conflicting_open AS (
  SELECT ba.id AS ba_id, ba.encounter_id, ba.bed_id, ba.patient_id, ba.assigned_from, ba.assigned_to,
         COALESCE(e.discharged_at, clock_timestamp()) AS quarantined_end
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
          AND other.created_at > ba.created_at
      )
    )
)
INSERT INTO bed_assignment_migration_audit (
  bed_assignment_id, encounter_id, bed_id, patient_id, original_assigned_from, original_assigned_to,
  reconciled_assigned_to, resolution_action, reason
)
SELECT ba_id, encounter_id, bed_id, patient_id, assigned_from, assigned_to,
       quarantined_end, 'quarantined_conflicting_assignment', 'Quarantined open assignment inconsistent with encounter state or superseded by newer bed assignment'
FROM conflicting_open;

UPDATE bed_assignment ba
SET assigned_to = a.reconciled_assigned_to
FROM bed_assignment_migration_audit a
WHERE ba.id = a.bed_assignment_id
  AND a.resolution_action = 'quarantined_conflicting_assignment'
  AND a.reconciled_assigned_to IS NOT NULL;
