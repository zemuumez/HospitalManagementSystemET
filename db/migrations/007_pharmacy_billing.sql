-- Historical dispenses without a captured price must not be priced retroactively.
ALTER TABLE pharmacy_movement ADD COLUMN unit_sale_price_minor bigint
 CHECK(unit_sale_price_minor BETWEEN 0 AND 1000000000);
CREATE TABLE pharmacy_invoice (
 movement_id uuid PRIMARY KEY REFERENCES pharmacy_movement(id),
 invoice_id uuid NOT NULL UNIQUE REFERENCES invoice(id),
 account_id uuid NOT NULL REFERENCES charge_account(id),
 discount_basis_points integer NOT NULL CHECK(discount_basis_points BETWEEN 0 AND 10000),
 created_by text NOT NULL REFERENCES "user"(id),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TRIGGER immutable_pharmacy_invoice BEFORE UPDATE OR DELETE ON pharmacy_invoice
 FOR EACH ROW EXECUTE FUNCTION protect_pharmacy_history();
