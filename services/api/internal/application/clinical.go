package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
	"time"
)

type ClinicalRepository interface {
	Beds(context.Context, int) ([]domain.Bed, error)
	CreateBed(context.Context, domain.Actor, domain.BedInput) (domain.Bed, error)
	Cases(context.Context, domain.Actor, int) ([]domain.Case, error)
	CreateCase(context.Context, domain.Actor, domain.CaseInput) (domain.Case, error)
	Encounters(context.Context, domain.Actor, string, int) ([]domain.Encounter, error)
	Admit(context.Context, domain.Actor, domain.EncounterInput, string) (domain.Encounter, error)
	Discharge(context.Context, domain.Actor, string, domain.Discharge, time.Time) (domain.Encounter, error)
	Notes(context.Context, domain.Actor, string) ([]domain.ClinicalNote, error)
	SignNote(context.Context, domain.Actor, string, domain.NoteInput, string) (domain.ClinicalNote, error)
}
type Clinical struct {
	Store ClinicalRepository
	Now   func() time.Time
}

func (c Clinical) Beds(ctx context.Context, a domain.Actor, page int) ([]domain.Bed, error) {
	if !a.Can("beds.read") {
		return nil, domain.ErrForbidden
	}
	if page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return c.Store.Beds(ctx, page)
}
func (c Clinical) CreateBed(ctx context.Context, a domain.Actor, b domain.BedInput) (domain.Bed, error) {
	if a.Role != "admin" {
		return domain.Bed{}, domain.ErrForbidden
	}
	if e := b.Validate(); e != nil {
		return domain.Bed{}, e
	}
	return c.Store.CreateBed(ctx, a, b)
}
func (c Clinical) Cases(ctx context.Context, a domain.Actor, page int) ([]domain.Case, error) {
	if !a.Can("clinical.read") {
		return nil, domain.ErrForbidden
	}
	if page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return c.Store.Cases(ctx, a, page)
}
func (c Clinical) CreateCase(ctx context.Context, a domain.Actor, i domain.CaseInput) (domain.Case, error) {
	if !a.Can("clinical.admit") {
		return domain.Case{}, domain.ErrForbidden
	}
	if e := i.Validate(); e != nil {
		return domain.Case{}, e
	}
	if a.Role == "doctor" && a.ID != i.DoctorID {
		return domain.Case{}, domain.ErrForbidden
	}
	return c.Store.CreateCase(ctx, a, i)
}
func (c Clinical) Encounters(ctx context.Context, a domain.Actor, kind string, page int) ([]domain.Encounter, error) {
	if !a.Can("clinical.read") {
		return nil, domain.ErrForbidden
	}
	if (kind != "ipd" && kind != "opd") || page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return c.Store.Encounters(ctx, a, kind, page)
}
func (c Clinical) Admit(ctx context.Context, a domain.Actor, i domain.EncounterInput, key string) (domain.Encounter, error) {
	if !a.Can("clinical.admit") {
		return domain.Encounter{}, domain.ErrForbidden
	}
	if len(key) < 16 || len(key) > 80 {
		return domain.Encounter{}, domain.ErrValidation
	}
	if e := i.Validate(c.Now()); e != nil {
		return domain.Encounter{}, e
	}
	return c.Store.Admit(ctx, a, i, key)
}
func (c Clinical) Discharge(ctx context.Context, a domain.Actor, id string, d domain.Discharge) (domain.Encounter, error) {
	if a.Role != "doctor" {
		return domain.Encounter{}, domain.ErrForbidden
	}
	d.Summary = strings.TrimSpace(d.Summary)
	if !domain.UUIDPattern.MatchString(id) || d.Version < 1 || len([]rune(d.Summary)) < 1 || len([]rune(d.Summary)) > 10000 {
		return domain.Encounter{}, domain.ErrValidation
	}
	return c.Store.Discharge(ctx, a, id, d, c.Now())
}
func (c Clinical) Notes(ctx context.Context, a domain.Actor, id string) ([]domain.ClinicalNote, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "patient" {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return nil, domain.ErrValidation
	}
	return c.Store.Notes(ctx, a, id)
}
func (c Clinical) SignNote(ctx context.Context, a domain.Actor, id string, n domain.NoteInput, key string) (domain.ClinicalNote, error) {
	if a.Role != "doctor" {
		return domain.ClinicalNote{}, domain.ErrForbidden
	}
	n.Body = strings.TrimSpace(n.Body)
	if !domain.UUIDPattern.MatchString(id) || len([]rune(n.Body)) < 1 || len([]rune(n.Body)) > 10000 || len(key) < 16 || len(key) > 80 {
		return domain.ClinicalNote{}, domain.ErrValidation
	}
	return c.Store.SignNote(ctx, a, id, n, key)
}
