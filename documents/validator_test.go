// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/coredocument"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/stretchr/testify/assert"
)

type MockValidator struct{}

func (m MockValidator) Validate(oldState Model, newState Model) error {
	return nil
}

type MockValidatorWithErrors struct{}

func (m MockValidatorWithErrors) Validate(oldState Model, newState Model) error {

	err := NewError("error_test", "error msg 1")
	err = errors.AppendError(err, NewError("error_test2", "error msg 2"))

	return err
}

type MockValidatorWithOneError struct{}

func (m MockValidatorWithOneError) Validate(oldState Model, newState Model) error {
	return errors.New("one error")
}

func TestValidatorInterface(t *testing.T) {
	var validator Validator

	// no error
	validator = MockValidator{}
	errs := validator.Validate(nil, nil)
	assert.Nil(t, errs, "")

	//one error
	validator = MockValidatorWithOneError{}
	errs = validator.Validate(nil, nil)
	assert.Error(t, errs, "error should be returned")
	assert.Equal(t, 1, errors.Len(errs), "errors should include one error")

	// more than one error
	validator = MockValidatorWithErrors{}
	errs = validator.Validate(nil, nil)
	assert.Error(t, errs, "error should be returned")
	assert.Equal(t, 2, errors.Len(errs), "errors should include two errors")
}

func TestValidatorGroup_Validate(t *testing.T) {

	var testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithOneError{},
		MockValidatorWithErrors{},
	}
	errs := testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 3, errors.Len(errs), "Validate should return 2 errors")

	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithErrors{},
		MockValidatorWithErrors{},
	}
	errs = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 4, errors.Len(errs), "Validate should return 4 errors")

	// empty group
	testValidatorGroup = ValidatorGroup{}
	errs = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 0, errors.Len(errs), "Validate should return no error")

	// group with no errors at all
	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidator{},
		MockValidator{},
	}
	errs = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 0, errors.Len(errs), "Validate should return no error")
}

func TestIsCurrencyValid(t *testing.T) {
	tests := []struct {
		cur   string
		valid bool
	}{
		{
			cur:   "EUR",
			valid: true,
		},

		{
			cur:   "INR",
			valid: true,
		},

		{
			cur:   "some currency",
			valid: false,
		},
	}

	for _, c := range tests {
		got := IsCurrencyValid(c.cur)
		assert.Equal(t, c.valid, got, "result must match")
	}
}

func TestUpdateVersionValidator(t *testing.T) {
	uvv := UpdateVersionValidator()

	// nil models
	err := uvv.Validate(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need both the old and new model")

	// old model pack core doc fail
	old := mockModel{}
	newM := mockModel{}
	old.On("PackCoreDocument").Return(nil, errors.New("error")).Once()
	err = uvv.Validate(old, newM)
	old.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch old core document")

	// newM model pack core doc fail
	oldCD := coredocument.New()
	oldCD.DocumentRoot = utils.RandomSlice(32)
	old.On("PackCoreDocument").Return(oldCD, nil).Once()
	newM.On("PackCoreDocument").Return(nil, errors.New("error")).Once()
	err = uvv.Validate(old, newM)
	old.AssertExpectations(t)
	newM.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch new core document")

	// mismatched identifiers
	newCD := coredocument.New()
	newCD.NextVersion = nil
	old.On("PackCoreDocument").Return(oldCD, nil).Once()
	newM.On("PackCoreDocument").Return(newCD, nil).Once()
	err = uvv.Validate(old, newM)
	old.AssertExpectations(t)
	newM.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, 5, errors.Len(err))

	// success
	newCD, err = coredocument.PrepareNewVersion(*oldCD, nil)
	assert.Nil(t, err)
	old.On("PackCoreDocument").Return(oldCD, nil).Once()
	newM.On("PackCoreDocument").Return(newCD, nil).Once()
	err = uvv.Validate(old, newM)
	old.AssertExpectations(t)
	newM.AssertExpectations(t)
	assert.Nil(t, err)
}

func Test_getCoreDocument(t *testing.T) {
	// nil document
	cd, err := getCoreDocument(nil)
	assert.Error(t, err)
	assert.Nil(t, cd)

	// pack core document fail
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("err")).Once()
	cd, err = getCoreDocument(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, cd)

	// success
	model = mockModel{}
	cd = coredocument.New()
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
	model.On("PackCoreDocument").Return(nil, errors.New("err")).Once()
	err := bv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// failed validator
	model = mockModel{}
	cd := coredocument.New()
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = bv.Validate(nil, model)
	assert.Error(t, err)
	assert.Equal(t, "cd_salts : Required field", errors.GetErrs(err)[0].Error())

	// success
	model = mockModel{}
	cd.DataRoot = utils.RandomSlice(32)
	assert.Nil(t, coredocument.FillSalts(cd))
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = bv.Validate(nil, model)
	assert.Nil(t, err)
}

func TestValidator_signingRootValidator(t *testing.T) {
	sv := signingRootValidator()

	// fail getCoreDoc
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("err")).Once()
	err := sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// missing signing_root
	cd := coredocument.New()
	assert.Nil(t, coredocument.FillSalts(cd))
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
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signing root mismatch")

	// success
	tree, err := coredocument.GetDocumentSigningTree(cd, cd.DataRoot)
	assert.Nil(t, err)
	cd.SigningRoot = tree.RootHash()
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	model.On("CalculateDataRoot").Return(cd.DataRoot, nil)
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestValidator_documentRootValidator(t *testing.T) {
	dv := documentRootValidator()

	// fail getCoreDoc
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("err")).Once()
	err := dv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// missing document root
	cd := coredocument.New()
	assert.Nil(t, coredocument.FillSalts(cd))
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
	tree, err := coredocument.GetDocumentRootTree(cd)
	assert.Nil(t, err)
	cd.DocumentRoot = tree.RootHash()
	model = mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = dv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestValidator_selfSignatureValidator(t *testing.T) {
	self, _ := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))

	idKeys := self.Keys[identity.KeyPurposeSigning]
	rfsv := readyForSignaturesValidator(self.ID[:], idKeys.PrivateKey, idKeys.PublicKey)

	// fail getCoreDoc
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("err")).Once()
	err := rfsv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// signature length mismatch
	cd := coredocument.New()
	assert.Nil(t, coredocument.FillSalts(cd))
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
	assert.Equal(t, 3, errors.Len(err))

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
	srv := &testingcommons.MockIdentityService{}
	ssv := signaturesValidator(srv)

	// fail getCoreDoc
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, errors.New("err")).Once()
	err := ssv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// signature length mismatch
	cd := coredocument.New()
	assert.Nil(t, coredocument.FillSalts(cd))
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

func TestValidator_anchoredValidator(t *testing.T) {
	av := anchoredValidator(mockRepo{})

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
	r := &mockRepo{}
	av = anchoredValidator(r)
	cd.CurrentVersion = anchorID[:]
	r.On("GetDocumentRootOf", anchorID).Return(nil, errors.New("error")).Once()
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
	r = &mockRepo{}
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
	r = &mockRepo{}
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

var (
	id1 = utils.RandomSlice(32)
	id2 = utils.RandomSlice(32)
	id3 = utils.RandomSlice(32)
	id4 = utils.RandomSlice(32)
	id5 = utils.RandomSlice(32)
)

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
			key: "[cd_salts : Required field]",
		},

		// salts wrong length previous root
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id4,
				CoredocumentSalts: []*coredocumentpb.DocumentSalt{
					{Value: id1},
					{Value: id2},
					{Value: id3},
					{Value: id5[5:]},
				},
			},
			key: "[cd_salts : Required field]",
		},

		// repeated identifiers
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id3,
				DataRoot:           id5,
				CoredocumentSalts: []*coredocumentpb.DocumentSalt{
					{Value: id1},
					{Value: id2},
					{Value: id3},
					{Value: id5},
				},
			},
			key: "[cd_overall : Identifier re-used]",
		},

		// repeated identifiers
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id2,
				DataRoot:           id5,
				CoredocumentSalts: []*coredocumentpb.DocumentSalt{
					{Value: id1},
					{Value: id2},
					{Value: id3},
					{Value: id5},
				},
			},
			key: "[cd_overall : Identifier re-used]",
		},

		// All okay
		{
			doc: &coredocumentpb.CoreDocument{
				DocumentRoot:       id1,
				DocumentIdentifier: id2,
				CurrentVersion:     id3,
				NextVersion:        id4,
				DataRoot:           id5,
				CoredocumentSalts: []*coredocumentpb.DocumentSalt{
					{Value: id1},
					{Value: id2},
					{Value: id3},
					{Value: id5},
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

		assert.Equal(t, c.key, err.Error())

	}
}

func TestPostAnchoredValidator(t *testing.T) {
	pav := PostAnchoredValidator(nil, nil)
	assert.Len(t, pav, 2)
}

func TestPreSignatureRequestValidator(t *testing.T) {
	self, _ := contextutil.Self(testingconfig.CreateAccountContext(t, cfg))
	idKeys := self.Keys[identity.KeyPurposeSigning]
	psv := PreSignatureRequestValidator(self.ID[:], idKeys.PrivateKey, idKeys.PublicKey)
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
