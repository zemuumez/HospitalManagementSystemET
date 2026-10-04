package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

func scanInventoryCategory(row pgx.Row) (domain.InventoryCategory, error) {
	var c domain.InventoryCategory
	e := row.Scan(&c.ID, &c.Name, &c.Description, &c.Active, &c.Version)
	return c, clinicalError(e)
}
func (s Store) InventoryCategories(ctx context.Context, page int) ([]domain.InventoryCategory, error) {
	rows, e := s.DB.Query(ctx, `SELECT id,name,description,active,version FROM inventory_category ORDER BY name,id LIMIT 25 OFFSET $1`, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.InventoryCategory{}
	for rows.Next() {
		c, e := scanInventoryCategory(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
func (s Store) SaveInventoryCategory(ctx context.Context, a domain.Actor, id string, i domain.InventoryCategoryInput) (domain.InventoryCategory, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.InventoryCategory{}, e
	}
	defer tx.Rollback(ctx)
	var out domain.InventoryCategory
	if id == "" {
		out, e = scanInventoryCategory(tx.QueryRow(ctx, `INSERT INTO inventory_category(name,description,active) VALUES($1,$2,$3) RETURNING id,name,description,active,version`, i.Name, i.Description, i.Active))
	} else {
		var version int
		e = tx.QueryRow(ctx, `SELECT version FROM inventory_category WHERE id=$1 FOR UPDATE`, id).Scan(&version)
		if e != nil {
			return out, clinicalError(e)
		}
		if version != i.Version {
			return out, domain.ErrStale
		}
		out, e = scanInventoryCategory(tx.QueryRow(ctx, `UPDATE inventory_category SET name=$2,description=$3,active=$4,version=version+1 WHERE id=$1 RETURNING id,name,description,active,version`, id, i.Name, i.Description, i.Active))
	}
	if e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "inventory_category.saved", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}

const inventoryItemFields = `id,category_id,name,unit,description,reorder_milli,active,version,balance_milli`

func scanInventoryItem(row pgx.Row) (domain.InventoryItem, error) {
	var i domain.InventoryItem
	e := row.Scan(&i.ID, &i.CategoryID, &i.Name, &i.Unit, &i.Description, &i.ReorderMilli, &i.Active, &i.Version, &i.BalanceMilli)
	return i, clinicalError(e)
}
func (s Store) InventoryItems(ctx context.Context, search string, low bool, page int) ([]domain.InventoryItem, error) {
	rows, e := s.DB.Query(ctx, `SELECT `+inventoryItemFields+` FROM inventory_item WHERE ($1='' OR strpos(lower(name),lower($1))>0) AND (NOT $2 OR (active AND balance_milli<=reorder_milli)) ORDER BY name,id LIMIT 25 OFFSET $3`, search, low, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.InventoryItem{}
	for rows.Next() {
		i, e := scanInventoryItem(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
func (s Store) SaveInventoryItem(ctx context.Context, a domain.Actor, id string, i domain.InventoryItemInput) (domain.InventoryItem, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.InventoryItem{}, e
	}
	defer tx.Rollback(ctx)
	var out domain.InventoryItem
	// Item locks precede category locks everywhere an existing item is changed.
	if id != "" {
		out, e = scanInventoryItem(tx.QueryRow(ctx, `SELECT `+inventoryItemFields+` FROM inventory_item WHERE id=$1 FOR UPDATE`, id))
		if e != nil {
			return out, e
		}
		if out.Version != i.Version || (!i.Active && out.BalanceMilli != 0) {
			return out, domain.ErrStale
		}
		if out.Unit != i.Unit {
			var used bool
			e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM inventory_movement WHERE item_id=$1)`, id).Scan(&used)
			if e != nil {
				return out, e
			}
			if used {
				return out, domain.ErrStale
			}
		}
	}
	var active bool
	e = tx.QueryRow(ctx, `SELECT active FROM inventory_category WHERE id=$1 FOR SHARE`, i.CategoryID).Scan(&active)
	if e != nil {
		return out, clinicalError(e)
	}
	if i.Active && !active {
		return out, domain.ErrStale
	}
	if id == "" {
		out, e = scanInventoryItem(tx.QueryRow(ctx, `INSERT INTO inventory_item(category_id,name,unit,description,reorder_milli,active) VALUES($1,$2,$3,$4,$5,$6) RETURNING `+inventoryItemFields, i.CategoryID, i.Name, i.Unit, i.Description, i.ReorderMilli, i.Active))
	} else {
		out, e = scanInventoryItem(tx.QueryRow(ctx, `UPDATE inventory_item SET category_id=$2,name=$3,unit=$4,description=$5,reorder_milli=$6,active=$7,version=version+1 WHERE id=$1 RETURNING `+inventoryItemFields, id, i.CategoryID, i.Name, i.Unit, i.Description, i.ReorderMilli, i.Active))
	}
	if e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "inventory_item.saved", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}

const inventoryMovementFields = `id,item_id,kind,quantity_milli,COALESCE(recipient_id,''),COALESCE(original_id::text,''),supplier,store_name,reference,cost_minor,restock,reason,delta_milli,created_at`

func scanInventoryMovement(row pgx.Row) (domain.InventoryMovement, error) {
	var m domain.InventoryMovement
	e := row.Scan(&m.ID, &m.ItemID, &m.Kind, &m.QuantityMilli, &m.RecipientID, &m.OriginalID, &m.Supplier, &m.StoreName, &m.Reference, &m.CostMinor, &m.Restock, &m.Reason, &m.DeltaMilli, &m.CreatedAt)
	return m, clinicalError(e)
}
func (s Store) MoveInventory(ctx context.Context, a domain.Actor, i domain.InventoryMovementInput, key string) (domain.InventoryMovement, error) {
	out := domain.InventoryMovement{}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if e = pharmacyLock(ctx, tx, a, key, "inventory"); e != nil {
		return out, e
	}
	wanted := requestHash(i)
	var old, hash string
	e = tx.QueryRow(ctx, `SELECT id,request_hash FROM inventory_movement WHERE actor_id=$1 AND request_key=$2`, a.ID, key).Scan(&old, &hash)
	if e == nil {
		if hash != wanted {
			return out, domain.ErrConflict
		}
		return scanInventoryMovement(tx.QueryRow(ctx, `SELECT `+inventoryMovementFields+` FROM inventory_movement WHERE id=$1`, old))
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return out, e
	}
	item, e := scanInventoryItem(tx.QueryRow(ctx, `SELECT `+inventoryItemFields+` FROM inventory_item WHERE id=$1 FOR UPDATE`, i.ItemID))
	if e != nil {
		return out, e
	}
	delta := -i.QuantityMilli
	if i.Kind == "receive" || i.Kind == "issue" {
		var active bool
		e = tx.QueryRow(ctx, `SELECT active FROM inventory_category WHERE id=$1 FOR SHARE`, item.CategoryID).Scan(&active)
		if e != nil {
			return out, e
		}
		if !item.Active || !active {
			return out, domain.ErrStale
		}
	}
	switch i.Kind {
	case "receive":
		delta = i.QuantityMilli
	case "issue":
		var active bool
		e = tx.QueryRow(ctx, `SELECT active AND role<>'patient' FROM staff_access WHERE user_id=$1 FOR SHARE`, i.RecipientID).Scan(&active)
		if e != nil {
			return out, clinicalError(e)
		}
		if !active {
			return out, domain.ErrValidation
		}
	case "return":
		var issued, returned int64
		e = tx.QueryRow(ctx, `SELECT quantity_milli,recipient_id,(SELECT COALESCE(sum(quantity_milli),0) FROM inventory_movement WHERE original_id=$1 AND kind='return') FROM inventory_movement WHERE id=$1 AND item_id=$2 AND kind='issue'`, i.OriginalID, i.ItemID).Scan(&issued, &i.RecipientID, &returned)
		if e != nil {
			return out, clinicalError(e)
		}
		if returned+i.QuantityMilli > issued {
			return out, domain.ErrStale
		}
		delta = 0
		if i.Restock {
			delta = i.QuantityMilli
		}
	}
	if item.BalanceMilli+delta < 0 || item.BalanceMilli+delta > 1000000000000 {
		return out, domain.ErrStale
	}
	out, e = scanInventoryMovement(tx.QueryRow(ctx, `INSERT INTO inventory_movement(item_id,kind,quantity_milli,delta_milli,recipient_id,original_id,supplier,store_name,reference,cost_minor,restock,reason,actor_id,request_key,request_hash) VALUES($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,'')::uuid,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING `+inventoryMovementFields, i.ItemID, i.Kind, i.QuantityMilli, delta, i.RecipientID, i.OriginalID, i.Supplier, i.StoreName, i.Reference, i.CostMinor, i.Restock, i.Reason, a.ID, key, wanted))
	if e != nil {
		return out, clinicalError(e)
	}
	if _, e = tx.Exec(ctx, `UPDATE inventory_item SET balance_milli=balance_milli+$2,version=version+1 WHERE id=$1`, i.ItemID, delta); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "inventory."+i.Kind, out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) InventoryMovements(ctx context.Context, a domain.Actor, id string, page int) ([]domain.InventoryMovement, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	rows, e := tx.Query(ctx, `SELECT `+inventoryMovementFields+` FROM inventory_movement WHERE item_id=$1 ORDER BY created_at DESC,id LIMIT 25 OFFSET $2`, id, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.InventoryMovement{}
	for rows.Next() {
		m, e := scanInventoryMovement(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, m)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if e = pharmacyAudit(ctx, tx, a, "inventory.movements_viewed", id); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
