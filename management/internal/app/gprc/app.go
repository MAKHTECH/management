package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/makhtech/management/internal/clients/sso"
	"github.com/makhtech/management/internal/config"
	grpcInt "github.com/makhtech/management/internal/grpc"
	"github.com/makhtech/management/internal/service"
	"github.com/makhtech/management/pkg/ratelimiter"
	managementv1 "github.com/makhtech/proto/gen/go/management"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	gRPCServer      *grpc.Server
	port            int
	authInterceptor *grpcInt.AuthInterceptor
}

func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var chainHandler grpc.UnaryHandler = handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			innerHandler := chainHandler
			currentInterceptor := interceptors[i]
			chainHandler = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return currentInterceptor(currentCtx, currentReq, info, innerHandler)
			}
		}
		return chainHandler(ctx, req)
	}
}

func chainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var chainHandler grpc.StreamHandler = handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			innerHandler := chainHandler
			currentInterceptor := interceptors[i]
			chainHandler = func(currentSrv interface{}, currentStream grpc.ServerStream) error {
				return currentInterceptor(currentSrv, currentStream, info, innerHandler)
			}
		}
		return chainHandler(srv, stream)
	}
}

func New(cfg *config.Config, ssoClient *sso.Client, rateLimiter *ratelimiter.TokenBucket, planSvc service.PlanService) *App {
	var opts []grpc.ServerOption
	var authInterceptor *grpcInt.AuthInterceptor

	if ssoClient != nil {
		authInterceptor = grpcInt.NewAuthInterceptor(ssoClient, rateLimiter)

		authInterceptor.SetPublicMethods(
			"/management.Management/ListPlans",
			"/management.Management/GetPlan",
		)

		opts = append(opts,
			grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor()),
			grpc.StreamInterceptor(authInterceptor.StreamInterceptor()),
		)

		slog.Info("auth interceptor enabled with rate limiting")
	} else {
		slog.Warn("auth interceptor disabled - SSO client not available")
	}

	gRPCServer := grpc.NewServer(opts...)

	// Включаем серверную рефлексию (полезно для отладки)
	reflection.Register(gRPCServer)

	// Регистрируем сервисы
	serverAPI := grpcInt.NewServerAPI(planSvc)
	managementv1.RegisterManagementServer(gRPCServer, serverAPI)

	slog.Info("management service registered")

	return &App{
		port:            cfg.GRPC.Port,
		gRPCServer:      gRPCServer,
		authInterceptor: authInterceptor,
	}
}

func (a *App) MustRun() {
	if err := a.run(); err != nil {
		panic(err)
	}
}

func (a *App) run() error {
	const op = "grpcapp.Run"

	slog.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	slog.Info("gRPC server is running", slog.String("addr", l.Addr().String()))

	if err = a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	slog.With(
		slog.String("op", op),
	).Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
