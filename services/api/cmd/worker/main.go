package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/delivery"
	"hms.local/api/internal/adapters/postgres"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
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
	dispatcher := application.Dispatcher{Store: postgres.Store{DB: db}, Transport: transport{}}
	slog.Info("Communications worker running")
	for ctx.Err() == nil {
		if err = dispatcher.Process(ctx); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("Message processing failed; no automatic redelivery")
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

type transport struct{}

func (transport) Send(ctx context.Context, m domain.MessageLease) (string, string, error) {
	result, err := delivery.Send(ctx, m.Channel, m.Recipient, m.Subject, m.Body)
	return result.Status, result.ProviderID, err
}
