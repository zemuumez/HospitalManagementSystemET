package domain

import (
	"regexp"
	"strings"
	"time"
)

var yearMadePattern = regexp.MustCompile(`^[0-9]{4}$`)

type Ambulance struct {
	ID            string    `json:"id"`
	VehicleNumber string    `json:"vehicle_number"`
	VehicleModel  string    `json:"vehicle_model"`
	YearMade      string    `json:"year_made"`
	DriverName    string    `json:"driver_name"`
	DriverLicense string    `json:"driver_license"`
	DriverContact string    `json:"driver_contact"`
	VehicleType   int       `json:"vehicle_type"` // 1: Contractual, 2: Owned
	IsAvailable   bool      `json:"is_available"`
	Note          string    `json:"note"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AmbulanceInput struct {
	VehicleNumber string `json:"vehicle_number"`
	VehicleModel  string `json:"vehicle_model"`
	YearMade      string `json:"year_made"`
	DriverName    string `json:"driver_name"`
	DriverLicense string `json:"driver_license"`
	DriverContact string `json:"driver_contact"`
	VehicleType   int    `json:"vehicle_type"`
	IsAvailable   *bool  `json:"is_available"`
	Note          string `json:"note"`
}

func (i *AmbulanceInput) Validate() error {
	i.VehicleNumber = strings.ToUpper(strings.TrimSpace(i.VehicleNumber))
	i.VehicleModel = strings.TrimSpace(i.VehicleModel)
	i.YearMade = strings.TrimSpace(i.YearMade)
	i.DriverName = strings.TrimSpace(i.DriverName)
	i.DriverLicense = strings.TrimSpace(i.DriverLicense)
	i.DriverContact = strings.TrimSpace(i.DriverContact)
	i.Note = strings.TrimSpace(i.Note)

	if i.VehicleNumber == "" || len(i.VehicleNumber) > 64 {
		return ErrValidation
	}
	if i.VehicleModel == "" || len(i.VehicleModel) > 128 {
		return ErrValidation
	}
	if !yearMadePattern.MatchString(i.YearMade) {
		return ErrValidation
	}
	if i.DriverName == "" || len(i.DriverName) > 128 {
		return ErrValidation
	}
	if i.DriverLicense == "" || len(i.DriverLicense) > 64 {
		return ErrValidation
	}
	if i.DriverContact == "" || len(i.DriverContact) > 32 {
		return ErrValidation
	}
	if i.VehicleType != 1 && i.VehicleType != 2 {
		return ErrValidation
	}
	return nil
}

type AmbulanceCall struct {
	ID             string    `json:"id"`
	AmbulanceID    string    `json:"ambulance_id"`
	VehicleModel   string    `json:"vehicle_model,omitempty"`
	VehicleNumber  string    `json:"vehicle_number,omitempty"`
	PatientID      string    `json:"patient_id"`
	PatientName    string    `json:"patient_name,omitempty"`
	DriverName     string    `json:"driver_name"`
	CallDate       time.Time `json:"call_date"`
	AmountMinor    int64     `json:"amount_minor"`
	Status         string    `json:"status"` // "dispatched", "completed", "cancelled"
	PickupLocation string    `json:"pickup_location"`
	Destination    string    `json:"destination"`
	Notes          string    `json:"notes"`
	InvoiceID      *string   `json:"invoice_id,omitempty"`
	Version        int       `json:"version"`
	CreatedBy      string    `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AmbulanceCallInput struct {
	AmbulanceID    string    `json:"ambulance_id"`
	PatientID      string    `json:"patient_id"`
	DriverName     string    `json:"driver_name"`
	CallDate       time.Time `json:"call_date"`
	AmountMinor    int64     `json:"amount_minor"`
	PickupLocation string    `json:"pickup_location"`
	Destination    string    `json:"destination"`
	Notes          string    `json:"notes"`
}

func (i *AmbulanceCallInput) Validate() error {
	i.AmbulanceID = strings.TrimSpace(i.AmbulanceID)
	i.PatientID = strings.TrimSpace(i.PatientID)
	i.DriverName = strings.TrimSpace(i.DriverName)
	i.PickupLocation = strings.TrimSpace(i.PickupLocation)
	i.Destination = strings.TrimSpace(i.Destination)
	i.Notes = strings.TrimSpace(i.Notes)

	if !UUIDPattern.MatchString(i.AmbulanceID) {
		return ErrValidation
	}
	if !UUIDPattern.MatchString(i.PatientID) {
		return ErrValidation
	}
	if i.CallDate.IsZero() {
		return ErrValidation
	}
	if i.AmountMinor < 0 {
		return ErrValidation
	}
	return nil
}

type AmbulanceCallUpdateInput struct {
	AmbulanceID    string    `json:"ambulance_id"`
	DriverName     string    `json:"driver_name"`
	CallDate       time.Time `json:"call_date"`
	AmountMinor    int64     `json:"amount_minor"`
	Status         string    `json:"status"`
	PickupLocation string    `json:"pickup_location"`
	Destination    string    `json:"destination"`
	Notes          string    `json:"notes"`
	Version        int       `json:"version"`
}

func (i *AmbulanceCallUpdateInput) Validate() error {
	i.AmbulanceID = strings.TrimSpace(i.AmbulanceID)
	i.DriverName = strings.TrimSpace(i.DriverName)
	i.Status = strings.ToLower(strings.TrimSpace(i.Status))
	i.PickupLocation = strings.TrimSpace(i.PickupLocation)
	i.Destination = strings.TrimSpace(i.Destination)
	i.Notes = strings.TrimSpace(i.Notes)

	if !UUIDPattern.MatchString(i.AmbulanceID) {
		return ErrValidation
	}
	if i.CallDate.IsZero() {
		return ErrValidation
	}
	if i.AmountMinor < 0 {
		return ErrValidation
	}
	if i.Status != "dispatched" && i.Status != "completed" && i.Status != "cancelled" {
		return ErrValidation
	}
	if i.Version <= 0 {
		return ErrValidation
	}
	return nil
}
