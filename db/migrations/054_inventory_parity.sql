-- Migration 054: Inventory Workspace Parity Extensions
-- Adds source fields for receipt attachments, business dates, staff issuer, department,
-- and linked receipt void reversal support with unique index protection.

ALTER TABLE inventory_movement
  ADD COLUMN IF NOT EXISTS attachment_url text NOT NULL DEFAULT '' CHECK (length(attachment_url) <= 255),
  ADD COLUMN IF NOT EXISTS issued_date text NOT NULL DEFAULT '' CHECK (length(issued_date) <= 30),
  ADD COLUMN IF NOT EXISTS return_due_date text NOT NULL DEFAULT '' CHECK (length(return_due_date) <= 30),
  ADD COLUMN IF NOT EXISTS issued_by text NOT NULL DEFAULT '' CHECK (length(issued_by) <= 128),
  ADD COLUMN IF NOT EXISTS department text NOT NULL DEFAULT '' CHECK (length(department) <= 100);

ALTER TABLE inventory_movement DROP CONSTRAINT IF EXISTS inventory_movement_kind_check;
ALTER TABLE inventory_movement DROP CONSTRAINT IF EXISTS inventory_movement_check;

ALTER TABLE inventory_movement ADD CONSTRAINT inventory_movement_kind_check
  CHECK (kind = ANY (ARRAY['receive'::text, 'issue'::text, 'return'::text, 'writeoff'::text, 'void_receipt'::text]));

ALTER TABLE inventory_movement ADD CONSTRAINT inventory_movement_check
  CHECK ((((kind = 'receive'::text) AND (delta_milli = quantity_milli) AND (length(reference) > 0) AND (recipient_id IS NULL) AND (original_id IS NULL)) OR
          ((kind = 'issue'::text) AND (delta_milli = (- quantity_milli)) AND (recipient_id IS NOT NULL) AND (original_id IS NULL)) OR
          ((kind = 'return'::text) AND (original_id IS NOT NULL) AND (recipient_id IS NOT NULL) AND (delta_milli = CASE WHEN restock THEN quantity_milli ELSE (0)::bigint END)) OR
          ((kind = 'writeoff'::text) AND (delta_milli = (- quantity_milli)) AND (recipient_id IS NULL)) OR
          ((kind = 'void_receipt'::text) AND (delta_milli = (- quantity_milli)) AND (recipient_id IS NULL) AND (original_id IS NOT NULL))));

CREATE UNIQUE INDEX IF NOT EXISTS inventory_movement_void_receipt ON inventory_movement(original_id) WHERE kind='void_receipt';
