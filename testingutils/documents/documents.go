// +build unit integration testworld

package testingdocuments

import (
	"context"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-substrate-rpc-client/v2/types"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	documents.Service
	mock.Mock
}

func (m *MockService) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Document, error) {
	args := m.Called(documentID)
	model, _ := args.Get(0).(documents.Document)
	return model, args.Error(1)
}

func (m *MockService) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Document, error) {
	args := m.Called(documentID, version)
	model, _ := args.Get(0).(documents.Document)
	return model, args.Error(1)
}

func (m *MockService) CreateProofs(ctx context.Context, documentID []byte, fields []string) (*documents.DocumentProof, error) {
	args := m.Called(ctx, documentID, fields)
	resp, _ := args.Get(0).(*documents.DocumentProof)
	return resp, args.Error(1)
}

func (m *MockService) CreateProofsForVersion(ctx context.Context, documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	args := m.Called(ctx, documentID, version, fields)
	resp, _ := args.Get(0).(*documents.DocumentProof)
	return resp, args.Error(1)
}

func (m *MockService) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Document, error) {
	args := m.Called(cd)
	return args.Get(0).(documents.Document), args.Error(1)
}

func (m *MockService) RequestDocumentSignature(ctx context.Context, model documents.Document, collaborator identity.DID) ([]*coredocumentpb.Signature, error) {
	args := m.Called()
	return args.Get(0).([]*coredocumentpb.Signature), args.Error(1)
}

func (m *MockService) ReceiveAnchoredDocument(ctx context.Context, model documents.Document, collaborator identity.DID) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockService) Commit(ctx context.Context, doc documents.Document) (gocelery.JobID, error) {
	args := m.Called(ctx, doc)
	jobID, _ := args.Get(0).(gocelery.JobID)
	return jobID, args.Error(1)
}

func (m *MockService) Derive(ctx context.Context, payload documents.UpdatePayload) (documents.Document, error) {
	args := m.Called(ctx, payload)
	model, _ := args.Get(0).(documents.Document)
	return model, args.Error(1)
}

type MockModel struct {
	documents.Document
	mock.Mock
}

func (m *MockModel) Scheme() string {
	args := m.Called()
	return args.String(0)
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

func (m *MockModel) GetAttributes() []documents.Attribute {
	args := m.Called()
	attrs, _ := args.Get(0).([]documents.Attribute)
	return attrs
}

func (m *MockModel) IsDIDCollaborator(did identity.DID) (bool, error) {
	args := m.Called(did)
	ok, _ := args.Get(0).(bool)
	return ok, args.Error(1)
}

func (m *MockModel) GetAccessTokens() ([]*coredocumentpb.AccessToken, error) {
	args := m.Called()
	ac, _ := args.Get(0).([]*coredocumentpb.AccessToken)
	return ac, args.Error(1)
}

func (m *MockModel) AttributeExists(key documents.AttrKey) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockModel) GetAttribute(key documents.AttrKey) (documents.Attribute, error) {
	args := m.Called(key)
	attr, _ := args.Get(0).(documents.Attribute)
	return attr, args.Error(1)
}

func (m *MockModel) AddAttributes(ca documents.CollaboratorsAccess, prepareNewVersion bool, attrs ...documents.Attribute) error {
	args := m.Called(ca, prepareNewVersion, attrs)
	return args.Error(0)
}

func (m *MockModel) GetStatus() documents.Status {
	args := m.Called()
	st, _ := args.Get(0).(documents.Status)
	return st
}

func (m *MockModel) CalculateTransitionRulesFingerprint() ([]byte, error) {
	args := m.Called()
	p, _ := args.Get(0).([]byte)
	return p, nil
}

type MockRegistry struct {
	mock.Mock
}

func (m MockRegistry) OwnerOf(registry common.Address, tokenID []byte) (common.Address, error) {
	args := m.Called(registry, tokenID)
	addr, _ := args.Get(0).(common.Address)
	return addr, args.Error(1)
}

func (m MockRegistry) OwnerOfOnCC(registry common.Address, tokenID []byte) (types.AccountID, error) {
	args := m.Called(registry, tokenID)
	addr, _ := args.Get(0).(types.AccountID)
	return addr, args.Error(1)
}

func (m MockRegistry) OwnerOfWithRetrial(registry common.Address, tokenID []byte) (common.Address, error) {
	args := m.Called(registry, tokenID)
	addr, _ := args.Get(0).(common.Address)
	return addr, args.Error(1)
}
