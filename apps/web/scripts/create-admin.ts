import { randomUUID } from "node:crypto";
import { hashPassword } from "better-auth/crypto";
import { pool } from "../src/lib/db.ts";
async function main() {
  const email = process.env.ADMIN_EMAIL?.trim().toLowerCase();
  const password = process.env.ADMIN_PASSWORD;
  if (!email || !password || password.length < 12)
    throw new Error(
      "Set ADMIN_EMAIL and ADMIN_PASSWORD (at least 12 characters)",
    );
  const hash = await hashPassword(password);
  const db = await pool.connect();
  try {
    await db.query("BEGIN");
    const id = randomUUID();
    await db.query(
      'INSERT INTO "user"(id,name,email,"emailVerified") VALUES($1,$2,$3,true)',
      [id, process.env.ADMIN_NAME ?? "Hospital administrator", email],
    );
    await db.query(
      'INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,\'credential\',$2,$3)',
      [randomUUID(), id, hash],
    );
    await db.query(
      "INSERT INTO staff_access(user_id,role) VALUES($1,'admin')",
      [id],
    );
    await db.query(
      "INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'identity.admin_provisioned',$1)",
      [id],
    );
    await db.query("COMMIT");
    console.log(
      "Administrator created. Existing accounts are never overwritten.",
    );
  } catch (error) {
    await db.query("ROLLBACK");
    throw error;
  } finally {
    db.release();
    await pool.end();
  }
}
main().catch(() => {
  console.error(
    "Could not create administrator. Check credentials, minimum password length, and duplicate email.",
  );
  process.exitCode = 1;
});
