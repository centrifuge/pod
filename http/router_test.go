//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	httpAuth "github.com/centrifuge/go-centrifuge/http/auth"
	httpV2 "github.com/centrifuge/go-centrifuge/http/v2"
	httpV3 "github.com/centrifuge/go-centrifuge/http/v3"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

func TestRouter_auth(t *testing.T) {
	authServiceMock := httpAuth.NewServiceMock(t)
	configServiceMock := config.NewServiceMock(t)

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := httpAuth.CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		proxyType.ProxyTypeName[proxyType.PodAuth],
	)
	assert.NoError(t, err)

	// missing auth
	r := httptest.NewRequest("POST", "/documents", nil)
	w := httptest.NewRecorder()

	auth(authServiceMock, configServiceMock)(nil).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusForbidden)
	assert.Contains(t, w.Body.String(), "Authentication failed")

	// ping
	r = httptest.NewRequest("POST", "/ping", nil)
	w = httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acc, err := contextutil.Account(r.Context())
		assert.ErrorIs(t, err, contextutil.ErrSelfNotFound)
		assert.Nil(t, acc)
		w.WriteHeader(http.StatusOK)
	})

	auth(authServiceMock, configServiceMock)(next).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK)

	// accounts
	r = httptest.NewRequest("GET", "/accounts/self", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()

	accHeader := &httpAuth.AccountHeader{
		Identity: delegatorAccountID,
		IsAdmin:  false,
	}

	authServiceMock.On("Validate", r.Context(), token).
		Return(accHeader, nil).
		Once()

	accountMock := config.NewAccountMock(t)

	configServiceMock.On("GetAccount", delegatorAccountID.ToBytes()).
		Return(accountMock, nil).
		Once()

	next = func(w http.ResponseWriter, r *http.Request) {
		acc, err := contextutil.Account(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, accountMock, acc)
		w.WriteHeader(http.StatusOK)
	}

	auth(authServiceMock, configServiceMock)(next).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK)

	// Auth service failure
	r = httptest.NewRequest("GET", "/accounts/self", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()

	authServiceMock.On("Validate", r.Context(), token).
		Return(nil, errors.New("error")).
		Once()

	auth(authServiceMock, configServiceMock)(nil).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusForbidden)

	// Config service failure
	r = httptest.NewRequest("GET", "/accounts/self", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()

	authServiceMock.On("Validate", r.Context(), token).
		Return(accHeader, nil).
		Once()

	configServiceMock.On("GetAccount", delegatorAccountID.ToBytes()).
		Return(nil, errors.New("error")).
		Once()

	auth(authServiceMock, configServiceMock)(nil).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusForbidden)
}

func TestRouter_auth_AdminPath(t *testing.T) {
	authServiceMock := httpAuth.NewServiceMock(t)
	configServiceMock := config.NewServiceMock(t)

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := httpAuth.CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		proxyType.ProxyTypeName[proxyType.PodAuth],
	)
	assert.NoError(t, err)

	// Not an admin
	r := httptest.NewRequest("POST", "/v2/accounts/generate", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acc, err := contextutil.Account(r.Context())
		assert.ErrorIs(t, err, contextutil.ErrSelfNotFound)
		assert.Nil(t, acc)
		w.WriteHeader(http.StatusOK)
	})

	accHeader := &httpAuth.AccountHeader{
		Identity: delegatorAccountID,
		IsAdmin:  false,
	}

	authServiceMock.On("Validate", r.Context(), token).
		Return(accHeader, nil).
		Once()

	auth(authServiceMock, configServiceMock)(next).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusForbidden)

	// Admin
	r = httptest.NewRequest("POST", "/v2/accounts/generate", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()

	accHeader = &httpAuth.AccountHeader{
		Identity: delegatorAccountID,
		IsAdmin:  true,
	}

	authServiceMock.On("Validate", r.Context(), token).
		Return(accHeader, nil).
		Once()

	next = func(w http.ResponseWriter, r *http.Request) {
		acc, err := contextutil.Account(r.Context())
		assert.ErrorIs(t, err, contextutil.ErrSelfNotFound)
		assert.Nil(t, acc)
		w.WriteHeader(http.StatusOK)
	}

	auth(authServiceMock, configServiceMock)(next).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK)

	// Validation error
	r = httptest.NewRequest("POST", "/v2/accounts/generate", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()

	authServiceMock.On("Validate", r.Context(), token).
		Return(nil, errors.New("error")).
		Once()

	auth(authServiceMock, configServiceMock)(nil).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusForbidden)
}

func TestRouter(t *testing.T) {
	cctx := map[string]interface{}{
		bootstrap.BootstrappedConfig:     config.NewConfigurationMock(t),
		config.BootstrappedConfigStorage: config.NewServiceMock(t),
		BootstrappedAuthService:          httpAuth.NewServiceMock(t),
		httpV2.BootstrappedService:       &httpV2.Service{},
		httpV3.BootstrappedService:       &httpV3.Service{},
	}

	ctx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, cctx)
	r, err := Router(ctx)
	assert.NoError(t, err)
	assert.Len(t, r.Middlewares(), 3)
	assert.Len(t, r.Routes(), 3)

	// health pattern
	assert.Equal(t, "/ping", r.Routes()[0].Pattern)
	// v2 routes
	assert.Len(t, r.Routes()[1].SubRoutes.Routes(), 25)
	// v3 routes
	assert.Len(t, r.Routes()[2].SubRoutes.Routes(), 6)
}
