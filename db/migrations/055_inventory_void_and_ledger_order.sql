-- Migration 055: Inventory Void Receipts, Sequence Ordering & Balance History
-- 1. Updates movement constraints to allow void_receipt
-- 2. Creates unique partial index for void_receipt reversals
-- 3. Adds deterministic per-item ledger sequence (ledger_seq) and running balance (balance_after_milli)
-- 4. Backfills existing movement records with deterministic sequence and balance

ALTER TABLE inventory_movement DROP CONSTRAINT IF EXISTS inventory_movement_kind_check;
ALTER TABLE inventory_movement ADD CONSTRAINT inventory_movement_kind_check
  CHECK (kind IN ('receive', 'issue', 'return', 'writeoff', 'void_receipt'));

ALTER TABLE inventory_movement DROP CONSTRAINT IF EXISTS inventory_movement_check;
ALTER TABLE inventory_movement ADD CONSTRAINT inventory_movement_check
  CHECK (
    (kind = 'receive' AND delta_milli = quantity_milli AND length(reference) > 0 AND recipient_id IS NULL AND original_id IS NULL) OR
    (kind = 'issue' AND delta_milli = -quantity_milli AND recipient_id IS NOT NULL AND original_id IS NULL) OR
    (kind = 'return' AND original_id IS NOT NULL AND recipient_id IS NOT NULL AND delta_milli = CASE WHEN restock THEN quantity_milli ELSE 0 END) OR
    (kind = 'writeoff' AND delta_milli = -quantity_milli AND recipient_id IS NULL AND original_id IS NULL) OR
    (kind = 'void_receipt' AND delta_milli = -quantity_milli AND recipient_id IS NULL AND original_id IS NOT NULL)
  );

CREATE UNIQUE INDEX IF NOT EXISTS inventory_movement_void_receipt
  ON inventory_movement(original_id)
  WHERE kind = 'void_receipt';

ALTER TABLE inventory_movement
  ADD COLUMN IF NOT EXISTS ledger_seq bigint NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS balance_after_milli bigint NOT NULL DEFAULT 0;

-- Backfill preexisting movements with deterministic ledger sequence and running balance.
-- Deterministic order handles timestamp ties: receipts (+delta) precede issues (-delta) so balance never goes negative.
ALTER TABLE inventory_movement DISABLE TRIGGER immutable_inventory_movement;

WITH ordered AS (
  SELECT id,
         row_number() OVER (
           PARTITION BY item_id
           ORDER BY created_at ASC,
                    CASE WHEN delta_milli > 0 THEN 0 ELSE 1 END,
                    id ASC
         ) AS seq,
         sum(delta_milli) OVER (
           PARTITION BY item_id
           ORDER BY created_at ASC,
                    CASE WHEN delta_milli > 0 THEN 0 ELSE 1 END,
                    id ASC
         ) AS bal
  FROM inventory_movement
)
UPDATE inventory_movement m
SET ledger_seq = o.seq,
    balance_after_milli = o.bal
FROM ordered o
WHERE m.id = o.id AND (m.ledger_seq = 0 OR m.balance_after_milli = 0);

ALTER TABLE inventory_movement ENABLE TRIGGER immutable_inventory_movement;

CREATE INDEX IF NOT EXISTS inventory_movement_item_ledger_seq
  ON inventory_movement(item_id, ledger_seq);
