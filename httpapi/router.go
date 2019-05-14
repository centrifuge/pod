package httpapi

import (
	"github.com/centrifuge/go-centrifuge/httpapi/health"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Router returns the http mux for the server.
// @title Centrifuge OS Node API
// @description Centrifuge OS Node API
// @version 0.0.4
// @contact.name Centrifuge
// @contact.url https://github.com/centrifuge/go-centrifuge
// @contact.email hello@centrifuge.io
// @BasePath /
// @license.name MIT
// @host localhost:8082
// @schemes http
func Router(config Config) *chi.Mux {
	r := chi.NewRouter()

	// add middlewares
	r.Use(middleware.Recoverer)
	r.Use(middleware.DefaultLogger)

	// health check
	health.Register(r, config)
	return r
}

// Config defines required methods for http API
// this will be the super set for the configs defined in sub packages
type Config interface {
	GetNetworkString() string
}
