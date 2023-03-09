//go:build testworld

package testworld

import (
	"testing"

	"github.com/centrifuge/pod/testworld/park/host"
	"github.com/stretchr/testify/assert"
)

func TestDocumentsAPI_EntityCreate(t *testing.T) {
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

	// Alice shares a document with Bob and Charlie
	payload := defaultEntityPayload(
		alice.GetMainAccount().GetAccountID().ToHexString(),
		[]string{
			bob.GetMainAccount().GetAccountID().ToHexString(),
			charlie.GetMainAccount().GetAccountID().ToHexString(),
		},
	)

	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	params := map[string]interface{}{
		"legal_name": "test company",
	}

	aliceClient.GetDocumentAndVerify(docID, params, nil)
	bobClient.GetDocumentAndVerify(docID, params, nil)
	charlieClient.GetDocumentAndVerify(docID, params, nil)
}

func TestDocumentsAPI_EntityCreateAndUpdate(t *testing.T) {
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

	// Alice shares document with Bob
	payload := entityCoreAPICreate(
		alice.GetMainAccount().GetAccountID().ToHexString(),
		[]string{
			bob.GetMainAccount().GetAccountID().ToHexString(),
		},
	)

	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	params := map[string]interface{}{
		"identity":   alice.GetMainAccount().GetAccountID().ToHexString(),
		"legal_name": "test company",
	}

	aliceClient.GetDocumentAndVerify(docID, params, createAttributes())
	bobClient.GetDocumentAndVerify(docID, params, createAttributes())
	charlieClient.NonExistingDocumentCheck(docID)

	// Bob updates the doc and adds Charlie
	payload = entityCoreAPIUpdate(
		[]string{
			alice.GetMainAccount().GetAccountID().ToHexString(),
			charlie.GetMainAccount().GetAccountID().ToHexString(),
		},
	)
	payload["document_id"] = docID

	docID, err = bobClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	params["legal_name"] = "updated company"

	aliceClient.GetDocumentAndVerify(docID, params, allAttributes())
	bobClient.GetDocumentAndVerify(docID, params, allAttributes())
	charlieClient.GetDocumentAndVerify(docID, params, allAttributes())
}

func TestDocumentsAPI_EntityShareGet(t *testing.T) {
	alice, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)
	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)
	bobClient, err := controller.GetClientForHost(t, host.Bob)
	assert.NoError(t, err)

	// Alice anchors Entity
	payload := defaultEntityPayload(
		alice.GetMainAccount().GetAccountID().ToHexString(),
		[]string{},
	)

	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	// Alice creates an EntityRelationship with Bob
	payload = defaultRelationshipPayload(
		alice.GetMainAccount().GetAccountID().ToHexString(),
		docID,
		bob.GetMainAccount().GetAccountID().ToHexString(),
	)

	relID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	response := bobClient.GetEntityWithRelation(relID)
	response.Path("$.data.legal_name").String().Equal("test company")
}
