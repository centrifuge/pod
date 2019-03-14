// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHost_AddExternalCollaborator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		docType  string
		testType testType
	}{
		{
			"Invoice_multiHost_AddExternalCollaborator",
			typeInvoice,
			multiHost,
		},
		{
			"Invoice_withinhost_AddExternalCollaborator",
			typeInvoice,
			withinHost,
		},
		{
			"Invoice_multiHostMultiAccount_AddExternalCollaborator",
			typeInvoice,
			multiHostMultiAccount,
		},
		{
			"PO_AddExternalCollaborator",
			typePO,
			multiHost,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.testType {
			case multiHost:
				addExternalCollaborator(t, test.docType)
			case multiHostMultiAccount:
				addExternalCollaborator_multiHostMultiAccount(t, test.docType)
			case withinHost:
				addExternalCollaborator_withinHost(t, test.docType)
			}
		})
	}
}

func addExternalCollaborator_withinHost(t *testing.T, documentType string) {
	bob := doctorFord.getHostTestSuite(t, "Bob")
	accounts := doctorFord.getHost("Bob").accounts
	a := accounts[0]
	b := accounts[1]
	c := accounts[2]

	// a shares document with b first
	res := createDocument(bob.httpExpect, a, documentType, http.StatusOK, defaultDocumentPayload(documentType, []string{b}))
	txID := getTransactionID(t, res)
	waitTillStatus(t, bob.httpExpect, a, txID, "success")

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(bob.httpExpect, a, documentType, params)
	getDocumentAndCheck(bob.httpExpect, b, documentType, params)
	nonExistingDocumentCheck(bob.httpExpect, c, documentType, params)

	//// let c update the document and fail
	res = failedUpdateDocument(bob.httpExpect, c, documentType, http.StatusInternalServerError, docIdentifier, updatedDocumentPayload(documentType, []string{a, c}))
	assert.NotNil(t, res)

	// b updates invoice and shares with c as well
	res = updateDocument(bob.httpExpect, b, documentType, http.StatusOK, docIdentifier, updatedDocumentPayload(documentType, []string{a, c}))
	txID = getTransactionID(t, res)
	waitTillStatus(t, bob.httpExpect, b, txID, "success")

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "EUR"
	getDocumentAndCheck(bob.httpExpect, a, documentType, params)
	getDocumentAndCheck(bob.httpExpect, b, documentType, params)
	getDocumentAndCheck(bob.httpExpect, c, documentType, params)
}

func addExternalCollaborator_multiHostMultiAccount(t *testing.T, documentType string) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	accounts := doctorFord.getHost("Bob").accounts
	a := accounts[0]
	b := accounts[1]
	c := accounts[2]
	charlie := doctorFord.getHostTestSuite(t, "Charlie")
	accounts2 := doctorFord.getHost("Charlie").accounts
	d := accounts2[0]
	e := accounts2[1]
	f := accounts2[2]

	// Alice shares document with Bobs accounts a and b
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusOK, defaultDocumentPayload(documentType, []string{a, b}))
	txID := getTransactionID(t, res)
	waitTillStatus(t, alice.httpExpect, alice.id.String(), txID, "success")

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(alice.httpExpect, alice.id.String(), documentType, params)
	getDocumentAndCheck(bob.httpExpect, a, documentType, params)
	getDocumentAndCheck(bob.httpExpect, b, documentType, params)
	nonExistingDocumentCheck(bob.httpExpect, c, documentType, params)

	//// let c update the document and fail
	res = failedUpdateDocument(bob.httpExpect, c, documentType, http.StatusInternalServerError, docIdentifier, updatedDocumentPayload(documentType, []string{alice.id.String(), b, c, d, e}))
	assert.NotNil(t, res)

	// Bob updates invoice and shares with bobs account c as well using account a and to accounts d and e of Charlie
	res = updateDocument(bob.httpExpect, a, documentType, http.StatusOK, docIdentifier, updatedDocumentPayload(documentType, []string{alice.id.String(), b, c, d, e}))
	txID = getTransactionID(t, res)
	waitTillStatus(t, bob.httpExpect, a, txID, "success")

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "EUR"
	getDocumentAndCheck(alice.httpExpect, alice.id.String(), documentType, params)
	// bobs accounts all have the document now
	getDocumentAndCheck(bob.httpExpect, a, documentType, params)
	getDocumentAndCheck(bob.httpExpect, b, documentType, params)
	getDocumentAndCheck(bob.httpExpect, c, documentType, params)
	getDocumentAndCheck(charlie.httpExpect, d, documentType, params)
	getDocumentAndCheck(charlie.httpExpect, e, documentType, params)
	nonExistingDocumentCheck(charlie.httpExpect, f, documentType, params)
}

func addExternalCollaborator(t *testing.T, documentType string) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusOK, defaultDocumentPayload(documentType, []string{bob.id.String()}))
	txID := getTransactionID(t, res)
	waitTillStatus(t, alice.httpExpect, alice.id.String(), txID, "success")

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(alice.httpExpect, alice.id.String(), documentType, params)
	getDocumentAndCheck(bob.httpExpect, bob.id.String(), documentType, params)
	nonExistingDocumentCheck(charlie.httpExpect, charlie.id.String(), documentType, params)

	//// let charlie update the document and fail
	//res = failedUpdateDocument(charlie.httpExpect, charlie.id.String(), documentType, http.StatusInternalServerError, docIdentifier, updatedDocumentPayload(documentType, []string{alice.id.String(), charlie.id.String()}))
	//assert.NotNil(t, res)

	// Bob updates invoice and shares with Charlie as well
	res = updateDocument(bob.httpExpect, bob.id.String(), documentType, http.StatusOK, docIdentifier, updatedDocumentPayload(documentType, []string{alice.id.String(), charlie.id.String()}))
	txID = getTransactionID(t, res)
	waitTillStatus(t, bob.httpExpect, bob.id.String(), txID, "success")

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "EUR"
	getDocumentAndCheck(alice.httpExpect, alice.id.String(), documentType, params)
	getDocumentAndCheck(bob.httpExpect, bob.id.String(), documentType, params)
	getDocumentAndCheck(charlie.httpExpect, charlie.id.String(), documentType, params)
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
	response := createDocument(kenny.httpExpect, kenny.id.String(), documentType, http.StatusOK, defaultInvoicePayload([]string{bob.id.String()}))
	txID := getTransactionID(t, response)
	waitTillStatus(t, kenny.httpExpect, kenny.id.String(), txID, "success")

	// check if Bob and Kenny received the document
	docIdentifier := getDocumentIdentifier(t, response)
	paramsV1 := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(kenny.httpExpect, kenny.id.String(), documentType, paramsV1)
	getDocumentAndCheck(bob.httpExpect, bob.id.String(), documentType, paramsV1)

	// Kenny gets killed
	kenny.host.kill()

	// Bob updates and sends to Kenny
	updatedPayload := updatedDocumentPayload(documentType, []string{kenny.id.String()})

	// Bob will anchor the document without Kennys signature
	response = updateDocument(bob.httpExpect, bob.id.String(), documentType, http.StatusOK, docIdentifier, updatedPayload)
	txID = getTransactionID(t, response)
	waitTillStatus(t, bob.httpExpect, bob.id.String(), txID, "failed")

	// check if bob saved the updated document
	paramsV2 := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "EUR",
	}
	getDocumentAndCheck(bob.httpExpect, bob.id.String(), documentType, paramsV2)

	// bring Kenny back to life
	doctorFord.reLive(t, kenny.name)

	// Kenny should NOT have latest version
	getDocumentAndCheck(kenny.httpExpect, kenny.id.String(), documentType, paramsV1)

}
