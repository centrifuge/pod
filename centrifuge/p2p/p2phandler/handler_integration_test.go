// +build integration

package p2phandler_test

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	cented25519 "github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519keys"
	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/centrifuge/go-centrifuge/centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/centrifuge/p2p/p2phandler"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

var handler = p2phandler.Handler{Notifier: &notification.WebhookSender{}}

func TestMain(m *testing.M) {
	cc.DONT_USE_FOR_UNIT_TESTS_TestFunctionalEthereumBootstrap()
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
	savedService := identity.IDService
	idConfig, err := cented25519.GetIDConfig()
	assert.Nil(t, err)
	centID, _ := identity.ToCentID(idConfig.ID)
	pubKey := idConfig.PublicKey
	b32Key, _ := utils.SliceToByte32(pubKey)
	idkey := &identity.EthereumIdentityKey{
		Key:       b32Key,
		Purposes:  []*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	id := &testingcommons.MockID{}
	srv := &testingcommons.MockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
	identity.IDService = srv

	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")

	id = &testingcommons.MockID{}
	srv = &testingcommons.MockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
	identity.IDService = srv

	req = getSignatureRequest(doc)
	resp, err = handler.RequestDocumentSignature(context.Background(), req)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.NotNil(t, err, "must not be nil")
	assert.Contains(t, err.Error(), "document already exists")

	identity.IDService = savedService
}

func TestHandler_RequestDocumentSignature_UpdateSucceeds(t *testing.T) {
	savedService := identity.IDService
	idConfig, err := cented25519.GetIDConfig()
	assert.Nil(t, err)
	centID, _ := identity.ToCentID(idConfig.ID)
	pubKey := idConfig.PublicKey
	b32Key, _ := utils.SliceToByte32(pubKey)
	idkey := &identity.EthereumIdentityKey{
		Key:       b32Key,
		Purposes:  []*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	id := &testingcommons.MockID{}
	srv := &testingcommons.MockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
	identity.IDService = srv

	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, doc.SigningRoot, sig.Signature), "signature must be valid")

	//Update document
	newDoc, err := coredocument.PrepareNewVersion(*doc, nil)
	assert.Nil(t, err)
	id = &testingcommons.MockID{}
	srv = &testingcommons.MockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
	identity.IDService = srv

	updateDocumentForP2Phandler(t, newDoc)
	newDoc = prepareDocumentForP2PHandler(t, newDoc)
	req = getSignatureRequest(newDoc)
	resp, err = handler.RequestDocumentSignature(context.Background(), req)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig = resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, newDoc.SigningRoot, sig.Signature), "signature must be valid")
	identity.IDService = savedService
}

func TestHandler_RequestDocumentSignature(t *testing.T) {
	savedService := identity.IDService
	idConfig, err := cented25519.GetIDConfig()
	assert.Nil(t, err)
	centID, _ := identity.ToCentID(idConfig.ID)
	pubKey := idConfig.PublicKey
	b32Key, _ := utils.SliceToByte32(pubKey)
	idkey := &identity.EthereumIdentityKey{
		Key:       b32Key,
		Purposes:  []*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	id := &testingcommons.MockID{}
	srv := &testingcommons.MockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", pubKey).Return(idkey, nil).Once()
	identity.IDService = srv

	doc := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(doc)
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, doc.SigningRoot, sig.Signature), "signature must be valid")

	identity.IDService = savedService
}

func TestHandler_SendAnchoredDocument_update_fail(t *testing.T) {
	centrifugeId := createIdentity(t)

	doc := prepareDocumentForP2PHandler(t, nil)

	// Anchor document
	secpIDConfig, err := secp256k1.GetIDConfig()
	anchorIDTyped, _ := anchors.NewAnchorID(doc.CurrentVersion)
	docRootTyped, _ := anchors.NewDocRoot(doc.DocumentRoot)
	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centrifugeId, docRootTyped)
	signature, _ := secp256k1.SignEthereum(messageToSign, secpIDConfig.PrivateKey)
	anchorConfirmations, err := anchors.CommitAnchor(anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
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
	secpIDConfig, err := secp256k1.GetIDConfig()
	anchorIDTyped, _ := anchors.NewAnchorID(doc.CurrentVersion)
	docRootTyped, _ := anchors.NewDocRoot(doc.DocumentRoot)
	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centrifugeId, docRootTyped)
	signature, _ := secp256k1.SignEthereum(messageToSign, secpIDConfig.PrivateKey)
	anchorConfirmations, err := anchors.CommitAnchor(anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
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
	config.Config.V.Set("identityId", centrifugeId.String())
	id, confirmations, err := identity.IDService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")
	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	assert.Equal(t, centrifugeId, watchRegisteredIdentity.Identity.GetCentrifugeID(), "Resulting Identity should have the same ID as the input")

	// Add Keys
	idConfig, err := cented25519.GetIDConfig()
	pubKey := idConfig.PublicKey
	confirmations, err = id.AddKeyToIdentity(context.Background(), identity.KeyPurposeSigning, pubKey)
	assert.Nil(t, err, "should not error out when adding key to identity")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchReceivedIdentity := <-confirmations
	assert.Equal(t, centrifugeId, watchReceivedIdentity.Identity.GetCentrifugeID(), "Resulting Identity should have the same ID as the input")

	secpIDConfig, err := secp256k1.GetIDConfig()
	secPubKey := secpIDConfig.PublicKey
	confirmations, err = id.AddKeyToIdentity(context.Background(), identity.KeyPurposeEthMsgAuth, secPubKey)
	assert.Nil(t, err, "should not error out when adding key to identity")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchReceivedIdentity = <-confirmations
	assert.Equal(t, centrifugeId, watchReceivedIdentity.Identity.GetCentrifugeID(), "Resulting Identity should have the same ID as the input")

	return centrifugeId
}

func prepareDocumentForP2PHandler(t *testing.T, doc *coredocumentpb.CoreDocument) *coredocumentpb.CoreDocument {
	idConfig, err := cented25519.GetIDConfig()
	assert.Nil(t, err)
	if doc == nil {
		doc = testingutils.GenerateCoreDocument()
	}
	tree, _ := coredocument.GetDocumentSigningTree(doc)
	doc.SigningRoot = tree.RootHash()
	sig := signatures.Sign(&config.IdentityConfig{
		ID:         idConfig.ID,
		PublicKey:  idConfig.PublicKey,
		PrivateKey: idConfig.PrivateKey,
	}, doc.SigningRoot)
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
			NetworkIdentifier: config.Config.GetNetworkID(),
		}, Document: doc,
	}
}

func getSignatureRequest(doc *coredocumentpb.CoreDocument) *p2ppb.SignatureRequest {
	return &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID(),
	}, Document: doc}
}
