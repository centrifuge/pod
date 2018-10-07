// +build unit

package invoice

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	legacyinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/legacy/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockInvoiceRepository implements storage.legacyRepo
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
		legacyRepo:       repo,
		coreDocProcessor: coreDocumentProcessor,
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

	anchoredDoc, err := s.AnchorInvoiceDocument(context.Background(), &legacyinvoicepb.AnchorInvoiceEnvelope{Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, doc.Document.CoreDocument.DocumentIdentifier, anchoredDoc.CoreDocument.DocumentIdentifier)
}

func TestInvoiceDocumentService_AnchorFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(errors.New("error anchors")).Once()

	anchoredDoc, err := s.AnchorInvoiceDocument(context.Background(), &legacyinvoicepb.AnchorInvoiceEnvelope{Document: doc.Document})

	mockRepo.AssertExpectations(t)
	mockCDP.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, anchoredDoc)
}

func TestInvoiceDocumentService_AnchorFailsWithNilDocument(t *testing.T) {
	_, s, _, _ := getTestSetupData()

	anchoredDoc, err := s.AnchorInvoiceDocument(context.Background(), &legacyinvoicepb.AnchorInvoiceEnvelope{})

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

	_, err := s.SendInvoiceDocument(context.Background(), &legacyinvoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})

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

	_, err := s.SendInvoiceDocument(context.Background(), &legacyinvoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})

	mockCDP.AssertExpectations(t)
	//the error handling in the send handler simply prints out the list of errors without much formatting
	//OK for now but could be done nicer in the future
	assert.Contains(t, err.Error(), "error sending error sending")
}

func TestInvoiceDocumentService_Send_StoreFails(t *testing.T) {
	doc, s, mockRepo, _ := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(errors.New("error storing")).Once()

	_, err := s.SendInvoiceDocument(context.Background(), &legacyinvoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})

	mockRepo.AssertExpectations(t)
	assert.Contains(t, err.Error(), "error storing")
}

func TestInvoiceDocumentService_Send_AnchorFails(t *testing.T) {
	doc, s, mockRepo, mockCDP := getTestSetupData()
	recipients := testingutils.GenerateP2PRecipients(2)

	mockRepo.On("Create", doc.Document.CoreDocument.DocumentIdentifier, doc.Document).Return(nil).Once()
	mockCDP.On("Anchor", mock.Anything).Return(errors.New("error anchors")).Once()

	_, err := s.SendInvoiceDocument(context.Background(), &legacyinvoicepb.SendInvoiceEnvelope{Recipients: recipients, Document: doc.Document})

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

	proofRequest := &legacyinvoicepb.CreateInvoiceProofEnvelope{
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

	proofRequest := &legacyinvoicepb.CreateInvoiceProofEnvelope{
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

	proofRequest := &legacyinvoicepb.CreateInvoiceProofEnvelope{
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

type mockService struct {
	Service
	mock.Mock
}

func (m *mockService) DeriveFromCreatePayload(payload *clientinvoicepb.InvoiceCreatePayload) (documents.Model, error) {
	args := m.Called(payload)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *mockService) Create(inv documents.Model) error {
	args := m.Called(inv)
	return args.Error(0)
}

func (m *mockService) GetLastVersion(identifier []byte) (documents.Model, error) {
	args := m.Called(identifier)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m *mockService) GetVersion(identifier []byte, version []byte) (documents.Model, error) {
	args := m.Called(identifier, version)
	data, _ := args.Get(0).(documents.Model)
	return data, args.Error(1)
}

func (m *mockService) DeriveInvoiceData(doc documents.Model) (*clientinvoicepb.InvoiceData, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*clientinvoicepb.InvoiceData)
	return data, args.Error(1)
}

func (m *mockService) DeriveInvoiceResponse(doc documents.Model) (*clientinvoicepb.InvoiceResponse, error) {
	args := m.Called(doc)
	data, _ := args.Get(0).(*clientinvoicepb.InvoiceResponse)
	return data, args.Error(1)
}

func getHandler() *grpcHandler {
	return &grpcHandler{service: &mockService{}}
}

func TestGRPCHandler_Create_derive_fail(t *testing.T) {
	// DeriveFrom payload fails
	h := getHandler()
	srv := h.service.(*mockService)
	srv.On("DeriveFromCreatePayload", mock.Anything).Return(nil, fmt.Errorf("derive failed")).Once()
	_, err := h.Create(context.Background(), nil)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "derive failed")
}

func TestGrpcHandler_Create_create_fail(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	srv.On("DeriveFromCreatePayload", mock.Anything).Return(new(InvoiceModel), nil).Once()
	srv.On("Create", mock.Anything).Return(fmt.Errorf("create failed")).Once()
	payload := &clientinvoicepb.InvoiceCreatePayload{Data: &clientinvoicepb.InvoiceData{GrossAmount: 300}}
	_, err := h.Create(context.Background(), payload)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "create failed")
}

type mockModel struct {
	documents.Model
	mock.Mock
	CoreDocument *coredocumentpb.CoreDocument
}

func (m *mockModel) PackCoreDocument() (*coredocumentpb.CoreDocument, error) {
	args := m.Called()
	cd, _ := args.Get(0).(*coredocumentpb.CoreDocument)
	return cd, args.Error(1)
}

func TestGrpcHandler_Create_coredocument_fail(t *testing.T) {
	return
	// TODO: Don't think it makes sense to test this. This logic will be tested in DeriveInvoiceResponse
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(mockModel)
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("core document failed"))
	srv.On("DeriveFromCreatePayload", mock.Anything).Return(model, nil)
	srv.On("Create", mock.Anything).Return(nil)
	payload := &clientinvoicepb.InvoiceCreatePayload{Data: &clientinvoicepb.InvoiceData{GrossAmount: 300}}
	_, err := h.Create(context.Background(), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "core document failed")
}

func TestGrpcHandler_Create(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(mockModel)
	payload := &clientinvoicepb.InvoiceCreatePayload{Data: &clientinvoicepb.InvoiceData{GrossAmount: 300}}
	response := &clientinvoicepb.InvoiceResponse{}
	srv.On("DeriveFromCreatePayload", mock.Anything).Return(model, nil)
	srv.On("DeriveInvoiceResponse", model).Return(response, nil)
	srv.On("Create", mock.Anything).Return(nil)
	res, err := h.Create(context.Background(), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, res, "must be non nil")
	assert.Equal(t, res, response)
}

func TestGrpcHandler_Get(t *testing.T) {
	identifier := "0x01010101"
	identifierBytes, _ := hexutil.Decode(identifier)
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(mockModel)
	payload := &clientinvoicepb.GetRequest{Identifier: identifier}
	response := &clientinvoicepb.InvoiceResponse{}
	srv.On("GetLastVersion", identifierBytes).Return(model, nil)
	srv.On("DeriveInvoiceResponse", model).Return(response, nil)
	res, err := h.Get(context.Background(), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, res, "must be non nil")
	assert.Equal(t, res, response)
}
func TestGrpcHandler_GetVersion(t *testing.T) {
	h := getHandler()
	srv := h.service.(*mockService)
	model := new(mockModel)
	payload := &clientinvoicepb.GetVersionRequest{Identifier: "0x01", Version: "0x00"}
	response := &clientinvoicepb.InvoiceResponse{}
	srv.On("GetVersion", []byte{0x01}, []byte{0x00}).Return(model, nil)
	srv.On("DeriveInvoiceResponse", model).Return(response, nil)
	res, err := h.GetVersion(context.Background(), payload)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, res, "must be non nil")
	assert.Equal(t, res, response)
}
