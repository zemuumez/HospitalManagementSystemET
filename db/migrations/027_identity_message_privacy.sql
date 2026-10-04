ALTER TABLE message_outbox ADD COLUMN audience text NOT NULL DEFAULT 'operational' CHECK(audience IN ('operational','identity'));
UPDATE message_outbox SET audience='identity' WHERE idempotency_key LIKE 'invitation:%';
CREATE INDEX operational_message_list ON message_outbox(created_at DESC) WHERE audience='operational';
