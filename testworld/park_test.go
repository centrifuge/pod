// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"
)

const TypeInvoice string = "invoice"
const TypePO string = "purchaseorder"

func TestHost_Happy(t *testing.T) {
	t.Parallel()
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// alice shares a document with bob and charlie
	res := createDocument(alice.httpExpect, TypeInvoice, http.StatusOK, defaultInvoicePayload([]string{bob.id.String(), charlie.id.String()}))


	docIdentifier := getDocumentIdentifier(t, res)

	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(alice.httpExpect,TypeInvoice, params)
	getDocumentAndCheck(bob.httpExpect,TypeInvoice, params)
	getDocumentAndCheck(charlie.httpExpect,TypeInvoice, params)
	fmt.Println("Host test success")
}
