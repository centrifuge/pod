package node

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/centrifuge/go-centrifuge/bootstrap"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap runs the severs.
// Note: this is a blocking call.
func (*Bootstrapper) Bootstrap(c map[string]interface{}) error {
	srvs, err := getServers(c)
	if err != nil {
		return fmt.Errorf("failed to load servers: %v", err)
	}

	n := New(srvs)
	feedback := make(chan error)
	// may be we can pass a context that exists in c here
	ctx, canc := context.WithCancel(context.WithValue(context.Background(), bootstrap.NodeObjRegistry, c))
	go n.Start(ctx, feedback)
	controlC := make(chan os.Signal, 1)
	signal.Notify(controlC, os.Interrupt)
	for {
		select {
		case err := <-feedback:
			panic(err)
		case sig := <-controlC:
			log.Info("Node shutting down because of ", sig)
			canc()
			err := <-feedback
			return err
		}
	}
}

func getServers(ctx map[string]interface{}) ([]Server, error) {
	p2pSrv, ok := ctx[bootstrap.BootstrappedP2PServer]
	if !ok {
		return nil, fmt.Errorf("p2p server not initialized")
	}

	apiSrv, ok := ctx[bootstrap.BootstrappedAPIServer]
	if !ok {
		return nil, fmt.Errorf("API server not initialized")
	}

	queueSrv, ok := ctx[bootstrap.BootstrappedQueueServer]
	if !ok {
		return nil, fmt.Errorf("queue server not initialized")
	}

	var servers []Server
	servers = append(servers, p2pSrv.(Server), apiSrv.(Server), queueSrv.(Server))
	return servers, nil
}
