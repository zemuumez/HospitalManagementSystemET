# Settings correction re-review

Reviewed HEAD: `28cd388c295e8f18d92129e2cd9de1d5bbaf2fe5` on `feat/settings-workspace-parity`.

Verdict: **R2 closed; R1 partially resolved, acceptance pending attachment corrections below.** Main was not changed. Keep Inventory on hold until this remaining lifecycle work passes review.

## Independent results

- All 17 isolated browser journeys passed, exit 0; isolated schema cleanup completed.
- Full Go/PostgreSQL suite with `-count=1`, TypeScript typecheck, all 8 frontend tests, and formatting passed.
- Operation Category and Custom Field edit dialogs, update endpoints, validation/permission tests, and browser reload checks are now present. R2 is closed for this review.
- Sequential pending-upload survival, direct in-use rejection, and cross-table shared-reference protection are covered and pass.
- Journey 17 actually switched to dark mode in this run: card background `rgb(18, 21, 31)`, heading `rgb(255, 255, 255)`. General Settings and Currencies Amharic checks passed. This is scoped evidence, not an exhaustive visual/source parity certification.
- An additional deterministic PostgreSQL concurrency probe failed: `successfully bound CMS image was retired: record not found`. The temporary probe was removed after execution; production and existing test files remain unchanged by this review.

## R1a — Binding and retirement still do not share a lock protocol (high)

Locations: `services/api/internal/adapters/postgres/cms_settings.go:62`, `:113`, `:288`, `:339`; `services/api/internal/adapters/postgres/attachments.go:153`.

`retireDisplacedTokensTx` checks references before obtaining its attachment row lock. More fundamentally, General Settings and Front CMS writes do not lock/validate newly referenced attachment tokens at all. A retirement transaction and a binding transaction can both commit successfully, leaving a persisted setting pointing to a deleted token. Merely moving the reference check below `FOR UPDATE` will not solve this unless binding participates in the same protocol.

Deterministic reproduction using the actual services and isolated PostgreSQL harness:

1. Upload image T and save `app_logo` pointing to T.
2. In a separate test transaction, hold `SELECT ... FROM secure_attachment WHERE token=T FOR UPDATE`.
3. Start `UpdateGeneralSettings({app_logo: ""})` in a goroutine. Wait via `pg_blocking_pids` until retirement is blocked on the held attachment lock, after its reference check.
4. Through `UpdateFrontCMSSettings`, save `review_banner` pointing to T. This succeeds while retirement is blocked.
5. Release the test lock. The first settings save also succeeds.
6. Query the CMS value: it still contains T's URL. Download T: **record not found**.

Required correction: adopt one transaction/lock order for all incoming asset references, displaced retirement, direct retirement, and abandoned cleanup. Lock relevant tokens in stable order before checking references or binding; validate that newly bound assets exist and are appropriate public nonclinical assets. Read displaced values under the appropriate settings locks as well. A save attempting to bind an already-retired token must fail without persisting a broken URL. Coordinate single-item settings methods too if they remain supported mutation paths. A normalized reference table with foreign keys is another valid implementation; it is not required if the locking protocol enforces the same invariant.

Keep a deterministic regression for the overlap above and both ordering outcomes: binding wins and retirement preserves the reference, or retirement wins and binding rejects the missing token. Exercise direct deletion and cleanup through the same protocol, retaining pending-upload and shared-reference cases.

## R1b — Abandoned cleanup exists but is not invoked

`CleanupAbandonedAttachments` is called only by tests and an otherwise unused compatibility wrapper. There is no application worker, scheduled job, or maintenance command invoking it. An upload abandoned by leaving the form therefore remains publicly retrievable indefinitely in normal operation, despite the new age-threshold method.

Wire a supported operational cleanup path with an explicit age threshold, bounded work per run, observable errors, and documented execution. Test that actual entry point with aged unreferenced, recent pending, referenced, and clinical fixtures. Apply R1a's reference/binding coordination to cleanup. Also make direct retirement file-removal errors observable; `AttachmentsService.RetireAttachment` still discards them even though the settings-save paths now log failures.

## Scope and next delivery

One focused attachment correction batch is sufficient; do not rebuild the edit dialogs or expand into Inventory. Rerun the existing gates plus the deterministic binding/retirement regression and operational cleanup test. Return commits and evidence for this specific remaining scope.

Evidence wording: the new ten-goroutine secret test overlaps marker-preservation and unrelated-field saves, but none of those goroutines replaces the secret. It should not be described as overlapping preserve-versus-replace coverage. The reviewed skip-marker implementation remains sound for the previously reported stale-secret overwrite; this wording issue is not a new acceptance blocker.
