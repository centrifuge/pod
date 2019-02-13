package testingdocuments

import (
	"context"
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	documents.Service
	mock.Mock
}

func (m *MockService) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
	args := m.Called(documentID)
	return args.Get(0).(documents.Model), args.Error(1)
}

func (m *MockService) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Model, error) {
	args := m.Called(documentID, version)
	return args.Get(0).(documents.Model), args.Error(1)
}

func (m *MockService) CreateProofs(ctx context.Context, documentID []byte, fields []string) (*documents.DocumentProof, error) {
	args := m.Called(documentID, fields)
	return args.Get(0).(*documents.DocumentProof), args.Error(1)
}

func (m *MockService) CreateProofsForVersion(ctx context.Context, documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	args := m.Called(documentID, version, fields)
	return args.Get(0).(*documents.DocumentProof), args.Error(1)
}

func (m *MockService) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	args := m.Called(cd)
	return args.Get(0).(documents.Model), args.Error(1)
}

func (m *MockService) RequestDocumentSignature(ctx context.Context, model documents.Model) (*coredocumentpb.Signature, error) {
	args := m.Called()
	return args.Get(0).(*coredocumentpb.Signature), args.Error(1)
}

func (m *MockService) ReceiveAnchoredDocument(ctx context.Context, model documents.Model, senderID []byte) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockService) Exists(ctx context.Context, documentID []byte) bool {
	args := m.Called()
	return args.Get(0).(bool)
}

type MockModel struct {
	documents.Model
	mock.Mock
	CoreDocumentModel *documents.CoreDocumentModel
}

func (m *MockModel) PackCoreDocument() (*documents.CoreDocumentModel, error) {
	args := m.Called()
	dm, _ := args.Get(0).(*documents.CoreDocumentModel)
	return dm, args.Error(1)
}

func (m *MockModel) JSON() ([]byte, error) {
	args := m.Called()
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}

func GenerateCoreDocumentModel() *documents.CoreDocumentModel {
	docModel := documents.NewCoreDocModel()
	identifier := utils.RandomSlice(32)
	dataSalts := &invoicepb.InvoiceDataSalts{}
	invData := &invoicepb.InvoiceData{}
	proofs.FillSalts(invData, dataSalts)

	serializedInv, _ := proto.Marshal(invData)
	serializedInvSalts, _ := proto.Marshal(dataSalts)
	salts := &coredocumentpb.CoreDocumentSalts{}
	doc := &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        utils.RandomSlice(32),
		CoredocumentSalts:  salts,
		DocumentRoot: identifier,
		EmbeddedData: &any.Any{
			TypeUrl: documenttypes.InvoiceDataTypeUrl,
			Value:   serializedInv,
		},
		EmbeddedDataSalts: &any.Any{
			TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
			Value:   serializedInvSalts,
		},
	}
	proofs.FillSalts(doc, salts)
	docModel.Document = doc
	return docModel
}
