# Settings workspace independent review

Reviewed branch: `feat/settings-workspace-parity`, submitted HEAD `cec596e`. Verdict: changes required; do not merge or begin Inventory yet.

## Independent verification

All 13 isolated Settings browser journeys passed, and the runner removed its isolated schema. Full Go tests with PostgreSQL enabled, typecheck, all 8 frontend unit tests and formatting passed. Passing checks do not establish the missing cases below. Findings are based on code inspection and supplied screenshots unless stated otherwise.

## S1 — Three Settings tabs render no content (high)

`apps/web/src/app/(hospital)/modules/[slug]/page.tsx` now routes operation-categories, operations and add-custom-fields to SettingsWorkspace. `apps/web/src/components/settings-workspace.tsx` renders none of those IDs: only the navigation appears. Restore existing connected components or implement the source-derived CRUD content. Merely including links does not preserve these workflows. Add browser CRUD, reload and permission coverage for all three tabs.

## S2 — Schedule and CMS batch updates are not atomic (high)

`CMSSettingsService.UpdateHospitalSchedules` and `UpdateFrontCMSSettings` validate and save each item sequentially. Each repository single-item method opens and commits its own transaction. A valid first item followed by an invalid second item therefore leaves the first change and audit event committed despite failure. A later database error has the same effect. Implement repository-level aggregate transactions covering all writes and audit events; validate the entire input before mutation. Reject duplicate day/key entries appropriately. Test invalid trailing items and an injected late database failure, asserting unchanged values and audit state. Do not label the existing successful-save test as proof of rollback.

## S3 — General-settings validation is only in the browser

`GeneralSettingInput.Validate` checks only nonempty key and key length. Direct requests can save blank required fields and malformed emails/currency/day/time values despite the documented field contract. Add an explicit key/type contract and server validation with appropriate partial-update semantics. HospitalScheduleDayInput compares time strings instead of parsing times; validate actual HH:MM values, including closed-day behavior. Test HTTP 422 and no mutation for malformed direct requests, not only HTML required-field behavior.

## S4 — Attachment lifecycle and visual retrieval are incomplete

Upload creates a public secure_attachment immediately. Remove only clears the setting string; replacement likewise leaves the old token/file publicly retrievable. There is no completed retirement/reference-management operation in this delivery. Define and implement safe asset replacement/removal, preserving shared references and retiring unreferenced assets after successful settings commit; handle failed/cancelled uploads without accumulating public orphan files. Keep clinical files private.

The supplied screenshots contain broken logo/favicon previews. Journey 6 checks HTTP 200 and image/png, but never decodes image bytes or asserts an image element has naturalWidth > 0. It tests clearing logo_url, not retirement or replacement. Use valid image fixtures and verify actual rendering after reload and a fresh session, replacement, removal and access to retired tokens. Diagnose the broken preview before claiming rendering is complete; this review does not assert an unproven root cause.

## S5 — Preserve-secret behavior can overwrite a concurrent replacement

`UpdateGeneralSettings` reads existing secrets before the write transaction and substitutes their plaintext for [CONFIGURED]/********. An unchanged-secret save can subsequently write that stale value over a concurrent replacement. Skip unchanged-secret fields instead of rewriting their old values, or resolve them under the same transaction/lock. Add an overlapping preserve-versus-replace regression and retain redacted GET/POST responses, explicit clearing and role-denial tests. Avoid interpreting successful sequential tests as concurrency coverage.

## S6 — UI parity and evidence coverage

The new component repeats the shell's tab navigation. The currency table applies text-white to names on white rows; supplied screenshots show unreadable currency names. Payment panels and schedule/module cards hardcode dark surfaces in light mode with mismatched text. Reuse theme tokens and canonical navigation; verify light/dark and English/Amharic views. Extend the source field/action matrix for every Settings/CMS tab, including the original image fields, rather than declaring all tabs complete after testing only Home text fields.

## Delivery instructions

Address S1–S6 in one correction batch with separate logical commits and focused tests. Update the checklist to distinguish persisted settings from downstream behavior verified in consumers. Rerun the expanded isolated browser suite, Go/PostgreSQL, typecheck, unit tests and formatting. Return commit hashes and screenshots for one review. Main remains unchanged by this review.
