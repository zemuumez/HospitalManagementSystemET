CREATE TABLE hospital_complaint (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id text NOT NULL,
  title varchar(191) NOT NULL,
  description text NOT NULL,
  status smallint NOT NULL DEFAULT 0 CHECK(status IN (0, 1, 2, 3)),
  response text NOT NULL DEFAULT '',
  resolved_by text,
  resolved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_hospital_complaint_patient ON hospital_complaint(patient_id);
CREATE INDEX idx_hospital_complaint_status ON hospital_complaint(status);

CREATE TABLE hospital_notice_board (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title varchar(191) NOT NULL,
  description text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE hospital_enquiry (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name varchar(191) NOT NULL,
  email varchar(191) NOT NULL,
  contact_no varchar(64) NOT NULL DEFAULT '',
  type smallint NOT NULL DEFAULT 1 CHECK(type IN (1, 2, 3, 4)),
  message text NOT NULL,
  viewed_by text,
  status smallint NOT NULL DEFAULT 0 CHECK(status IN (0, 1)),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_hospital_enquiry_status ON hospital_enquiry(status);

CREATE TABLE hospital_visitor (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  purpose smallint NOT NULL DEFAULT 1 CHECK(purpose IN (1, 2, 3)),
  name varchar(191) NOT NULL,
  phone varchar(64) NOT NULL DEFAULT '',
  id_card varchar(64) NOT NULL DEFAULT '',
  no_of_person integer NOT NULL DEFAULT 1 CHECK(no_of_person >= 1),
  date date NOT NULL,
  in_time varchar(8) NOT NULL DEFAULT '',
  out_time varchar(8) NOT NULL DEFAULT '',
  note text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE hospital_call_log (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(191) NOT NULL,
  phone varchar(64) NOT NULL DEFAULT '',
  date date NOT NULL,
  follow_up_date date,
  note text NOT NULL DEFAULT '',
  call_type smallint NOT NULL CHECK(call_type IN (1, 2)),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE hospital_postal (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  from_title varchar(191) NOT NULL DEFAULT '',
  to_title varchar(191) NOT NULL DEFAULT '',
  reference_no varchar(191) NOT NULL DEFAULT '',
  date date NOT NULL,
  address text NOT NULL DEFAULT '',
  type smallint NOT NULL CHECK(type IN (1, 2)),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- Seed initial hospital notice
INSERT INTO hospital_notice_board (id, title, description)
VALUES
  ('ea111111-1111-1111-1111-111111111111', 'Hospital Accreditation Renewal Completed', 'The Federal Ministry of Health has certified Addis Ababa Central Hospital with Grade-A Tertiary Clinical Excellence.')
ON CONFLICT (id) DO NOTHING;
