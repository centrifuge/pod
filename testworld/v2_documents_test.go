// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV2InvoiceCreateAndCommit_new_document(t *testing.T) {
	createNewDocument(t, func(dids []string) (map[string]interface{}, map[string]string) {
		params := map[string]string{
			"currency": "EUR",
			"number":   "12345",
		}
		return invoiceCoreAPICreate(dids), params
	}, func(dids []string) (map[string]interface{}, map[string]string) {
		// Alice updates the document
		payload := invoiceCoreAPIUpdate(dids)
		// update currency to USD and number to 56789
		data := payload["data"].(map[string]interface{})
		data["currency"] = "USD"
		data["number"] = "56789"
		payload["data"] = data
		return payload, map[string]string{
			"currency": "USD",
			"number":   "56789",
		}
	})
}

func TestV2InvoiceCreate_next_version(t *testing.T) {
	createNextDocument(t, invoiceCoreAPICreate)
}

func TestV2GenericCreateAndCommit_new_document(t *testing.T) {
	createNewDocument(t, func(dids []string) (map[string]interface{}, map[string]string) {
		return genericCoreAPICreate(dids), nil
	}, func(dids []string) (map[string]interface{}, map[string]string) {
		return genericCoreAPIUpdate(dids), nil
	})
}

func TestV2GenericCreate_next_version(t *testing.T) {
	createNextDocument(t, genericCoreAPICreate)
}

func TestV2EntityCreateAndCommit_new_document(t *testing.T) {
	createNewDocument(t, func(dids []string) (map[string]interface{}, map[string]string) {
		params := map[string]string{
			"legal_name": "test company",
			"identity":   dids[0],
		}
		return entityCoreAPICreate(dids[0], dids), params
	}, func(dids []string) (map[string]interface{}, map[string]string) {
		p := entityCoreAPIUpdate(dids)
		params := map[string]string{
			"legal_name": "updated company",
		}

		return p, params
	})
}

func TestV2EntityCreate_next_version(t *testing.T) {
	createNextDocument(t, func(dids []string) map[string]interface{} {
		var id string
		if len(dids) > 0 {
			id = dids[0]
		}
		return entityCoreAPICreate(id, dids)
	})
}

func createNewDocument(
	t *testing.T,
	createPayloadParams, updatePayloadParams func([]string) (map[string]interface{}, map[string]string)) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice prepares document to share with Bob
	payload, params := createPayloadParams([]string{bob.id.String()})
	res := createDocumentV2(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, payload)
	status := getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")

	checkDocumentParams(res, params)
	docID := getDocumentIdentifier(t, res)
	assert.NotEmpty(t, docID)

	// getting pending document should be successful
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusOK)

	// committed shouldn't be success
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "committed", http.StatusNotFound)

	// Alice updates the document
	payload, params = updatePayloadParams([]string{bob.id.String()})
	payload["document_id"] = docID
	res = updateDocumentV2(alice.httpExpect, alice.id.String(), "documents", http.StatusOK, payload)
	status = getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")
	checkDocumentParams(res, params)
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusOK)

	// Commits document and shares with Bob
	res = commitDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, docID)
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	assert.Equal(t, status, "success", message)
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docID, nil, updateAttributes())

	// pending document should fail
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusNotFound)

	// committed should be successful
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "committed", http.StatusOK)

	// Bob should have the document
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docID, nil, updateAttributes())

	// try to commit same document again - failure
	commitDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusBadRequest, docID)
}

func createNextDocument(t *testing.T, createPayload func([]string) map[string]interface{}) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Alice shares document with Bob
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, createPayload([]string{bob.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	assert.Equal(t, status, "success", message)
	docID := getDocumentIdentifier(t, res)
	versionID := getDocumentCurrentVersion(t, res)
	assert.Equal(t, docID, versionID, "failed to create a fresh document")
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docID, nil, createAttributes())

	// there should be no pending document with alice
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusNotFound)

	// bob creates a next pending version of the document
	payload := createPayload(nil)
	payload["document_id"] = docID
	res = createDocumentV2(bob.httpExpect, bob.id.String(), "documents", http.StatusCreated, payload)
	status = getDocumentStatus(t, res)
	assert.Equal(t, status, "pending", "document must be in pending status")
	edocID := getDocumentIdentifier(t, res)
	assert.Equal(t, docID, edocID, "document identifiers mismatch")
	eversionID := getDocumentCurrentVersion(t, res)
	assert.NotEqual(t, docID, eversionID, "document ID and versionID must not be equal")
	params := map[string]interface{}{
		"document_id": docID,
		"version_id":  eversionID,
	}

	// alice should not have this version
	nonExistingDocumentVersionCheck(alice.httpExpect, alice.id.String(), "documents", params)

	// bob has pending document
	getV2DocumentWithStatus(bob.httpExpect, bob.id.String(), docID, "pending", http.StatusOK)

	// commit the document
	// Commits document and shares with alice
	res = commitDocument(bob.httpExpect, bob.id.String(), "documents", http.StatusAccepted, docID)
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, bob.id.String(), txID)
	assert.Equal(t, status, "success", message)

	// bob shouldn't have any pending documents but has a committed one
	getV2DocumentWithStatus(bob.httpExpect, bob.id.String(), docID, "pending", http.StatusNotFound)
	getV2DocumentWithStatus(bob.httpExpect, bob.id.String(), docID, "committed", http.StatusOK)
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "committed", http.StatusOK)
}
