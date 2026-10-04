import { betterAuth } from "better-auth";
import { pool } from "./db";
import { sendAuthMail } from "./mail";
import { firebasePhone } from "./firebase-plugin";
import { hospitalMfa } from "./mfa";

const baseURL = process.env.BETTER_AUTH_URL ?? "http://127.0.0.1:3000";
if (
  !process.env.BETTER_AUTH_SECRET ||
  process.env.BETTER_AUTH_SECRET.length < 32
)
  throw new Error(
    "BETTER_AUTH_SECRET must contain at least 32 random characters",
  );
if (
  process.env.NODE_ENV === "production" &&
  process.env.FIREBASE_AUTH_EMULATOR_HOST
)
  throw new Error("Remove Firebase emulator settings in production");
export const auth = betterAuth({
  appName: "ULSHMS",
  baseURL,
  secret: process.env.BETTER_AUTH_SECRET,
  database: pool,
  trustedOrigins: [baseURL],
  emailAndPassword: {
    enabled: true,
    disableSignUp: true,
    minPasswordLength: 12,
    revokeSessionsOnPasswordReset: true,
    sendResetPassword: async ({ user, url }) =>
      sendAuthMail(
        user.email,
        "Reset your ULSHMS password",
        `Hello ${user.name},\n\nReset your password: ${url}\n\nIf you did not request this, ignore this email.`,
      ),
  },
  session: {
    expiresIn: 60 * 60 * 8,
    updateAge: 60 * 30,
    cookieCache: { enabled: false },
  },
  rateLimit: {
    enabled: true,
    storage: "database",
    window: 60,
    max: 100,
    customRules: {
      "/get-session": false,
      "/sign-in/email": { window: 60, max: 5 },
      "/request-password-reset": { window: 60, max: 3 },
      "/firebase/*": { window: 60, max: 5 },
    },
  },
  account: { accountLinking: { enabled: false } },
  databaseHooks: {
    session: {
      create: {
        before: async (session) => {
          const result = await pool.query(
            "SELECT user_id FROM staff_access WHERE user_id=$1 AND active",
            [session.userId],
          );
          return result.rowCount ? { data: session } : false;
        },
      },
    },
  },
  plugins: [firebasePhone(), hospitalMfa()],
});
