package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

// ─── Doctor departments ───────────────────────────────────────────────────────

func (s Store) DoctorDepartments(ctx context.Context, includeArchived bool) ([]domain.DoctorDepartment, error) {
	q := `SELECT id, title, description, archived, version, created_at, updated_at
	      FROM doctor_department`
	if !includeArchived {
		q += ` WHERE NOT archived`
	}
	q += ` ORDER BY title ASC`
	rows, err := s.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DoctorDepartment{}
	for rows.Next() {
		var d domain.DoctorDepartment
		if err := rows.Scan(&d.ID, &d.Title, &d.Description, &d.Archived, &d.Version, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (s Store) DoctorDepartment(ctx context.Context, id string) (domain.DoctorDepartment, error) {
	var d domain.DoctorDepartment
	err := s.DB.QueryRow(ctx, `
		SELECT id, title, description, archived, version, created_at, updated_at
		FROM doctor_department WHERE id = $1`, id).
		Scan(&d.ID, &d.Title, &d.Description, &d.Archived, &d.Version, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, domain.ErrNotFound
	}
	return d, err
}

func (s Store) CreateDoctorDepartment(ctx context.Context, a domain.Actor, i domain.DoctorDepartmentInput) (domain.DoctorDepartment, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.DoctorDepartment{}, err
	}
	defer tx.Rollback(ctx)

	var d domain.DoctorDepartment
	err = tx.QueryRow(ctx, `
		INSERT INTO doctor_department (title, description)
		VALUES ($1, $2)
		RETURNING id, title, description, archived, version, created_at, updated_at`,
		i.Title, i.Description).
		Scan(&d.ID, &d.Title, &d.Description, &d.Archived, &d.Version, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return d, domain.ErrConflict
		}
		return d, err
	}

	beforeSnap, _ := json.Marshal(domain.DoctorDepartment{})
	afterSnap, _ := json.Marshal(d)
	if _, err = tx.Exec(ctx, `
		INSERT INTO doctor_department_revision (department_id, version, actor_id, reason, before_snapshot, after_snapshot)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		d.ID, d.Version, a.ID, "created", beforeSnap, afterSnap); err != nil {
		return d, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_department.created', $2)`, a.ID, d.ID); err != nil {
		return d, err
	}
	return d, tx.Commit(ctx)
}

func (s Store) UpdateDoctorDepartment(ctx context.Context, a domain.Actor, id string, i domain.DoctorDepartmentInput) (domain.DoctorDepartment, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.DoctorDepartment{}, err
	}
	defer tx.Rollback(ctx)

	var before domain.DoctorDepartment
	err = tx.QueryRow(ctx, `
		SELECT id, title, description, archived, version, created_at, updated_at
		FROM doctor_department WHERE id = $1 FOR UPDATE`, id).
		Scan(&before.ID, &before.Title, &before.Description, &before.Archived, &before.Version, &before.CreatedAt, &before.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return before, domain.ErrNotFound
	}
	if err != nil {
		return before, err
	}
	if before.Archived {
		return before, domain.ErrStale
	}
	if before.Version != i.Version {
		return before, domain.ErrStale
	}

	var after domain.DoctorDepartment
	err = tx.QueryRow(ctx, `
		UPDATE doctor_department SET title=$1, description=$2, version=version+1, updated_at=now()
		WHERE id=$3
		RETURNING id, title, description, archived, version, created_at, updated_at`,
		i.Title, i.Description, id).
		Scan(&after.ID, &after.Title, &after.Description, &after.Archived, &after.Version, &after.CreatedAt, &after.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return after, domain.ErrConflict
		}
		return after, err
	}

	beforeSnap, _ := json.Marshal(before)
	afterSnap, _ := json.Marshal(after)
	if _, err = tx.Exec(ctx, `
		INSERT INTO doctor_department_revision (department_id, version, actor_id, reason, before_snapshot, after_snapshot)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		after.ID, after.Version, a.ID, i.Reason, beforeSnap, afterSnap); err != nil {
		return after, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_department.updated', $2)`, a.ID, after.ID); err != nil {
		return after, err
	}
	return after, tx.Commit(ctx)
}

func (s Store) ArchiveDoctorDepartment(ctx context.Context, a domain.Actor, id string, version int, reason string) (domain.DoctorDepartment, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.DoctorDepartment{}, err
	}
	defer tx.Rollback(ctx)

	var before domain.DoctorDepartment
	err = tx.QueryRow(ctx, `
		SELECT id, title, description, archived, version, created_at, updated_at
		FROM doctor_department WHERE id = $1 FOR UPDATE`, id).
		Scan(&before.ID, &before.Title, &before.Description, &before.Archived, &before.Version, &before.CreatedAt, &before.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return before, domain.ErrNotFound
	}
	if err != nil {
		return before, err
	}
	if before.Archived {
		return before, domain.ErrStale
	}
	if before.Version != version {
		return before, domain.ErrStale
	}

	var after domain.DoctorDepartment
	err = tx.QueryRow(ctx, `
		UPDATE doctor_department SET archived=true, version=version+1, updated_at=now()
		WHERE id=$1
		RETURNING id, title, description, archived, version, created_at, updated_at`, id).
		Scan(&after.ID, &after.Title, &after.Description, &after.Archived, &after.Version, &after.CreatedAt, &after.UpdatedAt)
	if err != nil {
		return after, err
	}

	beforeSnap, _ := json.Marshal(before)
	afterSnap, _ := json.Marshal(after)
	if _, err = tx.Exec(ctx, `
		INSERT INTO doctor_department_revision (department_id, version, actor_id, reason, before_snapshot, after_snapshot)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		after.ID, after.Version, a.ID, reason, beforeSnap, afterSnap); err != nil {
		return after, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_department.archived', $2)`, a.ID, after.ID); err != nil {
		return after, err
	}
	return after, tx.Commit(ctx)
}

func (s Store) DoctorDepartmentRevisions(ctx context.Context, a domain.Actor, id string, page int) ([]domain.DoctorDepartmentRevision, error) {
	limit := 25
	offset := (page - 1) * limit
	rows, err := s.DB.Query(ctx, `
		SELECT r.version, r.actor_id, r.reason, r.before_snapshot, r.after_snapshot, r.created_at
		FROM doctor_department_revision r
		WHERE r.department_id = $1
		ORDER BY r.version DESC
		LIMIT $2 OFFSET $3`, id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.DoctorDepartmentRevision{}
	for rows.Next() {
		var r domain.DoctorDepartmentRevision
		var beforeRaw, afterRaw []byte
		if err := rows.Scan(&r.Version, &r.ActorID, &r.Reason, &beforeRaw, &afterRaw, &r.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(beforeRaw, &r.Before)
		_ = json.Unmarshal(afterRaw, &r.After)
		out = append(out, r)
	}
	return out, nil
}

// ─── Doctor profile extensions ────────────────────────────────────────────────

func (s Store) DoctorExt(ctx context.Context, doctorID string) (domain.DoctorExt, error) {
	var d domain.DoctorExt
	err := s.DB.QueryRow(ctx, `
		SELECT p.user_id,
		       COALESCE(p.department_id::text, ''),
		       COALESCE(dd.title, ''),
		       COALESCE(p.description, ''),
		       COALESCE(p.photo_url, ''),
		       COALESCE(p.opd_charge, 0),
		       COALESCE(p.appointment_charge, 0),
		       p.version
		FROM doctor_profile p
		LEFT JOIN doctor_department dd ON dd.id = p.department_id
		WHERE p.user_id = $1`, doctorID).
		Scan(&d.DoctorID, &d.DepartmentID, &d.DepartmentTitle, &d.Description, &d.PhotoURL, &d.OpdCharge, &d.AppointmentCharge, &d.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, domain.ErrNotFound
	}
	return d, err
}

func (s Store) SaveDoctorExt(ctx context.Context, a domain.Actor, doctorID string, i domain.DoctorExtInput) (domain.DoctorExt, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.DoctorExt{}, err
	}
	defer tx.Rollback(ctx)

	var curVersion int
	err = tx.QueryRow(ctx, `SELECT version FROM doctor_profile WHERE user_id = $1 FOR UPDATE`, doctorID).Scan(&curVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DoctorExt{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.DoctorExt{}, err
	}
	if curVersion != i.Version {
		return domain.DoctorExt{}, domain.ErrStale
	}

	var deptID *string
	if i.DepartmentID != "" {
		deptID = &i.DepartmentID
	}

	_, err = tx.Exec(ctx, `
		UPDATE doctor_profile
		SET department_id=$1, description=$2, photo_url=$3,
		    opd_charge=$4, appointment_charge=$5, version=version+1
		WHERE user_id=$6`,
		deptID, i.Description, i.PhotoURL, i.OpdCharge, i.AppointmentCharge, doctorID)
	if err != nil {
		return domain.DoctorExt{}, err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'doctor_ext.updated', $2)`, a.ID, doctorID); err != nil {
		return domain.DoctorExt{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.DoctorExt{}, err
	}
	return s.DoctorExt(ctx, doctorID)
}

// ─── Hospital opening hours ───────────────────────────────────────────────────

func (s Store) HospitalHours(ctx context.Context) ([]domain.HospitalHoursEntry, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT weekday, open_minute, close_minute
		FROM hospital_hours ORDER BY weekday ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.HospitalHoursEntry{}
	for rows.Next() {
		var e domain.HospitalHoursEntry
		if err := rows.Scan(&e.Weekday, &e.OpenMinute, &e.CloseMinute); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (s Store) SaveHospitalHours(ctx context.Context, a domain.Actor, i domain.HospitalHoursInput) ([]domain.HospitalHoursEntry, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, `DELETE FROM hospital_hours`); err != nil {
		return nil, err
	}
	for _, h := range i.Hours {
		if _, err = tx.Exec(ctx, `
			INSERT INTO hospital_hours (weekday, open_minute, close_minute, updated_at)
			VALUES ($1, $2, $3, now())`,
			h.Weekday, h.OpenMinute, h.CloseMinute); err != nil {
			return nil, err
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'hospital_hours.updated', $1)`, a.ID); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.HospitalHours(ctx)
}

// ─── Hospital date overrides ──────────────────────────────────────────────────

func (s Store) HospitalDateOverrides(ctx context.Context, page int) ([]domain.HospitalDateOverride, error) {
	limit := 25
	offset := (page - 1) * limit
	rows, err := s.DB.Query(ctx, `
		SELECT id, override_date::text, closed, open_minute, close_minute, label, actor_id, created_at
		FROM hospital_date_override
		ORDER BY override_date DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.HospitalDateOverride{}
	for rows.Next() {
		var o domain.HospitalDateOverride
		if err := rows.Scan(&o.ID, &o.OverrideDate, &o.Closed, &o.OpenMinute, &o.CloseMinute, &o.Label, &o.ActorID, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

func (s Store) CreateHospitalDateOverride(ctx context.Context, a domain.Actor, i domain.HospitalDateOverrideInput) (domain.HospitalDateOverride, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.HospitalDateOverride{}, err
	}
	defer tx.Rollback(ctx)

	var o domain.HospitalDateOverride
	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_date_override (override_date, closed, open_minute, close_minute, label, actor_id)
		VALUES ($1::date, $2, $3, $4, $5, $6)
		RETURNING id, override_date::text, closed, open_minute, close_minute, label, actor_id, created_at`,
		i.OverrideDate, i.Closed, i.OpenMinute, i.CloseMinute, i.Label, a.ID).
		Scan(&o.ID, &o.OverrideDate, &o.Closed, &o.OpenMinute, &o.CloseMinute, &o.Label, &o.ActorID, &o.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return o, domain.ErrConflict
		}
		return o, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'hospital_date_override.created', $2)`, a.ID, o.ID); err != nil {
		return o, err
	}
	return o, tx.Commit(ctx)
}

func (s Store) DeleteHospitalDateOverride(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `DELETE FROM hospital_date_override WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'hospital_date_override.deleted', $2)`, a.ID, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ─── Patient profile extensions ───────────────────────────────────────────────

func (s Store) PatientProfileExt(ctx context.Context, patientID string) (domain.PatientProfileExt, error) {
	var e domain.PatientProfileExt
	err := s.DB.QueryRow(ctx, `
		SELECT COALESCE(father_name, ''), COALESCE(religion, ''),
		       COALESCE(referral_source, ''), COALESCE(notes, '')
		FROM patient_profile WHERE patient_id = $1`, patientID).
		Scan(&e.FatherName, &e.Religion, &e.ReferralSource, &e.Notes)
	if errors.Is(err, pgx.ErrNoRows) {
		return e, domain.ErrNotFound
	}
	return e, err
}

func (s Store) SavePatientProfileExt(ctx context.Context, a domain.Actor, patientID string, i domain.PatientProfileExt) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE patient_profile
		SET father_name=$1, religion=$2, referral_source=$3, notes=$4
		WHERE patient_id=$5`,
		i.FatherName, i.Religion, i.ReferralSource, i.Notes, patientID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'patient_profile_ext.updated', $2)`, a.ID, patientID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
