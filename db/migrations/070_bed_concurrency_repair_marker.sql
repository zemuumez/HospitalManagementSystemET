-- Migration 070: Durable Bed Concurrency Repair Marker & Stale Version Invalidation
-- Corrects upgrade gap where beds with configuration-only edits (via UpdateBed) had
-- hospital_bed.version > latest bed_state_event.version, causing migration 069 to skip them.
-- Uses a durable, once-only repair marker table to track invalidation and strictly prevent double-repair,
-- preserving current operational states.

CREATE TABLE IF NOT EXISTS bed_concurrency_repair_marker (
  bed_id uuid PRIMARY KEY REFERENCES hospital_bed(id),
  repaired_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  repaired_version integer NOT NULL
);

WITH latest_bed_state AS (
  SELECT DISTINCT ON (bed_id)
    bed_id,
    to_state,
    reason,
    version,
    recorded_at
  FROM bed_state_event
  ORDER BY bed_id, recorded_at DESC, version DESC
),
candidates AS (
  SELECT b.id, b.version
  FROM hospital_bed b
  JOIN latest_bed_state lbs ON lbs.bed_id = b.id
  WHERE b.state = lbs.to_state
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
    )
    AND NOT EXISTS (
      SELECT 1 FROM bed_concurrency_repair_marker m
      WHERE m.bed_id = b.id
    )
),
repaired AS (
  UPDATE hospital_bed b
  SET version = b.version + 1
  FROM candidates c
  WHERE b.id = c.id
  RETURNING b.id, b.version
)
INSERT INTO bed_concurrency_repair_marker (bed_id, repaired_version)
SELECT id, version FROM repaired
ON CONFLICT (bed_id) DO NOTHING;
