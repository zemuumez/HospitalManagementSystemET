package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

const medicineFields = `id,name,category,brand,unit,composition,side_effects,selling_price_minor`

func scanMedicine(row pgx.Row) (domain.Medicine, error) {
	var m domain.Medicine
	e := row.Scan(&m.ID, &m.Name, &m.Category, &m.Brand, &m.Unit, &m.Composition, &m.SideEffects, &m.SellingPriceMinor)
	return m, clinicalError(e)
}
func (s Store) Medicines(ctx context.Context, search string, page int) ([]domain.Medicine, error) {
	rows, e := s.DB.Query(ctx, `SELECT `+medicineFields+` FROM medicine WHERE active AND ($1='' OR strpos(lower(name),lower($1))>0) ORDER BY name,id LIMIT 25 OFFSET $2`, search, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Medicine{}
	for rows.Next() {
		m, e := scanMedicine(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s Store) CreateMedicine(ctx context.Context, a domain.Actor, i domain.MedicineInput) (domain.Medicine, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.Medicine{}, e
	}
	defer tx.Rollback(ctx)
	m, e := scanMedicine(tx.QueryRow(ctx, `INSERT INTO medicine(name,category,brand,unit,composition,side_effects,selling_price_minor,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING `+medicineFields, i.Name, i.Category, i.Brand, i.Unit, i.Composition, i.SideEffects, i.SellingPriceMinor, a.ID))
	if e != nil {
		return m, e
	}
	if e = pharmacyAudit(ctx, tx, a, "medicine.created", m.ID); e != nil {
		return m, e
	}
	return m, tx.Commit(ctx)
}
func pharmacyAudit(ctx context.Context, tx pgx.Tx, a domain.Actor, action, id string) error {
	_, e := tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,$2,$3)`, a.ID, action, id)
	return e
}

const batchFields = `id,medicine_id,lot,expiry_date::text,supplier,purchase_reference,unit_cost_minor,received_quantity,balance`

func scanBatch(row pgx.Row) (domain.MedicineBatch, error) {
	var b domain.MedicineBatch
	e := row.Scan(&b.ID, &b.MedicineID, &b.Lot, &b.ExpiryDate, &b.Supplier, &b.PurchaseReference, &b.UnitCostMinor, &b.Quantity, &b.Balance)
	return b, clinicalError(e)
}
func (s Store) MedicineBatches(ctx context.Context, id string, page int) ([]domain.MedicineBatch, error) {
	rows, e := s.DB.Query(ctx, `SELECT `+batchFields+` FROM medicine_batch WHERE medicine_id=$1 ORDER BY expiry_date,id LIMIT 25 OFFSET $2`, id, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.MedicineBatch{}
	for rows.Next() {
		b, e := scanBatch(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
func pharmacyLock(ctx context.Context, tx pgx.Tx, a domain.Actor, key, kind string) error {
	_, e := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, kind+":"+a.ID+":"+key)
	return e
}
func (s Store) ReceiveMedicine(ctx context.Context, a domain.Actor, i domain.BatchInput, key string) (domain.MedicineBatch, error) {
	out := domain.MedicineBatch{}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if e = pharmacyLock(ctx, tx, a, key, "receipt"); e != nil {
		return out, e
	}
	var oldID, hash string
	wanted := requestHash(i)
	e = tx.QueryRow(ctx, `SELECT id,request_hash FROM medicine_batch WHERE created_by=$1 AND request_key=$2`, a.ID, key).Scan(&oldID, &hash)
	if e == nil {
		if hash != wanted {
			return out, domain.ErrConflict
		}
		return scanBatch(tx.QueryRow(ctx, `SELECT `+batchFields+` FROM medicine_batch WHERE id=$1`, oldID))
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return out, e
	}
	var active bool
	if e = tx.QueryRow(ctx, `SELECT active FROM medicine WHERE id=$1 FOR SHARE`, i.MedicineID).Scan(&active); e != nil {
		return out, clinicalError(e)
	}
	if !active {
		return out, domain.ErrStale
	}
	out, e = scanBatch(tx.QueryRow(ctx, `INSERT INTO medicine_batch(medicine_id,lot,expiry_date,supplier,purchase_reference,unit_cost_minor,received_quantity,balance,created_by,request_key,request_hash) VALUES($1,$2,$3,$4,$5,$6,$7,$7,$8,$9,$10) RETURNING `+batchFields, i.MedicineID, i.Lot, i.ExpiryDate, i.Supplier, i.PurchaseReference, i.UnitCostMinor, i.Quantity, a.ID, key, wanted))
	if e != nil {
		return out, e
	}
	// Receipt movements use a separate namespace from caller-supplied movement keys.
	if _, e = tx.Exec(ctx, `INSERT INTO pharmacy_movement(batch_id,kind,quantity,actor_id,request_key,request_hash) VALUES($1,'receipt',$2,$3,$4,$5)`, out.ID, i.Quantity, a.ID, "receipt:"+out.ID, wanted); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "pharmacy.received", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}

const orderFields = `o.id,o.encounter_id,o.medicine_id,o.quantity,o.dose,o.route,o.frequency,o.duration_days,o.instructions,o.medicine_name,o.unit,o.signed_by,o.signed_at,EXISTS(SELECT 1 FROM medication_cancellation c WHERE c.order_id=o.id),(SELECT COALESCE(sum(quantity),0)::integer FROM pharmacy_movement m WHERE m.order_id=o.id AND m.kind='dispense')`

func scanOrder(row pgx.Row) (domain.MedicationOrder, error) {
	var o domain.MedicationOrder
	e := row.Scan(&o.ID, &o.EncounterID, &o.MedicineID, &o.Quantity, &o.Dose, &o.Route, &o.Frequency, &o.DurationDays, &o.Instructions, &o.MedicineName, &o.Unit, &o.SignedBy, &o.SignedAt, &o.Cancelled, &o.Dispensed)
	return o, clinicalError(e)
}
func (s Store) MedicationOrders(ctx context.Context, a domain.Actor, id string, page int) ([]domain.MedicationOrder, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	var authorized bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM encounter e JOIN patient p ON p.id=e.patient_id WHERE e.id=$1 AND ($2 IN ('admin','pharmacist') OR ($2='doctor' AND e.doctor_id=$3) OR ($2='patient' AND p.user_id=$3)))`, id, a.Role, a.ID).Scan(&authorized)
	if e != nil {
		return nil, e
	}
	if !authorized {
		return nil, domain.ErrNotFound
	}
	rows, e := tx.Query(ctx, `SELECT `+orderFields+` FROM medication_order o WHERE o.encounter_id=$1 ORDER BY o.signed_at DESC,o.id LIMIT 25 OFFSET $2`, id, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.MedicationOrder{}
	for rows.Next() {
		o, e := scanOrder(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out = append(out, o)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if e = pharmacyAudit(ctx, tx, a, "medication.viewed", id); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) SignMedication(ctx context.Context, a domain.Actor, i domain.MedicationInput, key string) (domain.MedicationOrder, error) {
	out := domain.MedicationOrder{}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if e = pharmacyLock(ctx, tx, a, key, "medication"); e != nil {
		return out, e
	}
	var oldID, hash string
	wanted := requestHash(i)
	e = tx.QueryRow(ctx, `SELECT id,request_hash FROM medication_order WHERE signed_by=$1 AND request_key=$2`, a.ID, key).Scan(&oldID, &hash)
	if e == nil {
		if hash != wanted {
			return out, domain.ErrConflict
		}
		return scanOrder(tx.QueryRow(ctx, `SELECT `+orderFields+` FROM medication_order o WHERE o.id=$1`, oldID))
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return out, e
	}
	var status string
	e = tx.QueryRow(ctx, `SELECT status FROM encounter WHERE id=$1 AND doctor_id=$2 FOR UPDATE`, i.EncounterID, a.ID).Scan(&status)
	if e != nil {
		return out, clinicalError(e)
	}
	if status != "active" {
		return out, domain.ErrStale
	}
	var name, unit string
	e = tx.QueryRow(ctx, `SELECT name,unit FROM medicine WHERE id=$1 AND active FOR SHARE`, i.MedicineID).Scan(&name, &unit)
	if e != nil {
		return out, clinicalError(e)
	}
	var id string
	e = tx.QueryRow(ctx, `INSERT INTO medication_order(encounter_id,medicine_id,medicine_name,unit,quantity,dose,route,frequency,duration_days,instructions,signed_by,request_key,request_hash) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id`, i.EncounterID, i.MedicineID, name, unit, i.Quantity, i.Dose, i.Route, i.Frequency, i.DurationDays, i.Instructions, a.ID, key, wanted).Scan(&id)
	if e != nil {
		return out, e
	}
	out, e = scanOrder(tx.QueryRow(ctx, `SELECT `+orderFields+` FROM medication_order o WHERE o.id=$1`, id))
	if e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "medication.signed", id); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) CancelMedication(ctx context.Context, a domain.Actor, id, reason string) error {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var found string
	e = tx.QueryRow(ctx, `SELECT id FROM medication_order WHERE id=$1 AND signed_by=$2 FOR UPDATE`, id, a.ID).Scan(&found)
	if e != nil {
		return clinicalError(e)
	}
	var old string
	e = tx.QueryRow(ctx, `SELECT reason FROM medication_cancellation WHERE order_id=$1`, id).Scan(&old)
	if e == nil {
		if old != reason {
			return domain.ErrConflict
		}
		return nil
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO medication_cancellation(order_id,reason,actor_id) VALUES($1,$2,$3)`, id, reason, a.ID); e != nil {
		return e
	}
	if e = pharmacyAudit(ctx, tx, a, "medication.cancelled", id); e != nil {
		return e
	}
	return tx.Commit(ctx)
}

const stockFields = `id,batch_id,COALESCE(order_id::text,''),kind,quantity,COALESCE(original_id::text,''),reason,created_at`

func scanStock(row pgx.Row) (domain.StockMovement, error) {
	var m domain.StockMovement
	e := row.Scan(&m.ID, &m.BatchID, &m.OrderID, &m.Kind, &m.Quantity, &m.OriginalID, &m.Reason, &m.CreatedAt)
	return m, clinicalError(e)
}
func (s Store) MoveStock(ctx context.Context, a domain.Actor, i domain.StockInput, key, today string) (domain.StockMovement, error) {
	out := domain.StockMovement{}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if e = pharmacyLock(ctx, tx, a, key, "stock"); e != nil {
		return out, e
	}
	var oldID, hash string
	wanted := requestHash(i)
	e = tx.QueryRow(ctx, `SELECT id,request_hash FROM pharmacy_movement WHERE actor_id=$1 AND request_key=$2`, a.ID, key).Scan(&oldID, &hash)
	if e == nil {
		if hash != wanted {
			return out, domain.ErrConflict
		}
		return scanStock(tx.QueryRow(ctx, `SELECT `+stockFields+` FROM pharmacy_movement WHERE id=$1`, oldID))
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return out, e
	}
	// All dispensing and cancellation serialize on the signed order first.
	var med string
	var prescribed int
	if i.OrderID != "" {
		e = tx.QueryRow(ctx, `SELECT medicine_id,quantity FROM medication_order WHERE id=$1 FOR UPDATE`, i.OrderID).Scan(&med, &prescribed)
		if e != nil {
			return out, clinicalError(e)
		}
	}
	var batchMed, expiry string
	var balance int
	e = tx.QueryRow(ctx, `SELECT medicine_id,expiry_date::text,balance FROM medicine_batch WHERE id=$1 FOR UPDATE`, i.BatchID).Scan(&batchMed, &expiry, &balance)
	if e != nil {
		return out, clinicalError(e)
	}
	if med != "" && med != batchMed {
		return out, domain.ErrValidation
	}
	var salePrice *int64
	delta := -i.Quantity
	switch i.Kind {
	case "dispense":
		var active bool
		e = tx.QueryRow(ctx, `SELECT active,selling_price_minor FROM medicine WHERE id=$1 FOR SHARE`, batchMed).Scan(&active, &salePrice)
		if e != nil {
			return out, e
		}
		if !active {
			return out, domain.ErrStale
		}
		var cancelled bool
		var dispensed int
		e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM medication_cancellation WHERE order_id=$1),(SELECT COALESCE(sum(quantity),0)::integer FROM pharmacy_movement WHERE order_id=$1 AND kind='dispense')`, i.OrderID).Scan(&cancelled, &dispensed)
		if e != nil {
			return out, e
		}
		if cancelled || expiry < today || dispensed+i.Quantity > prescribed {
			return out, domain.ErrStale
		}
	case "return":
		var issued, returned int
		e = tx.QueryRow(ctx, `SELECT quantity,(SELECT COALESCE(sum(quantity),0)::integer FROM pharmacy_movement WHERE original_id=$1 AND kind='return') FROM pharmacy_movement WHERE id=$1 AND batch_id=$2 AND order_id=$3 AND kind='dispense'`, i.OriginalID, i.BatchID, i.OrderID).Scan(&issued, &returned)
		if e != nil {
			return out, clinicalError(e)
		}
		if returned+i.Quantity > issued {
			return out, domain.ErrStale
		}
		// Patient returns are quarantined; never increase dispensable balance or refill allowance.
		delta = 0
	}
	if balance+delta < 0 {
		return out, domain.ErrStale
	}
	out, e = scanStock(tx.QueryRow(ctx, `INSERT INTO pharmacy_movement(batch_id,order_id,kind,quantity,original_id,reason,actor_id,request_key,request_hash,unit_sale_price_minor) VALUES($1,NULLIF($2,'')::uuid,$3,$4,NULLIF($5,'')::uuid,$6,$7,$8,$9,$10) RETURNING `+stockFields, i.BatchID, i.OrderID, i.Kind, i.Quantity, i.OriginalID, i.Reason, a.ID, key, wanted, salePrice))
	if e != nil {
		return out, clinicalError(e)
	}
	if _, e = tx.Exec(ctx, `UPDATE medicine_batch SET balance=balance+$2 WHERE id=$1`, i.BatchID, delta); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "pharmacy."+i.Kind, out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) StockMovements(ctx context.Context, a domain.Actor, batch string, page int) ([]domain.StockMovement, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	rows, e := tx.Query(ctx, `SELECT `+stockFields+` FROM pharmacy_movement WHERE batch_id=$1 ORDER BY created_at DESC,id LIMIT 25 OFFSET $2`, batch, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.StockMovement{}
	for rows.Next() {
		m, e := scanStock(rows)
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
	if e = pharmacyAudit(ctx, tx, a, "pharmacy.movements_viewed", batch); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
