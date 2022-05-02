//go:build unit
// +build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	v2 "github.com/centrifuge/go-centrifuge/http/v2"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	testingnfts "github.com/centrifuge/go-centrifuge/testingutils/nfts"
	"github.com/stretchr/testify/assert"
)

func TestRouter_auth(t *testing.T) {
	// missing auth
	r := httptest.NewRequest("POST", "/documents", nil)
	w := httptest.NewRecorder()
	h := auth(nil)(nil)
	h.ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusForbidden)
	assert.Contains(t, w.Body.String(), "'authorization' header missing")

	// ping
	r = httptest.NewRequest("POST", "/ping", nil)
	w = httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		v := ctx.Value(config.AccountHeaderKey)
		assert.Nil(t, v)
		w.WriteHeader(http.StatusOK)
	})
	auth(nil)(next).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK)

	// accounts
	r = httptest.NewRequest("POST", "/accounts/0x123456789", nil)
	w = httptest.NewRecorder()
	next = func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		v := ctx.Value(config.AccountHeaderKey)
		assert.Nil(t, v)
		w.WriteHeader(http.StatusOK)
	}
	auth(nil)(next).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK)

	// success
	did := testingidentity.GenerateRandomDID()
	r = httptest.NewRequest("POST", "/documents", nil)
	r.Header.Set("authorization", did.String())
	w = httptest.NewRecorder()
	next = func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		v := ctx.Value(config.AccountHeaderKey).(string)
		assert.Equal(t, did.String(), v)
		w.WriteHeader(http.StatusOK)
	}
	cfgSrv := new(configstore.MockService)
	cfgSrv.On("GetAccount", did[:]).Return(nil, nil)
	auth(cfgSrv)(next).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	cfgSrv.AssertExpectations(t)
}

func TestRouter(t *testing.T) {
	cctx := map[string]interface{}{
		bootstrap.BootstrappedNFTService: new(testingnfts.MockNFTService),
		bootstrap.BootstrappedConfig:     new(testingconfig.MockConfig),
		config.BootstrappedConfigStorage: new(configstore.MockService),
		v2.BootstrappedService:           v2.Service{},
	}

	ctx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, cctx)
	r, err := Router(ctx)
	assert.NoError(t, err)
	assert.Len(t, r.Middlewares(), 3)
	assert.Len(t, r.Routes(), 3)
	// health pattern
	assert.Equal(t, "/ping", r.Routes()[1].Pattern)
	// v2 routes
	assert.Len(t, r.Routes()[2].SubRoutes.Routes(), 28)
}
