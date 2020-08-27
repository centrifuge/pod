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

func TestVersionIDsValidator(t *testing.T) {
	uvv := versionIDsValidator()

	// nil models
	err := uvv.Validate(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need both the old and new model")

	old := new(mockModel)
	old.On("ID").Return(nil).Once()
	old.On("CurrentVersion").Return(nil).Once()
	old.On("NextVersion").Return(nil).Once()
	nm := new(mockModel)
	nm.On("ID").Return(utils.RandomSlice(32)).Once()
	nm.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	nm.On("NextVersion").Return(utils.RandomSlice(32)).Once()
	nm.On("PreviousVersion").Return(utils.RandomSlice(32)).Once()
	err = uvv.Validate(old, nm)
	assert.Error(t, err)
	old.AssertExpectations(t)
	nm.AssertExpectations(t)

	old = new(mockModel)
	pv := utils.RandomSlice(32)
	di := utils.RandomSlice(32)
	cv := utils.RandomSlice(32)
	nv := utils.RandomSlice(32)
	old.On("ID").Return(di).Once()
	old.On("CurrentVersion").Return(pv).Once()
	old.On("NextVersion").Return(cv).Once()
	nm = new(mockModel)
	nm.On("ID").Return(di).Once()
	nm.On("CurrentVersion").Return(cv).Once()
	nm.On("NextVersion").Return(nv).Once()
	nm.On("PreviousVersion").Return(pv).Once()
	err = uvv.Validate(old, nm)
	assert.NoError(t, err)
	old.AssertExpectations(t)
	nm.AssertExpectations(t)

}

func TestUpdateVersionValidator(t *testing.T) {
	uvv := UpdateVersionValidator(nil)
	assert.Len(t, uvv, 3)
}

func TestCreateVersionValidator(t *testing.T) {
	uvv := CreateVersionValidator(nil)
	assert.Len(t, uvv, 3)
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
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid document state transition: error")

	old.On("CollaboratorCanUpdate", updated, id1).Return(nil)
	err = tv.Validate(old.Model, updated)
	assert.NoError(t, err)
}

func TestValidator_SignatureValidator(t *testing.T) {
	account, err := contextutil.Account(testingconfig.CreateAccountContext(t, cfg))
	assert.NoError(t, err)
	anchorSrv := new(mockAnchorService)
	anchorSrv.On("GetAnchorData", mock.Anything).Return(utils.RandomSlice(32), time.Now(), nil)
	idService := new(testingcommons.MockIdentityService)
	sv := SignatureValidator(idService, anchorSrv)

	// fail to get signing root
	model := new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	model.On("CalculateSigningRoot").Return(nil, errors.New("error"))
	model.On("Timestamp").Return(time.Now().UTC(), nil)
	model.On("GetAttributes").Return(nil)
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
	model.On("Timestamp").Return(time.Now().UTC(), nil)
	model.On("GetAttributes").Return(nil)
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

	s2 := &coredocumentpb.Signature{
		Signature: utils.RandomSlice(32),
		SignerId:  utils.RandomSlice(identity.DIDLength),
		PublicKey: utils.RandomSlice(32),
	}

	did1, err := identity.NewDIDFromBytes(s.SignerId)
	assert.NoError(t, err)

	idService = new(testingcommons.MockIdentityService)
	sv = SignatureValidator(idService, anchorSrv)
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Author").Return(did1, nil)
	model.On("Timestamp").Return(tm, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("invalid signature")).Once()
	model.On("Signatures").Return().Once()
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, s)
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))

	// model author not found
	idService = new(testingcommons.MockIdentityService)
	sv = SignatureValidator(idService, anchorSrv)
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Author").Return(testingidentity.GenerateRandomDID(), nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	model.On("Timestamp").Return(tm, nil)
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	model.On("Signatures").Return().Once()
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, s)
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))
	assert.Contains(t, err.Error(), "author's signature missing on document")

	// signer not part of signing collaborators
	idService = new(testingcommons.MockIdentityService)
	sv = SignatureValidator(idService, anchorSrv)
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Author").Return(did1, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	model.On("Timestamp").Return(tm, nil)
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	model.On("Signatures").Return().Once()
	model.On("GetAttributes").Return(nil)
	model.sigs = append(model.sigs, s, s2)
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, 1, errors.Len(err))
	assert.Contains(t, err.Error(), "signer is not part of the signing collaborators")

	// model timestamp err
	idService = new(testingcommons.MockIdentityService)
	sv = SignatureValidator(idService, anchorSrv)
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Author").Return(did1, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	model.On("Timestamp").Return(tm, errors.New("some timestamp error"))
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	model.On("Signatures").Return().Once()
	model.sigs = append(model.sigs, s)
	err = sv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, 2, errors.Len(err))
	assert.Contains(t, err.Error(), "some timestamp error")

	// success
	idService = new(testingcommons.MockIdentityService)
	sv = SignatureValidator(idService, anchorSrv)
	s, err = account.SignMsg(sr)
	assert.NoError(t, err)
	acID := account.GetIdentityID()
	did1, err = identity.NewDIDFromBytes(acID)
	assert.NoError(t, err)
	model = new(mockModel)
	model.On("ID").Return(utils.RandomSlice(32))
	model.On("CurrentVersion").Return(utils.RandomSlice(32))
	model.On("NextVersion").Return(utils.RandomSlice(32))
	model.On("CalculateSigningRoot").Return(sr, nil)
	model.On("Author").Return(did1, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did1, testingidentity.GenerateRandomDID()}, nil)
	model.On("Timestamp").Return(tm, nil)
	idService.On("ValidateSignature", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	model.On("Signatures").Return().Once()
	model.On("GetAttributes").Return(nil)
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
	payload := ConsensusSignaturePayload(sr, false)
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
	did, err := identity.NewDIDFromBytes(s.SignerId)
	assert.NoError(t, err)
	model = new(mockModel)
	model.On("CalculateSigningRoot").Return(sr, nil).Once()
	model.On("Signatures").Return().Once()
	model.On("Author").Return(did, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did, testingidentity.GenerateRandomDID()}, nil)
	model.On("Timestamp").Return(tm, nil)
	model.sigs = append(model.sigs, s)
	srv = new(testingcommons.MockIdentityService)
	sid, err := identity.NewDIDFromBytes(s.SignerId)
	assert.NoError(t, err)
	srv.On("ValidateSignature", sid, s.PublicKey, s.Signature, payload, tm).Return(errors.New("error")).Once()
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
	model.On("Author").Return(did, nil)
	model.On("GetSignerCollaborators", mock.Anything).Return([]identity.DID{did, testingidentity.GenerateRandomDID()}, nil)
	model.On("Timestamp").Return(tm, nil)
	model.sigs = append(model.sigs, s)
	srv = new(testingcommons.MockIdentityService)
	srv.On("ValidateSignature", sid, s.PublicKey, s.Signature, payload, tm).Return(nil).Once()
	ssv = signaturesValidator(srv)
	err = ssv.Validate(nil, model)
	model.AssertExpectations(t)
	srv.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestPreAnchorValidator(t *testing.T) {
	pav := PreAnchorValidator(nil, nil)
	assert.Len(t, pav, 2)
}

func TestValidator_LatestVersionValidator(t *testing.T) {
	anchorSrv := mockAnchorService{}
	//zeros := [32]byte{}
	next := utils.RandomSlice(32)
	nextAid, err := anchors.ToAnchorID(next)

	nonZeroRoot, err := anchors.ToDocumentRoot(utils.RandomSlice(32))
	assert.NoError(t, err)
	zeros := [32]byte{}
	zeroRoot, err := anchors.ToDocumentRoot(zeros[:])
	assert.NoError(t, err)

	// failed to convert to anchor ID
	model := new(mockModel)
	model.On("NextVersion").Return(utils.RandomSlice(10)).Once()
	lv := LatestVersionValidator(anchorSrv)
	err = lv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentIdentifier, err))

	// successful
	anchorSrv.On("GetAnchorData", nextAid).Return(zeroRoot, time.Now(), errors.New("missing"))
	model = new(mockModel)
	model.On("NextVersion").Return(next).Once()
	lv = LatestVersionValidator(anchorSrv)
	err = lv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.NoError(t, err)

	// fail anchor exists
	model = new(mockModel)
	anchorSrv = mockAnchorService{}
	anchorSrv.On("GetAnchorData", nextAid).Return(nonZeroRoot, time.Now(), nil)
	model.On("NextVersion").Return(next).Once()
	lv = LatestVersionValidator(anchorSrv)
	err = lv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentNotLatest, err))
}

func TestValidator_CurrentVersionValidator(t *testing.T) {
	anchorSrv := mockAnchorService{}
	//zeros := [32]byte{}
	next := utils.RandomSlice(32)
	nextAid, err := anchors.ToAnchorID(next)

	nonZeroRoot, err := anchors.ToDocumentRoot(utils.RandomSlice(32))
	assert.NoError(t, err)
	zeros := [32]byte{}
	zeroRoot, err := anchors.ToDocumentRoot(zeros[:])
	assert.NoError(t, err)

	// failed to convert to anchor ID
	model := new(mockModel)
	model.On("CurrentVersion").Return(utils.RandomSlice(10)).Once()
	cv := currentVersionValidator(anchorSrv)
	err = cv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentIdentifier, err))

	// successful
	anchorSrv.On("GetAnchorData", nextAid).Return(zeroRoot, time.Now(), errors.New("missing"))
	model = new(mockModel)
	model.On("CurrentVersion").Return(next).Once()
	cv = currentVersionValidator(anchorSrv)
	err = cv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.NoError(t, err)

	// fail anchor exists
	model = new(mockModel)
	anchorSrv = mockAnchorService{}
	anchorSrv.On("GetAnchorData", nextAid).Return(nonZeroRoot, time.Now(), nil)
	model.On("CurrentVersion").Return(next).Once()
	cv = currentVersionValidator(anchorSrv)
	err = cv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentNotLatest, err))
}

func TestValidator_anchoredValidator(t *testing.T) {
	av := anchoredValidator(mockAnchorService{})

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
	r := &mockAnchorService{}
	av = anchoredValidator(r)
	r.On("GetAnchorData", anchorID).Return(nil, time.Now(), errors.New("error")).Once()
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
	r = &mockAnchorService{}
	av = anchoredValidator(r)
	r.On("GetAnchorData", anchorID).Return(docRoot, time.Now(), nil).Once()
	model = new(mockModel)
	model.On("CurrentVersion").Return(anchorID[:]).Once()
	model.On("CalculateDocumentRoot").Return(utils.RandomSlice(32), nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	r.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatched document roots")

	// anchored after max allowed time
	r = &mockAnchorService{}
	av = anchoredValidator(r)
	tm := time.Now()
	r.On("GetAnchorData", anchorID).Return(docRoot, tm, nil).Once()
	model = new(mockModel)
	model.On("CurrentVersion").Return(anchorID[:]).Once()
	model.On("CalculateDocumentRoot").Return(docRoot[:], nil).Once()
	model.On("Timestamp").Return(tm.Add(-MaxAuthoredToCommitDuration-1), nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	r.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document was anchored after max allowed time for anchor")

	// success
	r = &mockAnchorService{}
	av = anchoredValidator(r)
	r.On("GetAnchorData", anchorID).Return(docRoot, time.Now(), nil).Once()
	model = new(mockModel)
	model.On("CurrentVersion").Return(anchorID[:]).Once()
	model.On("CalculateDocumentRoot").Return(docRoot[:], nil).Once()
	model.On("Timestamp").Return(time.Now(), nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	r.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestPostAnchoredValidator(t *testing.T) {
	pav := PostAnchoredValidator(nil, nil)
	assert.Len(t, pav, 3)
}

func TestDocumentAuthorValidator(t *testing.T) {
	did := testingidentity.GenerateRandomDID()
	av := documentAuthorValidator(did)

	// fail
	model := new(mockModel)
	model.On("Author").Return(testingidentity.GenerateRandomDID(), nil).Once()
	err := av.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document sender is not the author")

	// success
	model = new(mockModel)
	model.On("Author").Return(did, nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestDocumentTimestampForSigningValidator(t *testing.T) {
	av := documentTimestampForSigningValidator()

	// fail
	model := new(mockModel)
	model.On("Timestamp").Return(time.Now().UTC().Add(-MaxAuthoredToCommitDuration), nil).Once()
	err := av.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document is too old to be signed")

	// success
	model = new(mockModel)
	model.On("Timestamp").Return(time.Now().UTC(), nil).Once()
	err = av.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestValidator_anchorRepoAddressValidator(t *testing.T) {
	addr := testingidentity.GenerateRandomDID().ToAddress()
	arv := anchorRepoAddressValidator(addr)

	model := new(mockModel)
	model.On("AnchorRepoAddress").Return(testingidentity.GenerateRandomDID().ToAddress()).Once()
	model.On("AnchorRepoAddress").Return(addr).Once()

	// failure
	err := arv.Validate(nil, model)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "anchor address is not the node configured address")

	// success
	assert.NoError(t, arv.Validate(nil, model))
	model.AssertExpectations(t)
}

func TestValidator_attributeSignatureValidator(t *testing.T) {
	asv := attributeValidator(nil, nil)

	// failed to get timestamp
	model := new(mockModel)
	model.On("Timestamp").Return(nil, errors.New("failed to get timestamp")).Once()
	err := asv.Validate(nil, model)
	assert.Error(t, err)
	model.AssertExpectations(t)

	attrs := []Attribute{
		{
			KeyLabel: "label 1",
			Value:    AttrVal{Type: AttrString},
		},

		{
			KeyLabel: "label 2",
			Value: AttrVal{
				Type: AttrSigned,
				Signed: Signed{
					Identity:        testingidentity.GenerateRandomDID(),
					DocumentVersion: utils.RandomSlice(20),
					Value:           utils.RandomSlice(32),
					Signature:       utils.RandomSlice(32),
					PublicKey:       utils.RandomSlice(32),
				},
			},
		},
	}
	// failed anchor id
	model = new(mockModel)
	model.On("Timestamp").Return(time.Now().UTC(), nil).Once()
	model.On("GetAttributes").Return(attrs).Once()
	err = asv.Validate(nil, model)
	assert.Error(t, err)
	model.AssertExpectations(t)

	// not anchored yet, failed Validation
	id := utils.RandomSlice(32)
	attrs[1].Value.Signed.DocumentVersion = id
	aid, err := anchors.ToAnchorID(id)
	assert.NoError(t, err)
	anchorSrv := new(mockAnchorService)
	anchorSrv.On("GetAnchorData", aid).Return(utils.RandomSlice(32), time.Now(), errors.New("failed to get")).Once()

	ts := time.Now().UTC()
	model = new(mockModel)
	model.On("Timestamp").Return(ts, nil).Once()
	model.On("GetAttributes").Return(attrs).Once()
	docID := utils.RandomSlice(32)
	model.On("ID").Return(docID).Once()

	signed := attrs[1].Value.Signed
	payload := attributeSignaturePayload(signed.Identity[:], docID, id, signed.Value)
	srv := new(testingcommons.MockIdentityService)
	srv.On("ValidateSignature", signed.Identity, signed.PublicKey, signed.Signature, payload, ts).Return(errors.New("failed")).Once()
	asv = attributeValidator(anchorSrv, srv)
	err = asv.Validate(nil, model)
	assert.Error(t, err)
	anchorSrv.AssertExpectations(t)
	srv.AssertExpectations(t)
	model.AssertExpectations(t)

	// success
	anchorSrv = new(mockAnchorService)
	anchorSrv.On("GetAnchorData", aid).Return(utils.RandomSlice(32), time.Now(), errors.New("failed to get")).Once()

	model = new(mockModel)
	model.On("Timestamp").Return(ts, nil).Once()
	model.On("GetAttributes").Return(attrs).Once()
	model.On("ID").Return(docID).Once()
	srv = new(testingcommons.MockIdentityService)
	srv.On("ValidateSignature", signed.Identity, signed.PublicKey, signed.Signature, payload, ts).Return(nil).Once()
	asv = attributeValidator(anchorSrv, srv)
	err = asv.Validate(nil, model)
	assert.NoError(t, err)
	anchorSrv.AssertExpectations(t)
	srv.AssertExpectations(t)
	model.AssertExpectations(t)
}
