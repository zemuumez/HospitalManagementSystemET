-- Migration 058: Smart Card Generation Request Idempotency
-- Durable tracking of card generation requests to guarantee idempotent replay,
-- prevent duplicate credential issuance, and protect against un-revocation.

CREATE TABLE IF NOT EXISTS smart_card_generation_request (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id        text NOT NULL REFERENCES "user"(id),
  idempotency_key text NOT NULL,
  template_id     uuid NOT NULL,
  mode            text NOT NULL CHECK (mode IN ('all', 'single', 'unassigned')),
  patient_id      uuid REFERENCES patient(id) ON DELETE SET NULL,
  response_json   jsonb NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE(actor_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_smart_card_gen_req_actor_key ON smart_card_generation_request(actor_id, idempotency_key);

-- If smart_card_generation_request was previously created by intermediate migration 057
-- with an ON DELETE CASCADE foreign key on template_id, drop that constraint:
DO $$
DECLARE
    r RECORD;
    v_tbl_oid oid;
BEGIN
    SELECT oid INTO v_tbl_oid FROM pg_class WHERE relname = 'smart_card_generation_request' AND relnamespace = (
        SELECT n.oid FROM pg_namespace n WHERE n.nspname = current_schema()
    );
    IF v_tbl_oid IS NOT NULL THEN
        FOR r IN (
            SELECT conname
            FROM pg_constraint
            WHERE conrelid = v_tbl_oid
              AND contype = 'f'
              AND conkey = ARRAY[(
                  SELECT attnum
                  FROM pg_attribute
                  WHERE attrelid = v_tbl_oid AND attname = 'template_id'
              )]
        ) LOOP
            EXECUTE 'ALTER TABLE smart_card_generation_request DROP CONSTRAINT ' || quote_ident(r.conname);
        END LOOP;
    END IF;
END $$;

ALTER TABLE IF EXISTS smart_card_generation_request
  DROP CONSTRAINT IF EXISTS smart_card_generation_request_template_id_fkey;
