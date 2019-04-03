// +build unit

package entityrelationship

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/transactions"
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
	repo := testRepo()
	mockAnchor := &mockAnchorRepo{}
	docSrv := documents.DefaultService(repo, mockAnchor, documents.NewServiceRegistry(), &idService)
	return idService, idFactory, DefaultService(
		docSrv,
		repo,
		queueSrv,
		ctx[transactions.BootstrappedService].(transactions.Manager), idFactory)
}

func TestService_DeriveFromCreatePayload(t *testing.T) {
	eSrv := service{}
	ctxh := testingconfig.CreateAccountContext(t, cfg)

	// nil payload
	m, err := eSrv.DeriveFromCreatePayload(ctxh, nil)
	assert.Nil(t, m)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(documents.ErrDocumentNil, err))

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
	eSrv := service{repo: testRepo()}
	_, cd := createCDWithEmbeddedEntityRelationship(t)
	m, err := eSrv.DeriveFromCoreDocument(cd)
	assert.NoError(t, err, "must return model")
	assert.NotNil(t, m, "model must be non-nil")
	relationship, ok := m.(*EntityRelationship)
	assert.True(t, ok, "must be true")
	assert.Equal(t, relationship.TargetIdentity.String(), "0x5F9132e0F92952abCb154A9b34563891ffe1AAcb")
	assert.Equal(t, relationship.OwnerIdentity.String(), selfDID.String())
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
	assert.True(t, testRepo().Exists(did[:], m.ID()))
	assert.True(t, testRepo().Exists(did[:], m.CurrentVersion()))
	idFactory.AssertExpectations(t)
}

func TestService_DeriveEntityData(t *testing.T) {
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
	eSrv := service{repo: testRepo()}

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
		TargetIdentity: "0x5f9132e0f92952abcb154a9b34563891ffe1aacb",
	}
	assert.Equal(t, payload.TargetIdentity, r.Relationship.TargetIdentity)
	assert.Equal(t, payload.OwnerIdentity, r.Relationship.OwnerIdentity)
}
