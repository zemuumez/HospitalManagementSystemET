package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

// A token locates a record; it never grants access. Patient ownership and care
// assignment are checked against current database records for every request.
func (s Store) AuthorizeAttachment(ctx context.Context, a domain.Actor, patientID, encounterID *string) error {
	if patientID == nil {
		if a.Role == "admin" || a.Can("settings.manage") {
			return nil
		}
		return domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(*patientID) {
		return domain.ErrValidation
	}
	var allowed bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM patient p
 WHERE p.id=$3::uuid
 AND ($4::uuid IS NULL OR EXISTS(SELECT 1 FROM encounter e WHERE e.id=$4::uuid AND e.patient_id=p.id))
 AND ($1='admin' OR ($1='patient' AND patient_portal_owner(p.id)=$2)
 OR ($1='doctor' AND (p.clinician_user_id=$2 OR EXISTS(SELECT 1 FROM encounter e WHERE e.patient_id=p.id AND e.doctor_id=$2 AND ($4::uuid IS NULL OR e.id=$4::uuid))))
 OR ($1 IN ('doctor','nurse') AND EXISTS(SELECT 1 FROM encounter e JOIN encounter_care_team c ON c.encounter_id=e.id WHERE e.patient_id=p.id AND c.staff_id=$2 AND c.revoked_at IS NULL AND ($4::uuid IS NULL OR e.id=$4::uuid)))
 OR ($1='lab_technician' AND $4::uuid IS NOT NULL AND EXISTS(SELECT 1 FROM diagnostic_order o WHERE o.encounter_id=$4::uuid))
 OR ($1='nurse' AND EXISTS(SELECT 1 FROM encounter e JOIN encounter_nurse n ON n.encounter_id=e.id WHERE e.patient_id=p.id AND n.nurse_id=$2 AND n.active AND ($4::uuid IS NULL OR e.id=$4::uuid)))))`, a.Role, a.ID, *patientID, encounterID).Scan(&allowed)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}

func (s Store) SaveAttachment(ctx context.Context, a domain.Actor, token string, input domain.CreateSecureAttachmentInput) (domain.SecureAttachment, error) {
	var att domain.SecureAttachment
	att.Token = token
	att.FileName = input.FileName
	att.MimeType = input.MimeType
	att.FileSizeBytes = input.FileSizeBytes
	att.StoragePath = input.StoragePath
	att.Sha256Hash = input.Sha256Hash
	att.UploaderID = a.ID
	att.PatientID = input.PatientID
	att.EncounterID = input.EncounterID
	att.IsPublic = input.IsPublic

	err := s.DB.QueryRow(ctx, `
		INSERT INTO secure_attachment (
			token, file_name, mime_type, file_size_bytes, storage_path,
			sha256_hash, uploader_id, patient_id, encounter_id, is_public
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`, token, input.FileName, input.MimeType, input.FileSizeBytes, input.StoragePath,
		input.Sha256Hash, a.ID, input.PatientID, input.EncounterID, input.IsPublic,
	).Scan(&att.ID, &att.CreatedAt)
	if err != nil {
		return domain.SecureAttachment{}, err
	}
	return att, nil
}

func (s Store) GetAttachmentByToken(ctx context.Context, token string) (domain.SecureAttachment, error) {
	var att domain.SecureAttachment
	var patientID, encounterID *string
	err := s.DB.QueryRow(ctx, `
		SELECT id, token, file_name, mime_type, file_size_bytes, storage_path,
		       sha256_hash, uploader_id, patient_id::text, encounter_id::text, is_public, created_at, patient_released
		FROM secure_attachment
		WHERE token = $1
	`, token).Scan(
		&att.ID, &att.Token, &att.FileName, &att.MimeType, &att.FileSizeBytes, &att.StoragePath,
		&att.Sha256Hash, &att.UploaderID, &patientID, &encounterID, &att.IsPublic, &att.CreatedAt, &att.PatientReleased,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.SecureAttachment{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.SecureAttachment{}, err
	}
	att.PatientID = patientID
	att.EncounterID = encounterID
	return att, nil
}

func (s Store) ListPatientAttachments(ctx context.Context, patientID string) ([]domain.SecureAttachment, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, token, file_name, mime_type, file_size_bytes, storage_path,
		       sha256_hash, uploader_id, patient_id::text, encounter_id::text, is_public, created_at, patient_released
		FROM secure_attachment
		WHERE canonical_patient_id(patient_id) = canonical_patient_id($1::uuid)
		ORDER BY created_at DESC
	`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.SecureAttachment
	for rows.Next() {
		var att domain.SecureAttachment
		var patID, encID *string
		if err := rows.Scan(
			&att.ID, &att.Token, &att.FileName, &att.MimeType, &att.FileSizeBytes, &att.StoragePath,
			&att.Sha256Hash, &att.UploaderID, &patID, &encID, &att.IsPublic, &att.CreatedAt, &att.PatientReleased,
		); err != nil {
			return nil, err
		}
		att.PatientID = patID
		att.EncounterID = encID
		list = append(list, att)
	}
	if list == nil {
		list = []domain.SecureAttachment{}
	}
	return list, nil
}

func (s Store) ReleaseAttachment(ctx context.Context, a domain.Actor, token string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var id string
	var released bool
	err = tx.QueryRow(ctx, `SELECT id,patient_released FROM secure_attachment WHERE token=$1 FOR UPDATE`, token).Scan(&id, &released)
	if err != nil {
		return err
	}
	if released {
		return nil
	}
	err = tx.QueryRow(ctx, `UPDATE secure_attachment sa SET patient_released=true WHERE token=$1 AND NOT EXISTS(SELECT 1 FROM diagnostic_report_file f WHERE f.file_url='/v1/attachments/'||sa.token||'/content' AND NOT f.patient_released) RETURNING id`, token).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrStale
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO attachment_release_event(attachment_id,actor_id) VALUES($1,$2)`, id, a.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s Store) DeleteAttachment(ctx context.Context, a domain.Actor, token string) (string, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var id, storagePath string
	var isPublic bool
	var patientID, encounterID *string
	err = tx.QueryRow(ctx, `
		SELECT id, storage_path, is_public, patient_id::text, encounter_id::text
		FROM secure_attachment
		WHERE token = $1
		FOR UPDATE
	`, token).Scan(&id, &storagePath, &isPublic, &patientID, &encounterID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	// Clinical files cannot be deleted through public settings retirement!
	if !isPublic || patientID != nil || encounterID != nil {
		return "", domain.ErrForbidden
	}

	var inUse bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM hospital_general_setting WHERE value LIKE '%' || $1 || '%'
		) OR EXISTS(
			SELECT 1 FROM front_cms_setting WHERE value LIKE '%' || $1 || '%'
		)
	`, token).Scan(&inUse)
	if err != nil {
		return "", err
	}
	if inUse {
		return "", domain.ErrInUse
	}

	_, err = tx.Exec(ctx, `DELETE FROM secure_attachment WHERE id = $1`, id)
	if err != nil {
		return "", err
	}
	if err = pharmacyAudit(ctx, tx, a, "attachment.retired", token); err != nil {
		return "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", err
	}
	return storagePath, nil
}
