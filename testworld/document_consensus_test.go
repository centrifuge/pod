// +build testworld

package testworld

import (
	"net/http"
	"strings"
	"testing"

	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/notification"
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
	res := createDocument(bob.httpExpect, a, documentType, http.StatusAccepted, defaultDocumentPayload(documentType, []string{b}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(bob.httpExpect, a, txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(t, bob.httpExpect, a, documentType, params, true)
	getDocumentAndCheck(t, bob.httpExpect, b, documentType, params, true)
	// account a completes job with a webhook
	msg, err := doctorFord.maeve.getReceivedMsg(a, int(notification.JobCompleted), txID)
	assert.NoError(t, err)
	assert.Equal(t, string(jobs.Success), msg.Status)

	// account b sends a webhook for received anchored doc
	msg, err = doctorFord.maeve.getReceivedMsg(b, int(notification.ReceivedPayload), docIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(a), strings.ToLower(msg.FromID))
	log.Debug("Host test success")
	nonExistingDocumentCheck(bob.httpExpect, c, documentType, params)

	// b updates invoice and shares with c as well
	res = updateDocument(bob.httpExpect, b, documentType, http.StatusAccepted, docIdentifier, updatedDocumentPayload(documentType, []string{a, c}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, b, txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "EUR"
	getDocumentAndCheck(t, bob.httpExpect, a, documentType, params, true)
	getDocumentAndCheck(t, bob.httpExpect, b, documentType, params, true)
	getDocumentAndCheck(t, bob.httpExpect, c, documentType, params, true)
	// account c sends a webhook for received anchored doc
	msg, err = doctorFord.maeve.getReceivedMsg(c, int(notification.ReceivedPayload), docIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(b), strings.ToLower(msg.FromID))
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
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusAccepted, defaultDocumentPayload(documentType, []string{a, b}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(t, alice.httpExpect, alice.id.String(), documentType, params, true)
	getDocumentAndCheck(t, bob.httpExpect, a, documentType, params, true)
	getDocumentAndCheck(t, bob.httpExpect, b, documentType, params, true)
	// alices main account completes job with a webhook
	msg, err := doctorFord.maeve.getReceivedMsg(alice.id.String(), int(notification.JobCompleted), txID)
	assert.NoError(t, err)
	assert.Equal(t, string(jobs.Success), msg.Status)

	// bobs account b sends a webhook for received anchored doc
	msg, err = doctorFord.maeve.getReceivedMsg(b, int(notification.ReceivedPayload), docIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(alice.id.String()), strings.ToLower(msg.FromID))
	nonExistingDocumentCheck(bob.httpExpect, c, documentType, params)

	// Bob updates invoice and shares with bobs account c as well using account a and to accounts d and e of Charlie
	res = updateDocument(bob.httpExpect, a, documentType, http.StatusAccepted, docIdentifier, updatedDocumentPayload(documentType, []string{alice.id.String(), b, c, d, e}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, a, txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "EUR"
	getDocumentAndCheck(t, alice.httpExpect, alice.id.String(), documentType, params, true)
	// bobs accounts all have the document now
	getDocumentAndCheck(t, bob.httpExpect, a, documentType, params, true)
	getDocumentAndCheck(t, bob.httpExpect, b, documentType, params, true)
	getDocumentAndCheck(t, bob.httpExpect, c, documentType, params, true)
	getDocumentAndCheck(t, charlie.httpExpect, d, documentType, params, true)
	getDocumentAndCheck(t, charlie.httpExpect, e, documentType, params, true)
	nonExistingDocumentCheck(charlie.httpExpect, f, documentType, params)
}

func addExternalCollaborator(t *testing.T, documentType string) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusAccepted, defaultDocumentPayload(documentType, []string{bob.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(t, alice.httpExpect, alice.id.String(), documentType, params, true)
	getDocumentAndCheck(t, bob.httpExpect, bob.id.String(), documentType, params, true)
	nonExistingDocumentCheck(charlie.httpExpect, charlie.id.String(), documentType, params)

	// Bob updates invoice and shares with Charlie as well
	res = updateDocument(bob.httpExpect, bob.id.String(), documentType, http.StatusAccepted, docIdentifier, updatedDocumentPayload(documentType, []string{alice.id.String(), charlie.id.String()}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, bob.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "EUR"
	getDocumentAndCheck(t, alice.httpExpect, alice.id.String(), documentType, params, true)
	getDocumentAndCheck(t, bob.httpExpect, bob.id.String(), documentType, params, true)
	getDocumentAndCheck(t, charlie.httpExpect, charlie.id.String(), documentType, params, true)
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
	response := createDocument(kenny.httpExpect, kenny.id.String(), documentType, http.StatusAccepted, defaultInvoicePayload([]string{bob.id.String()}))
	txID := getTransactionID(t, response)
	status, message := getTransactionStatusAndMessage(kenny.httpExpect, kenny.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// check if Bob and Kenny received the document
	docIdentifier := getDocumentIdentifier(t, response)
	paramsV1 := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(t, kenny.httpExpect, kenny.id.String(), documentType, paramsV1, true)
	getDocumentAndCheck(t, bob.httpExpect, bob.id.String(), documentType, paramsV1, true)

	// Kenny gets killed
	kenny.host.kill()

	// Bob updates and sends to Kenny
	updatedPayload := updatedDocumentPayload(documentType, []string{kenny.id.String()})

	// Bob will anchor the document without Kennys signature
	response = updateDocument(bob.httpExpect, bob.id.String(), documentType, http.StatusAccepted, docIdentifier, updatedPayload)
	txID = getTransactionID(t, response)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, bob.id.String(), txID)
	if status != "failed" {
		t.Error(message)
	}

	// check if bob saved the updated document
	paramsV2 := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "EUR",
	}
	getDocumentAndCheck(t, bob.httpExpect, bob.id.String(), documentType, paramsV2, true)

	// bring Kenny back to life
	doctorFord.reLive(t, kenny.name)

	// Kenny should NOT have latest version
	getDocumentAndCheck(t, kenny.httpExpect, kenny.id.String(), documentType, paramsV1, true)

}

func TestDocument_invalidAttributes(t *testing.T) {
	kenny := doctorFord.getHostTestSuite(t, "Kenny")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Kenny shares a document with Bob
	response := createDocument(kenny.httpExpect, kenny.id.String(), typeInvoice, http.StatusBadRequest, wrongInvoicePayload([]string{bob.id.String()}))

	errMsg := response.Raw()["message"].(string)
	assert.Contains(t, errMsg, "some invalid time stamp\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"some invalid ti")
}
