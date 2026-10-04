CREATE TABLE staff_profile (
 user_id text PRIMARY KEY REFERENCES staff_access(user_id),
 details jsonb NOT NULL DEFAULT '{}' CHECK(jsonb_typeof(details)='object'),
 version integer NOT NULL DEFAULT 1 CHECK(version>0)
);
INSERT INTO staff_profile(user_id) SELECT user_id FROM staff_access;
CREATE FUNCTION initialize_staff_profile() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN INSERT INTO staff_profile(user_id) VALUES(NEW.user_id); RETURN NEW; END $$;
CREATE TRIGGER initialize_staff_profile AFTER INSERT ON staff_access FOR EACH ROW EXECUTE FUNCTION initialize_staff_profile();
CREATE TABLE staff_profile_revision (
 user_id text NOT NULL REFERENCES staff_profile(user_id),
 version integer NOT NULL, actor_id text NOT NULL REFERENCES "user"(id),
 reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 500),
 before_snapshot jsonb NOT NULL, after_snapshot jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(user_id,version)
);
CREATE TRIGGER immutable_staff_revision BEFORE UPDATE OR DELETE ON staff_profile_revision FOR EACH ROW EXECUTE FUNCTION protect_pharmacy_history();
CREATE TABLE staff_role_event (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 user_id text NOT NULL REFERENCES staff_access(user_id), actor_id text NOT NULL REFERENCES "user"(id),
 previous_role text NOT NULL, next_role text NOT NULL, reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 500),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TRIGGER immutable_staff_role_event BEFORE UPDATE OR DELETE ON staff_role_event FOR EACH ROW EXECUTE FUNCTION protect_pharmacy_history();
