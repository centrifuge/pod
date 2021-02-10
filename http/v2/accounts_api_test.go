// +build unit

package v2

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
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	signingPub = "../../build/resources/signingKey.pub.pem"
	p2pPub     = "../../build/resources/p2pKey.pub.pem"
)

func TestHandler_GenerateAccount(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, body io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/accounts/generate", body).WithContext(ctx)
	}

	// empty body
	rctx := chi.NewRouteContext()
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	h := handler{}
	w, r := getHTTPReqAndResp(ctx, nil)
	h.GenerateAccount(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// failed generation
	data := map[string]interface{}{
		"centrifuge_chain_account": map[string]string{
			"id":            "0xc81ebbec0559a6acf184535eb19da51ed3ed8c4ac65323999482aaf9b6696e27",
			"secret":        "0xc166b100911b1e9f780bb66d13badf2c1edbe94a1220f1a0584c09490158be31",
			"ss_58_address": "5Gb6Zfe8K8NSKrkFLCgqs8LUdk7wKweXM5pN296jVqDpdziR",
		},
	}
	d, err := json.Marshal(data)
	assert.NoError(t, err)
	srv := new(configstore.MockService)
	srv.On("GenerateAccountAsync", mock.Anything).Return(nil, nil, errors.New("failed to generate account")).Once()
	h.srv.accountSrv = srv
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.GenerateAccount(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to generate account")
	srv.AssertExpectations(t)

	// success
	did := utils.RandomSlice(20)
	jobID := utils.RandomSlice(32)
	srv.On("GenerateAccountAsync", mock.Anything).Return(did, jobID, nil).Once()
	h.srv.accountSrv = srv
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.GenerateAccount(w, r)
	srv.AssertExpectations(t)
	assert.Equal(t, w.Code, http.StatusCreated)
	assert.Contains(t, w.Body.String(), hexutil.Encode(did))
	assert.Contains(t, w.Body.String(), hexutil.Encode(jobID))
}

func TestHandler_SignPayload(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/accounts/{account_id}/sign", b).WithContext(ctx)
	}
	// empty account_id
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1)
	rctx.URLParams.Values = make([]string, 1)
	rctx.URLParams.Keys[0] = coreapi.AccountIDParam
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.SignPayload(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrAccountIDInvalid.Error())

	// invalid account id
	rctx.URLParams.Values[0] = "invalid value"
	w, r = getHTTPReqAndResp(ctx, nil)
	h.SignPayload(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrAccountIDInvalid.Error())

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
	h.srv.accountSrv = srv
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
	h.srv.accountSrv = srv
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
	rctx.URLParams.Keys[0] = coreapi.AccountIDParam
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx)
	h := handler{}
	h.GetAccount(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrAccountIDInvalid.Error())

	// invalid account id
	rctx.URLParams.Values[0] = "invalid value"
	w, r = getHTTPReqAndResp(ctx)
	h.GetAccount(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), coreapi.ErrAccountIDInvalid.Error())

	// missing account
	accountID := utils.RandomSlice(20)
	rctx.URLParams.Values[0] = hexutil.Encode(accountID)
	srv := new(configstore.MockService)
	srv.On("GetAccount", accountID).Return(nil, errors.New("failed to get account")).Once()
	h.srv.accountSrv = srv
	w, r = getHTTPReqAndResp(ctx)
	h.GetAccount(w, r)
	srv.AssertExpectations(t)
	assert.Equal(t, w.Code, http.StatusNotFound)
	assert.Contains(t, w.Body.String(), coreapi.ErrAccountNotFound.Error())

	// missing path
	cfg := new(testingconfig.MockConfig)
	cfg.On("GetEthereumAccount", "name").Return(&config.AccountConfig{}, nil).Twice()
	cfg.On("GetEthereumDefaultAccountName").Return("dummyAcc").Twice()
	cfg.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Twice()
	cfg.On("GetIdentityID").Return(accountID, nil).Twice()
	cfg.On("GetP2PKeyPair").Return("p2p pub", "priv").Once()
	cfg.On("GetSigningKeyPair").Return(signingPub, "priv").Twice()
	cfg.On("GetEthereumContextWaitTimeout").Return(time.Second).Twice()
	cfg.On("GetPrecommitEnabled").Return(true).Twice()
	cfg.On("GetCentChainAccount").Return(config.CentChainAccount{}, nil).Twice()
	acc, err := configstore.NewAccount("name", cfg)
	assert.NoError(t, err)
	srv = new(configstore.MockService)
	srv.On("GetAccount", accountID).Return(acc, nil).Once()
	h.srv.accountSrv = srv
	w, r = getHTTPReqAndResp(ctx)
	h.GetAccount(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)

	// success
	cfg.On("GetP2PKeyPair").Return(p2pPub, "priv").Once()
	acc, err = configstore.NewAccount("name", cfg)
	assert.NoError(t, err)
	srv.On("GetAccount", accountID).Return(acc, nil).Once()
	h.srv.accountSrv = srv
	w, r = getHTTPReqAndResp(ctx)
	h.GetAccount(w, r)
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
	h.srv.accountSrv = srv
	w, r := getHTTPReqAndResp(ctx)
	h.GetAccounts(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)
	assert.Contains(t, w.Body.String(), "failed to get accounts")
	srv.AssertExpectations(t)

	// success
	accountID := utils.RandomSlice(20)
	// missing path
	cfg := new(testingconfig.MockConfig)
	cfg.On("GetEthereumAccount", "name").Return(&config.AccountConfig{}, nil).Twice()
	cfg.On("GetEthereumDefaultAccountName").Return("dummyAcc").Twice()
	cfg.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Twice()
	cfg.On("GetIdentityID").Return(accountID, nil).Twice()
	cfg.On("GetP2PKeyPair").Return("p2p pub", "priv").Once()
	cfg.On("GetSigningKeyPair").Return(signingPub, "priv").Twice()
	cfg.On("GetEthereumContextWaitTimeout").Return(time.Second).Twice()
	cfg.On("GetPrecommitEnabled").Return(true).Twice()
	cfg.On("GetCentChainAccount").Return(config.CentChainAccount{}, nil).Twice()
	acc, err := configstore.NewAccount("name", cfg)
	assert.NoError(t, err)
	srv = new(configstore.MockService)
	srv.On("GetAccounts").Return([]config.Account{acc}, nil).Once()
	h.srv.accountSrv = srv
	w, r = getHTTPReqAndResp(ctx)
	h.GetAccounts(w, r)
	assert.Equal(t, w.Code, http.StatusInternalServerError)

	// success
	cfg.On("GetP2PKeyPair").Return(p2pPub, "priv").Once()
	acc, err = configstore.NewAccount("name", cfg)
	assert.NoError(t, err)
	srv.On("GetAccounts").Return([]config.Account{acc}, nil).Once()
	h.srv.accountSrv = srv
	w, r = getHTTPReqAndResp(ctx)
	h.GetAccounts(w, r)
	srv.AssertExpectations(t)
	cfg.AssertExpectations(t)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), hexutil.Encode(accountID))
}
