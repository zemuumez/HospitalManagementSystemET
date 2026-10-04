-- Existing invoices are issued. New invoices must be sealed in the same
-- transaction as their lines; an unfinished invoice cannot be committed.
ALTER TABLE invoice ADD COLUMN sealed boolean NOT NULL DEFAULT true;
ALTER TABLE invoice ALTER COLUMN sealed SET DEFAULT false;

CREATE FUNCTION protect_issued_invoice() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_OP = 'DELETE' THEN
  RAISE EXCEPTION 'Issued invoices cannot be deleted' USING ERRCODE = '23514';
 END IF;
 IF OLD.sealed AND (
   (to_jsonb(NEW) - 'paid_minor' - 'version') IS DISTINCT FROM
   (to_jsonb(OLD) - 'paid_minor' - 'version')
 ) THEN
  RAISE EXCEPTION 'Issued invoice snapshots cannot be changed' USING ERRCODE = '23514';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER protect_invoice BEFORE UPDATE OR DELETE ON invoice
 FOR EACH ROW EXECUTE FUNCTION protect_issued_invoice();

CREATE FUNCTION protect_invoice_line() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE issued boolean;
BEGIN
 IF TG_OP <> 'INSERT' THEN
  RAISE EXCEPTION 'Invoice lines cannot be changed or deleted' USING ERRCODE = '23514';
 END IF;
 SELECT sealed INTO issued FROM invoice WHERE id=NEW.invoice_id FOR UPDATE;
 IF issued THEN
  RAISE EXCEPTION 'Cannot add lines to an issued invoice' USING ERRCODE = '23514';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER protect_invoice_lines BEFORE INSERT OR UPDATE OR DELETE ON invoice_line
 FOR EACH ROW EXECUTE FUNCTION protect_invoice_line();

CREATE FUNCTION validate_issued_invoice() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE row_invoice invoice%ROWTYPE; line_total bigint;
BEGIN
 SELECT * INTO row_invoice FROM invoice WHERE id=NEW.id;
 SELECT sum(quantity::bigint*unit_price_minor) INTO line_total
 FROM invoice_line WHERE invoice_id=NEW.id;
 IF NOT row_invoice.sealed OR line_total IS NULL OR line_total <> row_invoice.subtotal_minor
 OR row_invoice.total_minor <> line_total - (line_total*row_invoice.discount_basis_points+5000)/10000 THEN
  RAISE EXCEPTION 'Invoice must be sealed with matching line totals' USING ERRCODE = '23514';
 END IF;
 RETURN NULL;
END $$;
CREATE CONSTRAINT TRIGGER invoice_complete AFTER INSERT OR UPDATE ON invoice
 DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION validate_issued_invoice();
