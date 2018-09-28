// +build unit

package invoice

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockInvoiceRepository implements storage.Repository
type mockInvoiceRepository struct {
	mock.Mock
	replaceDoc *invoicepb.InvoiceDocument
}

func (m *mockInvoiceRepository) Exists(id []byte) bool {
	args := m.Called(id)
	return args.Get(0).(bool)
}

func (m *mockInvoiceRepository) GetKey(id []byte) []byte {
	args := m.Called(id)
	return args.Get(0).([]byte)
}

func (m *mockInvoiceRepository) GetByID(id []byte, doc proto.Message) (err error) {
	args := m.Called(id, doc)
	order := doc.(*invoicepb.InvoiceDocument)
	*order = *m.replaceDoc
	doc = order
	return args.Error(0)
}

func (m *mockInvoiceRepository) Create(id []byte, doc proto.Message) (err error) {
	args := m.Called(id, doc)
	return args.Error(0)
}

func (m *mockInvoiceRepository) Update(id []byte, doc proto.Message) (err error) {
	args := m.Called(id, doc)
	return args.Error(0)
}

func getMockedHandler() (handler *grpcHandler, repo *mockInvoiceRepository, coreDocumentProcessor *testingutils.MockCoreDocumentProcessor) {
	repo = new(mockInvoiceRepository)
	coreDocumentProcessor = new(testingutils.MockCoreDocumentProcessor)
	handler = &grpcHandler{
		Repository:            repo,
		CoreDocumentProcessor: coreDocumentProcessor,
	}
	return handler, repo, coreDocumentProcessor
}
func getTestSetupData() (doc *Invoice, srv *grpcHandler, repo *mockInvoiceRepository, coreDocumentProcessor *testingutils.MockCoreDocumentProcessor) {
	doc = &Invoice{Document: &invoicepb.InvoiceDocument{}}
	doc.Document.Data = &invoicepb.InvoiceData{
		InvoiceNumber:    "inv1234",
		SenderName:       "Jack",
		SenderZipcode:    "921007",
		SenderCountry:    "AUS",
		RecipientName:    "John",
		RecipientZipcode: "12345",
		RecipientCountry: "Germany",
		Currency:         "EUR",
		GrossAmount:      800,
	}
	salts := new(invoicepb.InvoiceDataSalts)
	proofs.FillSalts(doc.Document.Data, salts)
	doc.Document.CoreDocument = testingutils.GenerateCoreDocument()
	doc.Document.Salts = salts
	srv, repo, coreDocumentProcessor = getMockedHandler()
	return doc, srv, repo, coreDocumentProcessor
}

func TestInvoiceDocumentService_Anchor(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockRepo.On("Update", doc.Document.CoreDocument.DocumentIdentifier, mock.Anything).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(nil).Once()

	anchoredDoc, err := s.AnchorInvoiceDocument(context.Background(), &clientinvoicepb.AnchorInvoiceEnvelope{Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, doc.Document.CoreDocument.DocumentIdentifier, anchoredDoc.CoreDocument.DocumentIdentifier)
}

func TestInvoiceDocumentService_AnchorFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(errors.New("error anchors")).Once()

	anchoredDoc, err := s.AnchorInvoiceDocument(context.Background(), &clientinvoicepb.AnchorInvoiceEnvelope{Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, anchoredDoc)
}

func TestInvoiceDocumentService_AnchorFailsWithNilDocument(t *testing.T) {
	_, s, _, _ := getTestSetupData()

	anchoredDoc, err := s.AnchorInvoiceDocument(context.Background(), &clientinvoicepb.AnchorInvoiceEnvelope{})

	assert.Error(t, err)
	assert.Nil(t, anchoredDoc)
}

func TestInvoiceDocumentService_Send(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	recipients := testingutils.GenerateP2PRecipients(1)

	coredocumentrepository.GetRepository().Create(doc.Document.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockRepo.On("Update", doc.Document.CoreDocument.DocumentIdentifier, mock.Anything).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(nil).Once()
	mockCDP.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	_, err := s.SendInvoiceDocument(context.Background(), &clientinvoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestInvoiceDocumentService_SendFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	coredocumentrepository.GetRepository().Create(doc.Document.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockRepo.On("Update", doc.Document.CoreDocument.DocumentIdentifier, mock.Anything).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(nil).Once()
	mockCDP.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("error sending")).Twice()

	_, err := s.SendInvoiceDocument(context.Background(), &clientinvoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})

	mockCDP.AssertExpectations(t)
	//the error handling in the send handler simply prints out the list of errors without much formatting
	//OK for now but could be done nicer in the future
	assert.Contains(t, err.Error(), "error sending error sending")
}

func TestInvoiceDocumentService_Send_StoreFails(t *testing.T) {
	doc, s, mockRepo, _ := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(errors.New("error storing")).Once()

	_, err := s.SendInvoiceDocument(context.Background(), &clientinvoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	assert.Contains(t, err.Error(), "error storing")
}

func TestInvoiceDocumentService_Send_AnchorFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(errors.New("error anchors")).Once()

	_, err := s.SendInvoiceDocument(context.Background(), &clientinvoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Contains(t, err.Error(), "error anchors")
}

func TestInvoiceDocumentService_HandleCreateInvoiceProof(t *testing.T) {
	identifier := testingutils.Rand32Bytes()
	inv := Empty()
	inv.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
	}
	cdSalts := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(inv.Document.CoreDocument, cdSalts)
	inv.Document.CoreDocument.CoredocumentSalts = cdSalts

	inv.CalculateMerkleRoot()
	coredocument.CalculateDocumentRoot(inv.Document.CoreDocument)
	s, mockRepo, _ := getMockedHandler()

	proofRequest := &clientinvoicepb.CreateInvoiceProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "sender_country", "gross_amount"},
	}
	mockRepo.On("GetByID", proofRequest.DocumentIdentifier, new(invoicepb.InvoiceDocument)).Return(nil).Once()
	mockRepo.replaceDoc = inv.Document
	invoiceProof, err := s.CreateInvoiceProof(context.Background(), proofRequest)
	mockRepo.AssertExpectations(t)

	assert.Nil(t, err)
	assert.Equal(t, identifier, invoiceProof.DocumentIdentifier)
	assert.Equal(t, len(proofRequest.Fields), len(invoiceProof.FieldProofs))
	assert.Equal(t, proofRequest.Fields[0], invoiceProof.FieldProofs[0].Property)
	sha256Hash := sha256.New()
	fieldHash, err := proofs.CalculateHashForProofField(invoiceProof.FieldProofs[0], sha256Hash)

	valid, err := proofs.ValidateProofSortedHashes(fieldHash, invoiceProof.FieldProofs[0].SortedHashes, inv.Document.CoreDocument.DocumentRoot, sha256Hash)
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestInvoiceDocumentService_HandleCreateInvoiceProof_NotFilledSalts(t *testing.T) {
	identifier := testingutils.Rand32Bytes()
	inv := Empty()
	inv.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		CoredocumentSalts:  &coredocumentpb.CoreDocumentSalts{},
	}
	inv.Document.Salts = &invoicepb.InvoiceDataSalts{}
	s, mockRepo, mockProcessor := getMockedHandler()

	proofRequest := &clientinvoicepb.CreateInvoiceProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "sender_country", "gross_amount"},
	}
	// In this test we mock out the signing root generation and return empty hashes for the CoreDocumentProcessor.GetProofHashes
	mockProcessor.On("GetDataProofHashes", inv.Document.CoreDocument).Return([][]byte{}, nil).Once()
	mockRepo.On("GetByID", proofRequest.DocumentIdentifier, new(invoicepb.InvoiceDocument)).Return(nil).Once()
	mockRepo.replaceDoc = inv.Document

	invoiceProof, err := s.CreateInvoiceProof(context.Background(), proofRequest)
	mockRepo.AssertExpectations(t)
	assert.NotNil(t, err)
	assert.Nil(t, invoiceProof)
}

func TestInvoiceDocumentService_HandleCreateInvoiceProof_NotExistingInvoice(t *testing.T) {
	identifier := testingutils.Rand32Bytes()
	inv := Empty()
	inv.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
	}
	inv.CalculateMerkleRoot()

	s, mockRepo, _ := getMockedHandler()

	proofRequest := &clientinvoicepb.CreateInvoiceProofEnvelope{
		DocumentIdentifier: identifier,
		Fields:             []string{"currency", "sender_country", "gross_amount"},
	}
	//return an "empty" invoice as mock can't return nil
	mockRepo.On("GetByID", proofRequest.DocumentIdentifier, new(invoicepb.InvoiceDocument)).Return(errors.New("couldn't find invoice")).Once()
	mockRepo.replaceDoc = inv.Document
	invoiceProof, err := s.CreateInvoiceProof(context.Background(), proofRequest)
	mockRepo.AssertExpectations(t)

	assert.Nil(t, invoiceProof)
	assert.Error(t, err)
}
