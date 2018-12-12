// +build testworld

package testworld

import (
	"net/http"
	"testing"
)

func TestHost_AddExternalCollaborator_invoice(t *testing.T) {
	t.Parallel()
	addExternalCollaborator(t, TypeInvoice)
}

func TestHost_AddExternalCollaborator_po(t *testing.T) {
	t.Parallel()
	addExternalCollaborator(t, TypePO)
}

func addExternalCollaborator(t *testing.T, documentType string) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	res := createDocument(alice.httpExpect, documentType, http.StatusOK, defaultDocumentPayload(documentType, []string{bob.id.String()}))

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

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "EUR"
	getDocumentAndCheck(alice.httpExpect, documentType, params)
	getDocumentAndCheck(bob.httpExpect, documentType, params)
	getDocumentAndCheck(charlie.httpExpect, documentType, params)
}

func TestHost_CollaboratorTimeOut_invoice(t *testing.T) {
	t.Parallel()
	collaboratorTimeOut(t, TypeInvoice)
}

func TestHost_CollaboratorTimeOut_po(t *testing.T) {
	t.Parallel()
	collaboratorTimeOut(t, TypePO)
}

func collaboratorTimeOut(t *testing.T, documentType string) {

	kenny := doctorFord.getHostTestSuite(t, "Kenny")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Kenny shares a document with Bob
	response := createDocument(kenny.httpExpect, documentType, http.StatusOK, defaultInvoicePayload([]string{bob.id.String()}))

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
