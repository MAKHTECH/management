package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/makhtech/management/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	gRPCServer *grpc.Server
	port       int
}

// Добавление несколько interceptors
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

func New(cfg *config.Config) *App {

	gRPCServer := grpc.NewServer()

	// Включаем серверную рефлексию (полезно для отладки)
	reflection.Register(gRPCServer)

	// todo зарегать сервисы
	//grpc_auth.Register(gRPCServer, auth)
	//gprc_user.Register(gRPCServer, user)

	return &App{
		port:       cfg.GRPC.Port,
		gRPCServer: gRPCServer,
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
