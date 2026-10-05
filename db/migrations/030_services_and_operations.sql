CREATE TABLE charge_category (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(160) NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '',
  charge_type integer NOT NULL CHECK(charge_type IN (1, 2, 3, 4, 5)),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE hospital_charge (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  charge_type integer NOT NULL CHECK(charge_type IN (1, 2, 3, 4, 5)),
  charge_category_id uuid NOT NULL REFERENCES charge_category(id),
  code varchar(160) NOT NULL UNIQUE,
  standard_charge_minor bigint NOT NULL CHECK(standard_charge_minor >= 0),
  description text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE hospital_service (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(160) NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '',
  quantity integer NOT NULL DEFAULT 1 CHECK(quantity > 0),
  rate_minor bigint NOT NULL CHECK(rate_minor >= 0),
  status smallint NOT NULL DEFAULT 1 CHECK(status IN (0, 1)),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE operation_category (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(191) NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE hospital_operation (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  operation_category_id uuid NOT NULL REFERENCES operation_category(id),
  name varchar(191) NOT NULL,
  description text NOT NULL DEFAULT '',
  status smallint NOT NULL DEFAULT 1 CHECK(status IN (0, 1)),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE custom_field (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  module_name varchar(191) NOT NULL,
  field_type varchar(64) NOT NULL,
  field_name varchar(191) NOT NULL,
  is_required boolean NOT NULL DEFAULT false,
  values text NOT NULL DEFAULT '',
  grid integer NOT NULL DEFAULT 12 CHECK(grid BETWEEN 1 AND 12),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_custom_field_module ON custom_field(module_name);

CREATE TABLE hospital_module_setting (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  module_key varchar(128) NOT NULL UNIQUE,
  name varchar(191) NOT NULL,
  route varchar(191) NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- Seed basic charge categories
INSERT INTO charge_category (id, name, description, charge_type)
VALUES
  ('ca111111-1111-1111-1111-111111111111', 'Investigations', 'Clinical laboratory and imaging investigations', 1),
  ('ca222222-2222-2222-2222-222222222222', 'Operations', 'General surgical and day theater procedures', 2),
  ('ca333333-3333-3333-3333-333333333333', 'Clinical Procedures', 'Bedside clinical nursing and minor interventions', 5)
ON CONFLICT (name) DO NOTHING;

-- Seed sample services
INSERT INTO hospital_service (id, name, description, quantity, rate_minor, status)
VALUES
  ('ba111111-1111-1111-1111-111111111111', 'General Consultation', 'Standard outpatient specialist medical review', 1, 25000, 1),
  ('ba222222-2222-2222-2222-222222222222', 'Wound Dressing (Major)', 'Sterile major trauma or post-op dressing', 1, 15000, 1),
  ('ba333333-3333-3333-3333-333333333333', 'Nebulization Session', 'Respiratory distress nebulizer therapy', 1, 10000, 1)
ON CONFLICT (name) DO NOTHING;

-- Seed sample operation category & operation
INSERT INTO operation_category (id, name)
VALUES
  ('aa111111-1111-1111-1111-111111111111', 'General Surgery'),
  ('aa222222-2222-2222-2222-222222222222', 'Orthopedic Surgery')
ON CONFLICT (name) DO NOTHING;

INSERT INTO hospital_operation (id, operation_category_id, name, description, status)
VALUES
  ('bb111111-1111-1111-1111-111111111111', 'aa111111-1111-1111-1111-111111111111', 'Appendectomy', 'Removal of vermiform appendix', 1),
  ('bb222222-2222-2222-2222-222222222222', 'aa222222-2222-2222-2222-222222222222', 'Closed Fracture Reduction', 'Realignment of fractured bone without incision', 1)
ON CONFLICT DO NOTHING;

-- Seed core module settings
INSERT INTO hospital_module_setting (module_key, name, route, is_active)
VALUES
  ('appointments', 'Appointments', 'appointments', true),
  ('ipd', 'IPD Patients', 'ipd-patients', true),
  ('opd', 'OPD Patients', 'opd-patients', true),
  ('beds', 'Bed Management', 'beds', true),
  ('billing', 'Invoices & Billing', 'invoices', true),
  ('pharmacy', 'Medicines & Dispensing', 'medicines', true),
  ('diagnostics', 'Pathology & Radiology', 'pathology-tests', true),
  ('inventory', 'Item Inventory', 'items', true),
  ('ambulances', 'Ambulance Services', 'ambulances', true),
  ('attendance', 'Staff Attendance', 'attendance', true),
  ('services', 'Hospital Services', 'services', true)
ON CONFLICT (module_key) DO NOTHING;
