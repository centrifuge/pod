// +build unit integration testworld

package testingdocuments

import (
	"context"
	"math/big"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/ethereum/go-ethereum/common"
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

func (m *MockService) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Model, error) {
	args := m.Called(cd)
	return args.Get(0).(documents.Model), args.Error(1)
}

func (m *MockService) RequestDocumentSignature(ctx context.Context, model documents.Model, collaborator identity.DID) (*coredocumentpb.Signature, error) {
	args := m.Called()
	return args.Get(0).(*coredocumentpb.Signature), args.Error(1)
}

func (m *MockService) ReceiveAnchoredDocument(ctx context.Context, model documents.Model, collaborator identity.DID) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockService) Exists(ctx context.Context, documentID []byte) bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *MockService) CreateModel(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return model, jobID, args.Error(2)
}

func (m *MockService) UpdateModel(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return model, jobID, args.Error(2)
}

type MockModel struct {
	documents.Model
	mock.Mock
}

func (m *MockModel) GetData() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockModel) PreviousVersion() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *MockModel) CurrentVersion() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *MockModel) CurrentVersionPreimage() []byte {
	args := m.Called()
	id, _ := args.Get(0).([]byte)
	return id
}

func (m *MockModel) PackCoreDocument() (coredocumentpb.CoreDocument, error) {
	args := m.Called()
	dm, _ := args.Get(0).(coredocumentpb.CoreDocument)
	return dm, args.Error(1)
}

func (m *MockModel) UnpackCoreDocument(cd coredocumentpb.CoreDocument) error {
	args := m.Called(cd)
	return args.Error(0)
}

func (m *MockModel) JSON() ([]byte, error) {
	args := m.Called()
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}

func (m *MockModel) ID() []byte {
	args := m.Called()
	id, _ := args.Get(0).([]byte)
	return id
}

func (m *MockModel) NFTs() []*coredocumentpb.NFT {
	args := m.Called()
	dr, _ := args.Get(0).([]*coredocumentpb.NFT)
	return dr
}

func (m *MockModel) Author() (identity.DID, error) {
	args := m.Called()
	id, _ := args.Get(0).(identity.DID)
	return id, args.Error(1)
}

func (m *MockModel) Timestamp() (time.Time, error) {
	args := m.Called()
	dr, _ := args.Get(0).(time.Time)
	return dr, args.Error(1)
}

func (m *MockModel) GetCollaborators(filterIDs ...identity.DID) (documents.CollaboratorsAccess, error) {
	args := m.Called(filterIDs)
	cas, _ := args.Get(0).(documents.CollaboratorsAccess)
	return cas, args.Error(1)
}

type MockRegistry struct {
	mock.Mock
}

func (m MockRegistry) CurrentIndexOfToken(registry common.Address, tokenID []byte) (*big.Int, error) {
	args := m.Called(registry, tokenID)
	addr, _ := args.Get(0).(*big.Int)
	return addr, args.Error(1)
}

func (m MockRegistry) OwnerOf(registry common.Address, tokenID []byte) (common.Address, error) {
	args := m.Called(registry, tokenID)
	addr, _ := args.Get(0).(common.Address)
	return addr, args.Error(1)
}
