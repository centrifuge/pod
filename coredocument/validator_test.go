// +build unit

package coredocument

import (
	"fmt"
	"testing"

	"errors"

	"context"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
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
)

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
	oldCD.DocumentRoot = utils.RandomSlice(32)
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
	cd.DataRoot = utils.RandomSlice(32)
	assert.Nil(t, FillSalts(cd))
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
	assert.Nil(t, FillSalts(cd))
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signing root missing")

	// mismatch signing roots
	cd.SigningRoot = utils.RandomSlice(32)
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
	assert.Nil(t, FillSalts(cd))
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = dv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document root missing")

	// mismatch signing roots
	cd.DocumentRoot = utils.RandomSlice(32)
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
	ctxh, err := header.NewContextHeader(context.Background(), cfg)
	assert.Nil(t, err)
	idKeys := ctxh.Self().Keys[identity.KeyPurposeSigning]
	rfsv := readyForSignaturesValidator(ctxh.Self().ID[:], idKeys.PrivateKey, idKeys.PublicKey)

	// fail getCoreDoc
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("err")).Once()
	err = rfsv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// signature length mismatch
	cd := New()
	assert.Nil(t, FillSalts(cd))
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = rfsv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expecting only one signature")

	// mismatch
	cd.SigningRoot = utils.RandomSlice(32)
	s := &coredocumentpb.Signature{
		Signature: utils.RandomSlice(32),
		EntityId:  utils.RandomSlice(6),
		PublicKey: utils.RandomSlice(32),
	}
	cd.Signatures = append(cd.Signatures, s)
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = rfsv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Len(t, documents.ConvertToMap(err), 3)

	// success
	cd.SigningRoot = utils.RandomSlice(32)
	c, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	s = identity.Sign(c, identity.KeyPurposeSigning, cd.SigningRoot)
	cd.Signatures = []*coredocumentpb.Signature{s}
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = rfsv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestValidator_signatureValidator(t *testing.T) {
	srv := &testingcommons.MockIDService{}
	ssv := signaturesValidator(srv)

	// fail getCoreDoc
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("err")).Once()
	err := ssv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// signature length mismatch
	cd := New()
	assert.Nil(t, FillSalts(cd))
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = ssv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "atleast one signature expected")

	// failed validation
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(errors.New("fail")).Once()
	s := &coredocumentpb.Signature{EntityId: utils.RandomSlice(7)}
	cd.Signatures = append(cd.Signatures, s)
	err = ssv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature verification failed")

	// success
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	srv.On("ValidateSignature", mock.Anything, mock.Anything).Return(nil).Once()
	cd.SigningRoot = utils.RandomSlice(32)
	cd.Signatures = []*coredocumentpb.Signature{{}}

	err = ssv.Validate(nil, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestPreAnchorValidator(t *testing.T) {
	pav := PreAnchorValidator(nil)
	assert.Len(t, pav, 2)
}

type repo struct {
	mock.Mock
	anchors.AnchorRepository
}

func (r repo) GetDocumentRootOf(anchorID anchors.AnchorID) (anchors.DocumentRoot, error) {
	args := r.Called(anchorID)
	docRoot, _ := args.Get(0).(anchors.DocumentRoot)
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
	cd.CurrentVersion = utils.RandomSlice(32)
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get document root")

	// failed to get docRoot from chain
	anchorID, err := anchors.ToAnchorID(utils.RandomSlice(32))
	assert.Nil(t, err)
	r := &repo{}
	av = anchoredValidator(r)
	cd.CurrentVersion = anchorID[:]
	r.On("GetDocumentRootOf", anchorID).Return(nil, fmt.Errorf("error")).Once()
	cd.DocumentRoot = utils.RandomSlice(32)
	model = &mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	r.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get document root from chain")

	// mismatched doc roots
	docRoot := anchors.RandomDocumentRoot()
	r = &repo{}
	av = anchoredValidator(r)
	r.On("GetDocumentRootOf", anchorID).Return(docRoot, nil).Once()
	cd.DocumentRoot = utils.RandomSlice(32)
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

func TestValidate_baseValidator(t *testing.T) {
	tests := []struct {
		doc *coredocumentpb.CoreDocument
		key string
	}{
		// empty salts in document
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id4,
				DataRoot:           id5,
			},
			key: "cd_salts",
		},

		// salts missing previous root
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id4,
				DataRoot:           id5,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentVersion:     id2,
					NextVersion:        id3,
					DataRoot:           id4,
				},
			},
			key: "cd_salts",
		},

		// missing identifiers in core document
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id4,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentVersion:     id2,
					NextVersion:        id3,
					DataRoot:           id4,
					PreviousRoot:       id5,
				},
			},
			key: "cd_data_root",
		},

		// missing identifiers in core document and salts
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id4,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentVersion:     id2,
					NextVersion:        id3,
					DataRoot:           id4,
				},
			},
			key: "cd_data_root",
		},

		// repeated identifiers
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id3,
				DataRoot:           id5,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentVersion:     id2,
					NextVersion:        id3,
					DataRoot:           id4,
					PreviousRoot:       id5,
				},
			},
			key: "cd_overall",
		},

		// repeated identifiers
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id2,
				DataRoot:           id5,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentVersion:     id2,
					NextVersion:        id3,
					DataRoot:           id4,
					PreviousRoot:       id5,
				},
			},
			key: "cd_overall",
		},

		// All okay
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id4,
				DataRoot:           id5,
				CoredocumentSalts: &coredocumentpb.CoreDocumentSalts{
					DocumentIdentifier: id1,
					CurrentVersion:     id2,
					NextVersion:        id3,
					DataRoot:           id4,
					PreviousRoot:       id5,
				},
			},
		},
	}

	baseValidator := baseValidator()

	for _, c := range tests {

		model := mockModel{}
		model.On("PackCoreDocument", mock.Anything).Return(c.doc, nil).Once()

		err := baseValidator.Validate(nil, &model)
		if c.key == "" {
			assert.Nil(t, err)
			continue
		}

		assert.Contains(t, err.Error(), c.key)

	}
}

func TestPostAnchoredValidator(t *testing.T) {
	pav := PostAnchoredValidator(nil, nil)
	assert.Len(t, pav, 2)
}

func TestPreSignatureRequestValidator(t *testing.T) {
	ctxh, err := header.NewContextHeader(context.Background(), cfg)
	assert.Nil(t, err)
	idKeys := ctxh.Self().Keys[identity.KeyPurposeSigning]
	psv := PreSignatureRequestValidator(ctxh.Self().ID[:], idKeys.PrivateKey, idKeys.PublicKey)
	assert.Len(t, psv, 3)
}

func TestPostSignatureRequestValidator(t *testing.T) {
	psv := PostSignatureRequestValidator(nil)
	assert.Len(t, psv, 3)
}

func TestSignatureRequestValidator(t *testing.T) {
	srv := SignatureRequestValidator(nil)
	assert.Len(t, srv, 3)
}
