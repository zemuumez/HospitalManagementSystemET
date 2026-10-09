package application

import (
	"context"
	"hms.local/api/internal/domain"
)

type InventoryRepository interface {
	InventoryCategories(context.Context, int) ([]domain.InventoryCategory, error)
	SaveInventoryCategory(context.Context, domain.Actor, string, domain.InventoryCategoryInput) (domain.InventoryCategory, error)
	DeleteInventoryCategory(context.Context, domain.Actor, string) error
	InventoryItems(context.Context, string, bool, int) ([]domain.InventoryItem, error)
	SaveInventoryItem(context.Context, domain.Actor, string, domain.InventoryItemInput) (domain.InventoryItem, error)
	DeleteInventoryItem(context.Context, domain.Actor, string) error
	MoveInventory(context.Context, domain.Actor, domain.InventoryMovementInput, string) (domain.InventoryMovement, error)
	InventoryMovements(context.Context, domain.Actor, string, int) ([]domain.InventoryMovement, error)
}
type Inventory struct{ Store InventoryRepository }

func (i Inventory) Categories(ctx context.Context, a domain.Actor, page int) ([]domain.InventoryCategory, error) {
	if a.Role != "admin" {
		return nil, domain.ErrForbidden
	}
	if !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return i.Store.InventoryCategories(ctx, page)
}
func (i Inventory) SaveCategory(ctx context.Context, a domain.Actor, id string, in domain.InventoryCategoryInput) (domain.InventoryCategory, error) {
	if a.Role != "admin" {
		return domain.InventoryCategory{}, domain.ErrForbidden
	}
	if (id != "" && !domain.UUIDPattern.MatchString(id)) || (id != "" && in.Version < 1) {
		return domain.InventoryCategory{}, domain.ErrValidation
	}
	if e := in.Validate(); e != nil {
		return domain.InventoryCategory{}, e
	}
	return i.Store.SaveInventoryCategory(ctx, a, id, in)
}
func (i Inventory) DeleteCategory(ctx context.Context, a domain.Actor, id string) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.ErrValidation
	}
	return i.Store.DeleteInventoryCategory(ctx, a, id)
}
func (i Inventory) Items(ctx context.Context, a domain.Actor, search string, low bool, page int) ([]domain.InventoryItem, error) {
	if a.Role != "admin" {
		return nil, domain.ErrForbidden
	}
	if len(search) > 100 || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return i.Store.InventoryItems(ctx, search, low, page)
}
func (i Inventory) SaveItem(ctx context.Context, a domain.Actor, id string, in domain.InventoryItemInput) (domain.InventoryItem, error) {
	if a.Role != "admin" {
		return domain.InventoryItem{}, domain.ErrForbidden
	}
	if (id != "" && !domain.UUIDPattern.MatchString(id)) || (id != "" && in.Version < 1) {
		return domain.InventoryItem{}, domain.ErrValidation
	}
	if e := in.Validate(); e != nil {
		return domain.InventoryItem{}, e
	}
	return i.Store.SaveInventoryItem(ctx, a, id, in)
}
func (i Inventory) DeleteItem(ctx context.Context, a domain.Actor, id string) error {
	if a.Role != "admin" {
		return domain.ErrForbidden
	}
	if !domain.UUIDPattern.MatchString(id) {
		return domain.ErrValidation
	}
	return i.Store.DeleteInventoryItem(ctx, a, id)
}
func (i Inventory) Move(ctx context.Context, a domain.Actor, in domain.InventoryMovementInput, key string) (domain.InventoryMovement, error) {
	if a.Role != "admin" {
		return domain.InventoryMovement{}, domain.ErrForbidden
	}
	if !keyOK(key) {
		return domain.InventoryMovement{}, domain.ErrValidation
	}
	if e := in.Validate(); e != nil {
		return domain.InventoryMovement{}, e
	}
	return i.Store.MoveInventory(ctx, a, in, key)
}
func (i Inventory) Movements(ctx context.Context, a domain.Actor, id string, page int) ([]domain.InventoryMovement, error) {
	if a.Role != "admin" {
		return nil, domain.ErrForbidden
	}
	if (id != "" && !domain.UUIDPattern.MatchString(id)) || !pageOK(page) {
		return nil, domain.ErrValidation
	}
	return i.Store.InventoryMovements(ctx, a, id, page)
}
