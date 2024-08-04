package geyser_pkg

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// AuthenticationService handles the authentication logic by managing the gRPC context and access token.
type AuthenticationService struct {
	GRPCCtx     context.Context // Context for gRPC operations with metadata
	AccessToken string          // Access token for authentication
}

// NewAuthenticationService creates a new instance of AuthenticationService.
// It takes a context and an access token as input and returns an AuthenticationService with the gRPC context updated with the access token.
func NewAuthenticationService(
	ctx context.Context,
	accessToken string,
) *AuthenticationService {
	return &AuthenticationService{
		GRPCCtx:     metadata.NewOutgoingContext(ctx, metadata.Pairs("access-token", accessToken)),
		AccessToken: accessToken,
	}
}
