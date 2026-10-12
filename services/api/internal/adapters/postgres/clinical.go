package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"hms.local/api/internal/domain"
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
	rows, e := s.DB.Query(ctx, `SELECT b.id,b.name,t.name,b.type_id,b.charge_minor,b.state,b.version,b.state='ready' AND NOT EXISTS(SELECT 1 FROM encounter e WHERE e.bed_id=b.id AND e.status='active'),COALESCE(b.ward_name, 'General Ward') FROM hospital_bed b JOIN bed_type t ON t.id=b.type_id WHERE b.active ORDER BY b.name,b.id LIMIT 25 OFFSET $1`, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Bed{}
	for rows.Next() {
		var b domain.Bed
		if e = rows.Scan(&b.ID, &b.Name, &b.Type, &b.TypeID, &b.ChargeMinor, &b.State, &b.Version, &b.Available, &b.WardName); e != nil {
			return nil, e
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s Store) SearchBeds(ctx context.Context, a domain.Actor, search, ward, bedType string, available *bool, page, pageSize int) ([]domain.Bed, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize

	baseQuery := `
		FROM hospital_bed b
		JOIN bed_type t ON t.id = b.type_id
		LEFT JOIN bed_assignment ba ON ba.bed_id = b.id AND ba.assigned_to IS NULL
		LEFT JOIN encounter e ON (e.id = ba.encounter_id OR (e.bed_id = b.id AND e.status = 'active'))
		LEFT JOIN patient p ON p.id = COALESCE(e.patient_id, (SELECT patient_id FROM encounter WHERE id = ba.encounter_id))
		LEFT JOIN patient_profile pp ON pp.patient_id = p.id
		WHERE ($1 = $1 AND $2 = $2) AND b.active = true
	`
	args := []any{a.Role, a.ID}
	argIdx := 3

	if search = strings.TrimSpace(search); search != "" {
		baseQuery += fmt.Sprintf(` AND (
			b.name ILIKE $%d 
			OR t.name ILIKE $%d 
			OR b.ward_name ILIKE $%d 
			OR b.description ILIKE $%d
			OR (
				($1 = 'admin' OR $1 = 'receptionist'
				 OR ($1 = 'doctor' AND (e.doctor_id = $2 OR EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL)))
				 OR ($1 = 'nurse' AND (EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL) OR EXISTS (SELECT 1 FROM encounter_nurse en WHERE en.encounter_id = e.id AND en.nurse_id = $2 AND en.active)))
				) AND (p.given_name || ' ' || p.family_name) ILIKE $%d
			)
		)`, argIdx, argIdx, argIdx, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	if ward = strings.TrimSpace(ward); ward != "" && ward != "all" {
		baseQuery += fmt.Sprintf(` AND b.ward_name = $%d`, argIdx)
		args = append(args, ward)
		argIdx++
	}

	if bedType = strings.TrimSpace(bedType); bedType != "" && bedType != "all" {
		baseQuery += fmt.Sprintf(` AND (t.name = $%d OR t.id::text = $%d)`, argIdx, argIdx)
		args = append(args, bedType)
		argIdx++
	}

	if available != nil {
		if *available {
			baseQuery += ` AND b.state = 'ready' AND ba.id IS NULL AND e.id IS NULL`
		} else {
			baseQuery += ` AND (b.state != 'ready' OR ba.id IS NOT NULL OR e.id IS NOT NULL)`
		}
	}

	var total int
	countQuery := `SELECT COUNT(DISTINCT b.id) ` + baseQuery
	if err := s.DB.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQuery := fmt.Sprintf(`
		SELECT DISTINCT ON (b.ward_name, b.name, b.id)
		       b.id, b.name, t.name, b.type_id, b.charge_minor, b.state, b.version,
		       (b.state = 'ready' AND ba.id IS NULL AND e.id IS NULL) AS is_available,
		       COALESCE(b.ward_name, 'General Ward') AS ward_name,
		       COALESCE(b.description, '') AS description,
		       b.active,
		       CASE WHEN ($1 = 'admin' OR $1 = 'receptionist'
		         OR ($1 = 'doctor' AND (e.doctor_id = $2 OR EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL)))
		         OR ($1 = 'nurse' AND (EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL) OR EXISTS (SELECT 1 FROM encounter_nurse en WHERE en.encounter_id = e.id AND en.nurse_id = $2 AND en.active)))
		       ) THEN COALESCE(p.given_name || ' ' || p.family_name, '') ELSE '' END AS patient_name,
		       CASE WHEN ($1 = 'admin' OR $1 = 'receptionist'
		         OR ($1 = 'doctor' AND (e.doctor_id = $2 OR EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL)))
		         OR ($1 = 'nurse' AND (EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL) OR EXISTS (SELECT 1 FROM encounter_nurse en WHERE en.encounter_id = e.id AND en.nurse_id = $2 AND en.active)))
		       ) THEN COALESCE(p.phone, '') ELSE '' END AS patient_phone,
		       CASE WHEN ($1 = 'admin' OR $1 = 'receptionist'
		         OR ($1 = 'doctor' AND (e.doctor_id = $2 OR EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL)))
		         OR ($1 = 'nurse' AND (EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL) OR EXISTS (SELECT 1 FROM encounter_nurse en WHERE en.encounter_id = e.id AND en.nurse_id = $2 AND en.active)))
		       ) THEN COALESCE(pp.gender, '') ELSE '' END AS patient_gender,
		       CASE WHEN ($1 = 'admin' OR $1 = 'receptionist'
		         OR ($1 = 'doctor' AND (e.doctor_id = $2 OR EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL)))
		         OR ($1 = 'nurse' AND (EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL) OR EXISTS (SELECT 1 FROM encounter_nurse en WHERE en.encounter_id = e.id AND en.nurse_id = $2 AND en.active)))
		       ) THEN COALESCE(TO_CHAR(COALESCE(ba.assigned_from, e.admitted_at), 'DDth Mon, YYYY HH12:MI:SS AM'), '') ELSE '' END AS admission_date,
		       CASE WHEN ($1 = 'admin' OR $1 = 'receptionist'
		         OR ($1 = 'doctor' AND (e.doctor_id = $2 OR EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL)))
		         OR ($1 = 'nurse' AND (EXISTS (SELECT 1 FROM encounter_care_team ect WHERE ect.encounter_id = e.id AND ect.staff_id = $2 AND ect.revoked_at IS NULL) OR EXISTS (SELECT 1 FROM encounter_nurse en WHERE en.encounter_id = e.id AND en.nurse_id = $2 AND en.active)))
		       ) THEN COALESCE(COALESCE(ba.encounter_id::text, e.id::text), '') ELSE '' END AS encounter_id
		%s
		ORDER BY b.ward_name, b.name, b.id
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := s.DB.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.Bed{}
	for rows.Next() {
		var b domain.Bed
		if err := rows.Scan(
			&b.ID, &b.Name, &b.Type, &b.TypeID, &b.ChargeMinor, &b.State, &b.Version,
			&b.Available, &b.WardName, &b.Description, &b.Active,
			&b.PatientName, &b.PatientPhone, &b.PatientGender,
			&b.AdmissionDate, &b.EncounterID,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, b)
	}
	return out, total, rows.Err()
}

func (s Store) CreateBed(ctx context.Context, a domain.Actor, i domain.BedInput) (domain.Bed, error) {
	if i.WardName == "" {
		i.WardName = "General Ward"
	}
	b := domain.Bed{BedInput: i, Available: true, State: "ready", Version: 1, Active: true}
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
	e = tx.QueryRow(ctx, `
		INSERT INTO hospital_bed(name,bed_type,charge_minor,created_by,type_id,ward_name,description) 
		VALUES($1,$2,$3,$4,$5,$6,$7) 
		RETURNING id
	`, i.Name, b.Type, i.ChargeMinor, a.ID, b.TypeID, i.WardName, i.Description).Scan(&b.ID)
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

const encounterFields = `e.id,e.number,e.kind,e.case_id,COALESCE(e.bed_id::text,''),e.admitted_at,e.symptoms,e.patient_id,p.given_name||' '||p.family_name,e.doctor_id,u.name,COALESCE(b.name,''),e.status,e.version,e.discharged_at,e.discharge_summary,e.intake,
COALESCE((SELECT email FROM "user" WHERE id=patient_portal_owner(p.id)),''),u.email,
COALESCE((SELECT CASE WHEN i.paid_minor>=i.total_minor THEN 'Paid' ELSE 'Unpaid' END FROM encounter_billing eb JOIN invoice i ON i.id=eb.invoice_id WHERE eb.encounter_id=e.id),'Unbilled'),
(SELECT count(*) FROM encounter visits WHERE visits.patient_id=e.patient_id AND visits.doctor_id=e.doctor_id AND visits.kind='opd')`
const encounterFrom = ` FROM encounter e JOIN patient p ON p.id=e.patient_id JOIN "user" u ON u.id=e.doctor_id LEFT JOIN hospital_bed b ON b.id=e.bed_id `
const encounterScope = `($1 IN ('admin','receptionist') OR ($1='doctor' AND e.doctor_id=$2) OR ($1='patient' AND patient_portal_owner(p.id)=$2))`

func scanEncounter(row pgx.Row) (domain.Encounter, error) {
	var e domain.Encounter
	err := row.Scan(&e.ID, &e.Number, &e.Kind, &e.CaseID, &e.BedID, &e.AdmittedAt, &e.Symptoms, &e.PatientID, &e.PatientName, &e.DoctorID, &e.DoctorName, &e.BedName, &e.Status, &e.Version, &e.DischargedAt, &e.DischargeSummary, &e.Intake, &e.PatientEmail, &e.DoctorEmail, &e.BillStatus, &e.TotalVisits)
	return e, clinicalError(err)
}
func (s Store) Encounters(ctx context.Context, a domain.Actor, kind string, page, pageSize int) ([]domain.Encounter, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 25
	}
	offset := (page - 1) * pageSize
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `SELECT `+encounterFields+encounterFrom+` WHERE `+encounterScope+` AND e.kind=$3 ORDER BY e.admitted_at DESC,e.id LIMIT $4 OFFSET $5`, a.Role, a.ID, kind, pageSize, offset)
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
		if originalBed != i.BedID || old.Kind != i.Kind || old.CaseID != i.CaseID || !old.AdmittedAt.Equal(i.AdmittedAt) || old.Symptoms != i.Symptoms || old.Intake != i.Intake {
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
	err = tx.QueryRow(ctx, `INSERT INTO encounter(kind,case_id,patient_id,doctor_id,bed_id,admitted_at,symptoms,bed_charge_minor,created_by,request_key,intake) VALUES($1,$2,$3,$4,NULLIF($5,'')::uuid,$6,$7,$8,$9,$10,$11) RETURNING id`, i.Kind, i.CaseID, patientID, doctorID, i.BedID, i.AdmittedAt, i.Symptoms, charge, a.ID, key, i.Intake).Scan(&id)
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
	if out.BedID != "" {
		// Lock hospital bed
		var dummy string
		_ = tx.QueryRow(ctx, `SELECT id FROM hospital_bed WHERE id=$1 FOR UPDATE`, out.BedID).Scan(&dummy)
		// Close any open bed_assignment
		if _, err = tx.Exec(ctx, `UPDATE bed_assignment SET assigned_to=$2 WHERE encounter_id=$1 AND assigned_to IS NULL`, id, now); err != nil {
			return out, err
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE encounter SET status='discharged',discharged_at=$2,discharge_summary=$3,bed_id=NULL,version=version+1 WHERE id=$1`, id, now, d.Summary); err != nil {
		return out, err
	}
	if out.Kind == "ipd" {
		if _, err = tx.Exec(ctx, `INSERT INTO bed_event(encounter_id,kind,from_bed_id,charge_minor,actor_id,reason,encounter_version) VALUES($1,'discharge',$2,COALESCE((SELECT charge_minor FROM bed_event WHERE encounter_id=$1 ORDER BY encounter_version DESC LIMIT 1),0),$3,$4,$5)`, id, out.BedID, a.ID, d.Summary, out.Version+1); err != nil {
			return out, err
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'encounter.discharged',$2)`, a.ID, id); err != nil {
		return out, err
	}
	out.Status = "discharged"
	out.DischargedAt = &now
	out.DischargeSummary = d.Summary
	out.BedID = ""
	out.BedName = ""
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
