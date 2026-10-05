# Backend Ambulance and Ambulance Calls Contract

## 1. Overview

This module implements the backend lifecycle for hospital ambulances and emergency ambulance calls, supporting:
- Vehicle registration, specification tracking (model, year, vehicle type: owned vs contracted), and live availability tracking.
- Ambulance call dispatching, linking patients, vehicle assignment, driver snapshots, location tracking, and call statuses (`dispatched`, `completed`, `cancelled`).
- Concurrency protection preventing dual-dispatch of vehicles in transit.
- Source billing linkage into sealed invoices with immutable provenance (`ambulance_call_invoice`).
- RBAC permissions aligning with hospital roles and patient data ownership.

## 2. Authorization & Role Matrix

| Action | Endpoints | Permitted Roles | Notes |
|---|---|---|---|
| Manage Ambulances | `POST /v1/ambulances`, `PUT /v1/ambulances/{id}` | Admin, Case Manager, Receptionist | Full vehicle CRUD |
| View Ambulances | `GET /v1/ambulances`, `GET /v1/ambulances/{id}` | Admin, Case Manager, Receptionist, Doctor, Nurse | Active vehicle list |
| Dispatch / Manage Calls | `POST /v1/ambulance-calls`, `PUT /v1/ambulance-calls/{id}` | Admin, Case Manager, Receptionist | Vehicle locking and dispatch |
| View Ambulance Calls | `GET /v1/ambulance-calls`, `GET /v1/ambulance-calls/{id}` | Admin, Case Manager, Receptionist, Doctor, Nurse, Accountant, Patient | Patients scoped strictly to their own calls |
| Bill Ambulance Call | `POST /v1/ambulance-calls/{id}/bill` | Admin, Accountant | Derives patient and tariff into issued sealed invoice |

## 3. Endpoints

### 3.1 Ambulances

#### `GET /v1/ambulances`
- **Query Params**:
  - `page`: integer (default 1)
  - `available`: boolean (`true` or `false`, optional)
- **Response `200 OK`**:
  ```json
  {
    "ambulances": [
      {
        "id": "a1111111-1111-1111-1111-111111111111",
        "vehicle_number": "ETH-AMB-01",
        "vehicle_model": "Toyota HiAce Emergency",
        "year_made": "2023",
        "driver_name": "Abebe Kebede",
        "driver_license": "DL-ETH-98214",
        "driver_contact": "+251911223344",
        "vehicle_type": 2,
        "is_available": true,
        "note": "Fully equipped ICU transport",
        "version": 1,
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1
  }
  ```

#### `POST /v1/ambulances`
- **Request**:
  ```json
  {
    "vehicle_number": "ETH-AMB-03",
    "vehicle_model": "Nissan NV350 Caravan",
    "year_made": "2024",
    "driver_name": "Kassahun Desta",
    "driver_license": "DL-ETH-44551",
    "driver_contact": "+251911334455",
    "vehicle_type": 2,
    "is_available": true,
    "note": "Standard emergency response"
  }
  ```
- **Response `201 Created`**: Returns created ambulance object.

#### `PUT /v1/ambulances/{id}`
- **Request**: Same fields as POST, plus required `"version": 1`.
- **Response `200 OK`**: Returns updated ambulance with incremented version.
- **Error `409 Conflict`**: If version is stale.

---

### 3.2 Ambulance Calls

#### `GET /v1/ambulance-calls`
- **Query Params**:
  - `page`: integer (default 1)
  - `status`: string (`dispatched`, `completed`, `cancelled`, optional)
  - `patient_id`: uuid (optional)
  - `ambulance_id`: uuid (optional)
- **Response `200 OK`**:
  ```json
  {
    "ambulance_calls": [
      {
        "id": "c1111111-1111-1111-1111-111111111111",
        "ambulance_id": "a1111111-1111-1111-1111-111111111111",
        "vehicle_model": "Toyota HiAce Emergency",
        "vehicle_number": "ETH-AMB-01",
        "patient_id": "p1111111-1111-1111-1111-111111111111",
        "patient_name": "Almaz Ayana",
        "driver_name": "Abebe Kebede",
        "call_date": "2026-10-05T01:30:00Z",
        "amount_minor": 180000,
        "status": "dispatched",
        "pickup_location": "Bole Airport Terminal 2",
        "destination": "Hospital Emergency Triage",
        "notes": "Emergency escort required",
        "invoice_id": null,
        "version": 1,
        "created_by": "user-rec-01",
        "created_at": "2026-10-05T01:30:00Z",
        "updated_at": "2026-10-05T01:30:00Z"
      }
    ],
    "total": 1,
    "page": 1
  }
  ```

#### `POST /v1/ambulance-calls`
- **Headers**:
  - `Idempotency-Key`: optional string for deduplication
- **Request**:
  ```json
  {
    "ambulance_id": "a1111111-1111-1111-1111-111111111111",
    "patient_id": "p1111111-1111-1111-1111-111111111111",
    "driver_name": "Abebe Kebede",
    "call_date": "2026-10-05T01:30:00Z",
    "amount_minor": 180000,
    "pickup_location": "Bole Airport Terminal 2",
    "destination": "Hospital Emergency Triage",
    "notes": "Emergency escort required"
  }
  ```
- **Response `201 Created`**: Returns dispatched call.
- **Behavior**:
  - Row-level lock (`SELECT ... FOR UPDATE`) on the target ambulance.
  - Rejects with `409 Conflict` if the ambulance is already in-use or unavailable.
  - Atomically marks `ambulance.is_available = false`.

#### `PUT /v1/ambulance-calls/{id}`
- **Request**:
  ```json
  {
    "ambulance_id": "a1111111-1111-1111-1111-111111111111",
    "driver_name": "Abebe Kebede",
    "call_date": "2026-10-05T01:30:00Z",
    "amount_minor": 180000,
    "status": "completed",
    "pickup_location": "Bole Airport Terminal 2",
    "destination": "Hospital Emergency Triage",
    "notes": "Patient safely transferred to emergency team",
    "version": 1
  }
  ```
- **Response `200 OK`**: Returns updated call with incremented version.
- **Behavior**:
  - Transitioning status to `completed` or `cancelled` atomically marks the assigned vehicle as available again (`ambulance.is_available = true`).
  - Swapping vehicle assignments during an active dispatch frees the former vehicle and locks/claims the newly selected vehicle.

---

### 3.3 Tariffs and Billing Integration

#### `POST /v1/ambulance-calls/{id}/bill`
- **Permissions**: `billing.manage` (Admin, Accountant)
- **Headers**:
  - `Idempotency-Key`: string
  - `Origin`: required matching server origin
- **Request**:
  ```json
  {
    "accountId": "acc-uuid-1111",
    "discountBasisPoints": 500
  }
  ```
- **Response `200 OK`**:
  ```json
  {
    "id": "inv-uuid-2222",
    "number": "INV-202610-0042",
    "patientId": "p1111111-1111-1111-1111-111111111111",
    "patientName": "Almaz Ayana",
    "invoiceDate": "2026-10-05",
    "subtotalMinor": 180000,
    "discountBasisPoints": 500,
    "totalMinor": 171000,
    "paidMinor": 0,
    "lines": [
      {
        "accountId": "acc-uuid-1111",
        "accountName": "Ambulance Transportation",
        "description": "Ambulance Emergency Transport: Toyota HiAce Emergency (ETH-AMB-01)",
        "quantity": 1,
        "unitPriceMinor": 180000
      }
    ],
    "version": 1
  }
  ```
- **Billing Guarantees**:
  - Derives patient and price directly from the ambulance call record.
  - Sealed invoice with exact discount deduction.
  - Linked via `ambulance_call_invoice` table with database trigger `protect_retained_record()` preventing mutation or deletion.
  - Concurrent duplicate requests return the same invoice idempotently.

---

## 4. Limitations & Next Steps

- **GPS / Telematics Integration**: Live coordinates or GPS device feeds are not covered; calls record textual pickup and destination coordinates.
- **Frontend Screens**: Frontend integration for Ambulance and Ambulance Call tables and modals will be integrated in Section 4.
