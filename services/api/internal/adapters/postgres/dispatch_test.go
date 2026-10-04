package postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"sync"
	"testing"
)

type captureTransport struct{}

func (captureTransport) Send(ctx context.Context, m domain.MessageLease) (string, string, error) {
	return "captured", "", nil
}
func testDispatch(t *testing.T, db *pgxpool.Pool, s Store) {
	t.Helper()
	ctx := context.Background()
	insert := func(key string) string {
		t.Helper()
		var id string
		if e := db.QueryRow(ctx, `INSERT INTO message_outbox(actor_id,channel,recipient,body,idempotency_key) VALUES('admin','sms','+251911111111','Synthetic', $1) RETURNING id`, key).Scan(&id); e != nil {
			t.Fatal(e)
		}
		return id
	}
	id := insert("lease-test-1")
	var leases [2]domain.MessageLease
	var errs [2]error
	var wg sync.WaitGroup
	for n := 0; n < 2; n++ {
		wg.Add(1)
		go func(n int) { defer wg.Done(); leases[n], errs[n] = s.ClaimMessage(ctx) }(n)
	}
	wg.Wait()
	winner := 0
	if errs[0] != nil {
		winner = 1
	}
	if errs[winner] != nil || !errors.Is(errs[1-winner], pgx.ErrNoRows) || leases[winner].ID != id {
		t.Fatal("claim race", errs)
	}
	lease := leases[winner]
	if e := s.RenewMessage(ctx, lease); e != nil {
		t.Fatal(e)
	}
	if _, e := db.Exec(ctx, `UPDATE message_outbox SET lease_until=now()-interval '1 second' WHERE id=$1`, id); e != nil {
		t.Fatal(e)
	}
	if e := s.FinishMessage(ctx, lease, "sent", "provider"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("expired worker finalized", e)
	}
	n, e := s.RecoverMessages(ctx)
	if e != nil || n != 1 {
		t.Fatal("crash recovery", n, e)
	}
	if e = s.RenewMessage(ctx, lease); !errors.Is(e, domain.ErrStale) {
		t.Fatal("stale lease renewed", e)
	}
	if e = s.FinishMessage(ctx, lease, "sent", "provider"); !errors.Is(e, domain.ErrStale) {
		t.Fatal("stale worker overwrote recovery", e)
	}
	if _, e = s.ClaimMessage(ctx); !errors.Is(e, pgx.ErrNoRows) {
		t.Fatal("uncertain message resent", e)
	}
	var status string
	var attempts int
	if e = db.QueryRow(ctx, `SELECT status,attempts FROM message_outbox WHERE id=$1`, id).Scan(&status, &attempts); e != nil || status != "uncertain" || attempts != 1 {
		t.Fatal("recovery state", status, attempts, e)
	}
	id = insert("lease-test-2")
	dispatcher := application.Dispatcher{Store: s, Transport: captureTransport{}}
	if e = dispatcher.Process(ctx); e != nil {
		t.Fatal(e)
	}
	if e = db.QueryRow(ctx, `SELECT status,attempts FROM message_outbox WHERE id=$1`, id).Scan(&status, &attempts); e != nil || status != "captured" || attempts != 1 {
		t.Fatal("dispatch completion", status, attempts, e)
	}
	if e = dispatcher.Process(ctx); !errors.Is(e, pgx.ErrNoRows) {
		t.Fatal("captured message resent", e)
	}
}
