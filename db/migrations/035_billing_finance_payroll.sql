-- 035_billing_finance_payroll.sql
-- Expenses, Incomes, Employee Payroll, and Source Charges Billing Linkages

-- 1. Expense Heads and Expenses
CREATE TABLE hospital_expense_head (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(160) NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

INSERT INTO hospital_expense_head (name, description) VALUES
  ('Building Rent', 'Premises lease and facility rental costs'),
  ('Equipments & Maintenance', 'Medical and facility equipment maintenance and servicing'),
  ('Electricity & Utilities', 'Power grid and utility water billing'),
  ('Fuel & Transport', 'Power generator fuel and hospital transport operational costs'),
  ('Telephone & Internet', 'Hospital communications and telecom infrastructure'),
  ('Cleaning & Hygiene', 'Sanitation, laundry, and medical waste disposal supplies'),
  ('Hospital Supplies', 'Non-pharmaceutical consumable medical and office supplies'),
  ('Miscellaneous Expenses', 'Other administrative operational expenditures')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE hospital_expense (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  expense_head_id uuid NOT NULL REFERENCES hospital_expense_head(id) ON DELETE RESTRICT,
  name varchar(191) NOT NULL,
  invoice_number varchar(191) NOT NULL DEFAULT '',
  date timestamptz NOT NULL,
  amount_minor bigint NOT NULL CHECK(amount_minor > 0),
  description text NOT NULL DEFAULT '',
  recorded_by text NOT NULL REFERENCES "user"(id),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_hospital_expense_date ON hospital_expense(date);
CREATE INDEX idx_hospital_expense_head ON hospital_expense(expense_head_id);

-- 2. Income Heads and Incomes
CREATE TABLE hospital_income_head (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(160) NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

INSERT INTO hospital_income_head (name, description) VALUES
  ('Special Consultations', 'Executive clinical and specialist consultation receipts'),
  ('Canteen & Facility Lease', 'Third-party premises and kiosk lease income'),
  ('Transport & Ambulance Receipts', 'Third-party patient transport and ambulance service receipts'),
  ('Donations & Grants', 'Charitable health foundations and donor contributions'),
  ('Equipment Rental', 'Authorized external rental of medical diagnostic machinery'),
  ('Miscellaneous Income', 'Other ancillary hospital non-patient revenue')
ON CONFLICT (name) DO NOTHING;

CREATE TABLE hospital_income (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  income_head_id uuid NOT NULL REFERENCES hospital_income_head(id) ON DELETE RESTRICT,
  name varchar(191) NOT NULL,
  invoice_number varchar(191) NOT NULL DEFAULT '',
  date timestamptz NOT NULL,
  amount_minor bigint NOT NULL CHECK(amount_minor > 0),
  description text NOT NULL DEFAULT '',
  recorded_by text NOT NULL REFERENCES "user"(id),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_hospital_income_date ON hospital_income(date);
CREATE INDEX idx_hospital_income_head ON hospital_income(income_head_id);

-- 3. Employee Payroll
CREATE TABLE employee_payroll (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  payroll_number text NOT NULL UNIQUE,
  user_id text NOT NULL REFERENCES "user"(id),
  role varchar(64) NOT NULL,
  month varchar(32) NOT NULL,
  year integer NOT NULL CHECK(year BETWEEN 2000 AND 2100),
  basic_salary_minor bigint NOT NULL CHECK(basic_salary_minor >= 0),
  allowance_minor bigint NOT NULL DEFAULT 0 CHECK(allowance_minor >= 0),
  deductions_minor bigint NOT NULL DEFAULT 0 CHECK(deductions_minor >= 0),
  net_salary_minor bigint NOT NULL CHECK(net_salary_minor >= 0),
  status smallint NOT NULL DEFAULT 0 CHECK(status IN (0, 1)),
  payment_date timestamptz,
  created_by text NOT NULL REFERENCES "user"(id),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE(user_id, month, year)
);
CREATE INDEX idx_employee_payroll_user ON employee_payroll(user_id);
CREATE INDEX idx_employee_payroll_period ON employee_payroll(year, month);

-- 4. Source Charges Billing Linkage (Anti-Double-Billing)
CREATE TABLE service_invoice_link (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  invoice_id uuid NOT NULL REFERENCES invoice(id) ON DELETE CASCADE,
  source_type varchar(64) NOT NULL CHECK(source_type IN ('service', 'operation', 'ambulance', 'blood_issue')),
  source_id text NOT NULL,
  amount_minor bigint NOT NULL CHECK(amount_minor >= 0),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE(source_type, source_id)
);
CREATE INDEX idx_service_invoice_link_inv ON service_invoice_link(invoice_id);
