"use client";
import Link from "next/link";
import { LanguageSwitcher, useLanguage } from "@/components/language";
import { useState } from "react";
import { HeartPulse } from "lucide-react";
import { PhoneAuth } from "@/components/phone-auth";
import { authClient } from "@/lib/auth-client";
export default function Login() {
  const { t } = useLanguage();
  const [tab, setTab] = useState("email");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    setNotice("");
    const form = new FormData(e.currentTarget);
    try {
      const email = String(form.get("email"));
      if (tab === "reset") {
        const result = await authClient.requestPasswordReset({
          email,
          redirectTo: "/reset-password",
        });
        if (result.error) throw new Error(result.error.message);
        setNotice(
          "If this email has an account, a reset link will be sent to it.",
        );
      } else {
        const result = await authClient.signIn.email({
          email,
          password: String(form.get("password")),
          rememberMe: form.get("remember") === "on",
        });
        if (result.error)
          throw new Error(
            "Unable to sign in. Check your credentials or contact your administrator.",
          );
        window.location.assign("/dashboard");
      }
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Unable to connect. Please try again.",
      );
    } finally {
      setBusy(false);
    }
  }
  return (
    <main className="legacy-login">
      <div className="login-language">
        <Link href="/">{t("Front Site")}</Link>
        <LanguageSwitcher />
      </div>
      <div className="login-brand">
        <HeartPulse size={42} color="#6571ff" />
        <strong>ULSHMS</strong>
      </div>
      <section className="login-panel">
        <h1>{t(tab === "reset" ? "Forgot Password" : "Sign In")}</h1>
        {tab !== "reset" && (
          <div className="detail-tabs mb-6">
            {["email", "phone"].map((method) => (
              <button
                key={method}
                className={tab === method ? "active" : ""}
                onClick={() => {
                  setTab(method);
                  setError("");
                }}
              >
                {method === "email" ? t("Email & password") : t("Phone number")}
              </button>
            ))}
          </div>
        )}
        {error && (
          <p className="error mb-4" role="alert">
            {t(error)}
          </p>
        )}
        {notice && (
          <p className="success mb-4" role="status">
            {t(notice)}
          </p>
        )}
        {tab === "phone" ? (
          <PhoneAuth />
        ) : (
          <form onSubmit={submit} className="space-y-6">
            <label className="block">
              <span className="label">
                {t("Email address")} <b className="text-red-500">*</b>
              </span>
              <input
                name="email"
                type="email"
                autoComplete="username"
                className="field"
                placeholder={t("Email")}
                required
              />
            </label>
            {tab === "email" && (
              <div>
                <div className="flex justify-between">
                  <label className="label" htmlFor="password">
                    {t("Password")} <b className="text-red-500">*</b>
                  </label>
                  <button
                    type="button"
                    className="text-brand text-xs"
                    onClick={() => setTab("reset")}
                  >
                    {t("Forgot password?")}
                  </button>
                </div>
                <input
                  id="password"
                  name="password"
                  type="password"
                  autoComplete="current-password"
                  className="field"
                  placeholder={t("Password")}
                  required
                />
              </div>
            )}
            {tab === "email" && (
              <label className="remember-me">
                <input name="remember" type="checkbox" defaultChecked />{" "}
                {t("Remember Me")}
              </label>
            )}
            <button className="primary w-full" disabled={busy}>
              {busy
                ? "Please wait…"
                : tab === "reset"
                  ? "Send reset link"
                  : t("Sign In")}
            </button>
            {tab === "reset" && (
              <button
                type="button"
                className="text-brand"
                onClick={() => setTab("email")}
              >
                {t("Back to sign in")}
              </button>
            )}
          </form>
        )}
        <p className="mt-6 text-xs text-muted">
          {t("Need an account? Contact your hospital administrator.")}
        </p>
      </section>
      <p className="text-xs text-muted mt-8">
        {t("Hospital Management System")}
      </p>
    </main>
  );
}
