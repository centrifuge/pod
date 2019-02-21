// +build integration

package receiver_test

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/testingutils/documents"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/golang/protobuf/ptypes/any"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	cented25519 "github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
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

	dm := prepareDocumentForP2PHandler(t, nil)
	collab := centrifugeId.String()
	dm, err := dm.NewWithCollaborators([]string{collab})
	assert.NoError(t, err)
	dm, err = dm.PrepareNewVersion(nil)
	assert.NoError(t, err)
	req := getSignatureRequest(dm)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	// Add signature received
	doc := dm.Document
	doc.Signatures = append(doc.Signatures, resp.Signature)
	tree, err := dm.GetDocumentRootTree()
	assert.NoError(t, err)
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

	anchorReq := getAnchoredRequest(dm)
	anchorResp, err := handler.SendAnchoredDocument(ctxh, anchorReq, idConfig.ID[:])
	assert.Nil(t, err)
	assert.NotNil(t, anchorResp, "must be non nil")

	// Retrieve document from anchor repository with requester verification access type
	getReq := getDocumentRequestPeer(dm)
	getDocResp, err := handler.GetDocument(ctxh, getReq, centrifugeId)
	assert.Nil(t, err)
	assert.ObjectsAreEqual(getDocResp.Document, doc)

	// Retrieve document from anchor repository with access token verification access type
	// Retrieve document from anchor repository with nft verification access type
}

func TestHandler_HandleInterceptorReqSignature(t *testing.T) {
	centID := createIdentity(t)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	tc, err := contextutil.Account(ctxh)
	_, err = cfgService.CreateAccount(tc)
	assert.NoError(t, err)
	dm := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(dm)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctxh, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, req)

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
	doc := dm.Document
	assert.True(t, ed25519.Verify(sig.PublicKey, doc.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignature_verification_fail(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, nil)
	doc := dm.Document
	doc.SigningRoot = nil
	req := getSignatureRequest(dm)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.NotNil(t, err, "must be non nil")
	assert.Nil(t, resp, "must be nil")
	assert.Contains(t, err.Error(), "signing root missing")
}

func TestHandler_RequestDocumentSignature_AlreadyExists(t *testing.T) {
	dm := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(dm)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")

	req = getSignatureRequest(dm)
	resp, err = handler.RequestDocumentSignature(ctxh, req)
	assert.NotNil(t, err, "must not be nil")
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists.Error())
}

func TestHandler_RequestDocumentSignature_UpdateSucceeds(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(dm)
	doc := dm.Document
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, doc.SigningRoot, sig.Signature), "signature must be valid")
	//Update document
	newDM, err := dm.PrepareNewVersion(nil)
	assert.Nil(t, err)
	updateDocumentForP2Phandler(t, newDM)
	newDM = prepareDocumentForP2PHandler(t, newDM)
	req = getSignatureRequest(newDM)
	resp, err = handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig = resp.Signature
	newDoc := newDM.Document
	assert.True(t, ed25519.Verify(sig.PublicKey, newDoc.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignatureFirstTimeOnUpdatedDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, nil)
	newDM, err := dm.PrepareNewVersion(nil)
	assert.Nil(t, err)
	newDoc := newDM.Document
	assert.NotEqual(t, newDoc.DocumentIdentifier, newDoc.CurrentVersion)
	updateDocumentForP2Phandler(t, newDM)
	newDM = prepareDocumentForP2PHandler(t, newDM)
	req := getSignatureRequest(newDM)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, newDoc.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignature(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, nil)
	doc := dm.Document
	req := getSignatureRequest(dm)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig := resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, doc.SigningRoot, sig.Signature), "signature must be valid")
}

func TestHandler_SendAnchoredDocument_update_fail(t *testing.T) {
	centrifugeId := createIdentity(t)

	dm := prepareDocumentForP2PHandler(t, nil)

	// Anchor document
	doc := dm.Document
	idConfig, err := identity.GetIdentityConfig(cfg)
	anchorIDTyped, _ := anchors.ToAnchorID(doc.CurrentVersion)
	docRootTyped, _ := anchors.ToDocumentRoot(doc.DocumentRoot)
	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centrifugeId, docRootTyped)
	signature, _ := secp256k1.SignEthereum(messageToSign, idConfig.Keys[identity.KeyPurposeEthMsgAuth].PrivateKey)
	ctx := testingconfig.CreateAccountContext(t, cfg)
	anchorConfirmations, err := anchorRepo.CommitAnchor(ctx, anchorIDTyped, docRootTyped, centrifugeId, [][anchors.DocumentProofLength]byte{utils.RandomByte32()}, signature)
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")

	anchorReq := getAnchoredRequest(dm)
	anchorResp, err := handler.SendAnchoredDocument(ctx, anchorReq, idConfig.ID[:])
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelUpdateKeyNotFound.Error())
	assert.Nil(t, anchorResp)
}

func TestHandler_SendAnchoredDocument_EmptyDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, nil)
	req := getAnchoredRequest(doc)
	req.Document = nil
	id, err := cfg.GetIdentityID()
	assert.NoError(t, err)
	resp, err := handler.SendAnchoredDocument(ctxh, req, id)
	assert.NotNil(t, err)
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_SendAnchoredDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	centrifugeId := createIdentity(t)

	dm := prepareDocumentForP2PHandler(t, nil)
	req := getSignatureRequest(dm)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	// Add signature received
	doc := dm.Document
	doc.Signatures = append(doc.Signatures, resp.Signature)
	tree, _ := dm.GetDocumentRootTree()
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

	anchorReq := getAnchoredRequest(dm)
	anchorResp, err := handler.SendAnchoredDocument(ctxh, anchorReq, idConfig.ID[:])
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

func prepareDocumentForP2PHandler(t *testing.T, dm *documents.CoreDocumentModel) *documents.CoreDocumentModel {
	idConfig, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	if dm == nil {
		dm = testingdocuments.GenerateCoreDocumentModel()
	}

	m, err := docSrv.DeriveFromCoreDocumentModel(dm)
	assert.Nil(t, err)

	droot, err := m.CalculateDataRoot()
	assert.Nil(t, err)

	dm, err = m.PackCoreDocument()
	assert.NoError(t, err)

	tree, err := dm.GetDocumentSigningTree(droot)
	assert.NoError(t, err)
	doc := dm.Document
	doc.SigningRoot = tree.RootHash()
	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	tree, err = dm.GetDocumentRootTree()
	assert.NoError(t, err)
	doc.DocumentRoot = tree.RootHash()
	return dm
}

func updateDocumentForP2Phandler(t *testing.T, model *documents.CoreDocumentModel) {
	invData := &invoicepb.InvoiceData{}
	dataSalts, _ := documents.GenerateNewSalts(invData, "invoice", []byte{1, 0, 0, 0})

	serializedInv, err := proto.Marshal(invData)
	assert.NoError(t, err)
	doc := model.Document
	doc.EmbeddedData = &any.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   serializedInv,
	}
	doc.EmbeddedDataSalts = documents.ConvertToProtoSalts(dataSalts)
	cdSalts, _ := documents.GenerateNewSalts(doc, "", nil)
	doc.CoredocumentSalts = documents.ConvertToProtoSalts(cdSalts)
}

func getAnchoredRequest(dm *documents.CoreDocumentModel) *p2ppb.AnchorDocumentRequest {
	doc := *dm.Document
	return &p2ppb.AnchorDocumentRequest{Document: &doc}
}

func getSignatureRequest(dm *documents.CoreDocumentModel) *p2ppb.SignatureRequest {
	doc := *dm.Document
	return &p2ppb.SignatureRequest{Document: &doc}
}

func getDocumentRequestPeer(dm *documents.CoreDocumentModel) *p2ppb.GetDocumentRequest {
	doc := dm.Document
	return &p2ppb.GetDocumentRequest{
		DocumentIdentifier: doc.DocumentIdentifier,
		AccessType: p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}
}

func resolveSignatureResponse(t *testing.T, p2pEnv *protocolpb.P2PEnvelope) *p2ppb.SignatureResponse {
	signResp := new(p2ppb.SignatureResponse)
	dataEnv, err := p2pcommon.ResolveDataEnvelope(p2pEnv)
	assert.NoError(t, err)
	err = proto.Unmarshal(dataEnv.Body, signResp)
	assert.NoError(t, err)
	return signResp
}
