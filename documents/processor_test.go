//go:build unit

package documents

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/contextutil"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAnchorProcessor_Send(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	p2pConnectionTimeout := 1 * time.Second

	configMock.On("GetP2PConnectionTimeout").Return(p2pConnectionTimeout)

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	sendCtx, cancel := context.WithTimeout(ctx, p2pConnectionTimeout)
	defer cancel()

	testDoc := &coredocumentpb.CoreDocument{}

	p2pRes := &p2ppb.AnchorDocumentResponse{
		Accepted: true,
	}

	p2pClientMock.On(
		"SendAnchoredDocument",
		mock.IsType(sendCtx),
		accountID,
		&p2ppb.AnchorDocumentRequest{Document: testDoc},
	).Return(p2pRes, nil)

	err = ap.Send(ctx, testDoc, accountID)
	assert.NoError(t, err)
}

func TestAnchorProcessor_Send_Error(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	p2pConnectionTimeout := 1 * time.Second

	configMock.On("GetP2PConnectionTimeout").Return(p2pConnectionTimeout)

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	sendCtx, cancel := context.WithTimeout(ctx, p2pConnectionTimeout)
	defer cancel()

	testDoc := &coredocumentpb.CoreDocument{}

	p2pClientMock.On(
		"SendAnchoredDocument",
		mock.IsType(sendCtx),
		accountID,
		&p2ppb.AnchorDocumentRequest{Document: testDoc},
	).Return(nil, errors.New("error"))

	err = ap.Send(ctx, testDoc, accountID)
	assert.ErrorIs(t, err, ErrP2PDocumentSend)
}

func TestAnchorProcessor_PrepareForSignatureRequests(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("AddUpdateLog", accountID).
		Return(nil)
	documentMock.On("ExecuteComputeFields", computeFieldsTimeout).
		Return(nil)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil)

	signaturePayload := ConsensusSignaturePayload(signingRoot, false)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", signaturePayload).
		Return(signature, nil)

	documentMock.On("AppendSignatures", signature).Once()

	err = ap.PrepareForSignatureRequests(ctx, documentMock)
	assert.NoError(t, err)
}

func TestAnchorProcessor_PrepareForSignatureRequests_ContextAccountError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	documentMock := NewDocumentMock(t)

	err := ap.PrepareForSignatureRequests(context.Background(), documentMock)
	assert.ErrorIs(t, err, errors.ErrContextAccountRetrieval)
}

func TestAnchorProcessor_PrepareForSignatureRequests_AddUpdateLogError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)

	docError := errors.New("error")

	documentMock.On("AddUpdateLog", accountID).
		Return(docError)

	err = ap.PrepareForSignatureRequests(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentAddUpdateLog)
}

func TestAnchorProcessor_PrepareForSignatureRequests_ExecuteComputeFieldsError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("AddUpdateLog", accountID).
		Return(nil)

	docErr := errors.New("error")

	documentMock.On("ExecuteComputeFields", computeFieldsTimeout).
		Return(docErr)

	err = ap.PrepareForSignatureRequests(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentExecuteComputeFields)
}

func TestAnchorProcessor_PrepareForSignatureRequests_CalculateSigningRootError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("AddUpdateLog", accountID).
		Return(nil)
	documentMock.On("ExecuteComputeFields", computeFieldsTimeout).
		Return(nil)

	documentMock.On("CalculateSigningRoot").
		Return(nil, errors.New("error"))

	err = ap.PrepareForSignatureRequests(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentCalculateSigningRoot)
}

func TestAnchorProcessor_PrepareForSignatureRequests_SignMessageError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("AddUpdateLog", accountID).
		Return(nil)
	documentMock.On("ExecuteComputeFields", computeFieldsTimeout).
		Return(nil)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").
		Return(signingRoot, nil)

	signaturePayload := ConsensusSignaturePayload(signingRoot, false)

	signError := errors.New("error")

	accountMock.On("SignMsg", signaturePayload).
		Return(nil, signError)

	err = ap.PrepareForSignatureRequests(ctx, documentMock)
	assert.ErrorIs(t, err, ErrAccountSignMessage)
}

func TestAnchorProcessor_RequestSignatures(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)

	signatures := getTestSignatures(author, collaborators)

	documentMock.On("ID").Return(documentID)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)
	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("Signatures").Return(signatures)
	documentMock.On("Author").Return(author, nil)
	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)
	documentMock.On("GetAttributes").Return(nil)
	documentMock.On("GetComputeFieldsRules").Return(nil)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	p2pClientMock.On("GetSignaturesForDocument", ctx, documentMock).
		Return(signatures, nil, nil)

	documentMock.On("AppendSignatures", signatures[0], signatures[1], signatures[2])

	err = ap.RequestSignatures(ctx, documentMock)
	assert.NoError(t, err)
}

func TestAnchorProcessor_RequestSignatures_ValidationError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)

	signatures := getTestSignatures(author, collaborators)

	documentMock.On("ID").Return(documentID)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)
	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("Signatures").Return(signatures)
	documentMock.On("Author").Return(author, nil)
	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)
	documentMock.On("GetAttributes").Return(nil)
	documentMock.On("GetComputeFieldsRules").Return(nil)

	validateSignatureError := errors.New("error")

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(validateSignatureError)

	err = ap.RequestSignatures(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentValidation)
}

func TestAnchorProcessor_RequestSignatures_GetSignaturesForDocumentError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)

	signatures := getTestSignatures(author, collaborators)

	documentMock.On("ID").Return(documentID)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)
	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("Signatures").Return(signatures)
	documentMock.On("Author").Return(author, nil)
	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)
	documentMock.On("GetAttributes").Return(nil)
	documentMock.On("GetComputeFieldsRules").Return(nil)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	p2pClientError := errors.New("error")

	p2pClientMock.On("GetSignaturesForDocument", ctx, documentMock).
		Return(nil, nil, p2pClientError)

	err = ap.RequestSignatures(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentSignaturesRetrieval)
}

func TestAnchorProcessor_PrepareForAnchoring(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)

	signatures := getTestSignatures(author, collaborators)

	documentMock.On("ID").Return(documentID)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)
	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("Signatures").Return(signatures)
	documentMock.On("Author").Return(author, nil)
	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)
	documentMock.On("GetAttributes").Return(nil)
	documentMock.On("GetComputeFieldsRules").Return(nil)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	err = ap.PrepareForAnchoring(ctx, documentMock)
	assert.NoError(t, err)
}

func TestAnchorProcessor_PrepareForAnchoring_ValidationError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)

	signatures := getTestSignatures(author, collaborators)

	documentMock.On("ID").Return(documentID)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)
	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("Signatures").Return(signatures)
	documentMock.On("Author").Return(author, nil)
	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)
	documentMock.On("GetAttributes").Return(nil)
	documentMock.On("GetComputeFieldsRules").Return(nil)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(errors.New("error"))

	err = ap.PrepareForAnchoring(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentValidation)
}

func TestAnchorProcessor_PreAnchorDocument(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)

	documentMock.On("ID").Return(documentID, nil)
	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)

	anchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	docRoot, err := anchors.ToDocumentRoot(signingRoot)
	assert.NoError(t, err)

	anchorServiceMock.On("PreCommitAnchor", ctx, anchorID, docRoot).
		Return(nil)

	err = ap.PreAnchorDocument(ctx, documentMock)
	assert.NoError(t, err)
}

func TestAnchorProcessor_PreAnchorDocument_CalculateSigningRootError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	documentMock.On("CalculateSigningRoot").
		Return(nil, errors.New("error"))

	err := ap.PreAnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentCalculateSigningRoot)
}

func TestAnchorProcessor_PreAnchorDocument_ToAnchorIDError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	// Invalid length for the current version byte slice.
	currentVersion := utils.RandomSlice(16)

	signingRoot := utils.RandomSlice(32)

	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("CurrentVersion").Return(currentVersion)

	err := ap.PreAnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrAnchorIDCreation)
}

func TestAnchorProcessor_PreAnchorDocument_ToDocumentRootError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	currentVersion := utils.RandomSlice(32)

	// Invalid length for the signing root byte slice.
	signingRoot := utils.RandomSlice(16)

	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("CurrentVersion").Return(currentVersion)

	err := ap.PreAnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentRootCreation)
}

func TestAnchorProcessor_PreAnchorDocument_PreCommitError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)

	documentMock.On("ID").Return(documentID, nil)
	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)

	anchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	docRoot, err := anchors.ToDocumentRoot(signingRoot)
	assert.NoError(t, err)

	anchorServiceMock.On("PreCommitAnchor", ctx, anchorID, docRoot).
		Return(errors.New("error"))

	err = ap.PreAnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrPreCommitAnchor)
}

func TestAnchorProcessor_AnchorDocument(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	currentVersionPreImage := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)
	signaturesRoot := utils.RandomSlice(32)
	signatures := getTestSignatures(author, collaborators)

	mockDocumentPreAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		signatures,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	documentMock.On("CalculateDocumentRoot").Return(documentRoot, nil)
	documentMock.On("CurrentVersionPreimage").Return(currentVersionPreImage)
	documentMock.On("CalculateSignaturesRoot").Return(signaturesRoot, nil)

	anchorIDPreimage, err := anchors.ToAnchorID(currentVersionPreImage)
	assert.NoError(t, err)

	rootHash, err := anchors.ToDocumentRoot(documentRoot)
	assert.NoError(t, err)

	signaturesRootHash, err := utils.SliceToByte32(signaturesRoot)
	assert.NoError(t, err)

	anchorServiceMock.On("CommitAnchor", ctx, anchorIDPreimage, rootHash, signaturesRootHash).
		Return(nil)

	err = ap.AnchorDocument(ctx, documentMock)
	assert.NoError(t, err)
}

func TestAnchorProcessor_AnchorDocument_ValidationError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)
	signatures := getTestSignatures(author, collaborators)

	mockDocumentPreAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		signatures,
	)

	documentMock.On("CalculateDocumentRoot").Return(documentRoot, nil)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(errors.New("error"))

	err = ap.AnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentValidation)
}

func TestAnchorProcessor_AnchorDocument_CalculateDocumentRootError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)
	signatures := getTestSignatures(author, collaborators)

	mockDocumentPreAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		signatures,
	)
	documentMock.On("CalculateDocumentRoot").
		Once().
		Return(documentRoot, nil)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	documentMock.On("CalculateDocumentRoot").
		Once().
		Return(nil, errors.New("error"))

	err = ap.AnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentCalculateDocumentRoot)
}

func TestAnchorProcessor_AnchorDocument_ToDocumentRootError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)

	documentRoot := utils.RandomSlice(32)

	signatures := getTestSignatures(author, collaborators)

	mockDocumentPreAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		signatures,
	)

	documentMock.On("CalculateDocumentRoot").
		Once().
		Return(documentRoot, nil)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	// Invalid length for document root byte slice.
	invalidDocumentRoot := utils.RandomSlice(16)

	documentMock.On("CalculateDocumentRoot").
		Once().
		Return(invalidDocumentRoot, nil)

	err = ap.AnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentRootCreation)
}

func TestAnchorProcessor_AnchorDocument_ToAnchorIDError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	// Invalid length for current version pre-image byte slice
	currentVersionPreImage := utils.RandomSlice(16)

	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)
	signatures := getTestSignatures(author, collaborators)

	mockDocumentPreAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		signatures,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	documentMock.On("CalculateDocumentRoot").Return(documentRoot, nil)
	documentMock.On("CurrentVersionPreimage").Return(currentVersionPreImage)

	err = ap.AnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrAnchorIDCreation)
}

func TestAnchorProcessor_AnchorDocument_CalculateSignaturesRootError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	currentVersionPreImage := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)
	signatures := getTestSignatures(author, collaborators)

	mockDocumentPreAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		signatures,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	documentMock.On("CalculateDocumentRoot").Return(documentRoot, nil)
	documentMock.On("CurrentVersionPreimage").Return(currentVersionPreImage)

	documentMock.On("CalculateSignaturesRoot").Return(nil, errors.New("error"))

	err = ap.AnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentCalculateSignaturesRoot)
}

func TestAnchorProcessor_AnchorDocument_SignatureRootConversionError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	currentVersionPreImage := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	// Invalid length for signatures root byte slice
	signaturesRoot := utils.RandomSlice(33)

	signatures := getTestSignatures(author, collaborators)

	mockDocumentPreAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		signatures,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	documentMock.On("CalculateDocumentRoot").Return(documentRoot, nil)
	documentMock.On("CurrentVersionPreimage").Return(currentVersionPreImage)
	documentMock.On("CalculateSignaturesRoot").Return(signaturesRoot, nil)

	err = ap.AnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrSignaturesRootProofConversion)
}

func TestAnchorProcessor_AnchorDocument_CommitAnchorError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	currentVersionPreImage := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)
	signaturesRoot := utils.RandomSlice(32)
	signatures := getTestSignatures(author, collaborators)

	mockDocumentPreAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		signatures,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	documentMock.On("CalculateDocumentRoot").Return(documentRoot, nil)
	documentMock.On("CurrentVersionPreimage").Return(currentVersionPreImage)
	documentMock.On("CalculateSignaturesRoot").Return(signaturesRoot, nil)

	anchorIDPreimage, err := anchors.ToAnchorID(currentVersionPreImage)
	assert.NoError(t, err)

	rootHash, err := anchors.ToDocumentRoot(documentRoot)
	assert.NoError(t, err)

	signaturesRootHash, err := utils.SliceToByte32(signaturesRoot)
	assert.NoError(t, err)

	anchorServiceMock.On("CommitAnchor", ctx, anchorIDPreimage, rootHash, signaturesRootHash).
		Return(errors.New("error"))

	err = ap.AnchorDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrCommitAnchor)
}

func TestAnchorProcessor_SendDocument(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(author)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentPostAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		documentRoot,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	currentVersionAnchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	nextVersionAnchorID, err := anchors.ToAnchorID(nextVersion)
	assert.NoError(t, err)

	anchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorServiceMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorServiceMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)

	coreDocument := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(coreDocument, nil)

	p2pConnectionTimeout := 1 * time.Second

	configMock.On("GetP2PConnectionTimeout").Return(p2pConnectionTimeout)

	anchorDocumentRes := &p2ppb.AnchorDocumentResponse{
		Accepted: true,
	}

	p2pClientMock.On(
		"SendAnchoredDocument",
		mock.Anything,
		collaborators[0],
		&p2ppb.AnchorDocumentRequest{Document: coreDocument},
	).Return(anchorDocumentRes, nil)

	p2pClientMock.On(
		"SendAnchoredDocument",
		mock.Anything,
		collaborators[1],
		&p2ppb.AnchorDocumentRequest{Document: coreDocument},
	).Return(anchorDocumentRes, nil)

	err = ap.SendDocument(ctx, documentMock)
	assert.NoError(t, err)
}

func TestAnchorProcessor_SendDocument_ValidationError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentPostAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		documentRoot,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(errors.New("error"))

	currentVersionAnchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	nextVersionAnchorID, err := anchors.ToAnchorID(nextVersion)
	assert.NoError(t, err)

	anchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorServiceMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorServiceMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	err = ap.SendDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentValidation)
}

func TestAnchorProcessor_SendDocument_ContextIdentityError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentPostAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		documentRoot,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	currentVersionAnchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	nextVersionAnchorID, err := anchors.ToAnchorID(nextVersion)
	assert.NoError(t, err)

	anchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorServiceMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorServiceMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	err = ap.SendDocument(ctx, documentMock)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
}

func TestAnchorProcessor_SendDocument_SignerCollaboratorsError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(author)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentPostAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		documentRoot,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	currentVersionAnchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	nextVersionAnchorID, err := anchors.ToAnchorID(nextVersion)
	assert.NoError(t, err)

	anchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorServiceMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorServiceMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	documentMock.On("GetSignerCollaborators", author).Once().Return(nil, errors.New("error"))

	err = ap.SendDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentCollaboratorsRetrieval)
}

func TestAnchorProcessor_SendDocument_PackCoreDocumentError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(author)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentPostAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		documentRoot,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	currentVersionAnchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	nextVersionAnchorID, err := anchors.ToAnchorID(nextVersion)
	assert.NoError(t, err)

	anchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorServiceMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorServiceMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)

	documentMock.On("PackCoreDocument").
		Return(nil, errors.New("error"))

	err = ap.SendDocument(ctx, documentMock)
	assert.ErrorIs(t, err, ErrDocumentPackingCoreDocument)
}

func TestAnchorProcessor_SendDocument_SendError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	documentMock := NewDocumentMock(t)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(author)

	ctx = contextutil.WithAccount(ctx, accountMock)

	collaborators, err := getTestCollaborators(2)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentPostAnchoredValidatorCalls(
		documentMock,
		author,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		documentRoot,
	)

	identityServiceMock.On(
		"ValidateSignature",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	currentVersionAnchorID, err := anchors.ToAnchorID(currentVersion)
	assert.NoError(t, err)

	nextVersionAnchorID, err := anchors.ToAnchorID(nextVersion)
	assert.NoError(t, err)

	anchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorServiceMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorServiceMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)

	coreDocument := &coredocumentpb.CoreDocument{}

	documentMock.On("PackCoreDocument").
		Return(coreDocument, nil)

	p2pConnectionTimeout := 1 * time.Second

	configMock.On("GetP2PConnectionTimeout").Return(p2pConnectionTimeout)

	p2pClientMock.On(
		"SendAnchoredDocument",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil, errors.New("error"))

	err = ap.SendDocument(ctx, documentMock)
	assert.NoError(t, err)
}

func TestAnchorProcessor_RequestDocumentWithAccessToken(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	granterAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	delegatingDocumentIdentifier := utils.RandomSlice(32)

	accessTokenRequest := &p2ppb.AccessTokenRequest{DelegatingDocumentIdentifier: delegatingDocumentIdentifier, AccessTokenId: tokenIdentifier}

	request := &p2ppb.GetDocumentRequest{DocumentIdentifier: documentIdentifier,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: accessTokenRequest,
	}

	getDocRes := &p2ppb.GetDocumentResponse{}

	p2pClientMock.On("GetDocumentRequest", ctx, granterAccountID, request).
		Return(getDocRes, nil)

	res, err := ap.RequestDocumentWithAccessToken(
		ctx,
		granterAccountID,
		tokenIdentifier,
		documentIdentifier,
		delegatingDocumentIdentifier,
	)
	assert.NoError(t, err)
	assert.Equal(t, getDocRes, res)
}

func TestAnchorProcessor_RequestDocumentWithAccessToken_P2PClientError(t *testing.T) {
	p2pClientMock := NewClientMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	configMock := config.NewConfigurationMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	ap := NewAnchorProcessor(p2pClientMock, anchorServiceMock, configMock, identityServiceMock)

	ctx := context.Background()

	granterAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	delegatingDocumentIdentifier := utils.RandomSlice(32)

	accessTokenRequest := &p2ppb.AccessTokenRequest{DelegatingDocumentIdentifier: delegatingDocumentIdentifier, AccessTokenId: tokenIdentifier}

	request := &p2ppb.GetDocumentRequest{DocumentIdentifier: documentIdentifier,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: accessTokenRequest,
	}

	p2pClientMock.On("GetDocumentRequest", ctx, granterAccountID, request).
		Return(nil, errors.New("error"))

	res, err := ap.RequestDocumentWithAccessToken(
		ctx,
		granterAccountID,
		tokenIdentifier,
		documentIdentifier,
		delegatingDocumentIdentifier,
	)
	assert.ErrorIs(t, err, ErrP2PDocumentRetrieval)
	assert.Nil(t, res)
}

func TestConsensusSignaturePayload(t *testing.T) {
	dataRoot := utils.RandomSlice(11)

	unvalidatedFlag := byte(0)

	res := ConsensusSignaturePayload(dataRoot, false)
	assert.Equal(t, append(dataRoot, unvalidatedFlag), res)

	validatedFlag := byte(1)

	res = ConsensusSignaturePayload(dataRoot, true)
	assert.Equal(t, append(dataRoot, validatedFlag), res)
}

func mockDocumentPreAnchoredValidatorCalls(
	documentMock *DocumentMock,
	author *types.AccountID,
	collaborators []*types.AccountID,
	documentID []byte,
	currentVersion []byte,
	nextVersion []byte,
	signingRoot []byte,
	signatures []*coredocumentpb.Signature,
) {
	documentMock.On("ID").Return(documentID)
	documentMock.On("Author").Return(author, nil)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)
	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("Signatures").Return(signatures)
	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)
	documentMock.On("GetAttributes").Return(nil)
	documentMock.On("GetComputeFieldsRules").Return(nil)
}
