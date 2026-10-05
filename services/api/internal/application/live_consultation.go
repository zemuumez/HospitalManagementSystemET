package application

import (
	"context"
	"hms.local/api/internal/domain"
	"time"
)

type LiveConsultationStore interface {
	ValidateClinicalAttribution(context.Context, domain.Actor, string, string, *string) error
	LiveConsultations(context.Context, string, string, *int, int) ([]domain.LiveConsultation, int, error)
	LiveConsultation(context.Context, string) (domain.LiveConsultation, error)
	CreateLiveConsultation(context.Context, domain.Actor, domain.LiveConsultationInput) (domain.LiveConsultation, error)
	UpdateLiveConsultationStatus(context.Context, domain.Actor, string, int) (domain.LiveConsultation, error)

	LiveMeetings(context.Context, *int, int) ([]domain.LiveMeeting, int, error)
	LiveMeeting(context.Context, string) (domain.LiveMeeting, error)
	CreateLiveMeeting(context.Context, domain.Actor, domain.LiveMeetingInput) (domain.LiveMeeting, error)
	UpdateLiveMeetingStatus(context.Context, domain.Actor, string, int) (domain.LiveMeeting, error)

	ProviderSetting(context.Context, string) (domain.LiveProviderSetting, error)
	UpdateProviderSetting(context.Context, domain.Actor, domain.LiveProviderSettingInput) (domain.LiveProviderSetting, error)
}

type LiveConsultationService struct {
	Store LiveConsultationStore
	Now   func() time.Time
}

// --- Live Consultations ---

func (s LiveConsultationService) LiveConsultations(ctx context.Context, a domain.Actor, page int, statusFilter *int) ([]domain.LiveConsultation, int, error) {
	if !a.Can("live_consultations.read") {
		return nil, 0, domain.ErrForbidden
	}

	doctorFilter := ""
	patientFilter := ""

	if a.Role == "patient" {
		patientFilter = a.ID
	} else if a.Role == "doctor" {
		doctorFilter = a.ID
	}

	return s.Store.LiveConsultations(ctx, doctorFilter, patientFilter, statusFilter, page)
}

func (s LiveConsultationService) LiveConsultation(ctx context.Context, a domain.Actor, id string) (domain.LiveConsultation, error) {
	if !a.Can("live_consultations.read") {
		return domain.LiveConsultation{}, domain.ErrForbidden
	}
	c, err := s.Store.LiveConsultation(ctx, id)
	if err != nil {
		return domain.LiveConsultation{}, err
	}
	if a.Role == "patient" && c.PatientID != a.ID && (c.PatientUserID == nil || *c.PatientUserID != a.ID) {
		return domain.LiveConsultation{}, domain.ErrForbidden
	}
	if a.Role == "doctor" && c.DoctorID != a.ID && c.CreatedBy != a.ID {
		return domain.LiveConsultation{}, domain.ErrForbidden
	}
	return c, nil
}

func (s LiveConsultationService) CreateLiveConsultation(ctx context.Context, a domain.Actor, in domain.LiveConsultationInput) (domain.LiveConsultation, error) {
	if !a.Can("live_consultations.manage") {
		return domain.LiveConsultation{}, domain.ErrForbidden
	}
	if a.Role == "doctor" && in.DoctorID == "" {
		in.DoctorID = a.ID
	}
	if err := in.Validate(); err != nil {
		return domain.LiveConsultation{}, err
	}
	if err := s.Store.ValidateClinicalAttribution(ctx, a, in.PatientID, in.DoctorID, in.EncounterID); err != nil {
		return domain.LiveConsultation{}, err
	}
	return s.Store.CreateLiveConsultation(ctx, a, in)
}

func (s LiveConsultationService) UpdateLiveConsultationStatus(ctx context.Context, a domain.Actor, id string, status int) (domain.LiveConsultation, error) {
	if !a.Can("live_consultations.manage") {
		return domain.LiveConsultation{}, domain.ErrForbidden
	}
	if status < 0 || status > 2 {
		return domain.LiveConsultation{}, domain.ErrValidation
	}
	if _, err := s.LiveConsultation(ctx, a, id); err != nil {
		return domain.LiveConsultation{}, err
	}
	return s.Store.UpdateLiveConsultationStatus(ctx, a, id, status)
}

// --- Live Meetings ---

func (s LiveConsultationService) LiveMeetings(ctx context.Context, a domain.Actor, page int, statusFilter *int) ([]domain.LiveMeeting, int, error) {
	if !a.Can("live_meetings.read") {
		return nil, 0, domain.ErrForbidden
	}
	return s.Store.LiveMeetings(ctx, statusFilter, page)
}

func (s LiveConsultationService) LiveMeeting(ctx context.Context, a domain.Actor, id string) (domain.LiveMeeting, error) {
	if !a.Can("live_meetings.read") {
		return domain.LiveMeeting{}, domain.ErrForbidden
	}
	return s.Store.LiveMeeting(ctx, id)
}

func (s LiveConsultationService) CreateLiveMeeting(ctx context.Context, a domain.Actor, in domain.LiveMeetingInput) (domain.LiveMeeting, error) {
	if !a.Can("live_meetings.manage") {
		return domain.LiveMeeting{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.LiveMeeting{}, err
	}
	return s.Store.CreateLiveMeeting(ctx, a, in)
}

func (s LiveConsultationService) UpdateLiveMeetingStatus(ctx context.Context, a domain.Actor, id string, status int) (domain.LiveMeeting, error) {
	if !a.Can("live_meetings.manage") {
		return domain.LiveMeeting{}, domain.ErrForbidden
	}
	if status < 0 || status > 2 {
		return domain.LiveMeeting{}, domain.ErrValidation
	}
	return s.Store.UpdateLiveMeetingStatus(ctx, a, id, status)
}

// --- Provider Settings ---

func (s LiveConsultationService) ProviderSetting(ctx context.Context, a domain.Actor) (domain.LiveProviderSetting, error) {
	if !a.Can("live_consultations.manage") {
		return domain.LiveProviderSetting{}, domain.ErrForbidden
	}
	return s.Store.ProviderSetting(ctx, a.ID)
}

func (s LiveConsultationService) UpdateProviderSetting(ctx context.Context, a domain.Actor, in domain.LiveProviderSettingInput) (domain.LiveProviderSetting, error) {
	if !a.Can("live_consultations.manage") {
		return domain.LiveProviderSetting{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.LiveProviderSetting{}, err
	}
	return s.Store.UpdateProviderSetting(ctx, a, in)
}
