// +build testworld

package testworld

import "testing"

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

	alice := doctorFord.getHost("Alice")
	bob := doctorFord.getHost("Bob")
	eAlice := alice.createHttpExpectation(t)
	eBob := bob.createHttpExpectation(t)


	a, err := alice.id()
	if err != nil {
		t.Error(err)
	}

	b, err := bob.id()
	if err != nil {
		t.Error(err)
	}

	// alice shares an invoice bob
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

	// Bob updates and sends to Alice
	res, err = bob.updateInvoice(eBob, docIdentifier, map[string]interface{}{
		"data": map[string]interface{}{
			"invoice_number": "12324",
			"due_date":       "2018-09-26T23:12:37.902198664Z",
			"gross_amount":   "42",
			"currency":       "USD",
			"net_amount":     "42",
		},
		"collaborators": []string{a.String()},
	})
	if err != nil {
		t.Error(err)
	}

	fmt.Println(res)

	docIdentifier = res.Value("header").Path("$.document_id").String().NotEmpty().Raw()
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

}
