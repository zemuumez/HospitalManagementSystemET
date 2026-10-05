# Role–Action–Record Matrix

> **Single-hospital deployment.** All roles below are system-wide, not tenant-scoped.
> 
> Permissions are implemented in [`domain/hospital.go` → `Actor.Can()`](../services/api/internal/domain/hospital.go).

## Legend
- ✅ Full access
- 📖 Read-only
- 🔒 Own records only
- ➕ Create only
- ❌ No access

---

## Permissions by Role

| Permission | admin | doctor | nurse | receptionist | pharmacist | accountant | case_manager | lab_technician | patient |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Identity & Access** | | | | | | | | | |
| `staff.manage` | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `patients.read` | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | 🔒 |
| `patients.create` | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `messages.manage` | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Scheduling** | | | | | | | | | |
| `appointments.read` | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | 🔒 |
| `appointments.book` | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | 🔒 |
| **Clinical** | | | | | | | | | |
| `clinical.read` | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | 🔒 |
| `clinical.admit` | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `beds.read` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `prescriptions.manage` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `prescriptions.read` | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | 🔒 |
| **Nursing** | | | | | | | | | |
| `operations.manage` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `operations.read` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Diagnostics** | | | | | | | | | |
| `diagnostics.catalog` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| `diagnostics.read` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | 🔒 |
| **Pharmacy & Blood Bank** | | | | | | | | | |
| `pharmacy.catalog` | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| `pharmacy.manage` | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| `medication.read` | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | 🔒 |
| `blood_bank.manage` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| `blood_bank.read` | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | 🔒 |
| **Billing & Finance** | | | | | | | | | |
| `billing.read` | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | 🔒 |
| `billing.manage` | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `finance.read` | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `finance.manage` | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `payroll.read` | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `payroll.manage` | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `payroll.read_own` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| **Services & Inventory** | | | | | | | | | |
| `services.manage` | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| `services.read` | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Ambulance** | | | | | | | | | |
| `ambulance.manage` | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| `ambulance.read` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| `ambulance_call.manage` | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| `ambulance_call.read` | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | 🔒 |
| **Attendance** | | | | | | | | | |
| `attendance.clock` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| `attendance.read_own` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| `attendance.manage` | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Front Office** | | | | | | | | | |
| `front_office.manage` | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `front_office.read` | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ |
| `complaints.create` | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |
| `complaints.read` | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | 🔒 |
| `complaints.manage` | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ |
| `notices.read` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `notices.manage` | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Live Consultations** | | | | | | | | | |
| `live_consultations.manage` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `live_consultations.read` | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | 🔒 |
| `live_meetings.manage` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `live_meetings.read` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| **CMS & Settings** | | | | | | | | | |
| `cms.manage` | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `cms.read` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `settings.manage` | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `settings.read` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |

---

## Record-Level Scoping Rules

These rules are enforced at the application/adapter layer, not just permission gates:

| Record | Scope rule |
|---|---|
| `appointment` | patients see only their own; doctors see only theirs; admin/receptionist see all |
| `clinical_case` | patients see only linked records; doctors see assigned; admin/receptionist see all |
| `patient_profile` | patient sees own only; admin/receptionist can read any |
| `vitals` | nurse/doctor on the encounter; patient cannot read directly |
| `payroll_entry` | `payroll.read_own` → staff see own month; `payroll.read` → accountant/admin see all |
| `audit_event` | admin-only cursor-paginated review; self-audit of own actions allowed |
| `billing_invoice` | patient sees own; accountant/admin sees all |
| `ambulance_call` | patient sees own call; staff see all |
| `prescription_order` | patient sees own; pharmacist/doctor/nurse/admin see all |
| `blood_donation` | lab_technician/doctor/admin manage; patient sees own linked records |
| `complaint` | patient sees own; admin/receptionist/case_manager manage; staff read |
| `attendance_record` | `attendance.read_own` → own records; `attendance.manage` → all staff records |

---

## Field-Level Restrictions (Planned)

| Field | Restriction |
|---|---|
| `patient_profile.email` | Not exposed in list API; full profile requires `patients.read` |
| `patient_profile.emergency_*` | Excluded from patient-self-read to prevent disclosure |
| `staff_profile.date_of_birth` | Admin-only; not exposed in doctor-public profile |
| `payroll_entry.amount_minor` | Only visible to owner (`payroll.read_own`) or accountant/admin |
| `audit_event.request_body` | Stored redacted; restoration requires off-host archival process |

> **Status**: Field-level enforcement is partially implemented. Full enforcement is tracked in
> checklist item 3.B "Apply field-level privacy, team assignment, export authorization and
> minimum necessary record visibility."

---

## Open Security Items (Checklist 3.B)

The following items are not yet fully implemented and are being tracked:

- [ ] **CSRF/origin for every mutation** — origin header enforcement exists in `server.go` (line 149); cookie-based CSRF token not yet issued.
- [ ] **Public endpoint rate limits** — no per-IP rate limiter; relies on upstream reverse proxy for now.
- [ ] **Field-level privacy enforcement** — see table above.
- [ ] **Production CSP/security headers** — to be set in the reverse proxy / deployment config.
- [ ] **Phone-change/unlink anti-claim** — Firebase phone-change flow not yet audited.
- [ ] **Independent security assessment** — deferred until pre-production hardening phase.
