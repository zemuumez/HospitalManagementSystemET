package domain

import (
	"strings"
	"time"
)

type SecureAttachment struct {
	ID            string    `json:"id"`
	Token         string    `json:"token"`
	FileName      string    `json:"fileName"`
	MimeType      string    `json:"mimeType"`
	FileSizeBytes int64     `json:"fileSizeBytes"`
	StoragePath   string    `json:"-"`
	Sha256Hash    string    `json:"sha256Hash"`
	UploaderID    string    `json:"uploaderId"`
	PatientID     *string   `json:"patientId,omitempty"`
	EncounterID   *string   `json:"encounterId,omitempty"`
	IsPublic      bool      `json:"isPublic"`
	CreatedAt     time.Time `json:"createdAt"`
}

type CreateSecureAttachmentInput struct {
	FileName      string  `json:"fileName"`
	MimeType      string  `json:"mimeType"`
	FileSizeBytes int64   `json:"fileSizeBytes"`
	StoragePath   string  `json:"storagePath"`
	Sha256Hash    string  `json:"sha256Hash"`
	PatientID     *string `json:"patientId,omitempty"`
	EncounterID   *string `json:"encounterId,omitempty"`
	IsPublic      bool    `json:"isPublic"`
}

func (i *CreateSecureAttachmentInput) Validate() error {
	i.FileName = strings.TrimSpace(i.FileName)
	i.MimeType = strings.TrimSpace(strings.ToLower(i.MimeType))
	i.StoragePath = strings.TrimSpace(i.StoragePath)
	i.Sha256Hash = strings.TrimSpace(strings.ToLower(i.Sha256Hash))

	if len(i.FileName) < 1 || len(i.FileName) > 255 {
		return ErrValidation
	}
	if i.FileSizeBytes <= 0 || i.FileSizeBytes > 26214400 { // 25MB max
		return ErrValidation
	}
	if len(i.Sha256Hash) != 64 {
		return ErrValidation
	}
	if i.StoragePath == "" {
		return ErrValidation
	}

	allowedMimes := map[string]bool{
		"application/pdf":   true,
		"image/jpeg":        true,
		"image/png":         true,
		"image/webp":        true,
		"text/plain":        true,
		"text/csv":          true,
		"application/dicom": true,
		"application/zip":   true,
	}
	if !allowedMimes[i.MimeType] {
		return ErrValidation
	}

	if i.PatientID != nil && *i.PatientID != "" && !UUIDPattern.MatchString(*i.PatientID) {
		return ErrValidation
	}
	if i.EncounterID != nil && *i.EncounterID != "" && !UUIDPattern.MatchString(*i.EncounterID) {
		return ErrValidation
	}

	return nil
}
