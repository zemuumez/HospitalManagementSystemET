package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

const profileSelect = `SELECT p.id,p.medical_record_number,p.given_name,p.family_name,COALESCE(p.date_of_birth::text,''),p.phone,d.email,d.gender,d.blood_group,d.address1,d.address2,d.city,d.region,d.country,d.postal_code,d.emergency_name,d.emergency_phone,d.emergency_relationship,d.active,d.sms_consent,d.email_consent,d.version FROM patient p JOIN patient_profile d ON d.patient_id=p.id `

func scanProfile(row pgx.Row) (domain.PatientProfile, error) {
	var p domain.PatientProfile
	var mrn int64
	e := row.Scan(&p.ID, &mrn, &p.GivenName, &p.FamilyName, &p.DateOfBirth, &p.Phone, &p.Email, &p.Gender, &p.BloodGroup, &p.Address1, &p.Address2, &p.City, &p.Region, &p.Country, &p.PostalCode, &p.EmergencyName, &p.EmergencyPhone, &p.EmergencyRelationship, &p.Active, &p.SMSConsent, &p.EmailConsent, &p.Version)
	p.MRN = fmt.Sprintf("HMS-%06d", mrn)
	return p, clinicalError(e)
}
func (s Store) PatientProfile(ctx context.Context, a domain.Actor, id string) (domain.PatientProfile, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.PatientProfile{}, e
	}
	defer tx.Rollback(ctx)
	p, e := scanProfile(tx.QueryRow(ctx, profileSelect+`WHERE `+scope+` AND p.id=$3`, a.Role, a.ID, id))
	if e != nil {
		return p, e
	}
	if e = pharmacyAudit(ctx, tx, a, "patient.profile_viewed", id); e != nil {
		return p, e
	}
	return p, tx.Commit(ctx)
}
func (s Store) UpdatePatientProfile(ctx context.Context, a domain.Actor, id string, i domain.PatientProfileInput) (domain.PatientProfile, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.PatientProfile{}, e
	}
	defer tx.Rollback(ctx)
	before, e := scanProfile(tx.QueryRow(ctx, profileSelect+`WHERE `+scope+` AND p.id=$3 FOR UPDATE OF p,d`, a.Role, a.ID, id))
	if e != nil {
		return before, e
	}
	if before.Version != i.Version {
		return before, domain.ErrStale
	}
	if _, e = tx.Exec(ctx, `UPDATE patient SET given_name=$2,family_name=$3,date_of_birth=NULLIF($4,'')::date,phone=$5 WHERE id=$1`, id, i.GivenName, i.FamilyName, i.DateOfBirth, i.Phone); e != nil {
		return before, e
	}
	if _, e = tx.Exec(ctx, `UPDATE patient_profile SET email=$2,gender=$3,blood_group=$4,address1=$5,address2=$6,city=$7,region=$8,country=$9,postal_code=$10,emergency_name=$11,emergency_phone=$12,emergency_relationship=$13,active=$14,sms_consent=$15,email_consent=$16,version=version+1 WHERE patient_id=$1`, id, i.Email, i.Gender, i.BloodGroup, i.Address1, i.Address2, i.City, i.Region, i.Country, i.PostalCode, i.EmergencyName, i.EmergencyPhone, i.EmergencyRelationship, i.Active, i.SMSConsent, i.EmailConsent); e != nil {
		return before, e
	}
	after, e := scanProfile(tx.QueryRow(ctx, profileSelect+`WHERE p.id=$1`, id))
	if e != nil {
		return after, e
	}
	beforeJSON, e := json.Marshal(before)
	if e != nil {
		return after, e
	}
	afterJSON, e := json.Marshal(after)
	if e != nil {
		return after, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO patient_profile_revision(patient_id,version,actor_id,reason,before_snapshot,after_snapshot) VALUES($1,$2,$3,$4,$5,$6)`, id, after.Version, a.ID, i.Reason, beforeJSON, afterJSON); e != nil {
		return after, e
	}
	if e = pharmacyAudit(ctx, tx, a, "patient.profile_updated", id); e != nil {
		return after, e
	}
	return after, tx.Commit(ctx)
}
func (s Store) PatientRevisions(ctx context.Context, a domain.Actor, id string, page int) ([]domain.ProfileRevision, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	var exists bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM patient WHERE id=$1)`, id).Scan(&exists)
	if e != nil {
		return nil, e
	}
	if !exists {
		return nil, domain.ErrNotFound
	}
	rows, e := tx.Query(ctx, `SELECT version,actor_id,reason,before_snapshot,after_snapshot,created_at FROM patient_profile_revision WHERE patient_id=$1 ORDER BY version DESC LIMIT 25 OFFSET $2`, id, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.ProfileRevision{}
	for rows.Next() {
		var r domain.ProfileRevision
		var before, after []byte
		if e = rows.Scan(&r.Version, &r.ActorID, &r.Reason, &before, &after, &r.CreatedAt); e != nil {
			rows.Close()
			return nil, e
		}
		if e = json.Unmarshal(before, &r.Before); e != nil {
			rows.Close()
			return nil, e
		}
		if e = json.Unmarshal(after, &r.After); e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, r)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if e = pharmacyAudit(ctx, tx, a, "patient.revisions_viewed", id); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
