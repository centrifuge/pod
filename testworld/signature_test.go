// +build testworld

package testworld

import (
	"context"
	"net/http"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestHost_ValidSignature(t *testing.T) {
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ctxh := testingconfig.CreateAccountContext(t, eve.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, eve.host.idService, eve.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedPO(t, collaborators, eve.id, publicKey, privateKey, eve.host.config.GetContractAddress(config.AnchorRepo))

	signatures, signatureErrors, err := eve.host.p2pClient.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	assert.Nil(t, signatureErrors)
	assert.Equal(t, 1, len(signatures))
}

func TestHost_FakedSignature(t *testing.T) {
	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	actxh := testingconfig.CreateAccountContext(t, alice.host.config)
	ectxh := testingconfig.CreateAccountContext(t, eve.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, alice.host.idService, alice.id, actxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedPO(t, collaborators, eve.id, publicKey, privateKey, eve.host.config.GetContractAddress(config.AnchorRepo))

	signatures, signatureErrors, err := eve.host.p2pClient.GetSignaturesForDocument(ectxh, dm)
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Signature verification failed error")
	assert.Equal(t, 0, len(signatures))
}

func TestHost_RevokedSigningKey(t *testing.T) {
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ctxh := testingconfig.CreateAccountContext(t, eve.host.config)

	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, eve.host.idService, eve.id, ctxh)

	// Revoke Key
	key, err := utils.SliceToByte32(publicKey)
	assert.NoError(t, err)
	RevokeKey(t, eve.host.idService, key, eve.id, ctxh)

	// Eve creates document with Bob and signs with Revoked key
	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedPO(t, collaborators, eve.id, publicKey, privateKey, eve.host.config.GetContractAddress(config.AnchorRepo))

	signatures, signatureErrors, err := eve.host.p2pClient.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Signature verification failed error")
	assert.Equal(t, 0, len(signatures))

	// Bob creates document with Eve whose key is revoked
	keys, err := eve.host.idService.GetKeysByPurpose(eve.id, &(identity.KeyPurposeSigning.Value))
	assert.NoError(t, err)

	// Revoke Key
	RevokeKey(t, eve.host.idService, keys[0].GetKey(), eve.id, ctxh)

	res := createDocument(bob.httpExpect, bob.id.String(), typeInvoice, http.StatusOK, defaultInvoicePayload([]string{eve.id.String()}))
	txID := getTransactionID(t, res)
	status, message := getTransactionStatusAndMessage(bob.httpExpect, bob.id.String(), txID)
	if status != "failed" {
		t.Error(message)
	}
	assert.Contains(t, message, "failed to validate signatures")
}

// Helper Methods
func createCDWithEmbeddedPO(t *testing.T, collaborators [][]byte, identityDID identity.DID, publicKey []byte, privateKey []byte, anchorRepo common.Address) documents.Model {
	payload := testingdocuments.CreatePOPayload()
	var cs []string
	for _, c := range collaborators {
		cs = append(cs, hexutil.Encode(c))
	}
	payload.WriteAccess = &documentpb.WriteAccess{
		Collaborators: cs,
	}

	po := new(purchaseorder.PurchaseOrder)
	err := po.InitPurchaseOrderInput(payload, identityDID)
	assert.NoError(t, err)

	po.SetUsedAnchorRepoAddress(anchorRepo)
	err = po.AddUpdateLog(identityDID)
	assert.NoError(t, err)

	_, err = po.CalculateDataRoot()
	assert.NoError(t, err)

	sr, err := po.CalculateSigningRoot()
	assert.NoError(t, err)

	s, err := crypto.SignMessage(privateKey, sr, crypto.CurveSecp256K1)
	assert.NoError(t, err)

	sig := &coredocumentpb.Signature{
		SignatureId: append(identityDID[:], publicKey...),
		SignerId:    identityDID[:],
		PublicKey:   publicKey,
		Signature:   s,
	}
	po.AppendSignatures(sig)

	_, err = po.CalculateDocumentRoot()
	assert.NoError(t, err)

	return po
}

func RevokeKey(t *testing.T, idService identity.ServiceDID, key [32]byte, identityDID identity.DID, ctx context.Context) {
	err := idService.RevokeKey(ctx, key)
	assert.NoError(t, err)
	response, err := idService.GetKey(identityDID, key)
	assert.NoError(t, err)
	assert.NotEqual(t, utils.ByteSliceToBigInt([]byte{0}), response.RevokedAt, "Revoked key successfully")
}

func AddKey(t *testing.T, idService identity.ServiceDID, testKey identity.KeyDID, identityDID identity.DID, ctx context.Context) {
	err := idService.AddKey(ctx, testKey)
	assert.Nil(t, err, "Add Key should be successful")

	_, err = idService.GetKey(identityDID, testKey.GetKey())
	assert.Nil(t, err, "Get Key should be successful")

	err = idService.ValidateKey(ctx, identityDID, utils.Byte32ToSlice(testKey.GetKey()), testKey.GetPurpose(), nil)
	assert.Nil(t, err, "Key with purpose should exist")
}

func GetSigningKeyPair(t *testing.T, idService identity.ServiceDID, identityDID identity.DID, ctx context.Context) ([]byte, []byte) {
	// Generate PublicKey and PrivateKey
	publicKey, privateKey, err := secp256k1.GenerateSigningKeyPair()
	assert.NoError(t, err)

	address32Bytes := convertKeyTo32Bytes(publicKey)

	// Test Key
	testKey := identity.NewKey(address32Bytes, &(identity.KeyPurposeSigning.Value), utils.ByteSliceToBigInt([]byte{123}), 0)

	// Add Key
	AddKey(t, idService, testKey, identityDID, ctx)

	return utils.Byte32ToSlice(address32Bytes), privateKey
}

func convertKeyTo32Bytes(key []byte) [32]byte {
	address := common.HexToAddress(secp256k1.GetAddress(key))
	return utils.AddressTo32Bytes(address)
}
