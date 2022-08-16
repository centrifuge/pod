package http

import (
	"context"

	v3 "github.com/centrifuge/go-centrifuge/http/v3"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	auth2 "github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-centrifuge/http/health"
	v2 "github.com/centrifuge/go-centrifuge/http/v2"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Router returns the http mux for the server.
// @title Centrifuge OS Node API
// @description Centrifuge OS Node API
// @version 2.1.0
// @contact.name Centrifuge
// @contact.url https://github.com/centrifuge/go-centrifuge
// @contact.email hello@centrifuge.io
// @BasePath /
// @license.name MIT
// @host localhost:8082
// @schemes http
func Router(ctx context.Context) (*chi.Mux, error) {
	r := chi.NewRouter()
	cctx, ok := ctx.Value(bootstrap.NodeObjRegistry).(map[string]interface{})
	if !ok {
		return nil, errors.New("failed to get node object registry %s", bootstrap.NodeObjRegistry)
	}

	cfg, ok := cctx[bootstrap.BootstrappedConfig].(config.Configuration)
	if !ok {
		return nil, errors.New("failed to get %s", bootstrap.BootstrappedConfig)
	}

	cfgService, ok := cctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return nil, errors.New("failed to get %s", config.BootstrappedConfigStorage)
	}

	authService, ok := cctx[BootstrappedAuthService].(auth2.Service)
	if !ok {
		return nil, errors.New("failed to get %s", BootstrappedAuthService)
	}

	// add middlewares. do not change the order. Add any new middlewares to the bottom
	r.Use(middleware.Recoverer)
	r.Use(middleware.DefaultLogger)
	r.Use(auth(authService, cfgService))

	// health check
	health.Register(r, cfg)

	// v2 apis
	r.Route("/v2", func(r chi.Router) {
		v2.Register(cctx, r)
	})

	// v3 apis
	r.Route("/v3", func(r chi.Router) {
		v3.Register(cctx, r)
	})

	return r, nil
}
