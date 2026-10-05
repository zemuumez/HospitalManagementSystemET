-- Charge events snapshot a delivered catalogue service, not the catalogue row itself.
CREATE TABLE patient_service_charge (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), patient_id uuid NOT NULL REFERENCES patient(id),
 kind text NOT NULL CHECK(kind IN ('service','operation')), catalog_id uuid NOT NULL,
 description text NOT NULL, quantity integer NOT NULL CHECK(quantity BETWEEN 1 AND 10000),
 amount_minor bigint NOT NULL CHECK(amount_minor BETWEEN 1 AND 100000000000),
 created_by text NOT NULL REFERENCES "user"(id), request_key text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT clock_timestamp(), UNIQUE(created_by,request_key)
);
CREATE TRIGGER service_charge_alias_guard BEFORE INSERT ON patient_service_charge FOR EACH ROW EXECUTE FUNCTION reject_alias_patient_write();
CREATE TRIGGER immutable_service_charge BEFORE UPDATE OR DELETE ON patient_service_charge FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_service_charge_truncate BEFORE TRUNCATE ON patient_service_charge FOR EACH STATEMENT EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_service_invoice_link BEFORE UPDATE OR DELETE ON service_invoice_link FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_service_invoice_link_truncate BEFORE TRUNCATE ON service_invoice_link FOR EACH STATEMENT EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_discharge_summary BEFORE UPDATE OR DELETE ON discharge_summary FOR EACH ROW EXECUTE FUNCTION protect_retained_record();
CREATE TRIGGER immutable_discharge_summary_truncate BEFORE TRUNCATE ON discharge_summary FOR EACH STATEMENT EXECUTE FUNCTION protect_retained_record();
-- Retain money and patient associations after a charge is linked.
CREATE FUNCTION protect_linked_operational_charge() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_TABLE_NAME='appointment_billing' THEN
 IF OLD.invoice_id IS NOT NULL AND (NEW.fee_minor<>OLD.fee_minor OR NEW.invoice_id IS DISTINCT FROM OLD.invoice_id) THEN RAISE EXCEPTION 'linked charge is immutable' USING ERRCODE='23514'; END IF;
 END IF;
 IF TG_TABLE_NAME='encounter_billing' THEN
 IF (OLD.invoice_id IS NOT NULL OR OLD.financial_clearance) AND (NEW.total_minor<>OLD.total_minor OR NEW.bed_days<>OLD.bed_days OR NEW.bed_total_minor<>OLD.bed_total_minor OR NEW.doctor_fee_minor<>OLD.doctor_fee_minor OR NEW.procedure_fee_minor<>OLD.procedure_fee_minor OR NEW.other_charges_minor<>OLD.other_charges_minor OR (OLD.invoice_id IS NOT NULL AND NEW.invoice_id IS DISTINCT FROM OLD.invoice_id)) THEN RAISE EXCEPTION 'linked charge is immutable' USING ERRCODE='23514'; END IF;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER retain_appointment_fee BEFORE UPDATE ON appointment_billing FOR EACH ROW EXECUTE FUNCTION protect_linked_operational_charge();
CREATE TRIGGER retain_encounter_fee BEFORE UPDATE ON encounter_billing FOR EACH ROW EXECUTE FUNCTION protect_linked_operational_charge();
