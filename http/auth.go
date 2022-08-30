package http

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	auth2 "github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-centrifuge/utils/httputils"
	"github.com/go-chi/render"
)

var (
	adminPathRegex = regexp.MustCompile(`^/v2/accounts(|/generate|/0x[a-fA-F0-9]+)$`)
)

func isAdminPath(path string) bool {
	return adminPathRegex.MatchString(path)
}

func auth(authService auth2.Service, cfgService config.Service) func(handler http.Handler) http.Handler {
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
			// Header format -> "Authorization": "Bearer $jwt"
			authHeader := r.Header.Get("Authorization")
			bearer := strings.Split(authHeader, " ")
			if len(bearer) != 2 {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			}
			accHeader, err := authService.Validate(r.Context(), bearer[1])
			if err != nil {
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			}

			isAdminPath := isAdminPath(path)

			switch {
			case isAdminPath && !accHeader.IsAdmin:
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			case isAdminPath:
				handler.ServeHTTP(w, r)
				return
			}

			acc, err := cfgService.GetAccount(accHeader.Identity.ToBytes())
			if err != nil {
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
