package testingdocuments

import (
	"context"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
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

func GenerateCoreDocumentModelWithCollaborators(collaborators [][]byte) *documents.CoreDocumentModel {
	identifier := utils.RandomSlice(32)
	invData := &invoicepb.InvoiceData{}
	dataSalts, _ := documents.GenerateNewSalts(invData, "invoice", []byte{1, 0, 0, 0})

	serializedInv, _ := proto.Marshal(invData)
	doc := &coredocumentpb.CoreDocument{
		Collaborators:      collaborators,
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        utils.RandomSlice(32),
		EmbeddedData: &any.Any{
			TypeUrl: documenttypes.InvoiceDataTypeUrl,
			Value:   serializedInv,
		},
		EmbeddedDataSalts: documents.ConvertToProtoSalts(dataSalts),
	}
	cdSalts, _ := documents.GenerateNewSalts(doc, "", nil)
	doc.CoredocumentSalts = documents.ConvertToProtoSalts(cdSalts)
	dm := documents.NewCoreDocModel()
	mockModel := MockModel{
		CoreDocumentModel: dm,
	}
	dm.Document = doc
	return mockModel.CoreDocumentModel
}

func GenerateCoreDocumentModel() *documents.CoreDocumentModel {
	return GenerateCoreDocumentModelWithCollaborators(nil)
}
