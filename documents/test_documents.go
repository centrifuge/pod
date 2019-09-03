// +build integration unit testworld

package documents

import (
	"context"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
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
	Model
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

type MockService struct {
	Service
	mock.Mock
}

func (m *MockService) GetVersion(ctx context.Context, documentID []byte, version []byte) (Model, error) {
	args := m.Called(documentID, version)
	doc, _ := args.Get(0).(Model)
	return doc, args.Error(1)
}

func (m *MockService) Derive(ctx context.Context, payload UpdatePayload) (Model, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(Model)
	return doc, args.Error(1)
}

func (m *MockService) Validate(ctx context.Context, model Model, old Model) error {
	args := m.Called(ctx, model, old)
	return args.Error(0)
}

func (m *MockService) New(scheme string) (Model, error) {
	args := m.Called(scheme)
	doc, _ := args.Get(0).(Model)
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

func (m *MockModel) DeriveFromUpdatePayload(ctx context.Context, payload UpdatePayload) (Model, error) {
	args := m.Called(ctx, payload)
	doc, _ := args.Get(0).(Model)
	return doc, args.Error(1)
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Exists(accountID, id []byte) bool {
	args := m.Called(accountID, id)
	return args.Get(0).(bool)
}

func (m *MockRepository) Get(accountID, id []byte) (Model, error) {
	args := m.Called(accountID, id)
	doc, _ := args.Get(0).(Model)
	return doc, args.Error(0)
}

func (m *MockRepository) Create(accountID, id []byte, model Model) error {
	args := m.Called(accountID, id)
	return args.Error(0)
}

func (m *MockRepository) Update(accountID, id []byte, model Model) error {
	args := m.Called(accountID, id)
	return args.Error(0)
}

func (m *MockRepository) Register(model Model) {
	m.Called(model)
	return
}

func (m *MockRepository) GetLatest(accountID, docID []byte) (Model, error) {
	args := m.Called(accountID, docID)
	doc, _ := args.Get(0).(Model)
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
