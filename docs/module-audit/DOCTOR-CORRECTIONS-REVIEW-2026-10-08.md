# Doctor correction review

Reviewed `e33eeb54ba20579088817a19841a3824d28fa342` with `ebd3b3d` and `0a2861a`. **Verdict: one remaining validation defect before approval.** Main remains unchanged.

## Independently verified

- Go `test ./... -count=1 -v` with real loopback PostgreSQL enabled: passed.
- Connected isolated integration runner with `--integration`: passed both reported integration groups, including new coordinated doctor provisioning/editing cases. Temporary schema removed.
- Typecheck, eight frontend tests and formatting: passed.
- Source review confirms D1 private directory redaction, D5 invalid/conflicting status rejection and idempotent state handling, D2 seven-day 10:00–19:30 / 60-minute defaults, and D3 required designation/qualification/gender validation.
- D4 now writes identity, credential account, staff access, profile, schedule and audit in one database transaction. Duplicate email and invalid department paths are covered. This is a substantive improvement over the disconnected create flow.

## Remaining defect: malformed email accepted through Go profile editing

Location: `services/api/internal/domain/scheduling.go`, UpdateDoctorInput.Validate email branch.

The validator only checks that the string contains `@`. A direct executable domain probe returned validation success for each of `@`, `a@`, `a@@b`, and `a b@c`. UpdateDoctorProfile then writes the accepted value to the authentication user email. The Next staff PATCH route uses a stricter email validator, so the same identity field has conflicting contracts across the two routes. This violates the original required email-format rule and can set an unusable login address.

Fix only this validation gap: use the established strict bare-email validation convention, retaining lowercase normalization and the length bound. Reject display-name address syntax, whitespace and malformed local/domain forms; do not use a parser that silently accepts `Name <address@example.test>` as a bare login email. Make Next and Go agree on accepted inputs.

Regression requirements:

1. HTTP PUT doctor profile with the malformed examples returns 422.
2. User email, profile version and audit count remain unchanged after rejection.
3. Valid normalized email edit still succeeds; duplicate email still returns 409.
4. Apply the same examples to the coordinating Next staff PATCH route and confirm consistent rejection.

Commit this narrow correction, run focused tests plus relevant existing checks, push and return for re-review. Do not reopen completed corrections or expand to another module.

## Test coverage clarification

The submitted invalid-department test exits before inserting an identity, so it demonstrates early rejection rather than rollback after partial writes. Add a targeted forced late-write failure to the isolated provisioning regression: fail doctor_hours or doctor.created audit insertion after identity creation; assert user/account/access/profile/hours/audit are all absent, then remove the injected failure and retry successfully. Keep this inside an owned temporary schema. The transaction structure is present; this is the remaining evidence needed for its advertised partial-failure guarantee.

## Environment limitation

The main Go and integration suites completed successfully. A later additional database probe failed during schema setup with PostgreSQL SQLSTATE 53100 (no space left on device), before its doctor actions ran. Its database assertions are not claimed as executed. The malformed-email acceptance was independently reproduced by the domain probe without a database.

The review removed only its completed `.local/qa-9e89cd2269b59e8d80469df2` temporary browser build after verifying its path was under this worktree's `.local` directory. About 123 MB was then free on C:. No project source or shared development records were removed. Free additional disk space before another full integration run; do not delete database volumes or unrelated files as a test workaround.
