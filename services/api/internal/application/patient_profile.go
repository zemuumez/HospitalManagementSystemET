package application

import (
	"context"
	"hms.local/api/internal/domain"
)

func (h Hospital) PatientProfile(ctx context.Context, a domain.Actor, id string) (domain.PatientProfile, error) {
	if !a.Can("patients.read") {
		return domain.PatientProfile{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.PatientProfile{}, domain.ErrValidation
	}
	return h.Store.PatientProfile(ctx, a, id)
}
func (h Hospital) UpdatePatientProfile(ctx context.Context, a domain.Actor, id string, i domain.PatientProfileInput) (domain.PatientProfile, error) {
	if !a.Can("patients.create") {
		return domain.PatientProfile{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.PatientProfile{}, domain.ErrValidation
	}
	if e := i.Validate(h.Now()); e != nil {
		return domain.PatientProfile{}, e
	}
	return h.Store.UpdatePatientProfile(ctx, a, id, i)
}
func (h Hospital) PatientRevisions(ctx context.Context, a domain.Actor, id string, page int) ([]domain.ProfileRevision, error) {
	if a.Role != "admin" {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return h.Store.PatientRevisions(ctx, a, id, page)
}
