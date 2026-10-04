# Frontend source coverage

This inventory records extracted presentation fields, not completed behavioral parity. Generic fallback forms and special workflows still need individual verification.

| Screen | Form fields | Table source |
| --- | --- | --- |
| Patient ID Card / Patient ID Card Templates | 2 (extracted) | review/legacy/app/Livewire/PatientIdCardTemplateTable.php |
| Patient ID Card / Generate Patient ID Card | 2 (extracted) | review/legacy/app/Livewire/GeneratePatientIdCardTable.php |
| Users / Users | 16 (extracted) | review/legacy/app/Livewire/UserTable.php |
| Users / Admins | 10 (extracted) | review/legacy/app/Livewire/AdminTable.php |
| Users / Accountants | 17 (extracted) | review/legacy/app/Livewire/AccountantTable.php |
| Users / Nurses | 17 (extracted) | review/legacy/app/Livewire/NurseTable.php |
| Users / Lab Technicians | 17 (extracted) | review/legacy/app/Livewire/LabTechnicianTable.php |
| Users / Receptionists | 17 (extracted) | review/legacy/app/Livewire/ReceptionistTable.php |
| Users / Pharmacists | 17 (extracted) | review/legacy/app/Livewire/PharmacistTable.php |
| Appointments / Appointments | 9 (extracted) | review/legacy/app/Livewire/AppointmentTable.php |
| Appointments / Appointment Calendars | 2 (fallback) | needs verification |
| Appointments / Appointment Transaction | 2 (fallback) | review/legacy/app/Livewire/AppointmentTransactionTable.php |
| Appointments / Patient Queues | 2 (fallback) | review/legacy/app/Livewire/PatientQueueTable.php |
| IPD / OPD / IPD Patients | 13 (extracted) | review/legacy/app/Livewire/IpdPatientDepartmentTable.php |
| IPD / OPD / OPD Patients | 13 (extracted) | review/legacy/app/Livewire/OpdPatientDepartmentTable.php |
| Billing / Accounts | 4 (extracted) | review/legacy/app/Livewire/AccountTable.php |
| Billing / Employee Payrolls | 11 (extracted) | review/legacy/app/Livewire/EmployeePayrollTable.php |
| Billing / Invoices | 4 (extracted) | review/legacy/app/Livewire/InvoiceTable.php |
| Billing / Payments | 5 (extracted) | review/legacy/app/Livewire/PaymentTable.php |
| Billing / Payment Reports | 2 (fallback) | needs verification |
| Billing / Advanced Payments | 4 (extracted) | review/legacy/app/Livewire/AdvancedPaymentTable.php |
| Billing / Bills | 14 (extracted) | review/legacy/app/Livewire/BillTable.php |
| Billing / Manual Payment Approvals | 2 (fallback) | needs verification |
| Bed Management / Bed Status | 2 (fallback) | needs verification |
| Bed Management / Bed Assigns | 7 (extracted) | review/legacy/app/Livewire/BedAssignTable.php |
| Bed Management / Beds | 4 (extracted) | review/legacy/app/Livewire/BedTable.php |
| Bed Management / Bed Types | 2 (extracted) | review/legacy/app/Livewire/BedTypeTable.php |
| Blood Bank / Blood Banks | 2 (extracted) | review/legacy/app/Livewire/BloodBankTable.php |
| Blood Bank / Blood Donors | 5 (extracted) | review/legacy/app/Livewire/BloodDonorTable.php |
| Blood Bank / Blood Donations | 2 (extracted) | review/legacy/app/Livewire/BloodDonationTable.php |
| Blood Bank / Blood Issues | 7 (extracted) | review/legacy/app/Livewire/BloodIssueTable.php |
| Documents / Documents | 5 (extracted) | review/legacy/app/Livewire/DocumentTable.php |
| Documents / Document Types | 1 (extracted) | review/legacy/app/Livewire/DocumentTypeTable.php |
| Doctors / Doctors | 25 (extracted) | review/legacy/app/Livewire/DoctorTable.php |
| Doctors / Doctor Departments | 2 (extracted) | review/legacy/app/Livewire/DoctorDepartmentTable.php |
| Doctors / Schedules | 2 (extracted) | review/legacy/app/Livewire/ScheduleTable.php |
| Doctors / Doctor Holidays | 3 (extracted) | review/legacy/app/Livewire/DoctorHolidayTable.php |
| Doctors / Lunch Breaks | 4 (extracted) | review/legacy/app/Livewire/LunchBreakTable.php |
| Prescriptions / Prescriptions | 6 (extracted) | review/legacy/app/Livewire/PrescriptionTable.php |
| Diagnosis / Diagnosis Tests | 13 (extracted) | review/legacy/app/Livewire/PatientDiagnosisTestTable.php |
| Diagnosis / Diagnosis Categories | 2 (extracted) | needs verification |
| Enquiries / Enquiries | 2 (fallback) | needs verification |
| Finance / Incomes | 7 (extracted) | review/legacy/app/Livewire/IncomeTable.php |
| Finance / Expenses | 7 (extracted) | review/legacy/app/Livewire/ExpenseTable.php |
| Front Office / Call Logs | 6 (extracted) | review/legacy/app/Livewire/CallLogTable.php |
| Front Office / Visitors | 10 (extracted) | review/legacy/app/Livewire/VisitorTable.php |
| Front Office / Postal Receive / Dispatch | 6 (extracted) | needs verification |
| Front Office / Complaints | 2 (extracted) | review/legacy/app/Livewire/ComplaintTable.php |
| Front CMS / Front Settings | 3 (extracted) | needs verification |
| Front CMS / Services | 5 (extracted) | review/legacy/app/Livewire/ServiceTable.php |
| Front CMS / Notice Boards | 2 (extracted) | review/legacy/app/Livewire/NoticeBoardTable.php |
| Front CMS / Testimonials | 3 (extracted) | review/legacy/app/Livewire/TestimonialTable.php |
| Hospital Charges / Charge Categories | 3 (extracted) | needs verification |
| Hospital Charges / Charges | 5 (extracted) | review/legacy/app/Livewire/ChargeTable.php |
| Hospital Charges / Doctor OPD Charges | 2 (extracted) | review/legacy/app/Livewire/DoctorOPDChargeTable.php |
| Inventory / Item Categories | 1 (extracted) | needs verification |
| Inventory / Items | 4 (extracted) | review/legacy/app/Livewire/ItemTable.php |
| Inventory / Item Stocks | 8 (extracted) | review/legacy/app/Livewire/ItemStockTable.php |
| Inventory / Issued Items | 9 (extracted) | review/legacy/app/Livewire/IssuedItemTable.php |
| Live Consultations / Live Consultations | 11 (extracted) | review/legacy/app/Livewire/LiveConsultationTable.php |
| Live Consultations / Live Meetings | 2 (fallback) | needs verification |
| Live Consultations / Google Meet Consultations | 2 (fallback) | needs verification |
| Medicine / Medicine Categories | 2 (extracted) | needs verification |
| Medicine / Medicine Brands | 3 (extracted) | needs verification |
| Medicine / Medicines | 8 (extracted) | review/legacy/app/Livewire/MedicineTable.php |
| Medicine / Medicine Purchases | 7 (extracted) | review/legacy/app/Livewire/PurchaseMedicineTable.php |
| Medicine / Used Medicines | 2 (fallback) | review/legacy/app/Livewire/UsedMedicineTable.php |
| Medicine / Medicine Bills | 9 (extracted) | review/legacy/app/Livewire/MedicineBillTable.php |
| Patients / Patients | 19 (extracted) | review/legacy/app/Livewire/PatientTable.php |
| Patients / Patient Cases | 7 (extracted) | review/legacy/app/Livewire/PatientCaseTable.php |
| Patients / Case Handlers | 17 (extracted) | review/legacy/app/Livewire/CaseHandlerTable.php |
| Patients / Patient Admissions | 14 (extracted) | review/legacy/app/Livewire/PatientAdmissionTable.php |
| Pathology / Pathology Categories | 1 (extracted) | needs verification |
| Pathology / Pathology Units | 1 (extracted) | review/legacy/app/Livewire/PathologyUnitTable.php |
| Pathology / Pathology Parameters | 4 (extracted) | review/legacy/app/Livewire/PathologyParameterTable.php |
| Pathology / Pathology Tests | 11 (extracted) | review/legacy/app/Livewire/PathologyTestsTable.php |
| Reports / Birth Reports | 4 (extracted) | review/legacy/app/Livewire/BirthReportTable.php |
| Reports / Death Reports | 4 (extracted) | review/legacy/app/Livewire/DeathReportTable.php |
| Reports / Investigation Reports | 7 (extracted) | review/legacy/app/Livewire/InvestigationReportTable.php |
| Reports / Operation Reports | 4 (extracted) | review/legacy/app/Livewire/OperationReportTable.php |
| Radiology / Radiology Categories | 1 (extracted) | review/legacy/app/Livewire/RadiologyCategoriesTable.php |
| Radiology / Radiology Tests | 9 (extracted) | review/legacy/app/Livewire/RadiologyTestTable.php |
| Services / Insurances | 8 (extracted) | review/legacy/app/Livewire/InsuranceTable.php |
| Services / Packages | 3 (extracted) | review/legacy/app/Livewire/PackageTable.php |
| Services / Ambulances | 9 (extracted) | review/legacy/app/Livewire/AmbulanceTable.php |
| Services / Ambulance Calls | 5 (extracted) | review/legacy/app/Livewire/AmbulanceCallTable.php |
| SMS / Mail / SMS | 5 (extracted) | review/legacy/app/Livewire/SmsTable.php |
| SMS / Mail / Emails | 2 (fallback) | needs verification |
| SMS / Mail / Email Template | 2 (fallback) | review/legacy/app/Livewire/EmailTemplateTable.php |
| Settings / General Settings | 20 (extracted) | needs verification |
| Settings / Hospital Schedule | 2 (fallback) | needs verification |
| Settings / Currencies | 3 (extracted) | needs verification |
| Settings / Operation Categories | 1 (extracted) | needs verification |
| Settings / Operations | 2 (extracted) | review/legacy/app/Livewire/OperationTable.php |
| Settings / Payment Gateway | 17 (extracted) | needs verification |
| Settings / Custom Fields | 2 (fallback) | review/legacy/app/Livewire/AddCustomFieldTable.php |
| Vaccinations / Vaccinations | 3 (extracted) | review/legacy/app/Livewire/VaccinationTable.php |
| Vaccinations / Vaccinated Patients | 6 (extracted) | review/legacy/app/Livewire/VaccinatedPatientsTable.php |
| Odontogram / Odontogram | 3 (extracted) | review/legacy/app/Livewire/OdontogramTable.php |
| Add-ons / Add On | 2 (fallback) | review/legacy/app/Livewire/AddOnTable.php |
| Clinical Records / IPD Diagnoses | 4 (extracted) | needs verification |
| Clinical Records / IPD Consultant Registers | 2 (fallback) | review/legacy/app/Livewire/IpdConsultantRegisterTable.php |
| Clinical Records / IPD Operation | 12 (extracted) | review/legacy/app/Livewire/IpdOperationTable.php |
| Clinical Records / IPD Charges | 6 (extracted) | review/legacy/app/Livewire/IpdChargeTable.php |
| Clinical Records / IPD Prescriptions | 2 (extracted) | review/legacy/app/Livewire/IpdPrescriptionTable.php |
| Clinical Records / IPD Timelines | 5 (extracted) | needs verification |
| Clinical Records / IPD Payments | 5 (extracted) | review/legacy/app/Livewire/IpdPaymentTable.php |
| Clinical Records / IPD Bills | 2 (fallback) | needs verification |
| Clinical Records / OPD Diagnoses | 4 (extracted) | review/legacy/app/Livewire/OpdDiagnosesTable.php |
| Clinical Records / OPD Prescriptions | 2 (fallback) | review/legacy/app/Livewire/OpdPrescriptionTable.php |
| Clinical Records / OPD Timelines | 5 (extracted) | needs verification |
