package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/domain"
)

type Store struct{ DB *pgxpool.Pool }

func (s Store) Actor(ctx context.Context, id string) (domain.Actor, error) {
	var a domain.Actor
	err := s.DB.QueryRow(ctx, `SELECT u.id,u.name,a.role FROM "user" u JOIN staff_access a ON a.user_id=u.id WHERE u.id=$1 AND a.active`, id).Scan(&a.ID, &a.Name, &a.Role)
	return a, err
}

// Scope is enforced inside data access for lists AND aggregate counts.
const scope = `($1 IN ('admin','receptionist') OR ($1='doctor' AND clinician_user_id=$2) OR ($1='patient' AND patient_portal_owner(id)=$2))`

func (s Store) Patients(ctx context.Context, a domain.Actor, search string, page int) ([]domain.Patient, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `SELECT 
		patient.id,
		patient.medical_record_number,
		patient.given_name,
		patient.family_name,
		COALESCE(patient.date_of_birth::text,''),
		patient.phone,
		patient.created_at,
		COALESCE(patient_portal_owner(patient.id),''),
		COALESCE(patient.clinician_user_id,''),
		COALESCE(pp.email,''),
		COALESCE(pp.gender,'unknown'),
		COALESCE(pp.blood_group,''),
		COALESCE(pp.father_name,''),
		COALESCE(pp.active, true)
	FROM patient
	LEFT JOIN patient_profile pp ON pp.patient_id = patient.id
	WHERE canonical_patient_id(patient.id)=patient.id AND `+scope+` AND ($3='' OR strpos(lower(patient.given_name || ' ' || patient.family_name),lower($3))>0 OR patient.medical_record_number::text=$3) ORDER BY patient.created_at DESC,patient.id LIMIT 25 OFFSET $4`, a.Role, a.ID, search, (page-1)*25)
	if err != nil {
		return nil, err
	}
	result := []domain.Patient{}
	for rows.Next() {
		var p domain.Patient
		var mrn int64
		if err = rows.Scan(
			&p.ID,
			&mrn,
			&p.GivenName,
			&p.FamilyName,
			&p.DateOfBirth,
			&p.Phone,
			&p.CreatedAt,
			&p.UserID,
			&p.ClinicianID,
			&p.Email,
			&p.Gender,
			&p.BloodGroup,
			&p.FatherName,
			&p.Active,
		); err != nil {
			rows.Close()
			return nil, err
		}
		p.MRN = fmt.Sprintf("HMS-%06d", mrn)
		result = append(result, p)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action) VALUES($1,'patients.list_viewed')`, a.ID); err != nil {
		return nil, err
	}
	return result, tx.Commit(ctx)
}
func (s Store) RegisterPatient(ctx context.Context, a domain.Actor, in domain.PatientInput) (domain.Patient, error) {
	var p domain.Patient
	p.PatientInput = in
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return p, err
	}
	defer tx.Rollback(ctx)
	var mrn int64
	err = tx.QueryRow(ctx, `INSERT INTO patient(given_name,family_name,date_of_birth,phone) VALUES($1,$2,NULLIF($3,'')::date,$4) RETURNING id,medical_record_number,created_at`, in.GivenName, in.FamilyName, in.DateOfBirth, in.Phone).Scan(&p.ID, &mrn, &p.CreatedAt)
	if err != nil {
		return p, err
	}
	p.MRN = fmt.Sprintf("HMS-%06d", mrn)
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'patient.registered',$2)`, a.ID, p.ID); err != nil {
		return p, err
	}
	return p, tx.Commit(ctx)
}
func (s Store) Overview(ctx context.Context, a domain.Actor) (domain.Overview, error) {
	var o domain.Overview
	err := s.DB.QueryRow(ctx, `SELECT 
		count(*),
		count(*) FILTER(WHERE created_at::date=CURRENT_DATE),
		COALESCE((SELECT SUM(total_minor) FROM invoice), 0),
		COALESCE((SELECT SUM(bed_charge_minor) FROM encounter WHERE status='active'), 0),
		COALESCE((SELECT SUM(amount_minor) FROM invoice_payment WHERE direction='payment'), 0),
		COALESCE((SELECT SUM(amount_minor) FROM patient_advance_payment), 0),
		COALESCE((SELECT COUNT(*) FROM hospital_bed WHERE active=true), 0),
		COALESCE((SELECT COUNT(*) FROM hospital_bed b WHERE b.active=true AND NOT EXISTS (SELECT 1 FROM encounter e WHERE e.bed_id=b.id AND e.status='active')), 0),
		COALESCE((SELECT COUNT(*) FROM encounter WHERE status='active' AND bed_id IS NOT NULL), 0),
		COALESCE((SELECT COUNT(*) FROM staff_access WHERE role='doctor' AND active=true), 0),
		COALESCE((SELECT COUNT(*) FROM patient WHERE canonical_patient_id(id)=id), 0),
		COALESCE((SELECT COUNT(*) FROM staff_access WHERE role='nurse' AND active=true), 0),
		COALESCE((SELECT COUNT(*) FROM staff_access WHERE role='admin' AND active=true), 0),
		COALESCE((SELECT COUNT(*) FROM staff_access WHERE role='accountant' AND active=true), 0),
		COALESCE((SELECT COUNT(*) FROM staff_access WHERE role='lab_technician' AND active=true), 0),
		COALESCE((SELECT COUNT(*) FROM staff_access WHERE role='pharmacist' AND active=true), 0),
		COALESCE((SELECT COUNT(*) FROM staff_access WHERE role='receptionist' AND active=true), 0)
	FROM patient WHERE canonical_patient_id(id)=id AND `+scope, a.Role, a.ID).Scan(
		&o.PatientCount,
		&o.RegisteredToday,
		&o.InvoicesMinor,
		&o.BillsMinor,
		&o.PaymentsMinor,
		&o.AdvancePaymentsMinor,
		&o.TotalBeds,
		&o.AvailableBeds,
		&o.OccupiedBeds,
		&o.Doctors,
		&o.Patients,
		&o.Nurses,
		&o.Admins,
		&o.Accountants,
		&o.LabTechnicians,
		&o.Pharmacists,
		&o.Receptionists,
	)
	return o, err
}
func (s Store) Enqueue(ctx context.Context, a domain.Actor, in domain.MessageInput, key string) (domain.Message, error) {
	var m domain.Message
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return m, err
	}
	defer tx.Rollback(ctx)
	// Same actor/key returns the original message; no second send is enqueued.
	err = tx.QueryRow(ctx, `INSERT INTO message_outbox(actor_id,channel,recipient,subject,body,idempotency_key) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(actor_id,idempotency_key) DO UPDATE SET idempotency_key=EXCLUDED.idempotency_key WHERE message_outbox.channel=EXCLUDED.channel AND message_outbox.recipient=EXCLUDED.recipient AND message_outbox.subject=EXCLUDED.subject AND message_outbox.body=EXCLUDED.body RETURNING id,channel,recipient,status,created_at`, a.ID, in.Channel, in.Recipient, in.Subject, in.Body, key).Scan(&m.ID, &m.Channel, &m.Recipient, &m.Status, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, domain.ErrConflict
	}
	if err != nil {
		return m, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'message.queued',$2)`, a.ID, m.ID); err != nil {
		return m, err
	}
	return m, tx.Commit(ctx)
}
func (s Store) Messages(ctx context.Context) ([]domain.Message, error) {
	rows, err := s.DB.Query(ctx, `SELECT id,channel,recipient,status,created_at FROM message_outbox WHERE audience='operational' ORDER BY created_at DESC LIMIT 25`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.Message{}
	for rows.Next() {
		var m domain.Message
		if err = rows.Scan(&m.ID, &m.Channel, &m.Recipient, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}
