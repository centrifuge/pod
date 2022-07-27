//go:build integration || unit || testworld

package documents

import (
	"context"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/stretchr/testify/mock"
)

// GetTestCoreDocWithReset must only be used by tests for manipulations. It gets the embedded coredoc protobuf.
// All calls to this function will cause a regeneration of salts next time for precise-proof trees.
func (cd *CoreDocument) GetTestCoreDocWithReset() *coredocumentpb.CoreDocument {
	cd.Modified = true
	return &cd.Document
}

type MockModel struct {
	Document
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

func (m *MockModel) RemoveCollaborators(dids []identity.DID) error {
	args := m.Called(dids)
	return args.Error(0)
}

func (m *MockModel) AddRole(roleKey string, dids []identity.DID) (*coredocumentpb.Role, error) {
	args := m.Called(roleKey, dids)
	r, _ := args.Get(0).(*coredocumentpb.Role)
	return r, args.Error(1)
}

func (m *MockModel) GetRole(roleID []byte) (*coredocumentpb.Role, error) {
	args := m.Called(roleID)
	r, _ := args.Get(0).(*coredocumentpb.Role)
	return r, args.Error(1)
}

func (m *MockModel) UpdateRole(roleID []byte, dids []identity.DID) (*coredocumentpb.Role, error) {
	args := m.Called(roleID, dids)
	r, _ := args.Get(0).(*coredocumentpb.Role)
	return r, args.Error(1)
}

func (m *MockModel) AddTransitionRuleForAttribute(roleID []byte, key AttrKey) (*coredocumentpb.TransitionRule, error) {
	args := m.Called(roleID, key)
	r, _ := args.Get(0).(*coredocumentpb.TransitionRule)
	return r, args.Error(1)
}

func (m *MockModel) AddComputeFieldsRule(wasm []byte, fields []string, targetField string) (*coredocumentpb.TransitionRule, error) {
	args := m.Called(wasm, fields, targetField)
	r, _ := args.Get(0).(*coredocumentpb.TransitionRule)
	return r, args.Error(1)
}

func (m *MockModel) GetTransitionRule(ruleID []byte) (*coredocumentpb.TransitionRule, error) {
	args := m.Called(ruleID)
	r, _ := args.Get(0).(*coredocumentpb.TransitionRule)
	return r, args.Error(1)
}

func (m *MockModel) DeleteTransitionRule(ruleID []byte) error {
	args := m.Called(ruleID)
	return args.Error(0)
}

type MockService struct {
	Service
	mock.Mock
}

func (m *MockService) GetVersion(ctx context.Context, documentID []byte, version []byte) (Document, error) {
	args := m.Called(documentID, version)
	doc, _ := args.Get(0).(Document)
	return doc, args.Error(1)
}

func (m *MockService) GetCurrentVersion(ctx context.Context, docID []byte) (Document, error) {
	args := m.Called(ctx, docID)
	doc, _ := args.Get(0).(Document)
	return doc, args.Error(1)
}

func (m *MockService) Derive(ctx context.Context, payload UpdatePayload) (Document, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(Document)
	return doc, args.Error(1)
}

func (m *MockService) Validate(ctx context.Context, model Document, old Document) error {
	args := m.Called(ctx, model, old)
	return args.Error(0)
}

func (m *MockService) New(scheme string) (Document, error) {
	args := m.Called(scheme)
	doc, _ := args.Get(0).(Document)
	return doc, args.Error(1)
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

func (m *MockModel) GetCollaborators(filterIDs ...identity.DID) (CollaboratorsAccess, error) {
	args := m.Called(filterIDs)
	cas, _ := args.Get(0).(CollaboratorsAccess)
	return cas, args.Error(1)
}

func (m *MockModel) GetAttributes() []Attribute {
	args := m.Called()
	attrs, _ := args.Get(0).([]Attribute)
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

func (m *MockModel) AttributeExists(key AttrKey) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockModel) GetAttribute(key AttrKey) (Attribute, error) {
	args := m.Called(key)
	attr, _ := args.Get(0).(Attribute)
	return attr, args.Error(1)
}

func (m *MockModel) AddAttributes(ca CollaboratorsAccess, prepareNewVersion bool, attrs ...Attribute) error {
	args := m.Called(ca, prepareNewVersion, attrs)
	return args.Error(0)
}

func (m *MockModel) DeleteAttribute(key AttrKey, prepareNewVersion bool) error {
	args := m.Called(key, prepareNewVersion)
	return args.Error(0)
}

func (m *MockModel) GetStatus() Status {
	args := m.Called()
	return args.Get(0).(Status)
}

func (m *MockModel) SetStatus(st Status) error {
	args := m.Called(st)
	return args.Error(0)
}

func (m *MockModel) Patch(payload UpdatePayload) error {
	args := m.Called(payload)
	return args.Error(0)
}

func (m *MockModel) DeriveFromCreatePayload(ctx context.Context, payload CreatePayload) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *MockModel) DeriveFromClonePayload(ctx context.Context, d Document) error {
	args := m.Called(ctx, d)
	return args.Error(0)
}

func (m *MockModel) DeriveFromUpdatePayload(ctx context.Context, payload UpdatePayload) (Document, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(Document)
	return doc, args.Error(1)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Exists(accountID, id []byte) bool {
	args := m.Called(accountID, id)
	return args.Get(0).(bool)
}

func (m *MockRepository) Get(accountID, id []byte) (Document, error) {
	args := m.Called(accountID, id)
	doc, _ := args.Get(0).(Document)
	return doc, args.Error(0)
}

func (m *MockRepository) Create(accountID, id []byte, model Document) error {
	args := m.Called(accountID, id)
	return args.Error(0)
}

func (m *MockRepository) Update(accountID, id []byte, model Document) error {
	args := m.Called(accountID, id)
	return args.Error(0)
}

func (m *MockRepository) Register(model Document) {
	m.Called(model)
}

func (m *MockRepository) GetLatest(accountID, docID []byte) (Document, error) {
	args := m.Called(accountID, docID)
	doc, _ := args.Get(0).(Document)
	return doc, args.Error(1)
}

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[storage.BootstrappedDB]; !ok {
		return errors.New("initializing LevelDB repository failed")
	}
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}

func (b PostBootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	return b.Bootstrap(ctx)
}

func (PostBootstrapper) TestTearDown() error {
	return nil
}

type MockRequestProcessor struct {
	mock.Mock
}

func (m *MockRequestProcessor) RequestDocumentWithAccessToken(ctx context.Context, granterDID identity.DID, tokenIdentifier,
	documentIdentifier, delegatingDocumentIdentifier []byte) (*p2ppb.GetDocumentResponse, error) {
	args := m.Called(granterDID, tokenIdentifier, documentIdentifier, delegatingDocumentIdentifier)
	resp, _ := args.Get(0).(*p2ppb.GetDocumentResponse)
	return resp, args.Error(1)
}
