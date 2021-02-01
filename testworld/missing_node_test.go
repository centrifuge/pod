// +build testworld

package testworld

import (
	"net/http"
	"testing"

	"github.com/centrifuge/go-centrifuge/identity/ideth"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
)

func TestMissingNode_InvalidIdentity(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")

	// RandomDID
	randomDID := testingidentity.GenerateRandomDID()

	// Alice shares a document with a randomly generated DID
	res := createDocument(alice.httpExpect, alice.id.String(), typeDocuments, http.StatusAccepted, genericCoreAPICreate([]string{randomDID.String()}))

	// Transaction should fail with invalid identity
	errorMessage := "failed to send document to the node: bytecode for deployed identity contract " + randomDID.String() + " not correct"
	assertTransactionError(t, res, alice.httpExpect, alice.id.String(), errorMessage)
}

func TestMissingNode_MissingRoute(t *testing.T) {
	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")

	// RandomDID without P2P Discovery Key
	randomDID := ideth.DeployIdentity(t, alice.host.bootstrappedCtx, alice.host.config)

	// Alice shares a document with a randomly generated DID with missing P2P Key
	res := createDocument(alice.httpExpect, alice.id.String(), typeDocuments, http.StatusAccepted, genericCoreAPICreate([]string{randomDID.String()}))

	// Transaction should fail with missing p2p key error
	errorMessage := "failed to send document to the node: routing: not found"
	assertTransactionError(t, res, alice.httpExpect, alice.id.String(), errorMessage)
}

// Assert error thrown in the transaction status
func assertTransactionError(t *testing.T, res *httpexpect.Object, httpExpect *httpexpect.Expect, identityID string, errorMessage string) {
	txID := getJobID(t, res)
	status, message := getTransactionStatusAndMessage(httpExpect, identityID, txID)
	if status != "failed" {
		t.Error(message)
	}

	assert.Contains(t, message, errorMessage)
}
