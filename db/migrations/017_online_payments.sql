ALTER TABLE invoice_payment DROP CONSTRAINT invoice_payment_method_check;
ALTER TABLE invoice_payment ADD CONSTRAINT invoice_payment_method_check CHECK(method IN ('cash','bank','stripe'));
CREATE UNIQUE INDEX payment_stripe_reference ON invoice_payment(reference) WHERE method='stripe';
CREATE TABLE payment_checkout (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 invoice_id uuid NOT NULL UNIQUE REFERENCES invoice(id),
 actor_id text NOT NULL REFERENCES "user"(id),
 amount_minor bigint NOT NULL CHECK(amount_minor BETWEEN 1 AND 99999999),
 provider_id text NOT NULL DEFAULT '',
 state text NOT NULL DEFAULT 'pending' CHECK(state IN ('pending','ready','paid','review')),
 created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE UNIQUE INDEX payment_checkout_provider ON payment_checkout(provider_id) WHERE provider_id<>'';
CREATE TABLE payment_event (
 event_id text PRIMARY KEY,
 payload_hash text NOT NULL,
 intent_id text NOT NULL DEFAULT '',
 state text NOT NULL CHECK(state IN ('applied','review','ignored')),
 reason text NOT NULL,
 received_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX payment_event_review ON payment_event(received_at DESC) WHERE state='review';
CREATE TRIGGER immutable_payment_event BEFORE UPDATE OR DELETE ON payment_event FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TABLE payment_receipt (
 intent_id text PRIMARY KEY,
 checkout_id uuid NOT NULL UNIQUE REFERENCES payment_checkout(id),
 payment_id uuid NOT NULL UNIQUE REFERENCES invoice_payment(id),
 event_id text NOT NULL REFERENCES payment_event(event_id)
);
CREATE TRIGGER immutable_payment_receipt BEFORE UPDATE OR DELETE ON payment_receipt FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
