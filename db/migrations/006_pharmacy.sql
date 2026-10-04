CREATE TABLE medicine (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 name text NOT NULL UNIQUE CHECK(length(name) BETWEEN 2 AND 150),
 category text NOT NULL CHECK(length(category) BETWEEN 1 AND 100),
 brand text NOT NULL CHECK(length(brand) BETWEEN 1 AND 100),
 unit text NOT NULL CHECK(length(unit) BETWEEN 1 AND 40),
 composition text NOT NULL DEFAULT '' CHECK(length(composition)<=2000),
 side_effects text NOT NULL DEFAULT '' CHECK(length(side_effects)<=2000),
 selling_price_minor bigint NOT NULL CHECK(selling_price_minor BETWEEN 0 AND 1000000000),
 active boolean NOT NULL DEFAULT true,
 created_by text NOT NULL REFERENCES "user"(id)
);
CREATE TABLE medicine_batch (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 medicine_id uuid NOT NULL REFERENCES medicine(id),
 lot text NOT NULL CHECK(length(lot) BETWEEN 1 AND 100),
 expiry_date date NOT NULL,
 supplier text NOT NULL CHECK(length(supplier) BETWEEN 1 AND 200),
 purchase_reference text NOT NULL CHECK(length(purchase_reference) BETWEEN 1 AND 100),
 unit_cost_minor bigint NOT NULL CHECK(unit_cost_minor BETWEEN 0 AND 1000000000),
 received_quantity integer NOT NULL CHECK(received_quantity BETWEEN 1 AND 1000000),
 balance integer NOT NULL DEFAULT 0 CHECK(balance BETWEEN 0 AND 1000000),
 created_by text NOT NULL REFERENCES "user"(id),
 request_key text NOT NULL, request_hash text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(created_by,request_key),
 UNIQUE(medicine_id,lot,supplier,purchase_reference)
);
CREATE TABLE medication_order (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 encounter_id uuid NOT NULL REFERENCES encounter(id),
 medicine_id uuid NOT NULL REFERENCES medicine(id),
 medicine_name text NOT NULL,
 unit text NOT NULL,
 quantity integer NOT NULL CHECK(quantity BETWEEN 1 AND 10000),
 dose text NOT NULL CHECK(length(dose) BETWEEN 1 AND 100),
 route text NOT NULL CHECK(length(route) BETWEEN 1 AND 100),
 frequency text NOT NULL CHECK(length(frequency) BETWEEN 1 AND 100),
 duration_days integer NOT NULL CHECK(duration_days BETWEEN 1 AND 365),
 instructions text NOT NULL DEFAULT '' CHECK(length(instructions)<=2000),
 signed_by text NOT NULL REFERENCES "user"(id),
 signed_at timestamptz NOT NULL DEFAULT now(),
 request_key text NOT NULL, request_hash text NOT NULL,
 UNIQUE(signed_by,request_key)
);
CREATE TABLE medication_cancellation (
 order_id uuid PRIMARY KEY REFERENCES medication_order(id),
 reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 1000),
 actor_id text NOT NULL REFERENCES "user"(id),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE pharmacy_movement (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 batch_id uuid NOT NULL REFERENCES medicine_batch(id),
 order_id uuid REFERENCES medication_order(id),
 kind text NOT NULL CHECK(kind IN ('receipt','dispense','return','disposal')),
 quantity integer NOT NULL CHECK(quantity BETWEEN 1 AND 1000000),
 original_id uuid REFERENCES pharmacy_movement(id),
 reason text NOT NULL DEFAULT '' CHECK(length(reason)<=1000),
 actor_id text NOT NULL REFERENCES "user"(id),
 request_key text NOT NULL, request_hash text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(actor_id,request_key),
 CHECK((kind='dispense' AND order_id IS NOT NULL AND original_id IS NULL) OR
       (kind='return' AND order_id IS NOT NULL AND original_id IS NOT NULL AND length(reason)>0) OR
       (kind IN ('receipt','disposal') AND order_id IS NULL AND original_id IS NULL)),
 CHECK(kind<>'disposal' OR length(reason)>0)
);
CREATE UNIQUE INDEX batch_receipt_once ON pharmacy_movement(batch_id) WHERE kind='receipt';
CREATE INDEX pharmacy_order_movements ON pharmacy_movement(order_id,created_at);
CREATE INDEX pharmacy_batch_movements ON pharmacy_movement(batch_id,created_at);
CREATE INDEX medication_order_encounter ON medication_order(encounter_id,signed_at);
CREATE INDEX pharmacy_return_source ON pharmacy_movement(original_id) WHERE kind='return';
CREATE FUNCTION protect_pharmacy_history() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 RAISE EXCEPTION 'Pharmacy history is append-only' USING ERRCODE='23514';
END $$;
CREATE TRIGGER immutable_medication BEFORE UPDATE OR DELETE ON medication_order FOR EACH ROW EXECUTE FUNCTION protect_pharmacy_history();
CREATE TRIGGER immutable_cancellation BEFORE UPDATE OR DELETE ON medication_cancellation FOR EACH ROW EXECUTE FUNCTION protect_pharmacy_history();
CREATE TRIGGER immutable_stock_movement BEFORE UPDATE OR DELETE ON pharmacy_movement FOR EACH ROW EXECUTE FUNCTION protect_pharmacy_history();
CREATE FUNCTION protect_batch_snapshot() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP='DELETE' THEN RAISE EXCEPTION 'Batch history cannot be deleted' USING ERRCODE='23514'; END IF;
 IF (to_jsonb(NEW)-'balance') IS DISTINCT FROM (to_jsonb(OLD)-'balance') THEN
  RAISE EXCEPTION 'Batch receipt snapshot cannot be changed' USING ERRCODE='23514';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER immutable_batch BEFORE UPDATE OR DELETE ON medicine_batch FOR EACH ROW EXECUTE FUNCTION protect_batch_snapshot();

-- The materialized balance must reconcile with immutable movements at commit.
CREATE FUNCTION reconcile_pharmacy_batch() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE target uuid; actual integer; expected bigint; received bigint;
BEGIN
 IF TG_TABLE_NAME='medicine_batch' THEN target=NEW.id; ELSE target=NEW.batch_id; END IF;
 SELECT balance INTO actual FROM medicine_batch WHERE id=target;
 SELECT sum(CASE kind WHEN 'receipt' THEN quantity WHEN 'return' THEN 0 ELSE -quantity END),
 sum(CASE WHEN kind='receipt' THEN quantity ELSE 0 END) INTO expected,received
 FROM pharmacy_movement WHERE batch_id=target;
 IF expected IS NULL OR actual<>expected OR received<>(SELECT received_quantity FROM medicine_batch WHERE id=target) THEN
  RAISE EXCEPTION 'Stock balance must match movement history' USING ERRCODE='23514';
 END IF;
 RETURN NULL;
END $$;
CREATE CONSTRAINT TRIGGER batch_reconciled AFTER INSERT OR UPDATE ON medicine_batch
 DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reconcile_pharmacy_batch();
CREATE CONSTRAINT TRIGGER movement_reconciled AFTER INSERT ON pharmacy_movement
 DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reconcile_pharmacy_batch();
