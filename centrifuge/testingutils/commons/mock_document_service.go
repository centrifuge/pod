package testingcommons

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/documents"
	"github.com/stretchr/testify/mock"
)

type MockDocService struct {
	mock.Mock
}

func (m *MockDocService) GetLastVersion(documentID []byte) (documents.Model, error) {
	args := m.Called(documentID)
	return args.Get(0).(documents.Model), args.Error(1)
}

func (m *MockDocService) GetVersion(documentID []byte, version []byte) (documents.Model, error) {
	args := m.Called(documentID, version)
	return args.Get(0).(documents.Model), args.Error(1)
}

func (m *MockDocService) CreateProofs(documentID []byte, fields []string) (*documentpb.DocumentProof, error) {
	args := m.Called(documentID, fields)
	return args.Get(0).(*documentpb.DocumentProof), args.Error(1)
}

func (m *MockDocService) CreateProofsForVersion(documentID, version []byte, fields []string) (*documentpb.DocumentProof, error) {
	args := m.Called(documentID, version, fields)
	return args.Get(0).(*documentpb.DocumentProof), args.Error(1)
}

func (m *MockDocService) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	args := m.Called(cd)
	return args.Get(0).(documents.Model), args.Error(1)
}
