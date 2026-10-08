# Doctor profile and account core: independent acceptance

Reviewed HEAD: `d131cb917db62c4de1b4daf1ee2475b721f3d34e` on `feat/doctor-workspace-parity`.

## Verdict

Accepted for the bounded Step 1 doctor profile/account core scope. No remaining blocking findings from the preceding correction reviews. This is not acceptance of the entire doctor workspace or all HMS modules. Main was not changed during this review.

## Independently executed checks

- `go test ./... -count=1` passed with PostgreSQL test configuration enabled, including the domain and database-backed doctor regressions.
- `node --env-file=<existing local environment> scripts/verify-connected-isolated.mjs --integration` passed. The runner reported successful provisioning, scheduling, authentication, authorization, SMS capture and Mailpit checks, then removed its isolated schema.
- `npm run typecheck` passed.
- `npm test` passed all 8 frontend unit tests.
- `npm run format:check` passed.
- Additional direct validation of all 15 shared fixture cases against the installed Zod validator passed.

## Closed findings

The Go email pattern now matches the installed Next.js practical email validator. Both reject leading, trailing and consecutive dots in the unquoted local part. Go no longer silently trims outer whitespace, matching Next.js rejection behavior; accepted emails are lowercased and length-limited. Shared malformed cases run through Go PUT and Next.js PATCH/POST regression paths. Valid update, version increment and duplicate-email conflict coverage remain present.

The Next.js provisioning rollback test compares counts directly across user, account, staff_access, doctor_profile, staff_profile, doctor_hours and audit_event in the isolated schema, and confirms successful retry. These assertions no longer depend on looking up a deleted user.

The Go test injects an audit-insert failure after the profile/hours writes and verifies that profile, hours and audit records do not survive. Direct count queries now check errors. This supplies post-write rollback evidence in addition to archived-department validation coverage.

These checks close the remaining findings in DOCTOR-FINAL-REVIEW-2026-10-08.md. Earlier privacy, status idempotence, seven-day scheduling defaults, required fields and coordinated provisioning corrections remain covered by the passing suites.

## Next bounded task

After the accepted branch is merged, follow Step 2 in `docs/doctor-workspace-parity-plan.md`: doctor deletion protection and reference auditing. Map the original nine clinical-model checks plus payroll to actual target relations. Reuse existing guards where possible; implement missing enforcement and test admin/unauthorized access, every dependency blocking deletion, and unreferenced cleanup. Preserve clinical records. Return commits and evidence for review before proceeding to schedules, OPD charges and full frontend integration (Steps 3–6).

Keep the existing UI and source workflows. Do not treat this acceptance as proof that deletion, schedule aggregate parity, charge masters or browser workspace integration are complete.
