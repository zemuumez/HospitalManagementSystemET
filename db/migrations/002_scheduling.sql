-- Local hospital time is stored explicitly in the schedule. Instants use timestamptz.
CREATE TABLE doctor_profile (
 user_id text PRIMARY KEY REFERENCES staff_access(user_id),
 department text NOT NULL CHECK (length(department) BETWEEN 1 AND 100),
 timezone text NOT NULL DEFAULT 'Africa/Addis_Ababa' CHECK (timezone = 'Africa/Addis_Ababa'),
 slot_minutes integer NOT NULL DEFAULT 30 CHECK (slot_minutes BETWEEN 5 AND 120),
 version integer NOT NULL DEFAULT 1
);
CREATE TABLE doctor_hours (
 doctor_id text NOT NULL REFERENCES doctor_profile(user_id),
 weekday integer NOT NULL CHECK (weekday BETWEEN 0 AND 6),
 start_minute integer NOT NULL CHECK (start_minute BETWEEN 0 AND 1439),
 end_minute integer NOT NULL CHECK (end_minute BETWEEN 1 AND 1440 AND end_minute > start_minute),
 PRIMARY KEY (doctor_id,weekday,start_minute)
);
CREATE TABLE doctor_absence (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 doctor_id text NOT NULL REFERENCES doctor_profile(user_id),
 starts_at timestamptz NOT NULL,
 ends_at timestamptz NOT NULL CHECK (ends_at > starts_at),
 reason text NOT NULL DEFAULT '' CHECK (length(reason)<=200)
);
CREATE INDEX doctor_absence_range ON doctor_absence(doctor_id,starts_at,ends_at);
CREATE TABLE appointment (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 patient_id uuid NOT NULL REFERENCES patient(id),
 doctor_id text NOT NULL REFERENCES doctor_profile(user_id),
 starts_at timestamptz NOT NULL,
 ends_at timestamptz NOT NULL CHECK (ends_at > starts_at),
 status text NOT NULL DEFAULT 'booked' CHECK (status IN ('booked','arrived','completed','cancelled','no_show')),
 problem text NOT NULL DEFAULT '' CHECK (length(problem)<=2000),
 version integer NOT NULL DEFAULT 1,
 created_by text NOT NULL REFERENCES "user"(id),
 request_key text NOT NULL CHECK (length(request_key) BETWEEN 16 AND 80),
 notify_sms boolean NOT NULL DEFAULT false,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(created_by,request_key)
);
CREATE INDEX appointment_doctor_range ON appointment(doctor_id,starts_at,ends_at) WHERE status <> 'cancelled';
CREATE INDEX appointment_patient_range ON appointment(patient_id,starts_at,ends_at) WHERE status <> 'cancelled';
