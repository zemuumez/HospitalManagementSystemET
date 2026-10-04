package application

import (
	"context"
	"hms.local/api/internal/domain"
	"time"
)

type StaffRepository interface {
	StaffProfile(context.Context, domain.Actor, string) (domain.StaffProfile, error)
	UpdateStaffProfile(context.Context, domain.Actor, string, domain.StaffProfileInput) (domain.StaffProfile, error)
	ChangeStaffRole(context.Context, domain.Actor, string, domain.StaffRoleInput) error
}
type Staff struct {
	Store StaffRepository
	Now   func() time.Time
}

func (s Staff) Profile(ctx context.Context, a domain.Actor, id string) (domain.StaffProfile, error) {
	if a.Role != "admin" {
		return domain.StaffProfile{}, domain.ErrForbidden
	}
	if len(id) < 1 || len(id) > 128 {
		return domain.StaffProfile{}, domain.ErrValidation
	}
	return s.Store.StaffProfile(ctx, a, id)
}
func (s Staff) Update(ctx context.Context, a domain.Actor, id string, i domain.StaffProfileInput) (domain.StaffProfile, error) {
	if a.Role != "admin" {
		return domain.StaffProfile{}, domain.ErrForbidden
	}
	if len(id) < 1 || len(id) > 128 {
		return domain.StaffProfile{}, domain.ErrValidation
	}
	if e := i.Validate(s.Now()); e != nil {
		return domain.StaffProfile{}, e
	}
	return s.Store.UpdateStaffProfile(ctx, a, id, i)
}
func (s Staff) ChangeRole(ctx context.Context, a domain.Actor, id string, i domain.StaffRoleInput) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	if len(id) < 1 || len(id) > 128 {
		return domain.ErrValidation
	}
	if id == a.ID {
		return domain.ErrStale
	}
	if e := i.Validate(); e != nil {
		return e
	}
	return s.Store.ChangeStaffRole(ctx, a, id, i)
}
