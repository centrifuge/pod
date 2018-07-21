// +build unit

package purchaseorderservice

import (
	"context"
	"crypto/sha256"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

// ----- MOCKS -----
type MockPurchaseOrderRepository struct {
	mock.Mock
}

func (m *MockPurchaseOrderRepository) GetKey(id []byte) []byte {
	args := m.Called(id)
	return args.Get(0).([]byte)
}
func (m *MockPurchaseOrderRepository) FindById(id []byte) (doc *purchaseorderpb.PurchaseOrderDocument, err error) {
	args := m.Called(id)
	return args.Get(0).(*purchaseorderpb.PurchaseOrderDocument), args.Error(1)
}
func (m *MockPurchaseOrderRepository) Store(doc *purchaseorderpb.PurchaseOrderDocument) (err error) {
	args := m.Called(doc)
	return args.Error(0)
}
func (m *MockPurchaseOrderRepository) StoreOnce(doc *purchaseorderpb.PurchaseOrderDocument) (err error) {
	args := m.Called(doc)
	return args.Error(0)
}

// ----- END MOCKS -----

// ----- HELPER FUNCTIONS -----
func generateSendablePurchaseOrder() *purchaseorder.PurchaseOrder {
	doc := purchaseorder.NewEmptyPurchaseOrder()
	doc.Document.CoreDocument = testingutils.GenerateCoreDocument()
	return doc
}

func generateMockedOutPurchaseOrderService() (srv *PurchaseOrderDocumentService, repo *MockPurchaseOrderRepository, coreDocumentProcessor *testingutils.MockCoreDocumentProcessor) {
	repo = new(MockPurchaseOrderRepository)
	coreDocumentProcessor = new(testingutils.MockCoreDocumentProcessor)
	srv = &PurchaseOrderDocumentService{
		PurchaseOrderRepository: repo,
		CoreDocumentProcessor:   coreDocumentProcessor,
	}
	return
}

func getTestSetupData() (po *purchaseorder.PurchaseOrder, srv *PurchaseOrderDocumentService, repo *MockPurchaseOrderRepository, mockCoreDocumentProcessor *testingutils.MockCoreDocumentProcessor) {
	po = generateSendablePurchaseOrder()
	srv, repo, mockCoreDocumentProcessor = generateMockedOutPurchaseOrderService()
	return
}

// ----- END HELPER FUNCTIONS -----
func TestPurchaseOrderDocumentService_Anchor(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	mockRepo.On("StoreOnce", doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(nil).Once()

	anchoredDoc, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &purchaseorderpb.AnchorPurchaseOrderEnvelope{Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, doc.Document.CoreDocument.DocumentIdentifier, anchoredDoc.CoreDocument.DocumentIdentifier)
}

func TestPurchaseOrderDocumentService_AnchorFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	mockRepo.On("StoreOnce", doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(errors.New("error anchoring")).Once()

	anchoredDoc, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &purchaseorderpb.AnchorPurchaseOrderEnvelope{Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, anchoredDoc)
}

func TestPurchaseOrderDocumentService_AnchorFailsWithNilDocument(t *testing.T) {
	_, s, _, _ := getTestSetupData()

	anchoredDoc, err := s.HandleAnchorPurchaseOrderDocument(context.Background(), &purchaseorderpb.AnchorPurchaseOrderEnvelope{})

	assert.Error(t, err)
	assert.Nil(t, anchoredDoc)
}

func TestPurchaseOrderDocumentService_Send(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	recipients := testingutils.GenerateP2PRecipients(1)

	mockRepo.On("StoreOnce", doc.Document).Return(nil).Once()
	mockCDP.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(nil).Once()

	_, err := s.HandleSendPurchaseOrderDocument(context.Background(), &purchaseorderpb.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestPurchaseOrderDocumentService_Send_StoreFails(t *testing.T) {
	doc, s, mockRepo, _ := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("StoreOnce", doc.Document).Return(errors.New("error storing")).Once()

	_, err := s.HandleSendPurchaseOrderDocument(context.Background(), &purchaseorderpb.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	assert.Equal(t, "error storing", err.Error())
}

func TestPurchaseOrderDocumentService_Send_AnchorFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("StoreOnce", doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(errors.New("error anchoring")).Once()

	_, err := s.HandleSendPurchaseOrderDocument(context.Background(), &purchaseorderpb.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Equal(t, "error anchoring", err.Error())
}

func TestPurchaseOrderDocumentService_SendFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("StoreOnce", doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(nil).Once()
	mockCDP.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error sending")).Twice()

	_, err := s.HandleSendPurchaseOrderDocument(context.Background(), &purchaseorderpb.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Equal(t, "[error sending error sending]", err.Error())
}

func TestPurchaseOrderDocumentService_HandleCreatePurchaseOrderProof(t *testing.T) {
	s, mockRepo, _ := generateMockedOutPurchaseOrderService()

	identifier := testingutils.Rand32Bytes()
	order := purchaseorder.NewEmptyPurchaseOrder()
	order.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		// TODO: below should be actual merkle root
		DataRoot: testingutils.Rand32Bytes(),
	}
	order.CalculateMerkleRoot()

	proofRequest := &purchaseorderpb.CreatePurchaseOrderProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "order_country", "net_amount"},
	}

	mockRepo.On("FindById", proofRequest.DocumentIdentifier).Return(order.Document, nil).Once()

	purchaseorderProof, err := s.HandleCreatePurchaseOrderProof(context.Background(), proofRequest)

	mockRepo.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, identifier, purchaseorderProof.DocumentIdentifier)
	assert.Equal(t, len(proofRequest.Fields), len(purchaseorderProof.FieldProofs))
	assert.Equal(t, proofRequest.Fields[0], purchaseorderProof.FieldProofs[0].Property)
	sha256Hash := sha256.New()
	valid, err := proofs.ValidateProof(purchaseorderProof.FieldProofs[0], order.Document.CoreDocument.DocumentRoot, sha256Hash)
	assert.True(t, valid)
	assert.Nil(t, err)
}
