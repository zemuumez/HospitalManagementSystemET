export type AttendanceRow = { id: string; [key: string]: string };

export type StaffOption = {
  id: string;
  name: string;
  email: string;
  role: string;
};

export function assignmentRow(a: Record<string, any>): AttendanceRow {
  return {
    id: String(a.id),
    staffId: String(a.staffId),
    staff: a.staffName || a.staffId || "Staff Member",
    shiftId: String(a.shiftId),
    shift: a.shiftName || "Assigned Shift",
    from: a.effectiveFrom || "",
    to: a.effectiveTo || "",
    status: a.active === false ? "Inactive" : "Active",
    note: a.note || "",
    createdAt: a.createdAt || "",
  };
}

export async function saveShiftAssignment(
  payload: {
    staffId: string;
    shiftId: string;
    effectiveFrom: string;
    effectiveTo?: string;
    note?: string;
    active?: boolean;
  },
  request: typeof fetch = fetch,
): Promise<AttendanceRow> {
  const response = await request("/api/hms/attendance/assignments", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      staffId: payload.staffId,
      shiftId: payload.shiftId,
      effectiveFrom: payload.effectiveFrom,
      ...(payload.effectiveTo ? { effectiveTo: payload.effectiveTo } : {}),
      note: payload.note || "",
      active: payload.active !== false,
    }),
  });
  if (!response.ok) {
    throw new Error(
      `Duty assignment was not saved (${response.status}). Check fields and retry.`,
    );
  }
  return assignmentRow(await response.json());
}

export function leaveRow(l: Record<string, any>): AttendanceRow {
  const statusFormatted = l.status
    ? l.status.charAt(0).toUpperCase() + l.status.slice(1).toLowerCase()
    : "Pending";
  const typeFormatted = l.leaveType
    ? l.leaveType.charAt(0).toUpperCase() + l.leaveType.slice(1).toLowerCase()
    : "Casual";
  return {
    id: String(l.id),
    staffId: String(l.staffId),
    staff: l.staffName || l.staffId || "Staff Member",
    from: l.fromDate || "",
    to: l.toDate || "",
    days: String(l.days ?? 1),
    type: typeFormatted,
    leaveType: l.leaveType || "casual",
    status: statusFormatted,
    reason: l.reason || "",
    approver: l.approverName || (l.approverId ? "Admin" : ""),
    approverId: l.approverId || "",
    approverNotes: l.approverNotes || "",
    version: String(l.version ?? 1),
    actionedAt: l.actionedAt || "",
    createdAt: l.createdAt || "",
  };
}

export async function saveLeaveRequest(
  payload: {
    staffId?: string;
    fromDate: string;
    toDate: string;
    leaveType: string;
    reason: string;
  },
  request: typeof fetch = fetch,
): Promise<AttendanceRow> {
  const response = await request("/api/hms/attendance/leaves", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      ...(payload.staffId ? { staffId: payload.staffId } : {}),
      fromDate: payload.fromDate,
      toDate: payload.toDate,
      leaveType: payload.leaveType.toLowerCase(),
      reason: payload.reason,
    }),
  });
  if (!response.ok) {
    throw new Error(
      `Leave request was not submitted (${response.status}). Check the dates and retry.`,
    );
  }
  return leaveRow(await response.json());
}

export async function updateLeaveStatus(
  id: string,
  status: "approved" | "rejected" | "cancelled",
  version: number,
  approverNotes: string = "",
  request: typeof fetch = fetch,
): Promise<AttendanceRow> {
  const response = await request(
    `/api/hms/attendance/leaves/${encodeURIComponent(id)}/status`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        status,
        approverNotes,
        version,
      }),
    },
  );
  if (!response.ok) {
    throw new Error(
      response.status === 409
        ? "This leave request was updated by another administrator. Reload and retry."
        : `Could not update leave status (${response.status}).`,
    );
  }
  return leaveRow(await response.json());
}

export function recordRow(r: Record<string, any>): AttendanceRow {
  const checkIn = r.checkInAt
    ? new Date(r.checkInAt).toLocaleTimeString("en-GB", {
        hour: "2-digit",
        minute: "2-digit",
      })
    : "";
  const checkOut = r.checkOutAt
    ? new Date(r.checkOutAt).toLocaleTimeString("en-GB", {
        hour: "2-digit",
        minute: "2-digit",
      })
    : "";
  const statusFormatted = r.status
    ? r.status
        .replace("_", " ")
        .replace(/\b\w/g, (c: string) => c.toUpperCase())
    : "Present";
  const approvalFormatted = r.approvalStatus
    ? r.approvalStatus.charAt(0).toUpperCase() +
      r.approvalStatus.slice(1).toLowerCase()
    : "Approved";
  return {
    id: String(r.id),
    staffId: String(r.staffId),
    staff: r.staffName || r.staffId || "Staff Member",
    date: r.workDate || "",
    shiftId: String(r.shiftId || ""),
    shift: r.shiftName || "Assigned Shift",
    status: statusFormatted,
    approvalStatus: approvalFormatted,
    checkIn,
    checkOut,
    late: r.lateMinutes
      ? `${Math.floor(r.lateMinutes / 60)}h ${r.lateMinutes % 60}m`
      : "0m",
    early: r.earlyOutMinutes
      ? `${Math.floor(r.earlyOutMinutes / 60)}h ${r.earlyOutMinutes % 60}m`
      : "0m",
    worked: r.workedMinutes
      ? `${Math.floor(r.workedMinutes / 60)}h ${r.workedMinutes % 60}m`
      : "0m",
    overtime: r.overtimeMinutes
      ? `${Math.floor(r.overtimeMinutes / 60)}h ${r.overtimeMinutes % 60}m`
      : "0m",
    break: String(r.totalBreakMinutes ?? 0),
    breakTime: r.totalBreakMinutes
      ? `${Math.floor(r.totalBreakMinutes / 60)}h ${r.totalBreakMinutes % 60}m`
      : "0m",
    remarks: r.adminNotes || "",
    reason: r.adminNotes || "Attendance verification",
    version: String(r.version ?? 1),
    checkInAt: r.checkInAt || "",
    checkOutAt: r.checkOutAt || "",
  };
}

export async function updateRecordApproval(
  id: string,
  status: "approved" | "rejected",
  reason: string = "Reviewed by administrator",
  request: typeof fetch = fetch,
): Promise<AttendanceRow> {
  const response = await request(
    `/api/hms/attendance/records/${encodeURIComponent(id)}/approval`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        status,
        reason,
      }),
    },
  );
  if (!response.ok) {
    throw new Error(
      `Failed to update attendance request approval (${response.status}).`,
    );
  }
  return recordRow(await response.json());
}

export async function clockIn(
  shiftId?: string,
  request: typeof fetch = fetch,
): Promise<AttendanceRow> {
  const response = await request("/api/hms/attendance/clock-in", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(shiftId ? { shiftId } : {}),
  });
  if (!response.ok) {
    throw new Error(`Clock in failed (${response.status}).`);
  }
  return recordRow(await response.json());
}

export async function clockOut(
  request: typeof fetch = fetch,
): Promise<AttendanceRow> {
  const response = await request("/api/hms/attendance/clock-out", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({}),
  });
  if (!response.ok) {
    throw new Error(`Clock out failed (${response.status}).`);
  }
  return recordRow(await response.json());
}

export async function startBreak(
  reason: string = "Scheduled break",
  request: typeof fetch = fetch,
): Promise<Record<string, any>> {
  const response = await request("/api/hms/attendance/breaks/start", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ reason }),
  });
  if (!response.ok) {
    throw new Error(`Start break failed (${response.status}).`);
  }
  return await response.json();
}

export async function endBreak(
  request: typeof fetch = fetch,
): Promise<Record<string, any>> {
  const response = await request("/api/hms/attendance/breaks/end", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({}),
  });
  if (!response.ok) {
    throw new Error(`End break failed (${response.status}).`);
  }
  return await response.json();
}
