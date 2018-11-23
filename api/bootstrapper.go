package api

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
)

// Bootstrapper implements bootstrapper.Bootstrapper
type Bootstrapper struct{}

// Bootstrap initiates api server
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, ok := ctx[config.BootstrappedConfig].(Config)
	if !ok {
		return fmt.Errorf("config not initialised")
	}

	// just check to make sure that registry is initialised
	_, ok = ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	if !ok {
		return fmt.Errorf("service registry not initialised")
	}

	srv := apiServer{config: cfg}
	ctx[bootstrap.BootstrappedAPIServer] = srv
	return nil
}
