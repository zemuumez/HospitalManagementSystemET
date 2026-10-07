# Backend Packages Contract & Parity Specification

## 1. Overview & Baseline Comparison

This document establishes the authoritative contract for the **Hospital Packages** module, completing **Step 1: Package Backend and Contract** of the HMS parity roadmap.

### 1.1 Gap Analysis Table

| Domain Dimension | Legacy Source Evidence (`hms.sql` / Laravel) | Initial Target Baseline (`72c00a7`) | Target Implementation (`049_packages_catalog.sql`) |
|---|---|---|---|
| **Parent Fields** | `packages`: `id` (int unsigned), `name` (varchar 160, unique), `description` (text), `discount` (double, integer percentage in model rule), `total_amount` (double), `currency_symbol` (varchar 100). | None (table did not exist; frontend had hardcoded mock array `PKG-01`..`PKG-05` and local fallback saves). | `package`: `id` (uuid PK), `name` (varchar 160 NOT NULL UNIQUE), `description` (text NOT NULL DEFAULT ''), `discount` (integer NOT NULL CHECK(discount BETWEEN 0 AND 100)), `total_amount_minor` (bigint NOT NULL CHECK(>= 0)), `currency_symbol` (varchar 100 NOT NULL DEFAULT 'ETB'), `created_at`, `updated_at` (timestamptz). |
| **Child Fields** | `package_services`: `id` (bigint unsigned), `package_id` (FK packages), `service_id` (FK services), `quantity` (double, integer in model rule), `rate` (double), `amount` (double). | None (table did not exist). | `package_service`: `id` (uuid PK), `package_id` (uuid FK package ON DELETE CASCADE), `service_id` (uuid FK hospital_service ON DELETE RESTRICT), `quantity` (integer NOT NULL CHECK(> 0)), `rate_minor` (bigint NOT NULL CHECK(>= 0)), `amount_minor` (bigint NOT NULL CHECK(>= 0)), `created_at`, `updated_at` (timestamptz), UNIQUE(package_id, service_id). |
| **Validation & Totals** | Name unique, discount 0-100%, line amount = quantity × rate, subtotal = sum(amounts), total = subtotal - (subtotal × discount / 100). Submitted total was trusted in legacy controller. | Client calculated float amounts and submitted unvalidated payloads. | Authoritative server calculation in integer ETB minor units (cents): `amount_minor = quantity * rate_minor`, `subtotal = sum(amount_minor)`, `discount_amount = round((subtotal * discount + 50) / 100)`, `total_amount_minor = subtotal - discount_amount`. Client total amounts are ignored and computed authoritatively on server. |
| **Role Permissions & Scope** | Index/store/edit/update/destroy: `Admin\|Receptionist`. Show: `Admin\|Doctor\|Case Manager\|Patient\|Receptionist`. Export: `Admin\|Receptionist`. | No backend role checks for packages (returned 404). | `packages.manage`: `admin`, `receptionist` (create, update, delete). `packages.read`: `admin`, `receptionist`, `doctor`, `case_manager`, `patient` (list, show, export). Other roles (`nurse`, `accountant`, `lab_technician`) and anonymous users denied with 403 / 401. |
| **Endpoints** | Resource routes for `packages`, plus `export-packages`. | 404 on `/api/hms/packages`. | `GET /v1/packages`, `POST /v1/packages`, `GET /v1/packages/{id}`, `PUT /v1/packages/{id}`, `DELETE /v1/packages/{id}`, `GET /v1/packages-export`. |
| **Admission References & Deletion** | FK `patient_admissions.package_id`. Deletion blocked when referenced by admissions. | No admission FK to packages. `ipd_admission_details` had only free-text string snapshot `package_name`. | `ipd_admission_details.package_id` uuid REFERENCES `package(id)` ON DELETE RESTRICT. Application layer checks active admission linkage and returns readable 409 Conflict (`RECORD_IN_USE`: "Package is in use by patient admissions and cannot be deleted"). |
| **Testing Coverage** | No package tests in new repository. | 0 package tests. | 11 comprehensive test scenarios covering multi-line creation, line update/add/remove, duplicate name rejection, invalid references, invalid numbers, cross-parent line ID rejection, transactional rollback on child write failure, admission-linked deletion rejection, role permissions matrix, concurrency, and anonymous probe. |

---

## 2. Mathematical Calculation & Rounding Specification

1. **Line Amount**:
   $$\text{amount\_minor}_i = \text{quantity}_i \times \text{rate\_minor}_i$$
   where $\text{quantity}_i \in \mathbb{Z}^+$ and $\text{rate\_minor}_i \ge 0$.

2. **Subtotal**:
   $$\text{subtotal\_minor} = \sum_{i=1}^N \text{amount\_minor}_i$$

3. **Discount**:
   $$\text{discount\_amount\_minor} = \left\lfloor \frac{\text{subtotal\_minor} \times \text{discount} + 50}{100} \right\rfloor$$
   Explicit half-up integer rounding ensures exact minor-unit precision without floating-point accumulation errors.

4. **Total Amount**:
   $$\text{total\_amount\_minor} = \max(0, \text{subtotal\_minor} - \text{discount\_amount\_minor})$$

---

## 3. Security, Authorization & Error Codes

### 3.1 Role Matrix

| Endpoint | Method | Permission Required | Allowed Roles | Denied Roles (403) | Anonymous (401) |
|---|---|---|---|---|---|
| `/v1/packages` | `GET` | `packages.read` | `admin`, `receptionist`, `doctor`, `case_manager`, `patient` | `nurse`, `accountant`, `lab_technician` | Denied |
| `/v1/packages/{id}` | `GET` | `packages.read` | `admin`, `receptionist`, `doctor`, `case_manager`, `patient` | `nurse`, `accountant`, `lab_technician` | Denied |
| `/v1/packages-export` | `GET` | `packages.read` | `admin`, `receptionist`, `doctor`, `case_manager`, `patient` | `nurse`, `accountant`, `lab_technician` | Denied |
| `/v1/packages` | `POST` | `packages.manage` | `admin`, `receptionist` | `doctor`, `patient`, `nurse`, `accountant`, `case_manager`, `lab_technician` | Denied |
| `/v1/packages/{id}` | `PUT/PATCH` | `packages.manage` | `admin`, `receptionist` | `doctor`, `patient`, `nurse`, `accountant`, `case_manager`, `lab_technician` | Denied |
| `/v1/packages/{id}` | `DELETE` | `packages.manage` | `admin`, `receptionist` | `doctor`, `patient`, `nurse`, `accountant`, `case_manager`, `lab_technician` | Denied |

### 3.2 Error Responses

- **`400 Bad Request` (`INVALID_JSON`)**: Malformed JSON or unexpected fields.
- **`401 Unauthorized` (`UNAUTHENTICATED`)**: Missing or expired session cookie.
- **`403 Forbidden` (`FORBIDDEN`)**: Role lacks `packages.read` or `packages.manage`.
- **`404 Not Found` (`NOT_FOUND`)**: Package ID does not exist.
- **`409 Conflict` (`IDEMPOTENCY_CONFLICT` / `RECORD_IN_USE`)**:
  - Name already exists.
  - Cross-parent line ID submitted in package update.
  - Package is referenced in `ipd_admission_details` ("Package is in use by patient admissions and cannot be deleted").
- **`422 Unprocessable Entity` (`VALIDATION_FAILED`)**: Invalid quantity, negative rate, discount out of range [0, 100], empty lines, or inactive service reference.

---

## 4. Endpoints Specification

### 4.1 List Packages
- **Path**: `GET /v1/packages`
- **Query Parameters**:
  - `page`: integer (default 1)
  - `limit`: integer (default 25, max 100)
  - `search`: string (case-insensitive substring match on `name`)
- **Response `200 OK`**:
```json
{
  "packages": [
    {
      "id": "c1f78e45-1234-4567-89ab-cdef01234567",
      "name": "Executive Health Check",
      "description": "Comprehensive health checkup",
      "discount": 10,
      "total_amount_minor": 49500,
      "currency_symbol": "ETB",
      "services": [
        {
          "id": "e2a34567-89ab-cdef-0123-456789abcdef",
          "package_id": "c1f78e45-1234-4567-89ab-cdef01234567",
          "service_id": "ba111111-1111-1111-1111-111111111111",
          "service_name": "General Consultation",
          "quantity": 1,
          "rate_minor": 25000,
          "amount_minor": 25000,
          "created_at": "2026-10-08T02:00:00Z",
          "updated_at": "2026-10-08T02:00:00Z"
        }
      ],
      "created_at": "2026-10-08T02:00:00Z",
      "updated_at": "2026-10-08T02:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 25
}
```

### 4.2 Create Package
- **Path**: `POST /v1/packages`
- **Body**:
```json
{
  "name": "Executive Health Check",
  "description": "Comprehensive health checkup",
  "discount": 10,
  "services": [
    { "service_id": "ba111111-1111-1111-1111-111111111111", "quantity": 1, "rate_minor": 25000 },
    { "service_id": "ba222222-2222-2222-2222-222222222222", "quantity": 2, "rate_minor": 15000 }
  ]
}
```
- **Response `201 Created`**: Returns created `Package` with generated IDs and calculated line amounts and total.

### 4.3 Get Package Detail
- **Path**: `GET /v1/packages/{id}`
- **Response `200 OK`**: Returns single `Package` with full child lines and joined service names.

### 4.4 Update Package
- **Path**: `PUT /v1/packages/{id}`
- **Body**:
```json
{
  "name": "Executive Health Check (Revised)",
  "description": "Updated description",
  "discount": 15,
  "services": [
    { "id": "e2a34567-89ab-cdef-0123-456789abcdef", "service_id": "ba111111-1111-1111-1111-111111111111", "quantity": 2, "rate_minor": 25000 },
    { "service_id": "ba333333-3333-3333-3333-333333333333", "quantity": 1, "rate_minor": 10000 }
  ]
}
```
- **Transactional guarantees**:
  - Validates all line IDs belong strictly to this package. Cross-parent IDs are rejected with `409 Conflict`.
  - Recalculates line amounts and package totals.
  - Atomically updates existing lines, inserts new lines, deletes omitted lines, and updates parent in one transaction.
  - Writes audit event `package.updated`.

### 4.5 Delete Package
- **Path**: `DELETE /v1/packages/{id}`
- **Response `200 OK`**: `{"deleted": true}`
- **Protection**: Rejects deletion with `409 Conflict` if referenced in `ipd_admission_details`. Writes audit event `package.deleted`.
