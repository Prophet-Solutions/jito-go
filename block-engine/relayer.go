package block_engine

import (
	"context"

	auth_pb "github.com/Prophet-Solutions/block-engine-protos/auth"
	jito_pb "github.com/Prophet-Solutions/block-engine-protos/block_engine"
	"github.com/Prophet-Solutions/jito-go/pkg"
	block_engine_pkg "github.com/Prophet-Solutions/jito-go/pkg/block-engine"
	"github.com/gagliardetto/solana-go"
	"google.golang.org/grpc"
)

// NewRelayer initializes a new Relayer with the given context, gRPC address, and Solana private key.
// It establishes a gRPC connection, sets up authentication if a keyPair is provided, and returns the Relayer.
func NewRelayer(
	ctx context.Context,
	grpcAddr string,
	keyPair *solana.PrivateKey,
	opts ...grpc.DialOption,
) (*Relayer, error) {
	// Channel to handle errors during gRPC connection setup
	chErr := make(chan error)

	// Create a new gRPC connection using the provided address and options
	conn, err := pkg.CreateGRPCConnection(ctx, chErr, grpcAddr, opts...)
	if err != nil {
		return nil, err
	}

	// Create a new BlockEngineRelayerClient with the established connection
	blockEngineRelayerClient := jito_pb.NewBlockEngineRelayerClient(conn)
	var authService *block_engine_pkg.AuthenticationService

	// Set up authentication if a keyPair is provided
	if keyPair != nil {
		authService = block_engine_pkg.NewAuthenticationService(context.Background(), conn, keyPair)
		if err = authService.AuthenticateAndRefresh(auth_pb.Role_RELAYER); err != nil {
			return nil, err
		}
	}

	// Return the initialized Relayer
	return &Relayer{
		GRPCConn:              conn,
		Client:                blockEngineRelayerClient,
		AuthenticationService: authService,
		ErrChan:               chErr,
	}, nil
}

// SubscribeAccountsOfInterest subscribes to accounts of interest updates from the BlockEngineRelayer service.
// It returns a client for receiving these updates.
func (r *Relayer) SubscribeAccountsOfInterest(opts ...grpc.CallOption) (
	jito_pb.BlockEngineRelayer_SubscribeAccountsOfInterestClient, error) {
	return r.Client.SubscribeAccountsOfInterest(
		r.AuthenticationService.GRPCCtx,
		&jito_pb.AccountsOfInterestRequest{},
		opts...,
	)
}

// OnSubscribeAccountsOfInterest subscribes to accounts of interest updates and handles the incoming updates and errors.
// It returns channels for the updates and errors.
func (r *Relayer) OnSubscribeAccountsOfInterest(ctx context.Context) (
	<-chan *jito_pb.AccountsOfInterestUpdate, <-chan error, error) {
	// Subscribe to accounts of interest
	sub, err := r.SubscribeAccountsOfInterest()
	if err != nil {
		return nil, nil, err
	}

	// Channels to handle accounts of interest updates and errors
	chAccountOfInterest := make(chan *jito_pb.AccountsOfInterestUpdate)
	chErr := make(chan error)

	// Goroutine to receive updates and errors
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-r.AuthenticationService.GRPCCtx.Done():
				return
			default:
				resp, err := sub.Recv()
				if err != nil {
					chErr <- err
					continue
				}
				chAccountOfInterest <- resp
			}
		}
	}()

	return chAccountOfInterest, chErr, nil
}

// SubscribeProgramsOfInterest subscribes to programs of interest updates from the BlockEngineRelayer service.
// It returns a client for receiving these updates.
func (r *Relayer) SubscribeProgramsOfInterest(opts ...grpc.CallOption) (
	jito_pb.BlockEngineRelayer_SubscribeProgramsOfInterestClient, error) {
	return r.Client.SubscribeProgramsOfInterest(
		r.AuthenticationService.GRPCCtx,
		&jito_pb.ProgramsOfInterestRequest{},
		opts...,
	)
}

// OnSubscribeProgramsOfInterest subscribes to programs of interest updates and handles the incoming updates and errors.
// It returns channels for the updates and errors.
func (r *Relayer) OnSubscribeProgramsOfInterest(ctx context.Context) (
	<-chan *jito_pb.ProgramsOfInterestUpdate, <-chan error, error) {
	// Subscribe to programs of interest
	sub, err := r.SubscribeProgramsOfInterest()
	if err != nil {
		return nil, nil, err
	}

	// Channels to handle programs of interest updates and errors
	chProgramsOfInterest := make(chan *jito_pb.ProgramsOfInterestUpdate)
	chErr := make(chan error)

	// Goroutine to receive updates and errors
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				subInfo, err := sub.Recv()
				if err != nil {
					chErr <- err
					continue
				}
				chProgramsOfInterest <- subInfo
			}
		}
	}()

	return chProgramsOfInterest, chErr, nil
}

// StartExpiringPacketStream starts a stream for receiving expiring packet updates from the BlockEngineRelayer service.
// It returns a client for receiving these updates.
func (r *Relayer) StartExpiringPacketStream(opts ...grpc.CallOption) (
	jito_pb.BlockEngineRelayer_StartExpiringPacketStreamClient, error) {
	return r.Client.StartExpiringPacketStream(r.AuthenticationService.GRPCCtx, opts...)
}

// OnStartExpiringPacketStream starts a stream for receiving expiring packet updates and handles the incoming updates and errors.
// It returns channels for the updates and errors.
func (r *Relayer) OnStartExpiringPacketStream(ctx context.Context) (
	<-chan *jito_pb.StartExpiringPacketStreamResponse, <-chan error, error) {
	// Start the expiring packet stream
	sub, err := r.StartExpiringPacketStream()
	if err != nil {
		return nil, nil, err
	}

	// Channels to handle expiring packet updates and errors
	chPacket := make(chan *jito_pb.StartExpiringPacketStreamResponse)
	chErr := make(chan error)

	// Goroutine to receive updates and errors
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-r.AuthenticationService.GRPCCtx.Done():
				return
			default:
				resp, err := sub.Recv()
				if err != nil {
					chErr <- err
					continue
				}
				chPacket <- resp
			}
		}
	}()

	return chPacket, chErr, nil
}
