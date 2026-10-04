package application

import (
	"context"
	"hms.local/api/internal/domain"
	"strings"
	"time"
)

type PharmacyRepository interface {
	Medicines(context.Context, string, int) ([]domain.Medicine, error)
	CreateMedicine(context.Context, domain.Actor, domain.MedicineInput) (domain.Medicine, error)
	MedicineBatches(context.Context, string, int) ([]domain.MedicineBatch, error)
	ReceiveMedicine(context.Context, domain.Actor, domain.BatchInput, string) (domain.MedicineBatch, error)
	MedicationOrders(context.Context, domain.Actor, string, int) ([]domain.MedicationOrder, error)
	SignMedication(context.Context, domain.Actor, domain.MedicationInput, string) (domain.MedicationOrder, error)
	CancelMedication(context.Context, domain.Actor, string, string) error
	MoveStock(context.Context, domain.Actor, domain.StockInput, string, string) (domain.StockMovement, error)
	StockMovements(context.Context, domain.Actor, string, int) ([]domain.StockMovement, error)
}
type Pharmacy struct {
	Store PharmacyRepository
	Now   func() time.Time
}

func pageOK(page int) bool  { return page >= 1 && page <= 1000 }
func keyOK(key string) bool { return len(key) >= 16 && len(key) <= 80 }
func (p Pharmacy) Medicines(ctx context.Context, a domain.Actor, search string, page int) ([]domain.Medicine, error) {
	if !a.Can("pharmacy.catalog") {
		return nil, domain.ErrForbidden
	}
	if len(search) > 100 || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return p.Store.Medicines(ctx, search, page)
}
func (p Pharmacy) CreateMedicine(ctx context.Context, a domain.Actor, i domain.MedicineInput) (domain.Medicine, error) {
	if !a.Can("pharmacy.manage") {
		return domain.Medicine{}, domain.ErrForbidden
	}
	if e := i.Validate(); e != nil {
		return domain.Medicine{}, e
	}
	return p.Store.CreateMedicine(ctx, a, i)
}
func (p Pharmacy) Batches(ctx context.Context, a domain.Actor, id string, page int) ([]domain.MedicineBatch, error) {
	if !a.Can("pharmacy.manage") {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return p.Store.MedicineBatches(ctx, id, page)
}
func (p Pharmacy) Receive(ctx context.Context, a domain.Actor, i domain.BatchInput, key string) (domain.MedicineBatch, error) {
	if !a.Can("pharmacy.manage") {
		return domain.MedicineBatch{}, domain.ErrForbidden
	}
	if !keyOK(key) {
		return domain.MedicineBatch{}, domain.ErrValidation
	}
	if e := i.Validate(p.Now()); e != nil {
		return domain.MedicineBatch{}, e
	}
	return p.Store.ReceiveMedicine(ctx, a, i, key)
}
func (p Pharmacy) Orders(ctx context.Context, a domain.Actor, encounter string, page int) ([]domain.MedicationOrder, error) {
	if !a.Can("medication.read") {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(encounter) || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return p.Store.MedicationOrders(ctx, a, encounter, page)
}
func (p Pharmacy) Sign(ctx context.Context, a domain.Actor, i domain.MedicationInput, key string) (domain.MedicationOrder, error) {
	if a.Role != "doctor" {
		return domain.MedicationOrder{}, domain.ErrForbidden
	}
	if !keyOK(key) {
		return domain.MedicationOrder{}, domain.ErrValidation
	}
	if e := i.Validate(); e != nil {
		return domain.MedicationOrder{}, e
	}
	return p.Store.SignMedication(ctx, a, i, key)
}
func (p Pharmacy) Cancel(ctx context.Context, a domain.Actor, id, reason string) error {
	if a.Role != "doctor" {
		return domain.ErrForbidden
	}
	reason = strings.TrimSpace(reason)
	if !domain.UUIDPattern.MatchString(id) || len([]rune(reason)) < 1 || len([]rune(reason)) > 1000 {
		return domain.ErrValidation
	}
	return p.Store.CancelMedication(ctx, a, id, reason)
}
func (p Pharmacy) Move(ctx context.Context, a domain.Actor, i domain.StockInput, key string) (domain.StockMovement, error) {
	if !a.Can("pharmacy.manage") {
		return domain.StockMovement{}, domain.ErrForbidden
	}
	if !keyOK(key) {
		return domain.StockMovement{}, domain.ErrValidation
	}
	if e := i.Validate(); e != nil {
		return domain.StockMovement{}, e
	}
	return p.Store.MoveStock(ctx, a, i, key, p.Now().In(domain.HospitalLocation).Format("2006-01-02"))
}
func (p Pharmacy) Movements(ctx context.Context, a domain.Actor, batch string, page int) ([]domain.StockMovement, error) {
	if !a.Can("pharmacy.manage") {
		return nil, domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(batch) || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return p.Store.StockMovements(ctx, a, batch, page)
}
