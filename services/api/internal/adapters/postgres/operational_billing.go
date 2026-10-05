package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

// The source row is locked before the invoice. Prices come from a delivered
// service, never from client-supplied link metadata or a mutable catalogue.
func linkOperationalInvoice(ctx context.Context, tx pgx.Tx, a domain.Actor, kind, id, invoiceID string, expected *int64) (domain.ServiceInvoiceLink, error) {
	var out domain.ServiceInvoiceLink
	if !domain.UUIDPattern.MatchString(id) {
		return out, domain.ErrValidation
	}
	var query string
	switch kind {
	case "appointment":
		query = `SELECT ap.patient_id::text,b.fee_minor FROM appointment_billing b JOIN appointment ap ON ap.id=b.appointment_id WHERE ap.id=$1 FOR UPDATE OF b`
	case "encounter":
		query = `SELECT e.patient_id::text,b.total_minor FROM encounter_billing b JOIN encounter e ON e.id=b.encounter_id WHERE e.id=$1 FOR UPDATE OF b`
	case "ambulance":
		query = `SELECT patient_id::text,amount_minor FROM ambulance_call WHERE id=$1 AND status='completed' FOR UPDATE`
	case "blood_issue":
		query = `SELECT patient_id,amount_minor FROM blood_issue WHERE id=$1 FOR UPDATE`
	case "service", "operation":
		query = `SELECT patient_id::text,amount_minor FROM patient_service_charge WHERE id=$1 AND kind=$2 FOR UPDATE`
	default:
		return out, domain.ErrValidation
	}
	var patientID string
	var amount int64
	var err error
	if kind == "service" || kind == "operation" {
		err = tx.QueryRow(ctx, query, id, kind).Scan(&patientID, &amount)
	} else {
		err = tx.QueryRow(ctx, query, id).Scan(&patientID, &amount)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrValidation
	}
	if err != nil {
		return out, err
	}
	if expected != nil && *expected != amount {
		return out, domain.ErrValidation
	}
	var existing string
	err = tx.QueryRow(ctx, `SELECT id,invoice_id,source_type,source_id,amount_minor,created_at FROM service_invoice_link WHERE source_type=$1 AND source_id=$2`, kind, id).Scan(&out.ID, &out.InvoiceID, &out.SourceType, &out.SourceID, &out.AmountMinor, &out.CreatedAt)
	if err == nil {
		if out.InvoiceID != invoiceID {
			return out, domain.ErrConflict
		}
		return out, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return out, err
	}
	if kind == "ambulance" {
		err = tx.QueryRow(ctx, `SELECT invoice_id FROM ambulance_call_invoice WHERE call_id=$1`, id).Scan(&existing)
		if err == nil {
			return out, domain.ErrConflict
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return out, err
		}
	}
	var matching bool
	var subtotal, allocated int64
	err = tx.QueryRow(ctx, `SELECT canonical_patient_id(patient_id)=canonical_patient_id($2::uuid),subtotal_minor FROM invoice WHERE id=$1 FOR UPDATE`, invoiceID, patientID).Scan(&matching, &subtotal)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrValidation
	}
	if err != nil {
		return out, err
	}
	if !matching {
		return out, domain.ErrValidation
	}
	if err = tx.QueryRow(ctx, `SELECT COALESCE(sum(amount_minor),0) FROM service_invoice_link WHERE invoice_id=$1`, invoiceID).Scan(&allocated); err != nil {
		return out, err
	}
	if amount <= 0 || amount > subtotal-allocated {
		return out, domain.ErrValidation
	}
	err = tx.QueryRow(ctx, `INSERT INTO service_invoice_link(invoice_id,source_type,source_id,amount_minor) VALUES($1,$2,$3,$4) RETURNING id,invoice_id,source_type,source_id,amount_minor,created_at`, invoiceID, kind, id, amount).Scan(&out.ID, &out.InvoiceID, &out.SourceType, &out.SourceID, &out.AmountMinor, &out.CreatedAt)
	if err != nil {
		return out, clinicalError(err)
	}
	if err = pharmacyAudit(ctx, tx, a, "service_invoice_link.created", out.ID); err != nil {
		return out, err
	}
	return out, nil
}

func (s Store) CreatePatientServiceCharge(ctx context.Context, a domain.Actor, in domain.PatientServiceChargeInput, key string) (domain.PatientServiceCharge, error) {
	var out domain.PatientServiceCharge
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, a.ID+":service-charge:"+key); err != nil {
		return out, err
	}
	err = tx.QueryRow(ctx, `SELECT id,patient_id,kind,catalog_id,quantity,amount_minor,description,created_at FROM patient_service_charge WHERE created_by=$1 AND request_key=$2`, a.ID, key).Scan(&out.ID, &out.PatientID, &out.Kind, &out.CatalogID, &out.Quantity, &out.AmountMinor, &out.Description, &out.CreatedAt)
	if err == nil {
		if out.PatientID != in.PatientID || out.Kind != in.Kind || out.CatalogID != in.CatalogID || out.Quantity != in.Quantity || (in.Kind == "operation" && out.AmountMinor != in.AmountMinor) {
			return out, domain.ErrConflict
		}
		return out, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return out, err
	}
	out.PatientServiceChargeInput = in
	if in.Kind == "service" {
		err = tx.QueryRow(ctx, `SELECT name,rate_minor*$2::bigint FROM hospital_service WHERE id=$1 AND status=1`, in.CatalogID, in.Quantity).Scan(&out.Description, &out.AmountMinor)
	} else {
		err = tx.QueryRow(ctx, `SELECT name FROM hospital_operation WHERE id=$1 AND status=1`, in.CatalogID).Scan(&out.Description)
	}
	if err != nil {
		return out, clinicalError(err)
	}
	if out.AmountMinor <= 0 || out.AmountMinor > 100000000000 {
		return out, domain.ErrValidation
	}
	err = tx.QueryRow(ctx, `INSERT INTO patient_service_charge(patient_id,kind,catalog_id,quantity,amount_minor,description,created_by,request_key) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id,created_at`, in.PatientID, in.Kind, in.CatalogID, in.Quantity, out.AmountMinor, out.Description, a.ID, key).Scan(&out.ID, &out.CreatedAt)
	if err != nil {
		return out, clinicalError(err)
	}
	if err = pharmacyAudit(ctx, tx, a, "service.delivered", out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
