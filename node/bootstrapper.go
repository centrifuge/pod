package node

import (
	"context"
	"os"
	"os/signal"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct{}

// Bootstrap runs the servers.
// Note: this is a blocking call.
func (*Bootstrapper) Bootstrap(c map[string]interface{}) error {
	srvs, err := GetServers(c)
	if err != nil {
		cleanUp(c)
		return errors.New("failed to load servers: %v", err)
	}

	n := New(srvs)
	feedback := make(chan error)
	// may be we can pass a context that exists in c here
	ctx, canc := context.WithCancel(context.WithValue(context.Background(), bootstrap.NodeObjRegistry, c))
	go n.Start(ctx, feedback)
	controlC := make(chan os.Signal, 1)
	signal.Notify(controlC, os.Interrupt)
	select {
	case err := <-feedback:
		cleanUp(c)
		panic(err)
	case sig := <-controlC:
		log.Info("Node shutting down because of ", sig)
		canc()
		cleanUp(c)
		err := <-feedback
		return err
	}
}

func cleanUp(c map[string]interface{}) {
	// close the node db
	db := c[storage.BootstrappedDB].(storage.Repository)
	db.Close()
}

// GetServers gets the long running background services in the node as a list
func GetServers(ctx map[string]interface{}) ([]Server, error) {
	p2pSrv, ok := ctx[bootstrap.BootstrappedPeer]
	if !ok {
		return nil, errors.New("p2p server not initialized")
	}

	apiSrv, ok := ctx[bootstrap.BootstrappedAPIServer]
	if !ok {
		return nil, errors.New("API server not initialized")
	}

	dispatcher, ok := ctx[jobs.BootstrappedDispatcher].(Server)
	if !ok {
		return nil, errors.New("Node: dispatcher server not initialised")
	}

	var servers []Server
	servers = append(servers, p2pSrv.(Server), apiSrv.(Server), dispatcher)
	return servers, nil
}
