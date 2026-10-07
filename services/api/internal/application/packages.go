package application

import (
	"context"
	"hms.local/api/internal/domain"
	"time"
)

type PackagesStore interface {
	CreatePackage(ctx context.Context, a domain.Actor, in domain.PackageInput) (domain.Package, error)
	UpdatePackage(ctx context.Context, a domain.Actor, id string, in domain.PackageInput) (domain.Package, error)
	DeletePackage(ctx context.Context, a domain.Actor, id string) error
	Package(ctx context.Context, a domain.Actor, id string) (domain.Package, error)
	Packages(ctx context.Context, a domain.Actor, page int, limit int, search string) ([]domain.Package, int, error)
}

type PackagesService struct {
	Store PackagesStore
	Now   func() time.Time
}

func (s PackagesService) CreatePackage(ctx context.Context, a domain.Actor, in domain.PackageInput) (domain.Package, error) {
	if !a.Can("packages.manage") {
		return domain.Package{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.Package{}, err
	}
	return s.Store.CreatePackage(ctx, a, in)
}

func (s PackagesService) UpdatePackage(ctx context.Context, a domain.Actor, id string, in domain.PackageInput) (domain.Package, error) {
	if !a.Can("packages.manage") {
		return domain.Package{}, domain.ErrForbidden
	}
	if id == "" {
		return domain.Package{}, domain.ErrValidation
	}
	if err := in.Validate(); err != nil {
		return domain.Package{}, err
	}
	return s.Store.UpdatePackage(ctx, a, id, in)
}

func (s PackagesService) DeletePackage(ctx context.Context, a domain.Actor, id string) error {
	if !a.Can("packages.manage") {
		return domain.ErrForbidden
	}
	if id == "" {
		return domain.ErrValidation
	}
	return s.Store.DeletePackage(ctx, a, id)
}

func (s PackagesService) Package(ctx context.Context, a domain.Actor, id string) (domain.Package, error) {
	if !a.Can("packages.read") {
		return domain.Package{}, domain.ErrForbidden
	}
	if id == "" {
		return domain.Package{}, domain.ErrValidation
	}
	return s.Store.Package(ctx, a, id)
}

func (s PackagesService) Packages(ctx context.Context, a domain.Actor, page int, limit int, search string) ([]domain.Package, int, error) {
	if !a.Can("packages.read") {
		return nil, 0, domain.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}
	return s.Store.Packages(ctx, a, page, limit, search)
}
