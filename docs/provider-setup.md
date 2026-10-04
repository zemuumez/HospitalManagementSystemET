# Provider credentials and hospital setup

Production values are intentionally blank in [deployment/.env.example](../deployment/.env.example). This is a worksheet, not the development environment. Do not replace the working `.env.local` or commit a filled copy. Put actual values in the deployment platform's secret manager/environment settings; store service-account JSON outside the repository. No account, purchase, real SMS, payment or external deployment was performed for this guide.

## 1. What you need to provide

| Item | Who supplies it | Where it is used |
|---|---|---|
| Hospital HTTPS domain | Your domain/hosting account | BETTER_AUTH_URL, Firebase authorized domains |
| PostgreSQL connection | Database host administrator | DATABASE_URL in web/API/worker |
| Session signing secret | Generate locally | BETTER_AUTH_SECRET in web only |
| Firebase project/web app and server identity | Your Firebase/Google Cloud account | Phone-login proof verification |
| SMS-enabled sender and account credentials | Your SMS provider account | Operational messaging worker |
| Verified email sender and SMTP credentials | Your email provider account | Password reset and operational email |
| Payment provider merchant eligibility and credentials | Hospital's approved merchant account | Online payment adapter, when implemented |
| Hospital hours, tax rules, currency and clinical approval rules | Hospital administrator/accountant/clinical lead | Configuration and acceptance review |
| Production host, backup target, retention and operator contacts | Your infrastructure administrator | Deployment and recovery |

The Go API and PostgreSQL should be private services. Expose the HTTPS Next.js site through your reverse proxy. GO_API_URL is the private API address reachable from Next.js, not a public browser URL. Use separate development/staging/production databases and credentials.

Generate a signing secret locally with `node -e "console.log(require('node:crypto').randomBytes(32).toString('hex'))"`. Copy its output into your secret manager, not into a commit or chat. Rotation invalidates existing sessions unless a deliberate migration is implemented.

## 2. Firebase phone authentication

1. Create/select a project in the [Firebase console](https://console.firebase.google.com/), then register a Web app in Project settings.
2. Copy the Web app's apiKey, projectId and authDomain into the matching NEXT_PUBLIC_FIREBASE_* fields. Set FIREBASE_PROJECT_ID to the same project ID.
3. Open Authentication, enable Phone sign-in, configure your production domain under authorized domains, and allow only the SMS destination regions needed by your hospital.
4. Configure test phone numbers first. Confirm project billing/quota requirements in the console before enabling real SMS. Keep reCAPTCHA enabled for the deployed app.
5. Follow the [official phone-auth setup](https://firebase.google.com/docs/auth/web/phone-auth); provider UI labels may change. Firebase supplies login OTPs; it does not send operational hospital messages through this application's Twilio adapter.

For server verification, prefer a workload identity/Application Default Credentials supported by your host. Otherwise, Project settings > Service accounts provides the Admin SDK credential setup. Mount the private JSON on the web server and set GOOGLE_APPLICATION_CREDENTIALS to its absolute path. Never place this file in public assets or NEXT_PUBLIC_* variables. See [Admin SDK credentials](https://firebase.google.com/docs/admin/setup).

Production must omit both emulator variables entirely. Rebuild Next.js after changing NEXT_PUBLIC_* fields; restart the web service after changing server credentials. Test with an existing hospital account: sign in by email, link the verified phone, sign out, then phone-sign-in. Confirm an unlinked phone cannot take over an existing account. These successful live-provider checks remain pending until you configure your project.

## 3. Operational SMS with Twilio

1. Create/select a Twilio account and obtain its Account SID and Auth Token from the console.
2. Obtain an SMS-capable sender that is supported for your destination countries. Check provider country restrictions and sender registration requirements before selecting it.
3. Fill TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN and TWILIO_FROM. The current adapter uses a From sender, not a Messaging Service SID.
4. Set SMS_PROVIDER=twilio only in the operational worker environment after credentials are ready. Restart the worker.
5. Use one approved test recipient and inspect the provider message log. A locally recorded `sent` currently means provider acceptance, not confirmed handset delivery. Delivery callbacks/reconciliation remain separate checklist work.

Use the [Twilio quickstart](https://www.twilio.com/docs/messaging/quickstart) and [Message resource documentation](https://www.twilio.com/docs/messaging/api/message-resource). Development keeps SMS_PROVIDER=capture; no real message is sent. Production rejects capture mode. Blank/invalid Twilio credentials fail closed.

## 4. Email and Mailpit

Development already uses Mailpit: SMTP 127.0.0.1:1025 and inbox http://127.0.0.1:8025. It needs no external account and must not be exposed publicly.

For production:

1. Choose your transactional-email provider and verify the hospital sender/domain using the DNS records supplied by that provider.
2. Create SMTP credentials in that provider's dashboard. Use SMTP credentials, not an unrelated API token unless the provider explicitly identifies it as the SMTP password.
3. Fill SMTP_HOST, SMTP_USER and SMTP_PASSWORD. Use a STARTTLS endpoint on port 587 so both the web mailer and Go worker support it; leave SMTP_SECURE=false for STARTTLS.
4. Set MAIL_FROM to the web sender (display name allowed), and SMTP_FROM to the worker's plain email address. They must be authorized senders.
5. Restart web and worker. Request a password reset for your controlled test account and send one operational email to an approved recipient. Verify delivery and links use your HTTPS domain.

## 5. Payment credentials (preparation; adapter still pending)

Existing payment forms record money already received/refunded; they do not charge a card or move money. Blank reserved payment variables do not change that. Do not buy a provider plan merely to finish local development.

If Stripe is selected and supports the hospital's merchant jurisdiction:

1. Create a merchant account and sandbox. Complete the provider's requested business/bank verification before live use.
2. In the sandbox API-key page, copy the publishable key to NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY and secret key to STRIPE_SECRET_KEY. Never put the secret key in browser variables.
3. Once the application's signed webhook endpoint is implemented and documented, register its HTTPS URL in Stripe and copy that endpoint's signing secret to STRIPE_WEBHOOK_SECRET. Do not invent a callback URL or reuse an API key as a signing secret.
4. Use sandbox payments/webhook fixtures first. Live keys and live endpoint secrets are separate; activate only after amount/currency/replay/refund reconciliation tests pass.

See [Stripe keys](https://docs.stripe.com/keys) and [webhook setup](https://docs.stripe.com/webhooks). Other gateways require their own adapter, credential names and signature verification; credentials are not interchangeable. Provider choice remains configurable work, not an assumption of Ethiopian merchant support.

## 6. Values still intentionally unfilled

Keep all production secrets blank until you obtain them. You can continue using local email/password login, captured SMS, Mailpit and manual financial records. Provider setup is tracked separately from local backend implementation/testing; a blank worksheet is not proof of live delivery, real payment settlement or production readiness.

After configuring a provider, record only the environment, provider name, verification date and outcome in the deployment log. Never record secret values. Restrict and rotate compromised credentials in the provider console, update the secret manager and restart/rebuild the affected service.
