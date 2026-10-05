package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

// --- General Settings ---

func (s Store) GeneralSettings(ctx context.Context) (map[string]string, error) {
	rows, err := s.DB.Query(ctx, `SELECT key, value FROM hospital_general_setting ORDER BY key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		settings[k] = v
	}
	return settings, nil
}

func (s Store) UpdateGeneralSetting(ctx context.Context, a domain.Actor, in domain.GeneralSettingInput) (domain.HospitalGeneralSetting, error) {
	var out domain.HospitalGeneralSetting
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_general_setting (key, value, updated_at)
		VALUES ($1, $2, clock_timestamp())
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = clock_timestamp()
		RETURNING key, value, updated_at
	`, in.Key, in.Value).Scan(&out.Key, &out.Value, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'general_setting.updated', $2)`, a.ID, out.Key); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Hospital Schedules ---

func (s Store) HospitalSchedules(ctx context.Context) ([]domain.HospitalScheduleDay, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, day_of_week, start_time, end_time, is_closed, created_at, updated_at
		FROM hospital_schedule_day
		ORDER BY day_of_week ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.HospitalScheduleDay{}
	for rows.Next() {
		var d domain.HospitalScheduleDay
		if err := rows.Scan(&d.ID, &d.DayOfWeek, &d.StartTime, &d.EndTime, &d.IsClosed, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (s Store) UpdateHospitalSchedule(ctx context.Context, a domain.Actor, in domain.HospitalScheduleDayInput) (domain.HospitalScheduleDay, error) {
	var out domain.HospitalScheduleDay
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO hospital_schedule_day (day_of_week, start_time, end_time, is_closed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, clock_timestamp(), clock_timestamp())
		ON CONFLICT (day_of_week) DO UPDATE
		SET start_time = EXCLUDED.start_time, end_time = EXCLUDED.end_time, is_closed = EXCLUDED.is_closed, updated_at = clock_timestamp()
		RETURNING id, day_of_week, start_time, end_time, is_closed, created_at, updated_at
	`, in.DayOfWeek, in.StartTime, in.EndTime, in.IsClosed).
		Scan(&out.ID, &out.DayOfWeek, &out.StartTime, &out.EndTime, &out.IsClosed, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'hospital_schedule.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Front CMS Settings ---

func (s Store) FrontCMSSettings(ctx context.Context, typeFilter string) ([]domain.FrontCMSSetting, error) {
	query := `SELECT key, value, type, updated_at FROM front_cms_setting`
	args := []any{}
	if typeFilter != "" {
		query += ` WHERE type = $1`
		args = append(args, typeFilter)
	}
	query += ` ORDER BY key ASC`

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.FrontCMSSetting{}
	for rows.Next() {
		var c domain.FrontCMSSetting
		if err := rows.Scan(&c.Key, &c.Value, &c.Type, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (s Store) UpdateFrontCMSSetting(ctx context.Context, a domain.Actor, in domain.FrontCMSSettingInput) (domain.FrontCMSSetting, error) {
	var out domain.FrontCMSSetting
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO front_cms_setting (key, value, type, updated_at)
		VALUES ($1, $2, $3, clock_timestamp())
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, type = EXCLUDED.type, updated_at = clock_timestamp()
		RETURNING key, value, type, updated_at
	`, in.Key, in.Value, in.Type).Scan(&out.Key, &out.Value, &out.Type, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'front_cms.updated', $2)`, a.ID, out.Key); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Testimonials ---

func (s Store) Testimonials(ctx context.Context, statusFilter *int) ([]domain.CMSTestimonial, error) {
	query := `SELECT id, name, description, position, rating, status, created_at, updated_at FROM cms_testimonial`
	args := []any{}
	if statusFilter != nil {
		query += ` WHERE status = $1`
		args = append(args, *statusFilter)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []domain.CMSTestimonial{}
	for rows.Next() {
		var t domain.CMSTestimonial
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Position, &t.Rating, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func (s Store) Testimonial(ctx context.Context, id string) (domain.CMSTestimonial, error) {
	var t domain.CMSTestimonial
	err := s.DB.QueryRow(ctx, `
		SELECT id, name, description, position, rating, status, created_at, updated_at
		FROM cms_testimonial
		WHERE id = $1
	`, id).Scan(&t.ID, &t.Name, &t.Description, &t.Position, &t.Rating, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return t, domain.ErrNotFound
	}
	return t, err
}

func (s Store) CreateTestimonial(ctx context.Context, a domain.Actor, in domain.CMSTestimonialInput) (domain.CMSTestimonial, error) {
	out := domain.CMSTestimonial{
		Name:        in.Name,
		Description: in.Description,
		Position:    in.Position,
		Rating:      in.Rating,
		Status:      in.Status,
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO cms_testimonial (name, description, position, rating, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, out.Name, out.Description, out.Position, out.Rating, out.Status).Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'testimonial.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateTestimonial(ctx context.Context, a domain.Actor, id string, in domain.CMSTestimonialInput) (domain.CMSTestimonial, error) {
	var out domain.CMSTestimonial
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE cms_testimonial
		SET name = $1, description = $2, position = $3, rating = $4, status = $5, updated_at = clock_timestamp()
		WHERE id = $6
		RETURNING id, name, description, position, rating, status, created_at, updated_at
	`, in.Name, in.Description, in.Position, in.Rating, in.Status, id).
		Scan(&out.ID, &out.Name, &out.Description, &out.Position, &out.Rating, &out.Status, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'testimonial.updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) DeleteTestimonial(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `DELETE FROM cms_testimonial WHERE id = $1`, id)
	if err != nil {
		return clinicalError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'testimonial.deleted', $2)`, a.ID, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
