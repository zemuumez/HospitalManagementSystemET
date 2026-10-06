# Original → PostgreSQL table coverage

Every original dump table appears exactly once. Targets are **candidate storage counterparts**, not verified field or API parity. A dash means no dedicated counterpart was identified in the 48 checked-in migrations. Grouping is architectural interpretation; original facts are in TABLES.md and inventory.json.

| Domain | Original table | PostgreSQL candidates | Required reconciliation |
| --- | --- | --- | --- |
| Identity | `users` | user,patient,patient_profile,staff_profile,staff_access,account,twoFactor | Split identity from clinical profiles; map owner_type/owner_id, active state, demographic fields and role membership explicitly; never copy password/reset/MFA secrets blindly. |
| Identity | `departments` | staff_access | Legacy departments are Spatie ROLES, not clinical departments. Fixed Go role permissions are an alternative, not equivalent editable role/permission tables; reconcile per-user overrides and multiple roles. |
| Identity | `permissions` | staff_access | Legacy departments are Spatie ROLES, not clinical departments. Fixed Go role permissions are an alternative, not equivalent editable role/permission tables; reconcile per-user overrides and multiple roles. |
| Identity | `model_has_permissions` | staff_access | Legacy departments are Spatie ROLES, not clinical departments. Fixed Go role permissions are an alternative, not equivalent editable role/permission tables; reconcile per-user overrides and multiple roles. |
| Identity | `model_has_roles` | staff_access | Legacy departments are Spatie ROLES, not clinical departments. Fixed Go role permissions are an alternative, not equivalent editable role/permission tables; reconcile per-user overrides and multiple roles. |
| Identity | `role_has_permissions` | staff_access | Legacy departments are Spatie ROLES, not clinical departments. Fixed Go role permissions are an alternative, not equivalent editable role/permission tables; reconcile per-user overrides and multiple roles. |
| Identity | `admins` | staff_access,staff_profile | Consolidated role-specific identities; preserve old role-table IDs in import mapping, active/default/soft-delete semantics and payroll ownership. |
| Identity | `accountants` | staff_access,staff_profile | Consolidated role-specific identities; preserve old role-table IDs in import mapping, active/default/soft-delete semantics and payroll ownership. |
| Identity | `case_handlers` | staff_access,staff_profile | Consolidated role-specific identities; preserve old role-table IDs in import mapping, active/default/soft-delete semantics and payroll ownership. |
| Identity | `lab_technicians` | staff_access,staff_profile | Consolidated role-specific identities; preserve old role-table IDs in import mapping, active/default/soft-delete semantics and payroll ownership. |
| Identity | `nurses` | staff_access,staff_profile | Consolidated role-specific identities; preserve old role-table IDs in import mapping, active/default/soft-delete semantics and payroll ownership. |
| Identity | `pharmacists` | staff_access,staff_profile | Consolidated role-specific identities; preserve old role-table IDs in import mapping, active/default/soft-delete semantics and payroll ownership. |
| Identity | `receptionists` | staff_access,staff_profile | Consolidated role-specific identities; preserve old role-table IDs in import mapping, active/default/soft-delete semantics and payroll ownership. |
| Identity | `addresses` | patient_profile,staff_profile | Legacy polymorphic owner address; verify every owner type and address field rather than attaching all addresses to patients. |
| Identity | `password_reset_tokens` | session,verification,account | Better Auth replaces Laravel token storage; invalidate old tokens and implement reset/invitation flows with the new system. |
| Identity | `personal_access_tokens` | session,verification,account | Better Auth replaces Laravel token storage; invalidate old tokens and implement reset/invitation flows with the new system. |
| Identity | `failed_jobs` | message_outbox | Job failures and retry outbox are not identical; define retained failure details, replay and dead-letter operations. |
| Identity | `migrations` | - | Framework migration history is not hospital business data; retain source provenance in an import manifest, do not copy Laravel history into PostgreSQL migration history. |
| Identity | `subscriptions` | - | Cashier billing artifacts in dump; no matching app migration here. Single-hospital scope does not require SaaS subscription operations; record explicit exclusion and preserve any historical records separately. |
| Identity | `subscription_items` | - | Cashier billing artifacts in dump; no matching app migration here. Single-hospital scope does not require SaaS subscription operations; record explicit exclusion and preserve any historical records separately. |
| Staff and scheduling | `doctors` | doctor_profile,staff_profile,staff_access | Map doctor-role identity, qualifications, specialization and consultation fee; Google credential references need separate protected integration storage. |
| Staff and scheduling | `doctor_departments` | doctor_department | Clinical departments; distinct from legacy role table departments. Compare editable fields and delete-in-use behavior. |
| Staff and scheduling | `doctor_opd_charges` | doctor_profile,patient_service_charge | Doctor-specific OPD price records need explicit effective/default-price mapping and historical charge snapshots. |
| Staff and scheduling | `schedules` | doctor_hours | Recurring slots/day schedules; map per-patient duration, available_from/to and enabled days; preserve slot calculation semantics. |
| Staff and scheduling | `schedule_days` | doctor_hours | Recurring slots/day schedules; map per-patient duration, available_from/to and enabled days; preserve slot calculation semantics. |
| Staff and scheduling | `doctor_holidays` | doctor_absence | Map date ranges, reasons and doctor identities; confirm overlap and cancellation behavior. |
| Staff and scheduling | `lunch_breaks` | doctor_hours,doctor_absence | Recurring intra-day break is not automatically a whole-day absence; dedicated break intervals may be required. |
| Staff and scheduling | `hospital_schedules` | hospital_schedule_day,hospital_hours,hospital_date_override | Two current recurring-hour representations exist; choose a single authoritative schedule and map day/time formats. |
| Staff and scheduling | `appointments` | appointment,appointment_billing,public_appointment_request | Map all booking states, department/doctor relation, charge/payment status and custom fields; keep public requests separate from confirmed appointments. |
| Staff and scheduling | `patient_queues` | patient_queue | Preserve appointment linkage, display number and day/doctor queue semantics; unique ID is not queue number. |
| Staff and scheduling | `appointment_transactions` | appointment_billing,payment_checkout,payment_event,payment_receipt | Provider settlement mapping required; idempotent callback and amount/currency verification must survive retries. |
| Patients | `patients` | patient,patient_profile,patient_smart_card,patient_identity_alias | Legacy clinical ID differs from user ID. Preserve patient_unique_id, custom-field values and card-template reference. |
| Patients | `patient_id_card_templates` | - | Confirmed structural gap: current patient_smart_card.template_name is not a template entity with color and six visibility switches. |
| Patients | `patient_cases` | patient_case | Map business case number, patient, doctor, fee, date and status; retain case-number uniqueness independently of UUID. |
| Patients | `patient_admissions` | encounter,ipd_admission_details,bed_assignment | Separate original admission subsystem; do not assume identical to IPD register. Preserve admission number, package/insurance foreign keys, agent and guardian information. |
| Patients | `insurances` | - | Confirmed missing dedicated insurance catalog/disease-line tables; admission provider/policy text does not replace their CRUD or pricing. |
| Patients | `insurance_diseases` | - | Confirmed missing dedicated insurance catalog/disease-line tables; admission provider/policy text does not replace their CRUD or pricing. |
| Patients | `packages` | - | Confirmed missing dedicated package and service-line catalog; admission package_name/charge snapshot is insufficient for catalog CRUD. |
| Patients | `package_services` | - | Confirmed missing dedicated package and service-line catalog; admission package_name/charge snapshot is insufficient for catalog CRUD. |
| Clinical | `ipd_patient_departments` | encounter,ipd_admission_details,clinical_vitals,opd_follow_up | Merged encounter abstraction requires kind-specific contracts. Map legacy business numbers, nullable OPD case, vitals, dates, doctor, custom fields, charge, payment mode, bed and discharge/bill flags. |
| Clinical | `opd_patient_departments` | encounter,ipd_admission_details,clinical_vitals,opd_follow_up | Merged encounter abstraction requires kind-specific contracts. Map legacy business numbers, nullable OPD case, vitals, dates, doctor, custom fields, charge, payment mode, bed and discharge/bill flags. |
| Clinical | `ipd_diagnoses` | encounter_diagnosis,encounter_attachment,secure_attachment | Semantic gap: legacy report_type/report_date/description/document workflow differs from ICD diagnosis category/status; preserve both if needed. |
| Clinical | `opd_diagnoses` | encounter_diagnosis,encounter_attachment,secure_attachment | Semantic gap: legacy report_type/report_date/description/document workflow differs from ICD diagnosis category/status; preserve both if needed. |
| Clinical | `ipd_consultant_registers` | encounter_care_team,clinical_note | Consultant instruction and instructed/apply dates are not merely care-team membership; add instruction lifecycle and exact fields. |
| Clinical | `ipd_operation` | encounter_procedure,hospital_operation,operation_category | Map operation catalog/category, reference, assistants, anesthesia and dates; generic procedure name/findings is not full parity. |
| Clinical | `ipd_prescriptions` | prescription,medication_order | Preserve prescription headers per encounter, header/footer and status; define bridge to executable medication orders. |
| Clinical | `opd_prescriptions` | prescription,medication_order | Preserve prescription headers per encounter, header/footer and status; define bridge to executable medication orders. |
| Clinical | `ipd_prescription_items` | prescription_medicine,medication_order | Map category, medicine, dosage, dose interval/duration, instructions and meal fields; no unstructured concatenation that loses editability. |
| Clinical | `opd_prescription_items` | prescription_medicine,medication_order | Map category, medicine, dosage, dose interval/duration, instructions and meal fields; no unstructured concatenation that loses editability. |
| Clinical | `ipd_timelines` | clinical_note,encounter_attachment,secure_attachment | Need timeline title/date/description/document and patient-visible flag; generic clinical notes do not prove original visibility/editing behavior. |
| Clinical | `opd_timelines` | clinical_note,encounter_attachment,secure_attachment | Need timeline title/date/description/document and patient-visible flag; generic clinical notes do not prove original visibility/editing behavior. |
| Clinical | `ipd_charges` | patient_service_charge,invoice_line,hospital_charge | Map category/type/code/date and applied-charge snapshot; ensure encounter FK and original CRUD with ledger safeguards. |
| Clinical | `ipd_payments` | invoice_payment,payment_receipt | Encounter payment needs encounter association before/after invoice issuance; map method, date, notes and receipt attachment without forcing wrong lifecycle. |
| Clinical | `ipd_bills` | encounter_billing,invoice,invoice_line,invoice_payment | Bill summary/discount/tax/other charges/gross/net and discharge transaction need reconciliation; current encounter summary has different semantics. |
| Clinical | `odontograms` | patient_odontogram_entry | Confirmed structural mismatch: original chart header plus JSON has many charts per patient; current unique(patient_id,tooth_number) retains one current condition per tooth. Add chart/history parent and exact tooth-code mapping. |
| Clinical | `prescriptions` | prescription | General prescription header exists; audit every medical-history/vital/advice field, ownership, next visit and status. |
| Clinical | `prescriptions_medicines` | prescription_medicine | General prescription lines; compare dosage/day/time/comment and medicine snapshot fields separately from IPD/OPD items. |
| Diagnostics | `pathology_categories` | diagnostic_category | Consolidation requires modality discriminator and original catalog separation in dropdowns and authorization. |
| Diagnostics | `radiology_categories` | diagnostic_category | Consolidation requires modality discriminator and original catalog separation in dropdowns and authorization. |
| Diagnostics | `pathology_units` | diagnostic_unit | Map unit name and foreign-key use; prevent deleting units referenced by parameters/results. |
| Diagnostics | `pathology_parameters` | diagnostic_parameter | Map unit, normal range, description and reusable master semantics; current test-specific parameters may need a reusable master bridge. |
| Diagnostics | `pathology_parameter_items` | diagnostic_parameter | Legacy test-to-parameter association includes patient_result; separate catalog membership from patient result values without losing report data. |
| Diagnostics | `pathology_tests` | diagnostic_test,diagnostic_order | Separate catalog definition from patient test/order and charge snapshot; original pathology patient linkage must not be discarded. |
| Diagnostics | `radiology_tests` | diagnostic_test,diagnostic_order | Separate catalog definition from patient test/order and charge snapshot; original pathology patient linkage must not be discarded. |
| Diagnostics | `diagnosis_categories` | diagnostic_category,diagnosis_template | General diagnosis categories are a third workflow, not automatically pathology/radiology categories. |
| Diagnostics | `patient_diagnosis_tests` | diagnostic_order,diagnostic_result,diagnostic_value,diagnosis_template | Patient diagnosis report number and arbitrary property-name/value rows need explicit storage/printing contract; a template alone is not a completed report. |
| Diagnostics | `patient_diagnosis_properties` | diagnostic_order,diagnostic_result,diagnostic_value,diagnosis_template | Patient diagnosis report number and arbitrary property-name/value rows need explicit storage/printing contract; a template alone is not a completed report. |
| Diagnostics | `investigation_reports` | investigation_report,secure_attachment | Compare patient/date/doctor/status/report attachment and authorized download/print. |
| Diagnostics | `birth_reports` | birth_report,death_report,operation_report | Map patient/case/doctor relations, report dates and descriptive fields; preserve corrections and print snapshots. |
| Diagnostics | `death_reports` | birth_report,death_report,operation_report | Map patient/case/doctor relations, report dates and descriptive fields; preserve corrections and print snapshots. |
| Diagnostics | `operation_reports` | birth_report,death_report,operation_report | Map patient/case/doctor relations, report dates and descriptive fields; preserve corrections and print snapshots. |
| Diagnostics | `vaccinations` | vaccine_catalog,patient_vaccination | Separate master from administered dose; map serial/dose/date and patient vaccination records. |
| Diagnostics | `vaccinated_patients` | vaccine_catalog,patient_vaccination | Separate master from administered dose; map serial/dose/date and patient vaccination records. |
| Pharmacy | `categories` | medicine_category,medicine_brand | Medicine masters; categories is not inventory or charge category. |
| Pharmacy | `brands` | medicine_category,medicine_brand | Medicine masters; categories is not inventory or charge category. |
| Pharmacy | `medicines` | medicine,medicine_batch | Catalog vs stock/batch split; compare purchase/sale prices, salt/composition, quantity and expiry semantics. |
| Pharmacy | `purchase_medicines` | medicine_batch,pharmacy_movement | Confirmed missing dedicated purchase header/lines: batch receipts alone lack purchase number, tax/discount, payment status/type and notes workflow. |
| Pharmacy | `purchased_medicines` | medicine_batch,pharmacy_movement | Confirmed missing dedicated purchase header/lines: batch receipts alone lack purchase number, tax/discount, payment status/type and notes workflow. |
| Pharmacy | `medicine_bills` | pharmacy_invoice,invoice,invoice_line,pharmacy_movement | Original sale header/lines include tax, discount, patient/doctor, polymorphic owner, expiry and payment type/status; reconcile grouped sale and batch consumption. |
| Pharmacy | `sale_medicines` | pharmacy_invoice,invoice,invoice_line,pharmacy_movement | Original sale header/lines include tax, discount, patient/doctor, polymorphic owner, expiry and payment type/status; reconcile grouped sale and batch consumption. |
| Pharmacy | `used_medicines` | pharmacy_movement | Track prescription/operation consumption with reason and source reference; immutable movement rather than mutable balance only. |
| Inventory | `item_categories` | inventory_category,inventory_item | Map unit, description and catalog status; original item category distinct from medicine category. |
| Inventory | `items` | inventory_category,inventory_item | Map unit, description and catalog status; original item category distinct from medicine category. |
| Inventory | `item_stocks` | inventory_movement | Stock receipt document/supplier/store/date fields may require receipt entity beyond movement quantities. |
| Inventory | `issued_items` | inventory_movement,staff_access | Issue/return owner, department, date/return date and status need explicit lifecycle; cannot represent all as anonymous quantity deductions. |
| Blood bank | `blood_banks` | blood_bank,blood_donor,blood_donation,blood_issue | Table family exists; verify donor identity, blood group, bag balances, issue charges, patient/doctor relation and concurrency. |
| Blood bank | `blood_donors` | blood_bank,blood_donor,blood_donation,blood_issue | Table family exists; verify donor identity, blood group, bag balances, issue charges, patient/doctor relation and concurrency. |
| Blood bank | `blood_donations` | blood_bank,blood_donor,blood_donation,blood_issue | Table family exists; verify donor identity, blood group, bag balances, issue charges, patient/doctor relation and concurrency. |
| Blood bank | `blood_issues` | blood_bank,blood_donor,blood_donation,blood_issue | Table family exists; verify donor identity, blood group, bag balances, issue charges, patient/doctor relation and concurrency. |
| Finance | `accounts` | charge_account | Legacy financial accounts/payee usage differs from charge catalog; examine account type/status and ledger classification. |
| Finance | `invoices` | invoice,invoice_line | Map legacy business number, account lines, quantity/price/discount/status/currency and printable totals; compare cancellation vs destructive delete. |
| Finance | `invoice_items` | invoice,invoice_line | Map legacy business number, account lines, quantity/price/discount/status/currency and printable totals; compare cancellation vs destructive delete. |
| Finance | `bills` | invoice,invoice_line,encounter_billing | Original standalone bills/admission link are distinct from invoices and IPD bills; need source-type discriminator/adapters and preserved bill numbers. |
| Finance | `bill_items` | invoice,invoice_line,encounter_billing | Original standalone bills/admission link are distinct from invoices and IPD bills; need source-type discriminator/adapters and preserved bill numbers. |
| Finance | `payments` | hospital_expense,invoice_payment | Legacy payments are payment_date/account_id/pay_to outgoings, NOT automatically patient receipts. Dedicated payable/payment contract needed. |
| Finance | `advanced_payments` | patient_advance_payment | Base table exists in 047; receipt string/currency/date parity plus operational API, allocations, unallocated balance, reversals and print remain to verify/build. |
| Finance | `transactions` | payment_checkout,payment_event,payment_receipt | Map invoice/bill provider references and settlement states; preserve historical identifiers without importing reusable tokens. |
| Finance | `bill_transactions` | payment_checkout,payment_event,payment_receipt | Map invoice/bill provider references and settlement states; preserve historical identifiers without importing reusable tokens. |
| Finance | `expenses` | hospital_expense,hospital_expense_head,hospital_income,hospital_income_head | Map legacy integer head enums to master IDs, names, reference/date/amount/description and attachments. |
| Finance | `incomes` | hospital_expense,hospital_expense_head,hospital_income,hospital_income_head | Map legacy integer head enums to master IDs, names, reference/date/amount/description and attachments. |
| Finance | `employee_payrolls` | employee_payroll | Map polymorphic staff owner, payroll number/month/year/salary/allowance/deduction/net/status and print; define immutable posted payroll. |
| Services | `charge_categories` | charge_category,hospital_charge | Map type/category/code/standard charge and snapshot applied charges; preserve original dropdown dependencies. |
| Services | `charges` | charge_category,hospital_charge | Map type/category/code/standard charge and snapshot applied charges; preserve original dropdown dependencies. |
| Services | `services` | hospital_service | Service master and package membership; compare quantity/rate/status and currency. |
| Services | `operation_categories` | operation_category,hospital_operation | Operation master family; map charge/catalog association and in-use deletion behavior. |
| Services | `operations` | operation_category,hospital_operation | Operation master family; map charge/catalog association and in-use deletion behavior. |
| Services | `ambulances` | ambulance,ambulance_call,ambulance_call_invoice | Compare vehicle identifiers, availability/driver/contact, patient/date/amount and billed-call ledger link. |
| Services | `ambulance_calls` | ambulance,ambulance_call,ambulance_call_invoice | Compare vehicle identifiers, availability/driver/contact, patient/date/amount and billed-call ledger link. |
| Front office | `visitors` | hospital_visitor,hospital_call_log,hospital_postal | Map visitor ID/proof, purpose, contacts, dates/times, attachments and incoming/outgoing postal distinction. |
| Front office | `call_logs` | hospital_visitor,hospital_call_log,hospital_postal | Map visitor ID/proof, purpose, contacts, dates/times, attachments and incoming/outgoing postal distinction. |
| Front office | `postals` | hospital_visitor,hospital_call_log,hospital_postal | Map visitor ID/proof, purpose, contacts, dates/times, attachments and incoming/outgoing postal distinction. |
| Files | `documents` | secure_attachment,encounter_attachment | Missing dedicated document-type/document register semantics; binary storage alone lacks title/type/patient/uploader listing contract. |
| Files | `document_types` | secure_attachment,encounter_attachment | Missing dedicated document-type/document register semantics; binary storage alone lacks title/type/patient/uploader listing contract. |
| Files | `media` | secure_attachment | Legacy polymorphic media collections and conversions must map each owner/collection to secured file access, not public URLs. |
| CMS | `enquiries` | hospital_enquiry,hospital_notice_board,cms_testimonial,hospital_complaint | Map publication/status, attachments, patient/account identity and resolution; legacy complaints.patient_id FK targets users, not patients. |
| CMS | `notice_boards` | hospital_enquiry,hospital_notice_board,cms_testimonial,hospital_complaint | Map publication/status, attachments, patient/account identity and resolution; legacy complaints.patient_id FK targets users, not patients. |
| CMS | `testimonials` | hospital_enquiry,hospital_notice_board,cms_testimonial,hospital_complaint | Map publication/status, attachments, patient/account identity and resolution; legacy complaints.patient_id FK targets users, not patients. |
| CMS | `complaints` | hospital_enquiry,hospital_notice_board,cms_testimonial,hospital_complaint | Map publication/status, attachments, patient/account identity and resolution; legacy complaints.patient_id FK targets users, not patients. |
| CMS | `front_services` | hospital_service,front_cms_setting | Public services include presentation/media; cannot assume hospital clinical service catalog is interchangeable. |
| CMS | `front_settings` | front_cms_setting | Preserve original CMS keys/localized content, images, terms, appointment/map sections and public reads. |
| Settings | `settings` | hospital_general_setting | Map allowed keys, data types, language, branding and provider config; secrets must not be returned by general settings APIs. |
| Settings | `currency_settings` | hospital_general_setting | No dedicated currency master table; preserve code/name/icon and active/soft-delete behavior if currency CRUD remains in UI. |
| Settings | `modules` | hospital_module_setting | Module visibility is not authorization; map original stable module names/order and role restrictions. |
| Settings | `add_custom_fields` | custom_field | Definitions exist; enforce typed values per target record, required flags, options/grid, validation and historical definition changes. |
| Settings | `add_ons` | - | No dedicated add-on catalog/licensing table; record scope of original AddOn screen and build metadata contract without executing uploaded arbitrary code. |
| Messaging | `notifications` | message_outbox | Delivery outbox is not a user notification inbox: needs recipient/read state/type/deep link and ownership. |
| Messaging | `mails` | message_outbox | Map send/history/recipients/templates/status and delivery attempts; operational SMS separate from Firebase authentication. |
| Messaging | `sms` | message_outbox | Map send/history/recipients/templates/status and delivery attempts; operational SMS separate from Firebase authentication. |
| Messaging | `email_templates` | - | No dedicated email-template catalog table; preserve template name/subject/body/variables and safe preview/update/version behavior. |
| Integrations | `live_consultations` | live_consultation,live_meeting,live_meeting_candidate | Compare provider meeting IDs, participants, dates, status, authorized join/start and linked IPD/OPD entity. |
| Integrations | `live_meetings` | live_consultation,live_meeting,live_meeting_candidate | Compare provider meeting IDs, participants, dates, status, authorized join/start and linked IPD/OPD entity. |
| Integrations | `live_meetings_candidates` | live_consultation,live_meeting,live_meeting_candidate | Compare provider meeting IDs, participants, dates, status, authorized join/start and linked IPD/OPD entity. |
| Integrations | `user_zoom_credential` | live_consultation_provider_setting | Per-user secret/OAuth lifetime differs from shared provider setting; design encrypted secret references and refresh/revoke flow. |
| Integrations | `zoom_o_auth_credentials` | live_consultation_provider_setting | Per-user secret/OAuth lifetime differs from shared provider setting; design encrypted secret references and refresh/revoke flow. |
| Integrations | `google_calendar_integrations` | - | No dedicated calendar OAuth/list/event mapping tables; choose supported integration scope and implement sync ownership/retry/deletion semantics. |
| Integrations | `google_calendar_lists` | - | No dedicated calendar OAuth/list/event mapping tables; choose supported integration scope and implement sync ownership/retry/deletion semantics. |
| Integrations | `event_google_calendars` | - | No dedicated calendar OAuth/list/event mapping tables; choose supported integration scope and implement sync ownership/retry/deletion semantics. |
| Integrations | `user_google_event_schedules` | - | No dedicated calendar OAuth/list/event mapping tables; choose supported integration scope and implement sync ownership/retry/deletion semantics. |
| Beds | `bed_types` | bed_type,hospital_bed,bed_assignment,bed_event,bed_state_event | Preserve bed business ID/type/rate/description and assignment dates/patient/IPD references; compute occupancy transactionally and retain transfer/discharge history. |
| Beds | `beds` | bed_type,hospital_bed,bed_assignment,bed_event,bed_state_event | Preserve bed business ID/type/rate/description and assignment dates/patient/IPD references; compute occupancy transactionally and retain transfer/discharge history. |
| Beds | `bed_assigns` | bed_type,hospital_bed,bed_assignment,bed_event,bed_state_event | Preserve bed business ID/type/rate/description and assignment dates/patient/IPD references; compute occupancy transactionally and retain transfer/discharge history. |
