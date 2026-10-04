import nodemailer from "nodemailer";
export async function sendAuthMail(to: string, subject: string, text: string) {
  const host = process.env.SMTP_HOST;
  if (!host) throw new Error("SMTP_HOST is required");
  const transport = nodemailer.createTransport({
    host,
    port: Number(process.env.SMTP_PORT ?? 1025),
    secure: process.env.SMTP_SECURE === "true",
    requireTLS: process.env.NODE_ENV === "production",
    auth: process.env.SMTP_USER
      ? { user: process.env.SMTP_USER, pass: process.env.SMTP_PASSWORD }
      : undefined,
    connectionTimeout: 5000,
    socketTimeout: 10000,
  });
  await transport.sendMail({
    from: process.env.MAIL_FROM ?? "ULSHMS <noreply@hms.local>",
    to,
    subject,
    text,
  });
}
