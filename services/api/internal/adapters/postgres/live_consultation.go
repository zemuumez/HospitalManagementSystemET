package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
	"strings"
	"time"
)

// --- Live Consultations ---

func (s Store) LiveConsultations(ctx context.Context, doctorFilter, patientFilter string, statusFilter *int, page int) ([]domain.LiveConsultation, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	var conditions []string
	var args []any
	argIdx := 1

	if doctorFilter != "" {
		conditions = append(conditions, fmt.Sprintf("lc.doctor_id = $%d", argIdx))
		args = append(args, doctorFilter)
		argIdx++
	}
	if patientFilter != "" {
		conditions = append(conditions, fmt.Sprintf("(lc.patient_id = $%d OR lc.patient_id IN (SELECT id::text FROM patient WHERE patient_portal_owner(id) = $%d))", argIdx, argIdx))
		args = append(args, patientFilter)
		argIdx++
	}
	if statusFilter != nil {
		conditions = append(conditions, fmt.Sprintf("lc.status = $%d", argIdx))
		args = append(args, *statusFilter)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM live_consultation lc"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT lc.id, lc.doctor_id, lc.patient_id, lc.encounter_id, lc.consultation_title, lc.consultation_date,
		       lc.duration_minutes, lc.host_video, lc.participant_video, lc.type, lc.type_number, lc.platform_type,
		       lc.meeting_id, lc.password, lc.time_zone, lc.status, lc.description, lc.created_by, lc.created_at, lc.updated_at,
		       patient_portal_owner(p.id)
		FROM live_consultation lc
		LEFT JOIN patient p ON p.id::text = lc.patient_id
		%s
		ORDER BY lc.consultation_date DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.LiveConsultation{}
	for rows.Next() {
		var c domain.LiveConsultation
		var encID *string
		var patientUserID *string
		if err := rows.Scan(
			&c.ID, &c.DoctorID, &c.PatientID, &encID, &c.ConsultationTitle, &c.ConsultationDate,
			&c.DurationMinutes, &c.HostVideo, &c.ParticipantVideo, &c.Type, &c.TypeNumber, &c.PlatformType,
			&c.MeetingID, &c.Password, &c.TimeZone, &c.Status, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
			&patientUserID,
		); err != nil {
			return nil, 0, err
		}
		c.EncounterID = encID
		c.PatientUserID = patientUserID
		out = append(out, c)
	}
	return out, total, nil
}

func (s Store) LiveConsultation(ctx context.Context, id string) (domain.LiveConsultation, error) {
	var c domain.LiveConsultation
	var encID *string
	var patientUserID *string
	err := s.DB.QueryRow(ctx, `
		SELECT lc.id, lc.doctor_id, lc.patient_id, lc.encounter_id, lc.consultation_title, lc.consultation_date,
		       lc.duration_minutes, lc.host_video, lc.participant_video, lc.type, lc.type_number, lc.platform_type,
		       lc.meeting_id, lc.password, lc.time_zone, lc.status, lc.description, lc.created_by, lc.created_at, lc.updated_at,
		       patient_portal_owner(p.id)
		FROM live_consultation lc
		LEFT JOIN patient p ON p.id::text = lc.patient_id
		WHERE lc.id = $1
	`, id).Scan(
		&c.ID, &c.DoctorID, &c.PatientID, &encID, &c.ConsultationTitle, &c.ConsultationDate,
		&c.DurationMinutes, &c.HostVideo, &c.ParticipantVideo, &c.Type, &c.TypeNumber, &c.PlatformType,
		&c.MeetingID, &c.Password, &c.TimeZone, &c.Status, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
		&patientUserID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, domain.ErrNotFound
	}
	if err != nil {
		return c, err
	}
	c.EncounterID = encID
	c.PatientUserID = patientUserID
	return c, nil
}

func (s Store) CreateLiveConsultation(ctx context.Context, a domain.Actor, in domain.LiveConsultationInput) (domain.LiveConsultation, error) {
	out := domain.LiveConsultation{
		DoctorID:          in.DoctorID,
		PatientID:         in.PatientID,
		EncounterID:       in.EncounterID,
		ConsultationTitle: in.ConsultationTitle,
		ConsultationDate:  *in.ConsultationDate,
		DurationMinutes:   in.DurationMinutes,
		HostVideo:         in.HostVideo,
		ParticipantVideo:  in.ParticipantVideo,
		Type:              in.Type,
		TypeNumber:        in.TypeNumber,
		PlatformType:      in.PlatformType,
		MeetingID:         in.MeetingID,
		Password:          in.Password,
		TimeZone:          "Africa/Addis_Ababa",
		Status:            0,
		Description:       in.Description,
		CreatedBy:         a.ID,
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO live_consultation (
			doctor_id, patient_id, encounter_id, consultation_title, consultation_date,
			duration_minutes, host_video, participant_video, type, type_number, platform_type,
			meeting_id, password, time_zone, status, description, created_by
		)
		VALUES ($1, $2, $3::uuid, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 0, $15, $16)
		RETURNING id, created_at, updated_at
	`, out.DoctorID, out.PatientID, out.EncounterID, out.ConsultationTitle, out.ConsultationDate,
		out.DurationMinutes, out.HostVideo, out.ParticipantVideo, out.Type, out.TypeNumber, out.PlatformType,
		out.MeetingID, out.Password, out.TimeZone, out.Description, out.CreatedBy,
	).Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'live_consultation.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateLiveConsultationStatus(ctx context.Context, a domain.Actor, id string, status int) (domain.LiveConsultation, error) {
	var out domain.LiveConsultation
	var encID *string
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE live_consultation
		SET status = $1, updated_at = clock_timestamp()
		WHERE id = $2
		RETURNING id, doctor_id, patient_id, encounter_id, consultation_title, consultation_date,
		          duration_minutes, host_video, participant_video, type, type_number, platform_type,
		          meeting_id, password, time_zone, status, description, created_by, created_at, updated_at
	`, status, id).Scan(
		&out.ID, &out.DoctorID, &out.PatientID, &encID, &out.ConsultationTitle, &out.ConsultationDate,
		&out.DurationMinutes, &out.HostVideo, &out.ParticipantVideo, &out.Type, &out.TypeNumber, &out.PlatformType,
		&out.MeetingID, &out.Password, &out.TimeZone, &out.Status, &out.Description, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		return out, clinicalError(err)
	}
	out.EncounterID = encID

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'live_consultation.status_updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Live Meetings ---

func (s Store) LiveMeetings(ctx context.Context, statusFilter *int, page int) ([]domain.LiveMeeting, int, error) {
	if page < 1 {
		page = 1
	}
	limit := 25
	offset := (page - 1) * limit

	whereClause := ""
	args := []any{}
	if statusFilter != nil {
		whereClause = " WHERE status = $1"
		args = append(args, *statusFilter)
	}

	var total int
	err := s.DB.QueryRow(ctx, "SELECT count(*) FROM live_meeting"+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, title, meeting_date, duration_minutes, host_video, participant_video, platform_type,
		       meeting_id, password, time_zone, status, description, created_by, created_at, updated_at
		FROM live_meeting
		%s
		ORDER BY meeting_date DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []domain.LiveMeeting{}
	for rows.Next() {
		var m domain.LiveMeeting
		if err := rows.Scan(
			&m.ID, &m.Title, &m.MeetingDate, &m.DurationMinutes, &m.HostVideo, &m.ParticipantVideo, &m.PlatformType,
			&m.MeetingID, &m.Password, &m.TimeZone, &m.Status, &m.Description, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, nil
}

func (s Store) LiveMeeting(ctx context.Context, id string) (domain.LiveMeeting, error) {
	var m domain.LiveMeeting
	err := s.DB.QueryRow(ctx, `
		SELECT id, title, meeting_date, duration_minutes, host_video, participant_video, platform_type,
		       meeting_id, password, time_zone, status, description, created_by, created_at, updated_at
		FROM live_meeting
		WHERE id = $1
	`, id).Scan(
		&m.ID, &m.Title, &m.MeetingDate, &m.DurationMinutes, &m.HostVideo, &m.ParticipantVideo, &m.PlatformType,
		&m.MeetingID, &m.Password, &m.TimeZone, &m.Status, &m.Description, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, domain.ErrNotFound
	}
	if err != nil {
		return m, err
	}

	// Fetch candidate IDs
	cRows, err := s.DB.Query(ctx, `SELECT user_id FROM live_meeting_candidate WHERE live_meeting_id = $1`, id)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var uID string
			if err := cRows.Scan(&uID); err == nil {
				m.CandidateUserIDs = append(m.CandidateUserIDs, uID)
			}
		}
	}
	return m, nil
}

func (s Store) CreateLiveMeeting(ctx context.Context, a domain.Actor, in domain.LiveMeetingInput) (domain.LiveMeeting, error) {
	out := domain.LiveMeeting{
		Title:            in.Title,
		MeetingDate:      *in.MeetingDate,
		DurationMinutes:  in.DurationMinutes,
		HostVideo:        in.HostVideo,
		ParticipantVideo: in.ParticipantVideo,
		PlatformType:     in.PlatformType,
		MeetingID:        in.MeetingID,
		Password:         in.Password,
		TimeZone:         "Africa/Addis_Ababa",
		Status:           0,
		Description:      in.Description,
		CreatedBy:        a.ID,
		CandidateUserIDs: in.CandidateUserIDs,
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO live_meeting (
			title, meeting_date, duration_minutes, host_video, participant_video, platform_type,
			meeting_id, password, time_zone, status, description, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 0, $10, $11)
		RETURNING id, created_at, updated_at
	`, out.Title, out.MeetingDate, out.DurationMinutes, out.HostVideo, out.ParticipantVideo,
		out.PlatformType, out.MeetingID, out.Password, out.TimeZone, out.Description, out.CreatedBy,
	).Scan(&out.ID, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return out, clinicalError(err)
	}

	for _, candID := range in.CandidateUserIDs {
		candID = strings.TrimSpace(candID)
		if candID != "" {
			if _, err = tx.Exec(ctx, `
				INSERT INTO live_meeting_candidate (live_meeting_id, user_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, out.ID, candID); err != nil {
				return out, clinicalError(err)
			}
		}
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'live_meeting.created', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

func (s Store) UpdateLiveMeetingStatus(ctx context.Context, a domain.Actor, id string, status int) (domain.LiveMeeting, error) {
	var out domain.LiveMeeting
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		UPDATE live_meeting
		SET status = $1, updated_at = clock_timestamp()
		WHERE id = $2
		RETURNING id, title, meeting_date, duration_minutes, host_video, participant_video, platform_type,
		          meeting_id, password, time_zone, status, description, created_by, created_at, updated_at
	`, status, id).Scan(
		&out.ID, &out.Title, &out.MeetingDate, &out.DurationMinutes, &out.HostVideo, &out.ParticipantVideo, &out.PlatformType,
		&out.MeetingID, &out.Password, &out.TimeZone, &out.Status, &out.Description, &out.CreatedBy, &out.CreatedAt, &out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, domain.ErrNotFound
	}
	if err != nil {
		return out, clinicalError(err)
	}

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'live_meeting.status_updated', $2)`, a.ID, out.ID); err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

// --- Provider Settings ---

func (s Store) ProviderSetting(ctx context.Context, userID string) (domain.LiveProviderSetting, error) {
	var p domain.LiveProviderSetting
	var hasKey, hasSecret bool
	err := s.DB.QueryRow(ctx, `
		SELECT user_id, platform_type, (api_key != ''), (api_secret != ''), updated_at
		FROM live_consultation_provider_setting
		WHERE user_id = $1
	`, userID).Scan(&p.UserID, &p.PlatformType, &hasKey, &hasSecret, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LiveProviderSetting{
			UserID:       userID,
			PlatformType: "zoom",
			HasAPIKey:    false,
			HasAPISecret: false,
			UpdatedAt:    time.Now(),
		}, nil
	}
	if err != nil {
		return p, err
	}
	p.HasAPIKey = hasKey
	p.HasAPISecret = hasSecret
	return p, nil
}

func (s Store) UpdateProviderSetting(ctx context.Context, a domain.Actor, in domain.LiveProviderSettingInput) (domain.LiveProviderSetting, error) {
	var p domain.LiveProviderSetting
	var hasKey, hasSecret bool
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return p, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO live_consultation_provider_setting (user_id, platform_type, api_key, api_secret, updated_at)
		VALUES ($1, $2, $3, $4, clock_timestamp())
		ON CONFLICT (user_id) DO UPDATE
		SET platform_type = EXCLUDED.platform_type, api_key = EXCLUDED.api_key, api_secret = EXCLUDED.api_secret, updated_at = clock_timestamp()
		RETURNING user_id, platform_type, (api_key != ''), (api_secret != ''), updated_at
	`, a.ID, in.PlatformType, in.APIKey, in.APISecret).Scan(&p.UserID, &p.PlatformType, &hasKey, &hasSecret, &p.UpdatedAt)
	if err != nil {
		return p, clinicalError(err)
	}
	p.HasAPIKey = hasKey
	p.HasAPISecret = hasSecret

	if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'live_provider_setting.updated', $2)`, a.ID, a.ID); err != nil {
		return p, err
	}
	return p, tx.Commit(ctx)
}
