package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/delivery"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if os.Getenv("DATABASE_URL") == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("Database initialization failed")
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Communications worker running")
	for ctx.Err() == nil {
		if err = process(ctx, db); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("Message processing failed; no automatic redelivery")
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
func process(ctx context.Context, db *pgxpool.Pool) error {
	var id, channel, to, subject, body string
	// Atomic claim prevents multiple workers from dispatching the same queued row.
	err := db.QueryRow(ctx, `UPDATE message_outbox SET status='processing',updated_at=now() WHERE id=(SELECT id FROM message_outbox WHERE status='pending' ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING id,channel,recipient,subject,body`).Scan(&id, &channel, &to, &subject, &body)
	if err != nil {
		return err
	}
	result, sendErr := delivery.Send(ctx, channel, to, subject, body)
	// A crash or network ambiguity must be reconciled, not blindly retried.
	finish, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = db.Exec(finish, `UPDATE message_outbox SET status=$2,provider_id=$3,updated_at=now() WHERE id=$1`, id, result.Status, result.ProviderID)
	if err != nil {
		return err
	}
	if sendErr != nil {
		slog.Warn("Message delivery requires review", "message_id", id, "status", result.Status)
	}
	return nil
}
