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
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_MintNFT(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "expected no error")

	req := &MintNFTRequest{
		DocumentID:    []byte("document_id"),
		DocAttributes: getTestAttributeKeys(),
		ClassID:       types.U64(1234),
		Owner:         types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	classID1 := types.U64(1111)
	instanceID1 := types.NewU128(*big.NewInt(2222))

	classID2 := types.U64(1234)
	instanceID2 := types.NewU128(*big.NewInt(5678))

	encodedClassID1, err := types.EncodeToBytes(classID1)
	assert.NoError(t, err, "expected no error")
	encodedInstanceID1, err := types.EncodeToBytes(instanceID1)
	assert.NoError(t, err, "expected no error")

	encodedClassID2, err := types.EncodeToBytes(classID2)
	assert.NoError(t, err, "expected no error")
	encodedInstanceID2, err := types.EncodeToBytes(instanceID2)
	assert.NoError(t, err, "expected no error")

	ccNfts := []*coredocumentpb.NFT{
		{
			ClassId:    encodedClassID1,
			InstanceId: encodedInstanceID1,
		},
		{
			ClassId:    encodedClassID2,
			InstanceId: encodedInstanceID2,
		},
	}

	doc.On("CcNfts").
		Return(ccNfts)

	api.On("GetItemDetails", ctx, classID2, instanceID2).
		Return(nil, ErrItemDetailsNotFound)

	api.On("GetItemDetails", ctx, classID2, mock.Anything).
		Return(nil, ErrItemDetailsNotFound)

	did, err := identity.NewDIDFromBytes(testAcc.GetIdentityID())
	assert.NoError(t, err, "expected no error")

	dispatcher.On("Dispatch", did, mock.IsType(&gocelery.Job{})).
		Return(jobs.MockResult{}, nil)

	res, err := service.MintNFT(ctx, req)
	assert.NoError(t, err, "expected no error")
	assert.IsType(t, &MintNFTResponse{}, res, "types should match")
}

func TestService_MintNFT_InvalidRequests(t *testing.T) {
	invalidRequests := []*MintNFTRequest{
		nil,
		{
			DocumentID: nil,
			ClassID:    types.U64(1234),
		},
		{
			DocumentID: []byte("doc-id"),
			ClassID:    types.U64(0),
		},
	}

	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	for _, invalidRequest := range invalidRequests {
		res, err := service.MintNFT(context.Background(), invalidRequest)
		assert.ErrorIs(t, err, ErrRequestInvalid)
		assert.Nil(t, res)
	}
}

func TestService_MintNFT_NoNFTsPresent(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "expected no error")

	req := &MintNFTRequest{
		DocumentID:    []byte("document_id"),
		DocAttributes: getTestAttributeKeys(),
		ClassID:       types.U64(1234),
		Owner:         types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("CcNfts").
		Return(nil)

	api.On("GetItemDetails", ctx, types.U64(1234), mock.Anything).
		Return(nil, ErrItemDetailsNotFound)

	did, err := identity.NewDIDFromBytes(testAcc.GetIdentityID())
	assert.NoError(t, err, "expected no error")

	dispatcher.On("Dispatch", did, mock.IsType(&gocelery.Job{})).
		Return(jobs.MockResult{}, nil)

	res, err := service.MintNFT(ctx, req)
	assert.NoError(t, err, "expected no error")
	assert.IsType(t, &MintNFTResponse{}, res, "types should match")
}

func TestService_MintNFT_AccountError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	req := &MintNFTRequest{
		DocumentID:    []byte("document_id"),
		DocAttributes: getTestAttributeKeys(),
		ClassID:       types.U64(1234),
		Owner:         types.NewAccountID([]byte("account_id")),
	}

	res, err := service.MintNFT(context.Background(), req)
	assert.ErrorIs(t, err, ErrAccountFromContextRetrieval, "errors should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_DocError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	req := &MintNFTRequest{
		DocumentID:    []byte("document_id"),
		DocAttributes: getTestAttributeKeys(),
		ClassID:       types.U64(1234),
		Owner:         types.NewAccountID([]byte("account_id")),
	}

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(nil, errors.New("document error"))

	res, err := service.MintNFT(ctx, req)
	assert.ErrorIs(t, err, ErrDocumentRetrieval, "errors should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_InstanceAlreadyMinted(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	req := &MintNFTRequest{
		DocumentID:    []byte("document_id"),
		DocAttributes: getTestAttributeKeys(),
		ClassID:       types.U64(1234),
		Owner:         types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	classID1 := types.U64(1111)
	instanceID1 := types.NewU128(*big.NewInt(2222))

	classID2 := types.U64(1234)
	instanceID2 := types.NewU128(*big.NewInt(5678))

	encodedClassID1, err := types.EncodeToBytes(classID1)
	assert.NoError(t, err, "expected no error")
	encodedInstanceID1, err := types.EncodeToBytes(instanceID1)
	assert.NoError(t, err, "expected no error")

	encodedClassID2, err := types.EncodeToBytes(classID2)
	assert.NoError(t, err, "expected no error")
	encodedInstanceID2, err := types.EncodeToBytes(instanceID2)
	assert.NoError(t, err, "expected no error")

	ccNfts := []*coredocumentpb.CcNft{
		{
			ClassId:    encodedClassID1,
			InstanceId: encodedInstanceID1,
		},
		{
			ClassId:    encodedClassID2,
			InstanceId: encodedInstanceID2,
		},
	}

	doc.On("CcNfts").
		Return(ccNfts)

	doc.On("CurrentVersion").
		Return(utils.RandomSlice(32))

	instanceDetails := &types.InstanceDetails{}

	api.On("GetItemDetails", ctx, classID2, instanceID2).
		Return(instanceDetails, nil)

	res, err := service.MintNFT(ctx, req)
	assert.ErrorIs(t, err, ErrItemAlreadyMinted, "errors should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_InstanceIDGeneration_ContextError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	ctx, cancel := context.WithCancel(ctx)

	req := &MintNFTRequest{
		DocumentID:    []byte("document_id"),
		DocAttributes: getTestAttributeKeys(),
		ClassID:       types.U64(1234),
		Owner:         types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("CcNfts").
		Return(nil)

	instanceDetails := &types.InstanceDetails{}

	api.On("GetItemDetails", ctx, types.U64(1234), mock.Anything).
		Return(instanceDetails, nil)

	go func() {
		time.Sleep(3 * time.Second)
		cancel()
	}()

	res, err := service.MintNFT(ctx, req)

	assert.ErrorIs(t, err, ErrItemIDGeneration, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_InstanceIDGeneration_InstanceDetailsError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	req := &MintNFTRequest{
		DocumentID:    []byte("document_id"),
		DocAttributes: getTestAttributeKeys(),
		ClassID:       types.U64(1234),
		Owner:         types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("CcNfts").
		Return(nil)

	api.On("GetItemDetails", ctx, types.U64(1234), mock.Anything).
		Return(nil, errors.New("instance details error"))

	res, err := service.MintNFT(ctx, req)

	assert.ErrorIs(t, err, ErrItemIDGeneration, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_IdentityError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	mockAccount := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	req := &MintNFTRequest{
		DocumentID:    []byte("document_id"),
		DocAttributes: getTestAttributeKeys(),
		ClassID:       types.U64(1234),
		Owner:         types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("CcNfts").
		Return(nil)

	api.On("GetItemDetails", ctx, types.U64(1234), mock.Anything).
		Return(nil, ErrItemDetailsNotFound)

	mockAccount.On("GetIdentityID").
		Return([]byte{})

	res, err := service.MintNFT(ctx, req)
	assert.ErrorIs(t, err, ErrIdentityRetrieval, "errors should match")
	assert.Nil(t, res, "no response expected")
}

func TestService_MintNFT_DispatchError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "expected no error")

	req := &MintNFTRequest{
		DocumentID:    []byte("document_id"),
		DocAttributes: getTestAttributeKeys(),
		ClassID:       types.U64(1234),
		Owner:         types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	classID1 := types.U64(1111)
	instanceID1 := types.NewU128(*big.NewInt(2222))

	classID2 := types.U64(1234)
	instanceID2 := types.NewU128(*big.NewInt(5678))

	encodedClassID1, err := types.EncodeToBytes(classID1)
	assert.NoError(t, err, "expected no error")
	encodedInstanceID1, err := types.EncodeToBytes(instanceID1)
	assert.NoError(t, err, "expected no error")

	encodedClassID2, err := types.EncodeToBytes(classID2)
	assert.NoError(t, err, "expected no error")
	encodedInstanceID2, err := types.EncodeToBytes(instanceID2)
	assert.NoError(t, err, "expected no error")

	ccNfts := []*coredocumentpb.CcNft{
		{
			ClassId:    encodedClassID1,
			InstanceId: encodedInstanceID1,
		},
		{
			ClassId:    encodedClassID2,
			InstanceId: encodedInstanceID2,
		},
	}

	doc.On("CcNfts").
		Return(ccNfts)

	api.On("GetItemDetails", ctx, classID2, instanceID2).
		Return(nil, ErrItemDetailsNotFound)

	api.On("GetItemDetails", ctx, classID2, mock.Anything).
		Return(nil, ErrItemDetailsNotFound)

	did, err := identity.NewDIDFromBytes(testAcc.GetIdentityID())
	assert.NoError(t, err, "expected no error")

	dispatcher.On("Dispatch", did, mock.IsType(&gocelery.Job{})).
		Return(jobs.MockResult{}, errors.New("dispatch error"))

	res, err := service.MintNFT(ctx, req)
	assert.ErrorIs(t, err, ErrMintJobDispatch, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_OwnerOf(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := context.Background()

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	req := &OwnerOfRequest{
		ClassID:    classID,
		InstanceID: instanceID,
	}

	owner := types.NewAccountID(utils.RandomSlice(32))

	instanceDetails := &types.InstanceDetails{
		Owner: owner,
	}

	api.On("GetItemDetails", ctx, classID, instanceID).
		Return(instanceDetails, nil)

	res, err := service.OwnerOf(ctx, req)
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, classID, res.ClassID, "class IDs should be equal")
	assert.Equal(t, instanceID, res.InstanceID, "instance IDs should be equal")
	assert.Equal(t, owner, res.AccountID, "account IDs should be equal")
}

func TestService_OwnerOf_InvalidRequests(t *testing.T) {
	invalidRequests := []*OwnerOfRequest{
		nil,
		{
			ClassID:    types.U64(0),
			InstanceID: types.NewU128(*big.NewInt(5678)),
		},
		{
			ClassID:    types.U64(1234),
			InstanceID: types.NewU128(*big.NewInt(0)),
		},
	}
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := context.Background()

	for _, invalidRequest := range invalidRequests {
		res, err := service.OwnerOf(ctx, invalidRequest)
		assert.ErrorIs(t, err, ErrRequestInvalid, "errors should match")
		assert.Nil(t, res, "expected no response")
	}
}

func TestService_OwnerOf_InstanceDetailsError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := context.Background()

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	req := &OwnerOfRequest{
		ClassID:    classID,
		InstanceID: instanceID,
	}

	api.On("GetItemDetails", ctx, classID, instanceID).
		Return(nil, errors.New("instance details error"))

	res, err := service.OwnerOf(ctx, req)
	assert.ErrorIs(t, err, ErrItemDetailsRetrieval, "error should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_OwnerOf_InstanceDetailsNotFound(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := context.Background()

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	req := &OwnerOfRequest{
		ClassID:    classID,
		InstanceID: instanceID,
	}

	api.On("GetItemDetails", ctx, classID, instanceID).
		Return(nil, ErrItemDetailsNotFound)

	res, err := service.OwnerOf(ctx, req)
	assert.ErrorIs(t, err, ErrOwnerNotFound, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTClass(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "expected no error")

	did, err := identity.NewDIDFromBytes(testAcc.GetIdentityID())
	assert.NoError(t, err, "expected no error")

	classID := types.U64(1234)

	api.On("GetCollectionDetails", ctx, classID).
		Return(nil, ErrCollectionDetailsNotFound)

	dispatcher.On("Dispatch", did, mock.IsType(&gocelery.Job{})).
		Return(jobs.MockResult{}, nil)

	req := &CreateNFTCollectionRequest{ClassID: classID}

	res, err := service.CreateNFTCollection(ctx, req)
	assert.NoError(t, err)
	assert.IsType(t, &CreateNFTCollectionResponse{}, res)
}

func TestService_CreateNFTClassInvalidRequests(t *testing.T) {
	invalidRequests := []*CreateNFTCollectionRequest{
		nil,
		{
			ClassID: types.U64(0),
		},
	}

	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := context.Background()

	for _, invalidRequest := range invalidRequests {
		res, err := service.CreateNFTCollection(ctx, invalidRequest)
		assert.ErrorIs(t, err, ErrRequestInvalid, "errors should match")
		assert.Nil(t, res, "expected no response")
	}
}

func TestService_CreateNFTClass_AccountError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	classID := types.U64(1234)

	req := &CreateNFTCollectionRequest{ClassID: classID}

	res, err := service.CreateNFTCollection(context.Background(), req)
	assert.ErrorIs(t, err, ErrAccountFromContextRetrieval)
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTClass_IdentityError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	mockAccount := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	mockAccount.On("GetIdentityID").
		Return([]byte{})

	classID := types.U64(1234)

	req := &CreateNFTCollectionRequest{ClassID: classID}

	res, err := service.CreateNFTCollection(ctx, req)
	assert.ErrorIs(t, err, ErrIdentityRetrieval)
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTClass_ClassCheckError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	classID := types.U64(1234)

	api.On("GetCollectionDetails", ctx, classID).
		Return(nil, errors.New("class details error"))

	req := &CreateNFTCollectionRequest{ClassID: classID}

	res, err := service.CreateNFTCollection(ctx, req)
	assert.ErrorIs(t, err, ErrCollectionCheck)
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTClass_ClassExists(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	classID := types.U64(1234)

	api.On("GetCollectionDetails", ctx, classID).
		Return(&types.ClassDetails{}, nil)

	req := &CreateNFTCollectionRequest{ClassID: classID}

	res, err := service.CreateNFTCollection(ctx, req)
	assert.ErrorIs(t, err, ErrCollectionAlreadyExists)
	assert.Nil(t, res, "expected no response")
}

func TestService_CreateNFTClass_DispatchError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "expected no error")

	did, err := identity.NewDIDFromBytes(testAcc.GetIdentityID())
	assert.NoError(t, err, "expected no error")

	classID := types.U64(1234)

	api.On("GetCollectionDetails", ctx, classID).
		Return(nil, ErrCollectionDetailsNotFound)

	dispatcher.On("Dispatch", did, mock.IsType(&gocelery.Job{})).
		Return(jobs.MockResult{}, errors.New("dispatch error"))

	req := &CreateNFTCollectionRequest{ClassID: classID}

	res, err := service.CreateNFTCollection(ctx, req)
	assert.ErrorIs(t, err, ErrCreateCollectionJobDispatch)
	assert.Nil(t, res, "expected no response")
}

func TestService_InstanceMetadataOf(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	req := &GetItemMetadataRequest{
		ClassID:    classID,
		InstanceID: instanceID,
	}

	ctx := context.Background()

	instanceMetadata := &types.InstanceMetadata{}

	api.On("GetItemMetadata", ctx, req.ClassID, req.InstanceID).
		Return(instanceMetadata, nil)

	res, err := service.InstanceMetadataOf(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, instanceMetadata, res)
}

func TestService_InstanceMetadataOf_InvalidRequests(t *testing.T) {
	invalidRequests := []*GetItemMetadataRequest{
		nil,
		{
			ClassID:    types.U64(0),
			InstanceID: types.NewU128(*big.NewInt(5678)),
		},
		{
			ClassID:    types.U64(1234),
			InstanceID: types.NewU128(*big.NewInt(0)),
		},
	}

	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	ctx := context.Background()

	for _, invalidRequest := range invalidRequests {
		res, err := service.InstanceMetadataOf(ctx, invalidRequest)
		assert.ErrorIs(t, err, ErrRequestInvalid, "errors should match")
		assert.Nil(t, res, "expected no response")
	}
}

func TestService_InstanceMetadataOf_ApiError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	req := &GetItemMetadataRequest{
		ClassID:    classID,
		InstanceID: instanceID,
	}

	ctx := context.Background()

	api.On("GetItemMetadata", ctx, req.ClassID, req.InstanceID).
		Return(nil, errors.New("api err"))

	res, err := service.InstanceMetadataOf(ctx, req)
	assert.ErrorIs(t, err, ErrItemMetadataRetrieval, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_InstanceMetadataOf_ApiErrorNotFound(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := NewService(docSrv, dispatcher, api)

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	req := &GetItemMetadataRequest{
		ClassID:    classID,
		InstanceID: instanceID,
	}

	ctx := context.Background()

	api.On("GetItemMetadata", ctx, req.ClassID, req.InstanceID).
		Return(nil, ErrItemMetadataNotFound)

	res, err := service.InstanceMetadataOf(ctx, req)
	assert.ErrorIs(t, err, ErrItemMetadataNotFound, "errors should match")
	assert.Nil(t, res, "expected no response")
}

var (
	attrKeyLabels = []string{
		"key_1",
		"key_2",
		"key_3",
	}
)

func getTestAttributeKeys() []documents.AttrKey {
	var keys []documents.AttrKey

	for _, attrKeyLabel := range attrKeyLabels {
		key, _ := documents.AttrKeyFromLabel(attrKeyLabel)

		keys = append(keys, key)
	}

	return keys
}
