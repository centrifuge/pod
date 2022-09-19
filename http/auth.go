package http

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	httpAuth "github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/go-chi/render"
)

var (
	adminPathRegex = regexp.MustCompile(`^/accounts(|/generate|/0x[a-fA-F0-9]+)$`)
)

func isAdminPath(path string) bool {
	return adminPathRegex.MatchString(path)
}

func auth(authService httpAuth.Service, cfgService config.Service) func(handler http.Handler) http.Handler {
	skippedURLs := map[string]struct{}{
		"/ping": {},
	}

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			if _, ok := skippedURLs[path]; ok {
				handler.ServeHTTP(w, r)
				return
			}

			// Header format -> "Authorization": "Bearer $jw3t"
			authHeader := r.Header.Get("Authorization")
			bearer := strings.Split(authHeader, " ")
			if len(bearer) != 2 {
				log.Error("Invalid auth header")
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			}
			accHeader, err := authService.Validate(r.Context(), bearer[1])
			if err != nil {
				log.Errorf("Couldn't validate token: %s", err)
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			}

			if isAdminPath(path) {
				if !accHeader.IsAdmin {
					log.Error("Account not an admin")
					render.Status(r, http.StatusForbidden)
					render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
					return
				}

				handler.ServeHTTP(w, r)
				return
			}

			acc, err := cfgService.GetAccount(accHeader.Identity.ToBytes())
			if err != nil {
				log.Errorf("Couldn't retrieve account from storage: %s", err)
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			}

			ctx := contextutil.WithAccount(r.Context(), acc)

			r = r.WithContext(ctx)
			handler.ServeHTTP(w, r)
		})
	}
}
