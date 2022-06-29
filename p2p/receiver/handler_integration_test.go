//go:build integration
// +build integration

package receiver_test

import (
	"context"
	"flag"
	"os"
	"sync"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	protocolpb "github.com/centrifuge/centrifuge-protobufs/gen/go/protocol"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	cented25519 "github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs"
	p2pcommon "github.com/centrifuge/go-centrifuge/p2p/common"
	"github.com/centrifuge/go-centrifuge/p2p/receiver"
	"github.com/centrifuge/go-centrifuge/storage"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var (
	handler    *receiver.Handler
	anchorSrv  anchors.Service
	cfg        config.Configuration
	idService  identity.Service
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
	anchorSrv = ctx[anchors.BootstrappedAnchorService].(anchors.Service)
	idService = ctx[identity.BootstrappedDIDService].(identity.Service)
	handler = receiver.New(cfgService, receiver.HandshakeValidator(cfg.GetNetworkID(), idService), docSrv, new(testingdocuments.MockRegistry), idService)
	dispatcher := ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	ctxh, canc := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go dispatcher.Start(ctxh, wg, nil)
	defaultDID = ideth.DeployIdentity(new(testing.T), ctx, cfg)
	errors.MaskErrs = false
	result := m.Run()
	canc()
	wg.Wait()
	testingbootstrap.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestHandler_GetDocument_nonexistentIdentifier(t *testing.T) {
	b := utils.RandomSlice(32)
	req := &p2ppb.GetDocumentRequest{DocumentIdentifier: b}
	resp, err := handler.GetDocument(context.Background(), req, defaultDID)
	assert.Error(t, err, "must return error")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_HandleInterceptorReqSignature(t *testing.T) {
	tc, err := configstore.NewAccount("main", cfg)
	assert.Nil(t, err)
	acc := tc.(*configstore.Account)
	acc.IdentityID = defaultDID[:]
	ctxh, err := contextutil.New(context.Background(), acc)
	assert.Nil(t, err)
	_, err = cfgService.CreateAccount(acc)
	assert.NoError(t, err)
	doc, cd := prepareDocumentForP2PHandler(t, nil)
	p2pEnv, err := p2pcommon.PrepareP2PEnvelope(ctxh, cfg.GetNetworkID(), p2pcommon.MessageTypeRequestSignature, &p2ppb.SignatureRequest{Document: &cd})

	pub, _ := acc.GetP2PKeyPair()
	publicKey, err := cented25519.GetPublicSigningKey(pub)
	assert.NoError(t, err)
	var bPk [32]byte
	copy(bPk[:], publicKey)
	peerID, err := cented25519.PublicKeyToP2PKey(bPk)
	assert.NoError(t, err)

	p2pResp, err := handler.HandleInterceptor(ctxh, peerID, p2pcommon.ProtocolForDID(defaultDID), p2pEnv)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, p2pResp, "must be non nil")
	resp := resolveSignatureResponse(t, p2pResp)
	assert.NotNil(t, resp.Signatures[0].Signature, "must be non nil")
	sig := resp.Signatures[0]
	signingRoot, err := doc.CalculateSigningRoot()
	assert.NoError(t, err)
	payload := documents.ConsensusSignaturePayload(signingRoot, false)
	assert.True(t, cented25519.VerifySignature(sig.PublicKey, payload, sig.Signature), "signature must be valid")
}

func TestHandler_RequestDocumentSignature(t *testing.T) {
	tc, err := configstore.NewAccount("main", cfg)
	assert.Nil(t, err)
	acc := tc.(*configstore.Account)
	acc.IdentityID = defaultDID[:]

	ctxh, err := contextutil.New(context.Background(), acc)
	assert.Nil(t, err)

	doc, cd := prepareDocumentForP2PHandler(t, nil)

	// nil sigRequest
	id2 := testingidentity.GenerateRandomDID()
	_, err = handler.RequestDocumentSignature(ctxh, nil, defaultDID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil document provided")

	// requestDocumentSignature, no previous versions
	_, err = handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd}, defaultDID)
	assert.NoError(t, err)

	// we can update the document so that there are two versions in the repo
	doc, ncd := updateDocumentForP2Phandler(t, doc)
	assert.NotEqual(t, cd.DocumentIdentifier, ncd.CurrentVersion)

	// invalid transition for non-collaborator id
	_, err = handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &ncd}, id2)
	assert.Error(t, err)

	// valid transition for collaborator id
	resp, err := handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &ncd}, defaultDID)
	assert.NoError(t, err)
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signatures[0].Signature, "must be non nil")
	sig := resp.Signatures[0]
	signingRoot, err := doc.CalculateSigningRoot()
	assert.NoError(t, err)
	payload := documents.ConsensusSignaturePayload(signingRoot, true)
	assert.True(t, cented25519.VerifySignature(sig.PublicKey, payload, sig.Signature), "signature must be valid")

	// document already exists
	_, err = handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd}, defaultDID)
	assert.NotNil(t, err, "must not be nil")
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelCreateKeyExists.Error())
}

func TestHandler_SendAnchoredDocument_update_fail(t *testing.T) {
	doc, cd := prepareDocumentForP2PHandler(t, nil)
	ctx := testingconfig.CreateAccountContext(t, cfg)

	// Anchor document
	accDID, err := contextutil.AccountDID(ctx)
	assert.NoError(t, err)
	anchorIDTyped, err := anchors.ToAnchorID(cd.CurrentPreimage)
	assert.NoError(t, err)
	docRoot, err := doc.CalculateDocumentRoot()
	assert.NoError(t, err)
	docRootTyped, err := anchors.ToDocumentRoot(docRoot)
	assert.NoError(t, err)

	err = anchorSrv.CommitAnchor(ctx, anchorIDTyped, docRootTyped, utils.RandomByte32())
	assert.Nil(t, err)

	anchorResp, err := handler.SendAnchoredDocument(ctx, &p2ppb.AnchorDocumentRequest{Document: &cd}, accDID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), storage.ErrRepositoryModelUpdateKeyNotFound.Error())
	assert.Nil(t, anchorResp)
}

func TestHandler_SendAnchoredDocument_EmptyDocument(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	id, err := cfg.GetIdentityID()
	collaborator, err := identity.NewDIDFromBytes(id)
	assert.NoError(t, err)
	resp, err := handler.SendAnchoredDocument(ctxh, &p2ppb.AnchorDocumentRequest{}, collaborator)
	assert.NotNil(t, err)
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_SendAnchoredDocument(t *testing.T) {
	tc, err := configstore.NewAccount("main", cfg)
	assert.Nil(t, err)
	acc := tc.(*configstore.Account)
	acc.IdentityID = defaultDID[:]

	ctxh, err := contextutil.New(context.Background(), acc)
	assert.Nil(t, err)

	doc, cd := prepareDocumentForP2PHandler(t, nil)
	resp, err := handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &cd}, defaultDID)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	// Add signature received
	doc.AppendSignatures(resp.Signatures...)

	// Since we have changed the coredocument by adding signatures lets generate salts again
	rootHash, err := doc.CalculateDocumentRoot()
	assert.NoError(t, err)

	// Anchor document
	anchorIDTyped, err := anchors.ToAnchorID(doc.GetTestCoreDocWithReset().CurrentPreimage)
	assert.NoError(t, err)
	docRootTyped, err := anchors.ToDocumentRoot(rootHash)
	assert.NoError(t, err)

	err = anchorSrv.CommitAnchor(ctxh, anchorIDTyped, docRootTyped, utils.RandomByte32())
	assert.Nil(t, err)

	cd, err = doc.PackCoreDocument()
	assert.NoError(t, err)

	// this should succeed since this is the first document version
	anchorResp, err := handler.SendAnchoredDocument(ctxh, &p2ppb.AnchorDocumentRequest{Document: &cd}, defaultDID)
	assert.Nil(t, err)
	assert.NotNil(t, anchorResp, "must be non nil")
	assert.True(t, anchorResp.Accepted)

	// update the document
	npo, ncd := updateDocumentForP2Phandler(t, doc)
	resp, err = handler.RequestDocumentSignature(ctxh, &p2ppb.SignatureRequest{Document: &ncd}, defaultDID)
	assert.Nil(t, err)
	assert.NotNil(t, resp)

	// Add signature received
	npo.AppendSignatures(resp.Signatures...)
	rootHash, err = npo.CalculateDocumentRoot()
	assert.NoError(t, err)

	// Anchor document
	anchorIDTyped, err = anchors.ToAnchorID(npo.GetTestCoreDocWithReset().CurrentPreimage)
	assert.NoError(t, err)
	docRootTyped, err = anchors.ToDocumentRoot(rootHash)
	assert.NoError(t, err)
	err = anchorSrv.CommitAnchor(ctxh, anchorIDTyped, docRootTyped, utils.RandomByte32())
	assert.Nil(t, err)

	ncd, err = npo.PackCoreDocument()
	assert.NoError(t, err)

	// transition failure for random ID
	id := testingidentity.GenerateRandomDID()
	_, err = handler.SendAnchoredDocument(ctxh, &p2ppb.AnchorDocumentRequest{Document: &ncd}, id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid document state transition")

	anchorResp, err = handler.SendAnchoredDocument(ctxh, &p2ppb.AnchorDocumentRequest{Document: &ncd}, defaultDID)
	assert.Nil(t, err)
	assert.NotNil(t, anchorResp, "must be non nil")
	assert.True(t, anchorResp.Accepted)
}

func prepareDocumentForP2PHandler(t *testing.T, g *generic.Generic) (*generic.Generic, coredocumentpb.CoreDocument) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	accCfg, err := contextutil.Account(ctx)
	assert.NoError(t, err)
	acc := accCfg.(*configstore.Account)
	acc.IdentityID = defaultDID[:]
	accKeys, err := acc.GetKeys()
	assert.NoError(t, err)
	if g == nil {
		g = generic.InitGeneric(t, defaultDID, generic.CreateGenericPayload(t, nil))
	}
	g.SetUsedAnchorRepoAddress(cfg.GetContractAddress(config.AnchorRepo))
	err = g.AddUpdateLog(defaultDID)
	assert.NoError(t, err)
	sr, err := g.CalculateSigningRoot()
	assert.NoError(t, err)
	s, err := crypto.SignMessage(accKeys[identity.KeyPurposeSigning.Name].PrivateKey, documents.ConsensusSignaturePayload(sr, false), crypto.CurveEd25519)
	assert.NoError(t, err)
	sig := &coredocumentpb.Signature{
		SignatureId:         append(defaultDID[:], accKeys[identity.KeyPurposeSigning.Name].PublicKey...),
		SignerId:            defaultDID[:],
		PublicKey:           accKeys[identity.KeyPurposeSigning.Name].PublicKey,
		Signature:           s,
		TransitionValidated: false,
	}
	g.AppendSignatures(sig)
	_, err = g.CalculateDocumentRoot()
	assert.NoError(t, err)
	cd, err := g.PackCoreDocument()
	assert.NoError(t, err)
	return g, cd
}

func updateDocumentForP2Phandler(t *testing.T, g *generic.Generic) (*generic.Generic, coredocumentpb.CoreDocument) {
	cd, err := g.CoreDocument.PrepareNewVersion(nil, documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)
	g.CoreDocument = cd
	return prepareDocumentForP2PHandler(t, g)
}

func resolveSignatureResponse(t *testing.T, p2pEnv *protocolpb.P2PEnvelope) *p2ppb.SignatureResponse {
	signResp := new(p2ppb.SignatureResponse)
	dataEnv, err := p2pcommon.ResolveDataEnvelope(p2pEnv)
	assert.NoError(t, err)
	err = proto.Unmarshal(dataEnv.Body, signResp)
	assert.NoError(t, err)
	return signResp
}
