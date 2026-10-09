package application

import (
	"context"
	"hms.local/api/internal/domain"
	"log"
	"sort"
	"strings"
	"time"
)

type CMSSettingsStore interface {
	GeneralSettings(context.Context) (map[string]string, error)
	UpdateGeneralSetting(context.Context, domain.Actor, domain.GeneralSettingInput) (domain.HospitalGeneralSetting, error)
	UpdateGeneralSettings(context.Context, domain.Actor, []domain.GeneralSettingInput) ([]string, error)

	HospitalSchedules(context.Context) ([]domain.HospitalScheduleDay, error)
	UpdateHospitalSchedule(context.Context, domain.Actor, domain.HospitalScheduleDayInput) (domain.HospitalScheduleDay, error)
	UpdateHospitalSchedules(context.Context, domain.Actor, []domain.HospitalScheduleDayInput) error

	FrontCMSSettings(context.Context, string) ([]domain.FrontCMSSetting, error)
	UpdateFrontCMSSetting(context.Context, domain.Actor, domain.FrontCMSSettingInput) (domain.FrontCMSSetting, error)
	UpdateFrontCMSSettings(context.Context, domain.Actor, []domain.FrontCMSSettingInput) ([]string, error)

	CleanupAbandonedAttachments(context.Context, domain.Actor, time.Duration) ([]string, error)
	RetireUnreferencedAttachments(context.Context, domain.Actor) ([]string, error)

	Testimonials(context.Context, *int) ([]domain.CMSTestimonial, error)
	Testimonial(context.Context, string) (domain.CMSTestimonial, error)
	CreateTestimonial(context.Context, domain.Actor, domain.CMSTestimonialInput) (domain.CMSTestimonial, error)
	UpdateTestimonial(context.Context, domain.Actor, string, domain.CMSTestimonialInput) (domain.CMSTestimonial, error)
	DeleteTestimonial(context.Context, domain.Actor, string) error
}

type CMSSettingsService struct {
	Store CMSSettingsStore
	Files AttachmentFiles
	Now   func() time.Time
}

// --- General Settings ---

func (s CMSSettingsService) GeneralSettings(ctx context.Context, a domain.Actor) (map[string]string, error) {
	if a.ID != "" && !a.Can("settings.read") {
		return nil, domain.ErrForbidden
	}
	raw, err := s.Store.GeneralSettings(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		if domain.SettingsSecretKeys[k] {
			if strings.TrimSpace(v) != "" {
				out[k] = domain.SecretConfiguredPlaceholder
			} else {
				out[k] = ""
			}
		} else {
			out[k] = v
		}
	}
	return out, nil
}

func (s CMSSettingsService) UpdateGeneralSetting(ctx context.Context, a domain.Actor, in domain.GeneralSettingInput) (domain.HospitalGeneralSetting, error) {
	if !a.Can("settings.manage") {
		return domain.HospitalGeneralSetting{}, domain.ErrForbidden
	}
	if domain.SettingsSecretKeys[in.Key] {
		trimmed := strings.TrimSpace(in.Value)
		// S5: If secret was unchanged by user, do not mutate/overwrite with stale read
		if trimmed == domain.SecretConfiguredPlaceholder || trimmed == "********" {
			existing, err := s.Store.GeneralSettings(ctx)
			if err != nil {
				return domain.HospitalGeneralSetting{}, err
			}
			return domain.HospitalGeneralSetting{Key: in.Key, Value: existing[in.Key]}, nil
		}
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
		if domain.SettingsSecretKeys[in.Key] {
			trimmed := strings.TrimSpace(in.Value)
			// S5: Skip unchanged-secret fields completely from write transaction to eliminate concurrent overwrite races
			if trimmed == domain.SecretConfiguredPlaceholder || trimmed == "********" {
				continue
			}
		}
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
	displacedPaths, err := s.Store.UpdateGeneralSettings(ctx, a, inputs)
	if err != nil {
		return err
	}
	if s.Files != nil {
		for _, p := range displacedPaths {
			if err := s.Files.Remove(ctx, p); err != nil {
				log.Printf("failed to remove displaced attachment file %s: %v", p, err)
			}
		}
	}
	return nil
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

func (s CMSSettingsService) UpdateHospitalSchedules(ctx context.Context, a domain.Actor, days []domain.HospitalScheduleDayInput) error {
	if !a.Can("settings.manage") {
		return domain.ErrForbidden
	}
	// S2: Validate entire batch up-front before touching repository
	seenDays := make(map[int]bool)
	for _, in := range days {
		if err := in.Validate(); err != nil {
			return err
		}
		if seenDays[in.DayOfWeek] {
			return domain.ErrValidation
		}
		seenDays[in.DayOfWeek] = true
	}
	// Aggregate repository-level transaction covering all writes and audit events
	return s.Store.UpdateHospitalSchedules(ctx, a, days)
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
	// S2: Validate entire batch up-front before touching repository
	seenKeys := make(map[string]bool)
	for _, in := range settings {
		if err := in.Validate(); err != nil {
			return err
		}
		if seenKeys[in.Key] {
			return domain.ErrValidation
		}
		seenKeys[in.Key] = true
	}
	// Aggregate repository-level transaction covering all writes and audit events
	displacedPaths, err := s.Store.UpdateFrontCMSSettings(ctx, a, settings)
	if err != nil {
		return err
	}
	if s.Files != nil {
		for _, p := range displacedPaths {
			if err := s.Files.Remove(ctx, p); err != nil {
				log.Printf("failed to remove displaced CMS attachment file %s: %v", p, err)
			}
		}
	}
	return nil
}

func (s CMSSettingsService) CleanupAbandonedAttachments(ctx context.Context, a domain.Actor, olderThan time.Duration) error {
	if !a.Can("settings.manage") {
		return domain.ErrForbidden
	}
	paths, err := s.Store.CleanupAbandonedAttachments(ctx, a, olderThan)
	if err != nil {
		return err
	}
	if s.Files != nil {
		for _, p := range paths {
			if err := s.Files.Remove(ctx, p); err != nil {
				log.Printf("failed to remove abandoned attachment file %s: %v", p, err)
			}
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
