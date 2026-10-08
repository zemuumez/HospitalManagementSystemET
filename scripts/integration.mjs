import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { Pool } from "pg";
import { hashPassword } from "better-auth/crypto";

// The wrapper owns isolated services and drops the entire generated schema.
const schema = process.env.HMS_TEST_ISOLATED_SCHEMA;
assert.match(schema || "", /^hms_browser_[a-f0-9]{24}$/);
const base = process.env.BETTER_AUTH_URL;
if (!base || !new URL(base).hostname.match(/^(127\.0\.0\.1|localhost)$/))
  throw new Error("Integration checks require a loopback web server");
const db = new Pool({ connectionString: process.env.DATABASE_URL });
const ids = [];
const patients = [];
async function fixture(role) {
  const id = randomUUID();
  const password = randomUUID() + "Aa1!";
  const email = `test-${id}@example.test`;
  await db.query(
    'INSERT INTO "user"(id,name,email,"emailVerified") VALUES($1,$2,$3,true)',
    [id, `Integration ${role}`, email],
  );
  ids.push(id);
  await db.query(
    'INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,\'credential\',$2,$3)',
    [randomUUID(), id, await hashPassword(password)],
  );
  await db.query("INSERT INTO staff_access(user_id,role) VALUES($1,$2)", [
    id,
    role,
  ]);
  const response = await request("/api/auth/sign-in/email", {
    method: "POST",
    body: { email, password },
  });
  assert.equal(response.status, 200, await response.text());
  return {
    id,
    email,
    cookie: response.headers
      .getSetCookie()
      .map((c) => c.split(";")[0])
      .join("; "),
  };
}
function request(
  path,
  { cookie = "", method = "GET", body, origin = base, key } = {},
) {
  return fetch(base + path, {
    method,
    redirect: "manual",
    headers: {
      cookie,
      origin,
      "content-type": "application/json",
      ...(key ? { "Idempotency-Key": key } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
    signal: AbortSignal.timeout(20000),
  });
}
async function expectStatus(path, status, options) {
  const r = await request(path, options);
  assert.equal(r.status, status, `${path}: ${await r.text()}`);
}
try {
  assert.equal(
    (await db.query("SELECT current_schema() AS name")).rows[0].name,
    schema,
  );
  await expectStatus("/api/hms/patients", 401);
  const admin = await fixture("admin");
  const patient = await fixture("patient");
  const doctor = await fixture("doctor");
  await expectStatus("/api/hms/patients", 403, {
    method: "POST",
    cookie: admin.cookie,
    origin: "https://untrusted.example",
    body: {},
  });
  await expectStatus("/api/hms/patients", 403, {
    method: "POST",
    cookie: patient.cookie,
    body: {
      givenName: "Blocked",
      familyName: "Test",
      dateOfBirth: "2000-01-01",
      phone: "",
    },
  });
  await expectStatus("/api/hms/patients", 422, {
    method: "POST",
    cookie: admin.cookie,
    body: {
      givenName: "Invalid",
      familyName: "Test",
      dateOfBirth: "2099-01-01",
      phone: "",
    },
  });
  for (let n = 0; n < 2; n++) {
    const r = await request("/api/hms/patients", {
      cookie: admin.cookie,
      method: "POST",
      body: {
        givenName: "Synthetic",
        familyName: `Integration${n}`,
        dateOfBirth: "2000-01-01",
        phone: "",
      },
    });
    assert.equal(r.status, 201);
    const p = await r.json();
    patients.push(p.id);
  }
  await expectStatus(`/api/hms/patients/${patients[0]}`, 403, {
    cookie: patient.cookie,
    method: "PATCH",
    body: { userId: patient.id, clinicianId: doctor.id },
  });
  await expectStatus(`/api/hms/patients/${patients[0]}`, 200, {
    cookie: admin.cookie,
    method: "PATCH",
    body: { userId: patient.id, clinicianId: doctor.id },
  });
  await expectStatus(`/api/hms/patients/${patients[1]}`, 409, {
    cookie: admin.cookie,
    method: "PATCH",
    body: { userId: patient.id, clinicianId: doctor.id },
  });
  await expectStatus(`/api/hms/patients/${patients[0]}`, 409, {
    cookie: admin.cookie,
    method: "PATCH",
    body: { userId: "", clinicianId: doctor.id },
  });
  for (const actor of [patient, doctor]) {
    const result = await (
      await request("/api/hms/patients?search=Synthetic", {
        cookie: actor.cookie,
      })
    ).json();
    assert.equal(result.patients.length, 1);
    assert.equal(result.patients[0].id, patients[0]);
    const stats = await (
      await request("/api/hms/overview", { cookie: actor.cookie })
    ).json();
    assert.equal(stats.patientCount, 1);
  }

  // Scheduling contracts: source-derived hours, scoped records, races and retries.
  await expectStatus("/api/staff", 403, { cookie: patient.cookie });
  const staffResult = await request("/api/staff", {
    cookie: admin.cookie,
    method: "POST",
    body: {
      name: "Integration nurse",
      email: `nurse-${randomUUID()}@example.test`,
      role: "nurse",
      password: randomUUID() + "Aa1!",
    },
  });
  assert.equal(staffResult.status, 201, await staffResult.clone().text());
  const staffID = (await staffResult.json()).id;
  ids.push(staffID);
  await expectStatus("/api/staff", 409, {
    cookie: admin.cookie,
    method: "PATCH",
    body: { id: admin.id, active: false },
  });
  await expectStatus("/api/staff", 200, {
    cookie: admin.cookie,
    method: "PATCH",
    body: { id: staffID, active: false },
  });

  // D4: Doctor provisioning coordination, validation, rollback and editing
  const archivedDept = (
    await db.query(
      "INSERT INTO doctor_department (title, description, archived) VALUES ('Archived ENT', 'Old wing', true) RETURNING id",
    )
  ).rows[0].id;
  const activeDept = (
    await db.query(
      "INSERT INTO doctor_department (title, description, archived) VALUES ('Cardiology Clinic', 'Heart wing', false) RETURNING id",
    )
  ).rows[0].id;

  // 1. Missing required fields for doctor
  await expectStatus("/api/staff", 422, {
    cookie: admin.cookie,
    method: "POST",
    body: {
      name: "Dr. Incomplete",
      email: `doc-incomp-${randomUUID()}@example.test`,
      role: "doctor",
      password: randomUUID() + "Aa1!",
    },
  });

  // 2. Archived department failure cleanup: ensure complete rollback (no orphan user or staff_access)
  const failedDocEmail = `doc-fail-${randomUUID()}@example.test`;
  await expectStatus("/api/staff", 422, {
    cookie: admin.cookie,
    method: "POST",
    body: {
      name: "Dr. Failed Candidate",
      email: failedDocEmail,
      role: "doctor",
      password: randomUUID() + "Aa1!",
      departmentId: archivedDept,
      specialist: "Cardiologist",
      designation: "Consultant",
      qualification: "MBBS",
      gender: "male",
    },
  });
  assert.equal(
    (
      await db.query(
        'SELECT count(*)::int AS count FROM "user" WHERE email=$1',
        [failedDocEmail],
      )
    ).rows[0].count,
    0,
  );
  assert.equal(
    (
      await db.query(
        'SELECT count(*)::int AS count FROM account WHERE "userId" IN (SELECT id FROM "user" WHERE email=$1)',
        [failedDocEmail],
      )
    ).rows[0].count,
    0,
  );
  assert.equal(
    (
      await db.query(
        'SELECT count(*)::int AS count FROM staff_access WHERE user_id IN (SELECT id FROM "user" WHERE email=$1)',
        [failedDocEmail],
      )
    ).rows[0].count,
    0,
  );
  assert.equal(
    (
      await db.query(
        'SELECT count(*)::int AS count FROM doctor_profile WHERE user_id IN (SELECT id FROM "user" WHERE email=$1)',
        [failedDocEmail],
      )
    ).rows[0].count,
    0,
  );
  assert.equal(
    (
      await db.query(
        'SELECT count(*)::int AS count FROM staff_profile WHERE user_id IN (SELECT id FROM "user" WHERE email=$1)',
        [failedDocEmail],
      )
    ).rows[0].count,
    0,
  );
  assert.equal(
    (
      await db.query(
        'SELECT count(*)::int AS count FROM doctor_hours WHERE doctor_id IN (SELECT id FROM "user" WHERE email=$1)',
        [failedDocEmail],
      )
    ).rows[0].count,
    0,
  );
  assert.equal(
    (
      await db.query(
        "SELECT count(*)::int AS count FROM audit_event WHERE action='doctor.created' AND resource_id IN (SELECT id FROM \"user\" WHERE email=$1)",
        [failedDocEmail],
      )
    ).rows[0].count,
    0,
  );

  // 3. Retry with valid department succeeds and provisions complete profile + 7-day schedule
  const retryResult = await request("/api/staff", {
    cookie: admin.cookie,
    method: "POST",
    body: {
      name: "Dr. Coordinated Parity",
      email: failedDocEmail,
      role: "doctor",
      password: randomUUID() + "Aa1!",
      departmentId: activeDept,
      specialist: "Cardiologist",
      designation: "Senior Consultant",
      qualification: "MBBS, MD",
      gender: "male",
      dateOfBirth: "1985-05-15",
      bloodGroup: "O+",
      phone: "+251911223344",
      address1: "Bole Road",
      city: "Addis Ababa",
      zip: "1000",
      appointmentCharge: 500,
      opdCharge: 300,
    },
  });
  assert.equal(retryResult.status, 201, await retryResult.clone().text());
  const coordinatedDoctorID = (await retryResult.json()).id;
  ids.push(coordinatedDoctorID);

  // Verify DB linkage and default 7-day 10:00-19:30 schedule
  const profileRow = (
    await db.query(
      "SELECT specialist, designation, qualification, slot_minutes FROM doctor_profile WHERE user_id=$1",
      [coordinatedDoctorID],
    )
  ).rows[0];
  assert.equal(profileRow.specialist, "Cardiologist");
  assert.equal(profileRow.designation, "Senior Consultant");
  assert.equal(profileRow.qualification, "MBBS, MD");
  assert.equal(profileRow.slot_minutes, 60);

  const hoursRows = (
    await db.query(
      "SELECT weekday, start_minute, end_minute FROM doctor_hours WHERE doctor_id=$1 ORDER BY weekday",
      [coordinatedDoctorID],
    )
  ).rows;
  assert.equal(hoursRows.length, 7);
  for (let w = 0; w < 7; w++) {
    assert.equal(hoursRows[w].weekday, w);
    assert.equal(hoursRows[w].start_minute, 600);
    assert.equal(hoursRows[w].end_minute, 1170);
  }

  // 4. Duplicate email rejection
  await expectStatus("/api/staff", 409, {
    cookie: admin.cookie,
    method: "POST",
    body: {
      name: "Dr. Another Candidate",
      email: failedDocEmail,
      role: "doctor",
      password: randomUUID() + "Aa1!",
      departmentId: activeDept,
      specialist: "Neurologist",
      designation: "Consultant",
      qualification: "MBBS",
      gender: "female",
    },
  });

  // 5. Name and Email editing on PATCH /api/staff
  const updatedDocEmail = `doc-renamed-${randomUUID()}@example.test`;
  await expectStatus("/api/staff", 200, {
    cookie: admin.cookie,
    method: "PATCH",
    body: {
      id: coordinatedDoctorID,
      name: "Dr. Coordinated Renamed",
      email: updatedDocEmail,
    },
  });
  const updatedUserRow = (
    await db.query('SELECT name, email FROM "user" WHERE id=$1', [
      coordinatedDoctorID,
    ])
  ).rows[0];
  assert.equal(updatedUserRow.name, "Dr. Coordinated Renamed");
  assert.equal(updatedUserRow.email, updatedDocEmail);

  // Conflict on email edit
  await expectStatus("/api/staff", 409, {
    cookie: admin.cookie,
    method: "PATCH",
    body: {
      id: coordinatedDoctorID,
      email: admin.email,
    },
  });

  // 6. Strict email validation on PATCH /api/staff: malformed emails rejected with 422
  for (const badEmail of [
    "not-an-email",
    "missingdomain@",
    "@missinglocal.test",
    "missingdot@domain",
    "two@@domain.com",
    "has spaces@domain.com",
  ]) {
    await expectStatus("/api/staff", 422, {
      cookie: admin.cookie,
      method: "PATCH",
      body: {
        id: coordinatedDoctorID,
        email: badEmail,
      },
    });
  }

  // Verify email in database remains unchanged
  const unchangedUserRow = (
    await db.query('SELECT email FROM "user" WHERE id=$1', [
      coordinatedDoctorID,
    ])
  ).rows[0];
  assert.equal(unchangedUserRow.email, updatedDocEmail);
  const schedule = {
    id: doctor.id,
    name: "",
    department: "General medicine",
    slotMinutes: 30,
    version: 0,
    hours: Array.from({ length: 7 }, (_, weekday) => ({
      weekday,
      startMinute: 540,
      endMinute: 1020,
    })),
  };
  await expectStatus("/api/hms/doctors", 403, {
    cookie: patient.cookie,
    method: "POST",
    body: schedule,
  });
  await expectStatus("/api/hms/doctors", 200, {
    cookie: admin.cookie,
    method: "POST",
    body: schedule,
  });
  const day = new Date(Date.now() + 3 * 86400000).toISOString().slice(0, 10);
  const slotResult = await request(
    `/api/hms/slots?doctorId=${doctor.id}&date=${day}`,
    { cookie: admin.cookie },
  );
  assert.equal(slotResult.status, 200, await slotResult.clone().text());
  const slots = (await slotResult.json()).slots;
  assert.equal(slots.length, 16);
  const booking = {
    patientId: patients[0],
    doctorId: doctor.id,
    startsAt: slots[0],
    problem: "Synthetic appointment",
    notifySms: false,
  };
  await expectStatus("/api/hms/appointments", 404, {
    cookie: patient.cookie,
    method: "POST",
    key: randomUUID(),
    body: { ...booking, patientId: patients[1] },
  });
  const bookingKey = randomUUID();
  const booked = await Promise.all(
    [1, 2].map(() =>
      request("/api/hms/appointments", {
        cookie: admin.cookie,
        method: "POST",
        key: bookingKey,
        body: booking,
      }),
    ),
  );
  for (const r of booked) assert.equal(r.status, 201, await r.clone().text());
  const [first, duplicate] = await Promise.all(booked.map((r) => r.json()));
  assert.equal(first.id, duplicate.id);
  await expectStatus("/api/hms/appointments", 409, {
    cookie: admin.cookie,
    method: "POST",
    key: bookingKey,
    body: { ...booking, problem: "changed" },
  });
  await expectStatus("/api/hms/appointments", 409, {
    cookie: admin.cookie,
    method: "POST",
    key: randomUUID(),
    body: { ...booking, patientId: patients[1] },
  });
  const race = await Promise.all(
    [patients[0], patients[1]].map((patientId) =>
      request("/api/hms/appointments", {
        cookie: admin.cookie,
        method: "POST",
        key: randomUUID(),
        body: { ...booking, patientId, startsAt: slots[1] },
      }),
    ),
  );
  assert.deepEqual(race.map((r) => r.status).sort(), [201, 409]);
  await expectStatus("/api/hms/doctors", 409, {
    cookie: admin.cookie,
    method: "POST",
    body: { ...schedule, version: 1 },
  });
  const owned = await (
    await request("/api/hms/appointments", { cookie: patient.cookie })
  ).json();
  assert.ok(owned.appointments.every((a) => a.patientId === patients[0]));
  assert.ok(owned.appointments.some((a) => a.id === first.id));
  await expectStatus(`/api/hms/appointments/${first.id}`, 422, {
    cookie: admin.cookie,
    method: "PATCH",
    body: { status: "completed", version: 1 },
  });
  await expectStatus(`/api/hms/appointments/${first.id}`, 200, {
    cookie: patient.cookie,
    method: "PATCH",
    body: { status: "cancelled", version: 1 },
  });
  await expectStatus(`/api/hms/appointments/${first.id}`, 409, {
    cookie: admin.cookie,
    method: "PATCH",
    body: { status: "arrived", version: 1 },
  });
  const refreshed = await (
    await request(`/api/hms/slots?doctorId=${doctor.id}&date=${day}`, {
      cookie: admin.cookie,
    })
  ).json();
  assert.ok(refreshed.slots.includes(slots[0]));
  assert.ok(!refreshed.slots.includes(slots[1]));
  console.log(
    "PASS: admin-only staff provisioning and disablement; scheduling validation, ownership, concurrent booking, idempotency, schedule conflict, lifecycle and released slots.",
  );
  await db.query("UPDATE staff_access SET active=false WHERE user_id=$1", [
    patient.id,
  ]);
  await expectStatus("/api/hms/patients", 401, { cookie: patient.cookie });
  await db.query("UPDATE staff_access SET active=true WHERE user_id=$1", [
    patient.id,
  ]);
  await expectStatus("/api/auth/firebase/sign-in", 401, {
    method: "POST",
    body: { idToken: "this-is-not-a-valid-firebase-token" },
  });
  const key = randomUUID();
  const body = {
    channel: "sms",
    recipient: "+254700000099",
    subject: "Integration",
    body: "Synthetic development message. No live SMS.",
  };
  const responses = await Promise.all(
    [1, 2].map(() =>
      request("/api/hms/messages", {
        cookie: admin.cookie,
        method: "POST",
        key,
        body,
      }),
    ),
  );
  for (const r of responses) assert.equal(r.status, 202);
  const results = await Promise.all(responses.map((r) => r.json()));
  assert.equal(results[0].id, results[1].id);
  await expectStatus("/api/hms/messages", 409, {
    cookie: admin.cookie,
    method: "POST",
    key,
    body: { ...body, body: "Different request with reused key" },
  });
  const mail = await request("/api/hms/messages", {
    cookie: admin.cookie,
    method: "POST",
    key: randomUUID(),
    body: {
      channel: "email",
      recipient: "integration@example.test",
      subject: "ULSHMS integration check",
      body: "Development transport verification.",
    },
  });
  assert.equal(mail.status, 202);
  const mailResult = await mail.json();
  await expectStatus("/api/auth/request-password-reset", 200, {
    method: "POST",
    body: { email: patient.email, redirectTo: "/reset-password" },
  });
  let statuses;
  for (let attempt = 0; attempt < 15; attempt++) {
    statuses = (
      await db.query(
        "SELECT id,status FROM message_outbox WHERE id=ANY($1::uuid[])",
        [[results[0].id, mailResult.id]],
      )
    ).rows;
    if (statuses.every((r) => ["captured", "sent"].includes(r.status))) break;
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }
  assert.equal(statuses.find((r) => r.id === results[0].id).status, "captured");
  assert.equal(statuses.find((r) => r.id === mailResult.id).status, "sent");
  const mailbox = await (
    await fetch("http://127.0.0.1:8025/api/v1/messages")
  ).json();
  assert.ok(
    mailbox.messages.some((m) => m.Subject === "ULSHMS integration check"),
  );
  assert.ok(
    mailbox.messages.some((m) => m.Subject === "Reset your ULSHMS password"),
  );
  await expectStatus("/api/auth/sign-out", 200, {
    method: "POST",
    cookie: patient.cookie,
    body: {},
  });
  await expectStatus("/api/hms/patients", 401, { cookie: patient.cookie });
  console.log(
    "PASS: session auth, CSRF, role denial, patient/doctor record scope and aggregates, validation, disablement, invalid Firebase proof, concurrent message deduplication, captured SMS, Mailpit email/reset, logout revocation.",
  );
} finally {
  await db.end();
}
