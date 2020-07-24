// +build testworld

package testworld

import (
	"net/http"
	"testing"
)

func Test_CreateGetUpdateTransfers(t *testing.T) {
	t.Parallel()
	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	transferID, identifier := createDocumentWithTransfer(t, alice, bob)
	listTransfer(t, alice, bob, identifier)
	testUpdateTransfer(t, alice, bob, identifier, transferID)
}

func createDocumentWithTransfer(t *testing.T, alice, bob hostTestSuite) (transferId, docIdentifier string) {
	res := createDocument(alice.httpExpect, alice.id.String(), typeDocuments, http.StatusAccepted, genericCoreAPICreate([]string{bob.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, nil, createAttributes())

	// Alice creates a transfer designating Bob as the recipient
	res = createTransfer(alice.httpExpect, alice.id.String(), docIdentifier, http.StatusAccepted, defaultTransferPayload(alice.id.String(), bob.id.String()))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	transferId = getTransferID(t, res)
	params := map[string]interface{}{
		"document_id":    docIdentifier,
		"amount":         "300",
		"status":         "open",
		"scheduled_date": "2018-09-26T23:12:37Z",
	}

	// check if the transferAgreement is on the document
	getTransferAndCheck(alice.httpExpect, alice.id.String(), docIdentifier, transferId, params)
	return transferId, docIdentifier
}

func listTransfer(t *testing.T, alice, bob hostTestSuite, docIdentifier string) {
	var transfers []string
	for i := 0; i < 5; i++ {
		res := createTransfer(alice.httpExpect, alice.id.String(), docIdentifier, http.StatusAccepted, defaultTransferPayload(alice.id.String(), bob.id.String()))
		txID := getTransactionID(t, res)
		status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
		if status != "success" {
			t.Error(message)
		}

		transferId := getTransferID(t, res)
		transfers = append(transfers, transferId)
	}
	params := map[string]interface{}{
		"document_id":    docIdentifier,
		"amount":         "300",
		"status":         "open",
		"scheduled_date": "2018-09-26T23:12:37Z",
	}
	getListTransfersCheck(alice.httpExpect, alice.id.String(), docIdentifier, 6, params)
	getListTransfersCheck(bob.httpExpect, bob.id.String(), docIdentifier, 6, params)
}

func testUpdateTransfer(t *testing.T, alice, bob hostTestSuite, docID, transferID string) {
	res := updateTransfer(alice.httpExpect, alice.id.String(), http.StatusAccepted, docID, transferID, updateTransferPayload(alice.id.String(), bob.id.String()))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}
	params := map[string]interface{}{
		"document_id":    docID,
		"amount":         "400",
		"status":         "settled",
		"scheduled_date": "2018-09-26T23:12:37Z",
	}
	getTransferAndCheck(alice.httpExpect, alice.id.String(), docID, transferID, params)
}
