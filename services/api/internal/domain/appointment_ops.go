package domain

import (
	"strings"
	"time"
)

// ─── Patient Queue & Tokens ───────────────────────────────────────────────────

type PatientQueueItem struct {
	ID            string     `json:"id"`
	DoctorID      string     `json:"doctorId"`
	DoctorName    string     `json:"doctorName,omitempty"`
	PatientID     string     `json:"patientId"`
	PatientName   string     `json:"patientName,omitempty"`
	AppointmentID *string    `json:"appointmentId,omitempty"`
	QueueDate     string     `json:"queueDate"` // YYYY-MM-DD
	TokenNumber   int        `json:"tokenNumber"`
	Status        string     `json:"status"` // waiting, in_consultation, completed, skipped
	Notes         string     `json:"notes"`
	CreatedAt     time.Time  `json:"createdAt"`
	CalledAt      *time.Time `json:"calledAt,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
}

type EnqueuePatientInput struct {
	DoctorID      string  `json:"doctorId"`
	PatientID     string  `json:"patientId"`
	AppointmentID *string `json:"appointmentId,omitempty"`
	QueueDate     string  `json:"queueDate"`
	Notes         string  `json:"notes"`
}

func (i *EnqueuePatientInput) Validate() error {
	i.DoctorID = strings.TrimSpace(i.DoctorID)
	i.QueueDate = strings.TrimSpace(i.QueueDate)
	i.Notes = strings.TrimSpace(i.Notes)
	if i.DoctorID == "" || !UUIDPattern.MatchString(i.PatientID) {
		return ErrValidation
	}
	if i.AppointmentID != nil && !UUIDPattern.MatchString(*i.AppointmentID) {
		return ErrValidation
	}
	if _, err := time.Parse("2006-01-02", i.QueueDate); err != nil {
		return ErrValidation
	}
	if len(i.Notes) > 1000 {
		return ErrValidation
	}
	return nil
}

type UpdateQueueStatusInput struct {
	Status string `json:"status"` // in_consultation, completed, skipped
	Notes  string `json:"notes"`
}

func (i *UpdateQueueStatusInput) Validate() error {
	i.Status = strings.TrimSpace(i.Status)
	i.Notes = strings.TrimSpace(i.Notes)
	switch i.Status {
	case "in_consultation", "completed", "skipped":
		// valid
	default:
		return ErrValidation
	}
	if len(i.Notes) > 1000 {
		return ErrValidation
	}
	return nil
}

// ─── Public Appointment Request ───────────────────────────────────────────────

type PublicAppointmentRequest struct {
	ID              string     `json:"id"`
	PatientName     string     `json:"patientName"`
	PatientEmail    string     `json:"patientEmail"`
	PatientPhone    string     `json:"patientPhone"`
	DoctorID        string     `json:"doctorId"`
	DoctorName      string     `json:"doctorName,omitempty"`
	DepartmentID    *string    `json:"departmentId,omitempty"`
	DepartmentTitle string     `json:"departmentTitle,omitempty"`
	PreferredDate   string     `json:"preferredDate"` // YYYY-MM-DD
	Problem         string     `json:"problem"`
	Status          string     `json:"status"` // pending, confirmed, rejected
	RejectionReason string     `json:"rejectionReason,omitempty"`
	AppointmentID   *string    `json:"appointmentId,omitempty"`
	ReviewedBy      *string    `json:"reviewedBy,omitempty"`
	ReviewedAt      *time.Time `json:"reviewedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

type CreatePublicAppointmentRequestInput struct {
	PatientName   string  `json:"patientName"`
	PatientEmail  string  `json:"patientEmail"`
	PatientPhone  string  `json:"patientPhone"`
	DoctorID      string  `json:"doctorId"`
	DepartmentID  *string `json:"departmentId,omitempty"`
	PreferredDate string  `json:"preferredDate"`
	Problem       string  `json:"problem"`
}

func (i *CreatePublicAppointmentRequestInput) Validate() error {
	i.PatientName = strings.TrimSpace(i.PatientName)
	i.PatientEmail = strings.TrimSpace(i.PatientEmail)
	i.PatientPhone = strings.TrimSpace(i.PatientPhone)
	i.DoctorID = strings.TrimSpace(i.DoctorID)
	i.Problem = strings.TrimSpace(i.Problem)
	if len([]rune(i.PatientName)) < 1 || len([]rune(i.PatientName)) > 160 {
		return ErrValidation
	}
	if len(i.PatientPhone) < 6 || len(i.PatientPhone) > 64 {
		return ErrValidation
	}
	if i.DoctorID == "" {
		return ErrValidation
	}
	if i.DepartmentID != nil && !UUIDPattern.MatchString(*i.DepartmentID) {
		return ErrValidation
	}
	if _, err := time.Parse("2006-01-02", i.PreferredDate); err != nil {
		return ErrValidation
	}
	if len(i.Problem) > 2000 {
		return ErrValidation
	}
	return nil
}

type ReviewPublicAppointmentRequestInput struct {
	Action          string `json:"action"` // "confirm" or "reject"
	RejectionReason string `json:"rejectionReason,omitempty"`
	AppointmentID   string `json:"appointmentId,omitempty"` // when confirming, the created appointment ID
}

func (i *ReviewPublicAppointmentRequestInput) Validate() error {
	i.Action = strings.TrimSpace(i.Action)
	i.RejectionReason = strings.TrimSpace(i.RejectionReason)
	switch i.Action {
	case "confirm":
		if i.AppointmentID != "" && !UUIDPattern.MatchString(i.AppointmentID) {
			return ErrValidation
		}
	case "reject":
		if len(i.RejectionReason) > 500 {
			return ErrValidation
		}
	default:
		return ErrValidation
	}
	return nil
}

// ─── Appointment Billing & Fee ────────────────────────────────────────────────

type AppointmentBillingRecord struct {
	AppointmentID string    `json:"appointmentId"`
	FeeMinor      int64     `json:"feeMinor"`
	InvoiceID     *string   `json:"invoiceId,omitempty"`
	PaymentStatus string    `json:"paymentStatus"` // unpaid, partially_paid, paid, refunded
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
