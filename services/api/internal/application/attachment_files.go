package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"hms.local/api/internal/domain"
	"io"
	"net/http"
	"strings"
)

const MaxAttachmentBytes = 25 << 20

type AttachmentFiles interface {
	Put(context.Context, []byte) (string, error)
	Open(context.Context, string) (io.ReadCloser, error)
	Remove(context.Context, string) error
}

func (s AttachmentsService) Upload(ctx context.Context, a domain.Actor, name string, patientID, encounterID *string, data []byte) (domain.SecureAttachment, error) {
	if s.Files == nil {
		return domain.SecureAttachment{}, domain.ErrUnavailable
	}
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" && a.Role != "lab_technician" {
		return domain.SecureAttachment{}, domain.ErrForbidden
	}
	if err := s.Store.AuthorizeAttachment(ctx, a, patientID, encounterID); err != nil {
		return domain.SecureAttachment{}, err
	}
	if len(data) == 0 || len(data) > MaxAttachmentBytes || strings.ContainsAny(name, "/\\\r\n\x00") {
		return domain.SecureAttachment{}, domain.ErrValidation
	}
	mimeType := strings.Split(http.DetectContentType(data), ";")[0]
	// Executable/HTML/SVG and unrecognized binary uploads are rejected. More
	// specialized formats require a validated parser before enabling them.
	switch mimeType {
	case "application/pdf", "image/jpeg", "image/png", "image/webp", "text/plain":
	default:
		return domain.SecureAttachment{}, domain.ErrValidation
	}
	hash := sha256.Sum256(data)
	input := domain.CreateSecureAttachmentInput{FileName: name, MimeType: mimeType, FileSizeBytes: int64(len(data)), Sha256Hash: hex.EncodeToString(hash[:]), PatientID: patientID, EncounterID: encounterID, StoragePath: "pending"}
	if err := input.Validate(); err != nil {
		return domain.SecureAttachment{}, err
	}
	if s.Scanner == nil && s.RequireScan {
		return domain.SecureAttachment{}, domain.ErrUnavailable
	}
	if s.Scanner != nil {
		if err := s.Scanner.Scan(ctx, data); err != nil {
			return domain.SecureAttachment{}, err
		}
	}
	key, err := s.Files.Put(ctx, data)
	if err != nil {
		return domain.SecureAttachment{}, err
	}
	input.StoragePath = key
	att, err := s.CreateAttachment(ctx, a, input)
	if err != nil {
		_ = s.Files.Remove(ctx, key)
	}
	return att, err
}

func (s AttachmentsService) UploadPublic(ctx context.Context, a domain.Actor, name string, data []byte) (domain.SecureAttachment, error) {
	if s.Files == nil {
		return domain.SecureAttachment{}, domain.ErrUnavailable
	}
	if a.Role != "admin" && !a.Can("settings.manage") {
		return domain.SecureAttachment{}, domain.ErrForbidden
	}
	if len(data) == 0 || len(data) > MaxAttachmentBytes || strings.ContainsAny(name, "/\\\r\n\x00") {
		return domain.SecureAttachment{}, domain.ErrValidation
	}
	mimeType := strings.Split(http.DetectContentType(data), ";")[0]
	switch mimeType {
	case "image/jpeg", "image/png", "image/webp", "image/x-icon", "image/vnd.microsoft.icon":
	default:
		return domain.SecureAttachment{}, domain.ErrValidation
	}
	hash := sha256.Sum256(data)
	input := domain.CreateSecureAttachmentInput{
		FileName:      name,
		MimeType:      mimeType,
		FileSizeBytes: int64(len(data)),
		Sha256Hash:    hex.EncodeToString(hash[:]),
		IsPublic:      true,
		StoragePath:   "pending",
	}
	if err := input.Validate(); err != nil {
		return domain.SecureAttachment{}, err
	}
	if s.Scanner == nil && s.RequireScan {
		return domain.SecureAttachment{}, domain.ErrUnavailable
	}
	if s.Scanner != nil {
		if err := s.Scanner.Scan(ctx, data); err != nil {
			return domain.SecureAttachment{}, err
		}
	}
	key, err := s.Files.Put(ctx, data)
	if err != nil {
		return domain.SecureAttachment{}, err
	}
	input.StoragePath = key
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		_ = s.Files.Remove(ctx, key)
		return domain.SecureAttachment{}, err
	}
	token := hex.EncodeToString(raw)
	att, err := s.Store.SaveAttachment(ctx, a, token, input)
	if err != nil {
		_ = s.Files.Remove(ctx, key)
	}
	return att, err
}

func (s AttachmentsService) Download(ctx context.Context, a domain.Actor, token string) (domain.SecureAttachment, io.ReadCloser, error) {
	att, err := s.GetAttachmentByToken(ctx, a, token)
	if err != nil {
		return att, nil, err
	}
	if s.Files == nil {
		return att, nil, domain.ErrUnavailable
	}
	reader, err := s.Files.Open(ctx, att.StoragePath)
	return att, reader, err
}
