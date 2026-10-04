CREATE TABLE patient_profile (
 patient_id uuid PRIMARY KEY REFERENCES patient(id) ON DELETE CASCADE,
 email text NOT NULL DEFAULT '' CHECK(length(email)<=254),
 gender text NOT NULL DEFAULT 'unknown' CHECK(gender IN ('unknown','female','male','other')),
 blood_group text NOT NULL DEFAULT '' CHECK(blood_group IN ('','A+','A-','B+','B-','AB+','AB-','O+','O-')),
 address1 text NOT NULL DEFAULT '' CHECK(length(address1)<=250),
 address2 text NOT NULL DEFAULT '' CHECK(length(address2)<=250),
 city text NOT NULL DEFAULT '' CHECK(length(city)<=100),
 region text NOT NULL DEFAULT '' CHECK(length(region)<=100),
 country text NOT NULL DEFAULT '' CHECK(length(country)<=100),
 postal_code text NOT NULL DEFAULT '' CHECK(length(postal_code)<=30),
 emergency_name text NOT NULL DEFAULT '' CHECK(length(emergency_name)<=150),
 emergency_phone text NOT NULL DEFAULT '' CHECK(length(emergency_phone)<=20),
 emergency_relationship text NOT NULL DEFAULT '' CHECK(length(emergency_relationship)<=100),
 active boolean NOT NULL DEFAULT true,
 sms_consent boolean NOT NULL DEFAULT false,
 email_consent boolean NOT NULL DEFAULT false,
 version integer NOT NULL DEFAULT 1 CHECK(version>0)
);
INSERT INTO patient_profile(patient_id) SELECT id FROM patient;
CREATE FUNCTION initialize_patient_profile() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 INSERT INTO patient_profile(patient_id) VALUES(NEW.id);
 RETURN NEW;
END $$;
CREATE TRIGGER initialize_profile AFTER INSERT ON patient FOR EACH ROW EXECUTE FUNCTION initialize_patient_profile();
CREATE TABLE patient_profile_revision (
 patient_id uuid NOT NULL REFERENCES patient(id),
 version integer NOT NULL,
 actor_id text NOT NULL REFERENCES "user"(id),
 reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 500),
 before_snapshot jsonb NOT NULL,
 after_snapshot jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(patient_id,version)
);
CREATE TRIGGER immutable_patient_revision BEFORE UPDATE OR DELETE ON patient_profile_revision
 FOR EACH ROW EXECUTE FUNCTION protect_pharmacy_history();
CREATE INDEX patient_name_dob ON patient(lower(family_name),date_of_birth);
