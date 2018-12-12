// +build testworld

package testworld
/*
func TestHost_AddExternalCollaborator_invoice(t *testing.T) {
	t.Parallel()
	AddExternalCollaborator(t, TypeInvoice)
}


func addExternalCollaborator(t *testing.T, documentType string) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares invoice document with Bob first
	res := createDocument(alice.httpExpect,documentType, http.StatusOK, defaultDocumentPayload(documentType, []string{bob.id.String()}))


	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getInvoiceAndCheck(alice.httpExpect, params)
	getInvoiceAndCheck(bob.httpExpect, params)

	// Bob updates invoice and shares with Charlie as well
	res, err = bob.host.updateInvoice(bob.httpExpect, http.StatusOK, docIdentifier, updatedInvoicePayload([]string{alice.id.String(), charlie.id.String()}))
	if err != nil {
		t.Error(err)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "EUR"
	getInvoiceAndCheck(alice.httpExpect, params)
	getInvoiceAndCheck(bob.httpExpect, params)
	getInvoiceAndCheck(charlie.httpExpect, params)
}

/*
func TestHost_CollaboratorTimeOut(t *testing.T) {
	t.Parallel()
	kenny := doctorFord.getHostTestSuite(t, "Kenny")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	// Kenny shares an invoice with Bob
	response, err := kenny.host.createInvoice(kenny.httpExpect, http.StatusOK, defaultInvoicePayload([]string{bob.id.String()}))

	if err != nil {
		t.Error(err)
	}

	// check if Bob and Kenny received the document
	docIdentifier := getDocumentIdentifier(t, response)
	paramsV1 := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getInvoiceAndCheck(kenny.httpExpect, paramsV1)
	getInvoiceAndCheck(bob.httpExpect, paramsV1)

	// Kenny gets killed
	kenny.host.kill()

	// Bob updates and sends to Alice
	updatedPayload := updatedInvoicePayload([]string{kenny.id.String()})

	// Bob will anchor the document without Alice signature but will receive an error because kenny is dead
	response, err = bob.host.updateInvoice(bob.httpExpect, http.StatusInternalServerError, docIdentifier, updatedPayload)
	if err != nil {
		t.Error(err)
	}

	// check if bob saved the updated document
	paramsV2 := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "EUR",
	}
	getInvoiceAndCheck(bob.httpExpect, paramsV2)

	// bring Kenny back to life
	doctorFord.reLive(t, kenny.name)

	// Kenny should NOT have latest version
	getInvoiceAndCheck(kenny.httpExpect, paramsV1)

} */
