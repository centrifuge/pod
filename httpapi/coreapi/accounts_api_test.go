// +build unit

package coreapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestHandler_SignPayload(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/accounts/{account_id}/sign", b).WithContext(ctx)
	}
	// empty account_id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = accountIDParam
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.SignPayload(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrAccountIDInvalid.Error())

	// invalid account id
	rctx.URLParams.Values[0] = "invalid value"
	w, r = getHTTPReqAndResp(ctx, nil)
	h.SignPayload(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrAccountIDInvalid.Error())

	// empty body
	accountID := utils.RandomSlice(20)
	rctx.URLParams.Values[0] = hexutil.Encode(accountID)
	w, r = getHTTPReqAndResp(ctx, nil)
	h.SignPayload(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed signer
	payload := utils.RandomSlice(32)
	body := map[string]string{
		"payload": hexutil.Encode(payload),
	}
	d, err := json.Marshal(body)
	assert.NoError(t, err)
	srv := new(configstore.MockService)
	srv.On("Sign", accountID, payload).Return(nil, errors.New("failed to sign payload")).Once()
	h.srv.accountsSrv = srv
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.SignPayload(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to sign payload")
	srv.AssertExpectations(t)

	// success
	signature := utils.RandomSlice(32)
	pk := utils.RandomSlice(20)
	srv = new(configstore.MockService)
	srv.On("Sign", accountID, payload).Return(&coredocumentpb.Signature{
		SignerId:  accountID,
		Signature: signature,
		PublicKey: pk,
	}, nil).Once()
	h.srv.accountsSrv = srv
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.SignPayload(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), hexutil.Encode(payload))
	assert.Contains(t, w.Body.String(), hexutil.Encode(signature))
	assert.Contains(t, w.Body.String(), hexutil.Encode(pk))
	assert.Contains(t, w.Body.String(), hexutil.Encode(accountID))
	srv.AssertExpectations(t)
}

func TestHandler_GetAccount(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/accounts/{account_id}", nil).WithContext(ctx)
	}
	// empty account_id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = accountIDParam
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx)
	h := handler{}
	h.GetAccount(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrAccountIDInvalid.Error())

	// invalid account id
	rctx.URLParams.Values[0] = "invalid value"
	w, r = getHTTPReqAndResp(ctx)
	h.GetAccount(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrAccountIDInvalid.Error())

	// missing account
	accountID := utils.RandomSlice(20)
	rctx.URLParams.Values[0] = hexutil.Encode(accountID)
	srv := new(configstore.MockService)
	srv.On("GetAccount", accountID).Return(nil, errors.New("failed to get account")).Once()
	h.srv.accountsSrv = srv
	w, r = getHTTPReqAndResp(ctx)
	h.GetAccount(w, r)
	srv.AssertExpectations(t)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), ErrAccountNotFound.Error())

	// success
	cfg := new(testingconfig.MockConfig)
	cfg.On("GetEthereumAccount", "name").Return(&config.AccountConfig{}, nil).Once()
	cfg.On("GetEthereumDefaultAccountName").Return("dummyAcc").Once()
	cfg.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()
	cfg.On("GetIdentityID").Return(accountID, nil).Once()
	cfg.On("GetP2PKeyPair").Return("pub", "priv").Once()
	cfg.On("GetSigningKeyPair").Return("pub", "priv").Once()
	cfg.On("GetEthereumContextWaitTimeout").Return(time.Second).Once()
	cfg.On("GetPrecommitEnabled").Return(true).Once()
	acc, err := configstore.NewAccount("name", cfg)
	assert.NoError(t, err)
	srv = new(configstore.MockService)
	srv.On("GetAccount", accountID).Return(acc, nil).Once()
	h.srv.accountsSrv = srv
	w, r = getHTTPReqAndResp(ctx)
	h.GetAccount(w, r)
	srv.AssertExpectations(t)
	cfg.AssertExpectations(t)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), hexutil.Encode(accountID))
}

func TestHandler_GenerateAccount(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/accounts/generate", nil).WithContext(ctx)
	}

	// failed generation
	rctx := chi.NewRouteContext()
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}
	srv := new(configstore.MockService)
	srv.On("GenerateAccount").Return(nil, errors.New("failed to generate account")).Once()
	h.srv.accountsSrv = srv
	w, r := getHTTPReqAndResp(ctx)
	h.GenerateAccount(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to generate account")
	srv.AssertExpectations(t)

	// success
	accountID := utils.RandomSlice(20)
	cfg := new(testingconfig.MockConfig)
	cfg.On("GetEthereumAccount", "name").Return(&config.AccountConfig{}, nil).Once()
	cfg.On("GetEthereumDefaultAccountName").Return("dummyAcc").Once()
	cfg.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()
	cfg.On("GetIdentityID").Return(accountID, nil).Once()
	cfg.On("GetP2PKeyPair").Return("pub", "priv").Once()
	cfg.On("GetSigningKeyPair").Return("pub", "priv").Once()
	cfg.On("GetEthereumContextWaitTimeout").Return(time.Second).Once()
	cfg.On("GetPrecommitEnabled").Return(true).Once()
	acc, err := configstore.NewAccount("name", cfg)
	assert.NoError(t, err)
	srv = new(configstore.MockService)
	srv.On("GenerateAccount").Return(acc, nil).Once()
	h.srv.accountsSrv = srv
	w, r = getHTTPReqAndResp(ctx)
	h.GenerateAccount(w, r)
	srv.AssertExpectations(t)
	cfg.AssertExpectations(t)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), hexutil.Encode(accountID))
}

func TestHandler_GetAccounts(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/accounts", nil).WithContext(ctx)
	}

	// failed generation
	rctx := chi.NewRouteContext()
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}
	srv := new(configstore.MockService)
	srv.On("GetAccounts").Return(nil, errors.New("failed to get accounts")).Once()
	h.srv.accountsSrv = srv
	w, r := getHTTPReqAndResp(ctx)
	h.GetAccounts(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get accounts")
	srv.AssertExpectations(t)

	// success
	accountID := utils.RandomSlice(20)
	cfg := new(testingconfig.MockConfig)
	cfg.On("GetEthereumAccount", "name").Return(&config.AccountConfig{}, nil).Once()
	cfg.On("GetEthereumDefaultAccountName").Return("dummyAcc").Once()
	cfg.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()
	cfg.On("GetIdentityID").Return(accountID, nil).Once()
	cfg.On("GetP2PKeyPair").Return("pub", "priv").Once()
	cfg.On("GetSigningKeyPair").Return("pub", "priv").Once()
	cfg.On("GetEthereumContextWaitTimeout").Return(time.Second).Once()
	cfg.On("GetPrecommitEnabled").Return(true).Once()
	acc, err := configstore.NewAccount("name", cfg)
	assert.NoError(t, err)
	srv = new(configstore.MockService)
	srv.On("GetAccounts").Return([]config.Account{acc}, nil).Once()
	h.srv.accountsSrv = srv
	w, r = getHTTPReqAndResp(ctx)
	h.GetAccounts(w, r)
	srv.AssertExpectations(t)
	cfg.AssertExpectations(t)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), hexutil.Encode(accountID))
}
