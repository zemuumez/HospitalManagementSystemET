CREATE TABLE IF NOT EXISTS "user" (
  id text PRIMARY KEY, name text NOT NULL, email text NOT NULL UNIQUE,
  "emailVerified" boolean NOT NULL DEFAULT false, image text,
  "createdAt" timestamptz NOT NULL DEFAULT now(), "updatedAt" timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS session (
  id text PRIMARY KEY, "expiresAt" timestamptz NOT NULL, token text NOT NULL UNIQUE,
  "createdAt" timestamptz NOT NULL DEFAULT now(), "updatedAt" timestamptz NOT NULL DEFAULT now(),
  "ipAddress" text, "userAgent" text, "userId" text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_session_user ON session("userId");
CREATE TABLE IF NOT EXISTS account (
  id text PRIMARY KEY, "accountId" text NOT NULL, "providerId" text NOT NULL,
  "userId" text NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
  "accessToken" text, "refreshToken" text, "idToken" text,
  "accessTokenExpiresAt" timestamptz, "refreshTokenExpiresAt" timestamptz,
  scope text, password text,
  "createdAt" timestamptz NOT NULL DEFAULT now(), "updatedAt" timestamptz NOT NULL DEFAULT now(),
  UNIQUE("providerId", "accountId")
);
CREATE INDEX IF NOT EXISTS account_user ON account("userId");
CREATE TABLE IF NOT EXISTS verification (
  id text PRIMARY KEY, identifier text NOT NULL, value text NOT NULL,
  "expiresAt" timestamptz NOT NULL,
  "createdAt" timestamptz NOT NULL DEFAULT now(), "updatedAt" timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS verification_identifier ON verification(identifier);
CREATE TABLE IF NOT EXISTS "rateLimit" (
  id text PRIMARY KEY, key text NOT NULL UNIQUE, count integer NOT NULL, "lastRequest" bigint NOT NULL
);
CREATE TABLE IF NOT EXISTS staff_access (
  user_id text PRIMARY KEY REFERENCES "user"(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('admin','doctor','patient','nurse','receptionist','pharmacist','accountant','case_manager','lab_technician')),
  active boolean NOT NULL DEFAULT true
);
CREATE TABLE IF NOT EXISTS firebase_identity (
  firebase_uid text PRIMARY KEY, user_id text NOT NULL UNIQUE REFERENCES "user"(id) ON DELETE CASCADE,
  phone text NOT NULL UNIQUE, linked_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS consumed_phone_proof (
  digest text PRIMARY KEY, expires_at timestamptz NOT NULL
);
CREATE TABLE IF NOT EXISTS patient (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  medical_record_number bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
  given_name text NOT NULL CHECK (length(given_name) BETWEEN 1 AND 80),
  family_name text NOT NULL CHECK (length(family_name) BETWEEN 1 AND 80),
  date_of_birth date NOT NULL,
  phone text NOT NULL DEFAULT '',
  user_id text UNIQUE REFERENCES "user"(id),
  clinician_user_id text REFERENCES "user"(id),
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS patient_created ON patient(created_at DESC, id);
CREATE INDEX IF NOT EXISTS patient_clinician ON patient(clinician_user_id);
CREATE TABLE IF NOT EXISTS audit_event (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  actor_id text NOT NULL REFERENCES "user"(id), action text NOT NULL,
  resource_id text NOT NULL DEFAULT '', created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS message_outbox (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), actor_id text NOT NULL REFERENCES "user"(id),
  channel text NOT NULL CHECK(channel IN ('sms','email')), recipient text NOT NULL,
  subject text NOT NULL DEFAULT '', body text NOT NULL,
  idempotency_key text NOT NULL,
  status text NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','processing','sent','captured','failed','uncertain')),
  provider_id text NOT NULL DEFAULT '', created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(), UNIQUE(actor_id, idempotency_key)
);
CREATE INDEX IF NOT EXISTS outbox_pending ON message_outbox(created_at) WHERE status = 'pending';
