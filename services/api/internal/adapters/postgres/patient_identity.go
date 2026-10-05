package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

func (s Store) MergePatients(ctx context.Context, a domain.Actor, input domain.MergePatientInput) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return err
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Same lock as portal-link edits, followed by deterministic patient row locks.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(72841022)`); err != nil {
		return err
	}
	var primary, source string
	if err = tx.QueryRow(ctx, `SELECT canonical_patient_id(id)::text FROM patient WHERE id=$1`, input.PrimaryPatientID).Scan(&primary); err != nil {
		return clinicalError(err)
	}
	if err = tx.QueryRow(ctx, `SELECT canonical_patient_id(id)::text FROM patient WHERE id=$1`, input.MergedPatientID).Scan(&source); err != nil {
		return clinicalError(err)
	}
	if primary == source {
		return nil
	}
	rows, err := tx.Query(ctx, `SELECT id FROM patient WHERE canonical_patient_id(id) IN ($1::uuid,$2::uuid) ORDER BY id FOR UPDATE`, primary, source)
	if err != nil {
		return err
	}
	for rows.Next() {
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	var conflicts bool
	err = tx.QueryRow(ctx, `SELECT
 (SELECT count(DISTINCT user_id)>1 FROM patient WHERE canonical_patient_id(id) IN ($1::uuid,$2::uuid))
 OR EXISTS(SELECT 1 FROM encounter WHERE canonical_patient_id(patient_id) IN ($1::uuid,$2::uuid) AND status='active')
 OR EXISTS(SELECT 1 FROM appointment WHERE canonical_patient_id(patient_id) IN ($1::uuid,$2::uuid) AND status IN ('booked','arrived'))
 OR EXISTS(SELECT 1 FROM patient_queue WHERE canonical_patient_id(patient_id) IN ($1::uuid,$2::uuid) AND status IN ('waiting','in_consultation'))`, primary, source).Scan(&conflicts)
	if err != nil {
		return err
	}
	if conflicts {
		return domain.ErrStale
	}
	// Consent is the most restrictive across both identities. Keep primary demographics.
	var sms, email bool
	err = tx.QueryRow(ctx, `SELECT bool_and(d.sms_consent),bool_and(d.email_consent) FROM patient_profile d WHERE canonical_patient_id(patient_id) IN ($1::uuid,$2::uuid)`, primary, source).Scan(&sms, &email)
	if err != nil {
		return err
	}
	var snapshot []byte
	err = tx.QueryRow(ctx, `SELECT jsonb_agg(jsonb_build_object('patient',to_jsonb(p),'profile',(SELECT to_jsonb(d) FROM patient_profile d WHERE d.patient_id=p.id),'contact',(SELECT to_jsonb(c) FROM patient_contact_consent c WHERE c.patient_id=p.id)) ORDER BY p.id) FROM patient p WHERE canonical_patient_id(p.id) IN ($1::uuid,$2::uuid)`, primary, source).Scan(&snapshot)
	if err != nil {
		return err
	}
	var eventID string
	err = tx.QueryRow(ctx, `INSERT INTO patient_merge_event(primary_patient_id,merged_patient_id,reason,actor_id,merged_snapshot) VALUES($1,$2,$3,$4,$5) RETURNING id`, primary, source, input.Reason, a.ID, snapshot).Scan(&eventID)
	if err != nil {
		return err
	}
	before, err := scanProfile(tx.QueryRow(ctx, profileSelect+`WHERE p.id=$1 FOR UPDATE OF d`, primary))
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE patient_profile SET sms_consent=$2,email_consent=$3,version=version+1 WHERE patient_id=$1`, primary, sms, email); err != nil {
		return err
	}
	after, err := scanProfile(tx.QueryRow(ctx, profileSelect+`WHERE p.id=$1`, primary))
	if err != nil {
		return err
	}
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO patient_profile_revision(patient_id,version,actor_id,reason,before_snapshot,after_snapshot) VALUES($1,$2,$3,$4,$5,$6)`, primary, after.Version, a.ID, "Patient identity merge: "+input.Reason, beforeJSON, afterJSON); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO patient_contact_consent(patient_id,sms_consent,email_consent,data_sharing_consent) VALUES($1,$2,$3,false) ON CONFLICT(patient_id) DO UPDATE SET sms_consent=EXCLUDED.sms_consent,email_consent=EXCLUDED.email_consent,data_sharing_consent=false,updated_at=clock_timestamp()`, primary, sms, email); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE patient_identity_alias SET primary_id=$1,merge_event_id=$3 WHERE primary_id=$2`, primary, source, eventID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO patient_identity_alias(alias_id,primary_id,merge_event_id) VALUES($1,$2,$3)`, source, primary, eventID); err != nil {
		return err
	}
	if err = pharmacyAudit(ctx, tx, a, "patient.identity_merged", primary); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s Store) AuthorizePatientRecord(ctx context.Context, a domain.Actor, id string) error {
	var ok bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM patient WHERE id=$3 AND `+scope+`)`, a.Role, a.ID, id).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}
func (s Store) PatientIdentities(ctx context.Context, id string) ([]domain.PatientIdentity, error) {
	rows, err := s.DB.Query(ctx, `SELECT id,canonical_patient_id(id),medical_record_number,given_name||' '||family_name FROM patient WHERE canonical_patient_id(id)=canonical_patient_id($1::uuid) ORDER BY medical_record_number`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PatientIdentity{}
	for rows.Next() {
		var item domain.PatientIdentity
		var mrn int64
		if err = rows.Scan(&item.ID, &item.PrimaryID, &mrn, &item.Name); err != nil {
			return nil, err
		}
		item.MRN = fmt.Sprintf("HMS-%06d", mrn)
		out = append(out, item)
	}
	return out, rows.Err()
}

// Resolve only for reads. Writes use the canonical identifier returned to callers.
func (s Store) canonicalPatient(ctx context.Context, id string) (string, error) {
	var result string
	err := s.DB.QueryRow(ctx, `SELECT canonical_patient_id(id)::text FROM patient WHERE id=$1`, id).Scan(&result)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	return result, err
}
