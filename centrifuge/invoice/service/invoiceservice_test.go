// +build unit

package invoiceservice

import (
	"context"
	"crypto/sha256"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	cc.TestUnitBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

// Allows mocking out the storage to have no dependencies during the unit testing phase
type MockInvoiceRepository struct {
	mock.Mock
	returnFindInvoice *invoice.Invoice
}

func (m *MockInvoiceRepository) GetKey(id []byte) ([]byte) {
	return id
}
func (m *MockInvoiceRepository) FindById(id []byte) (inv *invoicepb.InvoiceDocument, err error) {
	return m.returnFindInvoice.Document, nil
}
func (m *MockInvoiceRepository) Store(inv *invoicepb.InvoiceDocument) (err error) {
	return nil
}

type MockCoreDocumentSender struct {
	mock.Mock
}

func (m *MockCoreDocumentSender) Send(cd *coredocument.CoreDocument, ctx context.Context, recipient string) (err error) {
	args := m.Called(cd, ctx, recipient)
	return args.Error(0)
}

func TestInvoiceDocumentService_Send(t *testing.T) {
	mockSender := new(MockCoreDocumentSender)
	s := InvoiceDocumentService{
		InvoiceRepository: new(MockInvoiceRepository),
		CoreDocumentSender: mockSender,
	}

	identifier := testingutils.Rand32Bytes()
	doc := invoice.NewEmptyInvoice()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		DataMerkleRoot:     testingutils.Rand32Bytes(),
	}

	mockSender.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	recipients := make([][]byte, 1)
	recipients[0] = []byte("abcd")
	sentDoc, err := s.HandleSendInvoiceDocument(context.Background(), &invoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})
	mockSender.AssertExpectations(t)
	assert.Nil(t, err)

	assert.Equal(t, sentDoc.CoreDocument.DocumentIdentifier, identifier,
		"DocumentIdentifier doesn't match")
}

func TestInvoiceDocumentService_HandleCreateInvoiceProof(t *testing.T) {
	mockRepo := new(MockInvoiceRepository)

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
	mockRepo.returnFindInvoice = inv
	s := InvoiceDocumentService{
		InvoiceRepository: mockRepo,
	}

	proofRequest := &invoicepb.CreateInvoiceProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "sender_country", "gross_amount"},
	}

	invoiceProof, err := s.HandleCreateInvoiceProof(context.Background(), proofRequest)
	assert.Nil(t, err)
	assert.Equal(t, identifier, invoiceProof.DocumentIdentifier)
	assert.Equal(t, len(proofRequest.Fields), len(invoiceProof.FieldProofs))
	assert.Equal(t, proofRequest.Fields[0], invoiceProof.FieldProofs[0].Property)
	sha256Hash := sha256.New()
	valid, err := proofs.ValidateProof(invoiceProof.FieldProofs[0], inv.Document.CoreDocument.DocumentRoot, sha256Hash)
	assert.True(t, valid)
	assert.Nil(t, err)
}
