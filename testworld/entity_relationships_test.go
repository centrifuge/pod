// +build testworld

package testworld

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHost_Entity_EntityRelationships(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice anchors entity
	docID := createAndCommitDocument(t, alice.httpExpect, alice.id.String(), defaultEntityPayload(alice.id.String(), []string{}))

	// Alice creates an EntityRelationship with Bob
	erID := createAndCommitDocument(t, alice.httpExpect, alice.id.String(),
		defaultRelationshipPayload(alice.id.String(), docID,
			bob.id.String()))

	// Charlie should not have access to the entity data
	nonExistingDocumentCheck(charlie.httpExpect, charlie.id.String(), erID)

	// Bob should have access to the Entity through the EntityRelationship
	response := getEntityWithRelation(bob.httpExpect, bob.id.String(), erID)
	response.Path("$.data.legal_name").String().Equal("test company")

	// Alice updates her entity
	payload := updatedEntityPayload(alice.id.String(), []string{})
	payload["document_id"] = docID
	docID = createAndCommitDocument(t, alice.httpExpect, alice.id.String(), payload)

	// Bob accesses the Entity through the EntityRelationship with Alice, this should return him the latest/updated Entity data
	response = getEntityWithRelation(bob.httpExpect, bob.id.String(), erID)
	response.Path("$.data.legal_name").String().Equal("edited test company")

	// Alice wants to list all relationships associated with her entity, this should return her one (with Bob)
	response = getEntityRelationships(alice.httpExpect, alice.id.String(), docID)
	relationships := response.Array()
	assert.Len(t, relationships.Raw(), 1)
	relationship := relationships.Element(0)
	relationship.Path("$.data.target_identity").String().Equal(bob.id.String())

	// Alice creates an EntityRelationship with Charlie
	cerID := createAndCommitDocument(t, alice.httpExpect, alice.id.String(),
		defaultRelationshipPayload(alice.id.String(), docID, charlie.id.String()))

	// Charlie should now have access to the Entity Data
	response = getEntityWithRelation(charlie.httpExpect, charlie.id.String(), cerID)
	response.Path("$.data.legal_name").String().Equal("edited test company")

	// Alice lists all relationship associated with her entity, this should return her two (with Bob and Charlie)
	response = getEntityRelationships(alice.httpExpect, alice.id.String(), docID)
	relationships = response.Array()
	assert.Len(t, relationships.Raw(), 2)

	// Alice revokes the EntityRelationship with Bob
	payload = defaultRelationshipPayload(alice.id.String(), docID, bob.id.String())
	payload["document_id"] = erID
	erID = createAndCommitDocument(t, alice.httpExpect, alice.id.String(), payload)

	// Bob should no longer have access to the EntityRelationship
	nonexistentEntityWithRelation(bob.httpExpect, bob.id.String(), erID)

	// Alice lists all relationships associated with her entity
	// This should return her one relationship with charlie
	response = getEntityRelationships(alice.httpExpect, alice.id.String(), docID)
	relationships = response.Array()
	assert.Len(t, relationships.Raw(), 1)
}
