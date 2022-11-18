//go:build integration

package entity

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

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
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/ipfs"
	"github.com/centrifuge/go-centrifuge/jobs"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/anchors"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	jobs2 "github.com/centrifuge/go-centrifuge/testingutils/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
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
	entityrelationship.Bootstrapper{},
	Bootstrapper{},
}

var (
	entityService    Service
	documentsService documents.Service
	documentsRepo    documents.Repository
	cfgService       config.Service
	dispatcher       jobs.Dispatcher
	anchorSrv        anchors.API
)

const (
	bootstrapAccountTimeout = 10 * time.Minute
)

func TestMain(m *testing.M) {
	serviceCtx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)

	entityService = serviceCtx[BootstrappedEntityService].(Service)
	documentsService = serviceCtx[documents.BootstrappedDocumentService].(documents.Service)
	documentsRepo = serviceCtx[documents.BootstrappedDocumentRepository].(documents.Repository)
	cfgService = serviceCtx[config.BootstrappedConfigStorage].(config.Service)
	dispatcher = serviceCtx[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)
	anchorSrv = serviceCtx[pallets.BootstrappedAnchorService].(anchors.API)

	ctx, cancel := context.WithTimeout(context.Background(), bootstrapAccountTimeout)
	defer cancel()

	if _, err := v2.BootstrapTestAccount(ctx, serviceCtx, keyrings.BobKeyRingPair); err != nil {
		panic(fmt.Errorf("couldn't create an account for Bob: %w", err))
	}

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_Service_GetEntityByRelationship(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	assert.Len(t, accs, 2)

	account1 := accs[0]

	account2 := accs[1]

	entityCoreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{},
		nil,
	)

	entity := &Entity{
		CoreDocument: entityCoreDoc,
		Data: Data{
			Identity: account1.GetIdentity(),
		},
	}

	// Commit the entity using Account 1.
	ctx := contextutil.WithAccount(context.Background(), account1)

	jobID, err := documentsService.Commit(ctx, entity)
	assert.NoError(t, err)
	assert.NotNil(t, jobID)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, account1.GetIdentity(), jobID)
	assert.NoError(t, err)

	// Store the entity for Account 2 as well.

	err = documentsRepo.Create(account2.GetIdentity().ToBytes(), entity.ID(), entity)
	assert.NoError(t, err)

	entityRelationShipCoreDoc, err := documents.NewCoreDocument(
		[]byte{0, 4, 0, 0},
		documents.CollaboratorsAccess{
			ReadCollaborators: []*types.AccountID{account1.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	entityRelationShipCoreDoc.AddUpdateLog(account1.GetIdentity())

	tokenIdentifier := utils.RandomSlice(32)
	roleIdentifier := utils.RandomSlice(32)

	tm, err := documents.AssembleTokenMessage(
		tokenIdentifier,
		account1.GetIdentity(),
		account2.GetIdentity(),
		roleIdentifier,
		entityCoreDoc.ID(),
		entityRelationShipCoreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := account1.SignMsg(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            account1.GetIdentity().ToBytes(),
		Grantee:            account2.GetIdentity().ToBytes(),
		RoleIdentifier:     roleIdentifier,
		DocumentIdentifier: entityCoreDoc.ID(),
		Signature:          signature.GetSignature(),
		Key:                account1.GetSigningPublicKey(),
		DocumentVersion:    entityRelationShipCoreDoc.CurrentVersion(),
	}

	entityRelationShipCoreDoc.Document.AccessTokens = []*coredocumentpb.AccessToken{accessToken}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: entityRelationShipCoreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: entityCoreDoc.ID(),
			OwnerIdentity:    account1.GetIdentity(),
			TargetIdentity:   account2.GetIdentity(),
		},
	}

	// Set status to committed to ensure that the document is also stored under its latest version.
	err = entityRelationship.SetStatus(documents.Committed)
	assert.NoError(t, err)

	err = documentsRepo.Create(account1.GetIdentity().ToBytes(), entityRelationship.ID(), entityRelationship)
	assert.NoError(t, err)

	err = documentsRepo.Create(account2.GetIdentity().ToBytes(), entityRelationship.ID(), entityRelationship)
	assert.NoError(t, err)

	ctx = contextutil.WithAccount(context.Background(), account2)

	res, err := entityService.GetEntityByRelationship(ctx, entityRelationship.ID())
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestIntegration_Service_GetCurrentVersion(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	assert.Len(t, accs, 2)

	account1 := accs[0]
	account2 := accs[1]

	ctx := contextutil.WithAccount(context.Background(), account1)

	coreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{
			ReadCollaborators: []*types.AccountID{account1.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	roleIdentifier := utils.RandomSlice(32)

	tm, err := documents.AssembleTokenMessage(
		tokenIdentifier,
		account1.GetIdentity(),
		account1.GetIdentity(),
		roleIdentifier,
		coreDoc.ID(),
		coreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := account1.SignMsg(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            account1.GetIdentity().ToBytes(),
		Grantee:            account1.GetIdentity().ToBytes(),
		RoleIdentifier:     roleIdentifier,
		DocumentIdentifier: coreDoc.ID(),
		Signature:          signature.GetSignature(),
		Key:                account1.GetSigningPublicKey(),
		DocumentVersion:    coreDoc.CurrentVersion(),
	}

	coreDoc.Document.AccessTokens = []*coredocumentpb.AccessToken{accessToken}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: coreDoc.ID(),
			OwnerIdentity:    account1.GetIdentity(),
			TargetIdentity:   account2.GetIdentity(),
		},
	}

	jobID, err := documentsService.Commit(ctx, entityRelationship)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, account1.GetIdentity(), jobID)
	assert.NoError(t, err)

	res, err := entityService.GetCurrentVersion(ctx, entityRelationship.ID())
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestIntegration_Service_GetCurrentVersion_IdentityNotCollaboratorError(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	assert.Len(t, accs, 2)

	account1 := accs[0]
	account2 := accs[1]

	ctx := contextutil.WithAccount(context.Background(), account1)

	publicKey, privateKey, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	publicKeyRaw, err := publicKey.Raw()
	assert.NoError(t, err)

	coreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{},
		nil,
	)
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	roleIdentifier := utils.RandomSlice(32)

	tm, err := documents.AssembleTokenMessage(
		tokenIdentifier,
		account1.GetIdentity(),
		account1.GetIdentity(),
		roleIdentifier,
		coreDoc.ID(),
		coreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := privateKey.Sign(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            account1.GetIdentity().ToBytes(),
		Grantee:            account1.GetIdentity().ToBytes(),
		RoleIdentifier:     roleIdentifier,
		DocumentIdentifier: coreDoc.ID(),
		Signature:          signature,
		Key:                publicKeyRaw,
		DocumentVersion:    coreDoc.CurrentVersion(),
	}

	coreDoc.Document.AccessTokens = []*coredocumentpb.AccessToken{accessToken}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: coreDoc.ID(),
			OwnerIdentity:    account1.GetIdentity(),
			TargetIdentity:   account2.GetIdentity(),
		},
	}

	jobID, err := documentsService.Commit(ctx, entityRelationship)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, account1.GetIdentity(), jobID)
	assert.NoError(t, err)

	res, err := entityService.GetCurrentVersion(ctx, entityRelationship.ID())
	assert.ErrorIs(t, err, ErrIdentityNotACollaborator)
	assert.Nil(t, res)
}

func TestIntegration_Service_GetCurrentVersion_DocumentNotFound(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	assert.Len(t, accs, 2)

	account1 := accs[0]
	account2 := accs[1]

	ctx := contextutil.WithAccount(context.Background(), account1)

	publicKey, privateKey, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	publicKeyRaw, err := publicKey.Raw()
	assert.NoError(t, err)

	coreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{
			ReadCollaborators: []*types.AccountID{account1.GetIdentity()},
		},
		nil,
	)
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	roleIdentifier := utils.RandomSlice(32)

	tm, err := documents.AssembleTokenMessage(
		tokenIdentifier,
		account1.GetIdentity(),
		account1.GetIdentity(),
		roleIdentifier,
		coreDoc.ID(),
		coreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := privateKey.Sign(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            account1.GetIdentity().ToBytes(),
		Grantee:            account1.GetIdentity().ToBytes(),
		RoleIdentifier:     roleIdentifier,
		DocumentIdentifier: coreDoc.ID(),
		Signature:          signature,
		Key:                publicKeyRaw,
		DocumentVersion:    coreDoc.CurrentVersion(),
	}

	coreDoc.Document.AccessTokens = []*coredocumentpb.AccessToken{accessToken}

	entityRelationship := &entityrelationship.EntityRelationship{
		CoreDocument: coreDoc,
		Data: entityrelationship.Data{
			EntityIdentifier: coreDoc.ID(),
			OwnerIdentity:    account1.GetIdentity(),
			TargetIdentity:   account2.GetIdentity(),
		},
	}

	res, err := entityService.GetCurrentVersion(ctx, entityRelationship.ID())
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestIntegration_Service_Validate(t *testing.T) {
	accs, err := cfgService.GetAccounts()
	assert.NoError(t, err)
	assert.NotEmpty(t, accs)
	assert.Len(t, accs, 2)

	account1 := accs[0]

	// There is an account created for Alice.
	entity := &Entity{
		Data: Data{
			Identity: account1.GetIdentity(),
		},
	}

	ctx := context.Background()

	err = entityService.Validate(ctx, entity, nil)
	assert.NoError(t, err)

	// No account should be present for.
	randomAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	entity = &Entity{
		Data: Data{
			Identity: randomAccountID,
		},
	}

	err = entityService.Validate(ctx, entity, nil)
	assert.ErrorIs(t, err, documents.ErrIdentityInvalid)

	// Nil model.
	err = entityService.Validate(ctx, nil, nil)
	assert.ErrorIs(t, err, documents.ErrDocumentNil)

	// Wrong model.
	entityRelationship := &entityrelationship.EntityRelationship{}

	err = entityService.Validate(ctx, entityRelationship, nil)
	assert.ErrorIs(t, err, documents.ErrDocumentInvalidType)

	// No identity.
	entity = &Entity{
		Data: Data{},
	}

	err = entityService.Validate(ctx, entity, nil)
	assert.ErrorIs(t, err, ErrEntityDataNoIdentity)
}
