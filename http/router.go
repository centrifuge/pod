package http

import (
	"context"
	"fmt"

	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/http/auth/access"
	"github.com/centrifuge/pod/http/health"
	v2 "github.com/centrifuge/pod/http/v2"
	v3 "github.com/centrifuge/pod/http/v3"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Router returns the http mux for the server.
// @title Centrifuge OS Node API
// @description Centrifuge OS Node API
// @version 3.0.0
// @contact.name Centrifuge
// @contact.url https://github.com/centrifuge/pod
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

	validationWrapperFactory, ok := cctx[BootstrappedValidationWrapperFactory].(access.ValidationWrapperFactory)
	if !ok {
		return nil, errors.New("failed to get %s", BootstrappedValidationWrapperFactory)
	}

	wrappers, err := validationWrapperFactory.GetValidationWrappers()

	if err != nil {
		return nil, fmt.Errorf("couldn't get default validation wrappers")
	}

	// add middlewares. do not change the order. Add any new middlewares to the bottom
	r.Use(middleware.Recoverer)
	r.Use(middleware.DefaultLogger)
	r.Use(auth(wrappers))

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
