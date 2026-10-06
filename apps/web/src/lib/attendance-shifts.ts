export type AttendanceRow = { id: string; [key: string]: string };

export function shiftRow(s: Record<string, any>): AttendanceRow {
  return {
    id: s.id,
    name: s.name,
    code: s.code,
    start: s.startTime.slice(0, 5),
    end: s.endTime.slice(0, 5),
    grace: String(s.gracePeriodMinutes),
    break: String(s.breakDurationMinutes),
    halfDay: String(s.halfDayMinutes),
    fullDay: String(s.fullDayMinutes),
    version: String(s.version),
    default: s.isDefault ? "Is Default" : "Not Default",
    status: s.active ? "Active" : "Inactive",
    staffCount: String(s.staffCount ?? 0),
  };
}

export async function saveAttendanceShift(
  row: AttendanceRow,
  request: typeof fetch = fetch,
) {
  const editing = Boolean(row.version);
  const response = await request(
    `/api/hms/attendance/shifts${editing ? `/${encodeURIComponent(row.id)}` : ""}`,
    {
      method: editing ? "PATCH" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: row.name,
        code: row.code,
        isDefault: row.default === "Is Default",
        startTime: row.start,
        endTime: row.end,
        gracePeriodMinutes: Number(row.grace || 0),
        breakDurationMinutes: Number(row.break || 0),
        halfDayMinutes: Number(row.halfDay),
        fullDayMinutes: Number(row.fullDay),
        active: row.status !== "Inactive",
        ...(editing ? { version: Number(row.version) } : {}),
      }),
    },
  );
  if (!response.ok) {
    throw new Error(
      response.status === 409
        ? "This shift changed. Reload before editing again."
        : `Shift was not saved (${response.status}). Check the fields and retry.`,
    );
  }
  return shiftRow(await response.json());
}
