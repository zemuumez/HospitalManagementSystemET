# Independent review: packages and insurance

Reviewed commits `ef74e5f`, `e32aff3`, `08dd639`, `1e0a86f` on `origin/feat/packages-insurance-parity`. Review performed on a separate `review/packages-insurance` branch. **Verdict: changes requested before proceeding to another module or merging into main.** No application fixes were made during this review.

## Independently rerun verification

- `node --env-file=<existing local development env> scripts/verify-connected-isolated.mjs --endpoint-repairs`: **74 checks passed**, including the submitted browser package/insurance flows; temporary schema removed. The shared development records were retained.
- `npm run typecheck`: passed.
- `npm test`: eight tests passed. Only one of these is the Amharic-coverage test; these are not eight separate localization tests.
- `go test ./... -count=1 -v` from services/api with `HMS_TEST_DATABASE_URL` set to the loopback development DSN: passed. `TestClinicalTransactions` ran against its own schema (not skipped), including the package and insurance helper tests.
- `npm run format:check`: **failed** for services-workspace.tsx and check-insurance-ui.mjs.
- Additional application/repository probes against a fresh schema reproduced the monetary and aggregate defects below. Raw observations are in [packages-insurance-review-probes.json](packages-insurance-review-probes.json).

Passing the submitted tests is reproducible, but they do not cover the following cases.

## Findings requiring corrections

### R1 — high: repeated child IDs corrupt package totals

Locations: `services/api/internal/domain/packages.go:72`; `services/api/internal/adapters/postgres/packages.go:191`.

The validator rejects repeated service IDs but accepts the same existing child ID twice when service IDs differ. Update calculates the parent total from both entries, then updates the same database row twice. Reproduction: create one line; update with that line's ID on two entries, using different valid services and rates 100 and 200, quantity 1, discount 0. Update succeeds and returns two lines/total 300, but reload returns one line with amount 200 and parent total 300.

Required: reject duplicate nonempty child IDs before any mutation; ensure persisted line totals reconcile to parent totals. Add an API/repository regression asserting validation error, unchanged original rows and no success audit event.

### R2 — high: unchecked integer overflow persists incorrect money

Locations: `services/api/internal/domain/packages.go:88`; `services/api/internal/domain/insurances.go:146`.

Multiplication, accumulation and `(base * discount + 50) / 100` use unchecked int64 arithmetic. Both aggregates accept a base of 100000000000000000 minor units with a 100% discount and persist total **184467440737095515**, rather than zero or rejecting an out-of-range input. The issue is server-side and survives database constraints because the overflow result is positive.

Required: define supported amount/quantity/line-count limits and use checked arithmetic or safe decimal/integer calculation. Reject overflow before writes. Test multiplication, addition, discount/rounding boundaries and browser-safe serialization limits. Do not clamp an overflowed negative result to zero and call it valid.

### R3 — medium: explicit zero service price is silently replaced

Locations: `services/api/internal/adapters/postgres/packages.go:44` and `:154`.

A submitted valid rate_minor=0 is treated as missing and overwritten by the service's catalog rate. Probe: service price 500, submitted package price 0, quantity 1 → persisted price/total 500. This changes the price the operator submitted; original source calculation uses the supplied line rate.

Required: distinguish omitted price from explicit zero, or require an explicit rate. Apply any default only when genuinely omitted. Test create and edit with zero-priced lines and nonzero catalog rates; the UI and server must agree.

### R4 — medium: export silently returns only the first 25 rows

Locations: `services/api/internal/adapters/httpapi/packages.go:14`, `insurances.go:14`; `services/api/internal/application/packages.go:72`, `insurances.go:73`.

Both export handlers request limit=1000 through a listing service that resets any limit above 100 to 25. The additional package probe created 32 records and observed total=32 but only 25 returned with exactly the export handler's service arguments. The insurance path has the same logic. Existing export tests check HTTP success, not complete output.

Required: implement bounded streaming/batched export with explicit filters and authorization, or clearly defined explicit limits rather than silent truncation. Test more than 25 and more than 100 records, counts, totals and the intended export format.

### R5 — medium: catalog pagination/search hides records beyond the first page

Locations: `apps/web/src/components/services-workspace.tsx:592`, `:595`, `:596`, `:1799`, `:1918`.

The workspace fetches packages/insurances without page/limit/search parameters, ignores API total and performs search on loaded arrays only. Insurance page-size selection does not change fetching/rendering; its only page button is a static 1. Thus a catalog above 25 records is incomplete and later records cannot be found. The service picker requests limit=100 but the service repository still hardcodes 25, so later services are unavailable too.

Required: wire server paging/search and actual totals for each list and a searchable/paged service lookup. Add browser tests with at least 30 packages, policies and services, including a match only on a later page, changing page size and creating a name that sorts beyond the first page.

### R6 — medium: failed catalog loads can appear connected and empty

Location: `apps/web/src/components/services-workspace.tsx:586` through `:761`.

Successful unrelated endpoint responses set the shared connected flag. A failed package/insurance HTTP response is otherwise ignored; initial arrays remain empty and can display “no policies” while the UI indicates connected. Network exceptions only set connected=false, without a module load error or retry state. Old rows can also remain after reload failure.

Required: independent module loading/error/success states, visible errors and retry, and distinguish a genuine empty 200 response from 403/500/network failure. Test partial failure (services 200, packages 500), complete network failure and reload after a previously successful list. Do not report success merely because a different workspace endpoint responded.

### R7 — medium: insurance status toggle is not retry-safe

Locations: `apps/web/src/components/services-workspace.tsx:1279`; `services/api/internal/adapters/postgres/insurances.go:324` (ToggleInsuranceStatus).

The UI has no in-flight guard for status requests, and the API blindly flips the stored value. Two rapid clicks or retry after a lost successful response reverse the first update. An atomic SQL toggle prevents a torn write but does not make the intended action idempotent. The supplied test deliberately toggles twice sequentially; it does not cover duplicate/retried requests.

Required: send desired status, validate 0/1 and set it idempotently (with a version policy if appropriate), disable the row control while pending and handle response loss. Add duplicate/retry/concurrent update tests and invalid-status validation (current InsuranceInput.Validate does not check Status).

## Coverage and documentation gaps to retain explicitly

- In-use deletion tests insert `ipd_admission_details` links directly through SQL, including an OPD fixture. They prove the FK/check behavior, not a working admission form/API selecting and persisting package/insurance IDs. No corresponding admission-ID integration was found outside the new catalog code/tests. Standalone original patient admissions remain a distinct unfinished workflow.
- The new read permissions expose catalogs to patient/case-manager roles as well as admin/receptionist/doctor. Catalog access may be an intentional policy, but document it and test every action for all nine roles. Existing browser flows are admin flows; the earlier nine-role checks do not automatically cover these newly added actions.
- Row locks serialize updates but do not reject stale form edits. Record the chosen last-write-wins/version policy and test it; do not claim stale-update protection from lock presence alone.
- REPAIRS.md still lists Steps 2–4 as future work despite the submitted commits. Refresh it and the module contracts/field register with actual evidence and unresolved gaps. Do not mark admission integration or full source parity done from direct SQL fixtures.
- Fix the two formatting failures and rerun the check. No unrelated refactor is required.

## Work order for the next builder

1. Fix R1–R3 with failing regression tests first; run focused domain/repository/API tests and commit.
2. Fix R4–R6 with populated catalog/export and failure-state tests; run browser integration/typecheck and commit.
3. Fix R7 and complete action-specific role/validation/retry coverage; update contracts/evidence and commit.
4. Run the existing 74-check suite, full Go suite with actual PostgreSQL enabled, frontend tests/typecheck and formatting. Report every commit and any skipped test. Stop for re-review before merging or starting another module.

Keep these changes on the submitted feature branch or a clearly identified correction branch. The review has not merged the feature into main.
