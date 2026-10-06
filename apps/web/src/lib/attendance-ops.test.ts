import test from "node:test";
import assert from "node:assert/strict";
import {
  assignmentRow,
  saveShiftAssignment,
  leaveRow,
  saveLeaveRequest,
  updateLeaveStatus,
  recordRow,
  updateRecordApproval,
} from "./attendance-ops.ts";

test("assignmentRow maps server assignment correctly", () => {
  const row = assignmentRow({
    id: "assign-1",
    staffId: "staff-101",
    staffName: "Dr. Abera",
    shiftId: "shift-202",
    shiftName: "Day Shift",
    effectiveFrom: "2026-10-01",
    effectiveTo: "2026-10-31",
    note: "Emergency duty",
    active: true,
  });
  assert.equal(row.id, "assign-1");
  assert.equal(row.staff, "Dr. Abera");
  assert.equal(row.shift, "Day Shift");
  assert.equal(row.status, "Active");
  assert.equal(row.note, "Emergency duty");
});

test("saveShiftAssignment sends POST payload to /api/hms/attendance/assignments", async () => {
  const payload = {
    staffId: "user-1",
    shiftId: "shift-1",
    effectiveFrom: "2026-10-01",
    note: "Primary cover",
    active: true,
  };
  const result = await saveShiftAssignment(payload, (async (url, options) => {
    assert.equal(url, "/api/hms/attendance/assignments");
    assert.equal(options?.method, "POST");
    const body = JSON.parse(String(options?.body));
    assert.equal(body.staffId, "user-1");
    assert.equal(body.shiftId, "shift-1");
    assert.equal(body.note, "Primary cover");
    assert.equal(body.active, true);
    return Response.json({
      id: "server-assign-id",
      staffId: body.staffId,
      staffName: "Nurse Selam",
      shiftId: body.shiftId,
      shiftName: "Morning Shift",
      effectiveFrom: body.effectiveFrom,
      note: body.note,
      active: true,
    });
  }) as typeof fetch);
  assert.equal(result.id, "server-assign-id");
  assert.equal(result.staff, "Nurse Selam");
  assert.equal(result.status, "Active");
});

test("leaveRow maps server leave and formats status/type", () => {
  const row = leaveRow({
    id: "leave-1",
    staffId: "staff-1",
    staffName: "Dr. Kebede",
    fromDate: "2026-10-15",
    toDate: "2026-10-18",
    days: 4,
    leaveType: "annual",
    status: "approved",
    approverName: "Medical Director",
    version: 2,
  });
  assert.equal(row.id, "leave-1");
  assert.equal(row.type, "Annual");
  assert.equal(row.status, "Approved");
  assert.equal(row.days, "4");
  assert.equal(row.approver, "Medical Director");
  assert.equal(row.version, "2");
});

test("saveLeaveRequest sends correct payload", async () => {
  const payload = {
    fromDate: "2026-11-01",
    toDate: "2026-11-03",
    leaveType: "Sick",
    reason: "Medical care",
  };
  const result = await saveLeaveRequest(payload, (async (url, options) => {
    assert.equal(url, "/api/hms/attendance/leaves");
    assert.equal(options?.method, "POST");
    const body = JSON.parse(String(options?.body));
    assert.equal(body.leaveType, "sick");
    assert.equal(body.reason, "Medical care");
    return Response.json({
      id: "leave-new-id",
      staffId: "user-current",
      staffName: "Current User",
      fromDate: body.fromDate,
      toDate: body.toDate,
      days: 3,
      leaveType: "sick",
      reason: body.reason,
      status: "pending",
      version: 1,
    });
  }) as typeof fetch);
  assert.equal(result.id, "leave-new-id");
  assert.equal(result.status, "Pending");
  assert.equal(result.days, "3");
});

test("updateLeaveStatus sends approval payload with version", async () => {
  const result = await updateLeaveStatus(
    "leave-123",
    "approved",
    1,
    "Approved",
    (async (url, options) => {
      assert.equal(url, "/api/hms/attendance/leaves/leave-123/status");
      assert.equal(options?.method, "POST");
      const body = JSON.parse(String(options?.body));
      assert.equal(body.status, "approved");
      assert.equal(body.version, 1);
      assert.equal(body.approverNotes, "Approved");
      return Response.json({
        id: "leave-123",
        staffId: "user-1",
        staffName: "Staff One",
        fromDate: "2026-10-10",
        toDate: "2026-10-12",
        days: 3,
        leaveType: "casual",
        status: "approved",
        approverId: "admin-1",
        approverName: "Admin",
        approverNotes: "Approved",
        version: 2,
      });
    }) as typeof fetch,
  );
  assert.equal(result.status, "Approved");
  assert.equal(result.version, "2");
});

test("recordRow maps attendance record and calculates readable times", () => {
  const row = recordRow({
    id: "rec-1",
    staffId: "user-1",
    staffName: "Dr. Aster",
    workDate: "2026-10-06",
    shiftName: "Day Shift",
    status: "present",
    approvalStatus: "submitted",
    lateMinutes: 15,
    earlyOutMinutes: 0,
    workedMinutes: 480,
    overtimeMinutes: 30,
    totalBreakMinutes: 45,
  });
  assert.equal(row.id, "rec-1");
  assert.equal(row.status, "Present");
  assert.equal(row.approvalStatus, "Submitted");
  assert.equal(row.late, "0h 15m");
  assert.equal(row.worked, "8h 0m");
  assert.equal(row.overtime, "0h 30m");
  assert.equal(row.breakTime, "0h 45m");
});

test("updateRecordApproval calls approval endpoint", async () => {
  const result = await updateRecordApproval(
    "rec-10",
    "approved",
    "Good to go",
    (async (url, options) => {
      assert.equal(url, "/api/hms/attendance/records/rec-10/approval");
      assert.equal(options?.method, "POST");
      const body = JSON.parse(String(options?.body));
      assert.equal(body.status, "approved");
      assert.equal(body.reason, "Good to go");
      return Response.json({
        id: "rec-10",
        staffId: "user-1",
        staffName: "Staff User",
        workDate: "2026-10-06",
        status: "present",
        approvalStatus: "approved",
      });
    }) as typeof fetch,
  );
  assert.equal(result.approvalStatus, "Approved");
});
