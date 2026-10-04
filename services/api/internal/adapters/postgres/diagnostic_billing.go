package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

func (s Store) BillDiagnostic(ctx context.Context, a domain.Actor, order string, i domain.SourceInvoiceInput, key, date string) (domain.Invoice, error) {
	var out domain.Invoice
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if e = pharmacyLock(ctx, tx, a, key, "invoice"); e != nil {
		return out, e
	}
	var status, patient, name string
	var price int64
	e = tx.QueryRow(ctx, `SELECT o.status,e.patient_id,t.name,t.charge_minor FROM diagnostic_order o JOIN encounter e ON e.id=o.encounter_id JOIN diagnostic_test t ON t.id=o.test_id WHERE o.id=$1 FOR UPDATE OF o`, order).Scan(&status, &patient, &name, &price)
	if e != nil {
		return out, clinicalError(e)
	}
	var existing, account string
	var discount int64
	e = tx.QueryRow(ctx, `SELECT invoice_id,account_id,discount_basis_points FROM diagnostic_invoice WHERE order_id=$1`, order).Scan(&existing, &account, &discount)
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
	if status != "released" {
		return out, domain.ErrStale
	}
	var used bool
	if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM invoice WHERE created_by=$1 AND request_key=$2)`, a.ID, key).Scan(&used); e != nil {
		return out, e
	}
	if used {
		return out, domain.ErrConflict
	}
	hash := requestHash(struct {
		DiagnosticOrder string
		Input           domain.SourceInvoiceInput
	}{order, i})
	id, e := insertSourceInvoice(ctx, tx, a, patient, name, 1, price, i, key, date, hash)
	if e != nil {
		return out, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO diagnostic_invoice(order_id,invoice_id,account_id,discount_basis_points,created_by) VALUES($1,$2,$3,$4,$5)`, order, id, i.AccountID, i.DiscountBasisPoints, a.ID); e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "diagnostic.invoice_issued", id); e != nil {
		return out, e
	}
	if e = tx.Commit(ctx); e != nil {
		return out, e
	}
	return s.Invoice(ctx, a, id)
}
