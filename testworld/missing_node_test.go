// +build testworld

package testworld

import (
	"context"
	"math/big"
	"net/http"
	"strings"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	res := createDocument(alice.httpExpect, alice.id.String(), typeInvoice, http.StatusOK, defaultInvoicePayload([]string{randomDID.String()}))

	// Validate document is anchored correctly
	validateAnchorDocument(t, res, alice.host.anchorRepo)

	// Transaction should fail with invalid identity
	errorMessage := "failed to send document to the node: bytecode for deployed identity contract " + randomDID.String() + " not correct"
	assertTransactionError(t, res, alice.httpExpect, alice.id.String(), errorMessage)
}

func TestMissingNode_MissingP2PKey(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")

	// RandomDID without P2P Discovery Key
	randomDID := createIdentity(t, alice.host.idFactory, alice.host.idService, alice.host.config)

	// Alice shares a document with a randomly generated DID with missing P2P Key
	res := createDocument(alice.httpExpect, alice.id.String(), typeInvoice, http.StatusOK, defaultInvoicePayload([]string{randomDID.String()}))

	// Validate document is anchored correctly
	validateAnchorDocument(t, res, alice.host.anchorRepo)

	// Transaction should fail with missing p2p key error
	errorMessage := "failed to send document to the node: error fetching p2p key: missing p2p key"
	assertTransactionError(t, res, alice.httpExpect, alice.id.String(), errorMessage)
}

// Helper Methods
func createIdentity(t *testing.T, idFactory identity.Factory, idService identity.ServiceDID, cfg config.Configuration) identity.DID {
	// Create Identity
	didAddr, err := idFactory.CalculateIdentityAddress(context.Background())
	assert.NoError(t, err)
	tc, err := configstore.NewAccount("main", cfg)
	assert.Nil(t, err)
	acc := tc.(*configstore.Account)
	acc.IdentityID = didAddr.Bytes()

	ctx, err := contextutil.New(context.Background(), tc)
	assert.Nil(t, err)
	did, err := idFactory.CreateIdentity(ctx)
	assert.Nil(t, err, "should not error out when creating identity")
	assert.Equal(t, did.String(), strings.ToLower(didAddr.String()), "Resulting Identity should have the same ID as the input")

	// Add Keys
	accKeys, err := tc.GetKeys()
	assert.NoError(t, err)
	sPk, err := utils.SliceToByte32(accKeys[identity.KeyPurposeSigning.Name].PublicKey)
	assert.NoError(t, err)
	keyDID := identity.NewKey(sPk, &(identity.KeyPurposeSigning.Value), big.NewInt(identity.KeyTypeECDSA), 0)
	err = idService.AddKey(ctx, keyDID)
	assert.Nil(t, err, "should not error out when adding key to identity")

	return *did
}

// Validate document is anchored correctly
func validateAnchorDocument(t *testing.T, res *httpexpect.Object, anchorRepo anchors.AnchorRepository) {
	versionID := getDocumentCurrentVersion(t, res)
	versionIDBytesArray, err := hexutil.Decode(versionID)
	assert.NoError(t, err)

	anchorID, err := anchors.ToAnchorID(versionIDBytesArray)
	assert.NoError(t, err)

	docRoot, anchoredAt, err := anchorRepo.GetAnchorData(anchorID)
	assert.NoError(t, err)
	assert.NotNil(t, docRoot)
	assert.NotNil(t, anchoredAt)
}

// Assert error thrown in the transaction status
func assertTransactionError(t *testing.T, res *httpexpect.Object, httpExpect *httpexpect.Expect, identityID string, errorMessage string) {
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(httpExpect, identityID, txID)
	if status != "failed" {
		t.Error(message)
	}

	assert.Contains(t, message, errorMessage)
}
