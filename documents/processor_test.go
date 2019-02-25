// +build unit

package documents

import (
	"context"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/ed25519"
)

type mockModel struct {
	mock.Mock
	Model
	sigs []*coredocumentpb.Signature
}

func (m *mockModel) DataRoot() ([]byte, error) {
	args := m.Called()
	dr, _ := args.Get(0).([]byte)
	return dr, args.Error(1)
}

func (m *mockModel) SigningRoot() ([]byte, error) {
	args := m.Called()
	sr, _ := args.Get(0).([]byte)
	return sr, args.Error(1)
}

func (m *mockModel) DocumentRoot() ([]byte, error) {
	args := m.Called()
	dr, _ := args.Get(0).([]byte)
	return dr, args.Error(1)
}

func (m *mockModel) AppendSignatures(sigs ...*coredocumentpb.Signature) {
	m.Called(sigs)
	m.sigs = sigs
}

func (m *mockModel) ID() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *mockModel) CurrentVersion() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *mockModel) NextVersion() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *mockModel) Signatures() []coredocumentpb.Signature {
	m.Called()
	var ss []coredocumentpb.Signature
	for _, s := range m.sigs {
		ss = append(ss, *s)
	}
	return ss
}

func (m *mockModel) GetCollaborators(filterIDs ...identity.CentID) ([]identity.CentID, error) {
	args := m.Called(filterIDs)
	cids, _ := args.Get(0).([]identity.CentID)
	return cids, args.Error(1)
}

func (m *mockModel) PackCoreDocument() (coredocumentpb.CoreDocument, error) {
	args := m.Called()
	cd, _ := args.Get(0).(coredocumentpb.CoreDocument)
	return cd, args.Error(1)
}

func TestDefaultProcessor_PrepareForSignatureRequests(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// failed to get self
	err := dp.PrepareForSignatureRequests(context.Background(), nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrSelfNotFound, err))

	// failed data root
	model := new(mockModel)
	model.On("DataRoot").Return(nil, errors.New("failed data root")).Once()
	err = dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed data root")

	// failed signing root
	model = new(mockModel)
	model.On("DataRoot").Return(utils.RandomSlice(32), nil).Once()
	model.On("SigningRoot").Return(nil, errors.New("failed signing root")).Once()
	err = dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed signing root")

	// success
	sr := utils.RandomSlice(32)
	model = new(mockModel)
	model.On("DataRoot").Return(utils.RandomSlice(32), nil).Once()
	model.On("SigningRoot").Return(sr, nil).Once()
	model.On("AppendSignatures", mock.Anything).Return().Once()
	err = dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, model.sigs)
	assert.Len(t, model.sigs, 1)
	sig := model.sigs[0]
	self, _ := contextutil.Self(ctxh)
	assert.True(t, ed25519.Verify(self.Keys[identity.KeyPurposeSigning].PublicKey, sr, sig.Signature))
}

type p2pClient struct {
	mock.Mock
	Client
}

func (p p2pClient) GetSignaturesForDocument(ctx context.Context, model Model) ([]*coredocumentpb.Signature, error) {
	args := p.Called(ctx, model)
	sigs, _ := args.Get(0).([]*coredocumentpb.Signature)
	return sigs, args.Error(1)
}

func (p p2pClient) SendAnchoredDocument(ctx context.Context, receiverID identity.CentID, in *p2ppb.AnchorDocumentRequest) (*p2ppb.AnchorDocumentResponse, error) {
	args := p.Called(ctx, receiverID, in)
	resp, _ := args.Get(0).(*p2ppb.AnchorDocumentResponse)
	return resp, args.Error(1)
}

func TestDefaultProcessor_RequestSignatures(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// failed to get self
	err := dp.RequestSignatures(context.Background(), nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(contextutil.ErrSelfNotFound, err))

	self, err := contextutil.Self(ctxh)
	assert.NoError(t, err)
	sr := utils.RandomSlice(32)
	keys := self.Keys[identity.KeyPurposeSigning]
	sig := crypto.Sign(self.ID[:], keys.PrivateKey, keys.PublicKey, sr)

	// validations failed
	model := new(mockModel)
	model.On("ID").Return([]byte{})
	model.On("CurrentVersion").Return([]byte{})
	model.On("NextVersion").Return([]byte{})
	model.On("SigningRoot").Return(nil, errors.New("error"))
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// failed signature collection
	model = new(mockModel)
	id := utils.RandomSlice(32)
	next := utils.RandomSlice(32)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.sigs = append(model.sigs, sig)
	c := p2pClient{}
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
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("AppendSignatures", []*coredocumentpb.Signature{sig}).Return().Once()
	model.sigs = append(model.sigs, sig)
	c = p2pClient{}
	c.On("GetSignaturesForDocument", ctxh, model).Return([]*coredocumentpb.Signature{sig}, nil).Once()
	dp.p2pClient = c
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestDefaultProcessor_PrepareForAnchoring(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	self, err := contextutil.Self(ctxh)
	assert.NoError(t, err)
	sr := utils.RandomSlice(32)
	keys := self.Keys[identity.KeyPurposeSigning]
	sig := crypto.Sign(self.ID[:], keys.PrivateKey, keys.PublicKey, sr)

	// validation failed
	model := new(mockModel)
	id := utils.RandomSlice(32)
	next := utils.RandomSlice(32)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIDService{}
	srv.On("ValidateSignature", sig, sr).Return(errors.New("validation failed")).Once()
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
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIDService{}
	srv.On("ValidateSignature", sig, sr).Return(nil).Once()
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

func (m mockRepo) CommitAnchor(ctx context.Context, anchorID anchors.AnchorID, documentRoot anchors.DocumentRoot, centID identity.CentID, documentProofs [][32]byte, signature []byte) (confirmations <-chan *anchors.WatchCommit, err error) {
	args := m.Called(anchorID, documentRoot, centID, documentProofs, signature)
	c, _ := args.Get(0).(chan *anchors.WatchCommit)
	return c, args.Error(1)
}

func (m mockRepo) GetDocumentRootOf(anchorID anchors.AnchorID) (anchors.DocumentRoot, error) {
	args := m.Called(anchorID)
	docRoot, _ := args.Get(0).(anchors.DocumentRoot)
	return docRoot, args.Error(1)
}

func TestDefaultProcessor_AnchorDocument(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	self, err := contextutil.Self(ctxh)
	assert.NoError(t, err)
	sr := utils.RandomSlice(32)
	keys := self.Keys[identity.KeyPurposeSigning]
	sig := crypto.Sign(self.ID[:], keys.PrivateKey, keys.PublicKey, sr)

	// validations failed
	id := utils.RandomSlice(32)
	next := utils.RandomSlice(32)
	model := new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("DocumentRoot").Return(nil, errors.New("error"))
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIDService{}
	srv.On("ValidateSignature", sig, sr).Return(nil).Once()
	dp.identityService = srv
	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pre anchor validation failed")

	// success
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("DocumentRoot").Return(utils.RandomSlice(32), nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIDService{}
	srv.On("ValidateSignature", sig, sr).Return(nil).Once()
	dp.identityService = srv
	repo := mockRepo{}
	ch := make(chan *anchors.WatchCommit, 1)
	ch <- new(anchors.WatchCommit)
	repo.On("CommitAnchor", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ch, nil).Once()
	dp.anchorRepository = repo
	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestDefaultProcessor_SendDocument(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	self, err := contextutil.Self(ctxh)
	assert.NoError(t, err)
	sr := utils.RandomSlice(32)
	keys := self.Keys[identity.KeyPurposeSigning]
	sig := crypto.Sign(self.ID[:], keys.PrivateKey, keys.PublicKey, sr)

	// validations failed
	id := utils.RandomSlice(32)
	aid, err := anchors.ToAnchorID(id)
	assert.NoError(t, err)
	next := utils.RandomSlice(32)
	model := new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("DocumentRoot").Return(utils.RandomSlice(32), nil)
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIDService{}
	srv.On("ValidateSignature", sig, sr).Return(nil).Once()
	dp.identityService = srv
	repo := mockRepo{}
	repo.On("GetDocumentRootOf", aid).Return(nil, errors.New("error"))
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
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("DocumentRoot").Return(dr[:], nil)
	model.On("GetCollaborators", mock.Anything).Return(nil, errors.New("error")).Once()
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIDService{}
	srv.On("ValidateSignature", sig, sr).Return(nil).Once()
	dp.identityService = srv
	repo = mockRepo{}
	repo.On("GetDocumentRootOf", aid).Return(dr, nil).Once()
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
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("DocumentRoot").Return(dr[:], nil)
	model.On("GetCollaborators", mock.Anything).Return([]identity.CentID{identity.RandomCentID()}, nil).Once()
	model.On("PackCoreDocument").Return(nil, errors.New("error")).Once()
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIDService{}
	srv.On("ValidateSignature", sig, sr).Return(nil).Once()
	dp.identityService = srv
	repo = mockRepo{}
	repo.On("GetDocumentRootOf", aid).Return(dr, nil).Once()
	dp.anchorRepository = repo
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	assert.Error(t, err)

	// send failed
	cd := coredocumentpb.CoreDocument{}
	cid := identity.RandomCentID()
	model = new(mockModel)
	model.On("ID").Return(id)
	model.On("CurrentVersion").Return(id)
	model.On("NextVersion").Return(next)
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("DocumentRoot").Return(dr[:], nil)
	model.On("GetCollaborators", mock.Anything).Return([]identity.CentID{cid}, nil).Once()
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIDService{}
	srv.On("ValidateSignature", sig, sr).Return(nil).Once()
	dp.identityService = srv
	repo = mockRepo{}
	repo.On("GetDocumentRootOf", aid).Return(dr, nil).Once()
	client := p2pClient{}
	client.On("SendAnchoredDocument", mock.Anything, cid, mock.Anything).Return(nil, errors.New("error")).Once()
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
	model.On("SigningRoot").Return(sr, nil)
	model.On("Signatures").Return()
	model.On("DocumentRoot").Return(dr[:], nil)
	model.On("GetCollaborators", mock.Anything).Return([]identity.CentID{cid}, nil).Once()
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.sigs = append(model.sigs, sig)
	srv = &testingcommons.MockIDService{}
	srv.On("ValidateSignature", sig, sr).Return(nil).Once()
	dp.identityService = srv
	repo = mockRepo{}
	repo.On("GetDocumentRootOf", aid).Return(dr, nil).Once()
	client = p2pClient{}
	client.On("SendAnchoredDocument", mock.Anything, cid, mock.Anything).Return(&p2ppb.AnchorDocumentResponse{Accepted: true}, nil).Once()
	dp.anchorRepository = repo
	dp.p2pClient = client
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	client.AssertExpectations(t)
	assert.NoError(t, err)
}
