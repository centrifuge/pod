//go:build unit
// +build unit

package entityrelationship

import (
	"context"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getServiceWithMockedLayers() (testingcommons.MockIdentityService, *identity.MockFactory, Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(did[:], nil)
	idService := testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	idFactory := new(identity.MockFactory)
	entityRepo := testEntityRepo()
	anchorSrv := &anchors.MockAnchorService{}
	anchorSrv.On("GetAnchorData", mock.Anything).Return(nil, errors.New("missing"))
	docSrv := documents.DefaultService(
		cfg, entityRepo, anchorSrv, documents.NewServiceRegistry(), &idService, nil)
	return idService, idFactory, DefaultService(
		docSrv,
		entityRepo,
		idFactory, anchorSrv)
}

// func TestService_GetEntityRelationships(t *testing.T) {
// 	ctxh := testingconfig.CreateAccountContext(t, cfg)
// 	_, idFactory, srv := getServiceWithMockedLayers()
// 	eSrv := srv.(service)
//
// 	// create
// 	idFactory.On("IdentityExists", mock.Anything).Return(true, nil)
// 	relationship := CreateRelationship(t, ctxh)
//
// 	old, _, _, err := eSrv.Create(ctxh, relationship)
// 	assert.NoError(t, err)
// 	assert.True(t, testEntityRepo().Exists(did[:], old.ID()))
// 	assert.True(t, testEntityRepo().Exists(did[:], old.CurrentVersion()))
//
// 	// get all relationships
// 	r, err := eSrv.GetEntityRelationships(ctxh, relationship.Data.EntityIdentifier)
// 	assert.NoError(t, err)
// 	assert.Len(t, r, 1)
// 	r, err = eSrv.GetEntityRelationships(ctxh, utils.RandomSlice(32))
// 	assert.NoError(t, err)
// 	assert.Equal(t, []documents.Document(nil), r)
// }

func TestService_DeriveFromCoreDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	eSrv := service{repo: testEntityRepo()}
	empty := coredocumentpb.CoreDocument{}
	_, err := eSrv.DeriveFromCoreDocument(empty)
	assert.Error(t, err)

	_, cd := CreateCDWithEmbeddedEntityRelationship(t, ctxh)
	m, err := eSrv.DeriveFromCoreDocument(cd)
	assert.NoError(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	relationship, ok := m.(*EntityRelationship)
	assert.True(t, ok, "must be true")
	assert.Equal(t, relationship.Data.TargetIdentity.String(), "0x5F9132e0F92952abCb154A9b34563891ffe1AAcb")
	assert.Equal(t, relationship.Data.OwnerIdentity.String(), did.String())
}

func TestService_ValidateError(t *testing.T) {
	srv := service{}
	err := srv.Validate(context.Background(), nil, nil)
	assert.Error(t, err)
}
