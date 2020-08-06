// +build testworld

package testworld

import (
	"net/http"
	"strings"
	"testing"

	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func setupTransitionRuleForCharlie(t *testing.T) (string, string) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice prepares document to share with Bob
	docPayload := genericCoreAPICreate([]string{bob.id.String()})
	res := createDocumentV2(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, docPayload)
	status := getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")
	docID := getDocumentIdentifier(t, res)
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
	obj := addTransitionRules(alice.httpExpect, alice.id.String(), docID, payload, http.StatusBadRequest)
	assert.Contains(t, obj.Path("$.message").String().Raw(), "role doesn't exist")

	ruleID := utils.RandomSlice(32)
	obj = getTransitionRule(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(ruleID), http.StatusNotFound)
	assert.Contains(t, obj.Path("$.message").String().Raw(), "transition rule missing")

	// delete an non existing rule
	delRes := deleteTransitionRule(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(ruleID), http.StatusNotFound)
	assert.Contains(t, delRes.JSON().Object().Path("$.message").String().Raw(), "transition rule missing")

	// create role
	obj = addRole(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(roleID), []string{charlie.id.String()}, http.StatusOK)
	r, cs := parseRole(obj)
	assert.Equal(t, r, hexutil.Encode(roleID))
	assert.Contains(t, cs, strings.ToLower(charlie.id.String()))

	// add transition rules
	obj = addTransitionRules(alice.httpExpect, alice.id.String(), docID, payload, http.StatusOK)
	tr := parseRules(t, obj)
	assert.Len(t, tr.Rules, 1)
	ruleID = tr.Rules[0].RuleID[:]
	obj = getTransitionRule(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(ruleID), http.StatusOK)
	rule := parseRule(t, obj)
	assert.Equal(t, tr.Rules[0], rule)

	// commit document
	res = commitDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, docID)
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	assert.Equal(t, status, "success", message)
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docID, nil, createAttributes())
	// pending document should fail
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusNotFound)
	// committed should be successful
	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "committed", http.StatusOK)
	// Bob should have the document
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docID, nil, createAttributes())
	obj = getTransitionRule(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(ruleID), http.StatusOK)
	rule = parseRule(t, obj)
	assert.Equal(t, tr.Rules[0], rule)
	return docID, rule.RuleID.String()
}

func TestTransitionRules(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")
	docID, ruleID := setupTransitionRuleForCharlie(t)

	// charlie updates the document with wrong attr key and tries to get full access
	p := genericCoreAPIUpdate([]string{charlie.id.String()})
	res := updateCoreAPIDocument(charlie.httpExpect, charlie.id.String(), "documents", docID, http.StatusAccepted, p)
	txID := getTransactionID(t, res)
	status, _ := getTransactionStatusAndMessage(charlie.httpExpect, charlie.id.String(), txID)
	if status == "success" {
		t.Error("document should not be updated")
	}

	// charlie updates the document with right attribute
	docID, ruleID = setupTransitionRuleForCharlie(t)
	p = genericCoreAPICreate(nil)
	p["attributes"] = coreapi.AttributeMapRequest{
		"oracle1": coreapi.AttributeRequest{
			Type:  "decimal",
			Value: "100.001",
		},
	}
	res = updateCoreAPIDocument(charlie.httpExpect, charlie.id.String(), "documents", docID, http.StatusAccepted, p)
	txID = getTransactionID(t, res)
	status, _ = getTransactionStatusAndMessage(charlie.httpExpect, charlie.id.String(), txID)
	if status != "success" {
		t.Error("document should be updated")
	}

	// alice deletes the rule
	p = genericCoreAPICreate(nil)
	p["document_id"] = docID
	// create a new draft of the existing document
	res = createDocumentV2(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, p)
	status = getDocumentStatus(t, res)
	assert.Equal(t, status, "pending")
	ndocID := getDocumentIdentifier(t, res)
	versionID := getDocumentCurrentVersion(t, res)
	assert.Equal(t, docID, ndocID, "Document ID should match")
	obj := getTransitionRule(alice.httpExpect, alice.id.String(), docID, ruleID, http.StatusOK)
	rule := parseRule(t, obj)
	assert.Equal(t, ruleID, rule.RuleID.String())
	deleteTransitionRule(alice.httpExpect, alice.id.String(), docID, ruleID, http.StatusNoContent).NoContent()

	// commit the document
	res = commitDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, docID)
	txID = getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	assert.Equal(t, status, "success", message)

	// charlie should not have latest document
	nonExistingGenericDocumentVersionCheck(charlie.httpExpect, charlie.id.String(), docID, versionID)
}
