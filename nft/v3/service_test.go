//go:build unit
// +build unit

package v3

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/errors"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/identity"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/go-centrifuge/jobs"

	"github.com/centrifuge/go-centrifuge/documents"
)

func TestService_MintNFT(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := newService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "expected no error")

	req := &MintNFTRequest{
		DocumentID: []byte("document_id"),
		PublicInfo: []string{"test_string"},
		ClassID:    types.U64(1234),
		Owner:      types.NewAccountID([]byte("account_id")),
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

	api.On("GetInstanceDetails", ctx, classID2, instanceID2).
		Return(nil, nil)

	api.On("GetInstanceDetails", ctx, classID2, mock.Anything).
		Return(nil, nil)

	did, err := identity.NewDIDFromBytes(testAcc.GetIdentityID())
	assert.NoError(t, err, "expected no error")

	dispatcher.On("Dispatch", did, mock.IsType(&gocelery.Job{})).
		Return(jobs.MockResult{}, nil)

	res, err := service.MintNFT(ctx, req)
	assert.NoError(t, err, "expected no error")
	assert.IsType(t, &MintNFTResponse{}, res, "types should match")
}

func TestService_MintNFT_NoNFTsPresent(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := newService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "expected no error")

	req := &MintNFTRequest{
		DocumentID: []byte("document_id"),
		PublicInfo: []string{"test_string"},
		ClassID:    types.U64(1234),
		Owner:      types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("CcNfts").
		Return(nil)

	api.On("GetInstanceDetails", ctx, types.U64(1234), mock.Anything).
		Return(nil, nil)

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

	service := newService(docSrv, dispatcher, api)

	req := &MintNFTRequest{
		DocumentID: []byte("document_id"),
		PublicInfo: []string{"test_string"},
		ClassID:    types.U64(1234),
		Owner:      types.NewAccountID([]byte("account_id")),
	}

	res, err := service.MintNFT(context.Background(), req)
	assert.ErrorIs(t, err, ErrAccountFromContextRetrieval, "errors should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_DocError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := newService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	req := &MintNFTRequest{
		DocumentID: []byte("document_id"),
		PublicInfo: []string{"test_string"},
		ClassID:    types.U64(1234),
		Owner:      types.NewAccountID([]byte("account_id")),
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

	service := newService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	req := &MintNFTRequest{
		DocumentID: []byte("document_id"),
		PublicInfo: []string{"test_string"},
		ClassID:    types.U64(1234),
		Owner:      types.NewAccountID([]byte("account_id")),
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

	api.On("GetInstanceDetails", ctx, classID2, instanceID2).
		Return(instanceDetails, nil)

	res, err := service.MintNFT(ctx, req)
	assert.ErrorIs(t, err, ErrInstanceAlreadyMinted, "errors should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_InstanceIDGeneration_ContextError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := newService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	ctx, cancel := context.WithCancel(ctx)

	req := &MintNFTRequest{
		DocumentID: []byte("document_id"),
		PublicInfo: []string{"test_string"},
		ClassID:    types.U64(1234),
		Owner:      types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("CcNfts").
		Return(nil)

	instanceDetails := &types.InstanceDetails{}

	api.On("GetInstanceDetails", ctx, types.U64(1234), mock.Anything).
		Return(instanceDetails, nil)

	go func() {
		time.Sleep(3 * time.Second)
		cancel()
	}()

	res, err := service.MintNFT(ctx, req)

	assert.ErrorIs(t, err, ErrInstanceIDGeneration, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_InstanceIDGeneration_InstanceDetailsError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := newService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	req := &MintNFTRequest{
		DocumentID: []byte("document_id"),
		PublicInfo: []string{"test_string"},
		ClassID:    types.U64(1234),
		Owner:      types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("CcNfts").
		Return(nil)

	api.On("GetInstanceDetails", ctx, types.U64(1234), mock.Anything).
		Return(nil, errors.New("instance details error"))

	res, err := service.MintNFT(ctx, req)

	assert.ErrorIs(t, err, ErrInstanceIDGeneration, "errors should match")
	assert.Nil(t, res, "expected no response")
}

func TestService_MintNFT_IdentityError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := newService(docSrv, dispatcher, api)

	mockAccount := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	req := &MintNFTRequest{
		DocumentID: []byte("document_id"),
		PublicInfo: []string{"test_string"},
		ClassID:    types.U64(1234),
		Owner:      types.NewAccountID([]byte("account_id")),
	}

	doc := documents.NewDocumentMock(t)

	docSrv.On("GetCurrentVersion", ctx, req.DocumentID).
		Return(doc, nil)

	doc.On("CcNfts").
		Return(nil)

	api.On("GetInstanceDetails", ctx, types.U64(1234), mock.Anything).
		Return(nil, nil)

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

	service := newService(docSrv, dispatcher, api)

	ctx := testingconfig.CreateAccountContext(t, cfg)

	testAcc, err := contextutil.Account(ctx)
	assert.NoError(t, err, "expected no error")

	req := &MintNFTRequest{
		DocumentID: []byte("document_id"),
		PublicInfo: []string{"test_string"},
		ClassID:    types.U64(1234),
		Owner:      types.NewAccountID([]byte("account_id")),
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

	api.On("GetInstanceDetails", ctx, classID2, instanceID2).
		Return(nil, nil)

	api.On("GetInstanceDetails", ctx, classID2, mock.Anything).
		Return(nil, nil)

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

	service := newService(docSrv, dispatcher, api)

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

	api.On("GetInstanceDetails", ctx, classID, instanceID).
		Return(instanceDetails, nil)

	res, err := service.OwnerOf(ctx, req)
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, classID, res.ClassID, "class IDs should be equal")
	assert.Equal(t, instanceID, res.InstanceID, "instance IDs should be equal")
	assert.Equal(t, owner, res.AccountID, "account IDs should be equal")
}

func TestService_OwnerOf_InstanceDetailsError(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := newService(docSrv, dispatcher, api)

	ctx := context.Background()

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	req := &OwnerOfRequest{
		ClassID:    classID,
		InstanceID: instanceID,
	}

	api.On("GetInstanceDetails", ctx, classID, instanceID).
		Return(nil, errors.New("instance details error"))

	res, err := service.OwnerOf(ctx, req)
	assert.ErrorIs(t, err, ErrInstanceDetailsRetrieval, "error should be equal")
	assert.Nil(t, res, "expected no response")
}

func TestService_OwnerOf_InstanceDetailsNotFound(t *testing.T) {
	docSrv := documents.NewServiceMock(t)
	dispatcher := jobs.NewDispatcherMock(t)
	api := NewUniquesAPIMock(t)

	service := newService(docSrv, dispatcher, api)

	ctx := context.Background()

	classID := types.U64(1234)
	instanceID := types.NewU128(*big.NewInt(5678))

	req := &OwnerOfRequest{
		ClassID:    classID,
		InstanceID: instanceID,
	}

	api.On("GetInstanceDetails", ctx, classID, instanceID).
		Return(nil, nil)

	res, err := service.OwnerOf(ctx, req)
	assert.ErrorIs(t, err, ErrInstanceDetailsNotFound, "errors should match")
	assert.Nil(t, res, "expected no response")
}
