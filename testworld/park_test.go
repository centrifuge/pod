// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHost_Happy(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// alice shares a document with bob and charlie
	res, err := alice.host.createInvoice(alice.httpExpect, http.StatusOK, defaultInvoicePayload([]string{bob.id.String(), charlie.id.String()}))
	if err != nil {
		t.Error(err)
	}

	docIdentifier := getDocumentIdentifier(t, res)

	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getInvoiceAndCheck(alice.httpExpect, params)
	getInvoiceAndCheck(bob.httpExpect, params)
	getInvoiceAndCheck(charlie.httpExpect, params)
	fmt.Println("Host test success")
}
