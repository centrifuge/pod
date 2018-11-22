// +build integration

package p2p_test

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/keytools/secp256k1"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/testingutils/coredocument"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

var (
	handler    p2ppb.P2PServiceServer
	anchorRepo anchors.AnchorRepository
	cfg        *config.Configuration
	idService  identity.Service
)

func TestMain(m *testing.M) {
	flag.Parse()
	ctx := cc.TestFunctionalEthereumBootstrap()
	registry := ctx[documents.BootstrappedRegistry].(*documents.ServiceRegistry)
	anchorRepo = ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	idService = ctx[identity.BootstrappedIDService].(identity.Service)
	cfg = ctx[config.BootstrappedConfig].(*config.Configuration)
	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
	cfg.Set("keys.ethauth.publicKey", "../build/resources/ethauth.pub.pem")
	cfg.Set("keys.ethauth.privateKey", "../build/resources/ethauth.key.pem")
	handler = p2p.GRPCHandler(cfg, registry)
	testingidentity.CreateIdentityWithKeys(cfg, idService)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestHandler_RequestDocumentSignature_verification_fail(t *testing.T) {
	doc := prepareDocumentForP2PHandler(t, nil)
	doc.SigningRoot = nil
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.NotNil(t, err, "must be non nil")
	assert.Nil(t, resp, "must be nil")
	assert.Contains(t, err.Error(), "signing root missing")
}

func TestHandler_RequestDocumentSignature_AlreadyExists(t *testing.T) {
	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")

	req = getSignatureRequest(doc)
	resp, err = handler.RequestDocumentSignature(context.Background(), req)
	assert.NotNil(t, err, "must not be nil")
	assert.Contains(t, err.Error(), "document already exists")
}

func TestHandler_RequestDocumentSignature_UpdateSucceeds(t *testing.T) {
	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
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
	resp, err = handler.RequestDocumentSignature(context.Background(), req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig = resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, newDoc.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignature(t *testing.T) {
	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
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
	anchorConfirmations, err := anchorRepo.CommitAnchor(anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")

	anchorReq := getAnchoredRequest(doc)
	anchorResp, err := handler.SendAnchoredDocument(context.Background(), anchorReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document doesn't exists")
	assert.Nil(t, anchorResp)
}

func TestHandler_SendAnchoredDocument_EmptyDocument(t *testing.T) {
	doc := prepareDocumentForP2PHandler(t, nil)
	req := getAnchoredRequest(doc)
	req.Document = nil
	resp, err := handler.SendAnchoredDocument(context.Background(), req)
	assert.NotNil(t, err)
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_SendAnchoredDocument(t *testing.T) {
	centrifugeId := createIdentity(t)

	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
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
	anchorConfirmations, err := anchorRepo.CommitAnchor(anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")

	anchorReq := getAnchoredRequest(doc)
	anchorResp, err := handler.SendAnchoredDocument(context.Background(), anchorReq)
	assert.Nil(t, err)
	assert.NotNil(t, anchorResp, "must be non nil")
	assert.True(t, anchorResp.Accepted)
	assert.Equal(t, anchorResp.CentNodeVersion, anchorReq.Header.CentNodeVersion)
}

func createIdentity(t *testing.T) identity.CentID {
	// Create Identity
	centrifugeId, _ := identity.ToCentID(utils.RandomSlice(identity.CentIDLength))
	cfg.Set("identityId", centrifugeId.String())
	id, confirmations, err := idService.CreateIdentity(centrifugeId)
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
	return &p2ppb.AnchorDocumentRequest{
		Header: &p2ppb.CentrifugeHeader{
			CentNodeVersion:   version.GetVersion().String(),
			NetworkIdentifier: cfg.GetNetworkID(),
		}, Document: doc,
	}
}

func getSignatureRequest(doc *coredocumentpb.CoreDocument) *p2ppb.SignatureRequest {
	return &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: cfg.GetNetworkID(),
	}, Document: doc}
}
