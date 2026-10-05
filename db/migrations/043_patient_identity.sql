-- Identity consolidation retains all original patient/MRN and ledger references.
CREATE TABLE patient_identity_alias (
 alias_id uuid PRIMARY KEY REFERENCES patient(id),
 primary_id uuid NOT NULL REFERENCES patient(id),
 merge_event_id uuid NOT NULL REFERENCES patient_merge_event(id),
 CHECK(alias_id<>primary_id)
);
CREATE INDEX patient_identity_primary ON patient_identity_alias(primary_id);
CREATE FUNCTION canonical_patient_id(value uuid) RETURNS uuid LANGUAGE sql STABLE AS $$
 SELECT COALESCE((SELECT primary_id FROM patient_identity_alias WHERE alias_id=value),value)
$$;
CREATE FUNCTION patient_portal_owner(value uuid) RETURNS text LANGUAGE sql STABLE AS $$
 SELECT p.user_id FROM patient p WHERE canonical_patient_id(p.id)=canonical_patient_id(value) AND p.user_id IS NOT NULL LIMIT 1
$$;
CREATE FUNCTION reject_alias_patient_write() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE identifier uuid;
BEGIN
 IF (to_jsonb(NEW)->>TG_ARGV[0]) !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$' THEN RETURN NEW; END IF;
 identifier := (to_jsonb(NEW)->>TG_ARGV[0])::uuid;
 PERFORM id FROM patient WHERE id=identifier FOR KEY SHARE;
 IF identifier IS NOT NULL AND canonical_patient_id(identifier)<>identifier THEN
   RAISE EXCEPTION 'Use the primary patient identifier' USING ERRCODE='23514';
 END IF;
 RETURN NEW;
END $$;
-- Block new work against retired identifiers, without rewriting historical rows.
DO $$ DECLARE r record; BEGIN
 FOR r IN SELECT DISTINCT c.table_name,c.column_name FROM information_schema.columns c
 WHERE c.table_schema=current_schema() AND c.column_name='patient_id' AND c.udt_name IN ('uuid','text')
 AND c.table_name NOT IN ('patient_identity_alias','patient_merge_event') LOOP
 EXECUTE format('CREATE TRIGGER reject_alias_patient_insert BEFORE INSERT ON %I FOR EACH ROW EXECUTE FUNCTION reject_alias_patient_write(''patient_id'')',r.table_name);
 END LOOP;
END $$;
CREATE TRIGGER reject_alias_patient_update BEFORE UPDATE ON patient
 FOR EACH ROW EXECUTE FUNCTION reject_alias_patient_write('id');
CREATE TRIGGER reject_alias_profile_update BEFORE UPDATE ON patient_profile
 FOR EACH ROW EXECUTE FUNCTION reject_alias_patient_write('patient_id');
CREATE TRIGGER retained_patient_merge_truncate BEFORE TRUNCATE ON patient_merge_event
 FOR EACH STATEMENT EXECUTE FUNCTION protect_retained_record();
