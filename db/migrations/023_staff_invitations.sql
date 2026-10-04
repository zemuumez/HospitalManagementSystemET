CREATE TABLE staff_invitation (
 id uuid PRIMARY KEY, user_id text NOT NULL REFERENCES staff_access(user_id),
 actor_id text NOT NULL REFERENCES "user"(id), role text NOT NULL,
 token_digest text NOT NULL UNIQUE CHECK(length(token_digest)=64),
 expires_at timestamptz NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
 consumed_at timestamptz, revoked_at timestamptz,
 CHECK(NOT(consumed_at IS NOT NULL AND revoked_at IS NOT NULL))
);
CREATE INDEX staff_invitation_pending ON staff_invitation(user_id) WHERE consumed_at IS NULL AND revoked_at IS NULL;
