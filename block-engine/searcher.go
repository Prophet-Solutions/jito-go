package block_engine

import (
	"context"
	"fmt"
	"math/rand"

	auth_pb "github.com/Prophet-Solutions/block-engine-protos/auth"
	jito_pb "github.com/Prophet-Solutions/block-engine-protos/searcher"
	"github.com/Prophet-Solutions/jito-go/pkg"
	block_engine_pkg "github.com/Prophet-Solutions/jito-go/pkg/block-engine"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"google.golang.org/grpc"
)

// NewSearcherClient initializes a new SearcherClient with the provided context, gRPC address,
// Jito and standard RPC clients, Solana private key, and additional gRPC dial options.
// It establishes a gRPC connection, sets up authentication if a keyPair is provided, and returns the SearcherClient.
func NewSearcherClient(
	ctx context.Context,
	grpcAddr string,
	jitoRPCClient, rpcClient *rpc.Client,
	keyPair *solana.PrivateKey,
	opts ...grpc.DialOption,
) (*SearcherClient, error) {
	// Channel to handle errors during gRPC connection setup
	chErr := make(chan error)

	// Create a new gRPC connection using the provided address and options
	conn, err := pkg.CreateGRPCConnection(ctx, chErr, grpcAddr, opts...)
	if err != nil {
		return nil, err
	}

	// Create a new SearcherServiceClient with the established connection
	searcherService := jito_pb.NewSearcherServiceClient(conn)
	var authService *block_engine_pkg.AuthenticationService

	// Set up authentication if a keyPair is provided
	if keyPair != nil {
		authService = block_engine_pkg.NewAuthenticationService(context.Background(), conn, keyPair)
		if err = authService.AuthenticateAndRefresh(auth_pb.Role_SEARCHER); err != nil {
			return nil, err
		}
	} else {
		authService = &block_engine_pkg.AuthenticationService{
			GRPCCtx: ctx,
		}
	}

	// Subscribe to bundle results if authentication is set up
	var subBundleRes jito_pb.SearcherService_SubscribeBundleResultsClient
	subBundleRes, err = searcherService.SubscribeBundleResults(
		authService.GRPCCtx,
		&jito_pb.SubscribeBundleResultsRequest{},
	)
	if err != nil {
		return nil, fmt.Errorf("could not perform bundle results subscription: %w", err)
	}

	// Return the initialized SearcherClient
	return &SearcherClient{
		GRPCConn:                 conn,
		RPCConn:                  rpcClient,
		JitoRPCConn:              jitoRPCClient,
		SearcherService:          searcherService,
		AuthenticationService:    authService,
		BundleStreamSubscription: subBundleRes,
		ErrChan:                  chErr,
	}, nil
}

// GetRegions retrieves the regions from the Searcher service.
// It returns a GetRegionsResponse or an error.
func (c *SearcherClient) GetRegions(opts ...grpc.CallOption) (*jito_pb.GetRegionsResponse, error) {
	return c.SearcherService.GetRegions(
		c.AuthenticationService.GRPCCtx,
		&jito_pb.GetRegionsRequest{},
		opts...,
	)
}

// GetConnectedLeaders retrieves the connected leaders from the Searcher service.
// It returns a ConnectedLeadersResponse or an error.
func (c *SearcherClient) GetConnectedLeaders(opts ...grpc.CallOption) (*jito_pb.ConnectedLeadersResponse, error) {
	return c.SearcherService.GetConnectedLeaders(
		c.AuthenticationService.GRPCCtx,
		&jito_pb.ConnectedLeadersRequest{},
		opts...,
	)
}

// GetNextScheduledLeader retrieves the next scheduled leader for the specified regions from the Searcher service.
// It returns a NextScheduledLeaderResponse or an error.
func (c *SearcherClient) GetNextScheduledLeader(regions []string, opts ...grpc.CallOption) (*jito_pb.NextScheduledLeaderResponse, error) {
	return c.SearcherService.GetNextScheduledLeader(
		c.AuthenticationService.GRPCCtx,
		&jito_pb.NextScheduledLeaderRequest{
			Regions: regions,
		},
		opts...,
	)
}

// GetConnectedLeadersRegioned retrieves the connected leaders for specified regions from the Searcher service.
// It returns a ConnectedLeadersRegionedResponse or an error.
func (c *SearcherClient) GetConnectedLeadersRegioned(regions []string, opts ...grpc.CallOption) (*jito_pb.ConnectedLeadersRegionedResponse, error) {
	return c.SearcherService.GetConnectedLeadersRegioned(
		c.AuthenticationService.GRPCCtx,
		&jito_pb.ConnectedLeadersRegionedRequest{
			Regions: regions,
		},
		opts...,
	)
}

// GetTipAccounts retrieves the tip accounts from the Searcher service.
// It returns a GetTipAccountsResponse or an error.
func (c *SearcherClient) GetTipAccounts(opts ...grpc.CallOption) (*jito_pb.GetTipAccountsResponse, error) {
	return c.SearcherService.GetTipAccounts(
		c.AuthenticationService.GRPCCtx,
		&jito_pb.GetTipAccountsRequest{},
		opts...,
	)
}

// GetRandomTipAccount retrieves a random tip account from the list of tip accounts.
// It returns the selected account or an error.
func (c *SearcherClient) GetRandomTipAccount(opts ...grpc.CallOption) (string, error) {
	resp, err := c.GetTipAccounts(opts...)
	if err != nil {
		return "", err
	}

	// Return a randomly selected account from the list of tip accounts
	return resp.Accounts[rand.Intn(len(resp.Accounts))], nil
}
