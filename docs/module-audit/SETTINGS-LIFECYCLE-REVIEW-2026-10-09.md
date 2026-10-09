# Settings attachment lifecycle review

Reviewed commit: `3acea1f8d5b3b144b7eae53d31839667f38ac84f` on `feat/settings-workspace-parity`.

Verdict: **changes required**. The previously reproduced binding/retirement race is addressed by shared attachment locking and validation. R2 remains closed. Three concrete issues in the latest lifecycle implementation remain below; no expansion into other workspaces is requested.

## Independent verification

All 17 isolated Settings browser journeys passed (exit 0), with successful isolated-schema cleanup. Full Go/PostgreSQL tests with `-count=1`, TypeScript typecheck, all 8 frontend tests, and formatting passed. The new supplied overlap and ordering tests passed as part of the Go suite.

Three additional temporary probes in the isolated PostgreSQL clinical harness reproduced the failures below. Probe code was removed after execution; production and existing test files were restored unchanged. Main was not modified by this review.

## L1 — Ordinary setting content is mistaken for an attachment reference

Location: `services/api/internal/adapters/postgres/cms_settings.go:192` and `:390` (and old-value extraction at `:187`/`:386`).

The unanchored `[a-f0-9]{32}` expression scans every non-secret general-setting value and every CMS value. The new existence validation treats any matching text as an attachment token. Legitimate text, a public provider key, or an external URL containing a 32-character hexadecimal identifier now fails with `ErrNotFound`, even though it contains no attachment reference.

Independent reproduction: call `UpdateGeneralSettings` with `about_us` equal to `Reference 0123456789abcdef0123456789abcdef for hospital information`. Expected: normal text saves. Actual: **record not found**. This is a regression caused by the new binding validation, not an invalid attachment upload.

Required correction: extract references only from actual supported attachment URLs/asset fields, using a precise parser or explicit reference representation. Use the same parsing contract for incoming and displaced references. Do not classify arbitrary hexadecimal substrings as assets. Keep validation for actual missing, retired, and clinical attachment references. Cover ordinary text, provider public-key values, external URLs, and genuine attachment URLs in regression tests.

## L2 — Worker cleanup cannot resolve its audit actor

Location: `services/api/cmd/worker/main.go:59`.

The worker executes `SELECT id FROM "user" WHERE role = 'admin' ...` and discards the query error. The application's `"user"` table has no `role` column; roles and activation live in `staff_access`. The independently executed query fails with **SQLSTATE 42703: column "role" does not exist**. The worker therefore constructs an admin actor with an empty ID. When cleanup finds an eligible asset, the audit insert cannot reference that user, so the deletion transaction cannot complete. The HTTP test supplies a valid actor and does not exercise worker startup.

Required correction: resolve a valid active administrator through `staff_access`, or use an explicitly supported system audit actor. Handle lookup failure/missing actor visibly. Fail the maintenance invocation appropriately; do not silently invent an empty identity. Test the real worker cleanup command against an isolated schema containing an aged orphan, and verify deletion plus a valid audit actor. Also test the no-eligible-actor case. Keep the communications worker's existing dispatch behavior intact.

## L3 — Referenced assets can permanently occupy the entire cleanup batch

Location: `services/api/internal/adapters/postgres/cms_settings.go:452–461`.

The candidate query orders all old public nonclinical assets and applies `LIMIT 100` before checking references. If the first 100 tokens are referenced, every invocation selects and skips those same rows. Later unreferenced assets are never considered, regardless of age or number of runs.

Independent reproduction: insert 101 aged public nonclinical attachments in ascending token order, reference the first 100 from CMS, and invoke cleanup twice. The unreferenced 101st attachment remains. Probe result: **unreferenced asset 101 remains after repeated cleanup because referenced assets occupy the first 100 slots**.

Required correction: filter likely unreferenced candidates before applying the limit, then recheck references under the acquired token locks; alternatively use a progressing bounded cursor. Preserve the shared binding/retirement protocol and `SKIP LOCKED` behavior. Add a more-than-100 fixture regression proving progress across runs while keeping referenced, recent pending, and clinical assets intact.

## Focused handoff

Fix L1–L3 in one lifecycle-only batch. Retain the now-working edit flows and attachment race protection. Run the existing gates plus the three regressions, including the actual worker CLI rather than only its service method. Return the commit and evidence for acceptance; keep main and Inventory unchanged until then.
