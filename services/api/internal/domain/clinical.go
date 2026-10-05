package domain

import (
	"strings"
	"time"
)

type BedInput struct {
	TypeID      string `json:"typeId"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	ChargeMinor int64  `json:"chargeMinor"`
}

func (b *BedInput) Validate() error {
	b.Name = strings.TrimSpace(b.Name)
	b.Type = strings.TrimSpace(b.Type)
	if len([]rune(b.Name)) < 1 || len([]rune(b.Name)) > 80 || (b.TypeID == "" && len([]rune(b.Type)) < 1) || (b.TypeID != "" && !UUIDPattern.MatchString(b.TypeID)) || len([]rune(b.Type)) > 80 || b.ChargeMinor < 0 || b.ChargeMinor > 1000000000 {
		return ErrValidation
	}
	return nil
}

type Bed struct {
	BedInput
	ID        string `json:"id"`
	Available bool   `json:"available"`
	State     string `json:"state"`
	Version   int    `json:"version"`
}
type CaseInput struct {
	PatientID   string `json:"patientId"`
	DoctorID    string `json:"doctorId"`
	Description string `json:"description"`
}

func (c *CaseInput) Validate() error {
	c.Description = strings.TrimSpace(c.Description)
	if !UUIDPattern.MatchString(c.PatientID) || c.DoctorID == "" || len(c.DoctorID) > 128 || len([]rune(c.Description)) > 2000 {
		return ErrValidation
	}
	return nil
}

type Case struct {
	CaseInput
	ID          string `json:"id"`
	Number      int64  `json:"number"`
	PatientName string `json:"patientName"`
	DoctorName  string `json:"doctorName"`
}

// EncounterIntake retains the original IPD/OPD registration fields. Charges here
// are registration quotations; posting an invoice remains a separate ledger action.
type EncounterIntake struct {
	// These are the original deployment's Tel and TAX registration text fields,
	// not a payment instruction or a calculated tax amount.
	Telephone            string  `json:"telephone"`
	TaxReference         string  `json:"taxReference"`
	Height               float64 `json:"height"`
	Weight               float64 `json:"weight"`
	BloodPressure        string  `json:"bloodPressure"`
	Notes                string  `json:"notes"`
	IdentificationNumber string  `json:"identificationNumber"`
	Reference            string  `json:"reference"`
	OldPatient           bool    `json:"oldPatient"`
	StandardChargeMinor  int64   `json:"standardChargeMinor"`
	PaymentMode          string  `json:"paymentMode"`
}

func (i *EncounterIntake) Validate() error {
	i.BloodPressure = strings.TrimSpace(i.BloodPressure)
	i.Notes = strings.TrimSpace(i.Notes)
	i.IdentificationNumber = strings.TrimSpace(i.IdentificationNumber)
	i.Reference = strings.TrimSpace(i.Reference)
	if len(i.Telephone) > 80 || len(i.TaxReference) > 200 || i.Height < 0 || i.Height > 300 || i.Weight < 0 || i.Weight > 1000 || len(i.BloodPressure) > 40 || len([]rune(i.Notes)) > 4000 || len([]rune(i.IdentificationNumber)) > 200 || len([]rune(i.Reference)) > 200 || i.StandardChargeMinor < 0 || i.StandardChargeMinor > 1000000000 {
		return ErrValidation
	}
	switch i.PaymentMode {
	case "", "cash", "bank_transfer", "card", "other":
	default:
		return ErrValidation
	}
	return nil
}

type EncounterInput struct {
	Intake     EncounterIntake `json:"intake"`
	Kind       string          `json:"kind"`
	CaseID     string          `json:"caseId"`
	BedID      string          `json:"bedId"`
	AdmittedAt time.Time       `json:"admittedAt"`
	Symptoms   string          `json:"symptoms"`
}

func (e *EncounterInput) Validate(now time.Time) error {
	if err := e.Intake.Validate(); err != nil {
		return err
	}
	e.AdmittedAt = e.AdmittedAt.UTC().Truncate(time.Microsecond)
	e.Symptoms = strings.TrimSpace(e.Symptoms)
	if !UUIDPattern.MatchString(e.CaseID) || (e.Kind != "ipd" && e.Kind != "opd") || (e.Kind == "ipd" && !UUIDPattern.MatchString(e.BedID)) || (e.Kind == "opd" && e.BedID != "") || e.AdmittedAt.After(now) || e.AdmittedAt.Before(now.AddDate(-1, 0, 0)) || len([]rune(e.Symptoms)) > 2000 {
		return ErrValidation
	}
	return nil
}

type Encounter struct {
	PatientEmail string `json:"patientEmail"`
	DoctorEmail  string `json:"doctorEmail"`
	BillStatus   string `json:"billStatus"`
	TotalVisits  int    `json:"totalVisits"`
	EncounterInput
	ID               string     `json:"id"`
	Number           int64      `json:"number"`
	PatientID        string     `json:"patientId"`
	PatientName      string     `json:"patientName"`
	DoctorID         string     `json:"doctorId"`
	DoctorName       string     `json:"doctorName"`
	BedName          string     `json:"bedName"`
	Status           string     `json:"status"`
	Version          int        `json:"version"`
	DischargedAt     *time.Time `json:"dischargedAt"`
	DischargeSummary string     `json:"dischargeSummary"`
}
type Discharge struct {
	Version int    `json:"version"`
	Summary string `json:"summary"`
}
type NoteInput struct {
	Body string `json:"body"`
}
type ClinicalNote struct {
	ID         string    `json:"id"`
	AuthorName string    `json:"authorName"`
	Body       string    `json:"body"`
	SignedAt   time.Time `json:"signedAt"`
}

type BedStateInput struct {
	State   string `json:"state"`
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}
type BedTransfer struct {
	BedID   string `json:"bedId"`
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}
type BedEvent struct {
	ID          string    `json:"id"`
	Kind        string    `json:"kind"`
	FromBedID   string    `json:"fromBedId"`
	ToBedID     string    `json:"toBedId"`
	ChargeMinor int64     `json:"chargeMinor"`
	ActorID     string    `json:"actorId"`
	Reason      string    `json:"reason"`
	RecordedAt  time.Time `json:"recordedAt"`
	Version     int       `json:"version"`
}

type BedTypeInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	Version     int    `json:"version"`
}
type BedType struct {
	BedTypeInput
	ID string `json:"id"`
}
