package block_engine

import (
	"fmt"

	block_engine_pb "github.com/Prophet-Solutions/block-engine-protos/block_engine"
	searcher_pb "github.com/Prophet-Solutions/block-engine-protos/searcher"
	pkg "github.com/Prophet-Solutions/jito-go/pkg/block-engine"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"google.golang.org/grpc"
)

// SearcherClient is a client for interacting with the Searcher service.
type SearcherClient struct {
	GRPCConn                 *grpc.ClientConn                                         // gRPC connection
	RPCConn                  *rpc.Client                                              // Standard RPC connection
	JitoRPCConn              *rpc.Client                                              // Jito RPC connection
	SearcherService          searcher_pb.SearcherServiceClient                        // Searcher service client
	BundleStreamSubscription searcher_pb.SearcherService_SubscribeBundleResultsClient // Bundle stream subscription
	AuthenticationService    *pkg.AuthenticationService                               // Authentication service
	ErrChan                  <-chan error                                             // Error channel
}

// Relayer is a client for interacting with the Block Engine Relayer service.
type Relayer struct {
	GRPCConn              *grpc.ClientConn                         // gRPC connection
	Client                block_engine_pb.BlockEngineRelayerClient // Relayer client
	AuthenticationService *pkg.AuthenticationService               // Authentication service
	ErrChan               <-chan error                             // Error channel
}

// Validator is a client for interacting with the Block Engine Validator service.
type Validator struct {
	GRPCConn              *grpc.ClientConn                           // gRPC connection
	Client                block_engine_pb.BlockEngineValidatorClient // Validator client
	AuthenticationService *pkg.AuthenticationService                 // Authentication service
	ErrChan               <-chan error                               // Error channel
}

// BundleResponse represents a response from sending a bundle.
type BundleResponse struct {
	BundleResponse *searcher_pb.SendBundleResponse // Response from sending bundle
	Signatures     []solana.Signature              // Signatures of transactions in the bundle
}

// BundleRejectionError represents an error when a bundle is rejected.
type BundleRejectionError struct {
	Message string // Error message
}

// Error implements the error interface for BundleRejectionError.
func (e BundleRejectionError) Error() string {
	return e.Message
}

// NewStateAuctionBidRejectedError creates a new error indicating that a bundle lost the state auction.
func NewStateAuctionBidRejectedError(auction string, tip uint64) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle lost state auction, auction: %s, tip %d lamports", auction, tip),
	}
}

// NewWinningBatchBidRejectedError creates a new error indicating that a bundle won the state auction but failed the global auction.
func NewWinningBatchBidRejectedError(auction string, tip uint64) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle won state auction but failed global auction, auction %s, tip %d lamports", auction, tip),
	}
}

// NewSimulationFailureError creates a new error indicating that a bundle failed simulation.
func NewSimulationFailureError(tx string, message string) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle simulation failure on tx %s, message: %s", tx, message),
	}
}

// NewInternalError creates a new internal error.
func NewInternalError(message string) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("internal error %s", message),
	}
}

// NewDroppedBundle creates a new error indicating that a bundle was dropped.
func NewDroppedBundle(message string) error {
	return BundleRejectionError{
		Message: fmt.Sprintf("bundle dropped %s", message),
	}
}
