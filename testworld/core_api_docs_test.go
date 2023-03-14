//go:build testworld

package testworld

import (
	"io/ioutil"
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

func TestDocumentsAPI_GenericCreateAndUpdate(t *testing.T) {
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

	// Alice shares document with Bob first
	payload := genericCoreAPICreate([]string{bob.GetMainAccount().GetAccountID().ToHexString()})

	docID, err := aliceClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	params := map[string]interface{}{}
	aliceClient.GetDocumentAndVerify(docID, params, createAttributes())
	bobClient.GetDocumentAndVerify(docID, params, createAttributes())
	charlieClient.NonExistingDocumentCheck(docID)

	// Bob updates purchase order and shares with Charlie as well
	payload = genericCoreAPIUpdate(
		[]string{
			alice.GetMainAccount().GetAccountID().ToHexString(),
			charlie.GetMainAccount().GetAccountID().ToHexString(),
		},
	)
	payload["document_id"] = docID

	docID, err = bobClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	aliceClient.GetDocumentAndVerify(docID, params, allAttributes())
	bobClient.GetDocumentAndVerify(docID, params, allAttributes())
	charlieClient.GetDocumentAndVerify(docID, params, allAttributes())
}

func TestDocumentsAPI_ComputeFields(t *testing.T) {
	alice, err := controller.GetHost(host.Alice)
	assert.NoError(t, err)
	bob, err := controller.GetHost(host.Bob)
	assert.NoError(t, err)

	aliceClient, err := controller.GetClientForHost(t, host.Alice)
	assert.NoError(t, err)
	bobClient, err := controller.GetClientForHost(t, host.Bob)
	assert.NoError(t, err)

	payload := genericCoreAPICreate([]string{alice.GetMainAccount().GetAccountID().ToHexString()})
	res := aliceClient.CreateDocument("documents", http.StatusCreated, payload)

	status := client.GetDocumentStatus(res)
	assert.Equal(t, status, "pending")

	docID := client.GetDocumentIdentifier(res)
	assert.NotEmpty(t, docID)

	wasm, err := ioutil.ReadFile("testingutils/compute_fields/simple_average.wasm")
	assert.NoError(t, err)

	// create role
	roleID := utils.RandomSlice(32)
	obj := aliceClient.AddRole(
		docID,
		hexutil.Encode(roleID),
		[]string{bob.GetMainAccount().GetAccountID().ToHexString()},
		http.StatusOK,
	)

	r, cs := client.ParseRole(obj)
	assert.Equal(t, r, hexutil.Encode(roleID))
	assert.Contains(t, cs, strings.ToLower(bob.GetMainAccount().GetAccountID().ToHexString()))

	// set compute fields
	rules := map[string][]map[string]interface{}{
		"compute_fields_rules": {
			{
				"wasm":                   hexutil.Encode(wasm),
				"attribute_labels":       []string{"test", "test1", "test2"},
				"target_attribute_label": "result",
			},
		},
		"attribute_rules": {
			{
				"key_label": "test",
				"role_id":   hexutil.Encode(roleID),
			},
			{
				"key_label": "test1",
				"role_id":   hexutil.Encode(roleID),
			},
			{
				"key_label": "test2",
				"role_id":   hexutil.Encode(roleID),
			},
		},
	}

	obj = aliceClient.AddTransitionRules(docID, rules, http.StatusOK)

	tr := client.ParseRules(t, obj)
	assert.Len(t, tr.Rules, 4)

	ruleID := tr.Rules[3].RuleID[:]
	obj = aliceClient.GetTransitionRule(docID, hexutil.Encode(ruleID), http.StatusOK)
	rule := client.ParseRule(t, obj)
	assert.Equal(t, tr.Rules[3], rule)

	// commits the document
	res = aliceClient.CommitDocument("documents", http.StatusAccepted, docID)
	jobID, err := client.GetJobID(res)
	assert.NoError(t, err)

	err = aliceClient.WaitForJobCompletion(jobID)
	assert.NoError(t, err)

	var result [32]byte

	aliceClient.GetDocumentAndVerify(docID, nil, withComputeFieldResultAttribute(result[:]))
	bobClient.GetDocumentAndVerify(docID, nil, withComputeFieldResultAttribute(result[:]))

	// bob adds the attributes
	payload = genericCoreAPICreate(nil)

	attrs := coreapi.AttributeMapRequest{
		"test": coreapi.AttributeRequest{
			Type:  "integer",
			Value: "1000",
		},
		"test1": coreapi.AttributeRequest{
			Type:  "integer",
			Value: "2000",
		},
		"test2": coreapi.AttributeRequest{
			Type:  "integer",
			Value: "3000",
		},
	}
	payload["attributes"] = attrs
	payload["document_id"] = docID

	_, err = bobClient.CreateAndCommitDocument(payload)
	assert.NoError(t, err)

	// result = encoded(risk(1)) + encoded((1000+2000+3000)/3 = 1000)
	result = [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x7, 0xd0}
	reqAttrs := withComputeFieldResultAttribute(result[:])

	for k, v := range attrs {
		reqAttrs[k] = v
	}

	aliceClient.GetDocumentAndVerify(docID, nil, reqAttrs)
	bobClient.GetDocumentAndVerify(docID, nil, reqAttrs)
}
