// +build unit

package coreapi

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
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	testingnfts "github.com/centrifuge/go-centrifuge/testingutils/nfts"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_MintNFT(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/nfts/mint", b).WithContext(ctx)
	}

	// empty data
	h := handler{}
	ctx := context.Background()
	w, r := getHTTPReqAndResp(ctx, nil)
	h.MintNFT(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	data := map[string]interface{}{
		"document_id":      hexutil.Encode(utils.RandomSlice(32)),
		"registry_address": hexutil.Encode(utils.RandomSlice(20)),
		"deposit_address":  hexutil.Encode(utils.RandomSlice(20)),
	}

	d, err := json.Marshal(data)
	assert.NoError(t, err)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	srv := new(testingnfts.MockNFTService)
	srv.On("MintNFT", ctx, mock.Anything).Return(nil, nil, errors.New("failed to mint nft")).Once()
	h.srv.nftService = srv
	h.MintNFT(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "failed to mint nft")
	srv.AssertExpectations(t)

	// success
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	srv = new(testingnfts.MockNFTService)
	tokenID := hexutil.Encode(utils.RandomSlice(32))
	srv.On("MintNFT", ctx, mock.Anything).Return(
		&nft.TokenResponse{
			TokenID: tokenID,
			JobID:   jobs.NewJobID().String(),
		}, nil, nil).Once()
	h.srv.nftService = srv
	h.MintNFT(w, r)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), tokenID)
	srv.AssertExpectations(t)
}

func TestHandler_TransferNFT(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context, b io.Reader) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("POST", "/nfts/{token_id}/transfer", b).WithContext(ctx)
	}

	// empty token
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 1, 1)
	rctx.URLParams.Values = make([]string, 1, 1)
	rctx.URLParams.Keys[0] = "token_id"
	rctx.URLParams.Values[0] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx, nil)
	h := handler{}
	h.TransferNFT(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidTokenID.Error())

	// invalid token
	rctx.URLParams.Values[0] = "somevvalue"
	w, r = getHTTPReqAndResp(ctx, nil)
	h = handler{}
	h.TransferNFT(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidTokenID.Error())

	// empty body
	tokenID := hexutil.Encode(utils.RandomSlice(32))
	rctx.URLParams.Values[0] = tokenID
	w, r = getHTTPReqAndResp(ctx, nil)
	h = handler{}
	h.TransferNFT(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "unexpected end of JSON input")

	// invalid registry
	body := map[string]interface{}{
		"registry_address": "some registry",
	}
	d, err := json.Marshal(body)
	assert.NoError(t, err)
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h = handler{}
	h.TransferNFT(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "cannot unmarshal hex string")

	// service fail
	reg := hexutil.Encode(utils.RandomSlice(20))
	to := hexutil.Encode(utils.RandomSlice(20))
	body = map[string]interface{}{
		"registry_address": reg,
		"to":               to,
	}
	d, err = json.Marshal(body)
	assert.NoError(t, err)
	srv := new(testingnfts.MockNFTService)
	srv.On("TransferFrom", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, errors.New("failed to transfer")).Once()
	h.srv.nftService = srv
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.TransferNFT(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to transfer")
	srv.AssertExpectations(t)

	// success
	srv = new(testingnfts.MockNFTService)
	srv.On("TransferFrom", ctx, mock.Anything, mock.Anything, mock.Anything).Return(&nft.TokenResponse{
		TokenID: tokenID,
		JobID:   jobs.NewJobID().String(),
	}, nil, nil).Once()
	h.srv.nftService = srv
	w, r = getHTTPReqAndResp(ctx, bytes.NewReader(d))
	h.TransferNFT(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), tokenID)
	srv.AssertExpectations(t)
}

func TestHandler_OwnerOfNFT(t *testing.T) {
	getHTTPReqAndResp := func(ctx context.Context) (*httptest.ResponseRecorder, *http.Request) {
		return httptest.NewRecorder(), httptest.NewRequest("GET", "/nfts/{token_id}/registry/{registry_address}/owner", nil).WithContext(ctx)
	}

	// empty token
	h := handler{}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Keys = make([]string, 2, 2)
	rctx.URLParams.Values = make([]string, 2, 2)
	rctx.URLParams.Keys[0] = tokenIDParam
	rctx.URLParams.Values[0] = ""
	rctx.URLParams.Keys[1] = registryAddressParam
	rctx.URLParams.Values[1] = ""
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	w, r := getHTTPReqAndResp(ctx)
	h.OwnerOfNFT(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidTokenID.Error())

	// invalid token id
	rctx.URLParams.Values[0] = "some value"
	w, r = getHTTPReqAndResp(ctx)
	h.OwnerOfNFT(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidTokenID.Error())

	// empty registry address
	rctx.URLParams.Values[0] = hexutil.Encode(utils.RandomSlice(32))
	w, r = getHTTPReqAndResp(ctx)
	h.OwnerOfNFT(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidRegistryAddress.Error())

	// invalid registry address
	rctx.URLParams.Values[1] = "some value"
	w, r = getHTTPReqAndResp(ctx)
	h.OwnerOfNFT(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), ErrInvalidRegistryAddress.Error())

	// owner failed
	registry := common.BytesToAddress(utils.RandomSlice(20))
	rctx.URLParams.Values[1] = registry.String()
	srv := new(testingnfts.MockNFTService)
	srv.On("OwnerOf", mock.Anything, mock.Anything).Return(nil, errors.New("failed to get owner")).Once()
	h.srv.nftService = srv
	w, r = getHTTPReqAndResp(ctx)
	h.OwnerOfNFT(w, r)
	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Contains(t, w.Body.String(), "failed to get owner")
	srv.AssertExpectations(t)

	// success
	owner := common.BytesToAddress(utils.RandomSlice(20))
	srv = new(testingnfts.MockNFTService)
	srv.On("OwnerOf", mock.Anything, mock.Anything).Return(owner, nil).Once()
	h.srv.nftService = srv
	w, r = getHTTPReqAndResp(ctx)
	h.OwnerOfNFT(w, r)
	assert.Equal(t, w.Code, http.StatusOK)
	assert.Contains(t, w.Body.String(), strings.ToLower(owner.String()))
	srv.AssertExpectations(t)
}
