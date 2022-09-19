//go:build integration

package entity

import (
	"context"
	"github.com/centrifuge/go-centrifuge/pallets"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"

	jobs2 "github.com/centrifuge/go-centrifuge/testingutils/jobs"

	"github.com/centrifuge/go-centrifuge/contextutil"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"

	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/integration_test"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	protocolIDDispatcher "github.com/centrifuge/go-centrifuge/dispatcher"
	"github.com/centrifuge/go-centrifuge/documents"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/ipfs_pinning"
	"github.com/centrifuge/go-centrifuge/jobs"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
)

var integrationTestBootstrappers = []bootstrap.TestBootstrapper{
	&testlogging.TestLoggingBootstrapper{},
	&config.Bootstrapper{},
	&leveldb.Bootstrapper{},
	jobs.Bootstrapper{},
	&configstore.Bootstrapper{},
	&integration_test.Bootstrapper{},
	centchain.Bootstrapper{},
	&pallets.Bootstrapper{},
	&protocolIDDispatcher.Bootstrapper{},
	&v2.Bootstrapper{},
	anchors.Bootstrapper{},
	documents.Bootstrapper{},
	pending.Bootstrapper{},
	&ipfs_pinning.Bootstrapper{},
	&nftv3.Bootstrapper{},
	p2p.Bootstrapper{},
	documents.PostBootstrapper{},
	entityrelationship.Bootstrapper{},
	Bootstrapper{},
}

var (
	entityService    Service
	documentsService documents.Service
	cfgService       config.Service
	dispatcher       jobs.Dispatcher
	anchorSrv        anchors.Service
)

func TestMain(m *testing.M) {
	ctx := bootstrap.RunTestBootstrappers(integrationTestBootstrappers, nil)

	entityService = ctx[BootstrappedEntityService].(Service)
	documentsService = ctx[documents.BootstrappedDocumentService].(documents.Service)
	cfgService = ctx[config.BootstrappedConfigStorage].(config.Service)
	dispatcher = ctx[jobs.BootstrappedJobDispatcher].(jobs.Dispatcher)
	anchorSrv = ctx[anchors.BootstrappedAnchorService].(anchors.Service)

	result := m.Run()

	bootstrap.RunTestTeardown(integrationTestBootstrappers)

	os.Exit(result)
}

func TestIntegration_Service_GetEntityByRelationship(t *testing.T) {
	aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	bobAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	acc, err := cfgService.GetAccount(aliceAccountID.ToBytes())
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	publicKey, privateKey, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	publicKeyRaw, err := publicKey.Raw()
	assert.NoError(t, err)

	coreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{
			ReadCollaborators: []*types.AccountID{aliceAccountID},
		},
		nil,
	)
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	roleIdentifier := utils.RandomSlice(32)

	tm, err := documents.AssembleTokenMessage(
		tokenIdentifier,
		aliceAccountID,
		aliceAccountID,
		roleIdentifier,
		coreDoc.ID(),
		coreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := privateKey.Sign(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            aliceAccountID.ToBytes(),
		Grantee:            aliceAccountID.ToBytes(),
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
			OwnerIdentity:    aliceAccountID,
			TargetIdentity:   bobAccountID,
		},
	}

	jobID, err := documentsService.Commit(ctx, entityRelationship)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	res, err := entityService.GetEntityByRelationship(ctx, entityRelationship.ID())
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestIntegration_Service_GetEntityByRelationship_DocumentNotStored(t *testing.T) {
	aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	bobAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	acc, err := cfgService.GetAccount(aliceAccountID.ToBytes())
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	publicKey, privateKey, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	publicKeyRaw, err := publicKey.Raw()
	assert.NoError(t, err)

	coreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{
			ReadCollaborators: []*types.AccountID{aliceAccountID},
		},
		nil,
	)
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	roleIdentifier := utils.RandomSlice(32)

	tm, err := documents.AssembleTokenMessage(
		tokenIdentifier,
		aliceAccountID,
		aliceAccountID,
		roleIdentifier,
		coreDoc.ID(),
		coreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := privateKey.Sign(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            aliceAccountID.ToBytes(),
		Grantee:            aliceAccountID.ToBytes(),
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
			OwnerIdentity:    aliceAccountID,
			TargetIdentity:   bobAccountID,
		},
	}

	// Not committing the document before this call to ensure that it's not in storage.

	res, err := entityService.GetEntityByRelationship(ctx, entityRelationship.ID())
	assert.ErrorIs(t, err, entityrelationship.ErrERNotFound)
	assert.Nil(t, res)
}

func TestIntegration_Service_GetEntityByRelationship_InvalidDocumentStored(t *testing.T) {
	aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	acc, err := cfgService.GetAccount(aliceAccountID.ToBytes())
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	coreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{},
		nil,
	)
	assert.NoError(t, err)

	entity := &Entity{
		CoreDocument: coreDoc,
		Data: Data{
			Identity: aliceAccountID,
		},
	}

	jobID, err := documentsService.Commit(ctx, entity)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	res, err := entityService.GetEntityByRelationship(ctx, entity.ID())
	assert.ErrorIs(t, err, entityrelationship.ErrNotEntityRelationship)
	assert.Nil(t, res)
}

func TestIntegration_Service_GetEntityByRelationship_AnchorProcessorError(t *testing.T) {
	aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	bobAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	acc, err := cfgService.GetAccount(aliceAccountID.ToBytes())
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	publicKey, privateKey, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	publicKeyRaw, err := publicKey.Raw()
	assert.NoError(t, err)

	coreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{
			ReadCollaborators: []*types.AccountID{aliceAccountID},
		},
		nil,
	)
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	roleIdentifier := utils.RandomSlice(32)

	tm, err := documents.AssembleTokenMessage(
		tokenIdentifier,
		aliceAccountID,
		aliceAccountID,
		roleIdentifier,
		coreDoc.ID(),
		coreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := privateKey.Sign(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier: tokenIdentifier,
		// Random granter ID to ensure failure during validation.
		Granter:            utils.RandomSlice(32),
		Grantee:            aliceAccountID.ToBytes(),
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
			OwnerIdentity:    aliceAccountID,
			TargetIdentity:   bobAccountID,
		},
	}

	jobID, err := documentsService.Commit(ctx, entityRelationship)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	res, err := entityService.GetEntityByRelationship(ctx, entityRelationship.ID())
	assert.True(t, errors.IsOfType(ErrP2PDocumentRequest, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_GetEntityByRelationship_ValidationError(t *testing.T) {
	aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	bobAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	acc, err := cfgService.GetAccount(aliceAccountID.ToBytes())
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	publicKey, privateKey, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	publicKeyRaw, err := publicKey.Raw()
	assert.NoError(t, err)

	coreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{
			ReadCollaborators: []*types.AccountID{aliceAccountID},
		},
		nil,
	)
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	roleIdentifier := utils.RandomSlice(32)

	tm, err := documents.AssembleTokenMessage(
		tokenIdentifier,
		aliceAccountID,
		aliceAccountID,
		roleIdentifier,
		coreDoc.ID(),
		coreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := privateKey.Sign(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            aliceAccountID.ToBytes(),
		Grantee:            aliceAccountID.ToBytes(),
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
			OwnerIdentity:    aliceAccountID,
			TargetIdentity:   bobAccountID,
		},
	}

	jobID, err := documentsService.Commit(ctx, entityRelationship)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	// Commit an anchor with the next preimage to ensure that the PostAnchoredValidator fails.
	anchorID, err := anchors.ToAnchorID(entityRelationship.NextPreimage())
	assert.NoError(t, err)

	docRoot, err := entityRelationship.CalculateDocumentRoot()
	assert.NoError(t, err)

	anchorRoot, err := anchors.ToDocumentRoot(docRoot)
	assert.NoError(t, err)

	err = anchorSrv.CommitAnchor(ctx, anchorID, anchorRoot, utils.RandomByte32())
	assert.NoError(t, err)

	res, err := entityService.GetEntityByRelationship(ctx, entityRelationship.ID())
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalid, err))
	assert.Nil(t, res)
}

func TestIntegration_Service_GetCurrentVersion(t *testing.T) {
	aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	bobAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	acc, err := cfgService.GetAccount(aliceAccountID.ToBytes())
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	publicKey, privateKey, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	publicKeyRaw, err := publicKey.Raw()
	assert.NoError(t, err)

	coreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{
			ReadCollaborators: []*types.AccountID{aliceAccountID},
		},
		nil,
	)
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	roleIdentifier := utils.RandomSlice(32)

	tm, err := documents.AssembleTokenMessage(
		tokenIdentifier,
		aliceAccountID,
		aliceAccountID,
		roleIdentifier,
		coreDoc.ID(),
		coreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := privateKey.Sign(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            aliceAccountID.ToBytes(),
		Grantee:            aliceAccountID.ToBytes(),
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
			OwnerIdentity:    aliceAccountID,
			TargetIdentity:   bobAccountID,
		},
	}

	jobID, err := documentsService.Commit(ctx, entityRelationship)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	res, err := entityService.GetCurrentVersion(ctx, entityRelationship.ID())
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestIntegration_Service_GetCurrentVersion_IdentityNotCollaboratorError(t *testing.T) {
	aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	bobAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	acc, err := cfgService.GetAccount(aliceAccountID.ToBytes())
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

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
		aliceAccountID,
		aliceAccountID,
		roleIdentifier,
		coreDoc.ID(),
		coreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := privateKey.Sign(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            aliceAccountID.ToBytes(),
		Grantee:            aliceAccountID.ToBytes(),
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
			OwnerIdentity:    aliceAccountID,
			TargetIdentity:   bobAccountID,
		},
	}

	jobID, err := documentsService.Commit(ctx, entityRelationship)
	assert.NoError(t, err)

	err = jobs2.WaitForJobToFinish(ctx, dispatcher, acc.GetIdentity(), jobID)
	assert.NoError(t, err)

	res, err := entityService.GetCurrentVersion(ctx, entityRelationship.ID())
	assert.ErrorIs(t, err, ErrIdentityNotACollaborator)
	assert.Nil(t, res)
}

func TestIntegration_Service_GetCurrentVersion_DocumentNotFound(t *testing.T) {
	aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	bobAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	acc, err := cfgService.GetAccount(aliceAccountID.ToBytes())
	assert.NoError(t, err)

	ctx := contextutil.WithAccount(context.Background(), acc)

	publicKey, privateKey, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	publicKeyRaw, err := publicKey.Raw()
	assert.NoError(t, err)

	coreDoc, err := documents.NewCoreDocument(
		compactPrefix(),
		documents.CollaboratorsAccess{
			ReadCollaborators: []*types.AccountID{aliceAccountID},
		},
		nil,
	)
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	roleIdentifier := utils.RandomSlice(32)

	tm, err := documents.AssembleTokenMessage(
		tokenIdentifier,
		aliceAccountID,
		aliceAccountID,
		roleIdentifier,
		coreDoc.ID(),
		coreDoc.CurrentVersion(),
	)
	assert.NoError(t, err)

	signature, err := privateKey.Sign(tm)
	assert.NoError(t, err)

	accessToken := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Granter:            aliceAccountID.ToBytes(),
		Grantee:            aliceAccountID.ToBytes(),
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
			OwnerIdentity:    aliceAccountID,
			TargetIdentity:   bobAccountID,
		},
	}

	res, err := entityService.GetCurrentVersion(ctx, entityRelationship.ID())
	assert.ErrorIs(t, err, documents.ErrDocumentNotFound)
	assert.Nil(t, res)
}

func TestIntegration_Service_Validate(t *testing.T) {
	aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	// There is an account created for Alice.
	entity := &Entity{
		Data: Data{
			Identity: aliceAccountID,
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
