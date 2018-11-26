package healthcheck

import (
	"context"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/health"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/golang/protobuf/ptypes/empty"
)

// Config defines methods required for the package healthcheck
type Config interface {
	GetNetworkString() string
}

// handler is the grpc handler that implements healthpb.HealthCheckServiceServer
type handler struct {
	config Config
}

// GRPCHandler returns the grpc implementation instance of healthpb.HealthCheckServiceServer
func GRPCHandler(config Config) healthpb.HealthCheckServiceServer {
	return handler{config}
}

// Ping returns the network node is connected to and version of the node
func (h handler) Ping(context.Context, *empty.Empty) (pong *healthpb.Pong, err error) {
	return &healthpb.Pong{
		Version: version.GetVersion().String(),
		Network: h.config.GetNetworkString(),
	}, nil
}
