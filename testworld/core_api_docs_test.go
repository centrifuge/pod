// +build testworld

package testworld

import (
	"net/http"
	"testing"
)

func TestCoreAPI_DocumentInvoiceCreateAndUpdate(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, invoiceCoreAPICreate([]string{bob.id.String()}))
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
		"document_id": docIdentifier,
		"currency":    "EUR",
	}

	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, params)
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, params)
	nonExistingGenericDocumentCheck(charlie.httpExpect, charlie.id.String(), docIdentifier)

	// Bob updates invoice and shares with Charlie as well
	res = updateCoreAPIDocument(bob.httpExpect, bob.id.String(), "documents", http.StatusCreated, invoiceCoreAPIUpdate(docIdentifier, []string{alice.id.String(), charlie.id.String()}))
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
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, params)
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, params)
	getGenericDocumentAndCheck(t, charlie.httpExpect, charlie.id.String(), docIdentifier, params)
}

func TestCoreAPI_DocumentPOCreateAndUpdate(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// Alice shares document with Bob first
	res := createDocument(alice.httpExpect, alice.id.String(), "documents", http.StatusCreated, poCoreAPICreate([]string{bob.id.String()}))
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
		"document_id": docIdentifier,
		"currency":    "EUR",
	}

	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, params)
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, params)
	nonExistingGenericDocumentCheck(charlie.httpExpect, charlie.id.String(), docIdentifier)

	// Bob updates purchase order and shares with Charlie as well
	res = updateCoreAPIDocument(bob.httpExpect, bob.id.String(), "documents", http.StatusCreated, poCoreAPIUpdate(docIdentifier, []string{alice.id.String(), charlie.id.String()}))
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
	getGenericDocumentAndCheck(t, alice.httpExpect, alice.id.String(), docIdentifier, params)
	getGenericDocumentAndCheck(t, bob.httpExpect, bob.id.String(), docIdentifier, params)
	getGenericDocumentAndCheck(t, charlie.httpExpect, charlie.id.String(), docIdentifier, params)
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
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "string",
				"value": "hello, world",
			},
		},
	}
}

func invoiceCoreAPIUpdate(docID string, collaborators []string) map[string]interface{} {
	payload := invoiceCoreAPICreate(collaborators)
	payload["document_id"] = docID
	payload["attributes"] = map[string]map[string]string{
		"decimal_test": {
			"type":  "decimal",
			"value": "100.001",
		},
	}
	return payload
}

func poCoreAPICreate(collaborators []string) map[string]interface{} {
	return map[string]interface{}{
		"scheme":       "purchaseorder",
		"write_access": collaborators,
		"data": map[string]interface{}{
			"number":         "12345",
			"status":         "unpaid",
			"total_amount":   "12.345",
			"recipient":      "0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7",
			"date_sent":      "2019-05-24T14:48:44.308854Z", // rfc3339nano
			"date_confirmed": "2019-05-24T14:48:44Z",        // rfc3339
			"currency":       "EUR",
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
		"attributes": map[string]map[string]string{
			"string_test": {
				"type":  "string",
				"value": "hello, world",
			},
		},
	}
}

func poCoreAPIUpdate(docID string, collaborators []string) map[string]interface{} {
	payload := poCoreAPICreate(collaborators)
	payload["document_id"] = docID
	payload["attributes"] = map[string]map[string]string{
		"decimal_test": {
			"type":  "decimal",
			"value": "100.001",
		},
	}
	return payload
}
