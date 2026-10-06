# Original model validation and state constants

Exact static model declarations. FormRequest overrides, repository checks and database constraints can add or change these rules. These expressions were not executed.

## Account

Source: `hms/app/Models/Account.php`; table: `accounts`

```php
'name' => 'required|string|unique:accounts,name',
        'type' => 'required|integer',
        'status' => 'nullable|integer',
        'description' => 'string|nullable',
```

```php
const INACTIVE = 0;
```

```php
const ACTIVE = 1;
```

```php
const ACTIVE_ALL = 2;
```

```php
const STATUS_ARR = [
        self::ACTIVE_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

```php
const DEBIT = 1;
```

```php
const CREDIT = 2;
```

```php
const TYPE_ALL = 0;
```

```php
const TYPE_ARR = [
        self::TYPE_ALL => 'All',
        self::DEBIT => 'Debit',
        self::CREDIT => 'Credit',
    ];
```

```php
const ACCOUNT_TYPES = [
        self::TYPE_ALL => 'All',
        self::CREDIT => 'Credit',
        self::DEBIT => 'Debit',
    ];
```

## Accountant

Source: `hms/app/Models/Accountant.php`; table: `accountants`

```php
'first_name' => 'required|string',
        'last_name' => 'required|string',
        'email' => 'required|email:filter|unique:users,email',
        'password' => 'required|same:password_confirmation|min:6',
        'designation' => 'required|string',
        'qualification' => 'required|string',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## AddCustomFields

Source: `hms/app/Models/AddCustomFields.php`; table: `add_custom_fields`

```php
'module_name' => 'required',
        'field_type' => 'required',
        'field_name' => 'required|string',
        'grid' => 'required|numeric|min:6|max:12',
```

```php
const MODULE_TYPE_ARR = [
        self::Appointment=> 'Appointment',
        self::IpdPatient => 'IPD Patient',
        self::OpdPatient => 'OPD Patient',
        self::Patient => 'Patient',
    ];
```

```php
const FIELD_TYPE_ARR = [
        0 => 'text',
        1 => 'textarea',
        2 => 'toggle',
        3 => 'number',
        4 => 'select',
        5 => 'multiSelect',
        6 => 'date',
        7 => 'date & Time',
    ];
```

## AddOn

Source: `hms/app/Models/AddOn.php`; table: `add_ons`

```php
// No static rules array extracted from this model.
```

## Address

Source: `hms/app/Models/Address.php`; table: `addresses`

```php
// No static rules array extracted from this model.
```

## AdvancedPayment

Source: `hms/app/Models/AdvancedPayment.php`; table: `advanced_payments`

```php
'patient_id' => 'required',
        'receipt_no' => 'required|string',
        'amount' => 'required',
        'date' => 'required|date',
```

## Ambulance

Source: `hms/app/Models/Ambulance.php`; table: `ambulances`

```php
'vehicle_number' => 'required|unique:ambulances,vehicle_number',
        'vehicle_model' => 'required',
        'driver_contact' => 'nullable',
        'year_made' => 'required|size:4',
        'driver_name' => 'required|string',
        'driver_license' => 'required|string',
```

```php
const STATUS_ALL = 2;
```

```php
const TRUE = 1;
```

```php
const FALSE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::TRUE => 'Available',
        self::FALSE => 'Not Available',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Available',
        2 => 'Not Available',
    ];
```

## AmbulanceCall

Source: `hms/app/Models/AmbulanceCall.php`; table: `ambulance_calls`

```php
'patient_id' => 'required|integer|min:1',
        'date' => 'required|string',
        'amount' => 'required',
```

## Appointment

Source: `hms/app/Models/Appointment.php`; table: `appointments`

```php
'patient_id' => 'required',
        'doctor_id' => 'required',
        'department_id' => 'required',
        'opd_date' => 'required',
        'problem' => 'string|nullable',
```

```php
const STATUS_ARR = [
        '2' => 'All',
        '0' => 'Pending',
        '1' => 'Completed',
        '3' => 'Cancelled',
        '4' => 'In Queue',
        '5' => 'Check In',
    ];
```

```php
const STATUS_PENDING = 0;
```

```php
const STATUS_COMPLETED = 1;
```

```php
const STATUS_ALL = 2;
```

```php
const STATUS_CANCELLED = 3;
```

```php
const STATUS_IN_QUEUE = 4;
```

```php
const STATUS_CHECK_IN = 5;
```

```php
const TYPE_STRIPE = 3;
```

```php
const TYPE_RAZORPAY = 4;
```

```php
const TYPE_PAYPAL = 5;
```

```php
const TYPE_CASH = 1;
```

```php
const OTHER = 6;
```

```php
const CHEQUE = 2;
```

```php
const PHONEPE = 7;
```

```php
const FLUTTERWAVE = 8;
```

```php
const PAYSTACK = 9;
```

```php
const PAYMENT_TYPES = [
        self::TYPE_STRIPE => 'Stripe',
        self::TYPE_RAZORPAY => 'RazorPay',
        self::TYPE_PAYPAL => 'Paypal',
        self::TYPE_CASH => 'Cash',
        self::OTHER => 'Other',
        self::CHEQUE => 'Cheque',
    ];
```

## AppointmentTransaction

Source: `hms/app/Models/AppointmentTransaction.php`; table: `appointment_transactions`

```php
// No static rules array extracted from this model.
```

```php
const PAYMENT_MODES = [
        1 => 'Cash',
        2 => 'Cheque',
        3 => 'Stripe',
        4 => 'Razorpay',
        5 => 'PayPal',
        6 => 'Other',
    ];
```

## Bed

Source: `hms/app/Models/Bed.php`; table: `beds`

```php
'bed_type' => 'required',
        'name' => 'required|unique:beds,name',
        'charge' => 'required',
        'description' => 'string|nullable',
```

```php
const NOTAVAILABLE = 0;
```

```php
const AVAILABLE = 1;
```

```php
const AVAILABLE_ALL = 2;
```

```php
const STATUS_ARR = [
        self::AVAILABLE_ALL => 'All',
        self::AVAILABLE => 'Available',
        self::NOTAVAILABLE => 'Not Available',
    ];
```

```php
const FILTER_INCOME_HEAD = [
        0 => 'All',
        1 => 'Available',
        2 => 'Not Available',
    ];
```

## BedAssign

Source: `hms/app/Models/BedAssign.php`; table: `bed_assigns`

```php
'bed_id' => 'required',
        'case_id' => 'nullable',
        'assign_date' => 'required',
        'discharge_date' => 'nullable|after:assign_date',
        'description' => 'nullable|string',
        'ipd_patient_department_id' => 'required',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## BedType

Source: `hms/app/Models/BedType.php`; table: `bed_types`

```php
'title' => 'required|unique:bed_types,title',
        'description' => 'nullable|string',
```

## Bill

Source: `hms/app/Models/Bill.php`; table: `bills`

```php
'patient_id' => 'required|integer|min:1',
        'bill_date' => 'required|string',
```

```php
const PAYEMENT_TYPE_ARR = [
        self::Stripe => 'Stripe',
        self::Manually => 'Manually',
        self::Razorpay => 'Razorpay',
        self::Flutterwave => 'Flutterwave',
        self::PhonePe => 'PhonePe',
        self::Paystack => 'Paystack',
    ];
```

## BillItems

Source: `hms/app/Models/BillItems.php`; table: `bill_items`

```php
'item_name' => 'required|string',
        'qty' => 'required|integer',
        'price' => 'required|regex:/^\d*(\.\d{1,2})?$/',
```

## BirthReport

Source: `hms/app/Models/BirthReport.php`; table: `birth_reports`

```php
'case_id' => 'required|unique:birth_reports,case_id',
        'doctor_id' => 'required',
        'date' => 'required',
        'description' => 'nullable|string',
```

## BloodBank

Source: `hms/app/Models/BloodBank.php`; table: `blood_banks`

```php
'blood_group' => 'required|unique:blood_banks,blood_group',
        'remained_bags' => 'required|numeric',
```

## BloodDonation

Source: `hms/app/Models/BloodDonation.php`; table: `blood_donations`

```php
'blood_donor_id' => 'required',
        'bags' => 'required|numeric|digits_between:1,100',
```

## BloodDonor

Source: `hms/app/Models/BloodDonor.php`; table: `blood_donors`

```php
'name' => 'required',
        'age' => 'required|numeric|digits_between:1,100',
        'blood_group' => 'required',
        'last_donate_date' => 'required',
```

## BloodIssue

Source: `hms/app/Models/BloodIssue.php`; table: `blood_issues`

```php
'issue_date' => 'required',
        'patient_id' => 'required',
        'doctor_id' => 'required',
        'donor_id' => 'required',
        'remarks' => 'string|nullable',
```

## Brand

Source: `hms/app/Models/Brand.php`; table: `brands`

```php
'name' => 'required|unique:brands,name',
        'email' => 'email|unique:brands,email|nullable',
        'phone' => 'nullable|numeric',
```

## CallLog

Source: `hms/app/Models/CallLog.php`; table: `call_logs`

```php
'name' => 'required',
```

```php
const INCOMING = 1;
```

```php
const OUTCOMING = 2;
```

```php
const CALLTYPE_ARR = [
        '0' => 'All',
        '1' => 'Incoming',
        '2' => 'Outgoing',
    ];
```

## CaseHandler

Source: `hms/app/Models/CaseHandler.php`; table: `case_handlers`

```php
'first_name' => 'required|string',
        'last_name' => 'required|string',
        'email' => 'required|email|unique:users,email',
        'password' => 'required|same:password_confirmation|min:6',
        'designation' => 'required|string',
        'qualification' => 'required|string',
        'address1' => 'nullable|string',
        'address2' => 'nullable|string',
        'city' => 'nullable|string',
        'zip' => 'nullable|integer',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

## Category

Source: `hms/app/Models/Category.php`; table: `categories`

```php
'name' => 'required|unique:categories,name',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

## Charge

Source: `hms/app/Models/Charge.php`; table: `charges`

```php
'charge_type' => 'required',
        'charge_category_id' => 'required',
        'code' => 'required|unique:charges,code',
        'standard_charge' => 'required',
```

## ChargeCategory

Source: `hms/app/Models/ChargeCategory.php`; table: `charge_categories`

```php
'name' => 'required|unique:charge_categories,name',
        'description' => 'nullable|string',
        'charge_type' => 'required',
```

```php
const CHARGE_TYPES = [
        1 => 'Investigations',
        2 => 'Operation Theatre',
        3 => 'Others',
        4 => 'Procedures',
        5 => 'Supplier',
    ];
```

```php
const FILTER_CHARGE_TYPES = [
        0 => 'All',
        4 => 'Procedures',
        1 => 'Investigations',
        5 => 'Supplier',
        2 => 'Operation Theatre',
        3 => 'Others',
    ];
```

## Complaint

Source: `hms/app/Models/Complaint.php`; table: `complaints`

```php
'title' => 'required|max:50',
        'description' => 'required|max:500',
```

```php
const STATUS_PENDING = 0;
```

```php
const STATUS_IN_PROGRESS = 1;
```

```php
const STATUS_RESOLVED = 2;
```

```php
const STATUS_REJECT = 3;
```

```php
const STATUS_HOLD = 4;
```

```php
const STATUS_ARR = [
        0 => 'Pending',
        1 => 'In Progress',
        2 => 'Resolved',
        3 => 'Reject',
        4 => 'Hold'
    ];
```

## CurrencySetting

Source: `hms/app/Models/CurrencySetting.php`; table: `currency_settings`

```php
'currency_name' => 'required|unique:currency_settings',
        'currency_icon' => 'required',
        'currency_code' => 'required|min:3|max:3',
```

## DeathReport

Source: `hms/app/Models/DeathReport.php`; table: `death_reports`

```php
'case_id' => 'required|unique:death_reports,case_id',
        'doctor_id' => 'required',
        'date' => 'required',
        'description' => 'nullable|string',
```

## Department

Source: `hms/app/Models/Department.php`; table: `departments`

```php
'name' => 'required|string',
```

```php
const INACTIVE = 0;
```

```php
const ACTIVE = 1;
```

```php
const ACTIVE_ALL = 2;
```

```php
const ACTIVE_ARR = [
        self::ACTIVE_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const ROLE = [
        0 => 'All',
        1 => 'Admin',
        2 => 'Doctor',
        3 => 'Patient',
        4 => 'Nurse',
        5 => 'Receptionist',
        6 => 'Pharmacist',
        7 => 'Accountant',
        8 => 'Case Manager',
        9 => 'Lab Technician',
    ];
```

## DiagnosisCategory

Source: `hms/app/Models/DiagnosisCategory.php`; table: `diagnosis_categories`

```php
'name' => 'required|unique:diagnosis_categories,name',
```

## Doctor

Source: `hms/app/Models/Doctor.php`; table: `doctors`

```php
'first_name' => 'required|string',
        'last_name' => 'required|string',
        'email' => 'required|email:filter|unique:users,email',
        'password' => 'required|same:password_confirmation|min:6',
        'designation' => 'required|string',
        'gender' => 'required',
        'qualification' => 'required|string',
        'dob' => 'nullable|date',
        'specialist' => 'required|string',
        'address1' => 'nullable|string',
        'address2' => 'nullable|string',
        'city' => 'nullable|string',
        'zip' => 'nullable|integer',
        'description' => 'nullable|string',
```

```php
const GOOGLE_JSON_FILE_PATH = 'google_json_file';
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 0;
```

```php
const INACTIVE = 1;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

## DoctorDepartment

Source: `hms/app/Models/DoctorDepartment.php`; table: `doctor_departments`

```php
'title' => 'required|unique:doctor_departments,title',
```

## DoctorHoliday

Source: `hms/app/Models/DoctorHoliday.php`; table: `doctor_holidays`

```php
'doctor_id' => 'required',
        'date' => 'required',
```

```php
const ALL = 0;
```

```php
const UPCOMING_HOLIDAY = 1;
```

```php
const PAST_HOLIDAY = 2;
```

```php
const TODAY = 3;
```

```php
const ALL_STATUS = [
        self::ALL => 'All',
        self::TODAY => 'Today',
        self::UPCOMING_HOLIDAY => 'Upcoming Holidays',
        self::PAST_HOLIDAY => 'Past Holidays',
    ];
```

## DoctorOPDCharge

Source: `hms/app/Models/DoctorOPDCharge.php`; table: `doctor_opd_charges`

```php
'doctor_id' => 'required|unique:doctor_opd_charges,doctor_id',
        'standard_charge' => 'required',
```

## Document

Source: `hms/app/Models/Document.php`; table: `documents`

```php
'title' => 'required|string',
        'document_type_id' => 'required|integer',
        'patient_id' => 'required|integer',
```

```php
const PATH = 'documents';
```

## DocumentType

Source: `hms/app/Models/DocumentType.php`; table: `document_types`

```php
'name' => 'required|string|unique:document_types,name',
```

## EmailTemplate

Source: `hms/app/Models/EmailTemplate.php`; table: `email_templates`

```php
// No static rules array extracted from this model.
```

```php
const APPOINTMENT_REMINDER_DOCTOR = 'Appointment Reminder To Doctor';
```

```php
const APPOINTMENT_REMINDER_PATIENT = 'Appointment Reminder To Patient';
```

```php
const APPOINTMENT_REMINDER = 'Appointment Reminder';
```

```php
const FORGOT_PASSWORD = 'Forgot Password';
```

```php
const HOSPITAL_ENQUIRY = 'Hospital Enquiry';
```

```php
const EMAIL_VERIFICATION = 'Email Verification';
```

## EmployeePayroll

Source: `hms/app/Models/EmployeePayroll.php`; table: `employee_payrolls`

```php
'sr_no' => 'required|numeric',
        'payroll_id' => 'required',
        'type' => 'required|numeric',
        'owner_id' => 'required',
        'month' => 'required',
        'year' => 'required',
        'net_salary' => 'required',
        'basic_salary' => 'required',
```

```php
const STATUS_ALL = 2;
```

```php
const PAID = 1;
```

```php
const NOT_PAID = 0;
```

```php
const STATUS = [0 => 'Unpaid', 1 => 'Paid'];
```

```php
const MONTHS = [
        1 => 'January',
        2 => 'February',
        3 => 'March',
        4 => 'April',
        5 => 'May',
        6 => 'June',
        7 => 'July',
        8 => 'August',
        9 => 'September',
        10 => 'October',
        11 => 'November',
        12 => 'December',
    ];
```

```php
const TYPES = [
        6 => 'Accountant',
        7 => 'Case Manager',
        2 => 'Doctor',
        3 => 'Lab Technician',
        1 => 'Nurse',
        5 => 'Pharmacist',
        4 => 'Receptionist',
    ];
```

```php
const CLASS_TYPES = [
        1 => Nurse::class,
        2 => Doctor::class,
        3 => LabTechnician::class,
        4 => Receptionist::class,
        5 => Pharmacist::class,
        6 => Accountant::class,
        7 => CaseHandler::class,
    ];
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::PAID => 'Paid',
        self::NOT_PAID => 'Unpaid',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Paid',
        2 => 'Unpaid',
    ];
```

```php
const PYAYROLLUSERS = [
        Doctor::class, Nurse::class, LabTechnician::class, Receptionist::class, Pharmacist::class, Accountant::class,
        CaseHandler::class,
    ];
```

## Enquiry

Source: `hms/app/Models/Enquiry.php`; table: `enquiries`

```php
'full_name' => 'required',
        'email' => 'required|email:filter',
        'contact_no' => 'required|numeric',
        'type' => 'required',
        'message' => 'required|max:5000',
```

```php
const ALL = 2;
```

```php
const READ = 1;
```

```php
const UNREAD = 0;
```

```php
const STATUS_ARR = [
        self::ALL => 'All',
        self::READ => 'Read',
        self::UNREAD => 'Unread',
    ];
```

```php
const TYPE_GENERAL = 'General Inquiry';
```

```php
const TYPE_FEEDBACK = 'Feedback/Suggestions';
```

```php
const TYPE_RESIDENTIAL = 'Residential Care';
```

## EventGoogleCalendar

Source: `hms/app/Models/EventGoogleCalendar.php`; table: `event_google_calendars`

```php
// No static rules array extracted from this model.
```

## Expense

Source: `hms/app/Models/Expense.php`; table: `expenses`

```php
'expense_head' => 'required|string',
        'name' => 'required|unique:expenses,name|string',
        'date' => 'required|date',
        'amount' => 'string|required',
        'description' => 'string|nullable',
        'invoice_number' => 'string|nullable',
```

```php
const PATH = 'expenses';
```

```php
const EXPENSE_HEAD = [
        1 => 'Building Rent',
        3 => 'Equipments',
        2 => 'Electricity Bill',
        4 => 'Power Generator Fuel Charge',
        6 => 'Telephone Bill',
        5 => 'Tea Expense',
    ];
```

```php
const FILTER_EXPENSE_HEAD = [
        0 => 'All',
        1 => 'Building Rent',
        3 => 'Equipments',
        2 => 'Electricity Bill',
        6 => 'Telephone Bill',
        4 => 'Power Generator Fuel Charge',
        5 => 'Tea Expense',
    ];
```

## FrontService

Source: `hms/app/Models/FrontService.php`; table: `front_services`

```php
'name' => 'required',
        'short_description' => 'required',
```

```php
const PATH = 'front-services';
```

## FrontSetting

Source: `hms/app/Models/FrontSetting.php`; table: `front_settings`

```php
'about_us_title' => 'required',
        'about_us_mission' => 'required',
        'about_us_image' => 'nullable',
```

```php
const PATH = 'front-settings';
```

```php
const HOME_IMAGE_PATH = 'homepage-image';
```

```php
const ABOUT_US = 1;
```

```php
const HOME_PAGE = 2;
```

```php
const APPOINTMENT = 3;
```

```php
const STATUS_ARR = [
        self::ABOUT_US => 'About Us',
    ];
```

## GoogleCalendarIntegration

Source: `hms/app/Models/GoogleCalendarIntegration.php`; table: `google_calendar_integrations`

```php
// No static rules array extracted from this model.
```

## GoogleCalendarList

Source: `hms/app/Models/GoogleCalendarList.php`; table: `google_calendar_lists`

```php
// No static rules array extracted from this model.
```

## HospitalSchedule

Source: `hms/app/Models/HospitalSchedule.php`; table: `hospital_schedules`

```php
// No static rules array extracted from this model.
```

```php
const WEEKDAY = [
        self::Mon => 'MON',
        self::Tue => 'TUE',
        self::Wed => 'WED',
        self::Thu => 'THU',
        self::Fri => 'FRI',
        self::Sat => 'SAT',
        self::Sun => 'SUN',
    ];
```

```php
const WEEKDAY_FULL_NAME = [
        self::Mon => 'Monday',
        self::Tue => 'Tuesday',
        self::Wed => 'Wednesday',
        self::Thu => 'Thursday',
        self::Fri => 'Friday',
        self::Sat => 'Saturday',
        self::Sun => 'Sunday',
    ];
```

## Income

Source: `hms/app/Models/Income.php`; table: `incomes`

```php
'income_head' => 'required|string',
        'name' => 'required|unique:incomes,name|string',
        'date' => 'required|date',
        'invoice_number' => 'string|nullable',
        'amount' => 'required',
        'description' => 'string|nullable',
```

```php
const PATH = 'income';
```

```php
const INCOME_HEAD = [
        1 => 'Canteen Rent',
        2 => 'Hospital Charges',
        3 => 'Special Campaign',
        4 => 'Vehicle Stand Charges',
    ];
```

```php
const FILTER_INCOME_HEAD = [
        0 => 'All',
        2 => 'Hospital Charges',
        3 => 'Special Campaign',
        1 => 'Canteen Rent',
        4 => 'Vehicle Stand Charges',
    ];
```

## Insurance

Source: `hms/app/Models/Insurance.php`; table: `insurances`

```php
'name' => 'required|unique:insurances,name',
        'service_tax' => 'required',
        'insurance_no' => 'required',
        'insurance_code' => 'required',
        'hospital_rate' => 'required',
        'discount' => 'required|integer',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARRAY = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## InsuranceDisease

Source: `hms/app/Models/InsuranceDisease.php`; table: `insurance_diseases`

```php
'insurance_id' => 'required',
        'disease_name' => 'required',
        'disease_charge' => 'required',
```

## InvestigationReport

Source: `hms/app/Models/InvestigationReport.php`; table: `investigation_reports`

```php
'patient_id' => 'required|unique:investigation_reports,patient_id',
        'date' => 'required|date',
        'title' => 'required|string',
        'doctor_id' => 'required',
        'status' => 'required',
```

```php
const STATUS = [self::NOT_SOLVED => 'Not Solved', self::SOLVED => 'Solved'];
```

```php
const STATUS_ALL = 0;
```

```php
const SOLVED = 1;
```

```php
const NOT_SOLVED = 2;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::SOLVED => 'Solved',
        self::NOT_SOLVED => 'Not Solved',
    ];
```

```php
const COLLECTION_REPORTS = 'reports';
```

## Invoice

Source: `hms/app/Models/Invoice.php`; table: `invoices`

```php
'patient_id' => 'required',
        'invoice_date' => 'required|date',
        'discount' => 'required|regex:/^\d+(\.\d{1,2})?$/',
```

```php
const PENDING = 1;
```

```php
const PAID = 0;
```

```php
const STATUS_ALL = 2;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::PENDING => 'Pending',
        self::PAID => 'Paid',
    ];
```

```php
const FILTER_STATUS_ARR = [
        2 => 'All',
        0 => 'Paid',
        1 => 'Pending',
    ];
```

## InvoiceItem

Source: `hms/app/Models/InvoiceItem.php`; table: `invoice_items`

```php
'account_id' => 'required|integer',
        'quantity' => 'required|integer',
        'price' => 'required|regex:/^\d+(\.\d{1,2})?$/',
        'description' => 'nullable|string',
```

## IpdBill

Source: `hms/app/Models/IpdBill.php`; table: `ipd_bills`

```php
'ipd_patient_department_id' => 'required',
        'total_payments' => 'required',
        'gross_total' => 'required',
        'discount_in_percentage' => 'required',
        'tax_in_percentage' => 'required',
        'other_charges' => 'required',
        'net_payable_amount' => 'required',
```

## IpdCharge

Source: `hms/app/Models/IpdCharge.php`; table: `ipd_charges`

```php
'date' => 'required',
        'charge_type_id' => 'required',
        'charge_category_id' => 'required',
        'charge_id' => 'required',
        'applied_charge' => 'required',
```

```php
const CHARGE_TYPES = [
        1 => 'Investigations',
        2 => 'Operation Theatre',
        3 => 'Others',
        4 => 'Procedures',
        5 => 'Supplier',
    ];
```

## IpdConsultantRegister

Source: `hms/app/Models/IpdConsultantRegister.php`; table: `ipd_consultant_registers`

```php
'applied_date' => 'required',
        'instruction' => 'required',
        'instruction.*' => 'required',
        'instruction_date' => 'required',
```

## IpdDiagnosis

Source: `hms/app/Models/IpdDiagnosis.php`; table: `ipd_diagnoses`

```php
'report_type' => 'required',
        'report_date' => 'required',
        'file' => 'nullable|mimes:jpeg,png,pdf,docx,doc',
```

```php
const IPD_DIAGNOSIS_PATH = 'ipd_diagnosis';
```

## IpdOperation

Source: `hms/app/Models/IpdOperation.php`; table: `ipd_operation`

```php
'operation_category_id' => 'required',
        'operation_id' => 'required',
        'operation_date' => 'required',
        'doctor_id' => 'required',
```

## IpdPatientDepartment

Source: `hms/app/Models/IpdPatientDepartment.php`; table: `ipd_patient_departments`

```php
'patient_id' => 'required',
        'admission_date' => 'required',
        'doctor_id' => 'required',
        'bed_type_id' => 'required',
        'bed_id' => 'required',
        'case_id' => 'required',
        'weight' => 'numeric|max:200|nullable',
        'height' => 'numeric|nullable',
        'bp' => 'numeric|max:200|nullable',
```

```php
const STATUS_ARR = [
        '' => 'All',
        0 => 'Active',
        1 => 'Discharged',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Discharged',
    ];
```

## IpdPayment

Source: `hms/app/Models/IpdPayment.php`; table: `ipd_payments`

```php
'payment_mode' => 'required',
        'date' => 'required|date',
        'amount' => 'required',
        'notes' => 'nullable|string',
```

```php
const IPD_PAYMENT_PATH = 'ipd_payments';
```

```php
const PAYMENT_MODES_STRIPE = 3;
```

```php
const PAYMENT_MODES_RAZORPAY = 4;
```

```php
const PAYMENT_MODES_FLUTTERWAVE = 8;
```

```php
const PAYMENT_MODES_PHONEPE = 5;
```

```php
const PAYMENT_MODES_PAYSTACK = 6;
```

```php
const PAYMENT_MODES = [
        1 => 'Cash',
        2 => 'Cheque',
        3 => 'Stripe',
        4 => 'Razorpay',
        8 => 'FlutterWave',
        6 => 'Paystack',
        5 => 'PhonePe',
    ];
```

## IpdPrescription

Source: `hms/app/Models/IpdPrescription.php`; table: `ipd_prescriptions`

```php
'date.*' => 'nullable',
        'charge_type_id.*' => 'nullable',
        'category_id.*' => 'required',
        'category_id.*' => 'required',
        'dose_interval.*' => 'required',
        'dosage.*' => 'required',
        'instruction.*' => 'required',
        'day.*' => 'required',
        'time.*' => 'required',
```

## IpdPrescriptionItem

Source: `hms/app/Models/IpdPrescriptionItem.php`; table: `ipd_prescription_items`

```php
'category_id' => 'required',
```

## IpdTimeline

Source: `hms/app/Models/IpdTimeline.php`; table: `ipd_timelines`

```php
'title' => 'required',
        'date' => 'required',
        'attachment' => 'nullable|mimes:jpeg,png,pdf,docx,doc',
```

```php
const IPD_TIMELINE_PATH = 'ipd_timelines';
```

## IssuedItem

Source: `hms/app/Models/IssuedItem.php`; table: `issued_items`

```php
'department_id' => 'required',
        'user_id' => 'required',
        'issued_by' => 'required',
        'issued_date' => 'required',
        'return_date' => 'nullable',
        'item_category_id' => 'required',
        'item_id' => 'required',
        'quantity' => 'required',
        'description' => 'nullable',
        'status' => 'nullable',
```

```php
const ITEM_RETURN = 0;
```

```php
const ITEM_RETURNED = 1;
```

```php
const STATUS_ALL = 2;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ITEM_RETURN => 'Return Item',
        self::ITEM_RETURNED => 'Returned',
    ];
```

## Item

Source: `hms/app/Models/Item.php`; table: `items`

```php
'name' => 'required|unique:items,name',
        'item_category_id' => 'required',
        'unit' => 'required',
        'description' => 'nullable',
```

## ItemCategory

Source: `hms/app/Models/ItemCategory.php`; table: `item_categories`

```php
'name' => 'required|unique:item_categories,name',
```

## ItemStock

Source: `hms/app/Models/ItemStock.php`; table: `item_stocks`

```php
'item_category_id' => 'required',
        'item_id' => 'required',
        'supplier_name' => 'string|nullable',
        'store_name' => 'string|nullable',
        'quantity' => 'numeric|required',
        'purchase_price' => 'required',
        'description' => 'nullable',
        'attachment' => 'nullable|mimes:jpeg,png,pdf,docx,doc',
```

```php
const PATH = 'item_stocks';
```

## LabTechnician

Source: `hms/app/Models/LabTechnician.php`; table: `lab_technicians`

```php
'first_name' => 'required|string',
        'last_name' => 'required|string',
        'email' => 'required|email:filter|unique:users,email',
        'password' => 'required|same:password_confirmation|min:6',
        'designation' => 'required|string',
        'qualification' => 'required|string',
        'address1' => 'nullable|string',
        'address2' => 'nullable|string',
        'city' => 'nullable|string',
        'zip' => 'nullable|integer',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## LiveConsultation

Source: `hms/app/Models/LiveConsultation.php`; table: `live_consultations`

```php
'patient_id' => 'required',
        'doctor_id' => 'required',
        'consultation_title' => 'required',
        'consultation_date' => 'required',
        'consultation_duration_minutes' => 'required|numeric|min:0|max:720',
        'type' => 'required',
        'type_number' => 'required',
        'platform_type' => 'required',
```

```php
const OPD = 0;
```

```php
const IPD = 1;
```

```php
const HOST_ENABLE = 1;
```

```php
const HOST_DISABLED = 0;
```

```php
const CLIENT_ENABLE = 1;
```

```php
const CLIENT_DISABLED = 0;
```

```php
const STATUS_AWAITED = 0;
```

```php
const STATUS_CANCELLED = 1;
```

```php
const STATUS_FINISHED = 2;
```

```php
const ZOOM = 1;
```

```php
const GOOGLE_MEET = 2;
```

```php
const PLATFORM_TYPE = [
        self::ZOOM => 'Zoom',
        self::GOOGLE_MEET => 'Google Meet',
    ];
```

```php
const PLATFORM_TYPE_ZOOM = [
        self::ZOOM => 'Zoom',
    ];
```

```php
const STATUS_TYPE = [
        self::OPD => 'OPD',
        self::IPD => 'IPD',
    ];
```

```php
const FILTER_STATUS = [
        0 => 'All',
        1 => 'Awaited',
        2 => 'Cancelled',
        3 => 'Finished',
    ];
```

## LiveMeeting

Source: `hms/app/Models/LiveMeeting.php`; table: `live_meetings`

```php
'consultation_title' => 'required',
        'consultation_date' => 'required',
        'consultation_duration_minutes' => 'required|min:0|max:720',
```

```php
const STATUS_AWAITED = 0;
```

```php
const STATUS_CANCELLED = 1;
```

```php
const STATUS_FINISHED = 2;
```

```php
const FILTER_STATUS = [
        0 => 'All',
        1 => 'Awaited',
        2 => 'Cancelled',
        3 => 'Finished',
    ];
```

## LunchBreak

Source: `hms/app/Models/LunchBreak.php`; table: `lunch_breaks`

```php
'doctor_id' => 'required',
        'break_from' => 'required',
        'break_to' => 'required',
```

## Mail

Source: `hms/app/Models/Mail.php`; table: `mails`

```php
'to' => 'required|email',
        'subject' => 'required|string',
        'message' => 'required',
        'attachments' => 'nullable|mimes:jpeg,gif,png,jpg,mp3',
```

## ManualBillPayment

Source: `hms/app/Models/ManualBillPayment.php`; table: `bill_transactions`

```php
// No static rules array extracted from this model.
```

```php
const STATUS_ARR = [
        self::Approved => 'Approved',
        self::Rejected => 'Rejected',
    ];
```

## Medicine

Source: `hms/app/Models/Medicine.php`; table: `medicines`

```php
'category_id' => 'required',
        'brand_id' => 'required',
        'name' => 'required|min:2|unique:medicines,name',
        'selling_price' => 'required',
        'buying_price' => 'required',
        'side_effects' => 'nullable',
        'salt_composition' => 'required|nullable|string',
        'quantity' => 'required|integer',
        'available_quantity' => 'required|integer|lte:quantity',
```

## MedicineBill

Source: `hms/app/Models/MedicineBill.php`; table: `medicine_bills`

```php
// No static rules array extracted from this model.
```

```php
const MEDICINE_BILL_CASH = 0;
```

```php
const MEDICINE_BILL_CHEQUE = 1;
```

```php
const MEDICINE_BILL_RAZORPAY = 2;
```

```php
const MEDICINE_BILL_PAYSTACK = 3;
```

```php
const MEDICINE_BILL_PHONEPE = 4;
```

```php
const MEDICINE_BILL_STRIPE = 5;
```

```php
const MEDICINE_BILL_FLUTTERWAVE = 6;
```

```php
const UNPAID = 0;
```

```php
const FULLPAID = 1;
```

```php
const PARTIALY_PAID = 2;
```

```php
const PAYMENT_STATUS_ARRAY = [
        self::UNPAID => 'Unpaid',
        self::FULLPAID => 'Full Paid',
        self::PARTIALY_PAID => 'Partially Paid',
    ];
```

```php
const PAYMENT_TYPE = [
        self::MEDICINE_BILL_CASH => 'Cash',
        self::MEDICINE_BILL_CHEQUE => 'Cheque',
    ];
```

## Module

Source: `hms/app/Models/Module.php`; table: `modules`

```php
'name' => 'required|min:2|unique:modules,name',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## NoticeBoard

Source: `hms/app/Models/NoticeBoard.php`; table: `notice_boards`

```php
'title' => 'required|string',
```

## Notification

Source: `hms/app/Models/Notification.php`; table: `notifications`

```php
// No static rules array extracted from this model.
```

```php
const ADMIN = 'Admin';
```

```php
const DOCTOR = 'Doctor';
```

```php
const RECEPTIONIST = 'Receptionist';
```

```php
const PATIENT = 'Patient';
```

```php
const NURSE = 'Nurse';
```

```php
const CASE_HANDLER = 'Case Manager';
```

```php
const PHARMACIST = 'Pharmacist';
```

```php
const ACCOUNTANT = 'Accountant';
```

```php
const LAB_TECHNICIAN = 'Lab Technician';
```

```php
const NOTIFICATION_FOR = [
        self::ADMIN => 1,
        self::DOCTOR => 2,
        self::RECEPTIONIST => 3,
        self::PATIENT => 4,
        self::NURSE => 5,
        self::CASE_HANDLER => 6,
        self::PHARMACIST => 7,
        self::ACCOUNTANT => 8,
        self::LAB_TECHNICIAN => 9,
    ];
```

```php
const NOTIFICATION_TYPE = [
        'Appointment' => 1,
        'Invoice' => 2,
        'Bills' => 3,
        'IPD Patient' => 4,
        'OPD Patient' => 5,
        'Prescription' => 6,
        'IPD Charge' => 7,
        'IPD Prescription' => 8,
        'OPD Diagnosis' => 9,
        'Employee Payrolls' => 10,
        'Advance Payment' => 11,
        'Patient' => 12,
        'Cases' => 13,
        'Visitor' => 14,
        'NoticeBoard' => 15,
        'Bed Assign' => 16,
        'Ambulance' => 17,
        'Service' => 18,
        'Income' => 19,
        'Expense' => 20,
        'Live Consultation' => 21,
        'Live Meeting' => 22,
        'OPD Prescription' => 23,
    ];
```

## Nurse

Source: `hms/app/Models/Nurse.php`; table: `nurses`

```php
'first_name' => 'required|string',
        'last_name' => 'required|string',
        'email' => 'required|email:filter|unique:users,email',
        'password' => 'required|same:password_confirmation|min:6',
        'designation' => 'required|string',
        'qualification' => 'required|string',
        'address1' => 'nullable|string',
        'address2' => 'nullable|string',
        'city' => 'nullable|string',
        'zip' => 'nullable|integer',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## Odontogram

Source: `hms/app/Models/Odontogram.php`; table: `odontograms`

```php
'odontogram' => 'required',
        'patient_id' => 'required',
        'doctor_id' => 'required',
        'description' => 'required',
```

## OpdDiagnosis

Source: `hms/app/Models/OpdDiagnosis.php`; table: `opd_diagnoses`

```php
'report_type' => 'required',
        'report_date' => 'required',
        'file' => 'nullable|mimes:jpeg,png,pdf,docx,doc',
```

```php
const OPD_DIAGNOSIS_PATH = 'opd_diagnosis';
```

## OpdPatientDepartment

Source: `hms/app/Models/OpdPatientDepartment.php`; table: `opd_patient_departments`

```php
'patient_id' => 'required',
        'appointment_date' => 'required',
        'doctor_id' => 'required',
        'standard_charge' => 'required',
        'payment_mode' => 'required',
        'weight' => 'numeric|max:200|nullable',
        'height' => 'numeric|nullable',
        'bp' => 'nullable|numeric|max:200',
```

```php
const PAYMENT_MODES = [
        1 => 'Cash',
        2 => 'Cheque',
    ];
```

## OpdPrescription

Source: `hms/app/Models/OpdPrescription.php`; table: `opd_prescriptions`

```php
'category_id.*' => 'required',
        'dosage.*' => 'required',
        'day.*' => 'required',
        'dose_interval.*' => 'required',
        'time.*' => 'required',
        'instruction.*' => 'required',
```

## OpdPrescriptionItem

Source: `hms/app/Models/OpdPrescriptionItem.php`; table: `opd_prescription_items`

```php
'category_id' => 'required',
```

## OpdTimeline

Source: `hms/app/Models/OpdTimeline.php`; table: `opd_timelines`

```php
'title' => 'required',
        'date' => 'required',
        'attachment' => 'nullable|mimes:jpeg,png,pdf,docx,doc',
```

```php
const OPD_TIMELINE_PATH = 'opd_timelines';
```

## Operation

Source: `hms/app/Models/Operation.php`; table: `operations`

```php
'operation_category_id' => 'required',
        'name' => 'required|unique:operations,name',
```

## OperationCategory

Source: `hms/app/Models/OperationCategory.php`; table: `operation_categories`

```php
'name' => 'required|unique:operation_categories,name',
```

## OperationReport

Source: `hms/app/Models/OperationReport.php`; table: `operation_reports`

```php
'case_id' => 'required|unique:operation_reports,case_id',
        'doctor_id' => 'required',
        'date' => 'required|date',
```

## Package

Source: `hms/app/Models/Package.php`; table: `packages`

```php
'name' => 'required|string|unique:packages,name',
        'discount' => 'required|integer',
        'total_amount' => 'required',
```

## PackageService

Source: `hms/app/Models/PackageService.php`; table: `package_services`

```php
'service_id' => 'required|integer',
        'quantity' => 'required|integer',
        'rate' => 'required|regex:/^\d*(\.\d{1,2})?$/',
```

## PathologyCategory

Source: `hms/app/Models/PathologyCategory.php`; table: `pathology_categories`

```php
'name' => 'required|unique:pathology_categories,name',
```

## PathologyParameter

Source: `hms/app/Models/PathologyParameter.php`; table: `pathology_parameters`

```php
'parameter_name' => 'required|unique:pathology_parameters,parameter_name',
        'reference_range' => 'required',
        'unit_id' => 'required',
```

## PathologyParameterItem

Source: `hms/app/Models/PathologyParameterItem.php`; table: `pathology_parameter_items`

```php
// No static rules array extracted from this model.
```

## PathologyTest

Source: `hms/app/Models/PathologyTest.php`; table: `pathology_tests`

```php
'test_name' => 'required|unique:pathology_tests,test_name',
        'short_name' => 'required',
        'test_type' => 'required',
        'category_id' => 'required',
        'charge_category_id' => 'required',
        'standard_charge' => 'required',
        'patient_id' => 'required'
```

## PathologyUnit

Source: `hms/app/Models/PathologyUnit.php`; table: `pathology_units`

```php
'name' => 'required|unique:pathology_units,name',
```

## Patient

Source: `hms/app/Models/Patient.php`; table: `patients`

```php
'first_name' => 'required|string',
        'last_name' => 'required|string',
        'email' => 'required|email:filter|unique:users,email',
        'password' => 'required|same:password_confirmation|min:6',
        'gender' => 'required',
        'dob' => 'nullable|date',
        'phone' => 'nullable|numeric',
        'address1' => 'nullable|string',
        'address2' => 'nullable|string',
        'city' => 'nullable|string',
        'zip' => 'nullable|integer',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## PatientAdmission

Source: `hms/app/Models/PatientAdmission.php`; table: `patient_admissions`

```php
'patient_id' => 'required',
        'doctor_id' => 'required',
        'admission_date' => 'required',
        'policy_no' => 'string|nullable',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## PatientCase

Source: `hms/app/Models/PatientCase.php`; table: `patient_cases`

```php
'patient_id' => 'required',
        'phone' => 'nullable|numeric',
        'doctor_id' => 'required',
        'date' => 'required',
        'description' => 'nullable',
        'fee' => 'required',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## PatientDiagnosisProperty

Source: `hms/app/Models/PatientDiagnosisProperty.php`; table: `patient_diagnosis_properties`

```php
// No static rules array extracted from this model.
```

## PatientDiagnosisTest

Source: `hms/app/Models/PatientDiagnosisTest.php`; table: `patient_diagnosis_tests`

```php
'patient_id' => 'required|unique:patient_diagnosis_tests,patient_id',
        'category_id' => 'required',
        'age' => 'integer|nullable',
        'height' => 'integer|nullable',
        'weight' => 'integer|nullable',
        'blood_pressure' => 'integer|nullable',
        'average_glucose' => 'string|nullable',
        'fasting_blood_sugar' => 'string|nullable',
        'urine_sugar' => 'string|nullable',
        'cholesterol' => 'string|nullable',
        'diabetes' => 'string|nullable',
```

## PatientIdCardTemplate

Source: `hms/app/Models/PatientIdCardTemplate.php`; table: `patient_id_card_templates`

```php
'name' => 'required|unique:patient_id_card_templates',
        'color' => 'required',
```

## PatientQueue

Source: `hms/app/Models/PatientQueue.php`; table: `patient_queues`

```php
// No static rules array extracted from this model.
```

## Payment

Source: `hms/app/Models/Payment.php`; table: `payments`

```php
'payment_date' => 'required',
        'account_id' => 'required',
        'pay_to' => 'required',
        'amount' => 'required',
        'description' => 'nullable|string',
```

## Permission

Source: `hms/app/Models/Permission.php`; table: `permissions`

```php
// No static rules array extracted from this model.
```

## Pharmacist

Source: `hms/app/Models/Pharmacist.php`; table: `pharmacists`

```php
'first_name' => 'required|string',
        'last_name' => 'required|string',
        'email' => 'required|email:filter|unique:users,email',
        'designation' => 'required|string',
        'qualification' => 'required|string',
        'password' => 'required|same:password_confirmation|min:6',
        'gender' => 'required',
        'dob' => 'nullable|date',
        'address1' => 'nullable|string',
        'address2' => 'nullable|string',
        'city' => 'nullable|string',
        'zip' => 'nullable|integer',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## Postal

Source: `hms/app/Models/Postal.php`; table: `postals`

```php
'from_title' => 'required_if:type,==,'.self::POSTAL_RECEIVE.'|string|nullable',
        'to_title' => 'required_if:type,==,'.self::POSTAL_DISPATCH.'|string|nullable',
        'reference_no' => 'string|nullable',
        'date' => 'date|nullable',
        'address' => 'string|nullable',
```

```php
const PATH = 'postal';
```

```php
const POSTAL_RECEIVE = 1;
```

```php
const POSTAL_DISPATCH = 2;
```

## Prescription

Source: `hms/app/Models/Prescription.php`; table: `prescriptions`

```php
'patient_id' => 'required',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const DAYS = 0;
```

```php
const MONTH = 1;
```

```php
const YEAR = 2;
```

```php
const TIME_ARR = [
        self::DAYS => 'Days',
        self::MONTH => 'Month',
        self::YEAR => 'Years',
    ];
```

```php
const AFETR_MEAL = 0;
```

```php
const BEFORE_MEAL = 1;
```

```php
const MEAL_ARR = [
        self::AFETR_MEAL => 'After Meal',
        self::BEFORE_MEAL => 'Before Meal',
    ];
```

```php
const ONE_TIME = 1;
```

```php
const TWO_TIME = 2;
```

```php
const THREE_TIME = 3;
```

```php
const FOUR_TIME = 4;
```

```php
const DOSE_INTERVAL = [
        self::ONE_TIME => 'Daily morning',
        self::TWO_TIME => 'Daily morning and evening',
        self::THREE_TIME => 'Daily morning, noon, and evening',
        self::FOUR_TIME => '4 times in a day',
    ];
```

```php
const ONE_DAY = 1;
```

```php
const THREE_DAY = 3;
```

```php
const ONE_WEEK = 7;
```

```php
const TWO_WEEK = 14;
```

```php
const ONE_MONTH = 30;
```

```php
const DOSE_DURATION = [
        self::ONE_DAY => 'Only one day',
        self::THREE_DAY => 'Upto Three days',
        self::ONE_WEEK => 'Upto One week',
        self::TWO_WEEK => 'Upto two week',
        self::ONE_MONTH => 'Upto one month',
    ];
```

## PrescriptionMedicineModal

Source: `hms/app/Models/PrescriptionMedicineModal.php`; table: `prescriptions_medicines`

```php
// No static rules array extracted from this model.
```

## PurchaseMedicine

Source: `hms/app/Models/PurchaseMedicine.php`; table: `purchase_medicines`

```php
// No static rules array extracted from this model.
```

```php
const PURCHASE_MEDICINE_CASH = 0;
```

```php
const PURCHASE_MEDICINE_CHEQUE = 1;
```

```php
const PURCHASE_MEDICINE_STRIPE = 5;
```

```php
const PURCHASE_MEDICINE_RAZORPAY = 2;
```

```php
const PURCHASE_MEDICINE_PAYSTACK = 3;
```

```php
const PURCHASE_MEDICINE_PHONEPE = 4;
```

```php
const PURCHASE_MEDICINE_FLUTTERWAVE = 6;
```

## PurchasedMedicine

Source: `hms/app/Models/PurchasedMedicine.php`; table: `purchased_medicines`

```php
// No static rules array extracted from this model.
```

## RadiologyCategory

Source: `hms/app/Models/RadiologyCategory.php`; table: `radiology_categories`

```php
'name' => 'required|unique:radiology_categories,name',
```

## RadiologyTest

Source: `hms/app/Models/RadiologyTest.php`; table: `radiology_tests`

```php
'test_name' => 'required|unique:radiology_tests,test_name',
        'short_name' => 'required',
        'test_type' => 'required',
        'category_id' => 'required',
        'charge_category_id' => 'required',
        'standard_charge' => 'required',
```

## Receptionist

Source: `hms/app/Models/Receptionist.php`; table: `receptionists`

```php
'first_name' => 'required|string',
        'last_name' => 'required|string',
        'email' => 'required|email:filter|unique:users,email',
        'password' => 'required|same:password_confirmation|min:6',
        'designation' => 'required|string',
        'qualification' => 'required|string',
        'address1' => 'nullable|string',
        'address2' => 'nullable|string',
        'city' => 'nullable|string',
        'zip' => 'nullable|integer',
```

```php
const STATUS_ALL = 0;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 2;
```

```php
const STATUS_INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

## SaleMedicine

Source: `hms/app/Models/SaleMedicine.php`; table: `sale_medicines`

```php
// No static rules array extracted from this model.
```

## Schedule

Source: `hms/app/Models/Schedule.php`; table: `schedules`

```php
'doctor_id' => 'required|unique:schedules,doctor_id',
        'available_on' => 'required',
        'available_from' => 'required',
        'available_to' => 'required',
        'per_patient_time' => 'required',
```

```php
const ALL = 0;
```

## ScheduleDay

Source: `hms/app/Models/ScheduleDay.php`; table: `schedule_days`

```php
'doctor_id' => 'required',
        'available_on' => 'required',
        'available_from' => 'required',
        'available_to' => 'required',
```

## Service

Source: `hms/app/Models/Service.php`; table: `services`

```php
'name' => 'required|unique:services,name',
        'quantity' => 'required|numeric',
        'rate' => 'required',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARRAY = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

## Setting

Source: `hms/app/Models/Setting.php`; table: `settings`

```php
'app_name' => 'required|string',
        'company_name' => 'required|string',
        'app_logo' => 'nullable|mimes:jpg,jpeg,png',
        'favicon' => 'nullable|mimes:jpg,jpeg,png',
        'hospital_email' => 'required|email:rfc,dns',
```

```php
const PATH = 'settings';
```

```php
const VIDEO_MEDIA_COLLECTION = 'patient_queue_theme_video';
```

```php
const CURRENCIES = [
        'eur' => 'Euro (EUR)',
        'aud' => 'Australia Dollar (AUD)',
        'inr' => 'India Rupee (INR)',
        'usd' => 'USA Dollar (USD)',
        'jpy' => 'Japanese Yen (JPY)',
        'gbp' => 'British Pound (GBP)',
        'cad' => 'Canadian Dollar (CAD)',
    ];
```

## Sms

Source: `hms/app/Models/Sms.php`; table: `sms`

```php
'message' => 'required|max:160',
```

```php
const ROLE_TYPES = [
        1 => 'Doctor',
        2 => 'Accountant',
        3 => 'Nurse',
        4 => 'LabTechnician',
        5 => 'Receptionist',
        6 => 'Pharmacist',
        7 => 'Case Handler',
        8 => 'Patient',
    ];
```

```php
const CLASS_TYPES = [
        1 => Doctor::class,
        2 => Accountant::class,
        3 => Nurse::class,
        4 => LabTechnician::class,
        5 => Receptionist::class,
        6 => Pharmacist::class,
        7 => CaseHandler::class,
        8 => Patient::class,
    ];
```

## Testimonial

Source: `hms/app/Models/Testimonial.php`; table: `testimonials`

```php
'name' => 'required|string',
        'description' => 'required|string',
```

```php
const PATH = 'testimonials';
```

## Transaction

Source: `hms/app/Models/Transaction.php`; table: `transactions`

```php
// No static rules array extracted from this model.
```

## UsedMedicine

Source: `hms/app/Models/UsedMedicine.php`; table: `used_medicines`

```php
// No static rules array extracted from this model.
```

## User

Source: `hms/app/Models/User.php`; table: `users`

```php
'first_name' => 'required|string',
        'last_name' => 'required|string',
        'email' => 'required|email:filter|unique:users,email',
        'password' => 'required|same:password_confirmation|min:6',
        'department_id' => 'required|string',
        'gender' => 'required|string',
        'dob' => 'nullable|date',
        'phone' => 'required|numeric',
        'address1' => 'nullable|string',
        'address2' => 'nullable|string',
        'city' => 'nullable|string',
        'zip' => 'nullable|integer',
        'image' => 'nullable|mimes:jpg,jpeg,png',
```

```php
const COLLECTION_PROFILE_PICTURES = 'profile_photo';
```

```php
const COLLECTION_MAIL_ATTACHMENTS = 'mail_attachments';
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 0;
```

```php
const INACTIVE = 1;
```

```php
const STATUS_ARR = [
        self::ACTIVE => 'Active',
        self::INACTIVE => 'InActive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```

```php
const THEME_DARK_MODE = 1;
```

```php
const THEME_LIGHT_MODE = 0;
```

```php
const LANGUAGES = [
        'ar' => 'Arabic',
        'zh' => 'Chinese',
        'en' => 'English',
        'fr' => 'French',
        'de' => 'German',
        'pt' => 'Portuguese',
        'ru' => 'Russian',
        'es' => 'Spanish',
        'tr' => 'Turkish',
    ];
```

```php
const MAIN_IPD = 'IPD';
```

```php
const MAIN_OPD = 'OPD';
```

```php
const MAIN_BED_MGT = 'MAIN_BED_MGT';
```

```php
const MAIN_BILLING_MGT = 'MAIN_BILLING_MGT';
```

```php
const MAIN_BLOOD_BANK_MGT = 'MAIN_BLOOD_BANK_MGT';
```

```php
const MAIN_DOCUMENT = 'MAIN_DOCUMENT';
```

```php
const MAIN_DOCTOR = 'MAIN_DOCTOR';
```

```php
const MAIN_PRESCRIPTION = 'MAIN_PRESCRIPTION';
```

```php
const MAIN_DIAGNOSIS = 'MAIN_DIAGNOSIS';
```

```php
const MAIN_FINANCE = 'MAIN_FINANCE';
```

```php
const MAIN_FRONT_OFFICE = 'MAIN_FRONT_OFFICE';
```

```php
const MAIN_HOSPITAL_CHARGE = 'MAIN_HOSPITAL_CHARGE';
```

```php
const MAIN_INVENTORY = 'MAIN_INVENTORY';
```

```php
const MAIN_LIVE_CONSULATION = 'MAIN_LIVE_CONSULATION';
```

```php
const MAIN_MEDICINES = 'MAIN_MEDICINES';
```

```php
const MAIN_PATIENT_CASE = 'MAIN_PATIENT_CASE';
```

```php
const MAIN_PATHOLOGY = 'MAIN_PATHOLOGY';
```

```php
const MAIN_REPORT = 'MAIN_REPORT';
```

```php
const MAIN_RADIOLOGY = 'MAIN_RADIOLOGY';
```

```php
const MAIN_SERVICE = 'MAIN_SERVICE';
```

```php
const MAIN_SMS_MAIL = 'MAIN_SMS_MAIL';
```

```php
const MAIN_DOCTOR_BED_MGT = 'MAIN_DOCTOR_BED_MGT';
```

```php
const MAIN_DOCTOR_PATIENT_CASE = 'MAIN_DOCTOR_PATIENT_CASE';
```

```php
const MAIN_CASE_MANGER_PATIENT_CASE = 'MAIN_CASE_MANGER_PATIENT_CASE';
```

```php
const MAIN_CASE_MANGER_SERVICE = 'MAIN_CASE_MANGER_SERVICE';
```

```php
const MAIN_ACCOUNT_MANAGER_MGT = 'MAIN_ACCOUNT_MANAGER_MGT';
```

```php
const MAIN_VACCINATION_MGT = 'MAIN_VACCINATION_MGT';
```

```php
const LANGUAGES_IMAGE = [
        'ar' => 'assets/img/iraq.svg',
        'zh' => 'assets/img/china.svg',
        'en' => 'assets/img/united-states.svg',
        'fr' => 'assets/img/france.svg',
        'de' => 'assets/img/germany.svg',
        'pt' => 'assets/img/portugal.svg',
        'ru' => 'assets/img/russia.svg',
        'es' => 'assets/img/spain.svg',
        'tr' => 'assets/img/turkey.svg',
    ];
```

```php
const IMG_COLUMN = 'image_url';
```

## UserGoogleEventSchedule

Source: `hms/app/Models/UserGoogleEventSchedule.php`; table: `user_google_event_schedules`

```php
// No static rules array extracted from this model.
```

## UserZoomCredential

Source: `hms/app/Models/UserZoomCredential.php`; table: `user_zoom_credential`

```php
'zoom_api_key' => 'required',
        'zoom_api_secret' => 'required',
```

## VaccinatedPatients

Source: `hms/app/Models/VaccinatedPatients.php`; table: `vaccinated_patients`

```php
'patient_id' => 'required',
        'vaccination_id' => 'required',
        'vaccination_serial_no' => 'string|nullable',
        'dose_number' => 'required|numeric|digits_between:1,50',
        'dose_given_date' => 'required',
```

## Vaccination

Source: `hms/app/Models/Vaccination.php`; table: `vaccinations`

```php
'name' => 'required',
        'manufactured_by' => 'required',
        'brand' => 'required',
```

## Visitor

Source: `hms/app/Models/Visitor.php`; table: `visitors`

```php
'purpose' => 'required|string',
        'name' => 'required|string',
        'id_card' => 'string|nullable',
        'no_of_person' => 'integer|nullable',
        'date' => 'date|nullable',
        'in_time' => 'string|nullable',
        'out_time' => 'string|nullable',
        'note' => 'string|nullable',
```

```php
const PATH = 'visitors';
```

```php
const PURPOSE = [
        3 => 'Visit',
        1 => 'Enquiry',
        2 => 'Seminar',
    ];
```

```php
const FILTER_PURPOSE = [
        0 => 'All',
        3 => 'Visit',
        1 => 'Enquiry',
        2 => 'Seminar',
    ];
```

## ZoomOAuth

Source: `hms/app/Models/ZoomOAuth.php`; table: `zoom_o_auth_credentials`

```php
// No static rules array extracted from this model.
```

## admin

Source: `hms/app/Models/admin.php`; table: `admins`

```php
'first_name' => 'required|string',
        'last_name' => 'required|string',
        'email' => 'required|email:filter|unique:users,email',
        'password' => 'required|same:password_confirmation|min:6',
        'image' => 'image|mimes:jpeg,jpg,png,gif',
```

```php
const STATUS_ALL = 2;
```

```php
const ACTIVE = 1;
```

```php
const INACTIVE = 0;
```

```php
const STATUS_ARR = [
        self::STATUS_ALL => 'All',
        self::ACTIVE => 'Active',
        self::INACTIVE => 'Deactive',
    ];
```

```php
const FILTER_STATUS_ARR = [
        0 => 'All',
        1 => 'Active',
        2 => 'Deactive',
    ];
```
