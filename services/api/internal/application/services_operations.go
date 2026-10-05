package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
	"time"
)

type ServicesOperationsStore interface {
	CreateChargeCategory(context.Context, domain.Actor, domain.ChargeCategoryInput) (domain.ChargeCategory, error)
	UpdateChargeCategory(context.Context, domain.Actor, string, domain.ChargeCategoryInput) (domain.ChargeCategory, error)
	ChargeCategories(context.Context, domain.Actor, int) ([]domain.ChargeCategory, error)
	ChargeCategory(context.Context, domain.Actor, string) (domain.ChargeCategory, error)

	CreateCharge(context.Context, domain.Actor, domain.HospitalChargeInput) (domain.HospitalCharge, error)
	UpdateCharge(context.Context, domain.Actor, string, domain.HospitalChargeInput) (domain.HospitalCharge, error)
	Charges(context.Context, domain.Actor, int, string) ([]domain.HospitalCharge, int, error)
	Charge(context.Context, domain.Actor, string) (domain.HospitalCharge, error)

	CreateService(context.Context, domain.Actor, domain.HospitalServiceInput) (domain.HospitalService, error)
	UpdateService(context.Context, domain.Actor, string, domain.HospitalServiceInput) (domain.HospitalService, error)
	Services(context.Context, domain.Actor, int, *int) ([]domain.HospitalService, int, error)
	Service(context.Context, domain.Actor, string) (domain.HospitalService, error)

	CreateOperationCategory(context.Context, domain.Actor, domain.OperationCategoryInput) (domain.OperationCategory, error)
	OperationCategories(context.Context, domain.Actor) ([]domain.OperationCategory, error)
	CreateOperation(context.Context, domain.Actor, domain.HospitalOperationInput) (domain.HospitalOperation, error)
	UpdateOperation(context.Context, domain.Actor, string, domain.HospitalOperationInput) (domain.HospitalOperation, error)
	Operations(context.Context, domain.Actor, int, string) ([]domain.HospitalOperation, int, error)
	Operation(context.Context, domain.Actor, string) (domain.HospitalOperation, error)

	CreateCustomField(context.Context, domain.Actor, domain.CustomFieldInput) (domain.CustomField, error)
	DeleteCustomField(context.Context, domain.Actor, string) error
	CustomFields(context.Context, domain.Actor, string) ([]domain.CustomField, error)

	ModuleSettings(context.Context, domain.Actor) ([]domain.HospitalModuleSetting, error)
	UpdateModuleSetting(context.Context, domain.Actor, string, bool) (domain.HospitalModuleSetting, error)
}

type ServicesOperationsService struct {
	Store ServicesOperationsStore
	Now   func() time.Time
}

// Charge Categories
func (s ServicesOperationsService) CreateChargeCategory(ctx context.Context, a domain.Actor, in domain.ChargeCategoryInput) (domain.ChargeCategory, error) {
	if !a.Can("services.manage") {
		return domain.ChargeCategory{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.ChargeCategory{}, err
	}
	return s.Store.CreateChargeCategory(ctx, a, in)
}

func (s ServicesOperationsService) UpdateChargeCategory(ctx context.Context, a domain.Actor, id string, in domain.ChargeCategoryInput) (domain.ChargeCategory, error) {
	if !a.Can("services.manage") {
		return domain.ChargeCategory{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.ChargeCategory{}, domain.ErrValidation
	}
	if err := in.Validate(); err != nil {
		return domain.ChargeCategory{}, err
	}
	return s.Store.UpdateChargeCategory(ctx, a, id, in)
}

func (s ServicesOperationsService) ChargeCategories(ctx context.Context, a domain.Actor, chargeType int) ([]domain.ChargeCategory, error) {
	if !a.Can("services.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.ChargeCategories(ctx, a, chargeType)
}

func (s ServicesOperationsService) ChargeCategory(ctx context.Context, a domain.Actor, id string) (domain.ChargeCategory, error) {
	if !a.Can("services.read") {
		return domain.ChargeCategory{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.ChargeCategory{}, domain.ErrValidation
	}
	return s.Store.ChargeCategory(ctx, a, id)
}

// Charges
func (s ServicesOperationsService) CreateCharge(ctx context.Context, a domain.Actor, in domain.HospitalChargeInput) (domain.HospitalCharge, error) {
	if !a.Can("services.manage") {
		return domain.HospitalCharge{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalCharge{}, err
	}
	return s.Store.CreateCharge(ctx, a, in)
}

func (s ServicesOperationsService) UpdateCharge(ctx context.Context, a domain.Actor, id string, in domain.HospitalChargeInput) (domain.HospitalCharge, error) {
	if !a.Can("services.manage") {
		return domain.HospitalCharge{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.HospitalCharge{}, domain.ErrValidation
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalCharge{}, err
	}
	return s.Store.UpdateCharge(ctx, a, id, in)
}

func (s ServicesOperationsService) Charges(ctx context.Context, a domain.Actor, page int, categoryID string) ([]domain.HospitalCharge, int, error) {
	if !a.Can("services.read") {
		return nil, 0, domain.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	return s.Store.Charges(ctx, a, page, categoryID)
}

func (s ServicesOperationsService) Charge(ctx context.Context, a domain.Actor, id string) (domain.HospitalCharge, error) {
	if !a.Can("services.read") {
		return domain.HospitalCharge{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.HospitalCharge{}, domain.ErrValidation
	}
	return s.Store.Charge(ctx, a, id)
}

// Services
func (s ServicesOperationsService) CreateService(ctx context.Context, a domain.Actor, in domain.HospitalServiceInput) (domain.HospitalService, error) {
	if !a.Can("services.manage") {
		return domain.HospitalService{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalService{}, err
	}
	return s.Store.CreateService(ctx, a, in)
}

func (s ServicesOperationsService) UpdateService(ctx context.Context, a domain.Actor, id string, in domain.HospitalServiceInput) (domain.HospitalService, error) {
	if !a.Can("services.manage") {
		return domain.HospitalService{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.HospitalService{}, domain.ErrValidation
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalService{}, err
	}
	return s.Store.UpdateService(ctx, a, id, in)
}

func (s ServicesOperationsService) Services(ctx context.Context, a domain.Actor, page int, status *int) ([]domain.HospitalService, int, error) {
	if !a.Can("services.read") {
		return nil, 0, domain.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	return s.Store.Services(ctx, a, page, status)
}

func (s ServicesOperationsService) Service(ctx context.Context, a domain.Actor, id string) (domain.HospitalService, error) {
	if !a.Can("services.read") {
		return domain.HospitalService{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.HospitalService{}, domain.ErrValidation
	}
	return s.Store.Service(ctx, a, id)
}

// Operation Categories & Operations
func (s ServicesOperationsService) CreateOperationCategory(ctx context.Context, a domain.Actor, in domain.OperationCategoryInput) (domain.OperationCategory, error) {
	if !a.Can("operations.manage") {
		return domain.OperationCategory{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.OperationCategory{}, err
	}
	return s.Store.CreateOperationCategory(ctx, a, in)
}

func (s ServicesOperationsService) OperationCategories(ctx context.Context, a domain.Actor) ([]domain.OperationCategory, error) {
	if !a.Can("operations.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.OperationCategories(ctx, a)
}

func (s ServicesOperationsService) CreateOperation(ctx context.Context, a domain.Actor, in domain.HospitalOperationInput) (domain.HospitalOperation, error) {
	if !a.Can("operations.manage") {
		return domain.HospitalOperation{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalOperation{}, err
	}
	return s.Store.CreateOperation(ctx, a, in)
}

func (s ServicesOperationsService) UpdateOperation(ctx context.Context, a domain.Actor, id string, in domain.HospitalOperationInput) (domain.HospitalOperation, error) {
	if !a.Can("operations.manage") {
		return domain.HospitalOperation{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.HospitalOperation{}, domain.ErrValidation
	}
	if err := in.Validate(); err != nil {
		return domain.HospitalOperation{}, err
	}
	return s.Store.UpdateOperation(ctx, a, id, in)
}

func (s ServicesOperationsService) Operations(ctx context.Context, a domain.Actor, page int, categoryID string) ([]domain.HospitalOperation, int, error) {
	if !a.Can("operations.read") {
		return nil, 0, domain.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	return s.Store.Operations(ctx, a, page, categoryID)
}

func (s ServicesOperationsService) Operation(ctx context.Context, a domain.Actor, id string) (domain.HospitalOperation, error) {
	if !a.Can("operations.read") {
		return domain.HospitalOperation{}, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.HospitalOperation{}, domain.ErrValidation
	}
	return s.Store.Operation(ctx, a, id)
}

// Custom Fields
func (s ServicesOperationsService) CreateCustomField(ctx context.Context, a domain.Actor, in domain.CustomFieldInput) (domain.CustomField, error) {
	if !a.Can("settings.manage") {
		return domain.CustomField{}, domain.ErrForbidden
	}
	if err := in.Validate(); err != nil {
		return domain.CustomField{}, err
	}
	return s.Store.CreateCustomField(ctx, a, in)
}

func (s ServicesOperationsService) DeleteCustomField(ctx context.Context, a domain.Actor, id string) error {
	if !a.Can("settings.manage") {
		return domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.ErrValidation
	}
	return s.Store.DeleteCustomField(ctx, a, id)
}

func (s ServicesOperationsService) CustomFields(ctx context.Context, a domain.Actor, moduleName string) ([]domain.CustomField, error) {
	if !a.Can("settings.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.CustomFields(ctx, a, strings.ToLower(strings.TrimSpace(moduleName)))
}

// Module Settings
func (s ServicesOperationsService) ModuleSettings(ctx context.Context, a domain.Actor) ([]domain.HospitalModuleSetting, error) {
	if !a.Can("settings.read") {
		return nil, domain.ErrForbidden
	}
	return s.Store.ModuleSettings(ctx, a)
}

func (s ServicesOperationsService) UpdateModuleSetting(ctx context.Context, a domain.Actor, key string, isActive bool) (domain.HospitalModuleSetting, error) {
	if !a.Can("settings.manage") {
		return domain.HospitalModuleSetting{}, domain.ErrForbidden
	}
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" || len(key) > 128 {
		return domain.HospitalModuleSetting{}, domain.ErrValidation
	}
	return s.Store.UpdateModuleSetting(ctx, a, key, isActive)
}
