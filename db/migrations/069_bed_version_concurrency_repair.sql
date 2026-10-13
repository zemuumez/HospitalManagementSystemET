-- Migration 069: Bed Version Concurrency Invalidation & Repair
-- Corrects concurrency gap where 068 restored bed operational states without incrementing hospital_bed.version.
-- Invalidates stale versions for beds in maintenance/unavailable states to reject concurrent stale writes,
-- while strictly preserving current operational states.

WITH latest_bed_state AS (
  SELECT DISTINCT ON (bed_id)
    bed_id,
    to_state,
    version,
    recorded_at
  FROM bed_state_event
  ORDER BY bed_id, recorded_at DESC, version DESC
)
UPDATE hospital_bed b
SET version = b.version + 1
FROM latest_bed_state lbs
WHERE b.id = lbs.bed_id
  AND b.state = lbs.to_state
  AND lbs.to_state IN ('maintenance', 'unavailable')
  AND b.version <= lbs.version
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
