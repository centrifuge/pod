//go:build unit

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/http/auth/access"
	authToken "github.com/centrifuge/pod/http/auth/token"
	httpV2 "github.com/centrifuge/pod/http/v2"
	httpV3 "github.com/centrifuge/pod/http/v3"
	"github.com/centrifuge/pod/testingutils/keyrings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRouter_auth(t *testing.T) {
	validationWrapperMock := access.NewValidationWrapperMock(t)

	validationWrappers := access.ValidationWrappers{validationWrapperMock}

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := authToken.CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		proxyType.ProxyTypeName[proxyType.PodAuth],
	)
	assert.NoError(t, err)

	// missing auth
	path := "/documents"

	r := httptest.NewRequest("POST", path, nil)
	w := httptest.NewRecorder()

	validationWrapperMock.On("Matches", path).
		Return(false).
		Once()

	auth(validationWrappers)(nil).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusForbidden)
	assert.Contains(t, w.Body.String(), "Authentication failed")

	// ping
	path = "/ping"

	r = httptest.NewRequest("POST", path, nil)
	w = httptest.NewRecorder()

	validationWrapperMock.On("Matches", path).
		Return(true).
		Once()

	validationWrapperMock.On("Validate", r).
		Return(nil).
		Once()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acc, err := contextutil.Account(r.Context())
		assert.ErrorIs(t, err, contextutil.ErrSelfNotFound)
		assert.Nil(t, acc)
		w.WriteHeader(http.StatusOK)
	})

	auth(validationWrappers)(next).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK)

	// accounts
	path = "/accounts/self"

	r = httptest.NewRequest("GET", path, nil)
	r.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()

	validationWrapperMock.On("Matches", path).
		Return(true).
		Once()

	accountMock := config.NewAccountMock(t)

	validationWrapperMock.On("Validate", r).
		Run(func(args mock.Arguments) {
			req, ok := args.Get(0).(*http.Request)
			assert.True(t, ok)

			ctx := contextutil.WithAccount(req.Context(), accountMock)

			*req = *req.WithContext(ctx)
		}).
		Return(nil).
		Once()

	next = func(w http.ResponseWriter, r *http.Request) {
		acc, err := contextutil.Account(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, accountMock, acc)
		w.WriteHeader(http.StatusOK)
	}

	auth(validationWrappers)(next).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusOK)

	// Auth service failure
	path = "/v3/investors"

	r = httptest.NewRequest("GET", path, nil)
	r.Header.Set("Authorization", "Bearer "+token)

	w = httptest.NewRecorder()

	validationWrapperMock.On("Matches", path).
		Return(true).
		Once()

	validationWrapperMock.On("Validate", r).
		Return(errors.New("error")).
		Once()

	auth(validationWrappers)(nil).ServeHTTP(w, r)
	assert.Equal(t, w.Code, http.StatusForbidden)
}

func TestRouter(t *testing.T) {
	validationWrapperMock := access.NewValidationWrapperMock(t)
	validationWrapperFactoryMock := access.NewValidationWrapperFactoryMock(t)

	validationWrapperFactoryMock.On("GetValidationWrappers").
		Return(access.ValidationWrappers{validationWrapperMock}, nil).
		Once()

	cctx := map[string]interface{}{
		bootstrap.BootstrappedConfig:         config.NewConfigurationMock(t),
		config.BootstrappedConfigStorage:     config.NewServiceMock(t),
		httpV2.BootstrappedService:           &httpV2.Service{},
		httpV3.BootstrappedService:           &httpV3.Service{},
		BootstrappedValidationWrapperFactory: validationWrapperFactoryMock,
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
	assert.Len(t, r.Routes()[2].SubRoutes.Routes(), 7)
}
