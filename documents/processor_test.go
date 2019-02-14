// +build unit

package documents

import (
	"context"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/ed25519"
)

func TestCoreDocumentProcessor_SendNilDocument(t *testing.T) {
	dp := DefaultProcessor(nil, nil, nil, cfg)
	err := dp.Send(nil, nil, [identity.CentIDLength]byte{})
	assert.Error(t, err, "should have thrown an error")
}

type mockModel struct {
	mock.Mock
	Model
}

func (m mockModel) PackCoreDocument() (*CoreDocumentModel, error) {
	args := m.Called()
	cd, _ := args.Get(0).(*CoreDocumentModel)
	return cd, args.Error(1)
}

func (m mockModel) UnpackCoreDocument(model *CoreDocumentModel) error {
	args := m.Called(model)
	return args.Error(0)
}

func (m mockModel) CalculateDataRoot() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func TestDefaultProcessor_PrepareForSignatureRequests(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("error")).Once()
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	err := dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack core document")

	dm := NewCoreDocModel()
	cd := dm.Document
	model = mockModel{}

	// failed to get id
	pub, _ := cfg.GetSigningKeyPair()
	cfg.Set("keys.signing.publicKey", "wrong path")
	cd.DataRoot = utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	dm.setCoreDocumentSalts()
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Once()
	cfg.Set("keys.signing.publicKey", pub)
	ctxh = testingconfig.CreateAccountContext(t, cfg)

	// failed unpack
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Once()
	model.On("UnpackCoreDocument", dm).Return(errors.New("error")).Once()
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	err = dp.PrepareForSignatureRequests(ctxh, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unpack the core document")
	model.AssertExpectations(t)

	// success
	cd.Signatures = nil
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Once()
	model.On("UnpackCoreDocument", dm).Return(nil).Once()
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	err = dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, cd.Signatures)
	assert.Len(t, cd.Signatures, 1)
	sig := cd.Signatures[0]
	self, _ := contextutil.Self(ctxh)
	assert.True(t, ed25519.Verify(self.Keys[identity.KeyPurposeSigning].PublicKey, cd.SigningRoot, sig.Signature))
}

type p2pClient struct {
	mock.Mock
	Client
}

func (p p2pClient) GetSignaturesForDocument(ctx context.Context, model *CoreDocumentModel) error {
	args := p.Called(ctx, model)
	return args.Error(0)
}

func TestDefaultProcessor_RequestSignatures(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("error")).Once()
	err := dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack core document")

	// validations failed
	dm := NewCoreDocModel()
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Times(4)
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate model for signature request")

	// failed signature collection
	dm = NewCoreDocModel()
	cd := dm.Document
	cd.DataRoot = utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	dm.setCoreDocumentSalts()
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Once()
	model.On("UnpackCoreDocument", dm).Return(nil).Once()
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	err = dp.PrepareForSignatureRequests(ctxh, model)
	assert.Nil(t, err)
	model.AssertExpectations(t)
	c := p2pClient{}
	c.On("GetSignaturesForDocument", ctxh, dm).Return(errors.New("error")).Once()
	dp.p2pClient = c
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Times(4)
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect signatures from the collaborators")

	// unpack fail
	c = p2pClient{}
	c.On("GetSignaturesForDocument", ctxh, dm).Return(nil).Once()
	dp.p2pClient = c
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Times(4)
	model.On("UnpackCoreDocument", dm).Return(errors.New("error")).Once()
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unpack core document")

	// success
	c = p2pClient{}
	c.On("GetSignaturesForDocument", ctxh, dm).Return(nil).Once()
	dp.p2pClient = c
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Times(4)
	model.On("UnpackCoreDocument", dm).Return(nil).Once()
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestDefaultProcessor_PrepareForAnchoring(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("error")).Once()
	err := dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack core document")

	// failed validations
	dm := NewCoreDocModel()
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Times(4)
	err = dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate signatures")

	// failed unpack
	dm = NewCoreDocModel()
	cd := dm.Document
	cd.DataRoot = utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	dm.setCoreDocumentSalts()
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	err = dm.CalculateSigningRoot(cd.DataRoot)
	assert.Nil(t, err)
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Times(4)
	model.On("UnpackCoreDocument", dm).Return(errors.New("error")).Once()
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	c, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	s := identity.Sign(c, identity.KeyPurposeSigning, cd.SigningRoot)
	cd.Signatures = []*coredocumentpb.Signature{s}
	assert.Nil(t, err)
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil)
	dp = DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	err = dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unpack core document")

	// success
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Times(4)
	model.On("UnpackCoreDocument", dm).Return(nil).Once()
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	err = dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, cd.DocumentRoot)
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

	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("error")).Once()
	err := dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack core document")

	// validations failed
	model = mockModel{}
	dm := NewCoreDocModel()
	model.On("PackCoreDocument").Return(dm, nil).Times(5)
	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pre anchor validation failed")

	// get ID failed
	dm = NewCoreDocModel()
	cd := dm.Document
	cd.DataRoot = utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	dm.setCoreDocumentSalts()
	assert.Nil(t, dm.CalculateSigningRoot(cd.DataRoot))
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Times(5)
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	c, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	s := identity.Sign(c, identity.KeyPurposeSigning, cd.SigningRoot)
	cd.Signatures = []*coredocumentpb.Signature{s}
	assert.Nil(t, dm.CalculateDocumentRoot())
	assert.Nil(t, err)
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil).Once()
	err = dp.AnchorDocument(context.Background(), model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "self value not found in the context")

	// success
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Times(5)
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil).Once()

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
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil)
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("error")).Once()
	err := dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack core document")

	// failed validations
	model = mockModel{}
	dm := NewCoreDocModel()
	model.On("PackCoreDocument").Return(dm, nil).Times(6)
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "post anchor validations failed")

	// failed send
	dm = NewCoreDocModel()
	cd := dm.Document
	cd.DataRoot = utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	cd.Collaborators = [][]byte{[]byte("some id")}
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	dm.setCoreDocumentSalts()
	assert.Nil(t, dm.CalculateSigningRoot(cd.DataRoot))
	model = mockModel{}
	model.On("PackCoreDocument").Return(dm, nil).Times(6)
	c, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	s := identity.Sign(c, identity.KeyPurposeSigning, cd.SigningRoot)
	cd.Signatures = []*coredocumentpb.Signature{s}
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	assert.Nil(t, dm.CalculateDocumentRoot())
	docRoot, err := anchors.ToDocumentRoot(cd.DocumentRoot)
	assert.Nil(t, err)
	repo := mockRepo{}
	repo.On("GetDocumentRootOf", mock.Anything).Return(docRoot, nil).Once()
	dp.anchorRepository = repo
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid length byte slice provided for centID")
}
