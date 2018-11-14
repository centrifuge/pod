package node

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/centrifuge/go-centrifuge/bootstrap"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(c map[string]interface{}) error {
	srvs, err := getServers(c)
	if err != nil {
		return fmt.Errorf("failed to load servers: %v", err)
	}

	n := NewNode(srvs)
	feedback := make(chan error)
	// may be we can pass a context that exists in c here
	ctx, canc := context.WithCancel(context.Background())
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

	return nil
}

func getServers(ctx map[string]interface{}) ([]Server, error) {
	p2pSrv, ok := ctx[bootstrap.BootstrappedP2PServer]
	if !ok {
		return nil, fmt.Errorf("p2p server not initialised")
	}

	apiSrv, ok := ctx[bootstrap.BootstrappedAPIServer]
	if !ok {
		return nil, fmt.Errorf("API server not initiliase")
	}

	var servers []Server
	servers = append(servers, p2pSrv.(Server))
	servers = append(servers, apiSrv.(Server))
	return servers, nil
}
