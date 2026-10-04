import type { BetterAuthPlugin } from "better-auth";
import {
  APIError,
  createAuthEndpoint,
  sessionMiddleware,
} from "better-auth/api";
import { hashPassword } from "better-auth/crypto";
import { createHash, randomBytes, randomUUID } from "node:crypto";
import { z } from "zod";
import type { PoolClient } from "pg";
import { pool } from "./db";
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
const digest = (token: string) =>
  createHash("sha256").update(token).digest("hex");
async function administrator(db: PoolClient, id: string) {
  await db.query("SELECT pg_advisory_xact_lock(72841022)");
  const result = await db.query(
    "SELECT user_id FROM staff_access WHERE user_id=$1 AND role='admin' AND active FOR UPDATE",
    [id],
  );
  if (!result.rowCount)
    throw new APIError("FORBIDDEN", {
      message: "Administrator access required",
    });
}
async function issue(
  db: PoolClient,
  actor: string,
  userId: string,
  email: string,
  role: string,
) {
  const id = randomUUID(),
    token = randomBytes(32).toString("hex");
  await db.query(
    "INSERT INTO staff_invitation(id,user_id,actor_id,role,token_digest,expires_at) VALUES($1,$2,$3,$4,$5,now()+interval '24 hours')",
    [id, userId, actor, role, digest(token)],
  );
  const link = new URL(
    "/accept-invitation",
    process.env.BETTER_AUTH_URL ?? "http://127.0.0.1:3000",
  );
  link.searchParams.set("token", token);
  await db.query(
    "INSERT INTO message_outbox(actor_id,channel,recipient,subject,body,idempotency_key) VALUES($1,'email',$2,$3,$4,$5)",
    [
      actor,
      email,
      "Your ULSHMS invitation",
      `A hospital administrator invited you to ULSHMS. Set your password using this single-use link within 24 hours: ${link.href}\n\nIf you did not expect this invitation, contact the hospital.`,
      "invitation:" + id,
    ],
  );
  await db.query(
    "INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'identity.invited',$2)",
    [actor, userId],
  );
  return { id, userId };
}
export const staffInvitations = () =>
  ({
    id: "hospital-invitations",
    endpoints: {
      listStaffInvitations: createAuthEndpoint(
        "/staff/invitations",
        {
          method: "GET",
          use: [sessionMiddleware],
          query: z.object({
            page: z.coerce.number().int().min(1).max(1000).default(1),
          }),
        },
        async (ctx) => {
          const db = await pool.connect();
          try {
            await db.query("BEGIN");
            await administrator(db, ctx.context.session.user.id);
            const result = await db.query(
              'SELECT i.id,i.user_id AS "userId",u.name,u.email,i.role,i.created_at AS "createdAt",i.expires_at AS "expiresAt",i.consumed_at AS "consumedAt",i.revoked_at AS "revokedAt" FROM staff_invitation i JOIN "user" u ON u.id=i.user_id ORDER BY i.created_at DESC,i.id LIMIT 25 OFFSET $1',
              [(ctx.query.page - 1) * 25],
            );
            await db.query(
              "INSERT INTO audit_event(actor_id,action) VALUES($1,'identity.invitations_viewed')",
              [ctx.context.session.user.id],
            );
            await db.query("COMMIT");
            ctx.setHeader("Cache-Control", "no-store");
            return ctx.json({
              invitations: result.rows,
              page: ctx.query.page,
              pageSize: 25,
            });
          } catch (e) {
            await db.query("ROLLBACK");
            throw e;
          } finally {
            db.release();
          }
        },
      ),
      inviteStaff: createAuthEndpoint(
        "/staff/invite",
        {
          method: "POST",
          use: [sessionMiddleware],
          body: z
            .object({
              name: z.string().trim().min(1).max(120),
              email: z
                .email()
                .max(254)
                .transform((v) => v.toLowerCase()),
              role: z.enum(roles),
            })
            .strict(),
        },
        async (ctx) => {
          const db = await pool.connect();
          try {
            await db.query("BEGIN");
            await administrator(db, ctx.context.session.user.id);
            const existing = await db.query(
              'SELECT id FROM "user" WHERE lower(email)=lower($1)',
              [ctx.body.email],
            );
            if (existing.rowCount)
              throw new APIError("CONFLICT", {
                message: "This email already has an account",
              });
            const id = randomUUID();
            await db.query(
              'INSERT INTO "user"(id,name,email) VALUES($1,$2,$3)',
              [id, ctx.body.name, ctx.body.email],
            );
            await db.query(
              "INSERT INTO staff_access(user_id,role,active) VALUES($1,$2,false)",
              [id, ctx.body.role],
            );
            const result = await issue(
              db,
              ctx.context.session.user.id,
              id,
              ctx.body.email,
              ctx.body.role,
            );
            await db.query("COMMIT");
            return ctx.json(result);
          } catch (e) {
            await db.query("ROLLBACK");
            throw e;
          } finally {
            db.release();
          }
        },
      ),
      resendStaffInvitation: createAuthEndpoint(
        "/staff/resend-invitation",
        {
          method: "POST",
          use: [sessionMiddleware],
          body: z.object({ id: z.uuid() }).strict(),
        },
        async (ctx) => {
          const db = await pool.connect();
          try {
            await db.query("BEGIN");
            await administrator(db, ctx.context.session.user.id);
            const result = await db.query(
              'SELECT i.user_id,i.role,u.email,a.active FROM staff_invitation i JOIN "user" u ON u.id=i.user_id JOIN staff_access a ON a.user_id=i.user_id WHERE i.id=$1 AND i.consumed_at IS NULL AND i.revoked_at IS NULL AND NOT EXISTS(SELECT 1 FROM account WHERE "userId"=i.user_id) FOR UPDATE OF i,a',
              [ctx.body.id],
            );
            const invite = result.rows[0];
            if (!invite || invite.active)
              throw new APIError("CONFLICT", {
                message: "Invitation cannot be renewed",
              });
            await db.query(
              "UPDATE staff_invitation SET revoked_at=now() WHERE user_id=$1 AND consumed_at IS NULL AND revoked_at IS NULL",
              [invite.user_id],
            );
            const renewed = await issue(
              db,
              ctx.context.session.user.id,
              invite.user_id,
              invite.email,
              invite.role,
            );
            await db.query("COMMIT");
            return ctx.json(renewed);
          } catch (e) {
            await db.query("ROLLBACK");
            throw e;
          } finally {
            db.release();
          }
        },
      ),
      revokeStaffInvitation: createAuthEndpoint(
        "/staff/revoke-invitation",
        {
          method: "POST",
          use: [sessionMiddleware],
          body: z.object({ id: z.uuid() }).strict(),
        },
        async (ctx) => {
          const db = await pool.connect();
          try {
            await db.query("BEGIN");
            await administrator(db, ctx.context.session.user.id);
            const result = await db.query(
              "UPDATE staff_invitation SET revoked_at=now() WHERE id=$1 AND consumed_at IS NULL AND revoked_at IS NULL RETURNING user_id",
              [ctx.body.id],
            );
            if (!result.rowCount)
              throw new APIError("CONFLICT", {
                message: "Invitation is no longer pending",
              });
            await db.query(
              "INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'identity.invitation_revoked',$2)",
              [ctx.context.session.user.id, result.rows[0].user_id],
            );
            await db.query("COMMIT");
            return ctx.json({ success: true });
          } catch (e) {
            await db.query("ROLLBACK");
            throw e;
          } finally {
            db.release();
          }
        },
      ),
      acceptStaffInvitation: createAuthEndpoint(
        "/staff/accept-invitation",
        {
          method: "POST",
          body: z
            .object({
              token: z.string().regex(/^[a-f0-9]{64}$/),
              password: z.string().min(12).max(128),
            })
            .strict(),
        },
        async (ctx) => {
          const hash = await hashPassword(ctx.body.password);
          const db = await pool.connect();
          try {
            await db.query("BEGIN");
            await db.query("SELECT pg_advisory_xact_lock(72841022)");
            const result = await db.query(
              "SELECT i.id,i.user_id,i.role,a.role AS current_role,a.active FROM staff_invitation i JOIN staff_access a ON a.user_id=i.user_id WHERE i.token_digest=$1 AND i.expires_at>clock_timestamp() AND i.consumed_at IS NULL AND i.revoked_at IS NULL FOR UPDATE OF i,a",
              [digest(ctx.body.token)],
            );
            const invite = result.rows[0];
            if (!invite || invite.active || invite.role !== invite.current_role)
              throw new APIError("BAD_REQUEST", {
                message: "Invitation is invalid or expired",
              });
            const exists = await db.query(
              'SELECT id FROM account WHERE "userId"=$1',
              [invite.user_id],
            );
            if (exists.rowCount)
              throw new APIError("BAD_REQUEST", {
                message: "Invitation is invalid or expired",
              });
            await db.query(
              `INSERT INTO account(id,"accountId","providerId","userId",password) VALUES($1,$2,'credential',$2,$3)`,
              [randomUUID(), invite.user_id, hash],
            );
            await db.query(
              'UPDATE "user" SET "emailVerified"=true,"updatedAt"=now() WHERE id=$1',
              [invite.user_id],
            );
            await db.query(
              "UPDATE staff_access SET active=true WHERE user_id=$1",
              [invite.user_id],
            );
            await db.query(
              "UPDATE staff_invitation SET consumed_at=now() WHERE id=$1",
              [invite.id],
            );
            await db.query(
              "INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'identity.invitation_accepted',$1)",
              [invite.user_id],
            );
            await db.query("COMMIT");
            return ctx.json({ success: true });
          } catch (e) {
            await db.query("ROLLBACK");
            throw e;
          } finally {
            db.release();
          }
        },
      ),
    },
    rateLimit: [
      {
        pathMatcher: (path: string) => path.startsWith("/staff/"),
        window: 60,
        max: 5,
      },
    ],
  }) satisfies BetterAuthPlugin;
