// +build unit

package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/nft"
	testingnfts "github.com/centrifuge/go-centrifuge/testingutils/nfts"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_MintNFTOnCC(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/nfts/registries/{registry_address}/mint", b).WithContext(ctx)
	}

	// empty registry tests
	h := handler{}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = coreapi.RegistryAddressParam
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h.MintNFTOnCC(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidRegistryAddress.Error())

	//  invalid registry
	rctx.URLParams.Values[0] = "some invalid"
	w, r = getHTTPReqAndResp(ctx, nil)
	h.MintNFTOnCC(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidRegistryAddress.Error())

	// empty data
	rctx.URLParams.Values[0] = hexutil.Encode(utils.RandomSlice(20))
	w, r = getHTTPReqAndResp(ctx, nil)
	h.MintNFTOnCC(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")
	data := map[string]interface{}{
		"document_id":     hexutil.Encode(utils.RandomSlice(32)),
		"deposit_address": hexutil.Encode(utils.RandomSlice(32)),
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	srv := new(testingnfts.MockNFTService)
	srv.On("MintNFTOnCC", ctx, mock.Anything).Return(nil, errors.New("failed to mint nft")).Once()
	h.srv.nftSrv = srv
	h.MintNFTOnCC(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "failed to mint nft")
	srv.AssertExpectations(t)

	// success
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	srv = new(testingnfts.MockNFTService)
	tokenID := hexutil.Encode(utils.RandomSlice(32))
	srv.On("MintNFTOnCC", ctx, mock.Anything).Return(
		&nft.TokenResponse{
			TokenID: tokenID,
			JobID:   hexutil.Encode(utils.RandomSlice(32)),
		}, nil).Once()
	h.srv.nftSrv = srv
	h.MintNFTOnCC(w, r)
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Contains(t, w.Body.String(), tokenID)
	srv.AssertExpectations(t)
}

func TestHandler_TransferNFTOnCC(t *testing.T) {
	var b io.Reader
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/nfts/registries/{registry_address}/tokens/{token_id}/transfer", b).WithContext(ctx)
	}

	// empty token and registry tests
	h := handler{}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 2)
	rctx.URLParams.Values = make([]string, 2)
	rctx.URLParams.Keys[0] = coreapi.TokenIDParam
	rctx.URLParams.Values[0] = ""
	rctx.URLParams.Keys[1] = coreapi.RegistryAddressParam
	rctx.URLParams.Values[1] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	testTokenIDAndRegistryAddress(t, rctx, getHTTPReqAndResp, h.TransferNFT)

	// empty body
	tokenID, err := nft.TokenIDFromString(rctx.URLParams.Values[0])
	assert.NoError(t, err)
	w, r := getHTTPReqAndResp(ctx)
	h = handler{}
	h.TransferNFTOnCC(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// Service fail
	to := hexutil.Encode(utils.RandomSlice(32))
	body := map[string]interface{}{
		"to": to,
	}
	d, err := json.Marshal(body)
	assert.NoError(t, err)
	srv := new(testingnfts.MockNFTService)
	srv.On("TransferNFT", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("failed to transfer")).Once()
	h.srv.nftSrv = srv
	b = bytes.NewReader(d)
	w, r = getHTTPReqAndResp(ctx)
	h.TransferNFTOnCC(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to transfer")
	srv.AssertExpectations(t)

	// success
	srv = new(testingnfts.MockNFTService)
	srv.On("TransferNFT", ctx, mock.Anything, mock.Anything, mock.Anything).Return(&nft.TokenResponse{
		TokenID: tokenID.String(),
		JobID:   "",
	}, nil).Once()
	h.srv.nftSrv = srv
	b = bytes.NewReader(d)
	w, r = getHTTPReqAndResp(ctx)
	h.TransferNFTOnCC(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), tokenID.String())
	srv.AssertExpectations(t)
}

func TestHandler_OwnerOfNFTOnCC(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/nfts/registries/{registry_address}/tokens/{token_id}/owner", nil).WithContext(ctx)
	}

	// empty token and registry tests
	h := handler{}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 2)
	rctx.URLParams.Values = make([]string, 2)
	rctx.URLParams.Keys[0] = coreapi.TokenIDParam
	rctx.URLParams.Values[0] = ""
	rctx.URLParams.Keys[1] = coreapi.RegistryAddressParam
	rctx.URLParams.Values[1] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	testTokenIDAndRegistryAddress(t, rctx, getHTTPReqAndResp, h.OwnerOfNFTOnCC)

	// owner failed
	srv := new(testingnfts.MockNFTService)
	srv.On("OwnerOfOnCC", mock.Anything, mock.Anything).Return(nil, errors.New("failed to get owner")).Once()
	h.srv.nftSrv = srv
	w, r := getHTTPReqAndResp(ctx)
	h.OwnerOfNFTOnCC(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to get owner")
	srv.AssertExpectations(t)

	// success
	owner := types.NewAccountID(utils.RandomSlice(32))
	srv = new(testingnfts.MockNFTService)
	srv.On("OwnerOfOnCC", mock.Anything, mock.Anything).Return(owner, nil).Once()
	h.srv.nftSrv = srv
	w, r = getHTTPReqAndResp(ctx)
	h.OwnerOfNFTOnCC(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), strings.ToLower(hexutil.Encode(owner[:])))
	srv.AssertExpectations(t)
}
