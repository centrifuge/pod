// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHost_Happy(t *testing.T) {
	t.Parallel()
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// alice shares a document with bob and charlie
	res := createDocument(alice.httpExpect, typeInvoice, http.StatusOK, defaultInvoicePayload([]string{bob.id.String(), charlie.id.String()}))
	txID := getTransactionID(t, res)
	waitTillSuccess(t, alice.httpExpect, txID)

	docIdentifier := getDocumentIdentifier(t, res)

	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(alice.httpExpect, typeInvoice, params)
	getDocumentAndCheck(bob.httpExpect, typeInvoice, params)
	getDocumentAndCheck(charlie.httpExpect, typeInvoice, params)
	fmt.Println("Host test success")
}
