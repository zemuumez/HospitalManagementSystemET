-- Migration 056: Inventory Legacy Ledger Cutover and Inferred History Remediation
-- 1. Adds legacy_unverified column to explicitly flag pre-cutover unverified movements
-- 2. Corrects any previously backfilled sequences/balances from migration 055 (from commit 8cc47e7)
--    by resetting existing records to ledger_seq = 0 and balance_after_milli = 0
-- 3. Creates inventory_ledger_cutover tracking table to record cutover provenance
-- 4. Preserves all item balances, movement records, quantities, and financial amounts

ALTER TABLE inventory_movement
  ADD COLUMN IF NOT EXISTS legacy_unverified boolean NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS inventory_ledger_cutover (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  cutover_at  timestamptz NOT NULL DEFAULT now(),
  description text NOT NULL
);

INSERT INTO inventory_ledger_cutover (description)
VALUES ('Migration 056 remediation: reset inferred legacy ledger sequences and established lock-serialized ledger provenance.');

-- For any databases that executed the earlier migration 055 backfill, reset inferred
-- sequences and running balances to 0, and mark legacy_unverified = true.
ALTER TABLE inventory_movement DISABLE TRIGGER immutable_inventory_movement;

UPDATE inventory_movement
SET ledger_seq = 0,
    balance_after_milli = 0,
    legacy_unverified = true;

ALTER TABLE inventory_movement ENABLE TRIGGER immutable_inventory_movement;
