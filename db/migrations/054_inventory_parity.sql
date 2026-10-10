-- Migration 054: Inventory Workspace Parity Extensions
-- Adds source fields for receipt attachments, business dates, staff issuer, and department.

ALTER TABLE inventory_movement
  ADD COLUMN IF NOT EXISTS attachment_url text NOT NULL DEFAULT '' CHECK (length(attachment_url) <= 255),
  ADD COLUMN IF NOT EXISTS issued_date text NOT NULL DEFAULT '' CHECK (length(issued_date) <= 30),
  ADD COLUMN IF NOT EXISTS return_due_date text NOT NULL DEFAULT '' CHECK (length(return_due_date) <= 30),
  ADD COLUMN IF NOT EXISTS issued_by text NOT NULL DEFAULT '' CHECK (length(issued_by) <= 128),
  ADD COLUMN IF NOT EXISTS department text NOT NULL DEFAULT '' CHECK (length(department) <= 100);
