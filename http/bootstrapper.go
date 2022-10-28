package http

import (
	"context"
	"fmt"
	"sync"

	"github.com/centrifuge/go-centrifuge/pallets"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	auth2 "github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
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
