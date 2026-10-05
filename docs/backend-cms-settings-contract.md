# Backend CMS and Hospital General Settings Contract

## 1. Overview

This module implements the backend lifecycle for hospital general configuration and content management:
- **Hospital General Settings**: Key-value platform settings covering hospital naming, contact information, branding logos/favicons, current currency (ETB), default language (`en`, `am`), social URLs, and patient queue display themes (`modern-blue`, `dark-emerald`, etc.).
- **Hospital Operating Schedules**: Weekly day-by-day operating hours (1=Monday to 7=Sunday) with opening and closing times and holiday/closed-day toggles.
- **Front CMS Settings**: Content sections for public hospital landing and informational pages (home hero, about hospital, certified doctors, mission, terms & conditions, privacy policy, and embedded contact maps).
- **Testimonials & Moderation**: Patient stories, reviews, and clinical testimonials with rating (1-5) and publication moderation status (`0`=draft/review, `1`=published).

---

## 2. Authorization & Role Matrix

| Resource | Read Endpoints | Manage Endpoints | Permitted Roles |
|---|---|---|---|
| **Hospital General Settings** | `GET /v1/general-settings` | `POST /v1/general-settings`, `PUT /v1/general-settings` | Read: Authenticated staff / Public. Manage: `admin` (`settings.manage`) |
| **Hospital Schedules** | `GET /v1/hospital-schedules` | `POST /v1/hospital-schedules`, `PUT /v1/hospital-schedules` | Read: Authenticated staff / Public. Manage: `admin` (`settings.manage`) |
| **Front CMS Settings** | `GET /v1/front-cms-settings` | `POST /v1/front-cms-settings`, `PUT /v1/front-cms-settings` | Read: Authenticated staff / Public. Manage: `admin` (`cms.manage`) |
| **Testimonials** | `GET /v1/testimonials` | `POST /v1/testimonials`, `PUT /v1/testimonials/{id}`, `DELETE /v1/testimonials/{id}` | Read: All roles (non-admin restricted to published only). Manage: `admin` (`cms.manage`) |

---

## 3. Endpoints

### 3.1 Hospital General Settings

#### `GET /v1/general-settings`
- **Response `200 OK`**:
  ```json
  {
    "general_settings": {
      "app_name": "Hospital Management System",
      "hospital_name": "Addis Ababa Central Hospital",
      "hospital_email": "info@hospital.et",
      "hospital_phone": "+251911000000",
      "hospital_address": "Bole Sub-City, Addis Ababa, Ethiopia",
      "current_currency": "ETB",
      "default_lang": "en",
      "logo_url": "/images/logo.png",
      "favicon_url": "/favicon.ico",
      "queue_theme": "modern-blue"
    }
  }
  ```

#### `POST /v1/general-settings` or `PUT /v1/general-settings`
- **Request Body**:
  ```json
  {
    "hospital_phone": "+251911999888",
    "queue_theme": "dark-emerald"
  }
  ```
- **Response `200 OK`**:
  ```json
  {
    "saved": true,
    "general_settings": { ... }
  }
  ```

---

### 3.2 Hospital Schedules

#### `GET /v1/hospital-schedules`
- **Response `200 OK`**:
  ```json
  {
    "hospital_schedules": [
      {
        "id": "e1111111-1111-1111-1111-111111111111",
        "day_of_week": 1,
        "start_time": "08:00",
        "end_time": "17:00",
        "is_closed": false,
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

#### `PUT /v1/hospital-schedules`
- **Request Body**:
  ```json
  {
    "day_of_week": 7,
    "start_time": "09:00",
    "end_time": "13:00",
    "is_closed": false
  }
  ```
- **Response `200 OK`**: Returns updated `HospitalScheduleDay`.

---

### 3.3 Front CMS Settings

#### `GET /v1/front-cms-settings`
- **Query Params**:
  - `type`: string (optional filter: `home`, `about`, `services`, `contact`, `terms`, `privacy`)
- **Response `200 OK`**:
  ```json
  {
    "front_cms_settings": [
      {
        "key": "home_title",
        "value": "Comprehensive Compassionate Healthcare in Addis Ababa",
        "type": "home",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

#### `POST /v1/front-cms-settings`
- **Request Body**:
  ```json
  {
    "key": "home_title",
    "value": "Leading Clinical Excellence in East Africa",
    "type": "home"
  }
  ```
- **Response `200 OK`**: Returns updated `FrontCMSSetting`.

---

### 3.4 CMS Testimonials

#### `GET /v1/testimonials`
- **Query Params**:
  - `status`: integer (optional filter: `1`=published, `0`=draft)
- **Response `200 OK`**:
  ```json
  {
    "testimonials": [
      {
        "id": "fa111111-1111-1111-1111-111111111111",
        "name": "Almaz Tadesse",
        "description": "The surgical and nursing team provided outstanding care during my recovery.",
        "position": "Patient",
        "rating": 5,
        "status": 1,
        "created_at": "2026-10-05T00:00:00Z",
        "updated_at": "2026-10-05T00:00:00Z"
      }
    ]
  }
  ```

#### `POST /v1/testimonials`
- **Request Body**:
  ```json
  {
    "name": "Tigist Alemu",
    "description": "The intensive care unit doctors were very attentive and kind.",
    "position": "Patient",
    "rating": 5,
    "status": 1
  }
  ```
- **Response `201 Created`**: Returns created `CMSTestimonial`.

#### `PUT /v1/testimonials/{id}`
- Updates fields. Returns updated `CMSTestimonial`.

#### `DELETE /v1/testimonials/{id}`
- Deletes testimonial record. Returns `{"deleted": true}`.
