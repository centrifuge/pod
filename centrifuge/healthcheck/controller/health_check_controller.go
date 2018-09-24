package healthcheckcontroller

import (
	"context"

	"github.com/centrifuge/go-centrifuge/centrifuge/healthcheck/service"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/health"
	"github.com/golang/protobuf/ptypes/empty"
)

// getHealthCheckService returns a new instance of HealthCheckService
func getHealthCheckService() *healthcheckservice.HealthCheckService {
	return &healthcheckservice.HealthCheckService{}
}

// HealthCheckController interfaces the grpc health check calls
type HealthCheckController struct{}

func (hcc *HealthCheckController) Ping(ctx context.Context, empty *empty.Empty) (pong *healthpb.Pong, err error) {
	return getHealthCheckService().Ping()
}
