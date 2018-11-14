// +build integration unit

package testingcoredocument

import (
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/mock"
)

func GenerateCoreDocument() *coredocumentpb.CoreDocument {
	identifier := utils.RandomSlice(32)
	salts := &coredocumentpb.CoreDocumentSalts{}
	doc := &coredocumentpb.CoreDocument{
		DataRoot:           utils.RandomSlice(32),
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        utils.RandomSlice(32),
		CoredocumentSalts:  salts,
		EmbeddedData: &any.Any{
			TypeUrl: documenttypes.InvoiceDataTypeUrl,
		},
		EmbeddedDataSalts: &any.Any{
			TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
		},
	}
	proofs.FillSalts(doc, salts)
	return doc
}

type MockCoreDocumentProcessor struct {
	mock.Mock
}

func (m *MockCoreDocumentProcessor) Send(ctx *header.ContextHeader, coreDocument *coredocumentpb.CoreDocument, recipient identity.CentID) (err error) {
	args := m.Called(coreDocument, ctx, recipient)
	return args.Error(0)
}

func (m *MockCoreDocumentProcessor) Anchor(
	ctx *header.ContextHeader,
	coreDocument *coredocumentpb.CoreDocument,
	saveState func(*coredocumentpb.CoreDocument) error) (err error) {
	args := m.Called(ctx, coreDocument, saveState)
	if saveState != nil {
		err := saveState(coreDocument)
		if err != nil {
			return err
		}
	}
	return args.Error(0)
}

func (m *MockCoreDocumentProcessor) PrepareForSignatureRequests(ctx *header.ContextHeader, model documents.Model) error {
	args := m.Called(model)
	return args.Error(0)
}

func (m *MockCoreDocumentProcessor) RequestSignatures(ctx *header.ContextHeader, model documents.Model) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *MockCoreDocumentProcessor) PrepareForAnchoring(model documents.Model) error {
	args := m.Called(model)
	return args.Error(0)
}

func (m *MockCoreDocumentProcessor) AnchorDocument(ctx *header.ContextHeader, model documents.Model) error {
	args := m.Called(model)
	return args.Error(0)
}

func (m *MockCoreDocumentProcessor) SendDocument(ctx *header.ContextHeader, model documents.Model) error {
	args := m.Called(ctx, model)
	return args.Error(0)
}

func (m *MockCoreDocumentProcessor) GetDataProofHashes(coreDocument *coredocumentpb.CoreDocument) (hashes [][]byte, err error) {
	args := m.Called(coreDocument)
	return args.Get(0).([][]byte), args.Error(1)
}
