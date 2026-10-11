# Smart Cards Workspace Parity & Acceptance Report

**Date:** 2026-10-11  
**Branch:** `feat/smart-cards-workspace-parity`  
**Commit SHA:** `55016ed07413f3cfe5b81f0bf3c599d63ab9c1c9`  
**Status:** Accepted & Verified  

---

## 1. Architectural Principles & Requirements Settled Upfront

1. **Reused Card Infrastructure & Single Authoritative Flow:**
   - Leveraged existing `patient_smart_card` table and `/v1/smart-cards/verify` verification contract.
   - Connected `patient_id_card_template` table via migration `057_patient_smart_cards_parity.sql`.
   - Single authoritative issuance pipeline in Go backend (`services/api/internal/adapters/postgres/patient_extensions.go`), guaranteeing unified token generation, issuance dates, and card number formatting.

2. **Separation of Display Identity from Access Credential:**
   - **Display Identifier:** Persistent `patient.patient_unique_id` formatted as `P-01001` generated via atomic PostgreSQL sequence `patient_unique_id_seq`. This identifier is durable, immutable across card reassignments, and never deleted when a card assignment is removed.
   - **Access Credential:** Cryptographically random `qr_token` stored in `patient_smart_card`. Only active credentials verify successfully.
   - **Selective Disclosure:** `/v1/smart-cards/verify` joins the card's template toggles (`email`, `phone`, `dob`, `blood_group`, `address`, `patient_unique_id`) and strips unauthorized fields from the verification response.

3. **Reassignment and Deletion Semantics:**
   - **Reassignment:** Generating a new card for a patient atomically marks prior active cards as `revoked` (reason: `superceded`), invalidating their QR tokens while preserving the patient record and `patient_unique_id`.
   - **Unassignment / Deletion:** Deleting a card assignment via `DELETE /v1/smart-cards/cards/{id}` clears `patient.smart_card_template_id`, marks the card record `revoked` (`unassigned`), and immediately revokes the QR credential (returning `{ "valid": false, "status": "revoked" }`).
   - **Referenced Template Deletion Protection:** Deleting a template referenced by any patient or active card returns `domain.ErrInUse`, mapped to HTTP `409 Conflict` (`RECORD_IN_USE`).

4. **Transactional & Retry-Safe Generation:**
   - Handled via `SELECT ... FROM patient WHERE ... FOR UPDATE` row-level locking.
   - All three generation modes supported:
     - `single`: Generates/reassigns card for one specific patient.
     - `unassigned`: Targets patients where `smart_card_template_id IS NULL` or no active card exists.
     - `all`: Batch generates/reassigns cards for all active patients.
   - Concurrent requests serialize cleanly under PostgreSQL row locks and sequence counters without duplicate keys or race conditions.

---

## 2. Source-Parity Matrix

| Feature / UI Element | Inspected Source Specification | Implementation in SmartCardsWorkspace | Parity Status |
| :--- | :--- | :--- | :--- |
| **Workspace Routing** | Accessible at `/modules/patient-id-card-template` & `/modules/generate-patient-id-card` | Routed in `apps/web/src/app/(hospital)/modules/[slug]/page.tsx` & `legacy-extras.tsx` | **MATCH** |
| **Subtab Switcher** | Tabs: "Smart Patient Card Templates" & "Generate Patient Smart Cards" | Interactive switcher with active classes and language-aware labels | **MATCH** |
| **Template Defaults** | Name, Header Color (`#f4bf16` Standard, `#6571ff` VIP), 6 switches | Seeded in migration 057; customizable via modal | **MATCH** |
| **Template 6 Toggles** | Email, Phone, DOB, Blood Group, Address, Patient Unique ID | Inline toggle switches on table + form switches in create/edit modal | **MATCH** |
| **Duplicate Template Name** | Case-insensitive uniqueness enforcement | Returns `409 Conflict` from API; UI displays "Template name must be unique" | **MATCH** |
| **Live Modal Preview** | Real-time card layout preview updating as name, color, and toggles change | `<SmartCardView template={editingTemplate} />` inside template modal | **MATCH** |
| **Generation Modes** | "All Patients", "Selected Patient", "Patients Without Cards" | Modal select dropdown switching dynamic inputs per mode | **MATCH** |
| **Card View Modal** | Visual card with header banner, initials avatar, demographic fields, and authentic QR code | High-resolution SVG QR code rendered via pure ISO/IEC 18004 generator | **MATCH** |
| **High-Res PNG Download** | Client-side 1000x460 canvas rendering with crisp typography and embedded QR | `downloadCardPng` using HTML5 2D Canvas context rendering | **MATCH** |
| **Print Button** | Trigger browser print dialog | `window.print()` targeting `.printable-card-area` | **MATCH** |
| **Theme / Dark Mode** | Clean contrast in light mode and `.legacy-dark` dark mode | Full dark mode styling via `.legacy-dark` rules | **MATCH** |
| **Localization** | Amharic translation for all headings, switches, modes, and actions | Complete translations in `apps/web/src/lib/am.json` | **MATCH** |

---

## 3. Verification Test Results

### 3.1 Go Backend PostgreSQL Isolated Test Suite (`services/api`)
Command: `$env:HMS_TEST_DATABASE_URL="postgresql://..."; go test ./...`
```text
ok      hms.local/api/internal/adapters/delivery      (cached)
ok      hms.local/api/internal/adapters/httpapi       (cached)
ok      hms.local/api/internal/adapters/postgres      29.254s
  === RUN   TestSmartCards_TemplateLifecycleAndIntegrity
  --- PASS: TestSmartCards_TemplateLifecycleAndIntegrity (2.12s)
  === RUN   TestSmartCards_GenerationModesReassignmentAndDeletion
  --- PASS: TestSmartCards_GenerationModesReassignmentAndDeletion (2.14s)
  === RUN   TestSmartCards_AuthorizedQRVerificationAndAccessSeparation
  --- PASS: TestSmartCards_AuthorizedQRVerificationAndAccessSeparation (2.17s)
  === RUN   TestSmartCards_ConcurrentGenerationSafety
  --- PASS: TestSmartCards_ConcurrentGenerationSafety (2.49s)
ok      hms.local/api/internal/adapters/privatefiles  (cached)
ok      hms.local/api/internal/adapters/stripe        (cached)
ok      hms.local/api/internal/domain                 (cached)
```
**Result:** 100% Passed.

### 3.2 Full Isolated Browser QA Suite (`scripts/verify-smart-cards-workspace.mjs`)
Command: `npm run test:smart-cards`
```text
=== Starting Smart Cards Workspace Isolated Browser Verification ===
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
Verifying QR token against /v1/smart-cards/verify?token=5cba4557d1bf01bded29b1b470f1e2e3...
Captured: 05_smart_cards_view_modal.png
Verified revoked QR token rejection successfully.
Starting Journey 5: Dark Mode and Amharic Localization...
Captured: 06_smart_cards_dark_mode.png
Switching language to Amharic...
Captured: 07_smart_cards_amharic_localization.png
=== All Smart Cards Workspace Isolated Browser Tests Passed Successfully! ===
```
**Result:** 5/5 Journeys Passed 100%.

### 3.3 Quality Gates
- `npm run typecheck`: Passed (0 errors).
- `npm run format:check`: Passed (All matched files use Prettier code style).
- Repository Hygiene: `main` and `Odontogram` remain 100% untouched.

---

## 4. Screenshot Evidence Artifacts

The following verification screenshots were recorded and stored in the artifact directory (`C:\Users\USER\.gemini\antigravity-ide\brain\226e38ec-3ff8-47a0-912a-af806b48919d`):

1. `01_smart_cards_templates_landing.png`: Smart Card Templates list with default templates (Standard, VIP) and 6 toggle switches.
2. `02_smart_cards_template_created.png`: Template creation and edit flow with live preview card and duplicate name handling.
3. `03_smart_cards_generated_list.png`: Issued cards table displaying persistent unique ID (`P-01001`), template color tags, and actions.
4. `04_smart_cards_fresh_session_persistence.png`: Fresh browser session demonstrating real database persistence without mocks.
5. `05_smart_cards_view_modal.png`: View Smart Card modal showing demographic data, authentic QR code matrix, and print/download buttons.
6. `06_smart_cards_dark_mode.png`: Dark mode styling (`.legacy-dark`) on Smart Card modal and table elements.
7. `07_smart_cards_amharic_localization.png`: Full Amharic localization (`የታካሚ ስማርት ካርድ ያዘጋጁ`, tabs, buttons, and switches).
