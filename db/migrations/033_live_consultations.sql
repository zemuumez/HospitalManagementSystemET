CREATE TABLE live_consultation (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  doctor_id text NOT NULL,
  patient_id text NOT NULL,
  encounter_id uuid,
  consultation_title varchar(191) NOT NULL,
  consultation_date timestamptz NOT NULL,
  duration_minutes integer NOT NULL CHECK(duration_minutes > 0),
  host_video boolean NOT NULL DEFAULT true,
  participant_video boolean NOT NULL DEFAULT true,
  type varchar(64) NOT NULL DEFAULT 'OPD',
  type_number varchar(64) NOT NULL DEFAULT '',
  platform_type varchar(64) NOT NULL DEFAULT 'zoom',
  meeting_id varchar(191) NOT NULL,
  password varchar(191) NOT NULL DEFAULT '',
  time_zone varchar(64) NOT NULL DEFAULT 'Africa/Addis_Ababa',
  status smallint NOT NULL DEFAULT 0 CHECK(status IN (0, 1, 2)),
  description text NOT NULL DEFAULT '',
  created_by text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_live_consultation_doctor ON live_consultation(doctor_id);
CREATE INDEX idx_live_consultation_patient ON live_consultation(patient_id);
CREATE INDEX idx_live_consultation_date ON live_consultation(consultation_date);

CREATE TABLE live_meeting (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title varchar(191) NOT NULL,
  meeting_date timestamptz NOT NULL,
  duration_minutes integer NOT NULL CHECK(duration_minutes > 0),
  host_video boolean NOT NULL DEFAULT true,
  participant_video boolean NOT NULL DEFAULT true,
  platform_type varchar(64) NOT NULL DEFAULT 'zoom',
  meeting_id varchar(191) NOT NULL,
  password varchar(191) NOT NULL DEFAULT '',
  time_zone varchar(64) NOT NULL DEFAULT 'Africa/Addis_Ababa',
  status smallint NOT NULL DEFAULT 0 CHECK(status IN (0, 1, 2)),
  description text NOT NULL DEFAULT '',
  created_by text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE live_meeting_candidate (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  live_meeting_id uuid NOT NULL REFERENCES live_meeting(id) ON DELETE CASCADE,
  user_id text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE(live_meeting_id, user_id)
);

CREATE TABLE live_consultation_provider_setting (
  user_id text PRIMARY KEY,
  platform_type varchar(64) NOT NULL DEFAULT 'zoom',
  api_key text NOT NULL DEFAULT '',
  api_secret text NOT NULL DEFAULT '',
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
