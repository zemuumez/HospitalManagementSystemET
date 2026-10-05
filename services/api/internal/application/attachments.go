package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"hms.local/api/internal/domain"
)

type AttachmentsRepository interface {
	ReleaseAttachment(context.Context, domain.Actor, string) error
	AuthorizeAttachment(context.Context, domain.Actor, *string, *string) error
	SaveAttachment(ctx context.Context, a domain.Actor, token string, input domain.CreateSecureAttachmentInput) (domain.SecureAttachment, error)
	GetAttachmentByToken(ctx context.Context, token string) (domain.SecureAttachment, error)
	ListPatientAttachments(ctx context.Context, patientID string) ([]domain.SecureAttachment, error)
}

type AttachmentScanner interface {
	Scan(context.Context, []byte) error
}

type AttachmentsService struct {
	Scanner     AttachmentScanner
	RequireScan bool
	Store       AttachmentsRepository
	Files       AttachmentFiles
}

func (s AttachmentsService) CreateAttachment(ctx context.Context, a domain.Actor, input domain.CreateSecureAttachmentInput) (domain.SecureAttachment, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" && a.Role != "lab_technician" {
		return domain.SecureAttachment{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.SecureAttachment{}, err
	}
	// Clinical attachments are always private, including for their uploader.
	if input.IsPublic {
		return domain.SecureAttachment{}, domain.ErrValidation
	}
	if err := s.Store.AuthorizeAttachment(ctx, a, input.PatientID, input.EncounterID); err != nil {
		return domain.SecureAttachment{}, err
	}

	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return domain.SecureAttachment{}, err
	}
	token := hex.EncodeToString(raw)

	return s.Store.SaveAttachment(ctx, a, token, input)
}

func (s AttachmentsService) GetAttachmentByToken(ctx context.Context, a domain.Actor, token string) (domain.SecureAttachment, error) {
	token = strings.TrimSpace(token)
	if len(token) < 32 {
		return domain.SecureAttachment{}, domain.ErrValidation
	}
	att, err := s.Store.GetAttachmentByToken(ctx, token)
	if err != nil {
		return domain.SecureAttachment{}, err
	}
	if a.Role == "patient" && !att.PatientReleased {
		return domain.SecureAttachment{}, domain.ErrForbidden
	}

	if err := s.Store.AuthorizeAttachment(ctx, a, att.PatientID, att.EncounterID); err != nil {
		return domain.SecureAttachment{}, err
	}
	return att, nil
}

func (s AttachmentsService) ListPatientAttachments(ctx context.Context, a domain.Actor, patientID string) ([]domain.SecureAttachment, error) {
	if !domain.UUIDPattern.MatchString(patientID) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizeAttachment(ctx, a, &patientID, nil); err != nil {
		return nil, err
	}
	items, err := s.Store.ListPatientAttachments(ctx, patientID)
	if err != nil {
		return nil, err
	}
	visible := []domain.SecureAttachment{}
	for _, att := range items {
		if a.Role == "patient" && !att.PatientReleased {
			continue
		}
		if err := s.Store.AuthorizeAttachment(ctx, a, att.PatientID, att.EncounterID); err == nil {
			visible = append(visible, att)
		} else if !errors.Is(err, domain.ErrForbidden) {
			return nil, err
		}
	}
	return visible, nil
}

func (s AttachmentsService) Release(ctx context.Context, a domain.Actor, token string) error {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.ErrForbidden
	}
	if _, err := s.GetAttachmentByToken(ctx, a, token); err != nil {
		return err
	}
	return s.Store.ReleaseAttachment(ctx, a, token)
}
