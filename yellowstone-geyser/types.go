package yellowstone_geyser

import (
	"context"
	"sync"

	pb "github.com/Prophet-Solutions/yellowstone-geyser-protos/geyser"
	"google.golang.org/grpc"
)

// YellowstoneGeyserClient is the main client struct that holds the gRPC connection
// and the GeyserClient for communicating with the Yellowstone Geyser service.
type GeyserClient struct {
	GRPCConn            *grpc.ClientConn // gRPC Connection
	Ctx                 context.Context  // Context for cancellation and deadlines
	Client              pb.GeyserClient  // Geyser client from protobuf
	Streams             sync.Map         // Active stream clients
	DefaultStreamClient *StreamClient    // Default stream client
	ErrCh               chan error       // Channel for errors
}

type StreamClient struct {
	Ctx              context.Context           // Context for cancellation and deadlines
	SubscribeClient  pb.Geyser_SubscribeClient // Geyser subscribe client
	SubscribeRequest *pb.SubscribeRequest      // Subscribe request
	UpdateCh         chan *pb.SubscribeUpdate  // Channel for updates
	ErrCh            chan error                // Channel for errors
}
