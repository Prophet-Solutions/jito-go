package block_engine

import (
	"context"

	auth_pb "github.com/Prophet-Solutions/block-engine-protos/auth"
	jito_pb "github.com/Prophet-Solutions/block-engine-protos/block_engine"
	bundle_pb "github.com/Prophet-Solutions/block-engine-protos/bundle"
	"github.com/Prophet-Solutions/jito-go/pkg"
	block_engine_pkg "github.com/Prophet-Solutions/jito-go/pkg/block-engine"
	"github.com/gagliardetto/solana-go"
	"google.golang.org/grpc"
)

// NewValidator initializes a new Validator with the provided context, gRPC address, and Solana private key.
// It establishes a gRPC connection, sets up authentication if a keyPair is provided, and returns the Validator.
func NewValidator(
	ctx context.Context,
	grpcAddr string,
	keyPair *solana.PrivateKey,
	opts ...grpc.DialOption,
) (*Validator, error) {
	// Channel to handle errors during gRPC connection setup
	chErr := make(chan error)

	// Create a new gRPC connection using the provided address and options
	conn, err := pkg.CreateGRPCConnection(ctx, chErr, grpcAddr, opts...)
	if err != nil {
		return nil, err
	}

	// Create a new BlockEngineValidatorClient with the established connection
	blockEngineValidatorClient := jito_pb.NewBlockEngineValidatorClient(conn)
	var authService *block_engine_pkg.AuthenticationService

	// Set up authentication if a keyPair is provided
	if keyPair != nil {
		authService = block_engine_pkg.NewAuthenticationService(context.Background(), conn, keyPair)
		if err = authService.AuthenticateAndRefresh(auth_pb.Role_VALIDATOR); err != nil {
			return nil, err
		}
	}

	// Return the initialized Validator
	return &Validator{
		GRPCConn:              conn,
		Client:                blockEngineValidatorClient,
		AuthenticationService: authService,
		ErrChan:               chErr,
	}, nil
}

// SubscribePackets subscribes to packet updates from the BlockEngineValidator service.
// It returns a client for receiving these updates.
func (v *Validator) SubscribePackets(
	opts ...grpc.CallOption,
) (jito_pb.BlockEngineValidator_SubscribePacketsClient, error) {
	return v.Client.SubscribePackets(
		v.AuthenticationService.GRPCCtx,
		&jito_pb.SubscribePacketsRequest{},
		opts...,
	)
}

// OnPacketSubscription subscribes to packet updates and handles the incoming updates and errors.
// It returns channels for the updates and errors.
func (v *Validator) OnPacketSubscription(
	ctx context.Context,
) (<-chan *jito_pb.SubscribePacketsResponse, <-chan error, error) {
	// Subscribe to packets
	sub, err := v.SubscribePackets()
	if err != nil {
		return nil, nil, err
	}

	// Channels to handle packet updates and errors
	chPackets := make(chan *jito_pb.SubscribePacketsResponse)
	chErr := make(chan error)

	// Goroutine to receive updates and errors
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				resp, err := sub.Recv()
				if err != nil {
					chErr <- err
					continue
				}

				chPackets <- resp
			}
		}
	}()

	return chPackets, chErr, nil
}

// SubscribeBundles subscribes to bundle updates from the BlockEngineValidator service.
// It returns a client for receiving these updates.
func (v *Validator) SubscribeBundles(
	opts ...grpc.CallOption,
) (jito_pb.BlockEngineValidator_SubscribeBundlesClient, error) {
	return v.Client.SubscribeBundles(
		v.AuthenticationService.GRPCCtx,
		&jito_pb.SubscribeBundlesRequest{},
		opts...,
	)
}

// OnBundleSubscription subscribes to bundle updates and handles the incoming updates and errors.
// It returns channels for the updates and errors.
func (v *Validator) OnBundleSubscription(ctx context.Context) (
	<-chan []*bundle_pb.BundleUuid, <-chan error, error) {
	// Subscribe to bundles
	sub, err := v.SubscribeBundles()
	if err != nil {
		return nil, nil, err
	}

	// Channels to handle bundle updates and errors
	chBundleUuid := make(chan []*bundle_pb.BundleUuid)
	chErr := make(chan error)

	// Goroutine to receive updates and errors
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-v.AuthenticationService.GRPCCtx.Done():
				return
			default:
				resp, err := sub.Recv()
				if err != nil {
					chErr <- err
					continue
				}

				chBundleUuid <- resp.Bundles
			}
		}
	}()

	return chBundleUuid, chErr, nil
}

// GetBlockBuilderFeeInfo retrieves the block builder fee information from the BlockEngineValidator service.
// It returns a BlockBuilderFeeInfoResponse or an error.
func (v *Validator) GetBlockBuilderFeeInfo(
	opts ...grpc.CallOption,
) (*jito_pb.BlockBuilderFeeInfoResponse, error) {
	return v.Client.GetBlockBuilderFeeInfo(
		v.AuthenticationService.GRPCCtx,
		&jito_pb.BlockBuilderFeeInfoRequest{},
		opts...,
	)
}
