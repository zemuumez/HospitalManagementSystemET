# Authenticator MFA backend

Migration 018 and `hospitalMfa()` enable Better Auth authenticator MFA. Existing users are not automatically enrolled. Enrollment and challenge screens belong to Section 4 and are not yet connected; do not enroll an ordinary development account until those screens are available.

## API flow

1. With an authenticated session, POST `/api/auth/two-factor/enable` with the current `password`. Securely display the returned `totpURI` as an authenticator QR and show `backupCodes` once. Never log either value.
2. POST `/api/auth/two-factor/verify-totp` with the authenticator `code`. Enrollment becomes active only after verification. Store replacement cookies; all prior sessions are revoked.
3. Email/password and `/api/auth/firebase/sign-in` return `twoFactorRedirect: true` for enrolled users. Their challenge cookie is not a hospital session and expires after five minutes.
4. POST `/api/auth/two-factor/verify-totp` with `code`, or `/api/auth/two-factor/verify-backup-code` with a recovery `code`. Successful verification consumes the challenge and issues a normal Better Auth session.
5. Recovery codes are single-use. Password-confirmed `/api/auth/two-factor/generate-backup-codes` replaces the recovery set. Store codes offline securely.
6. POST `/api/auth/two-factor/disable` with the current password from a sensitive authenticated session. Disabling revokes prior sessions and is audited. There is no administrator bypass/recovery endpoint.

All mutations require the configured trusted origin. Better Auth encrypts the authenticator secret and recovery-code storage using the auth secret. Preserve that secret securely with database recovery procedures. Device remembering has zero lifetime. Five failed sign-in MFA attempts lock MFA for fifteen minutes; endpoint throttling also applies. A password reset does not remove MFA. This is sign-in MFA, not a claim of per-action clinical or finance step-up enforcement.

## Verification

`npm run test:firebase` uses a disposable PostgreSQL schema, isolated web/API services and the actual Firebase Auth emulator. It verifies enrollment, prior-session revocation, password and phone challenges, unauthorized Go access before completion, invalid codes, consumed challenge rejection, single-use backup codes, lockout across new challenges, lock expiry, password-required disablement and enable/disable audit records. It resets only disposable rate-limit fixtures to test account lockout independently of IP throttling. No external credentials or real SMS are required.
