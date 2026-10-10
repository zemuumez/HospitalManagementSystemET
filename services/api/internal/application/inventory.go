package application

import (
	"context"
	"hms.local/api/internal/domain"
)

type InventoryRepository interface {
	InventoryCategories(context.Context, int, int, string) ([]domain.InventoryCategory, int, error)
	SaveInventoryCategory(context.Context, domain.Actor, string, domain.InventoryCategoryInput) (domain.InventoryCategory, error)
	DeleteInventoryCategory(context.Context, domain.Actor, string) error
	InventoryItems(context.Context, string, string, bool, int, int) ([]domain.InventoryItem, int, error)
	SaveInventoryItem(context.Context, domain.Actor, string, domain.InventoryItemInput) (domain.InventoryItem, error)
	DeleteInventoryItem(context.Context, domain.Actor, string) error
	MoveInventory(context.Context, domain.Actor, domain.InventoryMovementInput, string) (domain.InventoryMovement, error)
	InventoryMovements(context.Context, domain.Actor, string, string, int, int, string) ([]domain.InventoryMovement, int, error)
}
type Inventory struct{ Store InventoryRepository }

func (i Inventory) Categories(ctx context.Context, a domain.Actor, page int, limit int, search string) ([]domain.InventoryCategory, int, error) {
	if a.Role != "admin" {
		return nil, 0, domain.ErrForbidden
	}
	if !pageOK(page) || len(search) > 100 || (limit != 0 && (limit < 1 || limit > 500)) {
		return nil, 0, domain.ErrValidation
	}
	return i.Store.InventoryCategories(ctx, page, limit, search)
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
func (i Inventory) Items(ctx context.Context, a domain.Actor, search string, categoryID string, low bool, page int, limit int) ([]domain.InventoryItem, int, error) {
	if a.Role != "admin" {
		return nil, 0, domain.ErrForbidden
	}
	if len(search) > 100 || !pageOK(page) || (categoryID != "" && !domain.UUIDPattern.MatchString(categoryID)) || (limit != 0 && (limit < 1 || limit > 500)) {
		return nil, 0, domain.ErrValidation
	}
	return i.Store.InventoryItems(ctx, search, categoryID, low, page, limit)
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
func (i Inventory) Movements(ctx context.Context, a domain.Actor, id string, kind string, page int, limit int, search string) ([]domain.InventoryMovement, int, error) {
	if a.Role != "admin" {
		return nil, 0, domain.ErrForbidden
	}
	if (id != "" && !domain.UUIDPattern.MatchString(id)) || !pageOK(page) || (kind != "" && kind != "receive" && kind != "issue" && kind != "return" && kind != "writeoff") || (limit != 0 && (limit < 1 || limit > 500)) || len(search) > 100 {
		return nil, 0, domain.ErrValidation
	}
	return i.Store.InventoryMovements(ctx, a, id, kind, page, limit, search)
}
