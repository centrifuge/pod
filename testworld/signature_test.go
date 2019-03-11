// +build testworld

package testworld

import (
	"context"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHost_ValidSignature(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")

	ctxh := testingconfig.CreateAccountContext(t, alice.host.config)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedPO(t, collaborators, alice.id, ctxh)
	assert.Equal(t, 1, len(dm.Signatures()))

	signatures, _, _ := alice.host.p2pClient.GetSignaturesForDocument(ctxh, dm)
	assert.Equal(t, 1, len(signatures))
}

func TestHost_FakedSignature(t *testing.T) {
	t.Parallel()

	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	actxh := testingconfig.CreateAccountContext(t, alice.host.config)
	ectxh := testingconfig.CreateAccountContext(t, eve.host.config)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedPO(t, collaborators, eve.id, actxh)

	signatures, signatureErrors, _ := eve.host.p2pClient.GetSignaturesForDocument(ectxh, dm)
	assert.Error(t, signatureErrors[0], "Signature verification failed error")
	assert.Equal(t, 0, len(signatures))
}

func TestHost_RevokedSigningKey(t *testing.T) {
	t.Parallel()
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ctxh := testingconfig.CreateAccountContext(t, eve.host.config)

	keys, err := eve.host.idService.GetKeysByPurpose(eve.id, &(identity.KeyPurposeSigning.Value))
	assert.NoError(t, err)

	// Revoke Key
	eve.host.idService.RevokeKey(ctxh, keys[0])
	response, err := eve.host.idService.GetKey(eve.id, keys[0])
	assert.NotEqual(t, utils.ByteSliceToBigInt([]byte{0}), response.RevokedAt, "Revoked key successfully")

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedPO(t, collaborators, eve.id, ctxh)

	signatures, signatureErrors, _ := eve.host.p2pClient.GetSignaturesForDocument(ctxh, dm)
	assert.Error(t, signatureErrors[0], "Signature verification failed error")
	assert.Equal(t, 0, len(signatures))
}

// Helper Methods
func createCDWithEmbeddedPO(t *testing.T, collaborators [][]byte, identityDID identity.DID, ctx context.Context) documents.Model {
	payload := testingdocuments.CreatePOPayload()
	var cs []string
	for _, c := range collaborators {
		cs = append(cs, hexutil.Encode(c))
	}
	payload.Collaborators = cs

	po := new(purchaseorder.PurchaseOrder)
	err := po.InitPurchaseOrderInput(payload, identityDID.String())
	assert.NoError(t, err)

	_, err = po.CalculateDataRoot()
	assert.NoError(t, err)

	sr, err := po.CalculateSigningRoot()
	assert.NoError(t, err)

	accCfg, err := contextutil.Account(ctx)
	assert.NoError(t, err)
	acc := accCfg.(*configstore.Account)
	accKeys, err := acc.GetKeys()
	assert.NoError(t, err)

	s, err := crypto.SignMessage(accKeys[identity.KeyPurposeSigning.Name].PrivateKey, sr, crypto.CurveSecp256K1)
	assert.NoError(t, err)

	sig := &coredocumentpb.Signature{
		EntityId:  identityDID[:],
		PublicKey: accKeys[identity.KeyPurposeSigning.Name].PublicKey,
		Signature: s,
		Timestamp: utils.ToTimestamp(time.Now().UTC()),
	}
	po.AppendSignatures(sig)

	_, err = po.CalculateDocumentRoot()
	assert.NoError(t, err)

	return po
}
