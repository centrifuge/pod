package healthcheck

import (
	"context"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/health"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/golang/protobuf/ptypes/empty"
)

// handler is the grpc handler that implements healthpb.HealthCheckServiceServer
type handler struct {
	config config.Config
}

// GRPCHandler returns the grpc implementation instance of healthpb.HealthCheckServiceServer
func GRPCHandler(config config.Config) healthpb.HealthCheckServiceServer {
	return handler{config}
}

// Ping returns the network node is connected to and version of the node
func (h handler) Ping(context.Context, *empty.Empty) (pong *healthpb.Pong, err error) {
	return &healthpb.Pong{
		Version: version.GetVersion().String(),
		Network: h.config.GetNetworkString(),
	}, nil
}
