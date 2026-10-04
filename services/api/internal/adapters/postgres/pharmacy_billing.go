package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

func (s Store) BillDispensing(ctx context.Context, a domain.Actor, movement string, i domain.PharmacyInvoiceInput, key, date string) (domain.Invoice, error) {
	out := domain.Invoice{}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	// Share the invoice key lock namespace with manually created invoices.
	if e = pharmacyLock(ctx, tx, a, key, "invoice"); e != nil {
		return out, e
	}
	var quantity int64
	var price *int64
	var order string
	e = tx.QueryRow(ctx, `SELECT quantity,unit_sale_price_minor,order_id FROM pharmacy_movement WHERE id=$1 AND kind='dispense' FOR UPDATE`, movement).Scan(&quantity, &price, &order)
	if e != nil {
		return out, clinicalError(e)
	}
	var existing, account string
	var discount int64
	e = tx.QueryRow(ctx, `SELECT invoice_id,account_id,discount_basis_points FROM pharmacy_invoice WHERE movement_id=$1`, movement).Scan(&existing, &account, &discount)
	if e == nil {
		if account != i.AccountID || discount != i.DiscountBasisPoints {
			return out, domain.ErrConflict
		}
		if e = tx.Commit(ctx); e != nil {
			return out, e
		}
		return s.Invoice(ctx, a, existing)
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return out, e
	}
	var used bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM invoice WHERE created_by=$1 AND request_key=$2)`, a.ID, key).Scan(&used)
	if e != nil {
		return out, e
	}
	if used {
		return out, domain.ErrConflict
	}
	if price == nil || *price == 0 {
		return out, domain.ErrValidation
	}
	var patient, name string
	e = tx.QueryRow(ctx, `SELECT e.patient_id,o.medicine_name FROM medication_order o JOIN encounter e ON e.id=o.encounter_id WHERE o.id=$1`, order).Scan(&patient, &name)
	if e != nil {
		return out, e
	}
	e = tx.QueryRow(ctx, `SELECT name FROM charge_account WHERE id=$1 AND active FOR SHARE`, i.AccountID).Scan(&account)
	if e != nil {
		return out, clinicalError(e)
	}
	subtotal := quantity * *price
	total := subtotal - (subtotal*i.DiscountBasisPoints+5000)/10000
	if subtotal > 1000000000000 || total < 1 {
		return out, domain.ErrValidation
	}
	hash := requestHash(struct {
		Movement string
		Input    domain.PharmacyInvoiceInput
	}{movement, i})
	var id string
	e = tx.QueryRow(ctx, `INSERT INTO invoice(patient_id,invoice_date,subtotal_minor,discount_basis_points,total_minor,created_by,request_key,request_hash) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, patient, date, subtotal, i.DiscountBasisPoints, total, a.ID, key, hash).Scan(&id)
	if e != nil {
		return out, clinicalError(e)
	}
	if _, e = tx.Exec(ctx, `INSERT INTO invoice_line(invoice_id,position,account_id,account_name,description,quantity,unit_price_minor) VALUES($1,1,$2,$3,$4,$5,$6)`, id, i.AccountID, account, name, quantity, *price); e != nil {
		return out, e
	}
	if _, e = tx.Exec(ctx, `UPDATE invoice SET sealed=true WHERE id=$1`, id); e != nil {
		return out, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO pharmacy_invoice(movement_id,invoice_id,account_id,discount_basis_points,created_by) VALUES($1,$2,$3,$4,$5)`, movement, id, i.AccountID, i.DiscountBasisPoints, a.ID); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "pharmacy.invoice_issued", id); e != nil {
		return out, e
	}
	if e = tx.Commit(ctx); e != nil {
		return out, e
	}
	return s.Invoice(ctx, a, id)
}
