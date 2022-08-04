//go:build unit

package documents

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/notification"

	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	any2 "github.com/golang/protobuf/ptypes/any"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestService_GetCurrentVersion(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)

	repoMock.On("GetLatest", identity.ToBytes(), documentID).
		Once().
		Return(documentMock, nil)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	res, err := service.GetCurrentVersion(ctx, documentID)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_GetCurrentVersion_ContextAccountError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	documentID := utils.RandomSlice(32)

	res, err := service.GetCurrentVersion(context.Background(), documentID)
	assert.ErrorIs(t, err, ErrDocumentConfigAccount)
	assert.Nil(t, res)
}

func TestService_GetCurrentVersion_RepoError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)

	repoErr := errors.New("error")

	repoMock.On("GetLatest", identity.ToBytes(), documentID).
		Once().
		Return(nil, repoErr)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	res, err := service.GetCurrentVersion(ctx, documentID)
	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))
	assert.Nil(t, res)
}

func TestService_GetVersion(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)

	version := utils.RandomSlice(32)

	repoMock.On("Get", identity.ToBytes(), version).
		Once().
		Return(documentMock, nil)

	documentMock.On("ID").
		Once().
		Return(documentID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	res, err := service.GetVersion(ctx, documentID, version)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_GetVersion_ContextAccountError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	documentID := utils.RandomSlice(32)
	version := utils.RandomSlice(32)

	res, err := service.GetVersion(context.Background(), documentID, version)
	assert.ErrorIs(t, err, ErrDocumentConfigAccount)
	assert.Nil(t, res)
}

func TestService_GetVersion_RepoError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)

	version := utils.RandomSlice(32)

	repoMock.On("Get", identity.ToBytes(), version).
		Once().
		Return(nil, errors.New("error"))

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	res, err := service.GetVersion(ctx, documentID, version)
	assert.True(t, errors.IsOfType(ErrDocumentVersionNotFound, err))
	assert.Nil(t, res)
}

func TestService_GetVersion_DocumentIDMismatch(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)

	version := utils.RandomSlice(32)

	repoMock.On("Get", identity.ToBytes(), version).
		Once().
		Return(documentMock, nil)

	documentMock.On("ID").
		Once().
		Return(utils.RandomSlice(32))

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	res, err := service.GetVersion(ctx, documentID, version)
	assert.True(t, errors.IsOfType(ErrDocumentVersionNotFound, err))
	assert.Nil(t, res)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	serviceID := "service-id"

	serviceMock := NewServiceMock(t)

	err := serviceRegistry.Register(serviceID, serviceMock)
	assert.NoError(t, err)

	cd := coredocumentpb.CoreDocument{
		EmbeddedData: &any2.Any{
			TypeUrl: serviceID,
		},
	}

	documentMock := NewDocumentMock(t)

	serviceMock.On("DeriveFromCoreDocument", cd).
		Once().
		Return(documentMock, nil)

	res, err := service.DeriveFromCoreDocument(cd)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_DeriveFromCoreDocument_NilEmbeddedData(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	serviceID := "service-id"

	serviceMock := NewServiceMock(t)

	err := serviceRegistry.Register(serviceID, serviceMock)
	assert.NoError(t, err)

	cd := coredocumentpb.CoreDocument{}

	res, err := service.DeriveFromCoreDocument(cd)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestService_DeriveFromCoreDocument_RegistryServiceErr(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	serviceID := "service-id"

	cd := coredocumentpb.CoreDocument{
		EmbeddedData: &any2.Any{
			TypeUrl: serviceID,
		},
	}

	res, err := service.DeriveFromCoreDocument(cd)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestService_CreateProofs(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	mockDocumentPostAnchoredValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		documentRoot,
	)

	repoMock.On("GetLatest", identity.ToBytes(), documentID).
		Once().
		Return(documentMock, nil)

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

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	fields := []string{"cd_tree.document_type"}

	docProof := &DocumentProof{}

	documentMock.On("CreateProofs", fields).
		Return(docProof, nil)

	proof, err := service.CreateProofs(ctx, documentID, fields)
	assert.Nil(t, err)
	assert.NotNil(t, proof)

	assert.Equal(t, documentID, proof.DocumentID)
	assert.Equal(t, currentVersion, proof.VersionID)
}

func TestService_CreateProofs_ValidatorError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	mockDocumentPostAnchoredValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		documentRoot,
	)

	repoMock.On("GetLatest", identity.ToBytes(), documentID).
		Once().
		Return(documentMock, nil)

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

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	// Invalid timestamp.
	invalidDuration := 3 * time.Hour
	docTimestamp := anchorTime.Add(-invalidDuration)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	fields := []string{"cd_tree.document_type"}

	proof, err := service.CreateProofs(ctx, documentID, fields)
	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))
	assert.Nil(t, proof)
}

func TestService_CreateProofs_DocumentProofError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	mockDocumentPostAnchoredValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		documentRoot,
	)

	repoMock.On("GetLatest", identity.ToBytes(), documentID).
		Once().
		Return(documentMock, nil)

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

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	fields := []string{"cd_tree.document_type"}

	documentMock.On("CreateProofs", fields).
		Return(nil, errors.New("boom"))

	proof, err := service.CreateProofs(ctx, documentID, fields)
	assert.True(t, errors.IsOfType(ErrDocumentProof, err))
	assert.Nil(t, proof)
}

func TestService_CreateProofsForVersion(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	mockDocumentPostAnchoredValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
		documentRoot,
	)

	repoMock.On("Get", identity.ToBytes(), currentVersion).
		Once().
		Return(documentMock, nil)

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

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	fields := []string{"cd_tree.document_type"}

	docProof := &DocumentProof{}

	documentMock.On("CreateProofs", fields).
		Return(docProof, nil)

	proof, err := service.CreateProofsForVersion(ctx, documentID, currentVersion, fields)
	assert.Nil(t, err)
	assert.NotNil(t, proof)

	assert.Equal(t, documentID, proof.DocumentID)
	assert.Equal(t, currentVersion, proof.VersionID)
}

func TestService_CreateProofsForVersion_ContextAccountError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)

	fields := []string{"cd_tree.document_type"}

	proof, err := service.CreateProofsForVersion(context.Background(), documentID, currentVersion, fields)
	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))
	assert.Nil(t, proof)
}

func TestService_CreateProofsForVersion_DocumentVersionNotFoundError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repoMock.On("Get", identity.ToBytes(), currentVersion).
		Once().
		Return(nil, errors.New("error"))

	fields := []string{"cd_tree.document_type"}

	proof, err := service.CreateProofsForVersion(ctx, documentID, currentVersion, fields)
	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))
	assert.Nil(t, proof)
}

func TestService_CreateProofsForVersion_DocumentIDMismatch(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	identity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Once().
		Return(identity)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)

	documentMock.On("ID").
		Once().
		Return(utils.RandomSlice(32))

	repoMock.On("Get", identity.ToBytes(), currentVersion).
		Once().
		Return(documentMock, nil)

	fields := []string{"cd_tree.document_type"}

	proof, err := service.CreateProofsForVersion(ctx, documentID, currentVersion, fields)
	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))
	assert.Nil(t, proof)
}

func TestService_RequestDocumentSignature(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(nil)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentRequestDocumentSignatureValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, errors.New("error"))

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	signaturePayload := ConsensusSignaturePayload(signingRoot, false)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", signaturePayload).
		Once().
		Return(signature, nil)

	documentMock.On("AppendSignatures", signature)

	documentMock.On("SetStatus", Committing).
		Once().
		Return(nil)

	repoMock.On("Exists", accountID.ToBytes(), documentID).
		Once().
		Return(true)

	repoMock.On("Create", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Once().
		Return(nil)

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.NoError(t, err)
	assert.Equal(t, []*coredocumentpb.Signature{signature}, res)
}

func TestService_RequestDocumentSignature_OldDocumentPresent(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	previousVersion := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(previousVersion)

	oldDocumentMock := NewDocumentMock(t)

	repoMock.On("Get", accountID.ToBytes(), previousVersion).
		Once().
		Return(oldDocumentMock, nil)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentRequestDocumentSignatureValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
	)

	oldDocumentMock.On("CollaboratorCanUpdate", documentMock, documentAuthor).
		Once().
		Return(nil)

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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, errors.New("error"))

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	signaturePayload := ConsensusSignaturePayload(signingRoot, true)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", signaturePayload).
		Once().
		Return(signature, nil)

	documentMock.On("AppendSignatures", signature)

	documentMock.On("SetStatus", Committing).
		Once().
		Return(nil)

	repoMock.On("Exists", accountID.ToBytes(), documentID).
		Once().
		Return(true)

	repoMock.On("Create", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Once().
		Return(nil)

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.NoError(t, err)
	assert.Equal(t, []*coredocumentpb.Signature{signature}, res)
}

func TestService_RequestDocumentSignature_OldDocumentPresent_RepoError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	previousVersion := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(previousVersion)

	repoErr := errors.New("error")

	// Everything should work despite this error since we are only logging it
	// and continuing with the validation.
	repoMock.On("Get", accountID.ToBytes(), previousVersion).
		Once().
		Return(nil, repoErr)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentRequestDocumentSignatureValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, errors.New("error"))

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	signaturePayload := ConsensusSignaturePayload(signingRoot, false)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", signaturePayload).
		Once().
		Return(signature, nil)

	documentMock.On("AppendSignatures", signature)

	documentMock.On("SetStatus", Committing).
		Once().
		Return(nil)

	repoMock.On("Exists", accountID.ToBytes(), documentID).
		Once().
		Return(true)

	repoMock.On("Create", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Once().
		Return(nil)

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.NoError(t, err)
	assert.Equal(t, []*coredocumentpb.Signature{signature}, res)
}

func TestService_RequestDocumentSignature_DocIDAndCurrentVersionMismatch(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(nil)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentRequestDocumentSignatureValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, errors.New("error"))

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	signaturePayload := ConsensusSignaturePayload(signingRoot, false)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", signaturePayload).
		Once().
		Return(signature, nil)

	documentMock.On("AppendSignatures", signature)

	documentMock.On("SetStatus", Committing).
		Once().
		Return(nil)

	repoMock.On("Exists", accountID.ToBytes(), documentID).
		Once().
		Return(false)

	repoMock.On("Create", accountID.ToBytes(), documentID, documentMock).
		Once().
		Return(nil)

	repoMock.On("Create", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Once().
		Return(nil)

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.NoError(t, err)
	assert.Equal(t, []*coredocumentpb.Signature{signature}, res)
}

func TestService_RequestDocumentSignature_ContextAccountError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentMock := NewDocumentMock(t)

	res, err := service.RequestDocumentSignature(context.Background(), documentMock, documentAuthor)
	assert.ErrorIs(t, err, ErrDocumentConfigAccount)
	assert.Nil(t, res)
}

func TestService_RequestDocumentSignature_NilDocument(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	res, err := service.RequestDocumentSignature(ctx, nil, documentAuthor)
	assert.ErrorIs(t, err, ErrDocumentNil)
	assert.Nil(t, res)
}

func TestService_RequestDocumentSignature_ValidationError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(nil)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentRequestDocumentSignatureValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	// This will throw ErrDocumentIDReused causing validation to fail.
	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, nil)

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))
	assert.Nil(t, res)
}

func TestService_RequestDocumentSignature_SigningRootError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(nil)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock.On("ID").Return(documentID)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)
	documentMock.On("Signatures").Return(getTestSignatures(documentAuthor, collaborators))
	documentMock.On("Author").Return(documentAuthor, nil)
	documentMock.On("GetSignerCollaborators", documentAuthor).Return(collaborators, nil)
	documentMock.On("GetAttributes").Return(nil)
	documentMock.On("GetComputeFieldsRules").Return(nil)
	documentMock.On("Timestamp").Return(time.Now(), nil)

	// Called 2 times during validation.
	documentMock.On("CalculateSigningRoot").
		Times(2).
		Return(signingRoot, nil)

	// Called 1 time after validation, this is where we error out for this test.
	documentMock.On("CalculateSigningRoot").
		Times(1).
		Return(nil, errors.New("error"))

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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, errors.New("error"))

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestService_RequestDocumentSignature_SignMessageError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(nil)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentRequestDocumentSignatureValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, errors.New("error"))

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	signaturePayload := ConsensusSignaturePayload(signingRoot, false)

	signErr := errors.New("error")

	accountMock.On("SignMsg", signaturePayload).
		Once().
		Return(nil, signErr)

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.ErrorIs(t, err, signErr)
	assert.Nil(t, res)
}

func TestService_RequestDocumentSignature_SetStatusError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(nil)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentRequestDocumentSignatureValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, errors.New("error"))

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	signaturePayload := ConsensusSignaturePayload(signingRoot, false)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", signaturePayload).
		Once().
		Return(signature, nil)

	documentMock.On("AppendSignatures", signature)

	setStatusErr := errors.New("error")
	documentMock.On("SetStatus", Committing).
		Once().
		Return(setStatusErr)

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.ErrorIs(t, err, setStatusErr)
	assert.Nil(t, res)
}

func TestService_RequestDocumentSignature_DocIDAndCurrentVersionMismatch_DocIDCreateRepoError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(nil)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentRequestDocumentSignatureValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, errors.New("error"))

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	signaturePayload := ConsensusSignaturePayload(signingRoot, false)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", signaturePayload).
		Once().
		Return(signature, nil)

	documentMock.On("AppendSignatures", signature)

	documentMock.On("SetStatus", Committing).
		Once().
		Return(nil)

	repoMock.On("Exists", accountID.ToBytes(), documentID).
		Once().
		Return(false)

	repoErr := errors.New("error")

	repoMock.On("Create", accountID.ToBytes(), documentID, documentMock).
		Once().
		Return(repoErr)

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.True(t, errors.IsOfType(ErrDocumentPersistence, err))
	assert.Nil(t, res)
}

func TestService_RequestDocumentSignature_DocIDAndCurrentVersionMismatch_DocCurrentVersionRepoError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(nil)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentRequestDocumentSignatureValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, errors.New("error"))

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	signaturePayload := ConsensusSignaturePayload(signingRoot, false)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", signaturePayload).
		Once().
		Return(signature, nil)

	documentMock.On("AppendSignatures", signature)

	documentMock.On("SetStatus", Committing).
		Once().
		Return(nil)

	repoMock.On("Exists", accountID.ToBytes(), documentID).
		Once().
		Return(false)

	repoMock.On("Create", accountID.ToBytes(), documentID, documentMock).
		Once().
		Return(nil)

	repoErr := errors.New("error")

	repoMock.On("Create", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Once().
		Return(repoErr)

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.True(t, errors.IsOfType(ErrDocumentPersistence, err))
	assert.Nil(t, res)
}

func TestService_RequestDocumentSignature_RepoCreateError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := NewDocumentMock(t)
	documentMock.On("PreviousVersion").Return(nil)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	mockDocumentRequestDocumentSignatureValidatorCalls(
		documentMock,
		documentAuthor,
		collaborators,
		documentID,
		currentVersion,
		nextVersion,
		signingRoot,
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

	currentVersionAnchorTime := time.Now()

	anchorRoot, err := anchors.ToDocumentRoot(documentRoot)

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, currentVersionAnchorTime, errors.New("error"))

	nextVersionAnchorTime := time.Now()

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(anchorRoot, nextVersionAnchorTime, errors.New("error"))

	signaturePayload := ConsensusSignaturePayload(signingRoot, false)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", signaturePayload).
		Once().
		Return(signature, nil)

	documentMock.On("AppendSignatures", signature)

	documentMock.On("SetStatus", Committing).
		Once().
		Return(nil)

	repoMock.On("Exists", accountID.ToBytes(), documentID).
		Once().
		Return(true)

	repoErr := errors.New("error")

	repoMock.On("Create", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Once().
		Return(repoErr)

	res, err := service.RequestDocumentSignature(ctx, documentMock, documentAuthor)
	assert.True(t, errors.IsOfType(ErrDocumentPersistence, err))
	assert.Nil(t, res)
}

func TestService_ReceiveAnchoredDocument(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentMock.On("PreviousVersion").Return(nil)

	mockDocumentReceivedAnchoredDocumentValidatorCalls(
		documentMock,
		documentAuthor,
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

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	documentMock.On("SetStatus", Committed).
		Return(nil)

	repoMock.On("Update", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Return(nil)

	notifierMock.On("Send", ctx, mock.IsType(notification.Message{})).
		Return(nil)

	err = service.ReceiveAnchoredDocument(ctx, documentMock, documentAuthor)
	assert.NoError(t, err)

	// Sleep to ensure that the notifier is called.
	time.Sleep(1 * time.Second)
}

func TestService_ReceiveAnchoredDocument_OldDocumentPresent(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	previousVersion := utils.RandomSlice(32)

	documentMock.On("PreviousVersion").Return(previousVersion)

	oldDocumentMock := NewDocumentMock(t)

	repoMock.On("Get", accountID.ToBytes(), previousVersion).
		Return(oldDocumentMock, nil)

	oldDocumentMock.On("CollaboratorCanUpdate", documentMock, documentAuthor).
		Return(nil)

	mockDocumentReceivedAnchoredDocumentValidatorCalls(
		documentMock,
		documentAuthor,
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

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	documentMock.On("SetStatus", Committed).
		Return(nil)

	repoMock.On("Update", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Return(nil)

	notifierMock.On("Send", ctx, mock.IsType(notification.Message{})).
		Return(nil)

	err = service.ReceiveAnchoredDocument(ctx, documentMock, documentAuthor)
	assert.NoError(t, err)

	// Sleep to ensure that the notifier is called.
	time.Sleep(1 * time.Second)
}

func TestService_ReceiveAnchoredDocument_ContextAccountError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	err = service.ReceiveAnchoredDocument(context.Background(), documentMock, documentAuthor)
	assert.ErrorIs(t, err, ErrDocumentConfigAccount)
}

func TestService_ReceiveAnchoredDocument_DocumentNil(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	err = service.ReceiveAnchoredDocument(ctx, nil, documentAuthor)
	assert.ErrorIs(t, err, ErrDocumentNil)
}

func TestService_ReceiveAnchoredDocument_OldDocumentPresent_RepoError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	previousVersion := utils.RandomSlice(32)

	documentMock.On("PreviousVersion").Return(previousVersion)

	// This error will not interrupt the operation since we are logging it.
	repoMock.On("Get", accountID.ToBytes(), previousVersion).
		Return(nil, errors.New("error"))

	mockDocumentReceivedAnchoredDocumentValidatorCalls(
		documentMock,
		documentAuthor,
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

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	documentMock.On("SetStatus", Committed).
		Return(nil)

	repoMock.On("Update", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Return(nil)

	notifierMock.On("Send", ctx, mock.IsType(notification.Message{})).
		Return(nil)

	err = service.ReceiveAnchoredDocument(ctx, documentMock, documentAuthor)
	assert.NoError(t, err)

	// Sleep to ensure that the notifier is called.
	time.Sleep(1 * time.Second)
}

func TestService_ReceiveAnchoredDocument_ValidationError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentMock.On("PreviousVersion").Return(nil)

	mockDocumentReceivedAnchoredDocumentValidatorCalls(
		documentMock,
		documentAuthor,
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

	// This will cause a validation error.
	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, errors.New("error"))

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	err = service.ReceiveAnchoredDocument(ctx, documentMock, documentAuthor)
	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))
}

func TestService_ReceiveAnchoredDocument_SetStatusError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentMock.On("PreviousVersion").Return(nil)

	mockDocumentReceivedAnchoredDocumentValidatorCalls(
		documentMock,
		documentAuthor,
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

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	setStatusError := errors.New("error")
	documentMock.On("SetStatus", Committed).
		Return(setStatusError)

	err = service.ReceiveAnchoredDocument(ctx, documentMock, documentAuthor)
	assert.ErrorIs(t, err, setStatusError)
}

func TestService_ReceiveAnchoredDocument_RepoUpdateError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentMock.On("PreviousVersion").Return(nil)

	mockDocumentReceivedAnchoredDocumentValidatorCalls(
		documentMock,
		documentAuthor,
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

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	documentMock.On("SetStatus", Committed).
		Return(nil)

	repoError := errors.New("error")

	repoMock.On("Update", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Return(repoError)

	err = service.ReceiveAnchoredDocument(ctx, documentMock, documentAuthor)
	assert.True(t, errors.IsOfType(ErrDocumentPersistence, err))
}

func TestService_ReceiveAnchoredDocument_NotifierError(t *testing.T) {
	configMock := config.NewConfigurationMock(t)
	repoMock := NewRepositoryMock(t)
	anchorsMock := anchors.NewServiceMock(t)
	serviceRegistry := NewServiceRegistry()
	dispatcherMock := jobs.NewDispatcherMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	notifierMock := notification.NewSenderMock(t)

	service := NewService(
		configMock,
		repoMock,
		anchorsMock,
		serviceRegistry,
		dispatcherMock,
		identityServiceMock,
		notifierMock,
	)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentID := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

	documentMock := NewDocumentMock(t)
	documentAuthor, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators, err := getTestCollaborators(2)
	assert.NoError(t, err)

	documentMock.On("PreviousVersion").Return(nil)

	mockDocumentReceivedAnchoredDocumentValidatorCalls(
		documentMock,
		documentAuthor,
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

	anchorsMock.On("GetAnchorData", currentVersionAnchorID).
		Once().
		Return(anchorRoot, anchorTime, nil)

	anchorsMock.On("GetAnchorData", nextVersionAnchorID).
		Once().
		Return(nil, time.Time{}, errors.New("error"))

	docTimestamp := anchorTime.Add(3 * time.Hour)
	documentMock.On("Timestamp").
		Return(docTimestamp, nil)

	documentMock.On("SetStatus", Committed).
		Return(nil)

	repoMock.On("Update", accountID.ToBytes(), documentMock.CurrentVersion(), documentMock).
		Return(nil)

	// This will not interrupt the operation.
	notifierError := errors.New("error")
	notifierMock.On("Send", ctx, mock.IsType(notification.Message{})).
		Return(notifierError)

	err = service.ReceiveAnchoredDocument(ctx, documentMock, documentAuthor)
	assert.NoError(t, err)

	// Sleep to ensure that the notifier is called.
	time.Sleep(1 * time.Second)
}

func getTestCollaborators(count int) ([]*types.AccountID, error) {
	var collaborators []*types.AccountID

	for i := 0; i < count; i++ {
		accountID, err := testingcommons.GetRandomAccountID()
		if err != nil {
			return nil, err
		}

		collaborators = append(collaborators, accountID)
	}

	return collaborators, nil
}

func mockDocumentReceivedAnchoredDocumentValidatorCalls(
	documentMock *DocumentMock,
	author *types.AccountID,
	collaborators []*types.AccountID,
	documentID []byte,
	currentVersion []byte,
	nextVersion []byte,
	signingRoot []byte,
	documentRoot []byte,
) {
	// Transition validator is only called when the old document is also present.

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
}

func mockDocumentPostAnchoredValidatorCalls(
	documentMock *DocumentMock,
	author *types.AccountID,
	collaborators []*types.AccountID,
	documentID []byte,
	currentVersion []byte,
	nextVersion []byte,
	signingRoot []byte,
	documentRoot []byte,
) {
	documentMock.On("ID").Return(documentID)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)
	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("Signatures").Return(getTestSignatures(author, collaborators))
	documentMock.On("Author").Return(author, nil)
	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)
	documentMock.On("GetAttributes").Return(nil)
	documentMock.On("GetComputeFieldsRules").Return(nil)
	documentMock.On("CalculateDocumentRoot").Return(documentRoot, nil)
}

func mockDocumentRequestDocumentSignatureValidatorCalls(
	documentMock *DocumentMock,
	author *types.AccountID,
	collaborators []*types.AccountID,
	documentID []byte,
	currentVersion []byte,
	nextVersion []byte,
	signingRoot []byte,
) {
	documentMock.On("ID").Return(documentID)
	documentMock.On("CurrentVersion").Return(currentVersion)
	documentMock.On("NextVersion").Return(nextVersion)
	documentMock.On("CalculateSigningRoot").Return(signingRoot, nil)
	documentMock.On("Signatures").Return(getTestSignatures(author, collaborators))
	documentMock.On("Author").Return(author, nil)
	documentMock.On("GetSignerCollaborators", author).Return(collaborators, nil)
	documentMock.On("GetAttributes").Return(nil)
	documentMock.On("GetComputeFieldsRules").Return(nil)

	documentMock.On("Timestamp").Return(time.Now(), nil)
}

func getTestSignatures(author *types.AccountID, collaborators []*types.AccountID) []coredocumentpb.Signature {
	var signatures []coredocumentpb.Signature

	authorSignature := coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            author.ToBytes(),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: true,
	}

	signatures = append(signatures, authorSignature)

	for _, collaborator := range collaborators {
		collaboratorSignature := coredocumentpb.Signature{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            collaborator.ToBytes(),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: false,
		}

		signatures = append(signatures, collaboratorSignature)
	}

	return signatures
}

//func TestService_Validate(t *testing.T) {
//	r := NewServiceRegistry()
//	scheme := "generic"
//	srv := new(MockService)
//	srv.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
//	err := r.Register(scheme, srv)
//	assert.NoError(t, err)
//
//	// unsupported svc schema
//	m := new(mockModel)
//	m.On("Scheme", mock.Anything).Return("some scheme")
//	s := service{registry: r}
//	err = s.Validate(context.Background(), m, nil)
//	assert.Error(t, err)
//
//	// create validation error, already anchored
//	id := utils.RandomSlice(32)
//	ctxh := testingconfig.CreateAccountContext(t, cfg)
//	m = new(mockModel)
//	nid := utils.RandomSlice(32)
//	m.On("ID", mock.Anything).Return(id)
//	m.On("CurrentVersion").Return(id)
//	m.On("NextVersion").Return(nid)
//	m.On("PreviousVersion").Return(nid)
//	m.On("Scheme", mock.Anything).Return("generic")
//	anchorSrv := new(anchors.MockAnchorService)
//	anchorSrv.On("GetAnchorData", mock.Anything).Return(utils.RandomSlice(32), nil)
//	s.anchorSrv = anchorSrv
//	err = s.Validate(ctxh, m, nil)
//	assert.Error(t, err)
//
//	// create validation success
//	anchorSrv = new(anchors.MockAnchorService)
//	anchorSrv.On("GetAnchorData", mock.Anything).Return(id, errors.New("anchor data missing"))
//	s.anchorSrv = anchorSrv
//	err = s.Validate(ctxh, m, nil)
//	assert.NoError(t, err)
//
//	// Update validation error, already anchored
//	m1 := new(mockModel)
//	nid1 := utils.RandomSlice(32)
//	m1.On("ID", mock.Anything).Return(id)
//	m1.On("CurrentVersion").Return(nid)
//	m1.On("NextVersion").Return(nid1)
//	m1.On("PreviousVersion").Return(id)
//	m1.On("Scheme", mock.Anything).Return("generic")
//	anchorSrv = new(anchors.MockAnchorService)
//	anchorSrv.On("GetAnchorData", mock.Anything).Return(utils.RandomSlice(32), nil)
//	s.anchorSrv = anchorSrv
//	err = s.Validate(ctxh, m1, m)
//	assert.Error(t, err)
//
//	// update validation success
//	anchorSrv = new(anchors.MockAnchorService)
//	anchorSrv.On("GetAnchorData", mock.Anything).Return(id, errors.New("anchor data missing"))
//	s.anchorSrv = anchorSrv
//	err = s.Validate(ctxh, m1, m)
//	assert.NoError(t, err)
//
//	// specific document validation error
//	r = NewServiceRegistry()
//	srv = new(MockService)
//	srv.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("specific document error"))
//	err = r.Register(scheme, srv)
//	assert.NoError(t, err)
//	s.registry = r
//	err = s.Validate(ctxh, m1, m)
//	assert.Error(t, err)
//}
//
//func TestService_Commit(t *testing.T) {
//	r := NewServiceRegistry()
//	scheme := "generic"
//	srv := new(MockService)
//	srv.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
//	err := r.Register(scheme, srv)
//	assert.NoError(t, err)
//	s := service{registry: r}
//	m := new(mockModel)
//	id := utils.RandomSlice(32)
//	m.On("ID", mock.Anything).Return(id)
//	m.On("Scheme", mock.Anything).Return("generic")
//
//	// Account ID not set
//	_, err = s.Commit(context.Background(), m)
//	assert.Error(t, err)
//
//	// db error when fetching
//	mr := new(MockRepository)
//	mr.On("GetLatest", mock.Anything, mock.Anything).Return(nil, errors.New("some db error")).Once()
//	s.repo = mr
//	_, err = s.Commit(context.Background(), m)
//	assert.Error(t, err)
//
//	// Fail validation
//	nid := utils.RandomSlice(32)
//	m.On("CurrentVersion").Return(id)
//	m.On("NextVersion").Return(nid)
//	m.On("PreviousVersion").Return(nid)
//	mr.On("GetLatest", mock.Anything, mock.Anything).Return(nil, ErrDocumentVersionNotFound)
//	anchorSrv := new(anchors.MockAnchorService)
//	anchorSrv.On("GetAnchorData", mock.Anything).Return(utils.RandomSlice(32), nil)
//	s.anchorSrv = anchorSrv
//	ctxh := testingconfig.CreateAccountContext(t, cfg)
//	_, err = s.Commit(ctxh, m)
//	assert.Error(t, err)
//
//	// Error create model
//	anchorSrv = new(anchors.MockAnchorService)
//	anchorSrv.On("GetAnchorData", mock.Anything).Return(nil, errors.New("anchor data missing"))
//	s.anchorSrv = anchorSrv
//	m.On("SetStatus", mock.Anything).Return(nil)
//	mr.On("Exists", mock.Anything, mock.Anything).Return(false)
//	mr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(ErrDocumentPersistence).Once()
//	_, err = s.Commit(ctxh, m)
//	assert.Error(t, err)
//
//	// Error anchoring
//	dispatcher := new(jobs.MockDispatcher)
//	dispatcher.On("Dispatch", mock.Anything, mock.Anything).Return(nil, errors.New("dispatch failed")).Once()
//	s.dispatcher = dispatcher
//	mr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
//	_, err = s.Commit(ctxh, m)
//	assert.Error(t, err)
//
//	// Commit success
//	dispatcher.On("Dispatch", mock.Anything, mock.Anything).Return(new(jobs.MockResult), nil).Once()
//	_, err = s.Commit(ctxh, m)
//	assert.NoError(t, err)
//	dispatcher.AssertExpectations(t)
//	mr.AssertExpectations(t)
//}
//
//func TestService_Derive(t *testing.T) {
//	scheme := "generic"
//	attr, err := NewStringAttribute("test", AttrString, "value")
//	assert.NoError(t, err)
//	attrs := map[AttrKey]Attribute{
//		attr.Key: attr,
//	}
//	cid, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
//	assert.NoError(t, err)
//	payload := UpdatePayload{CreatePayload: CreatePayload{
//		Scheme:     scheme,
//		Attributes: attrs,
//		Collaborators: CollaboratorsAccess{
//			ReadCollaborators:      nil,
//			ReadWriteCollaborators: []identity.DID{cid},
//		},
//	}}
//	s := service{}
//
//	// unknown scheme
//	ctx := testingconfig.CreateAccountContext(t, cfg)
//	s.registry = NewServiceRegistry()
//	_, err = s.Derive(ctx, payload)
//	assert.Error(t, err)
//	assert.True(t, errors.IsOfType(ErrDocumentSchemeUnknown, err))
//
//	// derive failed
//	doc := new(MockModel)
//	docSrv := new(MockService)
//	docSrv.On("New", scheme).Return(doc, nil)
//	doc.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(errors.New("derive failed")).Once()
//	assert.NoError(t, s.registry.Register(scheme, docSrv))
//	_, err = s.Derive(ctx, payload)
//	assert.Error(t, err)
//	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))
//
//	// create successful
//	doc.On("DeriveFromCreatePayload", mock.Anything, mock.Anything).Return(nil).Once()
//	gdoc, err := s.Derive(ctx, payload)
//	assert.NoError(t, err)
//	assert.Equal(t, doc, gdoc)
//
//	// missing old version
//	docID := utils.RandomSlice(32)
//	repo := new(MockRepository)
//	repo.On("GetLatest", did[:], docID).Return(nil, ErrDocumentNotFound).Once()
//	s.repo = repo
//	payload.DocumentID = docID
//	_, err = s.Derive(ctx, payload)
//	assert.Error(t, err)
//	assert.True(t, errors.IsOfType(ErrDocumentNotFound, err))
//
//	// invalid type
//	doc.On("Scheme").Return("invalid").Once()
//	repo.On("GetLatest", did[:], docID).Return(doc, nil)
//	_, err = s.Derive(ctx, payload)
//	assert.Error(t, err)
//	assert.True(t, errors.IsOfType(ErrDocumentInvalidType, err))
//
//	// DeriveFromUpdatePayload failed
//	doc.On("Scheme").Return(scheme)
//	doc.On("DeriveFromUpdatePayload", mock.Anything, payload).Return(nil, ErrDocumentInvalid).Once()
//	_, err = s.Derive(ctx, payload)
//	assert.Error(t, err)
//	assert.True(t, errors.IsOfType(ErrDocumentInvalid, err))
//
//	// success
//	doc.On("DeriveFromUpdatePayload", mock.Anything, payload).Return(doc, nil).Once()
//	gdoc, err = s.Derive(ctx, payload)
//	assert.NoError(t, err)
//	assert.Equal(t, gdoc, doc)
//	doc.AssertExpectations(t)
//	repo.AssertExpectations(t)
//	docSrv.AssertExpectations(t)
//}
