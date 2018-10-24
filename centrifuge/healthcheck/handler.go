package healthcheck

import (
	"context"

	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/health"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes/empty"
)

// handler is the grpc handler that implements healthpb.HealthCheckServiceServer
type handler struct{}

// GRPCHandler returns the grpc implementation instance of healthpb.HealthCheckServiceServer
func GRPCHandler() healthpb.HealthCheckServiceServer {
	return handler{}
}

// Ping returns the network node is connected to and version of the node
func (handler) Ping(context.Context, *empty.Empty) (pong *healthpb.Pong, err error) {
	return &healthpb.Pong{
		Version: version.GetVersion().String(),
		Network: config.Config.GetNetworkString(),
	}, nil
}
