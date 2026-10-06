# Original controller workflow register

Rules/guards below are static source excerpts. They are not proof that all runtime paths were executed or that legacy behavior is safe to copy. Read the complete method for context. Models and full validation arrays are in ../legacy-schema/inventory.json.

## hms/app/Http/Controllers/API/AppointmentAPIController.php

Models: `Appointment` → `appointments`

### __construct (line 17)

`public function __construct(AppointmentRepository $appointmentRepo)`

### index (line 22)

`public function index(): JsonResponse`

### filter (line 34)

`public function filter(Request $request): JsonResponse`

### getDoctorDepartment (line 47)

`public function getDoctorDepartment(): JsonResponse`

### getDoctors (line 54)

`public function getDoctors($id): JsonResponse`

### bookingSlots (line 61)

`public function bookingSlots(Request $request): JsonResponse`

### create (line 74)

`public function create()`

### store (line 82)

`public function store(CreateAppointmentRequest $request): JsonResponse`

Guard/validation candidates:

```php
86: if (Appointment::where('opd_date', $input['opd_date'])->first()) {
92: if ($success) {
```

Write/transaction candidates:

```php
91: $success = $this->appointmentRepository->create($input);
```

### cancelAppointment (line 99)

`public function cancelAppointment(Request $request): JsonResponse`

Guard/validation candidates:

```php
103: if (! $appointment) {
```

Write/transaction candidates:

```php
107: $appointment->update(['is_completed' => Appointment::STATUS_CANCELLED]);
```

### show (line 117)

`public function show(int $id)`

### edit (line 127)

`public function edit(int $id)`

### update (line 137)

`public function update(Request $request, int $id)`

### destroy (line 148)

`public function destroy(Request $request): JsonResponse`

Guard/validation candidates:

```php
152: if (! $appointment) {
```

Write/transaction candidates:

```php
156: $appointment->delete($request->id);
```

## hms/app/Http/Controllers/API/AuthController.php

Models: `User` → `users`

### login (line 22)

`public function login(Request $request): JsonResponse`

Guard/validation candidates:

```php
27: if (empty($email) or empty($password)) {
32: if (empty($user)) {
36: if (! Hash::check($password, $user->password)) {
42: if ($user->hasRole('Doctor')) {
48: } elseif ($user->hasRole('Patient')) {
```

### logout (line 62)

`public function logout(): JsonResponse`

Write/transaction candidates:

```php
64: auth()->user()->tokens()->where('id', Auth::user()->currentAccessToken()->id)->delete();
```

### sendPasswordResetLinkEmail (line 72)

`public function sendPasswordResetLinkEmail(Request $request): JsonResponse`

Guard/validation candidates:

```php
75: $request->validate([
83: if (! $data['user']) {
98: if ($user) {
```

Write/transaction candidates:

```php
99: DB::table('password_reset_tokens')->where('email', $user->email)->update([
```

### resetPassword (line 118)

`public function resetPassword(Request $request): JsonResponse`

Guard/validation candidates:

```php
120: $request->validate([
129: if (! $tokenData) {
135: if (! $user) {
```

Write/transaction candidates:

```php
140: $user->save();
143: ->where('token', $request->token)->delete();
```

### changePassword (line 169)

`public function changePassword(Request $request): JsonResponse`

Guard/validation candidates:

```php
171: $request->validate([
178: if (! Hash::check($request->old_password, $user->password)) {
```

Write/transaction candidates:

```php
183: $user->save();
```

## hms/app/Http/Controllers/API/BillAPIController.php

Models: `Bill` → `bills`

### index (line 10)

`public function index(): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
12: if (getLoggedinPatient()) {
```

### show (line 23)

`public function show($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
27: if (! $bill) {
```

## hms/app/Http/Controllers/API/BirthReportAPIController.php

Models: `BirthReport` → `birth_reports`, `Doctor` → `doctors`

### index (line 11)

`public function index(): \Illuminate\Http\JsonResponse`

### show (line 24)

`public function show($id): \Illuminate\Http\JsonResponse`

### delete (line 34)

`public function delete($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
38: if (! $birthReport || $birthReport->doctor_id != getLoggedInUser()->owner_id) {
```

Write/transaction candidates:

```php
41: $birthReport->delete();
```

## hms/app/Http/Controllers/API/DiagnosisTestAPIController.php

Models: `PatientDiagnosisTest` → `patient_diagnosis_tests`

### __construct (line 19)

`public function __construct(PatientDiagnosisTestRepository $patientDiagnosisTestRepository)`

### index (line 24)

`public function index(): JsonResponse`

Guard/validation candidates:

```php
26: if (getLoggedinPatient() || getLoggedinDoctor()) {
29: if ($user->hasRole('Patient')) {
32: if ($user->hasRole('Doctor')) {
```

### show (line 46)

`public function show($id): JsonResponse`

Guard/validation candidates:

```php
48: if (getLoggedinPatient()) {
52: if (! $diagnosis) {
59: if (getLoggedinDoctor()) {
63: if (! $diagnosis) {
```

### destroy (line 74)

`public function destroy($id): JsonResponse`

Guard/validation candidates:

```php
77: if (empty($diagnosis_test) || $diagnosis_test->doctor_id != getLoggedInUser()->owner_id) {
```

Write/transaction candidates:

```php
80: $this->patientDiagnosisTestRepository->delete($id);
```

## hms/app/Http/Controllers/API/Doctor/DoctorAppointmentAPIController.php

Models: `Appointment` → `appointments`

### __construct (line 16)

`public function __construct(AppointmentRepository $appointmentRepo)`

### index (line 21)

`public function index(): \Illuminate\Http\JsonResponse`

### filter (line 32)

`public function filter(Request $request): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
38: if ($status == 'all') {
```

### confirmAppointment (line 88)

`public function confirmAppointment($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
91: if (! $appointment) {
```

Write/transaction candidates:

```php
95: $appointment->update(['is_completed' => Appointment::STATUS_COMPLETED]);
```

## hms/app/Http/Controllers/API/Doctor/DoctorBedAssignController.php

Models: `Bed` → `beds`, `BedAssign` → `bed_assigns`, `BedType` → `bed_types`, `IpdPatientDepartment` → `ipd_patient_departments`, `PatientCase` → `patient_cases`

### __construct (line 21)

`public function __construct(BedAssignRepository $bedAssignRepo)`

### index (line 26)

`public function index(): \Illuminate\Http\JsonResponse`

### filter (line 37)

`public function filter(Request $request): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
41: if ($status == 'all') {
```

### show (line 70)

`public function show($id): \Illuminate\Http\JsonResponse`

### patientCase (line 78)

`public function patientCase(): \Illuminate\Http\JsonResponse`

### ipdPatient (line 90)

`public function ipdPatient($caseId): \Illuminate\Http\JsonResponse`

### getBeds (line 100)

`public function getBeds(): \Illuminate\Http\JsonResponse`

### getEditBeds (line 107)

`public function getEditBeds(Request $request): \Illuminate\Http\JsonResponse`

### store (line 114)

`public function store(CreateBedAssignRequest $request): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
122: if (! empty($birthDate) && $assign_date < $birthDate) {
```

Write/transaction candidates:

```php
125: $this->bedAssignRepository->store($input);
```

### edit (line 130)

`public function edit($id): \Illuminate\Http\JsonResponse`

### update (line 137)

`public function update(Request $request, $id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
145: if (! empty($birthDate) && $assign_date < $birthDate) {
```

Write/transaction candidates:

```php
148: $this->bedAssignRepository->update($input, $bedAssign);
```

### delete (line 156)

`public function delete($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
159: if (! $bedAssign) {
```

Write/transaction candidates:

```php
162: $bedAssign->bed->update(['is_available' => 1]);
163: $this->bedAssignRepository->delete($bedAssign->id);
```

### showBedStatus (line 168)

`public function showBedStatus(): \Illuminate\Http\JsonResponse`

### showBedStatusDetail (line 182)

`public function showBedStatusDetail($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
188: if (! $data) {
192: if ($data->bedAssigns->isNotEmpty()) {
```

## hms/app/Http/Controllers/API/Doctor/DoctorDeathReportAPIController.php

Models: `DeathReport` → `death_reports`

### index (line 10)

`public function index(): \Illuminate\Http\JsonResponse`

### delete (line 22)

`public function delete($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
26: if (! $death_reports) {
```

Write/transaction candidates:

```php
30: $death_reports->delete();
```

## hms/app/Http/Controllers/API/Doctor/DoctorInvestigationReportController.php

Models: `InvestigationReport` → `investigation_reports`

### index (line 10)

`public function index(): \Illuminate\Http\JsonResponse`

### delete (line 23)

`public function delete($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
27: if (! $investigation_reports || $investigation_reports->doctor_id != getLoggedInUser()->owner_id) {
```

Write/transaction candidates:

```php
30: $investigation_reports->delete();
```

## hms/app/Http/Controllers/API/Doctor/DoctorLiveConsultationAPIController.php

Models: `IpdPatientDepartment` → `ipd_patient_departments`, `LiveConsultation` → `live_consultations`, `OpdPatientDepartment` → `opd_patient_departments`

### index (line 14)

`public function index(): \Illuminate\Http\JsonResponse`

### liveConsultancyMeeting (line 26)

`public function liveConsultancyMeeting($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
31: if (! $live_consultation) {
35: if($live_consultation->status == LiveConsultation::STATUS_CANCELLED || $live_consultation->status == LiveConsultation::STATUS_FINISHED){
```

### filter (line 43)

`public function filter(Request $request): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
47: if ($status == 'all') {
69: if(!$data){
85: if(!$data){
102: if(!$data){
```

### show (line 112)

`public function show($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
117: if (! $live_consultation) {
```

### getLiveConsultancy (line 125)

`public function getLiveConsultancy(): Builder`

### meeting (line 141)

`public function meeting($id): JsonResponse`

Guard/validation candidates:

```php
145: if (! $live_consultation) {
149: if ($live_consultation->status == 1 || $live_consultation->status == 2) {
```

## hms/app/Http/Controllers/API/Doctor/DoctorOperationAPIController.php

Models: `OperationReport` → `operation_reports`, `PatientCase` → `patient_cases`

### index (line 11)

`public function index(): \Illuminate\Http\JsonResponse`

### show (line 24)

`public function show($caseId): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
28: if (! $case_detail) {
```

### delete (line 35)

`public function delete($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
39: if (! $operation_report) {
```

Write/transaction candidates:

```php
43: $operation_report->delete();
```

## hms/app/Http/Controllers/API/Doctor/DoctorPatientAdmissionAPIController.php

Models: `PatientAdmission` → `patient_admissions`

### index (line 10)

`public function index(): \Illuminate\Http\JsonResponse`

### show (line 22)

`public function show($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
26: if (! $patient_admissions) {
```

### delete (line 33)

`public function delete($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
37: if (! $patient_admissions) {
```

Write/transaction candidates:

```php
41: $patient_admissions->delete();
```

## hms/app/Http/Controllers/API/DoctorAPIController.php

Models: `Doctor` → `doctors`, `Schedule` → `schedules`, `ScheduleDay` → `schedule_days`, `User` → `users`

### __construct (line 19)

`public function __construct(ScheduleRepository $scheduleRepo)`

### index (line 24)

`public function index(): JsonResponse`

### users (line 35)

`public function users(): JsonResponse`

### show (line 47)

`public function show($id): JsonResponse`

### doctorScheduleList (line 59)

`public function doctorScheduleList(): JsonResponse`

Guard/validation candidates:

```php
63: if (empty($schedules)) {
```

### doctorScheduleUpdate (line 89)

`public function doctorScheduleUpdate($id, Request $request): JsonResponse`

Guard/validation candidates:

```php
97: if ($schedule) {
```

Write/transaction candidates:

```php
95: $schedule = $this->scheduleRepository->update($input, $id);
```

## hms/app/Http/Controllers/API/DocumentAPIController.php

Models: `Document` → `documents`, `DocumentType` → `document_types`, `Patient` → `patients`, `PatientAdmission` → `patient_admissions`, `PatientCase` → `patient_cases`

### __construct (line 20)

`public function __construct(DocumentRepository $documentRepo)`

### index (line 25)

`public function index(): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
27: if (! getLoggedinPatient()) {
```

### getDocumentTypes (line 43)

`public function getDocumentTypes(): \Illuminate\Http\JsonResponse`

### create (line 50)

`public function create(Request $request): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
53: $request->validate([
```

Write/transaction candidates:

```php
57: $this->documentRepository->store($input);
```

### update (line 62)

`public function update(Request $request, $id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
65: $request->validate([
69: if (getLoggedinPatient()) {
72: if (! $documents) {
80: if (! $documents) {
```

### destroy (line 87)

`public function destroy($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
89: if (getLoggedinPatient()) {
95: if (! $documents) {
```

### downloadDocs (line 103)

`public function downloadDocs($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
106: if (! $document) {
```

### getPatientList (line 116)

`public function getPatientList(): \Illuminate\Http\JsonResponse`

### edit (line 140)

`public function edit($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
142: if (getLoggedinPatient()) {
148: if (! $document) {
```

## hms/app/Http/Controllers/API/InvoiceAPIController.php

Models: `Invoice` → `invoices`

### index (line 10)

`public function index(): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
12: if (getLoggedinPatient()) {
```

### show (line 23)

`public function show($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
27: if (! $invoice) {
```

## hms/app/Http/Controllers/API/LiveConsultationAPIController.php

Models: `LiveConsultation` → `live_consultations`

### __construct (line 23)

`public function __construct( LiveConsultationRepository $liveConsultationRepository, PatientCaseRepository $patientCaseRepository )`

### index (line 31)

`public function index(): JsonResponse`

### show (line 45)

`public function show($id): JsonResponse`

Guard/validation candidates:

```php
51: if (! $liveConsultation) {
```

### meeting (line 59)

`public function meeting($id): JsonResponse`

Guard/validation candidates:

```php
63: if (! $live_consultation) {
67: if ($live_consultation->status == 1 || $live_consultation->status == 2) {
```

### filter (line 75)

`public function filter(Request $request): JsonResponse`

Guard/validation candidates:

```php
86: if(!$data){
```

## hms/app/Http/Controllers/API/NoticeboardAPIController.php

Models: `NoticeBoard` → `notice_boards`

### index (line 10)

`public function index(): \Illuminate\Http\JsonResponse`

### show (line 22)

`public function show($id): \Illuminate\Http\JsonResponse`

## hms/app/Http/Controllers/API/PatientAdmissionAPIController.php

Models: `PatientAdmission` → `patient_admissions`

### index (line 10)

`public function index(): \Illuminate\Http\JsonResponse`

### show (line 22)

`public function show($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
27: if (! $admission) {
```

## hms/app/Http/Controllers/API/PatientCaseAPIController.php

Models: `PatientCase` → `patient_cases`

### index (line 11)

`public function index(): JsonResponse`

### show (line 23)

`public function show($id): JsonResponse`

Guard/validation candidates:

```php
27: if (! $patientCase) {
```

## hms/app/Http/Controllers/API/PayrollAPIController.php

Models: `Doctor` → `doctors`, `EmployeePayroll` → `employee_payrolls`

### index (line 12)

`public function index(): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
16: if ($type == \App\Models\Doctor::class) {
```

### show (line 30)

`public function show($id): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
34: if (! $payroll) {
```

## hms/app/Http/Controllers/API/PrescriptionAPIController.php

Models: `Doctor` → `doctors`, `Prescription` → `prescriptions`

### __construct (line 18)

`public function __construct(PrescriptionRepository $prescriptionRepo)`

### index (line 23)

`public function index(): JsonResponse`

### show (line 36)

`public function show($id): JsonResponse`

Guard/validation candidates:

```php
41: if (! $prescription) {
```

### DoctorPrescriptionList (line 48)

`public function DoctorPrescriptionList(): JsonResponse`

### prescriptionShow (line 60)

`public function prescriptionShow($id): JsonResponse`

Guard/validation candidates:

```php
62: if (getLoggedinDoctor()) {
65: if (! $prescription) {
75: if (getLoggedinPatient()) {
79: if (! $prescription) {
```

### destroy (line 90)

`public function destroy($id): JsonResponse`

Guard/validation candidates:

```php
93: if (empty($prescription)) {
96: if ($prescription->doctor_id == getLoggedInUser()->owner_id) {
```

Write/transaction candidates:

```php
97: $prescription->delete();
```

## hms/app/Http/Controllers/API/RegistrationController.php

Models: `Department` → `departments`, `Patient` → `patients`, `User` → `users`

### register (line 14)

`public function register(Request $request): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
16: $request->validate([
```

Write/transaction candidates:

```php
32: $user = User::create($data);
34: $patient = Patient::create(['user_id' => $user->id]);
36: $user->update(['owner_id' => $patient->id, 'owner_type' => Patient::class]);
```

## hms/app/Http/Controllers/API/UserAPIController.php

Models: `User` → `users`

### __construct (line 23)

`public function __construct(UserRepository $userRepository)`

### editProfile (line 28)

`public function editProfile(): JsonResponse`

### updateProfile (line 36)

`public function updateProfile(UpdateUserProfileRequest $request): JsonResponse`

### changePassword (line 45)

`public function changePassword(ChangePasswordRequest $request): JsonResponse`

### getProfile (line 57)

`public function getProfile(): JsonResponse`

## hms/app/Http/Controllers/API/VaccinatedPatientAPIController.php

Models: `VaccinatedPatients` → `vaccinated_patients`

### index (line 10)

`public function index(): \Illuminate\Http\JsonResponse`

Guard/validation candidates:

```php
12: if (getLoggedinPatient()) {
```

## hms/app/Http/Controllers/AccountController.php

Models: `Account` → `accounts`, `Payment` → `payments`

### __construct (line 17)

`public function __construct(AccountRepository $accountRepo)`

### index (line 22)

`public function index(Request $request)`

Views: `accounts.index`

### store (line 30)

`public function store(CreateAccountRequest $request)`

Write/transaction candidates:

```php
33: $this->accountRepository->create($input);
```

### show (line 38)

`public function show(Account $account)`

Views: `accounts.show`

### edit (line 45)

`public function edit(Account $account)`

### update (line 50)

`public function update(Account $account, UpdateAccountRequest $request)`

Write/transaction candidates:

```php
52: $this->accountRepository->update($request->all(), $account->id);
```

### destroy (line 57)

`public function destroy(Account $account)`

Guard/validation candidates:

```php
65: if ($result) {
```

Write/transaction candidates:

```php
69: $this->accountRepository->delete($account->id);
```

### activeDeactiveAccount (line 74)

`public function activeDeactiveAccount(Account $account)`

Write/transaction candidates:

```php
77: $account->save();
```

## hms/app/Http/Controllers/AccountantController.php

Models: `Accountant` → `accountants`, `EmployeePayroll` → `employee_payrolls`

### __construct (line 17)

`public function __construct(AccountantRepository $accountantRepo)`

### index (line 22)

`public function index()`

Views: `accountants.index`

### create (line 29)

`public function create()`

Views: `accountants.create`

### store (line 36)

`public function store(CreateAccountantRequest $request)`

Write/transaction candidates:

```php
41: $accountant = $this->accountantRepository->store($input);
```

### show (line 48)

`public function show(Accountant $accountant)`

Views: `accountants.show`

### edit (line 55)

`public function edit(Accountant $accountant)`

Views: `accountants.edit`

### update (line 63)

`public function update(Accountant $accountant, UpdateAccountantRequest $request)`

Write/transaction candidates:

```php
68: $accountant = $this->accountantRepository->update($accountant, $input);
```

### destroy (line 75)

`public function destroy(Accountant $accountant)`

Guard/validation candidates:

```php
79: if ($empPayRollResult) {
```

Write/transaction candidates:

```php
83: $accountant->user()->delete();
84: $accountant->address()->delete();
85: $accountant->delete();
```

### activeDeactiveStatus (line 90)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
94: $accountant->user()->update(['status' => $status]);
```

## hms/app/Http/Controllers/AddCustomFieldsController.php

Models: `AddCustomFields` → `add_custom_fields`

### index (line 14)

`public function index()`

Views: `add_custom_fields.index`

### store (line 23)

`public function store(CreateAddCustomFieldRequest $request)`

Write/transaction candidates:

```php
27: AddCustomFields::create($input);
```

### edit (line 36)

`public function edit($id)`

### update (line 46)

`public function update(CreateAddCustomFieldRequest $request, $id)`

Write/transaction candidates:

```php
52: $customField->update($input);
```

### destroy (line 60)

`public function destroy($id)`

Write/transaction candidates:

```php
62: AddCustomFields::where('id', $id)->delete();
```

## hms/app/Http/Controllers/AddOnController.php

Models: `AddOn` → `add_ons`

### index (line 13)

`public function index()`

Views: `add-on.index`

### extractZip (line 18)

`public function extractZip(Request $request)`

Guard/validation candidates:

```php
20: $request->validate([
29: if (is_dir($moduleFolder)) {
39: if ($zip->open($file) === TRUE) {
47: if (!in_array($isExistFile, $fileNames)) {
54: if (!empty($checkFiles)) {
58: if ($zip->open($file) === TRUE) {
```

### update (line 83)

`public function update($id)`

Guard/validation candidates:

```php
87: if (!$addOnModule) {
```

Write/transaction candidates:

```php
93: $addOnModule->save();
```

### delete (line 98)

`public function delete($id)`

Guard/validation candidates:

```php
102: if (!$addOnModule) {
108: if (is_dir($modulePath)) {
113: if (file_exists($modulesStatusesPath)) {
116: if (isset($content[$moduleName])) {
```

Write/transaction candidates:

```php
122: $addOnModule->delete();
```

### deleteDirectory (line 126)

`private function deleteDirectory($directory)`

Guard/validation candidates:

```php
129: if ($file === '.' || $file === '..') {
135: if (is_dir($filePath)) {
```

## hms/app/Http/Controllers/AdvancedPaymentController.php

Models: `AdvancedPayment` → `advanced_payments`

### __construct (line 17)

`public function __construct(AdvancedPaymentRepository $advancedPaymentRepo)`

### index (line 22)

`public function index()`

Views: `advanced_payments.index`

### store (line 29)

`public function store(CreateAdvancedPaymentRequest $request)`

Write/transaction candidates:

```php
34: $this->advancedPaymentRepository->create($input);
```

### show (line 41)

`public function show(AdvancedPayment $advancedPayment)`

Guard/validation candidates:

```php
45: if (empty($advancedPayment)) {
```

Views: `advanced_payments.show`

### edit (line 55)

`public function edit(AdvancedPayment $advancedPayment)`

### update (line 60)

`public function update(AdvancedPayment $advancedPayment, UpdateAdvancedPaymentRequest $request)`

Write/transaction candidates:

```php
65: $this->advancedPaymentRepository->update($input, $advancedPayment->id);
```

### destroy (line 71)

`public function destroy(AdvancedPayment $advancedPayment)`

Write/transaction candidates:

```php
73: $advancedPayment->delete();
```

## hms/app/Http/Controllers/AmbulanceCallController.php

Models: `Ambulance` → `ambulances`, `AmbulanceCall` → `ambulance_calls`

### __construct (line 28)

`public function __construct( AmbulanceCallRepository $ambulanceCallRepo, AmbulanceRepository $ambulanceRepo, PatientRepository $patientRepo )`

### index (line 38)

`public function index()`

Views: `ambulance_calls.index`

### create (line 43)

`public function create()`

Views: `ambulance_calls.create`

### store (line 51)

`public function store(CreateAmbulanceCallRequest $request)`

Guard/validation candidates:

```php
55: if ($request->has('amount')) {
```

Write/transaction candidates:

```php
59: $this->ambulanceCallRepository->create($input);
61: Ambulance::where('id', $input['ambulance_id'])->update(['is_available' => false]);
```

### show (line 68)

`public function show(AmbulanceCall $ambulanceCall)`

Views: `ambulance_calls.show`

### edit (line 73)

`public function edit(AmbulanceCall $ambulanceCall)`

Views: `ambulance_calls.edit`

### update (line 83)

`public function update(AmbulanceCall $ambulanceCall, UpdateAmbulanceCallRequest $request)`

Write/transaction candidates:

```php
87: $ambulanceCall = $this->ambulanceCallRepository->update($input, $ambulanceCall);
```

### destroy (line 94)

`public function destroy(AmbulanceCall $ambulanceCall)`

Write/transaction candidates:

```php
96: $this->ambulanceCallRepository->delete($ambulanceCall->id);
```

### getDriverName (line 101)

`public function getDriverName(Request $request)`

Guard/validation candidates:

```php
103: if (empty($request->get('id'))) {
```

### ambulanceCallExport (line 112)

`public function ambulanceCallExport()`

Guard/validation candidates:

```php
116: if (!$ambulanceCall) {
```

## hms/app/Http/Controllers/AmbulanceController.php

Models: `Ambulance` → `ambulances`, `AmbulanceCall` → `ambulance_calls`

### __construct (line 19)

`public function __construct(AmbulanceRepository $ambulanceRepo)`

### index (line 24)

`public function index()`

Views: `ambulances.index`

### create (line 31)

`public function create()`

Views: `ambulances.create`

### store (line 38)

`public function store(CreateAmbulanceRequest $request)`

Write/transaction candidates:

```php
44: $this->ambulanceRepository->create($input);
```

### show (line 52)

`public function show(Ambulance $ambulance)`

Views: `ambulances.show`

### edit (line 59)

`public function edit(Ambulance $ambulance)`

Views: `ambulances.edit`

### update (line 66)

`public function update(Ambulance $ambulance, UpdateAmbulanceRequest $request)`

Write/transaction candidates:

```php
72: $ambulance = $this->ambulanceRepository->update($input, $ambulance->id);
```

### destroy (line 79)

`public function destroy(Ambulance $ambulance)`

Guard/validation candidates:

```php
84: if ($result) {
```

Write/transaction candidates:

```php
88: $ambulance->delete($ambulance->id);
```

### isAvailableAmbulance (line 93)

`public function isAvailableAmbulance($id)`

Write/transaction candidates:

```php
97: $ambulance->update(['is_available' => $ambulance->is_available]);
```

### ambulanceExport (line 102)

`public function ambulanceExport()`

Guard/validation candidates:

```php
105: if (!$ambulances) {
```

## hms/app/Http/Controllers/AppBaseController.php

Models:

### sendResponse (line 9)

`public function sendResponse($result, $message)`

### sendError (line 14)

`public function sendError($error, $code = 404)`

### sendSuccess (line 19)

`public function sendSuccess($message)`

## hms/app/Http/Controllers/AppointmentCalendarController.php

Models: `Appointment` → `appointments`

### __construct (line 18)

`public function __construct( AppointmentCalendarRepository $appointmentCalendarRepo, AppointmentRepository $appointmentRepository )`

### index (line 26)

`public function index()`

Views: `appointment_calendars.index`

### calendarList (line 35)

`public function calendarList()`

### getAppointmentDetails (line 42)

`public function getAppointmentDetails(Appointment $appointment)`

## hms/app/Http/Controllers/AppointmentController.php

Models: `AddCustomFields` → `add_custom_fields`, `Appointment` → `appointments`

### __construct (line 24)

`public function __construct(AppointmentRepository $appointmentRepo, AppointmentTransactionRepository $appointmentTransactionRepo)`

### index (line 30)

`public function index()`

Views: `appointments.index`

### create (line 37)

`public function create()`

Views: `appointments.create`

### setFlutterWaveCredential (line 47)

`public function setFlutterWaveCredential()`

Guard/validation candidates:

```php
52: if (!$flutterwavePublicKey && !$flutterwaveSecretKey) {
```

### store (line 62)

`public function store(CreateAppointmentRequest $request)`

Guard/validation candidates:

```php
70: if ($request->user()->hasRole('Patient')) {
77: if (strpos($key, 'field') === 0) {
83: if ($input['payment_mode'] != 8 && $input['payment_mode'] != 7 && $input['payment_mode'] != 9) {
88: if ($input['payment_mode'] == 3 || $input['payment_mode'] == 4 || $input['payment_mode'] == 5) {
92: if ($input['payment_mode'] == 3) {
117: if (!in_array(strtoupper(getCurrentCurrency()), getFlutterWaveSupportedCurrencies())) {
128: if (strtoupper(getCurrentCurrency()) != 'INR') {
```

Write/transaction candidates:

```php
84: $data = $this->appointmentRepository->create($input);
89: $data->update(['payment_type' => 1]);
139: $data = $this->appointmentTransactionRepository->store($data);
```

### show (line 147)

`public function show(Appointment $appointment)`

Views: `appointments.show`

### edit (line 152)

`public function edit(Appointment $appointment)`

Views: `appointments.edit`

### update (line 164)

`public function update(Appointment $appointment, UpdateAppointmentRequest $request)`

Guard/validation candidates:

```php
171: if ($request->user()->hasRole('Patient')) {
177: if (strpos($key, 'field') === 0) {
```

Write/transaction candidates:

```php
183: $appointment = $this->appointmentRepository->update($input, $appointment->id);
```

### destroy (line 188)

`public function destroy(Appointment $appointment)`

Guard/validation candidates:

```php
192: if (getLoggedinPatient() && $appointment->patient_id != getLoggedInUser()->owner_id) {
```

Write/transaction candidates:

```php
195: $this->appointmentRepository->delete($appointment->id);
```

### getDoctors (line 201)

`public function getDoctors(Request $request)`

### getDoctorsCharge (line 208)

`public function getDoctorsCharge(Request $request)`

### getBookingSlot (line 215)

`public function getBookingSlot(Request $request)`

### appointmentExport (line 223)

`public function appointmentExport()`

Guard/validation candidates:

```php
226: if (getLoggedInUser()->hasRole('Doctor')) {
230: if (getLoggedInUser()->hasRole('Patient')) {
234: if ($appointments->count() == 0) {
```

### status (line 242)

`public function status(Appointment $appointment)`

Guard/validation candidates:

```php
244: if (getLoggedinDoctor() && $appointment->doctor_id != getLoggedInUser()->owner_id) {
```

Write/transaction candidates:

```php
248: $appointment->update(['is_completed' => $isCompleted]);
```

### cancelAppointment (line 254)

`public function cancelAppointment(Appointment $appointment)`

Guard/validation candidates:

```php
256: if ((getLoggedinPatient() && $appointment->patient_id != getLoggedInUser()->owner_id) || (getLoggedinDoctor() && $appointment->doctor_id != getLoggedInUser()->owner_id)) {
```

Write/transaction candidates:

```php
259: $appointment->update(['is_completed' => Appointment::STATUS_CANCELLED]);
```

## hms/app/Http/Controllers/AppointmentTransactionController.php

Models: `Appointment` → `appointments`

### __construct (line 19)

`public function __construct(AppointmentTransactionRepository $appointmentTransactionRepo)`

### index (line 29)

`public function index()`

Views: `appointment_transaction.index`

### appointmentStripePaymentSuccess (line 33)

`public function appointmentStripePaymentSuccess(Request $request)`

### webAppointmentStripePaymentSuccess (line 42)

`public function webAppointmentStripePaymentSuccess(Request $request)`

### appointmentRazorpayPayment (line 51)

`public function appointmentRazorpayPayment(Request $request)`

### appointmentRazorpayPaymentSuccess (line 58)

`public function appointmentRazorpayPaymentSuccess(Request $request)`

### webAppointmentRazorpayPayment (line 67)

`public function webAppointmentRazorpayPayment(Request $request)`

### WebAppointmentRazorpayPaymentSuccess (line 74)

`public function WebAppointmentRazorpayPaymentSuccess(Request $request)`

### paypalOnBoard (line 83)

`public function paypalOnBoard(Request $request)`

Guard/validation candidates:

```php
85: if (! in_array(strtoupper(getCurrentCurrency()), getPayPalSupportedCurrencies())) {
86: if($request->get('appointment_id')){
```

Write/transaction candidates:

```php
87: Appointment::find($request->get('appointment_id'))->delete();
```

### paypalSuccess (line 131)

`public function paypalSuccess(Request $request): RedirectResponse`

### paypalFailed (line 161)

`public function paypalFailed(Request $request): RedirectResponse`

Guard/validation candidates:

```php
164: if($appointmentId){
```

Write/transaction candidates:

```php
165: Appointment::find($appointmentId)->delete();
```

### webAppointmentPaypalOnBoard (line 172)

`public function webAppointmentPaypalOnBoard(Request $request)`

Guard/validation candidates:

```php
174: if (! in_array(strtoupper(getCurrentCurrency()), getPayPalSupportedCurrencies())) {
```

### webAppointmentPaypalSuccess (line 217)

`public function webAppointmentPaypalSuccess(Request $request): RedirectResponse`

### webAppointmentPaypalFailed (line 247)

`public function webAppointmentPaypalFailed(Request $request): RedirectResponse`

Guard/validation candidates:

```php
250: if($appointmentId){
```

Write/transaction candidates:

```php
251: Appointment::find($appointmentId)->delete();
```

### appointmentRazorPayPaymentFailed (line 258)

`public function appointmentRazorPayPaymentFailed(Request $request)`

Write/transaction candidates:

```php
262: $appointment->delete();
```

### appointmentStripeFailed (line 267)

`public function appointmentStripeFailed(Request $request)`

Guard/validation candidates:

```php
270: if($appointmentId){
```

Write/transaction candidates:

```php
271: Appointment::find($appointmentId)->delete();
```

### webAppointmentStripeFailed (line 278)

`public function webAppointmentStripeFailed(Request $request)`

Guard/validation candidates:

```php
281: if($appointmentId){
```

Write/transaction candidates:

```php
282: Appointment::find($appointmentId)->delete();
```

### webAppointmentRazorPayPaymentFailed (line 289)

`public function webAppointmentRazorPayPaymentFailed(Request $request)`

Write/transaction candidates:

```php
293: $appointment->delete();
```

### flutterWaveSuccess (line 297)

`public function flutterWaveSuccess(Request $request)`

Guard/validation candidates:

```php
302: if(!$flutterwavePublicKey && !$flutterwaveSecretKey){
311: if($request['status'] == 'cancelled'){
312: if(isset($request->input['web_appointment']) && $request->input['web_appointment'] == true){
326: if(isset($request->input['web_appointment']) && $request->input['web_appointment'] == true){
```

### phonePePaymentSuccess (line 338)

`public function phonePePaymentSuccess(Request $request)`

Guard/validation candidates:

```php
342: if(isset($request->input['web_appointment']) && $request->input['web_appointment'] == true){
```

### paystackPayment (line 354)

`public function paystackPayment(Request $request)`

Guard/validation candidates:

```php
356: if(!in_array(strtoupper(getCurrentCurrency()),getPayStackSupportedCurrencies())){
358: if(isset($request->data['web_appointment']) && $request->data['web_appointment'] == true){
390: if(isset($request->data['web_appointment']) && $request->data['web_appointment'] == true){
```

## hms/app/Http/Controllers/Auth/ConfirmPasswordController.php

Models:

### __construct (line 15)

`public function __construct()`

## hms/app/Http/Controllers/Auth/ForgotPasswordController.php

Models: `User` → `users`

### showLinkRequestForm (line 19)

`public function showLinkRequestForm()`

Views: `auth.passwords.email`

### sendResetLinkEmail (line 27)

`public function sendResetLinkEmail(Request $request)`

Guard/validation candidates:

```php
31: $request->validate([
37: if (! $user) {
```

## hms/app/Http/Controllers/Auth/LoginController.php

Models:

### __construct (line 22)

`public function __construct()`

### sendLoginResponse (line 27)

`protected function sendLoginResponse(Request $request)`

Guard/validation candidates:

```php
35: if ($user->email_verified_at == null) {
41: if ($user->enable_two_factor_authentication) {
47: if ($request->user()->hasRole('Admin')) {
50: if ($request->user()->hasRole(['Receptionist'])) {
52: } elseif ($request->user()->hasRole(['Doctor', 'Case Manager', 'Lab Technician', 'Pharmacist'])) {
54: } elseif ($request->user()->hasRole(['Patient'])) {
56: } elseif ($request->user()->hasRole(['Nurse'])) {
58: } elseif ($request->user()->hasRole(['Accountant'])) {
65: if (! isset($request->remember)) {
```

### sendFailedLoginResponse (line 80)

`protected function sendFailedLoginResponse(Request $request)`

### logout (line 88)

`public function logout(Request $request)`

Guard/validation candidates:

```php
92: $request->session()->invalidate();
96: if ($response = $this->loggedOut($request)) {
```

## hms/app/Http/Controllers/Auth/RegisterController.php

Models: `Department` → `departments`, `Patient` → `patients`, `User` → `users`

### __construct (line 22)

`public function __construct()`

### validator (line 27)

`protected function validator($data)`

### create (line 40)

`protected function create($data)`

Write/transaction candidates:

```php
48: $user = User::create($data);
49: $patient = Patient::create(['user_id' => $user->id]);
51: $user->update(['owner_id' => $patient->id, 'owner_type' => Patient::class]);
```

## hms/app/Http/Controllers/Auth/ResetPasswordController.php

Models:

### rules (line 15)

`protected function rules()`

## hms/app/Http/Controllers/Auth/TwofactorAuthenticationController.php

Models: `User` → `users`

### index (line 19)

`public function index()`

Guard/validation candidates:

```php
23: if ($user->google2fa_secret) {
```

Views: `two_auth.disable2fa`, `two_auth.enable2fa`

### generateTwoFactorSecret (line 30)

`public function generateTwoFactorSecret()`

### enable2FA (line 56)

`public function enable2FA(Request $request)`

Guard/validation candidates:

```php
58: $request->validate([
65: if (!$secret) {
74: if (!$valid) {
```

Write/transaction candidates:

```php
89: $user->save();
```

### regenerateRecoveryCodes (line 96)

`public function regenerateRecoveryCodes()`

Write/transaction candidates:

```php
105: $user->save();
```

### disable (line 111)

`public function disable(Request $request)`

Write/transaction candidates:

```php
118: $user->save();
```

### showVerifyForm (line 124)

`public function showVerifyForm()`

Guard/validation candidates:

```php
126: if (!session('2fa:user:id')) {
```

Views: `two_auth.verify2fa`

### verify (line 132)

`public function verify(Request $request)`

Guard/validation candidates:

```php
134: $request->validate([
140: if (!$user) {
148: if (ctype_digit($input) && strlen($input) === 6) {
149: if ($google2fa->verifyKey($user->google2fa_secret, $input)) {
167: if ($matched) {
```

Write/transaction candidates:

```php
173: $user->save();
```

### redirectUser (line 186)

`private function redirectUser($user)`

Guard/validation candidates:

```php
188: if ($user->hasRole('Admin')) {
190: } elseif ($user->hasRole(['Receptionist'])) {
192: } elseif ($user->hasRole(['Doctor', 'Case Manager', 'Lab Technician', 'Pharmacist'])) {
194: } elseif ($user->hasRole(['Patient'])) {
196: } elseif ($user->hasRole(['Nurse'])) {
198: } elseif ($user->hasRole(['Accountant'])) {
```

## hms/app/Http/Controllers/Auth/VerificationController.php

Models:

### redirectPath (line 12)

`public function redirectPath()`

Guard/validation candidates:

```php
14: if (method_exists($this, 'redirectTo')) {
```

### __construct (line 21)

`public function __construct()`

## hms/app/Http/Controllers/BedAssignController.php

Models: `BedAssign` → `bed_assigns`, `BedType` → `bed_types`, `IpdPatientDepartment` → `ipd_patient_departments`, `PatientAdmission` → `patient_admissions`

### __construct (line 23)

`public function __construct(BedAssignRepository $bedAssignRepo)`

### index (line 28)

`public function index()`

Views: `bed_assigns.index`

### create (line 35)

`public function create(Request $request)`

Views: `bed_assigns.create`

### store (line 44)

`public function store(CreateBedAssignRequest $request)`

Guard/validation candidates:

```php
53: if (! empty($input['birth_date']) && $assign_date < $input['birth_date']) {
```

Write/transaction candidates:

```php
59: $this->bedAssignRepository->store($input);
```

### show (line 66)

`public function show(BedAssign $bedAssign)`

Views: `bed_assigns.show`

### edit (line 71)

`public function edit(BedAssign $bedAssign)`

Views: `bed_assigns.edit`

### update (line 80)

`public function update(BedAssign $bedAssign, UpdateBedAssignRequest $request)`

Guard/validation candidates:

```php
89: if (! empty($birthDate) && $assign_date < $input['birth_date']) {
```

Write/transaction candidates:

```php
95: $bedAssign = $this->bedAssignRepository->update($input, $bedAssign);
```

### destroy (line 101)

`public function destroy(BedAssign $bedAssign)`

Write/transaction candidates:

```php
103: $bedAssign->bed->update(['is_available' => 1]);
104: $this->bedAssignRepository->delete($bedAssign->id);
```

### activeDeactiveStatus (line 109)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
113: $bedAssign->update(['status' => $status]);
114: $bedAssign->bed->update(['is_available' => 1]);
```

### bedStatus (line 119)

`public function bedStatus()`

Views: `bed_status.index`

### bedAssignExport (line 127)

`public function bedAssignExport()`

Guard/validation candidates:

```php
130: if (!$bedAssign) {
```

### getIpdPatientsList (line 137)

`public function getIpdPatientsList(Request $request)`

## hms/app/Http/Controllers/BedController.php

Models: `Bed` → `beds`, `BedAssign` → `bed_assigns`, `IpdPatientDepartment` → `ipd_patient_departments`

### __construct (line 21)

`public function __construct(BedRepository $bedRepo)`

### index (line 26)

`public function index()`

Views: `beds.index`

### store (line 35)

`public function store(CreateBedRequest $request)`

Write/transaction candidates:

```php
40: $this->bedRepository->store($input);
```

### show (line 45)

`public function show(Bed $bed)`

Views: `beds.show`

### edit (line 53)

`public function edit(Bed $bed)`

### update (line 58)

`public function update(Bed $bed, UpdateBedRequest $request)`

Write/transaction candidates:

```php
63: $bed = $this->bedRepository->update($input, $bed->id);
```

### destroy (line 68)

`public function destroy(Bed $bed)`

Guard/validation candidates:

```php
77: if ($result) {
```

Write/transaction candidates:

```php
81: $this->bedRepository->delete($bed->id);
```

### activeDeActiveStatus (line 86)

`public function activeDeActiveStatus($id)`

Write/transaction candidates:

```php
90: $bed->update(['status' => $bed->status]);
```

### createBulkBeds (line 95)

`public function createBulkBeds()`

Views: `beds.create_bulk_beds`

### storeBulkBeds (line 103)

`public function storeBulkBeds(CreateBulkBedRequest $request)`

### bedExport (line 111)

`public function bedExport()`

Guard/validation candidates:

```php
114: if (!$beds) {
```

## hms/app/Http/Controllers/BedTypeController.php

Models: `Bed` → `beds`, `BedType` → `bed_types`, `IpdPatientDepartment` → `ipd_patient_departments`

### __construct (line 17)

`public function __construct(BedTypeRepository $bedTypeRepo)`

### index (line 22)

`public function index()`

Views: `bed_types.index`

### store (line 27)

`public function store(CreateBedTypeRequest $request)`

Write/transaction candidates:

```php
31: $this->bedTypeRepository->create($input);
```

### show (line 36)

`public function show(BedType $bedType)`

Views: `bed_types.show`

### edit (line 43)

`public function edit(BedType $bedType)`

### update (line 48)

`public function update(BedType $bedType, UpdateBedTypeRequest $request)`

Write/transaction candidates:

```php
51: $bedType = $this->bedTypeRepository->update($input, $bedType->id);
```

### destroy (line 56)

`public function destroy(BedType $bedType)`

Guard/validation candidates:

```php
61: if ($bed || $ipdPatientDepartment) {
```

Write/transaction candidates:

```php
65: $this->bedTypeRepository->delete($bedType->id);
```

## hms/app/Http/Controllers/BillController.php

Models: `Bill` → `bills`, `Patient` → `patients`, `Setting` → `settings`

### __construct (line 22)

`public function __construct(BillRepository $billRepo)`

### index (line 27)

`public function index()`

Views: `bills.index`

### create (line 34)

`public function create()`

Views: `bills.create`

### store (line 41)

`public function store(CreateBillRequest $request)`

Guard/validation candidates:

```php
51: if (! empty($birthDate) && $billDate < $birthDate) {
```

Write/transaction candidates:

```php
44: DB::beginTransaction();
58: DB::commit();
60: DB::rollBack();
```

### show (line 68)

`public function show(Bill $bill)`

Guard/validation candidates:

```php
72: if ($bill->patientAdmission) {
```

Views: `bills.show`

### edit (line 81)

`public function edit(Bill $bill)`

Views: `bills.edit`

### update (line 91)

`public function update(Bill $bill, UpdateBillRequest $request)`

Guard/validation candidates:

```php
98: if (! empty($birthDate) && $billDate < $birthDate) {
```

### destroy (line 107)

`public function destroy(Bill $bill)`

Write/transaction candidates:

```php
109: $this->billRepository->delete($bill->id);
```

### getPatientAdmissionDetails (line 114)

`public function getPatientAdmissionDetails(Request $request)`

### convertToPdf (line 122)

`public function convertToPdf(Bill $bill)`

## hms/app/Http/Controllers/BirthReportController.php

Models: `BirthReport` → `birth_reports`, `DeathReport` → `death_reports`, `PatientCase` → `patient_cases`

### __construct (line 18)

`public function __construct(BirthReportRepository $birthReportRepo)`

### index (line 23)

`public function index()`

Views: `birth_reports.index`

### store (line 31)

`public function store(CreateBirthReportRequest $request)`

Guard/validation candidates:

```php
39: if (! empty($birthDate) && $selectBirthDate < $birthDate) {
45: if (! empty($isUserHasDead)) {
```

Write/transaction candidates:

```php
49: $this->birthReportRepository->store($input);
```

### show (line 54)

`public function show(BirthReport $birthReport)`

Views: `birth_reports.show`

### edit (line 64)

`public function edit(BirthReport $birthReport)`

Guard/validation candidates:

```php
66: if (getLoggedinDoctor() && checkRecordAccess($birthReport->doctor_id)) {
```

### update (line 73)

`public function update(BirthReport $birthReport, UpdateBirthReportRequest $request)`

Guard/validation candidates:

```php
79: if (! empty($birthDate) && $selectBirthDate < $birthDate) {
```

Write/transaction candidates:

```php
82: $birthReport = $this->birthReportRepository->update($request->all(), $birthReport);
```

### destroy (line 87)

`public function destroy(BirthReport $birthReport)`

Guard/validation candidates:

```php
89: if (getLoggedinDoctor() && checkRecordAccess($birthReport->doctor_id)) {
```

Write/transaction candidates:

```php
92: $this->birthReportRepository->delete($birthReport->id);
```

## hms/app/Http/Controllers/BloodBankController.php

Models: `BloodBank` → `blood_banks`, `BloodDonor` → `blood_donors`, `User` → `users`

### __construct (line 20)

`public function __construct(BloodBankRepository $bloodBankRepo)`

### index (line 25)

`public function index()`

Views: `blood_banks.index`

### store (line 30)

`public function store(CreateBloodBankRequest $request)`

Write/transaction candidates:

```php
33: $this->bloodBankRepository->create($input);
```

### edit (line 38)

`public function edit(BloodBank $bloodBank)`

### update (line 43)

`public function update(BloodBank $bloodBank, UpdateBloodBankRequest $request)`

Write/transaction candidates:

```php
46: $this->bloodBankRepository->update($input, $bloodBank->id);
```

### destroy (line 51)

`public function destroy(BloodBank $bloodBank)`

Guard/validation candidates:

```php
60: if ($result) {
```

Write/transaction candidates:

```php
64: $bloodBank->delete($bloodBank->id);
```

### bloodBankExport (line 69)

`public function bloodBankExport()`

Guard/validation candidates:

```php
72: if (!$bloodBank) {
```

## hms/app/Http/Controllers/BloodDonationController.php

Models: `BloodDonation` → `blood_donations`, `BloodDonor` → `blood_donors`

### __construct (line 19)

`public function __construct(BloodDonationRepository $bloodDonationRepository)`

### index (line 24)

`public function index()`

Views: `blood_donations.index`

### store (line 31)

`public function store(BloodDonationRequest $request)`

### edit (line 43)

`public function edit(BloodDonation $bloodDonation)`

### update (line 48)

`public function update(BloodDonationRequest $request, BloodDonation $bloodDonation)`

### destroy (line 60)

`public function destroy(BloodDonation $bloodDonation)`

Write/transaction candidates:

```php
63: $bloodDonation->delete($bloodDonation->id);
```

### bloodDonationExport (line 71)

`public function bloodDonationExport()`

Guard/validation candidates:

```php
74: if (!$bloodDonation) {
```

## hms/app/Http/Controllers/BloodDonorController.php

Models: `BloodDonation` → `blood_donations`, `BloodDonor` → `blood_donors`

### __construct (line 19)

`public function __construct(BloodDonorRepository $bloodDonorRepo)`

### index (line 24)

`public function index()`

Views: `blood_donors.index`

### store (line 31)

`public function store(CreateBloodDonorRequest $request)`

Write/transaction candidates:

```php
34: $this->bloodDonorRepository->create($input);
```

### edit (line 39)

`public function edit(BloodDonor $bloodDonor)`

### update (line 44)

`public function update(BloodDonor $bloodDonor, UpdateBloodDonorRequest $request)`

Write/transaction candidates:

```php
47: $this->bloodDonorRepository->update($input, $bloodDonor->id);
```

### destroy (line 52)

`public function destroy(BloodDonor $bloodDonor)`

Guard/validation candidates:

```php
57: if ($result) {
```

Write/transaction candidates:

```php
61: $bloodDonor->delete($bloodDonor->id);
```

### bloodDonorExport (line 66)

`public function bloodDonorExport()`

Guard/validation candidates:

```php
69: if (!$bloodDonors) {
```

## hms/app/Http/Controllers/BloodIssueController.php

Models: `BloodDonor` → `blood_donors`, `BloodIssue` → `blood_issues`

### __construct (line 24)

`public function __construct( BloodIssueRepository $bloodIssueRepository, PatientCaseRepository $patientCaseRepository )`

### index (line 32)

`public function index()`

Views: `blood_issues.index`

### store (line 41)

`public function store(BloodIssueRequest $request)`

Write/transaction candidates:

```php
46: $this->bloodIssueRepository->create($input);
```

### getBloodGroup (line 54)

`public function getBloodGroup(Request $request)`

### edit (line 65)

`public function edit(BloodIssue $bloodIssue)`

### update (line 70)

`public function update(BloodIssueRequest $request, BloodIssue $bloodIssue)`

Write/transaction candidates:

```php
75: $this->bloodIssueRepository->update($input, $bloodIssue->id);
```

### destroy (line 83)

`public function destroy(BloodIssue $bloodIssue)`

Write/transaction candidates:

```php
86: $bloodIssue->delete();
```

### export (line 94)

`public function export()`

Guard/validation candidates:

```php
97: if (!$bloodIssue) {
```

## hms/app/Http/Controllers/BrandController.php

Models: `Brand` → `brands`, `Medicine` → `medicines`

### __construct (line 19)

`public function __construct(BrandRepository $brandRepo)`

### index (line 24)

`public function index()`

Views: `brands.index`

### create (line 29)

`public function create()`

Views: `brands.create`

### store (line 34)

`public function store(CreateBrandRequest $request)`

Write/transaction candidates:

```php
38: $this->brandRepository->create($input);
```

### show (line 44)

`public function show(Brand $brand)`

Views: `brands.show`

### edit (line 51)

`public function edit(Brand $brand)`

Views: `brands.edit`

### update (line 56)

`public function update(Brand $brand, UpdateBrandRequest $request)`

Write/transaction candidates:

```php
60: $this->brandRepository->update($input, $brand->id);
```

### destroy (line 66)

`public function destroy(Brand $brand)`

Guard/validation candidates:

```php
74: if ($result) {
```

Write/transaction candidates:

```php
78: $brand->delete($brand->id);
```

### brandExport (line 83)

`public function brandExport()`

Guard/validation candidates:

```php
86: if (!$brands) {
```

## hms/app/Http/Controllers/CallLogController.php

Models: `CallLog` → `call_logs`

### __construct (line 20)

`public function __construct(CallLogRepository $callLogRepo)`

### index (line 25)

`public function index()`

Views: `call_logs.index`

### create (line 32)

`public function create()`

Views: `call_logs.create`

### store (line 37)

`public function store(CreateCallLogRequest $request)`

Write/transaction candidates:

```php
41: $this->CallLogRepository->create($input);
```

### edit (line 47)

`public function edit(CallLog $callLog)`

Views: `call_logs.edit`

### update (line 52)

`public function update(UpdateCallLogRequest $request, CallLog $callLog)`

Write/transaction candidates:

```php
56: $this->CallLogRepository->update($input, $callLog->id);
```

### destroy (line 62)

`public function destroy(CallLog $callLog)`

Write/transaction candidates:

```php
64: $callLog->delete();
```

### export (line 69)

`public function export()`

Guard/validation candidates:

```php
72: if (!$callLog) {
```

## hms/app/Http/Controllers/CaseHandlerController.php

Models: `CaseHandler` → `case_handlers`, `EmployeePayroll` → `employee_payrolls`

### __construct (line 19)

`public function __construct(CaseHandlerRepository $caseHandlerRepo)`

### index (line 24)

`public function index()`

Views: `case_handlers.index`

### create (line 31)

`public function create()`

Views: `case_handlers.create`

### store (line 38)

`public function store(CreateCaseHandlerRequest $request)`

Write/transaction candidates:

```php
42: $this->caseHandlerRepository->store($input);
```

### show (line 48)

`public function show(CaseHandler $caseHandler)`

Views: `case_handlers.show`

### edit (line 55)

`public function edit(CaseHandler $caseHandler)`

Views: `case_handlers.edit`

### update (line 63)

`public function update(CaseHandler $caseHandler, UpdateCaseHandlerRequest $request)`

Write/transaction candidates:

```php
67: $this->caseHandlerRepository->update($caseHandler, $input);
```

### destroy (line 73)

`public function destroy(CaseHandler $caseHandler)`

Guard/validation candidates:

```php
81: if ($result) {
```

Write/transaction candidates:

```php
85: $caseHandler->user()->delete();
86: $caseHandler->address()->delete();
87: $caseHandler->delete();
```

### activeDeactiveStatus (line 92)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
96: $caseHandler->user()->update(['status' => $status]);
```

### caseHandlerExport (line 101)

`public function caseHandlerExport()`

Guard/validation candidates:

```php
104: if (!$caseHandler) {
```

## hms/app/Http/Controllers/CategoryController.php

Models: `Category` → `categories`, `Medicine` → `medicines`

### __construct (line 16)

`public function __construct(CategoryRepository $categoryRepo)`

### index (line 21)

`public function index()`

Views: `categories.index`

### store (line 28)

`public function store(CreateCategoryRequest $request)`

Write/transaction candidates:

```php
32: $this->categoryRepository->create($input);
```

### show (line 37)

`public function show(Category $category)`

Views: `categories.show`

### edit (line 44)

`public function edit(Category $category)`

### update (line 49)

`public function update(Category $category, UpdateCategoryRequest $request)`

Write/transaction candidates:

```php
53: $this->categoryRepository->update($input, $category->id);
```

### destroy (line 58)

`public function destroy(Category $category)`

Guard/validation candidates:

```php
66: if ($result) {
```

Write/transaction candidates:

```php
70: $this->categoryRepository->delete($category->id);
```

### activeDeActiveCategory (line 75)

`public function activeDeActiveCategory($id)`

Write/transaction candidates:

```php
79: $category->save();
```

## hms/app/Http/Controllers/ChargeCategoryController.php

Models: `ChargeCategory` → `charge_categories`, `RadiologyTest` → `radiology_tests`

### __construct (line 16)

`public function __construct(ChargeCategoryRepository $chargeCategoryRepo)`

### index (line 21)

`public function index()`

Views: `charge_categories.index`

### store (line 29)

`public function store(CreateChargeCategoryRequest $request)`

Write/transaction candidates:

```php
33: $this->chargeCategoryRepository->create($input);
```

### show (line 38)

`public function show(ChargeCategory $chargeCategory)`

Views: `charge_categories.show`

### edit (line 45)

`public function edit(ChargeCategory $chargeCategory)`

### update (line 50)

`public function update(ChargeCategory $chargeCategory, UpdateChargeCategoryRequest $request)`

Write/transaction candidates:

```php
52: $chargeCategory = $this->chargeCategoryRepository->update($request->all(), $chargeCategory->id);
```

### destroy (line 57)

`public function destroy(ChargeCategory $chargeCategory)`

Guard/validation candidates:

```php
65: if ($result) {
```

Write/transaction candidates:

```php
69: $this->chargeCategoryRepository->delete($chargeCategory->id);
```

## hms/app/Http/Controllers/ChargeController.php

Models: `Charge` → `charges`, `ChargeCategory` → `charge_categories`

### __construct (line 20)

`public function __construct(ChargeRepository $chargeRepo)`

### index (line 25)

`public function index()`

Views: `charges.index`

### store (line 35)

`public function store(CreateChargeRequest $request)`

Write/transaction candidates:

```php
39: $charge = $this->chargeRepository->create($input);
```

### show (line 44)

`public function show(Charge $charge)`

Views: `charges.show`

### edit (line 52)

`public function edit(Charge $charge)`

### update (line 57)

`public function update(Charge $charge, UpdateChargeRequest $request)`

Write/transaction candidates:

```php
61: $charge = $this->chargeRepository->update($input, $charge->id);
```

### destroy (line 66)

`public function destroy(Charge $charge)`

Write/transaction candidates:

```php
68: $this->chargeRepository->delete($charge->id);
```

### getChargeCategory (line 73)

`public function getChargeCategory(Request $request)`

### chargeExport (line 82)

`public function chargeExport()`

Guard/validation candidates:

```php
85: if (!$charges) {
```

## hms/app/Http/Controllers/ComplaintController.php

Models: `Complaint` → `complaints`

### index (line 15)

`public function index()`

Views: `complaints.index`

### store (line 20)

`public function store(CreateComplaint $request)`

Write/transaction candidates:

```php
23: Complaint::create([
```

### edit (line 36)

`public function edit(Complaint $complaint)`

### update (line 41)

`public function update(UpdateComplaintRequest $request, Complaint $complaint)`

Write/transaction candidates:

```php
44: $complaint->update($request->only([
```

### destroy (line 54)

`public function destroy(Complaint $complaint)`

Write/transaction candidates:

```php
56: $complaint->delete();
```

### updateStatusResponse (line 63)

`public function updateStatusResponse(Request $request)`

Guard/validation candidates:

```php
65: $request->validate([
78: if (in_array($request->status, [
```

Write/transaction candidates:

```php
87: $complaint->update($updateData);
```

### responseEdit (line 94)

`public function responseEdit(Complaint $complaint)`

### show (line 99)

`public function show(Complaint $complaint)`

## hms/app/Http/Controllers/Controller.php

Models:

## hms/app/Http/Controllers/CurrencySettingController.php

Models: `CurrencySetting` → `currency_settings`, `Setting` → `settings`

### __construct (line 18)

`public function __construct(currency_settingRepository $currencySettingRepo)`

### index (line 23)

`public function index(Request $request)`

Views: `currency_settings.index`

### create (line 30)

`public function create()`

Views: `currency_settings.create`

### store (line 35)

`public function store(Createcurrency_settingRequest $request)`

Write/transaction candidates:

```php
39: $this->currencySettingRepository->create($input);
```

### show (line 44)

`public function show($id)`

Guard/validation candidates:

```php
48: if (empty($currencySetting)) {
```

Views: `currency_settings.show`

### edit (line 57)

`public function edit(CurrencySetting $currencySetting)`

### update (line 62)

`public function update(CurrencySetting $currencySetting, UpdateCurrencySettingRequest $request)`

Write/transaction candidates:

```php
66: $this->currencySettingRepository->update($input, $currencySetting->id);
```

### destroy (line 71)

`public function destroy(CurrencySetting $currencySetting)`

Guard/validation candidates:

```php
75: if ($currentCurrency && $currentCurrency->value === strtolower($currencyCode)) {
```

Write/transaction candidates:

```php
78: $this->currencySettingRepository->delete($currencySetting->id);
```

## hms/app/Http/Controllers/DeathReportController.php

Models: `BirthReport` → `birth_reports`, `DeathReport` → `death_reports`, `PatientCase` → `patient_cases`

### __construct (line 18)

`public function __construct(DeathReportRepository $deathReportRepo)`

### index (line 23)

`public function index()`

Views: `death_reports.index`

### store (line 31)

`public function store(CreateDeathReportRequest $request)`

Guard/validation candidates:

```php
37: if (empty($patientId)) {
43: if (! empty($birthDate) && $deathDate < $birthDate) {
```

Write/transaction candidates:

```php
47: $deathReport = $this->deathReportRepository->store($input);
```

### show (line 53)

`public function show(DeathReport $deathReport)`

Views: `death_reports.show`

### edit (line 63)

`public function edit(DeathReport $deathReport)`

Guard/validation candidates:

```php
65: if (checkRecordAccess($deathReport->doctor_id)) {
```

### update (line 72)

`public function update(DeathReport $deathReport, UpdateDeathReportRequest $request)`

Guard/validation candidates:

```php
79: if (! empty($birthDate) && $deathDate < $birthDate) {
```

Write/transaction candidates:

```php
83: $deathReport = $this->deathReportRepository->update($request->all(), $deathReport);
```

### destroy (line 88)

`public function destroy(DeathReport $deathReport)`

Guard/validation candidates:

```php
90: if (checkRecordAccess($deathReport->doctor_id)) {
```

Write/transaction candidates:

```php
93: $this->deathReportRepository->delete($deathReport->id);
```

## hms/app/Http/Controllers/DepartmentController.php

Models: `Department` → `departments`, `User` → `users`

### __construct (line 18)

`public function __construct(DepartmentRepository $departmentRepo)`

### index (line 23)

`public function index()`

Views: `departments.index`

### store (line 30)

`public function store(CreateDepartmentRequest $request)`

Write/transaction candidates:

```php
34: $this->departmentRepository->create($input);
```

### edit (line 39)

`public function edit(Department $department)`

### update (line 44)

`public function update(Department $department, UpdateDepartmentRequest $request)`

Write/transaction candidates:

```php
46: $this->departmentRepository->update($request->all(), $department->id);
```

### destroy (line 51)

`public function destroy(Department $department)`

Write/transaction candidates:

```php
53: $this->departmentRepository->delete($department->id);
```

### activeDeactiveDepartment (line 58)

`public function activeDeactiveDepartment(Department $department)`

Write/transaction candidates:

```php
61: $department->save();
```

### getUsersList (line 66)

`public function getUsersList(Request $request)`

Guard/validation candidates:

```php
68: if (empty($request->get('id'))) {
```

## hms/app/Http/Controllers/DiagnosisCategoryController.php

Models: `DiagnosisCategory` → `diagnosis_categories`, `PatientDiagnosisTest` → `patient_diagnosis_tests`

### __construct (line 18)

`public function __construct(DiagnosisCategoryRepository $categoryRepository)`

### index (line 23)

`public function index()`

Views: `diagnosis_categories.index`

### store (line 28)

`public function store(CreateDiagnosisCategoryRequest $request)`

Write/transaction candidates:

```php
31: $this->categoryRepository->create($input);
```

### show (line 36)

`public function show(DiagnosisCategory $diagnosisCategory)`

Views: `diagnosis_categories.show`

### edit (line 41)

`public function edit(DiagnosisCategory $diagnosisCategory)`

### update (line 46)

`public function update(UpdateDiagnosisCategoryRequest $request, DiagnosisCategory $diagnosisCategory)`

Write/transaction candidates:

```php
49: $this->categoryRepository->update($input, $diagnosisCategory->id);
```

### destroy (line 54)

`public function destroy(DiagnosisCategory $diagnosisCategory)`

Guard/validation candidates:

```php
62: if ($result) {
```

Write/transaction candidates:

```php
66: $diagnosisCategory->delete();
```

## hms/app/Http/Controllers/DoctorController.php

Models: `Appointment` → `appointments`, `BirthReport` → `birth_reports`, `DeathReport` → `death_reports`, `Doctor` → `doctors`, `EmployeePayroll` → `employee_payrolls`, `InvestigationReport` → `investigation_reports`, `IpdPatientDepartment` → `ipd_patient_departments`, `OperationReport` → `operation_reports`, `PatientAdmission` → `patient_admissions`, `PatientCase` → `patient_cases`, `Prescription` → `prescriptions`, `ScheduleDay` → `schedule_days`

### __construct (line 30)

`public function __construct(DoctorRepository $doctorRepo)`

### index (line 35)

`public function index()`

Views: `doctors.index`

### create (line 42)

`public function create()`

Views: `doctors.create`

### store (line 50)

`public function store(CreateDoctorRequest $request)`

Write/transaction candidates:

```php
54: $doctor = $this->doctorRepository->store($input);
59: ScheduleDay::create([
```

### show (line 73)

`public function show($doctorId)`

Guard/validation candidates:

```php
75: if (! getLoggedInUser()->hasRole('Receptionist') && checkRecordAccess($doctorId) && ! getLoggedInUser()->hasRole('Nurse')) {
78: if (! $doctor) {
87: if (! $data) {
```

Views: `errors.404`, `employees.doctors.show`, `errors.404`, `doctors.show`

### edit (line 95)

`public function edit(Doctor $doctor)`

Views: `doctors.edit`

### update (line 104)

`public function update(Doctor $doctor, UpdateDoctorRequest $request)`

Guard/validation candidates:

```php
106: if ($doctor->is_default == 1) {
112: if (empty($doctor)) {
```

Write/transaction candidates:

```php
120: $doctor = $this->doctorRepository->update($doctor, $input);
```

### destroy (line 126)

`public function destroy(Doctor $doctor)`

Guard/validation candidates:

```php
128: if ($doctor->is_default == 1) {
154: if ($result || $empPayRollResult) {
```

Write/transaction candidates:

```php
158: $doctor->user()->delete();
159: $doctor->address()->delete();
160: $doctor->delete();
```

### activeDeactiveStatus (line 165)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
169: $doctor->doctorUser()->update(['status' => $status]);
```

### doctorExport (line 174)

`public function doctorExport()`

Guard/validation candidates:

```php
177: if (!$doctors) {
```

## hms/app/Http/Controllers/DoctorDepartmentController.php

Models: `Doctor` → `doctors`, `DoctorDepartment` → `doctor_departments`

### __construct (line 17)

`public function __construct(DoctorDepartmentRepository $doctorDepartmentRepo)`

### index (line 22)

`public function index()`

Views: `doctor_departments.index`

### store (line 27)

`public function store(CreateDoctorDepartmentRequest $request)`

Write/transaction candidates:

```php
30: $this->doctorDepartmentRepository->create($input);
```

### show (line 35)

`public function show(DoctorDepartment $doctorDepartment)`

Guard/validation candidates:

```php
41: if (empty($doctorDepartment)) {
```

Views: `doctor_departments.show`

### edit (line 50)

`public function edit(DoctorDepartment $doctorDepartment)`

### update (line 55)

`public function update(DoctorDepartment $doctorDepartment, UpdateDoctorDepartmentRequest $request)`

Write/transaction candidates:

```php
58: $this->doctorDepartmentRepository->update($input, $doctorDepartment->id);
```

### destroy (line 63)

`public function destroy(DoctorDepartment $doctorDepartment)`

Guard/validation candidates:

```php
71: if ($result) {
```

Write/transaction candidates:

```php
75: $doctorDepartment->delete();
```

## hms/app/Http/Controllers/DoctorHolidayController.php

Models: `Doctor` → `doctors`, `DoctorHoliday` → `doctor_holidays`, `User` → `users`, `Appointment` → `appointments`

### __construct (line 20)

`public function __construct(HolidayRepository $holidayRepo)`

### index (line 25)

`public function index()`

Views: `doctor_holiday.index`

### create (line 30)

`public function create()`

Views: `doctor_holiday.create`

### store (line 37)

`public function store(CreateHolidayRequest $request)`

Guard/validation candidates:

```php
43: if($opdDates){
51: if ($holiday) {
```

Write/transaction candidates:

```php
49: $holiday = $this->holidayRepository->store($input);
```

### show (line 62)

`public function show($id)`

### edit (line 67)

`public function edit($id)`

### update (line 72)

`public function update(Request $request,$id)`

### destroy (line 77)

`public function destroy($id)`

Write/transaction candidates:

```php
79: $checkRecord = DoctorHoliday::destroy($id);
```

### holiday (line 84)

`public function holiday()`

Views: `holiday.index`

### doctorCreate (line 89)

`public function doctorCreate()`

Views: `holiday.create`

### doctorStore (line 97)

`public function doctorStore(CreateHolidayRequest $request)`

Guard/validation candidates:

```php
102: if ($holiday) {
```

Write/transaction candidates:

```php
100: $holiday = $this->holidayRepository->store($input);
```

### doctorDestroy (line 113)

`public function doctorDestroy($id)`

Guard/validation candidates:

```php
117: if ($doctorHoliday->doctor_id !== getLoggedInUser()->doctor->id) {
```

## hms/app/Http/Controllers/DoctorOPDChargeController.php

Models: `DoctorOPDCharge` → `doctor_opd_charges`

### __construct (line 20)

`public function __construct(DoctorOPDChargeRepository $doctorOPDChargeRepository)`

### index (line 25)

`public function index()`

Views: `doctor_opd_charges.index`

### store (line 32)

`public function store(CreateDoctorOPDChargeRequest $request)`

Write/transaction candidates:

```php
36: $this->doctorOPDChargeRepository->create($input);
```

### edit (line 41)

`public function edit(DoctorOPDCharge $doctorOPDCharge)`

### update (line 46)

`public function update(UpdateDoctorOPDChargeRequest $request, DoctorOPDCharge $doctorOPDCharge)`

Write/transaction candidates:

```php
50: $this->doctorOPDChargeRepository->update($input, $doctorOPDCharge->id);
```

### destroy (line 55)

`public function destroy(DoctorOPDCharge $doctorOPDCharge)`

Write/transaction candidates:

```php
57: $doctorOPDCharge->delete();
```

### doctorOPDChargeExport (line 62)

`public function doctorOPDChargeExport()`

Guard/validation candidates:

```php
65: if (!$doctorOPDCharge) {
```

## hms/app/Http/Controllers/DocumentController.php

Models: `Document` → `documents`

### __construct (line 18)

`public function __construct(DocumentRepository $documentRepo)`

### index (line 23)

`public function index()`

Views: `documents.index`

### store (line 30)

`public function store(CreateDocumentRequest $request)`

Write/transaction candidates:

```php
34: $this->documentRepository->store($input);
```

### show (line 39)

`public function show(Document $document)`

Views: `documents.show`

### edit (line 48)

`public function edit(Document $document)`

Guard/validation candidates:

```php
50: if (getLoggedinPatient() && checkRecordAccess($document->patient_id)) {
```

### update (line 57)

`public function update(Document $document, UpdateDocumentRequest $request)`

### destroy (line 64)

`public function destroy(Document $document)`

Guard/validation candidates:

```php
66: if (getLoggedinPatient() && checkRecordAccess($document->patient_id)) {
```

### downloadMedia (line 75)

`public function downloadMedia(Document $document)`

Guard/validation candidates:

```php
77: if (getLoggedinPatient() && checkRecordAccess($document->patient_id)) {
85: if (config('app.media_disc') === 'public') {
```

## hms/app/Http/Controllers/DocumentTypeController.php

Models: `Document` → `documents`, `DocumentType` → `document_types`

### __construct (line 16)

`public function __construct(DocumentTypeRepository $documentTypeRepo)`

### index (line 21)

`public function index()`

Views: `document_types.index`

### store (line 26)

`public function store(CreateDocumentTypeRequest $request)`

Write/transaction candidates:

```php
30: $this->documentTypeRepository->create($input);
```

### show (line 35)

`public function show(DocumentType $documentType)`

Guard/validation candidates:

```php
39: if (! getLoggedInUser()->hasRole('Admin')) {
```

Views: `document_types.show`

### edit (line 46)

`public function edit(DocumentType $documentType)`

### update (line 51)

`public function update(DocumentType $documentType, UpdateDocumentTypeRequest $request)`

Write/transaction candidates:

```php
53: $this->documentTypeRepository->update($request->all(), $documentType->id);
```

### destroy (line 58)

`public function destroy(DocumentType $documentType)`

Guard/validation candidates:

```php
66: if ($result) {
```

Write/transaction candidates:

```php
70: $this->documentTypeRepository->delete($documentType->id);
```

## hms/app/Http/Controllers/EmailTemplateController.php

Models: `EmailTemplate` → `email_templates`

### index (line 10)

`public function index()`

Views: `email-template.index`

### edit (line 14)

`public function edit($id)`

Views: `email-template.edit`

### update (line 21)

`public function update(Request $request, $id)`

Guard/validation candidates:

```php
24: $request->validate([
```

Write/transaction candidates:

```php
28: $template->update([
```

## hms/app/Http/Controllers/Employee/BillController.php

Models: `Bill` → `bills`

### __construct (line 24)

`public function __construct(BillRepository $billRepo)`

### index (line 29)

`public function index(Request $request)`

Guard/validation candidates:

```php
31: if ($request->ajax()) {
```

Views: `employees.bills.index`

### show (line 41)

`public function show(Bill $bill)`

Guard/validation candidates:

```php
43: if (checkRecordAccess($bill->patient_id)) {
```

Views: `errors.404`, `employees.bills.show`

### convertToPdf (line 56)

`public function convertToPdf(Bill $bill)`

## hms/app/Http/Controllers/Employee/DoctorController.php

Models: `Doctor` → `doctors`, `User` → `users`

### index (line 19)

`public function index(Request $request)`

Guard/validation candidates:

```php
21: if ($request->ajax()) {
```

Views: `employees.doctors.index`

### show (line 31)

`public function show($id)`

Guard/validation candidates:

```php
34: if (! $doctor) {
```

Views: `employees.doctors.show`

## hms/app/Http/Controllers/Employee/InvoiceController.php

Models: `Invoice` → `invoices`, `Setting` → `settings`

### __construct (line 20)

`public function __construct(InvoiceRepository $invoiceRepo)`

### index (line 25)

`public function index(Request $request)`

Views: `employees.invoices.index`

### show (line 33)

`public function show(Invoice $invoice)`

Guard/validation candidates:

```php
35: if (checkRecordAccess($invoice->patient_id)) {
```

Views: `errors.404`, `employees.invoices.show`

### convertToPdf (line 45)

`public function convertToPdf(Invoice $invoice)`

Guard/validation candidates:

```php
47: if (checkRecordAccess($invoice->patient_id)) {
```

Views: `errors.404`

## hms/app/Http/Controllers/Employee/NoticeBoardController.php

Models: `NoticeBoard` → `notice_boards`

### index (line 17)

`public function index(Request $request)`

Guard/validation candidates:

```php
19: if ($request->ajax()) {
```

Views: `employees.notice_boards.index`

### show (line 26)

`public function show($id)`

## hms/app/Http/Controllers/Employee/PatientAdmissionController.php

Models: `PatientAdmission` → `patient_admissions`

### index (line 14)

`public function index()`

Views: `employees.patient_admissions.index`

### show (line 21)

`public function show(PatientAdmission $patientAdmission)`

Guard/validation candidates:

```php
23: if (checkRecordAccess($patientAdmission->patient_id)) {
```

Views: `errors.404`, `employees.patient_admissions.show`

## hms/app/Http/Controllers/Employee/PatientDiagnosisTestController.php

Models: `PatientDiagnosisTest` → `patient_diagnosis_tests`

### __construct (line 22)

`public function __construct( PatientDiagnosisTestRepository $patientDiagnosisTestRepository )`

### index (line 28)

`public function index(Request $request)`

Views: `employees.patient_diagnosis_test.index`

### show (line 33)

`public function show(PatientDiagnosisTest $patientDiagnosisTest)`

Guard/validation candidates:

```php
35: if (checkRecordAccess($patientDiagnosisTest->patient_id)) {
```

Views: `errors.404`, `employees.patient_diagnosis_test.show`

### convertToPdf (line 44)

`public function convertToPdf(PatientDiagnosisTest $patientDiagnosisTest)`

## hms/app/Http/Controllers/Employee/PayrollController.php

Models: `EmployeePayroll` → `employee_payrolls`

### index (line 20)

`public function index(Request $request)`

Guard/validation candidates:

```php
22: if ($request->ajax()) {
```

Views: `employees.payrolls.index`

### userPayrollExport (line 29)

`public function userPayrollExport()`

Guard/validation candidates:

```php
37: if ($employeePayrolls->count() == 0) {
```

## hms/app/Http/Controllers/Employee/PrescriptionController.php

Models: `Prescription` → `prescriptions`

### index (line 22)

`public function index()`

Views: `employee_prescription_list.index`

### show (line 27)

`public function show($id)`

Views: `prescriptions.view`

### prescriptionExport (line 37)

`public function prescriptionExport()`

Guard/validation candidates:

```php
39: if (getLoggedInUser()->hasRole('Pharmacist')) {
41: if ($prescriptions->count() == 0) {
50: if ($prescriptions->count() == 0) {
```

## hms/app/Http/Controllers/EmployeePayrollController.php

Models: `EmployeePayroll` → `employee_payrolls`

### __construct (line 20)

`public function __construct(EmployeePayrollRepository $employeePayrollRepo)`

### index (line 25)

`public function index()`

Views: `employee_payrolls.index`

### create (line 32)

`public function create()`

Views: `employee_payrolls.create`

### store (line 45)

`public function store(CreateEmployeePayrollRequest $request)`

Write/transaction candidates:

```php
48: $this->employeePayrollRepository->create($input);
```

### show (line 55)

`public function show(EmployeePayroll $employeePayroll)`

Guard/validation candidates:

```php
57: if (checkRecordAccess($employeePayroll->owner_id)) {
```

Views: `errors.404`, `employee_payrolls.show`

### edit (line 64)

`public function edit(EmployeePayroll $employeePayroll)`

Views: `employee_payrolls.edit`

### update (line 73)

`public function update(EmployeePayroll $employeePayroll, UpdateEmployeePayrollRequest $request)`

Write/transaction candidates:

```php
76: $this->employeePayrollRepository->update($input, $employeePayroll->id);
```

### destroy (line 82)

`public function destroy(EmployeePayroll $employeePayroll)`

Write/transaction candidates:

```php
84: $employeePayroll->delete();
```

### getEmployeesList (line 89)

`public function getEmployeesList(Request $request)`

Guard/validation candidates:

```php
91: if (empty($request->get('id'))) {
95: if ($request->id == 2) {
```

### employeePayrollExport (line 106)

`public function employeePayrollExport()`

Guard/validation candidates:

```php
110: if (!$employeePayroll) {
```

### showModal (line 117)

`public function showModal(EmployeePayroll $employeePayroll)`

Guard/validation candidates:

```php
119: if ($employeePayroll->type_string == 'Doctor') {
```

## hms/app/Http/Controllers/EnquiryController.php

Models: `Enquiry` → `enquiries`

### __construct (line 16)

`public function __construct(EnquiryRepository $enqRepo)`

### index (line 21)

`public function index()`

Views: `enquiries.index`

### store (line 28)

`public function store(CreateEnquiryRequest $request)`

Write/transaction candidates:

```php
33: $this->enquiryRepository->store($input);
```

### show (line 41)

`public function show(Enquiry $enquiry)`

Guard/validation candidates:

```php
43: if ($enquiry->status == 0) {
```

Write/transaction candidates:

```php
44: $enquiry->update(['viewed_by' => getLoggedInUserId()]);
45: $enquiry->update(['status' => 1]);
```

Views: `enquiries.show`

### activeDeactiveStatus (line 51)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
56: $enquiry->update(['viewed_by' => $viewedStatus]);
57: $enquiry->update(['status' => $status]);
```

### contactUs (line 62)

`public function contactUs()`

Views: `web.home.contact_us`

## hms/app/Http/Controllers/ExpenseController.php

Models: `Expense` → `expenses`

### __construct (line 22)

`public function __construct(ExpenseRepository $expenseRepository)`

### index (line 27)

`public function index()`

Views: `expenses.index`

### store (line 37)

`public function store(CreateExpenseRequest $request)`

Write/transaction candidates:

```php
41: $this->expenseRepository->store($input);
```

### show (line 47)

`public function show(Expense $expense)`

Views: `expenses.show`

### edit (line 56)

`public function edit(Expense $expense)`

### update (line 61)

`public function update(UpdateExpenseRequest $request, Expense $expense)`

### destroy (line 68)

`public function destroy(Expense $expense)`

### downloadMedia (line 75)

`public function downloadMedia(Expense $expense)`

Guard/validation candidates:

```php
80: if (config('app.media_disc') === 'public') {
```

### expenseExport (line 96)

`public function expenseExport()`

Guard/validation candidates:

```php
99: if (!$expenses) {
```

## hms/app/Http/Controllers/FrontServiceController.php

Models: `FrontService` → `front_services`

### __construct (line 15)

`public function __construct(FrontServiceRepository $frontServiceRepository)`

### index (line 20)

`public function index()`

Views: `front_settings.front_services.index`

### store (line 25)

`public function store(FrontServiceRequest $request)`

Write/transaction candidates:

```php
29: $this->frontServiceRepository->store($input);
```

### edit (line 37)

`public function edit($id)`

### update (line 44)

`public function update($id, UpdateFrontServiceRequest $request)`

### destroy (line 55)

`public function destroy($id)`

Write/transaction candidates:

```php
60: $frontService->delete();
```

## hms/app/Http/Controllers/FrontSettingController.php

Models: `FrontSetting` → `front_settings`

### __construct (line 15)

`public function __construct(FrontSettingRepository $frontSettingRepository)`

### index (line 20)

`public function index(Request $request)`

Views: `front_settings.$sectionName`

### update (line 28)

`public function update(Request $request)`

## hms/app/Http/Controllers/GeneratePatientIdCardController.php

Models: `Patient` → `patients`, `PatientIdCardTemplate` → `patient_id_card_templates`

### __construct (line 19)

`public function __construct(GeneratePatientIdCardRepository $GeneratePatientIdCardRepositorie)`

### index (line 24)

`public function index()`

Views: `generate_patient_id_card.index`

### store (line 32)

`public function store(CreatePatientIdCardRequest $request)`

Write/transaction candidates:

```php
35: $this->GeneratePatientIdCardRepositorie->store($input);
```

### show (line 40)

`public function show($uniqueId)`

Guard/validation candidates:

```php
43: if(empty($patients->patient_unique_id)){
```

Write/transaction candidates:

```php
44: $patients->update(['patient_unique_id' => strtoupper(Patient::generateUniquePatientId())]);
```

### destroy (line 49)

`public function destroy($id)`

Write/transaction candidates:

```php
51: Patient::find($id)->update(['template_id' => null]);
```

### downloadIdCard (line 56)

`public function downloadIdCard($id)`

Guard/validation candidates:

```php
60: if(empty($patientIdCardData->patient_unique_id)){
68: if($arrUrl == "ui-avatars.com"){
```

Write/transaction candidates:

```php
61: $patientIdCardData->update(['patient_unique_id' => strtoupper(Patient::generateUniquePatientId())]);
```

### generateQrCode (line 82)

`public function generateQrCode($uniqueId)`

## hms/app/Http/Controllers/GoogleMeetLiveConsultationController.php

Models: `Doctor` → `doctors`, `EventGoogleCalendar` → `event_google_calendars`, `GoogleCalendarIntegration` → `google_calendar_integrations`, `GoogleCalendarList` → `google_calendar_lists`, `LiveConsultation` → `live_consultations`, `User` → `users`

### googleConfig (line 28)

`public function googleConfig()`

Guard/validation candidates:

```php
32: if (empty($name)) {
```

### index (line 45)

`public function index()`

Views: `goole_meet_consultation.index`

### store (line 53)

`public function store($liveConsultation, $accessToken, $meta)`

Guard/validation candidates:

```php
62: if ($accessToken) {
75: if ($liveConsultation->platform_type == LiveConsultation::GOOGLE_MEET) {
```

### oauth (line 102)

`public function oauth()`

Guard/validation candidates:

```php
108: if (empty($name) && file_exists(storage_path($name))) {
```

### redirect (line 120)

`public function redirect(Request $request)`

Guard/validation candidates:

```php
132: if ($exists) {
```

Write/transaction candidates:

```php
126: DB::beginTransaction();
133: GoogleCalendarIntegration::where('user_id',getLoggedInUserId())->delete();
134: GoogleCalendarList::where('user_id',getLoggedInUserId())->delete();
137: $googleCalendarIntegration = GoogleCalendarIntegration::create([
147: DB::commit();
150: DB::rollBack();
```

### fetchCalendarListAndSyncToDB (line 159)

`public function fetchCalendarListAndSyncToDB()`

Guard/validation candidates:

```php
166: if ($calendarListEntry->accessRole == 'owner') {
```

Write/transaction candidates:

```php
167: $googleCalendarList[] = GoogleCalendarList::create([
```

### eventGoogleCalendarStore (line 179)

`public function eventGoogleCalendarStore(Request $request)`

Guard/validation candidates:

```php
183: if($eventGoogleCalendars){
```

Write/transaction candidates:

```php
185: $eventGoogleCalendar->delete();
200: EventGoogleCalendar::create($data);
```

### syncGoogleCalendarList (line 206)

`public function syncGoogleCalendarList()`

Guard/validation candidates:

```php
220: if ($calendarListEntry->accessRole == 'owner') {
227: if (! $exists) {
```

Write/transaction candidates:

```php
228: $googleCalendarList[] = GoogleCalendarList::create([
238: EventGoogleCalendar::whereIn('google_calendar_id', $existingCalendars)->delete();
239: GoogleCalendarList::whereIn('google_calendar_id', $existingCalendars)->delete();
```

### getAccessToken (line 244)

`public function getAccessToken($userId)`

Guard/validation candidates:

```php
250: if(empty($user->gCredentials)){
256: if (is_array($accessToken) && count($accessToken) == 0) {
262: if (empty($accessToken['access_token'])) {
269: if ($this->client->isAccessTokenExpired()) {
272: if(isset($accessToken['refresh_token'])){
277: if (is_array($accessToken) && count($accessToken) == 0) {
283: if (empty($accessToken['access_token'])) {
```

Write/transaction candidates:

```php
288: $calendarRecord->update([
```

### disconnectGoogleCalendar (line 303)

`public function disconnectGoogleCalendar()`

Write/transaction candidates:

```php
305: EventGoogleCalendar::where('user_id',getLoggedInUserId())->delete();
306: GoogleCalendarIntegration::where('user_id',getLoggedInUserId())->delete();
307: GoogleCalendarList::where('user_id',getLoggedInUserId())->delete();
```

### validateGoogleCalendarJsonFile (line 314)

`public function validateGoogleCalendarJsonFile($input)`

Guard/validation candidates:

```php
328: if ($validator->fails()) {
```

### googleCalendarJsonFileStore (line 333)

`public function googleCalendarJsonFileStore(Request $request)`

Guard/validation candidates:

```php
339: if ($error) {
344: if (!empty($input['google_json_file'])) {
348: if ($googleCalendarJsonFileMedia !== null) {
```

Write/transaction candidates:

```php
358: $doctor->update(['google_json_file_path' => $googleJsonFilePath]);
```

## hms/app/Http/Controllers/HomeController.php

Models: `Accountant` → `accountants`, `AdvancedPayment` → `advanced_payments`, `Appointment` → `appointments`, `Bed` → `beds`, `Bill` → `bills`, `Doctor` → `doctors`, `Enquiry` → `enquiries`, `IpdBill` → `ipd_bills`, `IpdPatientDepartment` → `ipd_patient_departments`, `LabTechnician` → `lab_technicians`, `LiveConsultation` → `live_consultations`, `Module` → `modules`, `NoticeBoard` → `notice_boards`, `Nurse` → `nurses`, `Patient` → `patients`, `Payment` → `payments`, `Pharmacist` → `pharmacists`, `Receptionist` → `receptionists`, `Setting` → `settings`, `User` → `users`

### __construct (line 33)

`public function __construct(DashboardRepository $dashboardRepository)`

### index (line 39)

`public function index()`

Views: `home`

### dashboard (line 44)

`public function dashboard()`

Views: `dashboard.index`

### patientDashboard (line 70)

`public function patientDashboard()`

Views: `dashboard.patient-dashboard`

### dashboardChart (line 84)

`public function dashboardChart()`

### incomeExpenseReport (line 91)

`public function incomeExpenseReport(Request $request)`

## hms/app/Http/Controllers/HospitalScheduleController.php

Models: `HospitalSchedule` → `hospital_schedules`, `ScheduleDay` → `schedule_days`

### index (line 12)

`public function index()`

Views: `hospital_schedule.index`

### store (line 22)

`public function store(Request $request)`

Guard/validation candidates:

```php
26: if (isset($input['checked_week_days'])) {
36: if (strtotime($startTime) > strtotime($endTime)) {
```

Write/transaction candidates:

```php
30: HospitalSchedule::whereDayOfWeek($dayOfWeek)->delete();
```

### checkRecord (line 58)

`public function checkRecord(Request $request)`

Guard/validation candidates:

```php
62: if (isset($input['checked_week_days'])) {
71: if ($scheduleDayExists) {
```

## hms/app/Http/Controllers/IncomeController.php

Models: `Income` → `incomes`

### __construct (line 22)

`public function __construct(IncomeRepository $incomeRepository)`

### index (line 27)

`public function index()`

Views: `incomes.index`

### store (line 37)

`public function store(CreateIncomeRequest $request)`

Write/transaction candidates:

```php
41: $this->incomeRepository->store($input);
```

### show (line 47)

`public function show(Income $income)`

Views: `incomes.show`

### edit (line 56)

`public function edit(Income $income)`

### update (line 61)

`public function update(UpdateIncomeRequest $request, Income $income)`

### destroy (line 68)

`public function destroy(Income $income)`

### downloadMedia (line 75)

`public function downloadMedia(Income $income)`

Guard/validation candidates:

```php
80: if (config('app.media_disc') === 'public') {
```

### incomeExport (line 96)

`public function incomeExport()`

Guard/validation candidates:

```php
99: if (!$incomes) {
```

## hms/app/Http/Controllers/InsuranceController.php

Models: `Insurance` → `insurances`, `PatientAdmission` → `patient_admissions`

### __construct (line 21)

`public function __construct(InsuranceRepository $insuranceRepo)`

### index (line 26)

`public function index()`

Views: `insurances.index`

### create (line 33)

`public function create()`

Views: `insurances.create`

### store (line 38)

`public function store(CreateInsuranceRequest $request)`

Write/transaction candidates:

```php
45: DB::beginTransaction();
46: $insurance = $this->insuranceRepository->store($input);
47: DB::commit();
49: DB::rollBack();
```

### show (line 61)

`public function show(Insurance $insurance)`

Views: `insurances.show`

### edit (line 68)

`public function edit(Insurance $insurance)`

Views: `insurances.edit`

### update (line 75)

`public function update(Insurance $insurance, UpdateInsuranceRequest $request)`

Write/transaction candidates:

```php
82: DB::beginTransaction();
83: $insurance = $this->insuranceRepository->update($insurance, $input);
84: DB::commit();
86: DB::rollBack();
```

### destroy (line 96)

`public function destroy(Insurance $insurance)`

Guard/validation candidates:

```php
103: if ($result) {
```

Write/transaction candidates:

```php
108: $this->insuranceRepository->delete($insurance->id);
```

### activeDeactiveInsurance (line 116)

`public function activeDeactiveInsurance($id)`

Write/transaction candidates:

```php
120: $insurance->update(['status' => $insurance->status]);
```

### insuranceExport (line 125)

`public function insuranceExport()`

Guard/validation candidates:

```php
128: if (!$insurances) {
```

## hms/app/Http/Controllers/InvestigationReportController.php

Models: `InvestigationReport` → `investigation_reports`, `Patient` → `patients`

### __construct (line 20)

`public function __construct(InvestigationReportRepository $investigationReportRepo)`

### index (line 25)

`public function index()`

Views: `investigation_reports.index`

### create (line 32)

`public function create()`

Views: `investigation_reports.create`

### store (line 41)

`public function store(CreateInvestigationReportRequest $request)`

Guard/validation candidates:

```php
48: if (! empty($birthDate) && $reportDate < $birthDate) {
```

Write/transaction candidates:

```php
54: $this->investigationReportRepository->store($input);
```

### show (line 60)

`public function show(InvestigationReport $investigationReport)`

Guard/validation candidates:

```php
64: if (empty($investigationReport)) {
```

Views: `investigation_reports.show`

### edit (line 73)

`public function edit(InvestigationReport $investigationReport)`

Guard/validation candidates:

```php
75: if (checkRecordAccess($investigationReport->doctor_id)) {
```

Views: `errors.404`, `investigation_reports.edit`

### update (line 88)

`public function update(InvestigationReport $investigationReport, UpdateInvestigationReportRequest $request)`

Guard/validation candidates:

```php
90: if (empty($investigationReport)) {
101: if (! empty($birthDate) && $reportDate < $birthDate) {
```

Write/transaction candidates:

```php
107: $this->investigationReportRepository->update($input, $investigationReport->id);
```

### destroy (line 113)

`public function destroy(InvestigationReport $investigationReport)`

Guard/validation candidates:

```php
115: if (checkRecordAccess($investigationReport->doctor_id)) {
```

Write/transaction candidates:

```php
118: $investigationReport->delete();
```

### downloadMedia (line 124)

`public function downloadMedia(InvestigationReport $investigationReport)`

Guard/validation candidates:

```php
129: if (config('app.media_disc') === 'public') {
```

## hms/app/Http/Controllers/InvoiceController.php

Models: `Invoice` → `invoices`, `Setting` → `settings`

### __construct (line 19)

`public function __construct(InvoiceRepository $invoiceRepo)`

### index (line 24)

`public function index()`

Views: `invoices.index`

### create (line 31)

`public function create()`

Views: `invoices.create`

### store (line 38)

`public function store(CreateInvoiceRequest $request)`

Write/transaction candidates:

```php
41: DB::beginTransaction();
44: DB::commit();
46: DB::rollBack();
```

### show (line 54)

`public function show(Invoice $invoice)`

Views: `invoices.show`

### edit (line 62)

`public function edit(Invoice $invoice)`

Views: `invoices.edit`

### update (line 71)

`public function update(Invoice $invoice, UpdateInvoiceRequest $request)`

Write/transaction candidates:

```php
74: DB::beginTransaction();
76: DB::commit();
78: DB::rollBack();
```

### destroy (line 86)

`public function destroy(Invoice $invoice)`

Write/transaction candidates:

```php
88: $this->invoiceRepository->delete($invoice->id);
```

### convertToPdf (line 93)

`public function convertToPdf(Invoice $invoice)`

## hms/app/Http/Controllers/IpdBillController.php

Models: `IpdConsultantRegister` → `ipd_consultant_registers`, `IpdDiagnosis` → `ipd_diagnoses`, `IpdPatientDepartment` → `ipd_patient_departments`, `IpdPrescription` → `ipd_prescriptions`

### __construct (line 18)

`public function __construct(IpdBillRepository $ipdBillRepo)`

### store (line 23)

`public function store(CreateIpdBillRequest $request)`

### ipdBillConvertToPdf (line 31)

`public function ipdBillConvertToPdf(IpdPatientDepartment $ipdPatientDepartment)`

Guard/validation candidates:

```php
33: if(app()->getLocale() == "zh"){
```

### ipdDischargePatientToPdf (line 46)

`public function ipdDischargePatientToPdf(IpdPatientDepartment $ipdPatientDepartment)`

Guard/validation candidates:

```php
48: if(app()->getLocale() == "zh"){
```

## hms/app/Http/Controllers/IpdChargeController.php

Models: `IpdCharge` → `ipd_charges`

### __construct (line 19)

`public function __construct(IpdChargeRepository $ipdChargeRepo)`

### index (line 24)

`public function index(Request $request)`

Guard/validation candidates:

```php
26: if ($request->ajax()) {
```

### store (line 31)

`public function store(CreateIpdChargeRequest $request)`

Write/transaction candidates:

```php
36: $this->ipdChargeRepository->create($input);
```

### edit (line 42)

`public function edit(IpdCharge $ipdCharge)`

### update (line 47)

`public function update(IpdCharge $ipdCharge, UpdateIpdChargeRequest $request)`

Write/transaction candidates:

```php
52: $this->ipdChargeRepository->update($input, $ipdCharge->id);
```

### destroy (line 57)

`public function destroy(IpdCharge $ipdCharge)`

Write/transaction candidates:

```php
59: $ipdCharge->delete();
```

### getChargeCategoryList (line 64)

`public function getChargeCategoryList(Request $request)`

### getChargeList (line 71)

`public function getChargeList(Request $request)`

### getChargeStandardRate (line 78)

`public function getChargeStandardRate(Request $request)`

## hms/app/Http/Controllers/IpdConsultantRegisterController.php

Models: `IpdConsultantRegister` → `ipd_consultant_registers`

### __construct (line 18)

`public function __construct(IpdConsultantRegisterRepository $ipdConsultantRegisterRepo)`

### index (line 23)

`public function index(Request $request)`

Guard/validation candidates:

```php
25: if ($request->ajax()) {
```

### store (line 33)

`public function store(CreateIpdConsultantRegisterRequest $request)`

Guard/validation candidates:

```php
38: if ($input['doctor_id'][$i] == 0) {
45: if ($result) {
```

Write/transaction candidates:

```php
43: $result = $this->ipdConsultantRegisterRepository->store($input);
```

### edit (line 52)

`public function edit(IpdConsultantRegister $ipdConsultantRegister)`

### update (line 57)

`public function update(IpdConsultantRegister $ipdConsultantRegister, UpdateIpdConsultantRegisterRequest $request)`

Write/transaction candidates:

```php
60: $this->ipdConsultantRegisterRepository->update($input, $ipdConsultantRegister->id);
```

### destroy (line 65)

`public function destroy(IpdConsultantRegister $ipdConsultantRegister)`

Write/transaction candidates:

```php
67: $ipdConsultantRegister->delete();
```

## hms/app/Http/Controllers/IpdDiagnosisController.php

Models: `IpdDiagnosis` → `ipd_diagnoses`

### __construct (line 19)

`public function __construct(IpdDiagnosisRepository $ipdDiagnosisRepo)`

### index (line 24)

`public function index(Request $request)`

Guard/validation candidates:

```php
26: if ($request->ajax()) {
```

### store (line 31)

`public function store(CreateIpdDiagnosisRequest $request)`

Write/transaction candidates:

```php
34: $this->ipdDiagnosisRepository->store($input);
```

### edit (line 39)

`public function edit(IpdDiagnosis $ipdDiagnosis)`

### update (line 44)

`public function update(IpdDiagnosis $ipdDiagnosis, UpdateIpdDiagnosisRequest $request)`

### destroy (line 51)

`public function destroy(IpdDiagnosis $ipdDiagnosis)`

### downloadMedia (line 58)

`public function downloadMedia(IpdDiagnosis $ipdDiagnosis)`

Guard/validation candidates:

```php
62: if ($media != null) {
```

## hms/app/Http/Controllers/IpdOperationController.php

Models: `IpdOperation` → `ipd_operation`

### store (line 11)

`public function store(CreateIPDOperationRequest $request)`

Guard/validation candidates:

```php
17: if ($isExist) {
```

Write/transaction candidates:

```php
40: IpdOperation::create($input);
```

### edit (line 45)

`public function edit($id)`

### update (line 52)

`public function update($id, CreateIPDOperationRequest $request)`

Write/transaction candidates:

```php
55: $ipdOperation->update($request->all());
```

### delete (line 60)

`public function delete($id)`

Write/transaction candidates:

```php
62: IpdOperation::where('id', $id)->delete();
```

## hms/app/Http/Controllers/IpdPatientDepartmentController.php

Models: `AddCustomFields` → `add_custom_fields`, `Bed` → `beds`, `IpdCharge` → `ipd_charges`, `IpdPatientDepartment` → `ipd_patient_departments`, `IpdPayment` → `ipd_payments`, `PatientCase` → `patient_cases`

### __construct (line 24)

`public function __construct(IpdPatientDepartmentRepository $ipdPatientDepartmentRepo)`

### index (line 29)

`public function index()`

Views: `ipd_patient_departments.index`

### create (line 36)

`public function create()`

Views: `ipd_patient_departments.create`

### store (line 44)

`public function store(CreateIpdPatientDepartmentRequest $request)`

Guard/validation candidates:

```php
50: if($existsCaseId && $existsCaseId->discharge == 0){
```

Write/transaction candidates:

```php
56: $this->ipdPatientDepartmentRepository->store($input);
```

### show (line 63)

`public function show(IpdPatientDepartment $ipdPatientDepartment)`

Views: `ipd_patient_departments.show`

### edit (line 93)

`public function edit(IpdPatientDepartment $ipdPatientDepartment)`

Views: `ipd_patient_departments.edit`

### update (line 104)

`public function update(IpdPatientDepartment $ipdPatientDepartment, UpdateIpdPatientDepartmentRequest $request)`

Guard/validation candidates:

```php
108: if(isset($input['case_id'])){
109: if ($ipdPatientDepartment->case_id != $input['case_id']) {
112: if ($existingPatient) {
```

### destroy (line 125)

`public function destroy(IpdPatientDepartment $ipdPatientDepartment)`

### getPatientCasesList (line 132)

`public function getPatientCasesList(Request $request)`

### getPatientBedsList (line 139)

`public function getPatientBedsList(Request $request)`

## hms/app/Http/Controllers/IpdPaymentController.php

Models: `IpdPayment` → `ipd_payments`

### __construct (line 30)

`public function __construct(IpdPaymentRepository $ipdPaymentRepo)`

### index (line 40)

`public function index(Request $request)`

Guard/validation candidates:

```php
42: if ($request->ajax()) {
```

### store (line 47)

`public function store(CreateIpdPaymentRequest $request)`

Guard/validation candidates:

```php
52: if($input['payment_mode'] == IpdPayment::PAYMENT_MODES_STRIPE){
72: if(!in_array(strtoupper(getCurrentCurrency()),getFlutterWaveSupportedCurrencies())){
79: if(!$flutterwavePublicKey && !$flutterwaveSecretKey){
93: if (strtoupper(getCurrentCurrency()) != 'INR') {
```

Write/transaction candidates:

```php
110: $this->ipdPaymentRepository->store($input);
```

### edit (line 119)

`public function edit(IpdPayment $ipdPayment)`

### update (line 124)

`public function update(IpdPayment $ipdPayment, UpdateIpdPaymentRequest $request)`

### destroy (line 131)

`public function destroy(IpdPayment $ipdPayment)`

### downloadMedia (line 138)

`public function downloadMedia(IpdPayment $ipdPayment)`

Guard/validation candidates:

```php
142: if ($media != null) {
```

### ipdStripePaymentSuccess (line 152)

`public function ipdStripePaymentSuccess(Request $request)`

Guard/validation candidates:

```php
158: if(getLoggedinPatient()){
```

### ipdRazorpayPayment (line 165)

`public function ipdRazorpayPayment(Request $request)`

### ipdRazorpayPaymentSuccess (line 172)

`public function ipdRazorpayPaymentSuccess(Request $request)`

Guard/validation candidates:

```php
178: if(getLoggedinPatient()){
```

### ipdFlutterwavePaymentSuccess (line 185)

`public function ipdFlutterwavePaymentSuccess(Request $request)`

Guard/validation candidates:

```php
190: if(!$flutterwavePublicKey && !$flutterwaveSecretKey){
199: if($request['status'] == 'cancelled'){
202: if(getLoggedinPatient()){
212: if(getLoggedinPatient()){
```

### phonePePaymentSuccess (line 219)

`public function phonePePaymentSuccess(Request $request)`

Guard/validation candidates:

```php
225: if(getLoggedinPatient()){
```

### ipdPaystackPayment (line 232)

`public function ipdPaystackPayment(Request $request)`

Guard/validation candidates:

```php
234: if(!in_array(strtoupper(getCurrentCurrency()),getPayStackSupportedCurrencies())){
237: if(getLoggedinPatient()){
263: if(getLoggedinPatient()){
```

## hms/app/Http/Controllers/IpdPrescriptionController.php

Models: `IpdPrescription` → `ipd_prescriptions`, `Medicine` → `medicines`

### __construct (line 20)

`public function __construct(IpdPrescriptionRepository $ipdPrescriptionRepo)`

### index (line 25)

`public function index(Request $request)`

Guard/validation candidates:

```php
27: if ($request->ajax()) {
```

### store (line 32)

`public function store(CreateIpdPrescriptionRequest $request)`

Guard/validation candidates:

```php
42: if (! empty($duplicateIds)) {
50: if ($medicine->available_quantity < $qty) {
```

Write/transaction candidates:

```php
58: $this->ipdPrescriptionRepository->store($input);
```

### show (line 64)

`public function show(IpdPrescription $ipdPrescription)`

Views: `ipd_prescriptions.show_ipd_prescription_data`

### edit (line 69)

`public function edit(IpdPrescription $ipdPrescription)`

### update (line 77)

`public function update(IpdPrescription $ipdPrescription, UpdateIpdPrescriptionRequest $request)`

Guard/validation candidates:

```php
102: if (! empty($duplicateIds)) {
110: if ($medicine->available_quantity < $qty && ! array_key_exists($input['medicine_id'][$key], $result)) {
```

### destroy (line 122)

`public function destroy(IpdPrescription $ipdPrescription)`

Write/transaction candidates:

```php
124: $ipdPrescription->ipdPrescriptionItems()->delete();
125: $ipdPrescription->delete();
```

### getMedicineList (line 130)

`public function getMedicineList(Request $request)`

### getAvailableMedicineQuantity (line 137)

`public function getAvailableMedicineQuantity(Medicine $medicine)`

### convertToPDF (line 142)

`public function convertToPDF(IpdPrescription $ipdPrescription)`

## hms/app/Http/Controllers/IpdTimelineController.php

Models: `IpdTimeline` → `ipd_timelines`

### __construct (line 17)

`public function __construct(IpdTimelineRepository $ipdTimelineRepo)`

### index (line 22)

`public function index(Request $request)`

Views: `ipd_timelines.index`

### store (line 29)

`public function store(CreateIpdTimelineRequest $request)`

Write/transaction candidates:

```php
32: $this->ipdTimelineRepository->store($input);
```

### edit (line 37)

`public function edit(IpdTimeline $ipdTimeline)`

### update (line 42)

`public function update(IpdTimeline $ipdTimeline, UpdateIpdTimelineRequest $request)`

### destroy (line 49)

`public function destroy(IpdTimeline $ipdTimeline)`

### downloadMedia (line 56)

`public function downloadMedia(IpdTimeline $ipdTimeline)`

Guard/validation candidates:

```php
59: if ($media != null) {
```

## hms/app/Http/Controllers/IssuedItemController.php

Models: `IssuedItem` → `issued_items`

### __construct (line 16)

`public function __construct(IssuedItemRepository $issuedItemRepo)`

### index (line 21)

`public function index()`

Views: `issued_items.index`

### create (line 28)

`public function create()`

Views: `issued_items.create`

### store (line 35)

`public function store(CreateIssuedItemRequest $request)`

Write/transaction candidates:

```php
39: $this->issuedItemRepository->store($input);
```

### show (line 45)

`public function show(IssuedItem $issuedItem)`

Views: `issued_items.show`

### destroy (line 50)

`public function destroy(IssuedItem $issuedItem)`

### returnIssuedItem (line 57)

`public function returnIssuedItem(Request $request)`

## hms/app/Http/Controllers/ItemCategoryController.php

Models: `Item` → `items`, `ItemCategory` → `item_categories`

### __construct (line 18)

`public function __construct(ItemCategoryRepository $itemCategoryRepo)`

### index (line 23)

`public function index()`

Views: `item_categories.index`

### store (line 28)

`public function store(CreateItemCategoryRequest $request)`

Write/transaction candidates:

```php
31: $this->itemCategoryRepository->create($input);
```

### edit (line 36)

`public function edit(ItemCategory $itemCategory)`

### update (line 41)

`public function update(ItemCategory $itemCategory, UpdateItemCategoryRequest $request)`

Write/transaction candidates:

```php
44: $this->itemCategoryRepository->update($input, $itemCategory->id);
```

### destroy (line 49)

`public function destroy(ItemCategory $itemCategory)`

Guard/validation candidates:

```php
53: if ($result) {
```

Write/transaction candidates:

```php
56: $this->itemCategoryRepository->delete($itemCategory->id);
```

### getItemsList (line 61)

`public function getItemsList(Request $request)`

Guard/validation candidates:

```php
63: if (empty($request->get('id'))) {
```

## hms/app/Http/Controllers/ItemController.php

Models: `IssuedItem` → `issued_items`, `Item` → `items`, `ItemStock` → `item_stocks`

### __construct (line 19)

`public function __construct(ItemRepository $itemRepo)`

### index (line 24)

`public function index()`

Views: `items.index`

### create (line 29)

`public function create()`

Views: `items.create`

### store (line 36)

`public function store(CreateItemRequest $request)`

Write/transaction candidates:

```php
40: $this->itemRepository->create($input);
```

### show (line 46)

`public function show(Item $item)`

Views: `items.show`

### edit (line 51)

`public function edit(Item $item)`

Views: `items.edit`

### update (line 58)

`public function update(Item $item, UpdateItemRequest $request)`

Write/transaction candidates:

```php
62: $this->itemRepository->update($input, $item->id);
```

### destroy (line 68)

`public function destroy(Item $item)`

Guard/validation candidates:

```php
74: if ($result) {
```

Write/transaction candidates:

```php
77: $item->delete();
```

### getAvailableQuantity (line 82)

`public function getAvailableQuantity(Request $request)`

## hms/app/Http/Controllers/ItemStockController.php

Models: `ItemStock` → `item_stocks`

### __construct (line 16)

`public function __construct(ItemStockRepository $itemStockRepo)`

### index (line 21)

`public function index()`

Views: `item_stocks.index`

### create (line 26)

`public function create()`

Views: `item_stocks.create`

### store (line 34)

`public function store(CreateItemStockRequest $request)`

Write/transaction candidates:

```php
38: $this->itemStockRepository->store($input);
```

### show (line 44)

`public function show(ItemStock $itemStock)`

Views: `item_stocks.show`

### edit (line 49)

`public function edit(ItemStock $itemStock)`

Views: `item_stocks.edit`

### update (line 57)

`public function update(ItemStock $itemStock, UpdateItemStockRequest $request)`

Write/transaction candidates:

```php
61: $this->itemStockRepository->update($itemStock, $input);
```

### destroy (line 67)

`public function destroy(ItemStock $itemStock)`

### downloadMedia (line 74)

`public function downloadMedia(ItemStock $itemStock)`

## hms/app/Http/Controllers/LabTechnicianController.php

Models: `EmployeePayroll` → `employee_payrolls`, `LabTechnician` → `lab_technicians`

### __construct (line 19)

`public function __construct(LabTechnicianRepository $labTechnicianRepo)`

### index (line 24)

`public function index()`

Views: `lab_technicians.index`

### create (line 31)

`public function create()`

Views: `lab_technicians.create`

### store (line 38)

`public function store(CreateLabTechnicianRequest $request)`

Write/transaction candidates:

```php
42: $labTechnician = $this->labTechnicianRepository->store($input);
```

### show (line 49)

`public function show(LabTechnician $labTechnician)`

Views: `lab_technicians.show`

### edit (line 56)

`public function edit(LabTechnician $labTechnician)`

Views: `lab_technicians.edit`

### update (line 64)

`public function update(LabTechnician $labTechnician, UpdateLabTechnicianRequest $request)`

Write/transaction candidates:

```php
66: $labTechnician = $this->labTechnicianRepository->update($labTechnician, $request->all());
```

### destroy (line 73)

`public function destroy(LabTechnician $labTechnician)`

Guard/validation candidates:

```php
76: if ($empPayRollResult) {
```

Write/transaction candidates:

```php
79: $labTechnician->user()->delete();
80: $labTechnician->address()->delete();
81: $labTechnician->delete();
```

### activeDeactiveStatus (line 86)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
90: $labTechnician->user()->update(['status' => $status]);
```

### labTechnicianExport (line 95)

`public function labTechnicianExport()`

Guard/validation candidates:

```php
98: if (!$labTechnicians) {
```

## hms/app/Http/Controllers/LiveConsultationController.php

Models: `LiveConsultation` → `live_consultations`, `UserGoogleEventSchedule` → `user_google_event_schedules`, `UserZoomCredential` → `user_zoom_credential`

### __construct (line 32)

`public function __construct( LiveConsultationRepository $liveConsultationRepository, PatientCaseRepository $patientCaseRepository, ZoomRepository $zoomRepository )`

### index (line 42)

`public function index()`

Views: `live_consultations.index`

### store (line 52)

`public function store(LiveConsultationRequest $request)`

Guard/validation candidates:

```php
58: if($request->platform_type == LiveConsultation::GOOGLE_MEET){
80: if (isset($responseData['error'])) {
83: if($errorCode == 401){
```

Write/transaction candidates:

```php
56: DB::beginTransaction();
66: $this->liveConsultationRepository->store($request->all());
71: DB::commit();
76: DB::rollBack();
```

### edit (line 92)

`public function edit(LiveConsultation $liveConsultation)`

Guard/validation candidates:

```php
94: if (checkRecordAccess($liveConsultation->doctor_id)) {
```

### update (line 101)

`public function update(LiveConsultationRequest $request, LiveConsultation $liveConsultation)`

### destroy (line 112)

`public function destroy(LiveConsultation $liveConsultation)`

Guard/validation candidates:

```php
115: if (checkRecordAccess($liveConsultation->doctor_id)) {
119: if($liveConsultation->platform_type == LiveConsultation::GOOGLE_MEET){
```

Write/transaction candidates:

```php
121: $userGoogleEventCalendar->delete();
122: $liveConsultation->delete();
127: $liveConsultation->delete();
```

### getTypeNumber (line 136)

`public function getTypeNumber(Request $request)`

### getChangeStatus (line 147)

`public function getChangeStatus(Request $request)`

Guard/validation candidates:

```php
151: if (checkRecordAccess($liveConsultation->doctor_id)) {
156: if ($request->get('statusId') == LiveConsultation::STATUS_AWAITED) {
```

Write/transaction candidates:

```php
164: $liveConsultation->update([
```

### getLiveStatus (line 172)

`public function getLiveStatus(LiveConsultation $liveConsultation)`

Guard/validation candidates:

```php
174: if (getLoggedinDoctor() ? checkRecordAccess($liveConsultation->doctor_id) : checkRecordAccess($liveConsultation->patient_id)) {
```

### show (line 185)

`public function show(LiveConsultation $liveConsultation)`

Guard/validation candidates:

```php
187: if (getLoggedinDoctor() ? checkRecordAccess($liveConsultation->doctor_id) : checkRecordAccess($liveConsultation->patient_id)) {
```

### zoomCredential (line 199)

`public function zoomCredential($id)`

### zoomCredentialCreate (line 210)

`public function zoomCredentialCreate(CreateZoomCredentialRequest $request)`

### zoomConnect (line 221)

`public function zoomConnect(Request $request)`

Guard/validation candidates:

```php
226: if ($userZoomCredential == null) {
```

### zoomCallback (line 240)

`public function zoomCallback(Request $request)`

## hms/app/Http/Controllers/LiveMeetingController.php

Models: `LiveMeeting` → `live_meetings`

### __construct (line 21)

`public function __construct(LiveMeetingRepository $liveMeetingRepository, ZoomRepository $zoomRepository)`

### index (line 27)

`public function index()`

Views: `live_consultations.member_index`

### liveMeetingStore (line 35)

`public function liveMeetingStore(LiveMeetingRequest $request)`

Write/transaction candidates:

```php
38: $this->liveMeetingRepository->store($request->all());
```

### getChangeStatus (line 47)

`public function getChangeStatus(Request $request)`

Guard/validation candidates:

```php
52: if ($request->get('statusId') == LiveMeeting::STATUS_AWAITED) {
```

Write/transaction candidates:

```php
60: $liveMeeting->update([
```

### getLiveStatus (line 67)

`public function getLiveStatus(LiveMeeting $liveMeeting)`

Guard/validation candidates:

```php
69: if (getLoggedInUser()->hasRole('Admin') || collect($liveMeeting->members)->contains('id', getLoggedInUser()->id)) {
```

### edit (line 81)

`public function edit(LiveMeeting $liveMeeting)`

Guard/validation candidates:

```php
83: if (getLoggedInUser()->hasRole('Admin') || collect($liveMeeting->members)->contains('id', getLoggedInUser()->id)) {
```

### update (line 94)

`public function update(LiveMeetingRequest $request, LiveMeeting $liveMeeting)`

### show (line 105)

`public function show(LiveMeeting $liveMeeting)`

Guard/validation candidates:

```php
107: if (getLoggedInUser()->hasRole('Admin') || collect($liveMeeting->members)->contains('id', getLoggedInUser()->id)) {
```

### destroy (line 116)

`public function destroy(LiveMeeting $liveMeeting)`

Guard/validation candidates:

```php
119: if (getLoggedInUser()->hasRole('Admin') || collect($liveMeeting->members)->contains('id', getLoggedInUser()->id)) {
```

Write/transaction candidates:

```php
121: $liveMeeting->delete();
```

## hms/app/Http/Controllers/LunchBreakController.php

Models: `Doctor` → `doctors`, `LunchBreak` → `lunch_breaks`, `User` → `users`, `Appointment` → `appointments`

### __construct (line 18)

`public function __construct(LunchBreakRepository $lunchBreakRepository)`

### index (line 23)

`public function index()`

Views: `lunch_breaks.index`

### create (line 28)

`public function create()`

Views: `lunch_breaks.create`

### store (line 35)

`public function store(CreateLunchBreakRequest $request)`

Guard/validation candidates:

```php
39: if(isset($input['date']) && !empty($input['date'])){
42: if($opdDates){
51: if ($lunchBreak) {
```

Write/transaction candidates:

```php
49: $lunchBreak = $this->lunchBreakRepository->store($input);
```

### destroy (line 62)

`public function destroy($id)`

Write/transaction candidates:

```php
64: $checkRecord = LunchBreak::destroy($id);
```

## hms/app/Http/Controllers/MailController.php

Models:

### __construct (line 14)

`public function __construct(MailRepository $mailRepo)`

### index (line 19)

`public function index()`

Views: `mail.index`

### store (line 24)

`public function store(CreateMailRequest $request)`

Write/transaction candidates:

```php
27: $this->mailRepository->store($input);
```

## hms/app/Http/Controllers/ManualBillPaymentController.php

Models: `Bill` → `bills`, `ManualBillPayment` → `bill_transactions`

### __construct (line 22)

`public function __construct(ManualBillPaymentRepository $manualBillPaymentRepository)`

### index (line 30)

`public function index()`

Views: `manual_bill_payments.index`

### create (line 39)

`public function create()`

### store (line 47)

`public function store(Request $request)`

Guard/validation candidates:

```php
51: if ($request->payment_type == Bill::Stripe) {
69: if (!in_array(strtoupper(getCurrentCurrency()), getFlutterWaveSupportedCurrencies())) {
76: if (!$flutterwavePublicKey && !$flutterwaveSecretKey) {
90: if (strtoupper(getCurrentCurrency()) != 'INR') {
```

Write/transaction candidates:

```php
104: $this->manualBillPaymentRepository->create($input);
```

### show (line 113)

`public function show(string $id)`

### edit (line 121)

`public function edit(string $id)`

### update (line 129)

`public function update(Request $request, $id)`

### destroy (line 141)

`public function destroy(string $id)`

### paymentSuccess (line 146)

`public function paymentSuccess(Request $request)`

### onBoard (line 157)

`public function onBoard(Request $request)`

### razorpayPaymentSuccess (line 166)

`public function razorpayPaymentSuccess(Request $request)`

### flutterwavePaymentSuccess (line 177)

`public function flutterwavePaymentSuccess(Request $request)`

Guard/validation candidates:

```php
182: if (!$flutterwavePublicKey && !$flutterwaveSecretKey) {
191: if ($request['status'] == 'cancelled') {
```

### billPhonePePaymentSuccess (line 202)

`public function billPhonePePaymentSuccess(Request $request)`

### ManualPaystackOnBoard (line 212)

`public function ManualPaystackOnBoard(Request $request)`

Guard/validation candidates:

```php
214: if (!in_array(strtoupper(getCurrentCurrency()), getPayStackSupportedCurrencies())) {
```

### ManualPaystackSuccess (line 245)

`public function ManualPaystackSuccess(Request $request)`

## hms/app/Http/Controllers/MedicineBillController.php

Models: `Category` → `categories`, `Medicine` → `medicines`, `MedicineBill` → `medicine_bills`, `SaleMedicine` → `sale_medicines`

### __construct (line 38)

`public function __construct( PrescriptionRepository $prescriptionRepo, DoctorRepository $doctorRepository, MedicineRepository $medicineRepository, PatientRepository $patientRepo, MedicineBillRepository $medicineBillRepository, )`

### index (line 57)

`public function index()`

Views: `medicine-bills.index`

### create (line 62)

`public function create()`

Views: `medicine-bills.create`

### setFlutterWaveConfig (line 79)

`public function setFlutterWaveConfig()`

Guard/validation candidates:

```php
84: if(!$flutterwavePublicKey && !$flutterwaveSecretKey){
```

### store (line 94)

`public function store(CreateMedicineBillRequest $request)`

Guard/validation candidates:

```php
98: if($input['payment_type'] == MedicineBill::MEDICINE_BILL_STRIPE){
113: if (strtoupper(getCurrentCurrency()) != 'INR') {
127: if(!in_array(strtoupper(getCurrentCurrency()),getFlutterWaveSupportedCurrencies())){
```

### show (line 146)

`public function show(MedicineBill $medicineBill)`

Views: `medicine-bills.show`

### edit (line 153)

`public function edit(MedicineBill $medicineBill)`

Views: `medicine-bills.edit`

### update (line 172)

`public function update(MedicineBill $medicineBill, UpdateMedicineBillRequest $request)`

Guard/validation candidates:

```php
176: if (empty($input['medicine']) && $input['payment_status'] == false) {
```

Write/transaction candidates:

```php
181: $this->medicineBillRepository->update($medicineBill, $input);
```

### destroy (line 187)

`public function destroy(MedicineBill $medicineBill)`

Write/transaction candidates:

```php
189: $medicineBill->saleMedicine()->delete();
190: $medicineBill->delete();
```

### storePatient (line 195)

`public function storePatient(CreatePatientRequest $request)`

Write/transaction candidates:

```php
200: $this->patientRepository->store($input);
```

### convertToPDF (line 207)

`public function convertToPDF($id)`

### getMedicineCategory (line 217)

`public function getMedicineCategory(Category $category)`

### stripeSuccess (line 226)

`public function stripeSuccess(Request $request)`

### stripeFailed (line 235)

`public function stripeFailed(Request $request)`

### razorPayPayment (line 244)

`public function razorPayPayment(Request $request)`

### razorPayPaymentSuccess (line 251)

`public function razorPayPaymentSuccess(Request $request)`

### razorPayPaymentFailed (line 260)

`public function razorPayPaymentFailed(Request $request)`

Guard/validation candidates:

```php
268: if ($input['category_id']) {
276: if ($input['payment_status'] == 1) {
```

Write/transaction candidates:

```php
274: $saleMedicine->delete();
277: $medicine->update([
284: $medicineBill->delete();
```

### phonePePaymentSuccess (line 289)

`public function phonePePaymentSuccess(Request $request)`

### paystackPayment (line 298)

`public function paystackPayment(Request $request)`

Guard/validation candidates:

```php
300: if(!in_array(strtoupper(getCurrentCurrency()),getPayStackSupportedCurrencies())){
```

### flutterwavePaymentSuccess (line 329)

`public function flutterwavePaymentSuccess(Request $request)`

Guard/validation candidates:

```php
333: if($request['status'] == 'cancelled'){
```

## hms/app/Http/Controllers/MedicineController.php

Models: `Medicine` → `medicines`, `PurchasedMedicine` → `purchased_medicines`, `SaleMedicine` → `sale_medicines`

### __construct (line 20)

`public function __construct(MedicineRepository $medicineRepo)`

### index (line 25)

`public function index()`

Views: `medicines.index`

### create (line 30)

`public function create()`

Views: `medicines.create`

### store (line 37)

`public function store(CreateMedicineRequest $request)`

Write/transaction candidates:

```php
41: $this->medicineRepository->create($input);
```

### show (line 48)

`public function show(Medicine $medicine)`

Views: `medicines.show`

### edit (line 56)

`public function edit(Medicine $medicine)`

Views: `medicines.edit`

### update (line 64)

`public function update(Medicine $medicine, UpdateMedicineRequest $request)`

Write/transaction candidates:

```php
66: $this->medicineRepository->update($request->all(), $medicine->id);
```

### destroy (line 73)

`public function destroy(Medicine $medicine)`

Guard/validation candidates:

```php
76: if (! canAccessRecord(Medicine::class, $medicine->id)) {
83: if (isset($purchaseMedicine) && ! empty($purchaseMedicine)) {
86: if (isset($saleMedicine) && ! empty($saleMedicine)) {
```

Write/transaction candidates:

```php
84: $purchaseMedicine->map->delete();
87: $saleMedicine->map->delete();
90: $this->medicineRepository->delete($medicine->id);
```

### medicineExport (line 95)

`public function medicineExport()`

Guard/validation candidates:

```php
98: if (!$medicines) {
```

### showModal (line 105)

`public function showModal(Medicine $medicine)`

### checkUseOfMedicine (line 128)

`public function checkUseOfMedicine(Medicine $medicine)`

Guard/validation candidates:

```php
139: if ($result) {
```

## hms/app/Http/Controllers/NoticeBoardController.php

Models: `NoticeBoard` → `notice_boards`

### __construct (line 15)

`public function __construct(NoticeBoardRepository $noticeBoardRepo)`

### index (line 20)

`public function index()`

Views: `notice_boards.index`

### store (line 25)

`public function store(CreateNoticeBoardRequest $request)`

Write/transaction candidates:

```php
28: $this->noticeBoardRepository->create($input);
```

### show (line 34)

`public function show($id)`

### edit (line 41)

`public function edit(NoticeBoard $noticeBoard)`

### update (line 46)

`public function update(NoticeBoard $noticeBoard, UpdateNoticeBoardRequest $request)`

Write/transaction candidates:

```php
48: $this->noticeBoardRepository->update($request->all(), $noticeBoard->id);
```

### destroy (line 53)

`public function destroy(NoticeBoard $noticeBoard)`

Write/transaction candidates:

```php
55: $noticeBoard->delete();
```

## hms/app/Http/Controllers/NotificationController.php

Models: `Notification` → `notifications`

### readNotification (line 10)

`public function readNotification(Notification $notification)`

Write/transaction candidates:

```php
13: $notification->save();
```

### readAllNotification (line 18)

`public function readAllNotification()`

Write/transaction candidates:

```php
20: Notification::whereReadAt(null)->where('user_id', getLoggedInUserId())->update(['read_at' => Carbon::now()]);
```

## hms/app/Http/Controllers/NurseController.php

Models: `EmployeePayroll` → `employee_payrolls`, `Nurse` → `nurses`

### __construct (line 19)

`public function __construct(NurseRepository $nurseRepo)`

### index (line 24)

`public function index()`

Views: `nurses.index`

### create (line 31)

`public function create()`

Views: `nurses.create`

### store (line 38)

`public function store(CreateNurseRequest $request)`

Write/transaction candidates:

```php
43: $nurse = $this->nurseRepository->store($input);
```

### show (line 50)

`public function show(Nurse $nurse)`

Views: `nurses.show`

### edit (line 57)

`public function edit(Nurse $nurse)`

Views: `nurses.edit`

### update (line 65)

`public function update(Nurse $nurse, UpdateNurseRequest $request)`

Write/transaction candidates:

```php
70: $user = $this->nurseRepository->update($nurse, $input);
```

### destroy (line 77)

`public function destroy(Nurse $nurse)`

Guard/validation candidates:

```php
80: if ($empPayRollResult) {
```

Write/transaction candidates:

```php
83: $nurse->user()->delete();
84: $nurse->address()->delete();
85: $nurse->delete();
```

### activeDeactiveStatus (line 90)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
94: $nurse->user()->update(['status' => $status]);
```

### nurseExport (line 99)

`public function nurseExport()`

Guard/validation candidates:

```php
102: if (!$nurses) {
```

## hms/app/Http/Controllers/OdontogramController.php

Models: `Odontogram` → `odontograms`, `Setting` → `settings`

### __construct (line 19)

`public function __construct(OdontogramRepository $odontogramRepo)`

### index (line 24)

`public function index()`

Views: `odontogram.index`

### store (line 32)

`public function store(CreateOdontogramRequest $request)`

Write/transaction candidates:

```php
35: $this->odontogramRepository->store($input);
```

### edit (line 40)

`public function edit(Odontogram $odontogram)`

### update (line 45)

`public function update(UpdateOdontogramRequest $request, Odontogram $odontogram)`

### destroy (line 52)

`public function destroy(Odontogram $odontogram)`

Write/transaction candidates:

```php
55: $odontogram->delete();
```

### convertToPdf (line 59)

`public function convertToPdf(Odontogram $odontogram)`

## hms/app/Http/Controllers/OpdDiagnosisController.php

Models: `OpdDiagnosis` → `opd_diagnoses`

### __construct (line 18)

`public function __construct(OpdDiagnosisRepository $opdDiagnosisRepo)`

### index (line 23)

`public function index(Request $request)`

Guard/validation candidates:

```php
25: if ($request->ajax()) {
```

### store (line 30)

`public function store(CreateOpdDiagnosisRequest $request)`

Write/transaction candidates:

```php
33: $this->opdDiagnosisRepository->store($input);
```

### edit (line 39)

`public function edit(OpdDiagnosis $opdDiagnosis)`

### update (line 44)

`public function update(OpdDiagnosis $opdDiagnosis, UpdateOpdDiagnosisRequest $request)`

### destroy (line 51)

`public function destroy(OpdDiagnosis $opdDiagnosis)`

### downloadMedia (line 58)

`public function downloadMedia(OpdDiagnosis $opdDiagnosis)`

Guard/validation candidates:

```php
61: if ($media) {
```

## hms/app/Http/Controllers/OpdPatientDepartmentController.php

Models: `AddCustomFields` → `add_custom_fields`, `DoctorOPDCharge` → `doctor_opd_charges`, `OpdPatientDepartment` → `opd_patient_departments`

### __construct (line 19)

`public function __construct(OpdPatientDepartmentRepository $opdPatientDepartmentRepo)`

### index (line 24)

`public function index()`

Views: `opd_patient_departments.index`

### create (line 29)

`public function create(Request $request)`

Guard/validation candidates:

```php
34: if ($data['revisit']) {
```

Views: `opd_patient_departments.create`

### store (line 42)

`public function store(CreateOpdPatientDepartmentRequest $request)`

Write/transaction candidates:

```php
47: $this->opdPatientDepartmentRepository->store($input);
```

### show (line 54)

`public function show(OpdPatientDepartment $opdPatientDepartment)`

Views: `opd_patient_departments.show`

### edit (line 66)

`public function edit(OpdPatientDepartment $opdPatientDepartment)`

Views: `opd_patient_departments.edit`

### update (line 74)

`public function update(OpdPatientDepartment $opdPatientDepartment, UpdateOpdPatientDepartmentRequest $request)`

### destroy (line 83)

`public function destroy($id)`

Write/transaction candidates:

```php
86: $opdPatientDepartment->delete();
```

### getPatientCasesList (line 91)

`public function getPatientCasesList(Request $request)`

### getDoctorOPDCharge (line 98)

`public function getDoctorOPDCharge(Request $request)`

## hms/app/Http/Controllers/OpdPrescriptionController.php

Models: `Medicine` → `medicines`, `OpdPrescription` → `opd_prescriptions`

### __construct (line 21)

`public function __construct(OpdPresciptionRepository $opdPrescriptionRepo)`

### index (line 26)

`public function index(Request $request)`

Guard/validation candidates:

```php
28: if ($request->ajax()) {
```

### store (line 33)

`public function store(CreateOpdPrescriptionRequest $request)`

Guard/validation candidates:

```php
42: if (! empty($duplicateIds)) {
50: if ($medicine->available_quantity < $qty) {
```

Write/transaction candidates:

```php
58: $this->opdPrescriptionRepository->store($input);
```

### edit (line 64)

`public function edit(OpdPrescription $opdPrescription)`

### update (line 72)

`public function update(OpdPrescription $opdPrescription, UpdateOpdPrescriptionRequest $request)`

Guard/validation candidates:

```php
98: if (! empty($duplicateIds)) {
106: if ($medicine->available_quantity < $qty && ! array_key_exists($input['medicine_id'][$key], $result)) {
```

### show (line 118)

`public function show(OpdPrescription $opdPrescription)`

Views: `opd_prescriptions.show_opd_prescription_data`

### destroy (line 123)

`public function destroy(OpdPrescription $opdPrescription)`

Write/transaction candidates:

```php
125: $opdPrescription->opdPrescriptionItems()->delete();
126: $opdPrescription->delete();
```

### getMedicineList (line 131)

`public function getMedicineList(Request $request)`

### getAvailableMedicineQuantity (line 138)

`public function getAvailableMedicineQuantity(Medicine $medicine)`

### convertToPDF (line 143)

`public function convertToPDF(OpdPrescription $opdPrescription)`

## hms/app/Http/Controllers/OpdTimelineController.php

Models: `OpdTimeline` → `opd_timelines`

### __construct (line 16)

`public function __construct(OpdTimelineRepository $opdTimelineRepo)`

### index (line 21)

`public function index(Request $request)`

Views: `opd_timelines.index`

### store (line 28)

`public function store(CreateOpdTimelineRequest $request)`

Write/transaction candidates:

```php
31: $this->opdTimelineRepository->store($input);
```

### edit (line 36)

`public function edit(OpdTimeline $opdTimeline)`

### update (line 41)

`public function update(OpdTimeline $opdTimeline, UpdateOpdTimelineRequest $request)`

### destroy (line 48)

`public function destroy(OpdTimeline $opdTimeline)`

### downloadMedia (line 55)

`public function downloadMedia(OpdTimeline $opdTimeline)`

Guard/validation candidates:

```php
59: if ($media) {
```

## hms/app/Http/Controllers/OperationCategoryController.php

Models: `Operation` → `operations`, `OperationCategory` → `operation_categories`

### __construct (line 16)

`public function __construct(OperationCategoryRepository $categoryRepository)`

### index (line 21)

`public function index()`

Views: `operation_categories.index`

### store (line 26)

`public function store(StoreOperationCategoryRequest $request)`

Write/transaction candidates:

```php
29: $this->categoryRepository->create($input);
```

### edit (line 34)

`public function edit(OperationCategory $operationCategory)`

### update (line 39)

`public function update(UpdateOperationCategoryRequest $request, OperationCategory $operationCategory)`

Write/transaction candidates:

```php
42: $this->categoryRepository->update($input, $operationCategory->id);
```

### destroy (line 47)

`public function destroy(OperationCategory $operationCategory)`

Guard/validation candidates:

```php
55: if ($result) {
```

Write/transaction candidates:

```php
59: $operationCategory->delete();
```

### getOperationName (line 64)

`public function getOperationName(Request $request)`

## hms/app/Http/Controllers/OperationController.php

Models: `Operation` → `operations`, `OperationCategory` → `operation_categories`

### index (line 12)

`public function index()`

Views: `operations.index`

### store (line 19)

`public function store(CreateOperationRequest $request)`

Write/transaction candidates:

```php
23: Operation::create($input);
```

### edit (line 28)

`public function edit($id)`

### update (line 35)

`public function update(UpdateOperationRequest $request, $id)`

Write/transaction candidates:

```php
39: $operation->update($request->all());
```

### delete (line 44)

`public function delete($id)`

Write/transaction candidates:

```php
46: Operation::where('id', $id)->delete();
```

## hms/app/Http/Controllers/OperationReportController.php

Models: `OperationReport` → `operation_reports`, `PatientCase` → `patient_cases`

### __construct (line 17)

`public function __construct(OperationReportRepository $operationReportRepo)`

### index (line 22)

`public function index()`

Views: `operation_reports.index`

### store (line 30)

`public function store(CreateOperationReportRequest $request)`

Guard/validation candidates:

```php
37: if (! empty($birthDate) && $operationDate < $birthDate) {
```

Write/transaction candidates:

```php
41: $this->operationReportRepository->store($input);
```

### show (line 46)

`public function show(OperationReport $operationReport)`

Views: `operation_reports.show`

### edit (line 56)

`public function edit(OperationReport $operationReport)`

Guard/validation candidates:

```php
58: if (getLoggedinDoctor() && checkRecordAccess($operationReport->doctor_id)) {
```

### update (line 65)

`public function update(OperationReport $operationReport, UpdateOperationReportRequest $request)`

Guard/validation candidates:

```php
72: if (! empty($birthDate) && $operationDate < $birthDate) {
```

Write/transaction candidates:

```php
76: $this->operationReportRepository->update($input, $operationReport);
```

### destroy (line 81)

`public function destroy(OperationReport $operationReport)`

Write/transaction candidates:

```php
83: $operationReport->delete();
```

## hms/app/Http/Controllers/PackageController.php

Models: `Package` → `packages`, `PatientAdmission` → `patient_admissions`

### __construct (line 21)

`public function __construct(PackageRepository $packageRepo)`

### index (line 26)

`public function index()`

Views: `packages.index`

### create (line 31)

`public function create()`

Views: `packages.create`

### store (line 39)

`public function store(CreatePackageRequest $request)`

Write/transaction candidates:

```php
42: DB::beginTransaction();
43: $package = $this->packageRepository->store($request->all());
44: DB::commit();
46: DB::rollBack();
```

### show (line 54)

`public function show(Package $package)`

Views: `packages.show`

### edit (line 61)

`public function edit(Package $package)`

Views: `packages.edit`

### update (line 70)

`public function update(Package $package, UpdatePackageRequest $request)`

Write/transaction candidates:

```php
73: DB::beginTransaction();
75: DB::commit();
77: DB::rollBack();
```

### destroy (line 85)

`public function destroy(Package $package)`

Guard/validation candidates:

```php
92: if ($result) {
```

Write/transaction candidates:

```php
96: $this->packageRepository->delete($package->id);
```

### packageExport (line 101)

`public function packageExport()`

Guard/validation candidates:

```php
104: if (!$packages) {
```

## hms/app/Http/Controllers/PathologyCategoryController.php

Models: `PathologyCategory` → `pathology_categories`, `PathologyTest` → `pathology_tests`

### __construct (line 16)

`public function __construct(PathologyCategoryRepository $pathologyCategoryRepo)`

### index (line 21)

`public function index()`

Views: `pathology_categories.index`

### store (line 26)

`public function store(CreatePathologyCategoryRequest $request)`

Write/transaction candidates:

```php
29: $this->pathologyCategoryRepository->create($input);
```

### edit (line 34)

`public function edit(PathologyCategory $pathologyCategory)`

### update (line 39)

`public function update(PathologyCategory $pathologyCategory, UpdatePathologyCategoryRequest $request)`

Write/transaction candidates:

```php
42: $this->pathologyCategoryRepository->update($input, $pathologyCategory->id);
```

### destroy (line 47)

`public function destroy(PathologyCategory $pathologyCategory)`

Guard/validation candidates:

```php
54: if ($result) {
```

Write/transaction candidates:

```php
58: $pathologyCategory->delete();
```

## hms/app/Http/Controllers/PathologyParameterController.php

Models: `PathologyParameter` → `pathology_parameters`, `PathologyParameterItem` → `pathology_parameter_items`

### __construct (line 17)

`public function __construct(PathologyParameterRepository $pathologyParameterRepo)`

### index (line 22)

`public function index(Request $request)`

Views: `pathology_parameter.index`

### store (line 29)

`public function store(CreatePathologyParameterRequest $request)`

Write/transaction candidates:

```php
32: $this->pathologyParameterRepository->create($input);
```

### edit (line 37)

`public function edit(PathologyParameter $pathologyParameter)`

Guard/validation candidates:

```php
39: if (! canAccessRecord(PathologyParameter::class, $pathologyParameter->id)) {
```

### update (line 46)

`public function update(PathologyParameter $pathologyParameter, UpdatePathologyParameterRequest $request)`

Write/transaction candidates:

```php
49: $this->pathologyParameterRepository->update($input, $pathologyParameter->id);
```

### destroy (line 54)

`public function destroy(PathologyParameter $pathologyParameter)`

Guard/validation candidates:

```php
56: if (! canAccessRecord(PathologyParameter::class, $pathologyParameter->id)) {
65: if ($result) {
```

Write/transaction candidates:

```php
69: $pathologyParameter->delete();
```

## hms/app/Http/Controllers/PathologyTestController.php

Models: `Charge` → `charges`, `PathologyParameter` → `pathology_parameters`, `PathologyParameterItem` → `pathology_parameter_items`, `PathologyTest` → `pathology_tests`

### __construct (line 26)

`public function __construct(PathologyTestRepository $pathologyTestRepo, PatientRepository $patientRepository)`

### index (line 32)

`public function index()`

Views: `pathology_tests.index`

### create (line 37)

`public function create()`

Views: `pathology_tests.create`

### store (line 46)

`public function store(CreatePathologyTestRequest $request)`

Guard/validation candidates:

```php
55: if ($input['parameter_id']) {
57: if ($input['parameter_id'][$key] == null) {
62: if ($input['patient_result'][$key] == null) {
```

Write/transaction candidates:

```php
69: $this->pathologyTestRepository->store($input);
```

### show (line 76)

`public function show(PathologyTest $pathologyTest)`

Views: `pathology_tests.show`

### edit (line 83)

`public function edit(PathologyTest $pathologyTest)`

Views: `pathology_tests.edit`

### update (line 94)

`public function update(PathologyTest $pathologyTest, UpdatePathologyTestRequest $request)`

Guard/validation candidates:

```php
103: if ($input['parameter_id']) {
105: if ($input['parameter_id'][$key] == null) {
109: if ($input['patient_result'][$key] == null) {
```

Write/transaction candidates:

```php
116: $this->pathologyTestRepository->update($input, $pathologyTest);
```

### destroy (line 122)

`public function destroy(PathologyTest $pathologyTest)`

Write/transaction candidates:

```php
124: $pathologyTest->parameterItems()->delete();
125: $pathologyTest->delete();
```

### getStandardCharge (line 130)

`public function getStandardCharge($id)`

### pathologyTestExport (line 137)

`public function pathologyTestExport()`

Guard/validation candidates:

```php
140: if (!$pathologyTests) {
```

### showModal (line 147)

`public function showModal(PathologyTest $pathologyTest)`

### getPathologyParameter (line 170)

`public function getPathologyParameter($id)`

### convertToPDF (line 178)

`public function convertToPDF($id)`

## hms/app/Http/Controllers/PathologyUnitController.php

Models: `PathologyParameter` → `pathology_parameters`, `PathologyUnit` → `pathology_units`

### __construct (line 18)

`public function __construct(PathologyUnitRepository $pathologyUnitRepo)`

### index (line 23)

`public function index(Request $request)`

Views: `pathology_units.index`

### store (line 28)

`public function store(CreatePathologyUnitRequest $request)`

Write/transaction candidates:

```php
31: $this->pathologyUnitRepository->create($input);
```

### edit (line 36)

`public function edit(PathologyUnit $pathologyUnit)`

Guard/validation candidates:

```php
38: if (! canAccessRecord(PathologyUnit::class, $pathologyUnit->id)) {
```

### update (line 45)

`public function update(PathologyUnit $pathologyUnit, UpdatePathologyUnitRequest $request)`

Write/transaction candidates:

```php
48: $this->pathologyUnitRepository->update($input, $pathologyUnit->id);
```

### destroy (line 53)

`public function destroy(PathologyUnit $pathologyUnit)`

Guard/validation candidates:

```php
55: if (! canAccessRecord(PathologyUnit::class, $pathologyUnit->id)) {
64: if ($result) {
```

Write/transaction candidates:

```php
68: $pathologyUnit->delete();
```

## hms/app/Http/Controllers/Patient/IpdPatientDepartmentController.php

Models: `IpdPatientDepartment` → `ipd_patient_departments`, `IpdPayment` → `ipd_payments`

### __construct (line 20)

`public function __construct(IpdPatientDepartmentRepository $ipdPatientDepartmentRepo)`

### index (line 25)

`public function index(Request $request)`

Views: `ipd_patient_list.index`

### show (line 30)

`public function show(IpdPatientDepartment $ipdPatientDepartment)`

Guard/validation candidates:

```php
32: if (checkRecordAccess($ipdPatientDepartment->patient_id)) {
```

Views: `errors.404`, `ipd_patient_list.show`

## hms/app/Http/Controllers/Patient/OpdPatientDepartmentController.php

Models: `OpdPatientDepartment` → `opd_patient_departments`

### __construct (line 18)

`public function __construct(OpdPatientDepartmentRepository $opdPatientDepartmentRepo)`

### index (line 23)

`public function index()`

Views: `opd_patient_list.index`

### show (line 28)

`public function show(OpdPatientDepartment $opdPatientDepartment)`

Guard/validation candidates:

```php
30: if (checkRecordAccess($opdPatientDepartment->patient_id)) {
```

Views: `errors.404`, `opd_patient_list.show`

## hms/app/Http/Controllers/Patient/PatientCaseController.php

Models: `PatientCase` → `patient_cases`

### index (line 19)

`public function index(Request $request)`

Guard/validation candidates:

```php
21: if ($request->ajax()) {
```

Views: `patients_cases_list.index`

### show (line 33)

`public function show($id)`

Guard/validation candidates:

```php
38: if (empty($patientCase)) {
```

Views: `errors.404`, `patients_cases_list.show`

## hms/app/Http/Controllers/Patient/PrescriptionController.php

Models: `Prescription` → `prescriptions`

### index (line 18)

`public function index(Request $request)`

Views: `patients_prescription_list.index`

### show (line 25)

`public function show(int $id)`

Guard/validation candidates:

```php
29: if (checkRecordAccess($prescription->patient_id)) {
```

Views: `errors.404`, `patients_prescription_list.show`

### prescriptionExport (line 36)

`public function prescriptionExport()`

Guard/validation candidates:

```php
38: if (getLoggedInUser()->hasRole('Pharmacist')) {
40: if ($prescriptions->count() != 0) {
49: if ($prescriptions->count() == 0) {
```

## hms/app/Http/Controllers/Patient/VaccinatedController.php

Models: `OpdPatientDepartment` → `opd_patient_departments`

### __construct (line 20)

`public function __construct(OpdPatientDepartmentRepository $opdPatientDepartmentRepo)`

### index (line 25)

`public function index(Request $request)`

Guard/validation candidates:

```php
27: if ($request->ajax()) {
```

Views: `patient_vaccinated_list.index`

### show (line 34)

`public function show(OpdPatientDepartment $opdPatientDepartment)`

Views: `opd_patient_list.show`

## hms/app/Http/Controllers/PatientAdmissionController.php

Models: `Bill` → `bills`, `Patient` → `patients`, `PatientAdmission` → `patient_admissions`

### __construct (line 21)

`public function __construct(PatientAdmissionRepository $patientAdmissionRepo)`

### index (line 26)

`public function index()`

Views: `patient_admissions.index`

### create (line 33)

`public function create()`

Views: `patient_admissions.create`

### store (line 40)

`public function store(CreatePatientAdmissionRequest $request)`

Guard/validation candidates:

```php
47: if (! empty($birthDate) && $admissionDate < $birthDate) {
```

Write/transaction candidates:

```php
53: $this->patientAdmissionRepository->store($input);
```

### show (line 60)

`public function show(PatientAdmission $patientAdmission)`

Views: `patient_admissions.show`

### edit (line 65)

`public function edit(PatientAdmission $patientAdmission)`

Guard/validation candidates:

```php
67: if (getLoggedinDoctor()) {
68: if (getLoggedInUser()->owner_id == $patientAdmission->doctor_id) {
```

Views: `patient_admissions.edit`, `errors.404`, `patient_admissions.edit`

### update (line 93)

`public function update(PatientAdmission $patientAdmission, UpdatePatientAdmissionRequest $request)`

Guard/validation candidates:

```php
101: if (! empty($birthDate) && $admissionDate < $birthDate) {
```

Write/transaction candidates:

```php
107: $this->patientAdmissionRepository->update($input, $patientAdmission);
```

### destroy (line 114)

`public function destroy(PatientAdmission $patientAdmission)`

Guard/validation candidates:

```php
116: if (getLoggedinDoctor() && checkRecordAccess($patientAdmission->doctor_id)) {
127: if ($result) {
131: if (! empty($patientAdmission->bed_id)) {
```

Write/transaction candidates:

```php
134: $this->patientAdmissionRepository->delete($patientAdmission->id);
```

### activeDeactiveStatus (line 140)

`public function activeDeactiveStatus($id)`

Guard/validation candidates:

```php
144: if (! (getLoggedInUser()->hasRole('Receptionist') || getLoggedInUser()->hasRole('Case Manager')) && checkRecordAccess($patientAdmission->doctor_id)) {
```

Write/transaction candidates:

```php
148: $patientAdmission->update(['status' => $status]);
```

### patientAdmissionExport (line 154)

`public function patientAdmissionExport()`

Guard/validation candidates:

```php
157: if (getLoggedInUser()->hasRole('Doctor')) {
161: if ($patientAdmissions->isEmpty()) {
```

### showModal (line 168)

`public function showModal(PatientAdmission $patientAdmission)`

Guard/validation candidates:

```php
170: if (! (getLoggedInUser()->hasRole('Receptionist') || getLoggedInUser()->hasRole('Case Manager')) && checkRecordAccess($patientAdmission->doctor_id)) {
```

## hms/app/Http/Controllers/PatientCaseController.php

Models: `BedAssign` → `bed_assigns`, `BirthReport` → `birth_reports`, `DeathReport` → `death_reports`, `IpdPatientDepartment` → `ipd_patient_departments`, `OperationReport` → `operation_reports`, `Patient` → `patients`, `PatientCase` → `patient_cases`

### __construct (line 25)

`public function __construct(PatientCaseRepository $patientCaseManagerRepo)`

### index (line 30)

`public function index()`

Views: `patient_cases.index`

### create (line 37)

`public function create()`

Views: `patient_cases.create`

### store (line 45)

`public function store(CreatePatientCaseRequest $request)`

Guard/validation candidates:

```php
52: if (! empty($birthDate) && $caseDate < $birthDate) {
```

Write/transaction candidates:

```php
62: $this->patientCaseRepository->store($input);
```

### show (line 70)

`public function show(PatientCase $patientCase)`

Views: `patient_cases.show`

### edit (line 75)

`public function edit(PatientCase $patientCase)`

Views: `patient_cases.edit`

### update (line 83)

`public function update(PatientCase $patientCase, UpdatePatientCaseRequest $request)`

Guard/validation candidates:

```php
89: if (! empty($birthDate) && $caseDate < $birthDate) {
```

Write/transaction candidates:

```php
98: $patientCase = $this->patientCaseRepository->update($input, $patientCase->id);
```

### destroy (line 105)

`public function destroy(PatientCase $patientCase)`

Guard/validation candidates:

```php
116: if ($result) {
```

Write/transaction candidates:

```php
120: $this->patientCaseRepository->delete($patientCase->id);
```

### activeDeActiveStatus (line 125)

`public function activeDeActiveStatus($id)`

Write/transaction candidates:

```php
129: $patientCase->update(['status' => $patientCase->status]);
```

### patientCaseExport (line 134)

`public function patientCaseExport()`

Guard/validation candidates:

```php
137: if (!$patientCases) {
```

### showModal (line 144)

`public function showModal(PatientCase $patientCase)`

## hms/app/Http/Controllers/PatientController.php

Models: `AddCustomFields` → `add_custom_fields`, `AdvancedPayment` → `advanced_payments`, `Appointment` → `appointments`, `BedAssign` → `bed_assigns`, `Bill` → `bills`, `BirthReport` → `birth_reports`, `DeathReport` → `death_reports`, `InvestigationReport` → `investigation_reports`, `Invoice` → `invoices`, `IpdPatientDepartment` → `ipd_patient_departments`, `OperationReport` → `operation_reports`, `Patient` → `patients`, `PatientAdmission` → `patient_admissions`, `PatientCase` → `patient_cases`, `Prescription` → `prescriptions`, `Vaccination` → `vaccinations`

### __construct (line 36)

`public function __construct(PatientRepository $patientRepo)`

### index (line 41)

`public function index()`

Views: `patients.index`

### create (line 48)

`public function create()`

Views: `patients.create`

### store (line 56)

`public function store(CreatePatientRequest $request)`

Write/transaction candidates:

```php
60: $this->patientRepository->store($input);
```

### show (line 67)

`public function show($patientId)`

Guard/validation candidates:

```php
71: if (! $data) {
75: if (getLoggedinPatient() && checkRecordAccess($data->id)) {
81: if ($user->hasRole('Doctor')) {
```

Views: `errors.404`, `errors.404`, `patients.show`

### edit (line 93)

`public function edit(Patient $patient)`

Views: `patients.edit`

### update (line 101)

`public function update(Patient $patient, UpdatePatientRequest $request)`

Guard/validation candidates:

```php
103: if ($patient->is_default == 1) {
```

Write/transaction candidates:

```php
111: $this->patientRepository->update($input, $patient);
```

### destroy (line 118)

`public function destroy(Patient $patient)`

Guard/validation candidates:

```php
120: if ($patient->is_default == 1) {
141: if ($result) {
```

Write/transaction candidates:

```php
145: $patient->patientUser()->delete();
146: $patient->address()->delete();
147: $patient->delete();
```

### activeDeactiveStatus (line 152)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
156: $patient->patientUser()->update(['status' => $status]);
```

### patientExport (line 161)

`public function patientExport()`

Guard/validation candidates:

```php
164: if (!$patients) {
```

### getBirthDate (line 171)

`public function getBirthDate($id)`

## hms/app/Http/Controllers/PatientDiagnosisTestController.php

Models: `PatientDiagnosisProperty` → `patient_diagnosis_properties`, `PatientDiagnosisTest` → `patient_diagnosis_tests`

### __construct (line 34)

`public function __construct( PatientDiagnosisTestRepository $patientDiagnosisTestRepository, PatientRepository $patientRepository, DoctorRepository $doctorRepository )`

### index (line 44)

`public function index()`

Views: `patient_diagnosis_test.index`

### create (line 49)

`public function create()`

Views: `patient_diagnosis_test.create`

### store (line 62)

`public function store(CreatePatientDiagnosisTestRequest $request)`

Write/transaction candidates:

```php
65: $this->patientDiagnosisTestRepository->store($input);
```

### show (line 70)

`public function show(PatientDiagnosisTest $patientDiagnosisTest)`

Guard/validation candidates:

```php
72: if (getLoggedinDoctor() && checkRecordAccess($patientDiagnosisTest->doctor_id)) {
```

Views: `errors.404`, `patient_diagnosis_test.show`

### edit (line 81)

`public function edit(PatientDiagnosisTest $patientDiagnosisTest)`

Guard/validation candidates:

```php
83: if (getLoggedinDoctor() && checkRecordAccess($patientDiagnosisTest->doctor_id)) {
```

Views: `errors.404`, `patient_diagnosis_test.edit`

### update (line 98)

`public function update(UpdatePatientDiagnosisTestRequest $request, PatientDiagnosisTest $patientDiagnosisTest)`

### destroy (line 105)

`public function destroy(PatientDiagnosisTest $patientDiagnosisTest)`

Guard/validation candidates:

```php
107: if (getLoggedinDoctor() && checkRecordAccess($patientDiagnosisTest->doctor_id)) {
```

Write/transaction candidates:

```php
110: PatientDiagnosisProperty::wherePatientDiagnosisId($patientDiagnosisTest->id)->delete();
111: $this->patientDiagnosisTestRepository->delete($patientDiagnosisTest->id);
```

### convertToPdf (line 117)

`public function convertToPdf(PatientDiagnosisTest $patientDiagnosisTest)`

### patientDiagnosisTestExport (line 128)

`public function patientDiagnosisTestExport()`

Guard/validation candidates:

```php
131: if (!$patientDiagnosisTests) {
```

## hms/app/Http/Controllers/PatientIdCardTemplateController.php

Models: `PatientIdCardTemplate` → `patient_id_card_templates`, `Patient` → `patients`

### __construct (line 20)

`public function __construct(PatientIdCardTemplateRepository $PatientIdCardTemplateRepository)`

### index (line 25)

`public function index()`

Views: `patient_id_card_template.index`

### create (line 30)

`public function create()`

Views: `patient_id_card_template.create`

### store (line 35)

`public function store(CreatePatientIdCardTemplateRequest $request)`

Write/transaction candidates:

```php
38: $this->PatientIdCardTemplateRepository->create($input);
```

### edit (line 45)

`public function edit($id)`

Views: `patient_id_card_template.edit`

### update (line 52)

`public function update(UpdatePatientIdCardTemplateRequest $request, $id)`

Write/transaction candidates:

```php
54: $this->PatientIdCardTemplateRepository->update($id, $request->all());
```

### destroy (line 61)

`public function destroy($id)`

Write/transaction candidates:

```php
63: Patient::where('template_id',$id)->update(['template_id' => null]);
64: PatientIdCardTemplate::find($id)->delete();
```

### activeDeactiveStatus (line 69)

`public function activeDeactiveStatus(Request $request, $id)`

Guard/validation candidates:

```php
73: if(isset($request->color)){
```

Write/transaction candidates:

```php
74: $patientIdCardTemplateData->update(['color' => $request->color]);
76: $patientIdCardTemplateData->update([$request->name => $request->status]);
```

## hms/app/Http/Controllers/PatientQueueController.php

Models: `Appointment` → `appointments`, `PatientQueue` → `patient_queues`, `Setting` → `settings`, `User` → `users`

### index (line 14)

`public function index()`

Views: `patient_queues.index`

### show (line 19)

`public function show()`

Guard/validation candidates:

```php
25: if($setting != null && $setting->value == 1){
```

Views: `patient_queues.video-patient-queue-theme`, `patient_queues.patient-queue`

### changeStatus (line 32)

`public function changeStatus(Request $request)`

Write/transaction candidates:

```php
35: $appointment->update(['is_completed' => $request->status]);
43: PatientQueue::create($input);
```

### checkOutIn (line 48)

`public function checkOutIn(Request $request)`

Guard/validation candidates:

```php
53: if ($request->status == Appointment::STATUS_COMPLETED) {
```

Write/transaction candidates:

```php
54: $patientQueue->delete();
57: $appointment->update([
```

### refresh (line 64)

`public function refresh()`

Guard/validation candidates:

```php
73: if($setting != null && $setting->value == 1){
```

Views: `patient_queues.video_patient_queue_list`, `patient_queues.patient_queue_list`

## hms/app/Http/Controllers/PaymentController.php

Models: `Payment` → `payments`

### __construct (line 18)

`public function __construct(PaymentRepository $paymentRepo)`

### index (line 23)

`public function index()`

Views: `payments.index`

### create (line 28)

`public function create()`

Views: `payments.create`

### store (line 35)

`public function store(CreatePaymentRequest $request)`

Write/transaction candidates:

```php
39: $payment = $this->paymentRepository->create($input);
```

### show (line 46)

`public function show(Payment $payment)`

Views: `payments.show`

### edit (line 51)

`public function edit(Payment $payment)`

Views: `payments.edit`

### update (line 58)

`public function update(Payment $payment, UpdatePaymentRequest $request)`

Write/transaction candidates:

```php
62: $payment = $this->paymentRepository->update($input, $payment->id);
```

### destroy (line 69)

`public function destroy(Payment $payment)`

Write/transaction candidates:

```php
71: $this->paymentRepository->delete($payment->id);
```

### paymentExport (line 76)

`public function paymentExport()`

Guard/validation candidates:

```php
79: if (!$payments) {
```

### showModal (line 86)

`public function showModal(Payment $payment)`

## hms/app/Http/Controllers/PaymentGatewayController.php

Models: `Setting` → `settings`

### index (line 16)

`public function index()`

Views: `payment_gateway.index`

### create (line 26)

`public function create()`

### store (line 34)

`public function store(Request $request)`

### show (line 63)

`public function show(string $id)`

### edit (line 71)

`public function edit(string $id)`

### update (line 79)

`public function update(Request $request, string $id)`

### destroy (line 87)

`public function destroy(string $id)`

## hms/app/Http/Controllers/PaymentReportController.php

Models: `Account` → `accounts`, `Payment` → `payments`

### index (line 13)

`public function index()`

Views: `payment_reports.index`

### paymentReportExport (line 20)

`public function paymentReportExport()`

Guard/validation candidates:

```php
23: if (!$paymentReport) {
```

## hms/app/Http/Controllers/PharmacistController.php

Models: `EmployeePayroll` → `employee_payrolls`, `Pharmacist` → `pharmacists`

### __construct (line 19)

`public function __construct(PharmacistRepository $pharmacistRepo)`

### index (line 24)

`public function index()`

Views: `pharmacists.index`

### create (line 31)

`public function create()`

Views: `pharmacists.create`

### store (line 38)

`public function store(CreatePharmacistRequest $request)`

Write/transaction candidates:

```php
43: $this->pharmacistRepository->store($input);
```

### show (line 49)

`public function show(Pharmacist $pharmacist)`

Views: `pharmacists.show`

### edit (line 56)

`public function edit(Pharmacist $pharmacist)`

Views: `pharmacists.edit`

### update (line 64)

`public function update(Pharmacist $pharmacist, UpdatePharmacistRequest $request)`

Write/transaction candidates:

```php
68: $this->pharmacistRepository->update($input, $pharmacist);
```

### destroy (line 75)

`public function destroy(Pharmacist $pharmacist)`

Guard/validation candidates:

```php
79: if ($empPayRollResult) {
```

Write/transaction candidates:

```php
83: $pharmacist->user()->delete();
84: $pharmacist->delete();
85: $pharmacist->address()->delete();
```

### activeDeactiveStatus (line 90)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
94: $pharmacist->user()->update(['status' => $status]);
```

### pharmacistExport (line 99)

`public function pharmacistExport()`

Guard/validation candidates:

```php
102: if (!$pharmacists) {
```

## hms/app/Http/Controllers/PostalController.php

Models: `Postal` → `postals`

### __construct (line 20)

`public function __construct(PostalRepository $postalRepository)`

### index (line 25)

`public function index()`

Guard/validation candidates:

```php
27: if (Route::current()->getName() == 'receives.index') {
30: if (Route::current()->getName() == 'dispatches.index') {
```

Views: `postals.receives.index`, `postals.dispatches.index`

### store (line 35)

`public function store(PostalRequest $request)`

Guard/validation candidates:

```php
41: if (Route::current()->getName() == 'receives.store') {
45: if (Route::current()->getName() == 'dispatches.store') {
```

Write/transaction candidates:

```php
39: $this->postalRepository->store($input);
```

### edit (line 50)

`public function edit(Postal $postal)`

Guard/validation candidates:

```php
52: if (Route::current()->getName() == 'receives.edit') {
56: if (Route::current()->getName() == 'dispatches.edit') {
```

### update (line 61)

`public function update(PostalRequest $request, Postal $postal)`

Guard/validation candidates:

```php
65: if (Route::current()->getName() == 'receives.update') {
69: if (Route::current()->getName() == 'dispatches.update') {
```

### destroy (line 74)

`public function destroy(Postal $postal)`

### downloadMedia (line 81)

`public function downloadMedia(Postal $postal)`

### export (line 88)

`public function export()`

Guard/validation candidates:

```php
91: if (Route::current()->getName() == 'receives.excel') {
93: if (!$postalRecevices) {
100: if (Route::current()->getName() == 'dispatches.excel') {
102: if (!$postalDispatchs) {
```

## hms/app/Http/Controllers/PrescriptionController.php

Models: `Medicine` → `medicines`, `Prescription` → `prescriptions`, `Setting` → `settings`, `User` → `users`

### __construct (line 32)

`public function __construct( PrescriptionRepository $prescriptionRepo, DoctorRepository $doctorRepository, MedicineRepository $medicineRepository )`

### index (line 43)

`public function index()`

Views: `prescriptions.index`

### create (line 50)

`public function create()`

Views: `prescriptions.create`

### store (line 67)

`public function store(CreatePrescriptionRequest $request)`

Guard/validation candidates:

```php
72: if (!isset($input['medicine'])) {
81: if (!empty($duplicateIds)) {
92: if ($medicine->available_quantity < $qty) {
```

Write/transaction candidates:

```php
100: $prescription = $this->prescriptionRepository->create($input);
```

### show (line 108)

`public function show(Prescription $prescription)`

Guard/validation candidates:

```php
112: if (empty($prescription)) {
```

Views: `prescriptions.show`

### edit (line 121)

`public function edit(Prescription $prescription)`

Guard/validation candidates:

```php
123: if (getLoggedinDoctor() && checkRecordAccess($prescription->doctor_id)) {
```

Views: `errors.404`, `prescriptions.edit`

### update (line 145)

`public function update(Prescription $prescription, UpdatePrescriptionRequest $request)`

Guard/validation candidates:

```php
159: if (empty($prescription)) {
175: if (!empty($duplicateIds)) {
184: if (!array_key_exists($input['medicine'][$key], $result) && $medicine->available_quantity < $qty) {
```

### destroy (line 198)

`public function destroy(Prescription $prescription)`

Guard/validation candidates:

```php
200: if (checkRecordAccess($prescription->doctor_id)) {
205: if (empty($prescription)) {
```

Write/transaction candidates:

```php
210: $prescription->delete();
```

### activeDeactiveStatus (line 216)

`public function activeDeactiveStatus($id)`

Guard/validation candidates:

```php
220: if (getLoggedinDoctor() && checkRecordAccess($prescription->doctor_id)) {
```

Write/transaction candidates:

```php
224: $prescription->update(['status' => $status]);
```

### prescriptionsView (line 230)

`public function prescriptionsView($id)`

Guard/validation candidates:

```php
236: if (getLoggedinDoctor() && checkRecordAccess($prescription['prescription']->doctor_id)) {
```

Views: `errors.404`, `prescriptions.view`

### prescreptionMedicineStore (line 245)

`public function prescreptionMedicineStore(CreateMedicineRequest $request)`

Write/transaction candidates:

```php
249: $this->medicineRepository->create($input);
```

### convertToPDF (line 254)

`public function convertToPDF($id)`

### getAvailableMedicineQuantity (line 267)

`public function getAvailableMedicineQuantity(Medicine $medicine)`

### openAiPrompt (line 272)

`public function openAiPrompt(Request $request)`

Guard/validation candidates:

```php
277: if (empty(array_filter($patientDetails))) {
283: if ($value === null) {
335: if (empty($openAiKey)) {
339: if (empty($openAiModelName)) {
343: if (!$openAiKey) {
363: if (isset($data->json()['error'])) {
```

## hms/app/Http/Controllers/PurchaseMedicineController.php

Models: `Medicine` → `medicines`, `PurchasedMedicine` → `purchased_medicines`, `PurchaseMedicine` → `purchase_medicines`

### __construct (line 36)

`public function __construct(PurchaseMedicineRepository $purchaseMedicineRepo, MedicineRepository $medicineRepository, MedicineBillRepository $medicineBillRepository, IpdPaymentRepository $ipdPaymentRepo, AppointmentTransactionRepository $appointmentTransactionRepository,)`

### index (line 51)

`public function index()`

Views: `purchase-medicines.index`

### create (line 57)

`public function create()`

Views: `purchase-medicines.create`

### setFlutterWaveCredential (line 69)

`public function setFlutterWaveCredential()`

Guard/validation candidates:

```php
74: if (!$flutterwavePublicKey && !$flutterwaveSecretKey) {
```

### store (line 84)

`public function store(CreatePurchaseMedicineRequest $request)`

Guard/validation candidates:

```php
91: if ($input['payment_type'] == PurchaseMedicine::PURCHASE_MEDICINE_STRIPE) {
108: if (strtoupper(getCurrentCurrency()) != 'INR') {
117: if (!in_array(strtoupper(getCurrentCurrency()), getFlutterWaveSupportedCurrencies())) {
```

Write/transaction candidates:

```php
89: DB::beginTransaction();
93: $this->purchaseMedicineRepository->store($input);
95: DB::commit();
99: $this->purchaseMedicineRepository->store($input);
100: DB::commit();
129: $this->purchaseMedicineRepository->store($input);
130: DB::commit();
135: DB::rollBack();
```

### show (line 140)

`public function show(PurchaseMedicine $medicinePurchase)`

Views: `purchase-medicines.show`

### getMedicine (line 147)

`public function getMedicine(Medicine $medicine)`

### purchaseMedicineExport (line 155)

`public function purchaseMedicineExport()`

Guard/validation candidates:

```php
158: if (!$purchaseMedicines) {
```

### usedMedicine (line 169)

`public function usedMedicine()`

Views: `used-medicine.index`

### destroy (line 175)

`public function destroy(PurchaseMedicine $medicinePurchase)`

Write/transaction candidates:

```php
177: $medicinePurchase->delete();
```

### stripeSuccess (line 182)

`public function stripeSuccess(Request $request)`

### stripeFail (line 192)

`public function stripeFail(Request $request)`

Write/transaction candidates:

```php
198: $purchaseMedicine->delete();
206: $medicine->update($medicineQtyArray);
```

### razorPayInit (line 214)

`public function razorPayInit(Request $request)`

### razorPaySuccess (line 221)

`public function razorPaySuccess(Request $request)`

### razorPayFailed (line 230)

`public function razorPayFailed(Request $request)`

Write/transaction candidates:

```php
236: $purchaseMedicine->delete();
244: $medicine->update($medicineQtyArray);
```

### paystackPayment (line 250)

`public function paystackPayment(Request $request)`

Guard/validation candidates:

```php
252: if (!in_array(strtoupper(getCurrentCurrency()), getPayStackSupportedCurrencies())) {
```

### paystackPaymentSuccess (line 282)

`public function paystackPaymentSuccess(Request $request)`

Guard/validation candidates:

```php
286: if (isset($paymentDetails['data']['metadata']['data']['appointment_charge'])) {
292: if (isset($isWebAppointment) && $isWebAppointment == true) {
304: if (isset($paymentDetails['data']['metadata']['data']['purchase_no'])) {
313: if (isset($paymentDetails['data']['metadata']['data']['patient_id'])) {
321: if (isset($paymentDetails['data']['metadata']['ipd_patient_id'])) {
326: if (getLoggedinPatient()) {
```

### phonePePaymentSuccess (line 334)

`public function phonePePaymentSuccess(Request $request)`

### flutterwavePaymentSuccess (line 343)

`public function flutterwavePaymentSuccess(Request $request)`

Guard/validation candidates:

```php
347: if ($request['status'] == 'cancelled') {
```

## hms/app/Http/Controllers/RadiologyCategoryController.php

Models: `RadiologyCategory` → `radiology_categories`, `RadiologyTest` → `radiology_tests`

### __construct (line 16)

`public function __construct(RadiologyCategoryRepository $radiologyCategoryRepo)`

### index (line 21)

`public function index()`

Views: `radiology_categories.index`

### store (line 26)

`public function store(CreateRadiologyCategoryRequest $request)`

Write/transaction candidates:

```php
29: $this->radiologyCategoryRepository->create($input);
```

### edit (line 34)

`public function edit(RadiologyCategory $radiologyCategory)`

### update (line 39)

`public function update(RadiologyCategory $radiologyCategory, UpdateRadiologyCategoryRequest $request)`

Write/transaction candidates:

```php
42: $this->radiologyCategoryRepository->update($input, $radiologyCategory->id);
```

### destroy (line 47)

`public function destroy(RadiologyCategory $radiologyCategory)`

Guard/validation candidates:

```php
54: if ($result) {
```

Write/transaction candidates:

```php
58: $radiologyCategory->delete();
```

## hms/app/Http/Controllers/RadiologyTestController.php

Models: `Charge` → `charges`, `RadiologyTest` → `radiology_tests`

### __construct (line 19)

`public function __construct(RadiologyTestRepository $radiologyTestRepo)`

### index (line 24)

`public function index()`

Views: `radiology_tests.index`

### create (line 29)

`public function create()`

Views: `radiology_tests.create`

### store (line 36)

`public function store(CreateRadiologyTestRequest $request)`

Write/transaction candidates:

```php
42: $this->radiologyTestRepository->create($input);
```

### show (line 48)

`public function show(RadiologyTest $radiologyTest)`

Views: `radiology_tests.show`

### edit (line 53)

`public function edit(RadiologyTest $radiologyTest)`

Views: `radiology_tests.edit`

### update (line 61)

`public function update(RadiologyTest $radiologyTest, UpdateRadiologyTestRequest $request)`

Write/transaction candidates:

```php
67: $this->radiologyTestRepository->update($input, $radiologyTest->id);
```

### destroy (line 73)

`public function destroy(RadiologyTest $radiologyTest)`

Write/transaction candidates:

```php
75: $radiologyTest->delete();
```

### getStandardCharge (line 80)

`public function getStandardCharge($id)`

### getChargeCode (line 87)

`public function getChargeCode($id)`

### radiologyTestExport (line 94)

`public function radiologyTestExport()`

Guard/validation candidates:

```php
97: if (!$radiologyTests) {
```

### showModal (line 104)

`public function showModal(RadiologyTest $radiologyTest)`

## hms/app/Http/Controllers/ReceptionistController.php

Models: `EmployeePayroll` → `employee_payrolls`, `Receptionist` → `receptionists`

### __construct (line 19)

`public function __construct(ReceptionistRepository $receptionistRepo)`

### index (line 24)

`public function index()`

Views: `receptionists.index`

### create (line 31)

`public function create()`

Views: `receptionists.create`

### store (line 38)

`public function store(CreateReceptionistRequest $request)`

Write/transaction candidates:

```php
43: $receptionist = $this->receptionistRepository->store($input);
```

### show (line 50)

`public function show(Receptionist $receptionist)`

Views: `receptionists.show`

### edit (line 57)

`public function edit(Receptionist $receptionist)`

Views: `receptionists.edit`

### update (line 65)

`public function update(Receptionist $receptionist, UpdateReceptionistRequest $request)`

Write/transaction candidates:

```php
70: $receptionist = $this->receptionistRepository->update($receptionist, $input);
```

### destroy (line 77)

`public function destroy(Receptionist $receptionist)`

Guard/validation candidates:

```php
81: if ($empPayRollResult) {
```

Write/transaction candidates:

```php
85: $receptionist->user()->delete();
86: $receptionist->address()->delete();
87: $receptionist->delete();
```

### activeDeactiveStatus (line 92)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
96: $receptionist->user()->update(['status' => $status]);
```

### receptionistExport (line 101)

`public function receptionistExport()`

Guard/validation candidates:

```php
104: if (!$receptionists) {
```

## hms/app/Http/Controllers/ScheduleController.php

Models: `Appointment` → `appointments`, `Doctor` → `doctors`, `HospitalSchedule` → `hospital_schedules`, `Schedule` → `schedules`, `ScheduleDay` → `schedule_days`

### __construct (line 23)

`public function __construct(ScheduleRepository $scheduleRepo)`

### index (line 28)

`public function index()`

Guard/validation candidates:

```php
30: if (getLoggedInUser()->hasRole('Doctor')) {
```

Views: `schedules.index`, `schedules.index`

### create (line 40)

`public function create()`

Guard/validation candidates:

```php
46: if (count($hospitalSchedules) == 0) {
```

Views: `schedules.create`

### store (line 53)

`public function store(CreateScheduleRequest $request)`

Write/transaction candidates:

```php
57: $schedule = $this->scheduleRepository->store($input);
```

### show (line 64)

`public function show(Schedule $schedule)`

Guard/validation candidates:

```php
66: if (checkRecordAccess($schedule->doctor_id) && getLoggedInUser()->hasRole('Doctor')) {
```

Views: `errors.404`, `schedules.show`

### edit (line 75)

`public function edit(Schedule $schedule)`

Guard/validation candidates:

```php
77: if (getLoggedInUser()->hasRole('Doctor')) {
78: if (! (getLoggedInUser()->owner_id == $schedule->doctor_id)) {
```

Views: `schedules.edit`

### update (line 88)

`public function update(Schedule $schedule, UpdateScheduleRequest $request)`

Guard/validation candidates:

```php
93: if (getLoggedInUser()->hasRole('Doctor')) {
```

Write/transaction candidates:

```php
90: $schedule = $this->scheduleRepository->update($request->all(), $schedule->id);
```

### destroy (line 100)

`public function destroy(Schedule $schedule)`

Guard/validation candidates:

```php
105: if ($result) {
```

Write/transaction candidates:

```php
109: $this->scheduleRepository->delete($schedule->id);
110: ScheduleDay::whereScheduleId($schedule->id)->delete();
```

### doctorScheduleList (line 115)

`public function doctorScheduleList(Request $request)`

Guard/validation candidates:

```php
120: if ($value == $input['day_name']) {
127: if (! $check) {
```

### schedulesExport (line 136)

`public function schedulesExport()`

Guard/validation candidates:

```php
140: if (!$userSchedule) {
```

## hms/app/Http/Controllers/ServiceController.php

Models: `PackageService` → `package_services`, `Service` → `services`

### __construct (line 20)

`public function __construct(ServiceRepository $serviceRepo)`

### index (line 25)

`public function index()`

Views: `services.index`

### create (line 32)

`public function create()`

Views: `services.create`

### store (line 37)

`public function store(CreateServiceRequest $request)`

Write/transaction candidates:

```php
42: $this->serviceRepository->create($input);
```

### show (line 49)

`public function show(Service $service)`

Guard/validation candidates:

```php
53: if (empty($service)) {
```

Views: `services.show`

### edit (line 62)

`public function edit(Service $service)`

Views: `services.edit`

### update (line 67)

`public function update(Service $service, UpdateServiceRequest $request): RedirectResponse`

Guard/validation candidates:

```php
69: if (empty($service)) {
```

Write/transaction candidates:

```php
78: $this->serviceRepository->update($input, $service->id);
```

### destroy (line 84)

`public function destroy(Service $service)`

Guard/validation candidates:

```php
92: if ($result) {
```

Write/transaction candidates:

```php
96: $service->delete();
```

### activeDeActiveService (line 101)

`public function activeDeActiveService($id)`

Write/transaction candidates:

```php
105: $service->update(['status' => $service->status]);
```

### serviceExport (line 110)

`public function serviceExport()`

Guard/validation candidates:

```php
113: if (!$services) {
```

## hms/app/Http/Controllers/SettingController.php

Models: `Module` → `modules`, `Setting` → `settings`

### __construct (line 19)

`public function __construct(SettingRepository $settingRepo)`

### edit (line 24)

`public function edit(Request $request)`

Views: `settings.$sectionName`

### update (line 34)

`public function update(UpdateSettingRequest $request)`

### getModule (line 43)

`public function getModule(Request $request)`

Guard/validation candidates:

```php
45: if ($request->ajax()) {
```

### activeDeactiveStatus (line 50)

`public function activeDeactiveStatus(Module $module)`

Write/transaction candidates:

```php
53: $module->update(['is_active' => $is_active]);
```

### patientQueueThemeView (line 58)

`public function patientQueueThemeView()`

Guard/validation candidates:

```php
62: if (!isset($setting['patient_queue_theme'])) {
```

Views: `settings.queue_theme`

### patientQueueThemeUpdate (line 69)

`public function patientQueueThemeUpdate(Request $request)`

Guard/validation candidates:

```php
71: $request->validate([
77: if ($setting) {
```

Write/transaction candidates:

```php
78: $setting->update(['value' => $request->patient_queue_theme]);
80: Setting::create([
```

### settingsUploadThemeVideo (line 89)

`public function settingsUploadThemeVideo(Request $request)`

Guard/validation candidates:

```php
91: $request->validate([
95: if ($request->hasFile('patient_queue_theme_video')) {
```

Write/transaction candidates:

```php
104: $setting->update([
```

## hms/app/Http/Controllers/SmsController.php

Models: `Sms` → `sms`

### __construct (line 18)

`public function __construct(SmsRepository $smsRepository)`

### index (line 23)

`public function index()`

Views: `sms.index`

### store (line 30)

`public function store(CreateSmsRequest $request)`

Write/transaction candidates:

```php
33: $this->smsRepository->store($input);
```

### show (line 38)

`public function show(Sms $sms)`

Guard/validation candidates:

```php
40: if (getLoggedInUser()->hasRole('Admin')) {
45: if ($sms->send_by == getLoggedInUser()->id) {
```

Views: `sms.show`, `sms.show`, `errors.404`

### destroy (line 55)

`public function destroy(Sms $sms)`

Guard/validation candidates:

```php
57: if (getLoggedInUser()->hasRole('Admin')) {
62: if ($sms->send_by == getLoggedInUser()->id) {
```

Write/transaction candidates:

```php
58: $this->smsRepository->delete($sms->id);
63: $this->smsRepository->delete($sms->id);
```

### getUsersList (line 72)

`public function getUsersList(Request $request)`

Guard/validation candidates:

```php
74: if (empty($request->get('id'))) {
```

### showModal (line 88)

`public function showModal(Sms $sms)`

Guard/validation candidates:

```php
90: if (getLoggedInUser()->hasRole('Admin')) {
95: if ($sms->send_by == getLoggedInUser()->id) {
```

## hms/app/Http/Controllers/StripeController.php

Models: `IpdPatientDepartment` → `ipd_patient_departments`

### __construct (line 20)

`public function __construct(StripeRepository $stripeRepository)`

### createSession (line 25)

`public function createSession(Request $request)`

Guard/validation candidates:

```php
31: if (in_array(strtoupper(getCurrentCurrency()), zeroDecimalCurrencies())) {
```

Write/transaction candidates:

```php
40: $session = Session::create([
```

### paymentSuccess (line 68)

`public function paymentSuccess(Request $request)`

Guard/validation candidates:

```php
72: if (empty($sessionId)) {
```

### handleFailedPayment (line 83)

`public function handleFailedPayment()`

## hms/app/Http/Controllers/TestimonialController.php

Models: `Testimonial` → `testimonials`

### __construct (line 17)

`public function __construct(TestimonialRepository $testimonialRepository)`

### index (line 22)

`public function index()`

Views: `testimonials.index`

### store (line 27)

`public function store(TestimonialRequest $request)`

Write/transaction candidates:

```php
31: $this->testimonialRepository->store($input);
```

### edit (line 39)

`public function edit(Testimonial $testimonial)`

### update (line 44)

`public function update(Testimonial $testimonial, TestimonialRequest $request)`

### show (line 55)

`public function show(Testimonial $testimonial)`

### destroy (line 60)

`public function destroy(Testimonial $testimonial)`

## hms/app/Http/Controllers/UserController.php

Models: `Department` → `departments`, `DoctorDepartment` → `doctor_departments`, `Setting` → `settings`, `User` → `users`

### __construct (line 29)

`public function __construct(UserRepository $userRepo)`

### changePassword (line 34)

`public function changePassword(ChangePasswordRequest $request)`

### profileUpdate (line 47)

`public function profileUpdate(UpdateUserProfileRequest $request)`

### editProfile (line 60)

`public function editProfile()`

### updateLanguage (line 67)

`public function updateLanguage(Request $request)`

Write/transaction candidates:

```php
73: $user->update(['language' => $language]);
```

### index (line 78)

`public function index()`

Views: `users.index`

### create (line 86)

`public function create()`

Views: `users.create`

### store (line 95)

`public function store(CreateUserRequest $request)`

Write/transaction candidates:

```php
98: DB::beginTransaction();
101: $this->userRepository->store($input);
103: DB::commit();
107: DB::rollBack();
```

### show (line 113)

`public function show($user)`

Views: `users.show`

### edit (line 120)

`public function edit(User $user)`

Views: `users.edit`

### update (line 128)

`public function update(UpdateUserRequest $request, User $user)`

### destroy (line 140)

`public function destroy(User $user)`

### activeDeactiveStatus (line 147)

`public function activeDeactiveStatus($id)`

Write/transaction candidates:

```php
151: $user->update(['status' => $status]);
```

### showModal (line 156)

`public function showModal($user)`

### isVerified (line 163)

`public function isVerified($id)`

Write/transaction candidates:

```php
167: $user->update(['email_verified_at' => $emailVerified]);
```

### changeThemeMode (line 172)

`public function changeThemeMode()`

Guard/validation candidates:

```php
176: if ($user->thememode == User::THEME_LIGHT_MODE) {
```

Write/transaction candidates:

```php
182: $user->update();
```

## hms/app/Http/Controllers/VaccinatedPatientController.php

Models: `VaccinatedPatients` → `vaccinated_patients`

### __construct (line 21)

`public function __construct(VaccinatedPatientRepository $vaccinatedPatientRepository)`

### index (line 26)

`public function index()`

Views: `vaccinated_patients.index`

### store (line 33)

`public function store(CreateVaccinatedPatientRequest $request)`

Guard/validation candidates:

```php
39: if ($checkValidation) {
```

Write/transaction candidates:

```php
43: $this->vaccinatedPatientRepository->create($input);
```

### edit (line 51)

`public function edit(VaccinatedPatients $vaccinatedPatient)`

### update (line 56)

`public function update(UpdateVaccinatedPatientRequest $request, VaccinatedPatients $vaccinatedPatient)`

Guard/validation candidates:

```php
61: if (
69: if ($checkValidation) {
```

Write/transaction candidates:

```php
73: $this->vaccinatedPatientRepository->update($input, $vaccinatedPatient->id);
```

### destroy (line 81)

`public function destroy(VaccinatedPatients $vaccinatedPatient)`

Write/transaction candidates:

```php
84: $vaccinatedPatient->delete();
```

### vaccinatedPatientExport (line 92)

`public function vaccinatedPatientExport()`

Guard/validation candidates:

```php
95: if (!$vaccinatedPatients) {
```

## hms/app/Http/Controllers/VaccinationController.php

Models: `VaccinatedPatients` → `vaccinated_patients`, `Vaccination` → `vaccinations`

### __construct (line 22)

`public function __construct(VaccinationRepository $vaccinationRepository)`

### index (line 27)

`public function index()`

Views: `vaccinations.index`

### store (line 32)

`public function store(CreateVaccinationRequest $request)`

Write/transaction candidates:

```php
36: $this->vaccinationRepository->create($input);
```

### edit (line 44)

`public function edit(Vaccination $vaccination)`

### update (line 49)

`public function update(UpdateVaccinationRequest $request, Vaccination $vaccination)`

Write/transaction candidates:

```php
53: $this->vaccinationRepository->update($input, $vaccination->id);
```

### destroy (line 61)

`public function destroy(Vaccination $vaccination)`

Guard/validation candidates:

```php
70: if ($result) {
```

Write/transaction candidates:

```php
74: $vaccination->delete();
```

### vaccinationsExport (line 82)

`public function vaccinationsExport()`

Guard/validation candidates:

```php
85: if (!$vaccinations) {
```

## hms/app/Http/Controllers/VisitorController.php

Models: `Visitor` → `visitors`

### __construct (line 20)

`public function __construct(VisitorRepository $visitorRepo)`

### index (line 25)

`public function index()`

Views: `visitors.index`

### create (line 32)

`public function create()`

Views: `visitors.create`

### store (line 40)

`public function store(CreateVisitorRequest $request)`

Write/transaction candidates:

```php
56: $this->visitorRepository->store($input);
```

### edit (line 62)

`public function edit(Visitor $visitor)`

Views: `visitors.edit`

### update (line 71)

`public function update(UpdateVisitorRequest $request, Visitor $visitor)`

### destroy (line 81)

`public function destroy(Visitor $visitor)`

### downloadMedia (line 88)

`public function downloadMedia(Visitor $visitor)`

### export (line 95)

`public function export()`

Guard/validation candidates:

```php
98: if (!$visitors) {
```

## hms/app/Http/Controllers/Web/AppointmentController.php

Models: `Appointment` → `appointments`, `Department` → `departments`, `Patient` → `patients`, `User` → `users`

### __construct (line 28)

`public function __construct(AppointmentRepository $appointmentRepo, AppointmentTransactionRepository $appointmentTransactionRepo)`

### setFlutterWaveCredential (line 34)

`public function setFlutterWaveCredential()`

Guard/validation candidates:

```php
39: if(!$flutterwavePublicKey && !$flutterwaveSecretKey){
```

### store (line 49)

`public function store(CreateWebAppointmentRequest $request)`

Guard/validation candidates:

```php
58: if ($input['patient_type'] == 2 && ! empty($input['patient_type'])) {
61: if (strpos($key, 'field') === 0) {
67: if($input['payment_mode'] != 7 && $input['payment_mode'] != 8 && $input['payment_mode'] != 9){
71: if($input['payment_mode'] == 3 || $input['payment_mode'] == 4 || $input['payment_mode'] == 5){
75: if($input['payment_mode'] == 3){
103: if(!in_array(strtoupper(getCurrentCurrency()),getFlutterWaveSupportedCurrencies())){
116: if (strtoupper(getCurrentCurrency()) != 'INR') {
136: if ($input['patient_type'] == 1 && ! empty($input['patient_type'])) {
138: if ($emailExists) {
153: if (strpos($key, 'field') === 0) {
159: if (isset($input['email'])) {
170: if($input['payment_mode'] != 7 && $input['payment_mode'] != 8 && $input['payment_mode'] != 9){
183: if($input['payment_mode'] == 3){
211: if(!in_array(strtoupper(getCurrentCurrency()),getFlutterWaveSupportedCurrencies())){
225: if (strtoupper(getCurrentCurrency()) != 'INR') {
```

Write/transaction candidates:

```php
57: DB::beginTransaction();
68: $appointment = $this->appointmentRepository->create($input);
72: $appointment->update(['patient_type',1]);
76: DB::commit();
86: DB::commit();
93: DB::commit();
101: DB::commit();
114: DB::commit();
125: DB::commit();
131: $data = $this->appointmentTransactionRepository->store($appointment);
158: $user = User::create($userData);
163: $patient = Patient::create(['user_id' => $user->id,'patient_unique_id' => strtoupper(Patient::generateUniquePatientId())]);
167: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
171: $appointment = Appointment::create([
184: DB::commit();
194: DB::commit();
201: DB::commit();
209: DB::commit();
223: DB::commit();
235: DB::commit();
242: $data = $this->appointmentTransactionRepository->store($appointment);
243: DB::commit();
249: DB::commit();
253: DB::rollBack();
```

### getDoctors (line 259)

`public function getDoctors(Request $request)`

### getDoctorList (line 268)

`public function getDoctorList(Request $request)`

### getBookingSlot (line 276)

`public function getBookingSlot(Request $request)`

### getPatientDetails (line 285)

`public function getPatientDetails($email)`

Guard/validation candidates:

```php
289: if ($patient != null) {
```

### getDoctorsCharge (line 298)

`public function getDoctorsCharge(Request $request)`

## hms/app/Http/Controllers/Web/WebController.php

Models: `AddCustomFields` → `add_custom_fields`, `Bed` → `beds`, `Doctor` → `doctors`, `DoctorDepartment` → `doctor_departments`, `FrontService` → `front_services`, `FrontSetting` → `front_settings`, `HospitalSchedule` → `hospital_schedules`, `NoticeBoard` → `notice_boards`, `Nurse` → `nurses`, `Patient` → `patients`, `Setting` → `settings`, `Testimonial` → `testimonials`, `Vaccination` → `vaccinations`, `PatientCase` → `patient_cases`, `PatientAdmission` → `patient_admissions`, `Appointment` → `appointments`, `Bill` → `bills`, `Invoice` → `invoices`, `AdvancedPayment` → `advanced_payments`, `Document` → `documents`, `VaccinatedPatients` → `vaccinated_patients`, `PatientIdCardTemplate` → `patient_id_card_templates`

### __construct (line 51)

`public function __construct(AppointmentRepository $appointmentRepository, PatientRepository $patientRepo)`

### index (line 57)

`public function index()`

Views: `web.home.index`

### demo (line 79)

`public function demo()`

Views: `web.demo.index`

### modulesOfHms (line 84)

`public function modulesOfHms()`

Views: `web.modules_of_hms.index`

### changeLanguage (line 89)

`public function changeLanguage(Request $request)`

### aboutUs (line 101)

`public function aboutUs()`

Views: `web.home.about_us`

### appointmentFromOther (line 118)

`public function appointmentFromOther(Request $request)`

### appointment (line 125)

`public function appointment()`

Views: `web.home.appointment`

### services (line 134)

`public function services()`

Views: `web.home.services`

### doctors (line 141)

`public function doctors()`

Views: `web.home.doctors`

### termsOfService (line 152)

`public function termsOfService()`

Views: `web.home.terms-of-service`

### privacyPolicy (line 159)

`public function privacyPolicy()`

Views: `web.home.privacy-policy`

### workingHours (line 166)

`public function workingHours()`

Views: `web.home.working-hours`

### testimonials (line 175)

`public function testimonials()`

Views: `web.home.testimonials`

### setLanguage (line 182)

`public function setLanguage(Request $request)`

### showQrCodePatient (line 194)

`public function showQrCodePatient($uniqueId)`

Guard/validation candidates:

```php
202: if (!empty($user) && $user->hasRole('Doctor')) {
```

Views: `web.home.qr_code_patient`

### doctorDetails (line 222)

`public function doctorDetails($id)`

Views: `web.home.doctor-details`

## hms/app/Http/Controllers/adminController.php

Models: `User` → `users`

### __construct (line 16)

`public function __construct(adminRepository $adminRepo)`

### index (line 21)

`public function index()`

Views: `admins.index`

### create (line 26)

`public function create()`

Views: `admins.create`

### store (line 33)

`public function store(CreateAdminRequest $request)`

Write/transaction candidates:

```php
37: $this->adminRepository->store($input);
```

### show (line 44)

`public function show($id)`

Guard/validation candidates:

```php
48: if (empty($admin) && $admin->owner_type != \App\Models\admin::class) {
```

Views: `errors.404`, `admins.show`

### edit (line 55)

`public function edit(User $admin)`

Guard/validation candidates:

```php
59: if (empty($admin) && $admin->owner_type != \App\Models\admin::class) {
```

Views: `errors.404`, `admins.edit`

### update (line 66)

`public function update(User $admin, UpdateAdminRequest $request)`

Write/transaction candidates:

```php
71: $admin = $this->adminRepository->update($admin, $input);
```

### destroy (line 78)

`public function destroy(User $admin)`

Guard/validation candidates:

```php
80: if (empty($admin) && $admin->owner_type != \App\Models\admin::class) {
```

Write/transaction candidates:

```php
83: $admin->delete();
```

### activeDeactiveStatus (line 89)

`public function activeDeactiveStatus($id)`

Guard/validation candidates:

```php
93: if (empty($admin) && $admin->owner_type != \App\Models\admin::class) {
```

Write/transaction candidates:

```php
97: $admin->update(['status' => $status]);
```

## hms/app/Http/Requests/BloodDonationRequest.php

Models: `BloodDonation` → `blood_donations`

### authorize (line 16)

`public function authorize(): bool`

Guard/validation candidates:

```php
18: return true;
```

### rules (line 24)

`public function rules(): array`

Guard/validation candidates:

```php
26: return BloodDonation::$rules;
```

## hms/app/Http/Requests/BloodIssueRequest.php

Models: `BloodIssue` → `blood_issues`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return BloodIssue::$rules;
```

## hms/app/Http/Requests/ChangePasswordRequest.php

Models:

### authorize (line 15)

`public function authorize(): bool`

Guard/validation candidates:

```php
17: return true;
```

### rules (line 25)

`public function rules(): array`

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/CreateAccountRequest.php

Models: `Account` → `accounts`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Account::$rules;
```

### prepareForValidation (line 26)

`protected function prepareForValidation()`

### sanitize (line 31)

`public function sanitize()`

## hms/app/Http/Requests/CreateAccountantRequest.php

Models: `Accountant` → `accountants`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Accountant::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateAddCustomFieldRequest.php

Models: `AddCustomFields` → `add_custom_fields`

### authorize (line 14)

`public function authorize(): bool`

Guard/validation candidates:

```php
16: return true;
```

### rules (line 24)

`public function rules(): array`

Guard/validation candidates:

```php
26: $rules = AddCustomFields::$rules;
```

## hms/app/Http/Requests/CreateAdminRequest.php

Models: `admin` → `admins`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = admin::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateAdvancedPaymentRequest.php

Models: `AdvancedPayment` → `advanced_payments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return AdvancedPayment::$rules;
```

## hms/app/Http/Requests/CreateAmbulanceCallRequest.php

Models: `AmbulanceCall` → `ambulance_calls`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return AmbulanceCall::$rules;
```

### messages (line 26)

`public function messages()`

## hms/app/Http/Requests/CreateAmbulanceRequest.php

Models: `Ambulance` → `ambulances`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Ambulance::$rules;
```

## hms/app/Http/Requests/CreateAppointmentRequest.php

Models: `Appointment` → `appointments`, `Doctor` → `doctors`

### authorize (line 14)

`public function authorize(): bool`

Guard/validation candidates:

```php
16: return true;
```

### prepareForValidation (line 22)

`protected function prepareForValidation(): void`

Guard/validation candidates:

```php
27: if (getLoggedInUser()->hasRole('Patient')) {
```

### rules (line 35)

`public function rules(): array`

Guard/validation candidates:

```php
37: return Appointment::$rules;
```

## hms/app/Http/Requests/CreateBedAssignRequest.php

Models: `BedAssign` → `bed_assigns`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return BedAssign::$rules;
```

### messages (line 29)

`public function messages(): array`

## hms/app/Http/Requests/CreateBedRequest.php

Models: `Bed` → `beds`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Bed::$rules;
```

## hms/app/Http/Requests/CreateBedTypeRequest.php

Models: `BedType` → `bed_types`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return BedType::$rules;
```

## hms/app/Http/Requests/CreateBillRequest.php

Models: `Bill` → `bills`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Bill::$rules;
```

### messages (line 26)

`public function messages(): array`

## hms/app/Http/Requests/CreateBirthReportRequest.php

Models: `BirthReport` → `birth_reports`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return BirthReport::$rules;
```

## hms/app/Http/Requests/CreateBloodBankRequest.php

Models: `BloodBank` → `blood_banks`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return BloodBank::$rules;
```

## hms/app/Http/Requests/CreateBloodDonorRequest.php

Models: `BloodDonor` → `blood_donors`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return BloodDonor::$rules;
```

## hms/app/Http/Requests/CreateBrandRequest.php

Models: `Brand` → `brands`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Brand::$rules;
```

## hms/app/Http/Requests/CreateBulkBedRequest.php

Models:

### authorize (line 12)

`public function authorize(): bool`

Guard/validation candidates:

```php
14: return true;
```

### rules (line 20)

`public function rules(): array`

### messages (line 30)

`public function messages(): array`

## hms/app/Http/Requests/CreateCallLogRequest.php

Models: `CallLog` → `call_logs`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return CallLog::$rules;
```

## hms/app/Http/Requests/CreateCaseHandlerRequest.php

Models: `CaseHandler` → `case_handlers`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = CaseHandler::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateCategoryRequest.php

Models: `Category` → `categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Category::$rules;
```

## hms/app/Http/Requests/CreateChargeCategoryRequest.php

Models: `ChargeCategory` → `charge_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return ChargeCategory::$rules;
```

## hms/app/Http/Requests/CreateChargeRequest.php

Models: `Charge` → `charges`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Charge::$rules;
```

## hms/app/Http/Requests/CreateComplaint.php

Models: `Complaint` → `complaints`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return Complaint::$rules;
```

## hms/app/Http/Requests/CreateCurrencySettingRequest.php

Models: `CurrencySetting` → `currency_settings`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return CurrencySetting::$rules;
```

## hms/app/Http/Requests/CreateDeathReportRequest.php

Models: `DeathReport` → `death_reports`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return DeathReport::$rules;
```

## hms/app/Http/Requests/CreateDepartmentRequest.php

Models: `Department` → `departments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Department::$rules;
```

## hms/app/Http/Requests/CreateDiagnosisCategoryRequest.php

Models: `DiagnosisCategory` → `diagnosis_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return DiagnosisCategory::$rules;
```

## hms/app/Http/Requests/CreateDoctorDepartmentRequest.php

Models: `DoctorDepartment` → `doctor_departments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return DoctorDepartment::$rules;
```

## hms/app/Http/Requests/CreateDoctorOPDChargeRequest.php

Models: `DoctorOPDCharge` → `doctor_opd_charges`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return DoctorOPDCharge::$rules;
```

### messages (line 26)

`public function messages(): array`

## hms/app/Http/Requests/CreateDoctorRequest.php

Models: `Doctor` → `doctors`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Doctor::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateDocumentRequest.php

Models: `Document` → `documents`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Document::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateDocumentTypeRequest.php

Models: `DocumentType` → `document_types`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return DocumentType::$rules;
```

## hms/app/Http/Requests/CreateEmployeePayrollRequest.php

Models: `EmployeePayroll` → `employee_payrolls`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return EmployeePayroll::$rules;
```

## hms/app/Http/Requests/CreateEnquiryRequest.php

Models: `Enquiry` → `enquiries`

### authorize (line 14)

`public function authorize(): bool`

Guard/validation candidates:

```php
16: return true;
```

### rules (line 22)

`public function rules(): array`

Guard/validation candidates:

```php
24: $rules = Enquiry::$rules;
25: if (config('app.recaptcha.sitekey')) {
```

### attributes (line 35)

`public function attributes(): array`

## hms/app/Http/Requests/CreateExpenseRequest.php

Models: `Expense` → `expenses`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Expense::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateHolidayRequest.php

Models: `DoctorHoliday` → `doctor_holidays`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = DoctorHoliday::$rules;
```

## hms/app/Http/Requests/CreateIPDOperationRequest.php

Models: `IpdOperation` → `ipd_operation`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return IpdOperation::$rules;
```

### messages (line 31)

`public function messages(): array`

## hms/app/Http/Requests/CreateIncomeRequest.php

Models: `Income` → `incomes`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Income::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateInsuranceRequest.php

Models: `Insurance` → `insurances`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Insurance::$rules;
```

## hms/app/Http/Requests/CreateInvestigationReportRequest.php

Models: `InvestigationReport` → `investigation_reports`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = InvestigationReport::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateInvoiceRequest.php

Models: `Invoice` → `invoices`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Invoice::$rules;
```

### messages (line 26)

`public function messages(): array`

## hms/app/Http/Requests/CreateIpdBillRequest.php

Models: `IpdBill` → `ipd_bills`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return IpdBill::$rules;
```

## hms/app/Http/Requests/CreateIpdChargeRequest.php

Models: `IpdCharge` → `ipd_charges`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return IpdCharge::$rules;
```

## hms/app/Http/Requests/CreateIpdConsultantRegisterRequest.php

Models: `IpdConsultantRegister` → `ipd_consultant_registers`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return IpdConsultantRegister::$rules;
```

### messages (line 26)

`public function messages(): array`

## hms/app/Http/Requests/CreateIpdDiagnosisRequest.php

Models: `IpdDiagnosis` → `ipd_diagnoses`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return IpdDiagnosis::$rules;
```

## hms/app/Http/Requests/CreateIpdPatientDepartmentRequest.php

Models: `IpdPatientDepartment` → `ipd_patient_departments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return IpdPatientDepartment::$rules;
```

### messages (line 26)

`public function messages(): array`

## hms/app/Http/Requests/CreateIpdPaymentRequest.php

Models: `IpdBill` → `ipd_bills`, `IpdCharge` → `ipd_charges`, `IpdPatientDepartment` → `ipd_patient_departments`, `IpdPayment` → `ipd_payments`

### authorize (line 16)

`public function authorize(): bool`

Guard/validation candidates:

```php
18: return true;
```

### prepareForValidation (line 24)

`protected function prepareForValidation(): void`

### rules (line 33)

`public function rules(): array`

Guard/validation candidates:

```php
35: $rules = IpdPayment::$rules;
```

## hms/app/Http/Requests/CreateIpdPrescriptionRequest.php

Models: `IpdPrescription` → `ipd_prescriptions`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return IpdPrescription::$rules;
```

### messages (line 26)

`public function messages(): array`

## hms/app/Http/Requests/CreateIpdTimelineRequest.php

Models: `IpdTimeline` → `ipd_timelines`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return IpdTimeline::$rules;
```

## hms/app/Http/Requests/CreateIssuedItemRequest.php

Models: `IssuedItem` → `issued_items`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return IssuedItem::$rules;
```

## hms/app/Http/Requests/CreateItemCategoryRequest.php

Models: `ItemCategory` → `item_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return ItemCategory::$rules;
```

## hms/app/Http/Requests/CreateItemRequest.php

Models: `Item` → `items`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Item::$rules;
```

## hms/app/Http/Requests/CreateItemStockRequest.php

Models: `ItemStock` → `item_stocks`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return ItemStock::$rules;
```

## hms/app/Http/Requests/CreateLabTechnicianRequest.php

Models: `LabTechnician` → `lab_technicians`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = LabTechnician::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateLunchBreakRequest.php

Models: `LunchBreak` → `lunch_breaks`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return LunchBreak::$rules;
```

## hms/app/Http/Requests/CreateMailRequest.php

Models: `Mail` → `mails`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Mail::$rules;
```

## hms/app/Http/Requests/CreateMedicineBillRequest.php

Models:

### authorize (line 12)

`public function authorize(): bool`

Guard/validation candidates:

```php
14: return true;
```

### rules (line 22)

`public function rules(): array`

## hms/app/Http/Requests/CreateMedicineRequest.php

Models: `Medicine` → `medicines`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### prepareForValidation (line 18)

`protected function prepareForValidation()`

### rules (line 26)

`public function rules(): array`

Guard/validation candidates:

```php
28: return Medicine::$rules;
```

### messages (line 31)

`public function messages(): array`

### sanitize (line 39)

`public function sanitize()`

## hms/app/Http/Requests/CreateNoticeBoardRequest.php

Models: `NoticeBoard` → `notice_boards`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return NoticeBoard::$rules;
```

## hms/app/Http/Requests/CreateNurseRequest.php

Models: `Nurse` → `nurses`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Nurse::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateOdontogramRequest.php

Models: `Odontogram` → `odontograms`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return Odontogram::$rules;
```

## hms/app/Http/Requests/CreateOpdDiagnosisRequest.php

Models: `OpdDiagnosis` → `opd_diagnoses`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return OpdDiagnosis::$rules;
```

## hms/app/Http/Requests/CreateOpdPatientDepartmentRequest.php

Models: `OpdPatientDepartment` → `opd_patient_departments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return OpdPatientDepartment::$rules;
```

### messages (line 26)

`public function messages(): array`

## hms/app/Http/Requests/CreateOpdPrescriptionRequest.php

Models: `OpdPrescription` → `opd_prescriptions`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return OpdPrescription::$rules;
```

### messages (line 28)

`public function messages(): array`

## hms/app/Http/Requests/CreateOpdTimelineRequest.php

Models: `OpdTimeline` → `opd_timelines`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return OpdTimeline::$rules;
```

## hms/app/Http/Requests/CreateOperationReportRequest.php

Models: `OperationReport` → `operation_reports`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return OperationReport::$rules;
```

## hms/app/Http/Requests/CreateOperationRequest.php

Models: `Operation` → `operations`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return Operation::$rules;
```

### messages (line 28)

`public function messages(): array`

## hms/app/Http/Requests/CreatePackageRequest.php

Models: `Package` → `packages`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Package::$rules;
```

## hms/app/Http/Requests/CreatePathologyCategoryRequest.php

Models: `PathologyCategory` → `pathology_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return PathologyCategory::$rules;
```

## hms/app/Http/Requests/CreatePathologyParameterRequest.php

Models: `PathologyParameter` → `pathology_parameters`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return PathologyParameter::$rules;
```

## hms/app/Http/Requests/CreatePathologyTestRequest.php

Models: `PathologyTest` → `pathology_tests`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return PathologyTest::$rules;
```

## hms/app/Http/Requests/CreatePathologyUnitRequest.php

Models: `PathologyUnit` → `pathology_units`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return PathologyUnit::$rules;
```

## hms/app/Http/Requests/CreatePatientAdmissionRequest.php

Models: `PatientAdmission` → `patient_admissions`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return PatientAdmission::$rules;
```

## hms/app/Http/Requests/CreatePatientCaseRequest.php

Models: `PatientCase` → `patient_cases`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return PatientCase::$rules;
```

## hms/app/Http/Requests/CreatePatientDiagnosisTestRequest.php

Models: `PatientDiagnosisTest` → `patient_diagnosis_tests`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return PatientDiagnosisTest::$rules;
```

### messages (line 29)

`public function messages(): array`

## hms/app/Http/Requests/CreatePatientIdCardRequest.php

Models:

### authorize (line 12)

`public function authorize(): bool`

Guard/validation candidates:

```php
14: return true;
```

### rules (line 22)

`public function rules(): array`

## hms/app/Http/Requests/CreatePatientIdCardTemplateRequest.php

Models: `PatientIdCardTemplate` → `patient_id_card_templates`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return PatientIdCardTemplate::$rules;
```

## hms/app/Http/Requests/CreatePatientRequest.php

Models: `Patient` → `patients`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Patient::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreatePaymentRequest.php

Models: `Payment` → `payments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Payment::$rules;
```

## hms/app/Http/Requests/CreatePharmacistRequest.php

Models: `Pharmacist` → `pharmacists`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Pharmacist::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreatePrescriptionRequest.php

Models: `Prescription` → `prescriptions`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Prescription::$rules;
```

## hms/app/Http/Requests/CreatePurchaseMedicineRequest.php

Models:

### authorize (line 12)

`public function authorize(): bool`

Guard/validation candidates:

```php
14: return true;
```

### rules (line 18)

`public function rules(): array`

### messages (line 26)

`public function messages(): array`

## hms/app/Http/Requests/CreateRadiologyCategoryRequest.php

Models: `RadiologyCategory` → `radiology_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return RadiologyCategory::$rules;
```

## hms/app/Http/Requests/CreateRadiologyTestRequest.php

Models: `RadiologyTest` → `radiology_tests`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return RadiologyTest::$rules;
```

## hms/app/Http/Requests/CreateReceptionistRequest.php

Models: `Receptionist` → `receptionists`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Receptionist::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/CreateScheduleRequest.php

Models: `Schedule` → `schedules`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Schedule::$rules;
```

## hms/app/Http/Requests/CreateServiceRequest.php

Models: `Service` → `services`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Service::$rules;
```

## hms/app/Http/Requests/CreateSmsRequest.php

Models: `Sms` → `sms`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Sms::$rules;
```

## hms/app/Http/Requests/CreateUserRequest.php

Models: `User` → `users`, `DoctorDepartment` → `doctor_departments`

### authorize (line 17)

`public function authorize(): bool`

Guard/validation candidates:

```php
19: return true;
```

### rules (line 25)

`public function rules(): array`

Guard/validation candidates:

```php
27: $rules = User::$rules;
31: if ($this->department_id == 2) {
32: if (DoctorDepartment::count() == 0) {
```

### messages (line 47)

`public function messages(): array`

## hms/app/Http/Requests/CreateVaccinatedPatientRequest.php

Models: `VaccinatedPatients` → `vaccinated_patients`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return VaccinatedPatients::$rules;
```

## hms/app/Http/Requests/CreateVaccinationRequest.php

Models: `Vaccination` → `vaccinations`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Vaccination::$rules;
```

## hms/app/Http/Requests/CreateVisitorRequest.php

Models: `Visitor` → `visitors`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Visitor::$rules;
```

### messages (line 33)

`public function messages(): array`

## hms/app/Http/Requests/CreateWebAppointmentRequest.php

Models:

### authorize (line 12)

`public function authorize(): bool`

Guard/validation candidates:

```php
14: return true;
```

### rules (line 20)

`public function rules(): array`

Guard/validation candidates:

```php
29: if (request()->get('patient_type') == 1) {
32: if (request()->get('patient_type') == 2) {
```

## hms/app/Http/Requests/CreateZoomCredentialRequest.php

Models: `UserZoomCredential` → `user_zoom_credential`

### authorize (line 16)

`public function authorize(): bool`

Guard/validation candidates:

```php
18: return true;
```

### rules (line 24)

`public function rules(): array`

Guard/validation candidates:

```php
26: return UserZoomCredential::$rules;
```

## hms/app/Http/Requests/Createcurrency_settingRequest.php

Models: `CurrencySetting` → `currency_settings`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return CurrencySetting::$rules;
```

## hms/app/Http/Requests/FrontServiceRequest.php

Models: `FrontService` → `front_services`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = FrontService::$rules;
```

### messages (line 29)

`public function messages(): array`

## hms/app/Http/Requests/LiveConsultationRequest.php

Models: `LiveConsultation` → `live_consultations`

### authorize (line 16)

`public function authorize(): bool`

Guard/validation candidates:

```php
18: return true;
```

### rules (line 24)

`public function rules(): array`

Guard/validation candidates:

```php
26: return LiveConsultation::$rules;
```

## hms/app/Http/Requests/LiveMeetingRequest.php

Models: `LiveMeeting` → `live_meetings`

### authorize (line 16)

`public function authorize(): bool`

Guard/validation candidates:

```php
18: return true;
```

### rules (line 24)

`public function rules(): array`

Guard/validation candidates:

```php
26: return LiveMeeting::$rules;
```

## hms/app/Http/Requests/PostalRequest.php

Models: `Postal` → `postals`

### authorize (line 19)

`public function authorize(): bool`

Guard/validation candidates:

```php
21: return true;
```

### rules (line 27)

`public function rules(): array`

Guard/validation candidates:

```php
29: $rules = Postal::$rules;
```

### messages (line 38)

`public function messages(): array`

Guard/validation candidates:

```php
40: if (Route::current()->getName() == 'dispatches.store') {
```

## hms/app/Http/Requests/StoreOperationCategoryRequest.php

Models: `OperationCategory` → `operation_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return OperationCategory::$rules;
```

## hms/app/Http/Requests/TestimonialRequest.php

Models: `Testimonial` → `testimonials`

### authorize (line 16)

`public function authorize(): bool`

Guard/validation candidates:

```php
18: return true;
```

### rules (line 24)

`public function rules(): array`

Guard/validation candidates:

```php
26: $rules = Testimonial::$rules;
```

### messages (line 35)

`public function messages(): array`

## hms/app/Http/Requests/UpdateAccountRequest.php

Models: `Account` → `accounts`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Account::$rules;
```

### prepareForValidation (line 30)

`protected function prepareForValidation()`

### sanitize (line 35)

`public function sanitize()`

## hms/app/Http/Requests/UpdateAccountantRequest.php

Models: `Accountant` → `accountants`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Accountant::$rules;
```

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/UpdateAdminRequest.php

Models: `admin` → `admins`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = admin::$rules;
```

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/UpdateAdvancedPaymentRequest.php

Models: `AdvancedPayment` → `advanced_payments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = AdvancedPayment::$rules;
```

## hms/app/Http/Requests/UpdateAmbulanceCallRequest.php

Models: `AmbulanceCall` → `ambulance_calls`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = AmbulanceCall::$rules;
```

### messages (line 28)

`public function messages()`

## hms/app/Http/Requests/UpdateAmbulanceRequest.php

Models: `Ambulance` → `ambulances`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Ambulance::$rules;
```

## hms/app/Http/Requests/UpdateAppointmentRequest.php

Models: `Appointment` → `appointments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Appointment::$rules;
```

## hms/app/Http/Requests/UpdateBedAssignRequest.php

Models: `BedAssign` → `bed_assigns`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = BedAssign::$rules;
```

## hms/app/Http/Requests/UpdateBedRequest.php

Models: `Bed` → `beds`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Bed::$rules;
```

## hms/app/Http/Requests/UpdateBedTypeRequest.php

Models: `BedType` → `bed_types`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = BedType::$rules;
```

## hms/app/Http/Requests/UpdateBillRequest.php

Models: `Bill` → `bills`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Bill::$rules;
```

### messages (line 26)

`public function messages(): array`

## hms/app/Http/Requests/UpdateBirthReportRequest.php

Models: `BirthReport` → `birth_reports`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = BirthReport::$rules;
```

## hms/app/Http/Requests/UpdateBloodBankRequest.php

Models: `BloodBank` → `blood_banks`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = BloodBank::$rules;
```

## hms/app/Http/Requests/UpdateBloodDonorRequest.php

Models: `BloodDonor` → `blood_donors`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = BloodDonor::$rules;
```

## hms/app/Http/Requests/UpdateBrandRequest.php

Models: `Brand` → `brands`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Brand::$rules;
```

## hms/app/Http/Requests/UpdateCallLogRequest.php

Models: `CallLog` → `call_logs`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = CallLog::$rules;
```

## hms/app/Http/Requests/UpdateCaseHandlerRequest.php

Models: `CaseHandler` → `case_handlers`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = CaseHandler::$rules;
```

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/UpdateCategoryRequest.php

Models: `Category` → `categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Category::$rules;
```

## hms/app/Http/Requests/UpdateChargeCategoryRequest.php

Models: `ChargeCategory` → `charge_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = ChargeCategory::$rules;
```

## hms/app/Http/Requests/UpdateChargeRequest.php

Models: `Charge` → `charges`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Charge::$rules;
```

## hms/app/Http/Requests/UpdateComplaintRequest.php

Models: `Complaint` → `complaints`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

## hms/app/Http/Requests/UpdateCurrencySettingRequest.php

Models:

### authorize (line 12)

`public function authorize(): bool`

Guard/validation candidates:

```php
14: return true;
```

### rules (line 20)

`public function rules(): array`

## hms/app/Http/Requests/UpdateDeathReportRequest.php

Models: `DeathReport` → `death_reports`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = DeathReport::$rules;
```

## hms/app/Http/Requests/UpdateDepartmentRequest.php

Models: `Department` → `departments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### prepareForValidation (line 18)

`protected function prepareForValidation()`

### rules (line 26)

`public function rules(): array`

Guard/validation candidates:

```php
28: $rules = Department::$rules;
```

### sanitize (line 33)

`public function sanitize()`

## hms/app/Http/Requests/UpdateDiagnosisCategoryRequest.php

Models: `DiagnosisCategory` → `diagnosis_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = DiagnosisCategory::$rules;
```

## hms/app/Http/Requests/UpdateDoctorDepartmentRequest.php

Models: `DoctorDepartment` → `doctor_departments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = DoctorDepartment::$rules;
```

## hms/app/Http/Requests/UpdateDoctorOPDChargeRequest.php

Models: `DoctorOPDCharge` → `doctor_opd_charges`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = DoctorOPDCharge::$rules;
```

### messages (line 29)

`public function messages(): array`

## hms/app/Http/Requests/UpdateDoctorRequest.php

Models: `Doctor` → `doctors`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Doctor::$rules;
```

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/UpdateDocumentRequest.php

Models: `Document` → `documents`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Document::$rules;
```

### messages (line 33)

`public function messages(): array`

## hms/app/Http/Requests/UpdateDocumentTypeRequest.php

Models: `DocumentType` → `document_types`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = DocumentType::$rules;
```

## hms/app/Http/Requests/UpdateEmployeePayrollRequest.php

Models: `EmployeePayroll` → `employee_payrolls`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = EmployeePayroll::$rules;
```

## hms/app/Http/Requests/UpdateExpenseRequest.php

Models: `Expense` → `expenses`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Expense::$rules;
```

### messages (line 33)

`public function messages(): array`

## hms/app/Http/Requests/UpdateFrontServiceRequest.php

Models: `FrontService` → `front_services`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: $rules = FrontService::$rules;
```

## hms/app/Http/Requests/UpdateFrontSettingRequest.php

Models: `FrontSetting` → `front_settings`

### prepareForValidation (line 17)

`public function prepareForValidation()`

Guard/validation candidates:

```php
20: if (empty($description)) {
```

### authorize (line 30)

`public function authorize(): bool`

Guard/validation candidates:

```php
32: return true;
```

### rules (line 38)

`public function rules(): array`

Guard/validation candidates:

```php
40: return FrontSetting::$rules;
```

## hms/app/Http/Requests/UpdateIncomeRequest.php

Models: `Income` → `incomes`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Income::$rules;
```

### messages (line 33)

`public function messages(): array`

## hms/app/Http/Requests/UpdateInsuranceRequest.php

Models: `Insurance` → `insurances`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Insurance::$rules;
```

## hms/app/Http/Requests/UpdateInvestigationReportRequest.php

Models: `InvestigationReport` → `investigation_reports`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = InvestigationReport::$rules;
```

### messages (line 33)

`public function messages(): array`

## hms/app/Http/Requests/UpdateInvoiceRequest.php

Models: `Invoice` → `invoices`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Invoice::$rules;
```

### messages (line 28)

`public function messages(): array`

## hms/app/Http/Requests/UpdateIpdChargeRequest.php

Models: `IpdCharge` → `ipd_charges`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return IpdCharge::$rules;
```

## hms/app/Http/Requests/UpdateIpdConsultantRegisterRequest.php

Models: `IpdConsultantRegister` → `ipd_consultant_registers`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return IpdConsultantRegister::$rules;
```

### messages (line 26)

`public function messages()`

## hms/app/Http/Requests/UpdateIpdDiagnosisRequest.php

Models: `IpdDiagnosis` → `ipd_diagnoses`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = IpdDiagnosis::$rules;
```

## hms/app/Http/Requests/UpdateIpdPatientDepartmentRequest.php

Models: `IpdPatientDepartment` → `ipd_patient_departments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = IpdPatientDepartment::$rules;
25: if($this->route('ipdPatientDepartment')->discharge == 1){
```

### messages (line 35)

`public function messages(): array`

## hms/app/Http/Requests/UpdateIpdPaymentRequest.php

Models: `IpdCharge` → `ipd_charges`, `IpdPatientDepartment` → `ipd_patient_departments`, `IpdPayment` → `ipd_payments`

### authorize (line 15)

`public function authorize(): bool`

Guard/validation candidates:

```php
17: return true;
```

### prepareForValidation (line 23)

`protected function prepareForValidation(): void`

### rules (line 32)

`public function rules(): array`

Guard/validation candidates:

```php
34: $rules = IpdPayment::$rules;
```

## hms/app/Http/Requests/UpdateIpdPrescriptionRequest.php

Models: `IpdPrescription` → `ipd_prescriptions`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = IpdPrescription::$rules;
```

## hms/app/Http/Requests/UpdateIpdTimelineRequest.php

Models: `IpdTimeline` → `ipd_timelines`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = IpdTimeline::$rules;
```

## hms/app/Http/Requests/UpdateIssuedItemRequest.php

Models: `IssuedItem` → `issued_items`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = IssuedItem::$rules;
```

## hms/app/Http/Requests/UpdateItemCategoryRequest.php

Models: `ItemCategory` → `item_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = ItemCategory::$rules;
```

## hms/app/Http/Requests/UpdateItemRequest.php

Models: `Item` → `items`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Item::$rules;
```

## hms/app/Http/Requests/UpdateItemStockRequest.php

Models: `ItemStock` → `item_stocks`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = ItemStock::$rules;
```

## hms/app/Http/Requests/UpdateLabTechnicianRequest.php

Models: `LabTechnician` → `lab_technicians`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = LabTechnician::$rules;
```

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/UpdateLunchBreakRequest.php

Models: `LunchBreak` → `lunch_breaks`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return LunchBreak::$rules;
```

## hms/app/Http/Requests/UpdateMedicineBillRequest.php

Models:

### authorize (line 12)

`public function authorize(): bool`

Guard/validation candidates:

```php
14: return true;
```

### rules (line 22)

`public function rules(): array`

## hms/app/Http/Requests/UpdateMedicineRequest.php

Models: `Medicine` → `medicines`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### prepareForValidation (line 18)

`protected function prepareForValidation()`

### rules (line 26)

`public function rules(): array`

Guard/validation candidates:

```php
28: $rules = Medicine::$rules;
```

### messages (line 34)

`public function messages(): array`

### sanitize (line 42)

`public function sanitize()`

## hms/app/Http/Requests/UpdateNoticeBoardRequest.php

Models: `NoticeBoard` → `notice_boards`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = NoticeBoard::$rules;
```

## hms/app/Http/Requests/UpdateNurseRequest.php

Models: `Nurse` → `nurses`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Nurse::$rules;
```

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/UpdateOdontogramRequest.php

Models: `Odontogram` → `odontograms`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: return Odontogram::$rules;
```

## hms/app/Http/Requests/UpdateOpdDiagnosisRequest.php

Models: `OpdDiagnosis` → `opd_diagnoses`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = OpdDiagnosis::$rules;
```

## hms/app/Http/Requests/UpdateOpdPatientDepartmentRequest.php

Models: `OpdPatientDepartment` → `opd_patient_departments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return OpdPatientDepartment::$rules;
```

### messages (line 26)

`public function messages(): array`

## hms/app/Http/Requests/UpdateOpdPrescriptionRequest.php

Models: `OpdPrescription` → `opd_prescriptions`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: $rules = OpdPrescription::$rules;
```

## hms/app/Http/Requests/UpdateOpdTimelineRequest.php

Models: `OpdTimeline` → `opd_timelines`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = OpdTimeline::$rules;
```

## hms/app/Http/Requests/UpdateOperationCategoryRequest.php

Models: `OperationCategory` → `operation_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: $rules = OperationCategory::$rules;
```

## hms/app/Http/Requests/UpdateOperationReportRequest.php

Models: `OperationReport` → `operation_reports`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = OperationReport::$rules;
```

## hms/app/Http/Requests/UpdateOperationRequest.php

Models: `Operation` → `operations`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: $rules = Operation::$rules;
```

### messages (line 31)

`public function messages(): array`

## hms/app/Http/Requests/UpdatePackageRequest.php

Models: `Package` → `packages`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Package::$rules;
```

## hms/app/Http/Requests/UpdatePathologyCategoryRequest.php

Models: `PathologyCategory` → `pathology_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = PathologyCategory::$rules;
```

## hms/app/Http/Requests/UpdatePathologyParameterRequest.php

Models: `PathologyParameter` → `pathology_parameters`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: $rules = PathologyParameter::$rules;
```

## hms/app/Http/Requests/UpdatePathologyTestRequest.php

Models: `PathologyTest` → `pathology_tests`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = PathologyTest::$rules;
```

## hms/app/Http/Requests/UpdatePathologyUnitRequest.php

Models: `PathologyUnit` → `pathology_units`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: $rules = PathologyUnit::$rules;
```

## hms/app/Http/Requests/UpdatePatientAdmissionRequest.php

Models: `PatientAdmission` → `patient_admissions`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = PatientAdmission::$rules;
```

## hms/app/Http/Requests/UpdatePatientCaseRequest.php

Models: `PatientCase` → `patient_cases`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = PatientCase::$rules;
```

## hms/app/Http/Requests/UpdatePatientDiagnosisTestRequest.php

Models: `PatientDiagnosisTest` → `patient_diagnosis_tests`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = PatientDiagnosisTest::$rules;
```

### messages (line 32)

`public function messages(): array`

## hms/app/Http/Requests/UpdatePatientIdCardTemplateRequest.php

Models: `PatientIdCardTemplate` → `patient_id_card_templates`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 23)

`public function rules(): array`

Guard/validation candidates:

```php
25: $rules = PatientIdCardTemplate::$rules;
```

## hms/app/Http/Requests/UpdatePatientRequest.php

Models: `Patient` → `patients`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Patient::$rules;
```

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/UpdatePaymentRequest.php

Models: `Payment` → `payments`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Payment::$rules;
```

## hms/app/Http/Requests/UpdatePharmacistRequest.php

Models: `Pharmacist` → `pharmacists`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Pharmacist::$rules;
```

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/UpdatePrescriptionRequest.php

Models: `Prescription` → `prescriptions`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Prescription::$rules;
```

## hms/app/Http/Requests/UpdateRadiologyCategoryRequest.php

Models: `RadiologyCategory` → `radiology_categories`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = RadiologyCategory::$rules;
```

## hms/app/Http/Requests/UpdateRadiologyTestRequest.php

Models: `RadiologyTest` → `radiology_tests`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = RadiologyTest::$rules;
```

## hms/app/Http/Requests/UpdateReceptionistRequest.php

Models: `Receptionist` → `receptionists`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Receptionist::$rules;
```

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/UpdateScheduleRequest.php

Models: `Schedule` → `schedules`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Schedule::$rules;
```

## hms/app/Http/Requests/UpdateServiceRequest.php

Models: `Service` → `services`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Service::$rules;
```

## hms/app/Http/Requests/UpdateSettingRequest.php

Models: `Setting` → `settings`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Setting::$rules;
```

## hms/app/Http/Requests/UpdateUserProfileRequest.php

Models: `User` → `users`

### authorize (line 14)

`public function authorize(): bool`

Guard/validation candidates:

```php
16: return true;
```

### rules (line 24)

`public function rules(): array`

### messages (line 38)

`public function messages(): array`

## hms/app/Http/Requests/UpdateUserRequest.php

Models: `User` → `users`

### authorize (line 16)

`public function authorize(): bool`

Guard/validation candidates:

```php
18: return true;
```

### rules (line 24)

`public function rules(): array`

Guard/validation candidates:

```php
26: $rules = User::$rules;
```

### messages (line 34)

`public function messages(): array`

## hms/app/Http/Requests/UpdateVaccinatedPatientRequest.php

Models: `VaccinatedPatients` → `vaccinated_patients`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return VaccinatedPatients::$rules;
```

## hms/app/Http/Requests/UpdateVaccinationRequest.php

Models: `Vaccination` → `vaccinations`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: return Vaccination::$rules;
```

## hms/app/Http/Requests/UpdateVisitorRequest.php

Models: `Visitor` → `visitors`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = Visitor::$rules;
```

### messages (line 33)

`public function messages(): array`

## hms/app/Http/Requests/Updatecurrency_settingRequest.php

Models: `CurrencySetting` → `currency_settings`

### authorize (line 13)

`public function authorize(): bool`

Guard/validation candidates:

```php
15: return true;
```

### rules (line 21)

`public function rules(): array`

Guard/validation candidates:

```php
23: $rules = CurrencySetting::$rules;
```

## hms/app/Repositories/AccountRepository.php

Models: `Account` → `accounts`

### getFieldsSearchable (line 21)

`public function getFieldsSearchable(): array`

### model (line 26)

`public function model()`

## hms/app/Repositories/AccountantRepository.php

Models: `Accountant` → `accountants`, `Address` → `addresses`, `Department` → `departments`, `User` → `users`

### getFieldsSearchable (line 30)

`public function getFieldsSearchable(): array`

### model (line 35)

`public function model()`

### store (line 40)

`public function store($input, $mail = true)`

Guard/validation candidates:

```php
51: if ($mail) {
55: if (isset($input['image']) && ! empty($input['image'])) {
62: if (! empty($address = Address::prepareAddressArray($input))) {
69: return true;
```

Write/transaction candidates:

```php
49: $user = User::create($input);
58: $accountant = Accountant::create(['user_id' => $user->id]);
63: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
66: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
```

### update (line 75)

`public function update($accountant, $input)`

Guard/validation candidates:

```php
82: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
86: if (isset($input['image']) && ! empty($input['image'])) {
95: if (! empty($accountant->address)) {
96: if (empty($address = Address::prepareAddressArray($input))) {
101: if (! empty($address = Address::prepareAddressArray($input)) && empty($accountant->address)) {
108: return true;
```

Write/transaction candidates:

```php
92: $accountant->user->update($input);
93: $accountant->update($input);
97: $accountant->address->delete();
99: $accountant->address->update($input);
104: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
```

## hms/app/Repositories/AdvancedPaymentRepository.php

Models: `AdvancedPayment` → `advanced_payments`, `Notification` → `notifications`, `Patient` → `patients`

### getFieldsSearchable (line 25)

`public function getFieldsSearchable(): array`

### model (line 30)

`public function model()`

### getPatients (line 35)

`public function getPatients()`

### createNotification (line 42)

`public function createNotification($input)`

## hms/app/Repositories/AmbulanceCallRepository.php

Models: `Ambulance` → `ambulances`, `AmbulanceCall` → `ambulance_calls`

### getFieldsSearchable (line 23)

`public function getFieldsSearchable()`

### model (line 28)

`public function model()`

### update (line 33)

`public function update($input, $ambulanceCall)`

Guard/validation candidates:

```php
38: if ($input['ambulance_id'] == $ambulanceId) {
39: return true;
45: return true;
```

Write/transaction candidates:

```php
36: $ambulanceCall->update($input);
42: Ambulance::where('id', $ambulanceId)->update(['is_available' => true]);
43: Ambulance::where('id', $input['ambulance_id'])->update(['is_available' => false]);
```

## hms/app/Repositories/AmbulanceRepository.php

Models: `Ambulance` → `ambulances`, `CaseHandler` → `case_handlers`, `Notification` → `notifications`, `Receptionist` → `receptionists`, `User` → `users`

### getFieldsSearchable (line 30)

`public function getFieldsSearchable()`

### model (line 35)

`public function model()`

### getAmbulances (line 40)

`public function getAmbulances()`

### createNotification (line 47)

`public function createNotification()`

Guard/validation candidates:

```php
65: return true;
```

## hms/app/Repositories/AppointmentCalendarRepository.php

Models: `Appointment` → `appointments`, `User` → `users`

### model (line 16)

`public function model()`

### getAppointments (line 21)

`public function getAppointments()`

Guard/validation candidates:

```php
27: if ($user->hasRole('Doctor')) {
31: if ($user->hasRole(['Patient'])) {
```

## hms/app/Repositories/AppointmentRepository.php

Models: `Appointment` → `appointments`, `AppointmentTransaction` → `appointment_transactions`, `Department` → `departments`, `Doctor` → `doctors`, `DoctorDepartment` → `doctor_departments`, `Notification` → `notifications`, `Patient` → `patients`, `Receptionist` → `receptionists`, `Schedule` → `schedules`, `User` → `users`

### getFieldsSearchable (line 47)

`public function getFieldsSearchable()`

### model (line 52)

`public function model()`

### getPatients (line 57)

`public function getPatients()`

### getDoctors (line 65)

`public function getDoctors($id)`

### getDoctorDepartments (line 73)

`public function getDoctorDepartments()`

### getDoctorDepartmentForAPI (line 80)

`public function getDoctorDepartmentForAPI()`

### getDoctorDepartment (line 87)

`public function getDoctorDepartment($id)`

### sendAppointmentReminder (line 94)

`public function sendAppointmentReminder()`

Write/transaction candidates:

```php
122: $client->messages->create(
```

### getBookingSlot (line 135)

`public function getBookingSlot($inputs)`

Guard/validation candidates:

```php
147: if (isset($inputs['editId'])) {
```

### getBookingSlotAPI (line 156)

`public function getBookingSlotAPI($input)`

### appointmentParseIn (line 174)

`public function appointmentParseIn($time)`

### appointmentGetTimeIntervals (line 184)

`public function appointmentGetTimeIntervals($start_time, $end_time, $per_patient_time)`

### getTime (line 197)

`public function getTime($inputs)`

### sendAppointmentEmailBeforeOneHour (line 211)

`public function sendAppointmentEmailBeforeOneHour()`

Guard/validation candidates:

```php
240: return true;
```

### createNotification (line 243)

`public function createNotification($input)`

Guard/validation candidates:

```php
262: if ($notification == Notification::NOTIFICATION_FOR[Notification::PATIENT]) {
```

### createNewAppointment (line 279)

`public function createNewAppointment($input)`

Guard/validation candidates:

```php
295: if (strpos($key, 'field') === 0) {
301: if (isset($input['email'])) {
```

Write/transaction candidates:

```php
282: DB::beginTransaction();
300: $user = User::create($userData);
305: $patient = Patient::create(['user_id' => $user->id,'patient_unique_id' => strtoupper(Patient::generateUniquePatientId())]);
309: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
312: $appointment = Appointment::create([
322: DB::commit();
326: DB::rollBack();
```

### getDoctorLists (line 331)

`public function getDoctorLists()`

### getDoctorList (line 339)

`public function getDoctorList($id)`

### filter (line 348)

`public function filter($status)`

Guard/validation candidates:

```php
350: if ($status == 'all') {
365: return false;
```

### getDepartmentDoctorList (line 369)

`public function getDepartmentDoctorList($id)`

### getDoctorsAppointmentCharge (line 383)

`public function getDoctorsAppointmentCharge($id)`

## hms/app/Repositories/AppointmentTransactionRepository.php

Models: `Appointment` → `appointments`, `AppointmentTransaction` → `appointment_transactions`

### getFieldsSearchable (line 24)

`public function getFieldsSearchable(): array`

### model (line 29)

`public function model(): string`

### store (line 34)

`public function store($input)`

Guard/validation candidates:

```php
40: return true;
```

Write/transaction candidates:

```php
38: $appointment->update(['payment_status' => 1]);
```

### WebAppointmentstripeSession (line 48)

`public function WebAppointmentstripeSession($input)`

Write/transaction candidates:

```php
59: $session = Session::create([
```

### stripeSession (line 87)

`public function stripeSession($input)`

Write/transaction candidates:

```php
98: $session = Session::create([
```

### appointmentStripePaymentSuccess (line 127)

`public function appointmentStripePaymentSuccess($input)`

Guard/validation candidates:

```php
130: if (empty($sessionId)) {
155: return true;
```

Write/transaction candidates:

```php
140: DB::beginTransaction();
142: $appointmentTransaction = AppointmentTransaction::create([
152: $appointment->update(['is_completed' => false,'payment_status' => 1,'payment_type' => \App\Models\Appointment::TYPE_STRIPE]);
154: DB::commit();
157: DB::rollBack();
```

### TransactionRazorpayPayment (line 162)

`public function TransactionRazorpayPayment($input)`

Write/transaction candidates:

```php
181: $razorpayOrder = $api->order->create($orderData);
```

### TransactionRazorpayPaymentSuccess (line 191)

`public function TransactionRazorpayPaymentSuccess($input)`

Guard/validation candidates:

```php
197: if (count($input) && ! empty($input['razorpay_payment_id'])) {
205: if ($generatedSignature != $input['razorpay_signature']) {
222: return true;
227: return false;
```

Write/transaction candidates:

```php
199: DB::beginTransaction();
210: $appointmentTransaction = AppointmentTransaction::create([
219: $appointment->update(['is_completed' => false,'payment_status' => 1,'payment_type' => \App\Models\Appointment::TYPE_RAZORPAY]);
221: DB::commit();
224: DB::rollBack();
```

### paypalPaymentSuccess (line 230)

`public function paypalPaymentSuccess($response)`

Write/transaction candidates:

```php
234: DB::beginTransaction();
247: $transaction = AppointmentTransaction::create($transactionData);
250: $appointment->update(['is_completed' => false,'payment_status' => 1,'payment_type' => \App\Models\Appointment::TYPE_PAYPAL]);
252: DB::commit();
254: DB::rollBack();
```

### flutterWavePayment (line 260)

`public function flutterWavePayment($input)`

Guard/validation candidates:

```php
283: if ($payment['status'] !== 'success') {
285: if(isset($input['web_appointment']) && $input['web_appointment'] == true){
```

### flutterwavePaymentSuccess (line 302)

`public function flutterwavePaymentSuccess($input)`

Guard/validation candidates:

```php
309: if ($input['status'] ==  'successful')
329: return true;
```

Write/transaction candidates:

```php
305: DB::beginTransaction();
314: $appointment = Appointment::create($sessionData);
316: $appointmentTransaction = AppointmentTransaction::create([
325: $appointment->update(['is_completed' => 1,'payment_status' => 1,'payment_type' => \App\Models\Appointment::FLUTTERWAVE]);
327: DB::commit();
332: DB::rollBack();
```

### phonePePayment (line 338)

`public function phonePePayment($input)`

### phonePePaymentSuccess (line 410)

`public function phonePePaymentSuccess($input)`

Guard/validation candidates:

```php
474: return true;
479: return false;
```

Write/transaction candidates:

```php
457: DB::beginTransaction();
459: $appointment = Appointment::create($input['input']);
461: $appointmentTransaction = AppointmentTransaction::create([
470: $appointment->update(['is_completed' => 1,'payment_status' => 1,'payment_type' => \App\Models\Appointment::PHONEPE]);
472: DB::commit();
476: DB::rollBack();
```

### payStackPaymentSuccess (line 482)

`public function payStackPaymentSuccess($input)`

Guard/validation candidates:

```php
504: return true;
```

Write/transaction candidates:

```php
485: DB::beginTransaction();
489: $appointment = Appointment::create($sessionData['data']);
491: $appointmentTransaction = AppointmentTransaction::create([
500: $appointment->update(['is_completed' => 1,'payment_status' => 1,'payment_type' => \App\Models\Appointment::PAYSTACK]);
502: DB::commit();
507: DB::rollBack();
```

## hms/app/Repositories/BaseRepository.php

Models:

### __construct (line 14)

`public function __construct(Application $app)`

### makeModel (line 24)

`public function makeModel()`

Guard/validation candidates:

```php
28: if (! $model instanceof Model) {
```

### paginate (line 35)

`public function paginate($perPage, $columns = ['*'])`

### allQuery (line 42)

`public function allQuery($search = [], $skip = null, $limit = null)`

Guard/validation candidates:

```php
46: if (count($search)) {
48: if (in_array($key, $this->getFieldsSearchable())) {
54: if (! is_null($skip)) {
58: if (! is_null($limit)) {
```

### all (line 65)

`public function all($search = [], $skip = null, $limit = null, $columns = ['*'])`

### create (line 72)

`public function create($input)`

Write/transaction candidates:

```php
76: $model->save();
```

### find (line 81)

`public function find($id, $columns = ['*'])`

### update (line 88)

`public function update($input, $id)`

Write/transaction candidates:

```php
96: $model->save();
```

### delete (line 101)

`public function delete($id)`

Write/transaction candidates:

```php
107: return $model->delete();
```

## hms/app/Repositories/BedAssignRepository.php

Models: `Bed` → `beds`, `BedAssign` → `bed_assigns`, `Doctor` → `doctors`, `IpdPatientDepartment` → `ipd_patient_departments`, `Notification` → `notifications`, `Nurse` → `nurses`, `PatientCase` → `patient_cases`, `User` → `users`

### getFieldsSearchable (line 30)

`public function getFieldsSearchable()`

### model (line 35)

`public function model()`

### getBeds (line 40)

`public function getBeds()`

### getCases (line 48)

`public function getCases()`

Guard/validation candidates:

```php
52: if ($user->hasRole('Doctor')) {
```

### getIpdPatients (line 69)

`public function getIpdPatients($caseId)`

### store (line 76)

`public function store($input)`

Guard/validation candidates:

```php
87: return true;
```

Write/transaction candidates:

```php
82: BedAssign::create($input);
85: $bed->update(['is_available' => 0]);
```

### update (line 93)

`public function update($input, $bedAssign)`

Guard/validation candidates:

```php
98: if($oldBed){
107: if (isset($bedId)) {
115: return true;
```

Write/transaction candidates:

```php
99: $oldBed->ipdPatient()->update(['bed_id' => $input['bed_id']]);
100: $oldBed->bed()->update(['is_available' => 1]);
105: $bedAssign->update($input);
```

### getPatientBeds (line 121)

`public function getPatientBeds($bedAssign)`

### getPatientCases (line 130)

`public function getPatientCases($bedAssign)`

### createNotification (line 143)

`public function createNotification($input)`

## hms/app/Repositories/BedRepository.php

Models: `Bed` → `beds`, `BedType` → `bed_types`

### getFieldsSearchable (line 23)

`public function getFieldsSearchable()`

### model (line 28)

`public function model()`

### getBedTypes (line 33)

`public function getBedTypes()`

### getAssociateBedsList (line 40)

`public function getAssociateBedsList()`

### store (line 55)

`public function store($input)`

Guard/validation candidates:

```php
61: return true;
```

Write/transaction candidates:

```php
59: $bed = Bed::create($input);
```

### storeBulkBeds (line 67)

`public function storeBulkBeds($input)`

Guard/validation candidates:

```php
79: return true;
```

Write/transaction candidates:

```php
76: Bed::create($data);
```

## hms/app/Repositories/BedTypeRepository.php

Models: `BedType` → `bed_types`

### getFieldsSearchable (line 19)

`public function getFieldsSearchable()`

### model (line 24)

`public function model()`

## hms/app/Repositories/BillItemsRepository.php

Models: `Bill` → `bills`, `BillItems` → `bill_items`

### getFieldsSearchable (line 23)

`public function getFieldsSearchable()`

### model (line 28)

`public function model()`

### updateBillItem (line 33)

`public function updateBillItem($billItemInput, $billId)`

Guard/validation candidates:

```php
39: if (isset($data['id']) && ! empty($data['id'])) {
49: if (! (isset($billItemIds) && count($billItemIds))) {
```

Write/transaction candidates:

```php
41: $this->update($data, $data['id']);
44: $billItem = $bill->billItems()->save($billItem);
52: BillItems::whereNotIn('id', $billItemIds)->whereBillId($bill->id)->delete();
```

## hms/app/Repositories/BillRepository.php

Models: `Accountant` → `accountants`, `Bill` → `bills`, `BillItems` → `bill_items`, `Doctor` → `doctors`, `Medicine` → `medicines`, `Notification` → `notifications`, `Package` → `packages`, `Patient` → `patients`, `PatientAdmission` → `patient_admissions`, `Receptionist` → `receptionists`, `Setting` → `settings`, `User` → `users`

### getFieldsSearchable (line 35)

`public function getFieldsSearchable()`

### model (line 40)

`public function model()`

### getSyncList (line 45)

`public function getSyncList($isEditScreen)`

### getPatientList (line 54)

`public function getPatientList()`

### getAssociateMedicinesList (line 61)

`public function getAssociateMedicinesList()`

### getMedicinesList (line 76)

`public function getMedicinesList()`

### getPatientAdmissionIdList (line 85)

`public function getPatientAdmissionIdList($isEditScreen = false)`

Guard/validation candidates:

```php
90: if ($isEditScreen) {
```

### saveBill (line 106)

`public function saveBill($input)`

Guard/validation candidates:

```php
117: $validator = Validator::make($data, BillItems::$rules);
119: if ($validator->fails()) {
```

Write/transaction candidates:

```php
111: $bill = $this->create($input);
127: $bill->billItems()->save($billItem);
130: $bill->save();
```

### prepareInputForBillItem (line 135)

`public function prepareInputForBillItem($input)`

Guard/validation candidates:

```php
141: if (! (isset($items[$index]['price']) && $key == 'price')) {
```

### updateBill (line 151)

`public function updateBill($billId, $input)`

Guard/validation candidates:

```php
164: $validator = Validator::make($data, BillItems::$rules);
166: if ($validator->fails()) {
```

Write/transaction candidates:

```php
155: $bill = $this->update($input, $billId);
159: $billItem->delete();
179: $bill->save();
```

### patientAdmissionDetails (line 184)

`public function patientAdmissionDetails($inputs)`

Guard/validation candidates:

```php
195: if (isset($patientAdmission->package_id)) {
203: if (isset($inputs['editBillId'])) {
205: if (count($billGet) > 0) {
```

### getSyncListForCreate (line 215)

`public function getSyncListForCreate()`

### saveNotification (line 222)

`public function saveNotification($input)`

Guard/validation candidates:

```php
246: if ($notification == Notification::NOTIFICATION_FOR[Notification::PATIENT]) {
```

## hms/app/Repositories/BirthReportRepository.php

Models: `BirthReport` → `birth_reports`, `Doctor` → `doctors`, `PatientCase` → `patient_cases`

### getFieldsSearchable (line 27)

`public function getFieldsSearchable()`

### model (line 32)

`public function model()`

### getCases (line 37)

`public function getCases()`

Guard/validation candidates:

```php
41: if ($user->hasRole('Doctor')) {
```

### getDoctors (line 57)

`public function getDoctors()`

### store (line 64)

`public function store($input)`

Guard/validation candidates:

```php
72: return true;
```

Write/transaction candidates:

```php
70: $birthReport = BirthReport::create($input);
```

### update (line 78)

`public function update($input, $birthReport)`

Guard/validation candidates:

```php
87: return true;
```

Write/transaction candidates:

```php
85: $birthReport->update($input);
```

## hms/app/Repositories/BloodBankRepository.php

Models: `BloodBank` → `blood_banks`

### getFieldsSearchable (line 19)

`public function getFieldsSearchable()`

### model (line 24)

`public function model()`

## hms/app/Repositories/BloodDonationRepository.php

Models: `BloodDonation` → `blood_donations`

### getFieldsSearchable (line 17)

`public function getFieldsSearchable()`

### model (line 22)

`public function model()`

### createBloodDonation (line 27)

`public function createBloodDonation($input)`

Write/transaction candidates:

```php
30: $bloodDonation = $this->create($input);
35: $bloodBank->update(['remained_bags' => $remainedBags]);
37: $bloodDonation->bloodDonor->update(['last_donate_date' => $bloodDonation->created_at]);
```

### updateBloodDonation (line 43)

`public function updateBloodDonation($input, $bloodDonation)`

Write/transaction candidates:

```php
49: $bloodDonation = $this->update($input, $bloodDonation->id);
54: $bloodBank->update(['remained_bags' => $remainedBags]);
```

## hms/app/Repositories/BloodDonorRepository.php

Models: `BloodDonor` → `blood_donors`

### getFieldsSearchable (line 22)

`public function getFieldsSearchable()`

### model (line 27)

`public function model()`

## hms/app/Repositories/BloodIssueRepository.php

Models: `BloodDonor` → `blood_donors`, `BloodIssue` → `blood_issues`

### getFieldsSearchable (line 22)

`public function getFieldsSearchable()`

### model (line 27)

`public function model()`

### getBloodGroup (line 32)

`public function getBloodGroup($id)`

## hms/app/Repositories/BrandRepository.php

Models: `Brand` → `brands`

### getFieldsSearchable (line 21)

`public function getFieldsSearchable()`

### model (line 26)

`public function model()`

## hms/app/Repositories/CallLogRepository.php

Models: `CallLog` → `call_logs`

### getFieldsSearchable (line 18)

`public function getFieldsSearchable()`

### model (line 23)

`public function model()`

## hms/app/Repositories/CaseHandlerRepository.php

Models: `Address` → `addresses`, `CaseHandler` → `case_handlers`, `Department` → `departments`, `User` → `users`

### getFieldsSearchable (line 24)

`public function getFieldsSearchable()`

### model (line 29)

`public function model()`

### store (line 34)

`public function store($input, $mail = true)`

Guard/validation candidates:

```php
44: if ($mail) {
48: if (isset($input['image']) && ! empty($input['image'])) {
56: if (! empty($address = Address::prepareAddressArray($input))) {
64: return true;
```

Write/transaction candidates:

```php
42: $user = User::create($input);
52: $caseHandler = CaseHandler::create(['user_id' => $user->id]);
57: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
60: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
```

### update (line 70)

`public function update($caseHandler, $input)`

Guard/validation candidates:

```php
77: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
81: if (isset($input['image']) && ! empty($input['image'])) {
90: if (! empty($caseHandler->address)) {
91: if (empty($address = Address::prepareAddressArray($input))) {
96: if (! empty($address = Address::prepareAddressArray($input)) && empty($caseHandler->address)) {
103: return true;
```

Write/transaction candidates:

```php
87: $caseHandler->user->update($input);
88: $caseHandler->update($input);
92: $caseHandler->address->delete();
94: $caseHandler->address->update($input);
99: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
```

## hms/app/Repositories/CategoryRepository.php

Models: `Category` → `categories`

### getFieldsSearchable (line 18)

`public function getFieldsSearchable()`

### model (line 23)

`public function model()`

## hms/app/Repositories/ChargeCategoryRepository.php

Models: `ChargeCategory` → `charge_categories`

### getFieldsSearchable (line 19)

`public function getFieldsSearchable()`

### model (line 24)

`public function model()`

## hms/app/Repositories/ChargeRepository.php

Models: `Charge` → `charges`

### getFieldsSearchable (line 19)

`public function getFieldsSearchable()`

### model (line 24)

`public function model()`

## hms/app/Repositories/ComplaintRepository.php

Models: `Complaint` → `complaints`

### getFieldsSearchable (line 19)

`public function getFieldsSearchable()`

### model (line 24)

`public function model()`

## hms/app/Repositories/CurrencySettingRepository.php

Models: `CurrencySetting` → `currency_settings`

### getFieldsSearchable (line 18)

`public function getFieldsSearchable()`

### model (line 23)

`public function model()`

## hms/app/Repositories/DashboardRepository.php

Models: `Expense` → `expenses`, `Income` → `incomes`

### getIncomeExpenseReport (line 15)

`public function getIncomeExpenseReport($input)`

Guard/validation candidates:

```php
29: if ($chartDates == $incomeDates) {
44: if ($chartDates == $expenseDates) {
```

### getDate (line 58)

`public function getDate($startDate, $endDate)`

Guard/validation candidates:

```php
64: if (!($startDate && $endDate)) {
```

### incomeChartData (line 98)

`public function incomeChartData()`

Write/transaction candidates:

```php
100: $periods = CarbonPeriod::create(Carbon::now()->startOfYear(), '1 month', Carbon::now()->endOfYear());
```

### totalIncomeFilterReport (line 125)

`public function totalIncomeFilterReport($date)`

### totalExpenseFilterReport (line 132)

`public function totalExpenseFilterReport($date)`

## hms/app/Repositories/DeathReportRepository.php

Models: `DeathReport` → `death_reports`, `Doctor` → `doctors`, `PatientCase` → `patient_cases`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable()`

### model (line 33)

`public function model()`

### getCases (line 38)

`public function getCases()`

Guard/validation candidates:

```php
41: if ($user->hasRole('Doctor')) {
```

### getDoctors (line 56)

`public function getDoctors()`

### store (line 63)

`public function store($input)`

Guard/validation candidates:

```php
71: return true;
```

Write/transaction candidates:

```php
69: $deathReport = DeathReport::create($input);
```

### update (line 77)

`public function update($input, $deathReport)`

Guard/validation candidates:

```php
86: return true;
```

Write/transaction candidates:

```php
84: $deathReport->update($input);
```

## hms/app/Repositories/DepartmentRepository.php

Models: `Department` → `departments`

### getFieldsSearchable (line 19)

`public function getFieldsSearchable(): array`

### model (line 24)

`public function model()`

## hms/app/Repositories/DiagnosisCategoryRepository.php

Models: `DiagnosisCategory` → `diagnosis_categories`

### getFieldsSearchable (line 14)

`public function getFieldsSearchable()`

### model (line 19)

`public function model()`

## hms/app/Repositories/DoctorDepartmentRepository.php

Models: `DoctorDepartment` → `doctor_departments`

### getFieldsSearchable (line 19)

`public function getFieldsSearchable(): array`

### model (line 24)

`public function model()`

## hms/app/Repositories/DoctorOPDChargeRepository.php

Models: `Doctor` → `doctors`, `DoctorOPDCharge` → `doctor_opd_charges`

### getFieldsSearchable (line 15)

`public function getFieldsSearchable()`

### model (line 20)

`public function model()`

### getDoctors (line 25)

`public function getDoctors()`

## hms/app/Repositories/DoctorRepository.php

Models: `Address` → `addresses`, `Department` → `departments`, `Doctor` → `doctors`, `Schedule` → `schedules`, `User` → `users`

### getFieldsSearchable (line 29)

`public function getFieldsSearchable()`

### model (line 34)

`public function model()`

### store (line 39)

`public function store($input, $mail = true)`

Guard/validation candidates:

```php
50: if ($mail) {
54: if (isset($input['image']) && !empty($input['image'])) {
76: if (!empty($address = Address::prepareAddressArray($input))) {
86: return true;
```

Write/transaction candidates:

```php
48: $user = User::create(Arr::except($input, ['specialist', 'doctor_department_id']));
58: $doctor = Doctor::create([
66: $schedule = Schedule::create([
77: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
80: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
```

### update (line 89)

`public function update($doctor, $input)`

Guard/validation candidates:

```php
96: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && !empty($input['avatar_remove'])) {
100: if (isset($input['image']) && !empty($input['image'])) {
110: if (!empty($doctor->address)) {
111: if (empty($address = Address::prepareAddressArray($input))) {
116: if (!empty($address = Address::prepareAddressArray($input)) && empty($doctor->address)) {
123: return true;
```

Write/transaction candidates:

```php
107: $doctor->doctorUser->update($input);
108: $doctor->update($input);
112: $doctor->address->delete();
114: $doctor->address->update($input);
119: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
```

### getDoctors (line 129)

`public function getDoctors()`

### getDoctorAssociatedData (line 139)

`public function getDoctorAssociatedData($doctorId)`

Guard/validation candidates:

```php
145: if (!$data['doctorData']) {
146: return false;
```

## hms/app/Repositories/DocumentRepository.php

Models: `Document` → `documents`, `DocumentType` → `document_types`, `Patient` → `patients`

### getFieldsSearchable (line 26)

`public function getFieldsSearchable()`

### model (line 31)

`public function model()`

### getSyncList (line 36)

`public function getSyncList()`

Guard/validation candidates:

```php
39: if ($user->hasRole('Doctor')) {
42: if (! $user->hasRole('Patient')) {
```

### store (line 52)

`public function store($input)`

Guard/validation candidates:

```php
56: if (getLoggedinPatient()) {
61: if (isset($input['file']) && ! empty($input['file'])) {
```

Write/transaction candidates:

```php
60: $document = $this->create($input);
```

### updateDocument (line 69)

`public function updateDocument($input, $documentId)`

Guard/validation candidates:

```php
73: if (isset($input['file']) && ! empty($input['file'])) {
74: if ($document->media->first()) {
```

Write/transaction candidates:

```php
72: $document = $this->update($input, $documentId);
78: $document->update(['updated_at' => Carbon::now()->timestamp]);
```

### deleteDocument (line 85)

`public function deleteDocument($documentId)`

Guard/validation candidates:

```php
89: if ($document->media->first()) {
```

Write/transaction candidates:

```php
92: $this->delete($documentId);
```

## hms/app/Repositories/DocumentTypeRepository.php

Models: `DocumentType` → `document_types`

### getFieldsSearchable (line 18)

`public function getFieldsSearchable()`

### model (line 23)

`public function model()`

## hms/app/Repositories/EmployeePayrollRepository.php

Models: `Accountant` → `accountants`, `EmployeePayroll` → `employee_payrolls`, `Notification` → `notifications`, `User` → `users`

### getFieldsSearchable (line 34)

`public function getFieldsSearchable()`

### model (line 39)

`public function model()`

### create (line 44)

`public function create($input)`

Guard/validation candidates:

```php
54: return true;
```

Write/transaction candidates:

```php
52: parent::create($input);
```

### update (line 57)

`public function update($input, $id)`

Guard/validation candidates:

```php
67: return true;
```

### createNotification (line 70)

`public function createNotification($input)`

Guard/validation candidates:

```php
86: if ($notification == Notification::NOTIFICATION_FOR[EmployeePayroll::TYPES[$input['type']]]) {
```

## hms/app/Repositories/EnquiryRepository.php

Models: `Enquiry` → `enquiries`

### getFieldsSearchable (line 26)

`public function getFieldsSearchable()`

### model (line 31)

`public function model()`

### store (line 36)

`public function store($input)`

Guard/validation candidates:

```php
50: return true;
```

Write/transaction candidates:

```php
39: $enquiry = Enquiry::create($input);
```

## hms/app/Repositories/ExpenseRepository.php

Models: `Accountant` → `accountants`, `Expense` → `expenses`, `Notification` → `notifications`, `User` → `users`

### getFieldsSearchable (line 26)

`public function getFieldsSearchable()`

### model (line 31)

`public function model()`

### store (line 36)

`public function store($input)`

Guard/validation candidates:

```php
40: if (! empty($input['attachment'])) {
44: return true;
```

Write/transaction candidates:

```php
39: $expense = $this->create($input);
```

### updateExpense (line 50)

`public function updateExpense($input, $expenseId)`

Guard/validation candidates:

```php
56: if (! empty($input['attachment'])) {
60: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
```

Write/transaction candidates:

```php
54: $expense = $this->update($input, $expenseId);
```

### deleteDocument (line 68)

`public function deleteDocument($expenseId)`

Write/transaction candidates:

```php
73: $this->delete($expenseId);
```

### createNotification (line 79)

`public function createNotification($input)`

Guard/validation candidates:

```php
99: return true;
```

## hms/app/Repositories/FrontServiceRepository.php

Models: `FrontService` → `front_services`

### getFieldsSearchable (line 15)

`public function getFieldsSearchable()`

### model (line 20)

`public function model()`

### store (line 25)

`public function store($input)`

Guard/validation candidates:

```php
30: if (isset($input['icon']) && ! empty($input['icon'])) {
35: return true;
```

Write/transaction candidates:

```php
28: $frontService = FrontService::create($input);
```

### updateFrontService (line 41)

`public function updateFrontService($input, $frontServiceId)`

Guard/validation candidates:

```php
46: if (isset($input['icon']) && ! empty($input['icon'])) {
```

Write/transaction candidates:

```php
44: $frontService = $this->update($input, $frontServiceId);
```

## hms/app/Repositories/FrontSettingRepository.php

Models: `FrontSetting` → `front_settings`

### getFieldsSearchable (line 16)

`public function getFieldsSearchable()`

### model (line 21)

`public function model()`

### updateFrontSetting (line 26)

`public function updateFrontSetting($input)`

Guard/validation candidates:

```php
28: if (isset($input['about_us_image']) && ! empty($input['about_us_image'])) {
37: if (isset($input['home_page_image']) && ! empty($input['home_page_image'])) {
46: if (isset($input['home_page_certified_doctor_image']) && ! empty($input['home_page_certified_doctor_image'])) {
```

Write/transaction candidates:

```php
35: $frontSetting->update(['value' => $frontSetting->logo_url]);
44: $frontSetting->update(['value' => $frontSetting->logo_url]);
53: $frontSetting->update(['value' => $frontSetting->logo_url]);
```

## hms/app/Repositories/GeneratePatientIdCardRepository.php

Models: `Patient` → `patients`, `PatientIdCardTemplate` → `patient_id_card_templates`, `PatientAdmission` → `patient_admissions`

### getFieldsSearchable (line 19)

`public function getFieldsSearchable(): array`

### model (line 24)

`public function model()`

### getTemplates (line 29)

`public function getTemplates()`

### getPatients (line 34)

`public function getPatients()`

### store (line 39)

`public function store($input)`

Guard/validation candidates:

```php
41: if ($input['type'] == '1') {
54: return true;
```

Write/transaction candidates:

```php
51: $patient->update(['template_id' => $input['template_id'], 'patient_unique_id' => $uniqueId]);
```

## hms/app/Repositories/HolidayRepository.php

Models: `DoctorHoliday` → `doctor_holidays`

### getFieldsSearchable (line 20)

`public function getFieldsSearchable(): array`

### model (line 25)

`public function model()`

### store (line 30)

`public function store($input)`

Guard/validation candidates:

```php
35: if (! $doctor_holiday) {
38: return true;
40: return false;
```

Write/transaction candidates:

```php
36: DoctorHoliday::create($input);
```

## hms/app/Repositories/IncomeRepository.php

Models: `Accountant` → `accountants`, `Income` → `incomes`, `Notification` → `notifications`, `User` → `users`

### getFieldsSearchable (line 23)

`public function getFieldsSearchable()`

### model (line 28)

`public function model()`

### store (line 33)

`public function store($input)`

Guard/validation candidates:

```php
38: if (! empty($input['attachment'])) {
42: return true;
```

Write/transaction candidates:

```php
36: $income = $this->create($input);
```

### updateExpense (line 48)

`public function updateExpense($input, $incomeId)`

Guard/validation candidates:

```php
54: if (! empty($input['attachment'])) {
58: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
```

Write/transaction candidates:

```php
52: $income = $this->update($input, $incomeId);
```

### deleteDocument (line 66)

`public function deleteDocument($incomeId)`

Write/transaction candidates:

```php
71: $this->delete($incomeId);
```

### createNotification (line 77)

`public function createNotification($input)`

Guard/validation candidates:

```php
99: return true;
```

## hms/app/Repositories/InsuranceRepository.php

Models: `Insurance` → `insurances`, `InsuranceDisease` → `insurance_diseases`

### getFieldsSearchable (line 26)

`public function getFieldsSearchable()`

### model (line 31)

`public function model()`

### store (line 36)

`public function store($input)`

Guard/validation candidates:

```php
46: $validator = Validator::make($data, InsuranceDisease::$rules);
48: if ($validator->fails()) {
55: return true;
```

Write/transaction candidates:

```php
40: $insurance = Insurance::create(Arr::except($input, ['disease_name', 'disease_charge']));
52: $disease = InsuranceDisease::create($data);
```

### prepareInputForDiseaseItem (line 58)

`public function prepareInputForDiseaseItem($input)`

### getDisease (line 70)

`public function getDisease()`

### getInsuranceDisease (line 77)

`public function getInsuranceDisease($insuranceId)`

### update (line 84)

`public function update($insurance, $input)`

Guard/validation candidates:

```php
96: $validator = Validator::make($data, InsuranceDisease::$rules);
98: if ($validator->fails()) {
105: return true;
```

Write/transaction candidates:

```php
88: $insurance->update($input);
91: $disease->delete();
102: InsuranceDisease::create($data);
```

### delete (line 108)

`public function delete($insuranceId)`

Guard/validation candidates:

```php
116: return true;
```

Write/transaction candidates:

```php
112: $insurance->delete();
114: $insuranceDisease->delete();
```

## hms/app/Repositories/InvestigationReportRepository.php

Models: `Doctor` → `doctors`, `InvestigationReport` → `investigation_reports`, `Patient` → `patients`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable()`

### model (line 33)

`public function model()`

### getPatients (line 38)

`public function getPatients()`

Guard/validation candidates:

```php
41: if ($user->hasRole('Doctor')) {
```

### getDoctors (line 53)

`public function getDoctors()`

### store (line 60)

`public function store($input)`

Guard/validation candidates:

```php
65: if (! empty($input['attachment'])) {
70: return true;
```

Write/transaction candidates:

```php
63: $report = InvestigationReport::create($input);
```

### update (line 76)

`public function update($input, $id)`

Guard/validation candidates:

```php
82: if (! empty($input['attachment'])) {
87: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
91: return true;
```

Write/transaction candidates:

```php
80: $report->update($input);
```

## hms/app/Repositories/InvoiceItemRepository.php

Models: `Invoice` → `invoices`, `InvoiceItem` → `invoice_items`

### getFieldsSearchable (line 23)

`public function getFieldsSearchable()`

### model (line 28)

`public function model()`

### updateInvoiceItem (line 33)

`public function updateInvoiceItem($invoiceItemInput, $invoiceId)`

Guard/validation candidates:

```php
39: if (isset($data['id']) && ! empty($data['id'])) {
49: if (! (isset($invoiceItemIds) && count($invoiceItemIds))) {
```

Write/transaction candidates:

```php
41: $this->update($data, $data['id']);
44: $invoiceItem = $invoice->invoiceItems()->save($invoiceItem);
53: InvoiceItem::whereNotIn('id', $invoiceItemIds)->whereInvoiceId($invoice->id)->delete();
```

## hms/app/Repositories/InvoiceRepository.php

Models: `Account` → `accounts`, `Accountant` → `accountants`, `Invoice` → `invoices`, `InvoiceItem` → `invoice_items`, `Notification` → `notifications`, `Patient` → `patients`, `Receptionist` → `receptionists`, `Setting` → `settings`, `User` → `users`

### getFieldsSearchable (line 32)

`public function getFieldsSearchable()`

### model (line 37)

`public function model()`

### getSyncList (line 42)

`public function getSyncList()`

### getAssociateAccountList (line 55)

`public function getAssociateAccountList($result)`

### saveInvoice (line 68)

`public function saveInvoice($input)`

Guard/validation candidates:

```php
73: if ($invoiceExist) {
76: return false;
85: $validator = Validator::make($data, InvoiceItem::$rules);
87: if ($validator->fails()) {
```

Write/transaction candidates:

```php
79: $invoice = $this->create(Arr::only($input, ['patient_id', 'invoice_date', 'discount', 'status', 'invoice_id', 'currency_symbol']));
95: $invoice->invoiceItems()->save($invoiceItem);
98: $invoice->save();
```

### prepareInputForInvoiceItem (line 103)

`public function prepareInputForInvoiceItem($input)`

Guard/validation candidates:

```php
109: if (! (isset($items[$index]['price']) && $key == 'price')) {
```

### updateInvoice (line 119)

`public function updateInvoice($invoiceId, $input)`

Guard/validation candidates:

```php
128: $validator = Validator::make($data, InvoiceItem::$rules, [
132: if ($validator->fails()) {
```

Write/transaction candidates:

```php
123: $invoice = $this->update(Arr::only($input, ['patient_id', 'invoice_date', 'discount', 'status', 'currency_symbol']), $invoiceId);
145: $invoice->save();
```

### getSyncListForCreate (line 150)

`public function getSyncListForCreate($invoiceId = null)`

### saveNotification (line 157)

`public function saveNotification($input)`

Guard/validation candidates:

```php
180: if ($notification == Notification::NOTIFICATION_FOR[Notification::PATIENT]) {
```

## hms/app/Repositories/IpdBillRepository.php

Models: `BedAssign` → `bed_assigns`, `Doctor` → `doctors`, `IpdBill` → `ipd_bills`, `IpdCharge` → `ipd_charges`, `IpdPatientDepartment` → `ipd_patient_departments`, `IpdPayment` → `ipd_payments`, `Notification` → `notifications`, `Patient` → `patients`, `Receptionist` → `receptionists`, `Setting` → `settings`, `User` → `users`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable()`

### model (line 33)

`public function model()`

### saveBill (line 38)

`public function saveBill($input)`

Guard/validation candidates:

```php
40: if (isset($input['bill_id'])) {
77: if($bedAssign){
85: if ($ipdPatientDepartment->bedAssign) {
```

Write/transaction candidates:

```php
41: $bill = $this->update($input, $input['bill_id']);
47: $bill = $this->create($input);
73: $ipdPatientDepartment->save();
78: $bedAssign->update(['status' => 0]);
82: $ipdPatientDepartment->bed->update(['is_available' => 1]);
83: $ipdPatientDepartment->update(['discharge' => 1]);
86: BedAssign::where('id', $ipdPatientDepartment->bedAssign->id)->delete();
```

### getBillList (line 92)

`public function getBillList($ipdPatientDepartment)`

Guard/validation candidates:

```php
104: if ($ipdPatientBill) {
```

### getSyncListForCreate (line 136)

`public function getSyncListForCreate()`

## hms/app/Repositories/IpdChargeRepository.php

Models: `Charge` → `charges`, `ChargeCategory` → `charge_categories`, `IpdCharge` → `ipd_charges`, `IpdPatientDepartment` → `ipd_patient_departments`, `Notification` → `notifications`, `Receptionist` → `receptionists`

### getFieldsSearchable (line 31)

`public function getFieldsSearchable()`

### model (line 36)

`public function model()`

### getChargeCategories (line 41)

`public function getChargeCategories($chargeTypeId)`

### getCharges (line 46)

`public function getCharges($chargeCategoryId)`

### getChargeStandardRate (line 51)

`public function getChargeStandardRate($chargeId, $isEdit, $onceOnEditRender, $ipdChargeId)`

Guard/validation candidates:

```php
54: if (! $isEdit) {
57: if ($onceOnEditRender != null) {
61: if ($charge != null) {
```

### createNotification (line 70)

`public function createNotification($input)`

Guard/validation candidates:

```php
86: if ($notification == Notification::NOTIFICATION_FOR[Notification::PATIENT]) {
```

## hms/app/Repositories/IpdConsultantRegisterRepository.php

Models: `IpdConsultantRegister` → `ipd_consultant_registers`

### getFieldsSearchable (line 24)

`public function getFieldsSearchable()`

### model (line 29)

`public function model()`

### store (line 34)

`public function store($input)`

Guard/validation candidates:

```php
38: if ($input['applied_date'][$i] == null || $input['instruction_date'][$i] == null) {
39: return false;
54: return true;
```

Write/transaction candidates:

```php
48: $this->model->create($ipdConsultantInstruction);
```

## hms/app/Repositories/IpdDiagnosisRepository.php

Models: `IpdDiagnosis` → `ipd_diagnoses`

### getFieldsSearchable (line 22)

`public function getFieldsSearchable()`

### model (line 27)

`public function model()`

### store (line 32)

`public function store($input)`

Guard/validation candidates:

```php
36: if (isset($input['file']) && ! empty($input['file'])) {
```

Write/transaction candidates:

```php
35: $ipdDiagnosis = $this->create($input);
```

### updateIpdDiagnosis (line 45)

`public function updateIpdDiagnosis($input, $ipdDiagnosisId)`

Guard/validation candidates:

```php
49: if (isset($input['file']) && ! empty($input['file'])) {
54: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
```

Write/transaction candidates:

```php
48: $ipdDiagnosis = $this->update($input, $ipdDiagnosisId);
```

### deleteIpdDiagnosis (line 62)

`public function deleteIpdDiagnosis($ipdDiagnosisId)`

Write/transaction candidates:

```php
67: $this->delete($ipdDiagnosisId);
```

## hms/app/Repositories/IpdPatientDepartmentRepository.php

Models: `Bed` → `beds`, `BedAssign` → `bed_assigns`, `BedType` → `bed_types`, `Category` → `categories`, `Doctor` → `doctors`, `IpdCharge` → `ipd_charges`, `IpdConsultantRegister` → `ipd_consultant_registers`, `IpdDiagnosis` → `ipd_diagnoses`, `IpdOperation` → `ipd_operation`, `IpdPatientDepartment` → `ipd_patient_departments`, `IpdPayment` → `ipd_payments`, `IpdPrescription` → `ipd_prescriptions`, `IpdTimeline` → `ipd_timelines`, `Notification` → `notifications`, `OperationCategory` → `operation_categories`, `Patient` → `patients`, `PatientCase` → `patient_cases`, `Prescription` → `prescriptions`, `Setting` → `settings`

### getFieldsSearchable (line 52)

`public function getFieldsSearchable()`

### model (line 57)

`public function model()`

### getAssociatedData (line 62)

`public function getAssociatedData()`

### getPatientCases (line 87)

`public function getPatientCases($patientId)`

### getPatientBeds (line 92)

`public function getPatientBeds($bedTypeId, $isEdit, $bedId, $ipdPatientBedTypeId)`

Guard/validation candidates:

```php
95: if (! $isEdit) {
99: if ($bedTypeId == $ipdPatientBedTypeId) {
```

### getDoctorsData (line 108)

`public function getDoctorsData()`

### getPatientsData (line 116)

`public function getPatientsData()`

### getDoctorsList (line 124)

`public function getDoctorsList()`

### getMedicinesCategoriesData (line 140)

`public function getMedicinesCategoriesData()`

### getMedicineCategoriesList (line 145)

`public function getMedicineCategoriesList()`

### store (line 160)

`public function store($input)`

Guard/validation candidates:

```php
168: if (strpos($key, 'field') === 0) {
189: return true;
```

Write/transaction candidates:

```php
175: $ipdPatientDepartment = IpdPatientDepartment::create($input);
187: $bedAssign->store($bedAssignData);
```

### updateIpdPatientDepartment (line 192)

`public function updateIpdPatientDepartment($input, $ipdPatientDepartment)`

Guard/validation candidates:

```php
200: if (strpos($key, 'field') === 0) {
206: if (isset($input['discharge']) && $input['discharge']  == 1) {
225: return true;
242: if (empty($bedAssignUpdate)) {
252: return true;
```

Write/transaction candidates:

```php
207: $ipdPatientDepartment->update([
227: $ipdPatientDepartment = $this->update($input, $ipdPatientDepartment->id);
243: $bedAssigns = BedAssign::create($bedAssignData);
244: BedAssign::where('id', $bedAssigns->id)->update(['ipd_patient_department_id' => $ipdPatientDepartment->id]);
246: $bedAssign->update($bedAssignData, $bedAssignUpdate);
```

### deleteIpdPatientDepartment (line 255)

`public function deleteIpdPatientDepartment($ipdPatientDepartment)`

Guard/validation candidates:

```php
259: if ($ipdPatientDepartment->bedAssign) {
265: return true;
```

Write/transaction candidates:

```php
257: $ipdPatientDepartment->bed->update(['is_available' => 1]);
260: BedAssign::where('id', $ipdPatientDepartment->bedAssign->id)->delete();
263: $ipdPatientDepartment->delete();
```

### getSyncListForCreate (line 268)

`public function getSyncListForCreate()`

### createNotification (line 275)

`public function createNotification($input)`

### getConsultantRegister (line 291)

`public function getConsultantRegister($id)`

### getConsultantDoctor (line 298)

`public function getConsultantDoctor($id)`

### getIPDOperation (line 305)

`public function getIPDOperation($id)`

### getIPDTimeline (line 312)

`public function getIPDTimeline($id)`

Guard/validation candidates:

```php
314: if (Auth::user()->hasRole('Admin')) {
```

### getIPDPrescription (line 321)

`public function getIPDPrescription($id)`

### getIPDCharges (line 326)

`public function getIPDCharges($id)`

### getIPDPayment (line 331)

`public function getIPDPayment($id)`

### getIPDDiagnosis (line 336)

`public function getIPDDiagnosis($id)`

### getOperationCategoryList (line 341)

`public function getOperationCategoryList()`

### getDoseDurationList (line 346)

`public function getDoseDurationList()`

### getDoseIntervalList (line 361)

`public function getDoseIntervalList()`

### getMealList (line 376)

`public function getMealList()`

## hms/app/Repositories/IpdPaymentRepository.php

Models: `IpdPatientDepartment` → `ipd_patient_departments`, `IpdPayment` → `ipd_payments`, `Transaction` → `transactions`

### getFieldsSearchable (line 33)

`public function getFieldsSearchable()`

### model (line 38)

`public function model()`

### store (line 43)

`public function store($input)`

Guard/validation candidates:

```php
51: if ($ipdBill) {
58: if (isset($input['file']) && ! empty($input['file'])) {
63: return true;
```

Write/transaction candidates:

```php
45: $ipdPayment = $this->create($input);
55: $ipdBill->save();
```

### updateIpdPayment (line 66)

`public function updateIpdPayment($input, $ipdPaymentId)`

Guard/validation candidates:

```php
72: if (isset($input['file']) && ! empty($input['file'])) {
78: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
```

Write/transaction candidates:

```php
69: DB::beginTransaction();
70: $ipdPayment = $this->update($input, $ipdPaymentId);
81: DB::commit();
83: DB::rollBack();
```

### deleteIpdPayment (line 88)

`public function deleteIpdPayment($ipdPaymentId)`

Write/transaction candidates:

```php
93: $this->delete($ipdPaymentId);
```

### stripeSession (line 99)

`public function stripeSession($input)`

Write/transaction candidates:

```php
115: $session = Session::create([
```

### ipdStripePaymentSuccess (line 143)

`public function ipdStripePaymentSuccess($input)`

Guard/validation candidates:

```php
147: if (empty($sessionId)) {
172: if ($ipdBill) {
180: return true;
```

Write/transaction candidates:

```php
157: DB::beginTransaction();
159: $ipdPayment = IpdPayment::create([
176: $ipdBill->save();
179: DB::commit();
182: DB::rollBack();
```

### razorpayPayment (line 187)

`public function razorpayPayment($input)`

Write/transaction candidates:

```php
211: $razorpayOrder = $api->order->create($orderData);
```

### ipdRazorpayPaymentSuccess (line 218)

`public function ipdRazorpayPaymentSuccess($input)`

Guard/validation candidates:

```php
224: if (count($input) && ! empty($input['razorpay_payment_id'])) {
231: if ($generatedSignature != $input['razorpay_signature']) {
232: if(getLoggedinPatient()){
255: if ($ipdBill) {
263: return true;
268: return false;
```

Write/transaction candidates:

```php
226: DB::beginTransaction();
242: $ipdPayment = IpdPayment::create([
259: $ipdBill->save();
262: DB::commit();
265: DB::rollBack();
```

### phonePePayment (line 272)

`public function phonePePayment($input)`

### phonePePaymentSuccess (line 344)

`public function phonePePaymentSuccess($input)`

Guard/validation candidates:

```php
361: if ($ipdBill) {
368: return true;
```

Write/transaction candidates:

```php
348: $ipdPayment = IpdPayment::create([
365: $ipdBill->save();
```

### ipdPaystackPaymentSuccess (line 371)

`public function ipdPaystackPaymentSuccess($response)`

Guard/validation candidates:

```php
404: if ($ipdBill) {
```

Write/transaction candidates:

```php
374: DB::beginTransaction();
389: $transaction = Transaction::create($transactionData);
391: $ipdPayment = IpdPayment::create([
408: $ipdBill->save();
411: DB::commit();
414: DB::rollBack();
```

### flutterWavePayment (line 419)

`public function flutterWavePayment($input)`

Guard/validation candidates:

```php
453: if ($payment['status'] !== 'success') {
454: if(getLoggedinPatient()){
```

### ipdFlutterwavePaymentSuccess (line 466)

`public function ipdFlutterwavePaymentSuccess($input)`

Guard/validation candidates:

```php
471: if ($input['status'] ==  'successful')
491: if ($ipdBill) {
499: return true;
507: return false;
```

Write/transaction candidates:

```php
469: DB::beginTransaction();
478: $ipdPayment = IpdPayment::create([
495: $ipdBill->save();
498: DB::commit();
504: DB::rollBack();
```

## hms/app/Repositories/IpdPrescriptionRepository.php

Models: `IpdPatientDepartment` → `ipd_patient_departments`, `IpdPrescription` → `ipd_prescriptions`, `IpdPrescriptionItem` → `ipd_prescription_items`, `Medicine` → `medicines`, `MedicineBill` → `medicine_bills`, `Notification` → `notifications`, `SaleMedicine` → `sale_medicines`

### getFieldsSearchable (line 29)

`public function getFieldsSearchable()`

### model (line 34)

`public function model()`

### getMedicines (line 39)

`public function getMedicines($medicineCategoryId)`

### store (line 44)

`public function store($input)`

Guard/validation candidates:

```php
96: return true;
```

Write/transaction candidates:

```php
48: $ipdPrescription = $this->model->create($ipdPrescriptionArr);
52: $medicineBill = MedicineBill::create([
73: IpdPrescriptionItem::create($ipdPrescriptionItem);
85: SaleMedicine::create($saleMedicineArray);
87: $medicineBill->update([
```

### getIpdPrescriptionData (line 99)

`public function getIpdPrescriptionData($ipdPrescription)`

### updateIpdPrescriptionItems (line 110)

`public function updateIpdPrescriptionItems($input, $ipdPrescription)`

Guard/validation candidates:

```php
155: return true;
```

Write/transaction candidates:

```php
114: $medicineBill->saleMedicine()->delete();
116: $ipdPrescription->update($ipdPrescriptionArr);
117: $ipdPrescription->ipdPrescriptionItems()->delete();
133: IpdPrescriptionItem::create($ipdPrescriptionItem);
145: SaleMedicine::create($saleMedicineArray);
147: $medicineBill->update([
```

### createNotification (line 158)

`public function createNotification($input)`

Guard/validation candidates:

```php
171: if ($notification == Notification::NOTIFICATION_FOR[Notification::PATIENT]) {
```

## hms/app/Repositories/IpdTimelineRepository.php

Models: `IpdTimeline` → `ipd_timelines`

### getFieldsSearchable (line 24)

`public function getFieldsSearchable()`

### model (line 29)

`public function model()`

### getTimeLines (line 34)

`public function getTimeLines($ipdPatientDepartmentId)`

Guard/validation candidates:

```php
36: if (Auth::user()->hasRole('Admin')|| Auth::user()->hasRole('Doctor') || Auth::user()->hasRole('Receptionist')) {
```

### store (line 43)

`public function store($input)`

Guard/validation candidates:

```php
48: if (isset($input['attachment']) && ! empty($input['attachment'])) {
```

Write/transaction candidates:

```php
47: $ipdTimeline = $this->create($input);
```

### updateIpdTimeline (line 57)

`public function updateIpdTimeline($input, $ipdTimelineId)`

Guard/validation candidates:

```php
62: if (isset($input['attachment']) && ! empty($input['attachment'])) {
63: if ($ipdTimeline->media->first() != null) {
69: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
```

Write/transaction candidates:

```php
61: $ipdTimeline = $this->update($input, $ipdTimelineId);
```

### deleteIpdTimeline (line 77)

`public function deleteIpdTimeline($ipdTimelineId)`

Guard/validation candidates:

```php
81: if ($ipdTimeline->media->first() != null) {
```

Write/transaction candidates:

```php
84: $this->delete($ipdTimelineId);
```

## hms/app/Repositories/IssuedItemRepository.php

Models: `Department` → `departments`, `IssuedItem` → `issued_items`, `ItemCategory` → `item_categories`

### getFieldsSearchable (line 31)

`public function getFieldsSearchable()`

### model (line 36)

`public function model()`

### getAssociatedData (line 41)

`public function getAssociatedData()`

### store (line 51)

`public function store($input)`

Guard/validation candidates:

```php
62: return true;
```

Write/transaction candidates:

```php
54: DB::beginTransaction();
56: $issuedItem = IssuedItem::create($input);
58: $issuedItem->item()->update(['available_quantity' => $newItemAvailableQty]);
60: DB::commit();
64: DB::rollBack();
```

### destroyIssuedItemStock (line 70)

`public function destroyIssuedItemStock($issuedItem)`

Guard/validation candidates:

```php
72: if ($issuedItem->status == 0) {
```

Write/transaction candidates:

```php
74: $issuedItem->item()->update(['available_quantity' => $newItemAvailableQty]);
76: $this->delete($issuedItem->id);
```

### returnIssuedItem (line 79)

`public function returnIssuedItem($itemId)`

Guard/validation candidates:

```php
86: return true;
```

Write/transaction candidates:

```php
83: $issuedItem->item()->update(['available_quantity' => $newItemAvailableQty]);
84: $issuedItem->update(['return_date' => date('Y-m-d'), 'quantity' => 0, 'status' => IssuedItem::ITEM_RETURNED]);
```

## hms/app/Repositories/ItemCategoryRepository.php

Models: `ItemCategory` → `item_categories`

### getFieldsSearchable (line 18)

`public function getFieldsSearchable()`

### model (line 23)

`public function model()`

## hms/app/Repositories/ItemRepository.php

Models: `Item` → `items`, `ItemCategory` → `item_categories`

### getFieldsSearchable (line 23)

`public function getFieldsSearchable()`

### model (line 28)

`public function model()`

### getItemCategories (line 33)

`public function getItemCategories()`

## hms/app/Repositories/ItemStockRepository.php

Models: `ItemCategory` → `item_categories`, `ItemStock` → `item_stocks`

### getFieldsSearchable (line 31)

`public function getFieldsSearchable()`

### model (line 36)

`public function model()`

### getItemCategories (line 41)

`public function getItemCategories()`

### store (line 46)

`public function store($input)`

Guard/validation candidates:

```php
56: if (isset($input['attachment']) && ! empty($input['attachment'])) {
63: return true;
```

Write/transaction candidates:

```php
49: DB::beginTransaction();
53: $itemStock = $this->create($itemStockInputArray);
54: $itemStock->item()->update(['available_quantity' => $itemStockInputArray['quantity']]);
61: DB::commit();
65: DB::rollBack();
```

### update (line 71)

`public function update($itemStock, $input)`

Guard/validation candidates:

```php
78: if (is_numeric($itemStockInputArray['quantity'])) {
80: if ($itemStockInputArray['quantity'] !== $itemStock->quantity) {
81: if ($itemStockInputArray['quantity'] < $itemStock->quantity) {
92: if (isset($input['attachment']) && ! empty($input['attachment'])) {
93: if ($itemStock->media->first() != null) {
100: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
106: return true;
```

Write/transaction candidates:

```php
74: DB::beginTransaction();
87: $itemStock->item()->update(['available_quantity' => $newItemAvailableQty]);
89: $itemStock->update($itemStockInputArray);
104: DB::commit();
108: DB::rollBack();
```

### destroyItemStock (line 114)

`public function destroyItemStock($itemStock)`

Guard/validation candidates:

```php
122: if ($attachment->media->first() !== null) {
```

Write/transaction candidates:

```php
118: $itemStock->item()->update(['available_quantity' => $newItemAvailableQty]);
126: $this->delete($itemStock->id);
```

### downloadMedia (line 132)

`public function downloadMedia($itemStock)`

Guard/validation candidates:

```php
138: if (config('app.media_disc') === 'public') {
```

## hms/app/Repositories/LabTechnicianRepository.php

Models: `Address` → `addresses`, `Department` → `departments`, `LabTechnician` → `lab_technicians`, `User` → `users`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable()`

### model (line 33)

`public function model()`

### store (line 38)

`public function store($input, $mail = true)`

Guard/validation candidates:

```php
49: if ($mail) {
53: if (isset($input['image']) && ! empty($input['image'])) {
62: if (! empty($address = Address::prepareAddressArray($input))) {
69: return true;
```

Write/transaction candidates:

```php
47: $user = User::create($input);
57: $labTechnician = LabTechnician::create(['user_id' => $user->id]);
63: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
66: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
```

### update (line 75)

`public function update($labTechnician, $input)`

Guard/validation candidates:

```php
83: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
87: if (isset($input['image']) && ! empty($input['image'])) {
96: if (! empty($labTechnician->address)) {
97: if (empty($address = Address::prepareAddressArray($input))) {
102: if (! empty($address = Address::prepareAddressArray($input)) && empty($labTechnician->address)) {
109: return true;
```

Write/transaction candidates:

```php
93: $labTechnician->user->update($input);
94: $labTechnician->update($input);
98: $labTechnician->address->delete();
100: $labTechnician->address->update($input);
105: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
```

## hms/app/Repositories/LiveConsultationRepository.php

Models: `Doctor` → `doctors`, `IpdPatientDepartment` → `ipd_patient_departments`, `LiveConsultation` → `live_consultations`, `Notification` → `notifications`, `OpdPatientDepartment` → `opd_patient_departments`, `Patient` → `patients`, `User` → `users`, `UserGoogleEventSchedule` → `user_google_event_schedules`, `UserZoomCredential` → `user_zoom_credential`

### getFieldsSearchable (line 34)

`public function getFieldsSearchable()`

### model (line 39)

`public function model()`

### getTypeNumber (line 44)

`public function getTypeNumber($input)`

Guard/validation candidates:

```php
46: if ($input['consultation_type'] == LiveConsultation::OPD) {
```

### store (line 53)

`public function store($input)`

Guard/validation candidates:

```php
68: return true;
```

Write/transaction candidates:

```php
66: $zoomModel = LiveConsultation::create($input);
```

### edit (line 74)

`public function edit($input, $liveConsultation)`

Guard/validation candidates:

```php
80: if($input['platform_type'] == LiveConsultation::GOOGLE_MEET){
84: if ($newConsultationDateTime->ne($oldConsultationDateTime) || $input['consultation_duration_minutes'] != $liveConsultation->consultation_duration_minutes) {
111: return true;
117: return true;
```

Write/transaction candidates:

```php
78: DB::beginTransaction();
87: $oldGoogleMeetingData->delete();
88: $liveConsultation->delete();
96: $liveConsultation->update($input);
106: $zoomModel = $liveConsultation->update($input);
109: DB::commit();
113: DB::rollBack();
```

### createUserZoom (line 120)

`public function createUserZoom($input)`

Guard/validation candidates:

```php
131: return true;
```

### createNotification (line 137)

`public function createNotification($input = [])`

Guard/validation candidates:

```php
152: if ($notification == Notification::NOTIFICATION_FOR[Notification::PATIENT]) {
```

### filter (line 169)

`public function filter($status)`

Guard/validation candidates:

```php
171: if ($status == 'all') {
187: return false;
```

### googleMeetStore (line 191)

`public function googleMeetStore($input)`

Guard/validation candidates:

```php
202: return true;
```

Write/transaction candidates:

```php
199: $liveConsultation = LiveConsultation::create($input);
```

## hms/app/Repositories/LiveMeetingRepository.php

Models: `LiveMeeting` → `live_meetings`, `Notification` → `notifications`, `User` → `users`

### getFieldsSearchable (line 31)

`public function getFieldsSearchable()`

### model (line 36)

`public function model()`

### store (line 41)

`public function store($input)`

Guard/validation candidates:

```php
57: return true;
```

Write/transaction candidates:

```php
54: $zoomModel = LiveMeeting::create($input);
55: $zoomModel->members()->attach($input['staff_list']);
```

### edit (line 63)

`public function edit($input, $liveMeeting)`

Guard/validation candidates:

```php
75: return true;
```

Write/transaction candidates:

```php
72: $liveMeeting->update($input);
73: $liveMeeting->members()->sync($input['staff_list']);
```

### getUsers (line 81)

`public function getUsers()`

### createNotification (line 101)

`public function createNotification($input = [])`

## hms/app/Repositories/LunchBreakRepository.php

Models: `LunchBreak` → `lunch_breaks`

### getFieldsSearchable (line 23)

`public function getFieldsSearchable(): array`

### model (line 28)

`public function model()`

### store (line 33)

`public function store($input)`

Guard/validation candidates:

```php
35: if(isset($input['date'])){
46: if($from && $to){
52: if (! $doctor_break) {
61: return true;
63: return false;
```

Write/transaction candidates:

```php
53: LunchBreak::create([
```

## hms/app/Repositories/MailRepository.php

Models: `Mail` → `mails`

### getFieldsSearchable (line 25)

`public function getFieldsSearchable()`

### model (line 30)

`public function model()`

### store (line 35)

`public function store($input)`

Guard/validation candidates:

```php
39: if (isset($input['attachments']) && ! empty($input['attachments'])) {
62: return true;
```

Write/transaction candidates:

```php
46: $mail = Mail::create([
```

## hms/app/Repositories/ManualBillPaymentRepository.php

Models: `ManualBillPayment` → `bill_transactions`, `Bill` → `bills`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable(): array`

### model (line 33)

`public function model()`

### create (line 38)

`public function create($input)`

Guard/validation candidates:

```php
42: if(!empty($bill)){
53: return true;
```

Write/transaction candidates:

```php
43: ManualBillPayment::create([
50: $bill->update(['payment_mode' => $input['payment_type']]);
```

### updateTransaction (line 56)

`public function updateTransaction($input, $id)`

Guard/validation candidates:

```php
60: if ($input['payment_status'] == ManualBillPayment::Approved) {
```

Write/transaction candidates:

```php
61: $billTransaction->update(['status' => ManualBillPayment::Approved]);
62: $billTransaction->bill()->update(['status' => ManualBillPayment::Approved]);
64: $billTransaction->update(['status' => ManualBillPayment::Rejected]);
65: $billTransaction->bill()->update(['status' => ManualBillPayment::Rejected,'payment_mode' => null]);
```

### createStripeSession (line 70)

`public function createStripeSession($patientBill)`

Write/transaction candidates:

```php
74: $session = Session::create([
```

### stripePaymentSuccess (line 102)

`public function stripePaymentSuccess($sessionId)`

Guard/validation candidates:

```php
104: if (empty($sessionId)) {
115: if(!empty($bill)){
```

Write/transaction candidates:

```php
113: DB::beginTransaction();
116: ManualBillPayment::create([
125: $bill->update(['payment_mode' => Bill::Stripe, 'status' => '1']);
128: DB::commit();
130: DB::rollBack();
```

### razorpayPayment (line 137)

`public function razorpayPayment($billId)`

Write/transaction candidates:

```php
153: $razorpayOrder = $api->order->create($orderData);
```

### razorpayPaymentSuccess (line 161)

`public function razorpayPaymentSuccess($input)`

Guard/validation candidates:

```php
167: if (count($input) && ! empty($input['razorpay_payment_id'])) {
174: if ($generatedSignature != $input['razorpay_signature']) {
181: if(!empty($bill)){
200: return false;
```

Write/transaction candidates:

```php
169: DB::beginTransaction();
182: ManualBillPayment::create([
191: $bill->update(['payment_mode' => Bill::Razorpay, 'status' => '1']);
194: DB::commit();
196: DB::rollBack();
```

### flutterWavePayment (line 204)

`public function flutterWavePayment($input)`

Guard/validation candidates:

```php
234: if ($payment['status'] !== 'success') {
```

### phonePePayment (line 240)

`public function phonePePayment($input)`

### flutterwavePaymentSuccess (line 312)

`public function flutterwavePaymentSuccess($input)`

Guard/validation candidates:

```php
317: if ($input['status'] ==  'successful')
325: if(!empty($bill)){
346: return false;
```

Write/transaction candidates:

```php
315: DB::beginTransaction();
326: ManualBillPayment::create([
335: $bill->update(['payment_mode' => Bill::Flutterwave, 'status' => '1']);
338: DB::commit();
343: DB::rollBack();
```

### billPhonePePaymentSuccess (line 348)

`public function billPhonePePaymentSuccess($input)`

Guard/validation candidates:

```php
355: if(!empty($bill)){
375: return true;
380: return false;
```

Write/transaction candidates:

```php
351: DB::beginTransaction();
356: $payment = ManualBillPayment::create([
365: $bill->update(['payment_mode' => Bill::PhonePe, 'status' => '1']);
373: DB::commit();
377: DB::rollBack();
```

### PaystackPaymentSucess (line 383)

`public function PaystackPaymentSucess($response)`

Guard/validation candidates:

```php
391: if(!empty($bill)){
```

Write/transaction candidates:

```php
389: DB::beginTransaction();
392: $payment = ManualBillPayment::create([
402: $bill->update(['payment_mode' => Bill::Paystack, 'status' => '1']);
405: DB::commit();
407: DB::rollBack();
```

## hms/app/Repositories/MedicineBillRepository.php

Models: `Medicine` → `medicines`, `MedicineBill` → `medicine_bills`, `SaleMedicine` → `sale_medicines`

### getFieldsSearchable (line 41)

`public function getFieldsSearchable()`

### model (line 46)

`public function model()`

### store (line 51)

`public function store($input)`

Guard/validation candidates:

```php
54: if (isset($input['medicine'])) {
73: if ($input['category_id']) {
85: if ($input['payment_status'] == 1) {
94: return true;
```

Write/transaction candidates:

```php
55: $medicineBill = MedicineBill::create([
69: $medicineBill->update([
76: SaleMedicine::create([
86: $medicine->update([
```

### update (line 100)

`public function update($medicineBill, $input)`

Guard/validation candidates:

```php
107: if (empty($input['medicine'][$key]) && $input['payment_status'] == false) {
112: if (isset($saleMedincine->sale_quantity) && $input['quantity'][$key]) {
113: if ($saleMedincine->sale_quantity < $input['quantity'][$key] && $input['payment_status'] == 1) {
128: if ($input['payment_status'] && $medicineBill->payment_status == true) {
130: if (array_key_exists($key, $input['medicine'])) {
132: if ($updatedMedicine->available_quantity < $input['quantity'][$key]) {
169: if (! empty($duplicateIds)) {
179: if ($input['payment_status'] == true && $medicine->available_quantity < $qty && $medicineBill->payment_status == 0) {
184: if (! is_null($saleMedicine) && $input['payment_status'] == 1 && $medicineBill['payment_status'] == 1) {
186: if ($PreviousQty > $qty) {
193: if (! array_key_exists($input['medicine'][$key], $result) && $medicine->available_quantity < $qty && $input['payment_status'] == false) {
214: if ($input['category_id']) {
227: if ($input['payment_status'] == 1 && $beforeStatus == 0
242: return true;
```

Write/transaction candidates:

```php
103: DB::beginTransaction();
143: $deleteMedicine->update(['available_quantity' => $deleteMedicine->available_quantity + $saleMedicine->sale_quantity]);
147: $updatedMedicine->update([
187: $medicine->update([
199: $medicineBill->saleMedicine()->delete();
225: $saleMedicine->save();
229: $medicine->update([
235: DB::commit();
238: DB::rollBack();
```

### medicineBillStore (line 245)

`public function medicineBillStore($input)`

Guard/validation candidates:

```php
251: if (empty($input['medicine'])) {
262: if (! empty($duplicateIds)) {
270: if ($medicine->available_quantity < $qty) {
295: if ($input['category_id']) {
309: if ($input['payment_status'] == 1) {
```

Write/transaction candidates:

```php
249: DB::beginTransaction();
276: $medicineBill = MedicineBill::create([
291: $medicineBill->update([
307: $saleMedicine->save();
310: $medicine->update([
315: DB::commit();
320: DB::rollBack();
```

### stripeSession (line 325)

`public function stripeSession($input, $medicineBill)`

Write/transaction candidates:

```php
331: $session = Session::create([
```

### medicineBillstripeSuccess (line 358)

`public function medicineBillstripeSuccess($input)`

Guard/validation candidates:

```php
362: if (empty($sessionId)) {
370: if($sessionData){
371: return true;
374: return false;
```

### medicineBillstripeFailed (line 377)

`public function medicineBillstripeFailed($input)`

Guard/validation candidates:

```php
386: if ($input['category_id']) {
394: if ($input['payment_status'] == 1) {
404: return true;
```

Write/transaction candidates:

```php
392: $saleMedicine->delete();
395: $medicine->update([
402: $medicineBill->delete();
```

### razorPayPayment (line 410)

`public function razorPayPayment($input)`

Write/transaction candidates:

```php
425: $razorpayOrder = $api->order->create($orderData);
```

### razorPayPaymentSuccess (line 432)

`public function razorPayPaymentSuccess($input)`

Guard/validation candidates:

```php
436: if (count($input) && ! empty($input['razorpay_payment_id'])) {
441: if ($generatedSignature != $input['razorpay_signature']) {
448: return true;
```

### phonePePayment (line 452)

`public function phonePePayment($input)`

### phonePePaymentSuccess (line 524)

`public function phonePePaymentSuccess($input)`

Guard/validation candidates:

```php
531: if (empty($input['medicine'])) {
542: if (! empty($duplicateIds)) {
550: if ($medicine->available_quantity < $qty) {
575: if ($input['category_id']) {
589: if ($input['payment_status'] == 1) {
597: return true;
603: return false;
```

Write/transaction candidates:

```php
529: DB::beginTransaction();
556: $medicineBill = MedicineBill::create([
571: $medicineBill->update([
580: SaleMedicine::create([
590: $medicine->update([
595: DB::commit();
600: DB::rollBack();
```

### paystackPaymentSuccess (line 606)

`public function paystackPaymentSuccess($response)`

Guard/validation candidates:

```php
613: if (empty($input['medicine'])) {
624: if (! empty($duplicateIds)) {
632: if ($medicine->available_quantity < $qty) {
657: if ($input['category_id']) {
671: if ($input['payment_status'] == 1) {
679: return true;
```

Write/transaction candidates:

```php
611: DB::beginTransaction();
638: $medicineBill = MedicineBill::create([
653: $medicineBill->update([
662: SaleMedicine::create([
672: $medicine->update([
677: DB::commit();
682: DB::rollBack();
```

### flutterWavePayment (line 687)

`public function flutterWavePayment($input)`

Guard/validation candidates:

```php
712: if ($payment['status'] !== 'success') {
```

### flutterwavePaymentSuccess (line 721)

`public function flutterwavePaymentSuccess($input)`

Guard/validation candidates:

```php
726: if ($input['status'] ==  'successful')
733: if(isset($sessionData) && !empty($sessionData)){
735: if (empty($sessionData['medicine'])) {
746: if (! empty($duplicateIds)) {
754: if ($medicine->available_quantity < $qty) {
779: if ($sessionData['category_id']) {
793: if ($sessionData['payment_status'] == 1) {
802: return true;
810: return false;
```

Write/transaction candidates:

```php
724: DB::beginTransaction();
760: $medicineBill = MedicineBill::create([
775: $medicineBill->update([
784: SaleMedicine::create([
794: $medicine->update([
799: DB::commit();
807: DB::rollBack();
```

## hms/app/Repositories/MedicineRepository.php

Models: `Brand` → `brands`, `Category` → `categories`, `Medicine` → `medicines`, `Prescription` → `prescriptions`

### getFieldsSearchable (line 29)

`public function getFieldsSearchable()`

### model (line 34)

`public function model()`

### getSyncList (line 39)

`public function getSyncList()`

### getMedicineList (line 47)

`public function getMedicineList()`

### getMealList (line 62)

`public function getMealList()`

### getDoseDurationList (line 77)

`public function getDoseDurationList()`

### getDoseIntervalList (line 92)

`public function getDoseIntervalList()`

## hms/app/Repositories/NoticeBoardRepository.php

Models: `NoticeBoard` → `notice_boards`, `Notification` → `notifications`, `User` → `users`

### getFieldsSearchable (line 25)

`public function getFieldsSearchable()`

### model (line 30)

`public function model()`

### createNotification (line 35)

`public function createNotification()`

Guard/validation candidates:

```php
55: return true;
```

## hms/app/Repositories/NurseRepository.php

Models: `Address` → `addresses`, `Department` → `departments`, `Nurse` → `nurses`, `User` → `users`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable()`

### model (line 33)

`public function model()`

### store (line 38)

`public function store($input, $mail = true)`

Guard/validation candidates:

```php
48: if ($mail) {
52: if (isset($input['image']) && ! empty($input['image'])) {
59: if (! empty($address = Address::prepareAddressArray($input))) {
66: return true;
```

Write/transaction candidates:

```php
47: $user = User::create($input);
55: $nurse = Nurse::create(['user_id' => $user->id]);
60: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
63: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
```

### update (line 72)

`public function update($nurse, $input)`

Guard/validation candidates:

```php
78: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
81: if (isset($input['image']) && ! empty($input['image'])) {
90: if (! empty($nurse->address)) {
91: if (empty($address = Address::prepareAddressArray($input))) {
96: if (! empty($address = Address::prepareAddressArray($input)) && empty($nurse->address)) {
103: return true;
```

Write/transaction candidates:

```php
87: $nurse->user->update($input);
88: $nurse->update($input);
92: $nurse->address->delete();
94: $nurse->address->update($input);
99: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
```

## hms/app/Repositories/OdontogramRepository.php

Models: `Doctor` → `doctors`, `Odontogram` → `odontograms`, `Patient` → `patients`

### getFieldsSearchable (line 23)

`public function getFieldsSearchable()`

### model (line 28)

`public function model()`

### store (line 33)

`public function store($input)`

Write/transaction candidates:

```php
38: $this->create($input);
```

### updateData (line 44)

`public function updateData($input, $Id)`

Write/transaction candidates:

```php
54: $this->update($input, $Id);
```

### getPatients (line 61)

`public function getPatients()`

### getDoctorData (line 69)

`public function getDoctorData()`

## hms/app/Repositories/OpdDiagnosisRepository.php

Models: `Notification` → `notifications`, `OpdDiagnosis` → `opd_diagnoses`, `OpdPatientDepartment` → `opd_patient_departments`

### getFieldsSearchable (line 25)

`public function getFieldsSearchable()`

### model (line 30)

`public function model()`

### store (line 35)

`public function store($input)`

Guard/validation candidates:

```php
39: if (isset($input['file']) && ! empty($input['file'])) {
```

Write/transaction candidates:

```php
38: $opdDiagnosis = $this->create($input);
```

### updateOpdDiagnosis (line 48)

`public function updateOpdDiagnosis($input,$opdDiagnosisId)`

Guard/validation candidates:

```php
52: if (isset($input['file']) && ! empty($input['file'])) {
57: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
```

Write/transaction candidates:

```php
51: $opdDiagnosis = $this->update($input, $opdDiagnosisId);
```

### deleteOpdDiagnosis (line 65)

`public function deleteOpdDiagnosis($opdDiagnosisId)`

Write/transaction candidates:

```php
70: $this->delete($opdDiagnosisId);
```

### createNotification (line 76)

`public function createNotification($input)`

## hms/app/Repositories/OpdPatientDepartmentRepository.php

Models: `Category` → `categories`, `Doctor` → `doctors`, `Notification` → `notifications`, `OpdPatientDepartment` → `opd_patient_departments`, `OpdPrescription` → `opd_prescriptions`, `Patient` → `patients`, `PatientCase` → `patient_cases`, `Prescription` → `prescriptions`

### getFieldsSearchable (line 40)

`public function getFieldsSearchable()`

### model (line 45)

`public function model()`

### getAssociatedData (line 50)

`public function getAssociatedData()`

### getPatientCases (line 66)

`public function getPatientCases($patientId)`

### getDoctorsData (line 71)

`public function getDoctorsData()`

### getDoctorsList (line 76)

`public function getDoctorsList()`

### store (line 92)

`public function store($input)`

Guard/validation candidates:

```php
99: if (strpos($key, 'field') === 0) {
110: return true;
```

Write/transaction candidates:

```php
105: OpdPatientDepartment::create($input);
```

### updateOpdPatientDepartment (line 113)

`public function updateOpdPatientDepartment($input, $opdPatientDepartment)`

Guard/validation candidates:

```php
121: if (strpos($key, 'field') === 0) {
132: return true;
```

Write/transaction candidates:

```php
127: $opdPatientDepartment->update($input);
```

### createNotification (line 135)

`public function createNotification($input)`

Guard/validation candidates:

```php
141: if (isset($input['revisit'])) {
```

### getMedicinesCategoriesData (line 158)

`public function getMedicinesCategoriesData(): Collection`

### getMedicineCategoriesList (line 163)

`public function getMedicineCategoriesList()`

### getDoseDurationList (line 178)

`public function getDoseDurationList()`

### getDoseIntervalList (line 193)

`public function getDoseIntervalList()`

### getMealList (line 208)

`public function getMealList()`

## hms/app/Repositories/OpdPresciptionRepository.php

Models: `Medicine` → `medicines`, `MedicineBill` → `medicine_bills`, `Notification` → `notifications`, `OpdPatientDepartment` → `opd_patient_departments`, `OpdPrescription` → `opd_prescriptions`, `OpdPrescriptionItem` → `opd_prescription_items`, `SaleMedicine` → `sale_medicines`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable()`

### model (line 33)

`public function model()`

### getMedicines (line 38)

`public function getMedicines($medicineCategoryId)`

### store (line 43)

`public function store($input)`

Guard/validation candidates:

```php
99: return true;
```

Write/transaction candidates:

```php
50: $opdPrescription = OpdPrescription::create([
56: $medicineBill = MedicineBill::create([
77: OpdPrescriptionItem::create($opdPrescriptionItem);
89: SaleMedicine::create($saleMedicineArray);
91: $medicineBill->update([
```

### createNotification (line 102)

`public function createNotification($input)`

Guard/validation candidates:

```php
115: if ($notification == Notification::NOTIFICATION_FOR[Notification::PATIENT]) {
```

### getOpdPrescriptionData (line 132)

`public function getOpdPrescriptionData($opdPrescription)`

### updateopdPrescriptionItems (line 143)

`public function updateopdPrescriptionItems($input,$opdPrescription)`

Guard/validation candidates:

```php
190: return true;
```

Write/transaction candidates:

```php
147: $medicineBill->saleMedicine()->delete();
150: $opdPrescription->update($opdPrescriptionArr);
151: $opdPrescription->OpdPrescriptionItems()->delete();
168: OpdPrescriptionItem::create($opdPrescriptionItem);
180: SaleMedicine::create($saleMedicineArray);
182: $medicineBill->update([
```

## hms/app/Repositories/OpdTimelineRepository.php

Models: `OpdTimeline` → `opd_timelines`

### getFieldsSearchable (line 24)

`public function getFieldsSearchable()`

### model (line 29)

`public function model()`

### getTimeLines (line 34)

`public function getTimeLines($opdPatientDepartmentId)`

Guard/validation candidates:

```php
36: if (Auth::user()->hasRole('Admin')|| Auth::user()->hasRole('Doctor') || Auth::user()->hasRole('Receptionist')) {
```

### store (line 43)

`public function store($input)`

Guard/validation candidates:

```php
48: if (isset($input['attachment']) && ! empty($input['attachment'])) {
```

Write/transaction candidates:

```php
47: $opdTimeline = $this->create($input);
```

### updateOpdTimeline (line 57)

`public function updateOpdTimeline($input, $opdTimelineId)`

Guard/validation candidates:

```php
63: if (isset($input['attachment']) && ! empty($input['attachment'])) {
64: if ($opdTimeline->media->first() != null) {
70: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
```

Write/transaction candidates:

```php
61: $opdTimeline = $this->update($input, $opdTimelineId);
```

### deleteOpdTimeline (line 78)

`public function deleteOpdTimeline($opdTimelineId)`

Write/transaction candidates:

```php
83: $this->delete($opdTimelineId);
```

## hms/app/Repositories/OperationCategoryRepository.php

Models: `OperationCategory` → `operation_categories`

### getFieldsSearchable (line 13)

`public function getFieldsSearchable()`

### model (line 18)

`public function model()`

## hms/app/Repositories/OperationReportRepository.php

Models: `Doctor` → `doctors`, `OperationReport` → `operation_reports`, `PatientCase` → `patient_cases`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable()`

### model (line 33)

`public function model()`

### getDoctors (line 38)

`public function getDoctors()`

### getCases (line 45)

`public function getCases()`

Guard/validation candidates:

```php
48: if ($user->hasRole('Doctor')) {
```

### store (line 63)

`public function store($input)`

Guard/validation candidates:

```php
72: return true;
```

Write/transaction candidates:

```php
70: $operationReport = OperationReport::create($input);
```

### update (line 78)

`public function update($input, $operationReport)`

Guard/validation candidates:

```php
87: return true;
```

Write/transaction candidates:

```php
85: $operationReport->update($input);
```

## hms/app/Repositories/PackageRepository.php

Models: `Package` → `packages`, `PackageService` → `package_services`, `Service` → `services`

### getFieldsSearchable (line 26)

`public function getFieldsSearchable()`

### model (line 31)

`public function model()`

### getServicesList (line 36)

`public function getServicesList()`

### getServices (line 43)

`public function getServices()`

### store (line 57)

`public function store($input)`

Guard/validation candidates:

```php
67: $validator = Validator::make($data, PackageService::$rules);
69: if ($validator->fails()) {
```

Write/transaction candidates:

```php
61: $package = $this->create(Arr::except($input, ['service_id', 'quantity', 'rate']));
77: $package->packageServicesItems()->save($packageServiceItem);
80: $package->save();
```

### prepareInputForServicePackageItem (line 85)

`public function prepareInputForServicePackageItem($input)`

Guard/validation candidates:

```php
91: if (! (isset($items[$index]['rate']) && $key == 'rate')) {
```

### updatePackage (line 101)

`public function updatePackage($packageId, $input)`

Guard/validation candidates:

```php
111: $validator = Validator::make($data, PackageService::$rules, [
115: if ($validator->fails()) {
```

Write/transaction candidates:

```php
105: $package = $this->update($input, $packageId);
128: $package->save();
```

## hms/app/Repositories/PackageServiceItemsRepository.php

Models: `Package` → `packages`, `PackageService` → `package_services`

### getFieldsSearchable (line 23)

`public function getFieldsSearchable()`

### model (line 28)

`public function model()`

### updatePackageServiceItem (line 33)

`public function updatePackageServiceItem($packageServiceItemInput, $packageId)`

Guard/validation candidates:

```php
39: if (isset($data['id']) && ! empty($data['id'])) {
49: if (! (isset($packageServiceItemIds) && count($packageServiceItemIds))) {
```

Write/transaction candidates:

```php
41: $this->update($data, $data['id']);
44: $packageServiceItem = $package->packageServicesItems()->save($packageServiceItem);
52: PackageService::whereNotIn('id', $packageServiceItemIds)->wherePackageId($package->id)->delete();
```

## hms/app/Repositories/PathologyCategoryRepository.php

Models: `PathologyCategory` → `pathology_categories`

### getFieldsSearchable (line 17)

`public function getFieldsSearchable()`

### model (line 22)

`public function model()`

## hms/app/Repositories/PathologyParameterRepository.php

Models: `PathologyParameter` → `pathology_parameters`, `PathologyUnit` → `pathology_units`

### getFieldsSearchable (line 22)

`public function getFieldsSearchable(): array`

### model (line 27)

`public function model()`

### getPathologyUnitData (line 32)

`public function getPathologyUnitData()`

## hms/app/Repositories/PathologyTestRepository.php

Models: `ChargeCategory` → `charge_categories`, `PathologyCategory` → `pathology_categories`, `PathologyTest` → `pathology_tests`, `PathologyParameter` → `pathology_parameters`, `PathologyParameterItem` → `pathology_parameter_items`, `Setting` → `settings`

### getFieldsSearchable (line 35)

`public function getFieldsSearchable()`

### model (line 40)

`public function model()`

### store (line 45)

`public function store($input)`

Guard/validation candidates:

```php
63: if ($input['parameter_id']) {
```

Write/transaction candidates:

```php
47: DB::beginTransaction();
49: $pathologyTest = PathologyTest::create([
65: PathologyParameterItem::create([
73: DB::commit();
76: DB::rollBack();
```

### update (line 81)

`public function update($input, $pathologyTest)`

Guard/validation candidates:

```php
100: if ($input['parameter_id']) {
```

Write/transaction candidates:

```php
83: DB::beginTransaction();
85: $pathologyTest->update([
98: $pathologyTest->parameterItems()->delete();
102: PathologyParameterItem::create([
110: DB::commit();
113: DB::rollBack();
```

### getPathologyAssociatedData (line 118)

`public function getPathologyAssociatedData()`

### getParameterDataList (line 127)

`public function getParameterDataList()`

### getParameterItemData (line 142)

`public function getParameterItemData($id)`

### getSettingList (line 147)

`public function getSettingList(): array`

## hms/app/Repositories/PathologyUnitRepository.php

Models: `PathologyUnit` → `pathology_units`

### getFieldsSearchable (line 18)

`public function getFieldsSearchable(): array`

### model (line 23)

`public function model()`

## hms/app/Repositories/PatientAdmissionRepository.php

Models: `Bed` → `beds`, `Insurance` → `insurances`, `Package` → `packages`, `PatientAdmission` → `patient_admissions`

### getFieldsSearchable (line 34)

`public function getFieldsSearchable()`

### model (line 39)

`public function model()`

### getSyncList (line 44)

`public function getSyncList($patientAdmission = null)`

Guard/validation candidates:

```php
55: if (isset($patientAdmission)) {
```

### setBedAvailable (line 66)

`public function setBedAvailable($bedId)`

Guard/validation candidates:

```php
71: return true;
```

Write/transaction candidates:

```php
69: $bed->update(['is_available' => 1]);
```

### setBedUnAvailable (line 74)

`public function setBedUnAvailable($bedId)`

Guard/validation candidates:

```php
79: return true;
```

Write/transaction candidates:

```php
77: $bed->update(['is_available' => 0]);
```

### store (line 82)

`public function store($input)`

Guard/validation candidates:

```php
92: if (isset($input['bed_id'])) {
96: return true;
```

Write/transaction candidates:

```php
90: PatientAdmission::create($input);
```

### update (line 102)

`public function update($input, $patientAdmission)`

Guard/validation candidates:

```php
115: if (isset($bedId)) {
118: if (isset($input['bed_id'])) {
122: if (isset($input['bed_id']) && (isset($input['discharge_date']))) {
126: return true;
```

Write/transaction candidates:

```php
113: $patientAdmission->update($input);
```

## hms/app/Repositories/PatientCaseRepository.php

Models: `CaseHandler` → `case_handlers`, `Doctor` → `doctors`, `Notification` → `notifications`, `Patient` → `patients`, `PatientCase` → `patient_cases`, `Receptionist` → `receptionists`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable()`

### model (line 33)

`public function model()`

### getPatients (line 38)

`public function getPatients()`

### getDoctors (line 45)

`public function getDoctors()`

### store (line 52)

`public function store($input)`

Guard/validation candidates:

```php
58: return true;
```

Write/transaction candidates:

```php
56: $patientCase = PatientCase::create($input);
```

### createNotification (line 64)

`public function createNotification($input)`

Guard/validation candidates:

```php
86: if ($notification == Notification::NOTIFICATION_FOR[Notification::PATIENT]) {
```

## hms/app/Repositories/PatientDiagnosisTestRepository.php

Models: `DiagnosisCategory` → `diagnosis_categories`, `PatientDiagnosisProperty` → `patient_diagnosis_properties`, `PatientDiagnosisTest` → `patient_diagnosis_tests`, `Setting` → `settings`

### getFieldsSearchable (line 21)

`public function getFieldsSearchable()`

### model (line 26)

`public function model()`

### getUniqueReportNumber (line 31)

`public static function getUniqueReportNumber()`

Guard/validation candidates:

```php
36: if ($isExist) {
```

### getDiagnosisCategory (line 45)

`public function getDiagnosisCategory()`

### getPatientDiagnosisTestProperty (line 52)

`public function getPatientDiagnosisTestProperty($patientDiagnosisTestId)`

### store (line 59)

`public function store($input)`

Guard/validation candidates:

```php
75: if (isset($input['property_name']) && ! empty($input['property_name'])) {
80: if (! empty($data['property_name'])) {
87: return true;
```

Write/transaction candidates:

```php
61: $patientDiagnosisTest = PatientDiagnosisTest::create(Arr::only($input,
68: PatientDiagnosisProperty::create([
82: PatientDiagnosisProperty::create($data);
```

### updatePatientDiagnosis (line 90)

`public function updatePatientDiagnosis($input, $patientDiagnosisTest)`

Guard/validation candidates:

```php
109: if (isset($input['property_name']) && ! empty($input['property_name'])) {
114: if (! empty($data['property_name'])) {
121: return true;
```

Write/transaction candidates:

```php
92: $patientDiagnosisTest->update(Arr::only($input,
96: $diagnosisProperty->delete();
102: PatientDiagnosisProperty::create([
116: PatientDiagnosisProperty::create($data);
```

### prepareInputForPatientDiagnosisTest (line 124)

`public function prepareInputForPatientDiagnosisTest($input)`

### getSettingList (line 136)

`public function getSettingList()`

## hms/app/Repositories/PatientIdCardTemplateRepository.php

Models: `PatientIdCardTemplate` → `patient_id_card_templates`

### getFieldsSearchable (line 20)

`public function getFieldsSearchable()`

### model (line 25)

`public function model()`

### create (line 30)

`public function create($input)`

Guard/validation candidates:

```php
41: return true;
```

Write/transaction candidates:

```php
39: PatientIdCardTemplate::create($input);
```

### update (line 44)

`public function update($id, $input)`

Guard/validation candidates:

```php
56: return true;
```

Write/transaction candidates:

```php
54: $PatientIdCardTemplate->update($input);
```

## hms/app/Repositories/PatientRepository.php

Models: `Address` → `addresses`, `Department` → `departments`, `Notification` → `notifications`, `Patient` → `patients`, `Receptionist` → `receptionists`, `User` → `users`, `PatientAdmission` → `patient_admissions`

### getFieldsSearchable (line 29)

`public function getFieldsSearchable()`

### model (line 34)

`public function model()`

### store (line 39)

`public function store($input, $mail = true)`

Guard/validation candidates:

```php
51: if ($mail) {
55: if (isset($input['image']) && ! empty($input['image'])) {
61: if (strpos($key, 'field') === 0) {
71: if (! empty($address = Address::prepareAddressArray($input))) {
81: return true;
```

Write/transaction candidates:

```php
49: $user = User::create($input);
66: $patient = Patient::create(['user_id' => $user->id, 'patient_unique_id' => strtoupper(Patient::generateUniquePatientId()), 'custom_field' => !empty($jsonFields) ? $jsonFields : null]);
72: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
75: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
```

### update (line 84)

`public function update($input, $patient)`

Guard/validation candidates:

```php
91: if (strpos($key, 'field') === 0) {
97: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
100: if (isset($input['image']) && ! empty($input['image'])) {
110: if (! empty($patient->address)) {
111: if (empty($address = Address::prepareAddressArray($input))) {
116: if (! empty($address = Address::prepareAddressArray($input)) && empty($patient->address)) {
126: return true;
```

Write/transaction candidates:

```php
107: $patient->patientUser->update($input);
108: $patient->update($input);
112: $patient->address->delete();
114: $patient->address->update($input);
119: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
```

### getPatients (line 129)

`public function getPatients()`

### getPatientAssociatedData (line 137)

`public function getPatientAssociatedData($patientId)`

### createNotification (line 157)

`public function createNotification($input)`

Guard/validation candidates:

```php
169: if (isset($key)) {
```

## hms/app/Repositories/PaymentRepository.php

Models: `Account` → `accounts`, `Payment` → `payments`

### getFieldsSearchable (line 22)

`public function getFieldsSearchable()`

### model (line 27)

`public function model()`

### getAccounts (line 32)

`public function getAccounts()`

## hms/app/Repositories/PharmacistRepository.php

Models: `Address` → `addresses`, `Department` → `departments`, `Pharmacist` → `pharmacists`, `User` → `users`

### getFieldsSearchable (line 25)

`public function getFieldsSearchable()`

### model (line 30)

`public function model()`

### store (line 35)

`public function store($input, $mail = true)`

Guard/validation candidates:

```php
46: if ($mail) {
50: if (isset($input['image']) && ! empty($input['image'])) {
58: if (! empty($address = Address::prepareAddressArray($input))) {
68: return true;
```

Write/transaction candidates:

```php
45: $user = User::create($input);
54: $pharmacist = Pharmacist::create(['user_id' => $user->id]);
59: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
62: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
```

### update (line 71)

`public function update($input, $pharmacist)`

Guard/validation candidates:

```php
78: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
82: if (isset($input['image']) && ! empty($input['image'])) {
91: if (! empty($pharmacist->address)) {
92: if (empty($address = Address::prepareAddressArray($input))) {
97: if (! empty($address = Address::prepareAddressArray($input)) && empty($pharmacist->address)) {
107: return true;
```

Write/transaction candidates:

```php
88: $pharmacist->user->update($input);
89: $pharmacist->update($input);
93: $pharmacist->address->delete();
95: $pharmacist->address->update($input);
100: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
```

## hms/app/Repositories/PostalRepository.php

Models: `Postal` → `postals`

### getFieldsSearchable (line 22)

`public function getFieldsSearchable()`

### model (line 27)

`public function model()`

### store (line 32)

`public function store($input)`

Guard/validation candidates:

```php
37: if (! empty($input['attachment'])) {
44: return true;
```

Write/transaction candidates:

```php
35: $postal = $this->create($input);
```

### updatePostal (line 50)

`public function updatePostal($input, $postalId)`

Guard/validation candidates:

```php
55: if (! empty($input['attachment'])) {
62: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
```

Write/transaction candidates:

```php
53: $postal = $this->update($input, $postalId);
```

### deleteDocument (line 70)

`public function deleteDocument($postalId)`

Write/transaction candidates:

```php
75: $this->delete($postalId);
```

### downloadMedia (line 81)

`public function downloadMedia($postal)`

Guard/validation candidates:

```php
86: if (config('app.media_disc') === 'public') {
```

## hms/app/Repositories/PrescriptionRepository.php

Models: `Medicine` → `medicines`, `MedicineBill` → `medicine_bills`, `Notification` → `notifications`, `Patient` → `patients`, `Prescription` → `prescriptions`, `PrescriptionMedicineModal` → `prescriptions_medicines`, `SaleMedicine` → `sale_medicines`, `Setting` → `settings`

### getFieldsSearchable (line 46)

`public function getFieldsSearchable()`

### model (line 51)

`public function model()`

### getPatients (line 56)

`public function getPatients()`

Guard/validation candidates:

```php
60: if ($user->hasRole('Doctor')) {
```

### update (line 72)

`public function update($prescription, $input)`

Guard/validation candidates:

```php
77: return true;
```

Write/transaction candidates:

```php
75: $prescription->update($input);
```

### createNotification (line 83)

`public function createNotification($input)`

### createPrescription (line 99)

`public function createPrescription($input, $prescription)`

Guard/validation candidates:

```php
105: if (isset($input['medicine'])) {
```

Write/transaction candidates:

```php
106: $medicineBill = MedicineBill::create([
126: $prescriptionMedicine = PrescriptionMedicineModal::create($PrescriptionItem);
137: SaleMedicine::create($saleMedicineArray);
139: $medicineBill->update([
```

### updatePrescription (line 149)

`public function updatePrescription($prescription, $input)`

Guard/validation candidates:

```php
160: if (! empty($input['medicine'])) {
```

Write/transaction candidates:

```php
153: $prescription->update($prescriptionMedicineArr);
155: $prescription->getMedicine()->delete();
156: $medicineBill->saleMedicine()->delete();
171: $prescriptionMedicine = PrescriptionMedicineModal::create($PrescriptionItem);
182: SaleMedicine::create($saleMedicineArray);
184: $medicineBill->update([
```

### prepareInputForServicePackageItem (line 196)

`public function prepareInputForServicePackageItem($prescriptionMedicineArr)`

### getData (line 208)

`public function getData($id)`

Guard/validation candidates:

```php
212: if (empty($data)) {
```

### getMedicineData (line 220)

`public function getMedicineData($id)`

### getSyncListForCreate (line 232)

`public function getSyncListForCreate($prescriptionId = null)`

### getSettingList (line 239)

`public function getSettingList()`

### getMedicines (line 246)

`public function getMedicines()`

## hms/app/Repositories/PurchaseMedicineRepository.php

Models: `Accountant` → `accountants`, `Address` → `addresses`, `Category` → `categories`, `Medicine` → `medicines`, `PurchasedMedicine` → `purchased_medicines`, `PurchaseMedicine` → `purchase_medicines`, `User` → `users`

### getFieldsSearchable (line 35)

`public function getFieldsSearchable()`

### model (line 40)

`public function model()`

### getMedicine (line 45)

`public function getMedicine()`

### getMedicineList (line 52)

`public function getMedicineList()`

### getCategoryList (line 67)

`public function getCategoryList()`

### getCategory (line 82)

`public function getCategory()`

### store (line 90)

`public function store($input)`

Guard/validation candidates:

```php
121: return true;
```

Write/transaction candidates:

```php
93: DB::beginTransaction();
95: $purchaseMedicine = PurchaseMedicine::create($purchaseMedicineArray);
110: PurchasedMedicine::create($purchasedMedicineArray);
116: $medicine->update($medicineQtyArray);
119: DB::commit();
123: DB::rollBack();
```

### update (line 128)

`public function update($accountant, $input)`

Guard/validation candidates:

```php
135: if (isset($input['image']) && ! empty($input['image'])) {
139: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
148: if (! empty($accountant->address)) {
149: if (empty($address = Address::prepareAddressArray($input))) {
154: if (! empty($address = Address::prepareAddressArray($input)) && empty($accountant->address)) {
161: return true;
```

Write/transaction candidates:

```php
145: $accountant->user->update($input);
146: $accountant->update($input);
150: $accountant->address->delete();
152: $accountant->address->update($input);
157: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
```

### stripeSession (line 167)

`public function stripeSession($input)`

Write/transaction candidates:

```php
171: $session = Session::create([
```

### purchaseMedicinestripeSuccess (line 198)

`public function purchaseMedicinestripeSuccess($input)`

Guard/validation candidates:

```php
202: if (empty($sessionId)) {
210: if($sessionData){
211: return true;
214: return false;
```

### razorPayPayment (line 217)

`public function razorPayPayment($input)`

Write/transaction candidates:

```php
232: $razorpayOrder = $api->order->create($orderData);
```

### razorPaySuccess (line 239)

`public function razorPaySuccess($input)`

Guard/validation candidates:

```php
243: if (count($input) && ! empty($input['razorpay_payment_id'])) {
248: if ($generatedSignature != $input['razorpay_signature']) {
255: return true;
```

### paystackPaymentSuccess (line 259)

`public function paystackPaymentSuccess($response)`

Guard/validation candidates:

```php
293: return true;
```

Write/transaction candidates:

```php
264: DB::beginTransaction();
267: $purchaseMedicine = PurchaseMedicine::create($purchaseMedicineArray);
282: PurchasedMedicine::create($purchasedMedicineArray);
288: $medicine->update($medicineQtyArray);
291: DB::commit();
296: DB::rollBack();
```

### phonePePayment (line 301)

`public function phonePePayment($input)`

### phonePePaymentSuccess (line 373)

`public function phonePePaymentSuccess($input)`

Guard/validation candidates:

```php
406: return true;
411: return false;
```

Write/transaction candidates:

```php
378: DB::beginTransaction();
381: $purchaseMedicine = PurchaseMedicine::create($purchaseMedicineArray);
395: PurchasedMedicine::create($purchasedMedicineArray);
401: $medicine->update($medicineQtyArray);
404: DB::commit();
408: DB::rollBack();
```

### flutterWavePayment (line 414)

`public function flutterWavePayment($input)`

Guard/validation candidates:

```php
439: if ($payment['status'] !== 'success') {
```

### flutterwavePaymentSuccess (line 448)

`public function flutterwavePaymentSuccess($input)`

Guard/validation candidates:

```php
453: if ($input['status'] ==  'successful')
460: if(isset($sessionData) && !empty($sessionData)){
488: return true;
494: return false;
```

Write/transaction candidates:

```php
451: DB::beginTransaction();
462: $purchaseMedicine = PurchaseMedicine::create($purchaseMedicineArray);
476: PurchasedMedicine::create($purchasedMedicineArray);
482: $medicine->update($medicineQtyArray);
486: DB::commit();
491: DB::rollBack();
```

## hms/app/Repositories/RadiologyCategoryRepository.php

Models: `RadiologyCategory` → `radiology_categories`

### getFieldsSearchable (line 18)

`public function getFieldsSearchable()`

### model (line 23)

`public function model()`

## hms/app/Repositories/RadiologyTestRepository.php

Models: `ChargeCategory` → `charge_categories`, `RadiologyCategory` → `radiology_categories`, `RadiologyTest` → `radiology_tests`

### getFieldsSearchable (line 27)

`public function getFieldsSearchable()`

### model (line 32)

`public function model()`

### getRadiologyAssociatedData (line 37)

`public function getRadiologyAssociatedData()`

## hms/app/Repositories/ReceptionistRepository.php

Models: `Address` → `addresses`, `Department` → `departments`, `Receptionist` → `receptionists`, `User` → `users`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable()`

### model (line 33)

`public function model()`

### store (line 38)

`public function store($input, $mail = true)`

Guard/validation candidates:

```php
49: if ($mail) {
53: if (isset($input['image']) && ! empty($input['image'])) {
62: if (! empty($address = Address::prepareAddressArray($input))) {
69: return true;
```

Write/transaction candidates:

```php
48: $user = User::create($input);
57: $receptionist = Receptionist::create(['user_id' => $user->id]);
63: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
66: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
```

### update (line 75)

`public function update($receptionist, $input)`

Guard/validation candidates:

```php
81: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
84: if (isset($input['image']) && ! empty($input['image'])) {
93: if (! empty($receptionist->address)) {
94: if (empty($address = Address::prepareAddressArray($input))) {
99: if (! empty($address = Address::prepareAddressArray($input)) && empty($receptionist->address)) {
106: return true;
```

Write/transaction candidates:

```php
90: $receptionist->user->update($input);
91: $receptionist->update($input);
95: $receptionist->address->delete();
97: $receptionist->address->update($input);
102: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
```

## hms/app/Repositories/ScheduleRepository.php

Models: `Doctor` → `doctors`, `DoctorHoliday` → `doctor_holidays`, `HospitalSchedule` → `hospital_schedules`, `LunchBreak` → `lunch_breaks`, `Schedule` → `schedules`, `ScheduleDay` → `schedule_days`

### getFieldsSearchable (line 27)

`public function getFieldsSearchable()`

### model (line 32)

`public function model()`

### getData (line 37)

`public function getData()`

Guard/validation candidates:

```php
43: if ($user->hasRole('Doctor')) {
```

### prepareInputForScheduleDayItem (line 54)

`public function prepareInputForScheduleDayItem($input)`

### store (line 66)

`public function store($input)`

Guard/validation candidates:

```php
79: return true;
```

Write/transaction candidates:

```php
68: $schedule = Schedule::create($input);
76: $scheduleDay = ScheduleDay::create($data);
```

### update (line 82)

`public function update($input, $id)`

Guard/validation candidates:

```php
99: return true;
```

Write/transaction candidates:

```php
85: $schedule->update($input);
96: $scheduleDay->update($data);
```

### getDoctorSchedule (line 102)

`public function getDoctorSchedule($input)`

Guard/validation candidates:

```php
111: if(isset($input['date'])){
114: if($data['break']->count() == 0){
```

## hms/app/Repositories/ServiceRepository.php

Models: `Accountant` → `accountants`, `Notification` → `notifications`, `Receptionist` → `receptionists`, `Service` → `services`, `User` → `users`

### getFieldsSearchable (line 28)

`public function getFieldsSearchable()`

### model (line 33)

`public function model()`

### createNotification (line 38)

`public function createNotification()`

Guard/validation candidates:

```php
56: return true;
```

## hms/app/Repositories/SettingRepository.php

Models: `Setting` → `settings`

### getFieldsSearchable (line 22)

`public function getFieldsSearchable()`

### model (line 27)

`public function model()`

### getSyncList (line 32)

`public function getSyncList()`

### updateSetting (line 37)

`public function updateSetting($input)`

Guard/validation candidates:

```php
40: if (isset($input['app_logo']) && ! empty($input['app_logo'])) {
48: if (isset($input['favicon']) && ! empty($input['favicon'])) {
60: if ($country_code->value == $input['country_code']) {
77: if ($setting) {
```

Write/transaction candidates:

```php
46: $setting->update(['value' => $setting->logo_url]);
54: $setting->update(['value' => $setting->logo_url]);
78: $setting->update(['value' => $value]);
80: Setting::create([
85: Setting::where('key', $key)->update(['value' => $value]);
```

## hms/app/Repositories/SmsRepository.php

Models: `Sms` → `sms`, `User` → `users`

### getFieldsSearchable (line 22)

`public function getFieldsSearchable()`

### model (line 27)

`public function model()`

### store (line 32)

`public function store($input)`

Guard/validation candidates:

```php
34: if (! isset($input['number'])) {
```

### sendSMS (line 44)

`public function sendSMS($sendTo, $regionCode, $phone, $message)`

Write/transaction candidates:

```php
51: $sms = Sms::create([
59: $client->messages->create(
```

## hms/app/Repositories/StripeRepository.php

Models: `IpdPayment` → `ipd_payments`, `Transaction` → `transactions`

### patientPaymentSuccess (line 15)

`public function patientPaymentSuccess($sessionId)`

Guard/validation candidates:

```php
54: if ($ipdBill) {
```

Write/transaction candidates:

```php
35: DB::beginTransaction();
37: $transaction = Transaction::create($transactionData);
48: $ipdPayment->store($ipdPaymentData);
57: $ipdBill->save();
60: $ipdPatientDepartment->save();
63: DB::commit();
65: DB::rollBack();
```

## hms/app/Repositories/TestimonialRepository.php

Models: `Testimonial` → `testimonials`

### getFieldsSearchable (line 19)

`public function getFieldsSearchable()`

### model (line 24)

`public function model()`

### store (line 29)

`public function store($input)`

Guard/validation candidates:

```php
34: if (! empty($input['profile'])) {
40: return true;
```

Write/transaction candidates:

```php
32: $testimonial = $this->create($input);
```

### updateTestimonial (line 46)

`public function updateTestimonial($input, $testimonialId)`

Guard/validation candidates:

```php
50: if (! empty($input['profile'])) {
```

Write/transaction candidates:

```php
49: $testimonial = $this->update($input, $testimonialId);
```

### deleteTestimonial (line 61)

`public function deleteTestimonial($testimonial)`

Write/transaction candidates:

```php
65: $this->delete($testimonial->id);
```

## hms/app/Repositories/UserRepository.php

Models: `Accountant` → `accountants`, `AdvancedPayment` → `advanced_payments`, `Appointment` → `appointments`, `BedAssign` → `bed_assigns`, `Bill` → `bills`, `BirthReport` → `birth_reports`, `CaseHandler` → `case_handlers`, `DeathReport` → `death_reports`, `Doctor` → `doctors`, `DoctorDepartment` → `doctor_departments`, `EmployeePayroll` → `employee_payrolls`, `InvestigationReport` → `investigation_reports`, `Invoice` → `invoices`, `IpdPatientDepartment` → `ipd_patient_departments`, `LabTechnician` → `lab_technicians`, `Nurse` → `nurses`, `OperationReport` → `operation_reports`, `Patient` → `patients`, `PatientAdmission` → `patient_admissions`, `PatientCase` → `patient_cases`, `Pharmacist` → `pharmacists`, `Prescription` → `prescriptions`, `Receptionist` → `receptionists`, `Schedule` → `schedules`, `Setting` → `settings`, `User` → `users`

### getFieldsSearchable (line 52)

`public function getFieldsSearchable()`

### model (line 57)

`public function model()`

### profileUpdate (line 62)

`public function profileUpdate($input)`

Guard/validation candidates:

```php
66: if (empty($input['image']) && $input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
69: if (isset($input['image']) && ! empty($input['image'])) {
75: return true;
```

Write/transaction candidates:

```php
73: $user->update($input);
```

### changePassword (line 81)

`public function changePassword($input)`

Guard/validation candidates:

```php
85: if (! Hash::check($input['password_current'], $user->password)) {
91: return true;
```

Write/transaction candidates:

```php
89: $user->update($input);
```

### store (line 97)

`public function store($input)`

Guard/validation candidates:

```php
107: if (isset($input['image']) && ! empty($input['image'])) {
113: if ($input['department_id'] == 1) {
176: return true;
```

Write/transaction candidates:

```php
105: $user = User::create($input);
118: $doctor = Doctor::create([
129: $user->update([
133: $patient = Patient::create(['user_id' => $user->id,'patient_unique_id' => strtoupper(Patient::generateUniquePatientId())]);
138: $nurse = Nurse::create(['user_id' => $user->id]);
143: $receptionist = Receptionist::create(['user_id' => $user->id]);
148: $pharmacist = Pharmacist::create(['user_id' => $user->id]);
153: $accountant = Accountant::create(['user_id' => $user->id]);
158: $caseManager = CaseHandler::create(['user_id' => $user->id]);
163: $labTechnician = LabTechnician::create(['user_id' => $user->id]);
169: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType,]);
```

### updateUser (line 179)

`public function updateUser($input, $userId)`

Guard/validation candidates:

```php
183: if (isset($input['image']) && ! empty($input['image'])) {
191: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
195: return true;
```

Write/transaction candidates:

```php
182: $user = $this->update($input, $userId);
188: $user->update(['updated_at' => Carbon::now()->timestamp]);
```

### deleteUser (line 201)

`public function deleteUser($userId)`

Guard/validation candidates:

```php
206: if ($user->department_id == 2) {
215: if ($result || $empPayRollResult) {
230: if ($result) {
240: if ($empPayRollResult) {
249: if ($empPayRollResult) {
259: if ($empPayRollResult) {
269: if ($empPayRollResult) {
281: if ($empPayRollResult) {
```

Write/transaction candidates:

```php
222: Doctor::whereId($user->owner_id)->delete();
237: Patient::whereId($user->owner_id)->delete();
256: Receptionist::whereId($user->owner_id)->delete();
266: Pharmacist::whereId($user->owner_id)->delete();
276: Accountant::whereId($user->owner_id)->delete();
278: caseHandler::whereId($user->owner_id)->delete();
288: LabTechnician::whereId($user->owner_id)->delete();
292: $this->delete($userId);
```

### getUserData (line 298)

`public function getUserData($user)`

### profileApiUpdate (line 305)

`public function profileApiUpdate($input)`

Guard/validation candidates:

```php
310: if (isset($input['image']) && ! empty($input['image'])) {
```

Write/transaction candidates:

```php
314: $user->update($input);
```

## hms/app/Repositories/VaccinatedPatientRepository.php

Models: `Patient` → `patients`, `VaccinatedPatients` → `vaccinated_patients`, `Vaccination` → `vaccinations`

### getFieldsSearchable (line 25)

`public function getFieldsSearchable()`

### model (line 30)

`public function model()`

### getVaccinatedPatientData (line 35)

`public function getVaccinatedPatientData()`

Guard/validation candidates:

```php
41: if ($user->hasRole('Doctor')) {
```

## hms/app/Repositories/VaccinationRepository.php

Models: `Vaccination` → `vaccinations`

### getFieldsSearchable (line 20)

`public function getFieldsSearchable()`

### model (line 25)

`public function model()`

## hms/app/Repositories/VisitorRepository.php

Models: `Notification` → `notifications`, `Receptionist` → `receptionists`, `Visitor` → `visitors`

### getFieldsSearchable (line 26)

`public function getFieldsSearchable()`

### model (line 31)

`public function model()`

### store (line 36)

`public function store($input)`

Guard/validation candidates:

```php
41: if (! empty($input['attachment'])) {
65: return true;
```

Write/transaction candidates:

```php
39: $visitor = $this->create($input);
```

### updateVisitor (line 71)

`public function updateVisitor($input, $visitorId)`

Guard/validation candidates:

```php
76: if (! empty($input['attachment'])) {
83: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
```

Write/transaction candidates:

```php
74: $visitor = $this->update($input, $visitorId);
```

### deleteDocument (line 91)

`public function deleteDocument($visitorId)`

Write/transaction candidates:

```php
96: $this->delete($visitorId);
```

### downloadMedia (line 102)

`public function downloadMedia($visitor)`

Guard/validation candidates:

```php
107: if (config('app.media_disc') === 'public') {
```

## hms/app/Repositories/ZoomRepository.php

Models: `LiveConsultation` → `live_consultations`, `LiveMeeting` → `live_meetings`, `UserZoomCredential` → `user_zoom_credential`, `ZoomOAuth` → `zoom_o_auth_credentials`

### connectWithZoom (line 24)

`public function connectWithZoom($code)`

Guard/validation candidates:

```php
28: if (isset($userZoomCredential)) {
51: if (! $exist) {
64: return true;
```

Write/transaction candidates:

```php
52: ZoomOAuth::create([
58: $exist->update([
```

### updateZoomMeeting (line 67)

`public function updateZoomMeeting($data, $liveConsultation)`

Guard/validation candidates:

```php
97: if ($e->getCode() == 401) {
100: if ($e->getCode() == 0) {
```

### createZoomMeeting (line 107)

`public function createZoomMeeting($data)`

Guard/validation candidates:

```php
113: if(empty($zoomOAuth)){
139: if (401 == $e->getCode()) {
140: if (! isset($clientID)) {
146: if (! isset($userZoomCredential)) {
164: if ($e->getCode() == 401) {
167: if ($e->getCode() == 0) {
```

Write/transaction candidates:

```php
160: $zoomOAuth->update(['refresh_token' => $response->getBody()]);
```

### toZoomTimeFormat (line 175)

`public function toZoomTimeFormat($dateTime)`

### zoomGet (line 188)

`public function zoomGet($id)`

Guard/validation candidates:

```php
195: if (isset($liveCunsultation)) {
198: if (isset($liveMeeting)) {
211: if ($e->getCode() == 401) {
214: if ($e->getCode() == 0) {
```

### destroyZoomMeeting (line 221)

`public function destroyZoomMeeting($meetingId)`

Guard/validation candidates:

```php
241: if ($e->getCode() == 401) {
244: if ($e->getCode() == 400) {
247: if ($e->getCode() == 0) {
```

## hms/app/Repositories/adminRepository.php

Models: `Address` → `addresses`, `admin` → `admins`, `Department` → `departments`, `User` → `users`

### getFieldsSearchable (line 26)

`public function getFieldsSearchable()`

### model (line 31)

`public function model()`

### store (line 36)

`public function store($input, $mail = true)`

Guard/validation candidates:

```php
48: if (isset($input['image']) && ! empty($input['image'])) {
56: if (! empty($address = Address::prepareAddressArray($input))) {
63: return true;
```

Write/transaction candidates:

```php
46: $user = User::create($input);
52: $admin = Admin::create(['user_id' => $user->id]);
57: Address::create(array_merge($address, ['owner_id' => $ownerId, 'owner_type' => $ownerType]));
60: $user->update(['owner_id' => $ownerId, 'owner_type' => $ownerType]);
```

### update (line 69)

`public function update($admin, $input)`

Guard/validation candidates:

```php
72: if ($input['avatar_remove'] == 1 && isset($input['avatar_remove']) && ! empty($input['avatar_remove'])) {
76: if (isset($input['image']) && ! empty($input['image'])) {
84: return true;
```

Write/transaction candidates:

```php
82: $admin->update($input);
```

## hms/app/Repositories/currency_settingRepository.php

Models: `CurrencySetting` → `currency_settings`

### getFieldsSearchable (line 20)

`public function getFieldsSearchable()`

### model (line 25)

`public function model()`

### create (line 30)

`public function create($input)`

Guard/validation candidates:

```php
40: return true;
```

Write/transaction candidates:

```php
38: CurrencySetting::create($data);
```
