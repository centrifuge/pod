// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHost_FundingBasic(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	charlie := doctorFord.getHostTestSuite(t, "Charlie")

	// alice creates invoice
	res := createDocument(alice.httpExpect, alice.id.String(), typeInvoice, http.StatusOK, defaultInvoicePayload([]string{bob.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	docIdentifier := getDocumentIdentifier(t, res)

	params := map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
	}
	getDocumentAndCheck(t, alice.httpExpect, alice.id.String(), typeInvoice, params, true)
	getDocumentAndCheck(t, bob.httpExpect, bob.id.String(), typeInvoice, params, true)
	fmt.Println("Host test success")

	// alice adds a funding and shares with charlie
	res = createFunding(alice.httpExpect, alice.id.String(), docIdentifier, http.StatusOK, defaultFundingPayload([]string{charlie.id.String()}))
	txID = getTransactionID(t, res)
	status, message = getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "success" {
		t.Error(message)
	}

	fundingId := getFundingId(t,res)
	params = map[string]interface{}{
		"document_id": docIdentifier,
		"currency":    "USD",
		"amount": "20000",
		"apr":     "0.33",
	}

	// check if everybody received to funding
	getFundingAndCheck(alice.httpExpect,alice.id.String(),docIdentifier,fundingId, params)
	getFundingAndCheck(bob.httpExpect,bob.id.String(),docIdentifier,fundingId, params)
	getFundingAndCheck(charlie.httpExpect,charlie.id.String(),docIdentifier,fundingId, params)

}
