package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
)

func (c Clinical) AssignNurse(ctx context.Context, a domain.Actor, id string, i domain.NurseAssignment) (domain.EncounterNurse, error) {
	if a.Role != "admin" && a.Role != "doctor" {
		return domain.EncounterNurse{}, domain.ErrForbidden
	}
	i.Reason = strings.TrimSpace(i.Reason)
	if !domain.UUIDPattern.MatchString(id) || i.NurseID == "" || len(i.NurseID) > 128 || i.Version < 0 || (i.Version == 0 && !i.Active) || len([]rune(i.Reason)) < 1 || len([]rune(i.Reason)) > 2000 {
		return domain.EncounterNurse{}, domain.ErrValidation
	}
	return c.Store.AssignNurse(ctx, a, id, i)
}
func (c Clinical) Nurses(ctx context.Context, a domain.Actor, id string) ([]domain.EncounterNurse, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return nil, domain.ErrValidation
	}
	return c.Store.Nurses(ctx, a, id)
}
func (c Clinical) NursingEncounters(ctx context.Context, a domain.Actor, page int) ([]domain.Encounter, error) {
	if a.Role != "nurse" {
		return nil, domain.ErrForbidden
	}
	if page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return c.Store.NursingEncounters(ctx, a, page)
}
func (c Clinical) Vitals(ctx context.Context, a domain.Actor, id string, page int) ([]domain.Vitals, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "nurse" {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return c.Store.Vitals(ctx, a, id, page)
}
func (c Clinical) RecordVitals(ctx context.Context, a domain.Actor, id string, i domain.VitalsInput, key string) (domain.Vitals, error) {
	if a.Role != "doctor" && a.Role != "nurse" {
		return domain.Vitals{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || len(key) < 16 || len(key) > 80 {
		return domain.Vitals{}, domain.ErrValidation
	}
	if e := i.Validate(c.Now()); e != nil {
		return domain.Vitals{}, e
	}
	return c.Store.RecordVitals(ctx, a, id, i, key)
}
