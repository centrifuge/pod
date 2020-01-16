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
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	mockdoc "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHost_GetSignatureFromCollaboratorBasedOnWrongSignature(t *testing.T) {
	alice := doctorFord.getHostTestSuite(t, "Alice")
	mallory := doctorFord.getHostTestSuite(t, "Mallory")

	mctxh := testingconfig.CreateAccountContext(t, mallory.host.config)

	publicKey, privateKey := GetSigningKeyPair(t, mallory.host.idService, mallory.id, mctxh)

	collaborators := [][]byte{alice.id[:]}
	dm := createCDWithEmbeddedDocumentWithWrongSignature(t, collaborators, alice.id, publicKey, privateKey, mallory.host.config.GetContractAddress(config.AnchorRepo))

	signatures, signatureErrors, err := mallory.host.p2pClient.GetSignaturesForDocument(mctxh, dm)
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Signature verification failed error")
	assert.Equal(t, 0, len(signatures))
}

func TestHost_ReturnSignatureComputedBaseOnAnotherSigningRoot(t *testing.T) {
	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	mallory := doctorFord.getHostTestSuite(t, "Mallory")

	actxh := testingconfig.CreateAccountContext(t, alice.host.config)
	mctxh := testingconfig.CreateAccountContext(t, mallory.host.config)

	publicKey, privateKey := GetSigningKeyPair(t, alice.host.idService, alice.id, actxh)

	collaborators := [][]byte{mallory.id[:]}
	dm := createCDWithEmbeddedDocument(t, collaborators, alice.id, publicKey, privateKey, alice.host.config.GetContractAddress(config.AnchorRepo))

	dm2 := createCDWithEmbeddedDocument(t, collaborators, alice.id, publicKey, privateKey, alice.host.config.GetContractAddress(config.AnchorRepo))

	sr, err := dm2.CalculateSigningRoot()
	assert.NoError(t, err)

	publicKeyValid, privateKeyValid := GetSigningKeyPair(t, mallory.host.idService, mallory.id, mctxh)
	s, err := crypto.SignMessage(privateKeyValid, sr, crypto.CurveSecp256K1)
	assert.NoError(t, err)

	sig := &coredocumentpb.Signature{
		SignatureId: append(mallory.id[:], publicKeyValid...),
		SignerId:    mallory.id[:],
		PublicKey:   publicKeyValid,
		Signature:   s,
	}

	malloryDocMockSrv := mallory.host.bootstrappedCtx[documents.BootstrappedDocumentService].(*mockdoc.MockService)

	malloryDocMockSrv.On("RequestDocumentSignature", mock.Anything, mock.Anything, mock.Anything).Return([]*coredocumentpb.Signature{sig}, nil).Once()

	malloryDocMockSrv.On("DeriveFromCoreDocument", mock.Anything).Return(dm, nil).Once()

	signatures, signatureErrors, err := alice.host.p2pClient.GetSignaturesForDocument(actxh, dm)
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Signature verification failed error")
	assert.Equal(t, 0, len(signatures))
}

func TestHost_SignKeyNotInCollaboration(t *testing.T) {
	// Hosts
	alice := doctorFord.getHostTestSuite(t, "Alice")
	mallory := doctorFord.getHostTestSuite(t, "Mallory")

	actxh := testingconfig.CreateAccountContext(t, alice.host.config)
	mctxh := testingconfig.CreateAccountContext(t, mallory.host.config)

	publicKey, privateKey := GetSigningKeyPair(t, alice.host.idService, alice.id, actxh)

	collaborators := [][]byte{mallory.id[:]}
	dm := createCDWithEmbeddedDocument(t, collaborators, alice.id, publicKey, privateKey, alice.host.config.GetContractAddress(config.AnchorRepo))

	sr, err := dm.CalculateSigningRoot()
	assert.NoError(t, err)
	payload := documents.ConsensusSignaturePayload(sr, false)
	publicKeyValid, privateKeyValid := GetSigningKeyPair(t, mallory.host.idService, mallory.id, mctxh)
	s, err := crypto.SignMessage(privateKeyValid, payload, crypto.CurveSecp256K1)
	assert.NoError(t, err)

	sig := &coredocumentpb.Signature{
		SignatureId: append(mallory.id[:], publicKeyValid...),
		SignerId:    mallory.id[:],
		PublicKey:   publicKeyValid,
		Signature:   s,
	}

	malloryDocMockSrv := mallory.host.bootstrappedCtx[documents.BootstrappedDocumentService].(*mockdoc.MockService)

	malloryDocMockSrv.On("RequestDocumentSignature", mock.Anything, mock.Anything, mock.Anything).Return([]*coredocumentpb.Signature{sig}, nil).Once()

	malloryDocMockSrv.On("DeriveFromCoreDocument", mock.Anything).Return(dm, nil).Once()

	//Signature verification should success
	signatures, signatureErrors, err := alice.host.p2pClient.GetSignaturesForDocument(actxh, dm)

	assert.NoError(t, err)
	assert.Nil(t, signatureErrors)
	assert.Equal(t, 1, len(signatures))

	//Following simulate attack by Mallory with random keys pair
	//Random keys pairs should cause signature verification failure
	publicKey2, privateKey2 := GetRandomSigningKeyPair(t)
	s, err = crypto.SignMessage(privateKey2, payload, crypto.CurveSecp256K1)
	assert.NoError(t, err)

	sig = &coredocumentpb.Signature{
		SignatureId: append(mallory.id[:], publicKey2...),
		SignerId:    mallory.id[:],
		PublicKey:   publicKey2,
		Signature:   s,
	}

	// when got request on signature of document, mocking documents.Service of Mallory return a random signature
	malloryDocMockSrv.On("RequestDocumentSignature", mock.Anything, mock.Anything, mock.Anything).Return([]*coredocumentpb.Signature{sig}, nil).Once()

	malloryDocMockSrv.On("DeriveFromCoreDocument", mock.Anything).Return(dm, nil).Once()

	signatures, signatureErrors, err = alice.host.p2pClient.GetSignaturesForDocument(actxh, dm)
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Signature verification failed error")
	assert.Equal(t, 0, len(signatures))
}

func TestHost_ValidSignature(t *testing.T) {
	// Hosts
	bob := doctorFord.getHostTestSuite(t, "Bob")
	eve := doctorFord.getHostTestSuite(t, "Eve")

	ctxh := testingconfig.CreateAccountContext(t, eve.host.config)
	// Get PublicKey and PrivateKey
	publicKey, privateKey := GetSigningKeyPair(t, eve.host.idService, eve.id, ctxh)

	collaborators := [][]byte{bob.id[:]}
	dm := createCDWithEmbeddedDocument(t, collaborators, eve.id, publicKey, privateKey, eve.host.config.GetContractAddress(config.AnchorRepo))

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
	dm := createCDWithEmbeddedDocument(t, collaborators, eve.id, publicKey, privateKey, eve.host.config.GetContractAddress(config.AnchorRepo))

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
	dm := createCDWithEmbeddedDocument(t, collaborators, eve.id, publicKey, privateKey, eve.host.config.GetContractAddress(config.AnchorRepo))

	signatures, signatureErrors, err := eve.host.p2pClient.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "Signature verification failed error")
	assert.Equal(t, 0, len(signatures))

	res := createDocument(bob.httpExpect, bob.id.String(), typeDocuments, http.StatusAccepted, genericCoreAPICreate([]string{eve.id.String()}))
	txID := getTransactionID(t, res)
	status, _ := getTransactionStatusAndMessage(bob.httpExpect, bob.id.String(), txID)
	// Even though there was a signature validation error, as of now, we keep anchoring document
	assert.Equal(t, status, "success")
}

// Helper Methods
func createCDWithEmbeddedDocument(t *testing.T, collaborators [][]byte, identityDID identity.DID, publicKey []byte, privateKey []byte, anchorRepo common.Address) documents.Model {
	payload := generic.CreateGenericPayload(t, nil)
	var cs []identity.DID
	collabs, err := identity.BytesToDIDs(collaborators...)
	assert.NoError(t, err)
	for _, c := range collabs {
		cs = append(cs, *c)
	}
	payload.Collaborators.ReadWriteCollaborators = cs
	g := generic.InitGeneric(t, identityDID, payload)
	g.SetUsedAnchorRepoAddress(anchorRepo)
	err = g.AddUpdateLog(identityDID)
	assert.NoError(t, err)

	sr, err := g.CalculateSigningRoot()
	assert.NoError(t, err)
	signPayload := documents.ConsensusSignaturePayload(sr, false)
	s, err := crypto.SignMessage(privateKey, signPayload, crypto.CurveSecp256K1)
	assert.NoError(t, err)

	sig := &coredocumentpb.Signature{
		SignatureId:         append(identityDID[:], publicKey...),
		SignerId:            identityDID[:],
		PublicKey:           publicKey,
		Signature:           s,
		TransitionValidated: false,
	}
	g.AppendSignatures(sig)

	_, err = g.CalculateDocumentRoot()
	assert.NoError(t, err)

	return g
}

func createCDWithEmbeddedDocumentWithWrongSignature(t *testing.T, collaborators [][]byte, identityDID identity.DID, publicKey []byte, privateKey []byte, anchorRepo common.Address) documents.Model {
	payload := generic.CreateGenericPayload(t, nil)
	var cs []identity.DID
	collabs, err := identity.BytesToDIDs(collaborators...)
	assert.NoError(t, err)
	for _, c := range collabs {
		cs = append(cs, *c)
	}
	payload.Collaborators.ReadWriteCollaborators = cs

	g := generic.InitGeneric(t, identityDID, payload)
	g.SetUsedAnchorRepoAddress(anchorRepo)
	err = g.AddUpdateLog(identityDID)
	assert.NoError(t, err)

	//Wrong Signing Root will cause wrong signature
	sr, err := g.CalculateSignaturesRoot()
	assert.NoError(t, err)

	s, err := crypto.SignMessage(privateKey, sr, crypto.CurveSecp256K1)
	assert.NoError(t, err)

	sig := &coredocumentpb.Signature{
		SignatureId: append(identityDID[:], publicKey...),
		SignerId:    identityDID[:],
		PublicKey:   publicKey,
		Signature:   s,
	}
	g.AppendSignatures(sig)

	_, err = g.CalculateDocumentRoot()
	assert.NoError(t, err)

	return g
}

func RevokeKey(t *testing.T, idService identity.Service, key [32]byte, identityDID identity.DID, ctx context.Context) {
	err := idService.RevokeKey(ctx, key)
	assert.NoError(t, err)
	response, err := idService.GetKey(identityDID, key)
	assert.NoError(t, err)
	assert.NotEqual(t, utils.ByteSliceToBigInt([]byte{0}), response.RevokedAt, "Revoked key successfully")
}

func AddKey(t *testing.T, idService identity.Service, testKey identity.Key, identityDID identity.DID, ctx context.Context) {
	err := idService.AddKey(ctx, testKey)
	assert.Nil(t, err, "Add Key should be successful")

	_, err = idService.GetKey(identityDID, testKey.GetKey())
	assert.Nil(t, err, "Get Key should be successful")

	err = idService.ValidateKey(ctx, identityDID, utils.Byte32ToSlice(testKey.GetKey()), testKey.GetPurpose(), nil)
	assert.Nil(t, err, "Key with purpose should exist")
}

func GetSigningKeyPair(t *testing.T, idService identity.Service, identityDID identity.DID, ctx context.Context) ([]byte, []byte) {
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

func GetRandomSigningKeyPair(t *testing.T) ([]byte, []byte) {
	// Generate PublicKey and PrivateKey
	publicKey, privateKey, err := secp256k1.GenerateSigningKeyPair()
	assert.NoError(t, err)

	address32Bytes := convertKeyTo32Bytes(publicKey)

	return utils.Byte32ToSlice(address32Bytes), privateKey
}

func convertKeyTo32Bytes(key []byte) [32]byte {
	address := common.HexToAddress(secp256k1.GetAddress(key))
	return utils.AddressTo32Bytes(address)
}
