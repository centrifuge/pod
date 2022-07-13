package http

import (
	"context"
	"net/http"
	"strings"

	auth2 "github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-centrifuge/proxy"
	"github.com/vedhavyas/go-subkey/v2"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/health"
	v2 "github.com/centrifuge/go-centrifuge/http/v2"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// Router returns the http mux for the server.
// @title Centrifuge OS Node API
// @description Centrifuge OS Node API
// @version 2.0.0
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
		return nil, errors.New("failed to get %s", bootstrap.NodeObjRegistry)
	}

	cfg, ok := cctx[bootstrap.BootstrappedConfig].(Config)
	if !ok {
		return nil, errors.New("failed to get %s", bootstrap.BootstrappedConfig)
	}

	configSrv, ok := cctx[config.BootstrappedConfigStorage].(config.Service)
	if !ok {
		return nil, errors.New("failed to get %s", config.BootstrappedConfigStorage)
	}

	proxySrv, ok := cctx[proxy.BootstrappedProxyService].(proxy.Service)
	if !ok {
		return nil, errors.New("failed to get %s", proxy.ErrProxyRepoNotInitialised)
	}

	// add middlewares. do not change the order. Add any new middlewares to the bottom
	r.Use(middleware.Recoverer)
	r.Use(middleware.DefaultLogger)
	r.Use(auth(configSrv, proxySrv))

	// health check
	health.Register(r, cfg)

	// v2 apis
	r.Route("/v2", func(r chi.Router) {
		v2.Register(cctx, r)
	})

	r.Route("/beta", func(r chi.Router) {
		v2.RegisterBeta(cctx, r)
	})

	return r, nil
}

func auth(configSrv config.Service, proxySrv proxy.Service) func(handler http.Handler) http.Handler {
	// TODO(ved): regex would be a better alternative
	skippedURLs := []string{
		"/ping",
	}
	adminOnlyURLs := []string{
		"/accounts",
		"/accounts/generate", //TODO: Change to AddAccount later when ready
	}
	skipAuthentication := true // TODO: Read that flag from the node config
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if utils.ContainsString(skippedURLs, path) {
				handler.ServeHTTP(w, r)
				return
			}
			// Header format -> "Authorization": "Bearer $jwt"
			authHeader := r.Header.Get("Authorization")
			bearer := strings.Split(authHeader, " ")
			if len(bearer) != 2 {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			}
			authSrv := auth2.NewAuth(configSrv, proxySrv)
			// TODO: The reason why I pass for now the flag here, is so that in any case we will need the auth header key
			// to be used downstream
			accHeader, err := authSrv.Validate(bearer[1], skipAuthentication)
			if err != nil {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			}

			if !skipAuthentication && utils.ContainsString(adminOnlyURLs, path) && accHeader.ProxyType != proxy.NodeAdminRole {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			}

			// TODO remove this block as soon as we have converted the new DID types
			_, pk, err := subkey.SS58Decode(accHeader.Identity)
			if err != nil {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			}
			//

			ctx, err := contextutil.Context(context.WithValue(r.Context(), config.AccountHeaderKey, pk), configSrv)
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
