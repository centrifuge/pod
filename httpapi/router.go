package httpapi

import (
	"context"
	"net/http"

	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/userapi"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/httpapi/health"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// Router returns the http mux for the server.
// @title Centrifuge OS Node API
// @description Centrifuge OS Node API
// @version 0.0.5
// @contact.name Centrifuge
// @contact.url https://github.com/centrifuge/go-centrifuge
// @contact.email hello@centrifuge.io
// @BasePath /
// @license.name MIT
// @host localhost:8082
// @schemes http
func Router(
	config Config,
	configSrv config.Service,
	nftSrv nft.Service,
	docsSrv documents.Service,
	transferSrv transferdetails.Service,
	jobsSrv jobs.Manager) *chi.Mux {
	r := chi.NewRouter()

	// add middlewares. do not change the order. Add any new middlewares to the bottom
	r.Use(middleware.Recoverer)
	r.Use(middleware.DefaultLogger)
	r.Use(auth(configSrv))

	// health check
	health.Register(r, config)

	r.Route("/v1", func(r chi.Router) {
		// core apis
		coreapi.Register(r, nftSrv, configSrv, docsSrv, jobsSrv)
		// user apis
		userapi.Register(r, nftSrv, transferSrv)
	})
	return r
}

// Config defines required methods for http API
// this will be the super set for the configs defined in sub packages
type Config interface {
	GetNetworkString() string
}

func auth(configSrv config.Service) func(handler http.Handler) http.Handler {
	skippedURLs := []string{
		"/ping",
		"/accounts", // since we use default account DID for endpoints
	}
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if utils.ContainsString(skippedURLs, path) {
				handler.ServeHTTP(w, r)
				return
			}

			did := r.Header.Get("authorization")
			if !common.IsHexAddress(did) {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "'authorization' header missing"})
				return
			}

			ctx, err := contextutil.Context(context.WithValue(r.Context(), config.AccountHeaderKey, did), configSrv)
			if err != nil {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: err.Error()})
				return
			}
			r = r.WithContext(ctx)
			handler.ServeHTTP(w, r)
		})
	}
}
