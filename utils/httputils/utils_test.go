//go:build unit

package httputils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/stretchr/testify/assert"
)

func TestRespondIfError_noError(t *testing.T) {
	var err error
	code := http.StatusOK
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/documents", nil)
	defer func() {
		assert.Equal(t, w.Code, code)
		assert.Empty(t, w.Body)
	}()

	defer RespondIfError(&code, &err, w, r)
}

func TestRespondIfError_error(t *testing.T) {
	var err error
	code := http.StatusBadRequest
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/documents", nil)
	defer func() {
		assert.Equal(t, w.Code, code)
		assert.Contains(t, w.Body.String(), "bad request")
	}()

	defer RespondIfError(&code, &err, w, r)
	err = errors.New("bad request")
}
