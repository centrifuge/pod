package testingdocuments

import (
	"context"
	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/mock"
	"time"
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

func GenerateCoreDocumentModelWithCollaborators(collaborators [][]byte) (*documents.CoreDocumentModel, error) {
	dueDate := time.Now().Add(4 * 24 * time.Hour)
	invData := &invoicepb.InvoiceData{
		InvoiceNumber: "2132131",
		GrossAmount:   123,
		NetAmount:     123,
		Currency:      "EUR",
		DueDate:       &timestamp.Timestamp{Seconds: dueDate.Unix()},
	}
	dataSalts, _ := documents.GenerateNewSalts(invData, "invoice", []byte{1, 0, 0, 0})
	serializedInv, _ := proto.Marshal(invData)
	var dm *documents.CoreDocumentModel
	if collaborators != nil {
		var collabs []string
		for _, c := range collaborators {
			encoded := hexutil.Encode(c)
			collabs = append(collabs, encoded)
		}
		m, err := documents.NewWithCollaborators(collabs)
		if err != nil {
			return nil, err
		}
	dm = m
	} else {
			dm = documents.NewCoreDocModel()
		}
	dm.Document.EmbeddedData = &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   serializedInv,
	}
	dm.Document.EmbeddedDataSalts = documents.ConvertToProtoSalts(dataSalts)
	cdSalts, _ := documents.GenerateNewSalts(dm.Document, "", nil)
	dm.Document.CoredocumentSalts = documents.ConvertToProtoSalts(cdSalts)

	mockModel := MockModel{
		CoreDocumentModel: dm,
	}
	return mockModel.CoreDocumentModel, nil
}

func GenerateCoreDocumentModel() (*documents.CoreDocumentModel, error) {
	dm, err := GenerateCoreDocumentModelWithCollaborators(nil)
	if err != nil {
		return nil, err
	}
	return dm, nil
}
