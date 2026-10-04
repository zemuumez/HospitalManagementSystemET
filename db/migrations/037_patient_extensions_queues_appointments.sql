-- Migration 037: Patient queue & token allocation, public appointment requests,
-- appointment fee billing linkage, patient smart cards, guardian contacts/consent,
-- and patient merge events.

-- 1. Appointment billing & invoice linkage
ALTER TABLE service_invoice_link DROP CONSTRAINT IF EXISTS service_invoice_link_source_type_check;
ALTER TABLE service_invoice_link ADD CONSTRAINT service_invoice_link_source_type_check
  CHECK(source_type IN ('service', 'operation', 'ambulance', 'blood_issue', 'appointment'));

CREATE TABLE appointment_billing (
  appointment_id uuid PRIMARY KEY REFERENCES appointment(id) ON DELETE CASCADE,
  fee_minor      bigint NOT NULL CHECK (fee_minor >= 0),
  invoice_id     uuid REFERENCES invoice(id) ON DELETE SET NULL,
  payment_status varchar(32) NOT NULL DEFAULT 'unpaid' CHECK (payment_status IN ('unpaid', 'partially_paid', 'paid', 'refunded')),
  created_at     timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at     timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_appointment_billing_invoice ON appointment_billing(invoice_id);

-- 2. Patient daily queue & token allocation
CREATE TABLE patient_queue (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  doctor_id    text NOT NULL REFERENCES "user"(id),
  patient_id   uuid NOT NULL REFERENCES patient(id),
  appointment_id uuid REFERENCES appointment(id) ON DELETE SET NULL,
  queue_date   date NOT NULL,
  token_number integer NOT NULL CHECK (token_number > 0),
  status       varchar(32) NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'in_consultation', 'completed', 'skipped')),
  notes        text NOT NULL DEFAULT '' CHECK (length(notes) <= 1000),
  created_at   timestamptz NOT NULL DEFAULT clock_timestamp(),
  called_at    timestamptz,
  completed_at timestamptz,
  UNIQUE(doctor_id, queue_date, token_number),
  UNIQUE(doctor_id, queue_date, patient_id)
);
CREATE INDEX idx_patient_queue_lookup ON patient_queue(doctor_id, queue_date, status);
CREATE INDEX idx_patient_queue_patient ON patient_queue(patient_id, queue_date);

-- 3. Public appointment requests
CREATE TABLE public_appointment_request (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_name     varchar(160) NOT NULL CHECK (length(trim(patient_name)) >= 1),
  patient_email    varchar(191) NOT NULL DEFAULT '',
  patient_phone    varchar(64) NOT NULL CHECK (length(trim(patient_phone)) >= 6),
  doctor_id        text NOT NULL REFERENCES "user"(id),
  department_id    uuid REFERENCES doctor_department(id) ON DELETE SET NULL,
  preferred_date   date NOT NULL,
  problem          text NOT NULL DEFAULT '' CHECK (length(problem) <= 2000),
  status           varchar(32) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'rejected')),
  rejection_reason text NOT NULL DEFAULT '' CHECK (length(rejection_reason) <= 500),
  appointment_id   uuid REFERENCES appointment(id) ON DELETE SET NULL,
  reviewed_by      text REFERENCES "user"(id),
  reviewed_at      timestamptz,
  created_at       timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX idx_public_appt_req_status ON public_appointment_request(status, preferred_date);
CREATE INDEX idx_public_appt_req_doctor ON public_appointment_request(doctor_id, preferred_date);

-- 4. Patient smart cards
CREATE TABLE patient_smart_card (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  patient_id     uuid NOT NULL REFERENCES patient(id) ON DELETE CASCADE,
  card_number    text NOT NULL UNIQUE CHECK (length(trim(card_number)) >= 4),
  qr_token       text NOT NULL UNIQUE CHECK (length(trim(qr_token)) >= 16),
  template_name  text NOT NULL DEFAULT 'standard' CHECK (length(trim(template_name)) >= 1),
  status         varchar(32) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked', 'expired')),
  revoked_reason text NOT NULL DEFAULT '' CHECK (length(revoked_reason) <= 500),
  issued_at      timestamptz NOT NULL DEFAULT clock_timestamp(),
  expires_at     timestamptz NOT NULL,
  issued_by      text NOT NULL REFERENCES "user"(id)
);
CREATE INDEX idx_smart_card_patient ON patient_smart_card(patient_id, status);
CREATE INDEX idx_smart_card_qr ON patient_smart_card(qr_token) WHERE status = 'active';

-- 5. Patient guardian contacts and consent preferences
CREATE TABLE patient_contact_consent (
  patient_id           uuid PRIMARY KEY REFERENCES patient(id) ON DELETE CASCADE,
  guardian_name        text NOT NULL DEFAULT '' CHECK (length(guardian_name) <= 150),
  guardian_relation    text NOT NULL DEFAULT '' CHECK (length(guardian_relation) <= 100),
  guardian_phone       text NOT NULL DEFAULT '' CHECK (length(guardian_phone) <= 64),
  guardian_email       text NOT NULL DEFAULT '' CHECK (length(guardian_email) <= 191),
  sms_consent          boolean NOT NULL DEFAULT true,
  email_consent        boolean NOT NULL DEFAULT true,
  data_sharing_consent boolean NOT NULL DEFAULT true,
  updated_at           timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- 6. Patient merge audit event (immutable)
CREATE TABLE patient_merge_event (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  primary_patient_id uuid NOT NULL REFERENCES patient(id),
  merged_patient_id  uuid NOT NULL,
  reason             text NOT NULL CHECK (length(trim(reason)) BETWEEN 1 AND 500),
  actor_id           text NOT NULL REFERENCES "user"(id),
  merged_snapshot    jsonb NOT NULL,
  created_at         timestamptz NOT NULL DEFAULT clock_timestamp()
);
CREATE TRIGGER immutable_patient_merge BEFORE UPDATE OR DELETE ON patient_merge_event
  FOR EACH ROW EXECUTE FUNCTION protect_pharmacy_history();
