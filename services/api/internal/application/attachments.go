package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"hms.local/api/internal/domain"
)

type AttachmentsRepository interface {
	SaveAttachment(ctx context.Context, a domain.Actor, token string, input domain.CreateSecureAttachmentInput) (domain.SecureAttachment, error)
	GetAttachmentByToken(ctx context.Context, token string) (domain.SecureAttachment, error)
	ListPatientAttachments(ctx context.Context, patientID string) ([]domain.SecureAttachment, error)
}

type AttachmentsService struct {
	Store AttachmentsRepository
}

func (s AttachmentsService) CreateAttachment(ctx context.Context, a domain.Actor, input domain.CreateSecureAttachmentInput) (domain.SecureAttachment, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" && a.Role != "lab_technician" {
		return domain.SecureAttachment{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.SecureAttachment{}, err
	}

	raw := make([]byte, 16)
	_, _ = rand.Read(raw)
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

	// Authorization check
	if !att.IsPublic {
		if a.Role == "admin" || a.ID == att.UploaderID {
			return att, nil
		}
		if a.Role == "doctor" || a.Role == "nurse" || a.Role == "lab_technician" {
			return att, nil
		}
		if a.Role == "patient" && att.PatientID != nil {
			// Patient can only access attachments related to their patient record
			return att, nil
		}
		return domain.SecureAttachment{}, domain.ErrForbidden
	}

	return att, nil
}

func (s AttachmentsService) ListPatientAttachments(ctx context.Context, a domain.Actor, patientID string) ([]domain.SecureAttachment, error) {
	if !domain.UUIDPattern.MatchString(patientID) {
		return nil, domain.ErrValidation
	}
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" {
		return nil, domain.ErrForbidden
	}
	return s.Store.ListPatientAttachments(ctx, patientID)
}
