# Backend Pharmacy Masters, Blood Bank, and Prescriptions Contract

## 1. Overview

This module implements the backend lifecycle for pharmacy master data, blood transfusion banking, and structured clinical prescriptions:
- **Medicine Categories & Brands**: Master catalogs allowing classification of medicines and manufacturer/distributor tracking.
- **Blood Bank Inventory**: Strict inventory tracking across all 8 standard human blood groups (`A+`, `A-`, `B+`, `B-`, `AB+`, `AB-`, `O+`, `O-`). Atomically increments upon donation and decrements upon patient issue, with enforced no-negative-inventory guards.
- **Blood Donors & Donations**: Donor registry with age/gender/group validation, donation logs, bags counts, and audit trails.
- **Blood Issues**: Doctor-authorized blood dispensation to patients, fee tracking (`amount_minor` in ETB), optional billing invoice association, and patient privacy isolation.
- **Structured Prescriptions**: Comprehensive clinical prescription documents (food allergies, cardiac/hypertensive/diabetic history, vitals, clinical advice, follow-up scheduling, and itemized multi-medicine dosage instructions) with fulfillment status tracking (`0`=Pending, `1`=Dispensed).

---

## 2. Authorization & Role Matrix

| Resource | Read Endpoints | Manage Endpoints | Permitted Roles |
|---|---|---|---|
| **Medicine Categories** | `GET /v1/medicine-categories` | `POST /v1/medicine-categories`, `PUT /v1/medicine-categories/{id}` | Read: `admin`, `pharmacist`, `doctor` (`pharmacy.catalog`). Manage: `admin`, `pharmacist` (`pharmacy.manage`). |
| **Medicine Brands** | `GET /v1/medicine-brands` | `POST /v1/medicine-brands`, `PUT /v1/medicine-brands/{id}` | Read: `admin`, `pharmacist`, `doctor` (`pharmacy.catalog`). Manage: `admin`, `pharmacist` (`pharmacy.manage`). |
| **Blood Bank & Donors** | `GET /v1/blood-bank`, `GET /v1/blood-donors`, `GET /v1/blood-donations` | `POST /v1/blood-donors`, `POST /v1/blood-donations` | Read: Hospital staff (`admin`, `doctor`, `nurse`, `lab_technician`, `pharmacist`, `receptionist`) (`blood_bank.read`). Manage: `admin`, `lab_technician`, `doctor` (`blood_bank.manage`). |
| **Blood Issues** | `GET /v1/blood-issues`, `GET /v1/blood-issues/{id}` | `POST /v1/blood-issues` | Read: Staff / Patient (Patients strictly isolated to their own records). Manage: `admin`, `lab_technician`, `doctor` (`blood_bank.manage`). |
| **Prescriptions** | `GET /v1/prescriptions`, `GET /v1/prescriptions/{id}` | `POST /v1/prescriptions`, `PATCH /v1/prescriptions/{id}/status` | Read: `admin`, `doctor`, `nurse`, `pharmacist`, `patient` (`prescriptions.read`; patients strictly scoped). Create: `admin`, `doctor` (`prescriptions.manage`). Dispense: `admin`, `pharmacist`, `doctor`. |

---

## 3. Endpoints

### 3.1 Medicine Categories & Brands

#### `GET /v1/medicine-categories`
- **Query Params**: `page` (default 1), `search` (optional)
- **Response `200 OK`**:
  ```json
  {
    "categories": [
      {
        "id": "c1111111-1111-1111-1111-111111111111",
        "name": "Cardiovascular Agents",
        "is_active": true,
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1
  }
  ```

#### `POST /v1/medicine-categories`
- **Request Body**:
  ```json
  {
    "name": "Cardiovascular Agents",
    "is_active": true
  }
  ```
- **Response `201 Created`**: Returns created category.

#### `GET /v1/medicine-brands`
- **Query Params**: `page` (default 1), `search` (optional)
- **Response `200 OK`**:
  ```json
  {
    "brands": [
      {
        "id": "b1111111-1111-1111-1111-111111111111",
        "name": "Ethiopian Pharmaceuticals Manufacturing Sh.Co.",
        "email": "info@epharm.et",
        "phone": "+251114421111",
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1
  }
  ```

#### `POST /v1/medicine-brands`
- **Request Body**:
  ```json
  {
    "name": "Ethiopian Pharmaceuticals Manufacturing Sh.Co.",
    "email": "info@epharm.et",
    "phone": "+251114421111"
  }
  ```
- **Response `201 Created`**: Returns created brand.

---

### 3.2 Blood Bank & Donors

#### `GET /v1/blood-bank`
- **Response `200 OK`**:
  ```json
  {
    "blood_bank": [
      {"id": "...", "blood_group": "A+", "remained_bags": 12, "updated_at": "..."},
      {"id": "...", "blood_group": "A-", "remained_bags": 4, "updated_at": "..."},
      {"id": "...", "blood_group": "B+", "remained_bags": 8, "updated_at": "..."},
      {"id": "...", "blood_group": "B-", "remained_bags": 2, "updated_at": "..."},
      {"id": "...", "blood_group": "AB+", "remained_bags": 5, "updated_at": "..."},
      {"id": "...", "blood_group": "AB-", "remained_bags": 1, "updated_at": "..."},
      {"id": "...", "blood_group": "O+", "remained_bags": 20, "updated_at": "..."},
      {"id": "...", "blood_group": "O-", "remained_bags": 6, "updated_at": "..."}
    ]
  }
  ```

#### `POST /v1/blood-donors`
- **Request Body**:
  ```json
  {
    "name": "Mulugeta Tadesse",
    "age": 29,
    "gender": 0,
    "blood_group": "O+",
    "last_donate_date": "2026-10-05T10:00:00Z"
  }
  ```
- **Response `201 Created`**: Returns created donor.

#### `POST /v1/blood-donations`
- **Request Body**:
  ```json
  {
    "donor_id": "donor-uuid",
    "bags": 2,
    "donation_date": "2026-10-05T10:00:00Z"
  }
  ```
- **Response `201 Created`**: Automatically increments `blood_bank.remained_bags` for the donor's blood group.

#### `POST /v1/blood-issues`
- **Request Body**:
  ```json
  {
    "doctor_id": "doctor-user-id",
    "patient_id": "patient-uuid",
    "donor_id": "optional-donor-uuid",
    "blood_group": "O+",
    "bags": 1,
    "amount_minor": 50000,
    "issue_date": "2026-10-05T12:00:00Z",
    "remarks": "Transfusion for surgical care"
  }
  ```
- **Response `201 Created`**: Atomically validates inventory and decrements `blood_bank.remained_bags`. Returns `422 Unprocessable Entity` if inventory is insufficient.

---

### 3.3 Prescriptions

#### `POST /v1/prescriptions`
- **Request Body**:
  ```json
  {
    "patient_id": "patient-uuid",
    "doctor_id": "doctor-user-id",
    "encounter_id": "optional-encounter-uuid",
    "high_blood_pressure": "130/85",
    "diabetic": "No",
    "problem_description": "Acute bacterial pharyngitis and mild fever",
    "advice": "Take medications after meals with full glass of water. Rest 3 days.",
    "next_visit_qty": "7",
    "next_visit_time": "days",
    "medicines": [
      {
        "medicine_name": "Amoxicillin 500mg Capsule",
        "dosage": "1 cap three times daily",
        "day": "7",
        "time": "After Meals",
        "comment": "Complete entire course"
      },
      {
        "medicine_name": "Paracetamol 500mg Tablet",
        "dosage": "1 tab every 8 hours as needed",
        "day": "3",
        "time": "After Meals",
        "comment": "For fever and pain"
      }
    ]
  }
  ```
- **Response `201 Created`**: Returns created prescription with itemized medicines.

#### `PATCH /v1/prescriptions/{id}/status`
- **Request Body**:
  ```json
  {
    "status": 1
  }
  ```
- **Response `200 OK`**: Marks prescription as dispensed (`1`).

---

## 4. Audit Log Integration

All state mutations record immutable audit events:
- `medicine_category.created`, `medicine_category.updated`
- `medicine_brand.created`, `medicine_brand.updated`
- `blood_donor.created`
- `blood_donation.recorded`
- `blood_issue.created`
- `prescription.created`, `prescription.status_updated`
