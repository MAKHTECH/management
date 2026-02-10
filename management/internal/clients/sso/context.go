package sso

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	// AuthorizationHeader ключ для access token в metadata
	AuthorizationHeader = "authorization"
	// BearerPrefix префикс для Bearer токена
	BearerPrefix = "Bearer "
)

// contextWithAccessToken добавляет access token в контекст как gRPC metadata
func contextWithAccessToken(ctx context.Context, accessToken string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, AuthorizationHeader, BearerPrefix+accessToken)
}
