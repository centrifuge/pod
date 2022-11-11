//go:build unit

package v3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"

	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
)

func TestHandler_CommitAndMintNFT(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)

	owner, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)

	payload := coreapi.MintNFTV3Request{
		DocumentID: documentID,
		Owner:      owner,
		IPFSMetadata: nftv3.IPFSMetadata{
			Name:        "test-name",
			Description: "test-description",
			Image:       "test-image",
			DocumentAttributeKeys: []string{
				"attr_key1",
				"attr_key2",
			},
		},
		GrantReadAccess: true,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/nfts/collections/%d/commit_and_mint", testServer.URL, collectionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	mintReq := coreapi.ToNFTMintRequestV3(payload, collectionID)

	itemID := types.NewU128(*big.NewInt(2222))

	jobID := hexutil.Encode(utils.RandomSlice(32))

	mintRes := &nftv3.MintNFTResponse{
		JobID:  jobID,
		ItemID: itemID,
	}

	nftServiceMock.On("MintNFT", mock.Anything, mintReq, true).
		Return(mintRes, nil).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var coreapiRes coreapi.MintNFTV3Response

	err = json.Unmarshal(resBody, &coreapiRes)
	assert.NoError(t, err)

	assert.Equal(t, jobID, coreapiRes.Header.JobID)
	assert.Equal(t, documentID, coreapiRes.DocumentID.Bytes())
	assert.Equal(t, collectionID, coreapiRes.CollectionID)
	assert.Equal(t, itemID.String(), coreapiRes.ItemID)
	assert.Equal(t, owner, coreapiRes.Owner)
	assert.Equal(t, payload.IPFSMetadata, coreapiRes.IPFSMetadata)
}

func TestHandler_CommitAndMintNFT_InvalidCollectionIDParam(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	testURL := fmt.Sprintf(
		"%s/nfts/collections/%s/commit_and_mint",
		testServer.URL,
		"invalid-collection-id-param",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_CommitAndMintNFT_InvalidPayload(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)

	testURL := fmt.Sprintf("%s/nfts/collections/%d/commit_and_mint", testServer.URL, collectionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(utils.RandomSlice(32)))
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_CommitAndMintNFT_NFTServiceError(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)

	owner, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)

	payload := coreapi.MintNFTV3Request{
		DocumentID: documentID,
		Owner:      owner,
		IPFSMetadata: nftv3.IPFSMetadata{
			Name:        "test-name",
			Description: "test-description",
			Image:       "test-image",
			DocumentAttributeKeys: []string{
				"attr_key1",
				"attr_key2",
			},
		},
		GrantReadAccess: true,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/nfts/collections/%d/commit_and_mint", testServer.URL, collectionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	mintReq := coreapi.ToNFTMintRequestV3(payload, collectionID)

	nftServiceMock.On("MintNFT", mock.Anything, mintReq, true).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_MintNFT(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)

	owner, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)

	payload := coreapi.MintNFTV3Request{
		DocumentID: documentID,
		Owner:      owner,
		IPFSMetadata: nftv3.IPFSMetadata{
			Name:        "test-name",
			Description: "test-description",
			Image:       "test-image",
			DocumentAttributeKeys: []string{
				"attr_key1",
				"attr_key2",
			},
		},
		GrantReadAccess: true,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/nfts/collections/%d/mint", testServer.URL, collectionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	mintReq := coreapi.ToNFTMintRequestV3(payload, collectionID)

	itemID := types.NewU128(*big.NewInt(2222))

	jobID := hexutil.Encode(utils.RandomSlice(32))

	mintRes := &nftv3.MintNFTResponse{
		JobID:  jobID,
		ItemID: itemID,
	}

	nftServiceMock.On("MintNFT", mock.Anything, mintReq, false).
		Return(mintRes, nil).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var coreapiRes coreapi.MintNFTV3Response

	err = json.Unmarshal(resBody, &coreapiRes)
	assert.NoError(t, err)

	assert.Equal(t, jobID, coreapiRes.Header.JobID)
	assert.Equal(t, documentID, coreapiRes.DocumentID.Bytes())
	assert.Equal(t, collectionID, coreapiRes.CollectionID)
	assert.Equal(t, itemID.String(), coreapiRes.ItemID)
	assert.Equal(t, owner, coreapiRes.Owner)
	assert.Equal(t, payload.IPFSMetadata, coreapiRes.IPFSMetadata)
}

func TestHandler_MintNFT_InvalidCollectionIDParam(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	testURL := fmt.Sprintf(
		"%s/nfts/collections/%s/mint",
		testServer.URL,
		"invalid-collection-id-param",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_MintNFT_InvalidPayload(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)

	testURL := fmt.Sprintf("%s/nfts/collections/%d/mint", testServer.URL, collectionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(utils.RandomSlice(32)))
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_MintNFT_NFTServiceError(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)

	owner, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)

	payload := coreapi.MintNFTV3Request{
		DocumentID: documentID,
		Owner:      owner,
		IPFSMetadata: nftv3.IPFSMetadata{
			Name:        "test-name",
			Description: "test-description",
			Image:       "test-image",
			DocumentAttributeKeys: []string{
				"attr_key1",
				"attr_key2",
			},
		},
		GrantReadAccess: true,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/nfts/collections/%d/mint", testServer.URL, collectionID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	mintReq := coreapi.ToNFTMintRequestV3(payload, collectionID)

	nftServiceMock.On("MintNFT", mock.Anything, mintReq, false).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetNFTOwner(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))

	testURL := fmt.Sprintf("%s/nfts/collections/%d/items/%s/owner", testServer.URL, collectionID, itemID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	nftServiceMock.On("GetNFTOwner", collectionID, itemID).
		Return(accountID, nil).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var coreapiRes coreapi.GetNFTOwnerV3Response

	err = json.Unmarshal(resBody, &coreapiRes)
	assert.NoError(t, err)

	assert.Equal(t, collectionID, coreapiRes.CollectionID)
	assert.Equal(t, itemID.String(), coreapiRes.ItemID)
	assert.Equal(t, accountID, coreapiRes.Owner)
}

func TestHandler_GetNFTOwner_InvalidCollectionIDParam(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := "invalid-collection-id-param"
	itemID := types.NewU128(*big.NewInt(2222))

	testURL := fmt.Sprintf("%s/nfts/collections/%s/items/%s/owner", testServer.URL, collectionID, itemID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetNFTOwner_InvalidItemIDParam(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := "invalid-item-id-param"

	testURL := fmt.Sprintf("%s/nfts/collections/%d/items/%s/owner", testServer.URL, collectionID, itemID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetNFTOwner_NFTService_GenericError(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))

	testURL := fmt.Sprintf("%s/nfts/collections/%d/items/%s/owner", testServer.URL, collectionID, itemID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	nftServiceMock.On("GetNFTOwner", collectionID, itemID).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetNFTOwner_NFTService_NotFoundError(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))

	testURL := fmt.Sprintf("%s/nfts/collections/%d/items/%s/owner", testServer.URL, collectionID, itemID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	nftServiceMock.On("GetNFTOwner", collectionID, itemID).
		Return(nil, nftv3.ErrOwnerNotFound).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_CreateNFTCollection(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)

	payload := coreapi.CreateNFTCollectionV3Request{
		CollectionID: collectionID,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/nfts/collections", testServer.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	jobID := hexutil.Encode(utils.RandomSlice(32))

	srvRes := &nftv3.CreateNFTCollectionResponse{
		JobID:        jobID,
		CollectionID: collectionID,
	}

	nftServiceMock.On("CreateNFTCollection", mock.Anything, collectionID).
		Return(srvRes, nil).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var coreapiRes coreapi.CreateNFTCollectionV3Response

	err = json.Unmarshal(resBody, &coreapiRes)
	assert.NoError(t, err)

	assert.Equal(t, jobID, coreapiRes.Header.JobID)
	assert.Equal(t, collectionID, coreapiRes.CollectionID)
}

func TestHandler_CreateNFTCollection_InvalidPayload(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	testURL := fmt.Sprintf("%s/nfts/collections", testServer.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(utils.RandomSlice(32)))
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_CreateNFTCollection_NFTServiceError(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)

	payload := coreapi.CreateNFTCollectionV3Request{
		CollectionID: collectionID,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/nfts/collections", testServer.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	nftServiceMock.On("CreateNFTCollection", mock.Anything, collectionID).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_MetadataOfNFT(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))

	testURL := fmt.Sprintf("%s/nfts/collections/%d/items/%s/metadata", testServer.URL, collectionID, itemID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	deposit := types.NewU128(*big.NewInt(1234))
	metadataData := utils.RandomSlice(32)

	itemMetadata := &types.ItemMetadata{
		Deposit:  deposit,
		Data:     metadataData,
		IsFrozen: false,
	}

	nftServiceMock.On("GetItemMetadata", collectionID, itemID).
		Return(itemMetadata, nil).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var coreapiRes coreapi.ItemMetadataOfNFTV3Response

	err = json.Unmarshal(resBody, &coreapiRes)
	assert.NoError(t, err)

	assert.Equal(t, itemMetadata.Deposit.String(), coreapiRes.Deposit)
	assert.Equal(t, []byte(itemMetadata.Data), coreapiRes.Data.Bytes())
	assert.Equal(t, itemMetadata.IsFrozen, coreapiRes.IsFrozen)
}

func TestHandler_MetadataOfNFT_InvalidCollectionIDParam(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := "invalid-collection-id-param"
	itemID := types.NewU128(*big.NewInt(2222))

	testURL := fmt.Sprintf("%s/nfts/collections/%s/items/%s/metadata", testServer.URL, collectionID, itemID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_MetadataOfNFT_InvalidItemIDParam(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := "invalid-item-id-param"

	testURL := fmt.Sprintf("%s/nfts/collections/%d/items/%s/metadata", testServer.URL, collectionID, itemID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_MetadataOfNFT_NFTService_GenericError(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))

	testURL := fmt.Sprintf("%s/nfts/collections/%d/items/%s/metadata", testServer.URL, collectionID, itemID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	nftServiceMock.On("GetItemMetadata", collectionID, itemID).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_MetadataOfNFT_NFTService_NotFoundError(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))

	testURL := fmt.Sprintf("%s/nfts/collections/%d/items/%s/metadata", testServer.URL, collectionID, itemID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	nftServiceMock.On("GetItemMetadata", collectionID, itemID).
		Return(nil, nftv3.ErrItemMetadataNotFound).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_AttributeOfNFT(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))
	attributeName := "attr-name"

	testURL := fmt.Sprintf(
		"%s/nfts/collections/%d/items/%s/attribute/%s",
		testServer.URL,
		collectionID,
		itemID.String(),
		attributeName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	itemAttribute := utils.RandomSlice(32)

	nftServiceMock.On("GetItemAttribute", collectionID, itemID, attributeName).
		Return(itemAttribute, nil).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var coreapiRes coreapi.ItemAttributeOfNFTV3Response

	err = json.Unmarshal(resBody, &coreapiRes)
	assert.NoError(t, err)

	assert.Equal(t, itemAttribute, coreapiRes.Value.Bytes())
}

func TestHandler_AttributeOfNFT_InvalidCollectionIDParam(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := "invalid-collection-id-param"
	itemID := types.NewU128(*big.NewInt(2222))
	attributeName := "attr-name"

	testURL := fmt.Sprintf(
		"%s/nfts/collections/%s/items/%s/attribute/%s",
		testServer.URL,
		collectionID,
		itemID.String(),
		attributeName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AttributeOfNFT_InvalidItemIDParam(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := "invalid-item-id-param"
	attributeName := "attr-name"

	testURL := fmt.Sprintf(
		"%s/nfts/collections/%d/items/%s/attribute/%s",
		testServer.URL,
		collectionID,
		itemID,
		attributeName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AttributeOfNFT_NFTService_GenericError(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))
	attributeName := "attr-name"

	testURL := fmt.Sprintf(
		"%s/nfts/collections/%d/items/%s/attribute/%s",
		testServer.URL,
		collectionID,
		itemID.String(),
		attributeName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	nftServiceMock.On("GetItemAttribute", collectionID, itemID, attributeName).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AttributeOfNFT_NFTService_NotFoundError(t *testing.T) {
	nftServiceMock := nftv3.NewServiceMock(t)

	service := &Service{nftServiceMock}

	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))
	attributeName := "attr-name"

	testURL := fmt.Sprintf(
		"%s/nfts/collections/%d/items/%s/attribute/%s",
		testServer.URL,
		collectionID,
		itemID.String(),
		attributeName,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	nftServiceMock.On("GetItemAttribute", collectionID, itemID, attributeName).
		Return(nil, nftv3.ErrItemAttributeNotFound).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
