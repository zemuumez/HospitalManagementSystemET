# Backend Live Consultations, Meetings, and Provider Settings Contract

## 1. Overview

This module implements the backend lifecycle for remote clinical teleconsultations, hospital staff video conferences, and virtual provider integration settings:
- **Live Consultations**: Teleconsultations between clinicians and patients (OPD / IPD encounter linkage, Zoom / Google Meet / In-House platforms, meeting IDs, passcode generation, participant & host video preferences, duration in minutes, `Africa/Addis_Ababa` timezone, status lifecycle `0`=Awaited, `1`=Finished, `2`=Cancelled, and strict patient-scoping privacy rules).
- **Live Meetings**: Hospital-wide and department-level staff meetings, multi-candidate staff attendance invitations, meeting credentials, and lifecycle progression.
- **Provider Settings**: Clinician-specific virtual meeting provider configurations (API keys and secrets for Zoom and Meet), with security masking on retrieval and preservation of project rules (real credentials remain blank).

---

## 2. Authorization & Role Matrix

| Resource | Read Endpoints | Manage Endpoints | Permitted Roles |
|---|---|---|---|
| **Live Consultations** | `GET /v1/live-consultations`, `GET /v1/live-consultations/{id}` | `POST /v1/live-consultations`, `PUT /v1/live-consultations/{id}/status`, `PATCH /v1/live-consultations/{id}/status` | Read: `admin`, `doctor`, `nurse`, `patient` (`live_consultations.read`). Patients are strictly isolated to their own consultations. Doctors see consultations where they are the clinician or creator. Manage: `admin`, `doctor` (`live_consultations.manage`). |
| **Live Meetings** | `GET /v1/live-meetings`, `GET /v1/live-meetings/{id}` | `POST /v1/live-meetings`, `PUT /v1/live-meetings/{id}/status`, `PATCH /v1/live-meetings/{id}/status` | Read: Hospital staff roles (`admin`, `doctor`, `nurse`, `pharmacist`, `receptionist`, `accountant`, `case_manager`, `lab_technician`) (`live_meetings.read`). Manage: `admin`, `doctor` (`live_meetings.manage`). |
| **Provider Settings** | `GET /v1/live-consultations/provider-settings`, `GET /v1/live-consultation-settings` | `POST /v1/live-consultations/provider-settings`, `PUT /v1/live-consultations/provider-settings`, `POST /v1/live-consultation-settings` | Manage/Read Own: `admin`, `doctor` (`live_consultations.manage`). Returns masked credential presence (`has_api_key`, `has_api_secret`). |

---

## 3. Endpoints

### 3.1 Live Consultations

#### `GET /v1/live-consultations`
- **Query Params**:
  - `page`: integer (default `1`)
  - `status`: integer (optional: `0`=Awaited, `1`=Finished, `2`=Cancelled)
- **Response `200 OK`**:
  ```json
  {
    "live_consultations": [
      {
        "id": "c1111111-1111-1111-1111-111111111111",
        "doctor_id": "doctor-user-id",
        "patient_id": "patient-uuid",
        "encounter_id": "optional-encounter-uuid",
        "consultation_title": "Cardiology Teleconsultation",
        "consultation_date": "2026-10-06T10:00:00+03:00",
        "duration_minutes": 30,
        "host_video": true,
        "participant_video": true,
        "type": "OPD",
        "type_number": "OPD-1024",
        "platform_type": "zoom",
        "meeting_id": "HMS-a1b2c3d4",
        "password": "passcode",
        "time_zone": "Africa/Addis_Ababa",
        "status": 0,
        "description": "Routine teleconsultation checkup",
        "created_by": "doctor-user-id",
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1
  }
  ```

#### `POST /v1/live-consultations`
- **Request Body**:
  ```json
  {
    "doctor_id": "doctor-user-id",
    "patient_id": "patient-uuid",
    "encounter_id": "optional-encounter-uuid",
    "consultation_title": "Cardiology Teleconsultation",
    "consultation_date": "2026-10-06T10:00:00+03:00",
    "duration_minutes": 30,
    "host_video": true,
    "participant_video": true,
    "type": "OPD",
    "type_number": "OPD-1024",
    "platform_type": "zoom",
    "meeting_id": "optional-custom-meeting-id",
    "password": "optional-passcode",
    "description": "Routine teleconsultation checkup"
  }
  ```
- **Response `201 Created`**: Returns created `LiveConsultation` object.

#### `GET /v1/live-consultations/{id}`
- **Response `200 OK`**: Returns single `LiveConsultation` object.

#### `PATCH /v1/live-consultations/{id}/status`
- **Request Body**:
  ```json
  {
    "status": 1
  }
  ```
- **Response `200 OK`**: Returns updated `LiveConsultation` object.

---

### 3.2 Live Meetings

#### `GET /v1/live-meetings`
- **Query Params**:
  - `page`: integer (default `1`)
  - `status`: integer (optional)
- **Response `200 OK`**:
  ```json
  {
    "live_meetings": [
      {
        "id": "m1111111-1111-1111-1111-111111111111",
        "title": "Clinical Governance Committee",
        "meeting_date": "2026-10-07T14:00:00+03:00",
        "duration_minutes": 60,
        "host_video": true,
        "participant_video": true,
        "platform_type": "meet",
        "meeting_id": "MTG-f1e2d3c4",
        "password": "passcode",
        "time_zone": "Africa/Addis_Ababa",
        "status": 0,
        "description": "Quarterly antibiotic stewardship and safety review",
        "created_by": "admin-user-id",
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1
  }
  ```

#### `POST /v1/live-meetings`
- **Request Body**:
  ```json
  {
    "title": "Clinical Governance Committee",
    "meeting_date": "2026-10-07T14:00:00+03:00",
    "duration_minutes": 60,
    "host_video": true,
    "participant_video": true,
    "platform_type": "meet",
    "description": "Quarterly antibiotic stewardship and safety review",
    "candidate_user_ids": [
      "doctor-user-id",
      "nurse-user-id"
    ]
  }
  ```
- **Response `201 Created`**: Returns created `LiveMeeting` object with populated `candidate_user_ids`.

#### `GET /v1/live-meetings/{id}`
- **Response `200 OK`**: Returns single `LiveMeeting` object including candidate user IDs.

#### `PATCH /v1/live-meetings/{id}/status`
- **Request Body**:
  ```json
  {
    "status": 1
  }
  ```
- **Response `200 OK`**: Returns updated `LiveMeeting` object.

---

### 3.3 Provider Settings

#### `GET /v1/live-consultations/provider-settings`
- **Response `200 OK`**:
  ```json
  {
    "user_id": "doctor-user-id",
    "platform_type": "zoom",
    "has_api_key": true,
    "has_api_secret": true,
    "updated_at": "2026-10-05T00:00:00Z"
  }
  ```

#### `POST /v1/live-consultations/provider-settings`
- **Request Body**:
  ```json
  {
    "platform_type": "zoom",
    "api_key": "sample-zoom-api-key",
    "api_secret": "sample-zoom-api-secret"
  }
  ```
- **Response `200 OK`**: Returns saved `LiveProviderSetting` with credentials masked.

---

## 4. Audit Log Integration

All state mutations record immutable audit events:
- `live_consultation.created`
- `live_consultation.status_updated`
- `live_meeting.created`
- `live_meeting.status_updated`
- `live_consultation.provider_setting_updated`
