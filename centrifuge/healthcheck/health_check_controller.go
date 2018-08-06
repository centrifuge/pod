package healthcheck

import (
	"context"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/health"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/golang/protobuf/ptypes/empty"
)

// HealthCheckController interfaces the grpc health check calls
type HealthCheckController struct{}

func (hcc *HealthCheckController) Ping(ctx context.Context, empty *empty.Empty) (pong *healthpb.Pong, err error) {
	pong = new(healthpb.Pong)
	pong.Version = version.GetVersion().String()
	pong.Network = config.Config.GetNetworkString()
	return pong, nil
}
