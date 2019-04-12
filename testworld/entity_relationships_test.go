// +build testworld

package testworld

import (
	"net/http"
	"testing"
)

func TestHost_Entity_EntityRelationships(t *testing.T) {
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

	// Alice creates an EntityRelationship with Bob
	res = shareEntity(alice.httpExpect, alice.id.String(), entityIdentifier, http.StatusOK, defaultRelationshipPayload(alice.id.String(), bob.id.String()))
	relationshipIdentifier := getDocumentIdentifier(t, res)
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Charlie should not access to the entity data
	params := map[string]interface{}{
		"er_identifier": relationshipIdentifier,
	}
	response := getEntityWithRelation(charlie.httpExpect, charlie.id.String(), typeEntity, params)
	response.Path("$.data.entity.legal_name").String().Equal("test company")

	// Bob should have access to the Entity through the EntityRelationship
	response = getEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, params)
	response.Path("$.data.entity.legal_name").String().Equal("test company")

	// Alice updates her entity
	res = updateDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusOK, entityIdentifier, updatedEntityPayload(alice.id.String(), []string{}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	/// Bob accesses the Entity through the EntityRelationship with Alice, this should return him the latest/updated Entity data
	response = getEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, params)
	response.Path("$.data.entity.legal_name").String().Equal("edited test company")

	//// Alice wants to list all relationships associated with her entity, this should return her one (with Bob)
	//relationships, err := alice.host.entityService.ListEntityRelationships(ctxBob, entityIdentifierByte)
	//assert.Len(t, relationships, 1)
	//assert.Equal(t, relationships[0].TargetIdentity, bob.id)
	//
	//// Alice creates an EntityRelationship with Charlie
	//relationshipData = &entitypb2.RelationshipData{
	//	EntityIdentifier: entityIdentifier,
	//	OwnerIdentity:    alice.id.String(),
	//	TargetIdentity:   charlie.id.String(),
	//}
	//
	//relationship := entityrelationship.EntityRelationship{}
	//err = er.InitEntityRelationshipInput(ctxAlice, entityIdentifier, relationshipData)
	//assert.NoError(t, err)
	//
	//relationshipModel, _, isDone, err = alice.host.entityService.Share(ctxAlice, &relationship)
	//assert.NoError(t, err)
	//done = <-isDone
	//assert.True(t, done)
	//cd, err = relationshipModel.PackCoreDocument()
	//assert.NoError(t, err)
	//
	//relationshipIdentifier = cd.DocumentIdentifier
	//charlieModel, err := charlie.host.entityService.GetCurrentVersion(ctxCharlie, relationshipIdentifier)
	//assert.NoError(t, err)
	//assert.Equal(t, relationshipModel.CurrentVersion(), charlieModel.CurrentVersion())
	//
	//// Alice lists all relationship associated with her entity, this should return her two (with Bob and Charlie)
	//relationships, err = alice.host.entityService.ListEntityRelationships(ctxBob, entityIdentifierByte)
	//assert.Len(t, relationships, 2)
	//assert.Equal(t, relationships[0].TargetIdentity, bob.id)
	//assert.Equal(t, relationships[1].TargetIdentity, charlie.id)
	//
	//// Alice revokes the EntityRelationship with Bob
	//relationshipModel, _, isDone, err = alice.host.entityService.Revoke(ctxAlice, &er)
	//assert.NoError(t, err)
	//done = <-isDone
	//assert.True(t, done)
	//cd, err = relationshipModel.PackCoreDocument()
	//assert.NoError(t, err)
	//
	//// Bob should no longer have access to the EntityRelationship
	//bobModel, err = bob.host.entityService.GetCurrentVersion(ctxBob, relationshipIdentifier)
	//assert.Error(t, err)
	//
	//// Alice lists all relationships associated with her entity
	//// This should return her two relationships: one valid with Charlie, one revoked with Bob
	//relationships, err = alice.host.entityService.ListEntityRelationships(ctxBob, entityIdentifierByte)
	//assert.Len(t, relationships, 2)
	//assert.Equal(t, relationships[0].TargetIdentity, charlie.id)
	//assert.Equal(t, relationships[1].TargetIdentity, bob.id)
	//assert.Len(t, relationships[1].Document.AccessTokens, 0)
}
