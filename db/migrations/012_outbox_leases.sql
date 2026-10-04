ALTER TABLE message_outbox ADD COLUMN lease_token uuid;
ALTER TABLE message_outbox ADD COLUMN lease_until timestamptz;
ALTER TABLE message_outbox ADD COLUMN attempts integer NOT NULL DEFAULT 0 CHECK(attempts>=0);
-- Old processing rows have no provable owner or delivery result. Never resend them.
UPDATE message_outbox SET status='uncertain',updated_at=now() WHERE status='processing';
ALTER TABLE message_outbox ADD CONSTRAINT outbox_lease_state CHECK(
 (status='processing' AND lease_token IS NOT NULL AND lease_until IS NOT NULL)
 OR (status<>'processing' AND lease_token IS NULL AND lease_until IS NULL));
CREATE INDEX outbox_expired_lease ON message_outbox(lease_until) WHERE status='processing';
