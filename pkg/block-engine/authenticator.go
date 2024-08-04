package block_engine_pkg

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	jito_pb "github.com/Prophet-Solutions/block-engine-protos/auth"
	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthenticationService handles the authentication logic for interacting with the gRPC services.
type AuthenticationService struct {
	AuthService jito_pb.AuthServiceClient // Client for authentication service
	GRPCCtx     context.Context           // Context for gRPC operations
	KeyPair     *solana.PrivateKey        // Private key for signing challenges
	BearerToken string                    // Bearer token for authorization
	ExpiresAt   int64                     // Expiration time for the token
	ErrChan     chan error                // Channel for error handling
	mu          sync.Mutex                // Mutex for synchronizing token updates
}

// NewAuthenticationService creates a new instance of AuthenticationService.
func NewAuthenticationService(
	ctx context.Context,
	grpcConn *grpc.ClientConn,
	keyPair *solana.PrivateKey,
) *AuthenticationService {
	return &AuthenticationService{
		AuthService: jito_pb.NewAuthServiceClient(grpcConn),
		GRPCCtx:     ctx,
		KeyPair:     keyPair,
		ErrChan:     make(chan error, 1),
		mu:          sync.Mutex{},
	}
}

// AuthenticateAndRefresh handles the authentication and token refresh logic.
func (as *AuthenticationService) AuthenticateAndRefresh(role jito_pb.Role) error {
	// Generate authentication challenge
	respChallenge, err := as.AuthService.GenerateAuthChallenge(as.GRPCCtx,
		&jito_pb.GenerateAuthChallengeRequest{
			Role:   role,
			Pubkey: as.KeyPair.PublicKey().Bytes(),
		},
	)
	if err != nil {
		return err
	}

	challenge := fmt.Sprintf("%s-%s", as.KeyPair.PublicKey().String(), respChallenge.GetChallenge())

	// Generate signature for the challenge
	sig, err := as.generateChallengeSignature([]byte(challenge))
	if err != nil {
		return err
	}

	// Generate authentication tokens
	respToken, err := as.AuthService.GenerateAuthTokens(as.GRPCCtx, &jito_pb.GenerateAuthTokensRequest{
		Challenge:       challenge,
		SignedChallenge: sig,
		ClientPubkey:    as.KeyPair.PublicKey().Bytes(),
	})
	if err != nil {
		return err
	}

	// Update authorization metadata with the new token
	as.updateAuthorizationMetadata(respToken.AccessToken)

	// Goroutine to continuously refresh the access token before it expires
	go func() {
		for {
			select {
			case <-as.GRPCCtx.Done():
				as.ErrChan <- as.GRPCCtx.Err()
				return
			default:
				resp, err := as.AuthService.RefreshAccessToken(as.GRPCCtx, &jito_pb.RefreshAccessTokenRequest{
					RefreshToken: respToken.RefreshToken.Value,
				})
				if err != nil {
					as.ErrChan <- fmt.Errorf("failed to refresh access token: %w", err)
					time.Sleep(1 * time.Minute) // Retry after 1 minute on failure
					continue
				}

				as.updateAuthorizationMetadata(resp.AccessToken)
				sleepDuration := time.Until(resp.AccessToken.ExpiresAtUtc.AsTime()) - 15*time.Second
				if sleepDuration > 0 {
					time.Sleep(sleepDuration)
				}
			}
		}
	}()

	// Wait for the bearer token to be set
	for {
		if as.BearerToken != "" {
			break
		}
		log.Println("Waiting for challenge to solve.")
		time.Sleep(1 * time.Second)
	}

	return nil
}

// updateAuthorizationMetadata updates the gRPC context with the new authorization token.
func (as *AuthenticationService) updateAuthorizationMetadata(token *jito_pb.Token) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.GRPCCtx = metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token.Value))
	as.BearerToken = token.Value
	as.ExpiresAt = token.ExpiresAtUtc.Seconds
}

// generateChallengeSignature generates a signature for the given challenge using the private key.
func (as *AuthenticationService) generateChallengeSignature(challenge []byte) ([]byte, error) {
	sig, err := as.KeyPair.Sign(challenge)
	if err != nil {
		return nil, err
	}

	return base58.Decode(sig.String())
}
