// +build unit

package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

type mockConfig struct{}

func (mockConfig) GetNetworkString() string {
	return "test network"
}

func TestHandler_Ping(t *testing.T) {
	h := handler{mockConfig{}}
	r := httptest.NewRequest("GET", "/ping", nil)
	res := httptest.NewRecorder()
	h.Ping(res, r)
	assert.Equal(t, res.Code, http.StatusOK)
	dec := json.NewDecoder(res.Body)
	var pong Pong
	assert.NoError(t, dec.Decode(&pong))
	assert.Equal(t, pong.Network, "test network")
}

func TestRegister(t *testing.T) {
	r := chi.NewRouter()
	Register(r, mockConfig{})
	assert.Len(t, r.Routes(), 1)
	assert.Equal(t, r.Routes()[0].Pattern, "/ping")
}
