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
