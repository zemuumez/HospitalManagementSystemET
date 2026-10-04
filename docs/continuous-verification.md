# Continuous verification

`.github/workflows/verify.yml` runs on main pushes, pull requests and manual dispatch. It has read-only repository permissions and does not receive production secrets or deploy the application. PostgreSQL and Mailpit are disposable service containers; the password in the workflow belongs only to that isolated CI database. Authentication secrets are generated per run. SMS stays in capture mode.

Checks: fresh migrations, uncached Go unit/database/HTTP tests, Go vet and API/worker builds; web tests, formatting, build/typecheck; production-dependency npm audit; live authentication/API integration; existing connected browser journeys with Playwright Chromium. The database tests create/drop only their generated schemas. Browser fixtures use the disposable CI database. No Firebase project, SMS account or payment credentials are needed.

Node 24 and the Go version in go.mod are installed with the official [setup-node](https://github.com/actions/setup-node) and [setup-go](https://github.com/actions/setup-go) actions; [checkout](https://github.com/actions/checkout) does not persist credentials. Action major tags and container tags should be SHA/digest pinned as part of production supply-chain hardening. A workflow file being present does not prove a hosted run passed; inspect the repository Actions result before making it a required branch check.

Local equivalent commands and fixture boundaries are in the delivery checklist. Test:connected is existing UI regression coverage, not implementation of Section 4. Full per-module browser journeys remain to be added as those integrations are implemented. Firebase emulator success-path and dependency remediation are separate pending checks; npm audit may correctly fail when a vulnerability requires fixing.
