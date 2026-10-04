ALTER TABLE doctor_absence ADD COLUMN created_by text REFERENCES "user"(id);
ALTER TABLE doctor_absence ADD COLUMN cancelled_at timestamptz;
ALTER TABLE doctor_absence ADD COLUMN cancelled_by text REFERENCES "user"(id);
ALTER TABLE doctor_absence ADD COLUMN cancel_reason text NOT NULL DEFAULT '' CHECK(length(cancel_reason)<=200);
ALTER TABLE doctor_absence ADD COLUMN version integer NOT NULL DEFAULT 1 CHECK(version>0);
ALTER TABLE appointment ADD COLUMN original_starts_at timestamptz;
UPDATE appointment SET original_starts_at=starts_at;
ALTER TABLE appointment ALTER COLUMN original_starts_at SET NOT NULL;
CREATE FUNCTION retain_booking_time() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='INSERT' THEN NEW.original_starts_at=NEW.starts_at;
 ELSIF NEW.original_starts_at IS DISTINCT FROM OLD.original_starts_at THEN RAISE EXCEPTION 'Original booking time is retained';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER retain_booking_time BEFORE INSERT OR UPDATE ON appointment FOR EACH ROW EXECUTE FUNCTION retain_booking_time();
CREATE TABLE appointment_reschedule (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 appointment_id uuid NOT NULL REFERENCES appointment(id),
 from_start timestamptz NOT NULL,
 from_end timestamptz NOT NULL,
 to_start timestamptz NOT NULL,
 to_end timestamptz NOT NULL,
 actor_id text NOT NULL REFERENCES "user"(id),
 reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 2000),
 version integer NOT NULL,
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 UNIQUE(appointment_id,version)
);
CREATE TRIGGER immutable_appointment_reschedule BEFORE UPDATE OR DELETE ON appointment_reschedule FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
