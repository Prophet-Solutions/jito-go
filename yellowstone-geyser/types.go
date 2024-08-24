package yellowstone_geyser

import (
	pb "github.com/Prophet-Solutions/yellowstone-geyser-protos/geyser"
	"google.golang.org/grpc"
)

// YellowstoneGeyserClient is the main client struct that holds the gRPC connection
// and the GeyserClient for communicating with the Yellowstone Geyser service.
type YellowstoneGeyserClient struct {
	GRPCConn *grpc.ClientConn
	Stream   *pb.Geyser_SubscribeClient
	Client   pb.GeyserClient
}
