-- Enhance duty assignments and implement leave request tracking

ALTER TABLE attendance_shift_assignment
  ADD COLUMN IF NOT EXISTS note text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS active boolean NOT NULL DEFAULT true;

CREATE INDEX IF NOT EXISTS idx_attendance_shift_assignment_active
  ON attendance_shift_assignment(staff_id, active, effective_from);

CREATE TABLE IF NOT EXISTS attendance_leave_request (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  staff_id text NOT NULL REFERENCES "user"(id),
  from_date date NOT NULL,
  to_date date NOT NULL,
  days integer NOT NULL CHECK(days > 0),
  leave_type text NOT NULL CHECK(leave_type IN ('casual', 'annual', 'emergency', 'sick', 'other')),
  reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 1000),
  status text NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'approved', 'rejected', 'cancelled')),
  approver_id text REFERENCES "user"(id),
  approver_notes text NOT NULL DEFAULT '',
  actioned_at timestamptz,
  version integer NOT NULL DEFAULT 1 CHECK(version > 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT valid_leave_range CHECK (to_date >= from_date)
);

CREATE INDEX IF NOT EXISTS idx_attendance_leave_staff ON attendance_leave_request(staff_id, from_date DESC);
CREATE INDEX IF NOT EXISTS idx_attendance_leave_status ON attendance_leave_request(status);
