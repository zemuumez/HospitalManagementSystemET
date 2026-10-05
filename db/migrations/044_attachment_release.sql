ALTER TABLE secure_attachment ADD COLUMN patient_released boolean NOT NULL DEFAULT false;
CREATE TABLE attachment_release_event (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 attachment_id uuid NOT NULL REFERENCES secure_attachment(id),
 actor_id text NOT NULL REFERENCES "user"(id),
 created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER immutable_attachment_release BEFORE UPDATE OR DELETE ON attachment_release_event
 FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_attachment_release_truncate BEFORE TRUNCATE ON attachment_release_event
 FOR EACH STATEMENT EXECUTE FUNCTION protect_retained_record();

-- A single stored file cannot be attached to several orders and released through one.
CREATE UNIQUE INDEX diagnostic_report_file_storage_unique ON diagnostic_report_file(file_url);
