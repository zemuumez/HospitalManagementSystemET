package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

// insertSourceInvoice is shared by source workflows. Caller holds source/key locks.
func insertSourceInvoice(ctx context.Context, tx pgx.Tx, a domain.Actor, patient, name string, quantity, price int64, i domain.SourceInvoiceInput, key, date, hash string) (string, error) {
	if price < 1 || quantity < 1 || quantity > 1000000000000/price || i.DiscountBasisPoints < 0 || i.DiscountBasisPoints >= 10000 {
		return "", domain.ErrValidation
	}
	subtotal := quantity * price
	total := subtotal - (subtotal*i.DiscountBasisPoints+5000)/10000
	if total < 1 {
		return "", domain.ErrValidation
	}
	var account string
	if e := tx.QueryRow(ctx, `SELECT name FROM charge_account WHERE id=$1 AND active FOR SHARE`, i.AccountID).Scan(&account); e != nil {
		return "", clinicalError(e)
	}
	var id string
	e := tx.QueryRow(ctx, `INSERT INTO invoice(patient_id,invoice_date,subtotal_minor,discount_basis_points,total_minor,created_by,request_key,request_hash) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, patient, date, subtotal, i.DiscountBasisPoints, total, a.ID, key, hash).Scan(&id)
	if e != nil {
		return "", clinicalError(e)
	}
	if _, e = tx.Exec(ctx, `INSERT INTO invoice_line(invoice_id,position,account_id,account_name,description,quantity,unit_price_minor) VALUES($1,1,$2,$3,$4,$5,$6)`, id, i.AccountID, account, name, quantity, price); e != nil {
		return "", e
	}
	if _, e = tx.Exec(ctx, `UPDATE invoice SET sealed=true WHERE id=$1`, id); e != nil {
		return "", e
	}
	return id, nil
}
