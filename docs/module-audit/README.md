# Module-level source, permission and integration audit

Date: 2026-10-06. This is an audit of the original Laravel source and the current implementation, performed **before further backend/schema implementation**.

## What this establishes

The database inventory is now supplemented by source action/validation/authorization evidence and executable tests. This separates four questions that earlier completion claims mixed together:

1. What does the original route/form/request/model/repository require?
2. What does the current backend authorize and persist?
3. Does the current frontend call that contract and reload the same data?
4. Has that specific workflow been exercised, including failure paths and other roles?

Source indexing is not full semantic review. Passing an existing Go test is not proof of legacy equivalence. Loading a page is not CRUD acceptance. A characterization test can successfully reproduce a **product gap**; therefore passing characterization assertions does not make parity pass.

## Evidence files

| File | Scope |
| --- | --- |
| [LEGACY-ACTIONS.md](LEGACY-ACTIONS.md) | 855 original route declarations, including resource modifiers and enclosing middleware/prefix groups |
| [LEGACY-WORKFLOWS.md](LEGACY-WORKFLOWS.md) | 463 request/controller/repository classes, 2,218 method signatures and guard/write/transaction source pointers |
| [MODEL-RULES.md](MODEL-RULES.md) | Static validation arrays and state constants for the 130 original models |
| [CURRENT-ROLE-POLICY.json](CURRENT-ROLE-POLICY.json) | Current Go named-permission role matrix plus permission-advertising inconsistencies; direct role checks and row scope are separate |
| [FRONTEND-INTEGRATION.md](FRONTEND-INTEGRATION.md) | API/storage/preview/failure source pointers across 38 components and hospital pages |
| [source-evidence.json](source-evidence.json) | Machine-readable source navigation and evidence |
| [ACCEPTANCE-MATRIX.md](ACCEPTANCE-MATRIX.md) | Module-specific acceptance obligations and known blockers |
| `test-results.json` | Sanitized executable audit results, when the run is finalized; no cookies, passwords or patient fixture payloads |
| `PAGE-RESULTS.md` | Final admin page/hydration/transport results, when the run is finalized; not visual/CRUD acceptance |

The original ZIP was inspected, not executed. PHP/Composer were not available on PATH in this session. Consequently there is no claim of a running Laravel fixture suite, live-demo mutation tests, or equivalence across every original method. Runtime tests use generated accounts/data in isolated PostgreSQL schemas and temporary Next/Go services; the development database is retained.

## Confirmed findings

### Smart cards: the screen is not connected to the original contract

Original: `PatientIdCardTemplateController` creates/updates persisted templates and clears patient template references on deletion. `GeneratePatientIdCardRepository::store` has three generation modes: all patients, one selected user-linked patient, or patients lacking templates; it assigns the template and generates patient identifiers. `PatientIdCardTemplate::$rules` requires a unique name and color.

Current: `SmartCards` in `legacy-extras.tsx` uses `hms-card-templates` and `hms-smart-cards` in sessionStorage. A browser test creates a template, reloads it successfully in the same tab, verifies zero HMS mutation requests, then finds it absent in a fresh browser context with the same admin identity. Same-tab reload alone would have falsely suggested database persistence.

Required audit acceptance: create/edit/delete/toggle template fields, unique-name validation, generation-mode selection, patient eligibility, stable/replaced identifiers as explicitly decided, cross-session persistence, preview/download/QR and unauthorized access. Existing secure card issuance is a different contract and must be bridged deliberately.

### Odontogram: chart and clinical code information is lost

Original: chart ID, patient, doctor, required description and tooth-state JSON; create and update are distinct chart actions, with PDF by chart ID. Multiple records per patient are permitted. `CreateOdontogramRequest::authorize` returns true, while the enclosing route group permits Admin/Patient/Doctor. That is legacy evidence, not an instruction to allow patients to edit arbitrary clinical records.

Current authenticated test: writing the same tooth twice returns the same record ID and leaves one row containing the latest condition. The original aggregate payload is rejected by the current tooth endpoint. Frontend conversion also changes `Ce -> K, D -> K, B -> C, and PS -> F` on save/read round trip. The string transformations were executed from the actual component source; they are not inferred from screenshots.

Required acceptance: multiple independent charts, chart-level doctor/description/history, lossless original codes, edited/cleared tooth persistence, patient/doctor ownership and print matching the selected chart. Decide safe patient access explicitly rather than cloning a potentially overbroad legacy resource route.

### Clinical detail tabs: routes and data shapes diverge

Current `modules/[slug]/page.tsx` routes IPD diagnoses, consultant registers, prescriptions, charges, payments, bills and timelines to the same `EncounterRegister kind="ipd"`. OPD diagnoses/timelines similarly reuse the OPD register. An HTTP 200 for those URLs does not verify their intended workflow.

Original IPD requires case/bed/doctor/admission; OPD requires doctor/date/standard charge/payment mode and does not require case. Original diagnosis uses report type/date/description/media, consultants use applied/instruction dates and text, prescriptions have grouped header/footer/items, and timelines have patient-visible flags. Those are not interchangeable with generic ICD diagnoses or care-team membership.

### Missing or mismatched integration routes and false local success

Authenticated probes observed 404 for `/api/hms/packages`, `/api/hms/insurances`, `/api/hms/general-settings`, and `/api/hms/blood-banks`. Distinguish causes:

- Packages/insurances: current ServicesWorkspace calls these endpoints, and on failed create it inserts an in-memory row and closes the form. Source handlers show this fallback. Original package/service and insurance/disease lines are transactional persisted aggregates.
- General settings: the Go handler exists at `/v1/general-settings`, but the Next proxy allowlist does not expose `general-settings`. Do not treat that as a missing Go repository.
- Blood bank: Go exposes singular `/v1/blood-bank`; the plural alias is not a working aggregate endpoint. Check the caller's actual path before labeling its whole module broken.

### Role policy requires endpoint and ownership review

The named permission matrix alone is insufficient. For example `blood_bank.read` is advertised for patients, but aggregate stock/donor reads explicitly reject patients; their own blood-issue access is separate. The initial test assumption was corrected against the application service. That restriction is not a failure to be "fixed" by widening access.

Conversely, `cms.read` / `cms.manage` exist in `Actor.Can` but are omitted by `Actor.Permissions`, while `messages.manage` is advertised twice. Record these as current contract inconsistencies.

Original IPD register/view routes include Nurse/Lab Technician, whereas current `clinical.read` does not. Original admission management includes Case Manager. Compare the actual scoped workspace needs before deciding which permissions to add. Do not grant broad access solely because a legacy route allowed it.

Seed permission arrays are also only one layer: `AssignDefaultRoleToUserSeeder` grants broad `manage_*` names to several roles, while route middleware, request authorization and repository ownership checks can narrow them. Effective access must be tested with populated owned and unowned records.

### Shared workspaces make invalid list requests

The browser audit observed HTTP 422 from `/api/hms/doctor-absences` on the doctors/departments workspace and `/api/hms/inventory/movements` on inventory tabs. Source tracing explains both: `doctors-workspace.tsx:398` omits `doctorId`, which `Scheduling.Absences` requires; `inventory-workspace.tsx:225` omits the UUID `itemId` required by `Inventory.Movements`. Decide whether the UI should select a parent or the backend should expose a separately authorized aggregate list. Removing required filters without reviewing scope would be an unsafe shortcut.

The first 112-page navigation run recorded 11 network-idle timeouts, all with these 422 responses or the package/insurance 404s. A timeout is not proof that a page cannot render. The final evidence retains timeouts separately from observed HTTP failures and explicitly waits for the authenticated workspace before declaring a page successful.

## Test scope and interpretation

- Existing full Go suite was run with `HMS_TEST_DATABASE_URL` set to loopback, exercising the isolated-schema PostgreSQL suite. 153 named test/subtest pass events, no test skip events; packages with no tests are not coverage.
- Existing frontend unit suite: eight tests passed. TypeScript check passed. These do not test every original form.
- New module API audit: 178 assertions passed and eight contract findings were recorded. It uses all nine roles across 16 collection probes (IPD and OPD separately), anonymous rejection, patient creation/reload, unauthorized/unrelated doctor dental writes, invalid input and persistence characterization.
- New browser card test distinguishes sessionStorage persistence from actual shared-account persistence.
- New catalog-page audit covered the dashboard plus all 111 source-catalog module URLs as admin: **101 initial-render checks passed; 11 failed with network-idle timeouts and observed 422/404 API responses**. The strengthened run reproduced the same totals. Successful checks waited for authenticated rendering; failed navigation checks never reached that readiness assertion. No page is declared CRUD-complete. See PAGE-RESULTS.md for every route and test-results.json for the exact observations. The audit correctly exited nonzero and cleaned its temporary schema/services.
- Early audit runs found harness assumptions as well as product issues: login rate limiting, required encounter kind, singular blood-bank route, and the explicit patient aggregate restriction. Corrections were made to the **tests**, not production permissions. Rate limits remain enabled and the tests wait/pace their requests.
- External Firebase/SMS/payment credentials remain blank/unmodified. No new live-provider delivery or financial mutation was attempted.

## Running the audit

```powershell
python scripts/audit-module-source.py 'C:/Users/USER/Downloads/Telegram Desktop/hms.zip'
```

With the normal local environment loaded securely into the process, run:

```text
node scripts/verify-connected-isolated.mjs --module-audit
node scripts/verify-connected-isolated.mjs --module-ui-audit
```

Do not paste environment values into logs. Both wrappers require loopback PostgreSQL, generate isolated schema names, seed synthetic records, start temporary services and remove their own schema/services on exit. The audit uses Chrome through the existing Playwright test dependency; set `HMS_CHROME_PATH` if needed. Temporary API/worker binaries now include the run token to avoid collisions between concurrent QA runs.

Raw run artifacts are under ignored `.local/module-audit/`. Commit only sanitized summaries. A nonzero page-audit exit identifies observed failures, not permission to suppress or relabel them. Do not disable security to obtain a green audit.

## Gate before backend implementation resumes

The audit already proves that "frontend looks the same" is insufficient and identifies actionable blockers. It does **not** establish full original parity or full understanding of every dynamic path. Before declaring a module understood and ready, its acceptance sheet must identify:

- original action, inputs, conditional fields, state/side effects, errors and print/export behavior;
- effective role AND record-ownership rules, with a safe target decision for legacy weaknesses;
- target database/API/UI mapping;
- executable success/failure/other-role/reload checks, or a concrete missing-contract failure;
- evidence and commit reference.

Use ACCEPTANCE-MATRIX.md to work through the remaining modules. Missing actions are audit findings, not completed implementations. Fixes should follow as separate reviewed, tested commits once this evidence is accepted.
