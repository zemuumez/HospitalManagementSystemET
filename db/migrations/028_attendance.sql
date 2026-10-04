CREATE TABLE attendance_shift (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL UNIQUE CHECK(length(name) BETWEEN 1 AND 100),
  start_time time NOT NULL,
  end_time time NOT NULL,
  grace_period_minutes integer NOT NULL DEFAULT 15 CHECK(grace_period_minutes BETWEEN 0 AND 120),
  break_duration_minutes integer NOT NULL DEFAULT 60 CHECK(break_duration_minutes BETWEEN 0 AND 300),
  half_day_minutes integer NOT NULL DEFAULT 240 CHECK(half_day_minutes > 0 AND half_day_minutes <= 1440),
  full_day_minutes integer NOT NULL DEFAULT 480 CHECK(full_day_minutes >= half_day_minutes AND full_day_minutes <= 1440),
  is_overnight boolean NOT NULL DEFAULT false,
  active boolean NOT NULL DEFAULT true,
  version integer NOT NULL DEFAULT 1 CHECK(version > 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO attendance_shift (id, name, start_time, end_time, grace_period_minutes, break_duration_minutes, half_day_minutes, full_day_minutes, is_overnight, active)
VALUES 
('11111111-1111-1111-1111-111111111111', 'Day Shift', '08:00:00', '17:00:00', 15, 60, 240, 480, false, true),
('22222222-2222-2222-2222-222222222222', 'Night Shift', '20:00:00', '05:00:00', 15, 60, 240, 480, true, true);

CREATE TABLE attendance_shift_assignment (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  staff_id text NOT NULL REFERENCES "user"(id),
  shift_id uuid NOT NULL REFERENCES attendance_shift(id),
  effective_from date NOT NULL,
  effective_to date,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT valid_effective_range CHECK (effective_to IS NULL OR effective_to >= effective_from)
);

CREATE INDEX idx_attendance_shift_assignment_staff ON attendance_shift_assignment(staff_id, effective_from);

CREATE TABLE attendance_record (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  staff_id text NOT NULL REFERENCES "user"(id),
  work_date date NOT NULL,
  shift_id uuid NOT NULL REFERENCES attendance_shift(id),
  check_in_at timestamptz NOT NULL,
  check_out_at timestamptz,
  status text NOT NULL CHECK(status IN ('present', 'late', 'half_day', 'absent', 'checked_in')),
  approval_status text NOT NULL DEFAULT 'draft' CHECK(approval_status IN ('draft', 'submitted', 'approved', 'rejected')),
  total_break_minutes integer NOT NULL DEFAULT 0 CHECK(total_break_minutes >= 0),
  worked_minutes integer NOT NULL DEFAULT 0 CHECK(worked_minutes >= 0),
  late_minutes integer NOT NULL DEFAULT 0 CHECK(late_minutes >= 0),
  early_out_minutes integer NOT NULL DEFAULT 0 CHECK(early_out_minutes >= 0),
  overtime_minutes integer NOT NULL DEFAULT 0 CHECK(overtime_minutes >= 0),
  source text NOT NULL DEFAULT 'self_service' CHECK(source IN ('self_service', 'administrator')),
  admin_notes text NOT NULL DEFAULT '' CHECK(length(admin_notes) <= 1000),
  version integer NOT NULL DEFAULT 1 CHECK(version > 0),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT unique_staff_work_date UNIQUE(staff_id, work_date),
  CONSTRAINT valid_check_times CHECK (check_out_at IS NULL OR check_out_at >= check_in_at)
);

CREATE INDEX idx_attendance_record_staff_date ON attendance_record(staff_id, work_date DESC);
CREATE INDEX idx_attendance_record_date ON attendance_record(work_date DESC);
CREATE INDEX idx_attendance_record_approval ON attendance_record(approval_status);

CREATE TABLE attendance_break (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  attendance_record_id uuid NOT NULL REFERENCES attendance_record(id) ON DELETE CASCADE,
  start_at timestamptz NOT NULL,
  end_at timestamptz,
  duration_minutes integer CHECK(duration_minutes IS NULL OR duration_minutes >= 0),
  reason text NOT NULL DEFAULT '' CHECK(length(reason) <= 500),
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT valid_break_times CHECK (end_at IS NULL OR end_at >= start_at)
);

CREATE INDEX idx_attendance_break_record ON attendance_break(attendance_record_id, start_at);

CREATE TABLE attendance_correction (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  attendance_record_id uuid NOT NULL REFERENCES attendance_record(id) ON DELETE CASCADE,
  actor_id text NOT NULL REFERENCES "user"(id),
  reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 1000),
  before_snapshot jsonb NOT NULL,
  after_snapshot jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER immutable_attendance_correction BEFORE UPDATE OR DELETE ON attendance_correction
  FOR EACH ROW EXECUTE FUNCTION protect_retained_record();

CREATE TABLE attendance_approval_history (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  attendance_record_id uuid NOT NULL REFERENCES attendance_record(id) ON DELETE CASCADE,
  actor_id text NOT NULL REFERENCES "user"(id),
  from_status text NOT NULL,
  to_status text NOT NULL,
  reason text NOT NULL DEFAULT '' CHECK(length(reason) <= 1000),
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER immutable_attendance_approval_history BEFORE UPDATE OR DELETE ON attendance_approval_history
  FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
