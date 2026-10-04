package domain

import (
	"math"
	"strings"
	"time"
)

type NurseAssignment struct {
	NurseID string `json:"nurseId"`
	Active  bool   `json:"active"`
	Version int    `json:"version"`
	Reason  string `json:"reason"`
}
type EncounterNurse struct {
	NurseID string `json:"nurseId"`
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	Version int    `json:"version"`
}
type Measurements struct {
	HeightCm        *float64 `json:"heightCm,omitempty"`
	WeightKg        *float64 `json:"weightKg,omitempty"`
	TemperatureC    *float64 `json:"temperatureC,omitempty"`
	SystolicMmHg    *float64 `json:"systolicMmHg,omitempty"`
	DiastolicMmHg   *float64 `json:"diastolicMmHg,omitempty"`
	PulseBpm        *float64 `json:"pulseBpm,omitempty"`
	RespiratoryRate *float64 `json:"respiratoryRate,omitempty"`
	OxygenPercent   *float64 `json:"oxygenPercent,omitempty"`
}
type VitalsInput struct {
	ObservedAt       time.Time    `json:"observedAt"`
	Measurements     Measurements `json:"measurements"`
	Note             string       `json:"note"`
	CorrectionOf     string       `json:"correctionOf"`
	CorrectionReason string       `json:"correctionReason"`
}

func (i *VitalsInput) Validate(now time.Time) error {
	i.ObservedAt = i.ObservedAt.UTC().Truncate(time.Microsecond)
	i.Note = strings.TrimSpace(i.Note)
	i.CorrectionReason = strings.TrimSpace(i.CorrectionReason)
	if i.ObservedAt.IsZero() || i.ObservedAt.After(now) || len([]rune(i.Note)) > 2000 || len([]rune(i.CorrectionReason)) > 2000 {
		return ErrValidation
	}
	if (i.CorrectionOf == "" && i.CorrectionReason != "") || (i.CorrectionOf != "" && (!UUIDPattern.MatchString(i.CorrectionOf) || i.CorrectionReason == "")) {
		return ErrValidation
	}
	count := 0
	for _, value := range []*float64{i.Measurements.HeightCm, i.Measurements.WeightKg, i.Measurements.TemperatureC, i.Measurements.SystolicMmHg, i.Measurements.DiastolicMmHg, i.Measurements.PulseBpm, i.Measurements.RespiratoryRate, i.Measurements.OxygenPercent} {
		if value != nil {
			count++
			if math.IsNaN(*value) || math.IsInf(*value, 0) || *value < 0 || *value > 100000 {
				return ErrValidation
			}
		}
	}
	if count == 0 || (i.Measurements.OxygenPercent != nil && *i.Measurements.OxygenPercent > 100) {
		return ErrValidation
	}
	if (i.Measurements.SystolicMmHg == nil) != (i.Measurements.DiastolicMmHg == nil) {
		return ErrValidation
	}
	return nil
}

type Vitals struct {
	VitalsInput
	ID       string    `json:"id"`
	AuthorID string    `json:"authorId"`
	SignedAt time.Time `json:"signedAt"`
}
