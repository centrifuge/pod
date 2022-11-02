//go:build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/centrifuge/go-centrifuge/testworld/park/behavior/client"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestDocumentsAPI_NewPendingDocument(t *testing.T) {
	alice, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)
	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)
	charlie, err := controller.GetHost(host.Charlie)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)
	bobClient, err := controller.GetClientForHost(t, host.Bob)
	assert.NoError(t, err)
	charlieClient, err := controller.GetClientForHost(t, host.Charlie)
	assert.NoError(t, err)

	// Alice prepares document to share with Bob and Charlie
	payload := entityCoreAPICreate(
		alice.GetMainAccount().GetAccountID().ToHexString(),
		[]string{
			bob.GetMainAccount().GetAccountID().ToHexString(),
			charlie.GetMainAccount().GetAccountID().ToHexString(),
		},
	)

	res := aliceClient.CreateDocument("documents", http.StatusCreated, payload)
	status := client.GetDocumentStatus(res)
	assert.Equal(t, status, "pending")

	params := map[string]string{
		"legal_name": "test company",
		"identity":   alice.GetMainAccount().GetAccountID().ToHexString(),
	}

	client.CheckDocumentParams(res, params)

	label := "signed_attribute"
	client.SignedAttributeMissing(t, res, label)

	docID := client.GetDocumentIdentifier(res)
	assert.NotEmpty(t, docID)

	// getting pending document should be successful
	aliceClient.GetDocumentWithStatus(docID, "pending", http.StatusOK)

	// committed shouldn't be success
	aliceClient.GetDocumentWithStatus(docID, "committed", http.StatusNotFound)

	// add a signed attribute
	value := hexutil.Encode(utils.RandomSlice(32))
	res = aliceClient.AddSignedAttribute(docID, label, value, "bytes")
	client.SignedAttributeExists(t, res, label)

	// Alice updates the document
	payload = entityCoreAPIUpdate([]string{
		bob.GetMainAccount().GetAccountID().ToHexString(),
		charlie.GetMainAccount().GetAccountID().ToHexString(),
	})
	payload["document_id"] = docID

	res = aliceClient.UpdateDocument("documents", http.StatusOK, payload)

	status = client.GetDocumentStatus(res)
	assert.Equal(t, status, "pending")

	params["legal_name"] = "updated company"

	client.CheckDocumentParams(res, params)

	aliceClient.GetDocumentWithStatus(docID, "pending", http.StatusOK)

	// Alice removes Charlie from the list of collaborators
	charlieAccountIDHex := charlie.GetMainAccount().GetAccountID().ToHexString()
	aliceClient.RemoveCollaborators("documents", http.StatusOK, docID, charlieAccountIDHex)

	// Commits document and shares with Bob
	res = aliceClient.CommitDocument("documents", http.StatusAccepted, docID)

	jobID, err := client.GetJobID(res)
	assert.NoError(t, err)

	err = aliceClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	aliceClient.GetDocumentAndVerify(docID, nil, updateAttributes())

	// pending document should fail
	aliceClient.GetDocumentWithStatus(docID, "pending", http.StatusNotFound)

	// committed should be successful
	aliceClient.GetDocumentWithStatus(docID, "committed", http.StatusOK)

	// Bob should have the document
	bobClient.GetDocumentAndVerify(docID, nil, updateAttributes())

	// Charlie should not have the document
	charlieClient.NonExistingDocumentCheck(docID)

	// try to commit same document again - failure
	aliceClient.CommitDocument("documents", http.StatusBadRequest, docID)
}

func TestDocumentsAPI_NextVersionCreatedByCollaborator(t *testing.T) {
	alice, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)
	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)
	bobClient, err := controller.GetClientForHost(t, host.Bob)
	assert.NoError(t, err)

	// Alice shares document with Bob
	payload := entityCoreAPICreate(
		alice.GetMainAccount().GetAccountID().ToHexString(),
		[]string{
			bob.GetMainAccount().GetAccountID().ToHexString(),
		},
	)

	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	res := bobClient.GetDocumentAndVerify(docID, nil, createAttributes())

	versionID := client.GetDocumentCurrentVersion(res.Object())
	assert.Equal(t, docID, versionID, "failed to create a fresh document")

	// There should be no pending document with Alice
	aliceClient.GetDocumentWithStatus(docID, "pending", http.StatusNotFound)

	// Bob creates a next pending version of the document
	payload = entityCoreAPICreate(
		alice.GetMainAccount().GetAccountID().ToHexString(),
		[]string{
			bob.GetMainAccount().GetAccountID().ToHexString(),
		},
	)
	payload["document_id"] = docID

	doc := bobClient.CreateDocument("documents", http.StatusCreated, payload)

	status := client.GetDocumentStatus(doc)
	assert.Equal(t, status, "pending", "document must be in pending status")

	edocID := client.GetDocumentIdentifier(doc)
	assert.Equal(t, docID, edocID, "document identifiers mismatch")

	eversionID := client.GetDocumentCurrentVersion(doc)
	assert.NotEqual(t, docID, eversionID, "document ID and versionID must not be equal")

	// Alice should not have this version
	aliceClient.NonExistingDocumentVersionCheck(docID, eversionID)

	// Bob has pending document
	bobClient.GetDocumentWithStatus(docID, "pending", http.StatusOK)

	// Commits document and shares with alice
	doc = bobClient.CommitDocument("documents", http.StatusAccepted, docID)

	jobID, err := client.GetJobID(doc)
	assert.NoError(t, err)

	err = bobClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	// bob shouldn't have any pending documents but has a committed one
	bobClient.GetDocumentWithStatus(docID, "pending", http.StatusNotFound)
	bobClient.GetDocumentWithStatus(docID, "committed", http.StatusOK)
	aliceClient.GetDocumentWithStatus(docID, "committed", http.StatusOK)
}

func TestDocumentsAPI_ClonedDocument(t *testing.T) {
	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)
	bobClient, err := controller.GetClientForHost(t, host.Bob)
	assert.NoError(t, err)

	// Alice prepares document to share with Bob
	payload := genericCoreAPICreate(
		[]string{
			bob.GetMainAccount().GetAccountID().ToHexString(),
		},
	)

	res := aliceClient.CreateDocument("documents", http.StatusCreated, payload)
	status := client.GetDocumentStatus(res)
	assert.Equal(t, status, "pending")

	docID := client.GetDocumentIdentifier(res)
	assert.NotEmpty(t, docID)

	// getting pending document should be successful
	aliceClient.GetDocumentWithStatus(docID, "pending", http.StatusOK)

	// Commits template
	res = aliceClient.CommitDocument("documents", http.StatusAccepted, docID)

	jobID, err := client.GetJobID(res)
	assert.NoError(t, err)

	err = aliceClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	aliceClient.GetDocumentAndVerify(docID, nil, createAttributes())

	// Bob should have the template
	bobClient.GetDocumentAndVerify(docID, nil, createAttributes())

	// Bob clones the document from a payload with a template ID
	valid := map[string]interface{}{
		"scheme":      "generic",
		"document_id": docID,
	}

	res1 := bobClient.CloneDocument("documents", http.StatusCreated, valid)

	docID1 := client.GetDocumentIdentifier(res1)
	assert.NotEmpty(t, docID1)

	res = bobClient.CommitDocument("documents", http.StatusAccepted, docID1)

	jobID, err = client.GetJobID(res)

	err = bobClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	bobClient.GetClonedDocumentAndCheck(docID, docID1, nil, createAttributes())
}
