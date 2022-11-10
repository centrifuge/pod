//go:build unit

package pending

import (
	"context"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/gocelery/v2"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func TestService_Get(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	res, err := pendingDocService.Get(ctx, documentID, documents.Pending)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_Get_NonPendingDocument(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)

	ctx := context.Background()

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("GetCurrentVersion", ctx, documentID).
		Return(documentMock, nil).
		Once()

	res, err := pendingDocService.Get(ctx, documentID, documents.Committing)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_Get_NonPendingDocument_RetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)

	ctx := context.Background()

	documentServiceError := errors.New("error")

	documentServiceMock.On("GetCurrentVersion", ctx, documentID).
		Return(nil, documentServiceError).
		Once()

	res, err := pendingDocService.Get(ctx, documentID, documents.Committing)
	assert.ErrorIs(t, err, documentServiceError)
	assert.Nil(t, res)
}

func TestService_Get_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)

	ctx := context.Background()

	res, err := pendingDocService.Get(ctx, documentID, documents.Pending)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_Get_StorageRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.Get(ctx, documentID, documents.Pending)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_GetVersion(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	versionID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentServiceMock.On("GetVersion", ctx, documentID, versionID).
		Return(nil, errors.New("error")).
		Once()

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("CurrentVersion").
		Return(versionID).
		Once()

	res, err := pendingDocService.GetVersion(ctx, documentID, versionID)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_GetVersion_VersionFoundInDocService(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)
	versionID := utils.RandomSlice(32)

	ctx := context.Background()

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("GetVersion", ctx, documentID, versionID).
		Return(documentMock, nil).
		Once()

	res, err := pendingDocService.GetVersion(ctx, documentID, versionID)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_GetVersion_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)
	versionID := utils.RandomSlice(32)

	ctx := context.Background()

	documentServiceMock.On("GetVersion", ctx, documentID, versionID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.GetVersion(ctx, documentID, versionID)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_GetVersion_RetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	versionID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentServiceMock.On("GetVersion", ctx, documentID, versionID).
		Return(nil, errors.New("error")).
		Once()

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.GetVersion(ctx, documentID, versionID)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_GetVersion_VersionMismatch(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	versionID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentServiceMock.On("GetVersion", ctx, documentID, versionID).
		Return(nil, errors.New("error")).
		Once()

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("CurrentVersion").
		Return(utils.RandomSlice(32)).
		Once()

	res, err := pendingDocService.GetVersion(ctx, documentID, versionID)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_Create(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
	}

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("Derive", ctx, payload).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID).
		Once()

	repositoryMock.On("Create", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.Create(ctx, payload)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_Create_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	ctx := context.Background()

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
	}

	res, err := pendingDocService.Create(ctx, payload)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_Create_WithDocumentID_PendingDocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
		DocumentID: documentID,
	}

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("Derive", ctx, payload).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID).
		Once()

	repositoryMock.On("Create", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.Create(ctx, payload)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_Create_WithDocumentID_PendingDocumentFound(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
		DocumentID: documentID,
	}

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, nil).
		Once()

	res, err := pendingDocService.Create(ctx, payload)
	assert.ErrorIs(t, err, ErrPendingDocumentExists)
	assert.Nil(t, res)
}

func TestService_Create_DeriveError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
	}

	deriveError := errors.New("error")

	documentServiceMock.On("Derive", ctx, payload).
		Return(nil, deriveError).
		Once()

	res, err := pendingDocService.Create(ctx, payload)
	assert.ErrorIs(t, err, deriveError)
	assert.Nil(t, res)
}

func TestService_Create_RepositoryCreateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
	}

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("Derive", ctx, payload).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID).
		Once()

	repoError := errors.New("error")

	repositoryMock.On("Create", accountID.ToBytes(), documentID, documentMock).
		Return(repoError).
		Once()

	res, err := pendingDocService.Create(ctx, payload)
	assert.ErrorIs(t, err, repoError)
	assert.Equal(t, documentMock, res)
}

func TestService_Clone(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.ClonePayload{
		Scheme: "generic",
	}

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("DeriveClone", ctx, payload).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID).
		Once()

	repositoryMock.On("Create", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.Clone(ctx, payload)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_Clone_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)

	ctx := context.Background()

	payload := documents.ClonePayload{
		Scheme:     "generic",
		TemplateID: documentID,
	}

	res, err := pendingDocService.Clone(ctx, payload)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_Clone_WithTemplateID(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.ClonePayload{
		Scheme:     "generic",
		TemplateID: documentID,
	}

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("DeriveClone", ctx, payload).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID).
		Once()

	repositoryMock.On("Create", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.Clone(ctx, payload)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_Clone_WithTemplateID_PendingDocumentFound(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.ClonePayload{
		Scheme:     "generic",
		TemplateID: documentID,
	}

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, nil).
		Once()

	res, err := pendingDocService.Clone(ctx, payload)
	assert.ErrorIs(t, err, ErrPendingDocumentExists)
	assert.Nil(t, res)
}

func TestService_Clone_DeriveCloneError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.ClonePayload{
		Scheme: "generic",
	}

	deriveCloneError := errors.New("error")

	documentServiceMock.On("DeriveClone", ctx, payload).
		Return(nil, deriveCloneError).
		Once()

	res, err := pendingDocService.Clone(ctx, payload)
	assert.ErrorIs(t, err, deriveCloneError)
	assert.Nil(t, res)
}

func TestService_Clone_RepositoryCreateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.ClonePayload{
		Scheme: "generic",
	}

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("DeriveClone", ctx, payload).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID).
		Once()

	repoError := errors.New("error")

	repositoryMock.On("Create", accountID.ToBytes(), documentID, documentMock).
		Return(repoError).
		Once()

	res, err := pendingDocService.Clone(ctx, payload)
	assert.ErrorIs(t, err, repoError)
	assert.Equal(t, documentMock, res)
}

func TestService_Update(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
		DocumentID: documentID,
	}

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID).
		Once()

	documentMock.On("Patch", payload).
		Return(nil).
		Once()

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.Update(ctx, payload)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_Update_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)

	ctx := context.Background()

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
		DocumentID: documentID,
	}

	res, err := pendingDocService.Update(ctx, payload)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_Update_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
		DocumentID: documentID,
	}

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.Update(ctx, payload)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_Update_DocumentPatchError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
		DocumentID: documentID,
	}

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	patchError := errors.New("error")

	documentMock.On("Patch", payload).
		Return(patchError).
		Once()

	res, err := pendingDocService.Update(ctx, payload)
	assert.ErrorIs(t, err, patchError)
	assert.Nil(t, res)
}

func TestService_Update_RepositoryUpdateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        "generic",
			Collaborators: documents.CollaboratorsAccess{},
		},
		DocumentID: documentID,
	}

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID).
		Once()

	documentMock.On("Patch", payload).
		Return(nil).
		Once()

	repoError := errors.New("error")

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(repoError).
		Once()

	res, err := pendingDocService.Update(ctx, payload)
	assert.ErrorIs(t, err, repoError)
	assert.Equal(t, documentMock, res)
}

func TestService_Commit(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	expectedJobID := gocelery.JobID{}

	documentServiceMock.On("Commit", ctx, documentMock).
		Return(expectedJobID, nil).
		Once()

	repositoryMock.On("Delete", accountID.ToBytes(), documentID).
		Return(nil).
		Once()

	res, jobID, err := pendingDocService.Commit(ctx, documentID)
	assert.NoError(t, err)
	assert.Equal(t, expectedJobID, jobID)
	assert.Equal(t, documentMock, res)
}

func TestService_Commit_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)

	ctx := context.Background()

	res, jobID, err := pendingDocService.Commit(ctx, documentID)

	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, jobID)
	assert.Nil(t, res)
}

func TestService_Commit_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, jobID, err := pendingDocService.Commit(ctx, documentID)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, jobID)
	assert.Nil(t, res)
}

func TestService_Commit_CommitError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	commitError := errors.New("error")

	documentServiceMock.On("Commit", ctx, documentMock).
		Return(nil, commitError).
		Once()

	res, jobID, err := pendingDocService.Commit(ctx, documentID)
	assert.ErrorIs(t, err, commitError)
	assert.Nil(t, jobID)
	assert.Nil(t, res)
}

func TestService_Commit_RepositoryDeleteError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	expectedJobID := gocelery.JobID{}

	documentServiceMock.On("Commit", ctx, documentMock).
		Return(expectedJobID, nil).
		Once()

	deleteError := errors.New("error")

	repositoryMock.On("Delete", accountID.ToBytes(), documentID).
		Return(deleteError).
		Once()

	res, jobID, err := pendingDocService.Commit(ctx, documentID)
	assert.ErrorIs(t, err, deleteError)
	assert.Equal(t, expectedJobID, jobID)
	assert.Equal(t, documentMock, res)
}

func TestService_AddSignedAttribute(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	label := "test-label"
	value := utils.RandomSlice(32)
	attrType := documents.AttrInt256
	currentVersion := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID)
	documentMock.On("CurrentVersion").
		Return(currentVersion)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", mock.Anything).
		Return(signature, nil)

	attr, err := documents.NewSignedAttribute(
		label,
		accountID,
		accountMock,
		documentID,
		currentVersion,
		value,
		attrType,
	)
	assert.NoError(t, err)

	documentMock.On(
		"AddAttributes",
		documents.CollaboratorsAccess{},
		false,
		attr,
	).Return(nil).Once()

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.AddSignedAttribute(ctx, documentID, label, value, attrType)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_AddSignedAttribute_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)
	label := "test-label"
	value := utils.RandomSlice(32)
	attrType := documents.AttrInt256

	ctx := context.Background()

	res, err := pendingDocService.AddSignedAttribute(ctx, documentID, label, value, attrType)
	assert.ErrorIs(t, err, errors.ErrContextAccountRetrieval)
	assert.Nil(t, res)
}

func TestService_AddSignedAttribute_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	label := "test-label"
	value := utils.RandomSlice(32)
	attrType := documents.AttrInt256

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.AddSignedAttribute(ctx, documentID, label, value, attrType)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_AddSignedAttribute_NewSignedAttributeError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	label := "test-label"
	value := utils.RandomSlice(32)
	attrType := documents.AttrInt256
	currentVersion := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID)
	documentMock.On("CurrentVersion").
		Return(currentVersion)

	signatureError := errors.New("error")

	accountMock.On("SignMsg", mock.Anything).
		Return(nil, signatureError)

	res, err := pendingDocService.AddSignedAttribute(ctx, documentID, label, value, attrType)
	assert.ErrorIs(t, err, signatureError)
	assert.Nil(t, res)
}

func TestService_AddSignedAttribute_AddAttributesError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	label := "test-label"
	value := utils.RandomSlice(32)
	attrType := documents.AttrInt256
	currentVersion := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID)
	documentMock.On("CurrentVersion").
		Return(currentVersion)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", mock.Anything).
		Return(signature, nil)

	attr, err := documents.NewSignedAttribute(
		label,
		accountID,
		accountMock,
		documentID,
		currentVersion,
		value,
		attrType,
	)
	assert.NoError(t, err)

	addAttributesError := errors.New("error")

	documentMock.On(
		"AddAttributes",
		documents.CollaboratorsAccess{},
		false,
		attr,
	).Return(addAttributesError).Once()

	res, err := pendingDocService.AddSignedAttribute(ctx, documentID, label, value, attrType)
	assert.ErrorIs(t, err, addAttributesError)
	assert.Nil(t, res)
}

func TestService_AddSignedAttribute_RepositoryUpdateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	label := "test-label"
	value := utils.RandomSlice(32)
	attrType := documents.AttrInt256
	currentVersion := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("ID").
		Return(documentID)
	documentMock.On("CurrentVersion").
		Return(currentVersion)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", mock.Anything).
		Return(signature, nil)

	attr, err := documents.NewSignedAttribute(
		label,
		accountID,
		accountMock,
		documentID,
		currentVersion,
		value,
		attrType,
	)
	assert.NoError(t, err)

	documentMock.On(
		"AddAttributes",
		documents.CollaboratorsAccess{},
		false,
		attr,
	).Return(nil).Once()

	updateError := errors.New("error")

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(updateError).
		Once()

	res, err := pendingDocService.AddSignedAttribute(ctx, documentID, label, value, attrType)
	assert.ErrorIs(t, err, updateError)
	assert.Equal(t, documentMock, res)
}

func TestService_RemoveCollaborators(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountIDs := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("RemoveCollaborators", accountIDs).
		Return(nil).
		Once()

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.RemoveCollaborators(ctx, documentID, accountIDs)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_RemoveCollaborators_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountIDs := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := context.Background()

	res, err := pendingDocService.RemoveCollaborators(ctx, documentID, accountIDs)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_RemoveCollaborators_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountIDs := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.RemoveCollaborators(ctx, documentID, accountIDs)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_RemoveCollaborators_CollaboratorsRemovalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountIDs := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	removeCollaboratorsError := errors.New("error")

	documentMock.On("RemoveCollaborators", accountIDs).
		Return(removeCollaboratorsError).
		Once()

	res, err := pendingDocService.RemoveCollaborators(ctx, documentID, accountIDs)
	assert.ErrorIs(t, err, removeCollaboratorsError)
	assert.Nil(t, res)
}

func TestService_RemoveCollaborators_RepositoryUpdateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountIDs := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("RemoveCollaborators", accountIDs).
		Return(nil).
		Once()

	repositoryError := errors.New("error")

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(repositoryError).
		Once()

	res, err := pendingDocService.RemoveCollaborators(ctx, documentID, accountIDs)
	assert.ErrorIs(t, err, repositoryError)
	assert.Equal(t, documentMock, res)
}

func TestService_GetRole_PendingDocumentPresent(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	role := &coredocumentpb.Role{}

	documentMock.On("GetRole", roleID).
		Return(role, nil).
		Once()

	res, err := pendingDocService.GetRole(ctx, documentID, roleID)
	assert.NoError(t, err)
	assert.Equal(t, role, res)
}

func TestService_GetRole_PendingDocumentPresent_RoleRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	roleRetrievalError := errors.New("error")

	documentMock.On("GetRole", roleID).
		Return(nil, roleRetrievalError).
		Once()

	res, err := pendingDocService.GetRole(ctx, documentID, roleID)
	assert.ErrorIs(t, err, roleRetrievalError)
	assert.Nil(t, res)
}

func TestService_GetRole_PendingDocumentNotPresent(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("GetCurrentVersion", ctx, documentID).
		Return(documentMock, nil).
		Once()

	role := &coredocumentpb.Role{}

	documentMock.On("GetRole", roleID).
		Return(role, nil).
		Once()

	res, err := pendingDocService.GetRole(ctx, documentID, roleID)
	assert.NoError(t, err)
	assert.Equal(t, role, res)
}

func TestService_GetRole_PendingDocumentNotPresent_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	documentServiceMock.On("GetCurrentVersion", ctx, documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.GetRole(ctx, documentID, roleID)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_GetRole_PendingDocumentNotPresent_RoleRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("GetCurrentVersion", ctx, documentID).
		Return(documentMock, nil).
		Once()

	roleRetrievalError := errors.New("error")

	documentMock.On("GetRole", roleID).
		Return(nil, roleRetrievalError).
		Once()

	res, err := pendingDocService.GetRole(ctx, documentID, roleID)
	assert.ErrorIs(t, err, roleRetrievalError)
	assert.Nil(t, res)
}

func TestService_GetRole_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	ctx := context.Background()

	res, err := pendingDocService.GetRole(ctx, documentID, roleID)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_AddRole(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleKey := "role-keys"

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	role := &coredocumentpb.Role{}

	documentMock.On("AddRole", roleKey, collaborators).
		Return(role, nil).
		Once()

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.AddRole(ctx, documentID, roleKey, collaborators)
	assert.NoError(t, err)
	assert.Equal(t, role, res)
}

func TestService_AddRole_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)
	roleKey := "role-keys"

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := context.Background()

	res, err := pendingDocService.AddRole(ctx, documentID, roleKey, collaborators)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_AddRole_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleKey := "role-keys"

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.AddRole(ctx, documentID, roleKey, collaborators)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_AddRole_RoleAdditionError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleKey := "role-keys"

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	roleAdditionError := errors.New("error")

	documentMock.On("AddRole", roleKey, collaborators).
		Return(nil, roleAdditionError).
		Once()

	res, err := pendingDocService.AddRole(ctx, documentID, roleKey, collaborators)
	assert.ErrorIs(t, err, roleAdditionError)
	assert.Nil(t, res)
}

func TestService_AddRole_RepositoryUpdateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleKey := "role-keys"

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	role := &coredocumentpb.Role{}

	documentMock.On("AddRole", roleKey, collaborators).
		Return(role, nil).
		Once()

	repositoryUpdateError := errors.New("error")

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(repositoryUpdateError).
		Once()

	res, err := pendingDocService.AddRole(ctx, documentID, roleKey, collaborators)
	assert.ErrorIs(t, err, repositoryUpdateError)
	assert.Equal(t, role, res)
}

func TestService_UpdateRole(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	role := &coredocumentpb.Role{}

	documentMock.On("UpdateRole", roleID, collaborators).
		Return(role, nil).
		Once()

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.UpdateRole(ctx, documentID, roleID, collaborators)
	assert.NoError(t, err)
	assert.Equal(t, role, res)
}

func TestService_UpdateRole_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := context.Background()

	res, err := pendingDocService.UpdateRole(ctx, documentID, roleID, collaborators)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_UpdateRole_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.UpdateRole(ctx, documentID, roleID, collaborators)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_UpdateRole_RoleUpdateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	roleUpdateError := errors.New("error")

	documentMock.On("UpdateRole", roleID, collaborators).
		Return(nil, roleUpdateError).
		Once()

	res, err := pendingDocService.UpdateRole(ctx, documentID, roleID, collaborators)
	assert.ErrorIs(t, err, roleUpdateError)
	assert.Nil(t, res)
}

func TestService_UpdateRole_RepositoryUpdateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	roleID := utils.RandomSlice(32)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{
		accountID1,
		accountID2,
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	role := &coredocumentpb.Role{}

	documentMock.On("UpdateRole", roleID, collaborators).
		Return(role, nil).
		Once()

	repositoryUpdateError := errors.New("error")

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(repositoryUpdateError).
		Once()

	res, err := pendingDocService.UpdateRole(ctx, documentID, roleID, collaborators)
	assert.ErrorIs(t, err, repositoryUpdateError)
	assert.Equal(t, role, res)
}

func TestService_AddTransitionRules(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	attributeRule := AttributeRule{
		KeyLabel: "key-label-1",
		RoleID:   utils.RandomSlice(32),
	}

	computeFieldsRule := ComputeFieldsRule{
		WASM: utils.RandomSlice(32),
		AttributeLabels: []string{
			"attribute-label-1",
		},
		TargetAttributeLabel: "target-attribute-label-1",
	}

	documentID := utils.RandomSlice(32)

	addTransitionRules := AddTransitionRules{
		AttributeRules: []AttributeRule{
			attributeRule,
		},
		ComputeFieldsRules: []ComputeFieldsRule{
			computeFieldsRule,
		},
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	attributeKey, err := documents.AttrKeyFromLabel(attributeRule.KeyLabel)
	assert.NoError(t, err)

	transitionRule1 := &coredocumentpb.TransitionRule{}

	documentMock.On("AddTransitionRuleForAttribute", attributeRule.RoleID.Bytes(), attributeKey).
		Return(transitionRule1, nil).
		Once()

	transitionRule2 := &coredocumentpb.TransitionRule{}

	documentMock.On(
		"AddComputeFieldsRule",
		computeFieldsRule.WASM.Bytes(),
		computeFieldsRule.AttributeLabels,
		computeFieldsRule.TargetAttributeLabel,
	).Return(transitionRule2, nil).Once()

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.AddTransitionRules(ctx, documentID, addTransitionRules)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Contains(t, res, transitionRule1)
	assert.Contains(t, res, transitionRule2)
}

func TestService_AddTransitionRules_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	attributeRule := AttributeRule{
		KeyLabel: "key-label-1",
		RoleID:   utils.RandomSlice(32),
	}

	computeFieldsRule := ComputeFieldsRule{
		WASM: utils.RandomSlice(32),
		AttributeLabels: []string{
			"attribute-label-1",
		},
		TargetAttributeLabel: "target-attribute-label-1",
	}

	documentID := utils.RandomSlice(32)

	addTransitionRules := AddTransitionRules{
		AttributeRules: []AttributeRule{
			attributeRule,
		},
		ComputeFieldsRules: []ComputeFieldsRule{
			computeFieldsRule,
		},
	}

	ctx := context.Background()

	res, err := pendingDocService.AddTransitionRules(ctx, documentID, addTransitionRules)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_AddTransitionRules_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	attributeRule := AttributeRule{
		KeyLabel: "key-label-1",
		RoleID:   utils.RandomSlice(32),
	}

	computeFieldsRule := ComputeFieldsRule{
		WASM: utils.RandomSlice(32),
		AttributeLabels: []string{
			"attribute-label-1",
		},
		TargetAttributeLabel: "target-attribute-label-1",
	}

	documentID := utils.RandomSlice(32)

	addTransitionRules := AddTransitionRules{
		AttributeRules: []AttributeRule{
			attributeRule,
		},
		ComputeFieldsRules: []ComputeFieldsRule{
			computeFieldsRule,
		},
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.AddTransitionRules(ctx, documentID, addTransitionRules)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_AddTransitionRules_EmptyKeyLabelError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	attributeRule := AttributeRule{
		// An empty key label should cause an error.
		KeyLabel: "",
		RoleID:   utils.RandomSlice(32),
	}

	computeFieldsRule := ComputeFieldsRule{
		WASM: utils.RandomSlice(32),
		AttributeLabels: []string{
			"attribute-label-1",
		},
		TargetAttributeLabel: "target-attribute-label-1",
	}

	documentID := utils.RandomSlice(32)

	addTransitionRules := AddTransitionRules{
		AttributeRules: []AttributeRule{
			attributeRule,
		},
		ComputeFieldsRules: []ComputeFieldsRule{
			computeFieldsRule,
		},
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	res, err := pendingDocService.AddTransitionRules(ctx, documentID, addTransitionRules)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestService_AddTransitionRules_TransitionRuleAdditionError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	attributeRule := AttributeRule{
		KeyLabel: "key-label-1",
		RoleID:   utils.RandomSlice(32),
	}

	computeFieldsRule := ComputeFieldsRule{
		WASM: utils.RandomSlice(32),
		AttributeLabels: []string{
			"attribute-label-1",
		},
		TargetAttributeLabel: "target-attribute-label-1",
	}

	documentID := utils.RandomSlice(32)

	addTransitionRules := AddTransitionRules{
		AttributeRules: []AttributeRule{
			attributeRule,
		},
		ComputeFieldsRules: []ComputeFieldsRule{
			computeFieldsRule,
		},
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	attributeKey, err := documents.AttrKeyFromLabel(attributeRule.KeyLabel)
	assert.NoError(t, err)

	transitionRuleAdditionError := errors.New("error")

	documentMock.On("AddTransitionRuleForAttribute", attributeRule.RoleID.Bytes(), attributeKey).
		Return(nil, transitionRuleAdditionError).
		Once()

	res, err := pendingDocService.AddTransitionRules(ctx, documentID, addTransitionRules)
	assert.ErrorIs(t, err, transitionRuleAdditionError)
	assert.Nil(t, res)
}

func TestService_AddTransitionRules_ComputeFieldsRuleAdditionError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	attributeRule := AttributeRule{
		KeyLabel: "key-label-1",
		RoleID:   utils.RandomSlice(32),
	}

	computeFieldsRule := ComputeFieldsRule{
		WASM: utils.RandomSlice(32),
		AttributeLabels: []string{
			"attribute-label-1",
		},
		TargetAttributeLabel: "target-attribute-label-1",
	}

	documentID := utils.RandomSlice(32)

	addTransitionRules := AddTransitionRules{
		AttributeRules: []AttributeRule{
			attributeRule,
		},
		ComputeFieldsRules: []ComputeFieldsRule{
			computeFieldsRule,
		},
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	attributeKey, err := documents.AttrKeyFromLabel(attributeRule.KeyLabel)
	assert.NoError(t, err)

	transitionRule1 := &coredocumentpb.TransitionRule{}

	documentMock.On("AddTransitionRuleForAttribute", attributeRule.RoleID.Bytes(), attributeKey).
		Return(transitionRule1, nil).
		Once()

	computeFieldsRuleAdditionError := errors.New("error")

	documentMock.On(
		"AddComputeFieldsRule",
		computeFieldsRule.WASM.Bytes(),
		computeFieldsRule.AttributeLabels,
		computeFieldsRule.TargetAttributeLabel,
	).Return(nil, computeFieldsRuleAdditionError).Once()

	res, err := pendingDocService.AddTransitionRules(ctx, documentID, addTransitionRules)
	assert.ErrorIs(t, err, computeFieldsRuleAdditionError)
	assert.Nil(t, res)
}

func TestService_AddTransitionRules_RepositoryUpdateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	attributeRule := AttributeRule{
		KeyLabel: "key-label-1",
		RoleID:   utils.RandomSlice(32),
	}

	computeFieldsRule := ComputeFieldsRule{
		WASM: utils.RandomSlice(32),
		AttributeLabels: []string{
			"attribute-label-1",
		},
		TargetAttributeLabel: "target-attribute-label-1",
	}

	documentID := utils.RandomSlice(32)

	addTransitionRules := AddTransitionRules{
		AttributeRules: []AttributeRule{
			attributeRule,
		},
		ComputeFieldsRules: []ComputeFieldsRule{
			computeFieldsRule,
		},
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	attributeKey, err := documents.AttrKeyFromLabel(attributeRule.KeyLabel)
	assert.NoError(t, err)

	transitionRule1 := &coredocumentpb.TransitionRule{}

	documentMock.On("AddTransitionRuleForAttribute", attributeRule.RoleID.Bytes(), attributeKey).
		Return(transitionRule1, nil).
		Once()

	transitionRule2 := &coredocumentpb.TransitionRule{}

	documentMock.On(
		"AddComputeFieldsRule",
		computeFieldsRule.WASM.Bytes(),
		computeFieldsRule.AttributeLabels,
		computeFieldsRule.TargetAttributeLabel,
	).Return(transitionRule2, nil).Once()

	repositoryUpdateError := errors.New("error")

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(repositoryUpdateError).
		Once()

	res, err := pendingDocService.AddTransitionRules(ctx, documentID, addTransitionRules)
	assert.ErrorIs(t, err, repositoryUpdateError)
	assert.Len(t, res, 2)
	assert.Contains(t, res, transitionRule1)
	assert.Contains(t, res, transitionRule2)
}

func TestService_GetTransitionRule_PendingDocumentPresent(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	transitionRule := &coredocumentpb.TransitionRule{}

	documentMock.On("GetTransitionRule", ruleID).
		Return(transitionRule, nil).
		Once()

	res, err := pendingDocService.GetTransitionRule(ctx, documentID, ruleID)
	assert.NoError(t, err)
	assert.Equal(t, transitionRule, res)
}

func TestService_GetTransitionRule_PendingDocumentPresent_RuleRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	ruleRetrievalError := errors.New("error")

	documentMock.On("GetTransitionRule", ruleID).
		Return(nil, ruleRetrievalError).
		Once()

	res, err := pendingDocService.GetTransitionRule(ctx, documentID, ruleID)
	assert.ErrorIs(t, err, ruleRetrievalError)
	assert.Nil(t, res)
}

func TestService_GetTransitionRule_PendingDocumentNotPresent(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("GetCurrentVersion", ctx, documentID).
		Return(documentMock, nil).
		Once()

	transitionRule := &coredocumentpb.TransitionRule{}

	documentMock.On("GetTransitionRule", ruleID).
		Return(transitionRule, nil).
		Once()

	res, err := pendingDocService.GetTransitionRule(ctx, documentID, ruleID)
	assert.NoError(t, err)
	assert.Equal(t, transitionRule, res)
}

func TestService_GetTransitionRule_PendingDocumentNotPresent_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	documentServiceMock.On("GetCurrentVersion", ctx, documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.GetTransitionRule(ctx, documentID, ruleID)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_GetTransitionRule_PendingDocumentNotPresent_RuleRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("GetCurrentVersion", ctx, documentID).
		Return(documentMock, nil).
		Once()

	ruleRetrievalError := errors.New("error")

	documentMock.On("GetTransitionRule", ruleID).
		Return(nil, ruleRetrievalError).
		Once()

	res, err := pendingDocService.GetTransitionRule(ctx, documentID, ruleID)
	assert.ErrorIs(t, err, ruleRetrievalError)
	assert.Nil(t, res)
}

func TestService_GetTransitionRule_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := context.Background()

	res, err := pendingDocService.GetTransitionRule(ctx, documentID, ruleID)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_DeleteTransitionRule(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("DeleteTransitionRule", ruleID).
		Return(nil).
		Once()

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	err = pendingDocService.DeleteTransitionRule(ctx, documentID, ruleID)
	assert.NoError(t, err)
}

func TestService_DeleteTransitionRule_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := context.Background()

	err := pendingDocService.DeleteTransitionRule(ctx, documentID, ruleID)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
}

func TestService_DeleteTransitionRule_DocumentNotFoundError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	err = pendingDocService.DeleteTransitionRule(ctx, documentID, ruleID)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
}

func TestService_DeleteTransitionRule_RuleDeletionError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	deletionError := errors.New("error")

	documentMock.On("DeleteTransitionRule", ruleID).
		Return(deletionError).
		Once()

	err = pendingDocService.DeleteTransitionRule(ctx, documentID, ruleID)
	assert.ErrorIs(t, err, deletionError)
}

func TestService_DeleteTransitionRule_RepositoryUpdateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	ruleID := utils.RandomSlice(32)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On("DeleteTransitionRule", ruleID).
		Return(nil).
		Once()

	updateError := errors.New("error")

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(updateError).
		Once()

	err = pendingDocService.DeleteTransitionRule(ctx, documentID, ruleID)
	assert.ErrorIs(t, err, updateError)
}

func TestService_AddAttributes(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	attributes := []documents.Attribute{
		{
			KeyLabel: "key-label-1",
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value: documents.AttrVal{
				Type: documents.AttrString,
				Str:  "test-string-1",
			},
		},
		{
			KeyLabel: "key-label-2",
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value: documents.AttrVal{
				Type: documents.AttrString,
				Str:  "test-string-2",
			},
		},
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On(
		"AddAttributes",
		documents.CollaboratorsAccess{},
		false,
		attributes[0],
		attributes[1],
	).Return(nil).Once()

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.AddAttributes(ctx, documentID, attributes)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_AddAttributes_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)
	attributes := []documents.Attribute{
		{
			KeyLabel: "key-label-1",
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value: documents.AttrVal{
				Type: documents.AttrString,
				Str:  "test-string-1",
			},
		},
		{
			KeyLabel: "key-label-2",
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value: documents.AttrVal{
				Type: documents.AttrString,
				Str:  "test-string-2",
			},
		},
	}

	ctx := context.Background()

	res, err := pendingDocService.AddAttributes(ctx, documentID, attributes)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_AddAttributes_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	attributes := []documents.Attribute{
		{
			KeyLabel: "key-label-1",
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value: documents.AttrVal{
				Type: documents.AttrString,
				Str:  "test-string-1",
			},
		},
		{
			KeyLabel: "key-label-2",
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value: documents.AttrVal{
				Type: documents.AttrString,
				Str:  "test-string-2",
			},
		},
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.AddAttributes(ctx, documentID, attributes)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_AddAttributes_AttributesAdditionError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	attributes := []documents.Attribute{
		{
			KeyLabel: "key-label-1",
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value: documents.AttrVal{
				Type: documents.AttrString,
				Str:  "test-string-1",
			},
		},
		{
			KeyLabel: "key-label-2",
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value: documents.AttrVal{
				Type: documents.AttrString,
				Str:  "test-string-2",
			},
		},
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	attributeAdditionError := errors.New("error")

	documentMock.On(
		"AddAttributes",
		documents.CollaboratorsAccess{},
		false,
		attributes[0],
		attributes[1],
	).Return(attributeAdditionError).Once()

	res, err := pendingDocService.AddAttributes(ctx, documentID, attributes)
	assert.ErrorIs(t, err, attributeAdditionError)
	assert.Nil(t, res)
}

func TestService_AddAttributes_RepositoryUpdateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	attributes := []documents.Attribute{
		{
			KeyLabel: "key-label-1",
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value: documents.AttrVal{
				Type: documents.AttrString,
				Str:  "test-string-1",
			},
		},
		{
			KeyLabel: "key-label-2",
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value: documents.AttrVal{
				Type: documents.AttrString,
				Str:  "test-string-2",
			},
		},
	}

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On(
		"AddAttributes",
		documents.CollaboratorsAccess{},
		false,
		attributes[0],
		attributes[1],
	).Return(nil).Once()

	updateError := errors.New("error")

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(updateError).
		Once()

	res, err := pendingDocService.AddAttributes(ctx, documentID, attributes)
	assert.ErrorIs(t, err, updateError)
	assert.Equal(t, documentMock, res)
}

func TestService_DeleteAttribute(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	attributeKey := documents.AttrKey(utils.RandomByte32())

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On(
		"DeleteAttribute",
		attributeKey,
		false,
	).Return(nil).Once()

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(nil).
		Once()

	res, err := pendingDocService.DeleteAttribute(ctx, documentID, attributeKey)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_DeleteAttribute_IdentityRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	documentID := utils.RandomSlice(32)
	attributeKey := documents.AttrKey(utils.RandomByte32())

	ctx := context.Background()

	res, err := pendingDocService.DeleteAttribute(ctx, documentID, attributeKey)
	assert.ErrorIs(t, err, errors.ErrContextIdentityRetrieval)
	assert.Nil(t, res)
}

func TestService_DeleteAttribute_DocumentRetrievalError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	attributeKey := documents.AttrKey(utils.RandomByte32())

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(nil, errors.New("error")).
		Once()

	res, err := pendingDocService.DeleteAttribute(ctx, documentID, attributeKey)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_DeleteAttribute_AttributeDeletionError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	attributeKey := documents.AttrKey(utils.RandomByte32())

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	attributeDeletionError := errors.New("error")

	documentMock.On(
		"DeleteAttribute",
		attributeKey,
		false,
	).Return(attributeDeletionError).Once()

	res, err := pendingDocService.DeleteAttribute(ctx, documentID, attributeKey)
	assert.ErrorIs(t, err, attributeDeletionError)
	assert.Nil(t, res)
}

func TestService_DeleteAttribute_RepositoryUpdateError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := NewRepositoryMock(t)

	pendingDocService := NewService(documentServiceMock, repositoryMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	documentID := utils.RandomSlice(32)
	attributeKey := documents.AttrKey(utils.RandomByte32())

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentMock := documents.NewDocumentMock(t)

	repositoryMock.On("Get", accountID.ToBytes(), documentID).
		Return(documentMock, nil).
		Once()

	documentMock.On(
		"DeleteAttribute",
		attributeKey,
		false,
	).Return(nil).Once()

	updateError := errors.New("error")

	repositoryMock.On("Update", accountID.ToBytes(), documentID, documentMock).
		Return(updateError).
		Once()

	res, err := pendingDocService.DeleteAttribute(ctx, documentID, attributeKey)
	assert.ErrorIs(t, err, updateError)
	assert.Equal(t, documentMock, res)
}
