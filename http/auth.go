package http

import (
	"net/http"

	"github.com/centrifuge/pod/http/auth/access"
	"github.com/centrifuge/pod/utils/httputils"
	"github.com/go-chi/render"
)

func auth(validationServices access.ValidationServices) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := validationServices.Validate(r); err != nil {
				log.Errorf("Couldn't validate access for request: %s", err)
				render.Status(r, http.StatusForbidden)
				render.JSON(w, r, httputils.HTTPError{Message: "Authentication failed"})
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}
