// +build testworld

package testworld

import (
	"net/http"
	"testing"
)

func Test_CreateList(t *testing.T) {
	t.Parallel()
	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	_, identifier := createInvoiceWithFunding(t, alice, bob, charlie)
	listTest(t, alice, bob, charlie, identifier)
}
func Test_SignUpdate(t *testing.T) {
	t.Parallel()
	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	fundingId, identifier := createInvoiceWithFunding(t, alice, bob, charlie)
	signTest(t, alice, bob, charlie, fundingId, identifier)
	updateTest(t, alice, bob, charlie, fundingId, identifier)
}

func updateTest(t *testing.T, alice, bob, charlie hostTestSuite, fundingId, docIdentifier string) {
	// alice adds a funding and shares with charlie
	res := updateFunding(alice.httpExpect, alice.id.String(), fundingId, http.StatusAccepted, docIdentifier, updateFundingPayload(fundingId, alice.id.String(), charlie.id.String()))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	fundingId = getAgreementId(t, res)
	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
		"amount":      "10000",
		"apr":         "0.55",
	}

	// check if everybody received the update with an outdated signature
	getFundingWithSignatureAndCheck(alice.httpExpect, alice.id.String(), docIdentifier, fundingId, "true", "true", params)
	getFundingWithSignatureAndCheck(bob.httpExpect, bob.id.String(), docIdentifier, fundingId, "true", "true", params)
	getFundingWithSignatureAndCheck(charlie.httpExpect, charlie.id.String(), docIdentifier, fundingId, "true", "true", params)

}

func listTest(t *testing.T, alice, bob, charlie hostTestSuite, docIdentifier string) {
	var fundings []string
	for i := 0; i < 5; i++ {
		res := createFunding(alice.httpExpect, alice.id.String(), docIdentifier, http.StatusAccepted, defaultFundingPayload(alice.id.String(), bob.id.String()))
		txID := getTransactionID(t, res)
		status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
		if status != "success" {
			t.Error(message)
		}

		fundingId := getAgreementId(t, res)
		fundings = append(fundings, fundingId)

	}
	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
		"amount":      "20000",
		"apr":         "0.33",
	}

	getListFundingCheck(alice.httpExpect, alice.id.String(), docIdentifier, 6, params)
	getListFundingCheck(bob.httpExpect, bob.id.String(), docIdentifier, 6, params)
	getListFundingCheck(charlie.httpExpect, charlie.id.String(), docIdentifier, 6, params)
}

func signTest(t *testing.T, alice, bob, charlie hostTestSuite, fundingId, docIdentifier string) {
	// alice adds a funding and shares with charlie
	res := signFunding(alice.httpExpect, alice.id.String(), docIdentifier, fundingId, http.StatusAccepted)
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	fundingId = getAgreementId(t, res)
	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
		"amount":      "20000",
		"apr":         "0.33",
	}

	// check if everybody received to funding with a signature
	getFundingWithSignatureAndCheck(alice.httpExpect, alice.id.String(), docIdentifier, fundingId, "true", "false", params)
	getFundingWithSignatureAndCheck(bob.httpExpect, bob.id.String(), docIdentifier, fundingId, "true", "false", params)
	getFundingWithSignatureAndCheck(charlie.httpExpect, charlie.id.String(), docIdentifier, fundingId, "true", "false", params)

}

func createInvoiceWithFunding(t *testing.T, alice, bob, charlie hostTestSuite) (agreementId, docIdentifier string) {
	res := createDocument(alice.httpExpect, alice.id.String(), typeDocuments, http.StatusAccepted, defaultInvoicePayload([]string{bob.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier = getDocumentIdentifier(t, res)

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(t, alice.httpExpect, alice.id.String(), typeDocuments, params, true)
	getDocumentAndCheck(t, bob.httpExpect, bob.id.String(), typeDocuments, params, true)

	// alice adds a funding and shares with charlie
	res = createFunding(alice.httpExpect, alice.id.String(), docIdentifier, http.StatusAccepted, defaultFundingPayload(alice.id.String(), charlie.id.String()))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	agreementId = getAgreementId(t, res)
	params = map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
		"amount":      "20000",
		"apr":         "0.33",
	}

	// check if everybody received to funding
	getFundingAndCheck(alice.httpExpect, alice.id.String(), docIdentifier, agreementId, params)
	getFundingAndCheck(bob.httpExpect, bob.id.String(), docIdentifier, agreementId, params)
	getFundingAndCheck(charlie.httpExpect, charlie.id.String(), docIdentifier, agreementId, params)
	return agreementId, docIdentifier
}
