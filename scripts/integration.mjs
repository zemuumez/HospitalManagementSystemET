import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { Pool } from "pg";
import { hashPassword } from "better-auth/crypto";

// Requires local API/web/worker/Mailpit. Fixtures are synthetic and removed afterward.
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
  await db.query(
    "UPDATE patient SET user_id=$1,clinician_user_id=$2 WHERE id=$3",
    [patient.id, doctor.id, patients[0]],
  );
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
  await db.query("DELETE FROM patient WHERE id=ANY($1::uuid[])", [patients]);
  await db.query("DELETE FROM message_outbox WHERE actor_id=ANY($1::text[])", [
    ids,
  ]);
  await db.query("DELETE FROM audit_event WHERE actor_id=ANY($1::text[])", [
    ids,
  ]);
  await db.query('DELETE FROM "user" WHERE id=ANY($1::text[])', [ids]);
  await db.end();
}
