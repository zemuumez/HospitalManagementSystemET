import type { BetterAuthPlugin } from "better-auth";
import {
  APIError,
  createAuthEndpoint,
  sessionMiddleware,
} from "better-auth/api";
import { setSessionCookie } from "better-auth/cookies";
import { createHash } from "node:crypto";
import { z } from "zod";
import { pool } from "./db";
import { verifyFirebasePhone } from "./firebase-proof";

const body = z.object({ idToken: z.string().min(20).max(10000) });
async function consume(idToken: string, linkUserId?: string) {
  let proof;
  try {
    proof = await verifyFirebasePhone(idToken);
  } catch {
    throw new APIError("UNAUTHORIZED", {
      message: "Phone verification failed. Request a new code.",
    });
  }
  const db = await pool.connect();
  try {
    await db.query("BEGIN");
    await db.query("DELETE FROM consumed_phone_proof WHERE expires_at < now()");
    const used = await db.query(
      "INSERT INTO consumed_phone_proof(digest,expires_at) VALUES($1,$2) ON CONFLICT DO NOTHING RETURNING digest",
      [createHash("sha256").update(idToken).digest("hex"), proof.expires],
    );
    if (!used.rowCount) throw new Error("Proof already consumed");
    if (linkUserId) {
      const allowed = await db.query(
        "SELECT user_id FROM staff_access WHERE user_id=$1 AND active",
        [linkUserId],
      );
      if (!allowed.rowCount) throw new Error("Inactive account");
      // Never link on matching email/phone or overwrite another identity.
      await db.query(
        "INSERT INTO firebase_identity(firebase_uid,user_id,phone) VALUES($1,$2,$3)",
        [proof.uid, linkUserId, proof.phone],
      );
      await db.query(
        "INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'identity.phone_linked',$1)",
        [linkUserId],
      );
    }
    const match = await db.query<{ user_id: string }>(
      "SELECT f.user_id FROM firebase_identity f JOIN staff_access a ON a.user_id=f.user_id WHERE f.firebase_uid=$1 AND f.phone=$2 AND a.active",
      [proof.uid, proof.phone],
    );
    if (!match.rows[0]) throw new Error("Identity is not linked");
    await db.query("COMMIT");
    return match.rows[0].user_id;
  } catch {
    await db.query("ROLLBACK");
    throw new APIError("UNAUTHORIZED", {
      message:
        "Unable to use this phone identity. Sign in with email to link it, or contact your administrator.",
    });
  } finally {
    db.release();
  }
}

export const firebasePhone = () =>
  ({
    id: "firebase-phone",
    endpoints: {
      linkFirebasePhone: createAuthEndpoint(
        "/firebase/link",
        { method: "POST", body, use: [sessionMiddleware] },
        async (ctx) => {
          if (
            Date.now() -
              new Date(ctx.context.session.session.createdAt).getTime() >
            5 * 60 * 1000
          )
            throw new APIError("UNAUTHORIZED", {
              message: "Sign in again before linking your phone.",
            });
          await consume(ctx.body.idToken, ctx.context.session.user.id);
          return ctx.json({ linked: true });
        },
      ),
      signInFirebasePhone: createAuthEndpoint(
        "/firebase/sign-in",
        { method: "POST", body },
        async (ctx) => {
          const userId = await consume(ctx.body.idToken);
          const user = await ctx.context.internalAdapter.findUserById(userId);
          if (!user) throw new APIError("UNAUTHORIZED");
          const session =
            await ctx.context.internalAdapter.createSession(userId);
          if (!session) throw new APIError("INTERNAL_SERVER_ERROR");
          await setSessionCookie(ctx, { session, user });
          return ctx.json({ success: true });
        },
      ),
    },
  }) satisfies BetterAuthPlugin;
