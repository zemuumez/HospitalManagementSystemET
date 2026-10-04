package domain

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

var ErrForbidden = errors.New("forbidden")
var ErrValidation = errors.New("validation failed")
var ErrConflict = errors.New("idempotency key conflict")

type Actor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func (a Actor) Can(permission string) bool {
	switch permission {
	case "diagnostics.catalog":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "lab_technician"
	case "diagnostics.read":
		return a.Role == "admin" || a.Role == "doctor" || a.Role == "lab_technician" || a.Role == "patient"

	case "pharmacy.catalog":
		return a.Role == "admin" || a.Role == "pharmacist" || a.Role == "doctor"
	case "pharmacy.manage":
		return a.Role == "admin" || a.Role == "pharmacist"
	case "medication.read":
		return a.Role == "admin" || a.Role == "pharmacist" || a.Role == "doctor" || a.Role == "patient"

	case "billing.read":
		return a.Role == "admin" || a.Role == "accountant" || a.Role == "patient"
	case "billing.manage":
		return a.Role == "admin" || a.Role == "accountant"
	case "beds.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "nurse"
	case "clinical.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "patient"
	case "clinical.admit":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor"

	case "appointments.read", "appointments.book":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "patient"
	case "staff.manage":
		return a.Role == "admin"
	case "patients.read":
		return a.Role == "admin" || a.Role == "receptionist" || a.Role == "doctor" || a.Role == "patient"
	case "patients.create":
		return a.Role == "admin" || a.Role == "receptionist"
	case "messages.manage":
		return a.Role == "admin" || a.Role == "receptionist"
	}
	return false
}
func (a Actor) Permissions() []string {
	result := []string{}
	for _, p := range []string{"diagnostics.catalog", "diagnostics.read", "pharmacy.catalog", "pharmacy.manage", "medication.read", "patients.read", "patients.create", "messages.manage", "appointments.read", "appointments.book", "staff.manage", "beds.read", "clinical.read", "clinical.admit", "billing.read", "billing.manage"} {
		if a.Can(p) {
			result = append(result, p)
		}
	}
	return result
}

type PatientInput struct {
	GivenName   string `json:"givenName"`
	FamilyName  string `json:"familyName"`
	DateOfBirth string `json:"dateOfBirth"`
	Phone       string `json:"phone"`
}

var PhonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

func (p *PatientInput) Validate(now time.Time) error {
	p.GivenName = strings.TrimSpace(p.GivenName)
	p.FamilyName = strings.TrimSpace(p.FamilyName)
	p.Phone = strings.TrimSpace(p.Phone)
	dob, err := time.Parse("2006-01-02", p.DateOfBirth)
	if err != nil || dob.After(now) || dob.Before(now.AddDate(-150, 0, 0)) {
		return ErrValidation
	}
	if len([]rune(p.GivenName)) < 1 || len([]rune(p.GivenName)) > 80 || len([]rune(p.FamilyName)) < 1 || len([]rune(p.FamilyName)) > 80 {
		return ErrValidation
	}
	if p.Phone != "" && !PhonePattern.MatchString(p.Phone) {
		return ErrValidation
	}
	return nil
}

type Patient struct {
	UserID      string `json:"userId"`
	ClinicianID string `json:"clinicianId"`
	ID          string `json:"id"`
	MRN         string `json:"mrn"`
	PatientInput
	CreatedAt time.Time `json:"createdAt"`
}
type PatientAccess struct {
	UserID      string `json:"userId"`
	ClinicianID string `json:"clinicianId"`
}
type Overview struct {
	PatientCount    int `json:"patientCount"`
	RegisteredToday int `json:"registeredToday"`
}
type MessageInput struct {
	Channel   string `json:"channel"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

func (m *MessageInput) Validate() error {
	m.Recipient = strings.TrimSpace(m.Recipient)
	m.Subject = strings.TrimSpace(m.Subject)
	m.Body = strings.TrimSpace(m.Body)
	if len(m.Body) == 0 || len(m.Body) > 4000 || len(m.Subject) > 200 || strings.ContainsAny(m.Subject, "\r\n") {
		return ErrValidation
	}
	switch m.Channel {
	case "sms":
		if !PhonePattern.MatchString(m.Recipient) || len([]rune(m.Body)) > 480 {
			return ErrValidation
		}
	case "email":
		a, err := mail.ParseAddress(m.Recipient)
		if err != nil || a.Address != m.Recipient || m.Subject == "" {
			return ErrValidation
		}
	default:
		return ErrValidation
	}
	return nil
}

type Message struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	Recipient string    `json:"recipient"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}
