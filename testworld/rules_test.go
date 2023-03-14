//go:build testworld

package testworld

import (
	"net/http"
	"strings"
	"testing"

	"github.com/centrifuge/pod/http/coreapi"
	"github.com/centrifuge/pod/testworld/park/behavior/client"
	"github.com/centrifuge/pod/testworld/park/host"
	"github.com/centrifuge/pod/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestDocumentsAPI_TransitionRules(t *testing.T) {
	t.Parallel()

	charlie, err := controller.GetHost(host.Charlie)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)
	bobClient, err := controller.GetClientForHost(t, host.Bob)
	assert.NoError(t, err)
	charlieClient, err := controller.GetClientForHost(t, host.Charlie)
	assert.NoError(t, err)

	docID, _ := setupTransitionRuleForCharlie(t)

	// charlie updates the document with wrong attr key and tries to get full access
	payload := genericCoreAPIUpdate([]string{charlie.GetMainAccount().GetAccountID().ToHexString()})
	payload["document_id"] = docID

	charlieClient.CreateDocument("documents", http.StatusCreated, payload)

	res := charlieClient.CommitDocument("documents", http.StatusAccepted, docID)
	versionID := client.GetDocumentCurrentVersion(res)

	jobID, err := client.GetJobID(res)
	assert.NoError(t, err)

	err = charlieClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	// alice and bob would have not accepted the document update.
	aliceClient.NonExistingGenericDocumentVersionCheck(docID, versionID)
	bobClient.NonExistingGenericDocumentVersionCheck(docID, versionID)

	// charlie updates the document with right attribute
	docID, ruleID := setupTransitionRuleForCharlie(t)

	payload = genericCoreAPICreate(nil)
	payload["attributes"] = coreapi.AttributeMapRequest{
		"oracle1": coreapi.AttributeRequest{
			Type:  "decimal",
			Value: "100.001",
		},
	}
	payload["document_id"] = docID

	docID, err = charlieClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	// alice deletes the rule
	payload = genericCoreAPICreate(nil)
	payload["document_id"] = docID

	// create a new draft of the existing document
	res = aliceClient.CreateDocument("documents", http.StatusCreated, payload)

	status := client.GetDocumentStatus(res)
	assert.Equal(t, status, "pending")

	ndocID := client.GetDocumentIdentifier(res)
	versionID = client.GetDocumentCurrentVersion(res)

	assert.Equal(t, docID, ndocID, "Document ID should match")

	obj := aliceClient.GetTransitionRule(docID, ruleID, http.StatusOK)

	rule := client.ParseRule(t, obj)
	assert.Equal(t, ruleID, rule.RuleID.String())

	aliceClient.DeleteTransitionRule(docID, ruleID, http.StatusNoContent).NoContent()

	// commit the document
	res = aliceClient.CommitDocument("documents", http.StatusAccepted, docID)

	jobID, err = client.GetJobID(res)

	err = aliceClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	// charlie should not have the latest document
	charlieClient.NonExistingGenericDocumentVersionCheck(docID, versionID)
}

func setupTransitionRuleForCharlie(t *testing.T) (string, string) {
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

	// Alice prepares document to share with Bob
	docPayload := genericCoreAPICreate(
		[]string{
			bob.GetMainAccount().GetAccountID().ToHexString(),
			alice.GetMainAccount().GetAccountID().ToHexString(),
		},
	)
	res := aliceClient.CreateDocument("documents", http.StatusCreated, docPayload)

	status := client.GetDocumentStatus(res)
	assert.Equal(t, status, "pending")

	docID := client.GetDocumentIdentifier(res)
	roleID := utils.RandomSlice(32)
	payload := map[string][]map[string]interface{}{
		"attribute_rules": {
			{
				"key_label": "oracle1",
				"role_id":   hexutil.Encode(roleID),
			},
		},
	}

	// no role
	obj := aliceClient.AddTransitionRules(docID, payload, http.StatusBadRequest)
	assert.Contains(t, obj.Path("$.message").String().Raw(), "role doesn't exist")

	ruleID := utils.RandomSlice(32)
	obj = aliceClient.GetTransitionRule(docID, hexutil.Encode(ruleID), http.StatusNotFound)
	assert.Contains(t, obj.Path("$.message").String().Raw(), "transition rule missing")

	// delete an non existing rule
	delRes := aliceClient.DeleteTransitionRule(docID, hexutil.Encode(ruleID), http.StatusNotFound)
	assert.Contains(t, delRes.JSON().Object().Path("$.message").String().Raw(), "transition rule missing")

	// create role
	obj = aliceClient.AddRole(
		docID,
		hexutil.Encode(roleID),
		[]string{charlie.GetMainAccount().GetAccountID().ToHexString()},
		http.StatusOK,
	)

	r, cs := client.ParseRole(obj)
	assert.Equal(t, r, hexutil.Encode(roleID))
	assert.Contains(t, cs, strings.ToLower(charlie.GetMainAccount().GetAccountID().ToHexString()))

	// add transition rules
	obj = aliceClient.AddTransitionRules(docID, payload, http.StatusOK)

	tr := client.ParseRules(t, obj)
	assert.Len(t, tr.Rules, 1)

	ruleID = tr.Rules[0].RuleID[:]
	obj = aliceClient.GetTransitionRule(docID, hexutil.Encode(ruleID), http.StatusOK)

	rule := client.ParseRule(t, obj)
	assert.Equal(t, tr.Rules[0], rule)

	// commit document
	res = aliceClient.CommitDocument("documents", http.StatusAccepted, docID)

	jobID, err := client.GetJobID(res)
	assert.NoError(t, err)

	err = aliceClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	aliceClient.GetDocumentAndVerify(docID, nil, createAttributes())

	// pending document should fail
	aliceClient.GetDocumentWithStatus(docID, "pending", http.StatusNotFound)

	// committed should be successful
	aliceClient.GetDocumentWithStatus(docID, "committed", http.StatusOK)

	// Bob should have the document
	bobClient.GetDocumentAndVerify(docID, nil, createAttributes())

	obj = aliceClient.GetTransitionRule(docID, hexutil.Encode(ruleID), http.StatusOK)

	rule = client.ParseRule(t, obj)
	assert.Equal(t, tr.Rules[0], rule)

	return docID, rule.RuleID.String()
}
