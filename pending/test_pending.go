// +build integration unit

package pending

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/gocelery/v2"
	"github.com/stretchr/testify/mock"
)

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}

type MockService struct {
	mock.Mock
	Service
}

func (m *MockService) Create(ctx context.Context, payload documents.UpdatePayload) (documents.Document, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}

func (m *MockService) Clone(ctx context.Context, payload documents.ClonePayload) (documents.Document, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}

func (m *MockService) Update(ctx context.Context, payload documents.UpdatePayload) (documents.Document, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}

func (m *MockService) Commit(ctx context.Context, docID []byte) (documents.Document, gocelery.JobID, error) {
	args := m.Called(ctx, docID)
	doc, _ := args.Get(0).(documents.Document)
	jobID, _ := args.Get(1).(gocelery.JobID)
	return doc, jobID, args.Error(2)
}

func (m *MockService) Get(ctx context.Context, docID []byte, st documents.Status) (documents.Document, error) {
	args := m.Called(ctx, docID, st)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}

func (m *MockService) GetVersion(ctx context.Context, docID, versionID []byte) (documents.Document, error) {
	args := m.Called(ctx, docID, versionID)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}

func (m *MockService) AddSignedAttribute(ctx context.Context, docID []byte, label string, value []byte, valType documents.AttributeType) (documents.Document, error) {
	args := m.Called(ctx, docID, label, value)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}

func (m *MockService) RemoveCollaborators(ctx context.Context, docID []byte, dids []identity.DID) (documents.Document, error) {
	args := m.Called(ctx, docID, dids)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}

func (m *MockService) GetRole(ctx context.Context, docID, roleID []byte) (*coredocumentpb.Role, error) {
	args := m.Called(ctx, docID, roleID)
	r, _ := args.Get(0).(*coredocumentpb.Role)
	return r, args.Error(1)
}

func (m *MockService) AddRole(ctx context.Context, docID []byte, roleKey string, collab []identity.DID) (*coredocumentpb.Role, error) {
	args := m.Called(ctx, docID, roleKey, collab)
	r, _ := args.Get(0).(*coredocumentpb.Role)
	return r, args.Error(1)
}

func (m *MockService) UpdateRole(ctx context.Context, docID, roleID []byte, collab []identity.DID) (*coredocumentpb.Role, error) {
	args := m.Called(ctx, docID, roleID, collab)
	r, _ := args.Get(0).(*coredocumentpb.Role)
	return r, args.Error(1)
}

func (m *MockService) AddTransitionRules(ctx context.Context, docID []byte, addRule AddTransitionRules) ([]*coredocumentpb.TransitionRule, error) {
	args := m.Called(ctx, docID, addRule)
	r, _ := args.Get(0).([]*coredocumentpb.TransitionRule)
	return r, args.Error(1)
}

func (m *MockService) GetTransitionRule(ctx context.Context, docID, ruleID []byte) (*coredocumentpb.TransitionRule, error) {
	args := m.Called(ctx, docID, ruleID)
	r, _ := args.Get(0).(*coredocumentpb.TransitionRule)
	return r, args.Error(1)
}

func (m *MockService) DeleteTransitionRule(ctx context.Context, docID, ruleID []byte) error {
	args := m.Called(ctx, docID, ruleID)
	return args.Error(0)
}

func (m *MockService) AddAttributes(
	ctx context.Context,
	docID []byte, attrs []documents.Attribute) (documents.Document, error) {
	args := m.Called(ctx, docID, attrs)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}

func (m *MockService) DeleteAttribute(ctx context.Context, docID []byte, key documents.AttrKey) (documents.Document, error) {
	args := m.Called(ctx, docID, key)
	doc, _ := args.Get(0).(documents.Document)
	return doc, args.Error(1)
}
