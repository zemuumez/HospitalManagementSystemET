package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

// ─── Patient Contact Consent ──────────────────────────────────────────────────

func (s Store) PatientContactConsent(ctx context.Context, patientID string) (domain.PatientContactConsent, error) {
	var c domain.PatientContactConsent
	c.PatientID = patientID
	err := s.DB.QueryRow(ctx, `
		SELECT guardian_name, guardian_relation, guardian_phone, guardian_email,
		       sms_consent, email_consent, data_sharing_consent, updated_at
		FROM patient_contact_consent
		WHERE patient_id = canonical_patient_id($1::uuid)
	`, patientID).Scan(
		&c.GuardianName, &c.GuardianRelation, &c.GuardianPhone, &c.GuardianEmail,
		&c.SmsConsent, &c.EmailConsent, &c.DataSharingConsent, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// Return default consent record
		c.SmsConsent = false
		c.EmailConsent = false
		c.DataSharingConsent = false
		c.UpdatedAt = time.Now().UTC()
		return c, nil
	}
	if err != nil {
		return domain.PatientContactConsent{}, err
	}
	return c, nil
}

func (s Store) SavePatientContactConsent(ctx context.Context, a domain.Actor, c domain.PatientContactConsent) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO patient_contact_consent (
			patient_id, guardian_name, guardian_relation, guardian_phone, guardian_email,
			sms_consent, email_consent, data_sharing_consent, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, clock_timestamp())
		ON CONFLICT (patient_id) DO UPDATE SET
			guardian_name = EXCLUDED.guardian_name,
			guardian_relation = EXCLUDED.guardian_relation,
			guardian_phone = EXCLUDED.guardian_phone,
			guardian_email = EXCLUDED.guardian_email,
			sms_consent = EXCLUDED.sms_consent,
			email_consent = EXCLUDED.email_consent,
			data_sharing_consent = EXCLUDED.data_sharing_consent,
			updated_at = clock_timestamp()
	`, c.PatientID, c.GuardianName, c.GuardianRelation, c.GuardianPhone, c.GuardianEmail,
		c.SmsConsent, c.EmailConsent, c.DataSharingConsent)
	return err
}

// ─── ID Card Templates ────────────────────────────────────────────────────────

func (s Store) ListTemplates(ctx context.Context) ([]domain.PatientIDCardTemplate, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, color, email, phone, dob, blood_group, address, patient_unique_id, created_at, updated_at
		FROM patient_id_card_template
		ORDER BY created_at ASC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.PatientIDCardTemplate
	for rows.Next() {
		var t domain.PatientIDCardTemplate
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Color, &t.Email, &t.Phone, &t.DOB,
			&t.BloodGroup, &t.Address, &t.PatientUniqueID, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	if list == nil {
		list = []domain.PatientIDCardTemplate{}
	}
	return list, nil
}

func (s Store) GetTemplate(ctx context.Context, id string) (domain.PatientIDCardTemplate, error) {
	var t domain.PatientIDCardTemplate
	err := s.DB.QueryRow(ctx, `
		SELECT id, name, color, email, phone, dob, blood_group, address, patient_unique_id, created_at, updated_at
		FROM patient_id_card_template
		WHERE id = $1
	`, id).Scan(
		&t.ID, &t.Name, &t.Color, &t.Email, &t.Phone, &t.DOB,
		&t.BloodGroup, &t.Address, &t.PatientUniqueID, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PatientIDCardTemplate{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PatientIDCardTemplate{}, err
	}
	return t, nil
}

func (s Store) CreateTemplate(ctx context.Context, a domain.Actor, in domain.CreatePatientIDCardTemplateInput) (domain.PatientIDCardTemplate, error) {
	var dup bool
	err := s.DB.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM patient_id_card_template WHERE lower(name) = lower($1))
	`, in.Name).Scan(&dup)
	if err != nil {
		return domain.PatientIDCardTemplate{}, err
	}
	if dup {
		return domain.PatientIDCardTemplate{}, domain.ErrConflict
	}

	email := true
	if in.Email != nil {
		email = *in.Email
	}
	phone := true
	if in.Phone != nil {
		phone = *in.Phone
	}
	dob := true
	if in.DOB != nil {
		dob = *in.DOB
	}
	bloodGroup := true
	if in.BloodGroup != nil {
		bloodGroup = *in.BloodGroup
	}
	address := true
	if in.Address != nil {
		address = *in.Address
	}
	patientUniqueID := true
	if in.PatientUniqueID != nil {
		patientUniqueID = *in.PatientUniqueID
	}

	var out domain.PatientIDCardTemplate
	err = s.DB.QueryRow(ctx, `
		INSERT INTO patient_id_card_template (
			name, color, email, phone, dob, blood_group, address, patient_unique_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, color, email, phone, dob, blood_group, address, patient_unique_id, created_at, updated_at
	`, in.Name, in.Color, email, phone, dob, bloodGroup, address, patientUniqueID).Scan(
		&out.ID, &out.Name, &out.Color, &out.Email, &out.Phone, &out.DOB,
		&out.BloodGroup, &out.Address, &out.PatientUniqueID, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate key") {
			return domain.PatientIDCardTemplate{}, domain.ErrConflict
		}
		return domain.PatientIDCardTemplate{}, err
	}
	return out, nil
}

func (s Store) UpdateTemplate(ctx context.Context, a domain.Actor, id string, in domain.UpdatePatientIDCardTemplateInput) (domain.PatientIDCardTemplate, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.PatientIDCardTemplate{}, err
	}
	defer tx.Rollback(ctx)

	var cur domain.PatientIDCardTemplate
	err = tx.QueryRow(ctx, `
		SELECT id, name, color, email, phone, dob, blood_group, address, patient_unique_id
		FROM patient_id_card_template
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(
		&cur.ID, &cur.Name, &cur.Color, &cur.Email, &cur.Phone, &cur.DOB,
		&cur.BloodGroup, &cur.Address, &cur.PatientUniqueID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PatientIDCardTemplate{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PatientIDCardTemplate{}, err
	}

	name := cur.Name
	if in.Name != nil {
		name = *in.Name
		var dup bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM patient_id_card_template WHERE lower(name) = lower($1) AND id <> $2)
		`, name, id).Scan(&dup)
		if err != nil {
			return domain.PatientIDCardTemplate{}, err
		}
		if dup {
			return domain.PatientIDCardTemplate{}, domain.ErrConflict
		}
	}

	color := cur.Color
	if in.Color != nil {
		color = *in.Color
	}
	email := cur.Email
	if in.Email != nil {
		email = *in.Email
	}
	phone := cur.Phone
	if in.Phone != nil {
		phone = *in.Phone
	}
	dob := cur.DOB
	if in.DOB != nil {
		dob = *in.DOB
	}
	bloodGroup := cur.BloodGroup
	if in.BloodGroup != nil {
		bloodGroup = *in.BloodGroup
	}
	address := cur.Address
	if in.Address != nil {
		address = *in.Address
	}
	patientUniqueID := cur.PatientUniqueID
	if in.PatientUniqueID != nil {
		patientUniqueID = *in.PatientUniqueID
	}

	var out domain.PatientIDCardTemplate
	err = tx.QueryRow(ctx, `
		UPDATE patient_id_card_template
		SET name = $1, color = $2, email = $3, phone = $4, dob = $5, blood_group = $6,
		    address = $7, patient_unique_id = $8, updated_at = clock_timestamp()
		WHERE id = $9
		RETURNING id, name, color, email, phone, dob, blood_group, address, patient_unique_id, created_at, updated_at
	`, name, color, email, phone, dob, bloodGroup, address, patientUniqueID, id).Scan(
		&out.ID, &out.Name, &out.Color, &out.Email, &out.Phone, &out.DOB,
		&out.BloodGroup, &out.Address, &out.PatientUniqueID, &out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return domain.PatientIDCardTemplate{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.PatientIDCardTemplate{}, err
	}
	return out, nil
}

func (s Store) DeleteTemplate(ctx context.Context, a domain.Actor, id string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Check if referenced by any patient or active smart card
	var inUse bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM patient WHERE smart_card_template_id = $1::uuid
			UNION ALL
			SELECT 1 FROM patient_smart_card WHERE template_id = $1::uuid AND status = 'active'
		)
	`, id).Scan(&inUse)
	if err != nil {
		return err
	}
	if inUse {
		return domain.ErrInUse
	}

	tag, err := tx.Exec(ctx, `DELETE FROM patient_id_card_template WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return tx.Commit(ctx)
}

// ─── Patient Smart Cards ──────────────────────────────────────────────────────

func (s Store) PatientSmartCards(ctx context.Context, patientID string) ([]domain.PatientSmartCard, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT sc.id, sc.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       COALESCE(p.patient_unique_id, '') AS patient_unique_id,
		       p.medical_record_number::text AS mrn, COALESCE(pp.blood_group, '') AS blood_group,
		       sc.card_number, sc.qr_token, sc.template_name, COALESCE(sc.template_id::text, '') AS template_id,
		       sc.status, sc.revoked_reason, sc.issued_at, sc.expires_at, sc.issued_by
		FROM patient_smart_card sc
		JOIN patient p ON p.id = sc.patient_id
		LEFT JOIN patient_profile pp ON pp.patient_id = sc.patient_id
		WHERE canonical_patient_id(sc.patient_id) = canonical_patient_id($1::uuid)
		ORDER BY sc.issued_at DESC
	`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []domain.PatientSmartCard
	for rows.Next() {
		var c domain.PatientSmartCard
		if err := rows.Scan(
			&c.ID, &c.PatientID, &c.PatientName, &c.PatientUniqueID, &c.MRN, &c.BloodGroup,
			&c.CardNumber, &c.QRToken, &c.TemplateName, &c.TemplateID, &c.Status, &c.RevokedReason,
			&c.IssuedAt, &c.ExpiresAt, &c.IssuedBy,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	if cards == nil {
		cards = []domain.PatientSmartCard{}
	}
	return cards, nil
}

func (s Store) IssueSmartCard(ctx context.Context, a domain.Actor, card domain.PatientSmartCard) (domain.PatientSmartCard, error) {
	err := s.DB.QueryRow(ctx, `
		INSERT INTO patient_smart_card (
			patient_id, card_number, qr_token, template_name, status, expires_at, issued_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, issued_at
	`, card.PatientID, card.CardNumber, card.QRToken, card.TemplateName, card.Status, card.ExpiresAt, card.IssuedBy,
	).Scan(&card.ID, &card.IssuedAt)
	if err != nil {
		return domain.PatientSmartCard{}, err
	}
	return card, nil
}

func (s Store) RevokeSmartCard(ctx context.Context, a domain.Actor, cardID string, reason string) error {
	tag, err := s.DB.Exec(ctx, `
		UPDATE patient_smart_card
		SET status = 'revoked', revoked_reason = $1
		WHERE id = $2 AND status = 'active'
	`, reason, cardID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s Store) GenerateSmartCards(ctx context.Context, a domain.Actor, in domain.GenerateSmartCardsInput) (domain.GenerateSmartCardsResult, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return domain.GenerateSmartCardsResult{}, err
	}
	defer tx.Rollback(ctx)

	// Fetch template
	var tmpl domain.PatientIDCardTemplate
	err = tx.QueryRow(ctx, `
		SELECT id, name, color, email, phone, dob, blood_group, address, patient_unique_id
		FROM patient_id_card_template
		WHERE id = $1
	`, in.TemplateID).Scan(
		&tmpl.ID, &tmpl.Name, &tmpl.Color, &tmpl.Email, &tmpl.Phone, &tmpl.DOB,
		&tmpl.BloodGroup, &tmpl.Address, &tmpl.PatientUniqueID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.GenerateSmartCardsResult{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.GenerateSmartCardsResult{}, err
	}

	type patientRow struct {
		id       string
		uniqueID string
		name     string
		email    string
		phone    string
		dob      string
		blood    string
		address  string
		mrn      string
	}

	var query string
	var args []any

	switch in.Mode {
	case "all":
		query = `
			SELECT p.id, COALESCE(p.patient_unique_id, ''),
			       (p.given_name || ' ' || p.family_name) AS name,
			       COALESCE(pp.email, '') AS email,
			       p.phone,
			       to_char(p.date_of_birth, 'YYYY-MM-DD') AS dob,
			       COALESCE(pp.blood_group, '') AS blood_group,
			       CASE WHEN pp.address1 <> '' THEN pp.address1 || CASE WHEN pp.city <> '' THEN ', ' || pp.city ELSE '' END ELSE COALESCE(pp.city, '') END AS address,
			       p.medical_record_number::text AS mrn
			FROM patient p
			LEFT JOIN patient_profile pp ON pp.patient_id = p.id
			ORDER BY p.created_at ASC, p.id ASC
			FOR UPDATE OF p
		`
	case "single":
		query = `
			SELECT p.id, COALESCE(p.patient_unique_id, ''),
			       (p.given_name || ' ' || p.family_name) AS name,
			       COALESCE(pp.email, '') AS email,
			       p.phone,
			       to_char(p.date_of_birth, 'YYYY-MM-DD') AS dob,
			       COALESCE(pp.blood_group, '') AS blood_group,
			       CASE WHEN pp.address1 <> '' THEN pp.address1 || CASE WHEN pp.city <> '' THEN ', ' || pp.city ELSE '' END ELSE COALESCE(pp.city, '') END AS address,
			       p.medical_record_number::text AS mrn
			FROM patient p
			LEFT JOIN patient_profile pp ON pp.patient_id = p.id
			WHERE p.id = $1
			FOR UPDATE OF p
		`
		args = append(args, in.PatientID)
	case "unassigned":
		query = `
			SELECT p.id, COALESCE(p.patient_unique_id, ''),
			       (p.given_name || ' ' || p.family_name) AS name,
			       COALESCE(pp.email, '') AS email,
			       p.phone,
			       to_char(p.date_of_birth, 'YYYY-MM-DD') AS dob,
			       COALESCE(pp.blood_group, '') AS blood_group,
			       CASE WHEN pp.address1 <> '' THEN pp.address1 || CASE WHEN pp.city <> '' THEN ', ' || pp.city ELSE '' END ELSE COALESCE(pp.city, '') END AS address,
			       p.medical_record_number::text AS mrn
			FROM patient p
			LEFT JOIN patient_profile pp ON pp.patient_id = p.id
			WHERE p.smart_card_template_id IS NULL OR NOT EXISTS (
				SELECT 1 FROM patient_smart_card sc WHERE sc.patient_id = p.id AND sc.status = 'active'
			)
			ORDER BY p.created_at ASC, p.id ASC
			FOR UPDATE OF p
		`
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return domain.GenerateSmartCardsResult{}, err
	}

	var candidates []patientRow
	for rows.Next() {
		var pr patientRow
		if err := rows.Scan(
			&pr.id, &pr.uniqueID, &pr.name, &pr.email, &pr.phone, &pr.dob, &pr.blood, &pr.address, &pr.mrn,
		); err != nil {
			rows.Close()
			return domain.GenerateSmartCardsResult{}, err
		}
		candidates = append(candidates, pr)
	}
	rows.Close()

	var issuedList []domain.IssuedCardRecord
	now := time.Now().UTC()
	expiresAt := now.AddDate(1, 0, 0) // 1 year validity

	for _, p := range candidates {
		uniqueID := p.uniqueID
		if uniqueID == "" {
			var seq int64
			if err := tx.QueryRow(ctx, `SELECT nextval('patient_unique_id_seq')`).Scan(&seq); err != nil {
				return domain.GenerateSmartCardsResult{}, err
			}
			uniqueID = fmt.Sprintf("P-%05d", seq)
			_, err = tx.Exec(ctx, `
				UPDATE patient
				SET patient_unique_id = $1, smart_card_template_id = $2
				WHERE id = $3
			`, uniqueID, tmpl.ID, p.id)
		} else {
			_, err = tx.Exec(ctx, `
				UPDATE patient
				SET smart_card_template_id = $1
				WHERE id = $2
			`, tmpl.ID, p.id)
		}
		if err != nil {
			return domain.GenerateSmartCardsResult{}, err
		}

		// Reassignment: Revoke previous active card if any
		_, err = tx.Exec(ctx, `
			UPDATE patient_smart_card
			SET status = 'revoked', revoked_reason = 'Reassigned new card/template'
			WHERE patient_id = $1 AND status = 'active'
		`, p.id)
		if err != nil {
			return domain.GenerateSmartCardsResult{}, err
		}

		// Issue new card
		rawRand := make([]byte, 16)
		if _, err := rand.Read(rawRand); err != nil {
			return domain.GenerateSmartCardsResult{}, err
		}
		qrToken := hex.EncodeToString(rawRand)
		cardNumber := fmt.Sprintf("SC-%s-%s", uniqueID, hex.EncodeToString(rawRand[:3]))

		var cardID string
		var issuedAt time.Time
		err = tx.QueryRow(ctx, `
			INSERT INTO patient_smart_card (
				patient_id, card_number, qr_token, template_name, template_id, status, issued_at, expires_at, issued_by
			) VALUES ($1, $2, $3, $4, $5, 'active', clock_timestamp(), $6, $7)
			RETURNING id, issued_at
		`, p.id, cardNumber, qrToken, tmpl.Name, tmpl.ID, expiresAt, a.ID).Scan(&cardID, &issuedAt)
		if err != nil {
			return domain.GenerateSmartCardsResult{}, err
		}

		issuedList = append(issuedList, domain.IssuedCardRecord{
			ID:              cardID,
			PatientID:       p.id,
			PatientName:     p.name,
			PatientUniqueID: uniqueID,
			CardNumber:      cardNumber,
			QRToken:         qrToken,
			TemplateID:      tmpl.ID,
			TemplateName:    tmpl.Name,
			TemplateColor:   tmpl.Color,
			Status:          "active",
			IssuedAt:        issuedAt,
			ExpiresAt:       expiresAt,
			IssuedBy:        a.ID,
			Email:           p.email,
			Phone:           p.phone,
			DOB:             p.dob,
			BloodGroup:      p.blood,
			Address:         p.address,
			MRN:             p.mrn,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.GenerateSmartCardsResult{}, err
	}

	return domain.GenerateSmartCardsResult{
		GeneratedCount: len(issuedList),
		Cards:          issuedList,
	}, nil
}

func (s Store) ListIssuedCards(ctx context.Context, a domain.Actor, filter domain.ListIssuedCardsFilter) (domain.IssuedCardsPage, error) {
	where := []string{"sc.status = 'active'"}
	args := []any{}
	argIdx := 1

	if filter.TemplateID != "" {
		where = append(where, fmt.Sprintf("sc.template_id = $%d::uuid", argIdx))
		args = append(args, filter.TemplateID)
		argIdx++
	}

	if filter.Search != "" {
		pattern := "%" + strings.ToLower(filter.Search) + "%"
		where = append(where, fmt.Sprintf(`(
			lower(p.given_name || ' ' || p.family_name) LIKE $%d OR
			lower(COALESCE(p.patient_unique_id, '')) LIKE $%d OR
			lower(sc.card_number) LIKE $%d OR
			lower(sc.template_name) LIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx))
		args = append(args, pattern)
		argIdx++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var totalCount int
	countQuery := fmt.Sprintf(`
		SELECT count(*)
		FROM patient_smart_card sc
		JOIN patient p ON p.id = sc.patient_id
		%s
	`, whereClause)
	if err := s.DB.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return domain.IssuedCardsPage{}, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	query := fmt.Sprintf(`
		SELECT sc.id, sc.patient_id,
		       (p.given_name || ' ' || p.family_name) AS patient_name,
		       COALESCE(p.patient_unique_id, '') AS patient_unique_id,
		       sc.card_number, sc.qr_token,
		       COALESCE(sc.template_id::text, '') AS template_id,
		       COALESCE(t.name, sc.template_name) AS template_name,
		       COALESCE(t.color, '#f4bf16') AS template_color,
		       sc.status, sc.issued_at, sc.expires_at, sc.issued_by,
		       COALESCE(pp.email, '') AS email,
		       p.phone,
		       to_char(p.date_of_birth, 'YYYY-MM-DD') AS dob,
		       COALESCE(pp.blood_group, '') AS blood_group,
		       CASE WHEN pp.address1 <> '' THEN pp.address1 || CASE WHEN pp.city <> '' THEN ', ' || pp.city ELSE '' END ELSE COALESCE(pp.city, '') END AS address,
		       p.medical_record_number::text AS mrn
		FROM patient_smart_card sc
		JOIN patient p ON p.id = sc.patient_id
		LEFT JOIN patient_profile pp ON pp.patient_id = sc.patient_id
		LEFT JOIN patient_id_card_template t ON t.id = sc.template_id
		%s
		ORDER BY sc.issued_at DESC, sc.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.PageSize, offset)

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return domain.IssuedCardsPage{}, err
	}
	defer rows.Close()

	var cards []domain.IssuedCardRecord
	for rows.Next() {
		var c domain.IssuedCardRecord
		if err := rows.Scan(
			&c.ID, &c.PatientID, &c.PatientName, &c.PatientUniqueID, &c.CardNumber, &c.QRToken,
			&c.TemplateID, &c.TemplateName, &c.TemplateColor, &c.Status, &c.IssuedAt, &c.ExpiresAt, &c.IssuedBy,
			&c.Email, &c.Phone, &c.DOB, &c.BloodGroup, &c.Address, &c.MRN,
		); err != nil {
			return domain.IssuedCardsPage{}, err
		}
		cards = append(cards, c)
	}
	if cards == nil {
		cards = []domain.IssuedCardRecord{}
	}

	return domain.IssuedCardsPage{
		Cards:      cards,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s Store) DeleteCardAssignment(ctx context.Context, a domain.Actor, cardID string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var patientID string
	var status string
	err = tx.QueryRow(ctx, `
		SELECT patient_id, status
		FROM patient_smart_card
		WHERE id = $1
		FOR UPDATE
	`, cardID).Scan(&patientID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}

	// Preserve patient and patient_unique_id, but clear template assignment
	_, err = tx.Exec(ctx, `
		UPDATE patient
		SET smart_card_template_id = NULL
		WHERE id = $1
	`, patientID)
	if err != nil {
		return err
	}

	// Revoke the card so its QR credential can no longer be verified
	_, err = tx.Exec(ctx, `
		UPDATE patient_smart_card
		SET status = 'revoked', revoked_reason = 'Card assignment removed'
		WHERE id = $1
	`, cardID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s Store) VerifySmartCardQR(ctx context.Context, qrToken string) (domain.SmartCardVerificationResult, error) {
	var (
		cardID, patientID, cardNumber, status, revokedReason string
		expiresAt                                            time.Time
		patientName, patientUniqueID                         string
		email, phone, dob, bloodGroup, address               string
		tEmail, tPhone, tDOB, tBloodGroup, tAddress          *bool
		tPatientUniqueID                                     *bool
	)

	err := s.DB.QueryRow(ctx, `
		SELECT sc.id, sc.patient_id, sc.card_number, sc.status, sc.revoked_reason, sc.expires_at,
		       (p.given_name || ' ' || p.family_name) AS patient_name,
		       COALESCE(p.patient_unique_id, '') AS patient_unique_id,
		       COALESCE(pp.email, '') AS email,
		       p.phone,
		       to_char(p.date_of_birth, 'YYYY-MM-DD') AS dob,
		       COALESCE(pp.blood_group, '') AS blood_group,
		       CASE WHEN pp.address1 <> '' THEN pp.address1 || CASE WHEN pp.city <> '' THEN ', ' || pp.city ELSE '' END ELSE COALESCE(pp.city, '') END AS address,
		       t.email, t.phone, t.dob, t.blood_group, t.address, t.patient_unique_id
		FROM patient_smart_card sc
		JOIN patient p ON p.id = sc.patient_id
		LEFT JOIN patient_profile pp ON pp.patient_id = sc.patient_id
		LEFT JOIN patient_id_card_template t ON t.id = sc.template_id
		WHERE sc.qr_token = $1
	`, qrToken).Scan(
		&cardID, &patientID, &cardNumber, &status, &revokedReason, &expiresAt,
		&patientName, &patientUniqueID,
		&email, &phone, &dob, &bloodGroup, &address,
		&tEmail, &tPhone, &tDOB, &tBloodGroup, &tAddress, &tPatientUniqueID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SmartCardVerificationResult{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.SmartCardVerificationResult{}, err
	}

	if status != "active" {
		return domain.SmartCardVerificationResult{
			Valid:         false,
			Status:        status,
			RevokedReason: revokedReason,
		}, nil
	}

	if time.Now().UTC().After(expiresAt) {
		return domain.SmartCardVerificationResult{
			Valid:  false,
			Status: "expired",
		}, nil
	}

	res := domain.SmartCardVerificationResult{
		Valid:       true,
		Status:      "active",
		CardNumber:  cardNumber,
		ExpiresAt:   expiresAt,
		PatientName: patientName,
	}

	// Expose only authorized fields based on template toggles
	allowUnique := tPatientUniqueID == nil || *tPatientUniqueID
	allowEmail := tEmail == nil || *tEmail
	allowPhone := tPhone == nil || *tPhone
	allowDOB := tDOB == nil || *tDOB
	allowBlood := tBloodGroup == nil || *tBloodGroup
	allowAddress := tAddress == nil || *tAddress

	if allowUnique {
		res.PatientUniqueID = patientUniqueID
	}
	if allowEmail {
		res.Email = email
	}
	if allowPhone {
		res.Phone = phone
	}
	if allowDOB {
		res.DOB = dob
	}
	if allowBlood {
		res.BloodGroup = bloodGroup
	}
	if allowAddress {
		res.Address = address
	}

	return res, nil
}

// ─── Duplicate Patients & Merge ───────────────────────────────────────────────

func (s Store) FindDuplicatePatients(ctx context.Context, query string) ([]domain.DuplicatePatientCandidate, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT p.id, (p.given_name || ' ' || p.family_name) AS name,
		       COALESCE(p.phone, '') AS phone, COALESCE(pp.contact_email, '') AS email,
		       COALESCE(pp.gender, '') AS gender,
		       CASE
		         WHEN p.phone = $1 AND p.phone <> '' THEN 'exact_phone'
		         WHEN lower(pp.contact_email) = lower($1) AND pp.contact_email <> '' THEN 'exact_email'
		         ELSE 'name_match'
		       END AS match_rule
		FROM patient p
		LEFT JOIN patient_profile pp ON pp.patient_id = p.id
		WHERE (p.phone = $1 AND p.phone <> '')
		   OR (lower(pp.contact_email) = lower($1) AND pp.contact_email <> '')
		   OR (strpos(lower(p.given_name || ' ' || p.family_name), lower($1)) > 0)
		ORDER BY match_rule, p.created_at DESC
		LIMIT 20
	`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.DuplicatePatientCandidate
	for rows.Next() {
		var c domain.DuplicatePatientCandidate
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Email, &c.Gender, &c.MatchRule); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	if list == nil {
		list = []domain.DuplicatePatientCandidate{}
	}
	return list, nil
}
