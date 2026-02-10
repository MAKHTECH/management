package app

import (
	"context"
	"log/slog"
	"time"

	grpcapp "github.com/makhtech/management/internal/app/gprc"
	"github.com/makhtech/management/internal/clients/sso"
	"github.com/makhtech/management/internal/config"
	"github.com/makhtech/management/internal/repository/postgres"
	planService "github.com/makhtech/management/internal/service/plan"
	"github.com/makhtech/management/pkg/ratelimiter"
)

type App struct {
	GRPCSrv     *grpcapp.App
	SSOClient   *sso.Client
	RateLimiter *ratelimiter.TokenBucket
}

func New(cfg *config.Config, db *postgres.Database) *App {
	// Создаём Rate Limiter
	rateLimiterCfg := ratelimiter.Config{
		Rate:            cfg.RateLimiter.GetRate(),
		Capacity:        cfg.RateLimiter.GetCapacity(),
		CleanupInterval: cfg.RateLimiter.GetCleanupInterval(),
	}
	rl := ratelimiter.New(rateLimiterCfg)

	slog.Info("rate limiter initialized",
		slog.Int("rate", rateLimiterCfg.Rate),
		slog.Int("capacity", rateLimiterCfg.Capacity),
	)

	// Создаём SSO клиент
	address, timeout, insecure := cfg.SSO.ToSSOClientConfig()
	ssoClient, err := sso.New(context.Background(), sso.Config{
		Address:  address,
		Timeout:  timeout,
		Insecure: insecure,
	})
	if err != nil {
		slog.Warn("failed to connect to SSO service, continuing without SSO",
			slog.String("error", err.Error()),
			slog.String("address", address),
		)
		// Не паникуем - SSO может быть недоступен при старте
		ssoClient = nil
	}

	// Создаём репозитории
	planRepo := postgres.NewPlanRepository(db)

	// Создаём сервисы
	planSvc := planService.New(planRepo, slog.Default())

	// Создаём gRPC App с SSO клиентом, Rate Limiter и сервисами
	grpcApp := grpcapp.New(cfg, ssoClient, rl, planSvc)

	return &App{
		GRPCSrv:     grpcApp,
		SSOClient:   ssoClient,
		RateLimiter: rl,
	}
}

// Stop останавливает все компоненты приложения
func (a *App) Stop() {
	if a.SSOClient != nil {
		if err := a.SSOClient.Close(); err != nil {
			slog.Warn("failed to close SSO client", slog.String("error", err.Error()))
		}
	}
	a.GRPCSrv.Stop()
}

// MustConnectSSO пытается подключиться к SSO сервису с ретраями
func (a *App) MustConnectSSO(cfg *config.Config, maxRetries int) {
	if a.SSOClient != nil {
		return // Уже подключены
	}

	address, timeout, insecure := cfg.SSO.ToSSOClientConfig()

	for i := 0; i < maxRetries; i++ {
		ssoClient, err := sso.New(context.Background(), sso.Config{
			Address:  address,
			Timeout:  timeout,
			Insecure: insecure,
		})
		if err != nil {
			slog.Warn("SSO connection attempt failed",
				slog.Int("attempt", i+1),
				slog.Int("max_retries", maxRetries),
				slog.String("error", err.Error()),
			)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		a.SSOClient = ssoClient
		slog.Info("SSO client connected successfully")
		return
	}

	slog.Error("failed to connect to SSO service after all retries")
}
