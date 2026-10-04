"use client";
import { useState } from "react";
import Link from "next/link";
import { authClient } from "@/lib/auth-client";
export default function ResetPassword() {
  const [message, setMessage] = useState("");
  const [done, setDone] = useState(false);
  const [busy, setBusy] = useState(false);
  return (
    <main className="flex min-h-screen items-center justify-center p-6">
      <form
        className="card w-full max-w-md space-y-5 p-8"
        onSubmit={async (e) => {
          e.preventDefault();
          setBusy(true);
          try {
            const password = new FormData(e.currentTarget).get(
              "password",
            ) as string;
            const token = new URLSearchParams(window.location.search).get(
              "token",
            );
            if (!token) throw new Error("Invalid reset link");
            const result = await authClient.resetPassword({
              newPassword: password,
              token,
            });
            if (result.error) throw new Error(result.error.message);
            setDone(true);
            setMessage("Password updated. You can now sign in.");
          } catch (err) {
            setMessage(
              err instanceof Error ? err.message : "Unable to reset password",
            );
          } finally {
            setBusy(false);
          }
        }}
      >
        <h1 className="text-2xl font-semibold">Choose a new password</h1>
        {message && (
          <p role="status" className={done ? "success" : "error"}>
            {message}
          </p>
        )}
        {!done && (
          <>
            <label className="label" htmlFor="new-password">
              New password · at least 12 characters
            </label>
            <input
              id="new-password"
              className="field"
              name="password"
              type="password"
              minLength={12}
              required
              autoComplete="new-password"
            />
            <button className="primary w-full" disabled={busy}>
              Update password
            </button>
          </>
        )}
        <Link className="block text-sm text-brand" href="/login">
          Return to sign in
        </Link>
      </form>
    </main>
  );
}
