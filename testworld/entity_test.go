// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHost_BasicEntity(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares a document with Bob and Charlie
	res := createDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusAccepted, defaultEntityPayload(alice.id.String(), []string{bob.id.String(), charlie.id.String()}))
	docIdentifier := getDocumentIdentifier(t, res)
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"legal_name":  "test company",
	}
	getEntityAndCheck(alice.httpExpect, alice.id.String(), typeEntity, params)
	getEntityAndCheck(bob.httpExpect, bob.id.String(), typeEntity, params)
	getEntityAndCheck(charlie.httpExpect, charlie.id.String(), typeEntity, params)
	fmt.Println("Host test success")
}

func TestHost_EntityShareGet(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice anchors Entity
	res := createDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusAccepted, defaultEntityPayload(alice.id.String(), []string{}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}
	entityIdentifier := getDocumentIdentifier(t, res)

	// Alice creates an EntityRelationship with Bob
	resB := shareEntity(alice.httpExpect, alice.id.String(), entityIdentifier, http.StatusAccepted, defaultRelationshipPayload(entityIdentifier, bob.id.String()))
	relationshipIdentifier := getDocumentIdentifier(t, resB)
	txID = getTransactionID(t, resB)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}
	params := map[string]interface{}{
		"r_identifier": relationshipIdentifier,
	}
	response := getEntityWithRelation(bob.httpExpect, bob.id.String(), typeEntity, params)
	response.Path("$.data.entity.legal_name").String().Equal("test company")
}
