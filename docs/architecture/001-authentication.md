# Authentication decision

Accepted direction from the user: one hospital, Better Auth authentication, Firebase phone authentication AND operational SMS. This supersedes multi-hospital/subscription proposals and the earlier Firebase-owned session proposal. The broader legacy parity review continues alongside implementation; it has not been declared complete.

Better Auth runs in Next.js and owns email/password credentials, password resets and eight-hour database-backed sessions. Public self-registration and automatic account linking are disabled. A server-side provisioning command creates the initial administrator. Roles and activation are stored separately in `staff_access`; clients cannot set them through a profile update. Core Better Auth tables are in the checked-in SQL migration.

The Go API validates the incoming cookie by calling Better Auth's fixed `/api/auth/get-session` URL. It rejects missing/expired/mismatched sessions and resolves the user's active role from PostgreSQL on every request. No identity header is trusted. Cookie caching is disabled. Business queries additionally scope patient records to the patient identity or assigned clinician; only admin/receptionist can register patients in this first slice.

The browser calls the same-origin Next.js `/api/hms/*` proxy. Both the proxy and Go validate mutation origins. The proxy exposes a fixed endpoint allowlist and forwards only the session cookie, content type, trusted origin and idempotency key. Go is loopback-only by default; deployment must preserve the private backend network boundary and strip/validate proxy headers. No public CORS wildcard is enabled.

## Firebase bridge

Firebase's browser SDK owns reCAPTCHA, SMS OTP issuance and verification. The client submits a Firebase ID token to a custom Better Auth endpoint. Firebase Admin verifies signature, audience/issuer, expiry and revocation. The bridge additionally requires `sign_in_provider=phone`, an E.164 phone number and authentication within five minutes.

Linking requires an existing active Better Auth session created within five minutes, plus the fresh Firebase proof. The bridge writes a unique binding between Firebase UID, local account and verified phone. It never discovers a staff account by a matching phone or email, and it does not overwrite an existing binding. Sign-in requires that binding and an active local account, then creates a normal Better Auth session/cookie. A token digest is consumed transactionally to reject replay of the same proof. Changing/removing a binding needs a future audited recovery workflow; it is intentionally not a generic profile edit.

Without Firebase project configuration, email sign-in works and the phone control explains that configuration is needed. No Firebase credentials were supplied, so real SMS OTP has not been activated. Emulator configuration is provided and prohibited in production. Phone proof policy tests are not a substitute for an end-to-end Firebase emulator/production integration test.

## Operational communications

Go owns a durable PostgreSQL outbox. Authorized admin/receptionist requests require an idempotency key. Identical retries return the original row; a reused key with different content is rejected. A separate Go worker atomically claims pending messages. Development SMS is captured locally, and email uses Mailpit SMTP. No real recipient is contacted during development tests.

A Twilio adapter is available only when explicitly selected and configured. `sent` means the SMTP server/provider accepted the message, not delivery to an inbox/handset. Network ambiguity is `uncertain`. A process crash can leave `processing` rows. Neither is blindly retried: reconcile with the provider before requeueing to avoid duplicate messages. Delivery callbacks, scheduling, preferences, bulk sends and an administrative retry/reconciliation UI remain later work.

## Security remaining before production

This is a tested foundation, not a security certification or production release. Remaining work includes mandatory staff MFA/step-up, account recovery and deprovisioning UI, full permission matrix, clinical record/file policies, clinical audit retention, CSP/nonces, deploy-time TLS/secret management, production database roles/backups, provider delivery callbacks, abuse/load testing and an independent security review. Do not deploy the local superuser/database configuration as production infrastructure.

References: [Better Auth plugins](https://better-auth.com/docs/concepts/plugins), [sessions](https://better-auth.com/docs/concepts/session-management), [Firebase phone auth](https://firebase.google.com/docs/auth/web/phone-auth), [Firebase token verification](https://firebase.google.com/docs/auth/admin/verify-id-tokens), [Mailpit](https://mailpit.axllent.org/docs/configuration/).
