// +build unit

package userapi

import (
	"context"
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransferService struct {
	transferdetails.Service
	mock.Mock
}

func (m *MockTransferService) CreateTransferDetail(ctx context.Context, payload transferdetails.CreateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return model, jobID, args.Error(2)
}

func (m *MockTransferService) UpdateTransferDetail(ctx context.Context, payload transferdetails.UpdateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return model, jobID, args.Error(2)
}

func (m *MockTransferService) DeriveTransferDetail(ctx context.Context, model documents.Model, transferID []byte) (*transferdetails.TransferDetail, documents.Model, error) {
	args := m.Called(ctx, model, transferID)
	td, _ := args.Get(0).(*transferdetails.TransferDetail)
	nm, _ := args.Get(1).(documents.Model)
	return td, nm, args.Error(2)
}

func (m *MockTransferService) DeriveTransferList(ctx context.Context, model documents.Model) (*transferdetails.TransferDetailList, documents.Model, error) {
	args := m.Called(ctx, model)
	td, _ := args.Get(0).(*transferdetails.TransferDetailList)
	nm, _ := args.Get(1).(documents.Model)
	return td, nm, args.Error(2)
}

func newCoreAPIService(docSrv documents.Service) coreapi.Service {
	return coreapi.NewService(docSrv, nil, nil, nil)
}

func TestService_CreateTransferDetail(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	transferSrv.On("CreateTransferDetail", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	nm, _, err := service.CreateTransferDetail(context.Background(), transferdetails.CreateTransferDetailRequest{
		DocumentID: "test_id",
		Data:       transferdetails.Data{},
	})
	assert.NoError(t, err)
	assert.Equal(t, m, nm)
}

func TestService_UpdateDocument(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	transferSrv.On("UpdateTransferDetail", mock.Anything, mock.Anything).Return(m, jobs.NewJobID(), nil)
	nm, _, err := service.UpdateTransferDetail(context.Background(), transferdetails.UpdateTransferDetailRequest{
		DocumentID: "test_id",
		TransferID: "test_transfer",
		Data:       transferdetails.Data{},
	})
	assert.NoError(t, err)
	assert.Equal(t, m, nm)
}

func TestService_GetCurrentTransferDetail(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	td := new(transferdetails.TransferDetail)
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(m, nil)
	transferSrv.On("DeriveTransferDetail", mock.Anything, mock.Anything, mock.Anything).Return(td, m, nil)
	ntd, nm, err := service.GetCurrentTransferDetail(context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, td, ntd)
	assert.Equal(t, m, nm)
}

func TestService_GetCurrentTransferDetailList(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	td := new(transferdetails.TransferDetailList)
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(m, nil)
	transferSrv.On("DeriveTransferList", mock.Anything, mock.Anything).Return(td, m, nil)
	ntd, nm, err := service.GetCurrentTransferDetailsList(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, td, ntd)
	assert.Equal(t, m, nm)
}

func TestService_GetVersionTransferDetail(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	td := new(transferdetails.TransferDetail)
	docSrv.On("GetVersion", mock.Anything, mock.Anything, mock.Anything).Return(m, nil)
	transferSrv.On("DeriveTransferDetail", mock.Anything, mock.Anything, mock.Anything).Return(td, m, nil)
	ntd, nm, err := service.GetVersionTransferDetail(context.Background(), nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, td, ntd)
	assert.Equal(t, m, nm)
}

func TestService_GetVersionTransferDetailList(t *testing.T) {
	transferSrv := new(MockTransferService)
	docSrv := new(testingdocuments.MockService)
	service := Service{coreAPISrv: newCoreAPIService(docSrv), transferDetailsService: transferSrv}
	m := new(testingdocuments.MockModel)
	td := new(transferdetails.TransferDetailList)
	docSrv.On("GetVersion", mock.Anything, mock.Anything, mock.Anything).Return(m, nil)
	transferSrv.On("DeriveTransferList", mock.Anything, mock.Anything).Return(td, m, nil)
	ntd, nm, err := service.GetVersionTransferDetailsList(context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, td, ntd)
	assert.Equal(t, m, nm)
}

func TestService_CreateEntity(t *testing.T) {
	ctx := context.Background()
	did := testingidentity.GenerateRandomDID()
	req := CreateEntityRequest{
		WriteAccess: []identity.DID{did},
		Data: entity.Data{
			Identity:  &did,
			LegalName: "John Doe",
			Addresses: []entity.Address{
				{
					IsMain:  true,
					Country: "Germany",
					Label:   "home",
				},
			},
		},
		Attributes: coreapi.AttributeMapRequest{
			"string_test": {
				Type:  "invalid",
				Value: "hello, world!",
			},

			"decimal_test": {
				Type:  "decimal",
				Value: "100001.001",
			},
		},
	}
	s := Service{}

	// invalid attribute map
	_, _, err := s.CreateEntity(ctx, req)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrNotValidAttrType, err))

	// success
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	docSrv.On("CreateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil)
	s.coreAPISrv = newCoreAPIService(docSrv)
	strAttr := req.Attributes["string_test"]
	strAttr.Type = "string"
	req.Attributes["string_test"] = strAttr
	_, _, err = s.CreateEntity(ctx, req)
	assert.NoError(t, err)
	docSrv.AssertExpectations(t)
}

func TestService_UpdateEntity(t *testing.T) {
	ctx := context.Background()
	did := testingidentity.GenerateRandomDID()
	req := CreateEntityRequest{
		WriteAccess: []identity.DID{did},
		Data: entity.Data{
			Identity:  &did,
			LegalName: "John Doe",
			Addresses: []entity.Address{
				{
					IsMain:  true,
					Country: "Germany",
					Label:   "home",
				},
			},
		},
		Attributes: coreapi.AttributeMapRequest{
			"string_test": {
				Type:  "invalid",
				Value: "hello, world!",
			},

			"decimal_test": {
				Type:  "decimal",
				Value: "100001.001",
			},
		},
	}
	s := Service{}

	docID := utils.RandomSlice(32)

	// invalid attribute map
	_, _, err := s.UpdateEntity(ctx, docID, req)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrNotValidAttrType, err))

	// success
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	docSrv.On("UpdateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil)
	s.coreAPISrv = newCoreAPIService(docSrv)
	strAttr := req.Attributes["string_test"]
	strAttr.Type = "string"
	req.Attributes["string_test"] = strAttr
	_, _, err = s.UpdateEntity(ctx, docID, req)
	assert.NoError(t, err)
	docSrv.AssertExpectations(t)
}

func TestService_CreateInvoice(t *testing.T) {
	ctx := context.Background()
	did := testingidentity.GenerateRandomDID()
	req := CreateInvoiceRequest{
		WriteAccess: []identity.DID{did},
		Data: invoice.Data{
			Number:    "12345",
			Status:    "unpaid",
			Recipient: &did,
			Currency:  "EUR",
			Attachments: []*documents.BinaryAttachment{
				{
					Name:     "test",
					FileType: "pdf",
					Size:     1000202,
					Data:     byteutils.HexBytes(utils.RandomSlice(32)),
					Checksum: byteutils.HexBytes(utils.RandomSlice(32)),
				},
			},
		},
		Attributes: coreapi.AttributeMapRequest{
			"string_test": {
				Type:  "invalid",
				Value: "hello, world!",
			},

			"decimal_test": {
				Type:  "decimal",
				Value: "100001.001",
			},
		},
	}
	s := Service{}

	// invalid attribute map
	_, _, err := s.CreateInvoice(ctx, req)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrNotValidAttrType, err))

	// success
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	docSrv.On("CreateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil)
	s.coreAPISrv = newCoreAPIService(docSrv)
	strAttr := req.Attributes["string_test"]
	strAttr.Type = "string"
	req.Attributes["string_test"] = strAttr
	_, _, err = s.CreateInvoice(ctx, req)
	assert.NoError(t, err)
	docSrv.AssertExpectations(t)
}

func TestService_UpdateInvoice(t *testing.T) {
	ctx := context.Background()
	did := testingidentity.GenerateRandomDID()
	req := CreateInvoiceRequest{
		WriteAccess: []identity.DID{did},
		Data: invoice.Data{
			Number:    "12345",
			Status:    "unpaid",
			Recipient: &did,
			Currency:  "EUR",
			Attachments: []*documents.BinaryAttachment{
				{
					Name:     "test",
					FileType: "pdf",
					Size:     1000202,
					Data:     byteutils.HexBytes(utils.RandomSlice(32)),
					Checksum: byteutils.HexBytes(utils.RandomSlice(32)),
				},
			},
		},
		Attributes: coreapi.AttributeMapRequest{
			"string_test": {
				Type:  "invalid",
				Value: "hello, world!",
			},

			"decimal_test": {
				Type:  "decimal",
				Value: "100001.001",
			},
		},
	}
	s := Service{}
	docID := utils.RandomSlice(32)

	// invalid attribute map
	_, _, err := s.UpdateInvoice(ctx, docID, req)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrNotValidAttrType, err))

	// success
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	docSrv.On("UpdateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil)
	s.coreAPISrv = newCoreAPIService(docSrv)
	strAttr := req.Attributes["string_test"]
	strAttr.Type = "string"
	req.Attributes["string_test"] = strAttr
	_, _, err = s.UpdateInvoice(ctx, docID, req)
	assert.NoError(t, err)
	docSrv.AssertExpectations(t)
}

func TestService_ShareEntity(t *testing.T) {
	// failed to convert
	ctx := context.Background()
	s := Service{}
	_, _, err := s.ShareEntity(ctx, nil, ShareEntityRequest{})
	assert.Error(t, err)

	// success
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	s.coreAPISrv = newCoreAPIService(docSrv)
	did := testingidentity.GenerateRandomDID()
	did1 := testingidentity.GenerateRandomDID()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, did.String())
	docSrv.On("CreateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil).Once()
	docID := byteutils.HexBytes(utils.RandomSlice(32))
	m1, _, err := s.ShareEntity(ctx, docID, ShareEntityRequest{TargetIdentity: did1})
	assert.NoError(t, err)
	assert.Equal(t, m, m1)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestService_RevokeRelationship(t *testing.T) {
	// failed to convert
	ctx := context.Background()
	s := Service{}
	_, _, err := s.RevokeRelationship(ctx, nil, ShareEntityRequest{})
	assert.Error(t, err)

	// success
	docSrv := new(testingdocuments.MockService)
	m := new(testingdocuments.MockModel)
	s.coreAPISrv = newCoreAPIService(docSrv)
	did := testingidentity.GenerateRandomDID()
	did1 := testingidentity.GenerateRandomDID()
	ctx = context.WithValue(ctx, config.AccountHeaderKey, did.String())
	docSrv.On("UpdateModel", ctx, mock.Anything).Return(m, jobs.NewJobID(), nil).Once()
	docID := byteutils.HexBytes(utils.RandomSlice(32))
	m1, _, err := s.RevokeRelationship(ctx, docID, ShareEntityRequest{TargetIdentity: did1})
	assert.NoError(t, err)
	assert.Equal(t, m, m1)
	m.AssertExpectations(t)
	docSrv.AssertExpectations(t)
}

func TestService_GetRequiredInvoiceUnpaidProofFields(t *testing.T) {
	//missing account in context
	ctxh := context.Background()
	proofList, err := getRequiredInvoiceUnpaidProofFields(ctxh)
	assert.Error(t, err)
	assert.Nil(t, proofList)

	//error identity keys
	tc, err := configstore.NewAccount("main", cfg)
	assert.Nil(t, err)
	acc := tc.(*configstore.Account)
	acc.EthereumAccount = &config.AccountConfig{
		Key: "blabla",
	}
	ctxh, err = contextutil.New(ctxh, acc)
	assert.Nil(t, err)
	proofList, err = getRequiredInvoiceUnpaidProofFields(ctxh)
	assert.Error(t, err)
	assert.Nil(t, proofList)

	//success assertions
	tc, err = configstore.NewAccount("main", cfg)
	assert.Nil(t, err)
	ctxh, err = contextutil.New(ctxh, tc)
	assert.Nil(t, err)
	proofList, err = getRequiredInvoiceUnpaidProofFields(ctxh)
	assert.NoError(t, err)
	assert.Len(t, proofList, 8)
	accDIDBytes := tc.GetIdentityID()
	keys, err := tc.GetKeys()
	assert.NoError(t, err)
	signerID := hexutil.Encode(append(accDIDBytes, keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signatureSender := fmt.Sprintf("%s.signatures[%s]", documents.SignaturesTreePrefix, signerID)
	assert.Equal(t, signatureSender, proofList[6])
}
