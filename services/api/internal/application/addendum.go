package application

import (
	"context"
	"hms.local/api/internal/domain"
)

func (c Clinical) Addenda(ctx context.Context, a domain.Actor, id string, page int) ([]domain.Addendum, error) {
	if a.Role != "admin" && a.Role != "doctor" && a.Role != "patient" {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || page < 1 || page > 1000 {
		return nil, domain.ErrValidation
	}
	return c.Store.Addenda(ctx, a, id, page)
}
func (c Clinical) AddAddendum(ctx context.Context, a domain.Actor, id string, i domain.AddendumInput, key string) (domain.Addendum, error) {
	if a.Role != "doctor" {
		return domain.Addendum{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || len(key) < 16 || len(key) > 80 {
		return domain.Addendum{}, domain.ErrValidation
	}
	if e := i.Validate(); e != nil {
		return domain.Addendum{}, e
	}
	return c.Store.AddAddendum(ctx, a, id, i, key)
}
