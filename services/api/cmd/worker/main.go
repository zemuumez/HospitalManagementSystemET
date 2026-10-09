package main

import (
	"context"
	"errors"
	"flag"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/delivery"
	"hms.local/api/internal/adapters/postgres"
	"hms.local/api/internal/adapters/privatefiles"
	"hms.local/api/internal/application"
	"hms.local/api/internal/domain"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	cleanupFlag := flag.Bool("cleanup-attachments", false, "Run abandoned attachment cleanup maintenance and exit")
	olderThanFlag := flag.Duration("older-than", 24*time.Hour, "Threshold age for abandoned attachments cleanup")
	flag.Parse()

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

	var files application.AttachmentFiles
	attachmentDir := os.Getenv("HMS_ATTACHMENT_DIR")
	if attachmentDir == "" && os.Getenv("APP_ENV") != "production" {
		attachmentDir = filepath.Join(".", ".local", "attachments")
		_ = os.MkdirAll(attachmentDir, 0700)
	}
	if attachmentDir != "" {
		storage, e := privatefiles.New(attachmentDir)
		if e == nil {
			defer storage.Close()
			files = storage
		}
	}

	store := postgres.Store{DB: db}
	cmsSettings := application.CMSSettingsService{Store: store, Files: files, Now: time.Now}

	// Resolve admin user for operational audit trail
	var adminID string
	_ = db.QueryRow(ctx, `SELECT id FROM "user" WHERE role = 'admin' ORDER BY id ASC LIMIT 1`).Scan(&adminID)
	systemAdmin := domain.Actor{ID: adminID, Role: "admin"}

	// If invoked as a maintenance command:
	if *cleanupFlag {
		cleaned, err := cmsSettings.RunOperationalAttachmentCleanup(ctx, systemAdmin, *olderThanFlag)
		if err != nil {
			slog.Error("Operational attachment cleanup failed", "error", err)
			os.Exit(1)
		}
		slog.Info("Operational attachment cleanup completed", "cleaned", cleaned, "older_than", *olderThanFlag)
		return
	}

	dispatcher := application.Dispatcher{Store: store, Transport: transport{}}
	slog.Info("Communications worker running")

	cleanupTicker := time.NewTicker(time.Hour)
	defer cleanupTicker.Stop()

	for ctx.Err() == nil {
		if err = dispatcher.Process(ctx); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			slog.Error("Message processing failed; no automatic redelivery")
		}
		select {
		case <-ctx.Done():
			return
		case <-cleanupTicker.C:
			if cleaned, err := cmsSettings.RunOperationalAttachmentCleanup(ctx, systemAdmin, 24*time.Hour); err != nil {
				slog.Error("Scheduled abandoned attachment cleanup failed", "error", err)
			} else if cleaned > 0 {
				slog.Info("Scheduled abandoned attachment cleanup completed", "cleaned", cleaned)
			}
		case <-time.After(time.Second):
		}
	}
}

type transport struct{}

func (transport) Send(ctx context.Context, m domain.MessageLease) (string, string, error) {
	result, err := delivery.Send(ctx, m.Channel, m.Recipient, m.Subject, m.Body)
	return result.Status, result.ProviderID, err
}
