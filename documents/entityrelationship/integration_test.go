//go:build integration

package entityrelationship

import (
	"context"
	"os"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	protocolIDDispatcher "github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/ipfs_pinning"
	"github.com/centrifuge/go-centrifuge/jobs"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	jobs2 "github.com/centrifuge/go-centrifuge/testingutils/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&jobs.Bootstrapper{},
	&configstore.Bootstrapper{},
	&integration_test.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&protocolIDDispatcher.Bootstrapper{},
	&v2.Bootstrapper{},
	documents.Bootstrapper{},
	pending.Bootstrapper{},
	&ipfs_pinning.Bootstrapper{},
	&nftv3.Bootstrapper{},
	&p2p.Bootstrapper{},
	documents.PostBootstrapper{},
	Bootstrapper{},
}

var (
	docSrv                    documents.Service
	entityRelationshipService Service
	dispatcher                jobs.Dispatcher
	documentsRepository       documents.Repository
	dbRepository              storage.Repository
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	docSrv = ctx[documents.BootstrappedDocumentService].(documents.Service)
	entityRelationshipService = ctx[BootstrappedEntityRelationshipService].(Service)
	dispatcher = ctx[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)
	documentsRepository = ctx[documents.BootstrappedDocumentRepository].(documents.Repository)
	dbRepository = ctx[storage.BootstrappedDB].(storage.Repository)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_Service_GetEntityRelationships(t *testing.T) {
	ownerAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	targetAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(ownerAccountID)

	accountMock.On("GetPrecommitEnabled").
		Return(false)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	entityID := utils.RandomSlice(32)

	erCd1, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	erCd1.Document.AccessTokens = []*coredocumentpb.AccessToken{
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

	entityRelationship1 := &EntityRelationship{
		CoreDocument: erCd1,
		Data: Data{
			OwnerIdentity:    ownerAccountID,
			EntityIdentifier: entityID,
			TargetIdentity:   targetAccountID,
		},
	}

	jobID, err := docSrv.Commit(ctx, entityRelationship1)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, ownerAccountID, jobID)
	assert.NoError(t, err)

	erCd2, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship2 := &EntityRelationship{
		CoreDocument: erCd2,
		Data: Data{
			OwnerIdentity:    ownerAccountID,
			EntityIdentifier: entityID,
			TargetIdentity:   targetAccountID,
		},
	}

	jobID, err = docSrv.Commit(ctx, entityRelationship2)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, ownerAccountID, jobID)
	assert.NoError(t, err)

	res, err := entityRelationshipService.GetEntityRelationships(ctx, entityID)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, entityRelationship1.ID(), res[0].ID())
}

func TestIntegration_Service_GetEntityRelationships_ListRelationshipsError(t *testing.T) {
	ownerAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(ownerAccountID)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	entityID := utils.RandomSlice(32)

	res, err := entityRelationshipService.GetEntityRelationships(ctx, entityID)
	assert.Nil(t, err)
	assert.Nil(t, res)
}

func TestIntegration_Service_GetEntityRelationships_GetCurrentVersionError(t *testing.T) {
	ownerAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	targetAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(ownerAccountID)

	accountMock.On("GetPrecommitEnabled").
		Return(false)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	entityID := utils.RandomSlice(32)

	erCd1, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	erCd1.Document.AccessTokens = []*coredocumentpb.AccessToken{
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

	entityRelationship1 := &EntityRelationship{
		CoreDocument: erCd1,
		Data: Data{
			OwnerIdentity:    ownerAccountID,
			EntityIdentifier: entityID,
			TargetIdentity:   targetAccountID,
		},
	}

	// Not waiting for the commit to finish in order to ensure that we can't retrieve the current version of the document.
	_, err = docSrv.Commit(ctx, entityRelationship1)
	assert.NoError(t, err)

	res, err := entityRelationshipService.GetEntityRelationships(ctx, entityID)
	assert.True(t, errors.IsOfType(ErrDocumentsStorageRetrieval, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_Validate(t *testing.T) {
	ownerAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	targetAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx := context.Background()

	entityID := utils.RandomSlice(32)

	erCd1, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship1 := &EntityRelationship{
		CoreDocument: erCd1,
		Data: Data{
			OwnerIdentity:    ownerAccountID,
			EntityIdentifier: entityID,
			TargetIdentity:   targetAccountID,
		},
	}

	err = entityRelationshipService.Validate(ctx, entityRelationship1, nil)
	assert.NoError(t, err)
}

func TestIntegration_Service_Validate_OwnerAccountIDError(t *testing.T) {
	ownerAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	targetAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx := context.Background()

	entityID := utils.RandomSlice(32)

	erCd1, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship1 := &EntityRelationship{
		CoreDocument: erCd1,
		Data: Data{
			OwnerIdentity:    ownerAccountID,
			EntityIdentifier: entityID,
			TargetIdentity:   targetAccountID,
		},
	}

	err = entityRelationshipService.Validate(ctx, entityRelationship1, nil)
	assert.True(t, errors.IsOfType(documents.ErrIdentityInvalid, err))
}

func TestIntegration_Service_Validate_TargetAccountIDError(t *testing.T) {
	ownerAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	targetAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	ctx := context.Background()

	entityID := utils.RandomSlice(32)

	erCd1, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship1 := &EntityRelationship{
		CoreDocument: erCd1,
		Data: Data{
			OwnerIdentity:    ownerAccountID,
			EntityIdentifier: entityID,
			TargetIdentity:   targetAccountID,
		},
	}

	err = entityRelationshipService.Validate(ctx, entityRelationship1, nil)
	assert.True(t, errors.IsOfType(documents.ErrIdentityInvalid, err))
}

func TestIntegration_Repository_FindEntityRelationshipIdentifier(t *testing.T) {
	repo := newDBRepository(dbRepository, documentsRepository)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	targetIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityIdentifier := utils.RandomSlice(32)

	cd, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship := &EntityRelationship{
		CoreDocument: cd,
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: entityIdentifier,
			TargetIdentity:   targetIdentity,
		},
	}

	key := documents.GetKey(ownerIdentity.ToBytes(), entityRelationship.ID())

	err = dbRepository.Create(key, entityRelationship)
	assert.NoError(t, err)

	res, err := repo.FindEntityRelationshipIdentifier(entityIdentifier, ownerIdentity, targetIdentity)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestIntegration_Repository_FindEntityRelationshipIdentifier_NoStorageResults(t *testing.T) {
	repo := newDBRepository(dbRepository, documentsRepository)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	targetIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityIdentifier := utils.RandomSlice(32)

	res, err := repo.FindEntityRelationshipIdentifier(entityIdentifier, ownerIdentity, targetIdentity)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestIntegration_Repository_FindEntityRelationshipIdentifier_NoMatchingResults(t *testing.T) {
	repo := newDBRepository(dbRepository, documentsRepository)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	targetIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityIdentifier := utils.RandomSlice(32)

	cd, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship1 := &EntityRelationship{
		CoreDocument: cd,
		Data: Data{
			OwnerIdentity: ownerIdentity,
			// Use a random identifier to ensure that no matching results are found.
			EntityIdentifier: utils.RandomSlice(32),
			TargetIdentity:   targetIdentity,
		},
	}

	key := documents.GetKey(ownerIdentity.ToBytes(), entityRelationship1.ID())

	err = dbRepository.Create(key, entityRelationship1)
	assert.NoError(t, err)

	cd, err = documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityRelationship2 := &EntityRelationship{
		CoreDocument: cd,
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: entityIdentifier,
			// Use a random account ID to ensure that no matching results are found.
			TargetIdentity: randomAccountID,
		},
	}

	key = documents.GetKey(ownerIdentity.ToBytes(), entityRelationship2.ID())

	err = dbRepository.Create(key, entityRelationship2)
	assert.NoError(t, err)

	res, err := repo.FindEntityRelationshipIdentifier(entityIdentifier, ownerIdentity, targetIdentity)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestIntegration_Repository_ListAllRelationships(t *testing.T) {
	repo := newDBRepository(dbRepository, documentsRepository)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	targetIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityIdentifier := utils.RandomSlice(32)

	cd, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship1 := &EntityRelationship{
		CoreDocument: cd,
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: entityIdentifier,
			TargetIdentity:   targetIdentity,
		},
	}

	key := documents.GetKey(ownerIdentity.ToBytes(), entityRelationship1.ID())

	err = dbRepository.Create(key, entityRelationship1)
	assert.NoError(t, err)

	cd, err = documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship2 := &EntityRelationship{
		CoreDocument: cd,
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: entityIdentifier,
			TargetIdentity:   targetIdentity,
		},
	}

	key = documents.GetKey(ownerIdentity.ToBytes(), entityRelationship2.ID())

	err = dbRepository.Create(key, entityRelationship2)
	assert.NoError(t, err)

	res, err := repo.ListAllRelationships(entityIdentifier, ownerIdentity)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	_, ok := res[string(entityRelationship1.ID())]
	assert.True(t, ok)
	_, ok = res[string(entityRelationship2.ID())]
	assert.True(t, ok)
}

func TestIntegration_Repository_ListAllRelationships_PartialResults(t *testing.T) {
	repo := newDBRepository(dbRepository, documentsRepository)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	targetIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityIdentifier := utils.RandomSlice(32)

	cd, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship1 := &EntityRelationship{
		CoreDocument: cd,
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: entityIdentifier,
			TargetIdentity:   targetIdentity,
		},
	}

	key := documents.GetKey(ownerIdentity.ToBytes(), entityRelationship1.ID())

	err = dbRepository.Create(key, entityRelationship1)
	assert.NoError(t, err)

	cd, err = documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship2 := &EntityRelationship{
		CoreDocument: cd,
		Data: Data{
			OwnerIdentity: ownerIdentity,
			// Use different entity identifier to ensure partial results.
			EntityIdentifier: utils.RandomSlice(32),
			TargetIdentity:   targetIdentity,
		},
	}

	key = documents.GetKey(ownerIdentity.ToBytes(), entityRelationship2.ID())

	err = dbRepository.Create(key, entityRelationship2)
	assert.NoError(t, err)

	res, err := repo.ListAllRelationships(entityIdentifier, ownerIdentity)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	_, ok := res[string(entityRelationship1.ID())]
	assert.True(t, ok)
	_, ok = res[string(entityRelationship2.ID())]
	assert.False(t, ok)
}

func TestIntegration_Repository_ListAllRelationships_NoMatchingResults(t *testing.T) {
	repo := newDBRepository(dbRepository, documentsRepository)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	targetIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityIdentifier := utils.RandomSlice(32)

	cd, err := documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship1 := &EntityRelationship{
		CoreDocument: cd,
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: utils.RandomSlice(32),
			TargetIdentity:   targetIdentity,
		},
	}

	key := documents.GetKey(ownerIdentity.ToBytes(), entityRelationship1.ID())

	err = dbRepository.Create(key, entityRelationship1)
	assert.NoError(t, err)

	cd, err = documents.NewCoreDocument(compactPrefix(), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	entityRelationship2 := &EntityRelationship{
		CoreDocument: cd,
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: utils.RandomSlice(32),
			TargetIdentity:   targetIdentity,
		},
	}

	key = documents.GetKey(ownerIdentity.ToBytes(), entityRelationship2.ID())

	err = dbRepository.Create(key, entityRelationship2)
	assert.NoError(t, err)

	res, err := repo.ListAllRelationships(entityIdentifier, ownerIdentity)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func TestIntegration_Repository_ListAllRelationships_NoStorageResults(t *testing.T) {
	repo := newDBRepository(dbRepository, documentsRepository)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityIdentifier := utils.RandomSlice(32)

	res, err := repo.ListAllRelationships(entityIdentifier, ownerIdentity)
	assert.NoError(t, err)
	assert.Nil(t, res)
}
