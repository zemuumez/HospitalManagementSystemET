import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { setTimeout as delay } from "node:timers/promises";
import { Pool } from "pg";
import { hashPassword } from "better-auth/crypto";
const base = process.env.BETTER_AUTH_URL,
  schema = process.env.HMS_TEST_ISOLATED_SCHEMA;
assert.match(schema || "", /^hms_browser_[a-f0-9]{24}$/);
assert.equal(new URL(base).hostname, "127.0.0.1");
assert.equal(new URL(process.env.DATABASE_URL).hostname, "127.0.0.1");
const db = new Pool({ connectionString: process.env.DATABASE_URL });
const id = randomUUID(),
  email = `recovery-${id}@example.test`,
  oldPassword = randomUUID() + "Aa1!",
  newPassword = randomUUID() + "Bb2!";
const post = (path, body, cookie = "") =>
  fetch(base + path, {
    method: "POST",
    headers: { origin: base, "content-type": "application/json", cookie },
    body: JSON.stringify(body),
  });
const cookieOf = (response) =>
  response.headers
    .getSetCookie()
    .map((v) => v.split(";")[0])
    .join("; ");
async function resetEmail() {
  for (let n = 0; n < 15; n++) {
    const response = await fetch("http://127.0.0.1:8025/api/v1/messages");
    assert.equal(response.status, 200, "Mailpit available");
    const box = await response.json();
    const message = box.messages.find(
      (m) =>
        m.Subject === "Reset your ULSHMS password" &&
        m.To?.some((to) => to.Address === email),
    );
    if (message) {
      const detail = await fetch(
        `http://127.0.0.1:8025/api/v1/message/${message.ID}`,
      );
      assert.equal(detail.status, 200);
      return detail.json();
    }
    await delay(300);
  }
  throw Error("No password-reset email for the synthetic account");
}
try {
  assert.equal(
    (await db.query("SELECT current_schema() AS name")).rows[0].name,
    schema,
  );
  await db.query('INSERT INTO "user"(id,name,email) VALUES($1,$1,$2)', [
    id,
    email,
  ]);
  await db.query(
    'INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,\'credential\',$2,$3)',
    [randomUUID(), id, await hashPassword(oldPassword)],
  );
  await db.query(
    "INSERT INTO staff_access(user_id,role) VALUES($1,'patient')",
    [id],
  );
  let response = await post("/api/auth/sign-in/email", {
    email,
    password: oldPassword,
  });
  assert.equal(response.status, 200);
  const oldCookie = cookieOf(response);
  response = await post("/api/auth/sign-in/email", {
    email,
    password: oldPassword,
  });
  assert.equal(response.status, 200);
  const secondOldCookie = cookieOf(response);
  response = await post("/api/auth/request-password-reset", {
    email,
    redirectTo: "/reset-password",
  });
  assert.equal(response.status, 200);
  const message = await resetEmail();
  const raw = message.Text.match(/https?:\/\/[^\s]+/)?.[0];
  assert.ok(raw, "reset link delivered");
  const link = new URL(raw);
  assert.equal(link.origin, base);
  assert.ok(link.pathname.includes("/reset-password/"));
  response = await fetch(link, { redirect: "manual" });
  assert.equal(response.status, 302, "email link redirects to reset form");
  const redirect = new URL(response.headers.get("location"), base);
  const token = redirect.searchParams.get("token");
  assert.ok(token, "reset token supplied to form");
  response = await post("/api/auth/reset-password", { token, newPassword });
  assert.equal(response.status, 200, "password reset completed");
  response = await post("/api/auth/reset-password", {
    token,
    newPassword: randomUUID() + "Cc3!",
  });
  assert.equal(response.status, 400, "reset token cannot be reused");
  response = await fetch(base + "/api/hms/me", {
    headers: { cookie: oldCookie },
  });
  assert.equal(response.status, 401, "old session revoked after reset");
  response = await fetch(base + "/api/hms/me", {
    headers: { cookie: secondOldCookie },
  });
  assert.equal(response.status, 401, "all existing sessions revoked");
  response = await post("/api/auth/sign-in/email", {
    email,
    password: oldPassword,
  });
  assert.equal(response.status, 401, "old password rejected");
  response = await post("/api/auth/sign-in/email", {
    email,
    password: newPassword,
  });
  assert.equal(response.status, 200, "new password accepted");
  const freshCookie = cookieOf(response);
  response = await post("/api/auth/sign-out", {}, freshCookie);
  assert.equal(response.status, 200);
  response = await fetch(base + "/api/hms/me", {
    headers: { cookie: freshCookie },
  });
  assert.equal(response.status, 401, "logout revokes replacement session");
  response = await post("/api/auth/request-password-reset", {
    email,
    redirectTo: "/reset-password",
  });
  assert.equal(response.status, 200);
  const verification = (
    await db.query(
      "SELECT identifier FROM verification WHERE value=$1 AND identifier LIKE 'reset-password:%' ORDER BY \"createdAt\" DESC LIMIT 1",
      [id],
    )
  ).rows[0];
  assert.ok(verification, "second reset record exists");
  await db.query(
    "UPDATE verification SET \"expiresAt\"=now()-interval '1 minute' WHERE identifier=$1",
    [verification.identifier],
  );
  response = await post("/api/auth/reset-password", {
    token: verification.identifier.slice("reset-password:".length),
    newPassword: randomUUID() + "Dd4!",
  });
  assert.equal(response.status, 400, "expired reset token rejected");
  response = await post("/api/auth/request-password-reset", {
    email: `unknown-${id}@example.test`,
    redirectTo: "/reset-password",
  });
  assert.equal(
    response.status,
    200,
    "unknown accounts receive generic response",
  );
  response = await post("/api/auth/request-password-reset", {
    email,
    redirectTo: "/reset-password",
  });
  assert.equal(
    response.status,
    429,
    "password reset requests are rate limited",
  );
  console.log(
    "PASS: Mailpit reset link, reset completion, single-use/expiry, old-password denial, session revocation, replacement login/logout and generic unknown-account response and request throttling.",
  );
} finally {
  await db.end();
}
