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

// ─── Patient Smart Card ───────────────────────────────────────────────────────

type PatientSmartCard struct {
	ID            string    `json:"id"`
	PatientID     string    `json:"patientId"`
	PatientName   string    `json:"patientName,omitempty"`
	MRN           string    `json:"mrn,omitempty"`
	BloodGroup    string    `json:"bloodGroup,omitempty"`
	CardNumber    string    `json:"cardNumber"`
	QRToken       string    `json:"qrToken"`
	TemplateName  string    `json:"templateName"`
	Status        string    `json:"status"` // active, revoked, expired
	RevokedReason string    `json:"revokedReason,omitempty"`
	IssuedAt      time.Time `json:"issuedAt"`
	ExpiresAt     time.Time `json:"expiresAt"`
	IssuedBy      string    `json:"issuedBy"`
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
