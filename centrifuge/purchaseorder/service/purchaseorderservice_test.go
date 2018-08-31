// +build unit

package purchaseorderservice

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	clientpurchaseorderpb "github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

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

func generateMockedOutPurchaseOrderService() (srv *PurchaseOrderDocumentService, repo *mockPurchaseOrderRepository, coreDocumentProcessor *testingutils.MockCoreDocumentProcessor) {
	repo = new(mockPurchaseOrderRepository)
	coreDocumentProcessor = new(testingutils.MockCoreDocumentProcessor)
	srv = &PurchaseOrderDocumentService{
		Repository:            repo,
		CoreDocumentProcessor: coreDocumentProcessor,
	}
	return
}

func getTestSetupData() (po *purchaseorder.PurchaseOrder, srv *PurchaseOrderDocumentService, repo *mockPurchaseOrderRepository, mockCoreDocumentProcessor *testingutils.MockCoreDocumentProcessor) {
	po = &purchaseorder.PurchaseOrder{Document: &purchaseorderpb.PurchaseOrderDocument{}}
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
	proofs.FillSalts(salts)
	po.Document.Salts = salts
	po.Document.CoreDocument = testingutils.GenerateCoreDocument()
	srv, repo, mockCoreDocumentProcessor = generateMockedOutPurchaseOrderService()
	return
}

func TestPurchaseOrderDocumentService_Anchor(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(nil).Once()

	anchoredDoc, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.AnchorPurchaseOrderEnvelope{Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, doc.Document.CoreDocument.DocumentIdentifier, anchoredDoc.CoreDocument.DocumentIdentifier)
}

func TestPurchaseOrderDocumentService_AnchorFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(errors.New("error anchoring")).Once()

	anchoredDoc, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.AnchorPurchaseOrderEnvelope{Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, anchoredDoc)
}

func TestPurchaseOrderDocumentService_AnchorFailsWithNilDocument(t *testing.T) {
	_, s, _, _ := getTestSetupData()

	anchoredDoc, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.AnchorPurchaseOrderEnvelope{})

	assert.Error(t, err)
	assert.Nil(t, anchoredDoc)
}

func TestPurchaseOrderDocumentService_Send(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	recipients := testingutils.GenerateP2PRecipients(1)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	mockCDP.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(nil).Once()

	_, err := s.HandleSendPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestPurchaseOrderDocumentService_Send_StoreFails(t *testing.T) {
	doc, s, mockRepo, _ := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(errors.New("error storing")).Once()

	_, err := s.HandleSendPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	assert.Contains(t, err.Error(), "error storing")
}

func TestPurchaseOrderDocumentService_Send_AnchorFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(errors.New("error anchoring")).Once()

	_, err := s.HandleSendPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Contains(t, err.Error(), "error anchoring")
}

func TestPurchaseOrderDocumentService_SendFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(nil).Once()
	mockCDP.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error sending")).Twice()

	_, err := s.HandleSendPurchaseOrderDocument(context.Background(), &clientpurchaseorderpb.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Equal(t, "[1][error sending error sending]", err.Error())
}

func TestPurchaseOrderDocumentService_HandleCreatePurchaseOrderProof(t *testing.T) {
	s, mockRepo, _ := generateMockedOutPurchaseOrderService()

	identifier := testingutils.Rand32Bytes()
	order := purchaseorder.Empty()
	order.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
	}
	order.CalculateMerkleRoot()
	mockRepo.replaceDoc = order.Document

	proofRequest := &clientpurchaseorderpb.CreatePurchaseOrderProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "order_country", "net_amount"},
	}

	mockRepo.On("GetByID", proofRequest.DocumentIdentifier, new(purchaseorderpb.PurchaseOrderDocument)).Return(nil).Once()
	purchaseOrderProof, err := s.HandleCreatePurchaseOrderProof(context.Background(), proofRequest)
	mockRepo.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, identifier, purchaseOrderProof.DocumentIdentifier)
	assert.Equal(t, len(proofRequest.Fields), len(purchaseOrderProof.FieldProofs))
	assert.Equal(t, proofRequest.Fields[0], purchaseOrderProof.FieldProofs[0].Property)
	sha256Hash := sha256.New()
	fmt.Println(order.Document.CoreDocument.DataRoot)
	valid, err := proofs.ValidateProof(purchaseOrderProof.FieldProofs[0], order.Document.CoreDocument.DataRoot, sha256Hash)
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestInvoiceDocumentService_HandleCreatePurchaseOrderProof_NotFilledSalts(t *testing.T) {
	s, mockRepo, _ := generateMockedOutPurchaseOrderService()

	identifier := testingutils.Rand32Bytes()
	order := purchaseorder.Empty()
	order.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
	}
	mockRepo.replaceDoc = order.Document
	order.Document.Salts = &purchaseorderpb.PurchaseOrderDataSalts{}
	proofRequest := &clientpurchaseorderpb.CreatePurchaseOrderProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "order_country", "net_amount"},
	}

	mockRepo.On("GetByID", proofRequest.DocumentIdentifier, new(purchaseorderpb.PurchaseOrderDocument)).Return(nil).Once()
	purchaseOrderProof, err := s.HandleCreatePurchaseOrderProof(context.Background(), proofRequest)
	mockRepo.AssertExpectations(t)
	assert.NotNil(t, err)
	assert.Nil(t, purchaseOrderProof)
}
