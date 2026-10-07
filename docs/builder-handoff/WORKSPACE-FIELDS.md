# Original screen field inventory

This appendix contains all 111 records in the committed source-derived catalog. It is not a claim that every dynamic/conditional field was extracted. Read MODEL-RULES.md and table definitions for server requirements. `extracted: false` records are not source-confirmed forms. Current target overrides and newer attendance screens are supplied separately in reference/legacy.ts; those overrides are implementation evidence, not proof of original rules.

## Patient ID Card Templates (`patient-id-card-template`)

Group: Patient ID Card. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/patient_id_card_template`. Table pointer: `review/legacy/app/Livewire/PatientIdCardTemplateTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| color | Color | color | True |

Table columns: ["Name", "Color", "Email", "Phone", "Date of Birth", "Blood Group", "Address", "Patient Unique ID"].

## Generate Patient ID Card (`generate-patient-id-card`)

Group: Patient ID Card. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/generate_patient_id_card`. Table pointer: `review/legacy/app/Livewire/GeneratePatientIdCardTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| template_id | Select Template | select | True |
| patient_id | Select Patient | select | True |

Table columns: ["Patients", "Patient Unique ID", "Template ID"].

## Users (`users`)

Group: Users. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/users`. Table pointer: `review/legacy/app/Livewire/UserTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| first_name | First Name | text | True |
| last_name | Last Name | text | True |
| email | Email | email | True |
| department_id | Role | text | False |
| doctor_department | Doctor Departments | text | False |
| phone | Phone | tel | True |
| dob | Date of Birth | date | False |
| gender | Gender | select | True |
| status | Status | select | False |
| password | Password | password | True |
| password_confirmation | Password Confirmation | password | True |
| facebook_url | Facebook URL | text | False |
| instagram_url | Instagram URL | text | False |
| twitter_url | Twitter URL | text | False |
| linkedIn_url | Linked In URL | text | False |
| image | Profile | file | False |

Table columns: ["Users", "Role", "Email", "Status"].

## Admins (`admins`)

Group: Users. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/admins`. Table pointer: `review/legacy/app/Livewire/AdminTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| first_name | First Name | text | True |
| last_name | Last Name | text | True |
| email | Email | email | True |
| phone | Phone | tel | False |
| dob | Date of Birth | date | False |
| gender | Gender | select | True |
| status | Status | select | False |
| password | Password | password | True |
| password_confirmation | Password Confirmation | password | True |
| image | Profile | file | False |

Table columns: ["Name", "Phone", "Status", "Last Name", "Email"].

## Accountants (`accountants`)

Group: Users. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/accountants`. Table pointer: `review/legacy/app/Livewire/AccountantTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| first_name | First Name | text | True |
| last_name | Last Name | text | True |
| email | Email | email | True |
| phone | Phone | tel | False |
| blood_group | Blood Group | select | False |
| designation | Designation | text | True |
| qualification | Qualification | text | True |
| dob | Date of Birth | date | False |
| gender | Gender | select | True |
| status | Status | select | False |
| password | Password | password | True |
| password_confirmation | Password Confirmation | password | True |
| image | Profile | file | False |
| address1 | Address 1 | text | False |
| address2 | Address 2 | text | False |
| city | City | text | False |
| zip | Zip | text | False |

Table columns: ["Accountants", "Phone", "Status", "Last Name", "Email"].

## Nurses (`nurses`)

Group: Users. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/nurses`. Table pointer: `review/legacy/app/Livewire/NurseTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| first_name | First Name | text | True |
| last_name | Last Name | text | True |
| email | Email | email | True |
| designation | Designation | text | True |
| phone | Phone | tel | False |
| gender | Gender | select | True |
| status | Status | select | False |
| qualification | Qualification | text | True |
| dob | Date of Birth | date | False |
| blood_group | Blood Group | select | False |
| password | Password | password | True |
| password_confirmation | Password Confirmation | password | True |
| image | Profile | file | False |
| address1 | Address 1 | text | False |
| address2 | Address 2 | text | False |
| city | City | text | False |
| zip | Zip | text | False |

Table columns: ["Nurses", "Phone", "Qualification", "Birth Date", "Status"].

## Lab Technicians (`lab-technicians`)

Group: Users. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/lab_technicians`. Table pointer: `review/legacy/app/Livewire/LabTechnicianTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| first_name | First Name | text | True |
| last_name | Last Name | text | True |
| email | Email | email | True |
| designation | Designation | text | True |
| phone | Phone | tel | True |
| qualification | Qualification | text | True |
| dob | Date of Birth | date | False |
| blood_group | Blood Group | select | False |
| password | Password | password | True |
| password_confirmation | Password Confirmation | password | True |
| gender | Gender | select | True |
| status | Status | select | True |
| image | Profile | file | False |
| address1 | Address 1 | text | False |
| address2 | Address 2 | text | False |
| city | City | text | False |
| zip | Zip | text | False |

Table columns: ["Lab Technicians", "Designation", "Status"].

## Receptionists (`receptionists`)

Group: Users. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/receptionists`. Table pointer: `review/legacy/app/Livewire/ReceptionistTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| first_name | First Name | text | True |
| last_name | Last Name | text | True |
| email | Email | email | True |
| designation | Designation | text | True |
| phone | Phone | tel | False |
| gender | Gender | select | True |
| status | Status | select | False |
| qualification | Qualification | text | True |
| dob | Date of Birth | date | False |
| blood_group | Blood Group | select | False |
| password | Password | password | True |
| password_confirmation | Password Confirmation | password | True |
| image | Profile | file | False |
| address1 | Address 1 | text | False |
| address2 | Address 2 | text | False |
| city | City | text | False |
| zip | Zip | text | False |

Table columns: ["Receptionists", "Designation", "Phone", "Status", "Last Name"].

## Pharmacists (`pharmacists`)

Group: Users. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/pharmacists`. Table pointer: `review/legacy/app/Livewire/PharmacistTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| first_name | First Name | text | True |
| last_name | Last Name | text | True |
| email | Email | email | True |
| designation | Designation | text | True |
| phone | Phone | tel | True |
| qualification | Qualification | text | True |
| dob | Date of Birth | date | False |
| blood_group | Blood Group | select | False |
| password | Password | password | True |
| password_confirmation | Password Confirmation | password | True |
| gender | Gender | select | True |
| status | Status | select | False |
| image | Profile | file | False |
| address1 | Address 1 | text | False |
| address2 | Address 2 | text | False |
| city | City | text | False |
| zip | Zip | text | False |

Table columns: ["Pharmacists", "Blood Group", "Status"].

## Appointments (`appointments`)

Group: Appointments. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/appointments`. Table pointer: `review/legacy/app/Livewire/AppointmentTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| department_id | Doctor Department | select | True |
| doctor_id | Doctor | select | True |
| opd_date | Date | date | True |
| appointment_charge | Appointment Charge | text | False |
| payment_mode | Payment Mode | select | False |
| problem | Description | textarea | False |
| timeslot | Time | text | False |
| status | Status | select | False |

Table columns: ["Patient", "Doctor", "Doctor Department", "Date", "Status"].

## Appointment Calendars (`appointment-calendars`)

Group: Appointments. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/appointment_calendars`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## Appointment Transaction (`appointment-transaction`)

Group: Appointments. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/appointment_transaction`. Table pointer: `review/legacy/app/Livewire/AppointmentTransactionTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Patient", "Doctor", "Appointment Date", "Payment Status", "Payment Mode", "Amount"].

## Patient Queues (`patient-queues`)

Group: Appointments. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/patient_queues`. Table pointer: `review/legacy/app/Livewire/PatientQueueTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Patient", "Doctor", "Date", "Status"].

## IPD Patients (`ipd-patient-departments`)

Group: IPD / OPD. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/ipd_patient_departments`. Table pointer: `review/legacy/app/Livewire/IpdPatientDepartmentTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient ID | select | True |
| case_id | Case ID | select | False |
| ipd_number | IPD Number | text | False |
| height | Height | number | False |
| weight | Weight | number | False |
| bp | Bp | number | False |
| admission_date | Admission Date | date | True |
| doctor_id | Doctor ID | select | True |
| bed_type_id | Bed Type ID | select | True |
| bed_id | Bed ID | select | True |
| is_old_patient | Is Old Patient | text | False |
| symptoms | Symptoms | textarea | False |
| notes | Notes | textarea | False |

Table columns: ["IPD Number", "Doctor ID", "Admission Date", "Bed ID", "Discharge", "Bill Status", "Net Amount"].

## OPD Patients (`opd-patient-departments`)

Group: IPD / OPD. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/opd_patient_departments`. Table pointer: `review/legacy/app/Livewire/OpdPatientDepartmentTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient ID | select | True |
| case_id | Case ID | select | True |
| opd_number | OPD Number | text | False |
| height | Height | number | False |
| weight | Weight | number | False |
| bp | Bp | number | False |
| appointment_date | Appointment Date | date | True |
| doctor_id | Doctor ID | select | True |
| standard_charge | Standard Charge | text | True |
| payment_mode | Payment Mode | select | True |
| symptoms | Symptoms | textarea | False |
| notes | Notes | textarea | False |
| is_old_patient | Is Old Patient | text | False |

Table columns: ["OPD Number", "Doctor ID", "Appointment Date", "Standard Charge", "Payment Mode", "Phone", "Total Visits"].

## Accounts (`accounts`)

Group: Billing. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/accounts`. Table pointer: `review/legacy/app/Livewire/AccountTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Account | text | True |
| description | Description | textarea | False |
| status | Status | select | False |
| type | Type | select | False |

Table columns: ["Account", "Type", "Status"].

## Employee Payrolls (`employee-payrolls`)

Group: Billing. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/employee_payrolls`. Table pointer: `review/legacy/app/Livewire/EmployeePayrollTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| sr_no | Sr No | text | True |
| payroll_id | Payroll ID | text | True |
| type | Role | select | True |
| owner_id | Employee | select | True |
| month | Month | text | True |
| year | Year | text | True |
| status | Status | select | True |
| basic_salary | Basic Salary | text | True |
| allowance | Allowance | text | True |
| deductions | Deductions | text | True |
| net_salary | Net Salary | text | True |

Table columns: ["Sr No", "Payroll ID", "Employee", "Month", "Year", "Net Salary", "Status"].

## Invoices (`invoices`)

Group: Billing. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/invoices`. Table pointer: `review/legacy/app/Livewire/InvoiceTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| invoice_date | Invoice Date | date | True |
| discount | Discount | number | True |
| status | Status | select | True |

Table columns: ["Invoice ID", "Patient", "Invoice Date", "Amount", "Status"].

## Payments (`payments`)

Group: Billing. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/payments`. Table pointer: `review/legacy/app/Livewire/PaymentTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| account_id | Account | select | True |
| payment_date | Payment Date | date | True |
| pay_to | Pay To | text | True |
| amount | Amount | text | True |
| description | Description | textarea | False |

Table columns: ["Account", "Payment Date", "Pay To", "Amount"].

## Payment Reports (`payment-reports`)

Group: Billing. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/payment_reports`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## Advanced Payments (`advanced-payments`)

Group: Billing. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/advanced_payments`. Table pointer: `review/legacy/app/Livewire/AdvancedPaymentTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| receipt_no | Receipt No | text | True |
| amount | Amount | text | True |
| date | Date | date | True |

Table columns: ["Receipt No", "Patient", "Date", "Amount"].

## Bills (`bills`)

Group: Billing. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/bills`. Table pointer: `review/legacy/app/Livewire/BillTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_admission_id | Admission ID | select | True |
| bill_date | Bill Date | date | True |
| name | Patient | text | False |
| email | Patient Email | email | False |
| phone | Patient Cell No | text | False |
| gender | Patient Gender | select | False |
| dob | Patient Date of Birth | date | False |
| doctor_id | Doctor | text | False |
| admission_date | Admission Date | date | False |
| discharge_date | Discharge Date | date | False |
| package_id | Package Name | text | False |
| insurance_id | Insurance Name | text | False |
| total_days | Total Days | text | False |
| policy_no | Policy No | text | True |

Table columns: ["Payment Mode", "Bill ID", "Patient", "Bill Status", "Bill Date", "Amount"].

## Manual Payment Approvals (`manual-bill-payments`)

Group: Billing. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/manual_bill_payments`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## Bed Status (`bed-status`)

Group: Bed Management. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/bed_status`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## Bed Assigns (`bed-assigns`)

Group: Bed Management. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/bed_assigns`. Table pointer: `review/legacy/app/Livewire/BedAssignTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| case_id | Case | select | True |
| ipd_patient_department_id | IPD Patient | select | True |
| bed_id | Bed | select | True |
| assign_date | Assign Date | date | True |
| discharge_date | Discharge Date | date | False |
| description | Description | textarea | False |
| status | Status | select | False |

Table columns: ["IPD Number", "Patient", "Bed", "Assign Date", "Discharge Date", "Status"].

## Beds (`beds`)

Group: Bed Management. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/beds`. Table pointer: `review/legacy/app/Livewire/BedTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Bed | text | True |
| bed_type | Bed Type | select | True |
| charge | Charge | text | True |
| description | Description | textarea | False |

Table columns: ["Bed ID", "Bed", "Bed Type", "Charge", "Available"].

## Bed Types (`bed-types`)

Group: Bed Management. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/bed_types`. Table pointer: `review/legacy/app/Livewire/BedTypeTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| title | Bed Type | text | True |
| description | Description | textarea | False |

Table columns: ["Bed Type"].

## Blood Banks (`blood-banks`)

Group: Blood Bank. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/blood_banks`. Table pointer: `review/legacy/app/Livewire/BloodBankTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| blood_group | Blood Group | text | True |
| remained_bags | Remained Bags | number | True |

Table columns: ["Blood Group", "Remained Bags"].

## Blood Donors (`blood-donors`)

Group: Blood Bank. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/blood_donors`. Table pointer: `review/legacy/app/Livewire/BloodDonorTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| age | Age | number | True |
| gender | Gender | select | False |
| blood_group | Blood Group | select | True |
| last_donate_date | Last Donation Date | date | True |

Table columns: ["Name", "Age", "Gender", "Blood Group", "Last Donation Date"].

## Blood Donations (`blood-donations`)

Group: Blood Bank. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/blood_donations`. Table pointer: `review/legacy/app/Livewire/BloodDonationTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| blood_donor_id | Donor Name | select | True |
| bags | Bags | number | True |

Table columns: ["Donor Name", "Bags"].

## Blood Issues (`blood-issues`)

Group: Blood Bank. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/blood_issues`. Table pointer: `review/legacy/app/Livewire/BloodIssueTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| issue_date | Issue Date | date | True |
| doctor_id | Doctor Name | select | True |
| patient_id | Patient Name | select | True |
| donor_id | Donor Name | select | True |
| blood_group | Blood Group | select | True |
| amount | Amount | text | True |
| remarks | Remarks | textarea | False |

Table columns: ["Patient", "Doctor", "Donor Name", "Issue Date", "Blood Group", "Amount"].

## Documents (`documents`)

Group: Documents. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/documents`. Table pointer: `review/legacy/app/Livewire/DocumentTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| title | Title | text | True |
| document_type_id | Document Type | select | True |
| patient_id | Patient | select | True |
| file | Attachment | file | False |
| notes | Notes | textarea | False |

Table columns: ["File Name", "Document Type", "Patient"].

## Document Types (`document-types`)

Group: Documents. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/document_types`. Table pointer: `review/legacy/app/Livewire/DocumentTypeTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Document Type | text | True |

Table columns: ["Document Type"].

## Doctors (`doctors`)

Group: Doctors. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/doctors`. Table pointer: `review/legacy/app/Livewire/DoctorTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| first_name | First Name | text | True |
| last_name | Last Name | text | True |
| doctor_department_id | Doctor Department | select | True |
| email | Email | email | True |
| designation | Designation | text | True |
| phone | Phone | tel | False |
| qualification | Qualification | text | True |
| dob | Date of Birth | date | False |
| blood_group | Blood Group | select | False |
| gender | Gender | select | True |
| status | Status | select | False |
| specialist | Specialist | text | True |
| password | Password | password | True |
| password_confirmation | Password Confirmation | password | True |
| appointment_charge | Appointment Charge | text | False |
| description | Description | textarea | False |
| image | Profile | file | False |
| address1 | Address 1 | text | False |
| address2 | Address 2 | text | False |
| city | City | text | False |
| zip | Zip | text | False |
| facebook_url | Facebook URL | text | False |
| twitter_url | Twitter URL | text | False |
| instagram_url | Instagram URL | text | False |
| linkedIn_url | Linked In URL | text | False |

Table columns: ["Qualification", "Doctor", "Specialist", "Status"].

## Doctor Departments (`doctor-departments`)

Group: Doctors. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/doctor_departments`. Table pointer: `review/legacy/app/Livewire/DoctorDepartmentTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| title | Doctor Department | text | True |
| description | Description | textarea | False |

Table columns: ["Doctor Department"].

## Schedules (`schedules`)

Group: Doctors. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/schedules`. Table pointer: `review/legacy/app/Livewire/ScheduleTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| doctor_id | Doctor | select | True |
| per_patient_time | Per Patient Time | text | True |

Table columns: ["Doctor", "Per Patient Time"].

## Doctor Holidays (`doctor-holiday`)

Group: Doctors. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/doctor_holiday`. Table pointer: `review/legacy/app/Livewire/DoctorHolidayTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| doctor_id | Doctor | select | True |
| date | Date | date | False |
| name | Reason | text | False |

Table columns: ["Doctor", "Reason", "Date"].

## Lunch Breaks (`lunch-breaks`)

Group: Doctors. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/lunch_breaks`. Table pointer: `review/legacy/app/Livewire/LunchBreakTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| doctor_id | Doctor | select | True |
| break_from | From | text | True |
| break_to | To | text | True |
| date | Date | date | False |

Table columns: ["Doctor", "Lunch Break", "Date"].

## Prescriptions (`prescriptions`)

Group: Prescriptions. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/prescriptions`. Table pointer: `review/legacy/app/Livewire/PrescriptionTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| doctor_id | Doctor | select | True |
| health_insurance | Health Insurance | text | False |
| low_income | Low Income | text | False |
| reference | Reference | text | False |
| status | Status | select | False |

Table columns: ["Patients", "Doctors", "Medical History", "Status"].

## Diagnosis Tests (`patient-diagnosis-test`)

Group: Diagnosis. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/patient_diagnosis_test`. Table pointer: `review/legacy/app/Livewire/PatientDiagnosisTestTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| doctor_id | Doctor | select | True |
| category_id | Diagnosis Category | select | True |
| report_number | Report Number | text | False |
| age | Age | number | False |
| height | Height | number | False |
| weight | Weight | number | False |
| average_glucose | Average Glucose | text | False |
| fasting_blood_sugar | Fasting Blood Sugar | text | False |
| urine_sugar | Urine Sugar | text | False |
| blood_pressure | Blood Pressure | number | False |
| diabetes | Diabetes | text | False |
| cholesterol | Cholesterol | text | False |

Table columns: ["Report Number", "Patient", "Doctor", "Diagnosis Category"].

## Diagnosis Categories (`diagnosis-categories`)

Group: Diagnosis. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/diagnosis_categories`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Diagnosis Category | text | True |
| description | Description | textarea | False |

Table columns: ["Diagnosis Category"].

## Enquiries (`enquiries`)

Group: Enquiries. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/enquiries`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## Incomes (`incomes`)

Group: Finance. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/incomes`. Table pointer: `review/legacy/app/Livewire/IncomeTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| income_head | Income Head | select | True |
| name | Name | text | True |
| date | Date | date | True |
| invoice_number | Invoice Number | text | False |
| amount | Amount | text | True |
| attachment | Attachment | file | False |
| description | Description | textarea | False |

Table columns: ["Invoice Number", "Name", "Income Head", "Date", "Amount", "Attachment"].

## Expenses (`expenses`)

Group: Finance. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/expenses`. Table pointer: `review/legacy/app/Livewire/ExpenseTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| expense_head | Expense Head | select | True |
| name | Name | text | True |
| date | Date | date | True |
| invoice_number | Invoice Number | text | False |
| amount | Amount | text | True |
| attachment | Attachment | file | False |
| description | Description | textarea | False |

Table columns: ["Invoice Number", "Name", "Expense Head", "Date", "Amount", "Attachment"].

## Call Logs (`call-logs`)

Group: Front Office. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/call_logs`. Table pointer: `review/legacy/app/Livewire/CallLogTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| phone | Phone | tel | False |
| date | Received On | date | False |
| follow_up_date | Follow Up Date | date | False |
| note | Note | textarea | False |
| call_type | Call Type | select | False |

Table columns: ["Name", "Phone", "Received On", "Follow Up Date", "Call Type"].

## Visitors (`visitors`)

Group: Front Office. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/visitors`. Table pointer: `review/legacy/app/Livewire/VisitorTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| purpose | Purpose | select | True |
| name | Name | text | True |
| phone | Phone | tel | False |
| id_card | Id Card | text | False |
| no_of_person | Number Of Person | number | False |
| date | Date | date | False |
| in_time | In Time | text | False |
| out_time | Out Time | text | False |
| note | Note | textarea | False |
| attachment | Attachment | file | False |

Table columns: ["Purpose", "Name", "Phone", "Id Card", "Number Of Person", "Date", "In Time", "Out Time", "Attachment"].

## Postal Receive / Dispatch (`postals`)

Group: Front Office. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/postals`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| to_title | To Title | text | True |
| reference_no | Reference No | text | False |
| date | Date | date | False |
| from_title | From Title | text | False |
| image | Attachment | file | False |
| address | Address | textarea | False |

Table columns: ["To Title", "Reference No", "Date", "From Title"].

## Complaints (`complaints`)

Group: Front Office. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/complaints`. Table pointer: `review/legacy/app/Livewire/ComplaintTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| title | Title | text | True |
| description | Description | textarea | True |

Table columns: ["Title", "User", "Description", "Status"].

## Front Settings (`front-settings`)

Group: Front CMS. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/front_settings`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| short_description | Short Description | textarea | True |
| icon | Icon | file | True |

Table columns: ["Name"].

## Services (`services`)

Group: Front CMS. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/services`. Table pointer: `review/legacy/app/Livewire/ServiceTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Service | text | True |
| quantity | Quantity | text | True |
| rate | Rate | text | True |
| description | Description | textarea | False |
| status | Status | select | False |

Table columns: ["Service", "Quantity", "Rate", "Status"].

## Notice Boards (`notice-boards`)

Group: Front CMS. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/notice_boards`. Table pointer: `review/legacy/app/Livewire/NoticeBoardTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| title | Title | text | True |
| description | Description | textarea | False |

Table columns: ["Title"].

## Testimonials (`testimonials`)

Group: Front CMS. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/testimonials`. Table pointer: `review/legacy/app/Livewire/TestimonialTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | True |
| profile | Profile | file | False |

Table columns: ["Profile", "Name", "Description"].

## Charge Categories (`charge-categories`)

Group: Hospital Charges. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/charge_categories`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Charge Category | text | True |
| description | Description | textarea | False |
| charge_type | Charge Type | select | True |

Table columns: ["Charge Category", "Charge Type"].

## Charges (`charges`)

Group: Hospital Charges. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/charges`. Table pointer: `review/legacy/app/Livewire/ChargeTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| charge_type | Charge Type | select | True |
| charge_category_id | Charge Category | select | True |
| code | Code | text | True |
| standard_charge | Standard Charge | text | True |
| description | Description | textarea | False |

Table columns: ["Code", "Charge Category", "Charge Type", "Standard Charge"].

## Doctor OPD Charges (`doctor-opd-charges`)

Group: Hospital Charges. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/doctor_opd_charges`. Table pointer: `review/legacy/app/Livewire/DoctorOPDChargeTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| doctor_id | Doctor | select | True |
| standard_charge | Standard Charge | text | True |

Table columns: ["Doctor", "Standard Charge"].

## Item Categories (`item-categories`)

Group: Inventory. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/item_categories`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |

Table columns: ["Name"].

## Items (`items`)

Group: Inventory. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/items`. Table pointer: `review/legacy/app/Livewire/ItemTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| item_category_id | Item Category | select | True |
| unit | Unit | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Item Category", "Unit", "Available Quantity"].

## Item Stocks (`item-stocks`)

Group: Inventory. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/item_stocks`. Table pointer: `review/legacy/app/Livewire/ItemStockTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| item_category_id | Item Category | select | True |
| item_id | Item | select | True |
| supplier_name | Supplier Name | text | False |
| store_name | Store Name | text | False |
| quantity | Quantity | text | True |
| purchase_price | Purchase Price | text | True |
| description | Description | textarea | False |
| image | Attachment | file | False |

Table columns: ["Item", "Item Category", "Quantity", "Purchase Price"].

## Issued Items (`issued-items`)

Group: Inventory. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/issued_items`. Table pointer: `review/legacy/app/Livewire/IssuedItemTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| department_id | Department ID | select | True |
| user_id | User ID | select | True |
| issued_by | Issued By | text | True |
| issued_date | Issued Date | date | True |
| return_date | Return Date | date | True |
| item_category_id | Item Category | select | True |
| item_id | Item | select | True |
| quantity | Quantity | number | True |
| description | Description | textarea | False |

Table columns: ["Item", "Item Category", "Issued Date", "Return Date", "Quantity", "Status"].

## Live Consultations (`live-consultations`)

Group: Live Consultations. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/live_consultations`. Table pointer: `review/legacy/app/Livewire/LiveConsultationTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| consultation_title | Consultation Title | text | True |
| consultation_date | Consultation Date | date | True |
| platform_type | Platform Type | select | True |
| consultation_duration_minutes | Consultation Duration Minutes | number | True |
| patient_id | Patient Name | select | True |
| doctor_id | Doctor Name | select | True |
| type | Type | select | True |
| type_number | Type Number | select | True |
| host_video | Host Video | select | True |
| participant_video | Client Video | select | True |
| description | Description | textarea | False |

Table columns: ["Consultation Title", "Date", "Created By", "Created For", "Patient", "Status", "Password"].

## Live Meetings (`live-consultations-live-meetings`)

Group: Live Consultations. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/live_consultations/live_meetings`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## Google Meet Consultations (`goole-meet-consultation`)

Group: Live Consultations. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/goole_meet_consultation`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## Medicine Categories (`categories`)

Group: Medicine. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/categories`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Category | text | False |
| active | Status | text | False |

Table columns: ["Category", "Status"].

## Medicine Brands (`brands`)

Group: Medicine. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/brands`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Brand | text | True |
| phone | Phone | tel | False |
| email | Email | email | False |

Table columns: ["Brand", "Phone", "Email"].

## Medicines (`medicines`)

Group: Medicine. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/medicines`. Table pointer: `review/legacy/app/Livewire/MedicineTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Medicine | text | True |
| category_id | Category | select | True |
| brand_id | Brand | select | True |
| salt_composition | Salt Composition | text | True |
| buying_price | Buying Price | text | True |
| selling_price | Selling Price | text | True |
| side_effects | Side Effects | textarea | False |
| description | Description | textarea | False |

Table columns: ["Medicine", "Brand", "Available Quantity", "Selling Price", "Buying Price"].

## Medicine Purchases (`purchase-medicines`)

Group: Medicine. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/purchase-medicines`. Table pointer: `review/legacy/app/Livewire/PurchaseMedicineTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| note | Note | textarea | False |
| total | Total | number | True |
| discount | Discount | number | False |
| tax | Tax | number | False |
| net_amount | Net Amount | number | True |
| payment_type | Payment Type | select | True |
| payment_note | Payment Note | textarea | False |

Table columns: ["Purchase Number", "Total", "Tax", "Payment Status", "Net Amount", "Payment Mode"].

## Used Medicines (`used-medicine`)

Group: Medicine. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/used-medicine`. Table pointer: `review/legacy/app/Livewire/UsedMedicineTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Medicines", "Used Quantity", "Used At", "Date"].

## Medicine Bills (`medicine-bills`)

Group: Medicine. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/medicine-bills`. Table pointer: `review/legacy/app/Livewire/MedicineBillTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| bill_date | Bill Date | date | True |
| note | Note | textarea | False |
| total | Total | text | False |
| discount | Discount | text | False |
| tax | Tax | number | False |
| net_amount | Net Amount | text | False |
| payment_type | Payment Type | select | True |
| payment_note | Payment Note | textarea | False |

Table columns: ["Bill Number", "Date", "Patient", "Doctor", "Payment Mode", "Net Amount", "Payment Status"].

## Patients (`patients`)

Group: Patients. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/patients`. Table pointer: `review/legacy/app/Livewire/PatientTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| first_name | First Name | text | True |
| last_name | Last Name | text | True |
| email | Email | email | True |
| dob | Date of Birth | date | False |
| phone | Phone | tel | True |
| gender | Gender | select | True |
| status | Status | select | False |
| blood_group | Blood Group | select | False |
| password | Password | password | True |
| password_confirmation | Password Confirmation | password | True |
| image | Profile | file | False |
| address1 | Address 1 | text | False |
| address2 | Address 2 | text | False |
| city | City | text | False |
| zip | Zip | text | False |
| facebook_url | Facebook URL | text | False |
| twitter_url | Twitter URL | text | False |
| instagram_url | Instagram URL | text | False |
| linkedIn_url | Linked In URL | text | False |

Table columns: ["Patients", "Phone", "Blood Group", "Status"].

## Patient Cases (`patient-cases`)

Group: Patients. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/patient_cases`. Table pointer: `review/legacy/app/Livewire/PatientCaseTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| doctor_id | Doctor | select | True |
| date | Case Date | date | True |
| phone | Phone | tel | False |
| status | Status | select | False |
| fee | Fee | text | True |
| description | Description | textarea | False |

Table columns: ["Case ID", "Doctor", "Case Date", "Fee", "Status"].

## Case Handlers (`case-handlers`)

Group: Patients. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/case_handlers`. Table pointer: `review/legacy/app/Livewire/CaseHandlerTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| first_name | First Name | text | True |
| last_name | Last Name | text | True |
| email | Email | email | True |
| designation | Designation | text | True |
| phone | Phone | tel | False |
| gender | Gender | select | True |
| status | Status | select | False |
| qualification | Qualification | text | True |
| dob | Date of Birth | date | False |
| blood_group | Blood Group | select | False |
| password | Password | password | True |
| password_confirmation | Password Confirmation | password | True |
| image | Profile | file | False |
| address1 | Address 1 | text | False |
| address2 | Address 2 | text | False |
| city | City | text | False |
| zip | Zip | text | False |

Table columns: ["Users", "Phone", "Qualification", "Birth Date", "Status"].

## Patient Admissions (`patient-admissions`)

Group: Patients. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/patient_admissions`. Table pointer: `review/legacy/app/Livewire/PatientAdmissionTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| doctor_id | Doctor | select | True |
| admission_date | Admission Date | date | True |
| discharge_date | Discharge Date | date | False |
| package_id | Package | select | False |
| insurance_id | Insurance | select | False |
| bed_id | Bed | select | False |
| policy_no | Policy No | text | False |
| agent_name | Agent Name | text | False |
| guardian_name | Guardian Name | text | False |
| guardian_relation | Guardian Relation | text | False |
| guardian_contact | Guardian Contact | text | False |
| guardian_address | Guardian Address | text | False |
| status | Status | select | False |

Table columns: ["Package", "Insurance", "Policy No", "Status"].

## Pathology Categories (`pathology-categories`)

Group: Pathology. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/pathology_categories`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |

Table columns: ["Name"].

## Pathology Units (`pathology-units`)

Group: Pathology. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/pathology_units`. Table pointer: `review/legacy/app/Livewire/PathologyUnitTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |

Table columns: ["Name"].

## Pathology Parameters (`pathology-parameter`)

Group: Pathology. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/pathology_parameter`. Table pointer: `review/legacy/app/Livewire/PathologyParameterTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| parameter_name | Name | text | True |
| reference_range | Reference Range | text | True |
| unit_id | Unit | select | True |
| description | Description | textarea | False |

Table columns: ["Name", "Reference Range", "Unit"].

## Pathology Tests (`pathology-tests`)

Group: Pathology. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/pathology_tests`. Table pointer: `review/legacy/app/Livewire/PathologyTestsTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| test_name | Test Name | text | True |
| short_name | Short Name | text | True |
| test_type | Test Type | text | True |
| category_id | Category Name | select | True |
| unit | Unit | number | False |
| subcategory | Subcategory | text | False |
| method | Method | text | False |
| report_days | Report Days | number | False |
| charge_category_id | Charge Category | select | True |
| standard_charge | Standard Charge | text | True |

Table columns: ["Test Name", "Patient", "Short Name", "Test Type", "Category Name", "Charge Category"].

## Birth Reports (`birth-reports`)

Group: Reports. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/birth_reports`. Table pointer: `review/legacy/app/Livewire/BirthReportTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| case_id | Case | select | True |
| doctor_id | Doctor | select | True |
| date | Date | date | True |
| description | Description | textarea | False |

Table columns: ["Case ID", "Patient", "Doctor", "Date"].

## Death Reports (`death-reports`)

Group: Reports. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/death_reports`. Table pointer: `review/legacy/app/Livewire/DeathReportTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| case_id | Case | select | True |
| doctor_id | Doctor | select | True |
| date | Date | date | True |
| description | Description | textarea | False |

Table columns: ["Case ID", "Patient", "Doctor", "Date"].

## Investigation Reports (`investigation-reports`)

Group: Reports. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/investigation_reports`. Table pointer: `review/legacy/app/Livewire/InvestigationReportTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| title | Title | text | True |
| patient_id | Patient | select | True |
| doctor_id | Doctor | select | True |
| date | Date | date | True |
| image | Attachment | file | False |
| status | Status | select | True |
| description | Description | textarea | False |

Table columns: ["User Details", "Users", "Date", "Title", "Status", "Attachment"].

## Operation Reports (`operation-reports`)

Group: Reports. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/operation_reports`. Table pointer: `review/legacy/app/Livewire/OperationReportTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| case_id | Case | select | True |
| doctor_id | Doctor | select | True |
| date | Date | date | True |
| description | Description | textarea | False |

Table columns: ["Case ID", "Patient", "Doctor", "Date"].

## Radiology Categories (`radiology-categories`)

Group: Radiology. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/radiology_categories`. Table pointer: `review/legacy/app/Livewire/RadiologyCategoriesTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |

Table columns: ["Name"].

## Radiology Tests (`radiology-tests`)

Group: Radiology. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/radiology_tests`. Table pointer: `review/legacy/app/Livewire/RadiologyTestTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| test_name | Test Name | text | True |
| short_name | Short Name | text | True |
| test_type | Test Type | text | True |
| category_id | Category Name | select | True |
| subcategory | Subcategory | text | False |
| report_days | Report Days | number | False |
| charge_category_id | Charge Category | select | True |
| charge_id | Charge | select | True |
| standard_charge | Standard Charge | text | True |

Table columns: ["Test Name", "Short Name", "Test Type", "Category Name", "Charge Category"].

## Insurances (`insurances`)

Group: Services. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/insurances`. Table pointer: `review/legacy/app/Livewire/InsuranceTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Insurance | text | True |
| service_tax | Service Tax | text | True |
| discount | Discount | number | False |
| insurance_no | Insurance No | text | True |
| insurance_code | Insurance Code | text | True |
| hospital_rate | Hospital Rate | text | True |
| remark | Remark | textarea | False |
| disease_name[] | Status | text | True |

Table columns: ["Insurance", "Service Tax", "Insurance No", "Insurance Code", "Hospital Rate", "Total", "Status"].

## Packages (`packages`)

Group: Services. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/packages`. Table pointer: `review/legacy/app/Livewire/PackageTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Package | text | True |
| discount | Discount | number | True |
| description | Description | textarea | True |

Table columns: ["Package", "Discount", "Total Amount"].

## Ambulances (`ambulances`)

Group: Services. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/ambulances`. Table pointer: `review/legacy/app/Livewire/AmbulanceTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| vehicle_number | Vehicle Number | text | True |
| vehicle_model | Vehicle Model | text | True |
| year_made | Year Made | text | True |
| driver_name | Driver Name | text | True |
| driver_contact | Driver Contact | tel | True |
| driver_license | Driver License | text | True |
| note | Note | textarea | False |
| vehicle_type | Vehicle Type | select | True |
| is_available | Is Available | text | False |

Table columns: ["Vehicle Number", "Vehicle Model", "Year Made", "Driver Name", "Driver License", "Driver Contact", "Vehicle Type", "Is Available"].

## Ambulance Calls (`ambulance-calls`)

Group: Services. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/ambulance_calls`. Table pointer: `review/legacy/app/Livewire/AmbulanceCallTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| ambulance_id | Vehicle Model | select | True |
| date | Date | date | True |
| driver_name | Driver Name | text | True |
| amount | Amount | text | True |

Table columns: ["Patient", "Vehicle Model", "Driver Name", "Date", "Amount"].

## SMS (`sms`)

Group: SMS / Mail. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/sms`. Table pointer: `review/legacy/app/Livewire/SmsTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| phone | Phone Number | tel | True |
| role | Role | select | True |
| number | Send SMS By Number Directly | text | False |
| send_to[] | Send To | select | True |
| message | Message | textarea | True |

Table columns: ["Send To", "Phone", "Region Code", "Send By"].

## Emails (`emails`)

Group: SMS / Mail. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/emails`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## Email Template (`email-template`)

Group: SMS / Mail. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/email-template`. Table pointer: `review/legacy/app/Livewire/EmailTemplateTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Email Subject", "Actions"].

## General Settings (`settings`)

Group: Settings. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/settings`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| app_name | App Name | text | True |
| company_name | Company Name | text | True |
| hospital_email | Hospital Email | email | True |
| hospital_phone | Hospital Phone | tel | True |
| hospital_from_day | Hospital From Day | text | True |
| hospital_from_time | Hospital From Time | text | True |
| hospital_address | Address | text | True |
| current_currency | Currency | text | True |
| app_logo | App Logo | file | True |
| country_phone | Country Code | text | False |
| default_lang | Default Language | text | False |
| about_us | About Us | textarea | True |
| favicon | Favicon | text | False |
| facebook_url | Facebook URL | text | False |
| twitter_url | Twitter URL | text | False |
| instagram_url | Instagram URL | text | False |
| linkedIn_url | Linked In URL | text | False |
| open_ai_key | Open Ai Key | text | False |
| model_name | Model Name | text | False |
| custom_serial_prefix | Custom Serial Prefix | text | False |

Table columns: ["App Name", "Company Name", "Hospital Email", "Hospital Phone", "Hospital From Day"].

## Hospital Schedule (`hospital-schedule`)

Group: Settings. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/hospital_schedule`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## Currencies (`currency-settings`)

Group: Settings. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/currency_settings`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| currency_name | Currency Name | text | False |
| currency_icon | Currency Icon | text | False |
| currency_code | Currency Code | text | False |

Table columns: ["Currency Name", "Currency Icon", "Currency Code"].

## Operation Categories (`operation-categories`)

Group: Settings. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/operation_categories`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |

Table columns: ["Name"].

## Operations (`operations`)

Group: Settings. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/operations`. Table pointer: `review/legacy/app/Livewire/OperationTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| operation_category_id | Operation Category | select | True |
| name | Name | text | True |

Table columns: ["Name", "Operation Category"].

## Payment Gateway (`payment-gateway`)

Group: Settings. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/payment_gateway`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| stripe_key | Stripe Key | text | False |
| stripe_secret | Stripe Secret | text | False |
| paypal_client_id | Paypal Client ID | text | False |
| paypal_secret | Paypal Secret | text | False |
| paypal_mode | Paypal Mode | text | False |
| razorpay_key | Razorpay Key | text | False |
| razorpay_secret | Razorpay Secret | text | False |
| flutterwave_public_key | Flutterwave Public Key | text | False |
| flutterwave_secret_key | Flutterwave Secret Key | text | False |
| phonepe_merchant_id | Phonepe Merchant ID | text | False |
| phonepe_merchant_user_id | Phonepe Merchant User ID | text | False |
| phonepe_env | Phonepe Env | text | False |
| phonepe_salt_key | Phonepe Salt Key | text | False |
| phonepe_salt_index | Phonepe Salt Index | text | False |
| phonepe_merchant_transaction_id | Phonepe Merchant Transaction ID | text | False |
| paystack_public_key | Paystack Public Key | text | False |
| paystack_secret_key | Paystack Secret Key | text | False |

Table columns: ["Stripe Key", "Stripe Secret", "Paypal Client ID", "Paypal Secret", "Paypal Mode"].

## Custom Fields (`add-custom-fields`)

Group: Settings. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/add_custom_fields`. Table pointer: `review/legacy/app/Livewire/AddCustomFieldTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Module Name", "Field Type", "Field Name", "Value"].

## Vaccinations (`vaccinations`)

Group: Vaccinations. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/vaccinations`. Table pointer: `review/legacy/app/Livewire/VaccinationTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| manufactured_by | Manufactured By | text | True |
| brand | Brand | text | True |

Table columns: ["Name", "Manufactured By", "Brand"].

## Vaccinated Patients (`vaccinated-patients`)

Group: Vaccinations. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/vaccinated_patients`. Table pointer: `review/legacy/app/Livewire/VaccinatedPatientsTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| vaccination_id | Vaccine | select | True |
| vaccination_serial_number | Serial No | text | False |
| dose_number | Does No | number | True |
| dose_given_date | Dose Given Date | date | True |
| description | Notes | textarea | False |

Table columns: ["Patient", "Vaccination", "Serial No", "Does No", "Dose Given Date"].

## Odontogram (`odontogram`)

Group: Odontogram. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/odontogram`. Table pointer: `review/legacy/app/Livewire/OdontogramTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| patient_id | Patient | select | True |
| doctor_id | Doctor | select | True |
| description | Description | textarea | True |

Table columns: ["Patient", "Doctor"].

## Add On (`add-on`)

Group: Add-ons. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/add-on`. Table pointer: `review/legacy/app/Livewire/AddOnTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## IPD Diagnoses (`ipd-diagnoses`)

Group: Clinical Records. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/ipd_diagnoses`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| report_type | Report Type | text | True |
| report_date | Report Date | date | True |
| description | Description | textarea | False |
| document | Document | file | False |

Table columns: ["Report Type", "Report Date"].

## IPD Consultant Registers (`ipd-consultant-registers`)

Group: Clinical Records. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/ipd_consultant_registers`. Table pointer: `review/legacy/app/Livewire/IpdConsultantRegisterTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Doctor ID", "Applied Date", "Instruction Date"].

## IPD Operation (`ipd-operation`)

Group: Clinical Records. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/ipd_operation`. Table pointer: `review/legacy/app/Livewire/IpdOperationTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| operation_date | Operation Date | date | True |
| operation_category_id | Operation Category | select | True |
| operation_id | Operation Name | select | True |
| doctor_id | IPD Consultant Doctor | select | True |
| assistant_consultant_1 | Assistant Consultant 1 | text | False |
| assistant_consultant_2 | Assistant Consultant 2 | text | False |
| anesthetist | Anesthetist | text | False |
| anesthesia_type | Anesthesia Type | text | False |
| ot_technician | Ot Technician | text | False |
| ot_assistant | Ot Assistant | text | False |
| remark | Remarks | textarea | False |
| result | Result | textarea | False |

Table columns: ["Reference ID", "Operation Date", "Operation Name", "Operation Category Name", "Ot Technician"].

## IPD Charges (`ipd-charges`)

Group: Clinical Records. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/ipd_charges`. Table pointer: `review/legacy/app/Livewire/IpdChargeTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| date | Date | date | True |
| charge_type_id | Charge Type ID | select | True |
| charge_category_id | Charge Category ID | select | True |
| charge_id | Charge ID | select | True |
| standard_charge | Standard Charge | text | False |
| applied_charge | Applied Charge | text | True |

Table columns: ["Date", "Charge Type ID", "Charge Category ID", "Charge ID", "Standard Charge", "Applied Charge"].

## IPD Prescriptions (`ipd-prescriptions`)

Group: Clinical Records. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/ipd_prescriptions`. Table pointer: `review/legacy/app/Livewire/IpdPrescriptionTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| header_note | Header Note | textarea | True |
| footer_note | Footer Note | textarea | False |

Table columns: ["IPD No", "Patients", "Doctors", "Created On"].

## IPD Timelines (`ipd-timelines`)

Group: Clinical Records. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/ipd_timelines`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| title | Title | text | True |
| date | Date | date | True |
| description | Description | textarea | False |
| visible_to_person | Visible To Person | text | False |
| document | Document | file | False |

Table columns: ["Title", "Date", "Visible To Person"].

## IPD Payments (`ipd-payments`)

Group: Clinical Records. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/ipd_payments`. Table pointer: `review/legacy/app/Livewire/IpdPaymentTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| amount | Amount | text | True |
| date | Date | date | True |
| payment_mode | Payment Mode | select | True |
| document | Document | file | False |
| notes | Notes | textarea | False |

Table columns: ["Date", "Amount", "Payment Mode", "Document", "Note"].

## IPD Bills (`ipd-bills`)

Group: Clinical Records. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/ipd_bills`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["Name", "Description"].

## OPD Diagnoses (`opd-diagnoses`)

Group: Clinical Records. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/opd_diagnoses`. Table pointer: `review/legacy/app/Livewire/OpdDiagnosesTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| report_type | Report Type | text | True |
| report_date | Report Date | date | True |
| description | Description | textarea | False |
| document | Document | file | False |

Table columns: ["Report Type", "Report Date", "Document", "Description"].

## OPD Prescriptions (`opd-prescriptions`)

Group: Clinical Records. Extraction flag: `False`.

Original source pointer: `review/legacy/resources/views/opd_prescriptions`. Table pointer: `review/legacy/app/Livewire/OpdPrescriptionTable.php`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| name | Name | text | True |
| description | Description | textarea | False |

Table columns: ["OPD Number", "Patients", "Doctors", "Created On"].

## OPD Timelines (`opd-timelines`)

Group: Clinical Records. Extraction flag: `True`.

Original source pointer: `review/legacy/resources/views/opd_timelines`. Table pointer: `None`. These provenance paths need not exist in the new repository.

| Field key | Label | UI type | Required flag |
| --- | --- | --- | --- |
| title | Title | text | True |
| date | Date | date | True |
| description | Description | textarea | False |
| visible_to_person | Visible To Person | text | False |
| document | Document | file | False |

Table columns: ["Title", "Date", "Visible To Person"].
