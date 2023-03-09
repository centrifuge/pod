//go:build unit

package documents

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/pallets/anchors"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/centrifuge/pod/testingutils/path"
	"github.com/centrifuge/pod/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

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

func TestValidator_versionIDsValidator(t *testing.T) {
	uvv := versionIDsValidator()

	// nil models
	err := uvv.Validate(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need both the old and new model")

	old := NewDocumentMock(t)
	old.On("ID").Return(nil).Once()
	old.On("CurrentVersion").Return(nil).Once()
	old.On("NextVersion").Return(nil).Once()

	nm := NewDocumentMock(t)
	nm.On("ID").Return(utils.RandomSlice(32)).Once()
	nm.On("CurrentVersion").Return(utils.RandomSlice(32)).Once()
	nm.On("NextVersion").Return(utils.RandomSlice(32)).Once()
	nm.On("PreviousVersion").Return(utils.RandomSlice(32)).Once()

	err = uvv.Validate(old, nm)
	assert.Error(t, err)

	old = NewDocumentMock(t)
	pv := utils.RandomSlice(32)
	di := utils.RandomSlice(32)
	cv := utils.RandomSlice(32)
	nv := utils.RandomSlice(32)

	old.On("ID").Return(di).Once()
	old.On("CurrentVersion").Return(pv).Once()
	old.On("NextVersion").Return(cv).Once()

	nm = NewDocumentMock(t)
	nm.On("ID").Return(di).Once()
	nm.On("CurrentVersion").Return(cv).Once()
	nm.On("NextVersion").Return(nv).Once()
	nm.On("PreviousVersion").Return(pv).Once()

	err = uvv.Validate(old, nm)
	assert.NoError(t, err)
}

func TestValidator_baseValidator(t *testing.T) {
	bv := baseValidator()

	err := bv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	model := NewDocumentMock(t)
	model.On("ID").Return(nil).Times(2)
	model.On("CurrentVersion").Return(nil).Times(1)
	model.On("NextVersion").Return(nil).Times(2)
	err = bv.Validate(nil, model)
	assert.Error(t, err)

	// success
	model = NewDocumentMock(t)
	model.On("ID").Return(utils.RandomSlice(32)).Times(2)
	model.On("CurrentVersion").Return(utils.RandomSlice(32)).Times(2)
	model.On("NextVersion").Return(utils.RandomSlice(32)).Times(3)
	err = bv.Validate(nil, model)
	assert.Nil(t, err)
}

func TestValidator_signingRootValidator(t *testing.T) {
	sv := signingRootValidator()

	err := sv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	// failed to get signing root
	model := NewDocumentMock(t)
	model.On("CalculateSigningRoot").
		Return(nil, errors.New("error")).
		Once()

	err = sv.Validate(nil, model)
	assert.True(t, errors.IsOfType(ErrDocumentCalculateSigningRoot, err))

	// invalid signing root
	model = NewDocumentMock(t)
	model.On("CalculateSigningRoot").
		Return(utils.RandomSlice(idSize-1), nil).
		Once()

	err = sv.Validate(nil, model)
	assert.ErrorIs(t, err, ErrInvalidSigningRoot)

	// success
	model = NewDocumentMock(t)
	model.On("CalculateSigningRoot").
		Return(utils.RandomSlice(idSize), nil).
		Once()

	err = sv.Validate(nil, model)
	assert.NoError(t, err)
}

func TestValidator_documentRootValidator(t *testing.T) {
	dv := documentRootValidator()

	err := dv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	// failed to get document root
	model := NewDocumentMock(t)
	model.On("CalculateDocumentRoot").
		Return(nil, errors.New("error")).
		Once()

	err = dv.Validate(nil, model)
	assert.True(t, errors.IsOfType(ErrDocumentCalculateDocumentRoot, err))

	// invalid signing root
	model = NewDocumentMock(t)
	model.On("CalculateDocumentRoot").
		Return(utils.RandomSlice(idSize-1), nil).
		Once()

	err = dv.Validate(nil, model)
	assert.ErrorIs(t, err, ErrInvalidDocumentRoot)

	// success
	model = NewDocumentMock(t)
	model.On("CalculateDocumentRoot").
		Return(utils.RandomSlice(idSize), nil).
		Once()

	err = dv.Validate(nil, model)
	assert.NoError(t, err)
}

func TestValidator_documentAuthorValidator(t *testing.T) {
	sender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	av := documentAuthorValidator(sender)

	err = av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	nonSender, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// fail
	model := NewDocumentMock(t)
	model.On("Author").
		Return(nil, errors.New("error")).
		Once()

	err = av.Validate(nil, model)
	assert.ErrorIs(t, err, ErrDocumentAuthorRetrieval)

	model = NewDocumentMock(t)
	model.On("Author").
		Return(nonSender, nil).
		Once()

	err = av.Validate(nil, model)
	assert.ErrorIs(t, err, ErrDocumentSenderNotAuthor)

	// success
	model = NewDocumentMock(t)
	model.On("Author").
		Return(sender, nil).
		Once()

	err = av.Validate(nil, model)
	assert.Nil(t, err)
}

func TestValidator_documentTimestampForSigningValidator(t *testing.T) {
	av := documentTimestampForSigningValidator()

	err := av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	// fail
	model := NewDocumentMock(t)
	model.On("Timestamp").
		Return(time.Now(), errors.New("error")).
		Once()

	err = av.Validate(nil, model)
	assert.True(t, errors.IsOfType(ErrDocumentTimestampRetrieval, err))

	model = NewDocumentMock(t)
	model.On("Timestamp").
		Return(time.Now().UTC().Add(-MaxAuthoredToCommitDuration), nil).
		Once()

	err = av.Validate(nil, model)
	assert.ErrorIs(t, err, ErrDocumentTooOldToSign)

	// success
	model = NewDocumentMock(t)
	model.On("Timestamp").
		Return(time.Now().UTC(), nil).
		Once()

	err = av.Validate(nil, model)
	assert.Nil(t, err)
}

func TestValidator_signatureValidator(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)
	ssv := signaturesValidator(identityServiceMock)

	err := ssv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	authorPublicKey := utils.RandomSlice(32)
	authorSignature := utils.RandomSlice(64)
	authorTransitionValidated := true

	collaboratorPublicKey := utils.RandomSlice(32)
	collaboratorSignature := utils.RandomSlice(64)
	collaboratorTransitionValidated := false

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentAuthor.ToBytes(),
			PublicKey:           authorPublicKey,
			Signature:           authorSignature,
			TransitionValidated: authorTransitionValidated,
		},
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentCollaborator.ToBytes(),
			PublicKey:           collaboratorPublicKey,
			Signature:           collaboratorSignature,
			TransitionValidated: collaboratorTransitionValidated,
		},
	}

	documentMock.On("Signatures").
		Return(signatures).
		Once()
	documentMock.On("Author").
		Return(documentAuthor, nil).
		Once()

	collaborators := []*types.AccountID{documentAuthor, documentCollaborator}

	documentMock.On("GetSignerCollaborators", documentAuthor).
		Return(collaborators, nil).
		Once()

	timestamp := time.Now()

	documentMock.On("Timestamp").
		Return(timestamp, nil)

	identityServiceMock.On(
		"ValidateDocumentSignature",
		documentAuthor,
		authorPublicKey,
		ConsensusSignaturePayload(signingRoot, authorTransitionValidated),
		authorSignature,
		timestamp,
	).Return(nil).Once()

	identityServiceMock.On(
		"ValidateDocumentSignature",
		documentCollaborator,
		collaboratorPublicKey,
		ConsensusSignaturePayload(signingRoot, collaboratorTransitionValidated),
		collaboratorSignature,
		timestamp,
	).Return(nil).Once()

	err = ssv.Validate(nil, documentMock)
	assert.NoError(t, err)
}

func TestValidator_signatureValidator_SigningRootError(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)
	ssv := signaturesValidator(identityServiceMock)

	err := ssv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	documentMock.On("CalculateSigningRoot").
		Return(nil, errors.New("error")).
		Once()

	err = ssv.Validate(nil, documentMock)
	assert.True(t, errors.IsOfType(ErrDocumentCalculateSigningRoot, err))
}

func TestValidator_signatureValidator_SignaturesError(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)
	ssv := signaturesValidator(identityServiceMock)

	err := ssv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	documentMock.On("Signatures").
		Return(nil).
		Once()

	err = ssv.Validate(nil, documentMock)
	assert.ErrorIs(t, err, ErrDocumentNoSignatures)
}

func TestValidator_signatureValidator_AuthorError(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)
	ssv := signaturesValidator(identityServiceMock)

	err := ssv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	authorPublicKey := utils.RandomSlice(32)
	authorSignature := utils.RandomSlice(64)
	authorTransitionValidated := true

	collaboratorPublicKey := utils.RandomSlice(32)
	collaboratorSignature := utils.RandomSlice(64)
	collaboratorTransitionValidated := false

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentAuthor.ToBytes(),
			PublicKey:           authorPublicKey,
			Signature:           authorSignature,
			TransitionValidated: authorTransitionValidated,
		},
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentCollaborator.ToBytes(),
			PublicKey:           collaboratorPublicKey,
			Signature:           collaboratorSignature,
			TransitionValidated: collaboratorTransitionValidated,
		},
	}

	documentMock.On("Signatures").
		Return(signatures).
		Once()
	documentMock.On("Author").
		Return(nil, errors.New("error")).
		Once()

	err = ssv.Validate(nil, documentMock)
	assert.ErrorIs(t, err, ErrDocumentAuthorRetrieval)
}

func TestValidator_signatureValidator_SignerCollaboratorsError(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)
	ssv := signaturesValidator(identityServiceMock)

	err := ssv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	authorPublicKey := utils.RandomSlice(32)
	authorSignature := utils.RandomSlice(64)
	authorTransitionValidated := true

	collaboratorPublicKey := utils.RandomSlice(32)
	collaboratorSignature := utils.RandomSlice(64)
	collaboratorTransitionValidated := false

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentAuthor.ToBytes(),
			PublicKey:           authorPublicKey,
			Signature:           authorSignature,
			TransitionValidated: authorTransitionValidated,
		},
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentCollaborator.ToBytes(),
			PublicKey:           collaboratorPublicKey,
			Signature:           collaboratorSignature,
			TransitionValidated: collaboratorTransitionValidated,
		},
	}

	documentMock.On("Signatures").
		Return(signatures).
		Once()
	documentMock.On("Author").
		Return(documentAuthor, nil).
		Once()

	documentMock.On("GetSignerCollaborators", documentAuthor).
		Return(nil, errors.New("error")).
		Once()

	err = ssv.Validate(nil, documentMock)
	assert.ErrorIs(t, err, ErrDocumentSignerCollaboratorsRetrieval)
}

func TestValidator_signatureValidator_SignerAccountIDError(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)
	ssv := signaturesValidator(identityServiceMock)

	err := ssv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	authorPublicKey := utils.RandomSlice(32)
	authorSignature := utils.RandomSlice(64)
	authorTransitionValidated := true

	collaboratorPublicKey := utils.RandomSlice(32)
	collaboratorSignature := utils.RandomSlice(64)
	collaboratorTransitionValidated := false

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            []byte("invalid-account-id-bytes"),
			PublicKey:           authorPublicKey,
			Signature:           authorSignature,
			TransitionValidated: authorTransitionValidated,
		},
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentCollaborator.ToBytes(),
			PublicKey:           collaboratorPublicKey,
			Signature:           collaboratorSignature,
			TransitionValidated: collaboratorTransitionValidated,
		},
	}

	documentMock.On("Signatures").
		Return(signatures).
		Once()
	documentMock.On("Author").
		Return(documentAuthor, nil).
		Once()

	collaborators := []*types.AccountID{documentAuthor, documentCollaborator}

	documentMock.On("GetSignerCollaborators", documentAuthor).
		Return(collaborators, nil).
		Once()

	timestamp := time.Now()

	documentMock.On("Timestamp").
		Return(timestamp, nil)

	identityServiceMock.On(
		"ValidateDocumentSignature",
		documentCollaborator,
		collaboratorPublicKey,
		ConsensusSignaturePayload(signingRoot, collaboratorTransitionValidated),
		collaboratorSignature,
		timestamp,
	).Return(nil).Once()

	err = ssv.Validate(nil, documentMock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "couldn't parse signer account ID")
}

func TestValidator_signatureValidator_NotCollaboratorNorAuthor(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)
	ssv := signaturesValidator(identityServiceMock)

	err := ssv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaboratorPublicKey := utils.RandomSlice(32)
	collaboratorSignature := utils.RandomSlice(64)
	collaboratorTransitionValidated := false

	signatures := []*coredocumentpb.Signature{
		{
			// This signer is not part of the collaborators.
			SignatureId:         utils.RandomSlice(32),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: false,
		},
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentCollaborator.ToBytes(),
			PublicKey:           collaboratorPublicKey,
			Signature:           collaboratorSignature,
			TransitionValidated: collaboratorTransitionValidated,
		},
	}

	documentMock.On("Signatures").
		Return(signatures).
		Once()
	documentMock.On("Author").
		Return(documentAuthor, nil).
		Once()

	collaborators := []*types.AccountID{documentAuthor, documentCollaborator}

	documentMock.On("GetSignerCollaborators", documentAuthor).
		Return(collaborators, nil).
		Once()

	timestamp := time.Now()

	documentMock.On("Timestamp").
		Return(timestamp, nil)

	identityServiceMock.On(
		"ValidateDocumentSignature",
		documentCollaborator,
		collaboratorPublicKey,
		ConsensusSignaturePayload(signingRoot, collaboratorTransitionValidated),
		collaboratorSignature,
		timestamp,
	).Return(nil).Once()

	err = ssv.Validate(nil, documentMock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signer is not part of the signing collaborators")
}

func TestValidator_signatureValidator_DocumentTimestampError(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)
	ssv := signaturesValidator(identityServiceMock)

	err := ssv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	authorPublicKey := utils.RandomSlice(32)
	authorSignature := utils.RandomSlice(64)
	authorTransitionValidated := true

	collaboratorPublicKey := utils.RandomSlice(32)
	collaboratorSignature := utils.RandomSlice(64)
	collaboratorTransitionValidated := false

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentAuthor.ToBytes(),
			PublicKey:           authorPublicKey,
			Signature:           authorSignature,
			TransitionValidated: authorTransitionValidated,
		},
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentCollaborator.ToBytes(),
			PublicKey:           collaboratorPublicKey,
			Signature:           collaboratorSignature,
			TransitionValidated: collaboratorTransitionValidated,
		},
	}

	documentMock.On("Signatures").
		Return(signatures).
		Once()
	documentMock.On("Author").
		Return(documentAuthor, nil).
		Once()

	collaborators := []*types.AccountID{documentAuthor, documentCollaborator}

	documentMock.On("GetSignerCollaborators", documentAuthor).
		Return(collaborators, nil).
		Once()

	documentMock.On("Timestamp").
		Return(time.Now(), errors.New("error"))

	err = ssv.Validate(nil, documentMock)
	assert.NotNil(t, err)
}

func TestValidator_signatureValidator_ValidationError(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)
	ssv := signaturesValidator(identityServiceMock)

	err := ssv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	authorPublicKey := utils.RandomSlice(32)
	authorSignature := utils.RandomSlice(64)
	authorTransitionValidated := true

	collaboratorPublicKey := utils.RandomSlice(32)
	collaboratorSignature := utils.RandomSlice(64)
	collaboratorTransitionValidated := false

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentAuthor.ToBytes(),
			PublicKey:           authorPublicKey,
			Signature:           authorSignature,
			TransitionValidated: authorTransitionValidated,
		},
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentCollaborator.ToBytes(),
			PublicKey:           collaboratorPublicKey,
			Signature:           collaboratorSignature,
			TransitionValidated: collaboratorTransitionValidated,
		},
	}

	documentMock.On("Signatures").
		Return(signatures).
		Once()
	documentMock.On("Author").
		Return(documentAuthor, nil).
		Once()

	collaborators := []*types.AccountID{documentAuthor, documentCollaborator}

	documentMock.On("GetSignerCollaborators", documentAuthor).
		Return(collaborators, nil).
		Once()

	timestamp := time.Now()

	documentMock.On("Timestamp").
		Return(timestamp, nil)

	errMsg := "test signature invalid"

	identityServiceMock.On(
		"ValidateDocumentSignature",
		documentAuthor,
		authorPublicKey,
		ConsensusSignaturePayload(signingRoot, authorTransitionValidated),
		authorSignature,
		timestamp,
	).Return(errors.New(errMsg)).Once()

	identityServiceMock.On(
		"ValidateDocumentSignature",
		documentCollaborator,
		collaboratorPublicKey,
		ConsensusSignaturePayload(signingRoot, collaboratorTransitionValidated),
		collaboratorSignature,
		timestamp,
	).Return(nil).Once()

	err = ssv.Validate(nil, documentMock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), errMsg)
}

func TestValidator_signatureValidator_AuthorNotFound(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)
	ssv := signaturesValidator(identityServiceMock)

	err := ssv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil).
		Once()

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentCollaborator1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentCollaborator2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborator1PublicKey := utils.RandomSlice(32)
	collaborator1Signature := utils.RandomSlice(64)
	collaborator1TransitionValidated := false

	collaborator2PublicKey := utils.RandomSlice(32)
	collaborator2Signature := utils.RandomSlice(64)
	collaborator2TransitionValidated := true

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentCollaborator1.ToBytes(),
			PublicKey:           collaborator1PublicKey,
			Signature:           collaborator1Signature,
			TransitionValidated: collaborator1TransitionValidated,
		},
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            documentCollaborator2.ToBytes(),
			PublicKey:           collaborator2PublicKey,
			Signature:           collaborator2Signature,
			TransitionValidated: collaborator2TransitionValidated,
		},
	}

	documentMock.On("Signatures").
		Return(signatures).
		Once()
	documentMock.On("Author").
		Return(documentAuthor, nil).
		Once()

	collaborators := []*types.AccountID{documentCollaborator1, documentCollaborator2}

	documentMock.On("GetSignerCollaborators", documentAuthor).
		Return(collaborators, nil).
		Once()

	timestamp := time.Now()

	documentMock.On("Timestamp").
		Return(timestamp, nil)

	identityServiceMock.On(
		"ValidateDocumentSignature",
		documentCollaborator1,
		collaborator1PublicKey,
		ConsensusSignaturePayload(signingRoot, collaborator1TransitionValidated),
		collaborator1Signature,
		timestamp,
	).Return(nil).Once()

	identityServiceMock.On(
		"ValidateDocumentSignature",
		documentCollaborator2,
		collaborator2PublicKey,
		ConsensusSignaturePayload(signingRoot, collaborator2TransitionValidated),
		collaborator2Signature,
		timestamp,
	).Return(nil).Once()

	err = ssv.Validate(nil, documentMock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "author's signature missing on document")
}

func TestValidator_anchoredValidator(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	av := anchoredValidator(anchorServiceMock)

	err := av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	currentVersion := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock.On("CurrentVersion").
		Return(currentVersion).
		Once()
	documentMock.On("CalculateDocumentRoot").
		Return(documentRoot, nil).
		Once()

	anchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	docRoot, err := anchors.ToDocumentRoot(documentRoot)
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(docRoot, time.Now(), nil)

	timestamp := time.Now()

	documentMock.On("Timestamp").
		Return(timestamp, nil).
		Once()

	err = av.Validate(nil, documentMock)
	assert.NoError(t, err)
}

func TestValidator_anchoredValidator_ToAnchorIDError(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	av := anchoredValidator(anchorServiceMock)

	err := av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	currentVersion := utils.RandomSlice(anchors.AnchorIDLength - 1)

	documentMock.On("CurrentVersion").
		Return(currentVersion).
		Once()

	err = av.Validate(nil, documentMock)
	assert.True(t, errors.IsOfType(ErrAnchorIDCreation, err))
}

func TestValidator_anchoredValidator_CalculateDocumentRootError(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	av := anchoredValidator(anchorServiceMock)

	err := av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	currentVersion := utils.RandomSlice(32)

	documentMock.On("CurrentVersion").
		Return(currentVersion).
		Once()
	documentMock.On("CalculateDocumentRoot").
		Return(nil, errors.New("error")).
		Once()

	err = av.Validate(nil, documentMock)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrDocumentCalculateDocumentRoot, err))
}

func TestValidator_anchoredValidator_DocumentRootParseError(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	av := anchoredValidator(anchorServiceMock)

	err := av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	currentVersion := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(anchors.AnchorIDLength - 1)

	documentMock.On("CurrentVersion").
		Return(currentVersion).
		Once()
	documentMock.On("CalculateDocumentRoot").
		Return(documentRoot, nil).
		Once()

	err = av.Validate(nil, documentMock)
	assert.True(t, errors.IsOfType(ErrDocumentRootCreation, err))
}

func TestValidator_anchoredValidator_AnchorServiceError(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	av := anchoredValidator(anchorServiceMock)

	err := av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	currentVersion := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock.On("CurrentVersion").
		Return(currentVersion).
		Once()
	documentMock.On("CalculateDocumentRoot").
		Return(documentRoot, nil).
		Once()

	anchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(nil, time.Now(), errors.New("error"))

	err = av.Validate(nil, documentMock)
	assert.True(t, errors.IsOfType(ErrDocumentAnchorDataRetrieval, err))
}

func TestValidator_anchoredValidator_DocumentRootMismatch(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	av := anchoredValidator(anchorServiceMock)

	err := av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	currentVersion := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock.On("CurrentVersion").
		Return(currentVersion).
		Once()
	documentMock.On("CalculateDocumentRoot").
		Return(documentRoot, nil).
		Once()

	anchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	randomDocRoot, err := anchors.ToDocumentRoot(utils.RandomSlice(32))
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(randomDocRoot, time.Now(), nil)

	err = av.Validate(nil, documentMock)
	assert.True(t, errors.IsOfType(ErrDocumentRootsMismatch, err))
}

func TestValidator_anchoredValidator_DocumentTimestampError(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	av := anchoredValidator(anchorServiceMock)

	err := av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	currentVersion := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock.On("CurrentVersion").
		Return(currentVersion).
		Once()
	documentMock.On("CalculateDocumentRoot").
		Return(documentRoot, nil).
		Once()

	anchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	docRoot, err := anchors.ToDocumentRoot(documentRoot)
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(docRoot, time.Now(), nil)

	timestamp := time.Now()

	documentMock.On("Timestamp").
		Return(timestamp, errors.New("error")).
		Once()

	err = av.Validate(nil, documentMock)
	assert.True(t, errors.IsOfType(ErrDocumentTimestampRetrieval, err))
}

func TestValidator_anchoredValidator_InvalidAnchorTime(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	av := anchoredValidator(anchorServiceMock)

	err := av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	documentMock := NewDocumentMock(t)

	currentVersion := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock.On("CurrentVersion").
		Return(currentVersion).
		Once()
	documentMock.On("CalculateDocumentRoot").
		Return(documentRoot, nil).
		Once()

	anchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	docRoot, err := anchors.ToDocumentRoot(documentRoot)
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(docRoot, time.Now(), nil)

	documentMock.On("Timestamp").
		Return(time.Now().Add(-2*MaxAuthoredToCommitDuration), nil).
		Once()

	err = av.Validate(nil, documentMock)
	assert.True(t, errors.IsOfType(ErrDocumentInvalidAnchorTime, err))
}

func TestValidator_versionNotAnchoredValidator(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	// Success
	id := utils.RandomSlice(anchors.AnchorIDLength)

	anchorID, err := anchors.ToAnchorID(id)
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(nil, time.Now(), errors.New("error")).
		Once()

	err = versionNotAnchoredValidator(anchorServiceMock, id)
	assert.NoError(t, err)

	// Invalid ID length
	id = utils.RandomSlice(anchors.AnchorIDLength - 1)

	err = versionNotAnchoredValidator(anchorServiceMock, id)
	assert.True(t, errors.IsOfType(ErrAnchorIDCreation, err))

	// Anchor present
	id = utils.RandomSlice(anchors.AnchorIDLength)

	anchorID, err = anchors.ToAnchorID(id)
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(nil, time.Now(), nil).
		Once()

	err = versionNotAnchoredValidator(anchorServiceMock, id)
	assert.ErrorIs(t, err, ErrDocumentIDReused)
}

func TestValidator_LatestVersionValidator(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	lv := LatestVersionValidator(anchorServiceMock)

	err := lv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	nextVersion := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentMock.On("NextVersion").
		Return(nextVersion).
		Times(2)

	anchorID, err := anchors.ToAnchorID(nextVersion)
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(nil, time.Now(), errors.New("error")).
		Once()

	err = lv.Validate(nil, documentMock)
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(nil, time.Now(), nil).
		Once()

	err = lv.Validate(nil, documentMock)
	assert.True(t, errors.IsOfType(ErrDocumentNotLatest, err))
}

func TestValidator_currentVersionValidator(t *testing.T) {
	anchorServiceMock := anchors.NewAPIMock(t)

	cv := currentVersionValidator(anchorServiceMock)

	err := cv.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	currentVersion := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentMock.On("CurrentVersion").
		Return(currentVersion).
		Times(2)

	anchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(nil, time.Now(), errors.New("error")).
		Once()

	err = cv.Validate(nil, documentMock)
	assert.NoError(t, err)

	anchorServiceMock.On("GetAnchorData", anchorID).
		Return(nil, time.Now(), nil).
		Once()

	err = cv.Validate(nil, documentMock)
	assert.True(t, errors.IsOfType(ErrDocumentNotLatest, err))
}

func TestValidator_attributeValidator(t *testing.T) {
	identityServiceMock := v2.NewServiceMock(t)

	av := attributeValidator(identityServiceMock)

	err := av.Validate(nil, nil)
	assert.ErrorIs(t, err, ErrModelNil)

	signerAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)
	value := []byte("value")
	signature := utils.RandomSlice(64)
	publicKey := utils.RandomSlice(32)

	attributes := []Attribute{
		{
			KeyLabel: "label1",
			Key:      utils.RandomByte32(),
			Value: AttrVal{
				Type: AttrString,
				Str:  "string",
			},
		},
		{
			KeyLabel: "label2",
			Key:      utils.RandomByte32(),
			Value: AttrVal{
				Type: AttrSigned,
				Signed: Signed{
					Identity:        signerAccountID,
					Type:            AttrString,
					DocumentVersion: documentVersion,
					Value:           value,
					Signature:       signature,
					PublicKey:       publicKey,
				},
			},
		},
	}

	documentMock := NewDocumentMock(t)

	documentMock.On("ID").
		Return(documentID)
	documentMock.On("GetAttributes").
		Return(attributes)

	timestamp := time.Now()

	documentMock.On("Timestamp").
		Return(timestamp, nil).
		Once()

	payload := attributeSignaturePayload(signerAccountID.ToBytes(), documentID, documentVersion, value)

	identityServiceMock.On(
		"ValidateDocumentSignature",
		signerAccountID,
		publicKey,
		payload,
		signature,
		timestamp,
	).Return(nil).Once()

	err = av.Validate(nil, documentMock)
	assert.NoError(t, err)

	documentMock.On("Timestamp").
		Return(timestamp, errors.New("error")).
		Once()

	err = av.Validate(nil, documentMock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "couldn't get model timestamp")

	documentMock.On("Timestamp").
		Return(timestamp, nil).
		Once()

	identityServiceMock.On(
		"ValidateDocumentSignature",
		signerAccountID,
		publicKey,
		payload,
		signature,
		timestamp,
	).Return(errors.New("error")).Once()

	err = av.Validate(nil, documentMock)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate signature for attribute")
}

func TestValidator_transitionValidator(t *testing.T) {
	id1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	updated := NewDocumentMock(t)

	// does not error out if there is no old document model (if new model is the first version of the document model)
	tv := transitionValidator(id1)
	err = tv.Validate(nil, updated)
	assert.NoError(t, err)

	old := NewDocumentMock(t)
	old.On("CollaboratorCanUpdate", updated, id1).
		Return(errors.New("error")).
		Once()

	err = tv.Validate(old, updated)
	assert.True(t, errors.IsOfType(ErrInvalidDocumentStateTransition, err))

	old.On("CollaboratorCanUpdate", updated, id1).
		Return(nil).
		Once()

	err = tv.Validate(old, updated)
	assert.NoError(t, err)
}

func TestValidator_computeFieldsValidator(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	// create a compute field rule
	wasmPath := path.AppendPathToProjectRoot("testingutils/compute_fields/simple_average.wasm")
	wasm := wasmLoader(t, wasmPath)

	rule, err := cd.AddComputeFieldsRule(wasm, []string{"test", "test2", "test3"}, "result")
	assert.NoError(t, err)

	// add required attributes
	cd, err = cd.AddAttributes(CollaboratorsAccess{}, false, nil, getValidComputeFieldAttrs(t)...)
	assert.NoError(t, err)

	// failed to set target
	oldKey := rule.ComputeTargetField
	rule.ComputeTargetField = nil

	doc := NewDocumentMock(t)
	doc.On("GetComputeFieldsRules").
		Return(cd.GetComputeFieldsRules()).
		Once()

	doc.On("GetAttributes").
		Return(cd.GetAttributes()).
		Twice()

	validator := computeFieldsValidator(10 * time.Second)

	err = validator.Validate(nil, doc)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrEmptyAttrLabel, err))

	// wrong target result
	rule.ComputeTargetField = oldKey

	doc.On("GetComputeFieldsRules").
		Return(cd.GetComputeFieldsRules()).
		Twice()

	err = validator.Validate(nil, doc)
	assert.EqualError(t, err, fmt.Sprintf("compute fields[%s] validation failed", hexutil.Encode(rule.RuleKey)))

	// successful validation
	targetKey, err := AttrKeyFromLabel("result")
	assert.NoError(t, err)

	cd, err = cd.AddAttributes(CollaboratorsAccess{}, false, nil, Attribute{
		KeyLabel: "result",
		Key:      targetKey,
		Value: AttrVal{
			Type:  AttrBytes,
			Bytes: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x7, 0xd0},
		},
	})
	assert.NoError(t, err)

	doc.On("GetAttributes").
		Return(cd.GetAttributes()).
		Once()

	err = validator.Validate(nil, doc)
	assert.NoError(t, err)
}

func Test_CreateVersionValidator(t *testing.T) {
	res := CreateVersionValidator(nil)

	vg, ok := res.(ValidatorGroup)
	assert.True(t, ok)
	assert.Len(t, vg, 3)

	assert.Equal(t, reflect.ValueOf(baseValidator()).Pointer(), reflect.ValueOf(vg[0]).Pointer())
	assert.Equal(t, reflect.ValueOf(currentVersionValidator(nil)).Pointer(), reflect.ValueOf(vg[1]).Pointer())
	assert.Equal(t, reflect.ValueOf(LatestVersionValidator(nil)).Pointer(), reflect.ValueOf(vg[2]).Pointer())
}

func Test_UpdateVersionValidator(t *testing.T) {
	res := UpdateVersionValidator(nil)

	vg, ok := res.(ValidatorGroup)
	assert.True(t, ok)
	assert.Len(t, vg, 3)

	assert.Equal(t, reflect.ValueOf(versionIDsValidator()).Pointer(), reflect.ValueOf(vg[0]).Pointer())
	assert.Equal(t, reflect.ValueOf(currentVersionValidator(nil)).Pointer(), reflect.ValueOf(vg[1]).Pointer())
	assert.Equal(t, reflect.ValueOf(LatestVersionValidator(nil)).Pointer(), reflect.ValueOf(vg[2]).Pointer())
}

func Test_PreAnchorValidator(t *testing.T) {
	res := PreAnchorValidator(nil)

	vg, ok := res.(ValidatorGroup)
	assert.True(t, ok)

	assertPreAnchorValidator(t, vg)
}

func Test_PostAnchoredValidator(t *testing.T) {
	res := PostAnchoredValidator(nil, nil)

	vg, ok := res.(ValidatorGroup)
	assert.True(t, ok)
	assertPostAnchorValidator(t, vg)
}

func Test_ReceivedAnchoredDocumentValidator(t *testing.T) {
	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	res := ReceivedAnchoredDocumentValidator(nil, nil, collaborator)

	vg, ok := res.(ValidatorGroup)
	assert.True(t, ok)
	assert.Len(t, vg, 2)

	postAnchoredValidator, ok := vg[1].(ValidatorGroup)
	assert.True(t, ok)
	assertPostAnchorValidator(t, postAnchoredValidator)

	assert.Equal(t, reflect.ValueOf(transitionValidator(collaborator)).Pointer(), reflect.ValueOf(vg[0]).Pointer())
}

func Test_RequestDocumentSignatureValidator(t *testing.T) {
	collaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	res := RequestDocumentSignatureValidator(nil, nil, collaborator)

	vg, ok := res.(ValidatorGroup)
	assert.True(t, ok)
	assert.Len(t, vg, 6)

	signatureValidator, ok := vg[5].(ValidatorGroup)
	assert.True(t, ok)
	assertSignatureValidators(t, signatureValidator)

	assert.Equal(t, reflect.ValueOf(documentTimestampForSigningValidator()).Pointer(), reflect.ValueOf(vg[0]).Pointer())
	assert.Equal(t, reflect.ValueOf(documentAuthorValidator(collaborator)).Pointer(), reflect.ValueOf(vg[1]).Pointer())
	assert.Equal(t, reflect.ValueOf(currentVersionValidator(nil)).Pointer(), reflect.ValueOf(vg[2]).Pointer())
	assert.Equal(t, reflect.ValueOf(LatestVersionValidator(nil)).Pointer(), reflect.ValueOf(vg[3]).Pointer())
	assert.Equal(t, reflect.ValueOf(transitionValidator(collaborator)).Pointer(), reflect.ValueOf(vg[4]).Pointer())
}

func Test_SignatureValidator(t *testing.T) {
	res := SignatureValidator(nil)

	vg, ok := res.(ValidatorGroup)
	assert.True(t, ok)

	assertSignatureValidators(t, vg)
}

func assertPreAnchorValidator(t *testing.T, vg ValidatorGroup) {
	assert.Len(t, vg, 2)

	signatureValidator, ok := vg[0].(ValidatorGroup)
	assert.True(t, ok)
	assertSignatureValidators(t, signatureValidator)

	assert.Equal(t, reflect.ValueOf(documentRootValidator()).Pointer(), reflect.ValueOf(vg[1]).Pointer())
}

func assertPostAnchorValidator(t *testing.T, vg ValidatorGroup) {
	assert.Len(t, vg, 3)

	preAnchorValidator, ok := vg[0].(ValidatorGroup)
	assert.True(t, ok)
	assertPreAnchorValidator(t, preAnchorValidator)

	assert.Equal(t, reflect.ValueOf(anchoredValidator(nil)).Pointer(), reflect.ValueOf(vg[1]).Pointer())
	assert.Equal(t, reflect.ValueOf(LatestVersionValidator(nil)).Pointer(), reflect.ValueOf(vg[2]).Pointer())
}

func assertSignatureValidators(t *testing.T, vg ValidatorGroup) {
	assert.Len(t, vg, 5)

	assert.Equal(t, reflect.ValueOf(baseValidator()).Pointer(), reflect.ValueOf(vg[0]).Pointer())
	assert.Equal(t, reflect.ValueOf(signingRootValidator()).Pointer(), reflect.ValueOf(vg[1]).Pointer())
	assert.Equal(t, reflect.ValueOf(signaturesValidator(nil)).Pointer(), reflect.ValueOf(vg[2]).Pointer())
	assert.Equal(t, reflect.ValueOf(attributeValidator(nil)).Pointer(), reflect.ValueOf(vg[3]).Pointer())
	assert.Equal(t, reflect.ValueOf(computeFieldsValidator(computeFieldsTimeout)).Pointer(), reflect.ValueOf(vg[4]).Pointer())
}

func TestValidatorGroup_Validate(t *testing.T) {
	tests := []struct {
		validator1Error    error
		validator2Error    error
		validator3Error    error
		expectedErrorCount int
	}{
		{
			validator1Error:    nil,
			validator2Error:    nil,
			validator3Error:    nil,
			expectedErrorCount: 0,
		},
		{
			validator1Error:    errors.New("error1"),
			validator2Error:    nil,
			validator3Error:    nil,
			expectedErrorCount: 1,
		},
		{
			validator1Error:    nil,
			validator2Error:    errors.New("error2"),
			validator3Error:    nil,
			expectedErrorCount: 1,
		},
		{
			validator1Error:    nil,
			validator2Error:    nil,
			validator3Error:    errors.New("error3"),
			expectedErrorCount: 1,
		},
		{
			validator1Error:    errors.New("error1"),
			validator2Error:    errors.New("error2"),
			validator3Error:    nil,
			expectedErrorCount: 2,
		},
		{
			validator1Error:    errors.New("error1"),
			validator2Error:    nil,
			validator3Error:    errors.New("error3"),
			expectedErrorCount: 2,
		},
		{
			validator1Error:    nil,
			validator2Error:    errors.New("error2"),
			validator3Error:    errors.New("error3"),
			expectedErrorCount: 2,
		},
		{
			validator1Error:    errors.New("error1"),
			validator2Error:    errors.New("error2"),
			validator3Error:    errors.New("error3"),
			expectedErrorCount: 3,
		},
	}

	validatorMock1 := NewValidatorMock(t)
	validatorMock2 := NewValidatorMock(t)
	validatorMock3 := NewValidatorMock(t)

	vg := ValidatorGroup{
		validatorMock1,
		validatorMock2,
		validatorMock3,
	}

	oldDocMock := NewDocumentMock(t)
	newDocMock := NewDocumentMock(t)

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			validatorMock1.On("Validate", oldDocMock, newDocMock).
				Return(test.validator1Error).
				Once()

			validatorMock2.On("Validate", oldDocMock, newDocMock).
				Return(test.validator2Error).
				Once()

			validatorMock3.On("Validate", oldDocMock, newDocMock).
				Return(test.validator3Error).
				Once()

			err := vg.Validate(oldDocMock, newDocMock)
			assert.Equal(t, errors.Len(err), test.expectedErrorCount)
		})
	}
}
