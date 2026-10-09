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

1. **Upload (`POST /v1/attachments`)**:
   - Supports system assets with `isPublic=true` and optional `patientId` (permitted for actors with `settings.manage` or role `admin`).
   - Mime type detection: `image/png`, `image/jpeg`, `image/webp`.
   - Size limit: 25 MiB.
   - Generates non-predictable token and persists in `secure_attachment`.
2. **Download (`GET /v1/attachments/{token}/content`)**:
   - When `is_public` is true, returns the file stream immediately without requiring patient ownership or care-team assignment, allowing browsers to render logos and favicons.
3. **Replacement & Removal**:
   - Setting `logo_url` or `favicon_url` to the new attachment content URL updates `hospital_general_setting`.
   - Setting to `""` or clicking Remove clears the setting reference.

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
