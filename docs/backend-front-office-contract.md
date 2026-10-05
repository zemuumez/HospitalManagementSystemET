# Backend Front Office, Complaints, Enquiries, Notices, Visitors, Call Logs, and Postals Contract

## 1. Overview

This module implements the backend lifecycle for hospital front-desk, communication, and patient feedback operations:
- **Complaints & Grievances**: Patient complaint filing, lifecycle management (`0`=Pending, `1`=In Progress, `2`=Resolved, `3`=Rejected), administrative resolution with feedback responses, and strict patient-scoping privacy rules.
- **Notice Board**: Platform-wide hospital announcements, departmental bulletins, and staff notices.
- **Front-Office Enquiries**: Public patient enquiries and feedback intake, read-status tracking (`0`=Unread, `1`=Read), and receptionist review auditing.
- **Visitor Logs**: Visitor gate register, visiting purpose classification (Visit, Enquiry, Seminar/Vendor), identity card records, visitor headcounts, date, check-in and check-out timestamps.
- **Call Logs**: Reception telephonic logs with caller metadata, incoming vs outgoing tracking, call dates, and follow-up schedule dates.
- **Postal Records**: Physical mail registry tracking incoming dispatch receipts and outgoing couriers with reference numbers, addresses, and date stamping.

---

## 2. Authorization & Role Matrix

| Resource | Read Endpoints | Manage Endpoints | Permitted Roles |
|---|---|---|---|
| **Complaints** | `GET /v1/complaints`, `GET /v1/complaints/{id}` | `POST /v1/complaints` (Create), `PUT /v1/complaints/{id}/resolve` (Resolve) | Read: Authenticated staff / Patient (Patient strictly scoped to their own complaints). Create: Patient, Admin, Receptionist. Resolve: `admin`, `receptionist`, `case_manager` (`complaints.manage`). |
| **Notice Board** | `GET /v1/notices` | `POST /v1/notices`, `DELETE /v1/notices/{id}` | Read: All authenticated users (`notices.read`). Manage: `admin` (`notices.manage`). |
| **Enquiries** | `GET /v1/enquiries` | `POST /v1/enquiries` (Public), `PUT /v1/enquiries/{id}/read` | Read/Mark Read: `admin`, `receptionist` (`front_office.read`, `front_office.manage`). Submit: Public / Unauthenticated allowed. |
| **Visitors** | `GET /v1/visitors` | `POST /v1/visitors`, `PUT /v1/visitors/{id}` | Read: Authenticated clinical & front-office staff. Manage: `admin`, `receptionist` (`front_office.manage`). |
| **Call Logs** | `GET /v1/call-logs` | `POST /v1/call-logs`, `PUT /v1/call-logs/{id}` | Read: Authenticated clinical & front-office staff. Manage: `admin`, `receptionist` (`front_office.manage`). |
| **Postals** | `GET /v1/postals` | `POST /v1/postals`, `PUT /v1/postals/{id}` | Read: Authenticated clinical & front-office staff. Manage: `admin`, `receptionist` (`front_office.manage`). |

---

## 3. Endpoints

### 3.1 Complaints

#### `GET /v1/complaints`
- **Query Params**:
  - `page`: integer (default 1)
  - `patient_id`: string (optional for staff; forced to authenticated user ID for patients)
- **Response `200 OK`**:
  ```json
  {
    "complaints": [
      {
        "id": "c1111111-1111-1111-1111-111111111111",
        "patient_id": "patient-user-id",
        "title": "Pharmacy wait time too long",
        "description": "Waited 45 minutes for prescribed medication.",
        "status": 0,
        "response": "",
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1
  }
  ```

#### `POST /v1/complaints`
- **Request Body**:
  ```json
  {
    "title": "Pharmacy wait time too long",
    "description": "Waited 45 minutes for prescribed medication."
  }
  ```
- **Response `201 Created`**: Returns created `HospitalComplaint`.

#### `PUT /v1/complaints/{id}/resolve`
- **Request Body**:
  ```json
  {
    "status": 2,
    "response": "Workflow reviewed and additional staff deployed to peak hours."
  }
  ```
- **Response `200 OK`**: Returns resolved `HospitalComplaint`.

---

### 3.2 Notice Board

#### `GET /v1/notices`
- **Response `200 OK`**:
  ```json
  {
    "notices": [
      {
        "id": "ea111111-1111-1111-1111-111111111111",
        "title": "Hospital Accreditation Renewal Completed",
        "description": "Grade-A Tertiary Clinical Excellence certified.",
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

#### `POST /v1/notices`
- **Request Body**:
  ```json
  {
    "title": "Departmental Clinical Review",
    "description": "Weekly mortality and morbidity clinical audit in Main Conference Hall."
  }
  ```
- **Response `201 Created`**: Returns created `HospitalNoticeBoard`.

---

### 3.3 Enquiries

#### `POST /v1/enquiries`
- **Request Body**:
  ```json
  {
    "full_name": "Helen Mekonnen",
    "email": "helen.m@example.com",
    "contact_no": "+251911445566",
    "type": 1,
    "message": "Do you offer weekend pediatric cardiology appointments?"
  }
  ```
- **Response `201 Created`**: Returns created `HospitalEnquiry`.

#### `PUT /v1/enquiries/{id}/read`
- **Response `200 OK`**: Returns updated `HospitalEnquiry` with `status: 1`.

---

### 3.4 Visitors

#### `GET /v1/visitors`
- **Query Params**: `page`, `date`
- **POST /v1/visitors**:
  ```json
  {
    "purpose": 1,
    "name": "Yonas Berhe",
    "phone": "+251911778899",
    "id_card": "KEBELE-09-8877",
    "no_of_person": 2,
    "date": "2026-10-05",
    "in_time": "14:00",
    "out_time": "15:30",
    "note": "Visiting Ward 3 Bed 102"
  }
  ```

---

### 3.5 Call Logs & Postals

#### `GET /v1/call-logs` & `POST /v1/call-logs`
- Call types: `1`=Incoming, `2`=Outgoing.

#### `GET /v1/postals` & `POST /v1/postals`
- Postal types: `1`=Receive, `2`=Dispatch.
