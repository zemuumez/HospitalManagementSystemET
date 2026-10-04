package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

func requestHash(value any) string {
	data, _ := json.Marshal(value)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
func (s Store) ChargeAccounts(ctx context.Context, page int) ([]domain.ChargeAccount, error) {
	rows, e := s.DB.Query(ctx, `SELECT id,name FROM charge_account WHERE active ORDER BY name,id LIMIT 25 OFFSET $1`, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.ChargeAccount{}
	for rows.Next() {
		var a domain.ChargeAccount
		if e = rows.Scan(&a.ID, &a.Name); e != nil {
			return nil, e
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
func (s Store) CreateChargeAccount(ctx context.Context, a domain.Actor, name string) (domain.ChargeAccount, error) {
	out := domain.ChargeAccount{Name: name}
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	e = tx.QueryRow(ctx, `INSERT INTO charge_account(name,created_by) VALUES($1,$2) RETURNING id`, name, a.ID).Scan(&out.ID)
	if e != nil {
		return out, clinicalError(e)
	}
	if _, e = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'charge_account.created',$2)`, a.ID, out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}

const invoiceFields = `i.id,i.number,i.patient_id,p.given_name||' '||p.family_name,i.invoice_date::text,i.subtotal_minor,i.discount_basis_points,i.total_minor,i.paid_minor,i.version`
const invoiceFrom = ` FROM invoice i JOIN patient p ON p.id=i.patient_id `
const invoiceScope = `($1 IN ('admin','accountant') OR ($1='patient' AND p.user_id=$2))`

func scanInvoice(row pgx.Row) (domain.Invoice, error) {
	i := domain.Invoice{Lines: []domain.InvoiceLine{}}
	e := row.Scan(&i.ID, &i.Number, &i.PatientID, &i.PatientName, &i.InvoiceDate, &i.SubtotalMinor, &i.DiscountBasisPoints, &i.TotalMinor, &i.PaidMinor, &i.Version)
	return i, clinicalError(e)
}
func invoiceLines(ctx context.Context, q querier, i *domain.Invoice) error {
	rows, e := q.Query(ctx, `SELECT account_id,account_name,description,quantity,unit_price_minor FROM invoice_line WHERE invoice_id=$1 ORDER BY position`, i.ID)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		var l domain.InvoiceLine
		if e = rows.Scan(&l.AccountID, &l.AccountName, &l.Description, &l.Quantity, &l.UnitPriceMinor); e != nil {
			return e
		}
		i.Lines = append(i.Lines, l)
	}
	return rows.Err()
}
func (s Store) Invoices(ctx context.Context, a domain.Actor, page int) ([]domain.Invoice, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	rows, e := tx.Query(ctx, `SELECT `+invoiceFields+invoiceFrom+`WHERE `+invoiceScope+` ORDER BY i.created_at DESC,i.id LIMIT 25 OFFSET $3`, a.Role, a.ID, (page-1)*25)
	if e != nil {
		return nil, e
	}
	out := []domain.Invoice{}
	for rows.Next() {
		i, err := scanInvoice(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, i)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action) VALUES($1,'invoices.list_viewed')`, a.ID); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) Invoice(ctx context.Context, a domain.Actor, id string) (domain.Invoice, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return domain.Invoice{}, e
	}
	defer tx.Rollback(ctx)
	out, e := scanInvoice(tx.QueryRow(ctx, `SELECT `+invoiceFields+invoiceFrom+`WHERE `+invoiceScope+` AND i.id=$3`, a.Role, a.ID, id))
	if e != nil {
		return out, e
	}
	if e = invoiceLines(ctx, tx, &out); e != nil {
		return out, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'invoice.viewed',$2)`, a.ID, id); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) CreateInvoice(ctx context.Context, a domain.Actor, i domain.InvoiceInput, subtotal, total int64, key string) (domain.Invoice, error) {
	out := domain.Invoice{}
	hash := requestHash(i)
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, "invoice:"+a.ID+":"+key); e != nil {
		return out, e
	}
	var oldID, oldHash string
	e = tx.QueryRow(ctx, `SELECT id,request_hash FROM invoice WHERE created_by=$1 AND request_key=$2`, a.ID, key).Scan(&oldID, &oldHash)
	if e == nil {
		if hash != oldHash {
			return out, domain.ErrConflict
		}
		if e = tx.Commit(ctx); e != nil {
			return out, e
		}
		return s.Invoice(ctx, a, oldID)
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return out, e
	}
	var exists bool
	e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM patient WHERE id=$1)`, i.PatientID).Scan(&exists)
	if e != nil {
		return out, e
	}
	if !exists {
		return out, domain.ErrNotFound
	}
	for n := range i.Lines {
		e = tx.QueryRow(ctx, `SELECT name FROM charge_account WHERE id=$1 AND active FOR SHARE`, i.Lines[n].AccountID).Scan(&i.Lines[n].AccountName)
		if e != nil {
			return out, clinicalError(e)
		}
	}
	var id string
	e = tx.QueryRow(ctx, `INSERT INTO invoice(patient_id,invoice_date,subtotal_minor,discount_basis_points,total_minor,created_by,request_key,request_hash) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, i.PatientID, i.InvoiceDate, subtotal, i.DiscountBasisPoints, total, a.ID, key, hash).Scan(&id)
	if e != nil {
		return out, clinicalError(e)
	}
	for n, l := range i.Lines {
		if _, e = tx.Exec(ctx, `INSERT INTO invoice_line(invoice_id,position,account_id,account_name,description,quantity,unit_price_minor) VALUES($1,$2,$3,$4,$5,$6,$7)`, id, n+1, l.AccountID, l.AccountName, l.Description, l.Quantity, l.UnitPriceMinor); e != nil {
			return out, e
		}
	}
	if _, e = tx.Exec(ctx, `UPDATE invoice SET sealed=true WHERE id=$1`, id); e != nil {
		return out, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,'invoice.issued',$2)`, a.ID, id); e != nil {
		return out, e
	}
	if e = tx.Commit(ctx); e != nil {
		return out, e
	}
	return s.Invoice(ctx, a, id)
}
func scanPayment(row pgx.Row) (domain.Payment, error) {
	var p domain.Payment
	e := row.Scan(&p.ID, &p.AmountMinor, &p.Method, &p.Reference, &p.OriginalPaymentID, &p.Reason, &p.Direction, &p.CreatedAt)
	return p, clinicalError(e)
}

const paymentFields = `id,amount_minor,method,reference,COALESCE(original_payment_id::text,''),reason,direction,created_at`

func (s Store) Payments(ctx context.Context, a domain.Actor, id string) ([]domain.Payment, error) {
	if _, e := s.Invoice(ctx, a, id); e != nil {
		return nil, e
	}
	rows, e := s.DB.Query(ctx, `SELECT `+paymentFields+` FROM invoice_payment WHERE invoice_id=$1 ORDER BY created_at DESC,id LIMIT 100`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Payment{}
	for rows.Next() {
		p, e := scanPayment(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s Store) RecordPayment(ctx context.Context, a domain.Actor, id string, i domain.PaymentInput, key string) (domain.Payment, error) {
	out := domain.Payment{}
	hash := requestHash(struct {
		Invoice string
		Input   domain.PaymentInput
	}{id, i})
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, "payment:"+a.ID+":"+key); e != nil {
		return out, e
	}
	var oldHash, oldID string
	e = tx.QueryRow(ctx, `SELECT id,request_hash FROM invoice_payment WHERE actor_id=$1 AND request_key=$2`, a.ID, key).Scan(&oldID, &oldHash)
	if e == nil {
		if oldHash != hash {
			return out, domain.ErrConflict
		}
		out, e = scanPayment(tx.QueryRow(ctx, `SELECT `+paymentFields+` FROM invoice_payment WHERE id=$1`, oldID))
		if e != nil {
			return out, e
		}
		return out, tx.Commit(ctx)
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return out, e
	}
	var paid, total int64
	e = tx.QueryRow(ctx, `SELECT paid_minor,total_minor FROM invoice WHERE id=$1 FOR UPDATE`, id).Scan(&paid, &total)
	if e != nil {
		return out, clinicalError(e)
	}
	direction := "payment"
	delta := i.AmountMinor
	if i.OriginalPaymentID != "" {
		direction = "refund"
		delta = -delta
		var amount, refunded int64
		e = tx.QueryRow(ctx, `SELECT amount_minor FROM invoice_payment WHERE id=$1 AND invoice_id=$2 AND direction='payment'`, i.OriginalPaymentID, id).Scan(&amount)
		if e != nil {
			return out, clinicalError(e)
		}
		e = tx.QueryRow(ctx, `SELECT COALESCE(sum(amount_minor),0)::bigint FROM invoice_payment WHERE original_payment_id=$1 AND direction='refund'`, i.OriginalPaymentID).Scan(&refunded)
		if e != nil {
			return out, e
		}
		if refunded+i.AmountMinor > amount {
			return out, domain.ErrStale
		}
	}
	if paid+delta < 0 || paid+delta > total {
		return out, domain.ErrStale
	}
	out, e = scanPayment(tx.QueryRow(ctx, `INSERT INTO invoice_payment(invoice_id,amount_minor,direction,method,reference,reason,original_payment_id,actor_id,request_key,request_hash) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,'')::uuid,$8,$9,$10) RETURNING `+paymentFields, id, i.AmountMinor, direction, i.Method, i.Reference, i.Reason, i.OriginalPaymentID, a.ID, key, hash))
	if e != nil {
		return out, clinicalError(e)
	}
	if _, e = tx.Exec(ctx, `UPDATE invoice SET paid_minor=paid_minor+$2,version=version+1 WHERE id=$1`, id, delta); e != nil {
		return out, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action,resource_id) VALUES($1,$2,$3)`, a.ID, "invoice."+direction+"_recorded", out.ID); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) BillingPatients(ctx context.Context, a domain.Actor, search string, page int) ([]domain.Patient, error) {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	rows, e := tx.Query(ctx, `SELECT id,medical_record_number::text,given_name,family_name FROM patient WHERE ($1='' OR strpos(lower(given_name||' '||family_name),lower($1))>0 OR medical_record_number::text=$1) ORDER BY given_name,family_name,id LIMIT 25 OFFSET $2`, search, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.Patient{}
	for rows.Next() {
		var p domain.Patient
		if e = rows.Scan(&p.ID, &p.MRN, &p.GivenName, &p.FamilyName); e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO audit_event(actor_id,action) VALUES($1,'billing_patient_index.viewed')`, a.ID); e != nil {
		return nil, e
	}
	return out, tx.Commit(ctx)
}
