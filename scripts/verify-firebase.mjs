import assert from "node:assert/strict";
import { createHmac, randomUUID } from "node:crypto";
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
  return { id, email, password, cookie: cookies(response) };
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
function totp(uri) {
  const encoded = new URL(uri).searchParams.get("secret");
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";
  let bits = "";
  for (const char of encoded.replace(/=+$/, "").toUpperCase())
    bits += alphabet.indexOf(char).toString(2).padStart(5, "0");
  const bytes = [];
  for (let i = 0; i + 8 <= bits.length; i += 8)
    bytes.push(parseInt(bits.slice(i, i + 8), 2));
  const counter = Buffer.alloc(8);
  counter.writeBigUInt64BE(BigInt(Math.floor(Date.now() / 30000)));
  const digest = createHmac("sha1", Buffer.from(bytes))
    .update(counter)
    .digest();
  const offset = digest[19] & 15;
  return ((digest.readUInt32BE(offset) & 0x7fffffff) % 1000000)
    .toString()
    .padStart(6, "0");
}
async function resetLimits() {
  // Only this generated disposable schema is writable by the suite.
  await db.query('DELETE FROM "rateLimit"');
}
async function checkMfa(user) {
  await resetLimits();
  let response = await post(
    "/api/auth/two-factor/enable",
    { password: user.password },
    user.cookie,
  );
  assert.equal(response.status, 200, "MFA enrollment");
  const { totpURI, backupCodes } = await response.json();
  assert.ok(totpURI && backupCodes.length);
  response = await post(
    "/api/auth/two-factor/verify-totp",
    { code: totp(totpURI) },
    user.cookie,
  );
  assert.equal(response.status, 200, "MFA enrollment verification");
  assert.equal(
    (await fetch(base + "/api/hms/me", { headers: { cookie: user.cookie } }))
      .status,
    401,
    "enrollment revokes pre-MFA sessions",
  );
  await resetLimits();
  response = await post("/api/auth/sign-in/email", {
    email: user.email,
    password: user.password,
  });
  assert.equal(response.status, 200);
  assert.equal(
    (await response.json()).twoFactorRedirect,
    true,
    "password login requires MFA",
  );
  const challenge = cookies(response);
  assert.equal(
    (await fetch(base + "/api/hms/me", { headers: { cookie: challenge } }))
      .status,
    401,
    "challenge is not an authenticated session",
  );
  response = await post(
    "/api/auth/two-factor/verify-totp",
    { code: "invalid" },
    challenge,
  );
  assert.equal(response.ok, false, "bad TOTP rejected");
  response = await post(
    "/api/auth/two-factor/verify-totp",
    { code: totp(totpURI) },
    challenge,
  );
  assert.equal(response.status, 200, "valid TOTP completes sign-in");
  assert.equal(
    (
      await fetch(base + "/api/hms/me", {
        headers: { cookie: cookies(response) },
      })
    ).status,
    200,
  );
  await resetLimits();
  response = await post(
    "/api/auth/two-factor/verify-totp",
    { code: totp(totpURI) },
    challenge,
  );
  assert.equal(response.ok, false, "consumed challenge rejected");
  response = await post("/api/auth/firebase/sign-in", {
    idToken: await phoneProof(),
  });
  assert.equal(response.status, 200);
  assert.equal(
    (await response.json()).twoFactorRedirect,
    true,
    "Firebase cannot bypass MFA",
  );
  const phoneChallenge = cookies(response);
  assert.equal(
    (await fetch(base + "/api/hms/me", { headers: { cookie: phoneChallenge } }))
      .status,
    401,
  );
  response = await post(
    "/api/auth/two-factor/verify-backup-code",
    { code: backupCodes[0] },
    phoneChallenge,
  );
  assert.equal(response.status, 200, "backup code completes Firebase MFA");
  assert.equal(
    (
      await fetch(base + "/api/hms/me", {
        headers: { cookie: cookies(response) },
      })
    ).status,
    200,
  );
  await resetLimits();
  response = await post("/api/auth/sign-in/email", {
    email: user.email,
    password: user.password,
  });
  const recoveryChallenge = cookies(response);
  response = await post(
    "/api/auth/two-factor/verify-backup-code",
    { code: backupCodes[0] },
    recoveryChallenge,
  );
  assert.equal(response.ok, false, "backup code is single use");
  assert.equal(
    (
      await db.query(
        "SELECT count(*)::int AS n FROM audit_event WHERE actor_id=$1 AND action='identity.mfa_enabled'",
        [user.id],
      )
    ).rows[0].n,
    1,
  );
  for (let attempt = 0; attempt < 5; attempt++) {
    await resetLimits();
    response = await post("/api/auth/sign-in/email", {
      email: user.email,
      password: user.password,
    });
    response = await post(
      "/api/auth/two-factor/verify-totp",
      { code: "invalid" },
      cookies(response),
    );
    assert.equal(response.ok, false);
  }
  assert.ok(
    (
      await db.query(
        'SELECT "lockedUntil">now() AS locked FROM "twoFactor" WHERE "userId"=$1',
        [user.id],
      )
    ).rows[0].locked,
    "failed attempts lock account MFA",
  );
  await resetLimits();
  response = await post("/api/auth/sign-in/email", {
    email: user.email,
    password: user.password,
  });
  response = await post(
    "/api/auth/two-factor/verify-totp",
    { code: totp(totpURI) },
    cookies(response),
  );
  assert.equal(response.ok, false, "valid code cannot bypass account lockout");
  // Advance only this disposable fixture's lock deadline, not production behavior.
  await db.query(
    'UPDATE "twoFactor" SET "lockedUntil"=now()-interval \'1 second\' WHERE "userId"=$1',
    [user.id],
  );
  await resetLimits();
  response = await post("/api/auth/sign-in/email", {
    email: user.email,
    password: user.password,
  });
  response = await post(
    "/api/auth/two-factor/verify-totp",
    { code: totp(totpURI) },
    cookies(response),
  );
  assert.equal(response.status, 200, "sign-in recovers after lock expiry");
  const authenticated = cookies(response);
  response = await post(
    "/api/auth/two-factor/disable",
    { password: "wrong-password" },
    authenticated,
  );
  assert.equal(response.ok, false, "password required to disable MFA");
  await resetLimits();
  response = await post(
    "/api/auth/two-factor/disable",
    { password: user.password },
    authenticated,
  );
  assert.equal(response.status, 200, "explicit password-confirmed disable");
  assert.equal(
    (await fetch(base + "/api/hms/me", { headers: { cookie: authenticated } }))
      .status,
    401,
    "disable revokes previous sessions",
  );
  assert.equal(
    (
      await db.query(
        "SELECT count(*)::int AS n FROM audit_event WHERE actor_id=$1 AND action='identity.mfa_disabled'",
        [user.id],
      )
    ).rows[0].n,
    1,
  );
  console.log(
    "PASS: authenticator enrollment, prior-session revocation, email and Firebase MFA enforcement, invalid code/challenge replay denial and single-use recovery code.",
  );
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
  await checkMfa(winner);
  await resetLimits();
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
