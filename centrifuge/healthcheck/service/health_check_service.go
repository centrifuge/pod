package healthcheckservice

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/health"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
)

type HealthCheckService struct{}

func (hcs *HealthCheckService) Ping() (pong *healthpb.Pong, err error) {
	pong = new(healthpb.Pong)
	pong.Version = version.GetVersion().String()
	pong.Network = config.Config.GetNetworkString()
	return pong, nil
}
