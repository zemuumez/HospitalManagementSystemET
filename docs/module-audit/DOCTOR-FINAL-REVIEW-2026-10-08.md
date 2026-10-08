# Doctor correction re-review — 2026-10-08

Reviewed commit: `8bfa5ab4c487ad74350fa1df19e47bf3a7170a48` on `feat/doctor-workspace-parity`.

## Independent verification

- `go test ./... -count=1` passed with `HMS_TEST_DATABASE_URL` configured for PostgreSQL isolated tests.
- Connected isolated runner `--integration` passed, including provisioning, scheduling, authentication, authorization, SMS capture and Mailpit checks. Its schema was removed by the runner.
- `npm run typecheck`, `npm test` (8 tests), and `npm run format:check` passed.
- No application code changed during this review; main was not merged.

## Remaining validation defect

The replacement Go EmailPattern still accepts invalid unquoted local parts. An executable probe calling UpdateDoctorInput.Validate returned success for each of:

| Input | Go validation | Installed Zod z.email used by Next.js |
| --- | --- | --- |
| `.doctor@example.com` | accepted | rejected |
| `doctor.@example.com` | accepted | rejected |
| `doc..tor@example.com` | accepted | rejected |

The previous seven malformed-address cases are fixed, but the requested Next.js/Go agreement is incomplete. These results are domain-validation probes, not claims of database mutation. The Go update path consumes this validation before updating the login identity.

Correction: define one practical email acceptance contract matching the installed Next.js validator, including normalization and length, and implement it in both paths. Do not describe the current regex as fully RFC-compliant. Add the three cases above to HTTP rejection/no-mutation tests, retaining valid normalized updates and duplicate-email 409 coverage. Use a shared fixture corpus exercised by both languages to prevent further drift.

## Rollback evidence

The Next.js transaction now inserts user, account and staff_access before rejecting an archived department. Inspection confirms the explicit ROLLBACK and the integration run confirms the failed email can subsequently be provisioned successfully. This addresses the requested post-account failure scenario.

However, new cleanup queries for account, staff_access, profiles, hours and audit events filter through `SELECT id FROM "user" WHERE email=$1`. Once that user is absent, all those queries necessarily return zero, even if orphan rows existed. In particular, the audit query cannot prove absence of an independently retained audit record. This is a test-evidence weakness, not an observed rollback defect.

Correction: in the isolated schema compare before/after row identities or counts directly in every relevant table, without resolving IDs through the removed user. Prefer an isolated test fixture or captured attempted ID when concurrent writes are possible. Assert database query errors rather than discarding them.

The added Go archived-department test rejects before inserting doctor_profile; it is useful validation coverage but does not exercise rollback after profile writes. Do not describe it as proof of post-write Go rollback. If claiming that guarantee, inject a test-only database failure on a later hours/audit write in the isolated schema and assert no earlier profile write survives.

## Next action

Keep follow-up limited to the email contract and independent rollback assertions. Run focused regressions, then the existing checks, commit and push to the same feature branch for review. Do not expand module scope or merge main yet.
