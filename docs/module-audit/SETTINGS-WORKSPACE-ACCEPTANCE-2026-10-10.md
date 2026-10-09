# Settings workspace acceptance

Verdict: **accepted for the reviewed workspace scope**, including the test-isolation correction on `codex/settings-acceptance-review`.

Submitted implementation: `0ea97a1` on `feat/settings-workspace-parity`.
Additional review correction: `1ace858` (`test(settings): isolate worker cleanup schema and fixture storage`). Merge the accepted review branch, or include that correction with the submitted implementation. Main was not changed during this review.

## Findings closed

- L1: ordinary hexadecimal text, provider public keys, and external URLs no longer undergo arbitrary attachment existence validation. Actual supported local attachment references retain existence and clinical-asset checks.
- L2: worker cleanup resolves an active administrator through `staff_access`, records a valid audit actor, and fails visibly when none exists. The worker CLI success and missing-actor cases passed independently with the corrected test harness.
- L3: cleanup filters referenced assets before its 100-item limit, then rechecks under attachment locks. The more-than-100-fixture regression passes.
- The earlier attachment binding/retirement locking corrections remain in place, with ordering and clinical-protection tests passing. The Operation Categories and Custom Fields edit findings remain closed.

## Independent verification

| Gate | Result |
| --- | --- |
| Isolated Settings browser suite | All 17 journeys passed; schema cleanup completed |
| Full Go suite with PostgreSQL and `-count=1` | Passed with review test-isolation correction |
| Real worker CLI regression | Freshly built binary; isolated-schema orphan deletion and audit actor passed; no-admin failure passed |
| TypeScript typecheck | Passed |
| Frontend unit tests | 8 tests passed |
| Formatting | Passed |

The browser suite ran against submitted implementation `0ea97a1`; the review changes only affect Go test setup. Go verification used the corrected setup. Subsequent changes were gofmt, comment clarification, and removal of an unreachable environment-based skip.

## Test-isolation correction made during review

The delivered CLI test spawned the worker with ambient `DATABASE_URL`, while its fixtures lived in the clinical harness's randomly generated schema. It also reused an existing worker binary when present and did not point the worker at fixture file storage. This could test stale code, miss the fixture, or run cleanup against development data.

Commit `1ace858` builds a fresh worker into a temporary directory, obtains and checks the live `hms_test_` schema, explicitly sets that schema in the worker connection URL, and supplies the fixture attachment directory. The no-admin URL replaces rather than appends the search-path parameter. No production implementation was changed during review.

During verification, an initial temporary correction using pgx `ConnString()` alone also omitted runtime search-path changes and the CLI missed its fixture. A read-only check of the default-schema audit log found no attachment retirements in the preceding 15 minutes. The successful rerun used the explicit isolated search path described above.

## Scope and next step

This accepts the reviewed Settings workspace and its correction batches. It does not certify every downstream settings consumer or whole-system/source pixel parity. The browser evidence includes selected light/dark and Amharic checks, not exhaustive visual coverage of every state.

Next: integrate this accepted branch into main through the established merge workflow, then begin Inventory Completion with a gap audit against the implementation checklist. Preserve the accepted Packages/Insurance, Doctors, and Settings behavior in subsequent regression work.
