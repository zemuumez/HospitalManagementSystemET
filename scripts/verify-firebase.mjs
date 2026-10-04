import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { setTimeout as delay } from "node:timers/promises";
import { Pool } from "pg";
import { hashPassword } from "better-auth/crypto";
import { initializeApp, deleteApp } from "firebase-admin/app";
import { getAuth } from "firebase-admin/auth";

const base = process.env.BETTER_AUTH_URL;
const project = process.env.FIREBASE_PROJECT_ID;
const host = process.env.FIREBASE_AUTH_EMULATOR_HOST;
const schema = process.env.HMS_TEST_ISOLATED_SCHEMA;
assert.match(schema || "", /^hms_browser_[a-f0-9]{24}$/);
assert.match(project || "", /^demo-hms-[a-f0-9]{12}$/);
assert.match(host || "", /^127\.0\.0\.1:\d+$/);
assert.equal(new URL(base).hostname, "127.0.0.1");
assert.equal(new URL(process.env.DATABASE_URL).hostname, "127.0.0.1");
const db = new Pool({ connectionString: process.env.DATABASE_URL });
async function post(path, body, cookie = "") {
  return fetch(base + path, {
    method: "POST",
    headers: {
      origin: base,
      "content-type": "application/json",
      ...(cookie ? { cookie } : {}),
    },
    body: JSON.stringify(body),
  });
}
function cookies(response) {
  return response.headers
    .getSetCookie()
    .map((v) => v.split(";")[0])
    .join("; ");
}
async function fixture() {
  const id = randomUUID(),
    password = randomUUID() + "Aa1!",
    email = `firebase-${id}@example.test`;
  await db.query('INSERT INTO "user"(id,name,email) VALUES($1,$1,$2)', [
    id,
    email,
  ]);
  await db.query(
    'INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,\'credential\',$2,$3)',
    [randomUUID(), id, await hashPassword(password)],
  );
  await db.query("INSERT INTO staff_access(user_id,role) VALUES($1,'admin')", [
    id,
  ]);
  const response = await post("/api/auth/sign-in/email", { email, password });
  assert.equal(response.status, 200, "fixture sign-in");
  return { id, cookie: cookies(response) };
}
async function phoneProof() {
  // Distinct OTP flows must cross a JWT issuance second to produce distinct proofs.
  await delay(1100);
  const root = `http://${host}`;
  const sent = await fetch(
    root +
      "/identitytoolkit.googleapis.com/v1/accounts:sendVerificationCode?key=fake-api-key",
    {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        phoneNumber: "+251911111111",
        recaptchaToken: "emulator-only",
      }),
    },
  );
  assert.equal(sent.status, 200, "emulator OTP request");
  const { sessionInfo } = await sent.json();
  const response = await fetch(
    root + `/emulator/v1/projects/${project}/verificationCodes`,
  );
  assert.equal(response.status, 200);
  const { verificationCodes } = await response.json();
  const code = verificationCodes.find((v) => v.sessionInfo === sessionInfo);
  assert.ok(code, "emulator verification code available");
  const signed = await fetch(
    root +
      "/identitytoolkit.googleapis.com/v1/accounts:signInWithPhoneNumber?key=fake-api-key",
    {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ sessionInfo, code: code.code }),
    },
  );
  assert.equal(signed.status, 200, "emulator OTP confirmation");
  const { idToken } = await signed.json();
  assert.ok(idToken);
  return idToken;
}
try {
  assert.equal(
    (await db.query("SELECT current_schema() AS name")).rows[0].name,
    schema,
  );
  const users = [await fixture(), await fixture()];
  const proof = await phoneProof();
  let response = await post("/api/auth/firebase/sign-in", { idToken: proof });
  assert.equal(
    response.status,
    401,
    "unlinked identity must not create account",
  );
  const linked = await Promise.all(
    users.map((user) =>
      post("/api/auth/firebase/link", { idToken: proof }, user.cookie),
    ),
  );
  assert.deepEqual(
    linked.map((r) => r.status).sort(),
    [200, 401],
    "one account wins concurrent phone linking",
  );
  const winner = users[linked[0].status === 200 ? 0 : 1],
    loser = users[linked[0].status === 200 ? 1 : 0];
  assert.equal(
    (await db.query("SELECT user_id FROM firebase_identity")).rows[0].user_id,
    winner.id,
  );
  response = await post("/api/auth/firebase/sign-in", { idToken: proof });
  assert.equal(response.status, 401, "consumed link proof cannot sign in");
  const fresh = await phoneProof();
  const signed = await Promise.all([
    post("/api/auth/firebase/sign-in", { idToken: fresh }),
    post("/api/auth/firebase/sign-in", { idToken: fresh }),
  ]);
  assert.deepEqual(
    signed.map((r) => r.status).sort(),
    [200, 401],
    "one session from concurrent proof replay",
  );
  const phoneCookie = cookies(signed.find((r) => r.status === 200));
  response = await fetch(base + "/api/auth/get-session", {
    headers: { cookie: phoneCookie },
  });
  assert.equal(response.status, 200);
  assert.equal((await response.json()).user.id, winner.id);
  response = await fetch(base + "/api/hms/me", {
    headers: { cookie: phoneCookie },
  });
  assert.equal(
    response.status,
    200,
    "Go accepts the phone-authenticated Better Auth session",
  );
  const freshAgain = await phoneProof();
  response = await post(
    "/api/auth/firebase/link",
    { idToken: freshAgain },
    loser.cookie,
  );
  assert.equal(response.status, 401, "phone binding cannot be taken over");
  await db.query(
    'UPDATE session SET "createdAt"=now()-interval \'10 minutes\' WHERE "userId"=$1',
    [loser.id],
  );
  response = await post(
    "/api/auth/firebase/link",
    { idToken: freshAgain },
    loser.cookie,
  );
  assert.equal(response.status, 401, "old session cannot bind a phone");
  await db.query("UPDATE staff_access SET active=false WHERE user_id=$1", [
    winner.id,
  ]);
  response = await post("/api/auth/firebase/sign-in", { idToken: freshAgain });
  assert.equal(
    response.status,
    401,
    "disabled account cannot sign in by phone",
  );
  response = await fetch(base + "/api/hms/me", {
    headers: { cookie: phoneCookie },
  });
  assert.equal(response.status, 401, "disabled account session denied by Go");
  await db.query('DELETE FROM "rateLimit"'); // Reset only disposable rate-limit fixtures for additional negative cases.
  await db.query("UPDATE staff_access SET active=true WHERE user_id=$1", [
    winner.id,
  ]);
  const revokeProof = await phoneProof();
  await delay(1100);
  const adminApp = initializeApp({ projectId: project }, "hms-emulator-test");
  try {
    const { firebase_uid: uid } = (
      await db.query(
        "SELECT firebase_uid FROM firebase_identity WHERE user_id=$1",
        [winner.id],
      )
    ).rows[0];
    await getAuth(adminApp).revokeRefreshTokens(uid);
    response = await post("/api/auth/firebase/sign-in", {
      idToken: revokeProof,
    });
    assert.equal(response.status, 401, "revoked Firebase token cannot sign in");
  } finally {
    await deleteApp(adminApp);
  }
  console.log(
    "PASS: real Firebase Auth emulator OTP, explicit linking race, phone sign-in, Better Auth/Go session integration, replay denial, takeover denial, recent-session guard, disabled-account and revoked-token denial.",
  );
} finally {
  await db.end();
}
