package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"hms.local/api/internal/domain"
	"time"
)

func clinicalError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var p *pgconn.PgError
	if errors.As(err, &p) && (p.Code == "23505" || p.Code == "23503") {
		return domain.ErrStale
	}
	return err
}
func (s Store) Beds(ctx context.Context, page int) ([]domain.Bed, error) {
	rows, e := s.DB.Query(ctx, `SELECT b.id,b.name,t.name,b.type_id,b.charge_minor,b.state,b.version,b.state='ready' AND NOT EXISTS(SELECT 1 FROM encounter e WHERE e.bed_id=b.id AND e.status='active') FROM hospital_bed b JOIN bed_type t ON t.id=b.type_id WHERE b.active ORDER BY b.name,b.id LIMIT 25 OFFSET $1`, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Bed{}
	for rows.Next() {
		var b domain.Bed
		if e = rows.Scan(&b.ID, &b.Name, &b.Type, &b.TypeID, &b.ChargeMinor, &b.State, &b.Version, &b.Available); e != nil {
			return nil, e
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
func (s Store) CreateBed(ctx context.Context, a domain.Actor, i domain.BedInput) (domain.Bed, error) {
	b := domain.Bed{BedInput: i, Available: true, State: "ready", Version: 1}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return b, e
	}
	defer tx.Rollback(ctx)
	if i.TypeID == "" {
		if _, e = tx.Exec(ctx, `INSERT INTO bed_type(name) VALUES($1) ON CONFLICT(name) DO NOTHING`, i.Type); e != nil {
			return b, clinicalError(e)
		}
		e = tx.QueryRow(ctx, `SELECT id,name FROM bed_type WHERE name=$1 AND active FOR SHARE`, i.Type).Scan(&b.TypeID, &b.Type)
	} else {
		e = tx.QueryRow(ctx, `SELECT id,name FROM bed_type WHERE id=$1 AND active FOR SHARE`, i.TypeID).Scan(&b.TypeID, &b.Type)
	}
	if e != nil {
		return b, clinicalError(e)
	}
	e = tx.QueryRow(ctx, `INSERT INTO hospital_bed(name,bed_type,charge_minor,created_by,type_id) VALUES($1,$2,$3,$4,$5) RETURNING id`, i.Name, b.Type, i.ChargeMinor, a.ID, b.TypeID).Scan(&b.ID)
	if e != nil {
		return b, clinicalError(e)
	}
	if _, e = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'bed.created',$2)`, a.ID, b.ID); e != nil {
		return b, e
	}
	return b, tx.Commit(ctx)
}

const caseScope = `($1 IN ('admin','receptionist') OR ($1='doctor' AND c.doctor_id=$2) OR ($1='patient' AND patient_portal_owner(p.id)=$2))`

func (s Store) Cases(ctx context.Context, a domain.Actor, page int) ([]domain.Case, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	rows, e := tx.Query(ctx, `SELECT c.id,c.number,c.patient_id,c.doctor_id,c.description,p.given_name||' '||p.family_name,u.name FROM patient_case c JOIN patient p ON p.id=c.patient_id JOIN "user" u ON u.id=c.doctor_id WHERE `+caseScope+` ORDER BY c.created_at DESC,c.id LIMIT 25 OFFSET $3`, a.Role, a.ID, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.Case{}
	for rows.Next() {
		var c domain.Case
		if e = rows.Scan(&c.ID, &c.Number, &c.PatientID, &c.DoctorID, &c.Description, &c.PatientName, &c.DoctorName); e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, c)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action) VALUES($1,'cases.list_viewed')`, a.ID); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) CreateCase(ctx context.Context, a domain.Actor, i domain.CaseInput) (domain.Case, error) {
	c := domain.Case{CaseInput: i}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return c, e
	}
	defer tx.Rollback(ctx)
	e = tx.QueryRow(ctx, `SELECT u.name FROM "user" u JOIN staff_access s ON s.user_id=u.id WHERE u.id=$1 AND s.active AND s.role='doctor' FOR SHARE OF s`, i.DoctorID).Scan(&c.DoctorName)
	if e != nil {
		return c, clinicalError(e)
	}
	e = tx.QueryRow(ctx, `SELECT given_name||' '||family_name FROM patient WHERE id=$1 AND ($2 IN ('admin','receptionist') OR clinician_user_id=$3) FOR SHARE`, i.PatientID, a.Role, a.ID).Scan(&c.PatientName)
	if e != nil {
		return c, clinicalError(e)
	}
	e = tx.QueryRow(ctx, `INSERT INTO patient_case(patient_id,doctor_id,description,created_by) VALUES($1,$2,$3,$4) RETURNING id,number`, i.PatientID, i.DoctorID, i.Description, a.ID).Scan(&c.ID, &c.Number)
	if e != nil {
		return c, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'case.created',$2)`, a.ID, c.ID); e != nil {
		return c, e
	}
	return c, tx.Commit(ctx)
}

const encounterFields = `e.id,e.number,e.kind,e.case_id,COALESCE(e.bed_id::text,''),e.admitted_at,e.symptoms,e.patient_id,p.given_name||' '||p.family_name,e.doctor_id,u.name,COALESCE(b.name,''),e.status,e.version,e.discharged_at,e.discharge_summary`
const encounterFrom = ` FROM encounter e JOIN patient p ON p.id=e.patient_id JOIN "user" u ON u.id=e.doctor_id LEFT JOIN hospital_bed b ON b.id=e.bed_id `
const encounterScope = `($1 IN ('admin','receptionist') OR ($1='doctor' AND e.doctor_id=$2) OR ($1='patient' AND patient_portal_owner(p.id)=$2))`

func scanEncounter(row pgx.Row) (domain.Encounter, error) {
	var e domain.Encounter
	err := row.Scan(&e.ID, &e.Number, &e.Kind, &e.CaseID, &e.BedID, &e.AdmittedAt, &e.Symptoms, &e.PatientID, &e.PatientName, &e.DoctorID, &e.DoctorName, &e.BedName, &e.Status, &e.Version, &e.DischargedAt, &e.DischargeSummary)
	return e, clinicalError(err)
}
func (s Store) Encounters(ctx context.Context, a domain.Actor, kind string, page int) ([]domain.Encounter, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE `+encounterScope+` AND e.kind=$3 ORDER BY e.admitted_at DESC,e.id LIMIT 25 OFFSET $4`, a.Role, a.ID, kind, (page-1)*25)
	if err != nil {
		return nil, err
	}
	out := []domain.Encounter{}
	for rows.Next() {
		e, err := scanEncounter(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, e)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action) VALUES($1,'encounters.list_viewed')`, a.ID); err != nil {
		return nil, err
	}
	return out, tx.Commit(ctx)
}
func (s Store) Admit(ctx context.Context, a domain.Actor, i domain.EncounterInput, key string) (domain.Encounter, error) {
	empty := domain.Encounter{}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, "admit:"+a.ID+":"+key); err != nil {
		return empty, err
	}
	old, err := scanEncounter(tx.QueryRow(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE e.created_by=$1 AND e.request_key=$2`, a.ID, key))
	if err == nil {
		var originalBed string
		if err = tx.QueryRow(ctx, `SELECT COALESCE(admission_bed_id::text,'') FROM encounter WHERE id=$1`, old.ID).Scan(&originalBed); err != nil {
			return empty, err
		}
		if originalBed != i.BedID || old.Kind != i.Kind || old.CaseID != i.CaseID || !old.AdmittedAt.Equal(i.AdmittedAt) || old.Symptoms != i.Symptoms {
			return empty, domain.ErrConflict
		}
		return old, tx.Commit(ctx)
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return empty, err
	}
	var patientID, doctorID string
	err = tx.QueryRow(ctx, `SELECT c.patient_id,c.doctor_id FROM patient_case c JOIN staff_access s ON s.user_id=c.doctor_id WHERE c.id=$1 AND ($2 IN ('admin','receptionist') OR c.doctor_id=$3) AND s.active AND s.role='doctor' FOR SHARE OF s`, i.CaseID, a.Role, a.ID).Scan(&patientID, &doctorID)
	if err != nil {
		return empty, clinicalError(err)
	}
	var charge int64
	if i.Kind == "ipd" {
		err = tx.QueryRow(ctx, `SELECT charge_minor FROM hospital_bed WHERE id=$1 AND active AND state='ready' FOR UPDATE`, i.BedID).Scan(&charge)
		if err != nil {
			return empty, clinicalError(err)
		}
	}
	var id string
	err = tx.QueryRow(ctx, `INSERT INTO encounter(kind,case_id,patient_id,doctor_id,bed_id,admitted_at,symptoms,bed_charge_minor,created_by,request_key) VALUES($1,$2,$3,$4,NULLIF($5,'')::uuid,$6,$7,$8,$9,$10) RETURNING id`, i.Kind, i.CaseID, patientID, doctorID, i.BedID, i.AdmittedAt, i.Symptoms, charge, a.ID, key).Scan(&id)
	if err != nil {
		return empty, clinicalError(err)
	}
	if i.Kind == "ipd" {
		if _, err = tx.Exec(ctx, `INSERT INTO bed_event(encounter_id,kind,to_bed_id,charge_minor,actor_id,encounter_version) VALUES($1,'admission',$2,$3,$4,1)`, id, i.BedID, charge, a.ID); err != nil {
			return empty, err
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'encounter.admitted',$2)`, a.ID, id); err != nil {
		return empty, err
	}
	out, err := scanEncounter(tx.QueryRow(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE e.id=$1`, id))
	if err != nil {
		return empty, err
	}
	return out, tx.Commit(ctx)
}
func (s Store) Discharge(ctx context.Context, a domain.Actor, id string, d domain.Discharge, now time.Time) (domain.Encounter, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.Encounter{}, err
	}
	defer tx.Rollback(ctx)
	out, err := scanEncounter(tx.QueryRow(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE e.id=$1 AND e.doctor_id=$2 FOR UPDATE OF e`, id, a.ID))
	if err != nil {
		return out, err
	}
	if out.Version != d.Version || out.Status != "active" {
		return out, domain.ErrStale
	}
	if now.Before(out.AdmittedAt) {
		return out, domain.ErrValidation
	}
	if _, err = tx.Exec(ctx, `UPDATE encounter SET status='discharged',discharged_at=$2,discharge_summary=$3,version=version+1 WHERE id=$1`, id, now, d.Summary); err != nil {
		return out, err
	}
	if out.Kind == "ipd" {
		if _, err = tx.Exec(ctx, `INSERT INTO bed_event(encounter_id,kind,from_bed_id,charge_minor,actor_id,encounter_version) SELECT id,'discharge',bed_id,COALESCE((SELECT charge_minor FROM bed_event WHERE encounter_id=encounter.id ORDER BY encounter_version DESC LIMIT 1),bed_charge_minor),$2,version FROM encounter WHERE id=$1`, id, a.ID); err != nil {
			return out, err
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'encounter.discharged',$2)`, a.ID, id); err != nil {
		return out, err
	}
	out.Status = "discharged"
	out.DischargedAt = &now
	out.DischargeSummary = d.Summary
	out.Version++
	return out, tx.Commit(ctx)
}
func (s Store) Notes(ctx context.Context, a domain.Actor, id string) ([]domain.ClinicalNote, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	_, err = scanEncounter(tx.QueryRow(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE `+encounterScope+` AND e.id=$3`, a.Role, a.ID, id))
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `SELECT n.id,u.name,n.body,n.signed_at FROM clinical_note n JOIN "user" u ON u.id=n.author_id WHERE encounter_id=$1 ORDER BY signed_at DESC,id LIMIT 100`, id)
	if err != nil {
		return nil, err
	}
	out := []domain.ClinicalNote{}
	for rows.Next() {
		var n domain.ClinicalNote
		if err = rows.Scan(&n.ID, &n.AuthorName, &n.Body, &n.SignedAt); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, n)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'clinical_notes.viewed',$2)`, a.ID, id); err != nil {
		return nil, err
	}
	return out, tx.Commit(ctx)
}
func (s Store) SignNote(ctx context.Context, a domain.Actor, id string, i domain.NoteInput, key string) (domain.ClinicalNote, error) {
	n := domain.ClinicalNote{Body: i.Body, AuthorName: a.Name}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return n, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, "note:"+a.ID+":"+key); err != nil {
		return n, err
	}
	var oldEncounter, oldBody string
	err = tx.QueryRow(ctx, `SELECT id,encounter_id,body,signed_at FROM clinical_note WHERE author_id=$1 AND request_key=$2`, a.ID, key).Scan(&n.ID, &oldEncounter, &oldBody, &n.SignedAt)
	if err == nil {
		if oldEncounter != id || oldBody != i.Body {
			return n, domain.ErrConflict
		}
		return n, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return n, err
	}
	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM encounter WHERE id=$1 AND doctor_id=$2 FOR UPDATE`, id, a.ID).Scan(&status)
	if err != nil {
		return n, clinicalError(err)
	}
	if status != "active" {
		return n, domain.ErrStale
	}
	err = tx.QueryRow(ctx, `INSERT INTO clinical_note(encounter_id,author_id,body,request_key) VALUES($1,$2,$3,$4) RETURNING id,signed_at`, id, a.ID, i.Body, key).Scan(&n.ID, &n.SignedAt)
	if err != nil {
		return n, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'clinical_note.signed',$2)`, a.ID, n.ID); err != nil {
		return n, err
	}
	return n, tx.Commit(ctx)
}
