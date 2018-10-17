// +build unit integration

package documents

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/common"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateProofs(documentID []byte, fields []string) (common.DocumentProof, error) {
	args := m.Called(documentID, fields)
	return args.Get(0).(common.DocumentProof), args.Error(1)
}

func (m *MockService) CreateProofsForVersion(documentID, version []byte, fields []string) (common.DocumentProof, error) {
	args := m.Called(documentID, version, fields)
	return args.Get(0).(common.DocumentProof), args.Error(1)
}

func (m *MockService) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (Model, error) {
	args := m.Called(cd)
	return args.Get(0).(Model), args.Error(1)
}

func (m *MockService) RequestDocumentSignature(model Model) (*coredocumentpb.Signature, error) {
	args := m.Called()
	return args.Get(0).(*coredocumentpb.Signature), args.Error(1)
}

func (m *MockService) ReceiveAnchoredDocument(model Model, headers *p2ppb.CentrifugeHeader) error {
	args := m.Called()
	return args.Error(0)
}
