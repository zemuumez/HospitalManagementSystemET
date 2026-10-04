CREATE TABLE appointment_status_event (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 appointment_id uuid NOT NULL REFERENCES appointment(id),
 previous_status text NOT NULL,next_status text NOT NULL,
 actor_id text NOT NULL REFERENCES "user"(id),
 reason text NOT NULL DEFAULT '' CHECK(length(reason)<=1000),
 version integer NOT NULL CHECK(version>0),
 recorded_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 UNIQUE(appointment_id,version)
);
CREATE TRIGGER immutable_appointment_status_event BEFORE UPDATE OR DELETE ON appointment_status_event FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
