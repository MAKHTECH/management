package grpc

import (
	"context"
	"log/slog"
	"strings"

	"github.com/makhtech/management/internal/clients/sso"
	"github.com/makhtech/management/pkg/ratelimiter"
	ssov1 "github.com/makhtech/proto/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Ключи для хранения информации о пользователе в context
type contextKey string

const (
	// UserContextKey ключ для хранения информации о пользователе в context
	UserContextKey contextKey = "user"
	// AccessTokenContextKey ключ для хранения access token в context
	AccessTokenContextKey contextKey = "access_token"
)

// UserInfo информация о пользователе, извлечённая из JWT
type UserInfo struct {
	UserID   int64
	Username string
	Email    string
	PhotoURL string
	Role     ssov1.Role
	AppID    int32
	Balance  int64
}

// AuthInterceptor interceptor для аутентификации и авторизации
type AuthInterceptor struct {
	ssoClient   *sso.Client
	rateLimiter *ratelimiter.TokenBucket
	// Методы, которые не требуют аутентификации
	publicMethods map[string]bool
}

// NewAuthInterceptor создаёт новый AuthInterceptor
func NewAuthInterceptor(ssoClient *sso.Client, rateLimiter *ratelimiter.TokenBucket) *AuthInterceptor {
	return &AuthInterceptor{
		ssoClient:     ssoClient,
		rateLimiter:   rateLimiter,
		publicMethods: make(map[string]bool),
	}
}

// SetPublicMethods устанавливает методы, которые не требуют аутентификации
func (i *AuthInterceptor) SetPublicMethods(methods ...string) {
	for _, method := range methods {
		i.publicMethods[method] = true
	}
}

// UnaryInterceptor возвращает gRPC UnaryServerInterceptor
func (i *AuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Проверяем, является ли метод публичным
		if i.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Извлекаем access token из metadata
		accessToken, err := extractAccessToken(ctx)
		if err != nil {
			slog.Warn("failed to extract access token",
				slog.String("method", info.FullMethod),
				slog.String("error", err.Error()),
			)
			return nil, err
		}

		// Применяем rate limiting
		if i.rateLimiter != nil {
			allowed, remaining := i.rateLimiter.Allow(accessToken)
			if !allowed {
				slog.Warn("rate limit exceeded",
					slog.String("method", info.FullMethod),
				)
				return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded, please try again later")
			}
			slog.Debug("rate limit check passed",
				slog.String("method", info.FullMethod),
				slog.Int("remaining", remaining),
			)
		}

		// Валидируем JWT через SSO сервис
		userResp, err := i.ssoClient.ValidateJWT(ctx, accessToken)
		if err != nil {
			slog.Warn("JWT validation failed",
				slog.String("method", info.FullMethod),
				slog.String("error", err.Error()),
			)
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Создаём UserInfo из ответа SSO
		userInfo := &UserInfo{
			UserID:   userResp.UserId,
			Username: userResp.Username,
			Email:    userResp.Email,
			PhotoURL: userResp.PhotoUrl,
			Role:     userResp.Role,
			AppID:    userResp.AppId,
			Balance:  userResp.Balance,
		}

		// Добавляем информацию о пользователе в context
		ctx = context.WithValue(ctx, UserContextKey, userInfo)
		ctx = context.WithValue(ctx, AccessTokenContextKey, accessToken)

		slog.Debug("user authenticated",
			slog.String("method", info.FullMethod),
			slog.Int64("user_id", userInfo.UserID),
			slog.String("username", userInfo.Username),
		)

		return handler(ctx, req)
	}
}

// StreamInterceptor возвращает gRPC StreamServerInterceptor
func (i *AuthInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Проверяем, является ли метод публичным
		if i.publicMethods[info.FullMethod] {
			return handler(srv, stream)
		}

		ctx := stream.Context()

		// Извлекаем access token из metadata
		accessToken, err := extractAccessToken(ctx)
		if err != nil {
			return err
		}

		// Применяем rate limiting
		if i.rateLimiter != nil {
			allowed, _ := i.rateLimiter.Allow(accessToken)
			if !allowed {
				return status.Error(codes.ResourceExhausted, "rate limit exceeded, please try again later")
			}
		}

		// Валидируем JWT через SSO сервис
		userResp, err := i.ssoClient.ValidateJWT(ctx, accessToken)
		if err != nil {
			return status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		// Создаём UserInfo из ответа SSO
		userInfo := &UserInfo{
			UserID:   userResp.UserId,
			Username: userResp.Username,
			Email:    userResp.Email,
			PhotoURL: userResp.PhotoUrl,
			Role:     userResp.Role,
			AppID:    userResp.AppId,
			Balance:  userResp.Balance,
		}

		// Оборачиваем stream с новым context
		wrappedStream := &wrappedServerStream{
			ServerStream: stream,
			ctx:          context.WithValue(context.WithValue(ctx, UserContextKey, userInfo), AccessTokenContextKey, accessToken),
		}

		return handler(srv, wrappedStream)
	}
}

// extractAccessToken извлекает access token из gRPC metadata
func extractAccessToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	authHeader := values[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format, expected 'Bearer <token>'")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

// wrappedServerStream обёртка над ServerStream для передачи изменённого context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// GetUserFromContext извлекает информацию о пользователе из context
func GetUserFromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(UserContextKey).(*UserInfo)
	return user, ok
}

// GetAccessTokenFromContext извлекает access token из context
func GetAccessTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(AccessTokenContextKey).(string)
	return token, ok
}

// MustGetUserFromContext извлекает информацию о пользователе из context
// Паникует, если пользователь не найден
func MustGetUserFromContext(ctx context.Context) *UserInfo {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		panic("user not found in context")
	}
	return user
}
