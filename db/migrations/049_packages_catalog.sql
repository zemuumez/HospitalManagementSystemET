-- Migration 049: Packages catalog, package services lines, and admission linkage
CREATE TABLE package (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(160) NOT NULL UNIQUE,
  description text NOT NULL DEFAULT '',
  discount integer NOT NULL CHECK (discount >= 0 AND discount <= 100),
  total_amount_minor bigint NOT NULL CHECK (total_amount_minor >= 0),
  currency_symbol varchar(100) NOT NULL DEFAULT 'ETB',
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_package_name ON package(name);

CREATE TABLE package_service (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  package_id uuid NOT NULL REFERENCES package(id) ON DELETE CASCADE,
  service_id uuid NOT NULL REFERENCES hospital_service(id) ON DELETE RESTRICT,
  quantity integer NOT NULL CHECK (quantity > 0),
  rate_minor bigint NOT NULL CHECK (rate_minor >= 0),
  amount_minor bigint NOT NULL CHECK (amount_minor >= 0),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  UNIQUE (package_id, service_id)
);

CREATE INDEX idx_package_service_package ON package_service(package_id);
CREATE INDEX idx_package_service_service ON package_service(service_id);

-- Link to IPD admission details for admission references and deletion protection
ALTER TABLE ipd_admission_details
  ADD COLUMN IF NOT EXISTS package_id uuid REFERENCES package(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_ipd_admission_details_package ON ipd_admission_details(package_id);
