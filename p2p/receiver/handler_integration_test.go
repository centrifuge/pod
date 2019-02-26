// +build integration

package receiver_test

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	cented25519 "github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

var (
	handler    *receiver.Handler
	anchorRepo anchors.AnchorRepository
	cfg        config.Configuration
	idService  identity.Service
	cfgService config.Service
	docSrv     documents.Service
)

func TestMain(m *testing.M) {
	flag.Parse()
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgService = ctx[config.BootstrappedConfigStorage].(config.Service)
	docSrv = ctx[documents.BootstrappedDocumentService].(documents.Service)
	anchorRepo = ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	idService = ctx[identity.BootstrappedIDService].(identity.Service)
	handler = receiver.New(cfgService, receiver.HandshakeValidator(cfg.GetNetworkID(), idService), docSrv)
	testingidentity.CreateIdentityWithKeys(cfg, idService)
	result := m.Run()
	testingbootstrap.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestHandler_GetDocument_nonexistentIdentifier(t *testing.T) {
	b := utils.RandomSlice(32)
	centrifugeId := createIdentity(t)
	req := &p2ppb.GetDocumentRequest{DocumentIdentifier: b}
	resp, err := handler.GetDocument(context.Background(), req, centrifugeId)
	assert.Error(t, err, "must return error")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_GetDocumentSucceeds(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	centrifugeId := createIdentity(t)

	po, cd := prepareDocumentForP2PHandler(t, nil)
	resp, err := handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd})
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	// Add signature received
	po.AppendSignatures(resp.Signature)
	tree, err := po.CoreDocument.DocumentRootTree()
	assert.NoError(t, err)
	po.CoreDocument.Document.DocumentRoot = tree.RootHash()

	// Anchor document
	idConfig, err := identity.GetIdentityConfig(cfg)
	anchorIDTyped, _ := anchors.ToAnchorID(po.CurrentVersion())
	docRootTyped, _ := anchors.ToDocumentRoot(po.Document.DocumentRoot)
	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centrifugeId, docRootTyped)
	signature, _ := secp256k1.SignEthereum(messageToSign, idConfig.Keys[identity.KeyPurposeEthMsgAuth].PrivateKey)
	anchorConfirmations, err := anchorRepo.CommitAnchor(ctxh, anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")

	cd, err = po.PackCoreDocument()
	assert.NoError(t, err)
	anchorResp, err := handler.SendAnchoredDocument(ctxh, &p2ppb.AnchorDocumentRequest{Document: &cd}, idConfig.ID[:])
	assert.Nil(t, err)
	assert.NotNil(t, anchorResp, "must be non nil")

	// Retrieve document from anchor repository with document_identifier
	getDocResp, err := handler.GetDocument(ctxh, &p2ppb.GetDocumentRequest{DocumentIdentifier: po.ID()}, centrifugeId)
	assert.Nil(t, err)
	assert.ObjectsAreEqual(getDocResp.Document, cd)
}

func TestHandler_HandleInterceptorReqSignature(t *testing.T) {
	centID := createIdentity(t)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	tc, err := contextutil.Account(ctxh)
	_, err = cfgService.CreateAccount(tc)
	assert.NoError(t, err)
	_, cd := prepareDocumentForP2PHandler(t, nil)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctxh, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})

	pub, _ := cfg.GetP2PKeyPair()
	publicKey, err := cented25519.GetPublicSigningKey(pub)
	assert.NoError(t, err)
	var bPk [32]byte
	copy(bPk[:], publicKey)
	peerID, err := cented25519.PublicKeyToP2PKey(bPk)
	assert.NoError(t, err)

	p2pResp, err := handler.HandleInterceptor(ctxh, peerID, p2pcommon.ProtocolForCID(centID), p2pEnv)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, p2pResp, "must be non nil")
	resp := resolveSignatureResponse(t, p2pResp)
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, cd.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignature_AlreadyExists(t *testing.T) {
	_, cd := prepareDocumentForP2PHandler(t, nil)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	resp, err := handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd})
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")

	resp, err = handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd})
	assert.NotNil(t, err, "must not be nil")
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists.Error())
}

func TestHandler_RequestDocumentSignature_UpdateSucceeds(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	po, cd := prepareDocumentForP2PHandler(t, nil)
	resp, err := handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd})
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, cd.SigningRoot, sig.Signature), "signature must be valid")

	//Update document
	po, cd = updateDocumentForP2Phandler(t, po)
	resp, err = handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd})
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig = resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, cd.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignatureFirstTimeOnUpdatedDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	po, cd := prepareDocumentForP2PHandler(t, nil)
	po, cd = updateDocumentForP2Phandler(t, po)
	assert.NotEqual(t, cd.DocumentIdentifier, cd.CurrentVersion)
	resp, err := handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd})
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, cd.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignature(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	_, cd := prepareDocumentForP2PHandler(t, nil)
	resp, err := handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd})
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, cd.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_SendAnchoredDocument_update_fail(t *testing.T) {
	centrifugeId := createIdentity(t)

	_, cd := prepareDocumentForP2PHandler(t, nil)

	// Anchor document
	idConfig, err := identity.GetIdentityConfig(cfg)
	anchorIDTyped, _ := anchors.ToAnchorID(cd.CurrentVersion)
	docRootTyped, _ := anchors.ToDocumentRoot(cd.DocumentRoot)
	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centrifugeId, docRootTyped)
	signature, _ := secp256k1.SignEthereum(messageToSign, idConfig.Keys[identity.KeyPurposeEthMsgAuth].PrivateKey)
	ctx := testingconfig.CreateAccountContext(t, cfg)
	anchorConfirmations, err := anchorRepo.CommitAnchor(ctx, anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")

	anchorResp, err := handler.SendAnchoredDocument(ctx, &p2ppb.AnchorDocumentRequest{Document: &cd}, idConfig.ID[:])
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelUpdateKeyNotFound.Error())
	assert.Nil(t, anchorResp)
}

func TestHandler_SendAnchoredDocument_EmptyDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	id, err := cfg.GetIdentityID()
	assert.NoError(t, err)
	resp, err := handler.SendAnchoredDocument(ctxh, &p2ppb.AnchorDocumentRequest{}, id)
	assert.NotNil(t, err)
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_SendAnchoredDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	centrifugeId := createIdentity(t)

	po, cd := prepareDocumentForP2PHandler(t, nil)
	resp, err := handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd})
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	// Add signature received
	po.AppendSignatures(resp.Signature)
	tree, err := po.DocumentRootTree()
	po.Document.DocumentRoot = tree.RootHash()

	// Anchor document
	idConfig, err := identity.GetIdentityConfig(cfg)
	anchorIDTyped, _ := anchors.ToAnchorID(po.Document.CurrentVersion)
	docRootTyped, _ := anchors.ToDocumentRoot(po.Document.DocumentRoot)
	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centrifugeId, docRootTyped)
	signature, _ := secp256k1.SignEthereum(messageToSign, idConfig.Keys[identity.KeyPurposeEthMsgAuth].PrivateKey)
	anchorConfirmations, err := anchorRepo.CommitAnchor(ctxh, anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")

	cd, err = po.PackCoreDocument()
	assert.NoError(t, err)
	anchorResp, err := handler.SendAnchoredDocument(ctxh, &p2ppb.AnchorDocumentRequest{Document: &cd}, idConfig.ID[:])
	assert.Nil(t, err)
	assert.NotNil(t, anchorResp, "must be non nil")
	assert.True(t, anchorResp.Accepted)
}

func createIdentity(t *testing.T) identity.CentID {
	// Create Identity
	centrifugeId, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	cfg.Set("identityId", centrifugeId.String())
	id, confirmations, err := idService.CreateIdentity(testingconfig.CreateAccountContext(t, cfg), centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")
	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	assert.Equal(t, centrifugeId, watchRegisteredIdentity.Identity.CentID(), "Resulting Identity should have the same ID as the input")

	idConfig, err := identity.GetIdentityConfig(cfg)
	// Add Keys
	pubKey := idConfig.Keys[identity.KeyPurposeP2P].PublicKey
	confirmations, err = id.AddKeyToIdentity(context.Background(), identity.KeyPurposeP2P, pubKey)
	assert.Nil(t, err, "should not error out when adding key to identity")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchReceivedIdentity := <-confirmations
	assert.Equal(t, centrifugeId, watchReceivedIdentity.Identity.CentID(), "Resulting Identity should have the same ID as the input")

	sPubKey := idConfig.Keys[identity.KeyPurposeSigning].PublicKey
	confirmations, err = id.AddKeyToIdentity(context.Background(), identity.KeyPurposeSigning, sPubKey)
	assert.Nil(t, err, "should not error out when adding key to identity")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchReceivedIdentity = <-confirmations
	assert.Equal(t, centrifugeId, watchReceivedIdentity.Identity.CentID(), "Resulting Identity should have the same ID as the input")

	secPubKey := idConfig.Keys[identity.KeyPurposeEthMsgAuth].PublicKey
	confirmations, err = id.AddKeyToIdentity(context.Background(), identity.KeyPurposeEthMsgAuth, secPubKey)
	assert.Nil(t, err, "should not error out when adding key to identity")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchReceivedIdentity = <-confirmations
	assert.Equal(t, centrifugeId, watchReceivedIdentity.Identity.CentID(), "Resulting Identity should have the same ID as the input")

	return centrifugeId
}

func prepareDocumentForP2PHandler(t *testing.T, po *purchaseorder.PurchaseOrder) (*purchaseorder.PurchaseOrder, coredocumentpb.CoreDocument) {
	idConfig, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	if po == nil {
		payalod := testingdocuments.CreatePOPayload()
		po = new(purchaseorder.PurchaseOrder)
		err = po.InitPurchaseOrderInput(payalod, idConfig.ID.String())
		assert.NoError(t, err)
	}
	_, err = po.DataRoot()
	assert.NoError(t, err)
	sr, err := po.SigningRoot()
	assert.NoError(t, err)
	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, sr)
	po.AppendSignatures(sig)
	_, err = po.DocumentRoot()
	assert.NoError(t, err)
	cd, err := po.PackCoreDocument()
	assert.NoError(t, err)
	return po, cd
}

func updateDocumentForP2Phandler(t *testing.T, po *purchaseorder.PurchaseOrder) (*purchaseorder.PurchaseOrder, coredocumentpb.CoreDocument) {
	cd, err := po.CoreDocument.PrepareNewVersion(nil, true)
	assert.NoError(t, err)
	po.CoreDocument = cd
	return prepareDocumentForP2PHandler(t, po)
}

func resolveSignatureResponse(t *testing.T, p2pEnv *protocolpb.P2PEnvelope) *p2ppb.SignatureResponse {
	signResp := new(p2ppb.SignatureResponse)
	dataEnv, err := p2pcommon.ResolveDataEnvelope(p2pEnv)
	assert.NoError(t, err)
	err = proto.Unmarshal(dataEnv.Body, signResp)
	assert.NoError(t, err)
	return signResp
}
