//go:build testworld
// +build testworld

package testworld

//func setupTransitionRuleForCharlie(t *testing.T) (string, string) {
//	alice := doctorFord.getHostTestSuite(t, "Alice")
//	bob := doctorFord.getHostTestSuite(t, "Bob")
//	charlie := doctorFord.getHostTestSuite(t, "Charlie")
//
//	// Alice prepares document to share with Bob
//	docPayload := genericCoreAPICreate([]string{bob.id.String(), alice.id.String()})
//	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, docPayload)
//	status := getDocumentStatus(t, res)
//	assert.Equal(t, status, "pending")
//	docID := getDocumentIdentifier(t, res)
//	roleID := utils.RandomSlice(32)
//	payload := map[string][]map[string]interface{}{
//		"attribute_rules": {
//			{
//				"key_label": "oracle1",
//				"role_id":   hexutil.Encode(roleID),
//			},
//		},
//	}
//
//	// no role
//	obj := addTransitionRules(alice.httpExpect, alice.id.String(), docID, payload, http.StatusBadRequest)
//	assert.Contains(t, obj.Path("$.message").String().Raw(), "role doesn't exist")
//
//	ruleID := utils.RandomSlice(32)
//	obj = getTransitionRule(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(ruleID), http.StatusNotFound)
//	assert.Contains(t, obj.Path("$.message").String().Raw(), "transition rule missing")
//
//	// delete an non existing rule
//	delRes := deleteTransitionRule(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(ruleID), http.StatusNotFound)
//	assert.Contains(t, delRes.JSON().Object().Path("$.message").String().Raw(), "transition rule missing")
//
//	// create role
//	obj = addRole(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(roleID), []string{charlie.id.String()}, http.StatusOK)
//	r, cs := parseRole(obj)
//	assert.Equal(t, r, hexutil.Encode(roleID))
//	assert.Contains(t, cs, strings.ToLower(charlie.id.String()))
//
//	// add transition rules
//	obj = addTransitionRules(alice.httpExpect, alice.id.String(), docID, payload, http.StatusOK)
//	tr := parseRules(t, obj)
//	assert.Len(t, tr.Rules, 1)
//	ruleID = tr.Rules[0].RuleID[:]
//	obj = getTransitionRule(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(ruleID), http.StatusOK)
//	rule := parseRule(t, obj)
//	assert.Equal(t, tr.Rules[0], rule)
//
//	// commit document
//	res = commitDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, docID)
//	jobID := getJobID(t, res)
//	err := waitForJobComplete(doctorFord.maeve, alice.httpExpect, alice.id.String(), jobID)
//	assert.NoError(t, err)
//	getDocumentAndVerify(t, alice.httpExpect, alice.id.String(), docID, nil, createAttributes())
//	// pending document should fail
//	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "pending", http.StatusNotFound)
//	// committed should be successful
//	getV2DocumentWithStatus(alice.httpExpect, alice.id.String(), docID, "committed", http.StatusOK)
//	// Bob should have the document
//	getDocumentAndVerify(t, bob.httpExpect, bob.id.String(), docID, nil, createAttributes())
//	obj = getTransitionRule(alice.httpExpect, alice.id.String(), docID, hexutil.Encode(ruleID), http.StatusOK)
//	rule = parseRule(t, obj)
//	assert.Equal(t, tr.Rules[0], rule)
//	return docID, rule.RuleID.String()
//}
//
//func TestTransitionRules(t *testing.T) {
//	alice := doctorFord.getHostTestSuite(t, "Alice")
//	bob := doctorFord.getHostTestSuite(t, "Bob")
//	charlie := doctorFord.getHostTestSuite(t, "Charlie")
//	docID, _ := setupTransitionRuleForCharlie(t)
//
//	// charlie updates the document with wrong attr key and tries to get full access
//	p := genericCoreAPIUpdate([]string{charlie.id.String()})
//	p["document_id"] = docID
//	createDocument(charlie.httpExpect, charlie.id.String(), "documents", http.StatusCreated, p)
//	res := commitDocument(charlie.httpExpect, charlie.id.String(), "documents", http.StatusAccepted, docID)
//	versionID := getDocumentCurrentVersion(t, res)
//	jobID := getJobID(t, res)
//	err := waitForJobComplete(doctorFord.maeve, charlie.httpExpect, charlie.id.String(), jobID)
//	assert.NoError(t, err)
//	// alice and bob would have not accepted the document update.
//	nonExistingGenericDocumentVersionCheck(alice.httpExpect, alice.id.String(), docID, versionID)
//	nonExistingGenericDocumentVersionCheck(bob.httpExpect, bob.id.String(), docID, versionID)
//
//	// charlie updates the document with right attribute
//	docID, ruleID := setupTransitionRuleForCharlie(t)
//	p = genericCoreAPICreate(nil)
//	p["attributes"] = coreapi.AttributeMapRequest{
//		"oracle1": coreapi.AttributeRequest{
//			Type:  "decimal",
//			Value: "100.001",
//		},
//	}
//	p["document_id"] = docID
//	docID = createAndCommitDocument(t, doctorFord.maeve, charlie.httpExpect, charlie.id.String(), p)
//
//	// alice deletes the rule
//	p = genericCoreAPICreate(nil)
//	p["document_id"] = docID
//	// create a new draft of the existing document
//	res = createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, p)
//	status := getDocumentStatus(t, res)
//	assert.Equal(t, status, "pending")
//	ndocID := getDocumentIdentifier(t, res)
//	versionID = getDocumentCurrentVersion(t, res)
//	assert.Equal(t, docID, ndocID, "Document ID should match")
//	obj := getTransitionRule(alice.httpExpect, alice.id.String(), docID, ruleID, http.StatusOK)
//	rule := parseRule(t, obj)
//	assert.Equal(t, ruleID, rule.RuleID.String())
//	deleteTransitionRule(alice.httpExpect, alice.id.String(), docID, ruleID, http.StatusNoContent).NoContent()
//
//	// commit the document
//	res = commitDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, docID)
//	jobID = getJobID(t, res)
//	err = waitForJobComplete(doctorFord.maeve, alice.httpExpect, alice.id.String(), jobID)
//	assert.NoError(t, err)
//
//	// charlie should not have latest document
//	nonExistingGenericDocumentVersionCheck(charlie.httpExpect, charlie.id.String(), docID, versionID)
//}
