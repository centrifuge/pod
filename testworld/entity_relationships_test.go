//go:build testworld

package testworld

import (
	"testing"

	"github.com/centrifuge/pod/testworld/park/host"
	"github.com/stretchr/testify/assert"
)

func TestDocumentsAPI_EntityRelationships(t *testing.T) {
	alice, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)
	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)
	charlie, err := controller.GetHost(host.Charlie)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)
	bobClient, err := controller.GetClientForHost(t, host.Bob)
	assert.NoError(t, err)
	charlieClient, err := controller.GetClientForHost(t, host.Charlie)
	assert.NoError(t, err)

	// Alice anchors entity
	payload := defaultEntityPayload(alice.GetMainAccount().GetAccountID().ToHexString(), []string{})
	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	// Alice creates an EntityRelationship with Bob
	payload = defaultRelationshipPayload(
		alice.GetMainAccount().GetAccountID().ToHexString(),
		docID,
		bob.GetMainAccount().GetAccountID().ToHexString(),
	)
	erID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	// Bob should have access to the Entity through the EntityRelationship
	response := bobClient.GetEntityWithRelation(erID)
	response.Path("$.data.legal_name").String().Equal("test company")

	// Charlie should not have access to the entity data
	charlieClient.NonExistingDocumentCheck(erID)

	// Alice updates her entity
	payload = updatedEntityPayload(alice.GetMainAccount().GetAccountID().ToHexString(), []string{})
	payload["document_id"] = docID

	docID, err = aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	// Bob accesses the Entity through the EntityRelationship with Alice, this should return him the latest/updated Entity data
	response = bobClient.GetEntityWithRelation(erID)
	response.Path("$.data.legal_name").String().Equal("edited test company")

	// Alice wants to list all relationships associated with her entity, this should return her one (with Bob)
	response = aliceClient.GetEntityRelationships(docID)
	relationships := response.Array()
	assert.Len(t, relationships.Raw(), 1)

	relationship := relationships.Element(0)
	relationship.Path("$.data.target_identity").String().Equal(bob.GetMainAccount().GetAccountID().ToHexString())

	// Alice creates an EntityRelationship with Charlie
	payload = defaultRelationshipPayload(
		alice.GetMainAccount().GetAccountID().ToHexString(),
		docID,
		charlie.GetMainAccount().GetAccountID().ToHexString(),
	)
	cerID, err := aliceClient.CreateAndCommitDocument(payload)

	// Charlie should now have access to the Entity Data
	response = charlieClient.GetEntityWithRelation(cerID)
	response.Path("$.data.legal_name").String().Equal("edited test company")

	// Alice lists all relationship associated with her entity, this should return her two (with Bob and Charlie)
	response = aliceClient.GetEntityRelationships(docID)
	relationships = response.Array()
	assert.Len(t, relationships.Raw(), 2)

	// Alice revokes the EntityRelationship with Bob
	payload = defaultRelationshipPayload(
		alice.GetMainAccount().GetAccountID().ToHexString(),
		docID,
		bob.GetMainAccount().GetAccountID().ToHexString(),
	)
	payload["document_id"] = erID

	erID, err = aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	// Bob should no longer have access to the EntityRelationship
	bobClient.NonexistentEntityWithRelation(erID)

	// Alice lists all relationships associated with her entity
	// This should return her one relationship with charlie
	response = aliceClient.GetEntityRelationships(docID)
	relationships = response.Array()
	assert.Len(t, relationships.Raw(), 1)
}
