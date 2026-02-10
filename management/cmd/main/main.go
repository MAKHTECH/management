package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/makhtech/management/cmd/migrator"
	"github.com/makhtech/management/internal/app"
	"github.com/makhtech/management/internal/config"
	"github.com/makhtech/management/internal/repository/postgres"
	"github.com/makhtech/management/pkg/directories"
	"github.com/makhtech/management/pkg/logging"
)

func main() {
	logging.SetupLogger()

	cfg := config.MustLoad()

	slog.Info("starting application",
		slog.Any("config", cfg),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	pgCfg := cfg.Database.ToPostgresConfig()
	postgresPort, _ := strconv.Atoi(cfg.Database.Port)
	db, err := postgres.New(ctx, pgCfg)
	if err != nil {
		slog.Error("failed to connect to repository", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Применяем миграции при старте приложения
	migrationsPath := directories.FindDirectoryName("migrations")
	err = migrator.ApplyMigrations(
		migrator.PostgresConfig{
			Host:     pgCfg.Host,
			Port:     postgresPort,
			User:     pgCfg.User,
			Password: pgCfg.Password,
			DBName:   pgCfg.DBName,
			SSLMode:  pgCfg.SSLMode,
		},
		migrationsPath,
		"migrations",
	)
	if err != nil {
		slog.Error("failed to apply migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("repository pool stats",
		slog.Int("total_conns", int(db.Stats().TotalConns())),
		slog.Int("idle_conns", int(db.Stats().IdleConns())),
	)

	application := app.New(cfg, db)
	go application.GRPCSrv.MustRun()

	slog.Info("GRPC server is running on port", slog.Int("port", cfg.GRPC.Port))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("application started, waiting for shutdown signal...")

	<-quit

	slog.Info("shutting down gracefully...")

	db.Close()

	slog.Info("application stopped")
}
