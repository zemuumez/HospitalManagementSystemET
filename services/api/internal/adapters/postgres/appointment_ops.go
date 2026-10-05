package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

// ─── Patient Queue ────────────────────────────────────────────────────────────

func (s Store) DoctorQueue(ctx context.Context, doctorID string, queueDate string) ([]domain.PatientQueueItem, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT q.id, q.doctor_id, u.name AS doctor_name, q.patient_id,
		       (p.given_name || ' ' || p.family_name) AS patient_name,
		       q.appointment_id::text, q.queue_date::text, q.token_number,
		       q.status, q.notes, q.created_at, q.called_at, q.completed_at
		FROM patient_queue q
		JOIN "user" u ON u.id = q.doctor_id
		JOIN patient p ON p.id = q.patient_id
		WHERE q.doctor_id = $1 AND q.queue_date = $2::date
		ORDER BY q.token_number ASC
	`, doctorID, queueDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.PatientQueueItem
	for rows.Next() {
		var item domain.PatientQueueItem
		var apptID *string
		if err := rows.Scan(
			&item.ID, &item.DoctorID, &item.DoctorName, &item.PatientID,
			&item.PatientName, &apptID, &item.QueueDate, &item.TokenNumber,
			&item.Status, &item.Notes, &item.CreatedAt, &item.CalledAt, &item.CompletedAt,
		); err != nil {
			return nil, err
		}
		item.AppointmentID = apptID
		items = append(items, item)
	}
	if items == nil {
		items = []domain.PatientQueueItem{}
	}
	return items, nil
}

func (s Store) EnqueuePatient(ctx context.Context, a domain.Actor, input domain.EnqueuePatientInput) (domain.PatientQueueItem, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.PatientQueueItem{}, err
	}
	defer tx.Rollback(ctx)

	// The transaction lock covers the empty queue as well as existing rows.
	// Row-locking only existing queue entries cannot serialize the first token.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended(json_build_array('patient-queue', $1::text, $2::text)::text, 0))`, input.DoctorID, input.QueueDate); err != nil {
		return domain.PatientQueueItem{}, err
	}

	var valid bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM patient p JOIN staff_access d ON d.user_id=$2 AND d.role='doctor' AND d.active WHERE p.id=$1 AND ($3::uuid IS NULL OR EXISTS(SELECT 1 FROM appointment ap WHERE ap.id=$3 AND ap.patient_id=p.id AND ap.doctor_id=$2 AND (ap.starts_at AT TIME ZONE 'Africa/Addis_Ababa')::date=$4::date AND ap.status IN ('booked','arrived'))))`, input.PatientID, input.DoctorID, input.AppointmentID, input.QueueDate).Scan(&valid)
	if err != nil {
		return domain.PatientQueueItem{}, err
	}
	if !valid {
		return domain.PatientQueueItem{}, domain.ErrValidation
	}
	// Check if already in queue today
	var existingID string
	err = tx.QueryRow(ctx, `
		SELECT id FROM patient_queue
		WHERE doctor_id = $1 AND queue_date = $2::date AND patient_id = $3
	`, input.DoctorID, input.QueueDate, input.PatientID).Scan(&existingID)
	if err == nil {
		return domain.PatientQueueItem{}, domain.ErrConflict
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.PatientQueueItem{}, err
	}

	// Calculate next sequential token for this doctor and date
	var nextToken int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(token_number), 0) + 1
		FROM patient_queue
		WHERE doctor_id = $1 AND queue_date = $2::date
	`, input.DoctorID, input.QueueDate).Scan(&nextToken)
	if err != nil {
		return domain.PatientQueueItem{}, err
	}

	var item domain.PatientQueueItem
	item.DoctorID = input.DoctorID
	item.PatientID = input.PatientID
	item.AppointmentID = input.AppointmentID
	item.QueueDate = input.QueueDate
	item.TokenNumber = nextToken
	item.Status = "waiting"
	item.Notes = input.Notes

	err = tx.QueryRow(ctx, `
		INSERT INTO patient_queue (
			doctor_id, patient_id, appointment_id, queue_date, token_number, status, notes
		) VALUES ($1, $2, $3, $4::date, $5, $6, $7)
		RETURNING id, created_at
	`, input.DoctorID, input.PatientID, input.AppointmentID, input.QueueDate, nextToken, item.Status, input.Notes,
	).Scan(&item.ID, &item.CreatedAt)
	if err != nil {
		return domain.PatientQueueItem{}, err
	}

	// Fetch doctor name and patient name
	_ = tx.QueryRow(ctx, `SELECT name FROM "user" WHERE id = $1`, input.DoctorID).Scan(&item.DoctorName)
	_ = tx.QueryRow(ctx, `SELECT (given_name || ' ' || family_name) FROM patient WHERE id = $1`, input.PatientID).Scan(&item.PatientName)

	if err := tx.Commit(ctx); err != nil {
		return domain.PatientQueueItem{}, err
	}
	return item, nil
}

func (s Store) UpdateQueueStatus(ctx context.Context, a domain.Actor, queueID string, input domain.UpdateQueueStatusInput) (domain.PatientQueueItem, error) {
	now := time.Now().UTC()
	var calledAt, completedAt *time.Time
	if input.Status == "in_consultation" {
		calledAt = &now
	} else if input.Status == "completed" || input.Status == "skipped" {
		completedAt = &now
	}

	var item domain.PatientQueueItem
	var apptID *string
	err := s.DB.QueryRow(ctx, `
		UPDATE patient_queue
		SET status = $1,
		    notes = CASE WHEN $2 <> '' THEN $2 ELSE notes END,
		    called_at = COALESCE($3, called_at),
		    completed_at = COALESCE($4, completed_at)
		WHERE id = $5 AND ($6 IN ('admin','receptionist') OR doctor_id=$7)
 AND (status=$1 OR (status='waiting' AND $1 IN ('in_consultation','skipped')) OR (status='in_consultation' AND $1 IN ('completed','skipped')))
		RETURNING id, doctor_id, patient_id, appointment_id::text, queue_date::text, token_number,
		          status, notes, created_at, called_at, completed_at
	`, input.Status, input.Notes, calledAt, completedAt, queueID, a.Role, a.ID).Scan(
		&item.ID, &item.DoctorID, &item.PatientID, &apptID, &item.QueueDate, &item.TokenNumber,
		&item.Status, &item.Notes, &item.CreatedAt, &item.CalledAt, &item.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PatientQueueItem{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PatientQueueItem{}, err
	}
	item.AppointmentID = apptID

	_ = s.DB.QueryRow(ctx, `SELECT name FROM "user" WHERE id = $1`, item.DoctorID).Scan(&item.DoctorName)
	_ = s.DB.QueryRow(ctx, `SELECT (given_name || ' ' || family_name) FROM patient WHERE id = $1`, item.PatientID).Scan(&item.PatientName)
	return item, nil
}

// ─── Public Appointment Requests ──────────────────────────────────────────────

func (s Store) CreatePublicAppointmentRequest(ctx context.Context, input domain.CreatePublicAppointmentRequestInput) (domain.PublicAppointmentRequest, error) {
	var req domain.PublicAppointmentRequest
	req.PatientName = input.PatientName
	req.PatientEmail = input.PatientEmail
	req.PatientPhone = input.PatientPhone
	req.DoctorID = input.DoctorID
	req.DepartmentID = input.DepartmentID
	req.PreferredDate = input.PreferredDate
	req.Problem = input.Problem
	req.Status = "pending"

	err := s.DB.QueryRow(ctx, `
		INSERT INTO public_appointment_request (
			patient_name, patient_email, patient_phone, doctor_id, department_id,
			preferred_date, problem, status
		) VALUES ($1, $2, $3, $4, $5, $6::date, $7, $8)
		RETURNING id, created_at
	`, input.PatientName, input.PatientEmail, input.PatientPhone, input.DoctorID, input.DepartmentID,
		input.PreferredDate, input.Problem, req.Status,
	).Scan(&req.ID, &req.CreatedAt)
	if err != nil {
		return domain.PublicAppointmentRequest{}, err
	}
	return req, nil
}

func (s Store) ListPublicAppointmentRequests(ctx context.Context, status string, page int) ([]domain.PublicAppointmentRequest, error) {
	offset := (page - 1) * 25
	rows, err := s.DB.Query(ctx, `
		SELECT r.id, r.patient_name, r.patient_email, r.patient_phone,
		       r.doctor_id, u.name AS doctor_name, r.department_id::text,
		       COALESCE(d.title, '') AS department_title,
		       r.preferred_date::text, r.problem, r.status, r.rejection_reason,
		       r.appointment_id::text, r.reviewed_by, r.reviewed_at, r.created_at
		FROM public_appointment_request r
		JOIN "user" u ON u.id = r.doctor_id
		LEFT JOIN doctor_department d ON d.id = r.department_id
		WHERE ($1 = '' OR r.status = $1)
		ORDER BY r.created_at DESC
		LIMIT 25 OFFSET $2
	`, status, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.PublicAppointmentRequest
	for rows.Next() {
		var req domain.PublicAppointmentRequest
		var deptID, apptID *string
		if err := rows.Scan(
			&req.ID, &req.PatientName, &req.PatientEmail, &req.PatientPhone,
			&req.DoctorID, &req.DoctorName, &deptID,
			&req.DepartmentTitle,
			&req.PreferredDate, &req.Problem, &req.Status, &req.RejectionReason,
			&apptID, &req.ReviewedBy, &req.ReviewedAt, &req.CreatedAt,
		); err != nil {
			return nil, err
		}
		req.DepartmentID = deptID
		req.AppointmentID = apptID
		list = append(list, req)
	}
	if list == nil {
		list = []domain.PublicAppointmentRequest{}
	}
	return list, nil
}

func (s Store) ReviewPublicAppointmentRequest(ctx context.Context, a domain.Actor, id string, input domain.ReviewPublicAppointmentRequestInput) (domain.PublicAppointmentRequest, error) {
	now := time.Now().UTC()
	var newStatus string
	var apptID *string
	if input.Action == "confirm" {
		newStatus = "confirmed"
		if input.AppointmentID != "" {
			apptID = &input.AppointmentID
		}
	} else {
		newStatus = "rejected"
	}

	var req domain.PublicAppointmentRequest
	var deptID, scannedApptID *string
	err := s.DB.QueryRow(ctx, `
		UPDATE public_appointment_request
		SET status = $1,
		    rejection_reason = $2,
		    appointment_id = $3,
		    reviewed_by = $4,
		    reviewed_at = $5
		WHERE id = $6 AND status = 'pending' AND ($1='rejected' OR EXISTS(SELECT 1 FROM appointment a WHERE a.id=$3 AND a.doctor_id=public_appointment_request.doctor_id AND (a.starts_at AT TIME ZONE 'Africa/Addis_Ababa')::date=public_appointment_request.preferred_date AND a.status IN ('booked','arrived')))
		RETURNING id, patient_name, patient_email, patient_phone, doctor_id, department_id::text,
		          preferred_date::text, problem, status, rejection_reason, appointment_id::text,
		          reviewed_by, reviewed_at, created_at
	`, newStatus, input.RejectionReason, apptID, a.ID, now, id).Scan(
		&req.ID, &req.PatientName, &req.PatientEmail, &req.PatientPhone, &req.DoctorID, &deptID,
		&req.PreferredDate, &req.Problem, &req.Status, &req.RejectionReason, &scannedApptID,
		&req.ReviewedBy, &req.ReviewedAt, &req.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PublicAppointmentRequest{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PublicAppointmentRequest{}, err
	}
	req.DepartmentID = deptID
	req.AppointmentID = scannedApptID
	return req, nil
}

// ─── Appointment Billing ──────────────────────────────────────────────────────

func (s Store) AppointmentBilling(ctx context.Context, appointmentID string) (domain.AppointmentBillingRecord, error) {
	var rec domain.AppointmentBillingRecord
	var invoiceID *string
	err := s.DB.QueryRow(ctx, `
		SELECT appointment_id::text, fee_minor, invoice_id::text, CASE WHEN invoice_id IS NULL THEN payment_status WHEN EXISTS(SELECT 1 FROM invoice i WHERE i.id=invoice_id AND i.paid_minor>=i.total_minor) THEN 'paid' ELSE 'unpaid' END, created_at, updated_at
		FROM appointment_billing
		WHERE appointment_id = $1
	`, appointmentID).Scan(
		&rec.AppointmentID, &rec.FeeMinor, &invoiceID, &rec.PaymentStatus, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// If not explicitly set, fetch doctor appointment charge
		var charge float64
		_ = s.DB.QueryRow(ctx, `
			SELECT COALESCE(dp.appointment_charge, 0)
			FROM appointment a
			LEFT JOIN doctor_profile dp ON dp.user_id = a.doctor_id
			WHERE a.id = $1
		`, appointmentID).Scan(&charge)
		return domain.AppointmentBillingRecord{
			AppointmentID: appointmentID,
			FeeMinor:      int64(charge * 100),
			PaymentStatus: "unpaid",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}, nil
	}
	if err != nil {
		return domain.AppointmentBillingRecord{}, clinicalError(err)
	}
	rec.InvoiceID = invoiceID
	return rec, nil
}

func (s Store) SetAppointmentFee(ctx context.Context, a domain.Actor, appointmentID string, feeMinor int64) (domain.AppointmentBillingRecord, error) {
	var rec domain.AppointmentBillingRecord
	var invoiceID *string
	err := s.DB.QueryRow(ctx, `
		INSERT INTO appointment_billing (appointment_id, fee_minor, payment_status, created_at, updated_at)
		VALUES ($1, $2, 'unpaid', clock_timestamp(), clock_timestamp())
		ON CONFLICT (appointment_id) DO UPDATE SET
			fee_minor = EXCLUDED.fee_minor,
			updated_at = clock_timestamp()
 WHERE appointment_billing.invoice_id IS NULL
		RETURNING appointment_id::text, fee_minor, invoice_id::text, payment_status, created_at, updated_at
	`, appointmentID, feeMinor).Scan(
		&rec.AppointmentID, &rec.FeeMinor, &invoiceID, &rec.PaymentStatus, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		return domain.AppointmentBillingRecord{}, clinicalError(err)
	}
	rec.InvoiceID = invoiceID
	return rec, nil
}

func (s Store) LinkAppointmentInvoice(ctx context.Context, a domain.Actor, appointmentID string, invoiceID string) (domain.AppointmentBillingRecord, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.AppointmentBillingRecord{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = linkOperationalInvoice(ctx, tx, a, "appointment", appointmentID, invoiceID, nil); err != nil {
		return domain.AppointmentBillingRecord{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE appointment_billing SET invoice_id=$2,updated_at=clock_timestamp() WHERE appointment_id=$1`, appointmentID, invoiceID); err != nil {
		return domain.AppointmentBillingRecord{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.AppointmentBillingRecord{}, err
	}
	return s.AppointmentBilling(ctx, appointmentID)
}
