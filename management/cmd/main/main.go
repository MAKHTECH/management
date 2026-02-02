package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/makhkets/managment/internal/config"
	"github.com/makhkets/managment/pkg/database/postgres"
	"github.com/makhkets/managment/pkg/logging"
)

func main() {
	logging.SetupLogger()

	cfg := config.MustLoad()

	slog.Info("starting application",
		slog.String("env", cfg.Env),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	db, err := postgres.New(ctx, cfg.Database.ToPostgresConfig())
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("database pool stats",
		slog.Int("total_conns", int(db.Stats().TotalConns())),
		slog.Int("idle_conns", int(db.Stats().IdleConns())),
	)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("application started, waiting for shutdown signal...")

	<-quit

	slog.Info("shutting down gracefully...")

	db.Close()

	slog.Info("application stopped")
}
