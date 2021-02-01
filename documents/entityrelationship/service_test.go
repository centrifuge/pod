// +build unit

package entityrelationship

import (
	"context"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testinganchors "github.com/centrifuge/go-centrifuge/testingutils/anchors"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getServiceWithMockedLayers() (testingcommons.MockIdentityService, *identity.MockFactory, Service) {
	c := &testingconfig.MockConfig{}
	c.On("GetIdentityID").Return(did[:], nil)
	idService := testingcommons.MockIdentityService{}
	idService.On("IsSignedWithPurpose", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil).Once()
	queueSrv := new(testingutils.MockQueue)
	queueSrv.On("EnqueueJob", mock.Anything, mock.Anything).Return(&gocelery.AsyncResult{}, nil)
	idFactory := new(identity.MockFactory)
	entityRepo := testEntityRepo()
	anchorSrv := &testinganchors.MockAnchorService{}
	anchorSrv.On("GetAnchorData", mock.Anything).Return(nil, errors.New("missing"))
	docSrv := documents.DefaultService(
		cfg, entityRepo, anchorSrv, documents.NewServiceRegistry(), &idService, nil, nil, nil)
	return idService, idFactory, DefaultService(
		docSrv,
		entityRepo,
		queueSrv,
		ctx[jobs.BootstrappedService].(jobs.Manager), idFactory, anchorSrv)
}

func TestService_Update(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, idFactory, srv := getServiceWithMockedLayers()
	eSrv := srv.(service)

	// missing last version
	model, _ := CreateCDWithEmbeddedEntityRelationship(t, ctxh)
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
	relationship := CreateRelationship(t, ctxh)

	old, _, _, err := eSrv.Create(ctxh, relationship)
	assert.NoError(t, err)
	assert.True(t, testEntityRepo().Exists(did[:], old.ID()))
	assert.True(t, testEntityRepo().Exists(did[:], old.CurrentVersion()))

	// derive update payload
	m := new(EntityRelationship)
	err = m.revokeRelationship(old.(*EntityRelationship), *relationship.Data.TargetIdentity)
	assert.NoError(t, err)

	updated, _, _, err := eSrv.Update(ctxh, m)
	assert.NoError(t, err)
	assert.Equal(t, updated.PreviousVersion(), old.CurrentVersion())
	assert.True(t, testEntityRepo().Exists(did[:], updated.ID()))
	assert.True(t, testEntityRepo().Exists(did[:], updated.CurrentVersion()))
	assert.True(t, testEntityRepo().Exists(did[:], updated.PreviousVersion()))
}

func TestService_GetEntityRelationships(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, idFactory, srv := getServiceWithMockedLayers()
	eSrv := srv.(service)

	// create
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil)
	relationship := CreateRelationship(t, ctxh)

	old, _, _, err := eSrv.Create(ctxh, relationship)
	assert.NoError(t, err)
	assert.True(t, testEntityRepo().Exists(did[:], old.ID()))
	assert.True(t, testEntityRepo().Exists(did[:], old.CurrentVersion()))

	// get all relationships
	r, err := eSrv.GetEntityRelationships(ctxh, relationship.Data.EntityIdentifier)
	assert.NoError(t, err)
	assert.Len(t, r, 1)
	r, err = eSrv.GetEntityRelationships(ctxh, utils.RandomSlice(32))
	assert.NoError(t, err)
	assert.Equal(t, []documents.Document(nil), r)
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
	relationship := CreateRelationship(t, ctxh)
	m, _, _, err = eSrv.Create(ctxh, relationship)
	assert.NoError(t, err)
	assert.True(t, testEntityRepo().Exists(did[:], m.ID()))
	assert.True(t, testEntityRepo().Exists(did[:], m.CurrentVersion()))
	idFactory.AssertExpectations(t)
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	eSrv := service{repo: testEntityRepo()}
	empty := coredocumentpb.CoreDocument{}
	m, err := eSrv.DeriveFromCoreDocument(empty)
	assert.Error(t, err)

	_, cd := CreateCDWithEmbeddedEntityRelationship(t, ctxh)
	m, err = eSrv.DeriveFromCoreDocument(cd)
	assert.NoError(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	relationship, ok := m.(*EntityRelationship)
	assert.True(t, ok, "must be true")
	assert.Equal(t, relationship.Data.TargetIdentity.String(), "0x5F9132e0F92952abCb154A9b34563891ffe1AAcb")
	assert.Equal(t, relationship.Data.OwnerIdentity.String(), did.String())
}

type mockRepo struct {
	repository
	mock.Mock
}

func (m *mockRepo) Create(acc, id []byte, model documents.Document) error {
	args := m.Called(acc, id, model)
	return args.Error(0)
}

func (m *mockRepo) FindEntityRelationshipIdentifier(entityIdentifier []byte, ownerDID, targetDID identity.DID) ([]byte, error) {
	args := m.Called(entityIdentifier, ownerDID, targetDID)
	d, _ := args.Get(0).([]byte)
	return d, args.Error(1)
}
func TestService_CreateModel(t *testing.T) {
	payload := documents.CreatePayload{}
	srv := service{}

	// invalid data
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, _, err := srv.CreateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// validator failed
	idFactory := new(identity.MockFactory)
	idFactory.On("IdentityExists", mock.Anything).Return(false, nil).Once()
	payload.Data = validData(t, did)
	srv.factory = idFactory
	_, _, err = srv.CreateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// failed to create
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil)
	repo := new(mockRepo)
	repo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("failed to save")).Once()
	srv.repo = repo
	_, _, err = srv.CreateModel(ctxh, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentPersistence, err))

	// success
	srv.repo = testEntityRepo()
	jm := testingjobs.MockJobManager{}
	jm.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), nil)
	srv.jobManager = jm
	m, _, err := srv.CreateModel(ctxh, payload)
	assert.NoError(t, err)
	assert.NotNil(t, m)
	jm.AssertExpectations(t)
	idFactory.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestService_UpdateModel(t *testing.T) {
	payload := documents.UpdatePayload{}
	_, _, gsrv := getServiceWithMockedLayers()
	srv := gsrv.(service)
	ctx := context.Background()

	// invalid payload
	_, _, err := srv.UpdateModel(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// missing relationship
	payload.Data = validData(t, did)
	r := new(mockRepo)
	r.On("FindEntityRelationshipIdentifier", mock.Anything, did, mock.Anything).Return(
		nil, errors.New("failed to find relationship")).Once()
	srv.repo = r
	_, _, err = srv.UpdateModel(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// missing version
	erid := utils.RandomSlice(32)
	r.On("FindEntityRelationshipIdentifier", mock.Anything, did, mock.Anything).Return(erid, nil).Once()
	_, _, err = srv.UpdateModel(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNotFound, err))

	// missing token
	ctx = testingconfig.CreateAccountContext(t, cfg)
	old, _ := CreateCDWithEmbeddedEntityRelationship(t, ctx)
	err = testEntityRepo().Create(did[:], old.ID(), old)
	assert.NoError(t, err)
	r.On("FindEntityRelationshipIdentifier", mock.Anything, did, mock.Anything).Return(old.ID(), nil)
	_, _, err = srv.UpdateModel(ctx, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), documents.ErrAccessTokenNotFound.Error())

	// validation failed
	id := testingidentity.GenerateRandomDID()
	p := documents.AccessTokenParams{
		Grantee:            id.String(),
		DocumentIdentifier: hexutil.Encode(old.ID()),
	}
	cd, err := old.(*EntityRelationship).AddAccessToken(ctx, p)
	assert.NoError(t, err)
	old.(*EntityRelationship).CoreDocument.Document.AccessTokens = cd.Document.AccessTokens
	assert.NoError(t, testEntityRepo().Update(did[:], old.ID(), old))
	idFactory := new(identity.MockFactory)
	idFactory.On("IdentityExists", mock.Anything).Return(false, nil).Once()
	srv.factory = idFactory
	payload.Data = validDataWithTargetDID(t, did, id)
	_, _, err = srv.UpdateModel(ctx, payload)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))

	// create failed
	idFactory.On("IdentityExists", mock.Anything).Return(true, nil)
	r.On("Create", did[:], old.NextVersion(), mock.Anything).Return(errors.New("failed to create")).Once()
	_, _, err = srv.UpdateModel(ctx, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create")

	// success
	r.On("Create", did[:], old.NextVersion(), mock.Anything).Return(nil)
	jm := testingjobs.MockJobManager{}
	jm.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan error), nil)
	srv.jobManager = jm
	_, _, err = srv.UpdateModel(ctx, payload)
	assert.NoError(t, err)
	r.AssertExpectations(t)
	idFactory.AssertExpectations(t)
	jm.AssertExpectations(t)
}

func TestService_ValidateError(t *testing.T) {
	srv := service{}
	err := srv.Validate(context.Background(), nil, nil)
	assert.Error(t, err)
}
