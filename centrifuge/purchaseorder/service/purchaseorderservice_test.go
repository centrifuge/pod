// +build unit

package purchaseorderservice

import (
	"context"
	"crypto/sha256"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/go-errors/errors"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cc.TestUnitBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

// ----- MOCKS -----
type MockPurchaseOrderRepository struct {
	mock.Mock
}

func (m *MockPurchaseOrderRepository) GetKey(id []byte) ([]byte) {
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

// ----- END MOCKS -----

// ----- HELPER FUNCTIONS -----
func generateSendablePurchaseOrder() (*purchaseorder.PurchaseOrder) {
	doc := purchaseorder.NewEmptyPurchaseOrder()
	doc.Document.CoreDocument = testingutils.GenerateCoreDocument()
	return doc
}

func generateMockedOutPurchaseOrderService() (srv *PurchaseOrderDocumentService, repo *MockPurchaseOrderRepository, sender *testingutils.MockCoreDocumentSender, anchorer *testingutils.MockCoreDocumentAnchorer) {
	repo = new(MockPurchaseOrderRepository)
	sender = new(testingutils.MockCoreDocumentSender)
	anchorer = new(testingutils.MockCoreDocumentAnchorer)
	srv = &PurchaseOrderDocumentService{
		PurchaseOrderRepository: repo,
		CoreDocumentSender:      sender,
		CoreDocumentAnchorer:    anchorer,
	}
	return
}

// ----- END HELPER FUNCTIONS -----

func TestPurchaseOrderDocumentService_Send(t *testing.T) {
	s, mockRepo, mockSender, _ := generateMockedOutPurchaseOrderService()

	doc := generateSendablePurchaseOrder()
	recipients := testingutils.GenerateP2PRecipients(1)

	mockRepo.On("Store", doc.Document).Return(nil).Once()
	mockSender.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	_, err := s.HandleSendPurchaseOrderDocument(context.Background(), &purchaseorderpb.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestPurchaseOrderDocumentService_SendFails(t *testing.T) {
	s, mockRepo, mockSender, _ := generateMockedOutPurchaseOrderService()

	doc := generateSendablePurchaseOrder()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("Store", doc.Document).Return(nil).Once()
	mockSender.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error sending")).Twice()

	_, err := s.HandleSendPurchaseOrderDocument(context.Background(), &purchaseorderpb.SendPurchaseOrderEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
	assert.Equal(t, "[error sending error sending]", err.Error())
}

func TestPurchaseOrderDocumentService_HandleCreatePurchaseOrderProof(t *testing.T) {
	s, mockRepo, _, _ := generateMockedOutPurchaseOrderService()

	identifier := testingutils.Rand32Bytes()
	order := purchaseorder.NewEmptyPurchaseOrder()
	order.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		// TODO: below should be actual merkle root
		DataMerkleRoot: testingutils.Rand32Bytes(),
	}
	order.CalculateMerkleRoot()


	proofRequest := &purchaseorderpb.CreatePurchaseOrderProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "country", "amount"},
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
