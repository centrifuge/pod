package http

import (
	"context"
	"fmt"
	"sync"

	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/errors"
	auth2 "github.com/centrifuge/pod/http/auth"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pallets/proxy"
)

const (
	BootstrappedAuthService = "BootstrappedAuthService"
)

// Bootstrapper implements bootstrapper.Bootstrapper
type Bootstrapper struct {
	testServerWg        sync.WaitGroup
	testServerCtx       context.Context
	testServerCtxCancel context.CancelFunc
}

// Bootstrap initiates api server
func (b *Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	cfgService, ok := ctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return errors.New("config storage not initialised")
	}

	proxyAPI, ok := ctx[pallets.BootstrappedProxyAPI].(proxy.API)
	if !ok {
		return errors.New("proxy API not initialised")
	}

	cfg, err := cfgService.GetConfig()

	if err != nil {
		return fmt.Errorf("couldn't retrieve config: %s", err)
	}

	authService := auth2.NewService(cfg.IsAuthenticationEnabled(), proxyAPI, cfgService)

	ctx[BootstrappedAuthService] = authService

	srv := apiServer{config: cfg}
	ctx[bootstrap.BootstrappedAPIServer] = srv
	return nil
}
