ALTER TABLE clinical_note ADD CONSTRAINT clinical_note_encounter_unique UNIQUE(id,encounter_id);
CREATE TABLE clinical_addendum (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), encounter_id uuid NOT NULL REFERENCES encounter(id),
 note_id uuid, kind text NOT NULL CHECK(kind IN ('note','discharge')),
 author_id text NOT NULL REFERENCES "user"(id), body text NOT NULL CHECK(length(body) BETWEEN 1 AND 10000),
 reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 1000), signed_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 request_key text NOT NULL CHECK(length(request_key) BETWEEN 16 AND 80),
 FOREIGN KEY(note_id,encounter_id) REFERENCES clinical_note(id,encounter_id),
 CHECK((kind='note' AND note_id IS NOT NULL) OR (kind='discharge' AND note_id IS NULL)),
 UNIQUE(author_id,request_key)
);
CREATE INDEX clinical_addendum_encounter ON clinical_addendum(encounter_id,signed_at DESC,id);
CREATE TRIGGER immutable_clinical_addendum BEFORE UPDATE OR DELETE ON clinical_addendum FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE FUNCTION retain_discharge_summary() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.status='discharged' AND (NEW.status IS DISTINCT FROM OLD.status OR NEW.discharge_summary IS DISTINCT FROM OLD.discharge_summary OR NEW.discharged_at IS DISTINCT FROM OLD.discharged_at) THEN
  RAISE EXCEPTION 'Signed discharge is retained; append an addendum';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER retain_discharge_summary BEFORE UPDATE ON encounter FOR EACH ROW EXECUTE FUNCTION retain_discharge_summary();
