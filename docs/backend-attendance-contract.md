# Attendance backend contract

## 1. Overview and source parity

Original reference audit:
- Inspection of the legacy Laravel repository (`review/legacy/` and `hms.zip`) reveals no native attendance models or migrations; the core InfyOm HMS distribution omitted staff attendance.
- The reference specification (`review/specification.txt`) only included Doctor Schedules, Breaks, and Holidays.
- Attendance was added in newer deployments as an add-on module, documented in screenshot 3 and `docs/frontend-parity-audit-2026-10-04.md`.
- Screenshot 3 depicts two screens: **Attendance** (staff personal records and check-in/out) and **Manage Attendance** (administrator view of all staff records with manual add/edit).
- Display columns: `Staff`, `Date`, `Shift`, `Status`, `Check In`, `Check Out`, `Late`, `Early Out`, `Worked`, `Over Time`, `Break Time`, `Action`.

This increment establishes the typed Go backend contract for Section 3.I, enforcing domain rules, permissions, shifts, break tracking, approval state transitions, immutable audit/correction history, and precise calculation rules in `Africa/Addis_Ababa`. UI integration and payroll remain Section 4 work.

---

## 2. Data model and migration 028

### `attendance_shift`
Defines operational shift profiles:
- `id` (UUID, primary key)
- `name` (text, unique, 1–100 chars)
- `start_time` (time of day, e.g. `08:00:00`)
- `end_time` (time of day, e.g. `17:00:00`)
- `grace_period_minutes` (integer 0–120, default 15)
- `break_duration_minutes` (integer 0–300, default 60)
- `half_day_minutes` (integer > 0, default 240)
- `full_day_minutes` (integer >= half_day_minutes, default 480)
- `is_overnight` (boolean, true if shift spans past midnight)
- `active` (boolean, default true)
- `version` (integer optimistic lock)
- Default seeds: `Day Shift` (`08:00`–`17:00`, regular) and `Night Shift` (`20:00`–`05:00`, overnight).

### `attendance_shift_assignment`
Assigns staff to specific shifts:
- `id` (UUID, primary key)
- `staff_id` (foreign key to `"user"(id)`)
- `shift_id` (foreign key to `attendance_shift(id)`)
- `effective_from` (date in EAT)
- `effective_to` (date in EAT, nullable; must be >= `effective_from`)

### `attendance_record`
Stores the daily attendance session:
- `id` (UUID, primary key)
- `staff_id` (foreign key to `"user"(id)`)
- `work_date` (date in `Africa/Addis_Ababa`, unique per staff member)
- `shift_id` (foreign key to `attendance_shift(id)`)
- `check_in_at` (timestamptz, required)
- `check_out_at` (timestamptz, nullable while active)
- `status` (`checked_in`, `present`, `late`, `half_day`, `absent`)
- `approval_status` (`draft`, `submitted`, `approved`, `rejected`)
- `total_break_minutes`, `worked_minutes`, `late_minutes`, `early_out_minutes`, `overtime_minutes` (integers >= 0)
- `source` (`self_service`, `administrator`)
- `admin_notes` (text, max 1000 chars)
- `version` (integer optimistic lock)
- Constraints: `UNIQUE(staff_id, work_date)`, `CHECK(check_out_at IS NULL OR check_out_at >= check_in_at)`.

### `attendance_break`
Tracks breaks taken during attendance:
- `id` (UUID, primary key)
- `attendance_record_id` (foreign key with `ON DELETE CASCADE`)
- `start_at` (timestamptz, required)
- `end_at` (timestamptz, nullable while active)
- `duration_minutes` (integer >= 0)
- `reason` (text, max 500 chars)

### `attendance_correction`
Immutable audit log of administrator-entered changes:
- `id` (UUID, primary key)
- `attendance_record_id` (foreign key with `ON DELETE CASCADE`)
- `actor_id` (foreign key to `"user"(id)`)
- `reason` (text, 1–1000 chars, required)
- `before_snapshot` (jsonb)
- `after_snapshot` (jsonb)
- `created_at` (timestamptz)
- Protected by trigger: `immutable_attendance_correction` executes `protect_retained_record()` preventing SQL `UPDATE` or `DELETE`.

### `attendance_approval_history`
Immutable record of approval state transitions:
- `id` (UUID, primary key)
- `attendance_record_id` (foreign key with `ON DELETE CASCADE`)
- `actor_id` (foreign key to `"user"(id)`)
- `from_status` (text)
- `to_status` (text)
- `reason` (text, max 1000 chars)
- `created_at` (timestamptz)
- Protected by trigger: `immutable_attendance_approval_history` executes `protect_retained_record()` preventing SQL `UPDATE` or `DELETE`.

---

## 3. Calculation rules and timezones

All times and operational work dates are calculated in hospital timezone `Africa/Addis_Ababa` (UTC+3, EAT, no DST):
1. **Scheduled Start & End**:
   - `scheduledStart = workDate + shift.start_time`
   - `scheduledEnd = (workDate + (1 day if is_overnight else 0)) + shift.end_time`
2. **Lateness**:
   - Grace threshold: `scheduledStart + shift.grace_period_minutes`
   - If `check_in_at > grace_threshold`: `late_minutes = int(check_in_at - scheduledStart)`
   - Else: `late_minutes = 0`
3. **Early Out**:
   - If `check_out_at < scheduledEnd`: `early_out_minutes = int(scheduledEnd - check_out_at)`
   - Else: `early_out_minutes = 0`
4. **Worked Time**:
   - Elapsed minutes: `check_out_at - check_in_at`
   - `worked_minutes = max(0, elapsed - total_break_minutes)`
5. **Overtime**:
   - If `worked_minutes > shift.full_day_minutes`: `overtime_minutes = worked_minutes - shift.full_day_minutes`
   - Else: `overtime_minutes = 0`
6. **Attendance Status**:
   - Open record (`check_out_at IS NULL`): `checked_in`
   - Closed record:
     - If `worked_minutes >= shift.full_day_minutes`: `late` (if `late_minutes > 0`) else `present`
     - Else if `worked_minutes >= shift.half_day_minutes`: `half_day`
     - Else: `absent` (worked time below half-day threshold)
7. **Overnight Shifts**:
   - An overnight shift beginning at 20:00 on Date D and ending at 05:00 on Date D+1 maintains its operational `work_date` as Date D. Cross-midnight intervals do not reset or produce negative values.

---

## 4. Permissions and routes

Role policy:
- `admin`: full management (`attendance.manage`, `attendance.clock`, `attendance.read_own`)
- Staff roles (`doctor`, `nurse`, `receptionist`, `pharmacist`, `accountant`, `case_manager`, `lab_technician`): self-service clock in/out, breaks, and viewing own records (`attendance.clock`, `attendance.read_own`)
- `patient`: forbidden from all attendance endpoints (returns 403 Forbidden).

### Shift management
- `GET /v1/attendance/shifts`: any active staff member; returns all shifts.
- `POST /v1/attendance/shifts`: admin only; creates a shift definition.
- `PATCH /v1/attendance/shifts/{id}`: admin only; updates shift with version check (`version` required).

### Shift assignment
- `GET /v1/attendance/assignments?staffId=...`: admin can query any staff member; regular staff can only query their own ID.
- `POST /v1/attendance/assignments`: admin only; `{staffId, shiftId, effectiveFrom, effectiveTo}`.

### Self-service attendance
- `GET /v1/attendance/today`: returns authenticated staff member's record for today (or `{"record": null}`).
- `POST /v1/attendance/clock-in`: authenticated staff member clocks in. Accepts optional `{"shiftId": "..."}`. Idempotency key supported.
- `POST /v1/attendance/clock-out`: authenticated staff member clocks out. Auto-closes any active break. Idempotency key supported.
- `POST /v1/attendance/breaks/start`: authenticated staff member starts break (`{"reason": "..."}`). Rejects if already on break or not clocked in.
- `POST /v1/attendance/breaks/end`: authenticated staff member ends active break. Computes break duration.

### Attendance listing and history
- `GET /v1/attendance/records?page=1&pageSize=25&staffId=...&workDate=...&fromDate=...&toDate=...&status=...&approvalStatus=...`:
  - Admin: view all records or filter by staff.
  - Staff: filter is forced to `staffId = actor.ID` (cross-staff viewing rejected).
- `GET /v1/attendance/records/{id}`: view record detail. Scoped to owner or admin.
- `GET /v1/attendance/records/{id}/history`: view complete corrections and approval history. Scoped to owner or admin.

### Administrator operations
- `POST /v1/attendance/records`: admin only; creates manual attendance record (`staffId`, `workDate`, `shiftId`, `checkInAt`, `checkOutAt`, `totalBreakMinutes`, `adminNotes`, `reason`).
- `PATCH /v1/attendance/records/{id}`: admin only; reasoned correction (`version`, `shiftId`, `checkInAt`, `checkOutAt`, `totalBreakMinutes`, `adminNotes`, `reason`). Modifying an approved record automatically logs the change, captures `before_snapshot`/`after_snapshot`, and reverts approval status to `submitted`.
- `POST /v1/attendance/records/{id}/approval`: admin only; update approval status (`status`: `submitted` / `approved` / `rejected`, `reason`).
- `GET /v1/attendance/summary?date=YYYY-MM-DD`: admin only; returns aggregate counts (`totalCount`, `presentCount`, `lateCount`, `halfDayCount`, `absentCount`, `totalWorkedMinutes`, `totalOvertimeMinutes`, `pendingApprovalCount`).

---

## 5. Security, concurrency and protections

1. **Duplicate check-in prevention**: Database unique constraint `unique_staff_work_date` and row lock on staff access serialize clock-in requests, rejecting competing submissions with 409 Conflict.
2. **Break integrity**: Only one active break allowed at a time; auto-closed upon clock-out. Cannot start a break without an active attendance session.
3. **Optimistic concurrency**: Shift updates and attendance corrections require matching `version` integer; concurrent edits reject stale submissions with 409 Conflict.
4. **Correction provenance**: Reasoned corrections never silently overwrite approved data; before/after snapshots are stored in `attendance_correction` and protected against SQL modifications.
5. **No cross-staff leakage**: Non-admin staff queries automatically bind `staffId = actor.ID`. Direct ID lookups verify record ownership.

---

## 6. Remaining work and limitations

- Frontend integration for Attendance and Manage Attendance screens is deferred to Section 4.
- Biometric hardware / RFID clock-in integration is not included in this increment.
- Payroll calculation and overtime monetary compensation rules are deliberately excluded (must be derived from documented labor regulations and financial policies in Section 3.H).

### Verified retry repair (2026-10-05)
Clock-in, clock-out, break start/end and manual creation now persist keyed responses atomically with mutations. Identical actor/operation/key/input replays return the original result; changed input conflicts. Keys are optional for compatibility, with a maximum of 200 bytes; clients should always send a stable key for retries. Migration 041 reserves 029-040 for the separate operational branch. Regression tests reproduced the original clock-in retry failure, then passed for all five operations. Full uncached Go/PostgreSQL tests and go vet passed.
