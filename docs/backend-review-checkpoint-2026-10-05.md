# Backend review checkpoint — 2026-10-05

> Current outcome: the operational branch is merged and pushed as `d618b0e`; all development migrations through 045 are applied, API readiness is 200, database reconciliation passes, and [hosted CI is successful](https://github.com/zemuumez/HospitalManagementSystemET/actions/runs/37254122802). Section 3 still has explicitly open requirements. Earlier attendance-only and pending-merge statements below are historical and superseded by this outcome.

Section 3 is still in progress. Section 4 new frontend integration has not started.

## Accepted and running: attendance

The submitted `feat/attendance-backend` commit `a29232b` is an ancestor of `feat/section-3-operational-modules`; it is not a separate implementation to merge twice.

Attendance was reviewed in an isolated checkout. Three failures were reproduced with database regression tests: keyed retry conflicts, overlapping overnight open records, and approval of unfinished attendance. Repairs were committed individually:

- `0221ff5`: transactional replay responses for clock-in/out, break start/end and manual creation; changed payloads conflict; concurrent matching retries return the same result.
- `353128e`: one-open-record database guard, midnight lookup of the open overnight record, and checkout requirements for submission/approval.

Both fixes and the original attendance work are merged and pushed to `main`. Development migrations 028, 041 and 042 are applied. Numbers 029–040 are reserved for the separately reviewed operational branch. The migration runner tracks filenames individually, so later acceptance of those migrations does not skip them.

The local API was rebuilt as `services/api/bin/api-attendance.exe`, restarted, and returned HTTP 200 from `/readyz`. The existing worker was not changed. Attendance frontend integration remains open.

Verification passed:

- Uncached Go tests against a disposable PostgreSQL schema, including concurrency and negative authorization tests; `go vet`.
- `npm run test:integration`: authentication, session revocation, CSRF, role/record scope, staff provisioning, scheduling, concurrent booking, idempotency, captured SMS and Mailpit workflows. The isolated schema was removed after the run.
- [Hosted CI for 353128e](https://github.com/zemuumez/HospitalManagementSystemET/actions/runs/37248545137): successful.

## Repairs preserved on the operational review branch

[fix/operational-review](https://github.com/zemuumez/HospitalManagementSystemET/tree/fix/operational-review) contains the submitted operational work plus attendance repairs and these individually tested changes:

| Commit | Repair | Evidence |
|---|---|---|
| `c8fc2e4` | Serialize doctor/date queue token allocation | Before repair, 11 of 16 concurrent registrations failed with duplicate tokens; after repair, all receive unique tokens |
| `9790cfa` | Private file upload/download and patient/care-assignment authorization | Real temporary-file and PostgreSQL tests; multipart HTTP round-trip; cross-patient/doctor denial; size, traversal and HTML rejection; server checksums |
| `4869cf8` | Block unsafe partial patient merges | Original merge abort reproduced on clinical history; guarded endpoint leaves records and merge events unchanged |
| `728091b` | Record-level access checks across clinical-care extensions | Unrelated patient/doctor/nurse reads reproduced before repair; negative access, assigned-doctor/admin access and patient/encounter consistency tests pass |

All Go/PostgreSQL tests and vet pass on this review branch. The review branch has **not** been merged into main or applied to the development database. Passing its current tests does not establish complete module parity or production readiness.

## Remaining acceptance work

1. Implement a complete patient identity merge policy that preserves original clinical/financial identifiers, resolves portal ownership and consent conflicts, updates all scoped readers, and supports audited reconciliation. The unsafe endpoint is unavailable; the feature is not complete.
2. Finish authorization and lifecycle review of every new operational endpoint, particularly queues/public requests, prescription attribution, diagnostics release, clinical delegation, signed records, billing linkage and payroll.
3. Complete private-file malware scanning/quarantine, retention, orphan reconciliation, throttling, expanded formats and coordinated file/database restore. File configuration is documented on the review branch in `docs/private-attachment-storage.md`.
4. Verify original-source field/workflow parity rather than treating a table, CRUD endpoint or runbook as completion.
5. Perform the outstanding import reconciliation, performance/capacity runs, restore drill, staging checks and user acceptance. Unsupported completion claims for these items have been reopened in the review branch checklist.
6. Merge only verified increments; then complete Section 3 acceptance before beginning Section 4.

Real Firebase, SMS, payment and production delivery credentials remain blank. Existing setup instructions are in `provider-setup.md`. No external message, payment charge or production deployment was performed.


## Operational review update

Additional verified commits:

- `a69352c`: patient identity aliases preserve original clinical/financial identifiers, resolve portal ownership and consent conflicts, and prevent new writes to merged aliases.
- `c0975a4`: scoped appointment/prescription/consultation/diagnostic access and explicit reviewed file release.
- `346075c`: real delivered-service billing sources, payment-derived appointment state, controlled financial clearance, immutable signed summaries and guarded bed transfers.
- `63894f3`: production malware-scan gate, bounded public request quotas, active care-team delegation/revocation, source-correct prescription status, and operational lifecycle checks.
- `d978587`: authenticated encrypted backups, empty-target transactional restore, real-schema reconciliation and corrected acceptance claims.

The uncached Go/PostgreSQL suite and vet pass through migration 045. Connected API integration and the synthetic encrypted backup/restore drill pass. New-module browser integration, full original parity and the open Section 3 deployment/business requirements are not implied by these tests. Real provider credentials remain blank.


## Final local acceptance evidence

The reviewed operational branch passed on 2026-10-05:

- Uncached Go unit, PostgreSQL transaction/concurrency/authorization and HTTP tests through migration 045; `go vet`.
- Web unit tests, formatting, production build and TypeScript checks.
- Connected API integration: session/role/record isolation, CSRF, booking concurrency, captured SMS, Mailpit and revocation.
- Existing connected browser journeys: billing/payment, bed/case/admission/discharge, staff/schedules/patient booking, reload persistence and mobile layout.
- Firebase Auth emulator and MFA, password recovery, and staff invitations.
- Real-schema reconciliation during every connected suite; encrypted synthetic database restore with wrong-key, tampering and nonempty-target denial.
- Production dependency audit: zero vulnerabilities.

These results support integrating the submitted development branch. They do not establish full Section 3 completion: the corrected checklist retains outstanding tax/insurance/accounting workflows, complete original-module parity, provider provisioning, private-file recovery/retention, production deployment/security review and user acceptance. New-module frontend integration remains Section 4.
