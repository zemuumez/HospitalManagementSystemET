package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"hms.local/api/internal/adapters/httpapi"
	"hms.local/api/internal/adapters/postgres"
	"hms.local/api/internal/application"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	authURL := os.Getenv("BETTER_AUTH_URL")
	u, err := url.Parse(authURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || os.Getenv("DATABASE_URL") == "" {
		slog.Error("Set DATABASE_URL and BETTER_AUTH_URL")
		os.Exit(1)
	}
	if os.Getenv("APP_ENV") == "production" && u.Scheme != "https" {
		slog.Error("Production requires HTTPS authentication")
		os.Exit(1)
	}
	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("Invalid database configuration")
		os.Exit(1)
	}
	cfg.MaxConns = 10
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error("Database initialization failed")
		os.Exit(1)
	}
	defer db.Close()
	if err = db.Ping(ctx); err != nil {
		slog.Error("Database unavailable")
		os.Exit(1)
	}
	store := postgres.Store{DB: db}
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	handler := httpapi.Server{App: application.Hospital{Store: store, Now: time.Now}, Scheduling: application.Scheduling{Store: store, Now: time.Now}, Clinical: application.Clinical{Store: store, Now: time.Now}, Billing: application.Billing{Store: store, Now: time.Now}, Pharmacy: application.Pharmacy{Store: store, Now: time.Now}, Diagnostics: application.Diagnostics{Store: store}, Actors: store, AuthURL: authURL, Origin: authURL, Client: &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}.Handler()
	server := &http.Server{Addr: addr, Handler: http.TimeoutHandler(handler, 15*time.Second, `{"error":"Request timed out"}`), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	slog.Info("Hospital API listening", "address", addr)
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Server failed")
		os.Exit(1)
	}
}
