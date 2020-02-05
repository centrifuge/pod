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
			typeDocuments,
			multiHost,
		},
		{
			"Invoice_withinhost_AddExternalCollaborator",
			typeDocuments,
			withinHost,
		},
		{
			"Invoice_multiHostMultiAccount_AddExternalCollaborator",
			typeDocuments,
			multiHostMultiAccount,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			switch test.testType {
			case multiHost:
				addExternalCollaborator(t, test.docType)
			case multiHostMultiAccount:
				addExternalCollaboratorMultiHostMultiAccount(t, test.docType)
			case withinHost:
				addExternalCollaboratorWithinHost(t, test.docType)
			}
		})
	}
}

func addExternalCollaboratorWithinHost(t *testing.T, documentType string) {
	bob := doctorFord.getHostTestSuite(t, "Bob")
	accounts := doctorFord.getHost("Bob").accounts
	a := accounts[0]
	b := accounts[1]
	c := accounts[2]

	// a shares document with b first
	res := createDocument(bob.httpExpect, a, documentType, http.StatusAccepted, genericCoreAPICreate([]string{b}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(bob.httpExpect, a, txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	getGenericDocumentAndCheck(t, bob.httpExpect, a, docIdentifier, nil, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, b, docIdentifier, nil, createAttributes())
	// account a completes job with a webhook
	msg, err := doctorFord.maeve.getReceivedMsg(a, int(notification.JobCompleted), txID)
	assert.NoError(t, err)
	assert.Equal(t, string(jobs.Success), msg.Status)

	// account b sends a webhook for received anchored doc
	msg, err = doctorFord.maeve.getReceivedMsg(b, int(notification.ReceivedPayload), docIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(a), strings.ToLower(msg.FromID))
	log.Debug("Host test success")
	nonExistingDocumentCheck(bob.httpExpect, c, docIdentifier)

	// b updates invoice and shares with c as well
	res = updateDocument(bob.httpExpect, b, documentType, http.StatusAccepted, docIdentifier, genericCoreAPIUpdate([]string{a, c}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, b, txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	getGenericDocumentAndCheck(t, bob.httpExpect, a, docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, b, docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, c, docIdentifier, nil, allAttributes())
	// account c sends a webhook for received anchored doc
	msg, err = doctorFord.maeve.getReceivedMsg(c, int(notification.ReceivedPayload), docIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(b), strings.ToLower(msg.FromID))
}

func addExternalCollaboratorMultiHostMultiAccount(t *testing.T, documentType string) {
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
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusAccepted, genericCoreAPICreate([]string{a, b}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, a, docIdentifier, nil, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, b, docIdentifier, nil, createAttributes())
	// alices main account completes job with a webhook
	msg, err := doctorFord.maeve.getReceivedMsg(alice.id.String(), int(notification.JobCompleted), txID)
	assert.NoError(t, err)
	assert.Equal(t, string(jobs.Success), msg.Status)

	// bobs account b sends a webhook for received anchored doc
	msg, err = doctorFord.maeve.getReceivedMsg(b, int(notification.ReceivedPayload), docIdentifier)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(alice.id.String()), strings.ToLower(msg.FromID))
	nonExistingDocumentCheck(bob.httpExpect, c, docIdentifier)

	// Bob updates invoice and shares with bobs account c as well using account a and to accounts d and e of Charlie
	res = updateDocument(bob.httpExpect, a, documentType, http.StatusAccepted, docIdentifier, genericCoreAPIUpdate([]string{alice.id.String(), b, c, d, e}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, a, txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, allAttributes())
	// bobs accounts all have the document now
	getGenericDocumentAndCheck(t, bob.httpExpect, a, docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, b, docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, c, docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, charlie.httpExpect, d, docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, charlie.httpExpect, e, docIdentifier, nil, allAttributes())
	nonExistingDocumentCheck(charlie.httpExpect, f, docIdentifier)
}

func addExternalCollaborator(t *testing.T, documentType string) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusAccepted, genericCoreAPICreate([]string{bob.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, nil, createAttributes())
	nonExistingDocumentCheck(charlie.httpExpect, charlie.id.String(), docIdentifier)

	// Bob updates invoice and shares with Charlie as well
	res = updateDocument(bob.httpExpect, bob.id.String(), documentType, http.StatusAccepted, docIdentifier, genericCoreAPIUpdate([]string{alice.id.String(), charlie.id.String()}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, bob.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, charlie.httpExpect, charlie.id.String(), docIdentifier, nil, allAttributes())
}

func TestHost_CollaboratorTimeOut(t *testing.T) {
	// Run only locally since this creates resource issues for the entire test suite
	t.SkipNow()
	t.Parallel()

	//currently can't be run in parallel (because of node kill)
	collaboratorTimeOut(t, typeDocuments)
}

func collaboratorTimeOut(t *testing.T, documentType string) {
	kenny := doctorFord.getHostTestSuite(t, "Kenny")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Kenny shares a document with Bob
	response := createDocument(kenny.httpExpect, kenny.id.String(), documentType, http.StatusAccepted, genericCoreAPICreate([]string{bob.id.String()}))
	txID := getTransactionID(t, response)
	status, message := getTransactionStatusAndMessage(kenny.httpExpect, kenny.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	// check if Bob and Kenny received the document
	docIdentifier := getDocumentIdentifier(t, response)
	getGenericDocumentAndCheck(t, kenny.httpExpect, kenny.id.String(), docIdentifier, nil, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, nil, createAttributes())

	// Kenny gets killed
	kenny.host.kill()

	// Bob updates and sends to Kenny
	updatedPayload := genericCoreAPIUpdate([]string{kenny.id.String()})

	// Bob will anchor the document without Kennys signature
	response = updateDocument(bob.httpExpect, bob.id.String(), documentType, http.StatusAccepted, docIdentifier, updatedPayload)
	txID = getTransactionID(t, response)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, bob.id.String(), txID)
	if status != "failed" {
		t.Error(message)
	}

	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, nil, allAttributes())

	// bring Kenny back to life
	doctorFord.reLive(t, kenny.name)

	// Kenny should NOT have latest version
	getGenericDocumentAndCheck(t, kenny.httpExpect, kenny.id.String(), docIdentifier, nil, createAttributes())
}

func TestDocument_invalidAttributes(t *testing.T) {
	kenny := doctorFord.getHostTestSuite(t, "Kenny")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Kenny shares a document with Bob
	response := createDocument(kenny.httpExpect, kenny.id.String(), typeDocuments, http.StatusBadRequest, wrongGenericDocumentPayload([]string{bob.id.String()}))
	errMsg := response.Raw()["message"].(string)
	assert.Contains(t, errMsg, "some invalid time stamp\" as \"2006-01-02T15:04:05.999999999Z07:00\": cannot parse \"some invalid ti")
}

func TestDocument_latestDocumentVersion(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")
	kenny := doctorFord.getHostTestSuite(t, "Kenny")
	documentType := typeDocuments

	// alice creates a document with bob and kenny
	res := createDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusAccepted, genericCoreAPICreate([]string{bob.id.String(), kenny.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	versionID := getDocumentCurrentVersion(t, res)
	if versionID != docIdentifier {
		t.Errorf("docID(%s) != versionID(%s)\n", docIdentifier, versionID)
	}

	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, nil, createAttributes())
	getGenericDocumentAndCheck(t, kenny.httpExpect, kenny.id.String(), docIdentifier, nil, createAttributes())
	nonExistingDocumentCheck(charlie.httpExpect, charlie.id.String(), docIdentifier)

	// Bob updates invoice and shares with Charlie as well but kenny is offline and miss the update
	kenny.host.kill()
	res = updateDocument(bob.httpExpect, bob.id.String(), documentType, http.StatusAccepted, docIdentifier, genericCoreAPIUpdate([]string{charlie.id.String()}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, bob.id.String(), txID)
	if status != "failed" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	versionID = getDocumentCurrentVersion(t, res)
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, charlie.httpExpect, charlie.id.String(), docIdentifier, nil, allAttributes())
	// bring kenny back and should not have the latest version
	doctorFord.reLive(t, kenny.name)
	nonExistingDocumentVersionCheck(kenny.httpExpect, kenny.id.String(), docIdentifier, versionID)

	// alice updates document
	res = updateDocument(alice.httpExpect, alice.id.String(), documentType, http.StatusAccepted, docIdentifier, genericCoreAPIUpdate(nil))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	versionID = getDocumentCurrentVersion(t, res)

	// everyone should have the latest version
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, charlie.httpExpect, charlie.id.String(), docIdentifier, nil, allAttributes())
	getGenericDocumentAndCheck(t, kenny.httpExpect, kenny.id.String(), docIdentifier, nil, allAttributes())
}
