package yellowstone_geyser

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/Prophet-Solutions/jito-go/pkg"
	pb "github.com/Prophet-Solutions/yellowstone-geyser-protos/geyser"
	"google.golang.org/grpc"
)

func NewClient(
	ctx context.Context,
	grpcAddr string,
	opts ...grpc.DialOption,
) (*GeyserClient, error) {
	// Establish the gRPC connection using the provided context, address, and options.
	conn, err := pkg.CreateGRPCConnection(ctx, nil, grpcAddr, opts...)
	if err != nil {
		return nil, err
	}

	geyserClient := pb.NewGeyserClient(conn)
	if geyserClient == nil {
		return nil, fmt.Errorf("failed to create Geyser Client")
	}

	subscribe, err := geyserClient.Subscribe(ctx, grpc.MaxCallRecvMsgSize(16*1024*1024), grpc.MaxCallSendMsgSize(16*1024*1024))
	if err != nil {
		return nil, err
	}

	return &GeyserClient{
		GRPCConn: conn,
		Ctx:      ctx,
		Client:   geyserClient,
		Streams:  sync.Map{},
		DefaultStreamClient: &StreamClient{
			SubscribeClient: subscribe,
			Ctx:             ctx,
			UpdateCh:        make(chan *pb.SubscribeUpdate), // Unbuffered channel to prevent overflowing
			ErrCh:           make(chan error, 10),
		},
		ErrCh: make(chan error, 10),
	}, nil
}

func (c *GeyserClient) NewSubscribeClient(ctx context.Context, clientName string) error {
	stream, err := c.Client.Subscribe(ctx)
	if err != nil {
		return err
	}

	streamClient := &StreamClient{
		Ctx:             ctx,
		SubscribeClient: stream,
		SubscribeRequest: &pb.SubscribeRequest{
			Accounts:           make(map[string]*pb.SubscribeRequestFilterAccounts),
			Slots:              make(map[string]*pb.SubscribeRequestFilterSlots),
			Transactions:       make(map[string]*pb.SubscribeRequestFilterTransactions),
			TransactionsStatus: make(map[string]*pb.SubscribeRequestFilterTransactions),
			Blocks:             make(map[string]*pb.SubscribeRequestFilterBlocks),
			BlocksMeta:         make(map[string]*pb.SubscribeRequestFilterBlocksMeta),
			Entry:              make(map[string]*pb.SubscribeRequestFilterEntry),
			AccountsDataSlice:  make([]*pb.SubscribeRequestAccountsDataSlice, 0),
		},
		UpdateCh: make(chan *pb.SubscribeUpdate), // Unbuffered channel to prevent overflowing
		ErrCh:    make(chan error, 10),
	}

	c.Streams.Store(clientName, streamClient)
	go streamClient.listen()

	return nil
}

// listen starts listening for responses and errors.
func (s *StreamClient) listen() {
	for {
		select {
		case <-s.Ctx.Done():
			return
		default:
			recv, err := s.SubscribeClient.Recv()
			if err != nil {
				s.ErrCh <- err
			} else {
				s.UpdateCh <- recv
			}
		}
	}
}

func (gc *GeyserClient) GetBlockHeight(
	ctx context.Context,
	commitment *pb.CommitmentLevel,
) (*pb.GetBlockHeightResponse, error) {
	return gc.Client.GetBlockHeight(ctx, &pb.GetBlockHeightRequest{
		Commitment: commitment.Enum(),
	})
}

func (gc *GeyserClient) GetLatestBlockhash(
	ctx context.Context,
	commitment *pb.CommitmentLevel,
) (*pb.GetLatestBlockhashResponse, error) {
	return gc.Client.GetLatestBlockhash(ctx, &pb.GetLatestBlockhashRequest{
		Commitment: commitment.Enum(),
	})
}

func (gc *GeyserClient) GetSlot(
	ctx context.Context,
	commitment *pb.CommitmentLevel,
) (*pb.GetSlotResponse, error) {
	return gc.Client.GetSlot(ctx, &pb.GetSlotRequest{
		Commitment: commitment.Enum(),
	})
}

func (gc *GeyserClient) GetVersion(
	ctx context.Context,
) (*pb.GetVersionResponse, error) {
	return gc.Client.GetVersion(ctx, &pb.GetVersionRequest{})
}

func (gc *GeyserClient) IsBlockhashValid(
	ctx context.Context,
	blockhash string,
	commitment *pb.CommitmentLevel,
) (*pb.IsBlockhashValidResponse, error) {
	return gc.Client.IsBlockhashValid(ctx, &pb.IsBlockhashValidRequest{
		Blockhash:  blockhash,
		Commitment: commitment.Enum(),
	})
}

func (gc *GeyserClient) Ping(
	ctx context.Context,
	count int32,
) (*pb.PongResponse, error) {
	return gc.Client.Ping(ctx, &pb.PingRequest{Count: count})
}

func (c *GeyserClient) SetDefaultSubscribeClient(client pb.Geyser_SubscribeClient) *GeyserClient {
	c.DefaultStreamClient.SubscribeClient = client
	return c
}

func (s *StreamClient) SubscribeAccounts(filterName string, req *pb.SubscribeRequestFilterAccounts) error {
	s.SubscribeRequest.Accounts[filterName] = req
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) AppendAccounts(filterName string, accounts ...string) error {
	s.SubscribeRequest.Accounts[filterName].Account = append(s.SubscribeRequest.Accounts[filterName].Account, accounts...)
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) UnsubscribeAccountsByID(filterName string) error {
	delete(s.SubscribeRequest.Accounts, filterName)
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) UnsubscribeAccounts(filterName string, accounts ...string) error {
	for _, account := range accounts {
		s.SubscribeRequest.Accounts[filterName].Account = slices.DeleteFunc(s.SubscribeRequest.Accounts[filterName].Account, func(a string) bool {
			return a == account
		})
	}
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) UnsubscribeAllAccounts(filterName string) error {
	delete(s.SubscribeRequest.Accounts, filterName)
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) SubscribeSlots(filterName string, req *pb.SubscribeRequestFilterSlots) error {
	s.SubscribeRequest.Slots[filterName] = req
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) UnsubscribeSlots(filterName string) error {
	delete(s.SubscribeRequest.Slots, filterName)
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) SubscribeTransaction(filterName string, req *pb.SubscribeRequestFilterTransactions) error {
	s.SubscribeRequest.Transactions[filterName] = req
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) UnsubscribeTransaction(filterName string) error {
	delete(s.SubscribeRequest.Transactions, filterName)
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) SubscribeTransactionStatus(filterName string, req *pb.SubscribeRequestFilterTransactions) error {
	s.SubscribeRequest.TransactionsStatus[filterName] = req
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) UnsubscribeTransactionStatus(filterName string) error {
	delete(s.SubscribeRequest.TransactionsStatus, filterName)
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) SubscribeBlocks(filterName string, req *pb.SubscribeRequestFilterBlocks) error {
	s.SubscribeRequest.Blocks[filterName] = req
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) UnsubscribeBlocks(filterName string) error {
	delete(s.SubscribeRequest.Blocks, filterName)
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) SubscribeBlocksMeta(filterName string, req *pb.SubscribeRequestFilterBlocksMeta) error {
	s.SubscribeRequest.BlocksMeta[filterName] = req
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) UnsubscribeBlocksMeta(filterName string) error {
	delete(s.SubscribeRequest.BlocksMeta, filterName)
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) SubscribeEntry(filterName string, req *pb.SubscribeRequestFilterEntry) error {
	s.SubscribeRequest.Entry[filterName] = req
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) UnsubscribeEntry(filterName string) error {
	delete(s.SubscribeRequest.Entry, filterName)
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) SubscribeAccountDataSlice(req []*pb.SubscribeRequestAccountsDataSlice) error {
	s.SubscribeRequest.AccountsDataSlice = req
	return s.SubscribeClient.Send(s.SubscribeRequest)
}

func (s *StreamClient) UnsubscribeAccountDataSlice() error {
	s.SubscribeRequest.AccountsDataSlice = nil
	return s.SubscribeClient.Send(s.SubscribeRequest)
}
