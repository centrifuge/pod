// +build testworld

package testworld

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/stretchr/testify/assert"
)

func TestMissingNode_InvalidIdentity(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")

	// RandomDID
	randomDID := testingidentity.GenerateRandomDID()

	// alice shares a document with bob and charlie
	res := createDocument(alice.httpExpect, alice.id.String(), typeInvoice, http.StatusOK, defaultInvoicePayload([]string{randomDID.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(alice.httpExpect, alice.id.String(), txID)
	if status != "failed" {
		t.Error(message)
	}

	errorMessage := "failed to send document to the node: bytecode for deployed identity contract " + randomDID.String() + " not correct"
	assert.Contains(t, message, errorMessage)
}
