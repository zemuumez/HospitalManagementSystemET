CREATE TABLE hospital_bed (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 name text NOT NULL UNIQUE CHECK(length(name) BETWEEN 1 AND 80),
 bed_type text NOT NULL CHECK(length(bed_type) BETWEEN 1 AND 80),
 charge_minor bigint NOT NULL CHECK(charge_minor BETWEEN 0 AND 1000000000),
 active boolean NOT NULL DEFAULT true,
 created_by text NOT NULL REFERENCES "user"(id)
);
CREATE TABLE patient_case (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 number bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
 patient_id uuid NOT NULL REFERENCES patient(id),
 doctor_id text NOT NULL REFERENCES staff_access(user_id),
 description text NOT NULL DEFAULT '' CHECK(length(description)<=2000),
 created_by text NOT NULL REFERENCES "user"(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(id,patient_id,doctor_id)
);
CREATE TABLE encounter (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 number bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
 kind text NOT NULL CHECK(kind IN ('ipd','opd')),
 case_id uuid NOT NULL,
 patient_id uuid NOT NULL,
 doctor_id text NOT NULL,
 bed_id uuid REFERENCES hospital_bed(id),
 admitted_at timestamptz NOT NULL,
 status text NOT NULL DEFAULT 'active' CHECK(status IN ('active','discharged')),
 discharged_at timestamptz,
 discharge_summary text NOT NULL DEFAULT '' CHECK(length(discharge_summary)<=10000),
 symptoms text NOT NULL DEFAULT '' CHECK(length(symptoms)<=2000),
 bed_charge_minor bigint NOT NULL DEFAULT 0 CHECK(bed_charge_minor>=0),
 version integer NOT NULL DEFAULT 1,
 created_by text NOT NULL REFERENCES "user"(id),
 request_key text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(case_id,patient_id,doctor_id) REFERENCES patient_case(id,patient_id,doctor_id),
 CHECK((kind='ipd' AND bed_id IS NOT NULL) OR (kind='opd' AND bed_id IS NULL)),
 CHECK((status='active' AND discharged_at IS NULL) OR (status='discharged' AND discharged_at>=admitted_at AND length(discharge_summary)>0)),
 UNIQUE(created_by,request_key)
);
CREATE UNIQUE INDEX one_active_bed_occupant ON encounter(bed_id) WHERE status='active' AND bed_id IS NOT NULL;
CREATE UNIQUE INDEX one_active_inpatient_stay ON encounter(patient_id) WHERE status='active' AND kind='ipd';
CREATE INDEX encounter_patient ON encounter(patient_id,admitted_at DESC);
CREATE INDEX encounter_doctor ON encounter(doctor_id,admitted_at DESC);
CREATE TABLE clinical_note (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 encounter_id uuid NOT NULL REFERENCES encounter(id),
 author_id text NOT NULL REFERENCES "user"(id),
 body text NOT NULL CHECK(length(body) BETWEEN 1 AND 10000),
 signed_at timestamptz NOT NULL DEFAULT now(),
 request_key text NOT NULL,
 UNIQUE(author_id,request_key)
);
-- Signed notes are append-only; corrections are separate signed entries.
CREATE FUNCTION protect_signed_note() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'Signed clinical notes are immutable';
END $$;
CREATE TRIGGER immutable_clinical_note BEFORE UPDATE OR DELETE ON clinical_note FOR EACH ROW EXECUTE FUNCTION protect_signed_note();
