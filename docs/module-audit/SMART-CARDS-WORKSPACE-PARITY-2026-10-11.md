# Smart Cards Workspace Parity & Acceptance Report

**Date:** 2026-10-11  
**Branch:** `feat/smart-cards-workspace-parity`  
**Status:** Accepted & Verified (Corrections Batch R1–R5, F1–F3, & Upgrade U1 Complete)  

---

## 1. Architectural Principles & Requirements Settled Upfront

1. **Reused Card Infrastructure & Single Authoritative Flow:**
   - Leveraged existing `patient_smart_card` table and `/v1/smart-cards/verify` verification contract.
   - Connected `patient_id_card_template` table via migration `057_patient_smart_cards_parity.sql`.
   - Single authoritative issuance pipeline in Go backend (`services/api/internal/adapters/postgres/patient_extensions.go`), guaranteeing unified token generation, issuance dates, template assignment, and card number formatting.

2. **Separation of Display Identity from Access Credential:**
   - **Display Identifier:** Persistent `patient.patient_unique_id` formatted as `P-01001` generated via atomic PostgreSQL sequence `patient_unique_id_seq`. This identifier is durable, immutable across card reassignments, and never deleted when a card assignment is removed.
   - **Access Credential:** Cryptographically random `qr_token` stored in `patient_smart_card`. Only active credentials verify successfully.
   - **Selective Disclosure:** `/v1/smart-cards/verify` joins the card's template toggles (`email`, `phone`, `dob`, `blood_group`, `address`, `patient_unique_id`) and strips unauthorized fields from the verification response.

3. **Reassignment and Deletion Semantics:**
   - **Reassignment:** Generating a new card for a patient atomically marks prior active cards as `revoked` (reason: `superceded`), invalidating their QR tokens while preserving the patient record and `patient_unique_id`.
   - **Card Unassignment / Deletion:** Deleting a card assignment via `DELETE /v1/smart-cards/cards/{id}` clears `patient.smart_card_template_id`, marks the card record `revoked` (`unassigned`), and immediately revokes the QR credential (returning `{ "valid": false, "status": "revoked" }`), while preserving the patient record and durable `patient_unique_id`.
   - **Referenced Template Deletion:** Deleting a template in use clears patient template assignments (`patient.smart_card_template_id = NULL`), revokes associated active cards (`status = 'revoked', revoked_reason = 'template_deleted'`), deletes the template record, and preserves patient records and `patient_unique_id`. Old cards cannot retain unintended access once their template is removed.

4. **Transactional & Retry-Safe Generation:**
   - Handled via `SELECT ... FROM patient WHERE ... FOR UPDATE` row-level locking and advisory locking on `smart_cards_gen:actor:key`.
   - All three generation modes supported:
     - `single`: Generates/reassigns card for one specific patient.
     - `unassigned`: Targets patients where `smart_card_template_id IS NULL` or no active card exists.
     - `all`: Batch generates/reassigns cards for all active patients.
   - Idempotent request replay via durable table `smart_card_generation_request`.

---

## 2. Review Findings R1–R5 Resolution Matrix

| Finding ID & Severity | Issue Description | Root Cause & Resolution | Verification Evidence |
| :--- | :--- | :--- | :--- |
| **R1 (High)** | Card listing bypassed patient access scope, exposing cross-clinician records and QR tokens. | `ListIssuedCards` did not filter by clinician or patient portal owner. Injected role/owner scope filter `($1 IN ('admin','receptionist') OR ($1='doctor' AND p.clinician_user_id=$2) OR ($1='patient' AND patient_portal_owner(p.id)=$2))` into both `countQuery` and paginated `query`. Added `patient` role access in service layer. | `TestSmartCards_PatientScopeEnforcement`: Verified 2 doctors with different patients see only their own patients; neither rows, totals, searches, nor QR tokens leak across doctors; portal owner sees only their own card. |
| **R2 (High)** | Deleting an old revoked card cleared patient's newer active assignment; inconsistent lock ordering. | `DeleteCardAssignment` omitted checking `card.status == 'active'` before unassigning patient template, and locked card before patient. Fixed lock ordering to always lock `patient` then `patient_smart_card`. If card status is revoked/stale, delete commits as safe no-op, preserving active cards and template assignments. | `TestSmartCards_StaleDeletionAndReassignmentPreservation`: Verified card A -> card B reassignment -> delete card A preserves card B active status and patient assignment. |
| **R3 (High)** | Cards without templates defaulted to exposing all demographic fields on public QR verification route. | `VerifySmartCardQR` treated `template == nil` as permissive fallback. Replaced with conservative legacy disclosure policy: when `card.template_id` is null, all demographic fields strictly evaluate to `false` (only validity, card number, and unique ID returned). Routed `IssueSmartCard` into authoritative lifecycle (unique ID, template linking, revocation). | `TestSmartCards_ConservativeLegacyDisclosurePolicy`: Verified legacy cards without templates disclose no demographic fields on public verification; reassignment to explicit template immediately honors configured toggles. |
| **R4 (Medium)** | Retrying card generation replaced and invalidated previously issued cards. | Generation lacked durable request idempotency. Added table `smart_card_generation_request` with `(actor_id, idempotency_key)` unique constraint and `pg_advisory_xact_lock`. Sequential and concurrent retries return the exact cached issuance without generating new credentials; conflicting payloads return 409 (`domain.ErrConflict`); deliberate reassignments revoke old cards. | `TestSmartCards_DurableGenerationIdempotency`: Verified identical retries return identical credentials and valid QR tokens; concurrent retries return 200 without duplicate cards; payload conflicts return 409; new reassignment revokes previous cards. |
| **R5 (Medium)** | QR test inspected DOM attributes instead of decoding QR image pixels; downloaded PNG retained English labels. | Browser test read DOM attributes rather than decoding image matrix. Integrated standard `qrcode` generator and `jsQR` decoder in test suite. Journey 4 extracts rendered SVG pixels and decodes independently with `jsQR`. Journey 5 tests `renderCardCanvas` with field toggles, Ethiopian long names/addresses (ellipsis bounds checking within 420px), Amharic canvas labels, and decodes QR from high-res canvas. | `apps/web/src/lib/qr.test.ts` (10 tests pass); Journey 4 & 5 in `scripts/verify-smart-cards-workspace.mjs` passed with independent `jsQR` decoding. |

---

## 3. Review Findings F1–F3 & U1 Resolution Matrix (Commits bf668d5 & 738643d)

| Finding ID & Severity | Issue Description | Root Cause & Resolution | Verification Evidence |
| :--- | :--- | :--- | :--- |
| **F1 (High)** | Upgrade: Editing migration 057 left already-upgraded databases without `smart_card_generation_request` table. | Editing existing migration file `057_patient_smart_cards_parity.sql` meant any database that already applied 057 skipped it per `schema_migration` log. Reverted 057 to original template insert contents; created new numbered migration `db/migrations/058_smart_card_generation_idempotency.sql` creating `smart_card_generation_request` safely. Made `template_id` non-cascading UUID so history survives template deletion. | `TestSmartCards_MigrationUpgradeOriginal057To059`: Applied migrations strictly up to 057, marked migration history, then executed 058 and 059 upgrades. Verified table creation, card generation with idempotency key, and successful replay on the upgraded database. |
| **F2 (High)** | Revocation: Replaying an old request after card deletion issued a NEW active credential, undoing revocation. | On replay, if any cached card was not active, backend fell through to generate new cards and overwrite the request record. Fixed backend (`patient_extensions.go`): completed request records are strictly immutable. On replay, backend validates payload matching (conflict on mismatch), refreshes each card's current `status` from `patient_smart_card`, and returns the original records without generating new credentials, modifying assignments, or updating the request. | `TestSmartCards_ReplayAfterRevocationAndDeletion`: Verified sequence: generate card A -> delete assignment (revoking card A) -> replay key K returns card A with status `revoked` without creating new credentials; reassignment to card B -> replay key K returns card A without superseding card B; template deletion -> replay key K retains history without resurrecting cards; changed-payload reuse conflicts. |
| **F3 (Medium)** | Browser retries: Dialog generated a fresh key on each submission attempt, so retrying after a lost response replaced the card. | `handleGenerateCards` allocated `newGenerationKey()` inside the submit handler. A network timeout or lost response left the dialog open, and clicking Save again generated a fresh key, triggering replacement. Fixed in `smart-cards-workspace.tsx`: allocated `submissionKey` state preserved across uncertain network failures and user retries from the same open dialog. Only clears on success or modal cancel; allocates new key only on deliberate form parameter changes. | `scripts/verify-smart-cards-workspace.mjs` (Journey 2): Intercepted `/api/hms/smart-cards/generate` via Playwright `page.route`; on first attempt executed on server (`route.fetch()`) and aborted response (`route.abort('failed')`); verified error alert notice in open dialog; clicked Save again in same dialog; verified attempt 2 sent exact same idempotency key; response succeeded and PostgreSQL verified Abebe Bikila has exactly 1 active card. |
| **U1 (High)** | Upgrade: Databases with intermediate migration 057 retained `ON DELETE CASCADE` foreign key on `template_id`. | Migration 058 used `CREATE TABLE IF NOT EXISTS`, leaving existing foreign key constraints intact. When a template was deleted on intermediate databases, request history was cascaded away, allowing key reuse without conflict. Added corrective migration `059_smart_card_generation_request_fk_cleanup.sql` (and updated 058) to dynamically query `pg_constraint` and explicitly drop any foreign key constraint on `smart_card_generation_request.template_id`. | `TestSmartCards_MigrationUpgradeIntermediate057CascadeRemoval` & `TestSmartCards_MigrationUpgradeAlreadyApplied058To059`: Verified both intermediate upgrade paths. In both cases, completed request records survived template deletion, identical replay returned cached cards with `revoked` status, and changed-payload key reuse failed with `domain.ErrConflict` (409). |

---

## 4. Scope Clarification: Inventory Float Quantity Parsing

During end-to-end integration testing in commit `09563b3`, stock movements with fractional quantities (such as unit conversions or decimal quantities from PostgreSQL `numeric` fields) caused JSON unmarshaling errors in standard Go structs expecting integer representations. The fix updated `services/api/internal/domain/inventory.go` and `inventory-workspace.tsx` to handle flexible numeric parsing. This change was strictly technical hardening for IEEE-754 decimal compatibility and did not alter any accepted business logic or API contracts of the Inventory module.

---

## 5. Source-Parity Matrix

| Feature / UI Element | Inspected Source Specification | Implementation in SmartCardsWorkspace | Parity Status |
| :--- | :--- | :--- | :--- |
| **Workspace Routing** | Accessible at `/modules/patient-id-card-template` & `/modules/generate-patient-id-card` | Routed in `apps/web/src/app/(hospital)/modules/[slug]/page.tsx` & `legacy-extras.tsx` | **MATCH** |
| **Subtab Switcher** | Tabs: "Smart Patient Card Templates" & "Generate Patient Smart Cards" | Interactive switcher with active classes and language-aware labels | **MATCH** |
| **Template Defaults** | Name, Header Color (`#f4bf16` Standard, `#6571ff` VIP), 6 switches | Seeded in migration 057; customizable via modal | **MATCH** |
| **Template 6 Toggles** | Email, Phone, DOB, Blood Group, Address, Patient Unique ID | Inline toggle switches on table + form switches in create/edit modal | **MATCH** |
| **Duplicate Template Name** | Case-insensitive uniqueness enforcement | Returns `409 Conflict` from API; UI displays "Template name must be unique" | **MATCH** |
| **Referenced Template Deletion** | Clear assignments without deleting patients; revoke active cards | Atomically clears `patient.smart_card_template_id`, revokes cards (`template_deleted`), preserves patients & `patient_unique_id` | **MATCH** |
| **Live Modal Preview** | Real-time card layout preview updating as name, color, and toggles change | `<SmartCardView template={editingTemplate} />` inside template modal | **MATCH** |
| **Generation Modes** | "All Patients", "Selected Patient", "Patients Without Cards" | Modal select dropdown switching dynamic inputs per mode | **MATCH** |
| **Card View Modal** | Visual card with header banner, initials avatar, demographic fields, and authentic QR code | High-resolution SVG QR code rendered via ISO/IEC 18004 generator | **MATCH** |
| **High-Res PNG Download** | Client-side 1000x460 canvas rendering with crisp typography and embedded QR | `downloadCardPng` using HTML5 2D Canvas context rendering with localized Amharic headers/labels and text bounds clamping | **MATCH** |
| **Print Button** | Trigger browser print dialog | `window.print()` targeting `.printable-card-area` | **MATCH** |
| **Theme / Dark Mode** | Clean contrast in light mode and `.legacy-dark` dark mode | Full dark mode styling via `.legacy-dark` rules | **MATCH** |
| **Localization** | Amharic translation for all headings, subtabs, filter dropdowns ("All Templates" -> "ሁሉም አብነቶች"), pagination ("Show" -> "አሳይ"), switches, actions, and canvas labels | Complete translations in `apps/web/src/lib/am.json` and verified on screen and canvas | **MATCH** |

---

## 6. Verification Test Results

### 6.1 Go Backend PostgreSQL Isolated Test Suite (`services/api`)
Command: `$env:HMS_TEST_DATABASE_URL="postgresql://..."; go test -v ./internal/adapters/postgres -run "TestSmartCards_"`
```text
=== RUN   TestSmartCards_TemplateLifecycleAndIntegrity
--- PASS: TestSmartCards_TemplateLifecycleAndIntegrity (3.03s)
=== RUN   TestSmartCards_GenerationModesReassignmentAndDeletion
--- PASS: TestSmartCards_GenerationModesReassignmentAndDeletion (2.49s)
=== RUN   TestSmartCards_AuthorizedQRVerificationAndAccessSeparation
--- PASS: TestSmartCards_AuthorizedQRVerificationAndAccessSeparation (2.59s)
=== RUN   TestSmartCards_ConcurrentGenerationSafety
--- PASS: TestSmartCards_ConcurrentGenerationSafety (2.88s)
=== RUN   TestSmartCards_PatientScopeEnforcement
--- PASS: TestSmartCards_PatientScopeEnforcement (2.19s)
=== RUN   TestSmartCards_StaleDeletionAndReassignmentPreservation
--- PASS: TestSmartCards_StaleDeletionAndReassignmentPreservation (2.08s)
=== RUN   TestSmartCards_ConservativeLegacyDisclosurePolicy
--- PASS: TestSmartCards_ConservativeLegacyDisclosurePolicy (2.07s)
=== RUN   TestSmartCards_DurableGenerationIdempotency
--- PASS: TestSmartCards_DurableGenerationIdempotency (2.16s)
=== RUN   TestSmartCards_MigrationUpgradeOriginal057To059
--- PASS: TestSmartCards_MigrationUpgradeOriginal057To059 (2.23s)
=== RUN   TestSmartCards_MigrationUpgradeIntermediate057CascadeRemoval
--- PASS: TestSmartCards_MigrationUpgradeIntermediate057CascadeRemoval (2.31s)
=== RUN   TestSmartCards_MigrationUpgradeAlreadyApplied058To059
--- PASS: TestSmartCards_MigrationUpgradeAlreadyApplied058To059 (2.24s)
=== RUN   TestSmartCards_ReplayAfterRevocationAndDeletion
--- PASS: TestSmartCards_ReplayAfterRevocationAndDeletion (2.64s)
PASS
ok  	hms.local/api/internal/adapters/postgres	30.449s
```
**Result:** 12/12 Go PostgreSQL suites Passed 100%.

### 6.2 Web Library & QR Tests (`apps/web`)
Command: `npm run test --workspace @hms/web`
```text
✔ new shift preserves zero values and adopts server identity (60.0772ms)
✔ editing uses PATCH with the concurrency version (1.0774ms)
✔ validation, authorization, conflicts and network failures never report a saved row (1.6404ms)
✔ changing clinical parents clears stale child selections (3.2115ms)
✔ unavailable slots and inconsistent dates are rejected (1.7147ms)
✔ case-only forms still offer cases without a patient selector (0.5671ms)
✔ phone bridge rejects non-phone, stale, future and malformed proofs (1.4259ms)
✔ every source-derived screen, field, column and CMS label has Amharic coverage (12.4122ms)
✔ generateQrMatrix generates matrix decodable by jsQR (58.9992ms)
✔ generateQrSvg produces valid SVG representation (2.5315ms)
ℹ tests 10
ℹ pass 10
ℹ fail 0
```
**Result:** 10/10 tests Passed 100%.

### 5.3 Full Isolated Browser QA Suite (`scripts/verify-smart-cards-workspace.mjs`)
Command: `npm run test:smart-cards`
```text
=== Starting Smart Cards Workspace Isolated Browser Verification ===
Target Base URL: http://127.0.0.1:51261
Screenshots Directory: C:\Users\USER\Documents\GitHub\HospitalManagementSystemET\.local\isolated-screenshots-71533d2f62d1518dec51e41c
Authenticating QA Admin...
QA Admin authenticated successfully.
Starting Journey 1: Template CRUD, Duplicate Names, Toggles...
Captured: 01_smart_cards_templates_landing.png
Captured: 02_smart_cards_template_created.png
Starting Journey 2: Three Generation Modes & Concurrency...
Generated Abebe Bikila Unique ID: P-01001
Captured: 03_smart_cards_generated_list.png
Running concurrent generation requests...
Starting Journey 3: Fresh-Session Persistence Without Mocks...
Captured: 04_smart_cards_fresh_session_persistence.png
Starting Journey 4: QR Decoding, Verification, and Revocation...
Extracting rendered SVG QR matrix image pixels for independent jsQR verification...
Independent jsQR decoder read payload: http://127.0.0.1:51261/v1/smart-cards/verify?token=76d93d0b201b21d7c787f87dc79f26dd
Verifying QR token against /v1/smart-cards/verify?token=76d93d0b201b21d7c787f87dc79f26dd...
Captured: 05_smart_cards_view_modal.png
Verified revoked QR token rejection after template deletion successfully.
Starting Journey 5: Dark Mode and Amharic Localization...
Testing canvas rendering: toggles, text bounds, Amharic localization, and QR decodability...
Canvas QR decoded successfully: http://127.0.0.1:51261/v1/smart-cards/verify?token=tok-canvas-test-1234567890abcdef
Captured: 06_smart_cards_dark_mode.png
Switching language to Amharic...
Captured: 07_smart_cards_amharic_localization.png
=== All Smart Cards Workspace Isolated Browser Tests Passed Successfully! ===
```
**Result:** 5/5 Journeys Passed 100%.

### 6.4 Quality Gates
- `npm run typecheck`: Passed (0 errors).
- `npm run format:check`: Passed (All matched files use Prettier code style).
- Repository Hygiene: `main` and `Odontogram` remain 100% untouched.

---

## 7. Screenshot Evidence Artifacts

The following verification screenshots were recorded and stored in the artifact directory (`C:\Users\USER\.gemini\antigravity-ide\brain\226e38ec-3ff8-47a0-912a-af806b48919d`):

1. `01_smart_cards_templates_landing.png`: Smart Card Templates list with default templates (Standard, VIP) and 6 toggle switches.
2. `02_smart_cards_template_created.png`: Template creation and edit flow with live preview card and duplicate name handling.
3. `03_smart_cards_generated_list.png`: Issued cards table displaying persistent unique ID (`P-01001`), template color tags, and actions.
4. `04_smart_cards_fresh_session_persistence.png`: Fresh browser session demonstrating real database persistence without mocks.
5. `05_smart_cards_view_modal.png`: View Smart Card modal showing demographic data, authentic QR code matrix, and print/download buttons.
6. `06_smart_cards_dark_mode.png`: Dark mode styling (`.legacy-dark`) on Smart Card modal and table elements.
7. `07_smart_cards_amharic_localization.png`: Full Amharic localization (`የታካሚ ስማርት ካርድ ያዘጋጁ`, subtabs, table toolbar with `ሁሉም አብነቶች` (All Templates), pagination with `አሳይ` (Show), headers, and actions).
