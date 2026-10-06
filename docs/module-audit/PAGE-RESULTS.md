# Admin page audit results

112 planned pages; 112 visited. Results: {'pass': 101, 'api-failure': 0, 'fail': 11}.

Authenticated rendering/initial transport only. A pass is not CRUD or visual parity. Failures and HTTP errors remain unresolved; no application fixes were made during this audit.

Network-idle timeouts are test-observation failures, not proof of a crashed screen. HTTP status failures below were directly observed. Duplicate requests are collapsed in this table; raw counts remain in test-results.json.

| Page | Result | Observed API failures | Observation |
| --- | --- | --- | --- |
| `/dashboard` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/patient-id-card-template` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/generate-patient-id-card` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/users` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/admins` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/accountants` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/nurses` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/lab-technicians` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/receptionists` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/pharmacists` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/appointments` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/appointment-calendars` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/appointment-transaction` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/patient-queues` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/ipd-patient-departments` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/opd-patient-departments` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/accounts` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/employee-payrolls` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/invoices` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/payments` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/payment-reports` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/advanced-payments` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/bills` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/manual-bill-payments` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/bed-status` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/bed-assigns` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/beds` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/bed-types` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/blood-banks` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/blood-donors` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/blood-donations` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/blood-issues` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/documents` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/document-types` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/doctors` | fail | 422 `/api/hms/doctor-absences` | page.goto: Timeout 45000ms exceeded. |
| `/modules/doctor-departments` | fail | 422 `/api/hms/doctor-absences` | page.goto: Timeout 45000ms exceeded. |
| `/modules/schedules` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/doctor-holiday` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/lunch-breaks` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/prescriptions` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/patient-diagnosis-test` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/diagnosis-categories` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/enquiries` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/incomes` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/expenses` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/call-logs` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/visitors` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/postals` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/complaints` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/front-settings` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/services` | fail | 404 `/api/hms/insurances`, 404 `/api/hms/packages` | page.goto: Timeout 45000ms exceeded. |
| `/modules/notice-boards` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/testimonials` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/charge-categories` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/charges` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/doctor-opd-charges` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/item-categories` | fail | 422 `/api/hms/inventory/movements` | page.goto: Timeout 45000ms exceeded. |
| `/modules/items` | fail | 422 `/api/hms/inventory/movements` | page.goto: Timeout 45000ms exceeded. |
| `/modules/item-stocks` | fail | 422 `/api/hms/inventory/movements` | page.goto: Timeout 45000ms exceeded. |
| `/modules/issued-items` | fail | 422 `/api/hms/inventory/movements` | page.goto: Timeout 45000ms exceeded. |
| `/modules/live-consultations` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/live-consultations-live-meetings` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/goole-meet-consultation` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/categories` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/brands` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/medicines` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/purchase-medicines` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/used-medicine` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/medicine-bills` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/patients` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/patient-cases` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/case-handlers` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/patient-admissions` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/pathology-categories` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/pathology-units` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/pathology-parameter` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/pathology-tests` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/birth-reports` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/death-reports` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/investigation-reports` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/operation-reports` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/radiology-categories` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/radiology-tests` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/insurances` | fail | 404 `/api/hms/insurances`, 404 `/api/hms/packages` | page.goto: Timeout 45000ms exceeded. |
| `/modules/packages` | fail | 404 `/api/hms/insurances`, 404 `/api/hms/packages` | page.goto: Timeout 45000ms exceeded. |
| `/modules/ambulances` | fail | 404 `/api/hms/insurances`, 404 `/api/hms/packages` | page.goto: Timeout 45000ms exceeded. |
| `/modules/ambulance-calls` | fail | 404 `/api/hms/insurances`, 404 `/api/hms/packages` | page.goto: Timeout 45000ms exceeded. |
| `/modules/sms` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/emails` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/email-template` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/settings` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/hospital-schedule` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/currency-settings` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/operation-categories` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/operations` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/payment-gateway` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/add-custom-fields` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/vaccinations` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/vaccinated-patients` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/odontogram` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/add-on` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/ipd-diagnoses` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/ipd-consultant-registers` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/ipd-operation` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/ipd-charges` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/ipd-prescriptions` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/ipd-timelines` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/ipd-payments` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/ipd-bills` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/opd-diagnoses` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/opd-prescriptions` | pass | None observed | Authenticated workspace ready; initial check only |
| `/modules/opd-timelines` | pass | None observed | Authenticated workspace ready; initial check only |
