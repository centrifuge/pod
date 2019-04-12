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
	relationshipParams := map[string]interface{}{
		"r_identifier": relationshipIdentifier,
	}
	response := getEntityWithRelation(charlie.httpExpect, charlie.id.String(), typeEntity, relationshipParams)
	response.Path("$.data.entity.legal_name").String().Equal("test company")

	// Bob should have access to the Entity through the EntityRelationship
	response = getEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, relationshipParams)
	response.Path("$.data.entity.legal_name").String().Equal("test company")

	// Alice updates her entity
	res = updateDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusOK, entityIdentifier, updatedEntityPayload(alice.id.String(), []string{}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Bob accesses the Entity through the EntityRelationship with Alice, this should return him the latest/updated Entity data
	response = getEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, relationshipParams)
	response.Path("$.data.entity.legal_name").String().Equal("edited test company")

	// Alice wants to list all relationships associated with her entity, this should return her one (with Bob)
	entityParams := map[string]interface{}{
		"er_identifier": entityIdentifier,
	}
	response = listRelationships(alice.httpExpect, alice.id.String(), entityParams)

	// Alice creates an EntityRelationship with Charlie
	res = shareEntity(alice.httpExpect, alice.id.String(), entityIdentifier, http.StatusOK, defaultRelationshipPayload(alice.id.String(), charlie.id.String()))
	relationshipIdentifierC := getDocumentIdentifier(t, res)
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Charlie should now have access to the Entity Data
	relationshipParamsC := map[string]interface{}{
		"r_identifier": relationshipIdentifierC,
	}
	response = getEntityWithRelation(charlie.httpExpect, charlie.id.String(), typeEntity, relationshipParamsC)
	response.Path("$.data.entity.legal_name").String().Equal("test company")

	// Alice lists all relationship associated with her entity, this should return her two (with Bob and Charlie)
	response = listRelationships(alice.httpExpect, alice.id.String(), entityParams)

	// Alice revokes the EntityRelationship with Bob
	res = revokeEntity(alice.httpExpect, alice.id.String(), entityIdentifier, http.StatusOK, defaultRelationshipPayload(alice.id.String(), bob.id.String()))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Bob should no longer have access to the EntityRelationship
	response = getEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, relationshipParams)

	// Alice lists all relationships associated with her entity
	// This should return her two relationships: one valid with Charlie, one revoked with Bob
	response = listRelationships(alice.httpExpect, alice.id.String(), entityParams)
}
