package application

import (
	"context"
	"hms.local/api/internal/domain"
	"time"
)

type AmbulanceStore interface {
	CreateAmbulance(context.Context, domain.Actor, domain.AmbulanceInput) (domain.Ambulance, error)
	UpdateAmbulance(context.Context, domain.Actor, string, domain.AmbulanceInput, int) (domain.Ambulance, error)
	Ambulances(context.Context, domain.Actor, int, *bool) ([]domain.Ambulance, int, error)
	Ambulance(context.Context, domain.Actor, string) (domain.Ambulance, error)

	CreateAmbulanceCall(context.Context, domain.Actor, domain.AmbulanceCallInput, string) (domain.AmbulanceCall, error)
	UpdateAmbulanceCall(context.Context, domain.Actor, string, domain.AmbulanceCallUpdateInput) (domain.AmbulanceCall, error)
	AmbulanceCalls(context.Context, domain.Actor, int, string, string, string) ([]domain.AmbulanceCall, int, error)
	AmbulanceCall(context.Context, domain.Actor, string) (domain.AmbulanceCall, error)
	BillAmbulanceCall(context.Context, domain.Actor, string, domain.SourceInvoiceInput, string, string) (domain.Invoice, error)
}

type AmbulanceService struct {
	Store AmbulanceStore
	Now   func() time.Time
}

func (s AmbulanceService) CreateAmbulance(ctx context.Context, a domain.Actor, in domain.AmbulanceInput) (domain.Ambulance, error) {
	if !a.Can("ambulance.manage") {
		return domain.Ambulance{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.Ambulance{}, err
	}
	return s.Store.CreateAmbulance(ctx, a, in)
}

func (s AmbulanceService) UpdateAmbulance(ctx context.Context, a domain.Actor, id string, in domain.AmbulanceInput, version int) (domain.Ambulance, error) {
	if !a.Can("ambulance.manage") {
		return domain.Ambulance{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || version <= 0 {
		return domain.Ambulance{}, domain.ErrValidation
	}
	if err := in.Validate(); err != nil {
		return domain.Ambulance{}, err
	}
	return s.Store.UpdateAmbulance(ctx, a, id, in, version)
}

func (s AmbulanceService) Ambulances(ctx context.Context, a domain.Actor, page int, availableOnly *bool) ([]domain.Ambulance, int, error) {
	if !a.Can("ambulance.read") {
		return nil, 0, domain.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	return s.Store.Ambulances(ctx, a, page, availableOnly)
}

func (s AmbulanceService) Ambulance(ctx context.Context, a domain.Actor, id string) (domain.Ambulance, error) {
	if !a.Can("ambulance.read") {
		return domain.Ambulance{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.Ambulance{}, domain.ErrValidation
	}
	return s.Store.Ambulance(ctx, a, id)
}

func (s AmbulanceService) CreateCall(ctx context.Context, a domain.Actor, in domain.AmbulanceCallInput, key string) (domain.AmbulanceCall, error) {
	if !a.Can("ambulance_call.manage") {
		return domain.AmbulanceCall{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.AmbulanceCall{}, err
	}
	return s.Store.CreateAmbulanceCall(ctx, a, in, key)
}

func (s AmbulanceService) UpdateCall(ctx context.Context, a domain.Actor, id string, in domain.AmbulanceCallUpdateInput) (domain.AmbulanceCall, error) {
	if !a.Can("ambulance_call.manage") {
		return domain.AmbulanceCall{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.AmbulanceCall{}, domain.ErrValidation
	}
	if err := in.Validate(); err != nil {
		return domain.AmbulanceCall{}, err
	}
	return s.Store.UpdateAmbulanceCall(ctx, a, id, in)
}

func (s AmbulanceService) Calls(ctx context.Context, a domain.Actor, page int, status, patientID, ambulanceID string) ([]domain.AmbulanceCall, int, error) {
	if !a.Can("ambulance_call.read") {
		return nil, 0, domain.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	return s.Store.AmbulanceCalls(ctx, a, page, status, patientID, ambulanceID)
}

func (s AmbulanceService) Call(ctx context.Context, a domain.Actor, id string) (domain.AmbulanceCall, error) {
	if !a.Can("ambulance_call.read") {
		return domain.AmbulanceCall{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.AmbulanceCall{}, domain.ErrValidation
	}
	return s.Store.AmbulanceCall(ctx, a, id)
}

func (s AmbulanceService) BillCall(ctx context.Context, a domain.Actor, callID string, in domain.SourceInvoiceInput, key, date string) (domain.Invoice, error) {
	if !a.Can("billing.manage") {
		return domain.Invoice{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(callID) || !domain.UUIDPattern.MatchString(in.AccountID) {
		return domain.Invoice{}, domain.ErrValidation
	}
	if in.DiscountBasisPoints < 0 || in.DiscountBasisPoints > 9999 {
		return domain.Invoice{}, domain.ErrValidation
	}
	if date == "" && s.Now != nil {
		date = s.Now().Format("2006-01-02")
	}
	return s.Store.BillAmbulanceCall(ctx, a, callID, in, key, date)
}
