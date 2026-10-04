CREATE TABLE charge_account (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 name text NOT NULL UNIQUE CHECK(length(name) BETWEEN 1 AND 100),
 active boolean NOT NULL DEFAULT true,
 created_by text NOT NULL REFERENCES "user"(id)
);
CREATE TABLE invoice (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 number bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
 patient_id uuid NOT NULL REFERENCES patient(id),
 invoice_date date NOT NULL,
 currency text NOT NULL DEFAULT 'ETB' CHECK(currency='ETB'),
 subtotal_minor bigint NOT NULL CHECK(subtotal_minor BETWEEN 1 AND 1000000000000),
 discount_basis_points integer NOT NULL CHECK(discount_basis_points BETWEEN 0 AND 10000),
 total_minor bigint NOT NULL CHECK(total_minor BETWEEN 1 AND 1000000000000),
 paid_minor bigint NOT NULL DEFAULT 0 CHECK(paid_minor>=0 AND paid_minor<=total_minor),
 version integer NOT NULL DEFAULT 1,
 request_hash text NOT NULL,
 request_key text NOT NULL,
 created_by text NOT NULL REFERENCES "user"(id),
 created_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(created_by,request_key)
);
CREATE TABLE invoice_line (
 invoice_id uuid NOT NULL REFERENCES invoice(id),
 position integer NOT NULL CHECK(position BETWEEN 1 AND 100),
 account_id uuid NOT NULL REFERENCES charge_account(id),
 account_name text NOT NULL,
 description text NOT NULL CHECK(length(description)<=500),
 quantity integer NOT NULL CHECK(quantity BETWEEN 1 AND 10000),
 unit_price_minor bigint NOT NULL CHECK(unit_price_minor BETWEEN 0 AND 1000000000),
 PRIMARY KEY(invoice_id,position)
);
CREATE TABLE invoice_payment (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 invoice_id uuid NOT NULL REFERENCES invoice(id),
 amount_minor bigint NOT NULL CHECK(amount_minor>0),
 direction text NOT NULL CHECK(direction IN ('payment','refund')),
 method text NOT NULL CHECK(method IN ('cash','bank')),
 reference text NOT NULL DEFAULT '' CHECK(length(reference)<=100),
 reason text NOT NULL DEFAULT '' CHECK(length(reason)<=500),
 original_payment_id uuid REFERENCES invoice_payment(id),
 actor_id text NOT NULL REFERENCES "user"(id),
 request_key text NOT NULL,
 request_hash text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 CHECK((direction='payment' AND original_payment_id IS NULL) OR (direction='refund' AND original_payment_id IS NOT NULL AND length(reason)>0)),
 CHECK(method='cash' OR length(reference)>0),
 UNIQUE(actor_id,request_key)
);
CREATE UNIQUE INDEX payment_bank_reference ON invoice_payment(reference) WHERE method='bank';
CREATE INDEX invoice_payment_invoice ON invoice_payment(invoice_id,created_at);
CREATE INDEX invoice_patient ON invoice(patient_id,created_at DESC);
CREATE TRIGGER immutable_payment BEFORE UPDATE OR DELETE ON invoice_payment FOR EACH ROW EXECUTE FUNCTION protect_signed_note();
