# Settings Workspace Inventory, Contracts, and Gap Analysis

Date: 2026-10-09  
Branch: `feat/settings-workspace-parity`  
Reference: `docs/builder-handoff/IMPLEMENTATION-PLAN.md` (Phase 1, Item 3), `docs/module-audit/REPAIRS.md`, `docs/builder-handoff/WORKSPACE-FIELDS.md`

---

## 1. Executive Summary

This inventory maps every Settings and Front CMS tab, field, action, and access policy from the original application and current repository baseline. It defines concrete integration obligations to eliminate preview saves (`localStorage`/`sessionStorage`), wire real atomic persistence, protect provider secrets against browser leakage, and integrate authorized attachment storage for logos and favicons.

---

## 2. Inventory of Settings Tabs & Actions

| Tab / Slug | Title | Original Route / View | Database Table | Existing API Endpoints | Current Status / Blocking Gap |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **`settings`** | General Settings | `settings.edit`<br>`settings/general.blade.php` | `hospital_general_setting` | `GET /v1/general-settings`<br>`POST /v1/general-settings` | **Gap**: `SettingsForm` in `legacy-extras.tsx` saves to `localStorage` with preview alert. Logo/favicon use inlined data URLs. Provider secrets (`open_ai_key`) returned plaintext or unhandled. |
| **`hospital-schedule`** | Hospital Schedule | `hospital-schedules.index`<br>`hospital_schedule/fields.blade.php` | `hospital_schedule_day` | `GET /v1/hospital-schedules`<br>`POST/PUT /v1/hospital-schedules` | **Gap**: `legacy-screen.tsx` uses `sessionStorage.getItem("hms-hospital-schedule")`. Backend exposes day-by-day endpoint. Needs live API persistence. |
| **`modules-setting`** | Modules Setting | `module.index`<br>`settings/modules.blade.php` | `hospital_module_setting` | `GET /v1/modules-setting`<br>`PUT /v1/modules-setting/{key}` | **Gap**: `ModulesSettings` in `legacy-extras.tsx` saves to `localStorage.getItem("hms-disabled-modules")`. Backend exists but frontend not connected. |
| **`currency-settings`** | Currencies | `currency-settings.index`<br>`currency_settings/fields.blade.php` | `hospital_general_setting` (`current_currency`) | `GET /v1/general-settings` | Active currency is managed via `current_currency`. |
| **`operation-categories`** | Operation Categories | `operation-categories.index`<br>`operation_categories/create_modal.blade.php` | `operation_category` | `GET/POST/DELETE /v1/operation-categories` | Exists in Go `services_operations.go`; proxied via `allowedRoot`. |
| **`operations`** | Operations | `operations.index`<br>`operations/create_modal.blade.php` | `hospital_operation` | `GET/POST/DELETE /v1/operations` | Exists in Go `services_operations.go`; proxied via `allowedRoot`. |
| **`payment-gateway`** | Payment Gateways | `payment-gateways.index`<br>`payment_gateway/create.blade.php` | `hospital_general_setting` | `GET/POST /v1/general-settings` | **Gap**: `legacy-screen.tsx` previews only. Contains high-risk provider secrets (`stripe_secret`, `paypal_secret`, `razorpay_secret`, `flutterwave_secret_key`, etc.) requiring strict redaction and preservation. |
| **`add-custom-fields`** | Custom Field | `custom-fields.index`<br>`add_custom_fields/create.blade.php` | `custom_field` | `GET/POST/DELETE /v1/custom-fields` | Exists in Go `services_operations.go`. |
| **`patient-queue-theme`** | Patient Queue Theme | `queue-theme-view`<br>`settings/queue_theme.blade.php` | `hospital_general_setting` (`queue_theme_*`) | `GET/POST /v1/general-settings` | **Gap**: `QueueTheme` in `legacy-extras.tsx` uses `localStorage.getItem("hms-queue-theme")`. Needs live API persistence. |
| **`front-settings`** | Front CMS Settings | `front.settings.index`<br>`front_settings/index.blade.php` | `front_cms_setting` | `GET/POST /v1/front-cms-settings` | **Gap**: `SettingsForm` in `legacy-extras.tsx` uses `localStorage`. Proxy route missing `front-cms-settings` in `allowedRoot`. |

---

## 3. Field Register & Validation Rules

### 3.1 General Settings (`settings`)
- **Required Fields**:
  - `app_name`: String (1–128 chars). Application name displayed across headers and metadata.
  - `company_name`: String (1–128 chars). Operating entity or hospital company name.
  - `hospital_email`: Email format (valid RFC syntax, e.g. `info@hospital.et`).
  - `hospital_phone`: Telephone string (min 7 chars, e.g. `+251911000000`).
  - `hospital_from_day` / `hospital_start_day`: Day index or string (e.g. `1` for Monday).
  - `hospital_from_time` / `hospital_start_time`: Time string (`HH:MM`).
  - `hospital_address`: Non-empty physical address string.
  - `current_currency` / `currency`: 3-character code (default `'ETB'`).
  - `about_us`: Non-empty description string.
- **Optional Fields** (Cleared with explicit empty string `""`):
  - `country_phone` / `country_code`: Calling code (e.g. `+251`).
  - `default_lang` / `default_language`: ISO code (`"en"`, `"am"`).
  - `facebook_url`: Social link or `""`.
  - `twitter_url`: Social link or `""`.
  - `instagram_url`: Social link or `""`.
  - `linkedIn_url`: Social link or `""`.
  - `custom_serial_prefix`: Prefix for auto-generated identifiers.
  - `model_name`: AI model identifier (e.g. `gpt-4o`).
- **File Attachments**:
  - `app_logo` (`logo_url`): Uploaded image token `/v1/attachments/{token}/content` or system fallback.
  - `favicon` (`favicon_url`): Uploaded favicon token `/v1/attachments/{token}/content` or fallback.
- **Provider Secrets**:
  - `open_ai_key`: Sensitive API credential.

### 3.2 Payment Gateways (`payment-gateway`)
Stored in `hospital_general_setting` with strict secret masking:
- **Public Identifiers** (Plaintext allowed):
  - `stripe_key`
  - `paypal_client_id`, `paypal_mode`
  - `razorpay_key`
  - `flutterwave_public_key`
  - `phonepe_merchant_id`, `phonepe_merchant_user_id`, `phonepe_env`, `phonepe_salt_index`, `phonepe_merchant_transaction_id`
  - `paystack_public_key`
- **Protected Secrets** (NEVER return plaintext to browser):
  - `stripe_secret`
  - `paypal_secret`
  - `razorpay_secret`
  - `flutterwave_secret_key`
  - `phonepe_salt_key`
  - `paystack_secret_key`

---

## 4. Provider Secret Protection Specification

1. **Secret Key Set**:
   `{ "open_ai_key", "stripe_secret", "paypal_secret", "razorpay_secret", "flutterwave_secret_key", "phonepe_salt_key", "paystack_secret_key" }`
2. **Read / Retrieval (`GET /v1/general-settings`)**:
   - If key does not exist or stored value is empty string `""`: return `""` (unconfigured credentials remain blank).
   - If key exists and stored value is non-empty: return `"[CONFIGURED]"`. Plaintext is **never** sent to the client.
3. **Write / Mutation (`POST /v1/general-settings`)**:
   - If value is `"[CONFIGURED]"` or `"********"`: **Unchanged**. The backend preserves the existing secret in PostgreSQL.
   - If value is `""`: **Explicit Clearing**. The backend updates the value to `""`.
   - If value is a new non-empty string (not equal to placeholder): **Replacement**. The backend updates the database with the new secret.

---

## 5. Attachment Storage Lifecycle for System Assets

### 5.1 Upload and Public Asset Access
1. **Upload (`POST /v1/attachments`)**:
   - Supports system assets with `isPublic=true` and empty `patientId` (permitted for actors with `settings.manage` or role `admin`).
   - Mime type detection: `image/png`, `image/jpeg`, `image/webp`.
   - Size limit: 25 MiB.
   - Computes SHA-256 hash and assigns `uploader_id`.
   - Generates non-predictable token and persists in `secure_attachment`.
2. **Download (`GET /v1/attachments/{token}/content`)**:
   - When `is_public` is true, returns the file stream immediately without requiring patient ownership or care-team assignment, allowing browsers to render logos and favicons.

### 5.2 Binding, Replacement & Retirement Protocol (R1a)
To eliminate races between concurrent settings saves and attachment retirements:
1. **Precise Reference Parsing**:
   - Attachment tokens are extracted only from genuine local attachment URL paths (`/(?:api/hms/|v1/)?attachments/{token}(?:/content)?`) or designated general asset fields (`app_logo`, `logo_url`, `favicon`, `favicon_url`).
   - Ordinary setting text, provider public keys, and remote external URLs containing 32-character hexadecimal identifiers are preserved without false-positive token classification.
2. **Alphabetical Lock Ordering Protocol**:
   - All referenced and displacing attachment tokens are collected and deduplicated.
   - Tokens are acquired and locked in stable alphabetical order (`ORDER BY token ASC`) inside a PostgreSQL transaction (`SELECT token, is_public, patient_id FROM secure_attachment WHERE token = ANY(...) FOR UPDATE`).
   - Newly bound tokens are validated: must exist, must have `is_public = true`, and must have empty `patient_id` (rejecting clinical records with `ErrValidation` / 422).
   - If any bound token does not exist in `secure_attachment`, the transaction immediately aborts with `ErrNotFound` (404), preventing dangling image URLs from ever persisting.
3. **Atomic Displacement & Reference Verification**:
   - Previous setting values are read under `FOR UPDATE`.
   - Setting rows are updated atomically within the transaction.
   - For any token displaced by the update, the transaction verifies that no active setting in either `hospital_general_setting` or `front_cms_setting` still references the token before marking it for retirement.
   - If no references remain, the attachment row is deleted from `secure_attachment` under the held lock, and the disk file removal is dispatched.
4. **Single-Item Method Coordination**:
   - Single-item persistence methods (`UpdateGeneralSetting`, `UpdateFrontCMSSetting`) delegate directly through the unified batch transaction methods to guarantee identical row-locking, validation, and displaced-token cleanup semantics.

### 5.3 Operational Abandoned Attachment Cleanup (R1b)
Assets uploaded without being saved into settings or displaced before completion are cleaned via bounded, observable operational jobs:
1. **Cleanup Semantics (`CleanupAbandonedAttachments`)**:
   - Targets only public, non-clinical attachments (`is_public = true AND (patient_id IS NULL OR patient_id = '')`).
   - Enforces an age threshold (`created_at < NOW() - older_than`, default 24h).
   - Pre-filters candidate tokens using `NOT EXISTS` across both `hospital_general_setting` and `front_cms_setting` before applying `LIMIT 100`, preventing referenced assets from permanently occupying candidate slots.
   - Bounded work per run: `LIMIT 100` with non-blocking concurrency: `ORDER BY sa.token ASC FOR UPDATE OF sa SKIP LOCKED`.
   - Under the held row lock, re-verifies reference status across settings tables before deleting records and dispatching disk file removals.
2. **Observable Error Handling & Audit Identity**:
   - Worker resolves active administrators from `staff_access` joined with `"user"` for audit event recording (`actor_id = admin.ID`), visibly logging and exiting on missing administrators.
   - Both operational cleanup file removal (`RunOperationalAttachmentCleanup`) and direct retirement (`RetireAttachment`) log disk removal failures using structured logging (`log.Printf`) rather than discarding errors.
3. **Operational Invocation Channels**:
   - **Administrative REST Endpoint**:
     `POST /v1/attachments/cleanup?older_than=24h`
     - Requires `settings.manage` permission (returns 403 Forbidden for non-administrators).
     - Returns `{ "cleaned": <count>, "threshold": "<duration>" }`.
   - **Worker CLI One-Shot Command**:
     `worker -cleanup-attachments [-older-than=24h]`
     - Executes a single bounded cleanup sweep against PostgreSQL and local storage, exiting with code 0 upon completion.
   - **Worker Scheduled Background Job**:
     - Continuous daemon background ticker running hourly operational cleanup with a 24-hour retention threshold.

---

## 6. Role Authorization Matrix

| Action / Resource | Admin | Doctor | Receptionist | Nurse | Lab Tech | Accountant | Patient | Anonymous |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `GET /v1/general-settings` | 200 (redacted) | 403 | 403 | 403 | 403 | 403 | 403 | 401 |
| `POST /v1/general-settings` | 200 (persisted) | 403 | 403 | 403 | 403 | 403 | 403 | 401 |
| `GET /v1/hospital-schedules` | 200 | 403 | 403 | 403 | 403 | 403 | 403 | 401 |
| `POST /v1/hospital-schedules` | 200 | 403 | 403 | 403 | 403 | 403 | 403 | 401 |
| `GET /v1/modules-setting` | 200 | 403 | 403 | 403 | 403 | 403 | 403 | 401 |
| `PUT /v1/modules-setting/{key}` | 200 | 403 | 403 | 403 | 403 | 403 | 403 | 401 |
| Public logo/favicon download | 200 | 200 | 200 | 200 | 200 | 200 | 200 | 200 |
