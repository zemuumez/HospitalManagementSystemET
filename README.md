# ULSHMS

A single-hospital management system being rebuilt with Next.js, Tailwind CSS and Go. Better Auth owns accounts/sessions; Firebase verifies phone identities; Go owns hospital authorization and workflows.

The current UI is an original-style, frontend-first preview with 111 module routes and synthetic data. See [frontend preview and parity limits](docs/frontend-preview.md). The connected foundation includes email sign-in/reset, Firebase phone linking/sign-in preparation, scoped patient registration/directory at `/live-patients`, and operational email/SMS at `/communications`. The other module backends remain pending. See [discovery](docs/discovery/README.md), [delivery plan](docs/discovery/rebuild-plan.md), and the superseding [authentication decision](docs/architecture/001-authentication.md).

## Local development

Requirements: Node 24+, Go 1.26+, PostgreSQL 18, Mailpit. Docker Compose is optional. Never copy credentials from the legacy archive.

1. `npm ci` at repository root.
2. Copy `apps/web/.env.example` to `apps/web/.env.local`; set a random 32+ character `BETTER_AUTH_SECRET` and a working `DATABASE_URL`.
3. Start PostgreSQL and Mailpit. For Docker, set `POSTGRES_PASSWORD` and run `docker compose -f infra/development/compose.yaml up -d`. The database is on localhost:5433; Mailpit SMTP on localhost:1025 and UI on localhost:8025. Match the database password in the web environment file.
4. Run `npm run db:migrate`. Migrations are transactional and tracked; application startup does not mutate the schema.
5. Set `ADMIN_EMAIL`, `ADMIN_PASSWORD` (12+ characters) and optionally `ADMIN_NAME`, then `npm run admin:create`. This creates a new account only; it never resets or replaces an existing admin. Clear provisioning variables afterward.
6. In `services/api`, set `DATABASE_URL` and `BETTER_AUTH_URL=http://127.0.0.1:3000`, then `go run ./cmd/api`. API defaults to localhost:8080.
7. In a second terminal in `services/api`, set the same database URL, `SMS_PROVIDER=capture`, `SMTP_HOST=127.0.0.1`, `SMTP_PORT=1025`, and run `go run ./cmd/worker`.
8. At repository root, `npm run dev`; open http://127.0.0.1:3000. Use that exact hostname for cookie/origin consistency.

This workspace also has a private `.local` PostgreSQL cluster and ignored `.env.local` generated for verification. Local admin credentials are in ignored `.local/admin-login.txt`. Do not share or commit that file. Local servers may need to be restarted between sessions; no system service was installed.

## Phone authentication

For development, start the Firebase Auth Emulator using the Firebase CLI: `firebase emulators:start --only auth --project demo-ulshms --config infra/development/firebase.json`. Set the Firebase variables from `.env.example` and restart Next.js. The emulator supplies test OTP codes; it sends no real SMS. Sign in with email, visit Account & security, and link a test phone within five minutes of signing in. Sign out, then use phone sign-in.

For production, configure the Firebase project/web app, authorized domains and phone provider, supply ADC for Firebase Admin, and remove all emulator variables. No clinical records are stored in Firebase. Phone binding is explicit; a phone number match alone never selects a hospital account. Firebase is optional until configured; there is no fake-auth fallback.

## Mail and operational SMS

Password reset mail and operational email use the configured SMTP server. In development, inspect mail at http://localhost:8025. Operational SMS defaults to `captured` in the outbox. Actual sending requires `SMS_PROVIDER=twilio`, `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, and `TWILIO_FROM`. Capture mode refuses production. No credentials have been provisioned for real SMS.

Keep private clinical information out of free-form SMS/email. Queue status `sent` indicates provider acceptance only. Inspect `failed`, `uncertain`, or stale `processing` rows before any resend; the worker deliberately avoids automatic redelivery after ambiguous failures. The current UI supports composing and observing messages, not bulk sends or scheduled reminders.

## Checks

```text
npm run typecheck
npm test
npm run test:integration
npm run build
cd services/api
go test ./...
go vet ./...
```

The integration check requires the local web, API, worker, database and Mailpit to be running. It creates and removes synthetic fixtures and queues messages to test recipients. Run it only with `SMS_PROVIDER=capture` on the development worker and SMTP pointed at Mailpit; do not run it against live delivery providers.

On restricted Windows hosts, set GOPATH/GOMODCACHE/GOCACHE and npm cache to writable workspace directories. Versions are locked in package-lock.json and go.sum. The root overrides pin patched gRPC and UUID transitive dependencies; re-evaluate them when updating Firebase. Clinical behavior parity is tracked separately from infrastructure tests.

## Repository boundaries

The project repository is https://github.com/zemuumez/HospitalManagementSystemET. Commit each completed, verified implementation step with a descriptive message and push it to this repository. The initial commit records the existing rebuild as a baseline; the parity audit in `docs/frontend-parity-audit-2026-10-04.md` tracks remaining frontend work. Keep credentials, local databases, generated output, and legacy reference archives out of Git.

`apps/web` contains the UI and Better Auth; `services/api/internal/domain` has policy/validation without transport/database imports; `application` coordinates use cases through interfaces; `adapters` implement HTTP, persistence and delivery. API and worker entry points live in `cmd`. `db/migrations` owns the schema. `review` and `tmp` contain ignored legacy/reference work, not the new runtime.

Do not treat this foundation as the finished HMS or deploy it with development settings. Build the remaining modules as vertical slices with their legacy workflow audit, negative authorization tests and clinical/financial acceptance examples.
