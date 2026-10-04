ALTER TABLE diagnostic_test ADD COLUMN active boolean NOT NULL DEFAULT true,
 ADD COLUMN version integer NOT NULL DEFAULT 1 CHECK(version>0),
 ADD COLUMN revision integer NOT NULL DEFAULT 1 CHECK(revision>0),
 ADD COLUMN root_id uuid REFERENCES diagnostic_test(id),
 ADD COLUMN supersedes uuid UNIQUE REFERENCES diagnostic_test(id);
ALTER TABLE diagnostic_test DROP CONSTRAINT diagnostic_test_kind_name_key;
CREATE UNIQUE INDEX diagnostic_active_name ON diagnostic_test(kind,name) WHERE active;
CREATE INDEX diagnostic_test_series ON diagnostic_test(root_id,revision DESC);
DROP TRIGGER immutable_diagnostic_test ON diagnostic_test;
CREATE FUNCTION retain_diagnostic_definition() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='UPDATE' AND (to_jsonb(NEW)-'active'-'version')=(to_jsonb(OLD)-'active'-'version') AND OLD.active AND NOT NEW.active AND NEW.version=OLD.version+1 THEN RETURN NEW; END IF;
 RAISE EXCEPTION 'Diagnostic definitions are retained; create a revision' USING ERRCODE='23514';
END $$;
CREATE TRIGGER immutable_diagnostic_test BEFORE UPDATE OR DELETE ON diagnostic_test FOR EACH ROW EXECUTE FUNCTION retain_diagnostic_definition();
CREATE TABLE diagnostic_catalog_event (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,test_id uuid NOT NULL REFERENCES diagnostic_test(id),
 replacement_id uuid REFERENCES diagnostic_test(id),actor_id text NOT NULL REFERENCES "user"(id),
 reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 1000),created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TRIGGER immutable_diagnostic_catalog_event BEFORE UPDATE OR DELETE ON diagnostic_catalog_event FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
