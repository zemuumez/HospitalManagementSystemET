-- Migration 068: Bed State Preservation & Reliable Evidence Repair
-- Corrects erroneous 067 state-reset where unoccupied maintenance/unavailable beds were flipped to 'ready'.
-- Recovers prior intentional states from trigger-protected, immutable bed_state_event records.

WITH latest_bed_state AS (
  SELECT DISTINCT ON (bed_id)
    bed_id,
    to_state,
    reason,
    version,
    recorded_at
  FROM bed_state_event
  ORDER BY bed_id, recorded_at DESC, version DESC
)
UPDATE hospital_bed b
SET state = lbs.to_state,
    version = b.version + 1
FROM latest_bed_state lbs
WHERE b.id = lbs.bed_id
  AND b.state = 'ready'
  AND lbs.to_state IN ('maintenance', 'unavailable')
  AND NOT EXISTS (
    SELECT 1 FROM encounter e
    WHERE e.bed_id = b.id
      AND e.status = 'active'
  )
  AND NOT EXISTS (
    SELECT 1 FROM bed_assignment ba
    WHERE ba.bed_id = b.id
      AND ba.assigned_to IS NULL
  );
