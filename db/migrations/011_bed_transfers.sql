ALTER TABLE hospital_bed ADD COLUMN state text NOT NULL DEFAULT 'ready' CHECK(state IN ('ready','maintenance','unavailable'));
ALTER TABLE hospital_bed ADD COLUMN version integer NOT NULL DEFAULT 1 CHECK(version>0);
ALTER TABLE encounter ADD COLUMN admission_bed_id uuid REFERENCES hospital_bed(id);
UPDATE encounter SET admission_bed_id=bed_id;
CREATE TABLE bed_event (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 encounter_id uuid NOT NULL REFERENCES encounter(id),
 kind text NOT NULL CHECK(kind IN ('baseline','admission','transfer','discharge')),
 from_bed_id uuid REFERENCES hospital_bed(id),
 to_bed_id uuid REFERENCES hospital_bed(id),
 charge_minor bigint NOT NULL CHECK(charge_minor>=0),
 actor_id text NOT NULL REFERENCES "user"(id),
 reason text NOT NULL DEFAULT '' CHECK(length(reason)<=2000),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 encounter_version integer NOT NULL,
 UNIQUE(encounter_id,encounter_version)
);
-- Baselines describe known migration-time state, not an invented historical action.
INSERT INTO bed_event(encounter_id,kind,to_bed_id,charge_minor,actor_id,reason,encounter_version)
 SELECT id,'baseline',bed_id,bed_charge_minor,created_by,'Migration baseline; historical transfer details unavailable',version FROM encounter WHERE kind='ipd';
CREATE INDEX bed_event_history ON bed_event(encounter_id,encounter_version);
CREATE TRIGGER immutable_bed_event BEFORE UPDATE OR DELETE ON bed_event FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE FUNCTION retain_admission_bed() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='INSERT' THEN NEW.admission_bed_id=NEW.bed_id;
 ELSIF NEW.admission_bed_id IS DISTINCT FROM OLD.admission_bed_id THEN RAISE EXCEPTION 'Admission bed is retained';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER admission_bed_snapshot BEFORE INSERT OR UPDATE ON encounter FOR EACH ROW EXECUTE FUNCTION retain_admission_bed();
CREATE TABLE bed_state_event (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 bed_id uuid NOT NULL REFERENCES hospital_bed(id),
 from_state text NOT NULL,
 to_state text NOT NULL,
 actor_id text NOT NULL REFERENCES "user"(id),
 reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 2000),
 version integer NOT NULL,
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 UNIQUE(bed_id,version)
);
CREATE TRIGGER immutable_bed_state_event BEFORE UPDATE OR DELETE ON bed_state_event FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
