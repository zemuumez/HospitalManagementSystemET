package application

import (
	"context"
	"strings"
	"time"

	"hms.local/api/internal/domain"
)

type AppointmentOpsRepository interface {
	// Patient Queue
	DoctorQueue(ctx context.Context, doctorID string, queueDate string) ([]domain.PatientQueueItem, error)
	EnqueuePatient(ctx context.Context, a domain.Actor, input domain.EnqueuePatientInput) (domain.PatientQueueItem, error)
	UpdateQueueStatus(ctx context.Context, a domain.Actor, queueID string, input domain.UpdateQueueStatusInput) (domain.PatientQueueItem, error)

	// Public Appointment Requests
	CreatePublicAppointmentRequest(ctx context.Context, input domain.CreatePublicAppointmentRequestInput) (domain.PublicAppointmentRequest, error)
	ListPublicAppointmentRequests(ctx context.Context, status string, page int) ([]domain.PublicAppointmentRequest, error)
	ReviewPublicAppointmentRequest(ctx context.Context, a domain.Actor, id string, input domain.ReviewPublicAppointmentRequestInput) (domain.PublicAppointmentRequest, error)

	// Appointment Billing
	AppointmentBilling(ctx context.Context, appointmentID string) (domain.AppointmentBillingRecord, error)
	SetAppointmentFee(ctx context.Context, a domain.Actor, appointmentID string, feeMinor int64) (domain.AppointmentBillingRecord, error)
	LinkAppointmentInvoice(ctx context.Context, a domain.Actor, appointmentID string, invoiceID string) (domain.AppointmentBillingRecord, error)
}

type AppointmentOpsService struct {
	Store AppointmentOpsRepository
}

// ─── Patient Queue ────────────────────────────────────────────────────────────

func (s AppointmentOpsService) DoctorQueue(ctx context.Context, a domain.Actor, doctorID string, queueDate string) ([]domain.PatientQueueItem, error) {
	if !a.Can("appointments.read") {
		return nil, domain.ErrForbidden
	}
	doctorID = strings.TrimSpace(doctorID)
	queueDate = strings.TrimSpace(queueDate)
	if doctorID == "" {
		return nil, domain.ErrValidation
	}
	if _, err := time.Parse("2006-01-02", queueDate); err != nil {
		return nil, domain.ErrValidation
	}
	return s.Store.DoctorQueue(ctx, doctorID, queueDate)
}

func (s AppointmentOpsService) EnqueuePatient(ctx context.Context, a domain.Actor, input domain.EnqueuePatientInput) (domain.PatientQueueItem, error) {
	if !a.Can("appointments.book") {
		return domain.PatientQueueItem{}, domain.ErrForbidden
	}
	if err := input.Validate(); err != nil {
		return domain.PatientQueueItem{}, err
	}
	return s.Store.EnqueuePatient(ctx, a, input)
}

func (s AppointmentOpsService) UpdateQueueStatus(ctx context.Context, a domain.Actor, queueID string, input domain.UpdateQueueStatusInput) (domain.PatientQueueItem, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "receptionist" && a.Role != "nurse" {
		return domain.PatientQueueItem{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(queueID) {
		return domain.PatientQueueItem{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.PatientQueueItem{}, err
	}
	return s.Store.UpdateQueueStatus(ctx, a, queueID, input)
}

// ─── Public Appointment Requests ──────────────────────────────────────────────

func (s AppointmentOpsService) CreatePublicRequest(ctx context.Context, input domain.CreatePublicAppointmentRequestInput) (domain.PublicAppointmentRequest, error) {
	if err := input.Validate(); err != nil {
		return domain.PublicAppointmentRequest{}, err
	}
	return s.Store.CreatePublicAppointmentRequest(ctx, input)
}

func (s AppointmentOpsService) ListPublicRequests(ctx context.Context, a domain.Actor, status string, page int) ([]domain.PublicAppointmentRequest, error) {
	if !a.Can("appointments.read") {
		return nil, domain.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	return s.Store.ListPublicAppointmentRequests(ctx, status, page)
}

func (s AppointmentOpsService) ReviewPublicRequest(ctx context.Context, a domain.Actor, id string, input domain.ReviewPublicAppointmentRequestInput) (domain.PublicAppointmentRequest, error) {
	if !a.Can("appointments.book") {
		return domain.PublicAppointmentRequest{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.PublicAppointmentRequest{}, domain.ErrValidation
	}
	if err := input.Validate(); err != nil {
		return domain.PublicAppointmentRequest{}, err
	}
	return s.Store.ReviewPublicAppointmentRequest(ctx, a, id, input)
}

// ─── Appointment Billing ──────────────────────────────────────────────────────

func (s AppointmentOpsService) GetAppointmentBilling(ctx context.Context, a domain.Actor, appointmentID string) (domain.AppointmentBillingRecord, error) {
	if !a.Can("billing.read") && !a.Can("appointments.read") {
		return domain.AppointmentBillingRecord{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(appointmentID) {
		return domain.AppointmentBillingRecord{}, domain.ErrValidation
	}
	return s.Store.AppointmentBilling(ctx, appointmentID)
}

func (s AppointmentOpsService) SetAppointmentFee(ctx context.Context, a domain.Actor, appointmentID string, feeMinor int64) (domain.AppointmentBillingRecord, error) {
	if !a.Can("billing.manage") {
		return domain.AppointmentBillingRecord{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(appointmentID) || feeMinor < 0 {
		return domain.AppointmentBillingRecord{}, domain.ErrValidation
	}
	return s.Store.SetAppointmentFee(ctx, a, appointmentID, feeMinor)
}

func (s AppointmentOpsService) LinkAppointmentInvoice(ctx context.Context, a domain.Actor, appointmentID string, invoiceID string) (domain.AppointmentBillingRecord, error) {
	if !a.Can("billing.manage") {
		return domain.AppointmentBillingRecord{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(appointmentID) || !domain.UUIDPattern.MatchString(invoiceID) {
		return domain.AppointmentBillingRecord{}, domain.ErrValidation
	}
	return s.Store.LinkAppointmentInvoice(ctx, a, appointmentID, invoiceID)
}
