package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"hms.local/api/internal/domain"
)

const checkoutFields = `id,invoice_id,actor_id,amount_minor,provider_id,state,created_at`

func scanCheckout(row pgx.Row) (domain.Checkout, error) {
	var c domain.Checkout
	e := row.Scan(&c.ID, &c.InvoiceID, &c.ActorID, &c.AmountMinor, &c.ProviderID, &c.State, &c.CreatedAt)
	return c, clinicalError(e)
}
func (s Store) Checkout(ctx context.Context, a domain.Actor, id string) (domain.Checkout, error) {
	var out domain.Checkout
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return out, e
	}
	defer tx.Rollback(ctx)
	var balance int64
	e = tx.QueryRow(ctx, `SELECT i.total_minor-i.paid_minor FROM invoice i JOIN patient p ON p.id=i.patient_id WHERE i.id=$1 AND ($2 IN ('admin','accountant') OR ($2='patient' AND patient_portal_owner(p.id)=$3)) FOR UPDATE OF i`, id, a.Role, a.ID).Scan(&balance)
	if e != nil {
		return out, clinicalError(e)
	}
	out, e = scanCheckout(tx.QueryRow(ctx, `SELECT `+checkoutFields+` FROM payment_checkout WHERE invoice_id=$1`, id))
	if e == nil {
		if out.AmountMinor != balance {
			return out, domain.ErrStale
		}
		return out, tx.Commit(ctx)
	}
	if !errors.Is(e, domain.ErrNotFound) {
		return out, e
	}
	if balance < 1 || balance > 99999999 {
		return out, domain.ErrValidation
	}
	out, e = scanCheckout(tx.QueryRow(ctx, `INSERT INTO payment_checkout(invoice_id,actor_id,amount_minor) VALUES($1,$2,$3) RETURNING `+checkoutFields, id, a.ID, balance))
	if e != nil {
		return out, e
	}
	if e = pharmacyAudit(ctx, tx, a, "invoice.checkout_requested", id); e != nil {
		return out, e
	}
	return out, tx.Commit(ctx)
}
func (s Store) BindCheckout(ctx context.Context, c domain.Checkout, i domain.GatewayIntent) error {
	if i.AmountMinor != c.AmountMinor || i.Currency != "etb" || i.ID == "" {
		return domain.ErrValidation
	}
	result, e := s.DB.Exec(ctx, `UPDATE payment_checkout SET provider_id=$2,state='ready' WHERE id=$1 AND state IN ('pending','ready') AND (provider_id='' OR provider_id=$2)`, c.ID, i.ID)
	if e != nil {
		return clinicalError(e)
	}
	if result.RowsAffected() != 1 {
		return domain.ErrStale
	}
	return nil
}
func (s Store) ApplyPaymentEvent(ctx context.Context, event domain.GatewayEvent) error {
	tx, e := s.DB.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if _, e = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, "stripe-event:"+event.ID); e != nil {
		return e
	}
	var oldHash string
	e = tx.QueryRow(ctx, `SELECT payload_hash FROM payment_event WHERE event_id=$1`, event.ID).Scan(&oldHash)
	if e == nil {
		if oldHash != event.Hash {
			return domain.ErrConflict
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return e
	}
	save := func(state, reason string) error {
		_, err := tx.Exec(ctx, `INSERT INTO payment_event(event_id,payload_hash,intent_id,state,reason) VALUES($1,$2,$3,$4,$5)`, event.ID, event.Hash, event.IntentID, state, reason)
		return err
	}
	if event.Kind != "payment_intent.succeeded" || !domain.UUIDPattern.MatchString(event.CheckoutID) {
		if e = save("ignored", "Unrelated event"); e != nil {
			return e
		}
		return tx.Commit(ctx)
	}
	c, e := scanCheckout(tx.QueryRow(ctx, `SELECT `+checkoutFields+` FROM payment_checkout WHERE id=$1 FOR UPDATE`, event.CheckoutID))
	if errors.Is(e, domain.ErrNotFound) {
		if e = save("review", "Unknown checkout"); e != nil {
			return e
		}
		return tx.Commit(ctx)
	}
	if e != nil {
		return e
	}
	// A second event for the same provider intent cannot post a second payment.
	var posted bool
	if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM payment_receipt WHERE intent_id=$1)`, event.IntentID).Scan(&posted); e != nil {
		return e
	}
	if posted {
		if e = save("ignored", "Payment already posted"); e != nil {
			return e
		}
		return tx.Commit(ctx)
	}
	var paid, total int64
	if e = tx.QueryRow(ctx, `SELECT paid_minor,total_minor FROM invoice WHERE id=$1 FOR UPDATE`, c.InvoiceID).Scan(&paid, &total); e != nil {
		return e
	}
	reason := ""
	if event.InvoiceID != c.InvoiceID || event.AmountMinor != c.AmountMinor || event.Currency != "etb" || (c.ProviderID != "" && c.ProviderID != event.IntentID) {
		reason = "Provider values differ from checkout"
	} else if c.State == "paid" || c.State == "review" || paid+event.AmountMinor > total {
		reason = "Invoice balance or checkout requires reconciliation"
	}
	if reason != "" {
		if e = save("review", reason); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, `UPDATE payment_checkout SET state='review' WHERE id=$1 AND state<>'paid'`, c.ID); e != nil {
			return e
		}
		return tx.Commit(ctx)
	}
	if e = save("applied", "Verified provider payment"); e != nil {
		return e
	}
	var paymentID string
	e = tx.QueryRow(ctx, `INSERT INTO invoice_payment(invoice_id,amount_minor,direction,method,reference,actor_id,request_key,request_hash) VALUES($1,$2,'payment','stripe',$3,$4,$5,$6) RETURNING id`, c.InvoiceID, event.AmountMinor, event.IntentID, c.ActorID, "stripe:"+event.IntentID, event.Hash).Scan(&paymentID)
	if e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, `UPDATE invoice SET paid_minor=paid_minor+$2,version=version+1 WHERE id=$1`, c.InvoiceID, event.AmountMinor); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO payment_receipt(intent_id,checkout_id,payment_id,event_id) VALUES($1,$2,$3,$4)`, event.IntentID, c.ID, paymentID, event.ID); e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, `UPDATE payment_checkout SET state='paid',provider_id=$2 WHERE id=$1`, c.ID, event.IntentID); e != nil {
		return e
	}
	if e = pharmacyAudit(ctx, tx, domain.Actor{ID: c.ActorID}, "invoice.stripe_payment_verified", paymentID); e != nil {
		return e
	}
	return tx.Commit(ctx)
}
func (s Store) PaymentReviews(ctx context.Context, page int) ([]domain.PaymentReview, error) {
	rows, e := s.DB.Query(ctx, `SELECT event_id,intent_id,state,reason,received_at FROM payment_event WHERE state='review' ORDER BY received_at DESC,event_id LIMIT 25 OFFSET $1`, (page-1)*25)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []domain.PaymentReview{}
	for rows.Next() {
		var v domain.PaymentReview
		if e = rows.Scan(&v.EventID, &v.IntentID, &v.State, &v.Reason, &v.ReceivedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
