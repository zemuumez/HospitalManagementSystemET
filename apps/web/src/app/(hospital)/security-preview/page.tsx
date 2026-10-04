"use client";
import { useState } from "react";
import { useLanguage } from "@/components/language";
export default function SecurityPreview() {
  const { t } = useLanguage();
  const [stage, setStage] = useState("setup");
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [generation, setGeneration] = useState(1);
  const codes = Array.from(
    { length: 6 },
    (_, i) => `DEMO-${generation}-${String(i + 1).padStart(4, "0")}`,
  );
  function download() {
    const url = URL.createObjectURL(
      new Blob(
        ["FRONTEND DEMO ONLY — NOT VALID RECOVERY CODES\n" + codes.join("\n")],
        { type: "text/plain" },
      ),
    );
    const a = document.createElement("a");
    a.href = url;
    a.download = "demo-recovery-codes.txt";
    a.click();
    URL.revokeObjectURL(url);
  }
  return (
    <section className="legacy-card">
      <h1>{t("Two-Factor Authentication")}</h1>
      <p className="form-preview-note">
        {t(
          "Frontend demonstration only. This does not enable or disable account security.",
        )}
      </p>
      {stage === "enabled" ? (
        <>
          <h2>{t("Recovery Codes")}</h2>
          <p>{t("These sample codes cannot sign you in.")}</p>
          <ul className="my-5">
            {codes.map((c) => (
              <li key={c}>
                <code>{c}</code>
              </li>
            ))}
          </ul>
          <div className="flex flex-wrap gap-3">
            <button className="primary" onClick={download}>
              {t("Download Codes")}
            </button>
            <button
              className="secondary"
              onClick={() => setGeneration(generation + 1)}
            >
              {t("Regenerate Codes")}
            </button>
            <button
              className="secondary"
              onClick={() => {
                setCode("");
                setStage("disable");
              }}
            >
              {t("Disable Two-Factor Authentication")}
            </button>
          </div>
        </>
      ) : (
        <form
          className="max-w-lg mt-5"
          onSubmit={(e) => {
            e.preventDefault();
            if (code !== "123456") {
              setError("Incorrect demo code. Use 123456.");
              return;
            }
            setStage(stage === "disable" ? "setup" : "enabled");
            setCode("");
            setError("");
          }}
        >
          <h2>
            {t(
              stage === "disable"
                ? "Disable Two-Factor Authentication"
                : "Enable Two-Factor Authentication",
            )}
          </h2>
          {stage === "setup" && (
            <div className="my-5">
              <p>
                {t(
                  "Authenticator QR enrollment will be connected during backend integration.",
                )}
              </p>
              <p>
                {t("Demo verification code")}: <code>123456</code>
              </p>
            </div>
          )}
          <label>
            <span className="label">{t("One-Time Password")}</span>
            <input
              className="field"
              inputMode="numeric"
              pattern="[0-9]{6}"
              maxLength={6}
              required
              value={code}
              onChange={(e) => setCode(e.target.value)}
            />
          </label>
          {error && (
            <p role="alert" className="error">
              {t(error)}
            </p>
          )}
          <button className="primary mt-5">
            {t(stage === "disable" ? "Confirm Disable" : "Verify and Enable")}
          </button>
        </form>
      )}
    </section>
  );
}
