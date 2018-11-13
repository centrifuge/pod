package api

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
)

// Bootstrapper implements bootstrapper.Bootstrapper
type Bootstrapper struct{}

// Bootstrap initiates api server
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	if _, ok := ctx[bootstrap.BootstrappedConfig]; !ok {
		return fmt.Errorf("config not initialised")
	}

	cfg := ctx[bootstrap.BootstrappedConfig].(*config.Configuration)
	srv := apiServer{
		addr:    cfg.GetServerAddress(),
		port:    cfg.GetServerPort(),
		network: cfg.GetNetworkString(),
	}

	ctx[bootstrap.BootstrappedAPIServer] = srv
	return nil
}
