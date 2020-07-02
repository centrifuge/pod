// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestV2GenericCreateAndCommit_new_document(t *testing.T) {
	t.Parallel()
	createNewDocument(t, func(dids []string) (map[string]interface{}, map[string]string) {
		return genericCoreAPICreate(dids), nil
	}, func(dids []string) (map[string]interface{}, map[string]string) {
		return genericCoreAPIUpdate(dids), nil
	})
}

func TestV2GenericCreate_next_version(t *testing.T) {
	t.Parallel()
	createNextDocument(t, genericCoreAPICreate)
}

func TestV2EntityCreateAndCommit_new_document(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice prepares document to share with Bob and charlie
	payload, params := createPayloadParams([]string{bob.id.String(), charlie.id.String()})
	res := createDocumentV2(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, payload)
	status := getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")

	checkDocumentParams(res, params)
	label := "signed_attribute"
	signedAttributeMissing(t, res, label)
	docID := getDocumentIdentifier(t, res)
	assert.NotEmpty(t, docID)

	// getting pending document should be successful
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusOK)

	// committed shouldn't be success
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "committed", http.StatusNotFound)

	// add a signed attribute
	value := hexutil.Encode(utils.RandomSlice(32))
	res = addSignedAttribute(alice.httpExpect, alice.id.String(), docID, label, value, "bytes")
	signedAttributeExists(t, res, label)

	// Alice updates the document
	payload, params = updatePayloadParams([]string{bob.id.String(), charlie.id.String()})
	payload["document_id"] = docID
	res = updateDocumentV2(alice.httpExpect, alice.id.String(), "documents", http.StatusOK, payload)
	status = getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")
	checkDocumentParams(res, params)
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusOK)

	// alice removes charlie from the list of collaborators
	removeCollaborators(alice.httpExpect, alice.id.String(), "documents", http.StatusOK, docID, charlie.id.String())

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

	// charlie should not have the document
	nonExistingGenericDocumentCheck(charlie.httpExpect, charlie.id.String(), docID)

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
	// alice should not have this version
	nonExistingDocumentVersionCheck(alice.httpExpect, alice.id.String(), docID, eversionID)

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
