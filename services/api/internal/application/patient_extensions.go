package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"hms.local/api/internal/domain"
)

type PatientExtensionsRepository interface {
	AuthorizePatientRecord(context.Context, domain.Actor, string) error
	PatientIdentities(context.Context, string) ([]domain.PatientIdentity, error)
	// Guardian & consent
	PatientContactConsent(ctx context.Context, patientID string) (domain.PatientContactConsent, error)
	SavePatientContactConsent(ctx context.Context, a domain.Actor, c domain.PatientContactConsent) error

	// Smart cards
	PatientSmartCards(ctx context.Context, patientID string) ([]domain.PatientSmartCard, error)
	IssueSmartCard(ctx context.Context, a domain.Actor, card domain.PatientSmartCard) (domain.PatientSmartCard, error)
	RevokeSmartCard(ctx context.Context, a domain.Actor, cardID string, reason string) error
	VerifySmartCardQR(ctx context.Context, qrToken string) (domain.PatientSmartCard, error)

	// Duplicates & merge
	FindDuplicatePatients(ctx context.Context, query string) ([]domain.DuplicatePatientCandidate, error)
	MergePatients(ctx context.Context, a domain.Actor, input domain.MergePatientInput) error
}

type PatientExtensionsService struct {
	Store PatientExtensionsRepository
}

// ─── Guardian & Consent ───────────────────────────────────────────────────────

func (s PatientExtensionsService) GetConsent(ctx context.Context, a domain.Actor, patientID string) (domain.PatientContactConsent, error) {
	if !a.Can("patients.read") {
		return domain.PatientContactConsent{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(patientID) {
		return domain.PatientContactConsent{}, domain.ErrValidation
	}
	if err := s.Store.AuthorizePatientRecord(ctx, a, patientID); err != nil {
		return domain.PatientContactConsent{}, err
	}
	return s.Store.PatientContactConsent(ctx, patientID)
}

func (s PatientExtensionsService) SaveConsent(ctx context.Context, a domain.Actor, c domain.PatientContactConsent) error {
	if !a.Can("patients.write") {
		return domain.ErrForbidden
	}
	if err := c.Validate(); err != nil {
		return err
	}
	if err := s.Store.AuthorizePatientRecord(ctx, a, c.PatientID); err != nil {
		return err
	}
	return s.Store.SavePatientContactConsent(ctx, a, c)
}

// ─── Smart Cards ─────────────────────────────────────────────────────────────

func (s PatientExtensionsService) ListSmartCards(ctx context.Context, a domain.Actor, patientID string) ([]domain.PatientSmartCard, error) {
	if !a.Can("patients.read") {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(patientID) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizePatientRecord(ctx, a, patientID); err != nil {
		return nil, err
	}
	return s.Store.PatientSmartCards(ctx, patientID)
}

func (s PatientExtensionsService) IssueSmartCard(ctx context.Context, a domain.Actor, input domain.IssueSmartCardInput) (domain.PatientSmartCard, error) {
	if a.Role != "admin" && a.Role != "receptionist" {
		return domain.PatientSmartCard{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.PatientSmartCard{}, err
	}

	// Generate unique card number and cryptographic QR token
	rawRand := make([]byte, 16)
	_, _ = rand.Read(rawRand)
	qrToken := hex.EncodeToString(rawRand)

	now := time.Now().UTC()
	cardNumber := fmt.Sprintf("HMS-SC-%06d-%s", now.Unix()%1000000, hex.EncodeToString(rawRand[:2]))

	card := domain.PatientSmartCard{
		PatientID:    input.PatientID,
		CardNumber:   cardNumber,
		QRToken:      qrToken,
		TemplateName: input.TemplateName,
		Status:       "active",
		IssuedAt:     now,
		ExpiresAt:    now.AddDate(0, 0, input.ValidityDays),
		IssuedBy:     a.ID,
	}

	return s.Store.IssueSmartCard(ctx, a, card)
}

func (s PatientExtensionsService) RevokeSmartCard(ctx context.Context, a domain.Actor, cardID string, reason string) error {
	if a.Role != "admin" && a.Role != "receptionist" {
		return domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(cardID) {
		return domain.ErrValidation
	}
	reason = strings.TrimSpace(reason)
	if len(reason) == 0 || len(reason) > 500 {
		return domain.ErrValidation
	}
	return s.Store.RevokeSmartCard(ctx, a, cardID, reason)
}

func (s PatientExtensionsService) VerifySmartCardQR(ctx context.Context, qrToken string) (domain.PatientSmartCard, error) {
	qrToken = strings.TrimSpace(qrToken)
	if len(qrToken) < 16 || len(qrToken) > 128 {
		return domain.PatientSmartCard{}, domain.ErrValidation
	}
	return s.Store.VerifySmartCardQR(ctx, qrToken)
}

// ─── Duplicates & Merge ───────────────────────────────────────────────────────

func (s PatientExtensionsService) FindDuplicates(ctx context.Context, a domain.Actor, query string) ([]domain.DuplicatePatientCandidate, error) {
	if a.Role != "admin" && a.Role != "receptionist" {
		return nil, domain.ErrForbidden
	}
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 {
		return []domain.DuplicatePatientCandidate{}, nil
	}
	return s.Store.FindDuplicatePatients(ctx, query)
}

func (s PatientExtensionsService) MergePatients(ctx context.Context, a domain.Actor, input domain.MergePatientInput) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return err
	}
	return s.Store.MergePatients(ctx, a, input)
}

func (s PatientExtensionsService) Identities(ctx context.Context, a domain.Actor, id string) ([]domain.PatientIdentity, error) {
	if !domain.UUIDPattern.MatchString(id) {
		return nil, domain.ErrValidation
	}
	if err := s.Store.AuthorizePatientRecord(ctx, a, id); err != nil {
		return nil, err
	}
	return s.Store.PatientIdentities(ctx, id)
}
