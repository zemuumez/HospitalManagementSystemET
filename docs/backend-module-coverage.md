# Original module backend and integration coverage

Checkpoint: 2026-10-04, through commit `21e2f55`. See [the master checklist](backend-delivery-checklist.md) for acceptance criteria and QA evidence.

This inventory contains 111 source-catalog screens and 5 additional frontend screens. Each row remains open for full field/action parity; a partial backend is not a completed module. Source group/order is retained instead of alphabetic sorting.

| Screen / route identifier | Group | Backend and integration state | Full parity / QA |
|---|---|---|---|
| Patient ID Card Templates (`patient-id-card-template`) | Patient ID Card | Preview only; typed backend, integration and workflow QA pending | Open |
| Generate Patient ID Card (`generate-patient-id-card`) | Patient ID Card | Preview only; typed backend, integration and workflow QA pending | Open |
| Users (`users`) | Users | Provision/activate/deactivate; profiles and invitations pending | Open |
| Admins (`admins`) | Users | Preview only; typed backend, integration and workflow QA pending | Open |
| Accountants (`accountants`) | Users | Preview only; typed backend, integration and workflow QA pending | Open |
| Nurses (`nurses`) | Users | Preview only; typed backend, integration and workflow QA pending | Open |
| Lab Technicians (`lab-technicians`) | Users | Preview only; typed backend, integration and workflow QA pending | Open |
| Receptionists (`receptionists`) | Users | Preview only; typed backend, integration and workflow QA pending | Open |
| Pharmacists (`pharmacists`) | Users | Preview only; typed backend, integration and workflow QA pending | Open |
| Appointments (`appointments`) | Appointments | Booking, ownership, status and SMS queue; calendar/fees/reminders pending | Open |
| Appointment Calendars (`appointment-calendars`) | Appointments | Preview only; typed backend, integration and workflow QA pending | Open |
| Appointment Transaction (`appointment-transaction`) | Appointments | Preview only; typed backend, integration and workflow QA pending | Open |
| Patient Queues (`patient-queues`) | Appointments | Preview only; typed backend, integration and workflow QA pending | Open |
| IPD Patients (`ipd-patient-departments`) | IPD / OPD | Admission, bed allocation, signed notes and discharge; submodules pending | Open |
| OPD Patients (`opd-patient-departments`) | IPD / OPD | Encounter, signed notes and discharge; repeat visits/submodules pending | Open |
| Accounts (`accounts`) | Billing | Charge-account contracts in progress; unverified | Open |
| Employee Payrolls (`employee-payrolls`) | Billing | Preview only; typed backend, integration and workflow QA pending | Open |
| Invoices (`invoices`) | Billing | Billing contracts in progress; unverified | Open |
| Payments (`payments`) | Billing | Preview only; typed backend, integration and workflow QA pending | Open |
| Payment Reports (`payment-reports`) | Billing | Preview only; typed backend, integration and workflow QA pending | Open |
| Advanced Payments (`advanced-payments`) | Billing | Preview only; typed backend, integration and workflow QA pending | Open |
| Bills (`bills`) | Billing | Preview only; typed backend, integration and workflow QA pending | Open |
| Manual Payment Approvals (`manual-bill-payments`) | Billing | Preview only; typed backend, integration and workflow QA pending | Open |
| Bed Status (`bed-status`) | Bed Management | Availability list; original visual bed board integration pending | Open |
| Bed Assigns (`bed-assigns`) | Bed Management | Preview only; typed backend, integration and workflow QA pending | Open |
| Beds (`beds`) | Bed Management | Create/list and derived occupancy; types/maintenance/transfers pending | Open |
| Bed Types (`bed-types`) | Bed Management | Preview only; typed backend, integration and workflow QA pending | Open |
| Blood Banks (`blood-banks`) | Blood Bank | Preview only; typed backend, integration and workflow QA pending | Open |
| Blood Donors (`blood-donors`) | Blood Bank | Preview only; typed backend, integration and workflow QA pending | Open |
| Blood Donations (`blood-donations`) | Blood Bank | Preview only; typed backend, integration and workflow QA pending | Open |
| Blood Issues (`blood-issues`) | Blood Bank | Preview only; typed backend, integration and workflow QA pending | Open |
| Documents (`documents`) | Documents | Preview only; typed backend, integration and workflow QA pending | Open |
| Document Types (`document-types`) | Documents | Preview only; typed backend, integration and workflow QA pending | Open |
| Doctors (`doctors`) | Doctors | Profiles exist for scheduling; original Doctors screen still preview | Open |
| Doctor Departments (`doctor-departments`) | Doctors | Preview only; typed backend, integration and workflow QA pending | Open |
| Schedules (`schedules`) | Doctors | Weekly periods and slots; holidays/absence UI pending | Open |
| Doctor Holidays (`doctor-holiday`) | Doctors | Preview only; typed backend, integration and workflow QA pending | Open |
| Lunch Breaks (`lunch-breaks`) | Doctors | Preview only; typed backend, integration and workflow QA pending | Open |
| Prescriptions (`prescriptions`) | Prescriptions | Preview only; typed backend, integration and workflow QA pending | Open |
| Diagnosis Tests (`patient-diagnosis-test`) | Diagnosis | Preview only; typed backend, integration and workflow QA pending | Open |
| Diagnosis Categories (`diagnosis-categories`) | Diagnosis | Preview only; typed backend, integration and workflow QA pending | Open |
| Enquiries (`enquiries`) | Enquiries | Preview only; typed backend, integration and workflow QA pending | Open |
| Incomes (`incomes`) | Finance | Preview only; typed backend, integration and workflow QA pending | Open |
| Expenses (`expenses`) | Finance | Preview only; typed backend, integration and workflow QA pending | Open |
| Call Logs (`call-logs`) | Front Office | Preview only; typed backend, integration and workflow QA pending | Open |
| Visitors (`visitors`) | Front Office | Preview only; typed backend, integration and workflow QA pending | Open |
| Postal Receive / Dispatch (`postals`) | Front Office | Preview only; typed backend, integration and workflow QA pending | Open |
| Complaints (`complaints`) | Front Office | Preview only; typed backend, integration and workflow QA pending | Open |
| Front Settings (`front-settings`) | Front CMS | Preview only; typed backend, integration and workflow QA pending | Open |
| Services (`services`) | Front CMS | Preview only; typed backend, integration and workflow QA pending | Open |
| Notice Boards (`notice-boards`) | Front CMS | Preview only; typed backend, integration and workflow QA pending | Open |
| Testimonials (`testimonials`) | Front CMS | Preview only; typed backend, integration and workflow QA pending | Open |
| Charge Categories (`charge-categories`) | Hospital Charges | Preview only; typed backend, integration and workflow QA pending | Open |
| Charges (`charges`) | Hospital Charges | Preview only; typed backend, integration and workflow QA pending | Open |
| Doctor OPD Charges (`doctor-opd-charges`) | Hospital Charges | Preview only; typed backend, integration and workflow QA pending | Open |
| Item Categories (`item-categories`) | Inventory | Preview only; typed backend, integration and workflow QA pending | Open |
| Items (`items`) | Inventory | Preview only; typed backend, integration and workflow QA pending | Open |
| Item Stocks (`item-stocks`) | Inventory | Preview only; typed backend, integration and workflow QA pending | Open |
| Issued Items (`issued-items`) | Inventory | Preview only; typed backend, integration and workflow QA pending | Open |
| Live Consultations (`live-consultations`) | Live Consultations | Preview only; typed backend, integration and workflow QA pending | Open |
| Live Meetings (`live-consultations-live-meetings`) | Live Consultations | Preview only; typed backend, integration and workflow QA pending | Open |
| Google Meet Consultations (`goole-meet-consultation`) | Live Consultations | Preview only; typed backend, integration and workflow QA pending | Open |
| Medicine Categories (`categories`) | Medicine | Preview only; typed backend, integration and workflow QA pending | Open |
| Medicine Brands (`brands`) | Medicine | Preview only; typed backend, integration and workflow QA pending | Open |
| Medicines (`medicines`) | Medicine | Preview only; typed backend, integration and workflow QA pending | Open |
| Medicine Purchases (`purchase-medicines`) | Medicine | Preview only; typed backend, integration and workflow QA pending | Open |
| Used Medicines (`used-medicine`) | Medicine | Preview only; typed backend, integration and workflow QA pending | Open |
| Medicine Bills (`medicine-bills`) | Medicine | Preview only; typed backend, integration and workflow QA pending | Open |
| Patients (`patients`) | Patients | Essential registration and portal/doctor linkage; full fields/edit pending | Open |
| Patient Cases (`patient-cases`) | Patients | Create/list with patient and doctor; edit/closure/full case fields pending | Open |
| Case Handlers (`case-handlers`) | Patients | Preview only; typed backend, integration and workflow QA pending | Open |
| Patient Admissions (`patient-admissions`) | Patients | Preview only; typed backend, integration and workflow QA pending | Open |
| Pathology Categories (`pathology-categories`) | Pathology | Preview only; typed backend, integration and workflow QA pending | Open |
| Pathology Units (`pathology-units`) | Pathology | Preview only; typed backend, integration and workflow QA pending | Open |
| Pathology Parameters (`pathology-parameter`) | Pathology | Preview only; typed backend, integration and workflow QA pending | Open |
| Pathology Tests (`pathology-tests`) | Pathology | Preview only; typed backend, integration and workflow QA pending | Open |
| Birth Reports (`birth-reports`) | Reports | Preview only; typed backend, integration and workflow QA pending | Open |
| Death Reports (`death-reports`) | Reports | Preview only; typed backend, integration and workflow QA pending | Open |
| Investigation Reports (`investigation-reports`) | Reports | Preview only; typed backend, integration and workflow QA pending | Open |
| Operation Reports (`operation-reports`) | Reports | Preview only; typed backend, integration and workflow QA pending | Open |
| Radiology Categories (`radiology-categories`) | Radiology | Preview only; typed backend, integration and workflow QA pending | Open |
| Radiology Tests (`radiology-tests`) | Radiology | Preview only; typed backend, integration and workflow QA pending | Open |
| Insurances (`insurances`) | Services | Preview only; typed backend, integration and workflow QA pending | Open |
| Packages (`packages`) | Services | Preview only; typed backend, integration and workflow QA pending | Open |
| Ambulances (`ambulances`) | Services | Preview only; typed backend, integration and workflow QA pending | Open |
| Ambulance Calls (`ambulance-calls`) | Services | Preview only; typed backend, integration and workflow QA pending | Open |
| SMS (`sms`) | SMS / Mail | Outbox available through Communications; this original screen still preview | Open |
| Emails (`emails`) | SMS / Mail | SMTP outbox available through Communications; this original screen still preview | Open |
| Email Template (`email-template`) | SMS / Mail | Preview only; typed backend, integration and workflow QA pending | Open |
| General Settings (`settings`) | Settings | Preview only; typed backend, integration and workflow QA pending | Open |
| Hospital Schedule (`hospital-schedule`) | Settings | Preview only; typed backend, integration and workflow QA pending | Open |
| Currencies (`currency-settings`) | Settings | Preview only; typed backend, integration and workflow QA pending | Open |
| Operation Categories (`operation-categories`) | Settings | Preview only; typed backend, integration and workflow QA pending | Open |
| Operations (`operations`) | Settings | Preview only; typed backend, integration and workflow QA pending | Open |
| Payment Gateway (`payment-gateway`) | Settings | Preview only; typed backend, integration and workflow QA pending | Open |
| Custom Fields (`add-custom-fields`) | Settings | Preview only; typed backend, integration and workflow QA pending | Open |
| Vaccinations (`vaccinations`) | Vaccinations | Preview only; typed backend, integration and workflow QA pending | Open |
| Vaccinated Patients (`vaccinated-patients`) | Vaccinations | Preview only; typed backend, integration and workflow QA pending | Open |
| Odontogram (`odontogram`) | Odontogram | Preview only; typed backend, integration and workflow QA pending | Open |
| Add On (`add-on`) | Add-ons | Preview only; typed backend, integration and workflow QA pending | Open |
| IPD Diagnoses (`ipd-diagnoses`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| IPD Consultant Registers (`ipd-consultant-registers`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| IPD Operation (`ipd-operation`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| IPD Charges (`ipd-charges`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| IPD Prescriptions (`ipd-prescriptions`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| IPD Timelines (`ipd-timelines`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| IPD Payments (`ipd-payments`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| IPD Bills (`ipd-bills`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| OPD Diagnoses (`opd-diagnoses`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| OPD Prescriptions (`opd-prescriptions`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| OPD Timelines (`opd-timelines`) | Clinical Records | Preview only; typed backend, integration and workflow QA pending | Open |
| Attendance (`attendance`) | Attendance | Preview only; typed backend, integration and workflow QA pending | Open |
| Manage Attendance (`manage-attendance`) | Manage Attendance | Preview only; typed backend, integration and workflow QA pending | Open |
| Modules Setting (`modules-setting`) | Settings | Preview only; typed backend, integration and workflow QA pending | Open |
| Patient Queue Theme (`patient-queue-theme`) | Settings | Preview only; typed backend, integration and workflow QA pending | Open |
| Front CMS Services (`front-cms-services`) | Front CMS | Preview only; typed backend, integration and workflow QA pending | Open |

## Additional surfaces outside the module catalog

| Surface | Current state | Missing |
|---|---|---|
| Public home/content/doctors | Original-style frontend previews | Persist CMS, real public doctors, asset handling and publishing |
| Public appointment/contact/register | Preview form behavior | Verified booking/registration/enquiry flows and anti-abuse |
| Login/reset/account | Better Auth and phone bridge foundation | Firebase success path, recovery completion, MFA and full account lifecycle |
| Patient portal | Appointments/cases/IPD/OPD connected | Bills/invoices/reports/prescriptions/documents/vaccines and full summaries |
| Staff role dashboards | Mostly preview summaries | Real role-scoped aggregates and actionable work queues |
| Communications | Persistent authorized outbox | Callbacks, preferences, templates, scheduling and reconciliation |
| Security preview | Demonstration only | Replace with real security controls; never use demo codes as authentication |

## Closure rule

A row closes only after source contracts, migrations, service rules, authorization, original UI integration, errors/empty/loading states, localization, exports/files when applicable and relevant unit/database/API/browser tests are verified. Update this table after each committed implementation step.
