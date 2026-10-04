# Continuous verification

`.github/workflows/verify.yml` runs on main pushes, pull requests and manual dispatch. It has read-only repository permissions and does not receive production secrets or deploy the application. PostgreSQL and Mailpit are disposable service containers; the password in the workflow belongs only to that isolated CI database. Authentication secrets are generated per run. SMS stays in capture mode.

Checks: fresh migrations, uncached Go unit/database/HTTP tests, Go vet and API/worker builds; web tests, formatting, build/typecheck; production-dependency npm audit; live authentication/API integration; existing connected browser journeys with Playwright Chromium. The database tests create/drop only their generated schemas. Browser fixtures use the disposable CI database. No Firebase project, SMS account or payment credentials are needed.

Node 24 and the Go version in go.mod are installed with the official [setup-node](https://github.com/actions/setup-node) and [setup-go](https://github.com/actions/setup-go) actions; [checkout](https://github.com/actions/checkout) does not persist credentials. Action major tags and container tags should be SHA/digest pinned as part of production supply-chain hardening. A workflow file being present does not prove a hosted run passed; inspect the repository Actions result before making it a required branch check.

Local equivalent commands and fixture boundaries are in the delivery checklist. Test:connected is existing UI regression coverage, not implementation of Section 4. Full per-module browser journeys remain to be added as those integrations are implemented. Firebase emulator success-path and dependency remediation are separate pending checks; npm audit may correctly fail when a vulnerability requires fixing.

First hosted execution: [run 37225297032](https://github.com/zemuumez/HospitalManagementSystemET/actions/runs/37225297032) passed all steps for commit `3eed632`.

## Firebase emulator and dependency gate

Run `npm run test:firebase` with the ordinary local DATABASE_URL available in the ignored `.env.local`. The wrapper creates its own random database schema, copies the web app, allocates separate loopback web/API/Auth-emulator ports, uses a demo project and a private CLI config directory, and stops its processes/drops its schema afterward. It does not read production Firebase credentials or send SMS. OTP, explicit-link concurrency, Better Auth session issuance, Go session acceptance, replay/revocation/disablement and linking freshness are verified. Only synthetic rate-limit fixtures are reset between additional negative cases.

Firebase CLI 15.32.1 is a pinned development dependency. Its scoped overrides use basic-ftp 6.2.2, OpenTelemetry core 2.11.0 and Chokidar 4.0.3 to eliminate reported advisories; the Auth emulator suite is tested with that set. Chokidar's newer API has no glob support, so do not assume this validates unrelated Firebase CLI emulator/watch commands. Production Firebase SDKs are not replaced by that scoped override. CI now runs `npm audit --audit-level=high` across all dependencies. The full local audit returned zero vulnerabilities after the change.
