# Doctor workspace batch review

Reviewed HEAD: `7b94c0a4092d9da7590deac23226eb6f46b6afc7`, including deletion and Steps 3–6. Verdict: changes required before merge. This review does not certify exhaustive original UI parity.

## Findings and correction task

### W1 — Booking/availability concurrency (high)

In `services/api/internal/adapters/postgres/scheduling.go`, booking locks the doctor's staff_access row before checking availability. CreateDoctorHoliday (around line 1226) and CreateDoctorBreak (around line 1366) check appointments and insert exclusions without acquiring that same lock. Concurrent booking and exclusion creation can both observe no conflict and commit. DeleteDoctorSchedule likewise checks appointments without the booking lock before deleting hours. This is a code-derived race finding, not a claimed runtime reproduction.

Use a common doctor lock and consistent lock order for every operation changing or consuming availability, including rescheduling. Add deterministic concurrent booking-versus-holiday, booking-versus-break and booking-versus-schedule-deletion tests proving incompatible states cannot both commit.

### W2 — Frontend reintroduces incorrect defaults and fabricated clinical attributes (high)

`apps/web/src/components/doctors-workspace.tsx` around lines 463–469 sends `slotMinutes: 15` on every create, overriding the accepted original 60-minute default. Blank designation and qualification become "Senior Consultant" and "MD"; gender falls back to male. Remove synthetic credential fallbacks, require actual required input, and use the accepted backend schedule defaults unless explicitly configured. Test blank required fields and default creation through the browser and database.

### W3 — Original doctor form/detail coverage remains incomplete

The create payload omits DOB, blood group and address fields already supported by the accepted backend; the submitted screenshots show the abbreviated modal rather than complete original create/detail coverage. Reconcile every source-derived doctor field/action against the existing parity plan, then wire missing create/edit/detail fields. Do not mark the workspace complete based only on the subset exercised by the current browser script.

### W4 — Visible UI regressions

Submitted screenshots show repeated Doctor Holidays tabs, two competing navigation rows, very pale text on white backgrounds, and a developer-facing PostgreSQL status banner. The details component hardcodes near-white values (`#f8fafc`, `#e2e8f0`) on a white modal. Use existing light/dark theme tokens, one correctly ordered original navigation, and role-appropriate tabs/actions. Remove database implementation details from normal hospital UI. Verify readable light/dark screenshots, English/Amharic and the original page/form structure.

### W5 — Browser evidence is not reliable or isolated (high)

`scripts/verify-doctor-workspace.mjs` defaults to the development database and localhost:3000, contains embedded database/login credentials, writes to a machine-specific screenshot directory, and closes connections without removing created records. Do not run this script against normal development data during review. Remove embedded secrets and rotate the exposed development credentials; load credentials from the isolated runner and use temporary output paths.

Schedule verification around lines 205–223 only waits for `.billing-table` (also present in the form) and checks whether the doctor's name appears on the page. The supplied "schedule saved" screenshot still shows "Saving...". Other assertions search page-wide values that can match old rows. Wait for the specific successful response, assert the exact new record, reload, and verify persisted slot length/hours. Apply the same record-specific pattern to holidays, breaks and charge edits. Add failure/empty states, status revocation and unreferenced deletion browser cases. Do not equate a screenshot filename or log message with successful persistence.

### W6 — Required formatting check fails

Independent `npm run format:check` reports errors in doctors-workspace.tsx and verify-doctor-workspace.mjs. Format these files and rerun.

## Verification and scope

Independent Go suite with PostgreSQL enabled passed. Typecheck and all 8 frontend unit tests passed. The connected isolated integration runner passed, including deletion checks across Next.js and Go proxy paths, and removed its test schema. The submitted browser script was inspected, not executed, because it targets the normal development database and lacks fixture cleanup. Supplied screenshots were reviewed as delivery evidence, not represented as independently captured screenshots.

Complete W1–W6 as one correction batch with separate logical commits and tests, update the completion checklist honestly, and return one delivery for independent review. Keep main unchanged. Do not begin another workspace yet.
