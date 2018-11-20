// +build unit

package coredocument

import (
	"context"
	"fmt"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
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
	documents.Model
}

func (m mockModel) PackCoreDocument() (*coredocumentpb.CoreDocument, error) {
	args := m.Called()
	cd, _ := args.Get(0).(*coredocumentpb.CoreDocument)
	return cd, args.Error(1)
}

func (m mockModel) UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error {
	args := m.Called(cd)
	return args.Error(0)
}

func TestDefaultProcessor_PrepareForSignatureRequests(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	ctxh, err := header.NewContextHeader(context.Background(), cfg)
	assert.Nil(t, err)
	err = dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack core document")

	cd := new(coredocumentpb.CoreDocument)
	model = mockModel{}

	// failed to get id
	pub, _ := cfg.GetSigningKeyPair()
	cfg.Set("keys.signing.publicKey", "wrong path")
	cd = New()
	cd.DataRoot = utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	assert.Nil(t, FillSalts(cd))
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	ctxh, err = header.NewContextHeader(context.Background(), cfg)
	assert.NotNil(t, err)
	cfg.Set("keys.signing.publicKey", pub)
	ctxh, err = header.NewContextHeader(context.Background(), cfg)
	assert.Nil(t, err)

	// failed unpack
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.On("UnpackCoreDocument", cd).Return(fmt.Errorf("error")).Once()
	err = dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unpack the core document")

	// success
	cd.Signatures = nil
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.On("UnpackCoreDocument", cd).Return(nil).Once()
	err = dp.PrepareForSignatureRequests(ctxh, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, cd.Signatures)
	assert.Len(t, cd.Signatures, 1)
	sig := cd.Signatures[0]
	id := ctxh.Self()
	assert.True(t, ed25519.Verify(id.Keys[identity.KeyPurposeSigning].PublicKey, cd.SigningRoot, sig.Signature))
}

type p2pClient struct {
	mock.Mock
	client
}

func (p p2pClient) GetSignaturesForDocument(ctx *header.ContextHeader, identityService identity.Service, doc *coredocumentpb.CoreDocument) error {
	args := p.Called(ctx, doc)
	return args.Error(0)
}

func TestDefaultProcessor_RequestSignatures(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	dp := DefaultProcessor(srv, nil, nil, cfg).(defaultProcessor)
	ctx := context.Background()
	ctxh, err := header.NewContextHeader(ctx, cfg)
	assert.Nil(t, err)
	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack core document")

	// validations failed
	cd := new(coredocumentpb.CoreDocument)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate model for signature request")

	// failed signature collection
	cd = New()
	cd.DataRoot = utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	assert.Nil(t, FillSalts(cd))
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.On("UnpackCoreDocument", cd).Return(nil).Once()
	err = dp.PrepareForSignatureRequests(ctxh, model)
	assert.Nil(t, err)
	model.AssertExpectations(t)
	c := p2pClient{}
	c.On("GetSignaturesForDocument", ctxh, cd).Return(fmt.Errorf("error")).Once()
	dp.p2pClient = c
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect signatures from the collaborators")

	// unpack fail
	c = p2pClient{}
	c.On("GetSignaturesForDocument", ctxh, cd).Return(nil).Once()
	dp.p2pClient = c
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	model.On("UnpackCoreDocument", cd).Return(fmt.Errorf("error")).Once()
	err = dp.RequestSignatures(ctxh, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unpack core document")

	// success
	c = p2pClient{}
	c.On("GetSignaturesForDocument", ctxh, cd).Return(nil).Once()
	dp.p2pClient = c
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	model.On("UnpackCoreDocument", cd).Return(nil).Once()
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
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err := dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack core document")

	// failed validations
	cd := new(coredocumentpb.CoreDocument)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	err = dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate signatures")

	// failed unpack
	cd = New()
	cd.DataRoot = utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	assert.Nil(t, FillSalts(cd))
	err = CalculateSigningRoot(cd)
	assert.Nil(t, err)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	model.On("UnpackCoreDocument", cd).Return(fmt.Errorf("error")).Once()
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
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	model.On("UnpackCoreDocument", cd).Return(nil).Once()
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

func (m mockRepo) CommitAnchor(anchorID anchors.AnchorID, documentRoot anchors.DocumentRoot, centrifugeID identity.CentID, documentProofs [][32]byte, signature []byte) (<-chan *anchors.WatchCommit, error) {
	args := m.Called(anchorID, documentRoot, centrifugeID, documentProofs, signature)
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
	ctx := context.Background()
	ctxh, err := header.NewContextHeader(ctx, cfg)
	assert.Nil(t, err)

	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack core document")

	// validations failed
	model = mockModel{}
	cd := new(coredocumentpb.CoreDocument)
	model.On("PackCoreDocument").Return(cd, nil).Times(5)
	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pre anchor validation failed")

	// get ID failed
	cd = New()
	cd.DataRoot = utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	assert.Nil(t, FillSalts(cd))
	assert.Nil(t, CalculateSigningRoot(cd))
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(5)
	c, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	s := identity.Sign(c, identity.KeyPurposeSigning, cd.SigningRoot)
	cd.Signatures = []*coredocumentpb.Signature{s}
	assert.Nil(t, CalculateDocumentRoot(cd))
	assert.Nil(t, err)
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil).Once()
	oldID := cfg.GetString("identityId")
	cfg.Set("identityId", "wrong id")
	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get self cent ID")
	cfg.Set("identityId", "0x0102030405060708")

	// wrong ID
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(5)
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil).Once()

	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "centID invalid")
	cfg.Set("identityId", oldID)

	// failed anchor commit
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(5)
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil).Once()

	repo := mockRepo{}
	repo.On("CommitAnchor", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error")).Once()
	dp.anchorRepository = repo
	err = dp.AnchorDocument(ctxh, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	repo.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to commit anchor")

	// success
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(5)
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil).Once()

	repo = mockRepo{}
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
	ctx := context.Background()
	ctxh, err := header.NewContextHeader(ctx, cfg)
	assert.Nil(t, err)
	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack core document")

	// failed validations
	model = mockModel{}
	cd := new(coredocumentpb.CoreDocument)
	model.On("PackCoreDocument").Return(cd, nil).Times(6)
	err = dp.SendDocument(ctxh, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "post anchor validations failed")

	// failed send
	cd = New()
	cd.DataRoot = utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	cd.Collaborators = [][]byte{[]byte("some id")}
	assert.Nil(t, FillSalts(cd))
	assert.Nil(t, CalculateSigningRoot(cd))
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(6)
	c, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	s := identity.Sign(c, identity.KeyPurposeSigning, cd.SigningRoot)
	cd.Signatures = []*coredocumentpb.Signature{s}
	assert.Nil(t, CalculateDocumentRoot(cd))
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
