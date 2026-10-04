import assert from "node:assert/strict";
import { randomUUID } from "node:crypto";
import { setTimeout as delay } from "node:timers/promises";
import { Pool } from "pg";
import { hashPassword } from "better-auth/crypto";
const base = process.env.BETTER_AUTH_URL,
  schema = process.env.HMS_TEST_ISOLATED_SCHEMA;
assert.match(schema || "", /^hms_browser_[a-f0-9]{24}$/);
assert.equal(new URL(base).hostname, "127.0.0.1");
const db = new Pool({ connectionString: process.env.DATABASE_URL });
const post = (path, body, cookie = "", origin = base) =>
  fetch(base + "/api/auth" + path, {
    method: "POST",
    headers: { origin, cookie, "content-type": "application/json" },
    body: JSON.stringify(body),
  });
const cookies = (r) =>
  r.headers
    .getSetCookie()
    .map((c) => c.split(";")[0])
    .join("; ");
async function account(role) {
  const id = randomUUID(),
    email = `invite-${id}@example.test`,
    password = randomUUID() + "Aa1!";
  await db.query('INSERT INTO "user"(id,name,email) VALUES($1,$1,$2)', [
    id,
    email,
  ]);
  await db.query("INSERT INTO staff_access(user_id,role) VALUES($1,$2)", [
    id,
    role,
  ]);
  await db.query(
    `INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,'credential',$2,$3)`,
    [randomUUID(), id, await hashPassword(password)],
  );
  const r = await post("/sign-in/email", { email, password });
  assert.equal(r.status, 200);
  return { id, email, password, cookie: cookies(r) };
}
async function emailToken(email, exclude = "") {
  for (let n = 0; n < 40; n++) {
    const box = await (
      await fetch("http://127.0.0.1:8025/api/v1/messages")
    ).json();
    for (const message of box.messages.filter(
      (m) =>
        m.Subject === "Your ULSHMS invitation" &&
        m.To?.some((to) => to.Address === email),
    )) {
      const detail = await (
        await fetch(`http://127.0.0.1:8025/api/v1/message/${message.ID}`)
      ).json();
      const link = new URL(detail.Text.match(/https?:\/\/[^\s]+/)[0]);
      assert.equal(link.origin, base);
      const token = link.searchParams.get("token");
      if (token !== exclude) return token;
    }
    await delay(300);
  }
  throw Error("Synthetic invitation email not delivered");
}
try {
  assert.equal(
    (await db.query("SELECT current_schema() AS name")).rows[0].name,
    schema,
  );
  const admin = await account("admin"),
    patient = await account("patient");
  const email = `new-${randomUUID()}@example.test`,
    password = randomUUID() + "Aa1!";
  const input = { name: "Invited Nurse", email, role: "nurse" };
  assert.equal(
    (await post("/staff/invite", input, patient.cookie)).status,
    403,
  );
  assert.equal(
    (await post("/staff/invite", input, admin.cookie, "https://evil.example"))
      .status,
    403,
  );
  let r = await post("/staff/invite", input, admin.cookie);
  assert.equal(r.status, 200);
  const invite = await r.json();
  assert.ok(invite.id && invite.userId);
  assert.equal(invite.token, undefined);
  assert.equal(
    (
      await db.query("SELECT active FROM staff_access WHERE user_id=$1", [
        invite.userId,
      ])
    ).rows[0].active,
    false,
  );
  assert.equal((await post("/sign-in/email", { email, password })).status, 401);
  const token = await emailToken(email);
  assert.equal(
    (await post("/staff/accept-invitation", { token, password: "short" }))
      .status,
    400,
  );
  const accepted = await Promise.all([
    post("/staff/accept-invitation", { token, password }),
    post("/staff/accept-invitation", { token, password }),
  ]);
  assert.deepEqual(
    accepted.map((r) => r.status).sort(),
    [200, 400],
    "single-use concurrent acceptance",
  );
  assert.equal(
    (
      await db.query('SELECT "emailVerified" FROM "user" WHERE id=$1', [
        invite.userId,
      ])
    ).rows[0].emailVerified,
    true,
  );
  await db.query('DELETE FROM "rateLimit"');
  r = await post("/sign-in/email", { email, password });
  assert.equal(r.status, 200);
  const signed = cookies(r);
  r = await fetch(base + "/api/hms/me", { headers: { cookie: signed } });
  assert.equal(r.status, 200);
  assert.equal((await r.json()).user.role, "nurse");
  assert.equal(
    (await post("/staff/invite", input, admin.cookie)).status,
    409,
    "existing identity cannot be claimed",
  );
  assert.equal(
    (await post("/staff/revoke-invitation", { id: invite.id }, admin.cookie))
      .status,
    409,
    "accepted invite cannot revoke account",
  );
  const nextEmail = `next-${randomUUID()}@example.test`;
  r = await post("/staff/invite", { ...input, email: nextEmail }, admin.cookie);
  assert.equal(r.status, 200);
  const pending = await r.json();
  const oldToken = await emailToken(nextEmail);
  await db.query(
    "UPDATE staff_invitation SET expires_at=now()-interval '1 second' WHERE id=$1",
    [pending.id],
  );
  assert.equal(
    (await post("/staff/accept-invitation", { token: oldToken, password }))
      .status,
    400,
    "expired invitation rejected",
  );
  r = await post("/staff/resend-invitation", { id: pending.id }, admin.cookie);
  assert.equal(r.status, 200);
  const renewed = await r.json();
  const newToken = await emailToken(nextEmail, oldToken);
  assert.notEqual(newToken, oldToken);
  await db.query('DELETE FROM "rateLimit"');
  assert.equal(
    (await post("/staff/accept-invitation", { token: oldToken, password }))
      .status,
    400,
    "renewal invalidates old token",
  );
  assert.equal(
    (await post("/staff/revoke-invitation", { id: renewed.id }, admin.cookie))
      .status,
    200,
  );
  assert.equal(
    (await post("/staff/accept-invitation", { token: newToken, password }))
      .status,
    400,
    "revoked invitation rejected",
  );
  assert.equal(
    (
      await db.query("SELECT active FROM staff_access WHERE user_id=$1", [
        pending.userId,
      ])
    ).rows[0].active,
    false,
  );
  const actions = (
    await db.query("SELECT action FROM audit_event WHERE resource_id=$1", [
      invite.userId,
    ])
  ).rows.map((r) => r.action);
  assert.ok(
    actions.includes("identity.invited") &&
      actions.includes("identity.invitation_accepted"),
  );
  await db.query('DELETE FROM "rateLimit"');
  r = await fetch(base + "/api/auth/staff/invitations?page=1", {
    headers: { cookie: admin.cookie },
  });
  assert.equal(r.status, 200);
  const listing = await r.json();
  assert.equal(listing.invitations.length, 3);
  assert.ok(
    listing.invitations.some((v) => v.id === invite.id && v.consumedAt),
  );
  assert.equal(
    JSON.stringify(listing).includes(token),
    false,
    "list does not expose invitation bearer token",
  );
  assert.equal(
    (
      await fetch(base + "/api/auth/staff/invitations?page=1", {
        headers: { cookie: patient.cookie },
      })
    ).status,
    403,
  );
  assert.equal(
    (
      await fetch(base + "/api/auth/staff/invitations?page=0", {
        headers: { cookie: admin.cookie },
      })
    ).status,
    400,
  );
  await db.query('DELETE FROM "rateLimit"');
  const receptionist = await account("receptionist");
  r = await fetch(base + "/api/hms/messages", {
    headers: { cookie: receptionist.cookie },
  });
  assert.equal(r.status, 200);
  const operational = await r.json();
  assert.equal(
    JSON.stringify(operational).includes(email),
    false,
    "identity email metadata excluded from operational messages",
  );
  assert.equal(JSON.stringify(operational).includes(nextEmail), false);
  assert.equal(
    (
      await db.query(
        "SELECT count(*)::int AS count FROM message_outbox WHERE audience='identity'",
      )
    ).rows[0].count,
    3,
    "identity delivery remains queued/sent",
  );
  console.log(
    "PASS: invitation administrator/origin controls, Mailpit delivery, inactive pending identity, password policy, verified email, concurrent single-use acceptance, normal session/Go role, expiry, renewal, revocation and retained audit.",
  );
} finally {
  await db.end();
}
