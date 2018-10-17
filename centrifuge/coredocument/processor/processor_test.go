// +build unit

package coredocumentprocessor

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519keys"
	"github.com/centrifuge/go-centrifuge/centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/ed25519"
)

var dp defaultProcessor

func TestMain(m *testing.M) {
	dp = defaultProcessor{}
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers)
	result := m.Run()
	os.Exit(result)
}

func TestCoreDocumentProcessor_SendNilDocument(t *testing.T) {
	err := dp.Send(nil, nil, [identity.CentIDLength]byte{})
	assert.Error(t, err, "should have thrown an error")
}

func TestCoreDocumentProcessor_AnchorNilDocument(t *testing.T) {
	err := dp.Anchor(nil, nil, nil)
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
	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err := dp.PrepareForSignatureRequests(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack coredocument")

	// signing root failed
	cd := new(coredocumentpb.CoreDocument)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = dp.PrepareForSignatureRequests(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to calculate signing root")

	// failed to get id
	pub, _ := config.Config.GetSigningKeyPair()
	config.Config.V.Set("keys.signing.publicKey", "wrong path")
	cd = coredocument.New()
	cd.DataRoot = tools.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	coredocument.FillSalts(cd)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = dp.PrepareForSignatureRequests(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get keys for signing")
	config.Config.V.Set("keys.signing.publicKey", pub)

	// failed unpack
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.On("UnpackCoreDocument", cd).Return(fmt.Errorf("error")).Once()
	err = dp.PrepareForSignatureRequests(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unpack the core document")

	// success
	cd.Signatures = nil
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.On("UnpackCoreDocument", cd).Return(nil).Once()
	err = dp.PrepareForSignatureRequests(model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, cd.Signatures)
	assert.Len(t, cd.Signatures, 1)
	sig := cd.Signatures[0]
	id, err := ed25519keys.GetIDConfig()
	assert.Nil(t, err)
	assert.True(t, ed25519.Verify(id.PublicKey, cd.SigningRoot, sig.Signature))
}

type p2pClient struct {
	mock.Mock
	p2p.Client
}

func (p p2pClient) GetSignaturesForDocument(ctx context.Context, doc *coredocumentpb.CoreDocument) error {
	args := p.Called(ctx, doc)
	return args.Error(0)
}

func TestDefaultProcessor_RequestSignatures(t *testing.T) {
	// pack failed
	ctx := context.Background()
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err := dp.RequestSignatures(ctx, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack coredocument")

	// validations failed
	cd := new(coredocumentpb.CoreDocument)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	err = dp.RequestSignatures(ctx, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate model for signature request")

	// failed signature collection
	cd = coredocument.New()
	cd.DataRoot = tools.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	coredocument.FillSalts(cd)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.On("UnpackCoreDocument", cd).Return(nil).Once()
	err = dp.PrepareForSignatureRequests(model)
	assert.Nil(t, err)
	model.AssertExpectations(t)
	c := p2pClient{}
	c.On("GetSignaturesForDocument", ctx, cd).Return(fmt.Errorf("error")).Once()
	dp.P2PClient = c
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	err = dp.RequestSignatures(ctx, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to collect signatures from the collaborators")

	// unpack fail
	c = p2pClient{}
	c.On("GetSignaturesForDocument", ctx, cd).Return(nil).Once()
	dp.P2PClient = c
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	model.On("UnpackCoreDocument", cd).Return(fmt.Errorf("error")).Once()
	err = dp.RequestSignatures(ctx, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unpack core document")

	// success
	c = p2pClient{}
	c.On("GetSignaturesForDocument", ctx, cd).Return(nil).Once()
	dp.P2PClient = c
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	model.On("UnpackCoreDocument", cd).Return(nil).Once()
	err = dp.RequestSignatures(ctx, model)
	model.AssertExpectations(t)
	c.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestDefaultProcessor_PrepareForAnchoring(t *testing.T) {
	// pack failed
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err := dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack coredocument")

	// failed validations
	cd := new(coredocumentpb.CoreDocument)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	err = dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate signatures")

	// failed unpack
	cd = coredocument.New()
	cd.DataRoot = tools.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: "some type",
		Value:   []byte("some data"),
	}
	coredocument.FillSalts(cd)
	err = coredocument.CalculateSigningRoot(cd)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	model.On("UnpackCoreDocument", cd).Return(fmt.Errorf("error")).Once()
	c, err := ed25519keys.GetIDConfig()
	assert.Nil(t, err)
	s := signatures.Sign(c, cd.SigningRoot)
	cd.Signatures = []*coredocumentpb.Signature{s}
	pubkey, err := tools.SliceToByte32(c.PublicKey)
	assert.Nil(t, err)
	idkey := &identity.EthereumIdentityKey{
		Key:       pubkey,
		Purposes:  []*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	id := &testingcommons.MockID{}
	srv := &testingcommons.MockIDService{}
	centID, err := identity.ToCentID(c.ID)
	assert.Nil(t, err)
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubkey[:]).Return(idkey, nil).Once()
	identity.IDService = srv
	err = dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unpack core document")

	// success
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Times(4)
	model.On("UnpackCoreDocument", cd).Return(nil).Once()
	id = &testingcommons.MockID{}
	srv = &testingcommons.MockIDService{}
	centID, err = identity.ToCentID(c.ID)
	assert.Nil(t, err)
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubkey[:]).Return(idkey, nil).Once()
	identity.IDService = srv
	err = dp.PrepareForAnchoring(model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err)
	assert.NotNil(t, cd.DocumentRoot)
}
