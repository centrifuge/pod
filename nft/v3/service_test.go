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
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
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

	genericUtils.GetMock[*documents.ServiceMock](mocks).
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

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, itemID1).
		Return(nil, uniques.ErrItemDetailsNotFound)

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, mock.Anything).
		Return(nil, uniques.ErrItemDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	genericUtils.GetMock[*jobs.DispatcherMock](mocks).
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

	genericUtils.GetMock[*pending.ServiceMock](mocks).
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

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, itemID1).
		Return(nil, uniques.ErrItemDetailsNotFound)

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, mock.Anything).
		Return(nil, uniques.ErrItemDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	genericUtils.GetMock[*jobs.DispatcherMock](mocks).
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

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("NFTs").
		Return(nil)

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, mock.Anything).
		Return(nil, uniques.ErrItemDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	genericUtils.GetMock[*jobs.DispatcherMock](mocks).
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

	genericUtils.GetMock[*documents.ServiceMock](mocks).
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

	genericUtils.GetMock[*pending.ServiceMock](mocks).
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

	genericUtils.GetMock[*documents.ServiceMock](mocks).
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

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, itemID1).
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

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("NFTs").
		Return(nil)

	instanceDetails := &types.ItemDetails{}

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, mock.Anything).
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

	genericUtils.GetMock[*documents.ServiceMock](mocks).
		On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("NFTs").
		Return(nil)

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, mock.Anything).
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

	genericUtils.GetMock[*documents.ServiceMock](mocks).
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

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, itemID1).
		Return(nil, uniques.ErrItemDetailsNotFound)

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, mock.Anything).
		Return(nil, uniques.ErrItemDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	genericUtils.GetMock[*jobs.DispatcherMock](mocks).
		On("Dispatch", accountID, mock.IsType(&gocelery.Job{})).
		Return(resultMock, errors.New("dispatch error"))

	res, err := service.MintNFT(ctx, req, false)
	assert.ErrorIs(t, err, ErrMintJobDispatch, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_GetNFTOwner(t *testing.T) {
	service, mocks := getServiceMocks(t)

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	ownerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	instanceDetails := &types.ItemDetails{
		Owner: *ownerAccountID,
	}

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, itemID).
		Return(instanceDetails, nil)

	res, err := service.GetNFTOwner(collectionID, itemID)
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, ownerAccountID, res, "account IDs should be equal")
}

func TestService_GetNFTOwner_ItemDetailsError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, itemID).
		Return(nil, errors.New("instance details error"))

	res, err := service.GetNFTOwner(collectionID, itemID)
	assert.ErrorIs(t, err, ErrOwnerRetrieval, "error should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_GetNFTOwner_ItemDetailsNotFound(t *testing.T) {
	service, mocks := getServiceMocks(t)

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemDetails", collectionID, itemID).
		Return(nil, uniques.ErrItemDetailsNotFound)

	res, err := service.GetNFTOwner(collectionID, itemID)
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

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetCollectionDetails", collectionID).
		Return(nil, uniques.ErrCollectionDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	genericUtils.GetMock[*jobs.DispatcherMock](mocks).
		On("Dispatch", accountID, mock.IsType(&gocelery.Job{})).
		Return(resultMock, nil)

	res, err := service.CreateNFTCollection(ctx, collectionID)
	assert.NoError(t, err)
	assert.IsType(t, &CreateNFTCollectionResponse{}, res)
}

func TestService_CreateNFTCollection_AccountError(t *testing.T) {
	service, _ := getServiceMocks(t)

	collectionID := types.U64(1234)

	res, err := service.CreateNFTCollection(context.Background(), collectionID)
	assert.ErrorIs(t, err, errors.ErrContextAccountRetrieval)
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTCollection_CollectionCheckError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountMock := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1234)

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetCollectionDetails", collectionID).
		Return(nil, errors.New("class details error"))

	res, err := service.CreateNFTCollection(ctx, collectionID)
	assert.ErrorIs(t, err, ErrCollectionCheck)
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTCollection_CollectionExists(t *testing.T) {
	service, mocks := getServiceMocks(t)

	accountMock := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	collectionID := types.U64(1234)

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetCollectionDetails", collectionID).
		Return(&types.CollectionDetails{}, nil)

	res, err := service.CreateNFTCollection(ctx, collectionID)
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

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetCollectionDetails", collectionID).
		Return(nil, uniques.ErrCollectionDetailsNotFound)

	resultMock := jobs.NewResultMock(t)

	genericUtils.GetMock[*jobs.DispatcherMock](mocks).
		On("Dispatch", accountID, mock.IsType(&gocelery.Job{})).
		Return(resultMock, errors.New("dispatch error"))

	res, err := service.CreateNFTCollection(ctx, collectionID)
	assert.ErrorIs(t, err, ErrCreateCollectionJobDispatch)
	assert.Nil(t, res, "expected no response")
}

func TestService_GetItemMetadata(t *testing.T) {
	service, mocks := getServiceMocks(t)

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	instanceMetadata := &types.ItemMetadata{}

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemMetadata", collectionID, itemID).
		Return(instanceMetadata, nil)

	res, err := service.GetItemMetadata(collectionID, itemID)
	assert.NoError(t, err)
	assert.Equal(t, instanceMetadata, res)
}

func TestService_GetItemMetadata_ApiError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemMetadata", collectionID, itemID).
		Return(nil, errors.New("api err"))

	res, err := service.GetItemMetadata(collectionID, itemID)
	assert.ErrorIs(t, err, ErrItemMetadataRetrieval, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_GetItemMetadata_ApiErrorNotFound(t *testing.T) {
	service, mocks := getServiceMocks(t)

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemMetadata", collectionID, itemID).
		Return(nil, uniques.ErrItemMetadataNotFound)

	res, err := service.GetItemMetadata(collectionID, itemID)
	assert.ErrorIs(t, err, ErrItemMetadataNotFound, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_GetItemAttribute(t *testing.T) {
	service, mocks := getServiceMocks(t)

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))
	key := "test-key"

	itemAttribute := utils.RandomSlice(32)

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemAttribute", collectionID, itemID, []byte(key)).
		Return(itemAttribute, nil)

	res, err := service.GetItemAttribute(collectionID, itemID, key)
	assert.NoError(t, err)
	assert.Equal(t, itemAttribute, res)
}

func TestService_GetItemAttribute_ApiError(t *testing.T) {
	service, mocks := getServiceMocks(t)

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))
	key := "test-key"

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemAttribute", collectionID, itemID, []byte(key)).
		Return(nil, errors.New("error"))

	res, err := service.GetItemAttribute(collectionID, itemID, key)
	assert.ErrorIs(t, err, ErrItemAttributeRetrieval)
	assert.Nil(t, res)
}

func TestService_GetItemAttribute_ApiErrorNotFound(t *testing.T) {
	service, mocks := getServiceMocks(t)

	collectionID := types.U64(1234)
	itemID := types.NewU128(*big.NewInt(5678))
	key := "test-key"

	genericUtils.GetMock[*uniques.APIMock](mocks).
		On("GetItemAttribute", collectionID, itemID, []byte(key)).
		Return(nil, uniques.ErrItemAttributeNotFound)

	res, err := service.GetItemAttribute(collectionID, itemID, key)
	assert.ErrorIs(t, err, ErrItemAttributeNotFound)
	assert.Nil(t, res)
}

func getServiceMocks(t *testing.T) (Service, []any) {
	pendingDocServiceMock := pending.NewServiceMock(t)
	documentsServiceMock := documents.NewServiceMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)
	uniquesAPIMock := uniques.NewAPIMock(t)

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
