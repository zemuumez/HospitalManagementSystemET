# Packages/insurance correction re-review

Reviewed remote HEAD `699bab5223811c849bd98b0be9808d2c3920b091`, including corrections `91a2368`, `354a97e`, `af02a07`, `699bab5`. Main remains unchanged. Verdict: **substantial repairs verified; three remaining corrections before approval**.

## Independently executed checks

- Isolated real-login Next/Go/PostgreSQL/browser endpoint-repairs runner: **78 checks passed**; temporary schema removed.
- Go `test ./... -count=1 -v` with HMS_TEST_DATABASE_URL pointing to loopback PostgreSQL: passed; TestClinicalTransactions executed, not skipped.
- TypeScript typecheck: passed. Frontend tests: 8 passed. Formatting: passed.
- Additional application/repository probes against a fresh schema: duplicate package child IDs rejected; the previously overflowing amount rejected; explicit zero rate persists as zero. These confirm R1–R3 repairs. See [probe results](packages-insurance-corrections-probes.json).
- Populated list pagination/search and HTTP 500/retry browser cases passed, confirming the principal R5 list and R6 failure-state repairs. Explicit desired-status operations have new passing retry/concurrency tests.

## Remaining findings

### C1 — R7 still has a non-idempotent fallback

`services/api/internal/adapters/httpapi/insurances.go:61` only decodes the status body when ContentLength > 0. Empty body, `{}`, or null status reaches SetStatus with nil. Chunked/unknown-length bodies (ContentLength = -1) are not decoded either. `services/api/internal/application/insurances.go` permits nil and `services/api/internal/adapters/postgres/insurances.go:356` then executes the old SQL toggle.

Independent application/repository probe: starting status 1, two identical nil-status requests produced **1 → 0 → 1**. The valid explicit-status path is repaired, but the public fallback still reverses on retries. A chunked valid `{ "status": 0 }` can also be ignored and toggled instead of set.

Fix: require a non-null status of 0 or 1, decode irrespective of known ContentLength, reject missing/empty/null/invalid input without mutation, and remove the public nil-toggle path. Test missing/empty/null and unknown-length bodies over HTTP, explicit repeated desired-status requests, unchanged data/audit on validation failure, and roles. Do not preserve unsafe toggle semantics solely as backward compatibility for old tests.

### C2 — R5 service picker is still a first-page lookup

`apps/web/src/components/services-workspace.tsx:612` fetches only `/api/hms/services?page=1&limit=100`. Create/edit selectors use that array (around lines 2850 and 3085), with no additional lookup page requests. Therefore service 101 onward remains unavailable. The latest test seeds about 32 services, proving improvement beyond 25 but not a complete lookup. The claim in REPAIRS.md that the selector requests unconstrained active services is inaccurate: the request has limit=100 and no status filter.

Fix: use server-search/pagination for active service selection, or fetch all pages safely within a clearly enforced contract. Preserve existing selected archived references for display without permitting new invalid selections. Test at least 105 services, selecting and saving an active service outside the first 100, create/edit behavior, and a failed lookup with retry. Update the documentation to describe the actual implementation.

### C3 — R4 export cap still needs an explicit overflow contract

Both ExportPackages/ExportInsurances stop at 5,000 and return success with all accumulated rows. HTTP returns `{items,total}` without a truncated marker, continuation or error. The independent package probe observed **5,000 returned / 5,002 total**. The cap is mentioned in the repair notes, but the endpoint still produces a partial export with ordinary success. This is batched accumulation in memory, not streaming as REPAIRS.md claims.

Fix: retain a bounded export, but reject oversized selections with a clear error and instruction to filter, or provide explicit truncation/continuation that the consumer must handle. Alternatively implement a complete streamed export. Test >5,000 records for both types and filtered exports below the cap. Do not describe partial output as complete or in-memory batching as streaming.

## Next work order

1. Correct C1 and add the HTTP edge-case regressions; commit.
2. Correct C2 and test a service outside the first 100 in both forms; commit.
3. Correct C3, update evidence/contracts accurately, and test the cap/filtered behavior; commit.
4. Rerun the 78-check suite, Go tests with PostgreSQL enabled, typecheck, frontend tests and formatting. Stop for re-review; do not merge or start another module.

The remaining original admission integration is still a separate documented gap. These corrections do not require a broader application rewrite. This review changed documentation/evidence only.
