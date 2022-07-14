//go:build testworld
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
	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(), defaultEntityPayload(alice.id.ToHexString(),
		[]string{bob.id.ToHexString(), charlie.id.ToHexString()}))

	params := map[string]interface{}{
		"legal_name": "test company",
	}
	getDocumentAndVerify(t, alice.httpExpect, alice.id.ToHexString(), docID, params, nil)
	getDocumentAndVerify(t, bob.httpExpect, bob.id.ToHexString(), docID, params, nil)
	getDocumentAndVerify(t, charlie.httpExpect, charlie.id.ToHexString(), docID, params, nil)
}

func TestHost_EntityShareGet(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice anchors Entity
	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(), defaultEntityPayload(alice.id.ToHexString(), []string{}))

	// Alice creates an EntityRelationship with Bob
	relID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(),
		defaultRelationshipPayload(alice.id.ToHexString(), docID, bob.id.ToHexString()))

	response := getEntityWithRelation(bob.httpExpect, bob.id.ToHexString(), relID)
	response.Path("$.data.legal_name").String().Equal("test company")
}
