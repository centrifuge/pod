//go:build testworld

package testworld

//func TestHost_AddExternalCollaborator(t *testing.T) {
//	tests := []struct {
//		name     string
//		testType testType
//	}{
//		{
//			"Document_multiHost_AddExternalCollaborator",
//			multiHost,
//		},
//		{
//			"Document_withinhost_AddExternalCollaborator",
//			withinHost,
//		},
//		{
//			"Document_multiHostMultiAccount_AddExternalCollaborator",
//			multiHostMultiAccount,
//		},
//	}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			switch test.testType {
//			case multiHost:
//				addExternalCollaborator(t)
//			case multiHostMultiAccount:
//				addExternalCollaboratorMultiHostMultiAccount(t)
//			case withinHost:
//				addExternalCollaboratorWithinHost(t)
//			}
//		})
//	}
//}
//
//func addExternalCollaboratorWithinHost(t *testing.T) {
//	bob := doctorFord.getHostTestSuite(t, "Bob")
//	accounts := doctorFord.getTestAccount("Bob").accounts
//	a := accounts[0]
//	b := accounts[1]
//	c := accounts[2]
//
//	// a shares document with b first
//	docID := createAndCommitDocument(t, doctorFord.maeve, bob.httpExpect, a, genericCoreAPICreate([]string{b}))
//	getDocumentAndVerify(t, bob.httpExpect, a, docID, nil, createAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, b, docID, nil, createAttributes())
//
//	// account b sends a webhook for received anchored doc
//	msg, err := doctorFord.maeve.getReceivedDocumentMsg(b, docID)
//	assert.NoError(t, err)
//	assert.Equal(t, strings.ToLower(a), strings.ToLower(msg.Document.From.String()))
//	log.Debug("Host test success")
//	nonExistingDocumentCheck(bob.httpExpect, c, docID)
//
//	// b updates invoice and shares with c as well
//	payload := genericCoreAPIUpdate([]string{a, c})
//	payload["document_id"] = docID
//	docID = createAndCommitDocument(t, doctorFord.maeve, bob.httpExpect, b, payload)
//	getDocumentAndVerify(t, bob.httpExpect, a, docID, nil, allAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, b, docID, nil, allAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, c, docID, nil, allAttributes())
//	// account c sends a webhook for received anchored doc
//	msg, err = doctorFord.maeve.getReceivedDocumentMsg(c, docID)
//	assert.NoError(t, err)
//	assert.Equal(t, strings.ToLower(b), strings.ToLower(msg.Document.From.String()))
//}
//
//func addExternalCollaboratorMultiHostMultiAccount(t *testing.T) {
//	alice := doctorFord.getHostTestSuite(t, "Alice")
//	bob := doctorFord.getHostTestSuite(t, "Bob")
//	accounts := doctorFord.getTestAccount("Bob").accounts
//	a := accounts[0]
//	b := accounts[1]
//	c := accounts[2]
//	charlie := doctorFord.getHostTestSuite(t, "Charlie")
//	accounts2 := doctorFord.getTestAccount("Charlie").accounts
//	d := accounts2[0]
//	e := accounts2[1]
//	f := accounts2[2]
//
//	// Alice shares document with Bobs accounts a and b
//	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(), genericCoreAPICreate([]string{a, b}))
//	getDocumentAndVerify(t, alice.httpExpect, alice.id.ToHexString(), docID, nil, createAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, a, docID, nil, createAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, b, docID, nil, createAttributes())
//
//	// bobs account b sends a webhook for received anchored doc
//	msg, err := doctorFord.maeve.getReceivedDocumentMsg(b, docID)
//	assert.NoError(t, err)
//	assert.Equal(t, strings.ToLower(alice.id.ToHexString()), strings.ToLower(msg.Document.From.String()))
//	nonExistingDocumentCheck(bob.httpExpect, c, docID)
//
//	// Bob updates invoice and shares with bobs account c as well using account a and to accounts d and e of Charlie
//	payload := genericCoreAPIUpdate([]string{alice.id.ToHexString(), b, c, d, e})
//	payload["document_id"] = docID
//	docID = createAndCommitDocument(t, doctorFord.maeve, bob.httpExpect, a, payload)
//	getDocumentAndVerify(t, alice.httpExpect, alice.id.ToHexString(), docID, nil, allAttributes())
//	// bobs accounts all have the document now
//	getDocumentAndVerify(t, bob.httpExpect, a, docID, nil, allAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, b, docID, nil, allAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, c, docID, nil, allAttributes())
//	getDocumentAndVerify(t, charlie.httpExpect, d, docID, nil, allAttributes())
//	getDocumentAndVerify(t, charlie.httpExpect, e, docID, nil, allAttributes())
//	nonExistingDocumentCheck(charlie.httpExpect, f, docID)
//}
//
//func addExternalCollaborator(t *testing.T) {
//	alice := doctorFord.getHostTestSuite(t, "Alice")
//	bob := doctorFord.getHostTestSuite(t, "Bob")
//	charlie := doctorFord.getHostTestSuite(t, "Charlie")
//
//	// Alice shares document with Bob first
//	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(), genericCoreAPICreate([]string{bob.id.ToHexString()}))
//	getDocumentAndVerify(t, alice.httpExpect, alice.id.ToHexString(), docID, nil, createAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, bob.id.ToHexString(), docID, nil, createAttributes())
//	nonExistingDocumentCheck(charlie.httpExpect, charlie.id.ToHexString(), docID)
//
//	// Bob updates invoice and shares with Charlie as well
//	payload := genericCoreAPIUpdate([]string{alice.id.ToHexString(), charlie.id.ToHexString()})
//	payload["document_id"] = docID
//	docID = createAndCommitDocument(t, doctorFord.maeve, bob.httpExpect, bob.id.ToHexString(), payload)
//	getDocumentAndVerify(t, alice.httpExpect, alice.id.ToHexString(), docID, nil, allAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, bob.id.ToHexString(), docID, nil, allAttributes())
//	getDocumentAndVerify(t, charlie.httpExpect, charlie.id.ToHexString(), docID, nil, allAttributes())
//}
//
//func TestHost_CollaboratorTimeOut(t *testing.T) {
//	collaboratorTimeOut(t)
//}
//
//func collaboratorTimeOut(t *testing.T) {
//	kenny := doctorFord.getHostTestSuite(t, "Kenny")
//	bob := doctorFord.getHostTestSuite(t, "Bob")
//
//	// Kenny shares a document with Bob
//	docID := createAndCommitDocument(t, doctorFord.maeve, kenny.httpExpect, kenny.id.ToHexString(), genericCoreAPICreate([]string{bob.id.ToHexString()}))
//	getDocumentAndVerify(t, kenny.httpExpect, kenny.id.ToHexString(), docID, nil, createAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, bob.id.ToHexString(), docID, nil, createAttributes())
//
//	// Kenny gets killed
//	kenny.host.kill()
//
//	// Bob updates and sends to Kenny
//	// Bob will anchor the document without Kennys signature
//	payload := genericCoreAPIUpdate([]string{kenny.id.ToHexString()})
//	payload["document_id"] = docID
//	docID = createAndCommitDocument(t, doctorFord.maeve, bob.httpExpect, bob.id.ToHexString(), payload)
//	getDocumentAndVerify(t, bob.httpExpect, bob.id.ToHexString(), docID, nil, allAttributes())
//
//	// bring Kenny back to life
//	doctorFord.reLive(t, kenny.name)
//
//	// Kenny should NOT have latest version
//	getDocumentAndVerify(t, kenny.httpExpect, kenny.id.ToHexString(), docID, nil, createAttributes())
//}
//
//func TestDocument_invalidAttributes(t *testing.T) {
//	t.Parallel()
//	kenny := doctorFord.getHostTestSuite(t, "Kenny")
//	bob := doctorFord.getHostTestSuite(t, "Bob")
//
//	live, err := kenny.host.isLive(time.Second * 30)
//	assert.NoError(t, err)
//	assert.True(t, live)
//
//	// Kenny shares a document with Bob
//	response := createDocument(kenny.httpExpect, kenny.id.ToHexString(), typeDocuments, http.StatusBadRequest,
//		wrongGenericDocumentPayload([]string{bob.id.ToHexString()}))
//	errMsg := response.Raw()["message"].(string)
//	assert.Contains(t, errMsg, "some invalid time stamp\" as \"2006-01-02T15:04:05.999999999Z07:00\": cannot parse \"some invalid ti")
//}
//
//func TestDocument_latestDocumentVersion(t *testing.T) {
//	alice := doctorFord.getHostTestSuite(t, "Alice")
//	bob := doctorFord.getHostTestSuite(t, "Bob")
//	charlie := doctorFord.getHostTestSuite(t, "Charlie")
//	kenny := doctorFord.getHostTestSuite(t, "Kenny")
//
//	// alice creates a document with bob and kenny
//	docID := createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(),
//		genericCoreAPICreate([]string{alice.id.ToHexString(), bob.id.ToHexString(), kenny.id.ToHexString()}))
//	getDocumentAndVerify(t, alice.httpExpect, alice.id.ToHexString(), docID, nil, createAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, bob.id.ToHexString(), docID, nil, createAttributes())
//	getDocumentAndVerify(t, kenny.httpExpect, kenny.id.ToHexString(), docID, nil, createAttributes())
//	nonExistingDocumentCheck(charlie.httpExpect, charlie.id.ToHexString(), docID)
//
//	// Bob updates invoice and shares with Charlie as well but kenny is offline and miss the update
//	kenny.host.kill()
//	payload := genericCoreAPIUpdate([]string{charlie.id.ToHexString()})
//	payload["document_id"] = docID
//	docID = createAndCommitDocument(t, doctorFord.maeve, bob.httpExpect, bob.id.ToHexString(), payload)
//	getDocumentAndVerify(t, bob.httpExpect, bob.id.ToHexString(), docID, nil, allAttributes())
//	getDocumentAndVerify(t, alice.httpExpect, alice.id.ToHexString(), docID, nil, allAttributes())
//	getDocumentAndVerify(t, charlie.httpExpect, charlie.id.ToHexString(), docID, nil, allAttributes())
//	// bring kenny back and should not have the latest version
//	doctorFord.reLive(t, kenny.name)
//	getDocumentAndVerify(t, kenny.httpExpect, kenny.id.ToHexString(), docID, nil, createAttributes())
//
//	// alice updates document
//	payload = genericCoreAPIUpdate(nil)
//	payload["document_id"] = docID
//	docID = createAndCommitDocument(t, doctorFord.maeve, alice.httpExpect, alice.id.ToHexString(), payload)
//
//	// everyone should have the latest version
//	getDocumentAndVerify(t, alice.httpExpect, alice.id.ToHexString(), docID, nil, allAttributes())
//	getDocumentAndVerify(t, bob.httpExpect, bob.id.ToHexString(), docID, nil, allAttributes())
//	getDocumentAndVerify(t, charlie.httpExpect, charlie.id.ToHexString(), docID, nil, allAttributes())
//	getDocumentAndVerify(t, kenny.httpExpect, kenny.id.ToHexString(), docID, nil, allAttributes())
//}
