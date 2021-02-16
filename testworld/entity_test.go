// +build testworld

package testworld

import (
	"testing"
)

func TestHost_BasicEntity(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares a document with Bob and Charlie
	docID := createAndCommitDocument(t, alice.httpExpect, alice.id.String(), defaultEntityPayload(alice.id.String(),
		[]string{bob.id.String(), charlie.id.String()}))

	params := map[string]interface{}{
		"legal_name": "test company",
	}
	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, params, nil)
	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, params, nil)
	getDocumentAndVerify(t, charlie.httpExpect, charlie.id.String(), docID, params, nil)
}

func TestHost_EntityShareGet(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice anchors Entity
	docID := createAndCommitDocument(t, alice.httpExpect, alice.id.String(), defaultEntityPayload(alice.id.String(), []string{}))

	// Alice creates an EntityRelationship with Bob
	relID := createAndCommitDocument(t, alice.httpExpect, alice.id.String(),
		defaultRelationshipPayload(alice.id.String(), docID, bob.id.String()))

	response := getEntityWithRelation(bob.httpExpect, bob.id.String(), relID)
	response.Path("$.data.legal_name").String().Equal("test company")
}
