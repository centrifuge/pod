//go:build unit

package v3

import (
	"context"
	"math/big"
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pallets/uniques"
	"github.com/centrifuge/go-centrifuge/pending"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	mockUtils "github.com/centrifuge/go-centrifuge/testingutils/mocks"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_MintNFT_NonPendingDocument(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1111)
	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ipfsMetadata := IPFSMetadata{
		Name:        "test-name",
		Description: "test-desc",
		Image:       "test-image",
		DocumentAttributeKeys: []string{
			hexutil.Encode(utils.RandomSlice(32)),
		},
	}

	req := &MintNFTRequest{
		DocumentID:      []byte("document_id"),
		CollectionID:    collectionID,
		Owner:           ownerAccountID,
		IPFSMetadata:    ipfsMetadata,
		GrantReadAccess: true,
	}

	doc := documents.NewDocumentMock(t)

	mockUtils.GetMock[*documents.ServiceMock](mocks).
		On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	itemID1 := types.NewU128(*big.NewInt(2222))

	collectionID2 := types.U64(1234)
	itemID2 := types.NewU128(*big.NewInt(5678))

	encodedCollectionID1, err := codec.Encode(collectionID)
	assert.NoError(t, err, "expected no error")
	encodedItemID1, err := codec.Encode(itemID1)
	assert.NoError(t, err, "expected no error")

	encodedCollectionID2, err := codec.Encode(collectionID2)
	assert.NoError(t, err, "expected no error")
	encodedItemID2, err := codec.Encode(itemID2)
	assert.NoError(t, err, "expected no error")

	nfts := []*coredocumentpb.NFT{
		{
			CollectionId: encodedCollectionID1,
			ItemId:       encodedItemID1,
		},
		{
			CollectionId: encodedCollectionID2,
			ItemId:       encodedItemID2,
		},
	}

	doc.On("NFTs").
		Return(nfts)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, itemID1).
		Return(nil, uniques.ErrItemDetailsNotFound)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, mock.Anything).
		Return(nil, uniques.ErrItemDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	mockUtils.GetMock[*jobs.DispatcherMock](mocks).
		On("Dispatch", accountID, mock.IsType(&gocelery.Job{})).
		Run(func(args mock.Arguments) {
			job, ok := args.Get(1).(*gocelery.Job)
			assert.True(t, ok)

			assert.Equal(t, "Mint NFT on Centrifuge Chain", job.Desc)
			assert.Equal(t, mintNFTV3Job, job.Runner)
			assert.Equal(t, "add_nft_v3_to_document", job.Tasks[0].RunnerFunc)
		}).
		Return(resultMock, nil)

	res, err := service.MintNFT(ctx, req, false)
	assert.NoError(t, err, "expected no error")
	assert.IsType(t, &MintNFTResponse{}, res, "types should match")
}

func TestService_MintNFT_PendingDocument(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1111)
	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ipfsMetadata := IPFSMetadata{
		Name:        "test-name",
		Description: "test-desc",
		Image:       "test-image",
		DocumentAttributeKeys: []string{
			hexutil.Encode(utils.RandomSlice(32)),
		},
	}

	req := &MintNFTRequest{
		DocumentID:      []byte("document_id"),
		CollectionID:    collectionID,
		Owner:           ownerAccountID,
		IPFSMetadata:    ipfsMetadata,
		GrantReadAccess: true,
	}

	doc := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).
		On("Get", ctx, req.DocumentID, documents.Pending).
		Return(doc, nil)

	itemID1 := types.NewU128(*big.NewInt(2222))

	collectionID2 := types.U64(1234)
	itemID2 := types.NewU128(*big.NewInt(5678))

	encodedCollectionID1, err := codec.Encode(collectionID)
	assert.NoError(t, err, "expected no error")
	encodedItemID1, err := codec.Encode(itemID1)
	assert.NoError(t, err, "expected no error")

	encodedCollectionID2, err := codec.Encode(collectionID2)
	assert.NoError(t, err, "expected no error")
	encodedItemID2, err := codec.Encode(itemID2)
	assert.NoError(t, err, "expected no error")

	nfts := []*coredocumentpb.NFT{
		{
			CollectionId: encodedCollectionID1,
			ItemId:       encodedItemID1,
		},
		{
			CollectionId: encodedCollectionID2,
			ItemId:       encodedItemID2,
		},
	}

	doc.On("NFTs").
		Return(nfts)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, itemID1).
		Return(nil, uniques.ErrItemDetailsNotFound)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, mock.Anything).
		Return(nil, uniques.ErrItemDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	mockUtils.GetMock[*jobs.DispatcherMock](mocks).
		On("Dispatch", accountID, mock.IsType(&gocelery.Job{})).
		Run(func(args mock.Arguments) {
			job, ok := args.Get(1).(*gocelery.Job)
			assert.True(t, ok)

			assert.Equal(t, "Commit pending document and mint NFT on Centrifuge Chain", job.Desc)
			assert.Equal(t, commitAndMintNFTV3Job, job.Runner)
			assert.Equal(t, "commit_pending_document", job.Tasks[0].RunnerFunc)
		}).
		Return(resultMock, nil)

	res, err := service.MintNFT(ctx, req, true)
	assert.NoError(t, err, "expected no error")
	assert.IsType(t, &MintNFTResponse{}, res, "types should match")
}

func TestService_MintNFT_InvalidRequests(t *testing.T) {
	invalidRequests := []*MintNFTRequest{
		nil,
		{
			DocumentID:   nil,
			CollectionID: types.U64(1234),
		},
		{
			DocumentID:   []byte("doc-id"),
			CollectionID: types.U64(0),
		},
	}

	service, _ := getServiceMocks(t)

	for _, invalidRequest := range invalidRequests {
		res, err := service.MintNFT(context.Background(), invalidRequest, false)

		assert.True(t, errors.IsOfType(errors.ErrRequestInvalid, err))
		assert.Nil(t, res)
	}
}

func TestService_MintNFT_NoNFTsPresent(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1111)
	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ipfsMetadata := IPFSMetadata{
		Name:        "test-name",
		Description: "test-desc",
		Image:       "test-image",
		DocumentAttributeKeys: []string{
			hexutil.Encode(utils.RandomSlice(32)),
		},
	}

	req := &MintNFTRequest{
		DocumentID:      []byte("document_id"),
		CollectionID:    collectionID,
		Owner:           ownerAccountID,
		IPFSMetadata:    ipfsMetadata,
		GrantReadAccess: true,
	}

	doc := documents.NewDocumentMock(t)

	mockUtils.GetMock[*documents.ServiceMock](mocks).
		On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("NFTs").
		Return(nil)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, mock.Anything).
		Return(nil, uniques.ErrItemDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	mockUtils.GetMock[*jobs.DispatcherMock](mocks).
		On("Dispatch", accountID, mock.IsType(&gocelery.Job{})).
		Return(resultMock, nil)

	res, err := service.MintNFT(ctx, req, false)
	assert.NoError(t, err, "expected no error")
	assert.IsType(t, &MintNFTResponse{}, res, "types should match")
}

func TestService_MintNFT_AccountError(t *testing.T) {
	service, _ := getServiceMocks(t)

	collectionID := types.U64(1111)
	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ipfsMetadata := IPFSMetadata{
		Name:        "test-name",
		Description: "test-desc",
		Image:       "test-image",
		DocumentAttributeKeys: []string{
			hexutil.Encode(utils.RandomSlice(32)),
		},
	}

	req := &MintNFTRequest{
		DocumentID:      []byte("document_id"),
		CollectionID:    collectionID,
		Owner:           ownerAccountID,
		IPFSMetadata:    ipfsMetadata,
		GrantReadAccess: true,
	}

	res, err := service.MintNFT(context.Background(), req, false)
	assert.ErrorIs(t, err, errors.ErrContextAccountRetrieval, "errors should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_NonPendingDocument_DocServiceError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountMock := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1111)
	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ipfsMetadata := IPFSMetadata{
		Name:        "test-name",
		Description: "test-desc",
		Image:       "test-image",
		DocumentAttributeKeys: []string{
			hexutil.Encode(utils.RandomSlice(32)),
		},
	}

	req := &MintNFTRequest{
		DocumentID:      []byte("document_id"),
		CollectionID:    collectionID,
		Owner:           ownerAccountID,
		IPFSMetadata:    ipfsMetadata,
		GrantReadAccess: true,
	}

	mockUtils.GetMock[*documents.ServiceMock](mocks).
		On("GetCurrentVersion", ctx, req.DocumentID).
		Return(nil, errors.New("document error"))

	res, err := service.MintNFT(ctx, req, false)
	assert.ErrorIs(t, err, ErrDocumentRetrieval, "errors should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_PendingDocument_PendingDocServiceError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountMock := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1111)
	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ipfsMetadata := IPFSMetadata{
		Name:        "test-name",
		Description: "test-desc",
		Image:       "test-image",
		DocumentAttributeKeys: []string{
			hexutil.Encode(utils.RandomSlice(32)),
		},
	}

	req := &MintNFTRequest{
		DocumentID:      []byte("document_id"),
		CollectionID:    collectionID,
		Owner:           ownerAccountID,
		IPFSMetadata:    ipfsMetadata,
		GrantReadAccess: true,
	}

	mockUtils.GetMock[*pending.ServiceMock](mocks).
		On("Get", ctx, req.DocumentID, documents.Pending).
		Return(nil, errors.New("document error"))

	res, err := service.MintNFT(ctx, req, true)
	assert.ErrorIs(t, err, ErrDocumentRetrieval, "errors should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_InstanceAlreadyMinted(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountMock := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1111)
	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ipfsMetadata := IPFSMetadata{
		Name:        "test-name",
		Description: "test-desc",
		Image:       "test-image",
		DocumentAttributeKeys: []string{
			hexutil.Encode(utils.RandomSlice(32)),
		},
	}

	req := &MintNFTRequest{
		DocumentID:      []byte("document_id"),
		CollectionID:    collectionID,
		Owner:           ownerAccountID,
		IPFSMetadata:    ipfsMetadata,
		GrantReadAccess: true,
	}

	doc := documents.NewDocumentMock(t)

	mockUtils.GetMock[*documents.ServiceMock](mocks).
		On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	itemID1 := types.NewU128(*big.NewInt(2222))

	collectionID2 := types.U64(1234)
	itemID2 := types.NewU128(*big.NewInt(5678))

	encodedCollectionID1, err := codec.Encode(collectionID)
	assert.NoError(t, err, "expected no error")
	encodedItemID1, err := codec.Encode(itemID1)
	assert.NoError(t, err, "expected no error")

	encodedCollectionID2, err := codec.Encode(collectionID2)
	assert.NoError(t, err, "expected no error")
	encodedItemID2, err := codec.Encode(itemID2)
	assert.NoError(t, err, "expected no error")

	nfts := []*coredocumentpb.NFT{
		{
			CollectionId: encodedCollectionID1,
			ItemId:       encodedItemID1,
		},
		{
			CollectionId: encodedCollectionID2,
			ItemId:       encodedItemID2,
		},
	}

	doc.On("NFTs").
		Return(nfts)

	doc.On("CurrentVersion").
		Return(utils.RandomSlice(32))

	instanceDetails := &types.ItemDetails{}

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, itemID1).
		Return(instanceDetails, nil)

	res, err := service.MintNFT(ctx, req, false)
	assert.ErrorIs(t, err, ErrItemAlreadyMinted, "errors should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_ItemIDGeneration_ContextError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountMock := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	ctx, cancel := context.WithCancel(ctx)

	collectionID := types.U64(1111)
	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ipfsMetadata := IPFSMetadata{
		Name:        "test-name",
		Description: "test-desc",
		Image:       "test-image",
		DocumentAttributeKeys: []string{
			hexutil.Encode(utils.RandomSlice(32)),
		},
	}

	req := &MintNFTRequest{
		DocumentID:      []byte("document_id"),
		CollectionID:    collectionID,
		Owner:           ownerAccountID,
		IPFSMetadata:    ipfsMetadata,
		GrantReadAccess: true,
	}

	doc := documents.NewDocumentMock(t)

	mockUtils.GetMock[*documents.ServiceMock](mocks).
		On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("NFTs").
		Return(nil)

	instanceDetails := &types.ItemDetails{}

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, mock.Anything).
		Return(instanceDetails, nil)

	go func() {
		time.Sleep(3 * time.Second)
		cancel()
	}()

	res, err := service.MintNFT(ctx, req, false)

	assert.ErrorIs(t, err, ErrItemIDGeneration, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_ItemIDGeneration_ItemDetailsError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountMock := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1111)
	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ipfsMetadata := IPFSMetadata{
		Name:        "test-name",
		Description: "test-desc",
		Image:       "test-image",
		DocumentAttributeKeys: []string{
			hexutil.Encode(utils.RandomSlice(32)),
		},
	}

	req := &MintNFTRequest{
		DocumentID:      []byte("document_id"),
		CollectionID:    collectionID,
		Owner:           ownerAccountID,
		IPFSMetadata:    ipfsMetadata,
		GrantReadAccess: true,
	}

	doc := documents.NewDocumentMock(t)

	mockUtils.GetMock[*documents.ServiceMock](mocks).
		On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("NFTs").
		Return(nil)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, mock.Anything).
		Return(nil, errors.New("instance details error"))

	res, err := service.MintNFT(ctx, req, false)

	assert.ErrorIs(t, err, ErrItemIDGeneration, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_DispatchError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1111)
	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ipfsMetadata := IPFSMetadata{
		Name:        "test-name",
		Description: "test-desc",
		Image:       "test-image",
		DocumentAttributeKeys: []string{
			hexutil.Encode(utils.RandomSlice(32)),
		},
	}

	req := &MintNFTRequest{
		DocumentID:      []byte("document_id"),
		CollectionID:    collectionID,
		Owner:           ownerAccountID,
		IPFSMetadata:    ipfsMetadata,
		GrantReadAccess: true,
	}

	doc := documents.NewDocumentMock(t)

	mockUtils.GetMock[*documents.ServiceMock](mocks).
		On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	itemID1 := types.NewU128(*big.NewInt(2222))

	collectionID2 := types.U64(1234)
	itemID2 := types.NewU128(*big.NewInt(5678))

	encodedCollectionID1, err := codec.Encode(collectionID)
	assert.NoError(t, err, "expected no error")
	encodedItemID1, err := codec.Encode(itemID1)
	assert.NoError(t, err, "expected no error")

	encodedCollectionID2, err := codec.Encode(collectionID2)
	assert.NoError(t, err, "expected no error")
	encodedItemID2, err := codec.Encode(itemID2)
	assert.NoError(t, err, "expected no error")

	nfts := []*coredocumentpb.NFT{
		{
			CollectionId: encodedCollectionID1,
			ItemId:       encodedItemID1,
		},
		{
			CollectionId: encodedCollectionID2,
			ItemId:       encodedItemID2,
		},
	}

	doc.On("NFTs").
		Return(nfts)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, itemID1).
		Return(nil, uniques.ErrItemDetailsNotFound)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, mock.Anything).
		Return(nil, uniques.ErrItemDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	mockUtils.GetMock[*jobs.DispatcherMock](mocks).
		On("Dispatch", accountID, mock.IsType(&gocelery.Job{})).
		Return(resultMock, errors.New("dispatch error"))

	res, err := service.MintNFT(ctx, req, false)
	assert.ErrorIs(t, err, ErrMintJobDispatch, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_GetNFTOwner(t *testing.T) {
	service, mocks := getServiceMocks(t)

	ctx := context.Background()

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	req := &GetNFTOwnerRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	}

	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	instanceDetails := &types.ItemDetails{
		Owner: *ownerAccountID,
	}

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, itemID).
		Return(instanceDetails, nil)

	res, err := service.GetNFTOwner(ctx, req)
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, collectionID, res.CollectionID, "class IDs should be equal")
	assert.Equal(t, itemID, res.ItemID, "instance IDs should be equal")
	assert.Equal(t, ownerAccountID, res.AccountID, "account IDs should be equal")
}

func TestService_GetNFTOwner_InvalidRequests(t *testing.T) {
	invalidRequests := []*GetNFTOwnerRequest{
		nil,
		{
			CollectionID: types.U64(0),
			ItemID:       types.NewU128(*big.NewInt(5678)),
		},
		{
			CollectionID: types.U64(1234),
			ItemID:       types.NewU128(*big.NewInt(0)),
		},
	}

	service, _ := getServiceMocks(t)

	for _, invalidRequest := range invalidRequests {
		res, err := service.GetNFTOwner(context.Background(), invalidRequest)

		assert.True(t, errors.IsOfType(errors.ErrRequestInvalid, err))
		assert.Nil(t, res, "expected no response")
	}
}

func TestService_GetNFTOwner_ItemDetailsError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	ctx := context.Background()

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	req := &GetNFTOwnerRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	}

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, itemID).
		Return(nil, errors.New("instance details error"))

	res, err := service.GetNFTOwner(ctx, req)
	assert.ErrorIs(t, err, ErrOwnerRetrieval, "error should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_GetNFTOwner_ItemDetailsNotFound(t *testing.T) {
	service, mocks := getServiceMocks(t)

	ctx := context.Background()

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	req := &GetNFTOwnerRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	}

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemDetails", ctx, collectionID, itemID).
		Return(nil, uniques.ErrItemDetailsNotFound)

	res, err := service.GetNFTOwner(ctx, req)
	assert.ErrorIs(t, err, ErrOwnerNotFound, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTCollection(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1234)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetCollectionDetails", ctx, collectionID).
		Return(nil, uniques.ErrCollectionDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	mockUtils.GetMock[*jobs.DispatcherMock](mocks).
		On("Dispatch", accountID, mock.IsType(&gocelery.Job{})).
		Return(resultMock, nil)

	req := &CreateNFTCollectionRequest{CollectionID: collectionID}

	res, err := service.CreateNFTCollection(ctx, req)
	assert.NoError(t, err)
	assert.IsType(t, &CreateNFTCollectionResponse{}, res)
}

func TestService_CreateNFTCollection_InvalidRequests(t *testing.T) {
	invalidRequests := []*CreateNFTCollectionRequest{
		nil,
		{
			CollectionID: types.U64(0),
		},
	}

	service, _ := getServiceMocks(t)

	for _, invalidRequest := range invalidRequests {
		res, err := service.CreateNFTCollection(context.Background(), invalidRequest)

		assert.True(t, errors.IsOfType(errors.ErrRequestInvalid, err))
		assert.Nil(t, res, "expected no response")
	}
}

func TestService_CreateNFTCollection_AccountError(t *testing.T) {
	service, _ := getServiceMocks(t)

	collectionID := types.U64(1234)

	req := &CreateNFTCollectionRequest{CollectionID: collectionID}

	res, err := service.CreateNFTCollection(context.Background(), req)
	assert.ErrorIs(t, err, errors.ErrContextAccountRetrieval)
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTCollection_ClassCheckError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountMock := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1234)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetCollectionDetails", ctx, collectionID).
		Return(nil, errors.New("class details error"))

	req := &CreateNFTCollectionRequest{CollectionID: collectionID}

	res, err := service.CreateNFTCollection(ctx, req)
	assert.ErrorIs(t, err, ErrCollectionCheck)
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTCollection_ClassExists(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountMock := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1234)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetCollectionDetails", ctx, collectionID).
		Return(&types.CollectionDetails{}, nil)

	req := &CreateNFTCollectionRequest{CollectionID: collectionID}

	res, err := service.CreateNFTCollection(ctx, req)
	assert.ErrorIs(t, err, ErrCollectionAlreadyExists)
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTCollection_DispatchError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1234)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetCollectionDetails", ctx, collectionID).
		Return(nil, uniques.ErrCollectionDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	mockUtils.GetMock[*jobs.DispatcherMock](mocks).
		On("Dispatch", accountID, mock.IsType(&gocelery.Job{})).
		Return(resultMock, errors.New("dispatch error"))

	req := &CreateNFTCollectionRequest{CollectionID: collectionID}

	res, err := service.CreateNFTCollection(ctx, req)
	assert.ErrorIs(t, err, ErrCreateCollectionJobDispatch)
	assert.Nil(t, res, "expected no response")
}

func TestService_GetItemMetadata(t *testing.T) {
	service, mocks := getServiceMocks(t)

	ctx := context.Background()

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	req := &GetItemMetadataRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	}

	instanceMetadata := &types.ItemMetadata{}

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemMetadata", ctx, req.CollectionID, req.ItemID).
		Return(instanceMetadata, nil)

	res, err := service.GetItemMetadata(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, instanceMetadata, res)
}

func TestService_GetItemMetadata_InvalidRequests(t *testing.T) {
	invalidRequests := []*GetItemMetadataRequest{
		nil,
		{
			CollectionID: types.U64(0),
			ItemID:       types.NewU128(*big.NewInt(5678)),
		},
		{
			CollectionID: types.U64(1234),
			ItemID:       types.NewU128(*big.NewInt(0)),
		},
	}

	service, _ := getServiceMocks(t)

	for _, invalidRequest := range invalidRequests {
		res, err := service.GetItemMetadata(context.Background(), invalidRequest)

		assert.True(t, errors.IsOfType(errors.ErrRequestInvalid, err))
		assert.Nil(t, res, "expected no response")
	}
}

func TestService_GetItemMetadata_ApiError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	ctx := context.Background()

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	req := &GetItemMetadataRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	}

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemMetadata", ctx, req.CollectionID, req.ItemID).
		Return(nil, errors.New("api err"))

	res, err := service.GetItemMetadata(ctx, req)
	assert.ErrorIs(t, err, ErrItemMetadataRetrieval, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_GetItemMetadata_ApiErrorNotFound(t *testing.T) {
	service, mocks := getServiceMocks(t)

	ctx := context.Background()

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	req := &GetItemMetadataRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
	}

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemMetadata", ctx, req.CollectionID, req.ItemID).
		Return(nil, uniques.ErrItemMetadataNotFound)

	res, err := service.GetItemMetadata(ctx, req)
	assert.ErrorIs(t, err, ErrItemMetadataNotFound, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_GetItemAttribute(t *testing.T) {
	service, mocks := getServiceMocks(t)

	ctx := context.Background()

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	req := &GetItemAttributeRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
		Key:          "test-key",
	}

	itemAttribute := utils.RandomSlice(32)

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemAttribute", ctx, req.CollectionID, req.ItemID, []byte(req.Key)).
		Return(itemAttribute, nil)

	res, err := service.GetItemAttribute(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, itemAttribute, res)
}

func TestService_GetItemAttribute_InvalidRequests(t *testing.T) {
	invalidRequests := []*GetItemAttributeRequest{
		nil,
		{
			CollectionID: types.U64(0),
			ItemID:       types.NewU128(*big.NewInt(5678)),
			Key:          "test-key",
		},
		{
			CollectionID: types.U64(1234),
			ItemID:       types.NewU128(*big.NewInt(0)),
			Key:          "test-key",
		},
		{
			CollectionID: types.U64(1234),
			ItemID:       types.NewU128(*big.NewInt(5678)),
			Key:          "",
		},
	}

	service, _ := getServiceMocks(t)

	for _, invalidRequest := range invalidRequests {
		res, err := service.GetItemAttribute(context.Background(), invalidRequest)

		assert.True(t, errors.IsOfType(errors.ErrRequestInvalid, err))
		assert.Nil(t, res, "expected no response")
	}
}

func TestService_GetItemAttribute_ApiError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	ctx := context.Background()

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	req := &GetItemAttributeRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
		Key:          "test-key",
	}

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemAttribute", ctx, req.CollectionID, req.ItemID, []byte(req.Key)).
		Return(nil, errors.New("error"))

	res, err := service.GetItemAttribute(ctx, req)
	assert.ErrorIs(t, err, ErrItemAttributeRetrieval)
	assert.Nil(t, res)
}

func TestService_GetItemAttribute_ApiErrorNotFound(t *testing.T) {
	service, mocks := getServiceMocks(t)

	ctx := context.Background()

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	req := &GetItemAttributeRequest{
		CollectionID: collectionID,
		ItemID:       itemID,
		Key:          "test-key",
	}

	mockUtils.GetMock[*uniques.UniquesAPIMock](mocks).
		On("GetItemAttribute", ctx, req.CollectionID, req.ItemID, []byte(req.Key)).
		Return(nil, uniques.ErrItemAttributeNotFound)

	res, err := service.GetItemAttribute(ctx, req)
	assert.ErrorIs(t, err, ErrItemAttributeNotFound)
	assert.Nil(t, res)
}

func getServiceMocks(t *testing.T) (Service, []any) {
	pendingDocServiceMock := pending.NewServiceMock(t)
	documentsServiceMock := documents.NewServiceMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)
	uniquesAPIMock := uniques.NewUniquesAPIMock(t)

	service := NewService(
		pendingDocServiceMock,
		documentsServiceMock,
		dispatcherMock,
		uniquesAPIMock,
	)

	return service, []any{
		pendingDocServiceMock,
		documentsServiceMock,
		dispatcherMock,
		uniquesAPIMock,
	}
}
