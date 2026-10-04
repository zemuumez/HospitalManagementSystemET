package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

const nursingScope = `($1='admin' OR ($1='doctor' AND e.doctor_id=$2) OR ($1='nurse' AND EXISTS(SELECT 1 FROM encounter_nurse n JOIN staff_access s ON s.user_id=n.nurse_id WHERE n.encounter_id=e.id AND n.nurse_id=$2 AND n.active AND s.active AND s.role='nurse')))`

func (s Store) AssignNurse(ctx context.Context, a domain.Actor, id string, i domain.NurseAssignment) (domain.EncounterNurse, error) {
	out := domain.EncounterNurse{NurseID: i.NurseID, Active: i.Active}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if e = tx.QueryRow(ctx, `SELECT u.name FROM staff_access s JOIN "user" u ON u.id=s.user_id WHERE s.user_id=$1 AND s.role='nurse' AND (s.active OR NOT $2) FOR SHARE OF s`, i.NurseID, i.Active).Scan(&out.Name); e != nil {
		return out, clinicalError(e)
	}
	var status string
	if e = tx.QueryRow(ctx, `SELECT status FROM encounter WHERE id=$1 AND ($2='admin' OR doctor_id=$3) FOR UPDATE`, id, a.Role, a.ID).Scan(&status); e != nil {
		return out, clinicalError(e)
	}
	if i.Active && status != "active" {
		return out, domain.ErrStale
	}
	var version int
	e = tx.QueryRow(ctx, `SELECT version FROM encounter_nurse WHERE encounter_id=$1 AND nurse_id=$2`, id, i.NurseID).Scan(&version)
	if e != nil && !errors.Is(e, pgx.ErrNoRows) {
		return out, e
	}
	if version != i.Version {
		return out, domain.ErrStale
	}
	if i.Active && version == 0 {
		var count int
		if e = tx.QueryRow(ctx, `SELECT count(*) FROM encounter_nurse WHERE encounter_id=$1`, id).Scan(&count); e != nil {
			return out, e
		}
		if count >= 100 {
			return out, domain.ErrValidation
		}
	}
	out.Version = version + 1
	if _, e = tx.Exec(ctx, `INSERT INTO encounter_nurse(encounter_id,nurse_id,active,version) VALUES($1,$2,$3,$4) ON CONFLICT(encounter_id,nurse_id) DO UPDATE SET active=EXCLUDED.active,version=EXCLUDED.version`, id, i.NurseID, i.Active, out.Version); e != nil {
		return out, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO encounter_nurse_event(encounter_id,nurse_id,active,version,actor_id,reason) VALUES($1,$2,$3,$4,$5,$6)`, id, i.NurseID, i.Active, out.Version, a.ID, i.Reason); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "encounter.nurse_assignment", id); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) Nurses(ctx context.Context, a domain.Actor, id string) ([]domain.EncounterNurse, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	if _, e = scanEncounter(tx.QueryRow(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE `+nursingScope+` AND e.id=$3`, a.Role, a.ID, id)); e != nil {
		return nil, e
	}
	rows, e := tx.Query(ctx, `SELECT n.nurse_id,u.name,n.active,n.version FROM encounter_nurse n JOIN "user" u ON u.id=n.nurse_id WHERE n.encounter_id=$1 ORDER BY n.active DESC,u.name,n.nurse_id LIMIT 100`, id)
	if e != nil {
		return nil, e
	}
	out := []domain.EncounterNurse{}
	for rows.Next() {
		var v domain.EncounterNurse
		if e = rows.Scan(&v.NurseID, &v.Name, &v.Active, &v.Version); e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, v)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if e = pharmacyAudit(ctx, tx, a, "encounter.nurses_viewed", id); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) NursingEncounters(ctx context.Context, a domain.Actor, page int) ([]domain.Encounter, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	rows, e := tx.Query(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE `+nursingScope+` ORDER BY e.admitted_at DESC,e.id LIMIT 25 OFFSET $3`, a.Role, a.ID, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.Encounter{}
	for rows.Next() {
		v, e := scanEncounter(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, v)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if e = pharmacyAudit(ctx, tx, a, "nursing.encounters_viewed", ""); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}

const vitalFields = `id,author_id,observed_at,signed_at,measurements,note,COALESCE(correction_of::text,''),correction_reason`

func scanVitals(row pgx.Row) (domain.Vitals, error) {
	var v domain.Vitals
	var data []byte
	e := row.Scan(&v.ID, &v.AuthorID, &v.ObservedAt, &v.SignedAt, &data, &v.Note, &v.CorrectionOf, &v.CorrectionReason)
	if e != nil {
		return v, clinicalError(e)
	}
	e = json.Unmarshal(data, &v.Measurements)
	return v, e
}
func (s Store) Vitals(ctx context.Context, a domain.Actor, id string, page int) ([]domain.Vitals, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	if _, e = scanEncounter(tx.QueryRow(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE `+nursingScope+` AND e.id=$3`, a.Role, a.ID, id)); e != nil {
		return nil, e
	}
	rows, e := tx.Query(ctx, `SELECT `+vitalFields+` FROM clinical_vitals WHERE encounter_id=$1 ORDER BY signed_at DESC,id LIMIT 25 OFFSET $2`, id, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.Vitals{}
	for rows.Next() {
		v, e := scanVitals(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, v)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if e = pharmacyAudit(ctx, tx, a, "encounter.vitals_viewed", id); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) RecordVitals(ctx context.Context, a domain.Actor, id string, i domain.VitalsInput, key string) (domain.Vitals, error) {
	var out domain.Vitals
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if e = pharmacyLock(ctx, tx, a, key, "vitals"); e != nil {
		return out, e
	}
	enc, e := scanEncounter(tx.QueryRow(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE `+nursingScope+` AND e.id=$3 FOR UPDATE OF e`, a.Role, a.ID, id))
	if e != nil {
		return out, e
	}
	hash := requestHash(struct {
		EncounterID string
		Input       domain.VitalsInput
	}{id, i})
	var oldID, oldHash string
	e = tx.QueryRow(ctx, `SELECT id,request_hash FROM clinical_vitals WHERE author_id=$1 AND request_key=$2`, a.ID, key).Scan(&oldID, &oldHash)
	if e == nil {
		if hash != oldHash {
			return out, domain.ErrConflict
		}
		out, e = scanVitals(tx.QueryRow(ctx, `SELECT `+vitalFields+` FROM clinical_vitals WHERE id=$1`, oldID))
		if e != nil {
			return out, e
		}
		return out, tx.Commit(ctx)
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return out, e
	}
	if enc.Status != "active" {
		return out, domain.ErrStale
	}
	if i.ObservedAt.Before(enc.AdmittedAt) {
		return out, domain.ErrValidation
	}
	if i.CorrectionOf != "" {
		var original string
		if e = tx.QueryRow(ctx, `SELECT id FROM clinical_vitals WHERE id=$1 AND encounter_id=$2`, i.CorrectionOf, id).Scan(&original); e != nil {
			return out, clinicalError(e)
		}
	}
	data, e := json.Marshal(i.Measurements)
	if e != nil {
		return out, e
	}
	out, e = scanVitals(tx.QueryRow(ctx, `INSERT INTO clinical_vitals(encounter_id,author_id,observed_at,measurements,note,correction_of,correction_reason,request_key,request_hash) VALUES($1,$2,$3,$4,$5,NULLIF($6,'')::uuid,$7,$8,$9) RETURNING `+vitalFields, id, a.ID, i.ObservedAt, data, i.Note, i.CorrectionOf, i.CorrectionReason, key, hash))
	if e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "encounter.vitals_signed", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
