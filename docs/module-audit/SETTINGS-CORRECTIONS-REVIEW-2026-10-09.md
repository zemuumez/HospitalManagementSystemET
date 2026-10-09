# Settings correction batch independent review

Submitted branch: `feat/settings-workspace-parity`.
Reviewed HEAD: `075679e49c0285b69d1fa33b66c1339d76476cc0`.
Verdict: **changes required before acceptance or merge**. Two remaining correction areas are listed below. Main was not changed by this review.

## Independent verification

- All 17 supplied isolated Settings browser journeys passed (exit 0). The runner reported successful isolated-schema cleanup.
- Full Go suite passed with PostgreSQL enabled and `-count=1`.
- TypeScript typecheck, all 8 frontend tests, and formatting passed.
- Two additional temporary PostgreSQL regression probes both failed as described in R1. These probes were run against the existing isolated clinical test harness, then removed; implementation files remain unchanged.
- Original Laravel templates in the supplied `hms.zip` were inspected to confirm the missing edit actions in R2.

The submitted code fixes the original blank tabs, adds aggregate schedule/CMS transactions, strengthens required-field/time validation, skips unchanged secret markers in the batch write, removes duplicate navigation, and improves the previously reported light-mode styling. These improvements do not resolve the two remaining issues.

## R1 — Attachment retirement deletes pending and referenced assets (high)

Locations: `services/api/internal/adapters/postgres/cms_settings.go:228`, `services/api/internal/adapters/postgres/attachments.go:153`, and the retirement call after general/CMS saves in `services/api/internal/application/cms_settings.go`.

`RetireUnreferencedAttachments` scans **every** public nonclinical attachment after any successful general/CMS save, including uploads that another administrator has just selected but has not yet saved into a setting. Upload and settings submission are separate requests in the UI. A second session saving an unrelated company name, currency, or CMS field can therefore delete the first session's pending upload. The first session can then save the now-invalid URL successfully. There is no pending-upload protection or age threshold.

Independent reproduction using the actual services and PostgreSQL:

1. Upload a public image with `AttachmentsService.UploadPublic` and retain its token without saving a settings reference.
2. Call `CMSSettingsService.UpdateGeneralSettings` with only `company_name` changed, with file storage configured as in the API.
3. Download the first token. Expected: pending image remains available. Actual: **record not found**.

The new direct retirement operation has a second unsafe path: `DeleteAttachment` checks public/nonclinical status but never checks settings references. Uploading an image, saving it as `app_logo`, then calling `RetireAttachment` succeeds while `app_logo` still points to it. The second independent probe failed with **retirement succeeded while app_logo still references the asset**.

The sweep also selects candidates outside its deletion transaction and deletes by ID without rechecking references. A reference added between selection and deletion is not protected.

Required correction: retire only assets displaced by the successful update, preserve all shared references, and coordinate attachment binding and deletion transactionally. Protect pending uploads; handle abandoned uploads with a separate bounded expiration/cleanup lifecycle. Reject direct retirement of referenced assets. Do not silently lose failed file removals: retain enough state for retry or report cleanup failures. Keep clinical attachment protections intact.

Acceptance coverage: pending upload plus unrelated save; overlapping bind/retire; shared reference in general settings and CMS; direct retirement while referenced; ordinary replacement/removal; abandoned-upload cleanup; old-token 404 only after safe retirement; actual image decoding after reload/fresh session.

## R2 — Operation Categories and Custom Fields still cannot be edited

Locations: `apps/web/src/components/settings-workspace.tsx:2435` and `:3060`; `services/api/internal/adapters/httpapi/services_operations.go` operation-category/custom-field handlers.

Both restored tabs implement creation, listing/filtering, and deletion, but neither offers an edit action or submits an update request. Their HTTP handlers likewise have no update route. Operations has an edit flow; the other two do not. Journeys 14 and 16 never edit their created records, so their passing "CRUD" messages do not establish update support.

Original source evidence:

- `hms/resources/views/operation_categories/action.blade.php` contains the `operation-category-edit-btn`; `edit_modal.blade.php` provides a PATCH name-edit form.
- `hms/resources/views/add_custom_fields/action.blade.php` contains `updateCustomFieldBtn`; `edit_custom_field_modal.blade.php` provides the original edit dialog.

Required correction: implement source-derived edit dialogs, authorized update routes, validation and persistence for both tabs. Preserve record identity and existing references; deleting/recreating a record is not an edit substitute. Cover edits after reload and a fresh session, rejected updates without mutation, and non-admin denial. Include both flows in the browser suite before claiming full CRUD parity.

## Evidence limits to address in the same batch

Journey 17 checks two light-mode computed colors and one Amharic heading. It never switches to dark mode; it does not prove dark-mode or every-tab localization parity. Add the previously requested light/dark checks and record the actual scope of localization/source-field verification. The secret regression is sequential stale-client coverage, not an overlapping concurrency test, although skipping marker writes resolves the inspected batch overwrite mechanism. Report test scope accurately.

The final handoff should contain one correction batch for R1/R2 and the associated evidence updates, with logical commits, expanded isolated verification, and screenshots taken after loading completes. Keep Inventory on hold until Settings is accepted. No changes to main were made here.
