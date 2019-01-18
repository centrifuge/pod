package api

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config/configstore"
)

// Bootstrapper implements bootstrapper.Bootstrapper
type Bootstrapper struct{}

// Bootstrap initiates api server
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfg, err := configstore.RetrieveConfig(true, ctx)
	if err != nil {
		return err
	}

	srv := apiServer{config: cfg}
	ctx[bootstrap.BootstrappedAPIServer] = srv
	return nil
}
