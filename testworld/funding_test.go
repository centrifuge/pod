// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHost_Funding(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// alice shares a document with bob and charlie
	res := createDocument(alice.httpExpect, alice.id.String(), typeInvoice, http.StatusOK, defaultInvoicePayload([]string{bob.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(t, alice.httpExpect, alice.id.String(), typeInvoice, params, true)
	getDocumentAndCheck(t, bob.httpExpect, bob.id.String(), typeInvoice, params, true)
	fmt.Println("Host test success")

	// alice adds a funding
}
