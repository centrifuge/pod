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
	res := createDocument(alice.httpExpect, alice.id.String(), typeEntity, http.StatusOK, defaultEntityPayload(alice.id.String(), []string{bob.id.String(), charlie.id.String()}))
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
