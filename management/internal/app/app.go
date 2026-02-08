package app

import (
	grpcapp "github.com/makhtech/management/internal/app/gprc"
	"github.com/makhtech/management/internal/config"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(cfg *config.Config) *App {

	// todo зарегать сервисы тут
	//authService := auth.New(log, cfg, kafkaProducer, postgresRepo, postgresRepo, postgresRepo, redisRepo)

	grpcApp := grpcapp.New(cfg)

	return &App{
		GRPCSrv: grpcApp,
	}
}
