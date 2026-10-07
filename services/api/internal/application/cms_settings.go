package application

import (
	"context"
	"hms.local/api/internal/domain"
	"sort"
	"time"
)

type CMSSettingsStore interface {
	GeneralSettings(context.Context) (map[string]string, error)
	UpdateGeneralSetting(context.Context, domain.Actor, domain.GeneralSettingInput) (domain.HospitalGeneralSetting, error)
	UpdateGeneralSettings(context.Context, domain.Actor, []domain.GeneralSettingInput) error

	HospitalSchedules(context.Context) ([]domain.HospitalScheduleDay, error)
	UpdateHospitalSchedule(context.Context, domain.Actor, domain.HospitalScheduleDayInput) (domain.HospitalScheduleDay, error)

	FrontCMSSettings(context.Context, string) ([]domain.FrontCMSSetting, error)
	UpdateFrontCMSSetting(context.Context, domain.Actor, domain.FrontCMSSettingInput) (domain.FrontCMSSetting, error)

	Testimonials(context.Context, *int) ([]domain.CMSTestimonial, error)
	Testimonial(context.Context, string) (domain.CMSTestimonial, error)
	CreateTestimonial(context.Context, domain.Actor, domain.CMSTestimonialInput) (domain.CMSTestimonial, error)
	UpdateTestimonial(context.Context, domain.Actor, string, domain.CMSTestimonialInput) (domain.CMSTestimonial, error)
	DeleteTestimonial(context.Context, domain.Actor, string) error
}

type CMSSettingsService struct {
	Store CMSSettingsStore
	Now   func() time.Time
}

// --- General Settings ---

func (s CMSSettingsService) GeneralSettings(ctx context.Context, a domain.Actor) (map[string]string, error) {
	if a.ID != "" && !a.Can("settings.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.GeneralSettings(ctx)
}

func (s CMSSettingsService) UpdateGeneralSetting(ctx context.Context, a domain.Actor, in domain.GeneralSettingInput) (domain.HospitalGeneralSetting, error) {
	if !a.Can("settings.manage") {
		return domain.HospitalGeneralSetting{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalGeneralSetting{}, err
	}
	return s.Store.UpdateGeneralSetting(ctx, a, in)
}

func (s CMSSettingsService) UpdateGeneralSettings(ctx context.Context, a domain.Actor, settings map[string]string) error {
	if !a.Can("settings.manage") {
		return domain.ErrForbidden
	}
	inputs := make([]domain.GeneralSettingInput, 0, len(settings))
	seen := make(map[string]bool)
	for k, v := range settings {
		in := domain.GeneralSettingInput{Key: k, Value: v}
		if err := in.Validate(); err != nil {
			return err
		}
		if seen[in.Key] {
			return domain.ErrValidation
		}
		seen[in.Key] = true
		inputs = append(inputs, in)
	}
	// Validate the entire form before writing, and lock keys in stable order.
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].Key < inputs[j].Key })
	return s.Store.UpdateGeneralSettings(ctx, a, inputs)
}

// --- Hospital Schedules ---

func (s CMSSettingsService) HospitalSchedules(ctx context.Context, a domain.Actor) ([]domain.HospitalScheduleDay, error) {
	if a.ID != "" && !a.Can("settings.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.HospitalSchedules(ctx)
}

func (s CMSSettingsService) UpdateHospitalSchedule(ctx context.Context, a domain.Actor, in domain.HospitalScheduleDayInput) (domain.HospitalScheduleDay, error) {
	if !a.Can("settings.manage") {
		return domain.HospitalScheduleDay{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalScheduleDay{}, err
	}
	return s.Store.UpdateHospitalSchedule(ctx, a, in)
}

// --- Front CMS Settings ---

func (s CMSSettingsService) FrontCMSSettings(ctx context.Context, a domain.Actor, typeFilter string) ([]domain.FrontCMSSetting, error) {
	if a.ID != "" && !a.Can("cms.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.FrontCMSSettings(ctx, typeFilter)
}

func (s CMSSettingsService) UpdateFrontCMSSetting(ctx context.Context, a domain.Actor, in domain.FrontCMSSettingInput) (domain.FrontCMSSetting, error) {
	if !a.Can("cms.manage") {
		return domain.FrontCMSSetting{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.FrontCMSSetting{}, err
	}
	return s.Store.UpdateFrontCMSSetting(ctx, a, in)
}

func (s CMSSettingsService) UpdateFrontCMSSettings(ctx context.Context, a domain.Actor, settings []domain.FrontCMSSettingInput) error {
	if !a.Can("cms.manage") {
		return domain.ErrForbidden
	}
	for _, in := range settings {
		if err := in.Validate(); err != nil {
			return err
		}
		if _, err := s.Store.UpdateFrontCMSSetting(ctx, a, in); err != nil {
			return err
		}
	}
	return nil
}

// --- Testimonials ---

func (s CMSSettingsService) Testimonials(ctx context.Context, a domain.Actor, statusFilter *int) ([]domain.CMSTestimonial, error) {
	if a.ID != "" && !a.Can("cms.read") {
		return nil, domain.ErrForbidden
	}
	// If unauthenticated or non-admin, only published testimonials (status=1)
	if a.ID == "" || !a.Can("cms.manage") {
		published := 1
		statusFilter = &published
	}
	return s.Store.Testimonials(ctx, statusFilter)
}

func (s CMSSettingsService) Testimonial(ctx context.Context, a domain.Actor, id string) (domain.CMSTestimonial, error) {
	if a.ID != "" && !a.Can("cms.read") {
		return domain.CMSTestimonial{}, domain.ErrForbidden
	}
	return s.Store.Testimonial(ctx, id)
}

func (s CMSSettingsService) CreateTestimonial(ctx context.Context, a domain.Actor, in domain.CMSTestimonialInput) (domain.CMSTestimonial, error) {
	if !a.Can("cms.manage") {
		return domain.CMSTestimonial{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.CMSTestimonial{}, err
	}
	return s.Store.CreateTestimonial(ctx, a, in)
}

func (s CMSSettingsService) UpdateTestimonial(ctx context.Context, a domain.Actor, id string, in domain.CMSTestimonialInput) (domain.CMSTestimonial, error) {
	if !a.Can("cms.manage") {
		return domain.CMSTestimonial{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.CMSTestimonial{}, err
	}
	return s.Store.UpdateTestimonial(ctx, a, id, in)
}

func (s CMSSettingsService) DeleteTestimonial(ctx context.Context, a domain.Actor, id string) error {
	if !a.Can("cms.manage") {
		return domain.ErrForbidden
	}
	return s.Store.DeleteTestimonial(ctx, a, id)
}
