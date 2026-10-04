import { applicationDefault, getApps, initializeApp } from "firebase-admin/app";
import { getAuth } from "firebase-admin/auth";

export function assertPhoneProof(
  claims: {
    phone_number?: string;
    auth_time: number;
    firebase: { sign_in_provider: string };
  },
  now = Date.now() / 1000,
) {
  if (
    claims.firebase.sign_in_provider !== "phone" ||
    !claims.phone_number ||
    !/^\+[1-9]\d{7,14}$/.test(claims.phone_number)
  )
    throw new Error("Phone authentication required");
  if (now - claims.auth_time > 300 || claims.auth_time > now + 30)
    throw new Error("Fresh phone authentication required");
  return claims.phone_number;
}

export async function verifyFirebasePhone(idToken: string) {
  if (!process.env.FIREBASE_PROJECT_ID)
    throw new Error("Firebase is not configured");
  if (
    process.env.NODE_ENV === "production" &&
    process.env.FIREBASE_AUTH_EMULATOR_HOST
  )
    throw new Error("Firebase emulator is prohibited in production");
  const app =
    getApps()[0] ??
    initializeApp({
      projectId: process.env.FIREBASE_PROJECT_ID,
      ...(process.env.FIREBASE_AUTH_EMULATOR_HOST
        ? {}
        : { credential: applicationDefault() }),
    });
  const claims = await getAuth(app).verifyIdToken(idToken, true);
  const phone = assertPhoneProof(claims);
  return { uid: claims.uid, phone, expires: new Date(claims.exp * 1000) };
}
