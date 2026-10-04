# Backend Services, Charge Categories, Operations, Custom Fields, and Module Settings Contract

## 1. Overview

This module implements the backend for operational hospital catalog configuration:
- **Charge Categories & Charges**: Categorization (Investigation, Operation, Bed, Doctor, Other) and standard charges in Ethiopian Birr minor units (cents).
- **Hospital Services**: Billable and procedural hospital services with base quantity, unit rates in ETB cents, and active status toggling.
- **Operations & Categories**: Surgical and clinical procedural taxonomy and definitions with category references.
- **Custom Fields**: Dynamic form extensions configured per module (patient, appointment, ipd, opd, etc.), field types (text, number, select, date, boolean, textarea), validation rules, and responsive grid spans (1-12).
- **Module Settings**: Platform-wide hospital module activation toggles governing routing and sub-system availability.

---

## 2. Authorization & Role Matrix

| Domain Area | Read Endpoints | Write/Manage Endpoints | Allowed Roles |
|---|---|---|---|
| **Charge Categories** | `GET /v1/charge-categories` | `POST /v1/charge-categories`, `PUT /v1/charge-categories/{id}` | Read: All authenticated staff. Manage: `admin`, `accountant` (`services.manage`) |
| **Hospital Charges** | `GET /v1/charges` | `POST /v1/charges`, `PUT /v1/charges/{id}` | Read: All authenticated staff. Manage: `admin`, `accountant` (`services.manage`) |
| **Hospital Services** | `GET /v1/services` | `POST /v1/services`, `PUT /v1/services/{id}` | Read: All authenticated staff. Manage: `admin`, `accountant` (`services.manage`) |
| **Operation Categories** | `GET /v1/operation-categories` | `POST /v1/operation-categories` | Read: All authenticated staff. Manage: `admin`, `doctor` (`operations.manage`) |
| **Operations** | `GET /v1/operations` | `POST /v1/operations`, `PUT /v1/operations/{id}` | Read: All authenticated staff. Manage: `admin`, `doctor` (`operations.manage`) |
| **Custom Fields** | `GET /v1/custom-fields` | `POST /v1/custom-fields`, `DELETE /v1/custom-fields/{id}` | Read: All authenticated staff. Manage: `admin` (`settings.manage`) |
| **Module Settings** | `GET /v1/module-settings` | `PUT /v1/module-settings/{key}` | Read: All authenticated staff. Manage: `admin` (`settings.manage`) |

---

## 3. Endpoints

### 3.1 Charge Categories

#### `GET /v1/charge-categories`
- **Query Params**:
  - `type`: integer (optional filter: 1=Investigation, 2=Operation, 3=Bed, 4=Doctor, 5=Other)
- **Response `200 OK`**:
  ```json
  {
    "charge_categories": [
      {
        "id": "ca111111-1111-1111-1111-111111111111",
        "name": "Investigations",
        "description": "Clinical laboratory and imaging investigations",
        "charge_type": 1,
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

#### `POST /v1/charge-categories`
- **Request Body**:
  ```json
  {
    "name": "Physiotherapy & Rehabilitation",
    "description": "Musculoskeletal therapy and physical training",
    "charge_type": 5
  }
  ```
- **Response `201 Created`**: Returns created `ChargeCategory`.

---

### 3.2 Hospital Charges

#### `GET /v1/charges`
- **Query Params**:
  - `page`: integer (default 1)
  - `search`: string (optional code or description filter)
- **Response `200 OK`**:
  ```json
  {
    "charges": [
      {
        "id": "d1111111-1111-1111-1111-111111111111",
        "charge_type": 5,
        "charge_category_id": "ca111111-1111-1111-1111-111111111111",
        "category_name": "Investigations",
        "code": "CHG-CBC-01",
        "standard_charge_minor": 15000,
        "description": "Complete Blood Count",
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 25
  }
  ```

#### `POST /v1/charges`
- **Request Body**:
  ```json
  {
    "charge_type": 5,
    "charge_category_id": "ca111111-1111-1111-1111-111111111111",
    "code": "CHG-DENT-01",
    "standard_charge_minor": 30000,
    "description": "Simple Tooth Extraction"
  }
  ```
- **Response `201 Created`**: Returns created `HospitalCharge`.

---

### 3.3 Hospital Services

#### `GET /v1/services`
- **Query Params**:
  - `page`: integer (default 1)
  - `status`: integer (optional: `1` for active, `0` for inactive)
- **Response `200 OK`**:
  ```json
  {
    "services": [
      {
        "id": "ba111111-1111-1111-1111-111111111111",
        "name": "General Consultation",
        "description": "Standard outpatient specialist medical review",
        "quantity": 1,
        "rate_minor": 25000,
        "status": 1,
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 25
  }
  ```

#### `POST /v1/services`
- **Request Body**:
  ```json
  {
    "name": "Emergency Wound Suture",
    "description": "Primary wound debridement and closure",
    "quantity": 1,
    "rate_minor": 20000,
    "status": 1
  }
  ```
- **Response `201 Created`**: Returns created `HospitalService`.

---

### 3.4 Operations & Categories

#### `GET /v1/operation-categories`
- **Response `200 OK`**:
  ```json
  {
    "operation_categories": [
      {
        "id": "aa111111-1111-1111-1111-111111111111",
        "name": "General Surgery",
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

#### `GET /v1/operations`
- **Query Params**:
  - `page`: integer (default 1)
  - `search`: string (optional name filter)
- **Response `200 OK`**:
  ```json
  {
    "operations": [
      {
        "id": "bb111111-1111-1111-1111-111111111111",
        "operation_category_id": "aa111111-1111-1111-1111-111111111111",
        "operation_category_name": "General Surgery",
        "name": "Appendectomy",
        "description": "Removal of vermiform appendix",
        "status": 1,
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 25
  }
  ```

---

### 3.5 Custom Fields

#### `GET /v1/custom-fields`
- **Query Params**:
  - `module`: string (e.g. `patient`, `appointment`, `ipd`, `opd`)
- **Response `200 OK`**:
  ```json
  {
    "custom_fields": [
      {
        "id": "e1111111-1111-1111-1111-111111111111",
        "module_name": "patient",
        "field_type": "text",
        "field_name": "National Identification Number",
        "is_required": true,
        "values": "",
        "grid": 6,
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

#### `POST /v1/custom-fields`
- **Request Body**:
  ```json
  {
    "module_name": "patient",
    "field_type": "select",
    "field_name": "Blood Donor Registry",
    "is_required": false,
    "values": "Yes,No,Eligible",
    "grid": 6
  }
  ```
- **Response `201 Created`**: Returns created `CustomField`.

---

### 3.6 Module Settings

#### `GET /v1/module-settings` (or `/v1/modules-setting`)
- **Response `200 OK`**:
  ```json
  {
    "module_settings": [
      {
        "id": "f1111111-1111-1111-1111-111111111111",
        "module_key": "ambulances",
        "name": "Ambulance Services",
        "route": "ambulances",
        "is_active": true,
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

#### `PUT /v1/module-settings/{key}`
- **Request Body**:
  ```json
  {
    "is_active": false
  }
  ```
- **Response `200 OK`**: Returns updated `HospitalModuleSetting`.
