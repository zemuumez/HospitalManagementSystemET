package postgres

import (
	"context"
	"hms.local/api/internal/domain"
)

func (s Store) RecoverMessages(ctx context.Context) (int64, error) {
	result, e := s.DB.Exec(ctx, `UPDATE message_outbox SET status='uncertain',lease_token=NULL,lease_until=NULL,updated_at=clock_timestamp() WHERE id IN (SELECT id FROM message_outbox WHERE status='processing' AND lease_until<clock_timestamp() ORDER BY lease_until FOR UPDATE SKIP LOCKED LIMIT 100)`)
	return result.RowsAffected(), e
}
func (s Store) ClaimMessage(ctx context.Context) (domain.MessageLease, error) {
	var m domain.MessageLease
	e := s.DB.QueryRow(ctx, `UPDATE message_outbox SET status='processing',lease_token=gen_random_uuid(),lease_until=clock_timestamp()+interval '45 seconds',attempts=attempts+1,updated_at=clock_timestamp() WHERE id=(SELECT id FROM message_outbox WHERE status='pending' ORDER BY created_at,id FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id,lease_token,channel,recipient,subject,body`).Scan(&m.ID, &m.Token, &m.Channel, &m.Recipient, &m.Subject, &m.Body)
	return m, e
}
func (s Store) RenewMessage(ctx context.Context, m domain.MessageLease) error {
	result, e := s.DB.Exec(ctx, `UPDATE message_outbox SET lease_until=clock_timestamp()+interval '45 seconds',updated_at=clock_timestamp() WHERE id=$1 AND lease_token=$2 AND status='processing' AND lease_until>=clock_timestamp()`, m.ID, m.Token)
	if e != nil {
		return e
	}
	if result.RowsAffected() != 1 {
		return domain.ErrStale
	}
	return nil
}
func (s Store) FinishMessage(ctx context.Context, m domain.MessageLease, status, provider string) error {
	if status != "sent" && status != "captured" && status != "failed" && status != "uncertain" {
		return domain.ErrValidation
	}
	result, e := s.DB.Exec(ctx, `UPDATE message_outbox SET status=$3,provider_id=$4,lease_token=NULL,lease_until=NULL,updated_at=clock_timestamp() WHERE id=$1 AND lease_token=$2 AND status='processing' AND lease_until>=clock_timestamp()`, m.ID, m.Token, status, provider)
	if e != nil {
		return e
	}
	if result.RowsAffected() != 1 {
		return domain.ErrStale
	}
	return nil
}
