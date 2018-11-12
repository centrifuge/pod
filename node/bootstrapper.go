package node

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/centrifuge/go-centrifuge/api"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/p2p"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(c map[string]interface{}) error {
	if _, ok := c[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := c[bootstrap.BootstrappedConfig].(*config.Configuration)

	services, err := defaultServerList(cfg)
	if err != nil {
		return fmt.Errorf("failed to get default server list: %v", err)
	}

	n := NewNode(services)
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

func defaultServerList(cfg *config.Configuration) ([]Server, error) {
	return []Server{api.NewCentAPIServer(cfg), p2p.NewCentP2PServer(cfg)}, nil
}
