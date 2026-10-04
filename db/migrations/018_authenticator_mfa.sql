ALTER TABLE "user" ADD COLUMN "twoFactorEnabled" boolean NOT NULL DEFAULT false;
CREATE TABLE "twoFactor" (
 id text PRIMARY KEY,
 secret text NOT NULL,
 "backupCodes" text NOT NULL,
 "userId" text NOT NULL UNIQUE REFERENCES "user"(id) ON DELETE CASCADE,
 verified boolean NOT NULL DEFAULT true,
 "failedVerificationCount" integer NOT NULL DEFAULT 0 CHECK("failedVerificationCount">=0),
 "lockedUntil" timestamptz
);
CREATE FUNCTION audit_mfa_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW."twoFactorEnabled" IS DISTINCT FROM OLD."twoFactorEnabled" THEN
  DELETE FROM session WHERE "userId"=NEW.id;
  DELETE FROM verification WHERE value=NEW.id AND (identifier LIKE 'trust-device-%' OR identifier LIKE '2fa-%');
  INSERT INTO audit_event(actor_id,action,resource_id) VALUES(NEW.id,CASE WHEN NEW."twoFactorEnabled" THEN 'identity.mfa_enabled' ELSE 'identity.mfa_disabled' END,NEW.id);
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER user_mfa_change AFTER UPDATE OF "twoFactorEnabled" ON "user" FOR EACH ROW EXECUTE FUNCTION audit_mfa_change();
