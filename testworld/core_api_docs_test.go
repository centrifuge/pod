// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestCoreAPI_DocumentInvoiceCreateAndUpdate(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, invoiceCoreAPICreate([]string{bob.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"currency": "EUR",
	}

	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, params, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, params, createAttributes())
	nonExistingGenericDocumentCheck(charlie.httpExpect, charlie.id.String(), docIdentifier)

	// Bob updates invoice and shares with Charlie as well
	res = updateCoreAPIDocument(bob.httpExpect, bob.id.String(), "documents", docIdentifier, http.StatusAccepted, invoiceCoreAPIUpdate([]string{alice.id.String(), charlie.id.String()}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, bob.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	params["currency"] = "EUR"
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, params, allAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, params, allAttributes())
	getGenericDocumentAndCheck(t, charlie.httpExpect, charlie.id.String(), docIdentifier, params, allAttributes())
}

func TestCoreAPI_DocumentGenericCreateAndUpdate(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, genericCoreAPICreate([]string{bob.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{}
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, params, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, params, createAttributes())
	nonExistingGenericDocumentCheck(charlie.httpExpect, charlie.id.String(), docIdentifier)

	// Bob updates purchase order and shares with Charlie as well
	res = updateCoreAPIDocument(bob.httpExpect, bob.id.String(), "documents", docIdentifier, http.StatusAccepted, genericCoreAPIUpdate([]string{alice.id.String(), charlie.id.String()}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(bob.httpExpect, bob.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, params, allAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, params, allAttributes())
	getGenericDocumentAndCheck(t, charlie.httpExpect, charlie.id.String(), docIdentifier, params, allAttributes())
}

func TestCoreAPI_DocumentEntityCreateAndUpdate(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusAccepted, entityCoreAPICreate(alice.id.String(), []string{bob.id.String(), charlie.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)
	if docIdentifier == "" {
		t.Error("docIdentifier empty")
	}

	params := map[string]interface{}{
		"identity":   alice.id.String(),
		"legal_name": "test company",
	}

	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, params, createAttributes())
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, params, createAttributes())
	getGenericDocumentAndCheck(t, charlie.httpExpect, charlie.id.String(), docIdentifier, params, createAttributes())
}

func invoiceCoreAPICreate(collaborators []string) map[string]interface{} {
	return map[string]interface{}{
		"scheme":       "invoice",
		"write_access": collaborators,
		"data": map[string]interface{}{
			"number":       "12345",
			"status":       "unpaid",
			"gross_amount": "12.345",
			"recipient":    "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
			"date_due":     "2019-05-24T14:48:44.308854Z", // rfc3339nano
			"date_paid":    "2019-05-24T14:48:44Z",        // rfc3339
			"currency":     "EUR",
			"attachments": []map[string]interface{}{
				{
					"name":      "test",
					"file_type": "pdf",
					"size":      1000202,
					"data":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
					"checksum":  "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF3",
				},
			},
		},
		"attributes": createAttributes(),
	}
}

func invoiceCoreAPIUpdate(collaborators []string) map[string]interface{} {
	payload := invoiceCoreAPICreate(collaborators)
	payload["attributes"] = updateAttributes()
	return payload
}

func entityCoreAPICreate(identity string, collaborators []string) map[string]interface{} {
	p := map[string]interface{}{
		"scheme":       "entity",
		"write_access": collaborators,
		"attributes":   createAttributes(),
	}

	data := map[string]interface{}{
		"legal_name": "test company",
		"contacts": []map[string]interface{}{
			{
				"name": "test name",
			},
		},

		"payment_details": []map[string]interface{}{
			{
				"predefined": true,
				"bank_payment_method": map[string]interface{}{
					"identifier":  hexutil.Encode(utils.RandomSlice(32)),
					"holder_name": "John Doe",
				},
			},
		},
	}

	if identity != "" {
		data["identity"] = identity
	}

	p["data"] = data
	return p
}

func entityCoreAPIUpdate(collabs []string) map[string]interface{} {
	p := map[string]interface{}{
		"scheme":       "entity",
		"write_access": collabs,
		"data": map[string]interface{}{
			"legal_name": "updated company",
		},
		"attributes": updateAttributes(),
	}

	return p
}

func genericCoreAPICreate(collaborators []string) map[string]interface{} {
	return map[string]interface{}{
		"scheme":       "generic",
		"write_access": collaborators,
		"data":         map[string]interface{}{},
		"attributes":   createAttributes(),
	}
}

func genericCoreAPIUpdate(collaborators []string) map[string]interface{} {
	payload := genericCoreAPICreate(collaborators)
	payload["attributes"] = updateAttributes()
	return payload
}

func createAttributes() map[string]map[string]string {
	return map[string]map[string]string{
		"string_test": {
			"type":  "string",
			"value": "hello, world",
		},
	}
}

func updateAttributes() map[string]map[string]string {
	return map[string]map[string]string{
		"decimal_test": {
			"type":  "decimal",
			"value": "100.001",
		},
	}
}

func allAttributes() map[string]map[string]string {
	attrs := createAttributes()
	for k, v := range updateAttributes() {
		attrs[k] = v
	}

	return attrs
}
