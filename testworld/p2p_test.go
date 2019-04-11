// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	entitypb2 "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestHost_P2PGetDocumentWithToken(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice anchors entity
	res := createDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusOK, defaultEntityPayload(alice.id.String(), []string{}))
	entityIdentifier := getDocumentIdentifier(t, res)
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Bob should not have access to the entity data yet
	ctxBob := testingconfig.CreateAccountContext(t, bob.host.config)
	bobModel, err := bob.host.entityService.GetCurrentVersion(ctxBob, alice.id[:])
	assert.Error(t, err)

	// Alice creates an EntityRelationship with Bob using the entity service
	ctxAlice := testingconfig.CreateAccountContext(t, alice.host.config)
	relationshipData := &entitypb2.RelationshipData{
		EntityIdentifier: entityIdentifier,
		OwnerIdentity:    alice.id.String(),
		TargetIdentity:   bob.id.String(),
	}

	er := entityrelationship.EntityRelationship{}
	err = er.InitEntityRelationshipInput(ctxAlice, entityIdentifier, relationshipData)
	assert.NoError(t, err)

	relationshipModel, _, isDone, err := alice.host.entityService.Share(ctxAlice, &er)
	assert.NoError(t, err)
	done := <-isDone
	assert.True(t, done)
	cd, err := relationshipModel.PackCoreDocument()
	assert.NoError(t, err)

	// Now, Bob should have the EntityRelationship
	relationshipIdentifier := cd.DocumentIdentifier
	bobModel, err = bob.host.entityService.GetCurrentVersion(ctxBob, relationshipIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, relationshipModel.CurrentVersion(), bobModel.CurrentVersion())

	// Bob accesses Entity directly on p2p
	accessTokenRequest := &p2ppb.AccessTokenRequest{DelegatingDocumentIdentifier: relationshipIdentifier, AccessTokenId: cd.AccessTokens[0].Identifier}
	entityIdentifierByte, err := hexutil.Decode(entityIdentifier) // remove 0x
	assert.NoError(t, err)
	request := &p2ppb.GetDocumentRequest{DocumentIdentifier: entityIdentifierByte,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: accessTokenRequest,
	}

	response, err := bob.host.p2pClient.GetDocumentRequest(ctxBob, alice.id, request)
	assert.NoError(t, err)
	assert.Equal(t, response.Document.DocumentIdentifier, entityIdentifierByte)

	// Alice updates her entity
	res = updateDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusOK, entityIdentifier, updatedDocumentPayload(typeEntity, []string{}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Bob accesses the Entity through the EntityRelationship with Alice, this should return him the latest/updated Entity data
	response, err = bob.host.p2pClient.GetDocumentRequest(ctxBob, alice.id, request)
	assert.NoError(t, err)
	assert.Equal(t, response.Document.DocumentIdentifier, entityIdentifierByte)
	assert.Equal(t, response.Document.PreviousVersion, entityIdentifierByte)

	// Charlie does not have an EntityRelationship with Alice, but requests her EntityData with an access token from Bob
	ctxCharlie := testingconfig.CreateAccountContext(t, charlie.host.config)
	response, err = charlie.host.p2pClient.GetDocumentRequest(ctxCharlie, alice.id, request)
	assert.Error(t, err)

	// Alice wants to list all relationships associated with her entity, this should return her one (with Bob)

	// Alice creates an EntityRelationship with Charlie
	relationshipData = &entitypb2.RelationshipData{
		EntityIdentifier: entityIdentifier,
		OwnerIdentity:    alice.id.String(),
		TargetIdentity:   charlie.id.String(),
	}

	relationship := entityrelationship.EntityRelationship{}
	err = er.InitEntityRelationshipInput(ctxAlice, entityIdentifier, relationshipData)
	assert.NoError(t, err)

	relationshipModel, _, isDone, err = alice.host.entityService.Share(ctxAlice, &relationship)
	assert.NoError(t, err)
	done = <-isDone
	assert.True(t, done)
	cd, err = relationshipModel.PackCoreDocument()
	assert.NoError(t, err)

	relationshipIdentifier = cd.DocumentIdentifier
	charlieModel, err := charlie.host.entityService.GetCurrentVersion(ctxCharlie, relationshipIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, relationshipModel.CurrentVersion(), charlieModel.CurrentVersion())

	// Alice lists all relationship associated with her entity, this should return her two (with Bob and Charlie)

	// Alice revokes the EntityRelationship with Bob
	relationshipModel, _, isDone, err = alice.host.entityService.Revoke(ctxAlice, &er)
	assert.NoError(t, err)
	done = <-isDone
	assert.True(t, done)
	cd, err = relationshipModel.PackCoreDocument()
	assert.NoError(t, err)

	// Bob should no longer have access to the EntityRelationship
	bobModel, err = bob.host.entityService.GetCurrentVersion(ctxBob, relationshipIdentifier)
	assert.Error(t, err)

	// Alice lists all relationships associated with her entity
	// This should return her two (with Bob, which is revoked, and with Charlie, which is still valid)

}
