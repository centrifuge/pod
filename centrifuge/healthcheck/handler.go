package healthcheck

import (
	"context"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/health"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes/empty"
)

// Handler implements HealthCheckServiceServer
type Handler struct{}

// Ping return current version and network of the node
func (hcs *Handler) Ping(context.Context, *empty.Empty) (pong *healthpb.Pong, err error) {
	pong = &healthpb.Pong{
		Version: version.GetVersion().String(),
		Network: config.Config.GetNetworkString(),
	}
	return pong, nil
}
