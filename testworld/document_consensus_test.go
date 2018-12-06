// +build testworld

package testworld

import (
	"testing"
)

func TestHost_AddExternalCollaborator(t *testing.T) {
	alice := doctorFord.getHost("Alice")
	bob := doctorFord.getHost("Bob")
	charlie := doctorFord.getHost("Charlie")
	eAlice := alice.createHttpExpectation(t)
	eBob := bob.createHttpExpectation(t)
	eCharlie := charlie.createHttpExpectation(t)

	a, err := alice.id()
	if err != nil {
		t.Error(err)
	}

	b, err := bob.id()
	if err != nil {
		t.Error(err)
	}

	c, err := charlie.id()
	if err != nil {
		t.Error(err)
	}

	// Alice shares invoice document with Bob first
	res, err := alice.createInvoice(eAlice, map[string]interface{}{
		"data": map[string]interface{}{
			"invoice_number": "12324",
			"due_date":       "2018-09-26T23:12:37.902198664Z",
			"gross_amount":   "40",
			"currency":       "GBP",
			"net_amount":     "40",
		},
		"collaborators": []string{b.String()},
	})
	if err != nil {
		t.Error(err)
	}
	docIdentifier := res.Value("header").Path("$.document_id").String().NotEmpty().Raw()
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "GBP",
	}
	getInvoiceAndCheck(eAlice, params)
	getInvoiceAndCheck(eBob, params)

	// Bob updates invoice and shares with Charlie as well
	res, err = bob.updateInvoice(eBob, docIdentifier, map[string]interface{}{
		"data": map[string]interface{}{
			"invoice_number": "12324",
			"due_date":       "2018-09-26T23:12:37.902198664Z",
			"gross_amount":   "40",
			"currency":       "USD",
			"net_amount":     "40",
		},
		"collaborators": []string{a.String(), c.String()},
	})
	if err != nil {
		t.Error(err)
	}
	docIdentifier = res.Value("header").Path("$.document_id").String().NotEmpty().Raw()
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "USD"
	getInvoiceAndCheck(eAlice, params)
	getInvoiceAndCheck(eBob, params)
	getInvoiceAndCheck(eCharlie, params)
}

func TestHost_CollaboratorTimeOut(t *testing.T) {

	alice := getHostTestSuite(t, "Alice")
	bob := getHostTestSuite(t, "Bob")

	// alice shares an invoice bob
	response, err := alice.host.createInvoice(alice.expect, defaultInvoicePayload([]string{bob.id.String()}))

	if err != nil {
		t.Error(err)
	}

	// check if bob and alice received the document
	docIdentifier := getDocumentIdentifier(t, response)
	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getInvoiceAndCheck(alice.expect, params)
	getInvoiceAndCheck(bob.expect, params)

	// Bob updates and sends to Alice
	updatedPayload := updatedInvoicePayload([]string{alice.id.String()})
	response, err = bob.host.updateInvoice(bob.expect, docIdentifier, updatedPayload)
	if err != nil {
		t.Error(err)
	}

	docIdentifier = getDocumentIdentifier(t, response)
}
