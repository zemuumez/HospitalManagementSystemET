# Frontend review checkpoints

## Step 1 — role navigation and personal records

Compared `resources/views/layouts/menu.blade.php` (all nine role blocks), `livewire/patient-dashboard.blade.php`, employee view directories and patient/employee Livewire table column definitions. Implemented ordered role menus, persistent role preview, patient dashboard tiles/recent appointments and dedicated personal-record routes, staff doctor directory, notices and payroll views. The staff dashboard is a preview navigation hub: the source does not define a distinct dashboard for every employee role.

Patient and employee previews use synthetic fixtures only. IPD/OPD empty subordinate tabs do not imply completed clinical histories. Role preview is not authorization. Existing backend permissions are unchanged. Optional Holiday add-on is not included in these menu mappings.

Verification: TypeScript passed; browser checked all nine menus for unresolved destinations, opened every patient portal and its detail dialog, and confirmed role persistence after refresh. Browser assertion was updated to wait for client hydration.

Live comparison: `https://hms.infyom.com/dashboard` failed through the web reader and timed out after 60 seconds in Chrome on this pass. Deployed role sessions remain unverified. Source and supplied screenshots are the available comparison evidence.

## Step 2 — missing views and workflow behavior

Restored public doctor detail/overview/schedule pages and booking links from `web/home/doctor-details.blade.php`. Appointment availability is a labeled synthetic schedule, with Sunday unavailable and doctor/date changes resetting the selected slot. Hospital forms now filter doctors by department, cases by patient, and beds by type; parent changes clear stale child values. Appointment saves reject unavailable/duplicate slots. Admission/discharge and birth-date validation runs locally.

Restored prescription fields from `physical-info-fields.blade.php` and `other-fields.blade.php`, previously omitted by the form extractor. Prescription details show saved line items. Discharge summaries persist per sample record; configured preview custom fields are injected into matching module forms. Two-factor setup/recovery/regeneration/disable screens follow the source interaction stages, explicitly as a demo: there is no real QR enrollment, secret or account-security mutation.

Correction to the earlier audit: `departments/add_modal.blade.php` says “New Role”, and `routes/web.php:1102–1103` comments out its routes. It is a dormant role-management fragment, not an active missing hospital-department workflow. The separate Doctor Departments screen already exists.

Verification: TypeScript and three unit tests passed. Browser checked dependent doctor and slot choices/reset, prescription partial fields, two-factor demo errors/enable/regenerate/disable, and public doctor schedule/booking. Real stock, availability, clinical enforcement, custom-field backend schemas and document PDF generation remain backend work. This checkpoint does not assert exhaustive clinical parity.

## Step 3 — English and Amharic coverage

Added Amharic entries for all catalog screen titles, groups, columns and form labels, the omitted prescription partials, CMS fields, role names/personal views, form actions, and common validation/feedback. Static accessibility labels and dropdown text use the same translator; option values stay canonical English so language changes do not change stored values. Dates displayed in record tables use localized Gregorian formatting; this does not switch the clinical calendar to the Ethiopian calendar or change currency denomination.

Added a coverage test for all source-derived catalog/CMS/prescription labels. Browser checks passed for Amharic prescriptions, language switching with entered data preserved, patient dashboard, refresh persistence, 390px overflow and translated browser validation. Screenshots inspected. Proper names, medicine names, identifiers and user-authored content are preserved. Provider-returned error messages can still fall back to English. Native-speaker/clinical review of the new Amharic wording remains required; there is no original Amharic pack to compare.

## Step 4 — audit, fixes and reproducible checks

Added `npm run test:frontend` using a pinned Playwright dependency. It requires an existing development account through environment variables, refuses remote targets, and uses a fresh browser context. Only synthetic frontend records are changed. It records screenshots, CSV/card downloads and machine-readable results under ignored `.local/frontend-verification`.

The audit covers 111 source catalog routes and their available create forms, five supplemental screens, fourteen public URLs, nine role navigation sets and personal detail dialogs, schedule validation/persistence, per-patient dental persistence, template rename/card references, custom-field injection, CRUD/search/export, appointment conflicts, role-specific billing tabs, queue-theme persistence, and desktop/390px page overflow.

Defects found and fixed: schedule/dental/queue settings did not restore saved values; dental state was not patient-specific; template renaming orphaned card references; personal-table sorting reversed rows instead of sorting the chosen column; custom fields were injected into the wrong form branch; case-only forms lost their case options after dependent-select changes; the admin chart overflowed mobile width. Full-page editors now account for the sticky header when scrolled into view. Role tabs were compared with `layouts/sub_menu.blade.php`, and visibility now follows those source role blocks. Initial role selection follows the signed-in role unless a preview role was explicitly saved.

Test corrections are separate from product fixes: required-radio fixture handling, special screens without an H1, ambiguous labels, and waiting for React readiness and navigation before checking role menus.

Final verification: **157 unique browser checks passed** across the catalog run and the targeted follow-up. The 111 catalog checks passed; the follow-up passed all 46 public/role/regression/layout checks, resolving two test-timing failures. Both runs had zero browser runtime errors. Desktop/mobile screenshots were inspected. TypeScript, five unit tests, formatting and the production build are checked separately. The new Playwright dependency audit reported zero vulnerabilities.

## Remaining parity boundaries before claiming a complete original clone

- The deployed demo still cannot be reached from this environment; `/dashboard`, `/login` and `/` attempts failed. Its role sessions and any add-on/version differences have not been visually verified.
- Personal IPD/OPD subordinate tabs have empty sample histories. They do not constitute a full longitudinal clinical record. Some specialized modules still use the shared source-derived form/table renderer.
- Print uses browser layouts and CSV/card downloads. Original clinical/billing PDF templates and signed document exports need a separate output-layout pass; browser printing is not exact PDF parity.
- Real appointment availability, stock, clinical/financial constraints, payment-provider flows, patient self-service submission, secure attachments, QR patient lookup and actual two-factor enrollment require backend contracts and enforcement. Preview behavior must not be treated as the final business specification.
- Proper names and user-authored content remain untranslated. Amharic needs native-speaker/clinical review, and provider-generated errors may remain English. Calendar dates remain Gregorian.

The original discovery ledger remains authoritative about source files not yet exhaustively reviewed. This pass does not claim to have read every line of the archive or reproduced every original business rule.
