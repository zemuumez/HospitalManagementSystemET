# Staff invitations

Migration 023 and the Better Auth hospital-invitations plugin add backend invitation workflows. Existing password provisioning remains supported. Invite endpoints run under `/api/auth` and use the configured trusted origin, Better Auth database sessions and endpoint throttling (five requests per endpoint/IP per minute). The accepting account gets no session automatically and signs in normally afterward.

## Administrator endpoints

- POST `/staff/invite`: `{name,email,role}` using the existing nine role IDs. A new identity is created inactive with no password account. An existing email cannot be claimed or overwritten. The response contains invitation/user IDs, never the bearer token.
- GET `/staff/invitations?page=1`: administrator-only, 25 entries per page, including created/expiry/consumed/revoked times and recipient metadata. Tokens and digests are excluded. Reads are audited and no-store.
- POST `/staff/resend-invitation`: `{id}` revokes the old pending link and queues a fresh 24-hour link. It only works for inactive accounts with no credential/provider account. Expired pending invitations can be renewed.
- POST `/staff/revoke-invitation`: `{id}` revokes an unconsumed pending link. It does not disable an already accepted account; use normal staff access administration for that.

## Acceptance

POST `/staff/accept-invitation` with `{token,password}`. The password must contain 12–128 characters. The random 256-bit token is stored as a SHA-256 digest in the invitation table. Expired/revoked/consumed links, active identities and changed invitation roles are rejected. The transaction serializes acceptance with access administration, creates the Better Auth password hash, marks the invited email verified, activates its issued role, consumes the invitation and records audit attribution. Concurrent attempts yield exactly one success.

Email delivery uses the existing durable operational outbox. In development the worker sends only to Mailpit. The stored message contains the bearer link, so database/mailbox/backups are sensitive. Migration 027 assigns invitations to the identity audience and excludes them from the operational message list, including recipient/delivery metadata. That list already omits message bodies; invitation tokens were not returned through it. The delivery worker can still send identity messages, and the administrator-only invitation list exposes no token. SMTP credentials stay blank in the production worksheet. Link expiry is checked during acceptance even if email is delivered late. Uncertain email delivery uses existing outbox reconciliation policy; renewing an invitation invalidates a possibly delivered old link.

The email points to `/accept-invitation`; that acceptance screen and invitation management screens are Section 4 work and are not connected yet. Do not roll out invitations to real staff until those screens are implemented. No external email was sent in verification.

`npm run test:invitations` starts disposable web/API/worker services and validates actual Mailpit delivery, admin/origin denial, inactive pending accounts, password bounds, verified email, concurrent acceptance, session/Go role, expiry, renewal, revocation, existing-email takeover denial and private list pagination. Only generated test schemas are removed afterward.
