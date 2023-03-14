//go:build integration

package entityrelationship

import (
	"context"
	"os"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/bootstrap"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/pod/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/pod/centchain"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/config/configstore"
	"github.com/centrifuge/pod/contextutil"
	protocolIDDispatcher "github.com/centrifuge/pod/dispatcher"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	"github.com/centrifuge/pod/ipfs"
	"github.com/centrifuge/pod/jobs"
	nftv3 "github.com/centrifuge/pod/nft/v3"
	"github.com/centrifuge/pod/p2p"
	"github.com/centrifuge/pod/pallets"
	"github.com/centrifuge/pod/pending"
	"github.com/centrifuge/pod/storage"
	"github.com/centrifuge/pod/storage/leveldb"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	genericUtils "github.com/centrifuge/pod/testingutils/generic"
	jobs2 "github.com/centrifuge/pod/testingutils/jobs"
	"github.com/centrifuge/pod/testingutils/keyrings"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&integration_test.Bootstrapper{},
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	&configstore.Bootstrapper{},
	&jobs.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&protocolIDDispatcher.Bootstrapper{},
	&v2.AccountTestBootstrapper{},
	documents.Bootstrapper{},
	pending.Bootstrapper{},
	&ipfs.TestBootstrapper{},
	&nftv3.Bootstrapper{},
	&p2p.Bootstrapper{},
	documents.PostBootstrapper{},
	Bootstrapper{},
}

var (
	cfgService                config.Service
	docSrv                    documents.Service
	entityRelationshipService Service
	dispatcher                jobs.Dispatcher
	documentsRepository       documents.Repository
	dbRepository              storage.Repository
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)
	cfgService = genericUtils.GetService[config.Service](ctx)
	docSrv = genericUtils.GetService[documents.Service](ctx)
	entityRelationshipService = genericUtils.GetService[Service](ctx)
	dispatcher = genericUtils.GetService[jobs.Dispatcher](ctx)
	documentsRepository = genericUtils.GetService[documents.Repository](ctx)
	dbRepository = genericUtils.GetService[storage.Repository](ctx)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_Service_GetEntityRelationships(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)

	acc := accs[0]

	ownerAccountID := acc.GetIdentity()

	targetAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

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
	repository := newDBRepository(dbRepository, documentsRepository)

	ownerIdentity, err := types.NewAccountID(keyrings.EveKeyRingPair.PublicKey)
	assert.NoError(t, err)

	targetIdentity, err := types.NewAccountID(keyrings.FerdieKeyRingPair.PublicKey)
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

	dbRepository.Register(entityRelationship)

	res, err := repository.FindEntityRelationshipIdentifier(entityIdentifier, ownerIdentity, targetIdentity)
	assert.NoError(t, err)
	assert.Equal(t, res, entityRelationship.ID())
}

func TestIntegration_Repository_FindEntityRelationshipIdentifier_NoStorageResults(t *testing.T) {
	repository := newDBRepository(dbRepository, documentsRepository)

	ownerIdentity, err := types.NewAccountID(keyrings.EveKeyRingPair.PublicKey)
	assert.NoError(t, err)

	targetIdentity, err := types.NewAccountID(keyrings.FerdieKeyRingPair.PublicKey)
	assert.NoError(t, err)

	entityRelationShipIdentifier := utils.RandomSlice(32)

	res, err := repository.FindEntityRelationshipIdentifier(entityRelationShipIdentifier, ownerIdentity, targetIdentity)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestIntegration_Repository_FindEntityRelationshipIdentifier_NoMatchingResults(t *testing.T) {
	repository := newDBRepository(dbRepository, documentsRepository)

	dbRepository.Register(&EntityRelationship{})

	ownerIdentity, err := types.NewAccountID(keyrings.EveKeyRingPair.PublicKey)
	assert.NoError(t, err)

	targetIdentity, err := types.NewAccountID(keyrings.FerdieKeyRingPair.PublicKey)
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

	res, err := repository.FindEntityRelationshipIdentifier(entityIdentifier, ownerIdentity, targetIdentity)
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestIntegration_Repository_ListAllRelationships(t *testing.T) {
	repository := newDBRepository(dbRepository, documentsRepository)

	dbRepository.Register(&EntityRelationship{})

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

	res, err := repository.ListAllRelationships(entityIdentifier, ownerIdentity)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	_, ok := res[string(entityRelationship1.ID())]
	assert.True(t, ok)
	_, ok = res[string(entityRelationship2.ID())]
	assert.True(t, ok)
}

func TestIntegration_Repository_ListAllRelationships_PartialResults(t *testing.T) {
	repository := newDBRepository(dbRepository, documentsRepository)

	repository.Register(&EntityRelationship{})

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

	res, err := repository.ListAllRelationships(entityIdentifier, ownerIdentity)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	_, ok := res[string(entityRelationship1.ID())]
	assert.True(t, ok)
	_, ok = res[string(entityRelationship2.ID())]
	assert.False(t, ok)
}

func TestIntegration_Repository_ListAllRelationships_NoMatchingResults(t *testing.T) {
	repository := newDBRepository(dbRepository, documentsRepository)

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

	res, err := repository.ListAllRelationships(entityIdentifier, ownerIdentity)
	assert.NoError(t, err)
	assert.Nil(t, res)
}

func TestIntegration_Repository_ListAllRelationships_NoStorageResults(t *testing.T) {
	repository := newDBRepository(dbRepository, documentsRepository)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityIdentifier := utils.RandomSlice(32)

	res, err := repository.ListAllRelationships(entityIdentifier, ownerIdentity)
	assert.NoError(t, err)
	assert.Nil(t, res)
}
