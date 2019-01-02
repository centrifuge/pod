package api

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/errors"
)

// Bootstrapper implements bootstrapper.Bootstrapper
type Bootstrapper struct{}

// Bootstrap initiates api server
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, ok := ctx[bootstrap.BootstrappedConfig].(Config)
	if !ok {
		return errors.New("config not initialised")
	}

	srv := apiServer{config: cfg}
	ctx[bootstrap.BootstrappedAPIServer] = srv
	return nil
}
