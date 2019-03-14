// +build unit

package documents

import (
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	old := new(mockModel)
	old.On("CalculateDocumentRoot").Return(nil, errors.New("errors")).Once()
	err = uvv.Validate(old, new(mockModel))
	assert.Error(t, err)
	old.AssertExpectations(t)

	old = new(mockModel)
	old.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil).Once()
	old.On("ID").Return(nil).Once()
	old.On("CurrentVersion").Return(nil).Once()
	old.On("NextVersion").Return(nil).Once()
	nm := new(mockModel)
	nm.On("ID").Return(utils.RandomSlice(32)).Once()
	nm.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	nm.On("NextVersion").Return(utils.RandomSlice(32)).Once()
	nm.On("PreviousVersion").Return(utils.RandomSlice(32)).Once()
	nm.On("PreviousDocumentRoot").Return(utils.RandomSlice(32)).Once()
	err = uvv.Validate(old, nm)
	assert.Error(t, err)
	old.AssertExpectations(t)
	nm.AssertExpectations(t)

	old = new(mockModel)
	dpr := utils.RandomSlice(32)
	pv := utils.RandomSlice(32)
	di := utils.RandomSlice(32)
	cv := utils.RandomSlice(32)
	nv := utils.RandomSlice(32)
	old.On("CalculateDocumentRoot").Return(dpr, nil).Once()
	old.On("ID").Return(di).Once()
	old.On("CurrentVersion").Return(pv).Once()
	old.On("NextVersion").Return(cv).Once()
	nm = new(mockModel)
	nm.On("ID").Return(di).Once()
	nm.On("CurrentVersion").Return(cv).Once()
	nm.On("NextVersion").Return(nv).Once()
	nm.On("PreviousVersion").Return(pv).Once()
	nm.On("PreviousDocumentRoot").Return(dpr).Once()
	err = uvv.Validate(old, nm)
	assert.NoError(t, err)
	old.AssertExpectations(t)
	nm.AssertExpectations(t)

}

func TestValidator_baseValidator(t *testing.T) {
	bv := baseValidator()

	model := new(mockModel)
	model.On("ID").Return(nil).Times(2)
	model.On("CurrentVersion").Return(nil).Times(2)
	model.On("NextVersion").Return(nil).Times(3)
	err := bv.Validate(nil, model)
	assert.Error(t, err)

	// success
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32)).Times(2)
	model.On("CurrentVersion").Return(utils.RandomSlice(32)).Times(2)
	model.On("NextVersion").Return(utils.RandomSlice(32)).Times(3)
	err = bv.Validate(nil, model)
	assert.Nil(t, err)
}

func TestValidator_signingRootValidator(t *testing.T) {
	sv := signingRootValidator()

	// failed to get signing root
	model := new(mockModel)
	model.On("CalculateSigningRoot").Return(nil, errors.New("error")).Once()
	err := sv.Validate(nil, model)
	assert.Error(t, err)
	model.AssertExpectations(t)

	// invalid signing root
	model = new(mockModel)
	model.On("CalculateSigningRoot").Return(utils.RandomSlice(30), nil).Once()
	err = sv.Validate(nil, model)
	assert.Error(t, err)
	model.AssertExpectations(t)

	// success
	model = new(mockModel)
	model.On("CalculateSigningRoot").Return(utils.RandomSlice(32), nil).Once()
	err = sv.Validate(nil, model)
	assert.NoError(t, err)
	model.AssertExpectations(t)
}

func TestValidator_documentRootValidator(t *testing.T) {
	dv := documentRootValidator()

	// failed to get document root
	model := new(mockModel)
	model.On("CalculateDocumentRoot").Return(nil, errors.New("error")).Once()
	err := dv.Validate(nil, model)
	assert.Error(t, err)
	model.AssertExpectations(t)

	// invalid signing root
	model = new(mockModel)
	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(30), nil).Once()
	err = dv.Validate(nil, model)
	assert.Error(t, err)
	model.AssertExpectations(t)

	// success
	model = new(mockModel)
	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil).Once()
	err = dv.Validate(nil, model)
	assert.NoError(t, err)
	model.AssertExpectations(t)
}

func TestValidator_TransitionValidator(t *testing.T) {
	id1 := testingidentity.GenerateRandomDID()
	updated := new(mockModel)

	// does not error out if there is no old document model (if new model is the first version of the document model)
	tv := transitionValidator(id1)
	err := tv.Validate(nil, updated)
	assert.NoError(t, err)

	old := new(mockModel)
	old.On("CollaboratorCanUpdate", updated, id1).Return(errors.New("error"))
	err = tv.Validate(old, updated)
	assert.Contains(t, err.Error(), "invalid document state transition: error")

	old.On("CollaboratorCanUpdate", updated, id1).Return(nil)
	err = tv.Validate(old.Model, updated)
	assert.NoError(t, err)
}

func TestValidator_SignatureValidator(t *testing.T) {
	account, err := contextutil.Account(testingconfig.CreateAccountContext(t, cfg))
	assert.NoError(t, err)
	idService := new(testingcommons.MockIdentityService)
	sv := SignatureValidator(idService)

	// fail to get signing root
	model := new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	model.On("CalculateSigningRoot").Return(nil, errors.New("error"))
	err = sv.Validate(nil, model)
	assert.Error(t, err)
	model.AssertExpectations(t)

	// signature length mismatch
	sr := utils.RandomSlice(32)
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Signatures").Return().Once()
	err = sv.Validate(nil, model)
	assert.Error(t, err)
	model.AssertExpectations(t)
	assert.Contains(t, err.Error(), "atleast one signature expected")

	// mismatch
	tm := time.Now()
	s := &coredocumentpb.Signature{
		Signature: utils.RandomSlice(32),
		SignerId:  utils.RandomSlice(identity.DIDLength),
		PublicKey: utils.RandomSlice(32),
	}

	idService = new(testingcommons.MockIdentityService)
	sv = SignatureValidator(idService)
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Author").Return(identity.NewDIDFromBytes(s.SignerId))
	model.On("Timestamp").Return(tm, nil)
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("invalid signature")).Once()
	model.On("Signatures").Return().Once()
	model.sigs = append(model.sigs, s)
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))

	// model author not found
	idService = new(testingcommons.MockIdentityService)
	sv = SignatureValidator(idService)
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Author").Return(identity.NewDIDFromBytes(s.SignerId))
	model.On("Timestamp").Return(tm, nil)
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("invalid signature")).Once()
	model.On("Signatures").Return().Once()
	model.sigs = append(model.sigs, s)
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))

	// model timestamp err
	idService = new(testingcommons.MockIdentityService)
	sv = SignatureValidator(idService)
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Author").Return(identity.NewDIDFromBytes(s.SignerId))
	model.On("Timestamp").Return(tm, errors.New("some timestamp error"))
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("invalid signature")).Once()
	model.On("Signatures").Return().Once()
	model.sigs = append(model.sigs, s)
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))

	// success
	idService = new(testingcommons.MockIdentityService)
	sv = SignatureValidator(idService)
	s, err = account.SignMsg(sr)
	assert.NoError(t, err)
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Author").Return(identity.NewDIDFromBytes(s.SignerId))
	model.On("Timestamp").Return(tm, nil)
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	model.On("Signatures").Return().Once()
	model.sigs = append(model.sigs, s)
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.NoError(t, err)
}

func TestValidator_signatureValidator(t *testing.T) {
	srv := &testingcommons.MockIdentityService{}
	ssv := signaturesValidator(srv)

	// fail to get signing root
	model := new(mockModel)
	model.On("CalculateSigningRoot").Return(nil, errors.New("error")).Once()
	err := ssv.Validate(nil, model)
	assert.Error(t, err)
	model.AssertExpectations(t)

	// signature length mismatch
	sr := utils.RandomSlice(32)
	model = new(mockModel)
	model.On("CalculateSigningRoot").Return(sr, nil).Once()
	model.On("Signatures").Return().Once()
	err = ssv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "atleast one signature expected")

	// failed validation
	tm := time.Now().UTC()
	s := &coredocumentpb.Signature{
		Signature: utils.RandomSlice(32),
		SignerId:  utils.RandomSlice(identity.DIDLength),
		PublicKey: utils.RandomSlice(32),
	}
	model = new(mockModel)
	model.On("CalculateSigningRoot").Return(sr, nil).Once()
	model.On("Signatures").Return().Once()
	model.On("Author").Return(identity.NewDIDFromBytes(s.SignerId))
	model.On("Timestamp").Return(tm, nil)
	model.sigs = append(model.sigs, s)
	srv = new(testingcommons.MockIdentityService)
	srv.On("ValidateSignature", identity.NewDIDFromBytes(s.SignerId), s.PublicKey, s.Signature, sr, tm).Return(errors.New("error")).Once()
	ssv = signaturesValidator(srv)
	err = ssv.Validate(nil, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification failed")

	// success
	model = new(mockModel)
	model.On("CalculateSigningRoot").Return(sr, nil).Once()
	model.On("Signatures").Return().Once()
	model.On("Author").Return(identity.NewDIDFromBytes(s.SignerId))
	model.On("Timestamp").Return(tm, nil)
	model.sigs = append(model.sigs, s)
	srv = new(testingcommons.MockIdentityService)
	srv.On("ValidateSignature", identity.NewDIDFromBytes(s.SignerId), s.PublicKey, s.Signature, sr, tm).Return(nil).Once()
	ssv = signaturesValidator(srv)
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

	// failed anchorID
	model := new(mockModel)
	model.On("CurrentVersion").Return(nil).Once()
	err := av.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get anchorID")

	// failed docRoot
	model = new(mockModel)
	model.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	model.On("CalculateDocumentRoot").Return(nil, errors.New("error")).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get document root")

	// invalid doc root
	model = new(mockModel)
	model.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(30), nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get document root")

	// failed to get docRoot from chain
	anchorID, err := anchors.ToAnchorID(utils.RandomSlice(32))
	assert.Nil(t, err)
	r := &mockRepo{}
	av = anchoredValidator(r)
	r.On("GetDocumentRootOf", anchorID).Return(nil, errors.New("error")).Once()
	model = new(mockModel)
	model.On("CurrentVersion").Return(anchorID[:]).Once()
	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	r.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get document root for anchor")

	// mismatched doc roots
	docRoot := anchors.RandomDocumentRoot()
	r = &mockRepo{}
	av = anchoredValidator(r)
	r.On("GetDocumentRootOf", anchorID).Return(docRoot, nil).Once()
	model = new(mockModel)
	model.On("CurrentVersion").Return(anchorID[:]).Once()
	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	r.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatched document roots")

	// success
	r = &mockRepo{}
	av = anchoredValidator(r)
	r.On("GetDocumentRootOf", anchorID).Return(docRoot, nil).Once()
	model = new(mockModel)
	model.On("CurrentVersion").Return(anchorID[:]).Once()
	model.On("CalculateDocumentRoot").Return(docRoot[:], nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	r.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestPostAnchoredValidator(t *testing.T) {
	pav := PostAnchoredValidator(nil, nil)
	assert.Len(t, pav, 2)
}

func TestSignatureRequestValidator(t *testing.T) {
	srv := SignatureRequestValidator(testingidentity.GenerateRandomDID(), nil)
	assert.Len(t, srv, 2)
}
