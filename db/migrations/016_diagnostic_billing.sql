CREATE TABLE diagnostic_invoice (
 order_id uuid PRIMARY KEY REFERENCES diagnostic_order(id),
 invoice_id uuid NOT NULL UNIQUE REFERENCES invoice(id),
 account_id uuid NOT NULL REFERENCES charge_account(id),
 discount_basis_points integer NOT NULL CHECK(discount_basis_points BETWEEN 0 AND 9999),
 created_by text NOT NULL REFERENCES "user"(id),
 created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER immutable_diagnostic_invoice BEFORE UPDATE OR DELETE ON diagnostic_invoice FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
