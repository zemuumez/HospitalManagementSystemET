CREATE TABLE encounter_nurse (
 encounter_id uuid NOT NULL REFERENCES encounter(id),
 nurse_id text NOT NULL REFERENCES staff_access(user_id),
 active boolean NOT NULL,
 version integer NOT NULL CHECK(version>0),
 PRIMARY KEY(encounter_id,nurse_id)
);
CREATE TABLE encounter_nurse_event (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 encounter_id uuid NOT NULL REFERENCES encounter(id),
 nurse_id text NOT NULL REFERENCES staff_access(user_id),
 active boolean NOT NULL,
 version integer NOT NULL,
 actor_id text NOT NULL REFERENCES "user"(id),
 reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 2000),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 UNIQUE(encounter_id,nurse_id,version)
);
CREATE TRIGGER immutable_nurse_event BEFORE UPDATE OR DELETE ON encounter_nurse_event FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TABLE clinical_vitals (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 encounter_id uuid NOT NULL REFERENCES encounter(id),
 author_id text NOT NULL REFERENCES "user"(id),
 observed_at timestamptz NOT NULL,
 signed_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 measurements jsonb NOT NULL CHECK(jsonb_typeof(measurements)='object'),
 note text NOT NULL DEFAULT '' CHECK(length(note)<=2000),
 correction_of uuid UNIQUE REFERENCES clinical_vitals(id),
 correction_reason text NOT NULL DEFAULT '' CHECK(length(correction_reason)<=2000),
 request_key text NOT NULL,
 request_hash text NOT NULL,
 UNIQUE(author_id,request_key),
 CHECK((correction_of IS NULL AND correction_reason='') OR (correction_of IS NOT NULL AND length(correction_reason)>0))
);
CREATE INDEX clinical_vitals_encounter ON clinical_vitals(encounter_id,observed_at DESC);
CREATE TRIGGER immutable_clinical_vitals BEFORE UPDATE OR DELETE ON clinical_vitals FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
