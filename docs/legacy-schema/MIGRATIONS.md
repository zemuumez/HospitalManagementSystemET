# Migration evidence

All application migration up-directions, in filename order. Blueprint excerpts are structural evidence, not an executed final schema. Seeder call names are listed without seed contents. Configured permission table names resolve through `hms/config/permission.php`.

## 2014_10_12_000000_create_users_table.php

Source: `hms/database/migrations/2014_10_12_000000_create_users_table.php`

- `create` `users` at line 14

```php
$table->bigIncrements('id');
$table->bigInteger('department_id')->unsigned()->nullable();
$table->string('first_name', 100);
$table->string('last_name', 100)->nullable();
$table->string('email', 100)->unique();
$table->string('password', 100);
$table->string('designation', 100)->nullable();
$table->string('phone', 100)->nullable();
$table->integer('gender');
$table->string('qualification', 100)->nullable();
$table->string('blood_group', 100)->nullable();
$table->date('dob')->nullable();
$table->timestamp('email_verified_at')->nullable();
$table->integer('owner_id')->nullable();
$table->string('owner_type', 100)->nullable();
$table->boolean('status');
$table->string('language', 100)->default('en');
$table->rememberToken();
$table->timestamps();
```

## 2014_10_12_100000_create_password_resets_table.php

Source: `hms/database/migrations/2014_10_12_100000_create_password_resets_table.php`

- `create` `password_resets` at line 14

```php
$table->string('email', 160)->index();
$table->string('token');
$table->timestamp('created_at')->nullable();
```

## 2019_08_19_000000_create_failed_jobs_table.php

Source: `hms/database/migrations/2019_08_19_000000_create_failed_jobs_table.php`

- `create` `failed_jobs` at line 14

```php
$table->bigIncrements('id');
$table->text('connection');
$table->text('queue');
$table->longText('payload');
$table->longText('exception');
$table->timestamp('failed_at')->useCurrent();
```

## 2019_12_14_000001_create_personal_access_tokens_table.php

Source: `hms/database/migrations/2019_12_14_000001_create_personal_access_tokens_table.php`

- `create` `personal_access_tokens` at line 14

```php
$table->id();
$table->morphs('tokenable');
$table->string('name');
$table->string('token', 64)->unique();
$table->text('abilities')->nullable();
$table->timestamp('last_used_at')->nullable();
$table->timestamps();
```

## 2020_02_06_031618_create_categories_table.php

Source: `hms/database/migrations/2020_02_06_031618_create_categories_table.php`

- `create` `categories` at line 13

```php
$table->increments('id');
$table->string('name', 160);
$table->boolean('is_active')->default(false);
$table->timestamps();
```

## 2020_02_12_053840_create_doctor_departments_table.php

Source: `hms/database/migrations/2020_02_12_053840_create_doctor_departments_table.php`

- `create` `doctor_departments` at line 13

```php
$table->bigIncrements('id');
$table->string('title', 160);
$table->text('description')->nullable();
$table->timestamps();
```

## 2020_02_12_053932_create_departments_table.php

Source: `hms/database/migrations/2020_02_12_053932_create_departments_table.php`

- `create` `departments` at line 13

```php
$table->bigIncrements('id');
$table->string('name');
$table->boolean('is_active')->default(0);
$table->string('guard_name')->nullable();
$table->timestamps();
```

## 2020_02_13_042835_create_brands_table.php

Source: `hms/database/migrations/2020_02_13_042835_create_brands_table.php`

- `create` `brands` at line 13

```php
$table->increments('id');
$table->string('name', 160);
$table->string('email')->nullable();
$table->string('phone')->nullable();
$table->timestamps();
```

## 2020_02_13_053840_create_doctors_table.php

Source: `hms/database/migrations/2020_02_13_053840_create_doctors_table.php`

- `create` `doctors` at line 14

```php
$table->bigIncrements('id');
$table->unsignedBigInteger('user_id');
$table->unsignedBigInteger('doctor_department_id');
$table->string('specialist');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_department_id')->references('id')->on('doctor_departments');
```

## 2020_02_13_054103_create_patients_table.php

Source: `hms/database/migrations/2020_02_13_054103_create_patients_table.php`

- `create` `patients` at line 13

```php
$table->increments('id');
$table->unsignedBigInteger('user_id');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_13_094724_create_bills_table.php

Source: `hms/database/migrations/2020_02_13_094724_create_bills_table.php`

- `create` `bills` at line 14

```php
$table->increments('id');
$table->string('bill_id');
$table->unsignedInteger('patient_id');
$table->datetime('bill_date');
$table->double('amount', 8, 2)->nullable();
$table->timestamps();
```

## 2020_02_13_095024_create_medicines_table.php

Source: `hms/database/migrations/2020_02_13_095024_create_medicines_table.php`

- `create` `medicines` at line 13

```php
$table->increments('id');
$table->integer('category_id')->unsigned()->nullable();
$table->integer('brand_id')->unsigned()->nullable();
$table->string('name');
$table->double('selling_price');
$table->double('buying_price');
$table->string('salt_composition');
$table->text('description')->nullable();
$table->text('side_effects')->nullable();
$table->timestamps();
$table->foreign('category_id')->references('id')->on('categories') ->onDelete('set null') ->onUpdate('cascade');
$table->foreign('brand_id')->references('id')->on('brands') ->onDelete('set null') ->onUpdate('cascade');
```

## 2020_02_13_095125_create_bill_items_table.php

Source: `hms/database/migrations/2020_02_13_095125_create_bill_items_table.php`

- `create` `bill_items` at line 14

```php
$table->increments('id');
$table->string('item_name');
$table->unsignedInteger('bill_id');
$table->integer('qty')->unsigned();
$table->double('price', 8, 2);
$table->double('amount', 8, 2);
$table->timestamps();
$table->foreign('bill_id')->references('id')->on('bills') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_13_111857_create_nurses_table.php

Source: `hms/database/migrations/2020_02_13_111857_create_nurses_table.php`

- `create` `nurses` at line 13

```php
$table->increments('id');
$table->unsignedBigInteger('user_id');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_13_125601_create_addresses_table.php

Source: `hms/database/migrations/2020_02_13_125601_create_addresses_table.php`

- `create` `addresses` at line 14

```php
$table->bigIncrements('id');
$table->integer('owner_id')->nullable();
$table->string('owner_type')->nullable();
$table->string('address1')->nullable();
$table->string('address2')->nullable();
$table->string('city')->nullable();
$table->string('zip')->nullable();
$table->timestamps();
```

## 2020_02_13_141104_create_media_table.php

Source: `hms/database/migrations/2020_02_13_141104_create_media_table.php`

- `create` `media` at line 14

```php
$table->bigIncrements('id');
$table->morphs('model');
$table->string('collection_name', 160);
$table->string('name', 160);
$table->string('file_name', 160);
$table->string('mime_type', 160)->nullable();
$table->string('disk', 160);
$table->unsignedBigInteger('size');
$table->text('manipulations');
$table->text('custom_properties');
$table->text('responsive_images');
$table->unsignedInteger('order_column')->nullable();
$table->nullableTimestamps();
```

## 2020_02_14_051650_create_lab_technicians_table.php

Source: `hms/database/migrations/2020_02_14_051650_create_lab_technicians_table.php`

- `create` `lab_technicians` at line 14

```php
$table->bigIncrements('id');
$table->unsignedBigInteger('user_id');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_14_055353_create_appointments_table.php

Source: `hms/database/migrations/2020_02_14_055353_create_appointments_table.php`

- `create` `appointments` at line 14

```php
$table->bigIncrements('id');
$table->unsignedInteger('patient_id');
$table->unsignedBigInteger('doctor_id');
$table->unsignedBigInteger('department_id');
$table->dateTime('opd_date');
$table->text('problem')->nullable();
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('department_id')->references('id')->on('departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_14_091441_create_receptionists_table.php

Source: `hms/database/migrations/2020_02_14_091441_create_receptionists_table.php`

- `create` `receptionists` at line 13

```php
$table->increments('id');
$table->unsignedBigInteger('user_id');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_14_093246_create_pharmacists_table.php

Source: `hms/database/migrations/2020_02_14_093246_create_pharmacists_table.php`

- `create` `pharmacists` at line 13

```php
$table->increments('id');
$table->unsignedBigInteger('user_id');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_17_053450_create_accountants_table.php

Source: `hms/database/migrations/2020_02_17_053450_create_accountants_table.php`

- `create` `accountants` at line 13

```php
$table->increments('id');
$table->unsignedBigInteger('user_id');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_17_080856_create_bed_types_table.php

Source: `hms/database/migrations/2020_02_17_080856_create_bed_types_table.php`

- `create` `bed_types` at line 13

```php
$table->increments('id');
$table->string('title', 160);
$table->text('description')->nullable();
$table->timestamps();
```

## 2020_02_17_092326_create_blood_banks_table.php

Source: `hms/database/migrations/2020_02_17_092326_create_blood_banks_table.php`

- `create` `blood_banks` at line 13

```php
$table->increments('id');
$table->string('blood_group');
$table->bigInteger('remained_bags');
$table->timestamps();
```

## 2020_02_17_105627_create_beds_table.php

Source: `hms/database/migrations/2020_02_17_105627_create_beds_table.php`

- `create` `beds` at line 13

```php
$table->increments('id');
$table->unsignedInteger('bed_type');
$table->string('bed_id');
$table->string('name');
$table->text('description')->nullable();
$table->integer('charge');
$table->boolean('is_available')->default(1);
$table->timestamps();
$table->foreign('bed_type')->references('id')->on('bed_types') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2020_02_17_110620_create_blood_donors_table.php

Source: `hms/database/migrations/2020_02_17_110620_create_blood_donors_table.php`

- `create` `blood_donors` at line 13

```php
$table->increments('id');
$table->string('name');
$table->integer('age');
$table->integer('gender');
$table->string('blood_group');
$table->dateTime('last_donate_date');
$table->timestamps();
```

## 2020_02_17_135716_create_permission_tables.php

Source: `hms/database/migrations/2020_02_17_135716_create_permission_tables.php`

- `create` `permissions` at line 17

```php
$table->bigIncrements('id');
$table->string('name');
$table->string('guard_name');
$table->timestamps();
```

- `create` `model_has_permissions` at line 24

```php
$table->unsignedBigInteger('permission_id');
$table->string('model_type');
$table->unsignedBigInteger('model_id');
$table->index(['model_id', 'model_type'], 'model_has_permissions_model_id_model_type_index');
$table->foreign('permission_id') ->references('id') ->on('permissions') ->onDelete('cascade');
$table->primary(['permission_id', 'model_id', 'model_type'], 'model_has_permissions_permission_model_type_primary');
```

- `create` `model_has_roles` at line 42

```php
$table->unsignedBigInteger('role_id');
$table->string('model_type');
$table->unsignedBigInteger('model_id');
$table->index(['model_id', 'model_type'], 'model_has_roles_model_id_model_type_index');
$table->foreign('role_id') ->references('id') ->on('departments') ->onDelete('cascade');
$table->primary(['role_id', 'model_id', 'model_type'], 'model_has_roles_role_model_type_primary');
```

- `create` `role_has_permissions` at line 58

```php
$table->unsignedBigInteger('permission_id');
$table->unsignedBigInteger('role_id');
$table->foreign('permission_id') ->references('id') ->on('permissions') ->onDelete('cascade');
$table->foreign('role_id') ->references('id') ->on('departments') ->onDelete('cascade');
$table->primary(['permission_id', 'role_id'], 'role_has_permissions_permission_id_role_id_primary');
```

## 2020_02_18_042327_create_notice_boards_table.php

Source: `hms/database/migrations/2020_02_18_042327_create_notice_boards_table.php`

- `create` `notice_boards` at line 13

```php
$table->increments('id');
$table->string('title');
$table->text('description')->nullable();
$table->timestamps();
```

## 2020_02_18_042442_create_document_types_table.php

Source: `hms/database/migrations/2020_02_18_042442_create_document_types_table.php`

- `create` `document_types` at line 13

```php
$table->increments('id');
$table->string('name', 160);
$table->timestamps();
```

## 2020_02_18_060222_create_patient_cases_table.php

Source: `hms/database/migrations/2020_02_18_060222_create_patient_cases_table.php`

- `create` `patient_cases` at line 13

```php
$table->increments('id');
$table->string('case_id', 160)->unique();
$table->unsignedInteger('patient_id');
$table->string('phone')->nullable();
$table->unsignedBigInteger('doctor_id');
$table->datetime('date');
$table->double('fee');
$table->boolean('status')->default(0);
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_18_060223_create_operation_reports_table.php

Source: `hms/database/migrations/2020_02_18_060223_create_operation_reports_table.php`

- `create` `operation_reports` at line 13

```php
$table->increments('id');
$table->unsignedInteger('patient_id');
$table->string('case_id');
$table->unsignedBigInteger('doctor_id');
$table->dateTime('date');
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('doctor_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_18_064953_create_bed_assigns_table.php

Source: `hms/database/migrations/2020_02_18_064953_create_bed_assigns_table.php`

- `create` `bed_assigns` at line 13

```php
$table->increments('id');
$table->unsignedInteger('bed_id');
$table->unsignedInteger('patient_id');
$table->string('case_id');
$table->date('assign_date');
$table->date('discharge_date')->nullable();
$table->text('description')->nullable();
$table->boolean('status')->default(0);
$table->timestamps();
$table->foreign('bed_id')->references('id')->on('beds') ->onUpdate('cascade') ->onDelete('cascade');
$table->foreign('patient_id')->references('id')->on('patients') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2020_02_18_092202_create_documents_table.php

Source: `hms/database/migrations/2020_02_18_092202_create_documents_table.php`

- `create` `documents` at line 13

```php
$table->increments('id');
$table->string('title');
$table->integer('document_type_id');
$table->integer('patient_id');
$table->unsignedBigInteger('uploaded_by');
$table->timestamps();
$table->foreign('uploaded_by')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_18_094758_create_birth_reports_table.php

Source: `hms/database/migrations/2020_02_18_094758_create_birth_reports_table.php`

- `create` `birth_reports` at line 13

```php
$table->increments('id');
$table->unsignedInteger('patient_id');
$table->string('case_id');
$table->unsignedBigInteger('doctor_id');
$table->dateTime('date');
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_18_111020_create_death_reports_table.php

Source: `hms/database/migrations/2020_02_18_111020_create_death_reports_table.php`

- `create` `death_reports` at line 13

```php
$table->increments('id');
$table->unsignedInteger('patient_id');
$table->string('case_id');
$table->unsignedBigInteger('doctor_id');
$table->dateTime('date');
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_19_080336_create_employee_payrolls_table.php

Source: `hms/database/migrations/2020_02_19_080336_create_employee_payrolls_table.php`

- `create` `employee_payrolls` at line 13

```php
$table->increments('id');
$table->bigInteger('sr_no');
$table->string('payroll_id');
$table->integer('type');
$table->integer('owner_id');
$table->string('owner_type');
$table->string('month');
$table->integer('year');
$table->double('net_salary');
$table->integer('status');
$table->double('basic_salary');
$table->double('allowance');
$table->double('deductions');
$table->timestamps();
```

## 2020_02_19_134502_create_settings_table.php

Source: `hms/database/migrations/2020_02_19_134502_create_settings_table.php`

- `create` `settings` at line 13

```php
$table->increments('id');
$table->string('key');
$table->string('value')->nullable();
$table->timestamps();
```

## 2020_02_21_090236_create_investigation_reports_table.php

Source: `hms/database/migrations/2020_02_21_090236_create_investigation_reports_table.php`

- `create` `investigation_reports` at line 13

```php
$table->increments('id');
$table->unsignedInteger('patient_id');
$table->dateTime('date');
$table->string('title');
$table->text('description')->nullable();
$table->unsignedBigInteger('doctor_id');
$table->integer('status');
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_21_095439_create_accounts_table.php

Source: `hms/database/migrations/2020_02_21_095439_create_accounts_table.php`

- `create` `accounts` at line 13

```php
$table->increments('id');
$table->string('name', 160);
$table->tinyInteger('type');
$table->text('description')->nullable();
$table->boolean('status')->default(0);
$table->timestamps();
```

## 2020_02_22_070658_create_payments_table.php

Source: `hms/database/migrations/2020_02_22_070658_create_payments_table.php`

- `create` `payments` at line 13

```php
$table->increments('id');
$table->date('payment_date');
$table->unsignedInteger('account_id');
$table->string('pay_to');
$table->double('amount');
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('account_id')->references('id')->on('accounts') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2020_02_22_090112_create_insurances_table.php

Source: `hms/database/migrations/2020_02_22_090112_create_insurances_table.php`

- `create` `insurances` at line 13

```php
$table->increments('id');
$table->string('name', 160);
$table->double('service_tax');
$table->double('discount')->nullable();
$table->text('remark')->nullable();
$table->string('insurance_no');
$table->string('insurance_code');
$table->double('hospital_rate');
$table->double('total');
$table->boolean('status');
$table->timestamps();
```

## 2020_02_22_091537_create_insurance_disease_table.php

Source: `hms/database/migrations/2020_02_22_091537_create_insurance_disease_table.php`

- `create` `insurance_diseases` at line 14

```php
$table->bigIncrements('id');
$table->unsignedInteger('insurance_id');
$table->string('disease_name');
$table->double('disease_charge');
$table->timestamps();
$table->foreign('insurance_id')->references('id')->on('insurances') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2020_02_24_055136_create_invoices_table.php

Source: `hms/database/migrations/2020_02_24_055136_create_invoices_table.php`

- `create` `invoices` at line 13

```php
$table->increments('id');
$table->string('invoice_id');
$table->unsignedInteger('patient_id');
$table->date('invoice_date');
$table->double('amount', 8, 2)->default(0);
$table->double('discount', 8, 2)->default(0);
$table->boolean('status')->default(0);
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2020_02_24_055518_create_schedules_table.php

Source: `hms/database/migrations/2020_02_24_055518_create_schedules_table.php`

- `create` `schedules` at line 13

```php
$table->increments('id');
$table->unsignedBigInteger('doctor_id');
$table->time('per_patient_time');
$table->integer('serial_visibility');
$table->timestamps();
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_02_24_055702_create_invoice_items_table.php

Source: `hms/database/migrations/2020_02_24_055702_create_invoice_items_table.php`

- `create` `invoice_items` at line 13

```php
$table->increments('id');
$table->unsignedInteger('account_id');
$table->unsignedInteger('invoice_id');
$table->string('description')->nullable();
$table->integer('quantity');
$table->double('price', 8, 2);
$table->double('total', 8, 2);
$table->timestamps();
$table->foreign('account_id')->references('id')->on('accounts') ->onUpdate('cascade') ->onDelete('cascade');
$table->foreign('invoice_id')->references('id')->on('invoices') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2020_02_25_105042_create_services_table.php

Source: `hms/database/migrations/2020_02_25_105042_create_services_table.php`

- `create` `services` at line 13

```php
$table->increments('id');
$table->string('name', 160);
$table->text('description')->nullable();
$table->integer('quantity');
$table->integer('rate');
$table->integer('status');
$table->timestamps();
```

## 2020_02_25_131030_create_packages_table.php

Source: `hms/database/migrations/2020_02_25_131030_create_packages_table.php`

- `create` `packages` at line 13

```php
$table->increments('id');
$table->string('name', 160);
$table->text('description')->nullable();
$table->double('discount');
$table->double('total_amount');
$table->timestamps();
```

## 2020_02_25_131108_create_package_services_table.php

Source: `hms/database/migrations/2020_02_25_131108_create_package_services_table.php`

- `create` `package_services` at line 13

```php
$table->bigIncrements('id');
$table->unsignedInteger('package_id');
$table->unsignedInteger('service_id');
$table->double('quantity');
$table->double('rate');
$table->double('amount');
$table->timestamps();
$table->foreign('package_id')->references('id')->on('packages') ->onUpdate('cascade') ->onDelete('cascade');
$table->foreign('service_id')->references('id')->on('services') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2020_02_27_120948_create_patient_admissions_table.php

Source: `hms/database/migrations/2020_02_27_120948_create_patient_admissions_table.php`

- `create` `patient_admissions` at line 13

```php
$table->increments('id');
$table->string('patient_admission_id', 160)->unique();
$table->unsignedInteger('patient_id');
$table->unsignedBigInteger('doctor_id');
$table->datetime('admission_date');
$table->datetime('discharge_date')->nullable();
$table->unsignedInteger('package_id')->nullable();
$table->unsignedInteger('insurance_id')->nullable();
$table->unsignedInteger('bed_id')->nullable();
$table->string('policy_no')->nullable();
$table->string('agent_name')->nullable();
$table->string('guardian_name')->nullable();
$table->string('guardian_relation')->nullable();
$table->string('guardian_contact')->nullable();
$table->string('guardian_address')->nullable();
$table->boolean('status')->nullable();
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('package_id')->references('id')->on('packages') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('insurance_id')->references('id')->on('insurances') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('bed_id')->references('id')->on('beds') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2020_02_28_031410_create_case_handlers_table.php

Source: `hms/database/migrations/2020_02_28_031410_create_case_handlers_table.php`

- `create` `case_handlers` at line 13

```php
$table->increments('id');
$table->unsignedBigInteger('user_id');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_03_02_043813_create_advanced_payments_table.php

Source: `hms/database/migrations/2020_03_02_043813_create_advanced_payments_table.php`

- `create` `advanced_payments` at line 13

```php
$table->increments('id');
$table->unsignedBigInteger('patient_id');
$table->string('receipt_no');
$table->double('amount');
$table->date('date');
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('users');
```

## 2020_03_02_065845_add_patient_admission_id_to_bills.php

Source: `hms/database/migrations/2020_03_02_065845_add_patient_admission_id_to_bills.php`

- `table` `bills` at line 14

```php
$table->string('patient_admission_id');
```

## 2020_03_03_062243_add_patient_id_to_bills.php

Source: `hms/database/migrations/2020_03_03_062243_add_patient_id_to_bills.php`

- `table` `bills` at line 14

```php
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_03_03_113334_create_schedule_day_table.php

Source: `hms/database/migrations/2020_03_03_113334_create_schedule_day_table.php`

- `create` `schedule_days` at line 14

```php
$table->bigIncrements('id');
$table->unsignedBigInteger('doctor_id');
$table->unsignedInteger('schedule_id');
$table->string('available_on');
$table->time('available_from');
$table->time('available_to');
$table->timestamps();
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('schedule_id')->references('id')->on('schedules') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_03_26_052336_create_ambulances_table.php

Source: `hms/database/migrations/2020_03_26_052336_create_ambulances_table.php`

- `create` `ambulances` at line 13

```php
$table->increments('id');
$table->string('vehicle_number', 160);
$table->string('vehicle_model');
$table->string('year_made');
$table->string('driver_name');
$table->string('driver_license');
$table->string('driver_contact');
$table->string('note')->nullable();
$table->boolean('is_available')->default(1);
$table->integer('vehicle_type')->default(1);
$table->timestamps();
```

## 2020_03_26_081157_create_mails_table.php

Source: `hms/database/migrations/2020_03_26_081157_create_mails_table.php`

- `create` `mails` at line 14

```php
$table->bigIncrements('id');
$table->string('to');
$table->string('subject');
$table->text('message');
$table->string('attachments')->nullable();
$table->unsignedBigInteger('user_id');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_03_27_061641_create_enquiries_table.php

Source: `hms/database/migrations/2020_03_27_061641_create_enquiries_table.php`

- `create` `enquiries` at line 14

```php
$table->increments('id');
$table->string('full_name');
$table->string('email');
$table->string('contact_no')->nullable();
$table->tinyInteger('type')->nullable();
$table->text('message');
$table->unsignedBigInteger('viewed_by')->nullable();
$table->boolean('status')->default(0);
$table->timestamps();
$table->foreign('viewed_by')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_03_27_063148_create_ambulance_calls_table.php

Source: `hms/database/migrations/2020_03_27_063148_create_ambulance_calls_table.php`

- `create` `ambulance_calls` at line 13

```php
$table->increments('id');
$table->unsignedInteger('patient_id');
$table->unsignedInteger('ambulance_id');
$table->string('driver_name');
$table->date('date');
$table->double('amount', 8, 2);
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('ambulance_id')->references('id')->on('ambulances') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_03_31_122219_create_prescriptions_table.php

Source: `hms/database/migrations/2020_03_31_122219_create_prescriptions_table.php`

- `create` `prescriptions` at line 13

```php
$table->increments('id');
$table->unsignedInteger('patient_id');
$table->unsignedBigInteger('doctor_id')->nullable();
$table->string('food_allergies', 100)->nullable();
$table->string('tendency_bleed', 100)->nullable();
$table->string('heart_disease', 100)->nullable();
$table->string('high_blood_pressure', 100)->nullable();
$table->string('diabetic', 100)->nullable();
$table->string('surgery', 100)->nullable();
$table->string('accident', 100)->nullable();
$table->string('others', 100)->nullable();
$table->string('medical_history', 100)->nullable();
$table->string('current_medication', 100)->nullable();
$table->string('female_pregnancy', 100)->nullable();
$table->string('breast_feeding', 100)->nullable();
$table->string('health_insurance', 100)->nullable();
$table->string('low_income', 100)->nullable();
$table->string('reference', 100)->nullable();
$table->boolean('status')->nullable();
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_04_11_052629_create_charge_categories_table.php

Source: `hms/database/migrations/2020_04_11_052629_create_charge_categories_table.php`

- `create` `charge_categories` at line 13

```php
$table->increments('id');
$table->string('name', 160);
$table->text('description')->nullable();
$table->integer('charge_type');
$table->timestamps();
```

## 2020_04_11_053929_create_pathology_categories_table.php

Source: `hms/database/migrations/2020_04_11_053929_create_pathology_categories_table.php`

- `create` `pathology_categories` at line 13

```php
$table->increments('id');
$table->string('name', 160)->unique();
$table->timestamps();
```

## 2020_04_11_070859_create_radiology_categories_table.php

Source: `hms/database/migrations/2020_04_11_070859_create_radiology_categories_table.php`

- `create` `radiology_categories` at line 13

```php
$table->increments('id');
$table->string('name', 160)->unique();
$table->timestamps();
```

## 2020_04_11_090903_create_charges_table.php

Source: `hms/database/migrations/2020_04_11_090903_create_charges_table.php`

- `create` `charges` at line 13

```php
$table->increments('id');
$table->integer('charge_type');
$table->unsignedInteger('charge_category_id');
$table->string('code', 160);
$table->string('standard_charge');
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('charge_category_id')->references('id') ->on('charge_categories') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2020_04_13_050643_create_radiology_tests_table.php

Source: `hms/database/migrations/2020_04_13_050643_create_radiology_tests_table.php`

- `create` `radiology_tests` at line 13

```php
$table->increments('id');
$table->string('test_name', 160);
$table->string('short_name');
$table->string('test_type');
$table->unsignedInteger('category_id');
$table->string('subcategory')->nullable();
$table->integer('report_days')->nullable();
$table->unsignedInteger('charge_category_id');
$table->integer('standard_charge');
$table->timestamps();
$table->foreign('category_id')->references('id')->on('radiology_categories') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('charge_category_id')->references('id')->on('charge_categories') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_04_14_093339_create_pathology_tests_table.php

Source: `hms/database/migrations/2020_04_14_093339_create_pathology_tests_table.php`

- `create` `pathology_tests` at line 13

```php
$table->increments('id');
$table->string('test_name', 160);
$table->string('short_name');
$table->string('test_type');
$table->unsignedInteger('category_id');
$table->integer('unit')->nullable();
$table->string('subcategory')->nullable();
$table->string('method')->nullable();
$table->integer('report_days')->nullable();
$table->unsignedInteger('charge_category_id');
$table->integer('standard_charge');
$table->timestamps();
$table->foreign('category_id')->references('id')->on('pathology_categories') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('charge_category_id')->references('id')->on('charge_categories') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_04_24_111205_create_doctor_opd_charge_table.php

Source: `hms/database/migrations/2020_04_24_111205_create_doctor_opd_charge_table.php`

- `create` `doctor_opd_charges` at line 14

```php
$table->id();
$table->unsignedBigInteger('doctor_id');
$table->double('standard_charge');
$table->timestamps();
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_04_28_094118_create_expenses_table.php

Source: `hms/database/migrations/2020_04_28_094118_create_expenses_table.php`

- `create` `expenses` at line 14

```php
$table->id();
$table->integer('expense_head');
$table->string('name');
$table->string('invoice_number')->nullable();
$table->dateTime('date');
$table->double('amount');
$table->text('description')->nullable();
$table->timestamps();
```

## 2020_05_01_055137_create_incomes_table.php

Source: `hms/database/migrations/2020_05_01_055137_create_incomes_table.php`

- `create` `incomes` at line 14

```php
$table->id();
$table->integer('income_head');
$table->string('name');
$table->string('invoice_number')->nullable();
$table->dateTime('date');
$table->double('amount');
$table->text('description')->nullable();
$table->timestamps();
```

## 2020_05_11_083050_add_notes_documents_table.php

Source: `hms/database/migrations/2020_05_11_083050_add_notes_documents_table.php`

- `table` `documents` at line 14

```php
$table->text('notes')->nullable()->after('uploaded_by');
```

## 2020_05_12_075825_create_sms_table.php

Source: `hms/database/migrations/2020_05_12_075825_create_sms_table.php`

- `create` `sms` at line 14

```php
$table->id();
$table->unsignedBigInteger('send_to');
$table->string('phone_number');
$table->text('message');
$table->unsignedBigInteger('send_by');
$table->timestamps();
$table->foreign('send_to')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('send_by')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_06_22_071531_add_index_to_accounts_table.php

Source: `hms/database/migrations/2020_06_22_071531_add_index_to_accounts_table.php`

- `table` `accounts` at line 14

```php
$table->index('name');
```

## 2020_06_22_071943_add_index_to_doctor_opd_charges_table.php

Source: `hms/database/migrations/2020_06_22_071943_add_index_to_doctor_opd_charges_table.php`

- `table` `doctor_opd_charges` at line 14

```php
$table->index('created_at');
```

## 2020_06_22_072921_add_index_to_bed_assigns_table.php

Source: `hms/database/migrations/2020_06_22_072921_add_index_to_bed_assigns_table.php`

- `table` `bed_assigns` at line 14

```php
$table->index(['created_at', 'assign_date']);
```

## 2020_06_22_073042_add_index_to_medicines_table.php

Source: `hms/database/migrations/2020_06_22_073042_add_index_to_medicines_table.php`

- `table` `medicines` at line 14

```php
$table->index('name', 160);
```

## 2020_06_22_073457_add_index_to_employee_payrolls_table.php

Source: `hms/database/migrations/2020_06_22_073457_add_index_to_employee_payrolls_table.php`

- `table` `employee_payrolls` at line 14

```php
$table->index(['id', 'sr_no']);
```

## 2020_06_22_074937_add_index_to_notice_boards_table.php

Source: `hms/database/migrations/2020_06_22_074937_add_index_to_notice_boards_table.php`

- `table` `notice_boards` at line 14

```php
$table->index(['created_at', 'id']);
```

## 2020_06_22_075222_add_index_to_blood_donors_table.php

Source: `hms/database/migrations/2020_06_22_075222_add_index_to_blood_donors_table.php`

- `table` `blood_donors` at line 14

```php
$table->index(['created_at', 'last_donate_date']);
```

## 2020_06_22_075359_add_index_to_packages_table.php

Source: `hms/database/migrations/2020_06_22_075359_add_index_to_packages_table.php`

- `table` `packages` at line 14

```php
$table->index(['created_at', 'name']);
```

## 2020_06_22_075506_add_index_to_bed_types_table.php

Source: `hms/database/migrations/2020_06_22_075506_add_index_to_bed_types_table.php`

- `table` `bed_types` at line 14

```php
$table->index('title');
```

## 2020_06_22_075725_add_index_to_services_table.php

Source: `hms/database/migrations/2020_06_22_075725_add_index_to_services_table.php`

- `table` `services` at line 14

```php
$table->index('name');
```

## 2020_06_22_080944_add_index_to_invoices_table.php

Source: `hms/database/migrations/2020_06_22_080944_add_index_to_invoices_table.php`

- `table` `invoices` at line 14

```php
$table->index('invoice_date');
```

## 2020_06_22_081601_add_index_to_payments_table.php

Source: `hms/database/migrations/2020_06_22_081601_add_index_to_payments_table.php`

- `table` `payments` at line 14

```php
$table->index('payment_date');
```

## 2020_06_22_081802_add_index_to_advanced_payments_table.php

Source: `hms/database/migrations/2020_06_22_081802_add_index_to_advanced_payments_table.php`

- `table` `advanced_payments` at line 14

```php
$table->index('amount');
```

## 2020_06_22_081909_add_index_to_bills_table.php

Source: `hms/database/migrations/2020_06_22_081909_add_index_to_bills_table.php`

- `table` `bills` at line 14

```php
$table->index('bill_date');
```

## 2020_06_22_082548_add_index_to_beds_table.php

Source: `hms/database/migrations/2020_06_22_082548_add_index_to_beds_table.php`

- `table` `beds` at line 14

```php
$table->index('is_available');
```

## 2020_06_22_082942_add_index_to_blood_banks_table.php

Source: `hms/database/migrations/2020_06_22_082942_add_index_to_blood_banks_table.php`

- `table` `blood_banks` at line 14

```php
$table->index('remained_bags');
```

## 2020_06_22_083511_add_index_to_users_table.php

Source: `hms/database/migrations/2020_06_22_083511_add_index_to_users_table.php`

- `table` `users` at line 14

```php
$table->index('first_name');
```

## 2020_06_22_084750_add_index_to_patient_cases_table.php

Source: `hms/database/migrations/2020_06_22_084750_add_index_to_patient_cases_table.php`

- `table` `patient_cases` at line 14

```php
$table->index('date');
```

## 2020_06_22_084912_add_index_to_patient_admissions_table.php

Source: `hms/database/migrations/2020_06_22_084912_add_index_to_patient_admissions_table.php`

- `table` `patient_admissions` at line 14

```php
$table->index('admission_date');
```

## 2020_06_22_085036_add_index_to_document_types_table.php

Source: `hms/database/migrations/2020_06_22_085036_add_index_to_document_types_table.php`

- `table` `document_types` at line 14

```php
$table->index('name');
```

## 2020_06_22_085128_add_index_to_insurances_table.php

Source: `hms/database/migrations/2020_06_22_085128_add_index_to_insurances_table.php`

- `table` `insurances` at line 14

```php
$table->index('name');
```

## 2020_06_22_085317_add_index_to_ambulances_table.php

Source: `hms/database/migrations/2020_06_22_085317_add_index_to_ambulances_table.php`

- `table` `ambulances` at line 14

```php
$table->index('vehicle_number');
```

## 2020_06_22_090509_add_index_to_ambulance_calls_table.php

Source: `hms/database/migrations/2020_06_22_090509_add_index_to_ambulance_calls_table.php`

- `table` `ambulance_calls` at line 14

```php
$table->index('date');
```

## 2020_06_22_091253_add_index_to_doctor_departments_table.php

Source: `hms/database/migrations/2020_06_22_091253_add_index_to_doctor_departments_table.php`

- `table` `doctor_departments` at line 14

```php
$table->index('title');
```

## 2020_06_22_091455_add_index_to_appointments_table.php

Source: `hms/database/migrations/2020_06_22_091455_add_index_to_appointments_table.php`

- `table` `appointments` at line 14

```php
$table->index('opd_date');
```

## 2020_06_22_091617_add_index_to_birth_reports_table.php

Source: `hms/database/migrations/2020_06_22_091617_add_index_to_birth_reports_table.php`

- `table` `birth_reports` at line 14

```php
$table->index('date');
```

## 2020_06_22_091632_add_index_to_death_reports_table.php

Source: `hms/database/migrations/2020_06_22_091632_add_index_to_death_reports_table.php`

- `table` `death_reports` at line 14

```php
$table->index('date');
```

## 2020_06_22_091651_add_index_to_investigation_reports_table.php

Source: `hms/database/migrations/2020_06_22_091651_add_index_to_investigation_reports_table.php`

- `table` `investigation_reports` at line 14

```php
$table->index('date');
```

## 2020_06_22_091828_add_index_to_operation_reports_table.php

Source: `hms/database/migrations/2020_06_22_091828_add_index_to_operation_reports_table.php`

- `table` `operation_reports` at line 14

```php
$table->index('date');
```

## 2020_06_22_092018_add_index_to_categories_table.php

Source: `hms/database/migrations/2020_06_22_092018_add_index_to_categories_table.php`

- `table` `categories` at line 14

```php
$table->index('name');
```

## 2020_06_22_092149_add_index_to_brands_table.php

Source: `hms/database/migrations/2020_06_22_092149_add_index_to_brands_table.php`

- `table` `brands` at line 14

```php
$table->index('name');
```

## 2020_06_22_092324_add_index_to_pathology_tests_table.php

Source: `hms/database/migrations/2020_06_22_092324_add_index_to_pathology_tests_table.php`

- `table` `pathology_tests` at line 14

```php
$table->index('test_name');
```

## 2020_06_22_092338_add_index_to_pathology_categories_table.php

Source: `hms/database/migrations/2020_06_22_092338_add_index_to_pathology_categories_table.php`

- `table` `pathology_categories` at line 14

```php
$table->index('name');
```

## 2020_06_22_092347_add_index_to_radiology_tests_table.php

Source: `hms/database/migrations/2020_06_22_092347_add_index_to_radiology_tests_table.php`

- `table` `radiology_tests` at line 14

```php
$table->index('test_name');
```

## 2020_06_22_092357_add_index_to_radiology_categories_table.php

Source: `hms/database/migrations/2020_06_22_092357_add_index_to_radiology_categories_table.php`

- `table` `radiology_categories` at line 14

```php
$table->index('name');
```

## 2020_06_22_092651_add_index_to_expenses_table.php

Source: `hms/database/migrations/2020_06_22_092651_add_index_to_expenses_table.php`

- `table` `expenses` at line 14

```php
$table->index('date');
```

## 2020_06_22_092702_add_index_to_incomes_table.php

Source: `hms/database/migrations/2020_06_22_092702_add_index_to_incomes_table.php`

- `table` `incomes` at line 14

```php
$table->index('date');
```

## 2020_06_22_092855_add_index_to_charges_table.php

Source: `hms/database/migrations/2020_06_22_092855_add_index_to_charges_table.php`

- `table` `charges` at line 14

```php
$table->index('code');
```

## 2020_06_22_092905_add_index_to_charge_categories_table.php

Source: `hms/database/migrations/2020_06_22_092905_add_index_to_charge_categories_table.php`

- `table` `charge_categories` at line 14

```php
$table->index('name');
```

## 2020_06_22_093234_add_index_to_enquiries_table.php

Source: `hms/database/migrations/2020_06_22_093234_add_index_to_enquiries_table.php`

- `table` `enquiries` at line 14

```php
$table->index('created_at');
```

## 2020_06_24_044648_create_diagnosis_categories_table.php

Source: `hms/database/migrations/2020_06_24_044648_create_diagnosis_categories_table.php`

- `create` `diagnosis_categories` at line 14

```php
$table->bigIncrements('id');
$table->string('name', 160);
$table->text('description')->nullable();
$table->timestamps();
$table->index('name');
```

## 2020_06_25_080242_create_patient_diagnosis_tests_table.php

Source: `hms/database/migrations/2020_06_25_080242_create_patient_diagnosis_tests_table.php`

- `create` `patient_diagnosis_tests` at line 14

```php
$table->bigIncrements('id');
$table->unsignedInteger('patient_id');
$table->unsignedBigInteger('doctor_id');
$table->unsignedBigInteger('category_id');
$table->string('report_number');
$table->index('created_at');
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('category_id')->references('id')->on('diagnosis_categories') ->onDelete('cascade') ->onUpdate('cascade');
$table->timestamps();
```

## 2020_06_26_054352_create_patient_diagnosis_properties_table.php

Source: `hms/database/migrations/2020_06_26_054352_create_patient_diagnosis_properties_table.php`

- `create` `patient_diagnosis_properties` at line 14

```php
$table->increments('id');
$table->unsignedBigInteger('patient_diagnosis_id');
$table->string('property_name');
$table->string('property_value');
$table->timestamps();
$table->index('created_at');
$table->foreign('patient_diagnosis_id')->references('id')->on('patient_diagnosis_tests') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_07_15_044653_remove_serial_visibility_from_schedules_table.php

Source: `hms/database/migrations/2020_07_15_044653_remove_serial_visibility_from_schedules_table.php`

- `table` `schedules` at line 14

```php
$table->dropColumn('serial_visibility');
```

## 2020_07_15_121336_change_ambulances_table_column.php

Source: `hms/database/migrations/2020_07_15_121336_change_ambulances_table_column.php`

- `table` `ambulances` at line 14

```php
$table->text('note')->nullable()->change();
```

## 2020_07_22_052934_change_bed_assigns_table_column.php

Source: `hms/database/migrations/2020_07_22_052934_change_bed_assigns_table_column.php`

- `table` `bed_assigns` at line 14

```php
$table->datetime('assign_date')->change();
$table->datetime('discharge_date')->nullable()->change();
```

## 2020_07_29_095430_change_invoice_items_table_column.php

Source: `hms/database/migrations/2020_07_29_095430_change_invoice_items_table_column.php`

- `table` `invoice_items` at line 14

```php
$table->text('description')->change();
```

## 2020_08_26_081235_create_item_categories_table.php

Source: `hms/database/migrations/2020_08_26_081235_create_item_categories_table.php`

- `create` `item_categories` at line 14

```php
$table->increments('id');
$table->string('name', 160)->unique();
$table->timestamps();
```

## 2020_08_26_101134_create_items_table.php

Source: `hms/database/migrations/2020_08_26_101134_create_items_table.php`

- `create` `items` at line 14

```php
$table->increments('id');
$table->string('name');
$table->unsignedInteger('item_category_id');
$table->string('unit');
$table->text('description')->nullable();
$table->integer('available_quantity')->default(0);
$table->timestamps();
$table->foreign('item_category_id')->references('id')->on('item_categories') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_08_26_125032_create_item_stocks_table.php

Source: `hms/database/migrations/2020_08_26_125032_create_item_stocks_table.php`

- `create` `item_stocks` at line 14

```php
$table->increments('id');
$table->unsignedInteger('item_category_id');
$table->unsignedInteger('item_id');
$table->string('supplier_name')->nullable();
$table->string('store_name')->nullable();
$table->integer('quantity');
$table->double('purchase_price');
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('item_category_id')->references('id')->on('item_categories') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('item_id')->references('id')->on('items') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_08_27_141547_create_issued_items_table.php

Source: `hms/database/migrations/2020_08_27_141547_create_issued_items_table.php`

- `create` `issued_items` at line 14

```php
$table->increments('id');
$table->unsignedBigInteger('department_id');
$table->unsignedBigInteger('user_id');
$table->string('issued_by');
$table->date('issued_date');
$table->date('return_date')->nullable();
$table->unsignedInteger('item_category_id');
$table->unsignedInteger('item_id');
$table->integer('quantity');
$table->text('description')->nullable();
$table->boolean('status')->nullable()->default(false);
$table->timestamps();
$table->foreign('department_id')->references('id')->on('departments') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('item_category_id')->references('id')->on('item_categories') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('item_id')->references('id')->on('items') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_08_064222_create_ipd_patient_departments_table.php

Source: `hms/database/migrations/2020_09_08_064222_create_ipd_patient_departments_table.php`

- `create` `ipd_patient_departments` at line 13

```php
$table->increments('id');
$table->unsignedInteger('patient_id');
$table->string('ipd_number', 160)->unique();
$table->string('height')->nullable();
$table->string('weight')->nullable();
$table->string('bp')->nullable();
$table->text('symptoms')->nullable();
$table->text('notes')->nullable();
$table->datetime('admission_date');
$table->unsignedInteger('case_id');
$table->boolean('is_old_patient')->nullable()->default(false);
$table->unsignedBigInteger('doctor_id')->nullable();
$table->unsignedInteger('bed_type_id')->nullable();
$table->unsignedInteger('bed_id');
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('case_id')->references('id')->on('patient_cases') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('bed_type_id')->references('id')->on('bed_types') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('bed_id')->references('id')->on('beds') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_08_114627_create_ipd_diagnoses_table.php

Source: `hms/database/migrations/2020_09_08_114627_create_ipd_diagnoses_table.php`

- `create` `ipd_diagnoses` at line 13

```php
$table->increments('id');
$table->unsignedInteger('ipd_patient_department_id');
$table->string('report_type');
$table->datetime('report_date');
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('ipd_patient_department_id')->references('id')->on('ipd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_09_065624_create_ipd_consultant_registers_table.php

Source: `hms/database/migrations/2020_09_09_065624_create_ipd_consultant_registers_table.php`

- `create` `ipd_consultant_registers` at line 13

```php
$table->increments('id');
$table->unsignedInteger('ipd_patient_department_id');
$table->dateTime('applied_date');
$table->unsignedBigInteger('doctor_id');
$table->text('instruction');
$table->date('instruction_date');
$table->timestamps();
$table->foreign('ipd_patient_department_id')->references('id')->on('ipd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_09_135505_create_ipd_charges_table.php

Source: `hms/database/migrations/2020_09_09_135505_create_ipd_charges_table.php`

- `create` `ipd_charges` at line 13

```php
$table->increments('id');
$table->unsignedInteger('ipd_patient_department_id');
$table->date('date');
$table->integer('charge_type_id');
$table->unsignedInteger('charge_category_id');
$table->unsignedInteger('charge_id');
$table->integer('standard_charge')->nullable();
$table->integer('applied_charge');
$table->timestamps();
$table->foreign('ipd_patient_department_id')->references('id')->on('ipd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('charge_category_id')->references('id')->on('charge_categories') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('charge_id')->references('id')->on('charges') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_10_112306_create_ipd_prescriptions_table.php

Source: `hms/database/migrations/2020_09_10_112306_create_ipd_prescriptions_table.php`

- `create` `ipd_prescriptions` at line 13

```php
$table->increments('id');
$table->unsignedInteger('ipd_patient_department_id');
$table->text('header_note')->nullable();
$table->text('footer_note')->nullable();
$table->timestamps();
$table->foreign('ipd_patient_department_id')->references('id')->on('ipd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_10_114203_create_ipd_prescription_items_table.php

Source: `hms/database/migrations/2020_09_10_114203_create_ipd_prescription_items_table.php`

- `create` `ipd_prescription_items` at line 13

```php
$table->increments('id');
$table->unsignedInteger('ipd_prescription_id');
$table->unsignedInteger('category_id');
$table->unsignedInteger('medicine_id');
$table->string('dosage');
$table->text('instruction');
$table->timestamps();
$table->foreign('ipd_prescription_id')->references('id')->on('ipd_prescriptions') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('category_id')->references('id')->on('categories') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('medicine_id')->references('id')->on('medicines') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_11_045308_create_modules_table.php

Source: `hms/database/migrations/2020_09_11_045308_create_modules_table.php`

- `create` `modules` at line 14

```php
$table->id();
$table->string('name');
$table->boolean('is_active')->default(1);
$table->timestamps();
```

## 2020_09_12_050715_create_ipd_payments_table.php

Source: `hms/database/migrations/2020_09_12_050715_create_ipd_payments_table.php`

- `create` `ipd_payments` at line 14

```php
$table->increments('id');
$table->unsignedInteger('ipd_patient_department_id');
$table->integer('amount');
$table->date('date');
$table->tinyInteger('payment_mode');
$table->text('notes')->nullable();
$table->integer('transaction_id')->nullable();
$table->timestamps();
$table->foreign('ipd_patient_department_id')->references('id')->on('ipd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_12_071821_create_ipd_timelines_table.php

Source: `hms/database/migrations/2020_09_12_071821_create_ipd_timelines_table.php`

- `create` `ipd_timelines` at line 13

```php
$table->increments('id');
$table->unsignedInteger('ipd_patient_department_id');
$table->string('title');
$table->date('date');
$table->text('description')->nullable();
$table->boolean('visible_to_person')->default(true);
$table->timestamps();
$table->foreign('ipd_patient_department_id')->references('id')->on('ipd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_12_103003_create_ipd_bills_table.php

Source: `hms/database/migrations/2020_09_12_103003_create_ipd_bills_table.php`

- `create` `ipd_bills` at line 14

```php
$table->id();
$table->unsignedInteger('ipd_patient_department_id');
$table->integer('total_charges');
$table->integer('total_payments');
$table->integer('gross_total');
$table->integer('discount_in_percentage');
$table->integer('tax_in_percentage');
$table->integer('other_charges');
$table->integer('net_payable_amount');
$table->timestamps();
$table->foreign('ipd_patient_department_id')->references('id')->on('ipd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_14_083759_create_opd_patient_departments_table.php

Source: `hms/database/migrations/2020_09_14_083759_create_opd_patient_departments_table.php`

- `create` `opd_patient_departments` at line 14

```php
$table->increments('id');
$table->unsignedInteger('patient_id');
$table->string('opd_number', 160)->unique();
$table->string('height')->nullable();
$table->string('weight')->nullable();
$table->string('bp')->nullable();
$table->text('symptoms')->nullable();
$table->text('notes')->nullable();
$table->datetime('appointment_date');
$table->unsignedInteger('case_id')->nullable();
$table->boolean('is_old_patient')->nullable()->default(false);
$table->unsignedBigInteger('doctor_id')->nullable();
$table->double('standard_charge');
$table->tinyInteger('payment_mode');
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('case_id')->references('id')->on('patient_cases') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_14_144731_add_ipd_patient_department_id_to_bed_assigns_table.php

Source: `hms/database/migrations/2020_09_14_144731_add_ipd_patient_department_id_to_bed_assigns_table.php`

- `table` `bed_assigns` at line 13

```php
$table->unsignedInteger('ipd_patient_department_id')->nullable()->after('bed_id');
$table->foreign('ipd_patient_department_id')->references('id')->on('ipd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_15_064044_create_transactions_table.php

Source: `hms/database/migrations/2020_09_15_064044_create_transactions_table.php`

- `create` `transactions` at line 14

```php
$table->id();
$table->string('stripe_transaction_id');
$table->integer('amount');
$table->integer('user_id');
$table->string('status');
$table->text('meta')->nullable();
$table->timestamps();
```

## 2020_09_16_103204_create_opd_diagnoses_table.php

Source: `hms/database/migrations/2020_09_16_103204_create_opd_diagnoses_table.php`

- `create` `opd_diagnoses` at line 14

```php
$table->increments('id');
$table->unsignedInteger('opd_patient_department_id');
$table->string('report_type');
$table->datetime('report_date');
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('opd_patient_department_id')->references('id')->on('opd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_16_114031_create_opd_timelines_table.php

Source: `hms/database/migrations/2020_09_16_114031_create_opd_timelines_table.php`

- `create` `opd_timelines` at line 14

```php
$table->increments('id');
$table->unsignedInteger('opd_patient_department_id');
$table->string('title');
$table->date('date');
$table->text('description')->nullable();
$table->boolean('visible_to_person')->default(true);
$table->timestamps();
$table->foreign('opd_patient_department_id')->references('id')->on('opd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_09_23_045100_change_patient_diagnosis_properties_table.php

Source: `hms/database/migrations/2020_09_23_045100_change_patient_diagnosis_properties_table.php`

- `table` `patient_diagnosis_properties` at line 13

```php
$table->string('property_name')->nullable()->change();
$table->string('property_value')->nullable()->change();
```

## 2020_09_24_053229_add_ipd_bill_column_in_ipd_patient_departments_table.php

Source: `hms/database/migrations/2020_09_24_053229_add_ipd_bill_column_in_ipd_patient_departments_table.php`

- `table` `ipd_patient_departments` at line 14

```php
$table->boolean('bill_status')->default(false);
```

## 2020_10_09_085838_create_call_logs_table.php

Source: `hms/database/migrations/2020_10_09_085838_create_call_logs_table.php`

- `create` `call_logs` at line 14

```php
$table->id();
$table->string('name');
$table->string('phone')->nullable();
$table->date('date')->nullable();
$table->date('follow_up_date')->nullable();
$table->text('note')->nullable();
$table->integer('call_type');
$table->timestamps();
```

## 2020_10_12_125133_create_visitors_table.php

Source: `hms/database/migrations/2020_10_12_125133_create_visitors_table.php`

- `create` `visitors` at line 14

```php
$table->id();
$table->integer('purpose');
$table->string('name');
$table->string('phone')->nullable();
$table->string('id_card')->nullable();
$table->string('no_of_person')->nullable();
$table->date('date')->nullable();
$table->time('in_time')->nullable();
$table->time('out_time')->nullable();
$table->text('note')->nullable();
$table->timestamps();
```

## 2020_10_14_044134_create_postals_table.php

Source: `hms/database/migrations/2020_10_14_044134_create_postals_table.php`

- `create` `postals` at line 14

```php
$table->id();
$table->string('from_title')->nullable();
$table->string('to_title')->nullable();
$table->string('reference_no')->nullable();
$table->date('date')->nullable();
$table->text('address')->nullable();
$table->integer('type')->nullable();
$table->timestamps();
```

## 2020_10_30_043500_add_route_in_modules_table.php

Source: `hms/database/migrations/2020_10_30_043500_add_route_in_modules_table.php`

- `table` `modules` at line 14

```php
$table->string('route')->nullable()->after('is_active');
```

## 2020_10_31_062448_add_complete_in_appointments_table.php

Source: `hms/database/migrations/2020_10_31_062448_add_complete_in_appointments_table.php`

- `table` `appointments` at line 14

```php
$table->boolean('is_completed')->after('problem')->default(false);
```

## 2020_11_02_050736_create_testimonials_table.php

Source: `hms/database/migrations/2020_11_02_050736_create_testimonials_table.php`

- `create` `testimonials` at line 14

```php
$table->id();
$table->string('name');
$table->text('description');
$table->timestamps();
```

## 2020_11_07_121633_add_region_code_in_sms_table.php

Source: `hms/database/migrations/2020_11_07_121633_add_region_code_in_sms_table.php`

- `table` `sms` at line 16

```php
$table->unsignedBigInteger('send_to')->nullable()->change();
$table->string('region_code')->after('send_to')->nullable();
```

## 2020_11_19_093810_create_blood_donations_table.php

Source: `hms/database/migrations/2020_11_19_093810_create_blood_donations_table.php`

- `create` `blood_donations` at line 14

```php
$table->increments('id');
$table->unsignedInteger('blood_donor_id');
$table->integer('bags')->default(1);
$table->timestamps();
$table->foreign('blood_donor_id')->references('id')->on('blood_donors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_11_20_113830_create_blood_issues_table.php

Source: `hms/database/migrations/2020_11_20_113830_create_blood_issues_table.php`

- `create` `blood_issues` at line 14

```php
$table->bigIncrements('id');
$table->dateTime('issue_date');
$table->unsignedBigInteger('doctor_id');
$table->unsignedInteger('donor_id');
$table->unsignedInteger('patient_id');
$table->string('amount')->nullable();
$table->text('remarks')->nullable();
$table->timestamps();
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('donor_id')->references('id')->on('blood_donors') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_11_24_131253_create_notifications_table.php

Source: `hms/database/migrations/2020_11_24_131253_create_notifications_table.php`

- `create` `notifications` at line 14

```php
$table->increments('id');
$table->integer('type');
$table->integer('notification_for');
$table->unsignedBigInteger('user_id');
$table->string('title');
$table->text('text')->nullable();
$table->text('meta')->nullable();
$table->timestamp('read_at')->nullable();
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2020_12_28_131351_create_live_consultations_table.php

Source: `hms/database/migrations/2020_12_28_131351_create_live_consultations_table.php`

- `create` `live_consultations` at line 14

```php
$table->id();
$table->unsignedBigInteger('doctor_id');
$table->unsignedInteger('patient_id');
$table->string('consultation_title');
$table->dateTime('consultation_date');
$table->boolean('host_video');
$table->boolean('participant_video');
$table->string('consultation_duration_minutes');
$table->string('type');
$table->string('type_number');
$table->string('created_by');
$table->integer('status');
$table->text('description')->nullable();
$table->string('meeting_id');
$table->text('meta')->nullable();
$table->string('time_zone')->default(null);
$table->string('password');
$table->timestamps();
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2020_12_31_062506_create_live_meetings_table.php

Source: `hms/database/migrations/2020_12_31_062506_create_live_meetings_table.php`

- `create` `live_meetings` at line 14

```php
$table->bigIncrements('id');
$table->string('consultation_title');
$table->dateTime('consultation_date');
$table->string('consultation_duration_minutes');
$table->boolean('host_video');
$table->boolean('participant_video');
$table->text('description')->nullable();
$table->string('created_by');
$table->text('meta')->nullable();
$table->string('time_zone')->default(null);
$table->string('password');
$table->integer('status');
$table->timestamps();
```

## 2020_12_31_091242_create_live_meetings_candidates_table.php

Source: `hms/database/migrations/2020_12_31_091242_create_live_meetings_candidates_table.php`

- `create` `live_meetings_candidates` at line 14

```php
$table->bigIncrements('id');
$table->unsignedBigInteger('user_id');
$table->unsignedBigInteger('live_meeting_id');
$table->timestamps();
```

## 2021_01_05_100425_create_user_zoom_credential_table.php

Source: `hms/database/migrations/2021_01_05_100425_create_user_zoom_credential_table.php`

- `create` `user_zoom_credential` at line 14

```php
$table->id();
$table->unsignedBigInteger('user_id');
$table->string('zoom_api_key');
$table->string('zoom_api_secret');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2021_01_06_105407_add_metting_id_to_live_meetings_table.php

Source: `hms/database/migrations/2021_01_06_105407_add_metting_id_to_live_meetings_table.php`

- `table` `live_meetings` at line 14

```php
$table->string('meeting_id')->after('meta');
```

## 2021_02_23_065200_create_vaccinations_table.php

Source: `hms/database/migrations/2021_02_23_065200_create_vaccinations_table.php`

- `create` `vaccinations` at line 14

```php
$table->increments('id')->index();
$table->string('name');
$table->string('manufactured_by');
$table->string('brand');
$table->timestamps();
```

## 2021_02_23_065252_create_vaccinated_patients_table.php

Source: `hms/database/migrations/2021_02_23_065252_create_vaccinated_patients_table.php`

- `create` `vaccinated_patients` at line 14

```php
$table->increments('id')->index();
$table->unsignedInteger('patient_id')->index();
$table->unsignedInteger('vaccination_id')->index();
$table->string('vaccination_serial_number')->nullable();
$table->string('dose_number');
$table->dateTime('dose_given_date');
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('vaccination_id')->references('id')->on('vaccinations') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2021_04_05_085646_create_front_settings_table.php

Source: `hms/database/migrations/2021_04_05_085646_create_front_settings_table.php`

- `create` `front_settings` at line 14

```php
$table->id();
$table->string('key');
$table->text('value');
$table->text('type');
$table->timestamps();
```

## 2021_05_10_000000_add_uuid_to_failed_jobs_table.php

Source: `hms/database/migrations/2021_05_10_000000_add_uuid_to_failed_jobs_table.php`

- `table` `failed_jobs` at line 16

```php
$table->string('uuid')->after('id')->nullable()->unique();
```

## 2021_05_29_103036_add_conversions_disk_column_in_media_table.php

Source: `hms/database/migrations/2021_05_29_103036_add_conversions_disk_column_in_media_table.php`

- `table` `media` at line 14

```php
$table->string('conversions_disk', 160)->nullable();
$table->uuid('uuid')->nullable()->unique();
$table->text('generated_conversions');
$table->text('manipulations')->change();
```

- `table` `media` at line 20

```php
$table->text('manipulations')->change();
```

- `table` `media` at line 24

```php
$table->text('manipulations')->change();
```

- `table` `media` at line 30

```php
$table->text('custom_properties')->change();
```

- `table` `media` at line 34

```php
$table->text('custom_properties')->change();
```

- `table` `media` at line 40

```php
$table->text('responsive_images')->change();
```

- `table` `media` at line 44

```php
$table->text('responsive_images')->change();
```

## 2021_06_07_104022_change_patient_foreign_key_type_in_appointments_table.php

Source: `hms/database/migrations/2021_06_07_104022_change_patient_foreign_key_type_in_appointments_table.php`

- `table` `appointments` at line 14

```php
$table->dropForeign('appointments_patient_id_foreign');
```

- `table` `appointments` at line 18

```php
$table->unsignedInteger('patient_id')->change();
```

- `table` `appointments` at line 22

```php
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2021_06_08_073918_change_department_foreign_key_in_appointments_table.php

Source: `hms/database/migrations/2021_06_08_073918_change_department_foreign_key_in_appointments_table.php`

- `table` `appointments` at line 14

```php
$table->dropForeign('appointments_department_id_foreign');
```

- `table` `appointments` at line 18

```php
$table->unsignedBigInteger('department_id')->change();
```

- `table` `appointments` at line 22

```php
$table->foreign('department_id')->references('id')->on('doctor_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2021_06_21_082754_update_amount_datatype_in_bills_table.php

Source: `hms/database/migrations/2021_06_21_082754_update_amount_datatype_in_bills_table.php`

- `table` `bills` at line 17

```php
$table->decimal('amount', 16, 2)->nullable()->change();
```

## 2021_06_21_082845_update_amount_datatype_in_bill_items_table.php

Source: `hms/database/migrations/2021_06_21_082845_update_amount_datatype_in_bill_items_table.php`

- `table` `bill_items` at line 17

```php
$table->decimal('amount', 16, 2)->change();
```

## 2021_11_11_061443_create_front_services_table.php

Source: `hms/database/migrations/2021_11_11_061443_create_front_services_table.php`

- `create` `front_services` at line 14

```php
$table->id();
$table->string('name');
$table->text('short_description');
$table->timestamps();
```

## 2021_11_12_100750_create_hospital_schedules_table.php

Source: `hms/database/migrations/2021_11_12_100750_create_hospital_schedules_table.php`

- `create` `hospital_schedules` at line 14

```php
$table->id();
$table->string('day_of_week');
$table->string('start_time');
$table->string('end_time');
$table->timestamps();
```

## 2021_11_12_105805_add_social_details_in_users_table.php

Source: `hms/database/migrations/2021_11_12_105805_add_social_details_in_users_table.php`

- `table` `users` at line 14

```php
$table->string('facebook_url', 100)->after('remember_token')->nullable();
$table->string('twitter_url', 100)->after('facebook_url')->nullable();
$table->string('instagram_url', 100)->after('twitter_url')->nullable();
$table->string('linkedIn_url', 100)->after('instagram_url')->nullable();
```

## 2022_02_18_101938_add_darkmode_to_users_table.php

Source: `hms/database/migrations/2022_02_18_101938_add_darkmode_to_users_table.php`

- `table` `users` at line 14

```php
$table->string('thememode', 100)->default(\App\Models\User::THEME_LIGHT_MODE);
```

## 2022_04_09_064645_change_doctor_foreign_in_operation_reports_table.php

Source: `hms/database/migrations/2022_04_09_064645_change_doctor_foreign_in_operation_reports_table.php`

- `table` `operation_reports` at line 14

```php
$table->dropForeign('operation_reports_doctor_id_foreign');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2022_05_16_104947_add_default_length_in_table.php

Source: `hms/database/migrations/2022_05_16_104947_add_default_length_in_table.php`

- `table` `media` at line 14

```php
$table->string('model_type', 160)->change();
```

- `table` `model_has_roles` at line 17

```php
$table->string('model_type', 160)->change();
```

- `table` `model_has_permissions` at line 20

```php
$table->string('model_type', 160)->unique()->change();
```

- `table` `users` at line 23

```php
$table->string('stripe_id', 100)->change();
$table->string('card_brand', 100);
```

## 2022_07_29_200345_add_prescription_fields.php

Source: `hms/database/migrations/2022_07_29_200345_add_prescription_fields.php`

- `table` `prescriptions` at line 14

```php
$table->string('plus_rate', 100)->nullable();
$table->string('temperature', 100)->nullable();
$table->string('problem_description', 100)->nullable();
$table->string('test', 100)->nullable();
$table->string('advice', 100)->nullable();
$table->string('next_visit_qty', 100)->nullable();
$table->string('next_visit_time', 100)->nullable();
```

## 2022_08_01_204917_create_prescriptions_medicines_table.php

Source: `hms/database/migrations/2022_08_01_204917_create_prescriptions_medicines_table.php`

- `create` `prescriptions_medicines` at line 14

```php
$table->id();
$table->integer('prescription_id')->unsigned();
$table->integer('medicine')->unsigned();
$table->string('dosage')->nullable();
$table->string('day')->nullable();
$table->string('time')->nullable();
$table->string('comment')->nullable();
$table->foreign('prescription_id')->on('prescriptions')->references('id') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('medicine')->on('medicines')->references('id') ->onDelete('cascade') ->onUpdate('cascade');
$table->timestamps();
```

## 2022_08_26_225704_change_charges_standard_charge_column.php

Source: `hms/database/migrations/2022_08_26_225704_change_charges_standard_charge_column.php`

- `table` `charges` at line 14

```php
$table->bigInteger('standard_charge')->change();
```

## 2022_08_30_011825_change_item_unit_column.php

Source: `hms/database/migrations/2022_08_30_011825_change_item_unit_column.php`

- `table` `items` at line 14

```php
$table->bigInteger('unit')->change();
```

## 2022_09_06_202047_change_amount_at_blood_issue.php

Source: `hms/database/migrations/2022_09_06_202047_change_amount_at_blood_issue.php`

- `table` `blood_issues` at line 14

```php
$table->bigInteger('amount')->change();
```

## 2022_09_07_184901_change_dose_number_column.php

Source: `hms/database/migrations/2022_09_07_184901_change_dose_number_column.php`

- `table` `vaccinated_patients` at line 14

```php
$table->bigInteger('dose_number')->change();
```

## 2022_09_08_065652_add_country_code_field_in_settings.php

Source: `hms/database/migrations/2022_09_08_065652_add_country_code_field_in_settings.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2022_09_08_201840_defalut_new_module_seeder.php

Source: `hms/database/migrations/2022_09_08_201840_defalut_new_module_seeder.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2022_09_26_214705_create_admins_table.php

Source: `hms/database/migrations/2022_09_26_214705_create_admins_table.php`

- `create` `admins` at line 14

```php
$table->id('id');
$table->integer('user_id');
$table->timestamps();
$table->softDeletes();
```

## 2022_09_30_205212_create_currency_settings_table.php

Source: `hms/database/migrations/2022_09_30_205212_create_currency_settings_table.php`

- `create` `currency_settings` at line 14

```php
$table->id('id');
$table->string('currency_name');
$table->string('currency_icon');
$table->string('currency_code');
$table->timestamps();
$table->softDeletes();
```

## 2022_10_06_165905_create_admin_module_seeder_migration.php

Source: `hms/database/migrations/2022_10_06_165905_create_admin_module_seeder_migration.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2022_10_07_204913_create_default_currency_seeder_migration.php

Source: `hms/database/migrations/2022_10_07_204913_create_default_currency_seeder_migration.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2022_10_11_183203_create_change_field_type_employee_payroll.php

Source: `hms/database/migrations/2022_10_11_183203_create_change_field_type_employee_payroll.php`

- `table` `employee_payrolls` at line 14

```php
$table->string('net_salary')->change();
```

## 2022_11_02_163443_add_currency_field.php

Source: `hms/database/migrations/2022_11_02_163443_add_currency_field.php`

- `table` `employee_payrolls` at line 14

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `invoices` at line 17

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `invoice_items` at line 20

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `payments` at line 23

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `advanced_payments` at line 26

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `bills` at line 29

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `beds` at line 32

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `blood_issues` at line 35

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `patient_cases` at line 38

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `incomes` at line 41

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `expenses` at line 44

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `charges` at line 47

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `doctor_opd_charges` at line 50

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `ipd_charges` at line 53

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `ipd_payments` at line 56

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `opd_patient_departments` at line 59

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `item_stocks` at line 62

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `medicines` at line 65

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `pathology_tests` at line 68

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `radiology_tests` at line 71

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `insurances` at line 74

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `packages` at line 77

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `services` at line 80

```php
$table->string('currency_symbol', 100)->nullable();
```

- `table` `ambulance_calls` at line 83

```php
$table->string('currency_symbol', 100)->nullable();
```

## 2022_11_04_161324_add_default_column_at_admin.php

Source: `hms/database/migrations/2022_11_04_161324_add_default_column_at_admin.php`

- `table` `admins` at line 14

```php
$table->string('is_default')->after('user_id')->default(0)->nullable();
```

## 2022_11_12_014432_change_length_of_password_reset_table.php

Source: `hms/database/migrations/2022_11_12_014432_change_length_of_password_reset_table.php`

- `table` `password_resets` at line 14

```php
$table->string('token', 1000)->change();
```

## 2022_11_30_125757_change_net_salary_datatype_employee_payroll.php

Source: `hms/database/migrations/2022_11_30_125757_change_net_salary_datatype_employee_payroll.php`

- `table` `employee_payrolls` at line 14

```php
$table->bigInteger('net_salary')->change();
```

## 2022_12_10_064724_add_column_at_personal_access_token.php

Source: `hms/database/migrations/2022_12_10_064724_add_column_at_personal_access_token.php`

- `table` `personal_access_tokens` at line 14

```php
$table->timestamp('expires_at')->nullable()->after('last_used_at');
```

## 2022_12_23_064152_add_key_at_setting_table.php

Source: `hms/database/migrations/2022_12_23_064152_add_key_at_setting_table.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2023_01_12_112215_change_case_id_in_ipd_patient_department_table.php

Source: `hms/database/migrations/2023_01_12_112215_change_case_id_in_ipd_patient_department_table.php`

- `table` `ipd_patient_departments` at line 14

```php
$table->unsignedInteger('case_id')->nullable()->change();
```

## 2023_01_18_100008_create_operation_categories_table.php

Source: `hms/database/migrations/2023_01_18_100008_create_operation_categories_table.php`

- `create` `operation_categories` at line 14

```php
$table->bigIncrements('id');
$table->string('name')->unique();
$table->timestamps();
```

## 2023_01_18_130141_add_operation_category_in_modules.php

Source: `hms/database/migrations/2023_01_18_130141_add_operation_category_in_modules.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2023_01_20_041105_create_operations_table.php

Source: `hms/database/migrations/2023_01_20_041105_create_operations_table.php`

- `create` `operations` at line 14

```php
$table->id();
$table->unsignedBigInteger('operation_category_id');
$table->string('name')->unique();
$table->timestamps();
$table->foreign('operation_category_id')->on('operation_categories')->references('id') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2023_01_20_091507_create_operation_seeder_run_migration.php

Source: `hms/database/migrations/2023_01_20_091507_create_operation_seeder_run_migration.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2023_01_24_084500_create_ipd_operation_table.php

Source: `hms/database/migrations/2023_01_24_084500_create_ipd_operation_table.php`

- `create` `ipd_operation` at line 14

```php
$table->id();
$table->unsignedBigInteger('operation_category_id');
$table->unsignedBigInteger('operation_id');
$table->unsignedInteger('ipd_patient_department_id');
$table->date('operation_date');
$table->unsignedBigInteger('doctor_id');
$table->string('assistant_consultant_1');
$table->string('assistant_consultant_2');
$table->string('anesthetist');
$table->string('anesthesia_type');
$table->string('ot_technician');
$table->string('ot_assistant');
$table->string('remark');
$table->string('result');
$table->timestamps();
$table->foreign('operation_category_id')->references('id')->on('operation_categories') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('operation_id')->references('id')->on('operations') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('ipd_patient_department_id')->references('id')->on('ipd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2023_01_26_111156_change_ipd_operation_table.php

Source: `hms/database/migrations/2023_01_26_111156_change_ipd_operation_table.php`

- `table` `ipd_operation` at line 14

```php
$table->string('assistant_consultant_1')->nullable()->change();
$table->string('assistant_consultant_2')->nullable()->change();
$table->string('anesthetist')->nullable()->change();
$table->string('anesthesia_type')->nullable()->change();
$table->string('ot_technician')->nullable()->change();
$table->string('ot_assistant')->nullable()->change();
$table->string('remark')->nullable()->change();
$table->string('result')->nullable()->change();
```

## 2023_01_27_113441_add_reference_no_field_at_ipd_operation_table.php

Source: `hms/database/migrations/2023_01_27_113441_add_reference_no_field_at_ipd_operation_table.php`

- `table` `ipd_operation` at line 14

```php
$table->string('ref_no')->nullable()->after('id');
```

## 2023_06_30_115440_create_used_medicines_table.php

Source: `hms/database/migrations/2023_06_30_115440_create_used_medicines_table.php`

- `create` `used_medicines` at line 15

```php
$table->id();
$table->integer('stock_used');
$table->unsignedInteger('medicine_id')->nullable();
$table->integer('model_id');
$table->string('model_type');
$table->timestamps();
$table->foreign('medicine_id') ->references('id') ->on('medicines') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2023_06_30_115536_add_quantity_field_in_medicines_table.php

Source: `hms/database/migrations/2023_06_30_115536_add_quantity_field_in_medicines_table.php`

- `table` `medicines` at line 15

```php
$table->integer('quantity')->after('buying_price');
```

- `table` `medicines` at line 19

```php
$table->integer('available_quantity')->after('quantity');
```

## 2023_06_30_120038_create_purchase_medicines_table.php

Source: `hms/database/migrations/2023_06_30_120038_create_purchase_medicines_table.php`

- `create` `purchase_medicines` at line 14

```php
$table->id();
$table->string('purchase_no');
$table->float('tax', 25, 2);
$table->float('total', 25, 2);
$table->float('net_amount', 25, 2);
$table->integer('payment_type');
$table->float('discount', 25, 2);
$table->string('note')->nullable();
$table->string('payment_note')->nullable();
$table->timestamps();
```

## 2023_06_30_120054_create_purchased_medicines_table.php

Source: `hms/database/migrations/2023_06_30_120054_create_purchased_medicines_table.php`

- `create` `purchased_medicines` at line 14

```php
$table->id();
$table->unsignedBigInteger('purchase_medicines_id');
$table->integer('medicine_id')->unsigned()->nullable();
$table->dateTime('expiry_date')->nullable();
$table->string('lot_no');
$table->float('tax', 25, 2);
$table->integer('quantity');
$table->float('amount', 25, 2);
$table->timestamps();
$table->foreign('medicine_id')->references('id')->on('medicines') ->onDelete('set null') ->onUpdate('cascade');
$table->foreign('purchase_medicines_id')->references('id')->on('purchase_medicines') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2023_07_06_095020_add_dose_interval_field_in_prescriptions_medicines_table.php

Source: `hms/database/migrations/2023_07_06_095020_add_dose_interval_field_in_prescriptions_medicines_table.php`

- `table` `prescriptions_medicines` at line 14

```php
$table->integer('dose_interval')->after('time');
```

## 2023_07_06_130647_add_dose_interval_field_in_ipd_prescription_items_table.php

Source: `hms/database/migrations/2023_07_06_130647_add_dose_interval_field_in_ipd_prescription_items_table.php`

- `table` `ipd_prescription_items` at line 14

```php
$table->integer('dose_interval')->after('dosage');
$table->string('day')->after('dose_interval')->nullable();
```

## 2023_07_07_053436_create_medicine_bills_table.php

Source: `hms/database/migrations/2023_07_07_053436_create_medicine_bills_table.php`

- `create` `medicine_bills` at line 14

```php
$table->id();
$table->string('bill_number');
$table->unsignedInteger('patient_id');
$table->unsignedInteger('doctor_id')->nullable();
$table->string('model_type');
$table->string('model_id');
$table->float('discount', 25, 2);
$table->float('net_amount', 25, 2);
$table->float('total', 25, 2);
$table->float('tax_amount', 25, 2);
$table->integer('payment_status');
$table->integer('payment_type');
$table->string('note')->nullable();
$table->datetime('bill_date');
$table->timestamps();
```

## 2023_07_07_053704_create_sale_medicines_table.php

Source: `hms/database/migrations/2023_07_07_053704_create_sale_medicines_table.php`

- `create` `sale_medicines` at line 14

```php
$table->id();
$table->unsignedInteger('medicine_bill_id');
$table->unsignedInteger('medicine_id');
$table->integer('sale_quantity');
$table->float('sale_price', 25, 2);
$table->float('tax', 25, 2);
$table->dateTime('expiry_date');
$table->float('amount', 25, 2);
$table->timestamps();
```

## 2023_07_11_070125_add_time_field_in_ipd_prescription_items_table.php

Source: `hms/database/migrations/2023_07_11_070125_add_time_field_in_ipd_prescription_items_table.php`

- `table` `ipd_prescription_items` at line 14

```php
$table->string('time')->nullable();
```

## 2023_08_23_084806_create_zoom_o_auth_credentials_table.php

Source: `hms/database/migrations/2023_08_23_084806_create_zoom_o_auth_credentials_table.php`

- `create` `zoom_o_auth_credentials` at line 14

```php
$table->id();
$table->unsignedBigInteger('user_id');
$table->text('access_token');
$table->text('refresh_token');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onUpdate('CASCADE') ->onDelete('CASCADE');
```

## 2023_08_29_120924_add_charge_id_field_to_radiology_tests_table.php

Source: `hms/database/migrations/2023_08_29_120924_add_charge_id_field_to_radiology_tests_table.php`

- `table` `radiology_tests` at line 14

```php
$table->unsignedInteger('charge_id')->nullable();
$table->foreign('charge_id')->references('id')->on('charges') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2023_09_20_000000_rename_password_resets_table.php

Source: `hms/database/migrations/2023_09_20_000000_rename_password_resets_table.php`

- `rename` `password_resets` at line 13 → `password_reset_tokens`
## 2023_10_05_043821_create_pathology_units_table.php

Source: `hms/database/migrations/2023_10_05_043821_create_pathology_units_table.php`

- `create` `pathology_units` at line 14

```php
$table->increments('id');
$table->string('name');
$table->timestamps();
```

## 2023_10_05_053003_create_pathology_parameters_table.php

Source: `hms/database/migrations/2023_10_05_053003_create_pathology_parameters_table.php`

- `create` `pathology_parameters` at line 14

```php
$table->increments('id');
$table->string('parameter_name');
$table->string('reference_range');
$table->unsignedInteger('unit_id');
$table->text('description')->nullable();
$table->timestamps();
$table->foreign('unit_id')->references('id')->on('pathology_units') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2023_10_05_063557_create_pathology_parameter_items_table.php

Source: `hms/database/migrations/2023_10_05_063557_create_pathology_parameter_items_table.php`

- `create` `pathology_parameter_items` at line 14

```php
$table->id();
$table->unsignedInteger('pathology_id');
$table->text('patient_result');
$table->unsignedInteger('parameter_id');
$table->timestamps();
$table->foreign('pathology_id')->references('id')->on('pathology_tests') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('parameter_id')->references('id')->on('pathology_parameters') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2023_10_06_111045_create_doctor_holidays_table.php

Source: `hms/database/migrations/2023_10_06_111045_create_doctor_holidays_table.php`

- `create` `doctor_holidays` at line 14

```php
$table->id();
$table->string('name')->nullable();
$table->unsignedBigInteger('doctor_id');
$table->string('date');
$table->timestamps();
$table->foreign('doctor_id')->references('id')->on('doctors')->onDelete('cascade')->onUpdate('cascade');
```

## 2023_10_09_092910_create_lunch_breaks_table.php

Source: `hms/database/migrations/2023_10_09_092910_create_lunch_breaks_table.php`

- `create` `lunch_breaks` at line 14

```php
$table->id();
$table->unsignedBigInteger('doctor_id');
$table->time('break_from');
$table->time('break_to');
$table->boolean('every_day')->nullable();
$table->string('date')->nullable();
$table->timestamps();
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2023_10_20_104235_create_default_pathology_module.php

Source: `hms/database/migrations/2023_10_20_104235_create_default_pathology_module.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2023_10_30_044424_create_patient_id_card_templates_table.php

Source: `hms/database/migrations/2023_10_30_044424_create_patient_id_card_templates_table.php`

- `create` `patient_id_card_templates` at line 14

```php
$table->id();
$table->string('name');
$table->string('color');
$table->boolean('email')->default(1);
$table->boolean('phone')->default(1);
$table->boolean('dob')->default(1);
$table->boolean('blood_group')->default(1);
$table->boolean('address')->default(1);
$table->boolean('patient_unique_id')->default(1);
$table->timestamps();
```

## 2023_10_30_050649_add_description_to_doctors_table.php

Source: `hms/database/migrations/2023_10_30_050649_add_description_to_doctors_table.php`

- `table` `doctors` at line 14

```php
$table->text('description')->nullable()->after('specialist');
```

## 2023_10_30_053356_add_template_id_field_to_patients_table.php

Source: `hms/database/migrations/2023_10_30_053356_add_template_id_field_to_patients_table.php`

- `table` `patients` at line 14

```php
$table->unsignedBigInteger('template_id')->nullable()->after('user_id');
$table->foreign('template_id')->references('id')->on('patient_id_card_templates') ->onDelete('CASCADE') ->onUpdate('CASCADE');
```

## 2023_10_31_040501_add_patient_unique_id_field_in_patients_table.php

Source: `hms/database/migrations/2023_10_31_040501_add_patient_unique_id_field_in_patients_table.php`

- `table` `patients` at line 14

```php
$table->string('patient_unique_id')->nullable()->after('user_id');
```

## 2023_11_03_104055_create_default_language_seeder_migration.php

Source: `hms/database/migrations/2023_11_03_104055_create_default_language_seeder_migration.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2023_11_09_095046_create_run_default_patient_unique_id_seeder_table.php

Source: `hms/database/migrations/2023_11_09_095046_create_run_default_patient_unique_id_seeder_table.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2023_12_01_044623_add_status_column_in_bills_table.php

Source: `hms/database/migrations/2023_12_01_044623_add_status_column_in_bills_table.php`

- `table` `bills` at line 14

```php
$table->string('status')->nullable()->after('amount');
```

## 2023_12_01_045242_create_bill_transactions_table.php

Source: `hms/database/migrations/2023_12_01_045242_create_bill_transactions_table.php`

- `create` `bill_transactions` at line 14

```php
$table->id();
$table->string('transaction_id')->nullable();
$table->integer('payment_type')->comment('1 Stripe, 2 Manual');
$table->double('amount');
$table->unsignedInteger('bill_id');
$table->string('status')->nullable();
$table->text('meta')->nullable();
$table->boolean('is_manual_payment')->nullable();
$table->timestamps();
$table->foreign('bill_id')->references('id')->on('bills')->onUpdate('cascade')->onDelete('cascade');
```

## 2023_12_01_112531_add_payment_mode_column.php

Source: `hms/database/migrations/2023_12_01_112531_add_payment_mode_column.php`

- `table` `bills` at line 14

```php
$table->string('payment_mode')->nullable()->after('status');
```

## 2023_12_04_084628_add_discharge_filed_ipd_patient_departments_to_table.php

Source: `hms/database/migrations/2023_12_04_084628_add_discharge_filed_ipd_patient_departments_to_table.php`

- `table` `ipd_patient_departments` at line 14

```php
$table->boolean('discharge')->default(false)->after('admission_date');
```

## 2023_12_19_071947_create_add_custom_fields_table.php

Source: `hms/database/migrations/2023_12_19_071947_create_add_custom_fields_table.php`

- `create` `add_custom_fields` at line 14

```php
$table->id();
$table->string('module_name');
$table->string('field_type');
$table->string('field_name');
$table->boolean('is_required');
$table->text('values')->nullable();
$table->integer('grid')->default(12);
$table->timestamps();
```

## 2023_12_21_061835_add_custom_field_to_appointments_table.php

Source: `hms/database/migrations/2023_12_21_061835_add_custom_field_to_appointments_table.php`

- `table` `appointments` at line 14

```php
$table->string('custom_field')->after('is_completed')->nullable();
```

## 2023_12_21_105145_add_custom_field_to_ipd_patient_departments_table.php

Source: `hms/database/migrations/2023_12_21_105145_add_custom_field_to_ipd_patient_departments_table.php`

- `table` `ipd_patient_departments` at line 14

```php
$table->string('custom_field')->after('bed_id')->nullable();
```

## 2023_12_21_114522_add_custom_field_to_opd_patient_departments_table.php

Source: `hms/database/migrations/2023_12_21_114522_add_custom_field_to_opd_patient_departments_table.php`

- `table` `opd_patient_departments` at line 14

```php
$table->string('custom_field')->after('payment_mode')->nullable();
```

## 2023_12_21_121555_add_custom_field_to_patients_table.php

Source: `hms/database/migrations/2023_12_21_121555_add_custom_field_to_patients_table.php`

- `table` `patients` at line 14

```php
$table->string('custom_field')->after('template_id')->nullable();
```

## 2023_12_26_094330_create_ipd_patient_department_seeder_table.php

Source: `hms/database/migrations/2023_12_26_094330_create_ipd_patient_department_seeder_table.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2024_01_16_094958_add_ipd_patient_department_bed_assign_seeder_table.php

Source: `hms/database/migrations/2024_01_16_094958_add_ipd_patient_department_bed_assign_seeder_table.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2024_01_16_103253_create_ipd_patient_department_bed_assign_seeder_table.php

Source: `hms/database/migrations/2024_01_16_103253_create_ipd_patient_department_bed_assign_seeder_table.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2024_01_26_062011_add_appointment_charge_at_doctors_table.php

Source: `hms/database/migrations/2024_01_26_062011_add_appointment_charge_at_doctors_table.php`

- `table` `doctors` at line 14

```php
$table->integer('appointment_charge')->after('description')->nullable()->default(0);
```

## 2024_01_26_094657_create_appointment_transactions_table.php

Source: `hms/database/migrations/2024_01_26_094657_create_appointment_transactions_table.php`

- `create` `appointment_transactions` at line 14

```php
$table->id();
$table->string('transaction_id')->nullable();
$table->integer('payment_type');
$table->double('amount');
$table->unsignedBigInteger('appointment_id');
$table->timestamps();
$table->foreign('appointment_id')->references('id')->on('appointments')->onUpdate('cascade')->onDelete('cascade');
```

## 2024_01_26_113651_add_payment_status_at_appointments_table.php

Source: `hms/database/migrations/2024_01_26_113651_add_payment_status_at_appointments_table.php`

- `table` `appointments` at line 14

```php
$table->boolean('payment_status')->default(0)->after('custom_field');
$table->integer('payment_type')->after('payment_status');
```

## 2024_02_02_063538_create_setting_table_payement_field_sedder_table.php

Source: `hms/database/migrations/2024_02_02_063538_create_setting_table_payement_field_sedder_table.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2024_02_02_111802_create_setting_table_paystack_field_seeder_table.php

Source: `hms/database/migrations/2024_02_02_111802_create_setting_table_paystack_field_seeder_table.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2024_02_17_112301_create_google_calendar_integrations_table.php

Source: `hms/database/migrations/2024_02_17_112301_create_google_calendar_integrations_table.php`

- `create` `google_calendar_integrations` at line 14

```php
$table->id();
$table->unsignedBigInteger('user_id');
$table->text('access_token');
$table->text('meta');
$table->string('last_used_at')->default(null);
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2024_02_17_112840_create_google_calendar_lists_table.php

Source: `hms/database/migrations/2024_02_17_112840_create_google_calendar_lists_table.php`

- `create` `google_calendar_lists` at line 14

```php
$table->id();
$table->unsignedBigInteger('user_id');
$table->string('calendar_name');
$table->string('google_calendar_id');
$table->text('meta');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onUpdate('cascade') ->onDelete('cascade');
```

## 2024_02_17_113008_create_event_google_calendars_table.php

Source: `hms/database/migrations/2024_02_17_113008_create_event_google_calendars_table.php`

- `create` `event_google_calendars` at line 14

```php
$table->id();
$table->unsignedBigInteger('user_id');
$table->unsignedBigInteger('google_calendar_list_id');
$table->string('google_calendar_id');
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onUpdate('CASCADE') ->onDelete('CASCADE');
$table->foreign('google_calendar_list_id')->references('id')->on('google_calendar_lists') ->onUpdate('CASCADE') ->onDelete('CASCADE');
```

## 2024_02_17_113110_add_column_in_live_consultations_table.php

Source: `hms/database/migrations/2024_02_17_113110_add_column_in_live_consultations_table.php`

- `table` `live_consultations` at line 14

```php
$table->string('platform_type')->after('password');
```

## 2024_02_17_113257_create_user_google_event_schedules_table.php

Source: `hms/database/migrations/2024_02_17_113257_create_user_google_event_schedules_table.php`

- `create` `user_google_event_schedules` at line 14

```php
$table->id();
$table->unsignedBigInteger('user_id');
$table->string('google_live_consultation_id');
$table->string('google_calendar_id');
$table->string('google_event_id');
$table->string('google_meet_link')->nullable();
$table->timestamps();
$table->foreign('user_id')->references('id')->on('users') ->onUpdate('CASCADE') ->onDelete('CASCADE');
```

## 2024_02_23_043211_create_phone_pay_seeder_table.php

Source: `hms/database/migrations/2024_02_23_043211_create_phone_pay_seeder_table.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2024_02_26_105319_add_patient_id_to_pathology_tests_table.php

Source: `hms/database/migrations/2024_02_26_105319_add_patient_id_to_pathology_tests_table.php`

- `table` `pathology_tests` at line 14

```php
$table->unsignedInteger('patient_id')->nullable()->after('standard_charge');
$table->foreign('patient_id') ->references('id') ->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2024_02_28_095146_create_setting_table_payment_getaway_field_seeder_table.php

Source: `hms/database/migrations/2024_02_28_095146_create_setting_table_payment_getaway_field_seeder_table.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2024_03_01_115646_create_opd_prescriptions_table.php

Source: `hms/database/migrations/2024_03_01_115646_create_opd_prescriptions_table.php`

- `create` `opd_prescriptions` at line 14

```php
$table->id();
$table->unsignedInteger('opd_patient_department_id');
$table->text('header_note')->nullable();
$table->text('footer_note')->nullable();
$table->timestamps();
$table->foreign('opd_patient_department_id')->references('id')->on('opd_patient_departments') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2024_03_04_070757_create_opd_prescription_items_table.php

Source: `hms/database/migrations/2024_03_04_070757_create_opd_prescription_items_table.php`

- `create` `opd_prescription_items` at line 14

```php
$table->id();
$table->unsignedBigInteger('opd_prescription_id');
$table->unsignedInteger('category_id');
$table->unsignedInteger('medicine_id');
$table->string('dosage');
$table->integer('dose_interval');
$table->string('day')->nullable();
$table->string('time')->nullable();
$table->text('instruction');
$table->timestamps();
$table->foreign('opd_prescription_id')->references('id')->on('opd_prescriptions') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('category_id')->references('id')->on('categories') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('medicine_id')->references('id')->on('medicines') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2024_03_16_072822_add_payment_status_field_in_purchase_medicines_table.php

Source: `hms/database/migrations/2024_03_16_072822_add_payment_status_field_in_purchase_medicines_table.php`

- `table` `purchase_medicines` at line 14

```php
$table->boolean('payment_status')->default(1)->nullable()->after('payment_type');
```

## 2024_05_01_045327_change_token_field_in_password_reset_tokens_table.php

Source: `hms/database/migrations/2024_05_01_045327_change_token_field_in_password_reset_tokens_table.php`

- `table` `password_reset_tokens` at line 14

```php
$table->longText('token')->change();
```

## 2024_05_14_084350_add_google_json_file_path_field_in_doctor_table.php

Source: `hms/database/migrations/2024_05_14_084350_add_google_json_file_path_field_in_doctor_table.php`

- `table` `doctors` at line 14

```php
$table->longText('google_json_file_path')->after('appointment_charge')->nullable();
```

## 2024_05_15_113139_change_custom_field_to_patients_table.php

Source: `hms/database/migrations/2024_05_15_113139_change_custom_field_to_patients_table.php`

- `table` `patients` at line 14

```php
$table->longText('custom_field')->change();
```

## 2024_05_15_113152_change_custom_field_to_appointments_table.php

Source: `hms/database/migrations/2024_05_15_113152_change_custom_field_to_appointments_table.php`

- `table` `appointments` at line 14

```php
$table->longText('custom_field')->change();
```

## 2024_05_15_113203_change_custom_field_to_ipd_patient_departments_table.php

Source: `hms/database/migrations/2024_05_15_113203_change_custom_field_to_ipd_patient_departments_table.php`

- `table` `ipd_patient_departments` at line 14

```php
$table->longText('custom_field')->change();
```

## 2024_05_15_113219_change_custom_field_to_opd_patient_departments_table.php

Source: `hms/database/migrations/2024_05_15_113219_change_custom_field_to_opd_patient_departments_table.php`

- `table` `opd_patient_departments` at line 14

```php
$table->longText('custom_field')->change();
```

## 2024_05_20_072708_change_expiry_date_field_type_in_sale_medicines_table.php

Source: `hms/database/migrations/2024_05_20_072708_change_expiry_date_field_type_in_sale_medicines_table.php`

- `table` `sale_medicines` at line 14

```php
$table->date('expiry_date')->change();
```

## 2024_05_28_035934_change_amount_table_column.php

Source: `hms/database/migrations/2024_05_28_035934_change_amount_table_column.php`

- `table` `invoices` at line 14

```php
$table->decimal('amount', 16, 2)->default(0)->change();
```

## 2024_05_28_044844_change_price_table_column.php

Source: `hms/database/migrations/2024_05_28_044844_change_price_table_column.php`

- `table` `invoice_items` at line 14

```php
$table->decimal('price', 16, 2)->change();
$table->decimal('total', 16, 2)->change();
```

## 2024_05_28_085124_change_net_salary_table_column.php

Source: `hms/database/migrations/2024_05_28_085124_change_net_salary_table_column.php`

- `table` `employee_payrolls` at line 14

```php
$table->double('net_salary')->change();
```

## 2024_05_28_094532_change_charge_column_to_beds_table.php

Source: `hms/database/migrations/2024_05_28_094532_change_charge_column_to_beds_table.php`

- `table` `beds` at line 14

```php
$table->double('charge')->change();
```

## 2024_05_28_094617_change_amount_column_to_blood_issues_table.php

Source: `hms/database/migrations/2024_05_28_094617_change_amount_column_to_blood_issues_table.php`

- `table` `blood_issues` at line 14

```php
$table->double('amount')->nullable()->change();
```

## 2024_05_28_095246_change_standard_charge_column_to_charges_table.php

Source: `hms/database/migrations/2024_05_28_095246_change_standard_charge_column_to_charges_table.php`

- `table` `charges` at line 14

```php
$table->double('standard_charge')->change();
```

## 2024_05_28_095826_change_standard_charge_column_to_radiology_tests_table.php

Source: `hms/database/migrations/2024_05_28_095826_change_standard_charge_column_to_radiology_tests_table.php`

- `table` `radiology_tests` at line 14

```php
$table->double('standard_charge')->change();
```

## 2024_05_28_100518_change_standard_charge_column_to_pathology_tests_table.php

Source: `hms/database/migrations/2024_05_28_100518_change_standard_charge_column_to_pathology_tests_table.php`

- `table` `pathology_tests` at line 14

```php
$table->double('standard_charge')->change();
```

## 2024_05_30_035823_change_amount_column_to_ipd_charges_table.php

Source: `hms/database/migrations/2024_05_30_035823_change_amount_column_to_ipd_charges_table.php`

- `table` `ipd_charges` at line 14

```php
$table->double('standard_charge')->nullable()->change();
$table->double('applied_charge')->change();
```

## 2024_05_30_035950_change_amount_column_to_ipd_payments_table.php

Source: `hms/database/migrations/2024_05_30_035950_change_amount_column_to_ipd_payments_table.php`

- `table` `ipd_payments` at line 14

```php
$table->double('amount')->change();
```

## 2024_05_30_041541_change_charges_amount_column_to_ipd_bills_table.php

Source: `hms/database/migrations/2024_05_30_041541_change_charges_amount_column_to_ipd_bills_table.php`

- `table` `ipd_bills` at line 14

```php
$table->double('total_charges')->change();
$table->double('total_payments')->change();
$table->double('gross_total')->change();
$table->double('other_charges')->change();
$table->double('net_payable_amount')->change();
```

## 2024_06_04_072039_change_rate_field_datatype_in_services_table.php

Source: `hms/database/migrations/2024_06_04_072039_change_rate_field_datatype_in_services_table.php`

- `table` `services` at line 14

```php
$table->double('rate')->change();
```

## 2024_06_25_084154_add_open_ai_key_field_in_settings_table.php

Source: `hms/database/migrations/2024_06_25_084154_add_open_ai_key_field_in_settings_table.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2024_07_22_054344_change_appointment_charge_field_datatype_in_doctors_table.php

Source: `hms/database/migrations/2024_07_22_054344_change_appointment_charge_field_datatype_in_doctors_table.php`

- `table` `doctors` at line 14

```php
$table->double('appointment_charge')->change();
```

## 2024_09_02_093150_default_pyament_type_seeder.php

Source: `hms/database/migrations/2024_09_02_093150_default_pyament_type_seeder.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.

## 2024_11_11_062027_create_add_ons_table.php

Source: `hms/database/migrations/2024_11_11_062027_create_add_ons_table.php`

- `create` `add_ons` at line 14

```php
$table->id();
$table->string('name')->unique();
$table->boolean('status')->default(1);
$table->timestamps();
```

## 2025_04_23_125529_create_odontograms_table.php

Source: `hms/database/migrations/2025_04_23_125529_create_odontograms_table.php`

- `create` `odontograms` at line 14

```php
$table->id();
$table->json('odontogram')->nullable();
$table->unsignedInteger('patient_id');
$table->unsignedBigInteger('doctor_id');
$table->string('description')->nullable();
$table->timestamps();
$table->foreign('patient_id')->references('id')->on('patients') ->onDelete('cascade') ->onUpdate('cascade');
$table->foreign('doctor_id')->references('id')->on('doctors') ->onDelete('cascade') ->onUpdate('cascade');
```

## 2025_06_09_095853_create_patient_queues_table.php

Source: `hms/database/migrations/2025_06_09_095853_create_patient_queues_table.php`

- `create` `patient_queues` at line 14

```php
$table->id();
$table->foreignId('appointment_id')->constrained('appointments')->cascadeOnDelete()->cascadeOnUpdate();
$table->integer('no');
$table->timestamps();
```

## 2026_03_21_050434_add_two_factor_authentication_fields_in_users_table.php

Source: `hms/database/migrations/2026_03_21_050434_add_two_factor_authentication_fields_in_users_table.php`

- `table` `users` at line 14

```php
$table->boolean('enable_two_factor_authentication')->default(0)->after('email_verified_at');
$table->text('google2fa_secret')->nullable()->after('enable_two_factor_authentication');
$table->text('two_factor_recovery_codes')->nullable()->after('google2fa_secret');
```

## 2026_04_04_074403_create_complaints_table.php

Source: `hms/database/migrations/2026_04_04_074403_create_complaints_table.php`

- `create` `complaints` at line 14

```php
$table->id();
$table->foreignId('patient_id')->constrained('users')->cascadeOnDelete();
$table->string('title');
$table->text('description');
$table->tinyInteger('status')->default(0);
$table->text('response')->nullable();
$table->foreignId('resolved_by')->nullable()->constrained('users')->cascadeOnDelete();
$table->timestamp('resolved_at')->nullable();
$table->timestamps();
```

## 2026_04_13_124041_create_email_templates_table.php

Source: `hms/database/migrations/2026_04_13_124041_create_email_templates_table.php`

- `create` `email_templates` at line 14

```php
$table->id();
$table->string('template_name')->unique();
$table->text('email_subject');
$table->longText('email_content');
$table->json('dynamic_variables')->nullable();
$table->timestamps();
```

## 2026_04_20_114626_run_email_template_seeder.php

Source: `hms/database/migrations/2026_04_20_114626_run_email_template_seeder.php`

No literal/resolved Schema operation extracted; see source for data/seeder-only work.
