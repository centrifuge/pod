// +build unit

package entityrelationship

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/anchors"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getServiceWithMockedLayers() (testingcommons.MockIdentityService, *testingcommons.MockIdentityFactory, Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(did[:], nil)
	idService := testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)
	idFactory := new(testingcommons.MockIdentityFactory)
	entityRepo := testEntityRepo()
	anchorRepo := &testinganchors.MockAnchorRepo{}
	anchorRepo.On("GetAnchorData", mock.Anything).Return(nil, errors.New("missing"))
	docSrv := documents.DefaultService(cfg, entityRepo, anchorRepo, documents.NewServiceRegistry(), &idService)
	return idService, idFactory, DefaultService(
		docSrv,
		entityRepo,
		queueSrv,
		ctx[jobs.BootstrappedService].(jobs.Manager), idFactory, anchorRepo)
}

func TestService_Update(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, idFactory, srv := getServiceWithMockedLayers()
	eSrv := srv.(service)

	// missing last version
	model, _ := createCDWithEmbeddedEntityRelationship(t)
	_, _, _, err := eSrv.Update(ctxh, model)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))
	assert.NoError(t, testEntityRepo().Create(did[:], model.CurrentVersion(), model))

	// calculate data root fails
	nm := new(mockModel)
	nm.On("ID").Return(model.ID(), nil).Once()
	_, _, _, err = eSrv.Update(ctxh, nm)
	nm.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// create
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil)
	rp := testingdocuments.CreateRelationshipPayload()
	relationship, err := eSrv.DeriveFromCreatePayload(ctxh, rp)
	assert.NoError(t, err)

	old, _, _, err := eSrv.Create(ctxh, relationship)
	assert.NoError(t, err)
	assert.True(t, testEntityRepo().Exists(did[:], old.ID()))
	assert.True(t, testEntityRepo().Exists(did[:], old.CurrentVersion()))

	// derive update payload
	m, err := eSrv.DeriveFromUpdatePayload(ctxh, rp)
	assert.NoError(t, err)

	updated, _, _, err := eSrv.Update(ctxh, m)
	assert.NoError(t, err)
	assert.Equal(t, updated.PreviousVersion(), old.CurrentVersion())
	assert.True(t, testEntityRepo().Exists(did[:], updated.ID()))
	assert.True(t, testEntityRepo().Exists(did[:], updated.CurrentVersion()))
	assert.True(t, testEntityRepo().Exists(did[:], updated.PreviousVersion()))
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, idFactory, srv := getServiceWithMockedLayers()
	eSrv := srv.(service)

	// success
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil)
	rp := testingdocuments.CreateRelationshipPayload()
	relationship, err := eSrv.DeriveFromCreatePayload(ctxh, rp)
	assert.NoError(t, err)

	old, _, _, err := eSrv.Create(ctxh, relationship)
	assert.NoError(t, err)
	assert.True(t, testEntityRepo().Exists(did[:], old.ID()))
	assert.True(t, testEntityRepo().Exists(did[:], old.CurrentVersion()))

	// nil payload
	m, err := eSrv.DeriveFromUpdatePayload(ctxh, nil)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrPayloadNil, err))

	// invalid identity
	payload := &entitypb.RelationshipPayload{
		TargetIdentity: "some random string",
		Identifier:     rp.Identifier,
	}

	m, err = eSrv.DeriveFromUpdatePayload(ctxh, payload)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "malformed address provided")

	// invalid identifier
	payload.Identifier = "random string"
	m, err = eSrv.DeriveFromUpdatePayload(ctxh, payload)
	assert.Nil(t, m)
	assert.Error(t, err)

	// other DID in payload
	payload.Identifier = rp.Identifier
	payload.TargetIdentity = testingidentity.GenerateRandomDID().String()
	m, err = eSrv.DeriveFromUpdatePayload(ctxh, payload)
	assert.Nil(t, m)
	assert.Error(t, err)

	// valid payload
	m, err = eSrv.DeriveFromUpdatePayload(ctxh, rp)
	assert.NoError(t, err)
}

func TestService_GetEntityRelationships(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, idFactory, srv := getServiceWithMockedLayers()
	eSrv := srv.(service)

	// create
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil)
	rp := testingdocuments.CreateRelationshipPayload()
	relationship, err := eSrv.DeriveFromCreatePayload(ctxh, rp)
	assert.NoError(t, err)

	old, _, _, err := eSrv.Create(ctxh, relationship)
	assert.NoError(t, err)
	assert.True(t, testEntityRepo().Exists(did[:], old.ID()))
	assert.True(t, testEntityRepo().Exists(did[:], old.CurrentVersion()))

	// derive update payload
	m, err := eSrv.DeriveFromUpdatePayload(ctxh, rp)
	assert.NoError(t, err)

	updated, _, _, err := eSrv.Update(ctxh, m)
	assert.NoError(t, err)
	assert.True(t, testEntityRepo().Exists(did[:], updated.CurrentVersion()))

	// get all relationships
	entityID, err := hexutil.Decode(rp.Identifier)
	assert.NoError(t, err)
	r, err := eSrv.GetEntityRelationships(ctxh, entityID)
	assert.NoError(t, err)
	assert.Len(t, r, 1)
	r, err = eSrv.GetEntityRelationships(ctxh, utils.RandomSlice(32))
	assert.NoError(t, err)
	assert.Equal(t, []documents.Model(nil), r)
}

func TestService_Create(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, idFactory, srv := getServiceWithMockedLayers()
	eSrv := srv.(service)

	// calculate data root fails
	m, _, _, err := eSrv.Create(ctxh, &mockModel{})
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// success
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil)
	relationship, err := eSrv.DeriveFromCreatePayload(ctxh, testingdocuments.CreateRelationshipPayload())
	assert.NoError(t, err)
	m, _, _, err = eSrv.Create(ctxh, relationship)
	assert.NoError(t, err)
	assert.True(t, testEntityRepo().Exists(did[:], m.ID()))
	assert.True(t, testEntityRepo().Exists(did[:], m.CurrentVersion()))
	idFactory.AssertExpectations(t)
}

func TestService_DeriveFromCreatePayload(t *testing.T) {
	eSrv := service{}
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// nil payload
	m, err := eSrv.DeriveFromCreatePayload(ctxh, nil)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrPayloadNil, err))

	// invalid identity
	docID := hexutil.Encode(utils.RandomSlice(32))
	payload := &entitypb.RelationshipPayload{
		TargetIdentity: "some random string",
		Identifier:     docID,
	}

	m, err = eSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// success
	payload.TargetIdentity = testingidentity.GenerateRandomDID().String()
	m, err = eSrv.DeriveFromCreatePayload(ctxh, payload)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	er := m.(*EntityRelationship)
	assert.Equal(t, er.TargetIdentity.String(), payload.TargetIdentity)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	selfDID, err := contextutil.AccountDID(ctxh)
	assert.NoError(t, err)
	eSrv := service{repo: testEntityRepo()}
	empty := coredocumentpb.CoreDocument{}
	m, err := eSrv.DeriveFromCoreDocument(empty)
	assert.Error(t, err)

	_, cd := createCDWithEmbeddedEntityRelationship(t)
	m, err = eSrv.DeriveFromCoreDocument(cd)
	assert.NoError(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	relationship, ok := m.(*EntityRelationship)
	assert.True(t, ok, "must be true")
	assert.Equal(t, relationship.TargetIdentity.String(), "0x5F9132e0F92952abCb154A9b34563891ffe1AAcb")
	assert.Equal(t, relationship.OwnerIdentity.String(), selfDID.String())
}

func TestService_DeriveEntityRelationshipData(t *testing.T) {
	_, _, eSrv := getServiceWithMockedLayers()

	// some random model
	_, err := eSrv.DeriveEntityRelationshipData(&mockModel{})
	assert.Error(t, err, "Derive must fail")

	// success
	payload := testingdocuments.CreateRelationshipPayload()
	relationship, err := eSrv.DeriveFromCreatePayload(testingconfig.CreateAccountContext(t, cfg), payload)
	assert.NoError(t, err, "must be non nil")
	data, err := eSrv.DeriveEntityRelationshipData(relationship)
	assert.NoError(t, err, "Derive must succeed")
	assert.NotNil(t, data, "data must be non nil")
}

func TestService_DeriveEntityResponse(t *testing.T) {
	// success
	eSrv := service{repo: testEntityRepo()}

	// derive data failed
	m := new(mockModel)
	r, err := eSrv.DeriveEntityRelationshipResponse(m)
	m.AssertExpectations(t)
	assert.Nil(t, r)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	// success
	relationship, _ := createCDWithEmbeddedEntityRelationship(t)
	r, err = eSrv.DeriveEntityRelationshipResponse(relationship)
	assert.NoError(t, err)

	ctxh := testingconfig.CreateAccountContext(t, cfg)
	selfDID, err := contextutil.AccountDID(ctxh)
	assert.NoError(t, err)
	payload := &entitypb.RelationshipData{
		OwnerIdentity:  selfDID.String(),
		TargetIdentity: "0x5F9132e0F92952abCb154A9b34563891ffe1AAcb",
	}
	assert.Equal(t, payload.TargetIdentity, r.Relationship[0].TargetIdentity)
	assert.Equal(t, payload.OwnerIdentity, r.Relationship[0].OwnerIdentity)
}
