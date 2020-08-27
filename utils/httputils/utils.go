package httputils

import (
	"net/http"

	"github.com/go-chi/render"
)

// HTTPError contains the error message
type HTTPError struct {
	Message string `json:"message"`
}

// RespondIfError if err != nil, returns the HTTPError and code as API response
// no-op if the err is nil
func RespondIfError(code *int, err *error, w http.ResponseWriter, r *http.Request) {
	if *err == nil {
		return
	}

	e := *err
	render.Status(r, *code)
	render.JSON(w, r, HTTPError{Message: e.Error()})
}
