// +build testworld

package testworld

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/gavv/httpexpect"
)

func TestHost_Entity_EntityRelationships(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice anchors entity
	res := createDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusAccepted, defaultEntityPayload(alice.id.String(), []string{}))
	entityIdentifier := getDocumentIdentifier(t, res)
	txID := getJobID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Alice creates an EntityRelationship with Bob
	resB := shareEntity(alice.httpExpect, alice.id.String(), entityIdentifier, http.StatusAccepted, defaultRelationshipPayload(entityIdentifier, bob.id.String()))
	relationshipIdentifierB := getDocumentIdentifier(t, resB)
	txID = getJobID(t, resB)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Charlie should not have access to the entity data
	relationshipParams := map[string]interface{}{
		"r_identifier": relationshipIdentifierB,
	}
	response := nonexistentEntityWithRelation(charlie.httpExpect, charlie.id.String(), typeEntity, relationshipParams)

	// Bob should have access to the Entity through the EntityRelationship
	response = getEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, relationshipParams)
	response.Path("$.data.entity.legal_name").String().Equal("test company")

	// Alice updates her entity
	res = updateDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusAccepted, entityIdentifier, updatedEntityPayload(alice.id.String(), []string{}))
	txID = getJobID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Bob accesses the Entity through the EntityRelationship with Alice, this should return him the latest/updated Entity data
	response = getEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, relationshipParams)
	response.Path("$.data.entity.legal_name").String().Equal("edited test company")

	// Alice wants to list all relationships associated with her entity, this should return her one (with Bob)
	response = getEntity(alice.httpExpect, alice.id.String(), entityIdentifier)
	response.Path("$.data.relationships[0].active").Boolean().Equal(true)
	response.Path("$.data.relationships[0].target_identity").String().Equal(bob.id.String())

	// Alice creates an EntityRelationship with Charlie
	resC := shareEntity(alice.httpExpect, alice.id.String(), entityIdentifier, http.StatusAccepted, defaultRelationshipPayload(entityIdentifier, charlie.id.String()))
	relationshipIdentifierC := getDocumentIdentifier(t, resC)
	txID = getJobID(t, resC)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Charlie should now have access to the Entity Data
	relationshipParamsC := map[string]interface{}{
		"r_identifier": relationshipIdentifierC,
	}
	response = getEntityWithRelation(charlie.httpExpect, charlie.id.String(), typeEntity, relationshipParamsC)
	response.Path("$.data.entity.legal_name").String().Equal("edited test company")

	// Alice lists all relationship associated with her entity, this should return her two (with Bob and Charlie)
	response = getEntity(alice.httpExpect, alice.id.String(), entityIdentifier)
	cIdx, bIdx := checkRelationships(response, charlie.id.String(), bob.id.String())
	response.Path("$.data.relationships[" + cIdx + "].active").Boolean().Equal(true)
	response.Path("$.data.relationships[" + bIdx + "].active").Boolean().Equal(true)

	// Alice revokes the EntityRelationship with Bob
	resB = revokeEntity(alice.httpExpect, alice.id.String(), entityIdentifier, http.StatusAccepted, defaultRelationshipPayload(entityIdentifier, bob.id.String()))
	txID = getJobID(t, resB)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// Bob should no longer have access to the EntityRelationship
	response = nonexistentEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, relationshipParams)

	// Alice lists all relationships associated with her entity
	// This should return her two relationships: one valid with Charlie, one revoked with Bob
	response = getEntity(alice.httpExpect, alice.id.String(), entityIdentifier)
	cIdx, bIdx = checkRelationships(response, charlie.id.String(), bob.id.String())
	response.Path("$.data.relationships[" + cIdx + "].active").Boolean().Equal(true)
	//todo add check active for bob not existing

}

func checkRelationships(response *httpexpect.Value, charlieDID, bobDID string) (string, string) {
	response.Path("$.data.relationships").Array().Length().Equal(2)
	firstR := response.Path("$.data.relationships[0].target_identity").String().Raw()
	charlieIdx := 0
	if firstR != charlieDID {
		charlieIdx = 1

	}
	bIdx := strconv.Itoa(1 - charlieIdx)
	cIdx := strconv.Itoa(charlieIdx)
	response.Path("$.data.relationships[" + cIdx + "].target_identity").String().Equal(charlieDID)
	response.Path("$.data.relationships[" + bIdx + "].target_identity").String().Equal(bobDID)

	return cIdx, bIdx
}
