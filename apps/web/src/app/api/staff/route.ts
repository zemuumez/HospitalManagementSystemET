import { randomUUID } from "node:crypto";
import { hashPassword } from "better-auth/crypto";
import { z } from "zod";
import { auth } from "@/lib/auth";
import { pool } from "@/lib/db";

export const runtime = "nodejs";
const roles = [
  "admin",
  "doctor",
  "patient",
  "nurse",
  "receptionist",
  "pharmacist",
  "accountant",
  "case_manager",
  "lab_technician",
] as const;
const createInput = z
  .object({
    name: z.string().trim().min(1).max(120),
    email: z
      .email()
      .max(254)
      .transform((v) => v.toLowerCase()),
    role: z.enum(roles),
    password: z.string().min(12).max(128),
  })
  .strict();
const changeInput = z
  .object({ id: z.string().min(1).max(128), active: z.boolean() })
  .strict();
function json(body: unknown, status = 200) {
  return Response.json(body, {
    status,
    headers: { "Cache-Control": "no-store" },
  });
}
async function administrator(request: Request) {
  const session = await auth.api.getSession({ headers: request.headers });
  if (!session) return null;
  const result = await pool.query(
    "SELECT user_id FROM staff_access WHERE user_id=$1 AND role='admin' AND active",
    [session.user.id],
  );
  return result.rowCount ? session.user.id : null;
}
export async function GET(request: Request) {
  try {
    const actor = await administrator(request);
    if (!actor) return json({ error: "Administrator access required" }, 403);
    const url = new URL(request.url),
      page = Number(url.searchParams.get("page") || 1),
      search = url.searchParams.get("search") || "";
    if (
      !Number.isInteger(page) ||
      page < 1 ||
      page > 1000 ||
      search.length > 80
    )
      return json({ error: "Invalid search" }, 422);
    const result = await pool.query(
      `SELECT u.id,u.name,u.email,a.role,a.active FROM "user" u JOIN staff_access a ON a.user_id=u.id WHERE ($1='' OR strpos(lower(u.name||' '||u.email),lower($1))>0) ORDER BY u.name,u.id LIMIT 25 OFFSET $2`,
      [search, (page - 1) * 25],
    );
    return json({ users: result.rows, page, pageSize: 25 });
  } catch {
    return json({ error: "Unable to load hospital users" }, 503);
  }
}
async function mutate(request: Request) {
  if (
    request.headers.get("origin") !==
    (process.env.BETTER_AUTH_URL ?? "http://127.0.0.1:3000")
  )
    return json({ error: "Origin not allowed" }, 403);
  try {
    const actor = await administrator(request);
    if (!actor) return json({ error: "Administrator access required" }, 403);
    const reader = request.body?.getReader();
    let body = "",
      size = 0;
    const decoder = new TextDecoder();
    if (reader) {
      while (true) {
        const chunk = await reader.read();
        if (chunk.done) break;
        size += chunk.value.byteLength;
        if (size > 8192) {
          await reader.cancel();
          return json({ error: "Request too large" }, 413);
        }
        body += decoder.decode(chunk.value, { stream: true });
      }
      body += decoder.decode();
    }
    let input: unknown;
    try {
      input = JSON.parse(body);
    } catch {
      return json({ error: "Invalid JSON" }, 400);
    }
    const creating = request.method === "POST";
    const parsed = creating
      ? createInput.safeParse(input)
      : changeInput.safeParse(input);
    if (!parsed.success)
      return json(
        {
          error:
            "Check the supplied fields. Passwords need at least 12 characters.",
        },
        422,
      );
    const hash = creating
      ? await hashPassword(createInput.parse(input).password)
      : null;
    const db = await pool.connect();
    try {
      await db.query("BEGIN");
      // Serialize access changes to protect the final administrator and in-flight checks.
      await db.query("SELECT pg_advisory_xact_lock(72841022)");
      const current = await db.query(
        "SELECT user_id FROM staff_access WHERE user_id=$1 AND role='admin' AND active FOR UPDATE",
        [actor],
      );
      if (!current.rowCount) {
        await db.query("ROLLBACK");
        return json({ error: "Administrator access required" }, 403);
      }
      let id: string;
      if (creating) {
        const data = createInput.parse(input);
        id = randomUUID();
        await db.query('INSERT INTO "user"(id,name,email) VALUES($1,$2,$3)', [
          id,
          data.name,
          data.email,
        ]);
        await db.query(
          `INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,'credential',$2,$3)`,
          [randomUUID(), id, hash],
        );
        await db.query("INSERT INTO staff_access(user_id,role) VALUES($1,$2)", [
          id,
          data.role,
        ]);
      } else {
        const data = changeInput.parse(input);
        id = data.id;
        if (id === actor && !data.active) {
          await db.query("ROLLBACK");
          return json({ error: "You cannot disable your own account" }, 409);
        }
        const target = await db.query(
          "SELECT role,active FROM staff_access WHERE user_id=$1 FOR UPDATE",
          [id],
        );
        if (!target.rowCount) {
          await db.query("ROLLBACK");
          return json({ error: "User not found" }, 404);
        }
        if (target.rows[0].role === "admin" && !data.active) {
          const count = await db.query(
            "SELECT count(*)::int AS count FROM staff_access WHERE role='admin' AND active",
          );
          if (count.rows[0].count <= 1) {
            await db.query("ROLLBACK");
            return json(
              { error: "The hospital must retain an active administrator" },
              409,
            );
          }
        }
        await db.query("UPDATE staff_access SET active=$2 WHERE user_id=$1", [
          id,
          data.active,
        ]);
        if (!data.active)
          await db.query('DELETE FROM session WHERE "userId"=$1', [id]);
      }
      await db.query(
        "INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,$2,$3)",
        [
          actor,
          creating ? "identity.provisioned" : "identity.access_changed",
          id,
        ],
      );
      await db.query("COMMIT");
      return json({ id }, creating ? 201 : 200);
    } catch (error) {
      await db.query("ROLLBACK");
      if ((error as { code?: string }).code === "23505")
        return json({ error: "This email is already registered" }, 409);
      throw error;
    } finally {
      db.release();
    }
  } catch {
    return json({ error: "Unable to update hospital users" }, 503);
  }
}
export { mutate as POST, mutate as PATCH };
