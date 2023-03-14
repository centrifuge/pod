//go:build unit

package entity

import (
	"context"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	configMocks "github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/documents/entityrelationship"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/pallets/anchors"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestService_DeriveFromCoreDocument(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	entity := getTestEntityProto()

	b, err := proto.Marshal(entity)
	assert.NoError(t, err)

	embeddedData := &anypb.Any{
		TypeUrl: documenttypes.EntityDataTypeUrl,
		Value:   b,
	}

	coreDoc := &coredocumentpb.CoreDocument{
		EmbeddedData: embeddedData,
	}

	res, err := service.DeriveFromCoreDocument(coreDoc)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	// Invalid embedded data

	embeddedData.TypeUrl = ""

	res, err = service.DeriveFromCoreDocument(coreDoc)
	assert.True(t, errors.IsOfType(documents.ErrDocumentUnPackingCoreDocument, err))
	assert.Nil(t, res)
}

func TestService_GetEntityByRelationship(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	tokenIdentifier := utils.RandomSlice(32)
	granter := utils.RandomSlice(32)
	grantee := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)

	granterAccountID, err := types.NewAccountID(granter)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            granter,
		Grantee:            grantee,
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		Signature:          utils.RandomSlice(32),
		Key:                utils.RandomSlice(32),
		DocumentVersion:    utils.RandomSlice(32),
	}

	coreDoc := &documents.CoreDocument{
		Document: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			AccessTokens:       []*coredocumentpb.AccessToken{accessToken},
		},
	}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: documentIdentifier,
		},
	}

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(entityRelationship, nil).
		Once()

	res := &coredocumentpb.CoreDocument{}

	anchorProcessorRes := &p2ppb.GetDocumentResponse{
		Document: res,
	}

	anchorProcessorMock.On(
		"RequestDocumentWithAccessToken",
		ctx,
		granterAccountID,
		tokenIdentifier,
		documentIdentifier,
		documentIdentifier,
	).Return(anchorProcessorRes, nil).Once()

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("DeriveFromCoreDocument", res).
		Return(documentMock, nil).
		Once()

	validatorMock.On("Validate", nil, documentMock).
		Return(nil).
		Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, doc)
}

func TestService_GetEntityByRelationship_EntityRelationshipServiceNotFound(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(nil, errors.New("error")).
		Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.ErrorIs(t, err, entityrelationship.ErrERNotFound)
	assert.Nil(t, doc)
}

func TestService_GetEntityByRelationship_InvalidModel(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	// Invalid model
	entity := &Entity{}

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(entity, nil).
		Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.ErrorIs(t, err, entityrelationship.ErrNotEntityRelationship)
	assert.Nil(t, doc)
}

func TestService_GetEntityByRelationship_NoAccessToken(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	documentIdentifier := utils.RandomSlice(32)

	coreDoc := &documents.CoreDocument{
		Document: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			AccessTokens:       nil, // no token
		},
	}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: documentIdentifier,
		},
	}

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(entityRelationship, nil).
		Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.ErrorIs(t, err, entityrelationship.ErrERNoToken)
	assert.Nil(t, doc)
}

func TestService_GetEntityByRelationship_ErrInvalidIdentifier(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	tokenIdentifier := utils.RandomSlice(32)
	granter := utils.RandomSlice(32)
	grantee := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            granter,
		Grantee:            grantee,
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: utils.RandomSlice(32), // invalid identifier
		Signature:          utils.RandomSlice(32),
		Key:                utils.RandomSlice(32),
		DocumentVersion:    utils.RandomSlice(32),
	}

	coreDoc := &documents.CoreDocument{
		Document: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			AccessTokens:       []*coredocumentpb.AccessToken{accessToken},
		},
	}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: documentIdentifier,
		},
	}

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(entityRelationship, nil).
		Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.ErrorIs(t, err, entityrelationship.ErrERInvalidIdentifier)
	assert.Nil(t, doc)
}

func TestService_GetEntityByRelationship_ErrInvalidGranterAccountID(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	tokenIdentifier := utils.RandomSlice(32)
	granter := []byte("invalid-account-id-bytes")
	grantee := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            granter,
		Grantee:            grantee,
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		Signature:          utils.RandomSlice(32),
		Key:                utils.RandomSlice(32),
		DocumentVersion:    utils.RandomSlice(32),
	}

	coreDoc := &documents.CoreDocument{
		Document: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			AccessTokens:       []*coredocumentpb.AccessToken{accessToken},
		},
	}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: documentIdentifier,
		},
	}

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(entityRelationship, nil).
		Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.ErrorIs(t, err, documents.ErrGranterInvalidAccountID)
	assert.Nil(t, doc)
}

func TestService_GetEntityByRelationship_AnchorProcessorError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	tokenIdentifier := utils.RandomSlice(32)
	granter := utils.RandomSlice(32)
	grantee := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)

	granterAccountID, err := types.NewAccountID(granter)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            granter,
		Grantee:            grantee,
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		Signature:          utils.RandomSlice(32),
		Key:                utils.RandomSlice(32),
		DocumentVersion:    utils.RandomSlice(32),
	}

	coreDoc := &documents.CoreDocument{
		Document: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			AccessTokens:       []*coredocumentpb.AccessToken{accessToken},
		},
	}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: documentIdentifier,
		},
	}

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(entityRelationship, nil).
		Once()

	anchorProcessorMock.On(
		"RequestDocumentWithAccessToken",
		ctx,
		granterAccountID,
		tokenIdentifier,
		documentIdentifier,
		documentIdentifier,
	).Return(nil, errors.New("error")).Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.True(t, errors.IsOfType(ErrP2PDocumentRequest, err))
	assert.Nil(t, doc)
}

func TestService_GetEntityByRelationship_AnchorProcessorInvalidResponse(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	tokenIdentifier := utils.RandomSlice(32)
	granter := utils.RandomSlice(32)
	grantee := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)

	granterAccountID, err := types.NewAccountID(granter)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            granter,
		Grantee:            grantee,
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		Signature:          utils.RandomSlice(32),
		Key:                utils.RandomSlice(32),
		DocumentVersion:    utils.RandomSlice(32),
	}

	coreDoc := &documents.CoreDocument{
		Document: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			AccessTokens:       []*coredocumentpb.AccessToken{accessToken},
		},
	}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: documentIdentifier,
		},
	}

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(entityRelationship, nil).
		Once()

	anchorProcessorMock.On(
		"RequestDocumentWithAccessToken",
		ctx,
		granterAccountID,
		tokenIdentifier,
		documentIdentifier,
		documentIdentifier,
	).Return(nil, nil).Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.ErrorIs(t, err, documents.ErrDocumentInvalid)
	assert.Nil(t, doc)
}

func TestService_GetEntityByRelationship_AnchorProcessorInvalidResponse2(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	tokenIdentifier := utils.RandomSlice(32)
	granter := utils.RandomSlice(32)
	grantee := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)

	granterAccountID, err := types.NewAccountID(granter)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            granter,
		Grantee:            grantee,
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		Signature:          utils.RandomSlice(32),
		Key:                utils.RandomSlice(32),
		DocumentVersion:    utils.RandomSlice(32),
	}

	coreDoc := &documents.CoreDocument{
		Document: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			AccessTokens:       []*coredocumentpb.AccessToken{accessToken},
		},
	}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: documentIdentifier,
		},
	}

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(entityRelationship, nil).
		Once()

	anchorProcessorMock.On(
		"RequestDocumentWithAccessToken",
		ctx,
		granterAccountID,
		tokenIdentifier,
		documentIdentifier,
		documentIdentifier,
	).Return(&p2ppb.GetDocumentResponse{}, nil).Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.ErrorIs(t, err, documents.ErrDocumentInvalid)
	assert.Nil(t, doc)
}

func TestService_GetEntityByRelationship_DeriveError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	tokenIdentifier := utils.RandomSlice(32)
	granter := utils.RandomSlice(32)
	grantee := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)

	granterAccountID, err := types.NewAccountID(granter)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            granter,
		Grantee:            grantee,
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		Signature:          utils.RandomSlice(32),
		Key:                utils.RandomSlice(32),
		DocumentVersion:    utils.RandomSlice(32),
	}

	coreDoc := &documents.CoreDocument{
		Document: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			AccessTokens:       []*coredocumentpb.AccessToken{accessToken},
		},
	}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: documentIdentifier,
		},
	}

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(entityRelationship, nil).
		Once()

	res := &coredocumentpb.CoreDocument{}

	anchorProcessorRes := &p2ppb.GetDocumentResponse{
		Document: res,
	}

	anchorProcessorMock.On(
		"RequestDocumentWithAccessToken",
		ctx,
		granterAccountID,
		tokenIdentifier,
		documentIdentifier,
		documentIdentifier,
	).Return(anchorProcessorRes, nil).Once()

	documentServiceMock.On("DeriveFromCoreDocument", res).
		Return(nil, errors.New("error")).
		Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.True(t, errors.IsOfType(ErrDocumentDerive, err))
	assert.Nil(t, doc)
}

func TestService_GetEntityByRelationship_ValidatorError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	relationshipIdentifier := utils.RandomSlice(32)

	tokenIdentifier := utils.RandomSlice(32)
	granter := utils.RandomSlice(32)
	grantee := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)

	granterAccountID, err := types.NewAccountID(granter)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            granter,
		Grantee:            grantee,
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		Signature:          utils.RandomSlice(32),
		Key:                utils.RandomSlice(32),
		DocumentVersion:    utils.RandomSlice(32),
	}

	coreDoc := &documents.CoreDocument{
		Document: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			AccessTokens:       []*coredocumentpb.AccessToken{accessToken},
		},
	}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: documentIdentifier,
		},
	}

	entityRelationshipMock.On("GetCurrentVersion", ctx, relationshipIdentifier).
		Return(entityRelationship, nil).
		Once()

	res := &coredocumentpb.CoreDocument{}

	anchorProcessorRes := &p2ppb.GetDocumentResponse{
		Document: res,
	}

	anchorProcessorMock.On(
		"RequestDocumentWithAccessToken",
		ctx,
		granterAccountID,
		tokenIdentifier,
		documentIdentifier,
		documentIdentifier,
	).Return(anchorProcessorRes, nil).Once()

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On("DeriveFromCoreDocument", res).
		Return(documentMock, nil).
		Once()

	validatorMock.On("Validate", nil, documentMock).
		Return(errors.New("error")).
		Once()

	doc, err := service.GetEntityByRelationship(ctx, relationshipIdentifier)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))
	assert.Nil(t, doc)
}

func TestService_GetCurrentVersion(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	documentIdentifier := utils.RandomSlice(32)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := configMocks.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On(
		"GetCurrentVersion",
		ctx,
		documentIdentifier,
	).Return(documentMock, nil).Once()

	documentMock.On("IsCollaborator", accountID).
		Return(true, nil).
		Once()

	res, err := service.GetCurrentVersion(ctx, documentIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, documentMock, res)
}

func TestService_GetCurrentVersion_NoAccount(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	documentIdentifier := utils.RandomSlice(32)

	res, err := service.GetCurrentVersion(ctx, documentIdentifier)
	assert.True(t, errors.IsOfType(documents.ErrAccountNotFoundInContext, err))
	assert.Nil(t, res)
}

func TestService_GetCurrentVersion_DocumentNotFound(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	documentIdentifier := utils.RandomSlice(32)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := configMocks.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	documentServiceMock.On(
		"GetCurrentVersion",
		ctx,
		documentIdentifier,
	).Return(nil, errors.New("error")).Once()

	res, err := service.GetCurrentVersion(ctx, documentIdentifier)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestService_GetCurrentVersion_CollaboratorError(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	documentIdentifier := utils.RandomSlice(32)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := configMocks.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On(
		"GetCurrentVersion",
		ctx,
		documentIdentifier,
	).Return(documentMock, nil).Once()

	documentMock.On("IsCollaborator", accountID).
		Return(false, nil).
		Once()

	res, err := service.GetCurrentVersion(ctx, documentIdentifier)
	assert.ErrorIs(t, err, ErrIdentityNotACollaborator)
	assert.Nil(t, res)
}

func TestService_GetCurrentVersion_CollaboratorError2(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()
	documentIdentifier := utils.RandomSlice(32)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := configMocks.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID).
		Once()

	ctx = contextutil.WithAccount(ctx, accountMock)

	documentMock := documents.NewDocumentMock(t)

	documentServiceMock.On(
		"GetCurrentVersion",
		ctx,
		documentIdentifier,
	).Return(documentMock, nil).Once()

	documentMock.On("IsCollaborator", accountID).
		Return(true, errors.New("error")).
		Once()

	res, err := service.GetCurrentVersion(ctx, documentIdentifier)
	assert.ErrorIs(t, err, ErrIdentityNotACollaborator)
	assert.Nil(t, res)
}

func TestService_New(t *testing.T) {
	s := service{}

	doc, err := s.New("")
	assert.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestService_Validate(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	newEntity := &Entity{
		Data: Data{
			Identity: accountID,
		},
	}

	identityServiceMock.On(
		"ValidateAccount",
		accountID,
	).Return(nil).Once()

	err = service.Validate(ctx, newEntity, nil)
	assert.NoError(t, err)
}

func TestService_Validate_Error(t *testing.T) {
	documentServiceMock := documents.NewServiceMock(t)
	documentRepositoryMock := documents.NewRepositoryMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipMock := entityrelationship.NewServiceMock(t)
	anchorServiceMock := anchors.NewAPIMock(t)
	anchorProcessorMock := documents.NewAnchorProcessorMock(t)
	validatorMock := documents.NewValidatorMock(t)

	service := NewService(
		documentServiceMock,
		documentRepositoryMock,
		identityServiceMock,
		entityRelationshipMock,
		anchorServiceMock,
		anchorProcessorMock,
		func() documents.Validator {
			return validatorMock
		},
	)

	ctx := context.Background()

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	newEntity := &Entity{
		Data: Data{
			Identity: accountID,
		},
	}

	identityServiceMock.On(
		"ValidateAccount",
		accountID,
	).Return(errors.New("error")).Once()

	err = service.Validate(ctx, newEntity, nil)
	assert.ErrorIs(t, err, documents.ErrIdentityInvalid)
}
