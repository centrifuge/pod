//go:build integration

package p2p

import (
	"context"
	"encoding/json"
	"math/big"
	"math/rand"
	"os"
	"testing"

	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
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
	"github.com/centrifuge/go-centrifuge/documents/generic"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/ipfs_pinning"
	"github.com/centrifuge/go-centrifuge/jobs"
	nftv3 "github.com/centrifuge/go-centrifuge/nft/v3"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/uniques"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	jobsUtil "github.com/centrifuge/go-centrifuge/testingutils/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	p2pUtils "github.com/centrifuge/go-centrifuge/testingutils/p2p"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/stretchr/testify/assert"
)

var (
	peer1ServiceContext map[string]any
	peer2ServiceContext map[string]any

	peer1Account config.Account
	peer2Account config.Account
)

func TestMain(m *testing.M) {
	// Set up 2 peers with 2 separate configurations.

	peer1Bootstrappers := getIntegrationTestBootstrappers()
	peer1ServiceContext = bootstrap.RunTestBootstrappers(peer1Bootstrappers, nil)

	// Get the P2P address of peer 1, and use this address as a bootstrap peer in peer 2.
	peer1Cfg := genericUtils.GetService[config.Configuration](peer1ServiceContext)

	peer1Addr, err := p2pUtils.GetLocalP2PAddress(peer1Cfg)
	if err != nil {
		panic(err)
	}

	_, peer2CfgFile, err := config.CreateTestConfig(func(args map[string]any) {
		args["bootstraps"] = []string{peer1Addr}
		args["p2pPort"] = peer1Cfg.GetP2PPort() + 1
	})
	if err != nil {
		panic(err)
	}

	peer2ServiceContext = map[string]any{
		config.BootstrappedConfigFile: peer2CfgFile,
	}

	peer2Bootstrappers := getIntegrationTestBootstrappers()
	peer2ServiceContext = bootstrap.RunTestBootstrappers(peer2Bootstrappers, peer2ServiceContext)

	// Create the account used in peer 1 - Bob

	if peer1Account, err = v2.BootstrapTestAccount(peer1ServiceContext, keyrings.BobKeyRingPair); err != nil {
		panic(err)
	}

	// Create the account used in peer 2 - Charlie

	if peer2Account, err = v2.BootstrapTestAccount(peer2ServiceContext, keyrings.CharlieKeyRingPair); err != nil {
		panic(err)
	}

	code := m.Run()

	bootstrap.RunTestTeardown(peer1Bootstrappers)
	bootstrap.RunTestTeardown(peer2Bootstrappers)

	os.Exit(code)
}

func TestPeer_Integration_DocumentHandling_CommittedDocument(t *testing.T) {
	ctx := context.Background()

	peer1DocService := genericUtils.GetService[documents.Service](peer1ServiceContext)
	peer1Dispatcher := genericUtils.GetService[jobs.Dispatcher](peer1ServiceContext)
	peer1Peer := genericUtils.GetService[*p2pPeer](peer1ServiceContext)

	peer2DocRepo := genericUtils.GetService[documents.Repository](peer2ServiceContext)
	peer2Peer := genericUtils.GetService[*p2pPeer](peer2ServiceContext)

	// Create a generic document that has both peer 1 and peer 2 accounts as collaborators.

	testDoc, err := peer1DocService.
		Derive(ctx, documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: "generic",
				Collaborators: documents.CollaboratorsAccess{
					ReadWriteCollaborators: []*types.AccountID{
						peer1Account.GetIdentity(),
						peer2Account.GetIdentity(),
					},
				},
			},
		})
	assert.NoError(t, err)

	// Commit the document in the peer 1 doc service.

	jobID, err := peer1DocService.Commit(contextutil.WithAccount(context.Background(), peer1Account), testDoc)
	assert.NoError(t, err)

	err = jobsUtil.WaitForJobToFinish(ctx, peer1Dispatcher, peer1Account.GetIdentity(), jobID)
	assert.NoError(t, err)

	// Retrieve committed document.

	testDoc, err = peer1DocService.GetCurrentVersion(contextutil.WithAccount(context.Background(), peer1Account), testDoc.ID())
	assert.NoError(t, err)

	// Peer 1 sending document to peer 2.

	coreDocument, err := testDoc.PackCoreDocument()
	assert.NoError(t, err)

	sendAnchoredDocRes, err := peer1Peer.SendAnchoredDocument(
		contextutil.WithAccount(ctx, peer1Account),
		peer2Account.GetIdentity(),
		&p2ppb.AnchorDocumentRequest{
			Document: coreDocument,
		},
	)
	assert.NoError(t, err)
	assert.True(t, sendAnchoredDocRes.GetAccepted())

	docRes, err := peer2DocRepo.Get(peer2Account.GetIdentity().ToBytes(), testDoc.CurrentVersion())
	assert.NoError(t, err)

	assertExpectedDocMatchesActual(t, testDoc, docRes)

	// Peer 2 requesting document from peer 1.

	getDocumentRes, err := peer2Peer.GetDocumentRequest(
		contextutil.WithAccount(ctx, peer2Account),
		peer1Account.GetIdentity(),
		&p2ppb.GetDocumentRequest{
			DocumentIdentifier: testDoc.CurrentVersion(),
			AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, testDoc.ID(), getDocumentRes.GetDocument().GetDocumentIdentifier())

	// Peer 1 requesting document signatures from peer 2.

	// Add a new document attribute so that a new document version is prepared.
	attr, err := documents.NewStringAttribute("test-label", documents.AttrString, "test-attr-1")

	err = testDoc.AddAttributes(documents.CollaboratorsAccess{}, true, attr)
	assert.NoError(t, err)

	testDoc.AddUpdateLog(peer1Account.GetIdentity())

	signingRoot, err := testDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	localSignature, err := peer1Account.SignMsg(documents.ConsensusSignaturePayload(signingRoot, false))
	assert.NoError(t, err)

	testDoc.AppendSignatures(localSignature)

	signatures, signatureErrors, err := peer1Peer.GetSignaturesForDocument(
		contextutil.WithAccount(ctx, peer1Account),
		testDoc,
	)
	assert.NoError(t, err)
	assert.Nil(t, signatureErrors)
	assert.Len(t, signatures, 1)
	assert.Equal(t, peer2Account.GetIdentity().ToBytes(), signatures[0].GetSignerId())
}

func TestPeer_Integration_GetDocumentSignatures(t *testing.T) {
	ctx := context.Background()

	peer1DocService := genericUtils.GetService[documents.Service](peer1ServiceContext)
	peer1Peer := genericUtils.GetService[*p2pPeer](peer1ServiceContext)

	// Create a generic document that has both peer 1 and peer 2 accounts as collaborators.

	testDoc, err := peer1DocService.
		Derive(ctx, documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: "generic",
				Collaborators: documents.CollaboratorsAccess{
					ReadWriteCollaborators: []*types.AccountID{
						peer1Account.GetIdentity(),
						peer2Account.GetIdentity(),
					},
				},
			},
		})
	assert.NoError(t, err)

	// Peer 1 requesting a signature with an invalid document - it has no timestamp, no author, no author signatures.

	signatures, signatureErrors, err := peer1Peer.GetSignaturesForDocument(
		contextutil.WithAccount(ctx, peer1Account),
		testDoc,
	)
	assert.NoError(t, err)
	assert.NotNil(t, signatureErrors)
	assert.Nil(t, signatures)

	// Peer 1 requesting a signature with a valid document.
	testDoc.AddUpdateLog(peer1Account.GetIdentity())

	signingRoot, err := testDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	localSignature, err := peer1Account.SignMsg(documents.ConsensusSignaturePayload(signingRoot, false))
	assert.NoError(t, err)

	testDoc.AppendSignatures(localSignature)

	signatures, signatureErrors, err = peer1Peer.GetSignaturesForDocument(
		contextutil.WithAccount(ctx, peer1Account),
		testDoc,
	)
	assert.NoError(t, err)
	assert.Nil(t, signatureErrors)
	assert.Len(t, signatures, 1)
	assert.Equal(t, peer2Account.GetIdentity().ToBytes(), signatures[0].GetSignerId())
}

func TestPeer_Integration_SendAnchoredDocument(t *testing.T) {
	ctx := context.Background()

	peer1DocService := genericUtils.GetService[documents.Service](peer1ServiceContext)
	peer1Peer := genericUtils.GetService[*p2pPeer](peer1ServiceContext)
	peer1AnchorProcessor := genericUtils.GetService[documents.AnchorProcessor](peer1ServiceContext)

	peer2DocRepo := genericUtils.GetService[documents.Repository](peer2ServiceContext)

	// Create a generic document that has both peer 1 and peer 2 accounts as collaborators.

	testDoc, err := peer1DocService.
		Derive(ctx, documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: "generic",
				Collaborators: documents.CollaboratorsAccess{
					ReadWriteCollaborators: []*types.AccountID{
						peer1Account.GetIdentity(),
						peer2Account.GetIdentity(),
					},
				},
			},
		})
	assert.NoError(t, err)

	// Make sure the doc is valid by adding an update log and a signature.
	testDoc.AddUpdateLog(peer1Account.GetIdentity())

	signingRoot, err := testDoc.CalculateSigningRoot()
	assert.NoError(t, err)

	localSignature, err := peer1Account.SignMsg(documents.ConsensusSignaturePayload(signingRoot, false))
	assert.NoError(t, err)

	testDoc.AppendSignatures(localSignature)

	_, err = testDoc.CalculateDocumentRoot()
	assert.NoError(t, err)

	err = testDoc.SetStatus(documents.Committing)
	assert.NoError(t, err)

	coreDocument, err := testDoc.PackCoreDocument()
	assert.NoError(t, err)

	// Peer 1 sending non-anchored document to peer 2. This should fail since the document is not anchored.

	sendAnchoredDocRes, err := peer1Peer.SendAnchoredDocument(
		contextutil.WithAccount(ctx, peer1Account),
		peer2Account.GetIdentity(),
		&p2ppb.AnchorDocumentRequest{
			Document: coreDocument,
		},
	)
	assert.NotNil(t, err)
	assert.False(t, sendAnchoredDocRes.GetAccepted())

	// Anchor document.
	err = peer1AnchorProcessor.AnchorDocument(contextutil.WithAccount(ctx, peer1Account), testDoc)
	assert.NoError(t, err)

	// Peer 1 sending anchored document to peer 2. This should fail because peer 2 does not have a record
	// of this document in its storage.

	sendAnchoredDocRes, err = peer1Peer.SendAnchoredDocument(
		contextutil.WithAccount(ctx, peer1Account),
		peer2Account.GetIdentity(),
		&p2ppb.AnchorDocumentRequest{
			Document: coreDocument,
		},
	)
	assert.NotNil(t, err)
	assert.False(t, sendAnchoredDocRes.GetAccepted())

	// Store document in the peer 2 storage.

	err = peer2DocRepo.Create(peer2Account.GetIdentity().ToBytes(), testDoc.CurrentVersion(), testDoc)
	assert.NoError(t, err)

	// Successful send.

	sendAnchoredDocRes, err = peer1Peer.SendAnchoredDocument(
		contextutil.WithAccount(ctx, peer1Account),
		peer2Account.GetIdentity(),
		&p2ppb.AnchorDocumentRequest{
			Document: coreDocument,
		},
	)
	assert.NoError(t, err)
	assert.True(t, sendAnchoredDocRes.GetAccepted())
}

func TestPeer_Integration_GetDocumentRequest_RequesterVerification(t *testing.T) {
	ctx := context.Background()

	peer1DocService := genericUtils.GetService[documents.Service](peer1ServiceContext)
	peer1DocRepo := genericUtils.GetService[documents.Repository](peer1ServiceContext)

	peer2Peer := genericUtils.GetService[*p2pPeer](peer2ServiceContext)

	// Create a generic document that has both peer 1 and peer 2 accounts as collaborators.

	document1, err := peer1DocService.
		Derive(ctx, documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: "generic",
				Collaborators: documents.CollaboratorsAccess{
					ReadWriteCollaborators: []*types.AccountID{
						peer1Account.GetIdentity(),
						peer2Account.GetIdentity(),
					},
				},
			},
		})
	assert.NoError(t, err)

	err = document1.SetStatus(documents.Committed)
	assert.NoError(t, err)

	err = peer1DocRepo.Create(peer1Account.GetIdentity().ToBytes(), document1.CurrentVersion(), document1)
	assert.NoError(t, err)

	// Peer 2 requesting document 1 from peer 1.

	getDocumentRes, err := peer2Peer.GetDocumentRequest(
		contextutil.WithAccount(ctx, peer2Account),
		peer1Account.GetIdentity(),
		&p2ppb.GetDocumentRequest{
			DocumentIdentifier: document1.CurrentVersion(),
			AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, document1.ID(), getDocumentRes.GetDocument().GetDocumentIdentifier())

	// Create a generic document that only has the peer 1 account as collaborators.

	document2, err := peer1DocService.
		Derive(ctx, documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: "generic",
				Collaborators: documents.CollaboratorsAccess{
					ReadWriteCollaborators: []*types.AccountID{
						peer1Account.GetIdentity(),
					},
				},
			},
		})
	assert.NoError(t, err)

	err = document2.SetStatus(documents.Committed)
	assert.NoError(t, err)

	err = peer1DocRepo.Create(peer1Account.GetIdentity().ToBytes(), document2.CurrentVersion(), document2)
	assert.NoError(t, err)

	// Peer 2 requesting document 2 from peer 1.

	getDocumentRes, err = peer2Peer.GetDocumentRequest(
		contextutil.WithAccount(ctx, peer2Account),
		peer1Account.GetIdentity(),
		&p2ppb.GetDocumentRequest{
			DocumentIdentifier: document2.CurrentVersion(),
			AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
		},
	)
	assert.NotNil(t, err)
	assert.Nil(t, getDocumentRes)
}

func TestPeer_Integration_GetDocumentRequest_NFTOwnerVerification(t *testing.T) {
	ctx := context.Background()

	peer1DocService := genericUtils.GetService[documents.Service](peer1ServiceContext)
	peer1DocRepo := genericUtils.GetService[documents.Repository](peer1ServiceContext)

	peer2Peer := genericUtils.GetService[*p2pPeer](peer2ServiceContext)
	peer2UniquesAPI := genericUtils.GetService[uniques.API](peer2ServiceContext)

	// Create a generic document that has both peer 1 and peer 2 accounts as collaborators.

	document1, err := peer1DocService.
		Derive(ctx, documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: "generic",
				Collaborators: documents.CollaboratorsAccess{
					ReadWriteCollaborators: []*types.AccountID{
						peer1Account.GetIdentity(),
						peer2Account.GetIdentity(),
					},
				},
			},
		})
	assert.NoError(t, err)

	// Mint NFT for account 2.

	collectionID := types.U64(rand.Uint64())
	encodedCollectionID, err := codec.Encode(collectionID)
	assert.NoError(t, err)

	itemID := types.NewU128(*big.NewInt(rand.Int63()))
	encodedItemID, err := codec.Encode(itemID)
	assert.NoError(t, err)

	_, err = peer2UniquesAPI.CreateCollection(contextutil.WithAccount(ctx, peer2Account), collectionID)
	assert.NoError(t, err)

	_, err = peer2UniquesAPI.Mint(contextutil.WithAccount(ctx, peer2Account), collectionID, itemID, peer2Account.GetIdentity())
	assert.NoError(t, err)

	// Add NFT to document 1, set its status to committed and store it in the peer 1 doc repo.

	err = document1.AddNFT(true, collectionID, itemID)
	assert.NoError(t, err)

	err = document1.SetStatus(documents.Committed)
	assert.NoError(t, err)

	err = peer1DocRepo.Create(peer1Account.GetIdentity().ToBytes(), document1.CurrentVersion(), document1)
	assert.NoError(t, err)

	// Peer 2 requesting document 1 from peer 1 using valid NFT info.

	getDocumentRes, err := peer2Peer.GetDocumentRequest(
		contextutil.WithAccount(ctx, peer2Account),
		peer1Account.GetIdentity(),
		&p2ppb.GetDocumentRequest{
			DocumentIdentifier: document1.ID(),
			AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
			NftCollectionId:    encodedCollectionID,
			NftItemId:          encodedItemID,
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, document1.ID(), getDocumentRes.GetDocument().GetDocumentIdentifier())

	// Peer 2 requesting document 1 from peer 1 using invalid NFT info.

	getDocumentRes, err = peer2Peer.GetDocumentRequest(
		contextutil.WithAccount(ctx, peer2Account),
		peer1Account.GetIdentity(),
		&p2ppb.GetDocumentRequest{
			DocumentIdentifier: document1.CurrentVersion(),
			AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
			NftCollectionId:    utils.RandomSlice(32),
			NftItemId:          utils.RandomSlice(32),
		},
	)
	assert.NotNil(t, err)
	assert.Nil(t, getDocumentRes)

	// Create a generic document that only has the peer 1 account as collaborator.

	document2, err := peer1DocService.
		Derive(ctx, documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: "generic",
				Collaborators: documents.CollaboratorsAccess{
					ReadWriteCollaborators: []*types.AccountID{
						peer1Account.GetIdentity(),
					},
				},
			},
		})
	assert.NoError(t, err)

	// Add NFT to document 2.

	err = document2.AddNFT(true, collectionID, itemID)
	assert.NoError(t, err)

	err = document2.SetStatus(documents.Committed)
	assert.NoError(t, err)

	err = peer1DocRepo.Create(peer1Account.GetIdentity().ToBytes(), document2.CurrentVersion(), document2)
	assert.NoError(t, err)

	// Peer 2 requesting document 2 from peer 1 using valid NFT info.
	// This should fail since the account for peer 2 is not added as a collaborator.

	getDocumentRes, err = peer2Peer.GetDocumentRequest(
		contextutil.WithAccount(ctx, peer2Account),
		peer1Account.GetIdentity(),
		&p2ppb.GetDocumentRequest{
			DocumentIdentifier: document2.ID(),
			AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
			NftCollectionId:    utils.RandomSlice(32),
			NftItemId:          utils.RandomSlice(32),
		},
	)
	assert.NotNil(t, err)
	assert.Nil(t, getDocumentRes)
}

func TestPeer_Integration_GetDocumentRequest_AccessTokenVerification(t *testing.T) {
	ctx := context.Background()

	peer1DocService := genericUtils.GetService[documents.Service](peer1ServiceContext)
	peer1DocRepo := genericUtils.GetService[documents.Repository](peer1ServiceContext)

	peer2Peer := genericUtils.GetService[*p2pPeer](peer2ServiceContext)

	// Create a generic document with no collaborators.

	document1, err := peer1DocService.
		Derive(ctx, documents.UpdatePayload{
			CreatePayload: documents.CreatePayload{
				Scheme: "generic",
			},
		})
	assert.NoError(t, err)

	err = document1.SetStatus(documents.Committed)
	assert.NoError(t, err)

	err = peer1DocRepo.Create(peer1Account.GetIdentity().ToBytes(), document1.CurrentVersion(), document1)
	assert.NoError(t, err)

	// Create an entity relationship between document 1 and the peer 2 account.

	entityRelationshipData := &entityrelationship.Data{
		OwnerIdentity:    peer1Account.GetIdentity(),
		EntityIdentifier: document1.ID(),
		TargetIdentity:   peer2Account.GetIdentity(),
	}

	encodedData, err := json.Marshal(entityRelationshipData)
	assert.NoError(t, err)

	entityRelationship1, err := peer1DocService.
		Derive(
			contextutil.WithAccount(ctx, peer1Account),
			documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: "entity_relationship",
					Data:   encodedData,
				},
			})
	assert.NoError(t, err)

	// Log an update to ensure that the entity relationship has an author and a timestamp.
	entityRelationship1.AddUpdateLog(peer1Account.GetIdentity())

	// Set status to committed to ensure that we're saving the latest version as well.
	err = entityRelationship1.SetStatus(documents.Committed)
	assert.NoError(t, err)

	err = peer1DocRepo.Create(peer1Account.GetIdentity().ToBytes(), entityRelationship1.CurrentVersion(), entityRelationship1)
	assert.NoError(t, err)

	// Peer 2 requesting document 1 from peer 1 using a valid access token request.

	getDocumentRes, err := peer2Peer.GetDocumentRequest(
		contextutil.WithAccount(ctx, peer2Account),
		peer1Account.GetIdentity(),
		&p2ppb.GetDocumentRequest{
			DocumentIdentifier: document1.ID(),
			AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
			AccessTokenRequest: &p2ppb.AccessTokenRequest{
				DelegatingDocumentIdentifier: entityRelationship1.ID(),
				AccessTokenId:                entityRelationship1.GetAccessTokens()[0].GetIdentifier(),
			},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, document1.ID(), getDocumentRes.GetDocument().GetDocumentIdentifier())

	// Peer 2 requesting document 1 from peer 1 using an access token request with invalid token.

	getDocumentRes, err = peer2Peer.GetDocumentRequest(
		contextutil.WithAccount(ctx, peer2Account),
		peer1Account.GetIdentity(),
		&p2ppb.GetDocumentRequest{
			DocumentIdentifier: document1.ID(),
			AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
			AccessTokenRequest: &p2ppb.AccessTokenRequest{
				DelegatingDocumentIdentifier: entityRelationship1.ID(),
				AccessTokenId:                utils.RandomSlice(32),
			},
		},
	)
	assert.NotNil(t, err)

	// Peer 2 requesting document 1 from peer 1 using an access token request with an invalid identifier.

	getDocumentRes, err = peer2Peer.GetDocumentRequest(
		contextutil.WithAccount(ctx, peer2Account),
		peer1Account.GetIdentity(),
		&p2ppb.GetDocumentRequest{
			DocumentIdentifier: document1.ID(),
			AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
			AccessTokenRequest: &p2ppb.AccessTokenRequest{
				DelegatingDocumentIdentifier: utils.RandomSlice(32),
				AccessTokenId:                utils.RandomSlice(32),
			},
		},
	)
	assert.NotNil(t, err)

	// Create an entity relationship between document 1 and a random account.

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityRelationshipData = &entityrelationship.Data{
		OwnerIdentity:    peer1Account.GetIdentity(),
		EntityIdentifier: document1.ID(),
		TargetIdentity:   randomAccountID,
	}

	encodedData, err = json.Marshal(entityRelationshipData)
	assert.NoError(t, err)

	entityRelationship2, err := peer1DocService.
		Derive(
			contextutil.WithAccount(ctx, peer1Account),
			documents.UpdatePayload{
				CreatePayload: documents.CreatePayload{
					Scheme: "entity_relationship",
					Data:   encodedData,
				},
			})
	assert.NoError(t, err)

	// Log an update to ensure that the entity relationship has an author and a timestamp.
	entityRelationship2.AddUpdateLog(peer1Account.GetIdentity())

	// Set status to committed to ensure that we're saving the latest version as well.
	err = entityRelationship2.SetStatus(documents.Committed)
	assert.NoError(t, err)

	err = peer1DocRepo.Create(peer1Account.GetIdentity().ToBytes(), entityRelationship2.CurrentVersion(), entityRelationship2)
	assert.NoError(t, err)

	// Request the document from peer 2 using the entity relationship that was created for the random account.

	getDocumentRes, err = peer2Peer.GetDocumentRequest(
		contextutil.WithAccount(ctx, peer2Account),
		peer1Account.GetIdentity(),
		&p2ppb.GetDocumentRequest{
			DocumentIdentifier: document1.ID(),
			AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
			AccessTokenRequest: &p2ppb.AccessTokenRequest{
				DelegatingDocumentIdentifier: entityRelationship2.ID(),
				AccessTokenId:                entityRelationship2.GetAccessTokens()[0].GetIdentifier(),
			},
		},
	)
	assert.NotNil(t, err)
}

func getIntegrationTestBootstrappers() []bootstrap.TestBootstrapper {
	return []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		&jobs.Bootstrapper{},
		&integration_test.Bootstrapper{},
		centchain.Bootstrapper{},
		&pallets.Bootstrapper{},
		&protocolIDDispatcher.Bootstrapper{},
		&v2.AccountTestBootstrapper{},
		documents.Bootstrapper{},
		pending.Bootstrapper{},
		&ipfs_pinning.TestBootstrapper{},
		&nftv3.Bootstrapper{},
		&Bootstrapper{},
		documents.PostBootstrapper{},
		generic.Bootstrapper{},
		entityrelationship.Bootstrapper{},
	}
}

func assertExpectedDocMatchesActual(t *testing.T, expected documents.Document, actual documents.Document) {
	assert.Equal(t, expected.ID(), actual.ID())
	assert.Equal(t, expected.CurrentVersion(), actual.CurrentVersion())
	assert.Equal(t, expected.CurrentVersionPreimage(), actual.CurrentVersionPreimage())
	assert.Equal(t, expected.NextVersion(), actual.NextVersion())
	assert.Equal(t, expected.NextPreimage(), actual.NextPreimage())
	assert.Equal(t, expected.PreviousVersion(), actual.PreviousVersion())
	assert.Equal(t, len(expected.Signatures()), len(actual.Signatures()))
	assert.Equal(t, expected.Signatures()[0].GetSignature(), actual.Signatures()[0].GetSignature())
	assert.Equal(t, expected.Signatures()[0].GetPublicKey(), actual.Signatures()[0].GetPublicKey())
	assert.Equal(t, expected.Signatures()[0].GetSignerId(), actual.Signatures()[0].GetSignerId())
	assert.Equal(t, expected.Signatures()[0].GetSignatureId(), actual.Signatures()[0].GetSignatureId())
	assert.Equal(t, expected.Signatures()[0].GetTransitionValidated(), actual.Signatures()[0].GetTransitionValidated())
}
