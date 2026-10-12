import test from "node:test";
import assert from "node:assert/strict";
import {
  saveBedType,
  deleteBedType,
  saveBed,
  deleteBed,
  assignBed,
  dischargeBed,
} from "./bed_management.ts";
import am from "./am.json" with { type: "json" };

test("saveBedType: new bed type uses POST and returns server identity", async () => {
  const customFetch = (async (url, init) => {
    assert.equal(url, "/api/hms/bed-types");
    assert.equal(init?.method, "POST");
    const body = JSON.parse(String(init?.body));
    assert.equal(body.name, "VIP Deluxe");
    assert.equal(body.description, "Private executive suite");
    assert.equal(body.active, true);
    return Response.json({
      id: "bt-uuid-1",
      name: body.name,
      description: body.description,
      active: true,
      version: 1,
    });
  }) as typeof fetch;

  const res = await saveBedType(
    { name: "VIP Deluxe", description: "Private executive suite" },
    customFetch,
  );
  assert.equal(res.id, "bt-uuid-1");
  assert.equal(res.version, 1);
});

test("saveBedType: editing uses PATCH with optimistic concurrency version", async () => {
  const customFetch = (async (url, init) => {
    assert.equal(url, "/api/hms/bed-types/bt-uuid-1");
    assert.equal(init?.method, "PATCH");
    const body = JSON.parse(String(init?.body));
    assert.equal(body.version, 2);
    return Response.json({
      id: "bt-uuid-1",
      name: body.name,
      description: body.description,
      active: true,
      version: 3,
    });
  }) as typeof fetch;

  const res = await saveBedType(
    {
      id: "bt-uuid-1",
      name: "VIP Deluxe Renamed",
      description: "Updated description",
      version: 2,
    },
    customFetch,
  );
  assert.equal(res.version, 3);
});

test("deleteBedType: sends DELETE with version and surfaces in-use conflicts", async () => {
  const successFetch = (async (url, init) => {
    assert.equal(url, "/api/hms/bed-types/bt-uuid-1?version=3");
    assert.equal(init?.method, "DELETE");
    return Response.json({ ok: true });
  }) as typeof fetch;

  const res = await deleteBedType("bt-uuid-1", 3, successFetch);
  assert.equal(res.ok, true);

  const conflictFetch = (async () => {
    return new Response(
      JSON.stringify({ error: "bed type is currently referenced by beds" }),
      { status: 409, headers: { "Content-Type": "application/json" } },
    );
  }) as typeof fetch;

  await assert.rejects(
    async () => deleteBedType("bt-uuid-1", 3, conflictFetch),
    /bed type is currently referenced by beds/,
  );
});

test("saveBed: new bed uses POST, sets chargeMinor and default state", async () => {
  const customFetch = (async (url, init) => {
    assert.equal(url, "/api/hms/beds");
    assert.equal(init?.method, "POST");
    const body = JSON.parse(String(init?.body));
    assert.equal(body.name, "Bed-101");
    assert.equal(body.type, "VIP Deluxe");
    assert.equal(body.chargeMinor, 75000);
    assert.equal(body.state, "available");
    return Response.json({
      id: "b-uuid-101",
      name: body.name,
      type: body.type,
      chargeMinor: body.chargeMinor,
      state: body.state,
      available: true,
      version: 1,
    });
  }) as typeof fetch;

  const res = await saveBed(
    {
      name: "Bed-101",
      type: "VIP Deluxe",
      chargeMinor: 75000,
    },
    customFetch,
  );
  assert.equal(res.id, "b-uuid-101");
  assert.equal(res.available, true);
});

test("saveBed: editing supports maintenance state and stale-version rejection", async () => {
  const customFetch = (async (url, init) => {
    assert.equal(url, "/api/hms/beds/b-uuid-101");
    assert.equal(init?.method, "PATCH");
    const body = JSON.parse(String(init?.body));
    assert.equal(body.state, "maintenance");
    assert.equal(body.version, 1);
    return Response.json({
      id: "b-uuid-101",
      name: "Bed-101",
      type: "VIP Deluxe",
      chargeMinor: 75000,
      state: "maintenance",
      available: false,
      version: 2,
    });
  }) as typeof fetch;

  const res = await saveBed(
    {
      id: "b-uuid-101",
      name: "Bed-101",
      type: "VIP Deluxe",
      chargeMinor: 75000,
      state: "maintenance",
      version: 1,
    },
    customFetch,
  );
  assert.equal(res.state, "maintenance");
  assert.equal(res.available, false);
  assert.equal(res.version, 2);

  const staleFetch = (async () => {
    return new Response(
      JSON.stringify({ error: "conflict: stale bed version" }),
      { status: 409, headers: { "Content-Type": "application/json" } },
    );
  }) as typeof fetch;

  await assert.rejects(
    async () =>
      saveBed(
        {
          id: "b-uuid-101",
          name: "Bed-101",
          type: "VIP Deluxe",
          chargeMinor: 75000,
          version: 1,
        },
        staleFetch,
      ),
    /conflict: stale bed version/,
  );
});

test("deleteBed: protects occupied beds against deletion", async () => {
  const conflictFetch = (async () => {
    return new Response(
      JSON.stringify({
        error: "cannot delete bed: bed is currently assigned or occupied",
      }),
      { status: 409, headers: { "Content-Type": "application/json" } },
    );
  }) as typeof fetch;

  await assert.rejects(
    async () => deleteBed("b-uuid-101", conflictFetch),
    /cannot delete bed: bed is currently assigned or occupied/,
  );
});

test("assignBed: validates persistent UUIDs and calls POST /bed-assignments", async () => {
  await assert.rejects(
    async () =>
      assignBed({
        patientId: "",
        encounterId: "enc-uuid-1",
        bedId: "bed-uuid-1",
      }),
    /Patient ID, Encounter ID, and Bed ID are all required/,
  );

  const customFetch = (async (url, init) => {
    assert.equal(url, "/api/hms/bed-assignments");
    assert.equal(init?.method, "POST");
    const body = JSON.parse(String(init?.body));
    assert.equal(body.patientId, "pat-uuid-1");
    assert.equal(body.encounterId, "enc-uuid-1");
    assert.equal(body.bedId, "bed-uuid-1");
    return Response.json({
      id: "assign-uuid-1",
      patientId: body.patientId,
      encounterId: body.encounterId,
      bedId: body.bedId,
      isActive: true,
    });
  }) as typeof fetch;

  const res = await assignBed(
    {
      patientId: "pat-uuid-1",
      encounterId: "enc-uuid-1",
      bedId: "bed-uuid-1",
      notes: "Post-op admission",
    },
    customFetch,
  );
  assert.equal(res.id, "assign-uuid-1");
  assert.equal(res.isActive, true);
});

test("dischargeBed: calls discharge endpoint and releases bed", async () => {
  const customFetch = (async (url, init) => {
    assert.equal(url, "/api/hms/bed-assignments/assign-uuid-1/discharge");
    assert.equal(init?.method, "POST");
    const body = JSON.parse(String(init?.body));
    assert.equal(body.dischargeNotes, "Discharged home in stable condition");
    return Response.json({
      id: "assign-uuid-1",
      isActive: false,
      assignedTo: new Date().toISOString(),
    });
  }) as typeof fetch;

  const res = await dischargeBed(
    "assign-uuid-1",
    "Discharged home in stable condition",
    customFetch,
  );
  assert.equal(res.isActive, false);
  assert.ok(res.assignedTo);
});

test("Amharic localization dictionary covers all bed management workspace keys", () => {
  const requiredKeys = [
    "Bed Management",
    "Bed Status",
    "Bed Assigns",
    "Beds",
    "Bed Types",
    "New Bed",
    "Edit Bed",
    "New Bed Type",
    "Edit Bed Type",
    "New Bed Assign",
    "Bed Name",
    "Bed Type",
    "No Bed Available",
    "Available",
    "Occupied",
    "IPD NO",
    "PATIENT",
    "BED",
    "ASSIGN DATE",
    "DISCHARGE DATE",
    "STATUS",
    "ACTION",
    "BED ID",
    "CHARGE",
    "AVAILABLE",
    "DESCRIPTION",
  ];

  for (const key of requiredKeys) {
    assert.ok(
      (am as Record<string, string>)[key],
      `Missing Amharic translation for key "${key}"`,
    );
  }
});
