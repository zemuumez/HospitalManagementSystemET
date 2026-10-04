CREATE TABLE diagnostic_test (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 kind text NOT NULL CHECK(kind IN ('pathology','radiology')),
 name text NOT NULL CHECK(length(name) BETWEEN 1 AND 150),
 short_name text NOT NULL CHECK(length(short_name) BETWEEN 1 AND 50),
 category text NOT NULL CHECK(length(category) BETWEEN 1 AND 100),
 method text NOT NULL DEFAULT '' CHECK(length(method)<=200),
 report_days integer NOT NULL CHECK(report_days BETWEEN 0 AND 365),
 charge_minor bigint NOT NULL CHECK(charge_minor BETWEEN 0 AND 1000000000),
 created_by text NOT NULL REFERENCES "user"(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(kind,name)
);
CREATE TABLE diagnostic_parameter (
 test_id uuid NOT NULL REFERENCES diagnostic_test(id),
 position integer NOT NULL CHECK(position BETWEEN 1 AND 50),
 name text NOT NULL CHECK(length(name) BETWEEN 1 AND 150),
 unit text NOT NULL DEFAULT '' CHECK(length(unit)<=40),
 reference_range text NOT NULL DEFAULT '' CHECK(length(reference_range)<=500),
 value_type text NOT NULL CHECK(value_type IN ('number','text')),
 PRIMARY KEY(test_id,position)
);
CREATE TABLE diagnostic_order (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 accession bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
 encounter_id uuid NOT NULL REFERENCES encounter(id),
 test_id uuid NOT NULL REFERENCES diagnostic_test(id),
 ordered_by text NOT NULL REFERENCES "user"(id),
 indication text NOT NULL CHECK(length(indication) BETWEEN 1 AND 2000),
 status text NOT NULL DEFAULT 'ordered' CHECK(status IN ('ordered','collected','processing','review','signed','released','cancelled')),
 sample_reference text NOT NULL DEFAULT '' CHECK(length(sample_reference)<=100),
 version integer NOT NULL DEFAULT 1,
 request_key text NOT NULL, request_hash text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(ordered_by,request_key)
);
CREATE UNIQUE INDEX diagnostic_sample_reference ON diagnostic_order(sample_reference) WHERE sample_reference<>'';
CREATE INDEX diagnostic_encounter ON diagnostic_order(encounter_id,created_at DESC);
CREATE INDEX diagnostic_worklist ON diagnostic_order(status,created_at DESC);
CREATE TABLE diagnostic_result (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 order_id uuid NOT NULL REFERENCES diagnostic_order(id),
 revision integer NOT NULL,
 summary text NOT NULL CHECK(length(summary) BETWEEN 1 AND 5000),
 amendment_reason text NOT NULL DEFAULT '' CHECK(length(amendment_reason)<=1000),
 created_by text NOT NULL REFERENCES "user"(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(order_id,revision)
);
CREATE TABLE diagnostic_value (
 result_id uuid NOT NULL REFERENCES diagnostic_result(id),
 position integer NOT NULL,
 value text NOT NULL CHECK(length(value) BETWEEN 1 AND 1000),
 PRIMARY KEY(result_id,position)
);
CREATE TABLE diagnostic_review (
 result_id uuid NOT NULL REFERENCES diagnostic_result(id),
 action text NOT NULL CHECK(action IN ('sign','release','reject')),
 actor_id text NOT NULL REFERENCES "user"(id),
 reason text NOT NULL DEFAULT '' CHECK(length(reason)<=1000),
 created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(result_id,action)
);
CREATE TABLE diagnostic_event (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 order_id uuid NOT NULL REFERENCES diagnostic_order(id),
 from_status text NOT NULL,to_status text NOT NULL,
 reason text NOT NULL DEFAULT '' CHECK(length(reason)<=1000),
 actor_id text NOT NULL REFERENCES "user"(id),
 version integer NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(order_id,version)
);
CREATE FUNCTION protect_retained_record() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'Retained records are append-only' USING ERRCODE='23514'; END $$;
CREATE TRIGGER immutable_diagnostic_test BEFORE UPDATE OR DELETE ON diagnostic_test FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_diagnostic_parameter BEFORE UPDATE OR DELETE ON diagnostic_parameter FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_diagnostic_result BEFORE UPDATE OR DELETE ON diagnostic_result FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_diagnostic_value BEFORE UPDATE OR DELETE ON diagnostic_value FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_diagnostic_review BEFORE UPDATE OR DELETE ON diagnostic_review FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_diagnostic_event BEFORE UPDATE OR DELETE ON diagnostic_event FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
-- Late additions to a result would alter its signed content. Values must be
-- inserted in the same transaction as the result, before a deferred seal.
ALTER TABLE diagnostic_result ADD COLUMN sealed boolean NOT NULL DEFAULT false;
DROP TRIGGER immutable_diagnostic_result ON diagnostic_result;
CREATE FUNCTION protect_diagnostic_result() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='UPDATE' AND NOT OLD.sealed AND NEW.sealed AND
 (to_jsonb(NEW)-'sealed')=(to_jsonb(OLD)-'sealed') THEN RETURN NEW; END IF;
 RAISE EXCEPTION 'Result content is immutable' USING ERRCODE='23514';
END $$;
CREATE TRIGGER immutable_diagnostic_result BEFORE UPDATE OR DELETE ON diagnostic_result FOR EACH ROW EXECUTE FUNCTION protect_diagnostic_result();
CREATE FUNCTION check_diagnostic_value_insert() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE locked boolean;
BEGIN
 SELECT sealed INTO locked FROM diagnostic_result WHERE id=NEW.result_id FOR UPDATE;
 IF locked THEN RAISE EXCEPTION 'Cannot append to sealed result' USING ERRCODE='23514'; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER diagnostic_value_insert BEFORE INSERT ON diagnostic_value FOR EACH ROW EXECUTE FUNCTION check_diagnostic_value_insert();
CREATE FUNCTION check_diagnostic_result_complete() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE test uuid;locked boolean; expected integer;actual integer;
BEGIN
 SELECT o.test_id,r.sealed INTO test,locked FROM diagnostic_result r JOIN diagnostic_order o ON o.id=r.order_id WHERE r.id=NEW.id;
 SELECT count(*) INTO expected FROM diagnostic_parameter WHERE test_id=test;
 SELECT count(*) INTO actual FROM diagnostic_value WHERE result_id=NEW.id;
 IF NOT locked OR actual<>expected OR EXISTS(SELECT 1 FROM diagnostic_value v WHERE v.result_id=NEW.id AND NOT EXISTS(SELECT 1 FROM diagnostic_parameter p WHERE p.test_id=test AND p.position=v.position)) THEN
 RAISE EXCEPTION 'Result must be sealed and complete' USING ERRCODE='23514'; END IF;
 RETURN NULL;
END $$;
CREATE CONSTRAINT TRIGGER diagnostic_result_complete AFTER INSERT OR UPDATE ON diagnostic_result DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION check_diagnostic_result_complete();

CREATE FUNCTION protect_ordered_test_parameters() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 PERFORM id FROM diagnostic_test WHERE id=NEW.test_id FOR UPDATE;
 IF EXISTS(SELECT 1 FROM diagnostic_order WHERE test_id=NEW.test_id) THEN
 RAISE EXCEPTION 'Ordered test parameters cannot be extended' USING ERRCODE='23514'; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER protect_ordered_parameters BEFORE INSERT ON diagnostic_parameter FOR EACH ROW EXECUTE FUNCTION protect_ordered_test_parameters();
