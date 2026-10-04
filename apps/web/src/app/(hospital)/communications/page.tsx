"use client";
import { useLanguage } from "@/components/language";
import { useEffect, useState } from "react";
import { Mail, Send } from "lucide-react";
import { api, type Message } from "@/lib/api";
export default function Communications() {
  const { t } = useLanguage();

  const [messages, setMessages] = useState<Message[]>([]);
  const [channel, setChannel] = useState("email");
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState(false);
  const [key, setKey] = useState("");
  function load() {
    api<{ messages: Message[] }>("messages")
      .then((r) => setMessages(r.messages))
      .catch((e) => setError(e.message));
  }
  useEffect(() => {
    load();
    setKey(crypto.randomUUID());
  }, []);
  return (
    <div className="space-y-6">
      <div>
        <p className="eyebrow text-brand">{t("Stay connected")}</p>
        <h1 className="mt-2 text-[28px] font-semibold tracking-tight">
          {t("Communications")}
        </h1>
        <p className="mt-2 text-sm text-muted">
          {t("Send a clear, timely message to the people who need it.")}
        </p>
      </div>
      {error && (
        <p className="error" role="alert">
          {t(error)}
        </p>
      )}
      {notice && (
        <p className="success" role="status">
          {t(notice)}
        </p>
      )}
      <div className="grid items-start gap-6 xl:grid-cols-[1fr_1.1fr]">
        <form
          className="card space-y-5 p-6"
          onSubmit={async (e) => {
            e.preventDefault();
            const form = e.currentTarget;
            setBusy(true);
            setError("");
            setNotice("");
            try {
              await api("messages", {
                method: "POST",
                headers: { "Idempotency-Key": key },
                body: JSON.stringify({
                  ...Object.fromEntries(new FormData(form)),
                  channel,
                }),
              });
              setNotice("Message queued. Check delivery status below.");
              setKey(crypto.randomUUID());
              form.reset();
              load();
            } catch (err) {
              setError(
                err instanceof Error ? err.message : "Unable to queue message",
              );
            } finally {
              setBusy(false);
            }
          }}
        >
          <h2 className="flex items-center gap-2 font-semibold">
            <Mail size={18} className="text-brand" />
            {t("New message")}
          </h2>
          <div>
            <label htmlFor="channel" className="label">
              {t("Send via")}
            </label>
            <select
              className="field"
              id="channel"
              value={channel}
              onChange={(e) => setChannel(e.target.value)}
            >
              <option value="email">{t("Email")}</option>
              <option value="sms">{t("SMS")}</option>
            </select>
          </div>
          <div>
            <label htmlFor="recipient" className="label">
              {channel === "sms"
                ? "Phone number with country code"
                : t("Email address")}
            </label>
            <input
              className="field"
              id="recipient"
              name="recipient"
              type={channel === "sms" ? "tel" : "email"}
              required
            />
          </div>
          <div>
            <label className="label" htmlFor="subject">
              {t("Subject")}
              {channel === "sms" && "· internal reference"}
            </label>
            <input
              className="field"
              id="subject"
              name="subject"
              required={channel === "email"}
              maxLength={200}
            />
          </div>
          <div>
            <label className="label" htmlFor="body">
              {t("Message")}
            </label>
            <textarea
              className="field min-h-32"
              id="body"
              name="body"
              required
              maxLength={channel === "sms" ? 480 : 4000}
            />
          </div>
          <p className="text-xs leading-5 text-muted">
            {t(
              "Keep messages brief. Use the patient portal for private clinical information.",
            )}
          </p>
          <button className="primary" disabled={busy || !key}>
            <Send size={15} />
            {busy ? "Queuing…" : "Queue message"}
          </button>
        </form>
        <section className="card overflow-hidden">
          <div className="flex items-center justify-between border-b border-slate-100 p-5">
            <h2 className="font-semibold">{t("Recent messages")}</h2>
            <button className="text-xs text-brand" onClick={load}>
              {t("Refresh status")}
            </button>
          </div>
          {!messages.length ? (
            <p className="px-6 py-16 text-center text-sm text-muted">
              {t("Your sent and queued messages will appear here.")}
            </p>
          ) : (
            <ul className="divide-y divide-slate-100">
              {messages.map((m) => (
                <li
                  className="flex items-center justify-between gap-4 p-5"
                  key={m.id}
                >
                  <div className="min-w-0">
                    <p className="truncate text-xs font-medium">
                      {m.recipient}
                    </p>
                    <p className="mt-1 text-[10px] text-muted">
                      {m.channel.toUpperCase()} ·{" "}
                      {new Date(m.createdAt).toLocaleString()}
                    </p>
                  </div>
                  <span className="rounded-full bg-[#eef4f1] px-3 py-1 text-[10px] capitalize text-brand">
                    {m.status}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
  );
}
