import test from "node:test";
import assert from "node:assert/strict";
import { saveAttendanceShift } from "./attendance-shifts.ts";

const row = {
  id: "local-draft",
  name: "Day",
  start: "08:00",
  end: "17:00",
  grace: "0",
  break: "0",
  halfDay: "240",
  fullDay: "480",
  status: "Active",
};
const saved = {
  id: "server-id",
  name: "Day",
  startTime: "08:00:00",
  endTime: "17:00:00",
  gracePeriodMinutes: 0,
  breakDurationMinutes: 0,
  halfDayMinutes: 240,
  fullDayMinutes: 480,
  active: true,
  version: 1,
};
test("new shift preserves zero values and adopts server identity", async () => {
  const result = await saveAttendanceShift(row, (async (url, options) => {
    assert.equal(url, "/api/hms/attendance/shifts");
    assert.equal(options?.method, "POST");
    const body = JSON.parse(String(options?.body));
    assert.equal(body.halfDayMinutes, 240);
    assert.equal(body.fullDayMinutes, 480);
    assert.equal(body.gracePeriodMinutes, 0);
    assert.equal(body.breakDurationMinutes, 0);
    return Response.json(saved);
  }) as typeof fetch);
  assert.equal(result.id, "server-id");
  assert.equal(result.version, "1");
});
test("editing uses PATCH with the concurrency version", async () => {
  await saveAttendanceShift({ ...row, id: "server-id", version: "3" }, (async (
    url,
    options,
  ) => {
    assert.equal(url, "/api/hms/attendance/shifts/server-id");
    assert.equal(options?.method, "PATCH");
    assert.equal(JSON.parse(String(options?.body)).version, 3);
    return Response.json({ ...saved, version: 4 });
  }) as typeof fetch);
});
test("validation, authorization, conflicts and network failures never report a saved row", async () => {
  for (const status of [400, 401, 403, 409, 500]) {
    await assert.rejects(
      saveAttendanceShift(
        row,
        (async () => new Response(null, { status })) as typeof fetch,
      ),
    );
  }
  await assert.rejects(
    saveAttendanceShift(row, (async () => {
      throw new Error("offline");
    }) as typeof fetch),
  );
});
