# Doctor workspace correction re-review

Reviewed `14ebe7b` on feat/doctor-deletion-parity. Changes still required for edit-flow parity; no merge performed.

## Independent results

- Full Go suite `go test ./... -count=1` with PostgreSQL configuration: passed.
- Typecheck, 8 frontend tests and formatting: passed.
- Isolated `--doctor-workspace` runner: all 16 browser journeys passed after setting SCREENSHOT_DIR to a writable workspace-local directory. Its isolated schema was removed.
- The unmodified runner initially failed with EPERM writing into the hardcoded Antigravity screenshot directory. No application failure was inferred from that permission error.
- Independently captured details screenshot confirms readable light-theme fields, complete demographics/address display and the 60-minute default. Dark theme and complete original UI parity were not independently certified.

W1 locking corrections are present across holiday/break creation and schedule deletion; booking/rescheduling share the staff_access row lock (a row lock, not an advisory lock). Concurrent regression cases are included in the passing database suite. Fabricated required-field fallbacks and forced 15-minute creation have been removed. The isolated browser harness now verifies exact persisted records. Formatting is resolved.

## R1: Optional fields cannot be cleared

`apps/web/src/components/doctors-workspace.tsx:604` and adjacent update payload fields use `value || undefined` for phone, dateOfBirth, bloodGroup, address1/address2, city, zip and description. JSON.stringify drops these properties. The Go update contract treats omission as preserve-existing, so clearing an existing value reports success but restores the old value after reload. Send explicit empty values for deliberate clearing, supported by validation/storage; retain omission only for unchanged or unauthorized fields. Required designation/gender should reject blank edits rather than silently preserving old values.

Regression: populate optional fields, clear them through the edit UI, save, reload and assert the exact empty database/detail state. Verify required fields still reject blank input.

## R2: Doctor self-edit submits forbidden department

The update payload always includes `departmentId: editDocDeptId || undefined` (`doctors-workspace.tsx:601`). Existing doctors have that ID populated. `Scheduling.UpdateDoctorProfile` explicitly rejects every non-nil DepartmentID for doctor actors (`services/api/internal/application/scheduling.go:92`). Therefore the offered own-profile edit fails with 403 even when only phone or qualification is changed. Do not weaken the server guard: omit department for doctor actors and disable/hide that control appropriately. Add a browser self-edit/persistence test and keep a direct unauthorized department-change test. Current journey 16 only checks hidden/disabled controls and does not exercise self-edit.

## R3: Admin email editing remains absent

The edit modal has no email input/state and handleUpdateDoctor never includes email, despite the accepted backend email-update feature and original account edit requirements. Wire an admin email field through the existing endpoint; verify valid update persists, malformed values fail without mutation and duplicate values return conflict. Preserve existing role rules.

## Runner portability follow-up

Remove the personal Antigravity screenshot fallback. Set SCREENSHOT_DIR from the isolated wrapper to its own temporary artifact directory. Make the browser script require the isolated-schema marker before fixture writes, and avoid claiming guaranteed cleanup while swallowing cleanup errors outside that wrapper. Embedded credentials have been removed from current script content; credential rotation was not independently verified.

## Next action

Deliver these edits and regressions as one bounded correction batch on the same feature branch. Reuse the existing implementation. Keep main unchanged and do not expand to another module. The passing suites above establish substantial progress but do not cover the three edit-flow gaps identified by code inspection.
