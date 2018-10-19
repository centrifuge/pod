// +build unit

package coredocument

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519keys"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&config.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, nil)
	flag.Parse()
	config.Config.V.Set("keys.signing.publicKey", "../../example/resources/signature1.pub.pem")
	config.Config.V.Set("keys.signing.privateKey", "../../example/resources/signature1.key.pem")
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
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

func TestUpdateVersionValidator(t *testing.T) {
	uvv := UpdateVersionValidator()

	// nil models
	err := uvv.Validate(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need both the old and new model")

	// old model pack core doc fail
	old := mockModel{}
	new := mockModel{}
	old.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err = uvv.Validate(old, new)
	old.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch old core document")

	// new model pack core doc fail
	oldCD := New()
	oldCD.DocumentRoot = tools.RandomSlice(32)
	old.On("PackCoreDocument").Return(oldCD, nil).Once()
	new.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err = uvv.Validate(old, new)
	old.AssertExpectations(t)
	new.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch new core document")

	// mismatched identifiers
	newCD := New()
	newCD.NextVersion = nil
	old.On("PackCoreDocument").Return(oldCD, nil).Once()
	new.On("PackCoreDocument").Return(newCD, nil).Once()
	err = uvv.Validate(old, new)
	old.AssertExpectations(t)
	new.AssertExpectations(t)
	assert.Error(t, err)
	assert.Len(t, documents.ConvertToMap(err), 4)

	// success
	newCD, err = PrepareNewVersion(*oldCD, nil)
	assert.Nil(t, err)
	old.On("PackCoreDocument").Return(oldCD, nil).Once()
	new.On("PackCoreDocument").Return(newCD, nil).Once()
	err = uvv.Validate(old, new)
	old.AssertExpectations(t)
	new.AssertExpectations(t)
	assert.Nil(t, err)
}

func Test_getCoreDocument(t *testing.T) {
	// nil document
	cd, err := getCoreDocument(nil)
	assert.Error(t, err)
	assert.Nil(t, cd)

	// pack core document fail
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("err")).Once()
	cd, err = getCoreDocument(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, cd)

	// success
	model = mockModel{}
	cd = New()
	model.On("PackCoreDocument").Return(cd, nil).Once()
	got, err := getCoreDocument(model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, cd, got)
}

func TestValidator_baseValidator(t *testing.T) {
	bv := baseValidator()

	// fail getCoreDocument
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("err")).Once()
	err := bv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// failed validator
	model = mockModel{}
	cd := New()
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = bv.Validate(nil, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cd_salts : Required field")

	// success
	model = mockModel{}
	cd.DataRoot = tools.RandomSlice(32)
	FillSalts(cd)
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = bv.Validate(nil, model)
	assert.Nil(t, err)
}

func TestValidator_signingRootValidator(t *testing.T) {
	sv := signingRootValidator()

	// fail getCoreDoc
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("err")).Once()
	err := sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// missing signing_root
	cd := New()
	FillSalts(cd)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signing root missing")

	// mismatch signing roots
	cd.SigningRoot = tools.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signing root mismatch")

	// success
	tree, err := GetDocumentSigningTree(cd)
	assert.Nil(t, err)
	cd.SigningRoot = tree.RootHash()
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestValidator_documentRootValidator(t *testing.T) {
	dv := documentRootValidator()

	// fail getCoreDoc
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("err")).Once()
	err := dv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// missing document root
	cd := New()
	FillSalts(cd)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = dv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document root missing")

	// mismatch signing roots
	cd.DocumentRoot = tools.RandomSlice(32)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = dv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document root mismatch")

	// success
	tree, err := GetDocumentRootTree(cd)
	assert.Nil(t, err)
	cd.DocumentRoot = tree.RootHash()
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = dv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestValidator_selfSignatureValidator(t *testing.T) {
	rfsv := readyForSignaturesValidator()

	// fail getCoreDoc
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("err")).Once()
	err := rfsv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// signature length mismatch
	cd := New()
	FillSalts(cd)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = rfsv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expecting only one signature")

	// mismatch
	cd.SigningRoot = tools.RandomSlice(32)
	s := &coredocumentpb.Signature{
		Signature: tools.RandomSlice(32),
		EntityId:  tools.RandomSlice(6),
		PublicKey: tools.RandomSlice(32),
	}
	cd.Signatures = append(cd.Signatures, s)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = rfsv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Len(t, documents.ConvertToMap(err), 3)

	// success
	cd.SigningRoot = tools.RandomSlice(32)
	c, err := ed25519keys.GetIDConfig()
	assert.Nil(t, err)
	s = signatures.Sign(c, cd.SigningRoot)
	cd.Signatures = []*coredocumentpb.Signature{s}
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = rfsv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestValidator_signatureValidator(t *testing.T) {
	ssv := signaturesValidator()

	// fail getCoreDoc
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("err")).Once()
	err := ssv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// signature length mismatch
	cd := New()
	FillSalts(cd)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = ssv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "atleast one signature expected")

	// failed validation
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	s := &coredocumentpb.Signature{EntityId: tools.RandomSlice(7)}
	cd.Signatures = append(cd.Signatures, s)
	err = ssv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")

	// success
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	cd.SigningRoot = tools.RandomSlice(32)
	c, err := ed25519keys.GetIDConfig()
	assert.Nil(t, err)
	s = signatures.Sign(c, cd.SigningRoot)
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
	err = ssv.Validate(nil, model)
	model.AssertExpectations(t)
	id.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestPreAnchorValidator(t *testing.T) {
	pav := PreAnchorValidator()
	assert.Len(t, pav, 2)
}

type repo struct {
	mock.Mock
	anchors.AnchorRepository
}

func (r repo) GetDocumentRootOf(anchorID anchors.AnchorID) (anchors.DocRoot, error) {
	args := r.Called(anchorID)
	docRoot, _ := args.Get(0).(anchors.DocRoot)
	return docRoot, args.Error(1)
}

func TestValidator_anchoredValidator(t *testing.T) {
	av := anchoredValidator(repo{})

	// fail get core document
	err := av.Validate(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get core document")

	// failed anchorID
	model := &mockModel{}
	cd := &coredocumentpb.CoreDocument{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get anchorID")

	// failed docRoot
	model = &mockModel{}
	cd.CurrentVersion = tools.RandomSlice(32)
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get document root")

	// failed to get docRoot from chain
	anchorID, err := anchors.NewAnchorID(tools.RandomSlice(32))
	assert.Nil(t, err)
	r := &repo{}
	av = anchoredValidator(r)
	cd.CurrentVersion = anchorID[:]
	r.On("GetDocumentRootOf", anchorID).Return(nil, fmt.Errorf("error")).Once()
	cd.DocumentRoot = tools.RandomSlice(32)
	model = &mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	r.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get document root from chain")

	// mismatched doc roots
	docRoot := anchors.NewRandomDocRoot()
	r = &repo{}
	av = anchoredValidator(r)
	r.On("GetDocumentRootOf", anchorID).Return(docRoot, nil).Once()
	cd.DocumentRoot = tools.RandomSlice(32)
	model = &mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	r.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatched document roots")

	// success
	r = &repo{}
	av = anchoredValidator(r)
	r.On("GetDocumentRootOf", anchorID).Return(docRoot, nil).Once()
	cd.DocumentRoot = docRoot[:]
	model = &mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	r.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestPostAnchoredValidator(t *testing.T) {
	pav := PostAnchoredValidator(nil)
	assert.Len(t, pav, 2)
}

func TestPreSignatureRequestValidator(t *testing.T) {
	psv := PreSignatureRequestValidator()
	assert.Len(t, psv, 3)
}

func TestPostSignatureRequestValidator(t *testing.T) {
	psv := PostSignatureRequestValidator()
	assert.Len(t, psv, 3)
}

func TestSignatureRequestValidator(t *testing.T) {
	srv := SignatureRequestValidator()
	assert.Len(t, srv, 3)
}
