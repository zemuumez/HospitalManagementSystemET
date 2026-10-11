package domain

import (
	"strings"
	"time"
)

// ─── Patient Guardian & Consent ───────────────────────────────────────────────

type PatientContactConsent struct {
	PatientID          string    `json:"patientId"`
	GuardianName       string    `json:"guardianName"`
	GuardianRelation   string    `json:"guardianRelation"`
	GuardianPhone      string    `json:"guardianPhone"`
	GuardianEmail      string    `json:"guardianEmail"`
	SmsConsent         bool      `json:"smsConsent"`
	EmailConsent       bool      `json:"emailConsent"`
	DataSharingConsent bool      `json:"dataSharingConsent"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func (c *PatientContactConsent) Validate() error {
	c.GuardianName = strings.TrimSpace(c.GuardianName)
	c.GuardianRelation = strings.TrimSpace(c.GuardianRelation)
	c.GuardianPhone = strings.TrimSpace(c.GuardianPhone)
	c.GuardianEmail = strings.TrimSpace(c.GuardianEmail)
	if !UUIDPattern.MatchString(c.PatientID) {
		return ErrValidation
	}
	if len([]rune(c.GuardianName)) > 150 || len([]rune(c.GuardianRelation)) > 100 ||
		len([]rune(c.GuardianPhone)) > 64 || len([]rune(c.GuardianEmail)) > 191 {
		return ErrValidation
	}
	return nil
}

// ─── Patient Smart Card & Templates ──────────────────────────────────────────

type PatientIDCardTemplate struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Color           string    `json:"color"`
	Email           bool      `json:"email"`
	Phone           bool      `json:"phone"`
	DOB             bool      `json:"dob"`
	BloodGroup      bool      `json:"bloodGroup"`
	Address         bool      `json:"address"`
	PatientUniqueID bool      `json:"patientUniqueId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CreatePatientIDCardTemplateInput struct {
	Name            string `json:"name"`
	Color           string `json:"color"`
	Email           *bool  `json:"email,omitempty"`
	Phone           *bool  `json:"phone,omitempty"`
	DOB             *bool  `json:"dob,omitempty"`
	BloodGroup      *bool  `json:"bloodGroup,omitempty"`
	Address         *bool  `json:"address,omitempty"`
	PatientUniqueID *bool  `json:"patientUniqueId,omitempty"`
}

func (c *CreatePatientIDCardTemplateInput) Validate() error {
	c.Name = strings.TrimSpace(c.Name)
	c.Color = strings.TrimSpace(c.Color)
	if len(c.Name) < 1 || len(c.Name) > 100 {
		return ErrValidation
	}
	if len(c.Color) < 1 || len(c.Color) > 30 {
		return ErrValidation
	}
	return nil
}

type UpdatePatientIDCardTemplateInput struct {
	Name            *string `json:"name,omitempty"`
	Color           *string `json:"color,omitempty"`
	Email           *bool   `json:"email,omitempty"`
	Phone           *bool   `json:"phone,omitempty"`
	DOB             *bool   `json:"dob,omitempty"`
	BloodGroup      *bool   `json:"bloodGroup,omitempty"`
	Address         *bool   `json:"address,omitempty"`
	PatientUniqueID *bool   `json:"patientUniqueId,omitempty"`
}

func (u *UpdatePatientIDCardTemplateInput) Validate() error {
	if u.Name != nil {
		trimmed := strings.TrimSpace(*u.Name)
		if len(trimmed) < 1 || len(trimmed) > 100 {
			return ErrValidation
		}
		u.Name = &trimmed
	}
	if u.Color != nil {
		trimmed := strings.TrimSpace(*u.Color)
		if len(trimmed) < 1 || len(trimmed) > 30 {
			return ErrValidation
		}
		u.Color = &trimmed
	}
	return nil
}

type PatientSmartCard struct {
	ID              string    `json:"id"`
	PatientID       string    `json:"patientId"`
	PatientName     string    `json:"patientName,omitempty"`
	PatientUniqueID string    `json:"patientUniqueId,omitempty"`
	MRN             string    `json:"mrn,omitempty"`
	BloodGroup      string    `json:"bloodGroup,omitempty"`
	CardNumber      string    `json:"cardNumber"`
	QRToken         string    `json:"qrToken"`
	TemplateName    string    `json:"templateName"`
	TemplateID      string    `json:"templateId,omitempty"`
	Status          string    `json:"status"` // active, revoked, expired
	RevokedReason   string    `json:"revokedReason,omitempty"`
	IssuedAt        time.Time `json:"issuedAt"`
	ExpiresAt       time.Time `json:"expiresAt"`
	IssuedBy        string    `json:"issuedBy"`
}

type IssueSmartCardInput struct {
	PatientID    string `json:"patientId"`
	TemplateName string `json:"templateName"`
	ValidityDays int    `json:"validityDays"`
}

func (i *IssueSmartCardInput) Validate() error {
	i.TemplateName = strings.TrimSpace(i.TemplateName)
	if !UUIDPattern.MatchString(i.PatientID) {
		return ErrValidation
	}
	if i.TemplateName == "" {
		i.TemplateName = "standard"
	}
	if len(i.TemplateName) > 64 {
		return ErrValidation
	}
	if i.ValidityDays <= 0 || i.ValidityDays > 3650 {
		i.ValidityDays = 365 // default 1 year
	}
	return nil
}

type GenerateSmartCardsInput struct {
	TemplateID     string `json:"templateId"`
	Mode           string `json:"mode"` // "all", "single", "unassigned"
	PatientID      string `json:"patientId,omitempty"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

func (g *GenerateSmartCardsInput) Validate() error {
	g.Mode = strings.ToLower(strings.TrimSpace(g.Mode))
	g.IdempotencyKey = strings.TrimSpace(g.IdempotencyKey)
	if len(g.IdempotencyKey) > 128 {
		return ErrValidation
	}
	if !UUIDPattern.MatchString(g.TemplateID) {
		return ErrValidation
	}
	switch g.Mode {
	case "all", "unassigned":
		// valid
	case "single":
		if !UUIDPattern.MatchString(g.PatientID) {
			return ErrValidation
		}
	default:
		return ErrValidation
	}
	return nil
}

type IssuedCardRecord struct {
	ID              string    `json:"id"`
	PatientID       string    `json:"patientId"`
	PatientName     string    `json:"patientName"`
	PatientUniqueID string    `json:"patientUniqueId"`
	CardNumber      string    `json:"cardNumber"`
	QRToken         string    `json:"qrToken"`
	TemplateID      string    `json:"templateId"`
	TemplateName    string    `json:"templateName"`
	TemplateColor   string    `json:"templateColor"`
	Status          string    `json:"status"`
	IssuedAt        time.Time `json:"issuedAt"`
	ExpiresAt       time.Time `json:"expiresAt"`
	IssuedBy        string    `json:"issuedBy"`
	Email           string    `json:"email,omitempty"`
	Phone           string    `json:"phone,omitempty"`
	DOB             string    `json:"dob,omitempty"`
	BloodGroup      string    `json:"bloodGroup,omitempty"`
	Address         string    `json:"address,omitempty"`
	MRN             string    `json:"mrn,omitempty"`
}

type GenerateSmartCardsResult struct {
	GeneratedCount int                `json:"generatedCount"`
	Cards          []IssuedCardRecord `json:"cards"`
}

type ListIssuedCardsFilter struct {
	Search     string
	TemplateID string
	Page       int
	PageSize   int
}

type IssuedCardsPage struct {
	Cards      []IssuedCardRecord `json:"cards"`
	TotalCount int                `json:"totalCount"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
}

type SmartCardVerificationResult struct {
	Valid           bool      `json:"valid"`
	Status          string    `json:"status"`
	RevokedReason   string    `json:"revokedReason,omitempty"`
	CardNumber      string    `json:"cardNumber,omitempty"`
	ExpiresAt       time.Time `json:"expiresAt,omitempty"`
	PatientName     string    `json:"patientName,omitempty"`
	PatientUniqueID string    `json:"patientUniqueId,omitempty"`
	Email           string    `json:"email,omitempty"`
	Phone           string    `json:"phone,omitempty"`
	DOB             string    `json:"dob,omitempty"`
	BloodGroup      string    `json:"bloodGroup,omitempty"`
	Address         string    `json:"address,omitempty"`
}

// ─── Patient Duplicates & Merge ───────────────────────────────────────────────

type DuplicatePatientCandidate struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Gender    string `json:"gender"`
	MatchRule string `json:"matchRule"` // "exact_phone", "exact_email", "name_dob"
}

type MergePatientInput struct {
	PrimaryPatientID string `json:"primaryPatientId"`
	MergedPatientID  string `json:"mergedPatientId"`
	Reason           string `json:"reason"`
}

type PatientIdentity struct {
	ID        string `json:"id"`
	PrimaryID string `json:"primaryId"`
	MRN       string `json:"mrn"`
	Name      string `json:"name"`
}

func (m *MergePatientInput) Validate() error {
	m.Reason = strings.TrimSpace(m.Reason)
	if !UUIDPattern.MatchString(m.PrimaryPatientID) || !UUIDPattern.MatchString(m.MergedPatientID) {
		return ErrValidation
	}
	if m.PrimaryPatientID == m.MergedPatientID {
		return ErrValidation
	}
	if len([]rune(m.Reason)) < 1 || len([]rune(m.Reason)) > 500 {
		return ErrValidation
	}
	return nil
}
