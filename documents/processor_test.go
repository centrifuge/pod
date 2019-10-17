// +build unit

package documents

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockModel struct {
	mock.Mock
	Model
	sigs []*coredocumentpb.Signature
}

func (m *mockModel) Scheme() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockModel) SetStatus(st Status) error {
	args := m.Called(st)
	return args.Error(0)
}

func (m *mockModel) NFTs() []*coredocumentpb.NFT {
	args := m.Called()
	dr, _ := args.Get(0).([]*coredocumentpb.NFT)
	return dr
}

func (m *mockModel) CalculateDataRoot() ([]byte, error) {
	args := m.Called()
	dr, _ := args.Get(0).([]byte)
	return dr, args.Error(1)
}

func (m *mockModel) CalculateSigningRoot() ([]byte, error) {
	args := m.Called()
	sr, _ := args.Get(0).([]byte)
	return sr, args.Error(1)
}

func (m *mockModel) CalculateDocumentRoot() ([]byte, error) {
	args := m.Called()
	dr, _ := args.Get(0).([]byte)
	return dr, args.Error(1)
}

func (m *mockModel) CalculateSignaturesRoot() ([]byte, error) {
	args := m.Called()
	dr, _ := args.Get(0).([]byte)
	return dr, args.Error(1)
}

func (m *mockModel) PreviousDocumentRoot() []byte {
	args := m.Called()
	dr, _ := args.Get(0).([]byte)
	return dr
}

func (m *mockModel) AppendSignatures(sigs ...*coredocumentpb.Signature) {
	m.Called(sigs)
	m.sigs = sigs
}

func (m *mockModel) ID() []byte {
	args := m.Called()
	id, _ := args.Get(0).([]byte)
	return id
}

func (m *mockModel) CurrentVersion() []byte {
	args := m.Called()
	id, _ := args.Get(0).([]byte)
	return id
}

func (m *mockModel) CurrentVersionPreimage() []byte {
	args := m.Called()
	id, _ := args.Get(0).([]byte)
	return id
}

func (m *mockModel) NextVersion() []byte {
	args := m.Called()
	id, _ := args.Get(0).([]byte)
	return id
}

func (m *mockModel) PreviousVersion() []byte {
	args := m.Called()
	id, _ := args.Get(0).([]byte)
	return id
}

func (m *mockModel) Signatures() []coredocumentpb.Signature {
	m.Called()
	var ss []coredocumentpb.Signature
	for _, s := range m.sigs {
		ss = append(ss, *s)
	}
	return ss
}

func (m *mockModel) AddUpdateLog(account identity.DID) error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockModel) Author() (identity.DID, error) {
	args := m.Called()
	id, _ := args.Get(0).(identity.DID)
	return id, args.Error(1)
}

func (m *mockModel) Timestamp() (time.Time, error) {
	args := m.Called()
	dr, _ := args.Get(0).(time.Time)
	return dr, args.Error(1)
}

func (m *mockModel) GetCollaborators(filterIDs ...identity.DID) (CollaboratorsAccess, error) {
	args := m.Called(filterIDs)
	cas, _ := args.Get(0).(CollaboratorsAccess)
	return cas, args.Error(1)
}

func (m *mockModel) GetSignerCollaborators(filterIDs ...identity.DID) ([]identity.DID, error) {
	args := m.Called(filterIDs)
	cids, _ := args.Get(0).([]identity.DID)
	return cids, args.Error(1)
}

func (m *mockModel) PackCoreDocument() (coredocumentpb.CoreDocument, error) {
	args := m.Called()
	cd, _ := args.Get(0).(coredocumentpb.CoreDocument)
	return cd, args.Error(1)
}

func (m *mockModel) CollaboratorCanUpdate(new Model, collaborator identity.DID) error {
	args := m.Called(new, collaborator)
	return args.Error(0)
}

func (m *mockModel) SetUsedAnchorRepoAddress(addr common.Address) {
	m.Called(addr)
}

func (m *mockModel) AnchorRepoAddress() common.Address {
	args := m.Called()
	return args.Get(0).(common.Address)
}

func (m *mockModel) GetAttributes() []Attribute {
	args := m.Called()
	attrs, _ := args.Get(0).([]Attribute)
	return attrs
}

func TestDefaultProcessor_PrepareForSignatureRequests(t *testing.T) {
	srv := &testingcommons.MockIdentityService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)

	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// failed to get self
	err := dp.PrepareForSignatureRequests(context.Background(), nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrSelfNotFound, err))

	// failed data root
	model := new(mockModel)
	model.On("CalculateDataRoot").Return(nil, errors.New("failed data root")).Once()
	err = dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed data root")

	// failed signing root
	model = new(mockModel)
	model.On("CalculateDataRoot").Return(utils.RandomSlice(32), nil).Once()
	model.On("AddUpdateLog").Return(nil).Once()
	model.On("SetUsedAnchorRepoAddress", cfg.GetContractAddress(config.AnchorRepo)).Return().Once()
	model.On("CalculateSigningRoot").Return(nil, errors.New("failed signing root")).Once()
	err = dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed signing root")

	// success
	sr := utils.RandomSlice(32)
	model = new(mockModel)
	model.On("CalculateDataRoot").Return(utils.RandomSlice(32), nil).Once()
	model.On("CalculateSigningRoot").Return(sr, nil).Once()
	model.On("AddUpdateLog").Return(nil).Once()
	model.On("SetUsedAnchorRepoAddress", cfg.GetContractAddress(config.AnchorRepo)).Return().Once()
	model.On("AppendSignatures", mock.Anything).Return().Once()
	err = dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, model.sigs)
	assert.Len(t, model.sigs, 1)
	sig := model.sigs[0]
	self, err := contextutil.Account(ctxh)
	assert.NoError(t, err)
	keys, err := self.GetKeys()
	assert.NoError(t, err)
	assert.True(t, crypto.VerifyMessage(keys[identity.KeyPurposeSigning.Name].PublicKey, ConsensusSignaturePayload(sr, byte(0)), sig.Signature, crypto.CurveSecp256K1))
}

type p2pClient struct {
	mock.Mock
	Client
}

func (p *p2pClient) GetSignaturesForDocument(ctx context.Context, model Model) ([]*coredocumentpb.Signature, []error, error) {
	args := p.Called(ctx, model)
	sigs, _ := args.Get(0).([]*coredocumentpb.Signature)
	return sigs, nil, args.Error(1)
}

func (p *p2pClient) SendAnchoredDocument(ctx context.Context, receiverID identity.DID, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error) {
	args := p.Called(ctx, receiverID, in)
	resp, _ := args.Get(0).(*p2ppb.AnchorDocumentResponse)
	return resp, args.Error(1)
}

func TestDefaultProcessor_RequestSignatures(t *testing.T) {
	srv := &testingcommons.MockIdentityService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	self, err := contextutil.Account(ctxh)
	assert.NoError(t, err)
	did := self.GetIdentityID()
	sr := utils.RandomSlice(32)
	sig, err := self.SignMsg(sr)
	assert.NoError(t, err)

	did1, err := identity.NewDIDFromBytes(did)
	assert.NoError(t, err)

	// data validations failed
	model := new(mockModel)
	model.On("ID").Return([]byte{})
	model.On("CurrentVersion").Return([]byte{})
	model.On("NextVersion").Return([]byte{})
	model.On("CalculateSigningRoot").Return(nil, errors.New("error"))
	model.On("Timestamp").Return(time.Now().UTC(), nil)
	model.On("GetAttributes").Return(nil)
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// key validation failed
	model = new(mockModel)
	id := utils.RandomSlice(32)
	next := utils.RandomSlice(32)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("Author").Return(did1, nil)
	model.On("Timestamp").Return(time.Now(), nil)
	model.On("GetAttributes").Return(nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	model.sigs = append(model.sigs, sig)
	c := new(p2pClient)
	srv.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("cannot validate key")).Once()
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot validate key")

	// failed signature collection
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("Author").Return(did1, nil)
	model.On("Timestamp").Return(time.Now(), nil)
	model.On("GetAttributes").Return(nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	model.sigs = append(model.sigs, sig)
	c = new(p2pClient)
	srv.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	c.On("GetSignaturesForDocument", ctxh, model).Return(nil, errors.New("failed to get signatures")).Once()
	dp.p2pClient = c
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get signatures")

	// success
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("AppendSignatures", []*coredocumentpb.Signature{sig}).Return().Once()
	model.On("Author").Return(did1, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	model.On("Timestamp").Return(time.Now(), nil)
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	c = new(p2pClient)
	c.On("GetSignaturesForDocument", ctxh, model).Return([]*coredocumentpb.Signature{sig}, nil).Once()
	dp.p2pClient = c
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestDefaultProcessor_PrepareForAnchoring(t *testing.T) {
	srv := &testingcommons.MockIdentityService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	self, err := contextutil.Account(ctxh)
	assert.NoError(t, err)
	did := self.GetIdentityID()
	sr := utils.RandomSlice(32)
	sig, err := self.SignMsg(sr)
	assert.NoError(t, err)
	did1, err := identity.NewDIDFromBytes(did)
	assert.NoError(t, err)

	// validation failed
	model := new(mockModel)
	id := utils.RandomSlice(32)
	next := utils.RandomSlice(32)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("Author").Return(did1, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	tm := time.Now()
	model.On("Timestamp").Return(tm, nil)
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIdentityService{}
	cid, _ := identity.NewDIDFromBytes(did)
	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, sr, tm).Return(errors.New("validation failed")).Once()
	dp.identityService = srv
	err = dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err)

	// success
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("Author").Return(did1, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	model.On("Timestamp").Return(tm, nil)
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIdentityService{}
	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, sr, tm).Return(nil).Once()
	dp.identityService = srv
	err = dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.NoError(t, err)
}

type mockRepo struct {
	mock.Mock
	anchors.AnchorRepository
}

func (m mockRepo) CommitAnchor(ctx context.Context, anchorID anchors.AnchorID, documentRoot anchors.DocumentRoot, documentProof [32]byte) (done chan error, err error) {
	args := m.Called(anchorID, documentRoot, documentProof)
	c, _ := args.Get(0).(chan error)
	return c, args.Error(1)
}

func (m mockRepo) GetAnchorData(anchorID anchors.AnchorID) (docRoot anchors.DocumentRoot, anchoredTime time.Time, err error) {
	args := m.Called(anchorID)
	docRoot, _ = args.Get(0).(anchors.DocumentRoot)
	anchoredTime, _ = args.Get(1).(time.Time)
	return docRoot, anchoredTime, args.Error(2)
}

func TestDefaultProcessor_AnchorDocument(t *testing.T) {
	srv := &testingcommons.MockIdentityService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	self, err := contextutil.Account(ctxh)
	assert.NoError(t, err)
	did := self.GetIdentityID()
	assert.NoError(t, err)
	sr := utils.RandomSlice(32)
	sig, err := self.SignMsg(sr)
	assert.NoError(t, err)
	did1, err := identity.NewDIDFromBytes(did)
	assert.NoError(t, err)

	// validations failed
	id := utils.RandomSlice(32)
	next := utils.RandomSlice(32)
	model := new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("CalculateDocumentRoot").Return(nil, errors.New("error"))
	model.On("Author").Return(did1, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	tm := time.Now()
	model.On("Timestamp").Return(tm, nil)
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIdentityService{}
	cid, err := identity.NewDIDFromBytes(did)
	assert.NoError(t, err)
	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, sr, tm).Return(nil).Once()
	dp.identityService = srv
	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pre anchor validation failed")

	id = utils.RandomSlice(32)
	next = utils.RandomSlice(32)
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("CurrentVersionPreimage").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil)
	model.On("Author").Return(did1, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	model.On("Timestamp").Return(tm, nil)
	model.On("CalculateSignaturesRoot").Return(nil, errors.New("error"))
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIdentityService{}
	srv.On("ValidateSignature", did1, sig.PublicKey, sig.Signature, sr, tm).Return(nil).Once()
	dp.identityService = srv
	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get signature root")

	// success
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("CurrentVersionPreimage").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("CalculateSignaturesRoot").Return(utils.RandomSlice(32), nil)
	model.On("Signatures").Return()
	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil)
	model.On("Author").Return(did1, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	model.On("Timestamp").Return(tm, nil)
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIdentityService{}
	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, sr, tm).Return(nil).Once()
	dp.identityService = srv
	repo := mockRepo{}
	ch := make(chan error, 1)
	ch <- nil
	repo.On("CommitAnchor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ch, nil).Once()
	dp.anchorRepository = repo
	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestDefaultProcessor_SendDocument(t *testing.T) {
	srv := &testingcommons.MockIdentityService{}
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil).Once()
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	self, err := contextutil.Account(ctxh)
	assert.NoError(t, err)
	didb := self.GetIdentityID()
	assert.NoError(t, err)
	did1, err := identity.NewDIDFromBytes(didb)
	assert.NoError(t, err)
	sr := utils.RandomSlice(32)
	sig, err := self.SignMsg(sr)
	assert.NoError(t, err)
	zeros := [32]byte{}
	zeroRoot, err := anchors.ToDocumentRoot(zeros[:])

	// validations failed
	id := utils.RandomSlice(32)
	aid, err := anchors.ToAnchorID(id)
	assert.NoError(t, err)
	next := utils.RandomSlice(32)
	nextAid, err := anchors.ToAnchorID(next)
	assert.NoError(t, err)
	model := new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil)
	model.On("Author").Return(did1, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	tm := time.Now()
	model.On("Timestamp").Return(tm, nil)
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIdentityService{}
	cid, err := identity.NewDIDFromBytes(didb)
	assert.NoError(t, err)
	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, sr, tm).Return(nil).Once()
	dp.identityService = srv
	repo := mockRepo{}

	repo.On("GetAnchorData", nextAid).Return(zeroRoot, time.Now(), nil)
	repo.On("GetAnchorData", aid).Return(nil, time.Now(), errors.New("error"))
	dp.anchorRepository = repo
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "post anchor validations failed")

	// get collaborators failed
	dr, err := anchors.ToDocumentRoot(utils.RandomSlice(32))
	assert.NoError(t, err)

	assert.NoError(t, err)
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("CalculateDocumentRoot").Return(dr[:], nil)
	model.On("GetSignerCollaborators", mock.Anything).Return(nil, errors.New("error")).Once()
	model.On("Author").Return(did1, nil)
	model.On("Timestamp").Return(tm, nil)
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIdentityService{}
	dp.identityService = srv
	repo = mockRepo{}
	repo.On("GetAnchorData", nextAid).Return(zeroRoot, time.Now(), nil)
	repo.On("GetAnchorData", aid).Return(dr, time.Now(), nil)
	dp.anchorRepository = repo
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	assert.Error(t, err)

	// pack core document failed
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("CalculateDocumentRoot").Return(dr[:], nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{testingidentity.GenerateRandomDID()}, nil)
	model.On("PackCoreDocument").Return(nil, errors.New("error")).Once()
	model.On("Author").Return(did1, nil)
	model.On("Timestamp").Return(tm, nil)
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIdentityService{}
	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, sr, tm).Return(nil).Once()
	dp.identityService = srv
	repo = mockRepo{}
	repo.On("GetAnchorData", nextAid).Return(zeroRoot, time.Now(), errors.New("missing"))
	repo.On("GetAnchorData", aid).Return(dr, time.Now(), nil)
	dp.anchorRepository = repo
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	assert.Error(t, err)

	// send failed
	cd := coredocumentpb.CoreDocument{}
	did := testingidentity.GenerateRandomDID()
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("CalculateDocumentRoot").Return(dr[:], nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did}, nil)
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.On("Author").Return(did1, nil)
	model.On("Timestamp").Return(tm, nil)
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIdentityService{}
	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, sr, tm).Return(nil).Once()
	dp.identityService = srv
	repo = mockRepo{}
	repo.On("GetAnchorData", aid).Return(dr, time.Now(), nil)
	repo.On("GetAnchorData", nextAid).Return([32]byte{}, time.Now(), errors.New("missing"))
	client := new(p2pClient)
	client.On("SendAnchoredDocument", mock.Anything, did, mock.Anything).Return(nil, errors.New("error")).Once()
	dp.anchorRepository = repo
	dp.p2pClient = client
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	client.AssertExpectations(t)
	assert.Error(t, err)

	// successful
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("CalculateDocumentRoot").Return(dr[:], nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did}, nil)
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.On("Author").Return(did1, nil)
	model.On("Timestamp").Return(tm, nil)
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIdentityService{}
	srv.On("ValidateSignature", cid, sig.PublicKey, sig.Signature, sr, tm).Return(nil).Once()
	dp.identityService = srv
	repo = mockRepo{}
	repo.On("GetAnchorData", aid).Return(dr, time.Now(), nil)
	repo.On("GetAnchorData", nextAid).Return([32]byte{}, time.Now(), errors.New("missing"))
	client = new(p2pClient)
	client.On("SendAnchoredDocument", mock.Anything, did, mock.Anything).Return(&p2ppb.AnchorDocumentResponse{Accepted: true}, nil).Once()
	dp.anchorRepository = repo
	dp.p2pClient = client
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	client.AssertExpectations(t)
	assert.NoError(t, err)
}
