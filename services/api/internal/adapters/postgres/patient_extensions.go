package postgres

import (
	"context"
	"errors"
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

// ─── Patient Smart Cards ──────────────────────────────────────────────────────

func (s Store) PatientSmartCards(ctx context.Context, patientID string) ([]domain.PatientSmartCard, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT sc.id, sc.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       p.medical_record_number::text AS mrn, COALESCE(pp.blood_group, '') AS blood_group,
		       sc.card_number, sc.qr_token, sc.template_name, sc.status, sc.revoked_reason,
		       sc.issued_at, sc.expires_at, sc.issued_by
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
			&c.ID, &c.PatientID, &c.PatientName, &c.MRN, &c.BloodGroup,
			&c.CardNumber, &c.QRToken, &c.TemplateName, &c.Status, &c.RevokedReason,
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

func (s Store) VerifySmartCardQR(ctx context.Context, qrToken string) (domain.PatientSmartCard, error) {
	var c domain.PatientSmartCard
	err := s.DB.QueryRow(ctx, `
		SELECT sc.id, sc.patient_id, (p.given_name || ' ' || p.family_name) AS patient_name,
		       p.medical_record_number::text AS mrn, COALESCE(pp.blood_group, '') AS blood_group,
		       sc.card_number, sc.qr_token, sc.template_name, sc.status, sc.revoked_reason,
		       sc.issued_at, sc.expires_at, sc.issued_by
		FROM patient_smart_card sc
		JOIN patient p ON p.id = sc.patient_id
		LEFT JOIN patient_profile pp ON pp.patient_id = sc.patient_id
		WHERE sc.qr_token = $1
	`, qrToken).Scan(
		&c.ID, &c.PatientID, &c.PatientName, &c.MRN, &c.BloodGroup,
		&c.CardNumber, &c.QRToken, &c.TemplateName, &c.Status, &c.RevokedReason,
		&c.IssuedAt, &c.ExpiresAt, &c.IssuedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PatientSmartCard{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.PatientSmartCard{}, err
	}
	return c, nil
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
