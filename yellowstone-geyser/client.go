package yellowstone_geyser

import (
	"context"
	"encoding/json"

	"github.com/Prophet-Solutions/jito-go/pkg"
	pb "github.com/Prophet-Solutions/yellowstone-geyser-protos/geyser"
	"github.com/gagliardetto/solana-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewClient creates and returns a new YellowstoneGeyserClient.
// It establishes a gRPC connection to the Yellowstone Geyser service using the provided address and options.
//
// Parameters:
// - ctx: The context for managing the connection lifecycle.
// - grpcAddr: The address of the gRPC server.
// - opts: Optional gRPC dial options for customizing the connection.
//
// Returns:
// - A pointer to YellowstoneGeyserClient if successful, or an error if the connection fails.
func NewClient(
	ctx context.Context,
	grpcAddr string,
	opts ...grpc.DialOption,
) (*YellowstoneGeyserClient, error) {
	// Channel to receive errors during the connection establishment.
	chErr := make(chan error)

	// Establish the gRPC connection using the provided context, address, and options.
	conn, err := pkg.CreateGRPCConnection(ctx, chErr, grpcAddr, opts...)
	if err != nil {
		return nil, err
	}

	// Return the YellowstoneGeyserClient with the established connection.
	return &YellowstoneGeyserClient{
		GRPCConn: conn,
		Client:   pb.NewGeyserClient(conn),
		Stream:   nil,
	}, nil
}

// GetBlockHeight retrieves the current block height from the Geyser service.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - commitment: The commitment level to use for the request.
//
// Returns:
// - The response containing the block height, or an error if the request fails.
func (gc *YellowstoneGeyserClient) GetBlockHeight(
	ctx context.Context,
	commitment *pb.CommitmentLevel,
) (*pb.GetBlockHeightResponse, error) {
	return gc.Client.GetBlockHeight(ctx, &pb.GetBlockHeightRequest{
		Commitment: commitment.Enum(),
	})
}

// GetLatestBlockhash retrieves the latest blockhash from the Geyser service.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - commitment: The commitment level to use for the request.
//
// Returns:
// - The response containing the latest blockhash, or an error if the request fails.
func (gc *YellowstoneGeyserClient) GetLatestBlockhash(
	ctx context.Context,
	commitment *pb.CommitmentLevel,
) (*pb.GetLatestBlockhashResponse, error) {
	return gc.Client.GetLatestBlockhash(ctx, &pb.GetLatestBlockhashRequest{
		Commitment: commitment.Enum(),
	})
}

// GetSlot retrieves the current slot number from the Geyser service.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - commitment: The commitment level to use for the request.
//
// Returns:
// - The response containing the current slot number, or an error if the request fails.
func (gc *YellowstoneGeyserClient) GetSlot(
	ctx context.Context,
	commitment *pb.CommitmentLevel,
) (*pb.GetSlotResponse, error) {
	return gc.Client.GetSlot(ctx, &pb.GetSlotRequest{
		Commitment: commitment.Enum(),
	})
}

// GetVersion retrieves the current version of the Geyser service.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
//
// Returns:
// - The response containing the version information, or an error if the request fails.
func (gc *YellowstoneGeyserClient) GetVersion(
	ctx context.Context,
) (*pb.GetVersionResponse, error) {
	return gc.Client.GetVersion(ctx, &pb.GetVersionRequest{})
}

// IsBlockhashValid checks whether a given blockhash is valid according to the Geyser service.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - blockhash: The blockhash to validate.
// - commitment: The commitment level to use for the request.
//
// Returns:
// - The response indicating the validity of the blockhash, or an error if the request fails.
func (gc *YellowstoneGeyserClient) IsBlockhashValid(
	ctx context.Context,
	blockhash string,
	commitment *pb.CommitmentLevel,
) (*pb.IsBlockhashValidResponse, error) {
	return gc.Client.IsBlockhashValid(ctx, &pb.IsBlockhashValidRequest{
		Blockhash:  blockhash,
		Commitment: commitment.Enum(),
	})
}

// Ping sends a ping request to the Geyser service.
//
// Parameters:
// - ctx: The context for managing the request lifecycle.
// - count: The number of ping requests to send.
//
// Returns:
// - The response containing the pong result, or an error if the request fails.
func (gc *YellowstoneGeyserClient) Ping(
	ctx context.Context,
	count int32,
) (*pb.PongResponse, error) {
	return gc.Client.Ping(ctx, &pb.PingRequest{Count: count})
}

// Subscribe sets up a subscription with the Geyser service to receive real-time updates.
//
// Parameters:
// - ctx: The context for managing the subscription lifecycle.
// - token: Optional authentication token for the subscription.
// - jsonInput: Optional JSON input for customizing the subscription request.
// - slots: If true, subscribe to slot updates.
// - blocks: If true, subscribe to block updates.
// - blocksMeta: If true, subscribe to block metadata updates.
// - signature: Optional signature to subscribe to specific transaction updates.
// - accounts: If true, subscribe to account updates.
// - transactions: If true, subscribe to transaction updates.
// - voteTransactions: If true, include vote transactions in the subscription.
// - failedTransactions: If true, include failed transactions in the subscription.
// - accountsFilter: List of public keys to filter account updates by account addresses.
// - accountOwnersFilter: List of public keys to filter account updates by account owners.
// - transactionsAccountsInclude: List of public keys to include in transaction updates.
// - transactionsAccountsExclude: List of public keys to exclude from transaction updates.
//
// Returns:
// - A gRPC client stream for receiving subscription updates, or an error if the subscription fails.
func (gc *YellowstoneGeyserClient) Subscribe(
	ctx context.Context,
	token *string,
	jsonInput *string,
	slots bool,
	blocks bool,
	blocksMeta bool,
	signature *solana.Signature,
	accounts bool,
	transactions bool,
	voteTransactions bool,
	failedTransactions bool,
	accountsFilter []solana.PublicKey,
	accountOwnersFilter []solana.PublicKey,
	transactionsAccountsInclude []solana.PublicKey,
	transactionsAccountsExclude []solana.PublicKey,
) error {
	// Create an empty subscription request.
	var subscription pb.SubscribeRequest

	// Convert public keys to string format for filtering.
	stringAccountsFilter := pkg.ConvertBatchPublicKeyToString(accountsFilter)
	stringAccountOwnersFilter := pkg.ConvertBatchPublicKeyToString(accountOwnersFilter)
	stringTransactionsAccountsInclude := pkg.ConvertBatchPublicKeyToString(transactionsAccountsInclude)
	stringTransactionsAccountsExclude := pkg.ConvertBatchPublicKeyToString(transactionsAccountsExclude)

	// If JSON input is provided, unmarshal it into the subscription request.
	if jsonInput != nil && *jsonInput != "" {
		jsonData := []byte(*jsonInput)
		err := json.Unmarshal(jsonData, &subscription)
		if err != nil {
			return err
		}
	} else {
		// If no JSON is provided, start with an empty subscription.
		subscription = pb.SubscribeRequest{}
	}

	// Configure slot subscription if requested.
	if slots {
		if subscription.Slots == nil {
			subscription.Slots = make(map[string]*pb.SubscribeRequestFilterSlots)
		}
		subscription.Slots["slots"] = &pb.SubscribeRequestFilterSlots{}
	}

	// Configure block subscription if requested.
	if blocks {
		if subscription.Blocks == nil {
			subscription.Blocks = make(map[string]*pb.SubscribeRequestFilterBlocks)
		}
		subscription.Blocks["blocks"] = &pb.SubscribeRequestFilterBlocks{}
	}

	// Configure block metadata subscription if requested.
	if blocksMeta {
		if subscription.BlocksMeta == nil {
			subscription.BlocksMeta = make(map[string]*pb.SubscribeRequestFilterBlocksMeta)
		}
		subscription.BlocksMeta["block_meta"] = &pb.SubscribeRequestFilterBlocksMeta{}
	}

	// Configure account subscription if filters are provided or accounts subscription is requested.
	if (len(accountsFilter)+len(accountOwnersFilter)) > 0 || accounts {
		if subscription.Accounts == nil {
			subscription.Accounts = make(map[string]*pb.SubscribeRequestFilterAccounts)
		}
		subscription.Accounts["account_sub"] = &pb.SubscribeRequestFilterAccounts{}

		// Apply account address filters.
		if len(accountsFilter) > 0 {
			subscription.Accounts["account_sub"].Account = stringAccountsFilter
		}

		// Apply account owner filters.
		if len(accountOwnersFilter) > 0 {
			subscription.Accounts["account_sub"].Owner = stringAccountOwnersFilter
		}
	}

	// Configure transaction subscription.
	if subscription.Transactions == nil {
		subscription.Transactions = make(map[string]*pb.SubscribeRequestFilterTransactions)
	}

	// If a specific signature is provided, subscribe to the corresponding transaction.
	if signature != nil {
		tr := true
		subscription.Transactions["signature_sub"] = &pb.SubscribeRequestFilterTransactions{
			Failed: &tr,
			Vote:   &tr,
		}
		cSignature := signature.String()
		subscription.Transactions["signature_sub"].Signature = &cSignature
	}

	// Subscribe to generic transaction streams if requested.
	if transactions {
		subscription.Transactions["transactions_sub"] = &pb.SubscribeRequestFilterTransactions{
			Failed: &failedTransactions,
			Vote:   &voteTransactions,
		}

		// Apply inclusion and exclusion filters for transactions.
		subscription.Transactions["transactions_sub"].AccountInclude = stringTransactionsAccountsInclude
		subscription.Transactions["transactions_sub"].AccountExclude = stringTransactionsAccountsExclude
	}

	// Set up the subscription request context with the provided token if available.
	if token != nil && *token != "" {
		md := metadata.New(map[string]string{"x-token": *token})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Open a gRPC stream for the subscription.
	stream, err := gc.Client.Subscribe(ctx)
	if err != nil {
		return err
	}

	// Send the subscription request to the server.
	err = stream.Send(&subscription)
	if err != nil {
		return err
	}

	// Store the subscription request in the client.
	gc.Stream = &stream

	// Return the stream for receiving updates.
	return nil
}
