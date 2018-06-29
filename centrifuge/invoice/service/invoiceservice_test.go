// +build unit

package invoiceservice

import (
	"context"
	"crypto/sha256"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/go-errors/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"github.com/stretchr/testify/mock"
	"fmt"
)

func TestMain(m *testing.M) {
	cc.TestUnitBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

// ----- MOCKS -----
// Allows mocking out the storage to have no dependencies during the unit testing phase
type MockInvoiceRepository struct {
	mock.Mock
}

func (m *MockInvoiceRepository) GetKey(id []byte) ([]byte) {
	args := m.Called(id)
	return args.Get(0).([]byte)
}
func (m *MockInvoiceRepository) FindById(id []byte) (inv *invoicepb.InvoiceDocument, err error) {
	args := m.Called(id)
	return args.Get(0).(*invoicepb.InvoiceDocument) ,args.Error(1)
}
func (m *MockInvoiceRepository) Store(inv *invoicepb.InvoiceDocument) (err error) {
	args := m.Called(inv)
	return args.Error(0)
}

type MockCoreDocumentSender struct {
	mock.Mock
}
func (m *MockCoreDocumentSender) Send(coreDocument *coredocumentpb.CoreDocument, ctx context.Context, recipient string) (err error) {
	args := m.Called(coreDocument, ctx, recipient)
	return args.Error(0)
}
// ----- END MOCKS -----

// ----- HELPER FUNCTIONS -----
func generateP2PRecipients(quantity int) ([][]byte) {
	recipients := make([][]byte, quantity)

	for i := 0; i < quantity; i++ {
		recipients[0] = []byte(fmt.Sprintf("RecipientNo[%d]", quantity))
	}
	return recipients
}

func generateSendableInvoice() (*invoice.Invoice) {
	identifier := testingutils.Rand32Bytes()
	doc := invoice.NewEmptyInvoice()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		DataMerkleRoot:     testingutils.Rand32Bytes(),
	}
	return doc
}
// ----- END HELPER FUNCTIONS -----


// ----- TESTS -----
func TestInvoiceDocumentService_Send(t *testing.T) {
	mockSender := new(MockCoreDocumentSender)
	mockRepo := new(MockInvoiceRepository)
	s := InvoiceDocumentService{
		InvoiceRepository:  mockRepo,
		CoreDocumentSender: mockSender,
	}

	doc := generateSendableInvoice()
	recipients := generateP2PRecipients(1)

	mockRepo.On("Store", doc.Document).Return(nil).Once()
	mockSender.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	_, err := s.HandleSendInvoiceDocument(context.Background(), &invoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})
	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestInvoiceDocumentService_SendFails(t *testing.T) {
	mockSender := new(MockCoreDocumentSender)
	mockRepo := new(MockInvoiceRepository)
	s := InvoiceDocumentService{
		InvoiceRepository:  mockRepo,
		CoreDocumentSender: mockSender,
	}
	doc := generateSendableInvoice()
	recipients := generateP2PRecipients(2)

	mockRepo.On("Store", doc.Document).Return(nil).Once()
	sendError := errors.New("error sending")
	mockSender.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(sendError).Twice()
	_, err := s.HandleSendInvoiceDocument(context.Background(), &invoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})
	mockSender.AssertExpectations(t)

	//the error handling in the send handler simply prints out the list of errors without much formatting
	//OK for now but could be done nicer in the future
	assert.Equal(t, "[error sending error sending]", err.Error())
}


func TestInvoiceDocumentService_HandleCreateInvoiceProof(t *testing.T) {
	identifier := testingutils.Rand32Bytes()
	inv := invoice.NewEmptyInvoice()
	inv.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		// TODO: below should be actual merkle root
		DataMerkleRoot: testingutils.Rand32Bytes(),
	}
	inv.CalculateMerkleRoot()

	//mock the storage
	mockRepo := new(MockInvoiceRepository)
	s := InvoiceDocumentService{
		InvoiceRepository: mockRepo,
	}

	proofRequest := &invoicepb.CreateInvoiceProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "sender_country", "gross_amount"},
	}
	mockRepo.On("FindById", proofRequest.DocumentIdentifier).Return(inv.Document, nil).Once()

	invoiceProof, err := s.HandleCreateInvoiceProof(context.Background(), proofRequest)
	mockRepo.AssertExpectations(t)

	assert.Nil(t, err)
	assert.Equal(t, identifier, invoiceProof.DocumentIdentifier)
	assert.Equal(t, len(proofRequest.Fields), len(invoiceProof.FieldProofs))
	assert.Equal(t, proofRequest.Fields[0], invoiceProof.FieldProofs[0].Property)
	sha256Hash := sha256.New()
	valid, err := proofs.ValidateProof(invoiceProof.FieldProofs[0], inv.Document.CoreDocument.DocumentRoot, sha256Hash)
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestInvoiceDocumentService_HandleCreateInvoiceProof_NotExistingInvoice(t *testing.T) {
	identifier := testingutils.Rand32Bytes()
	inv := invoice.NewEmptyInvoice()
	inv.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		// TODO: below should be actual merkle root
		DataMerkleRoot: testingutils.Rand32Bytes(),
	}
	inv.CalculateMerkleRoot()

	//mock the storage
	mockRepo := new(MockInvoiceRepository)
	s := InvoiceDocumentService{
		InvoiceRepository: mockRepo,
	}

	proofRequest := &invoicepb.CreateInvoiceProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "sender_country", "gross_amount"},
	}
	//return an "empty" invoice as mock can't return nil
	mockRepo.On("FindById", proofRequest.DocumentIdentifier).Return(invoice.NewEmptyInvoice().Document, errors.New("couldn't find invoice")).Once()

	invoiceProof, err := s.HandleCreateInvoiceProof(context.Background(), proofRequest)
	mockRepo.AssertExpectations(t)

	assert.Nil(t, invoiceProof)
	assert.Error(t, err)
}
