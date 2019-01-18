// +build integration

package receiver_test

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/coredocument"
	cented25519 "github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/coredocument"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

var (
	handler    *receiver.Handler
	anchorRepo anchors.AnchorRepository
	cfg        config.Configuration
	idService  identity.Service
	cfgService config.Service
)

func TestMain(m *testing.M) {
	flag.Parse()
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgService = ctx[config.BootstrappedConfigStorage].(config.Service)
	docSrv := ctx[documents.BootstrappedDocumentService].(documents.Service)
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
	req := &p2ppb.GetDocumentRequest{DocumentIdentifier: b}
	resp, err := handler.GetDocument(context.Background(), req)
	assert.Error(t, err, "must return error")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_GetDocumentSucceeds(t *testing.T) {
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	centrifugeId := createIdentity(t)

	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	// Add signature received
	doc.Signatures = append(doc.Signatures, resp.Signature)
	tree, _ := coredocument.GetDocumentRootTree(doc)
	doc.DocumentRoot = tree.RootHash()

	// Anchor document
	idConfig, err := identity.GetIdentityConfig(cfg)
	anchorIDTyped, _ := anchors.ToAnchorID(doc.CurrentVersion)
	docRootTyped, _ := anchors.ToDocumentRoot(doc.DocumentRoot)
	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centrifugeId, docRootTyped)
	signature, _ := secp256k1.SignEthereum(messageToSign, idConfig.Keys[identity.KeyPurposeEthMsgAuth].PrivateKey)
	anchorConfirmations, err := anchorRepo.CommitAnchor(ctxh, anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")

	anchorReq := getAnchoredRequest(doc)
	anchorResp, err := handler.SendAnchoredDocument(ctxh, anchorReq, idConfig.ID[:])
	assert.Nil(t, err)
	assert.NotNil(t, anchorResp, "must be non nil")

	// Retrieve document from anchor repository with document_identifier
	getReq := getGetDocumentRequest(doc)
	getDocResp, err := handler.GetDocument(ctxh, getReq)
	assert.Nil(t, err)
	assert.ObjectsAreEqual(getDocResp.Document, doc)
}

func TestHandler_HandleInterceptorReqSignature(t *testing.T) {
	centID := createIdentity(t)
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	tc, err := contextutil.Account(ctxh)
	_, err = cfgService.CreateAccount(tc)
	assert.NoError(t, err)
	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctxh, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, req)

	pub, _ := cfg.GetSigningKeyPair()
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
	assert.True(t, ed25519.Verify(sig.PublicKey, doc.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignature_verification_fail(t *testing.T) {
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, nil)
	doc.SigningRoot = nil
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.NotNil(t, err, "must be non nil")
	assert.Nil(t, resp, "must be nil")
	assert.Contains(t, err.Error(), "signing root missing")
}

func TestHandler_RequestDocumentSignature_AlreadyExists(t *testing.T) {
	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")

	req = getSignatureRequest(doc)
	resp, err = handler.RequestDocumentSignature(ctxh, req)
	assert.NotNil(t, err, "must not be nil")
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists.Error())
}

func TestHandler_RequestDocumentSignature_UpdateSucceeds(t *testing.T) {
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, doc.SigningRoot, sig.Signature), "signature must be valid")
	//Update document
	newDoc, err := coredocument.PrepareNewVersion(*doc, nil)
	assert.Nil(t, err)
	updateDocumentForP2Phandler(t, newDoc)
	newDoc = prepareDocumentForP2PHandler(t, newDoc)
	req = getSignatureRequest(newDoc)
	resp, err = handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig = resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, newDoc.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignatureFirstTimeOnUpdatedDocument(t *testing.T) {
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, nil)
	newDoc, err := coredocument.PrepareNewVersion(*doc, nil)
	assert.Nil(t, err)
	assert.NotEqual(t, newDoc.DocumentIdentifier, newDoc.CurrentVersion)
	updateDocumentForP2Phandler(t, newDoc)
	newDoc = prepareDocumentForP2PHandler(t, newDoc)
	req := getSignatureRequest(newDoc)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, newDoc.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignature(t *testing.T) {
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, doc.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_SendAnchoredDocument_update_fail(t *testing.T) {
	centrifugeId := createIdentity(t)

	doc := prepareDocumentForP2PHandler(t, nil)

	// Anchor document
	idConfig, err := identity.GetIdentityConfig(cfg)
	anchorIDTyped, _ := anchors.ToAnchorID(doc.CurrentVersion)
	docRootTyped, _ := anchors.ToDocumentRoot(doc.DocumentRoot)
	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centrifugeId, docRootTyped)
	signature, _ := secp256k1.SignEthereum(messageToSign, idConfig.Keys[identity.KeyPurposeEthMsgAuth].PrivateKey)
	ctx := testingconfig.CreateTenantContext(t, cfg)
	anchorConfirmations, err := anchorRepo.CommitAnchor(ctx, anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")

	anchorReq := getAnchoredRequest(doc)
	anchorResp, err := handler.SendAnchoredDocument(ctx, anchorReq, idConfig.ID[:])
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelUpdateKeyNotFound.Error())
	assert.Nil(t, anchorResp)
}

func TestHandler_SendAnchoredDocument_EmptyDocument(t *testing.T) {
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, nil)
	req := getAnchoredRequest(doc)
	req.Document = nil
	id, _ := cfg.GetIdentityID()
	resp, err := handler.SendAnchoredDocument(ctxh, req, id)
	assert.NotNil(t, err)
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_SendAnchoredDocument(t *testing.T) {
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	centrifugeId := createIdentity(t)

	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	// Add signature received
	doc.Signatures = append(doc.Signatures, resp.Signature)
	tree, _ := coredocument.GetDocumentRootTree(doc)
	doc.DocumentRoot = tree.RootHash()

	// Anchor document
	idConfig, err := identity.GetIdentityConfig(cfg)
	anchorIDTyped, _ := anchors.ToAnchorID(doc.CurrentVersion)
	docRootTyped, _ := anchors.ToDocumentRoot(doc.DocumentRoot)
	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centrifugeId, docRootTyped)
	signature, _ := secp256k1.SignEthereum(messageToSign, idConfig.Keys[identity.KeyPurposeEthMsgAuth].PrivateKey)
	anchorConfirmations, err := anchorRepo.CommitAnchor(ctxh, anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")

	anchorReq := getAnchoredRequest(doc)
	anchorResp, err := handler.SendAnchoredDocument(ctxh, anchorReq, idConfig.ID[:])
	assert.Nil(t, err)
	assert.NotNil(t, anchorResp, "must be non nil")
	assert.True(t, anchorResp.Accepted)
}

func createIdentity(t *testing.T) identity.CentID {
	// Create Identity
	centrifugeId, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	cfg.Set("identityId", centrifugeId.String())
	id, confirmations, err := idService.CreateIdentity(testingconfig.CreateTenantContext(t, cfg), centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")
	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	assert.Equal(t, centrifugeId, watchRegisteredIdentity.Identity.CentID(), "Resulting Identity should have the same ID as the input")

	idConfig, err := identity.GetIdentityConfig(cfg)
	// Add Keys
	pubKey := idConfig.Keys[identity.KeyPurposeSigning].PublicKey
	confirmations, err = id.AddKeyToIdentity(context.Background(), identity.KeyPurposeSigning, pubKey)
	assert.Nil(t, err, "should not error out when adding key to identity")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchReceivedIdentity := <-confirmations
	assert.Equal(t, centrifugeId, watchReceivedIdentity.Identity.CentID(), "Resulting Identity should have the same ID as the input")

	secPubKey := idConfig.Keys[identity.KeyPurposeEthMsgAuth].PublicKey
	confirmations, err = id.AddKeyToIdentity(context.Background(), identity.KeyPurposeEthMsgAuth, secPubKey)
	assert.Nil(t, err, "should not error out when adding key to identity")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchReceivedIdentity = <-confirmations
	assert.Equal(t, centrifugeId, watchReceivedIdentity.Identity.CentID(), "Resulting Identity should have the same ID as the input")

	return centrifugeId
}

func prepareDocumentForP2PHandler(t *testing.T, doc *coredocumentpb.CoreDocument) *coredocumentpb.CoreDocument {
	idConfig, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	if doc == nil {
		doc = testingcoredocument.GenerateCoreDocument()
	}
	tree, _ := coredocument.GetDocumentSigningTree(doc)
	doc.SigningRoot = tree.RootHash()
	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	tree, _ = coredocument.GetDocumentRootTree(doc)
	doc.DocumentRoot = tree.RootHash()
	return doc
}

func updateDocumentForP2Phandler(t *testing.T, doc *coredocumentpb.CoreDocument) {
	salts := &coredocumentpb.CoreDocumentSalts{}
	doc.CoredocumentSalts = salts
	doc.DataRoot = utils.RandomSlice(32)
	doc.EmbeddedData = &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
	}
	doc.EmbeddedDataSalts = &any.Any{
		TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
	}
	err := proofs.FillSalts(doc, salts)
	assert.Nil(t, err)
}

func getAnchoredRequest(doc *coredocumentpb.CoreDocument) *p2ppb.AnchorDocumentRequest {
	return &p2ppb.AnchorDocumentRequest{Document: doc}
}

func getSignatureRequest(doc *coredocumentpb.CoreDocument) *p2ppb.SignatureRequest {
	return &p2ppb.SignatureRequest{Document: doc}
}

func getGetDocumentRequest(doc *coredocumentpb.CoreDocument) *p2ppb.GetDocumentRequest {
	return &p2ppb.GetDocumentRequest{DocumentIdentifier: doc.DocumentIdentifier}
}

func resolveSignatureResponse(t *testing.T, p2pEnv *protocolpb.P2PEnvelope) *p2ppb.SignatureResponse {
	signResp := new(p2ppb.SignatureResponse)
	dataEnv, err := p2pcommon.ResolveDataEnvelope(p2pEnv)
	assert.NoError(t, err)
	err = proto.Unmarshal(dataEnv.Body, signResp)
	assert.NoError(t, err)
	return signResp
}
