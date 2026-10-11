-- Migration 059: Smart Card Generation Request FK Cleanup
-- Explicitly remove any foreign key constraint on smart_card_generation_request.template_id
-- referencing patient_id_card_template (such as ON DELETE CASCADE from intermediate migration 057)
-- so that completed generation request records survive template deletion, preserving immutable
-- audit history and idempotency conflict checking.

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
