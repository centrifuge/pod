//go:build unit

package entityrelationship

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/errors"

	"google.golang.org/protobuf/proto"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestService_DeriveFromCoreDocument(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	entityRelationshippb := getTestEntityRelationshipProto()

	b, err := proto.Marshal(entityRelationshippb)
	assert.NoError(t, err)

	cd := &coredocumentpb.CoreDocument{
		EmbeddedData: &anypb.Any{
			TypeUrl: documenttypes.EntityRelationshipDataTypeUrl,
			Value:   b,
		},
	}

	res, err := service.DeriveFromCoreDocument(cd)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	// Invalid core document
	res, err = service.DeriveFromCoreDocument(&coredocumentpb.CoreDocument{})
	assert.True(t, errors.IsOfType(documents.ErrDocumentUnPackingCoreDocument, err))
	assert.Nil(t, res)
}

func TestService_GetEntityRelationships(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID).
		Once()

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	entityID := utils.RandomSlice(32)

	documentID1 := utils.RandomSlice(32)
	documentID2 := utils.RandomSlice(32)

	relationships := map[string][]byte{
		string(documentID1): documentID1,
		string(documentID2): documentID2,
	}

	repositoryMock.On("ListAllRelationships", entityID, accountID).
		Return(relationships, nil).
		Once()

	accessTokens1 := []*coredocumentpb.AccessToken{
		{
			Identifier:         utils.RandomSlice(32),
			Granter:            utils.RandomSlice(32),
			Grantee:            utils.RandomSlice(32),
			RoleIdentifier:     utils.RandomSlice(32),
			DocumentIdentifier: utils.RandomSlice(32),
			Signature:          utils.RandomSlice(32),
			Key:                utils.RandomSlice(32),
			DocumentVersion:    utils.RandomSlice(32),
		},
	}

	accessTokens2 := []*coredocumentpb.AccessToken{
		{
			Identifier:         utils.RandomSlice(32),
			Granter:            utils.RandomSlice(32),
			Grantee:            utils.RandomSlice(32),
			RoleIdentifier:     utils.RandomSlice(32),
			DocumentIdentifier: utils.RandomSlice(32),
			Signature:          utils.RandomSlice(32),
			Key:                utils.RandomSlice(32),
			DocumentVersion:    utils.RandomSlice(32),
		},
	}

	documentMock1 := documents.NewDocumentMock(t)
	documentMock1.On("GetAccessTokens").
		Return(accessTokens1).
		Once()

	documentMock2 := documents.NewDocumentMock(t)
	documentMock2.On("GetAccessTokens").
		Return(accessTokens2).
		Once()

	documentServiceMock.On("GetCurrentVersion", ctx, documentID1).
		Return(documentMock1, nil).
		Once()

	documentServiceMock.On("GetCurrentVersion", ctx, documentID2).
		Return(documentMock2, nil).
		Once()

	res, err := service.GetEntityRelationships(ctx, entityID)
	assert.NoError(t, err)
	assert.Contains(t, res, documentMock1)
	assert.Contains(t, res, documentMock2)
}

func TestService_GetEntityRelationships_NoResults(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID).
		Once()

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	entityID := utils.RandomSlice(32)

	documentID1 := utils.RandomSlice(32)
	documentID2 := utils.RandomSlice(32)

	relationships := map[string][]byte{
		string(documentID1): documentID1,
		string(documentID2): documentID2,
	}

	repositoryMock.On("ListAllRelationships", entityID, accountID).
		Return(relationships, nil).
		Once()

	documentMock1 := documents.NewDocumentMock(t)
	documentMock1.On("GetAccessTokens").
		Return(nil).
		Once()

	documentMock2 := documents.NewDocumentMock(t)
	documentMock2.On("GetAccessTokens").
		Return(nil).
		Once()

	documentServiceMock.On("GetCurrentVersion", ctx, documentID1).
		Return(documentMock1, nil).
		Once()

	documentServiceMock.On("GetCurrentVersion", ctx, documentID2).
		Return(documentMock2, nil).
		Once()

	res, err := service.GetEntityRelationships(ctx, entityID)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func TestService_GetEntityRelationships_PartialResults(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID).
		Once()

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	entityID := utils.RandomSlice(32)

	documentID1 := utils.RandomSlice(32)
	documentID2 := utils.RandomSlice(32)

	relationships := map[string][]byte{
		string(documentID1): documentID1,
		string(documentID2): documentID2,
	}

	repositoryMock.On("ListAllRelationships", entityID, accountID).
		Return(relationships, nil).
		Once()

	accessTokens1 := []*coredocumentpb.AccessToken{
		{
			Identifier:         utils.RandomSlice(32),
			Granter:            utils.RandomSlice(32),
			Grantee:            utils.RandomSlice(32),
			RoleIdentifier:     utils.RandomSlice(32),
			DocumentIdentifier: utils.RandomSlice(32),
			Signature:          utils.RandomSlice(32),
			Key:                utils.RandomSlice(32),
			DocumentVersion:    utils.RandomSlice(32),
		},
	}

	documentMock1 := documents.NewDocumentMock(t)
	documentMock1.On("GetAccessTokens").
		Return(accessTokens1).
		Once()

	documentMock2 := documents.NewDocumentMock(t)
	documentMock2.On("GetAccessTokens").
		Return(nil). // no access tokens, this should be omitted.
		Once()

	documentServiceMock.On("GetCurrentVersion", ctx, documentID1).
		Return(documentMock1, nil).
		Once()

	documentServiceMock.On("GetCurrentVersion", ctx, documentID2).
		Return(documentMock2, nil).
		Once()

	res, err := service.GetEntityRelationships(ctx, entityID)
	assert.NoError(t, err)
	assert.Contains(t, res, documentMock1)
	assert.NotContains(t, res, documentMock2)
}

func TestService_GetEntityRelationships_EntityIDNil(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	res, err := service.GetEntityRelationships(context.Background(), nil)
	assert.ErrorIs(t, err, ErrEntityIDNil)
	assert.Nil(t, res)
}

func TestService_GetEntityRelationships_MissingAccount(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	res, err := service.GetEntityRelationships(context.Background(), utils.RandomSlice(32))
	assert.ErrorIs(t, err, documents.ErrAccountNotFoundInContext)
	assert.Nil(t, res)
}

func TestService_GetEntityRelationships_RepoError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID).
		Once()

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	entityID := utils.RandomSlice(32)

	repositoryMock.On("ListAllRelationships", entityID, accountID).
		Return(nil, errors.New("error")).
		Once()

	res, err := service.GetEntityRelationships(ctx, entityID)
	assert.True(t, errors.IsOfType(ErrRelationshipsStorageRetrieval, err))
	assert.Nil(t, res)
}

func TestService_GetEntityRelationships_DocumentServiceError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID).
		Once()

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	entityID := utils.RandomSlice(32)

	documentID1 := utils.RandomSlice(32)
	documentID2 := utils.RandomSlice(32)

	relationships := map[string][]byte{
		string(documentID1): documentID1,
		string(documentID2): documentID2,
	}

	repositoryMock.On("ListAllRelationships", entityID, accountID).
		Return(relationships, nil).
		Once()

	accessTokens1 := []*coredocumentpb.AccessToken{
		{
			Identifier:         utils.RandomSlice(32),
			Granter:            utils.RandomSlice(32),
			Grantee:            utils.RandomSlice(32),
			RoleIdentifier:     utils.RandomSlice(32),
			DocumentIdentifier: utils.RandomSlice(32),
			Signature:          utils.RandomSlice(32),
			Key:                utils.RandomSlice(32),
			DocumentVersion:    utils.RandomSlice(32),
		},
	}

	documentMock1 := documents.NewDocumentMock(t)
	documentMock1.On("GetAccessTokens").
		Return(accessTokens1).
		Once()
	documentServiceMock.On("GetCurrentVersion", ctx, documentID1).
		Return(documentMock1, nil).
		Once()

	documentServiceMock.On("GetCurrentVersion", ctx, documentID2).
		Return(nil, errors.New("error")).
		Once()

	res, err := service.GetEntityRelationships(ctx, entityID)
	assert.True(t, errors.IsOfType(ErrDocumentsStorageRetrieval, err))
	assert.Nil(t, res, documentMock1)
}

func TestService_New(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	res, err := service.New("")
	assert.NoError(t, err)
	assert.IsType(t, &EntityRelationship{}, res)
}

func TestService_Validate(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	entityRelationship := getTestEntityRelationship(t, documents.CollaboratorsAccess{}, nil)

	ctx := context.Background()

	identityServiceMock.On("ValidateAccount", ctx, entityRelationship.Data.OwnerIdentity).
		Return(nil).
		Once()

	identityServiceMock.On("ValidateAccount", ctx, entityRelationship.Data.TargetIdentity).
		Return(nil).
		Once()

	err := service.Validate(ctx, entityRelationship, nil)
	assert.NoError(t, err)
}

func TestService_Validate_NilDocument(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	ctx := context.Background()

	err := service.Validate(ctx, nil, nil)
	assert.ErrorIs(t, err, documents.ErrDocumentNil)
}

func TestService_Validate_InvalidDocType(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	ctx := context.Background()

	documentMock := documents.NewDocumentMock(t)

	err := service.Validate(ctx, documentMock, nil)
	assert.ErrorIs(t, err, documents.ErrDocumentInvalidType)
}

func TestService_Validate_NilOwner(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	entityRelationship := getTestEntityRelationship(t, documents.CollaboratorsAccess{}, nil)
	entityRelationship.Data.OwnerIdentity = nil

	ctx := context.Background()

	err := service.Validate(ctx, entityRelationship, nil)
	assert.True(t, errors.IsOfType(documents.ErrIdentityInvalid, err))
}

func TestService_Validate_NilTarget(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	entityRelationship := getTestEntityRelationship(t, documents.CollaboratorsAccess{}, nil)
	entityRelationship.Data.TargetIdentity = nil

	ctx := context.Background()

	err := service.Validate(ctx, entityRelationship, nil)
	assert.True(t, errors.IsOfType(documents.ErrIdentityInvalid, err))
}

func TestService_Validate_InvalidAccount(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	repositoryMock := newRepositoryMock(t)
	anchorServiceMock := anchors.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	service := NewService(documentServiceMock, repositoryMock, anchorServiceMock, identityServiceMock)

	entityRelationship := getTestEntityRelationship(t, documents.CollaboratorsAccess{}, nil)

	ctx := context.Background()

	identityServiceMock.On("ValidateAccount", ctx, entityRelationship.Data.OwnerIdentity).
		Return(errors.New("error")).
		Once()

	err := service.Validate(ctx, entityRelationship, nil)
	assert.True(t, errors.IsOfType(documents.ErrIdentityInvalid, err))
}
