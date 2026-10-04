package domain

import (
	"net/mail"
	"strings"
	"time"
)

type PatientProfileInput struct {
	PatientInput
	Email                 string `json:"email"`
	Gender                string `json:"gender"`
	BloodGroup            string `json:"bloodGroup"`
	Address1              string `json:"address1"`
	Address2              string `json:"address2"`
	City                  string `json:"city"`
	Region                string `json:"region"`
	Country               string `json:"country"`
	PostalCode            string `json:"postalCode"`
	EmergencyName         string `json:"emergencyName"`
	EmergencyPhone        string `json:"emergencyPhone"`
	EmergencyRelationship string `json:"emergencyRelationship"`
	Active                bool   `json:"active"`
	SMSConsent            bool   `json:"smsConsent"`
	EmailConsent          bool   `json:"emailConsent"`
	Version               int    `json:"version"`
	Reason                string `json:"reason"`
}
type PatientProfile struct {
	PatientProfileInput
	ID  string `json:"id"`
	MRN string `json:"mrn"`
}
type ProfileRevision struct {
	Version   int            `json:"version"`
	ActorID   string         `json:"actorId"`
	Reason    string         `json:"reason"`
	Before    PatientProfile `json:"before"`
	After     PatientProfile `json:"after"`
	CreatedAt time.Time      `json:"createdAt"`
}

func (p *PatientProfileInput) Validate(now time.Time) error {
	if e := p.PatientInput.Validate(now); e != nil {
		return e
	}
	p.Email = strings.TrimSpace(p.Email)
	if p.Email != "" {
		email, e := mail.ParseAddress(p.Email)
		if e != nil || email.Address != p.Email || len(p.Email) > 254 {
			return ErrValidation
		}
	}
	switch p.Gender {
	case "unknown", "female", "male", "other":
	default:
		return ErrValidation
	}
	switch p.BloodGroup {
	case "", "A+", "A-", "B+", "B-", "AB+", "AB-", "O+", "O-":
	default:
		return ErrValidation
	}
	for _, f := range []struct {
		s   *string
		max int
	}{{&p.Address1, 250}, {&p.Address2, 250}, {&p.City, 100}, {&p.Region, 100}, {&p.Country, 100}, {&p.PostalCode, 30}, {&p.EmergencyName, 150}, {&p.EmergencyPhone, 20}, {&p.EmergencyRelationship, 100}} {
		if !validText(f.s, 0, f.max) {
			return ErrValidation
		}
	}
	if p.EmergencyPhone != "" && !PhonePattern.MatchString(p.EmergencyPhone) {
		return ErrValidation
	}
	if p.Version < 1 || p.Version > 1000000000 || !validText(&p.Reason, 1, 500) {
		return ErrValidation
	}
	return nil
}
