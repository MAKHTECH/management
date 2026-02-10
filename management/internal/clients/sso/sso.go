package sso

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ssov1 "github.com/makhtech/proto/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client обёртка над gRPC клиентами SSO-сервиса
type Client struct {
	authClient         ssov1.AuthClient
	userClient         ssov1.UserClient
	transactionsClient ssov1.TransactionsClient
	conn               *grpc.ClientConn
}

// Config конфигурация SSO клиента
type Config struct {
	Address      string        `json:"address"`
	Timeout      time.Duration `json:"timeout"`
	RetriesCount int           `json:"retries_count"`
	Insecure     bool          `json:"insecure"`
}

// New создаёт новый SSO клиент
func New(ctx context.Context, cfg Config) (*Client, error) {
	const op = "clients.sso.New"

	opts := []grpc.DialOption{}

	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Добавляем таймаут на подключение
	dialCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to SSO service: %w", op, err)
	}

	slog.Info("SSO client connected", slog.String("address", cfg.Address))

	return &Client{
		authClient:         ssov1.NewAuthClient(conn),
		userClient:         ssov1.NewUserClient(conn),
		transactionsClient: ssov1.NewTransactionsClient(conn),
		conn:               conn,
	}, nil
}

// Close закрывает соединение с SSO сервисом
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ValidateJWT валидирует JWT токен и возвращает информацию о пользователе
func (c *Client) ValidateJWT(ctx context.Context, accessToken string) (*ssov1.ValidateJWTResponse, error) {
	const op = "clients.sso.ValidateJWT"

	// Добавляем access token в metadata для передачи в SSO сервис
	ctx = contextWithAccessToken(ctx, accessToken)

	resp, err := c.userClient.ValidateJWT(ctx, &ssov1.ValidateJWTRequest{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resp, nil
}

// Auth возвращает Auth клиент для низкоуровневых операций
func (c *Client) Auth() ssov1.AuthClient {
	return c.authClient
}

// User возвращает User клиент для низкоуровневых операций
func (c *Client) User() ssov1.UserClient {
	return c.userClient
}

// Transactions возвращает Transactions клиент для низкоуровневых операций
func (c *Client) Transactions() ssov1.TransactionsClient {
	return c.transactionsClient
}
