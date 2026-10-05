CREATE TABLE ambulance (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  vehicle_number varchar(64) NOT NULL UNIQUE,
  vehicle_model varchar(128) NOT NULL,
  year_made varchar(4) NOT NULL CHECK(year_made ~ '^[0-9]{4}$'),
  driver_name varchar(128) NOT NULL,
  driver_license varchar(64) NOT NULL,
  driver_contact varchar(32) NOT NULL,
  vehicle_type smallint NOT NULL CHECK(vehicle_type IN (1, 2)),
  is_available boolean NOT NULL DEFAULT true,
  note text NOT NULL DEFAULT '',
  version integer NOT NULL DEFAULT 1 CHECK(version > 0),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE ambulance_call (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ambulance_id uuid NOT NULL REFERENCES ambulance(id),
  patient_id uuid NOT NULL REFERENCES patient(id),
  driver_name varchar(128) NOT NULL,
  call_date timestamptz NOT NULL,
  amount_minor bigint NOT NULL CHECK(amount_minor >= 0),
  status varchar(32) NOT NULL DEFAULT 'dispatched' CHECK(status IN ('dispatched', 'completed', 'cancelled')),
  pickup_location varchar(255) NOT NULL DEFAULT '',
  destination varchar(255) NOT NULL DEFAULT '',
  notes text NOT NULL DEFAULT '',
  version integer NOT NULL DEFAULT 1 CHECK(version > 0),
  created_by text NOT NULL REFERENCES "user"(id),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_ambulance_call_patient ON ambulance_call(patient_id, call_date DESC);
CREATE INDEX idx_ambulance_call_ambulance ON ambulance_call(ambulance_id, call_date DESC);
CREATE INDEX idx_ambulance_call_status ON ambulance_call(status);

CREATE TABLE ambulance_call_invoice (
  call_id uuid PRIMARY KEY REFERENCES ambulance_call(id),
  invoice_id uuid NOT NULL UNIQUE REFERENCES invoice(id),
  account_id uuid NOT NULL REFERENCES charge_account(id),
  discount_basis_points integer NOT NULL CHECK(discount_basis_points BETWEEN 0 AND 9999),
  created_by text NOT NULL REFERENCES "user"(id),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TRIGGER immutable_ambulance_call_invoice BEFORE UPDATE OR DELETE ON ambulance_call_invoice FOR EACH ROW EXECUTE FUNCTION protect_retained_record();

INSERT INTO ambulance (id, vehicle_number, vehicle_model, year_made, driver_name, driver_license, driver_contact, vehicle_type, is_available, note)
VALUES 
  ('a1111111-1111-1111-1111-111111111111', 'ETH-AMB-01', 'Toyota HiAce Emergency', '2023', 'Abebe Kebede', 'DL-ETH-98214', '+251911223344', 2, true, 'Fully equipped ICU transport'),
  ('a2222222-2222-2222-2222-222222222222', 'ETH-AMB-02', 'Mercedes Sprinter Van', '2022', 'Tadesse Alemu', 'DL-ETH-77312', '+251922334455', 1, true, 'Contracted standard transport')
ON CONFLICT (vehicle_number) DO NOTHING;
