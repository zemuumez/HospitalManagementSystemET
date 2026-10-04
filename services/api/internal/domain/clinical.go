package domain

import (
	"strings"
	"time"
)

type BedInput struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	ChargeMinor int64  `json:"chargeMinor"`
}

func (b *BedInput) Validate() error {
	b.Name = strings.TrimSpace(b.Name)
	b.Type = strings.TrimSpace(b.Type)
	if len([]rune(b.Name)) < 1 || len([]rune(b.Name)) > 80 || len([]rune(b.Type)) < 1 || len([]rune(b.Type)) > 80 || b.ChargeMinor < 0 || b.ChargeMinor > 1000000000 {
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
type EncounterInput struct {
	Kind       string    `json:"kind"`
	CaseID     string    `json:"caseId"`
	BedID      string    `json:"bedId"`
	AdmittedAt time.Time `json:"admittedAt"`
	Symptoms   string    `json:"symptoms"`
}

func (e *EncounterInput) Validate(now time.Time) error {
	e.AdmittedAt = e.AdmittedAt.UTC().Truncate(time.Microsecond)
	e.Symptoms = strings.TrimSpace(e.Symptoms)
	if !UUIDPattern.MatchString(e.CaseID) || (e.Kind != "ipd" && e.Kind != "opd") || (e.Kind == "ipd" && !UUIDPattern.MatchString(e.BedID)) || (e.Kind == "opd" && e.BedID != "") || e.AdmittedAt.After(now) || e.AdmittedAt.Before(now.AddDate(-1, 0, 0)) || len([]rune(e.Symptoms)) > 2000 {
		return ErrValidation
	}
	return nil
}

type Encounter struct {
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
