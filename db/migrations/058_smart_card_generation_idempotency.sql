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
