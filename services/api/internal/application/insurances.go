package application

import (
	"context"
	"hms.local/api/internal/domain"
	"time"
)

type InsurancesStore interface {
	CreateInsurance(ctx context.Context, a domain.Actor, in domain.InsuranceInput) (domain.Insurance, error)
	UpdateInsurance(ctx context.Context, a domain.Actor, id string, in domain.InsuranceInput) (domain.Insurance, error)
	DeleteInsurance(ctx context.Context, a domain.Actor, id string) error
	Insurance(ctx context.Context, a domain.Actor, id string) (domain.Insurance, error)
	Insurances(ctx context.Context, a domain.Actor, page int, limit int, search string) ([]domain.Insurance, int, error)
	ToggleInsuranceStatus(ctx context.Context, a domain.Actor, id string) (domain.Insurance, error)
}

type InsurancesService struct {
	Store InsurancesStore
	Now   func() time.Time
}

func (s InsurancesService) CreateInsurance(ctx context.Context, a domain.Actor, in domain.InsuranceInput) (domain.Insurance, error) {
	if !a.Can("insurances.manage") {
		return domain.Insurance{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.Insurance{}, err
	}
	return s.Store.CreateInsurance(ctx, a, in)
}

func (s InsurancesService) UpdateInsurance(ctx context.Context, a domain.Actor, id string, in domain.InsuranceInput) (domain.Insurance, error) {
	if !a.Can("insurances.manage") {
		return domain.Insurance{}, domain.ErrForbidden
	}
	if id == "" {
		return domain.Insurance{}, domain.ErrValidation
	}
	if err := in.Validate(); err != nil {
		return domain.Insurance{}, err
	}
	return s.Store.UpdateInsurance(ctx, a, id, in)
}

func (s InsurancesService) DeleteInsurance(ctx context.Context, a domain.Actor, id string) error {
	if !a.Can("insurances.manage") {
		return domain.ErrForbidden
	}
	if id == "" {
		return domain.ErrValidation
	}
	return s.Store.DeleteInsurance(ctx, a, id)
}

func (s InsurancesService) Insurance(ctx context.Context, a domain.Actor, id string) (domain.Insurance, error) {
	if !a.Can("insurances.read") {
		return domain.Insurance{}, domain.ErrForbidden
	}
	if id == "" {
		return domain.Insurance{}, domain.ErrValidation
	}
	return s.Store.Insurance(ctx, a, id)
}

func (s InsurancesService) Insurances(ctx context.Context, a domain.Actor, page int, limit int, search string) ([]domain.Insurance, int, error) {
	if !a.Can("insurances.read") {
		return nil, 0, domain.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 25
	} else if limit > 100 {
		limit = 100
	}
	return s.Store.Insurances(ctx, a, page, limit, search)
}

func (s InsurancesService) ExportInsurances(ctx context.Context, a domain.Actor, search string) ([]domain.Insurance, int, error) {
	if !a.Can("insurances.read") {
		return nil, 0, domain.ErrForbidden
	}
	const maxExport = 5000
	const batchSize = 100
	var all []domain.Insurance
	page := 1
	var totalCount int
	for {
		batch, total, err := s.Store.Insurances(ctx, a, page, batchSize, search)
		if err != nil {
			return nil, 0, err
		}
		totalCount = total
		all = append(all, batch...)
		if len(all) >= total || len(batch) == 0 || len(all) >= maxExport {
			break
		}
		page++
	}
	return all, totalCount, nil
}

func (s InsurancesService) ToggleStatus(ctx context.Context, a domain.Actor, id string) (domain.Insurance, error) {
	if !a.Can("insurances.manage") {
		return domain.Insurance{}, domain.ErrForbidden
	}
	if id == "" {
		return domain.Insurance{}, domain.ErrValidation
	}
	return s.Store.ToggleInsuranceStatus(ctx, a, id)
}
