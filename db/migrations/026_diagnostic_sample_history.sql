ALTER TABLE diagnostic_order DROP CONSTRAINT diagnostic_order_status_check;
ALTER TABLE diagnostic_order ADD CONSTRAINT diagnostic_order_status_check CHECK(status IN ('ordered','collected','processing','review','signed','released','cancelled','sample_rejected'));
CREATE TABLE diagnostic_sample (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),order_id uuid NOT NULL REFERENCES diagnostic_order(id),
 reference text NOT NULL UNIQUE CHECK(length(reference) BETWEEN 1 AND 100),
 collected_by text REFERENCES "user"(id),collected_at timestamptz,
 order_version integer NOT NULL CHECK(order_version>0),UNIQUE(order_id,order_version)
);
INSERT INTO diagnostic_sample(order_id,reference,collected_by,collected_at,order_version)
 SELECT o.id,o.sample_reference,e.actor_id,e.created_at,COALESCE(e.version,1)
 FROM diagnostic_order o LEFT JOIN LATERAL(SELECT actor_id,created_at,version FROM diagnostic_event WHERE order_id=o.id AND to_status='collected' ORDER BY version LIMIT 1)e ON true WHERE o.sample_reference<>'';
CREATE TABLE diagnostic_sample_rejection (
 sample_id uuid PRIMARY KEY REFERENCES diagnostic_sample(id),actor_id text NOT NULL REFERENCES "user"(id),
 reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 1000),rejected_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER immutable_diagnostic_sample BEFORE UPDATE OR DELETE ON diagnostic_sample FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_diagnostic_sample_rejection BEFORE UPDATE OR DELETE ON diagnostic_sample_rejection FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
