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
	"github.com/centrifuge/go-centrifuge/keytools/ed25519"
	"github.com/centrifuge/go-centrifuge/p2p"
)

type Bootstrapper struct {
}

func (*Bootstrapper) Bootstrap(c map[string]interface{}) error {
	if _, ok := c[bootstrap.BootstrappedConfig]; ok {
		services, err := defaultServerList()
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
	return errors.New("could not initialize node")
}

func defaultServerList() ([]Server, error) {
	publicKey, privateKey, err := ed25519.GetSigningKeyPairFromConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %v", err)
	}

	return []Server{
		api.NewCentAPIServer(
			config.Config().GetServerAddress(),
			config.Config().GetServerPort(),
			config.Config().GetNetworkString(),
		),
		p2p.NewCentP2PServer(
			config.Config().GetP2PPort(),
			config.Config().GetBootstrapPeers(),
			publicKey, privateKey,
		),
	}, nil
}
