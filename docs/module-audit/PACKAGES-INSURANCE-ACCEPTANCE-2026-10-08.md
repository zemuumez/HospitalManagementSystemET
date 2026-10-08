# Independent acceptance of catalog correction scope

Reviewed implementation: `d8e52f96323f14f63b08c57faf91bc36aefaa6d8` on `feat/packages-insurance-parity`, including `8d7c535`, `7463d82`, `d8e52f9`.

**Verdict: C1–C3 accepted for the reviewed scope. No remaining merge-blocking finding was identified in this targeted review.** This accepts the package/insurance catalog corrections, not full hospital workflow parity or production certification. Main was not changed or merged during review.

## Independent verification

Executed from a separate review branch:

- `node --env-file=<existing local development env> scripts/verify-connected-isolated.mjs --endpoint-repairs`: **79 checks passed**, exit 0. Includes real Better Auth/Next/Go/PostgreSQL/browser flows, pagination/search, service selection beyond 100, create/edit, archived selection display, lookup failure/retry and catalog failure isolation. The harness removed its temporary schema and retained shared development records.
- `go test ./... -count=1 -v` from services/api with `HMS_TEST_DATABASE_URL` configured: passed, exit 0. `TestClinicalTransactions` executed with PostgreSQL (14.95 seconds), not skipped. Includes HTTP status payload and export-overflow regressions.
- `npm run typecheck`: passed.
- `npm test`: eight tests passed, including the Amharic catalog-coverage test.
- `npm run format:check`: passed.

## Findings closed

### C1: required desired status

The HTTP handler now decodes bodies without checking ContentLength, requires a non-null 0/1 status, and passes an integer desired value to the application/repository. The nil-toggle path is removed. Regression coverage includes missing/empty body, empty object, null status, unchanged data/audit for rejected requests, unknown-length valid body, repeated desired statuses and concurrent requests.

Report clarification: empty/unparseable JSON bodies return **400**, while a parsed object with missing/null/invalid status returns **422**. Both reject without mutation; saying every empty-body case returns 422 was inaccurate but is not a functional blocker.

### C2: service lookup beyond the first page

The picker loads active services in 100-row batches. Independent browser execution passed creation/editing with service 102 in a catalog seeded with at least 106 active services, display of an existing archived selection, exclusion of that archived option from a new line, and failed lookup/retry recovery. The backend permits preservation of an already-referenced archived service during package update.

Retained limitation: the picker is bounded to 50 pages / 5,000 active services and uses client filtering over that loaded catalog; it is not an unlimited server-search selector. Record a separate enhancement if deployment requires more than 5,000 active services. This acceptance does not claim testing above that picker bound.

### C3: explicit bounded export

Package and insurance exports reject selections above 5,000 with `422`, `EXPORT_LIMIT_EXCEEDED`, total and limit unless truncation is explicitly requested. Opt-in truncation includes `truncated: true` and a warning; filtered exports below the cap report `truncated: false`. Real PostgreSQL regressions seed 5,005 records for each type and verify these branches. Exports use in-memory batched accumulation, not streaming.

## Boundaries and next step

- Earlier R1–R3 regressions remain in the passing suite: child-ID uniqueness/reconciliation, bounded money calculations and explicit zero pricing.
- Admission deletion protection is demonstrated with directly inserted references. Original standalone admission UI/API linkage and historical package/insurance use remain separate unfinished work.
- Last-write-wins updates, catalog role visibility and provider prerequisites retain their documented scope. This review does not establish comprehensive application-wide security or role parity.
- Do not rewrite historical review findings; this acceptance supersedes their open status for the named correction scope.
- Recommended next implementation slice after integration of the reviewed branch: doctor workspace contract/form/API parity, following docs/builder-handoff/IMPLEMENTATION-PLAN.md. Start with a gap table and one tested vertical step; do not mark the whole backend complete.

No application code or shared database schema was changed during this review. The feature branch is ready for the normal merge process; merging and deployment were not performed here.
