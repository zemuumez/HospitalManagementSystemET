# Frontend review checkpoints

## Step 1 — role navigation and personal records

Compared `resources/views/layouts/menu.blade.php` (all nine role blocks), `livewire/patient-dashboard.blade.php`, employee view directories and patient/employee Livewire table column definitions. Implemented ordered role menus, persistent role preview, patient dashboard tiles/recent appointments and dedicated personal-record routes, staff doctor directory, notices and payroll views. The staff dashboard is a preview navigation hub: the source does not define a distinct dashboard for every employee role.

Patient and employee previews use synthetic fixtures only. IPD/OPD empty subordinate tabs do not imply completed clinical histories. Role preview is not authorization. Existing backend permissions are unchanged. Optional Holiday add-on is not included in these menu mappings.

Verification: TypeScript passed; browser checked all nine menus for unresolved destinations, opened every patient portal and its detail dialog, and confirmed role persistence after refresh. Browser assertion was updated to wait for client hydration.

Live comparison: `https://hms.infyom.com/dashboard` failed through the web reader and timed out after 60 seconds in Chrome on this pass. Deployed role sessions remain unverified. Source and supplied screenshots are the available comparison evidence.
