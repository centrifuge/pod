// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"
)

func Test_CreateTransfer(t *testing.T) {
	t.Parallel()
	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	_, identifier := createInvoiceWithTransfer(t, alice, bob)
	fmt.Println(identifier)
	//listTest(t, alice, bob, charlie, identifier)
}

func createInvoiceWithTransfer(t *testing.T, alice, bob hostTestSuite) (transferId, docIdentifier string) {
	res := createDocument(alice.httpExpect, alice.id.String(), typeInvoice, http.StatusOK, defaultInvoicePayload([]string{bob.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(t, alice.httpExpect, alice.id.String(), typeInvoice, params, true)
	getDocumentAndCheck(t, bob.httpExpect, bob.id.String(), typeInvoice, params, true)

	// Alice creates a transfer designating Bob as the recipient
	res = createTransfer(alice.httpExpect, alice.id.String(), docIdentifier, http.StatusCreated, defaultTransferPayload(alice.id.String(), bob.id.String()))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	transferId = getTransferId(t, res)
	params = map[string]interface{}{
		"document_id": docIdentifier,
		"amount":      "300",
		"status":         "open",
		"scheduled_date": "2018-09-26T23:12:37Z",
	}

	// check if the transferAgreement is on the document
	getTransferAndCheck(alice.httpExpect, alice.id.String(), docIdentifier, transferId, params)
	return transferId, docIdentifier
}