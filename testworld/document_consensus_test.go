// +build testworld

package testworld

import (
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestHost_AddExternalCollaborator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		docType string
	}{
		{
			"Invoice_AddExternalCollaborator",
			typeInvoice,
		},
		{
			"PO_AddExternalCollaborator",
			typePO,
		},
	}
	for _, test := range tests {
		t.Run(test.docType, func(t *testing.T) {
			t.Parallel()
			addExternalCollaborator(t, test.docType)
		})
	}
}

func addExternalCollaborator(t *testing.T, documentType string) {
	// TODO remove this when we have retry for tasks
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	res := createDocument(alice.httpExpect, documentType, http.StatusOK, defaultDocumentPayload(documentType, []string{bob.id.String()}))
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
	getDocumentAndCheck(alice.httpExpect, documentType, params)
	getDocumentAndCheck(bob.httpExpect, documentType, params)

	// Bob updates invoice and shares with Charlie as well
	res = updateDocument(bob.httpExpect, documentType, http.StatusOK, docIdentifier, updatedDocumentPayload(documentType, []string{alice.id.String(), charlie.id.String()}))
	txID = getTransactionID(t, res)
	waitTillSuccess(t, bob.httpExpect, txID)

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "EUR"
	getDocumentAndCheck(alice.httpExpect, documentType, params)
	getDocumentAndCheck(bob.httpExpect, documentType, params)
	getDocumentAndCheck(charlie.httpExpect, documentType, params)
}

func TestHost_CollaboratorTimeOut(t *testing.T) {
	// Run only locally since this creates resource issues for the entire test suite
	t.SkipNow()
	t.Parallel()

	//currently can't be run in parallel (because of node kill)
	collaboratorTimeOut(t, typeInvoice)
	collaboratorTimeOut(t, typePO)
}

func collaboratorTimeOut(t *testing.T, documentType string) {

	kenny := doctorFord.getHostTestSuite(t, "Kenny")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Kenny shares a document with Bob
	response := createDocument(kenny.httpExpect, documentType, http.StatusOK, defaultInvoicePayload([]string{bob.id.String()}))
	txID := getTransactionID(t, response)
	waitTillSuccess(t, kenny.httpExpect, txID)

	// check if Bob and Kenny received the document
	docIdentifier := getDocumentIdentifier(t, response)
	paramsV1 := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(kenny.httpExpect, documentType, paramsV1)
	getDocumentAndCheck(bob.httpExpect, documentType, paramsV1)

	// Kenny gets killed
	kenny.host.kill()

	// Bob updates and sends to Alice
	updatedPayload := updatedDocumentPayload(documentType, []string{kenny.id.String()})

	// Bob will anchor the document without Alice signature but will receive an error because kenny is dead
	response = updateDocument(bob.httpExpect, documentType, http.StatusInternalServerError, docIdentifier, updatedPayload)
	txID = getTransactionID(t, response)
	waitTillSuccess(t, bob.httpExpect, txID)

	// check if bob saved the updated document
	paramsV2 := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "EUR",
	}
	getDocumentAndCheck(bob.httpExpect, documentType, paramsV2)

	// bring Kenny back to life
	doctorFord.reLive(t, kenny.name)

	// Kenny should NOT have latest version
	getDocumentAndCheck(kenny.httpExpect, documentType, paramsV1)

}
