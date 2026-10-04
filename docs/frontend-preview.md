# Original-style frontend preview

The user changed the delivery order: recreate the Laravel UI closely, finish frontend review first, and defer modernization. The teal redesign has been replaced.

## Available for review

- Poppins fonts copied from the supplied project, original 265px sidebar geometry, purple primary/active colors, horizontal module navigation, original-style tables, switches, form spacing, full-page large forms, detail tabs and smaller dialogs.
- 111 module routes, including nested clinical screens. The source inventory contains 674 extracted/default fields; 93 forms have fields extracted from Blade. Presentation overrides provide other screen types and sample selectors.
- Searchable module navigation, nine role menu previews, collapsible desktop navigation, mobile navigation and light/dark presentation.
- Search, sorting, status filters, pagination, CSV export, sample create/edit/delete, detail views and tab-scoped session storage. Password inputs are never saved in preview storage.
- Appointments/calendar, bed availability, patient/staff forms, IPD/OPD subforms, invoice line items and totals, medicine purchase lot/expiry fields, patient card previews, hospital schedule and the original odontogram SVG/marking palette.

Open `/patients` or `/dashboard`. Use the top **View as** selector to inspect role navigation. This selector changes presentation only; it does not grant API permissions.

## Data and scope

All expanded module screens use explicitly marked synthetic records. Their edits stay in session storage in the current browser tab. They do not create hospital records, send messages, upload files, charge payments or start consultations. File controls currently retain only the selected filename. Reference dropdown choices are sample values. These are interactive frontend previews, not connected operational workflows.

The existing authenticated patient API remains available at `/live-patients`; actual operational outbox UI remains at `/communications`. Authentication, account phone linking and Go authorization are unchanged. Existing credentials and sessions were preserved.

## Verification

- All 111 catalog routes returned HTTP 200 and rendered in authenticated Chrome.
- Sample create/edit/delete, search, reload persistence, patient full-page form, clinical diagnosis creation and role menu filtering passed.
- Medicine bill quantity/rate/discount/tax calculation and stored line-item reload passed (2 × 125, less 10, plus 10% = 264).
- Original odontogram tooth marking passed.
- Desktop and 390px mobile screenshots reviewed; mobile navigation and overflow checks passed. No browser page errors occurred during the checks.
- TypeScript and production build passed.

## Parity limits still to resolve

This preview must not be described as fully verified feature parity. The original Laravel application has not been run against its original database, so there is no screenshot-by-screenshot comparison with the user's configured installation. Fields were recovered from source, but conditional forms, dependent selectors, role-specific actions, localized text, report/print layouts, dynamic custom fields, public website/patient self-service flows and all workflow edge cases still need individual acceptance review. In particular, menu previews are presentation mappings, not a verified replica of the legacy permission matrix.

`frontend-coverage.md` is the extraction inventory; `scripts/extract-frontend.py` regenerates it. Source-derived metadata is in `legacy-catalog.json`, presentation overrides in `legacy.ts`, and the shared renderer in `legacy-screen.tsx`. Inventory coverage and passing route tests do not establish behavioral parity. Backend wiring and production security checks remain separate deliverables.
