CREATE TABLE inventory_category (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),name text NOT NULL UNIQUE CHECK(length(name) BETWEEN 1 AND 100),
 description text NOT NULL DEFAULT '' CHECK(length(description)<=1000),active boolean NOT NULL DEFAULT true,version integer NOT NULL DEFAULT 1
);
CREATE TABLE inventory_item (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),category_id uuid NOT NULL REFERENCES inventory_category(id),
 name text NOT NULL UNIQUE CHECK(length(name) BETWEEN 1 AND 150),unit text NOT NULL CHECK(length(unit) BETWEEN 1 AND 40),
 description text NOT NULL DEFAULT '' CHECK(length(description)<=2000),
 reorder_milli bigint NOT NULL DEFAULT 0 CHECK(reorder_milli BETWEEN 0 AND 1000000000000),
 balance_milli bigint NOT NULL DEFAULT 0 CHECK(balance_milli BETWEEN 0 AND 1000000000000),
 active boolean NOT NULL DEFAULT true,version integer NOT NULL DEFAULT 1
);
CREATE TABLE inventory_movement (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),item_id uuid NOT NULL REFERENCES inventory_item(id),
 kind text NOT NULL CHECK(kind IN ('receive','issue','return','writeoff')),
 quantity_milli bigint NOT NULL CHECK(quantity_milli BETWEEN 1 AND 1000000000000),
 delta_milli bigint NOT NULL,
 recipient_id text REFERENCES staff_access(user_id),original_id uuid REFERENCES inventory_movement(id),
 supplier text NOT NULL DEFAULT '' CHECK(length(supplier)<=200),store_name text NOT NULL DEFAULT '' CHECK(length(store_name)<=100),
 reference text NOT NULL DEFAULT '' CHECK(length(reference)<=100),
 cost_minor bigint NOT NULL DEFAULT 0 CHECK(cost_minor BETWEEN 0 AND 1000000000000),
 restock boolean NOT NULL DEFAULT false,reason text NOT NULL CHECK(length(reason) BETWEEN 1 AND 1000),
 actor_id text NOT NULL REFERENCES "user"(id),request_key text NOT NULL,request_hash text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),UNIQUE(actor_id,request_key),
 CHECK((kind='receive' AND delta_milli=quantity_milli AND length(reference)>0 AND recipient_id IS NULL AND original_id IS NULL) OR
 (kind='issue' AND delta_milli=-quantity_milli AND recipient_id IS NOT NULL AND original_id IS NULL) OR
 (kind='return' AND original_id IS NOT NULL AND recipient_id IS NOT NULL AND delta_milli=CASE WHEN restock THEN quantity_milli ELSE 0 END) OR
 (kind='writeoff' AND delta_milli=-quantity_milli AND recipient_id IS NULL AND original_id IS NULL)),
 CHECK(kind='return' OR NOT restock)
);
CREATE UNIQUE INDEX inventory_receipt_reference ON inventory_movement(item_id,supplier,reference) WHERE kind='receive';
CREATE INDEX inventory_item_history ON inventory_movement(item_id,created_at DESC);
CREATE INDEX inventory_original_return ON inventory_movement(original_id) WHERE kind='return';
CREATE TRIGGER immutable_inventory_movement BEFORE UPDATE OR DELETE ON inventory_movement FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE FUNCTION reconcile_inventory() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE target uuid;expected bigint;actual bigint;
BEGIN
 IF TG_TABLE_NAME='inventory_item' THEN target=NEW.id; ELSE target=NEW.item_id; END IF;
 SELECT COALESCE(sum(delta_milli),0) INTO expected FROM inventory_movement WHERE item_id=target;
 SELECT balance_milli INTO actual FROM inventory_item WHERE id=target;
 IF actual<>expected THEN RAISE EXCEPTION 'Inventory balance does not match movement ledger' USING ERRCODE='23514'; END IF;
 RETURN NULL;
END $$;
CREATE CONSTRAINT TRIGGER inventory_item_reconciled AFTER INSERT OR UPDATE ON inventory_item DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reconcile_inventory();
CREATE CONSTRAINT TRIGGER inventory_movement_reconciled AFTER INSERT ON inventory_movement DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION reconcile_inventory();
