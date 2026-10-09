package postgres

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

var tokenRegex = regexp.MustCompile(`[a-f0-9]{32}`)

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

type lockedAttachment struct {
	id          string
	storagePath string
	isPublic    bool
	isClinical  bool
}

// lockAndValidateTokensTx locks all candidate tokens in secure_attachment in stable alphabetical order.
// Newly bound tokens are validated to guarantee they exist and are appropriate public nonclinical assets.
func lockAndValidateTokensTx(ctx context.Context, tx pgx.Tx, boundTokens map[string]bool, displacedTokens map[string]bool) (map[string]lockedAttachment, error) {
	allTokensMap := make(map[string]bool)
	for tok := range boundTokens {
		allTokensMap[tok] = true
	}
	for tok := range displacedTokens {
		allTokensMap[tok] = true
	}
	if len(allTokensMap) == 0 {
		return nil, nil
	}
	allTokens := make([]string, 0, len(allTokensMap))
	for tok := range allTokensMap {
		allTokens = append(allTokens, tok)
	}
	sort.Strings(allTokens) // Stable alphabetical order prevents deadlocks

	locked := make(map[string]lockedAttachment, len(allTokens))
	for _, tok := range allTokens {
		var id, storagePath string
		var isPublic bool
		var patientID, encounterID *string
		err := tx.QueryRow(ctx, `
			SELECT id, storage_path, is_public, patient_id::text, encounter_id::text
			FROM secure_attachment
			WHERE token = $1
			FOR UPDATE
		`, tok).Scan(&id, &storagePath, &isPublic, &patientID, &encounterID)
		if errors.Is(err, pgx.ErrNoRows) {
			if boundTokens[tok] {
				// Attempting to bind a non-existent or retired token must fail without saving a broken URL
				return nil, domain.ErrNotFound
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		isClinical := (!isPublic || patientID != nil || encounterID != nil)
		if boundTokens[tok] && isClinical {
			// Clinical attachments cannot be bound to public hospital settings
			return nil, domain.ErrValidation
		}
		locked[tok] = lockedAttachment{
			id:          id,
			storagePath: storagePath,
			isPublic:    isPublic,
			isClinical:  isClinical,
		}
	}
	return locked, nil
}

// retireLockedDisplacedTokensTx checks references under held row locks and safely deletes unreferenced displaced public assets.
func retireLockedDisplacedTokensTx(ctx context.Context, tx pgx.Tx, a domain.Actor, displacedTokens map[string]bool, locked map[string]lockedAttachment) ([]string, error) {
	if len(displacedTokens) == 0 {
		return nil, nil
	}
	var displacedList []string
	for tok := range displacedTokens {
		displacedList = append(displacedList, tok)
	}
	sort.Strings(displacedList)

	var removedPaths []string
	for _, tok := range displacedList {
		att, exists := locked[tok]
		if !exists || att.isClinical {
			continue // Already deleted or strictly protected clinical record
		}
		var inUse bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM hospital_general_setting WHERE value LIKE '%' || $1 || '%'
			) OR EXISTS(
				SELECT 1 FROM front_cms_setting WHERE value LIKE '%' || $1 || '%'
			)
		`, tok).Scan(&inUse)
		if err != nil {
			return nil, err
		}
		if inUse {
			continue // Preserved due to active reference in another setting or CMS
		}
		if _, err = tx.Exec(ctx, `DELETE FROM secure_attachment WHERE id = $1`, att.id); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO audit_event (actor_id, action, resource_id)
			VALUES ($1, 'attachment.retired', $2)
		`, a.ID, tok); err != nil {
			return nil, err
		}
		removedPaths = append(removedPaths, att.storagePath)
	}
	return removedPaths, nil
}

func (s Store) UpdateGeneralSetting(ctx context.Context, a domain.Actor, in domain.GeneralSettingInput) (domain.HospitalGeneralSetting, error) {
	if !a.Can("settings.manage") {
		return domain.HospitalGeneralSetting{}, domain.ErrForbidden
	}
	_, err := s.UpdateGeneralSettings(ctx, a, []domain.GeneralSettingInput{in})
	if err != nil {
		return domain.HospitalGeneralSetting{}, err
	}
	var out domain.HospitalGeneralSetting
	err = s.DB.QueryRow(ctx, `SELECT key, value, updated_at FROM hospital_general_setting WHERE key = $1`, in.Key).Scan(&out.Key, &out.Value, &out.UpdatedAt)
	if err != nil {
		return domain.HospitalGeneralSetting{}, err
	}
	return out, nil
}

func (s Store) UpdateGeneralSettings(ctx context.Context, a domain.Actor, inputs []domain.GeneralSettingInput) ([]string, error) {
	if !a.Can("settings.manage") {
		return nil, domain.ErrForbidden
	}
	if len(inputs) == 0 {
		return nil, nil
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Sort input keys to ensure consistent row-lock order
	inputsCopy := make([]domain.GeneralSettingInput, len(inputs))
	copy(inputsCopy, inputs)
	sort.Slice(inputsCopy, func(i, j int) bool { return inputsCopy[i].Key < inputsCopy[j].Key })

	boundTokens := make(map[string]bool)
	allOldTokens := make(map[string]bool)
	allNewTokens := make(map[string]bool)

	for _, in := range inputsCopy {
		var prevVal string
		err := tx.QueryRow(ctx, `SELECT value FROM hospital_general_setting WHERE key = $1 FOR UPDATE`, in.Key).Scan(&prevVal)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		oldTokens := tokenRegex.FindAllString(prevVal, -1)
		for _, t := range oldTokens {
			allOldTokens[t] = true
		}
		if !domain.SettingsSecretKeys[in.Key] {
			newTokens := tokenRegex.FindAllString(in.Value, -1)
			for _, t := range newTokens {
				allNewTokens[t] = true
				boundTokens[t] = true
			}
		}
	}

	displacedTokens := make(map[string]bool)
	for t := range allOldTokens {
		if !allNewTokens[t] {
			displacedTokens[t] = true
		}
	}

	// 1. Lock all incoming and displaced tokens in stable alphabetical order and validate bound assets
	locked, err := lockAndValidateTokensTx(ctx, tx, boundTokens, displacedTokens)
	if err != nil {
		return nil, err
	}

	// 2. Persist updated setting values
	for _, in := range inputsCopy {
		if _, err = tx.Exec(ctx, `
			INSERT INTO hospital_general_setting(key, value, updated_at)
			VALUES($1, $2, clock_timestamp())
			ON CONFLICT(key) DO UPDATE
			SET value = EXCLUDED.value, updated_at = clock_timestamp()
		`, in.Key, in.Value); err != nil {
			return nil, clinicalError(err)
		}
		if err = pharmacyAudit(ctx, tx, a, "general_setting.updated", in.Key); err != nil {
			return nil, err
		}
	}

	// 3. Under the held row locks, retire unreferenced displaced tokens
	removedPaths, err := retireLockedDisplacedTokensTx(ctx, tx, a, displacedTokens, locked)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return removedPaths, nil
}

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

func (s Store) UpdateHospitalSchedules(ctx context.Context, a domain.Actor, days []domain.HospitalScheduleDayInput) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, in := range days {
		var id string
		err = tx.QueryRow(ctx, `
			INSERT INTO hospital_schedule_day (day_of_week, start_time, end_time, is_closed, created_at, updated_at)
			VALUES ($1, $2, $3, $4, clock_timestamp(), clock_timestamp())
			ON CONFLICT (day_of_week) DO UPDATE
			SET start_time = EXCLUDED.start_time, end_time = EXCLUDED.end_time, is_closed = EXCLUDED.is_closed, updated_at = clock_timestamp()
			RETURNING id
		`, in.DayOfWeek, in.StartTime, in.EndTime, in.IsClosed).Scan(&id)
		if err != nil {
			return clinicalError(err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'hospital_schedule.updated', $2)`, a.ID, id); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
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
	if !a.Can("cms.manage") {
		return domain.FrontCMSSetting{}, domain.ErrForbidden
	}
	_, err := s.UpdateFrontCMSSettings(ctx, a, []domain.FrontCMSSettingInput{in})
	if err != nil {
		return domain.FrontCMSSetting{}, err
	}
	var out domain.FrontCMSSetting
	err = s.DB.QueryRow(ctx, `SELECT key, value, type, updated_at FROM front_cms_setting WHERE key = $1`, in.Key).Scan(&out.Key, &out.Value, &out.Type, &out.UpdatedAt)
	if err != nil {
		return domain.FrontCMSSetting{}, err
	}
	return out, nil
}

func (s Store) UpdateFrontCMSSettings(ctx context.Context, a domain.Actor, settings []domain.FrontCMSSettingInput) ([]string, error) {
	if !a.Can("cms.manage") {
		return nil, domain.ErrForbidden
	}
	if len(settings) == 0 {
		return nil, nil
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Sort input keys to guarantee stable row-lock order
	settingsCopy := make([]domain.FrontCMSSettingInput, len(settings))
	copy(settingsCopy, settings)
	sort.Slice(settingsCopy, func(i, j int) bool { return settingsCopy[i].Key < settingsCopy[j].Key })

	boundTokens := make(map[string]bool)
	allOldTokens := make(map[string]bool)
	allNewTokens := make(map[string]bool)

	for _, in := range settingsCopy {
		var prevVal string
		err := tx.QueryRow(ctx, `SELECT value FROM front_cms_setting WHERE key = $1 FOR UPDATE`, in.Key).Scan(&prevVal)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
		oldTokens := tokenRegex.FindAllString(prevVal, -1)
		for _, t := range oldTokens {
			allOldTokens[t] = true
		}
		newTokens := tokenRegex.FindAllString(in.Value, -1)
		for _, t := range newTokens {
			allNewTokens[t] = true
			boundTokens[t] = true
		}
	}

	displacedTokens := make(map[string]bool)
	for t := range allOldTokens {
		if !allNewTokens[t] {
			displacedTokens[t] = true
		}
	}

	// 1. Lock all incoming and displaced tokens in stable alphabetical order and validate bound assets
	locked, err := lockAndValidateTokensTx(ctx, tx, boundTokens, displacedTokens)
	if err != nil {
		return nil, err
	}

	// 2. Persist updated settings values
	for _, in := range settingsCopy {
		_, err = tx.Exec(ctx, `
			INSERT INTO front_cms_setting (key, value, type, updated_at)
			VALUES ($1, $2, $3, clock_timestamp())
			ON CONFLICT (key) DO UPDATE
			SET value = EXCLUDED.value, type = EXCLUDED.type, updated_at = clock_timestamp()
		`, in.Key, in.Value, in.Type)
		if err != nil {
			return nil, clinicalError(err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO audit_event (actor_id, action, resource_id) VALUES ($1, 'front_cms.updated', $2)`, a.ID, in.Key); err != nil {
			return nil, err
		}
	}

	// 3. Under the held row locks, retire unreferenced displaced tokens
	removedPaths, err := retireLockedDisplacedTokensTx(ctx, tx, a, displacedTokens, locked)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return removedPaths, nil
}

func (s Store) CleanupAbandonedAttachments(ctx context.Context, a domain.Actor, olderThan time.Duration) ([]string, error) {
	if !a.Can("settings.manage") {
		return nil, domain.ErrForbidden
	}
	intervalSec := int(olderThan.Seconds())
	if intervalSec <= 0 {
		intervalSec = 86400 // default 24h
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Bounded work per run (LIMIT 100) and stable order (ORDER BY token ASC)
	rows, err := tx.Query(ctx, `
		SELECT id, token, storage_path, is_public, patient_id::text, encounter_id::text
		FROM secure_attachment
		WHERE is_public = true
		  AND patient_id IS NULL
		  AND encounter_id IS NULL
		  AND created_at < clock_timestamp() - make_interval(secs => $1)
		ORDER BY token ASC
		LIMIT 100
		FOR UPDATE SKIP LOCKED
	`, intervalSec)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type cand struct {
		id          string
		token       string
		storagePath string
		isPublic    bool
		patientID   *string
		encounterID *string
	}
	var candidates []cand
	for rows.Next() {
		var c cand
		if err := rows.Scan(&c.id, &c.token, &c.storagePath, &c.isPublic, &c.patientID, &c.encounterID); err != nil {
			return nil, err
		}
		candidates = append(candidates, c)
	}
	rows.Close()

	var removedPaths []string
	for _, c := range candidates {
		// Strictly protect clinical attachments
		if !c.isPublic || c.patientID != nil || c.encounterID != nil {
			continue
		}
		// Under the token's held row lock, check if still referenced
		var inUse bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM hospital_general_setting WHERE value LIKE '%' || $1 || '%'
			) OR EXISTS(
				SELECT 1 FROM front_cms_setting WHERE value LIKE '%' || $1 || '%'
			)
		`, c.token).Scan(&inUse)
		if err != nil {
			return nil, err
		}
		if inUse {
			continue // Preserved due to active reference in general settings or CMS
		}
		if _, err = tx.Exec(ctx, `DELETE FROM secure_attachment WHERE id = $1`, c.id); err != nil {
			return nil, err
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO audit_event (actor_id, action, resource_id)
			VALUES ($1, 'attachment.retired', $2)
		`, a.ID, c.token); err != nil {
			return nil, err
		}
		removedPaths = append(removedPaths, c.storagePath)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return removedPaths, nil
}

func (s Store) RetireUnreferencedAttachments(ctx context.Context, a domain.Actor) ([]string, error) {
	return s.CleanupAbandonedAttachments(ctx, a, 24*time.Hour)
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
