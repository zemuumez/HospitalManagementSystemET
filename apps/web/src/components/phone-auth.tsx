"use client";
import { useRef, useState } from "react";
import { initializeApp, getApps } from "firebase/app";
import {
  getAuth,
  connectAuthEmulator,
  RecaptchaVerifier,
  signInWithPhoneNumber,
  signOut,
  type ConfirmationResult,
} from "firebase/auth";
let emulatorConnected = false;
export function PhoneAuth({ link = false }: { link?: boolean }) {
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [confirmation, setConfirmation] = useState<ConfirmationResult>();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const verifier = useRef<RecaptchaVerifier | null>(null);
  if (!process.env.NEXT_PUBLIC_FIREBASE_API_KEY)
    return (
      <p className="rounded-lg bg-slate-50 p-4 text-sm text-muted">
        Phone sign-in will be available after your administrator connects
        Firebase. You can sign in with your email.
      </p>
    );
  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError("");
    const app =
      getApps()[0] ??
      initializeApp({
        apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
        projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID,
        authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
      });
    const firebase = getAuth(app);
    try {
      if (
        process.env.NEXT_PUBLIC_FIREBASE_AUTH_EMULATOR_URL &&
        !emulatorConnected
      ) {
        if (process.env.NODE_ENV === "production")
          throw new Error(
            "Development phone authentication cannot run in production.",
          );
        connectAuthEmulator(
          firebase,
          process.env.NEXT_PUBLIC_FIREBASE_AUTH_EMULATOR_URL,
        );
        emulatorConnected = true;
      }
      if (!confirmation) {
        verifier.current ??= new RecaptchaVerifier(
          firebase,
          "phone-recaptcha",
          { size: "invisible" },
        );
        setConfirmation(
          await signInWithPhoneNumber(firebase, phone, verifier.current),
        );
      } else {
        const credential = await confirmation.confirm(code);
        const response = await fetch(
          `/api/auth/firebase/${link ? "link" : "sign-in"}`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              idToken: await credential.user.getIdToken(),
            }),
          },
        );
        const result = await response.json();
        if (!response.ok)
          throw new Error(result.message ?? "Phone verification failed");
        if (link) {
          setSuccess(true);
          setConfirmation(undefined);
        } else window.location.assign("/dashboard");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to verify phone");
      verifier.current?.clear();
      verifier.current = null;
    } finally {
      if (confirmation) await signOut(firebase);
      setBusy(false);
    }
  }
  if (success)
    return (
      <p className="success" role="status">
        Phone linked. You can now use it to sign in.
      </p>
    );
  return (
    <form onSubmit={submit} className="space-y-4">
      <p className="text-sm leading-6 text-muted">
        {link
          ? "Link a phone you own. For your security, sign in again if your session is more than five minutes old."
          : "Use a phone already linked to your hospital account. First time? Sign in with email and link your phone in Account."}
      </p>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <div>
        <label className="label" htmlFor="phone">
          Phone number
        </label>
        <input
          id="phone"
          className="field"
          type="tel"
          autoComplete="tel"
          required
          pattern="\+[1-9][0-9]{7,14}"
          placeholder="+254 700 000 000"
          value={phone}
          disabled={!!confirmation}
          onChange={(e) => setPhone(e.target.value.replace(/\s/g, ""))}
        />
      </div>
      {confirmation && (
        <div>
          <label className="label" htmlFor="otp">
            Verification code
          </label>
          <input
            id="otp"
            className="field"
            inputMode="numeric"
            autoComplete="one-time-code"
            required
            pattern="[0-9]{6}"
            value={code}
            onChange={(e) => setCode(e.target.value)}
          />
        </div>
      )}
      <div id="phone-recaptcha" />
      <button className="primary w-full" disabled={busy}>
        {busy
          ? "Please wait…"
          : confirmation
            ? "Verify phone"
            : "Send verification code"}
      </button>
    </form>
  );
}
