package http

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	auth2 "github.com/centrifuge/go-centrifuge/http/auth"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	v2proxy "github.com/centrifuge/go-centrifuge/identity/v2/proxy"
)

const (
	BootstrappedAuthService = "BootstrappedAuthService"
)

// Bootstrapper implements bootstrapper.Bootstrapper
type Bootstrapper struct{}

// Bootstrap initiates api server
func (b Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfgService, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("config storage not initialised")
	}

	proxyAPI, ok := ctx[v2.BootstrappedProxyAPI].(v2proxy.API)
	if !ok {
		return errors.New("proxy API not initialised")
	}

	cfg, err := cfgService.GetConfig()

	if err != nil {
		return fmt.Errorf("couldn't retrieve config: %s", err)
	}

	authService := auth2.NewAuth(cfg.IsAuthenticationEnabled(), proxyAPI, cfgService)

	ctx[BootstrappedAuthService] = authService

	srv := apiServer{config: cfg}
	ctx[bootstrap.BootstrappedAPIServer] = srv
	return nil
}
