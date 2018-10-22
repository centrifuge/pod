// +build unit

package purchaseorder

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	legacy "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/legacy/purchaseorder"
	clientpurchaseorderpb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPurchaseOrderRepository struct {
	mock.Mock
	replaceDoc *purchaseorderpb.PurchaseOrderDocument
}

func (m *mockPurchaseOrderRepository) Exists(id []byte) bool {
	args := m.Called(id)
	return args.Get(0).(bool)
}

func (m *mockPurchaseOrderRepository) GetKey(id []byte) []byte {
	args := m.Called(id)
	return args.Get(0).([]byte)
}

func (m *mockPurchaseOrderRepository) GetByID(id []byte, doc proto.Message) (err error) {
	args := m.Called(id, doc)
	order := doc.(*purchaseorderpb.PurchaseOrderDocument)
	*order = *m.replaceDoc
	doc = order
	return args.Error(0)
}

func (m *mockPurchaseOrderRepository) Create(id []byte, doc proto.Message) (err error) {
	args := m.Called(id, doc)
	return args.Error(0)
}

func (m *mockPurchaseOrderRepository) Update(id []byte, doc proto.Message) (err error) {
	args := m.Called(id, doc)
	return args.Error(0)
}

func generateMockedOutPurchaseOrderService() (srv legacy.PurchaseOrderDocumentServiceServer, repo *mockPurchaseOrderRepository, coreDocumentProcessor *testingutils.MockCoreDocumentProcessor) {
	repo = new(mockPurchaseOrderRepository)
	coreDocumentProcessor = new(testingutils.MockCoreDocumentProcessor)
	srv = LegacyGRPCHandler(repo, coreDocumentProcessor)
	return srv, repo, coreDocumentProcessor
}

func getTestSetupData() (po *PurchaseOrder, srv legacy.PurchaseOrderDocumentServiceServer, repo *mockPurchaseOrderRepository, mockCoreDocumentProcessor *testingutils.MockCoreDocumentProcessor) {
	po = &PurchaseOrder{Document: &purchaseorderpb.PurchaseOrderDocument{}}
	po.Document.Data = &purchaseorderpb.PurchaseOrderData{
		PoNumber:         "po1234",
		OrderName:        "Jack",
		OrderZipcode:     "921007",
		OrderCountry:     "Australia",
		RecipientName:    "John",
		RecipientZipcode: "12345",
		RecipientCountry: "Germany",
		Currency:         "EUR",
		OrderAmount:      800,
	}
	salts := new(purchaseorderpb.PurchaseOrderDataSalts)
	proofs.FillSalts(po.Document.Data, salts)
	po.Document.Salts = salts
	po.Document.CoreDocument = testingutils.GenerateCoreDocument()
	srv, repo, mockCoreDocumentProcessor = generateMockedOutPurchaseOrderService()
	return po, srv, repo, mockCoreDocumentProcessor
}

func TestPurchaseOrderDocumentService_Anchor(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	anchoredDoc, err := s.AnchorPurchaseOrderDocument(context.Background(), &legacy.AnchorPurchaseOrderEnvelope{Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, doc.Document.CoreDocument.DocumentIdentifier, anchoredDoc.CoreDocument.DocumentIdentifier)
}

func TestPurchaseOrderDocumentService_AnchorFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error anchoring")).Once()

	anchoredDoc, err := s.AnchorPurchaseOrderDocument(context.Background(), &legacy.AnchorPurchaseOrderEnvelope{Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, anchoredDoc)
}

func TestPurchaseOrderDocumentService_AnchorFailsWithNilDocument(t *testing.T) {
	_, s, _, _ := getTestSetupData()

	anchoredDoc, err := s.AnchorPurchaseOrderDocument(context.Background(), &legacy.AnchorPurchaseOrderEnvelope{})

	assert.Error(t, err)
	assert.Nil(t, anchoredDoc)
}

func TestPurchaseOrderDocumentService_Send(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	recipients := testingutils.GenerateP2PRecipients(1)

	coredocumentrepository.GetRepository().Create(doc.Document.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	mockCDP.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	_, err := s.SendPurchaseOrderDocument(context.Background(), &legacy.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestPurchaseOrderDocumentService_Send_StoreFails(t *testing.T) {
	doc, s, mockRepo, _ := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(errors.New("error storing")).Once()

	_, err := s.SendPurchaseOrderDocument(context.Background(), &legacy.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	assert.Contains(t, err.Error(), "error storing")
}

func TestPurchaseOrderDocumentService_Send_AnchorFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error anchoring")).Once()

	_, err := s.SendPurchaseOrderDocument(context.Background(), &legacy.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Contains(t, err.Error(), "error anchoring")
}

func TestPurchaseOrderDocumentService_SendFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	coredocumentrepository.GetRepository().Create(doc.Document.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mockCDP.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error sending")).Twice()

	_, err := s.SendPurchaseOrderDocument(context.Background(), &legacy.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Equal(t, "[1][error sending error sending]", err.Error())
}

func TestPurchaseOrderDocumentService_HandleCreatePurchaseOrderProof(t *testing.T) {
	identifier := testingutils.Rand32Bytes()
	order := Empty()
	orderAny := &any.Any{
		TypeUrl: documenttypes.PurchaseOrderDataTypeUrl,
		Value:   []byte{},
	}
	order.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        testingutils.Rand32Bytes(),
		Collaborators:      [][]byte{{1, 1, 2, 4, 5, 6}, {1, 2, 3, 2, 3, 2}},
		EmbeddedData:       orderAny,
	}
	cdSalts := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(order.Document.CoreDocument, cdSalts)
	order.Document.CoreDocument.CoredocumentSalts = cdSalts

	order.CalculateMerkleRoot()
	coredocument.CalculateDocumentRoot(order.Document.CoreDocument)
	s, mockRepo, _ := generateMockedOutPurchaseOrderService()

	proofRequest := &legacy.CreatePurchaseOrderProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "order_country", "net_amount", "collaborators[0]", "document_type"},
	}
	mockRepo.On("GetByID", proofRequest.DocumentIdentifier, new(purchaseorderpb.PurchaseOrderDocument)).Return(nil).Once()
	mockRepo.replaceDoc = order.Document
	purchaseOrderProof, err := s.CreatePurchaseOrderProof(context.Background(), proofRequest)
	mockRepo.AssertExpectations(t)

	assert.Nil(t, err)
	assert.Equal(t, identifier, purchaseOrderProof.DocumentIdentifier)
	assert.Equal(t, len(proofRequest.Fields), len(purchaseOrderProof.FieldProofs))
	assert.Equal(t, proofRequest.Fields[0], purchaseOrderProof.FieldProofs[0].Property)
	sha256Hash := sha256.New()
	fieldHash, err := proofs.CalculateHashForProofField(purchaseOrderProof.FieldProofs[0], sha256Hash)

	valid, err := proofs.ValidateProofSortedHashes(fieldHash, purchaseOrderProof.FieldProofs[0].SortedHashes, order.Document.CoreDocument.DocumentRoot, sha256Hash)
	assert.True(t, valid)
	assert.Nil(t, err)

	// Collaborators[0] proof
	fieldHash, err = proofs.CalculateHashForProofField(purchaseOrderProof.FieldProofs[3], sha256Hash)
	assert.Nil(t, err)
	valid, err = proofs.ValidateProofSortedHashes(fieldHash, purchaseOrderProof.FieldProofs[3].SortedHashes, order.Document.CoreDocument.DocumentRoot, sha256Hash)
	assert.True(t, valid)
	assert.Nil(t, err)

	// Document Type Proof
	fieldHash, err = proofs.CalculateHashForProofField(purchaseOrderProof.FieldProofs[4], sha256Hash)
	assert.Nil(t, err)
	valid, err = proofs.ValidateProofSortedHashes(fieldHash, purchaseOrderProof.FieldProofs[4].SortedHashes, order.Document.CoreDocument.DocumentRoot, sha256Hash)
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestPurchaseOrderDocumentService_HandleCreatePurchaseOrderProof_NotFilledSalts(t *testing.T) {
	identifier := testingutils.Rand32Bytes()
	order := Empty()
	order.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        testingutils.Rand32Bytes(),
		CoredocumentSalts:  &coredocumentpb.CoreDocumentSalts{},
	}
	order.Document.Salts = &purchaseorderpb.PurchaseOrderDataSalts{}
	s, mockRepo, mockProcessor := generateMockedOutPurchaseOrderService()

	proofRequest := &legacy.CreatePurchaseOrderProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "order_country", "net_amount"},
	}
	// In this test we mock out the signing root generation and return empty hashes for the CoreDocumentProcessor.GetProofHashes
	mockProcessor.On("GetDataProofHashes", order.Document.CoreDocument).Return([][]byte{}, nil).Once()
	mockRepo.On("GetByID", proofRequest.DocumentIdentifier, new(purchaseorderpb.PurchaseOrderDocument)).Return(nil).Once()
	mockRepo.replaceDoc = order.Document

	purchaseOrderProof, err := s.CreatePurchaseOrderProof(context.Background(), proofRequest)
	mockRepo.AssertExpectations(t)
	assert.NotNil(t, err)
	assert.Nil(t, purchaseOrderProof)
}

func TestPurchaseOrderDocumentService_HandleCreatePurchaseOrderProof_NotExistingPurchaseOrder(t *testing.T) {
	identifier := testingutils.Rand32Bytes()
	order := Empty()
	order.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        testingutils.Rand32Bytes(),
	}
	order.CalculateMerkleRoot()

	s, mockRepo, _ := generateMockedOutPurchaseOrderService()

	proofRequest := &legacy.CreatePurchaseOrderProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "order_country", "net_amount"},
	}
	//return an "empty" invoice as mock can't return nil
	mockRepo.On("GetByID", proofRequest.DocumentIdentifier, new(purchaseorderpb.PurchaseOrderDocument)).Return(errors.New("couldn't find invoice")).Once()
	mockRepo.replaceDoc = order.Document
	purchaseOrderProof, err := s.CreatePurchaseOrderProof(context.Background(), proofRequest)
	mockRepo.AssertExpectations(t)

	assert.Nil(t, purchaseOrderProof)
	assert.Error(t, err)
}

type mockService struct {
	Service
	mock.Mock
}

func (m mockService) Create(ctx context.Context, doc documents.Model) (documents.Model, error) {
	args := m.Called(ctx, doc)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m mockService) DeriveFromCreatePayload(req *clientpurchaseorderpb.PurchaseOrderCreatePayload, ctxh *documents.ContextHeader) (documents.Model, error) {
	args := m.Called(req, ctxh)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m mockService) DerivePurchaseOrderResponse(doc documents.Model) (*clientpurchaseorderpb.PurchaseOrderResponse, error) {
	args := m.Called(doc)
	resp, _ := args.Get(0).(*clientpurchaseorderpb.PurchaseOrderResponse)
	return resp, args.Error(1)
}

func TestGRPCHandler_Create(t *testing.T) {
	h := grpcHandler{}
	req := testingdocuments.CreatePOPayload()
	ctx := context.Background()
	model := &testingdocuments.MockModel{}
	ctxh, err := documents.NewContextHeader()
	assert.Nil(t, err)

	// derive fails
	srv := mockService{}
	srv.On("DeriveFromCreatePayload", req, ctxh).Return(nil, fmt.Errorf("derive failed")).Once()
	h.service = srv
	resp, err := h.Create(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive failed")

	// create fails
	srv = mockService{}
	srv.On("DeriveFromCreatePayload", req, ctxh).Return(model, nil).Once()
	srv.On("Create", ctx, model).Return(nil, fmt.Errorf("create failed")).Once()
	h.service = srv
	resp, err = h.Create(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create failed")

	// derive response fails
	srv = mockService{}
	srv.On("DeriveFromCreatePayload", req, ctxh).Return(model, nil).Once()
	srv.On("Create", ctx, model).Return(model, nil).Once()
	srv.On("DerivePurchaseOrderResponse", model).Return(nil, fmt.Errorf("derive response fails")).Once()
	h.service = srv
	resp, err = h.Create(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derive response fails")

	// success
	eresp := &clientpurchaseorderpb.PurchaseOrderResponse{}
	srv = mockService{}
	srv.On("DeriveFromCreatePayload", req, ctxh).Return(model, nil).Once()
	srv.On("Create", ctx, model).Return(model, nil).Once()
	srv.On("DerivePurchaseOrderResponse", model).Return(eresp, nil).Once()
	h.service = srv
	resp, err = h.Create(ctx, req)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, eresp, resp)
}
