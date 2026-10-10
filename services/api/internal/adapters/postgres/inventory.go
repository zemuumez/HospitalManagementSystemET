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
func (s Store) InventoryCategories(ctx context.Context, page int, limit int, search string) ([]domain.InventoryCategory, int, error) {
	if limit <= 0 || limit > 500 {
		limit = 25
	}
	if page <= 0 {
		page = 1
	}
	var total int
	e := s.DB.QueryRow(ctx, `SELECT count(*) FROM inventory_category WHERE ($1='' OR strpos(lower(name),lower($1))>0)`, search).Scan(&total)
	if e != nil {
		return nil, 0, e
	}
	offset := (page - 1) * limit
	rows, e := s.DB.Query(ctx, `SELECT id,name,description,active,version FROM inventory_category WHERE ($1='' OR strpos(lower(name),lower($1))>0) ORDER BY name,id LIMIT $2 OFFSET $3`, search, limit, offset)
	if e != nil {
		return nil, 0, e
	}
	defer rows.Close()
	out := []domain.InventoryCategory{}
	for rows.Next() {
		c, e := scanInventoryCategory(rows)
		if e != nil {
			return nil, 0, e
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
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
func (s Store) DeleteInventoryCategory(ctx context.Context, a domain.Actor, id string) error {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var inUse bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM inventory_item WHERE category_id=$1)`, id).Scan(&inUse)
	if e != nil {
		return clinicalError(e)
	}
	if inUse {
		return domain.ErrInUse
	}
	res, e := tx.Exec(ctx, `DELETE FROM inventory_category WHERE id=$1`, id)
	if e != nil {
		return clinicalError(e)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if e = pharmacyAudit(ctx, tx, a, "inventory_category.deleted", id); e != nil {
		return e
	}
	return tx.Commit(ctx)
}

const inventoryItemFields = `id,category_id,name,unit,description,reorder_milli,active,version,balance_milli`

func scanInventoryItem(row pgx.Row) (domain.InventoryItem, error) {
	var i domain.InventoryItem
	e := row.Scan(&i.ID, &i.CategoryID, &i.Name, &i.Unit, &i.Description, &i.ReorderMilli, &i.Active, &i.Version, &i.BalanceMilli)
	return i, clinicalError(e)
}
func (s Store) InventoryItems(ctx context.Context, search string, categoryID string, low bool, page int, limit int) ([]domain.InventoryItem, int, error) {
	if limit <= 0 || limit > 500 {
		limit = 25
	}
	if page <= 0 {
		page = 1
	}
	var total int
	e := s.DB.QueryRow(ctx, `SELECT count(*) FROM inventory_item WHERE ($1='' OR strpos(lower(name),lower($1))>0) AND ($2='' OR category_id=NULLIF($2,'')::uuid) AND (NOT $3 OR (active AND balance_milli<=reorder_milli))`, search, categoryID, low).Scan(&total)
	if e != nil {
		return nil, 0, e
	}
	offset := (page - 1) * limit
	rows, e := s.DB.Query(ctx, `SELECT `+inventoryItemFields+` FROM inventory_item WHERE ($1='' OR strpos(lower(name),lower($1))>0) AND ($2='' OR category_id=NULLIF($2,'')::uuid) AND (NOT $3 OR (active AND balance_milli<=reorder_milli)) ORDER BY name,id LIMIT $4 OFFSET $5`, search, categoryID, low, limit, offset)
	if e != nil {
		return nil, 0, e
	}
	defer rows.Close()
	out := []domain.InventoryItem{}
	for rows.Next() {
		i, e := scanInventoryItem(rows)
		if e != nil {
			return nil, 0, e
		}
		out = append(out, i)
	}
	return out, total, rows.Err()
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
func (s Store) DeleteInventoryItem(ctx context.Context, a domain.Actor, id string) error {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var inUse bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM inventory_movement WHERE item_id=$1)`, id).Scan(&inUse)
	if e != nil {
		return clinicalError(e)
	}
	if inUse {
		return domain.ErrInUse
	}
	res, e := tx.Exec(ctx, `DELETE FROM inventory_item WHERE id=$1`, id)
	if e != nil {
		return clinicalError(e)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if e = pharmacyAudit(ctx, tx, a, "inventory_item.deleted", id); e != nil {
		return e
	}
	return tx.Commit(ctx)
}

const inventoryMovementFields = `m.id,m.item_id,m.kind,m.quantity_milli,COALESCE(m.recipient_id,''),COALESCE(m.original_id::text,''),m.supplier,m.store_name,m.reference,m.cost_minor,m.restock,m.reason,m.delta_milli,COALESCE((SELECT sum(r.quantity_milli) FROM inventory_movement r WHERE r.original_id=m.id AND r.kind='return'), 0)::bigint,m.attachment_url,m.issued_date,m.return_due_date,m.issued_by,m.department,COALESCE(it.name,''),m.created_at`

func scanInventoryMovement(row pgx.Row) (domain.InventoryMovement, error) {
	var m domain.InventoryMovement
	e := row.Scan(&m.ID, &m.ItemID, &m.Kind, &m.QuantityMilli, &m.RecipientID, &m.OriginalID, &m.Supplier, &m.StoreName, &m.Reference, &m.CostMinor, &m.Restock, &m.Reason, &m.DeltaMilli, &m.ReturnedMilli, &m.AttachmentURL, &m.IssuedDate, &m.ReturnDueDate, &m.IssuedBy, &m.Department, &m.ItemName, &m.CreatedAt)
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
		return scanInventoryMovement(tx.QueryRow(ctx, `SELECT `+inventoryMovementFields+` FROM inventory_movement m JOIN inventory_item it ON it.id=m.item_id WHERE m.id=$1`, old))
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
	var outID string
	e = tx.QueryRow(ctx, `INSERT INTO inventory_movement(item_id,kind,quantity_milli,delta_milli,recipient_id,original_id,supplier,store_name,reference,cost_minor,restock,reason,actor_id,request_key,request_hash,attachment_url,issued_date,return_due_date,issued_by,department) VALUES($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,'')::uuid,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20) RETURNING id`, i.ItemID, i.Kind, i.QuantityMilli, delta, i.RecipientID, i.OriginalID, i.Supplier, i.StoreName, i.Reference, i.CostMinor, i.Restock, i.Reason, a.ID, key, wanted, i.AttachmentURL, i.IssuedDate, i.ReturnDueDate, i.IssuedBy, i.Department).Scan(&outID)
	if e != nil {
		return out, clinicalError(e)
	}
	out, e = scanInventoryMovement(tx.QueryRow(ctx, `SELECT `+inventoryMovementFields+` FROM inventory_movement m JOIN inventory_item it ON it.id=m.item_id WHERE m.id=$1`, outID))
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
func (s Store) InventoryMovements(ctx context.Context, a domain.Actor, id string, kind string, page int, limit int, search string) ([]domain.InventoryMovement, int, error) {
	if a.Role != "admin" {
		return nil, 0, domain.ErrForbidden
	}
	if limit <= 0 || limit > 500 {
		limit = 25
	}
	if page <= 0 {
		page = 1
	}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, 0, e
	}
	defer tx.Rollback(ctx)

	var total int
	e = tx.QueryRow(ctx, `SELECT count(*) FROM inventory_movement m JOIN inventory_item it ON it.id=m.item_id WHERE ($1='' OR m.item_id=NULLIF($1,'')::uuid) AND ($2='' OR m.kind=$2) AND ($3='' OR m.reason ILIKE '%'||$3||'%' OR m.reference ILIKE '%'||$3||'%' OR m.supplier ILIKE '%'||$3||'%' OR it.name ILIKE '%'||$3||'%' OR m.department ILIKE '%'||$3||'%' OR m.issued_by ILIKE '%'||$3||'%')`, id, kind, search).Scan(&total)
	if e != nil {
		return nil, 0, e
	}

	offset := (page - 1) * limit
	rows, e := tx.Query(ctx, `SELECT `+inventoryMovementFields+` FROM inventory_movement m JOIN inventory_item it ON it.id=m.item_id WHERE ($1='' OR m.item_id=NULLIF($1,'')::uuid) AND ($2='' OR m.kind=$2) AND ($3='' OR m.reason ILIKE '%'||$3||'%' OR m.reference ILIKE '%'||$3||'%' OR m.supplier ILIKE '%'||$3||'%' OR it.name ILIKE '%'||$3||'%' OR m.department ILIKE '%'||$3||'%' OR m.issued_by ILIKE '%'||$3||'%') ORDER BY m.created_at DESC, m.id LIMIT $4 OFFSET $5`, id, kind, search, limit, offset)
	if e != nil {
		return nil, 0, e
	}
	defer rows.Close()

	out := []domain.InventoryMovement{}
	for rows.Next() {
		m, e := scanInventoryMovement(rows)
		if e != nil {
			return nil, 0, e
		}
		out = append(out, m)
	}
	if e = rows.Err(); e != nil {
		return nil, 0, e
	}
	if e = pharmacyAudit(ctx, tx, a, "inventory.movements_viewed", id); e != nil {
		return nil, 0, e
	}
	return out, total, tx.Commit(ctx)
}
