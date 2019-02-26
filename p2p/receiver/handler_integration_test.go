// +build integration

package receiver_test

import (
	"context"
	"flag"
	"math/big"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/testingutils/identity"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/ethereum/go-ethereum/common"

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
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

var (
	handler    *receiver.Handler
	anchorRepo anchors.AnchorRepository
	cfg        config.Configuration
	idService  identity.ServiceDID
	idFactory  identity.Factory
	cfgService config.Service
	docSrv     documents.Service
	defaultDID identity.DID
)

func TestMain(m *testing.M) {
	flag.Parse()
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgService = ctx[config.BootstrappedConfigStorage].(config.Service)
	docSrv = ctx[documents.BootstrappedDocumentService].(documents.Service)
	anchorRepo = ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	idService = ctx[identity.BootstrappedDIDService].(identity.ServiceDID)
	idFactory = ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	handler = receiver.New(cfgService, receiver.HandshakeValidator(cfg.GetNetworkID(), idService), docSrv)
	defaultDID = createIdentity(&testing.T{})
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

func updateDocument(t *testing.T, dm *documents.CoreDocumentModel, centrifugeId identity.DID, ctxh context.Context) {

	dm = prepareDocumentForP2PHandler(t, dm, centrifugeId)

	ed := dm.Document.EmbeddedData
	edsalts := dm.Document.EmbeddedDataSalts

	req := getSignatureRequest(dm)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.NoError(t, err)
	dm.Document.EmbeddedData = ed
	dm.Document.EmbeddedDataSalts = edsalts

	// Add signature received
	dm.Document.Signatures = append(dm.Document.Signatures, resp.Signature)
	tree, err := dm.GetDocumentRootTree()
	assert.NoError(t, err)
	dm.Document.DocumentRoot = tree.RootHash()

	cdSalts, _ := documents.GenerateNewSalts(dm.Document, "", nil)
	dm.Document.CoredocumentSalts = documents.ConvertToProtoSalts(cdSalts)
}

func TestHandler_GetDocumentSucceeds(t *testing.T) {
	// generate identity for testing
	didAddr, err := idFactory.CalculateIdentityAddress(context.Background())
	assert.NoError(t, err)
	did := identity.NewDID(*didAddr)
	tc, err := configstore.TempAccount("", cfg)
	assert.NoError(t, err)
	tcr := tc.(*configstore.Account)
	tcr.IdentityID = did[:]
	cid, err := testingidentity.CreateAccountIDWithKeys(cfg.GetEthereumContextWaitTimeout(), tcr, idService, idFactory)
	assert.NoError(t, err)

	// generate initial core doc with collaborators
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm, err := testingdocuments.GenerateCoreDocumentModelWithCollaborators([][]byte{cid[:]})
	assert.Nil(t, err)

	updateDocument(t, dm, cid, ctxh)

	// Retrieve document from repository with requester verification access type
	getReq := getDocumentRequestPeer(dm)
	getDocResp, err := handler.GetDocument(ctxh, getReq, cid)
	assert.Nil(t, err)
	assert.ObjectsAreEqual(getDocResp.Document, dm.Document)

	// Retrieve document from repository with access token verification access type
	// TODO : will fail until signature validation scheme is changed
	//docID := hexutil.Encode(dm.Document.DocumentIdentifier)
	//at := documentpb.AccessTokenParams{
	//	Grantee:            cid.String(),
	//	DocumentIdentifier: docID,
	//}
	//dm, err = dm.AddAccessTokenToReadRules(ctxh, at)
	//assert.NoError(t, err)
	//dm, err = dm.PrepareNewVersion(nil)
	//assert.NoError(t, err)
	//updateDocument(t, dm, cid, ctxh)
	//
	//role := dm.Document.Roles[1]
	//token := role.AccessTokens[0]
	//accessTokenReq := getDocumentRequestAccessToken(dm, token.Identifier)
	//getDocResp, err = handler.GetDocument(ctxh, accessTokenReq, cid)
	//assert.Nil(t, err)
	//assert.ObjectsAreEqual(getDocResp.Document, dm.Document)

	// Retrieve document from repository with nft verification access type
	// TODO: will currently always work because token owner is a collaborator
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(32)
	err = dm.AddNFTToReadRules(registry, tokenID)
	dm, err = dm.PrepareNewVersion(nil)
	assert.NoError(t, err)
	updateDocument(t, dm, cid, ctxh)

	nftReq := getDocumentRequestNft(dm, registry, tokenID)
	getDocResp, err = handler.GetDocument(ctxh, nftReq, cid)
	assert.Nil(t, err)
	assert.ObjectsAreEqual(getDocResp.Document, dm.Document)
}

func TestHandler_HandleInterceptorReqSignature(t *testing.T) {
	centID := createIdentity(t)
	tc, err := configstore.NewAccount("", cfg)
	assert.Nil(t, err)
	acc := tc.(*configstore.Account)
	acc.IdentityID = centID[:]
	ctxh, err := contextutil.New(context.Background(), acc)
	assert.Nil(t, err)
	_, err = cfgService.CreateAccount(acc)
	assert.NoError(t, err)
	dm := prepareDocumentForP2PHandler(t, nil, centID)
	req := getSignatureRequest(dm)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctxh, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, req)

	pub, _ := acc.GetP2PKeyPair()
	publicKey, err := cented25519.GetPublicSigningKey(pub)
	assert.NoError(t, err)
	var bPk [32]byte
	copy(bPk[:], publicKey)
	peerID, err := cented25519.PublicKeyToP2PKey(bPk)
	assert.NoError(t, err)

	p2pResp, err := handler.HandleInterceptor(ctxh, peerID, p2pcommon.ProtocolForDID(&centID), p2pEnv)
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
	dm := prepareDocumentForP2PHandler(t, nil, defaultDID)
	doc := dm.Document
	doc.SigningRoot = nil
	req := getSignatureRequest(dm)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.NotNil(t, err, "must be non nil")
	assert.Nil(t, resp, "must be nil")
	assert.Contains(t, err.Error(), "signing root missing")
}

func TestHandler_RequestDocumentSignature_AlreadyExists(t *testing.T) {
	dm := prepareDocumentForP2PHandler(t, nil, defaultDID)
	ed := dm.Document.EmbeddedData
	edsalts := dm.Document.EmbeddedDataSalts
	req := getSignatureRequest(dm)
	tc, err := configstore.NewAccount("", cfg)
	assert.Nil(t, err)
	acc := tc.(*configstore.Account)
	acc.IdentityID = defaultDID[:]
	ctxh, err := contextutil.New(context.Background(), acc)
	assert.Nil(t, err)

	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")

	dm.Document.EmbeddedData = ed
	dm.Document.EmbeddedDataSalts = edsalts

	req = getSignatureRequest(dm)
	resp, err = handler.RequestDocumentSignature(ctxh, req)
	assert.NotNil(t, err, "must not be nil")
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists.Error())
}

func TestHandler_RequestDocumentSignature_UpdateSucceeds(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, nil, defaultDID)
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
	newDM = prepareDocumentForP2PHandler(t, newDM, defaultDID)
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
	dm := prepareDocumentForP2PHandler(t, nil, defaultDID)
	newDM, err := dm.PrepareNewVersion(nil)
	assert.Nil(t, err)
	newDoc := newDM.Document
	assert.NotEqual(t, newDoc.DocumentIdentifier, newDoc.CurrentVersion)
	updateDocumentForP2Phandler(t, newDM)
	newDM = prepareDocumentForP2PHandler(t, newDM, defaultDID)
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
	dm := prepareDocumentForP2PHandler(t, nil, defaultDID)
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
	dm := prepareDocumentForP2PHandler(t, nil, centrifugeId)

	// Anchor document
	doc := dm.Document
	idConfig, err := identity.GetIdentityConfig(cfg)
	anchorIDTyped, _ := anchors.ToAnchorID(doc.CurrentVersion)
	docRootTyped, _ := anchors.ToDocumentRoot(doc.DocumentRoot)
	ctx := testingconfig.CreateAccountContext(t, cfg)
	anchorConfirmations, err := anchorRepo.CommitAnchor(ctx, anchorIDTyped, docRootTyped, [][anchors.DocumentProofLength]byte{utils.RandomByte32()})
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.True(t, watchCommittedAnchor, "No error should be thrown by context")

	anchorReq := getAnchoredRequest(dm)
	anchorResp, err := handler.SendAnchoredDocument(ctx, anchorReq, idConfig.ID[:])
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelUpdateKeyNotFound.Error())
	assert.Nil(t, anchorResp)
}

func TestHandler_SendAnchoredDocument_EmptyDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, nil, defaultDID)
	req := getAnchoredRequest(doc)
	req.Document = nil
	id, err := cfg.GetIdentityID()
	assert.NoError(t, err)
	resp, err := handler.SendAnchoredDocument(ctxh, req, id)
	assert.NotNil(t, err)
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_SendAnchoredDocument(t *testing.T) {
	tc, err := configstore.NewAccount("", cfg)
	assert.Nil(t, err)
	centrifugeId := createIdentity(t)
	acc := tc.(*configstore.Account)
	acc.IdentityID = centrifugeId[:]

	ctxh, err := contextutil.New(context.Background(), tc)
	assert.Nil(t, err)

	dm := prepareDocumentForP2PHandler(t, nil, centrifugeId)
	ed := dm.Document.EmbeddedData
	edsalts := dm.Document.EmbeddedDataSalts
	req := getSignatureRequest(dm)
	resp, err := handler.RequestDocumentSignature(ctxh, req)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	dm.Document.EmbeddedData = ed
	dm.Document.EmbeddedDataSalts = edsalts

	// Add signature received
	doc := dm.Document
	doc.Signatures = append(doc.Signatures, resp.Signature)
	tree, _ := dm.GetDocumentRootTree()
	doc.DocumentRoot = tree.RootHash()

	// Anchor document
	anchorIDTyped, _ := anchors.ToAnchorID(doc.CurrentVersion)
	docRootTyped, _ := anchors.ToDocumentRoot(doc.DocumentRoot)
	anchorConfirmations, err := anchorRepo.CommitAnchor(ctxh, anchorIDTyped, docRootTyped, [][anchors.DocumentProofLength]byte{utils.RandomByte32()})
	assert.Nil(t, err)

	watchCommittedAnchor := <-anchorConfirmations
	assert.True(t, watchCommittedAnchor, "No error should be thrown by context")

	anchorReq := getAnchoredRequest(dm)
	anchorResp, err := handler.SendAnchoredDocument(ctxh, anchorReq, centrifugeId[:])
	assert.Nil(t, err)
	assert.NotNil(t, anchorResp, "must be non nil")
	assert.True(t, anchorResp.Accepted)
}

func createIdentity(t *testing.T) identity.DID {
	// Create Identity
	didAddr, err := idFactory.CalculateIdentityAddress(context.Background())
	assert.NoError(t, err)
	tc, err := configstore.NewAccount("", cfg)
	assert.Nil(t, err)
	acc := tc.(*configstore.Account)
	acc.IdentityID = didAddr.Bytes()

	ctx, err := contextutil.New(context.Background(), tc)
	assert.Nil(t, err)
	did, err := idFactory.CreateIdentity(ctx)
	assert.Nil(t, err, "should not error out when creating identity")
	assert.Equal(t, did.String(), didAddr.String(), "Resulting Identity should have the same ID as the input")

	idConfig, err := identity.GetIdentityConfig(cfg)
	assert.NoError(t, err)
	// Add Keys
	pk, err := utils.SliceToByte32(idConfig.Keys[identity.KeyPurposeP2P].PublicKey)
	assert.NoError(t, err)
	keyDID := identity.NewKey(pk, big.NewInt(identity.KeyPurposeP2P), big.NewInt(identity.KeyTypeECDSA))
	err = idService.AddKey(ctx, keyDID)
	assert.Nil(t, err, "should not error out when adding key to identity")

	sPk, err := utils.SliceToByte32(idConfig.Keys[identity.KeyPurposeSigning].PublicKey)
	assert.NoError(t, err)
	keyDID = identity.NewKey(sPk, big.NewInt(identity.KeyPurposeSigning), big.NewInt(identity.KeyTypeECDSA))
	err = idService.AddKey(ctx, keyDID)
	assert.Nil(t, err, "should not error out when adding key to identity")

	secPk, err := utils.SliceToByte32(idConfig.Keys[identity.KeyPurposeEthMsgAuth].PublicKey)
	assert.NoError(t, err)
	keyDID = identity.NewKey(secPk, big.NewInt(identity.KeyPurposeEthMsgAuth), big.NewInt(identity.KeyTypeECDSA))
	err = idService.AddKey(ctx, keyDID)
	assert.Nil(t, err, "should not error out when adding key to identity")

	return *did
}

func prepareDocumentForP2PHandler(t *testing.T, dm *documents.CoreDocumentModel, did identity.DID) *documents.CoreDocumentModel {
	idConfig, err := identity.GetIdentityConfig(cfg)
	idConfig.ID = did
	assert.NoError(t, err)
	if dm == nil {
		dm, err = testingdocuments.GenerateCoreDocumentModel()
		assert.Nil(t, err)
	}
	m, err := docSrv.DeriveFromCoreDocumentModel(dm)
	assert.Nil(t, err)

	droot, err := m.CalculateDataRoot()
	assert.Nil(t, err)

	dm, err = m.PackCoreDocument()
	assert.NoError(t, err)

	doc := dm.Document
	tree, err := dm.GetDocumentSigningTree(droot)
	assert.NoError(t, err)
	doc.SigningRoot = tree.RootHash()
	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, doc.SigningRoot)
	doc.Signatures = []*coredocumentpb.Signature{sig}
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
	cdSalts, _ := documents.GenerateCoreDocSalts(doc)
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
	return &p2ppb.GetDocumentRequest{
		DocumentIdentifier: dm.Document.DocumentIdentifier,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_REQUESTER_VERIFICATION,
	}
}

func getDocumentRequestNft(dm *documents.CoreDocumentModel, registry common.Address, tokenID []byte) *p2ppb.GetDocumentRequest {
	return &p2ppb.GetDocumentRequest{
		DocumentIdentifier: dm.Document.DocumentIdentifier,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_NFT_OWNER_VERIFICATION,
		NftRegistryAddress: registry[:],
		NftTokenId:         tokenID,
	}
}

func getDocumentRequestAccessToken(dm *documents.CoreDocumentModel, tokenID []byte) *p2ppb.GetDocumentRequest {
	atr := &p2ppb.AccessTokenRequest{
		DelegatingDocumentIdentifier: dm.Document.DocumentIdentifier,
		AccessTokenId:                tokenID,
	}
	return &p2ppb.GetDocumentRequest{
		DocumentIdentifier: dm.Document.DocumentIdentifier,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: atr,
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
