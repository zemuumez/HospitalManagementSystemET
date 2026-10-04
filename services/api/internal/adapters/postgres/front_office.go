package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

// --- Complaints ---

func (s Store) Complaints(ctx context.Context, patientID string, page int) ([]domain.HospitalComplaint, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	whereClause := ""
	args := []any{}
	if patientID != "" {
		whereClause = " WHERE patient_id = $1"
		args = append(args, patientID)
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM hospital_complaint"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, patient_id, title, description, status, response, resolved_by, resolved_at, created_at, updated_at
		FROM hospital_complaint
		%s
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.HospitalComplaint{}
	for rows.Next() {
		var c domain.HospitalComplaint
		if err := rows.Scan(&c.ID, &c.PatientID, &c.Title, &c.Description, &c.Status, &c.Response, &c.ResolvedBy, &c.ResolvedAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, nil
}

func (s Store) Complaint(ctx context.Context, id string) (domain.HospitalComplaint, error) {
	var c domain.HospitalComplaint
	err := s.DB.QueryRow(ctx, `
		SELECT id, patient_id, title, description, status, response, resolved_by, resolved_at, created_at, updated_at
		FROM hospital_complaint
		WHERE id = $1
	`, id).Scan(&c.ID, &c.PatientID, &c.Title, &c.Description, &c.Status, &c.Response, &c.ResolvedBy, &c.ResolvedAt, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, domain.ErrNotFound
	}
	return c, err
}

func (s Store) CreateComplaint(ctx context.Context, a domain.Actor, in domain.ComplaintCreateInput) (domain.HospitalComplaint, error) {
	out := domain.HospitalComplaint{
		PatientID:   in.PatientID,
		Title:       in.Title,
		Description: in.Description,
		Status:      0,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_complaint (patient_id, title, description, status)
		VALUES ($1, $2, $3, 0)
		RETURNING id, created_at, updated_at
	`, out.PatientID, out.Title, out.Description).Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'complaint.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) ResolveComplaint(ctx context.Context, a domain.Actor, id string, in domain.ComplaintResolveInput) (domain.HospitalComplaint, error) {
	var out domain.HospitalComplaint
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE hospital_complaint
		SET status = $1, response = $2, resolved_by = $3, resolved_at = clock_timestamp(), updated_at = clock_timestamp()
		WHERE id = $4
		RETURNING id, patient_id, title, description, status, response, resolved_by, resolved_at, created_at, updated_at
	`, in.Status, in.Response, a.ID, id).
		Scan(&out.ID, &out.PatientID, &out.Title, &out.Description, &out.Status, &out.Response, &out.ResolvedBy, &out.ResolvedAt, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'complaint.resolved', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Notice Board ---

func (s Store) Notices(ctx context.Context) ([]domain.HospitalNoticeBoard, error) {
	rows, err := s.DB.Query(ctx, `SELECT id, title, description, created_at, updated_at FROM hospital_notice_board ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.HospitalNoticeBoard{}
	for rows.Next() {
		var n domain.HospitalNoticeBoard
		if err := rows.Scan(&n.ID, &n.Title, &n.Description, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}

func (s Store) CreateNotice(ctx context.Context, a domain.Actor, in domain.NoticeBoardInput) (domain.HospitalNoticeBoard, error) {
	out := domain.HospitalNoticeBoard{
		Title:       in.Title,
		Description: in.Description,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_notice_board (title, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`, out.Title, out.Description).Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'notice.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) DeleteNotice(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `DELETE FROM hospital_notice_board WHERE id = $1`, id)
	if err != nil {
		return clinicalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'notice.deleted', $2)`, a.ID, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// --- Enquiries ---

func (s Store) Enquiries(ctx context.Context, page int) ([]domain.HospitalEnquiry, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM hospital_enquiry").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.DB.Query(ctx, fmt.Sprintf(`
		SELECT id, full_name, email, contact_no, type, message, viewed_by, status, created_at, updated_at
		FROM hospital_enquiry
		ORDER BY created_at DESC
		LIMIT %d OFFSET %d
	`, limit, offset))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.HospitalEnquiry{}
	for rows.Next() {
		var e domain.HospitalEnquiry
		if err := rows.Scan(&e.ID, &e.FullName, &e.Email, &e.ContactNo, &e.Type, &e.Message, &e.ViewedBy, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, nil
}

func (s Store) SubmitEnquiry(ctx context.Context, in domain.EnquiryInput) (domain.HospitalEnquiry, error) {
	out := domain.HospitalEnquiry{
		FullName:  in.FullName,
		Email:     in.Email,
		ContactNo: in.ContactNo,
		Type:      in.Type,
		Message:   in.Message,
		Status:    0,
	}
	err := s.DB.QueryRow(ctx, `
		INSERT INTO hospital_enquiry (full_name, email, contact_no, type, message, status)
		VALUES ($1, $2, $3, $4, $5, 0)
		RETURNING id, created_at, updated_at
	`, out.FullName, out.Email, out.ContactNo, out.Type, out.Message).Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	return out, clinicalError(err)
}

func (s Store) MarkEnquiryRead(ctx context.Context, a domain.Actor, id string) (domain.HospitalEnquiry, error) {
	var out domain.HospitalEnquiry
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE hospital_enquiry
		SET status = 1, viewed_by = $1, updated_at = clock_timestamp()
		WHERE id = $2
		RETURNING id, full_name, email, contact_no, type, message, viewed_by, status, created_at, updated_at
	`, a.ID, id).Scan(&out.ID, &out.FullName, &out.Email, &out.ContactNo, &out.Type, &out.Message, &out.ViewedBy, &out.Status, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'enquiry.read', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Visitors ---

func (s Store) Visitors(ctx context.Context, page int, date string) ([]domain.HospitalVisitor, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	whereClause := ""
	args := []any{}
	if date != "" {
		whereClause = " WHERE date = $1"
		args = append(args, date)
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM hospital_visitor"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.DB.Query(ctx, fmt.Sprintf(`
		SELECT id, purpose, name, phone, id_card, no_of_person, date, in_time, out_time, note, created_at, updated_at
		FROM hospital_visitor
		%s
		ORDER BY date DESC, in_time DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.HospitalVisitor{}
	for rows.Next() {
		var v domain.HospitalVisitor
		var d time.Time
		if err := rows.Scan(&v.ID, &v.Purpose, &v.Name, &v.Phone, &v.IDCard, &v.NoOfPerson, &d, &v.InTime, &v.OutTime, &v.Note, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, err
		}
		v.Date = d.Format("2006-01-02")
		out = append(out, v)
	}
	return out, total, nil
}

func (s Store) CreateVisitor(ctx context.Context, a domain.Actor, in domain.VisitorInput) (domain.HospitalVisitor, error) {
	out := domain.HospitalVisitor{
		Purpose:    in.Purpose,
		Name:       in.Name,
		Phone:      in.Phone,
		IDCard:     in.IDCard,
		NoOfPerson: in.NoOfPerson,
		Date:       in.Date,
		InTime:     in.InTime,
		OutTime:    in.OutTime,
		Note:       in.Note,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_visitor (purpose, name, phone, id_card, no_of_person, date, in_time, out_time, note)
		VALUES ($1, $2, $3, $4, $5, $6::date, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`, out.Purpose, out.Name, out.Phone, out.IDCard, out.NoOfPerson, out.Date, out.InTime, out.OutTime, out.Note).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'visitor.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateVisitor(ctx context.Context, a domain.Actor, id string, in domain.VisitorInput) (domain.HospitalVisitor, error) {
	var out domain.HospitalVisitor
	var d time.Time
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE hospital_visitor
		SET purpose = $1, name = $2, phone = $3, id_card = $4, no_of_person = $5, date = $6::date, in_time = $7, out_time = $8, note = $9, updated_at = clock_timestamp()
		WHERE id = $10
		RETURNING id, purpose, name, phone, id_card, no_of_person, date, in_time, out_time, note, created_at, updated_at
	`, in.Purpose, in.Name, in.Phone, in.IDCard, in.NoOfPerson, in.Date, in.InTime, in.OutTime, in.Note, id).
		Scan(&out.ID, &out.Purpose, &out.Name, &out.Phone, &out.IDCard, &out.NoOfPerson, &d, &out.InTime, &out.OutTime, &out.Note, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		return out, clinicalError(err)
	}
	out.Date = d.Format("2006-01-02")

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'visitor.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Call Logs ---

func (s Store) CallLogs(ctx context.Context, page int, callType int) ([]domain.HospitalCallLog, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	whereClause := ""
	args := []any{}
	if callType > 0 {
		whereClause = " WHERE call_type = $1"
		args = append(args, callType)
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM hospital_call_log"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.DB.Query(ctx, fmt.Sprintf(`
		SELECT id, name, phone, date, follow_up_date, note, call_type, created_at, updated_at
		FROM hospital_call_log
		%s
		ORDER BY date DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.HospitalCallLog{}
	for rows.Next() {
		var l domain.HospitalCallLog
		var d time.Time
		var fud *time.Time
		if err := rows.Scan(&l.ID, &l.Name, &l.Phone, &d, &fud, &l.Note, &l.CallType, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, 0, err
		}
		l.Date = d.Format("2006-01-02")
		if fud != nil {
			formatted := fud.Format("2006-01-02")
			l.FollowUpDate = &formatted
		}
		out = append(out, l)
	}
	return out, total, nil
}

func (s Store) CreateCallLog(ctx context.Context, a domain.Actor, in domain.CallLogInput) (domain.HospitalCallLog, error) {
	out := domain.HospitalCallLog{
		Name:         in.Name,
		Phone:        in.Phone,
		Date:         in.Date,
		FollowUpDate: in.FollowUpDate,
		Note:         in.Note,
		CallType:     in.CallType,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_call_log (name, phone, date, follow_up_date, note, call_type)
		VALUES ($1, $2, $3::date, CASE WHEN $4::text = '' THEN NULL ELSE $4::date END, $5, $6)
		RETURNING id, created_at, updated_at
	`, out.Name, out.Phone, out.Date, stringVal(out.FollowUpDate), out.Note, out.CallType).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'call_log.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateCallLog(ctx context.Context, a domain.Actor, id string, in domain.CallLogInput) (domain.HospitalCallLog, error) {
	var out domain.HospitalCallLog
	var d time.Time
	var fud *time.Time
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE hospital_call_log
		SET name = $1, phone = $2, date = $3::date, follow_up_date = CASE WHEN $4::text = '' THEN NULL ELSE $4::date END, note = $5, call_type = $6, updated_at = clock_timestamp()
		WHERE id = $7
		RETURNING id, name, phone, date, follow_up_date, note, call_type, created_at, updated_at
	`, in.Name, in.Phone, in.Date, stringVal(in.FollowUpDate), in.Note, in.CallType, id).
		Scan(&out.ID, &out.Name, &out.Phone, &d, &fud, &out.Note, &out.CallType, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		return out, clinicalError(err)
	}
	out.Date = d.Format("2006-01-02")
	if fud != nil {
		formatted := fud.Format("2006-01-02")
		out.FollowUpDate = &formatted
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'call_log.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Postals ---

func (s Store) Postals(ctx context.Context, page int, postalType int) ([]domain.HospitalPostal, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	whereClause := ""
	args := []any{}
	if postalType > 0 {
		whereClause = " WHERE type = $1"
		args = append(args, postalType)
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM hospital_postal"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.DB.Query(ctx, fmt.Sprintf(`
		SELECT id, from_title, to_title, reference_no, date, address, type, created_at, updated_at
		FROM hospital_postal
		%s
		ORDER BY date DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.HospitalPostal{}
	for rows.Next() {
		var p domain.HospitalPostal
		var d time.Time
		if err := rows.Scan(&p.ID, &p.FromTitle, &p.ToTitle, &p.ReferenceNo, &d, &p.Address, &p.Type, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		p.Date = d.Format("2006-01-02")
		out = append(out, p)
	}
	return out, total, nil
}

func (s Store) CreatePostal(ctx context.Context, a domain.Actor, in domain.PostalInput) (domain.HospitalPostal, error) {
	out := domain.HospitalPostal{
		FromTitle:   in.FromTitle,
		ToTitle:     in.ToTitle,
		ReferenceNo: in.ReferenceNo,
		Date:        in.Date,
		Address:     in.Address,
		Type:        in.Type,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_postal (from_title, to_title, reference_no, date, address, type)
		VALUES ($1, $2, $3, $4::date, $5, $6)
		RETURNING id, created_at, updated_at
	`, out.FromTitle, out.ToTitle, out.ReferenceNo, out.Date, out.Address, out.Type).
		Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'postal.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdatePostal(ctx context.Context, a domain.Actor, id string, in domain.PostalInput) (domain.HospitalPostal, error) {
	var out domain.HospitalPostal
	var d time.Time
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE hospital_postal
		SET from_title = $1, to_title = $2, reference_no = $3, date = $4::date, address = $5, type = $6, updated_at = clock_timestamp()
		WHERE id = $7
		RETURNING id, from_title, to_title, reference_no, date, address, type, created_at, updated_at
	`, in.FromTitle, in.ToTitle, in.ReferenceNo, in.Date, in.Address, in.Type, id).
		Scan(&out.ID, &out.FromTitle, &out.ToTitle, &out.ReferenceNo, &d, &out.Address, &out.Type, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		return out, clinicalError(err)
	}
	out.Date = d.Format("2006-01-02")

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'postal.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func stringVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
