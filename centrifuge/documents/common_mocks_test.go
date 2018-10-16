package documents_test

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateProofs(documentID []byte, fields []string) (*documentpb.DocumentProof, error) {
	args := m.Called(documentID, fields)
	return args.Get(0).(*documentpb.DocumentProof), args.Get(1).(error)
}

func (m *MockService) CreateProofsForVersion(documentID, version []byte, fields []string) (*documentpb.DocumentProof, error) {
	args := m.Called(documentID, version, fields)
	return args.Get(0).(*documentpb.DocumentProof), args.Get(1).(error)
}

func (m *MockService) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	args := m.Called(cd)
	return args.Get(0).(documents.Model), args.Get(1).(error)
}

func (m *MockService) RequestDocumentSignature(model documents.Model) (*coredocumentpb.Signature, error) {
	args := m.Called()
	return args.Get(0).(*coredocumentpb.Signature), args.Get(1).(error)
}

func (m *MockService) ReceiveAnchoredDocument(model documents.Model, headers *p2ppb.CentrifugeHeader) error {
	args := m.Called()
	return args.Get(0).(error)
}
