//go:build integration
// +build integration

package p2p_test

import (
	"context"
	"flag"
	"os"
	"sync"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/bootstrap"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"

	"github.com/centrifuge/go-centrifuge/jobs"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

var (
	client     documents.Client
	cfg        config.Configuration
	idService  identity.Service
	idFactory  identity.Factory
	cfgStore   config.Service
	defaultDID identity.DID
	ethClient  ethereum.Client
	dispatcher jobs.Dispatcher
)

func TestMain(m *testing.M) {
	flag.Parse()
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgStore = ctx[config.BootstrappedConfigStorage].(config.Service)
	idService = ctx[identity.BootstrappedDIDService].(identity.Service)
	idFactory = ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	client = ctx[bootstrap.BootstrappedPeer].(documents.Client)
	ethClient = ctx[ethereum.BootstrappedEthereumClient].(ethereum.Client)
	dispatcher = ctx[jobs.BootstrappedDispatcher].(jobs.Dispatcher)
	ctxh, canc := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go dispatcher.Start(ctxh, wg, nil)
	tc, err := configstore.TempAccount("main", cfg)
	assert.NoError(&testing.T{}, err)
	didAddr, err := idFactory.NextIdentityAddress()
	assert.NoError(&testing.T{}, err)
	acc := tc.(*configstore.Account)
	acc.IdentityID = didAddr[:]
	did, err := testingidentity.CreateAccountIDWithKeys(
		cfg.GetEthereumContextWaitTimeout(), acc, idService, idFactory, ethClient, dispatcher)
	assert.NoError(&testing.T{}, err)
	defaultDID = did
	result := m.Run()
	canc()
	wg.Wait()
	testingbootstrap.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestClient_GetSignaturesForDocument(t *testing.T) {
	tc, _, err := createLocalCollaborator(t, false)
	assert.NoError(t, err)
	acc, err := configstore.NewAccount("main", cfg)
	assert.Nil(t, err)
	acci := acc.(*configstore.Account)
	acci.IdentityID = defaultDID[:]
	ctxh := contextutil.WithAccount(context.Background(), acci)
	dm := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	signs, _, err := client.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	assert.NotNil(t, signs)
}

func TestClient_GetSignaturesForDocumentValidationCheck(t *testing.T) {
	// Random DID cause signature verification failure
	tc, _, err := createLocalCollaborator(t, true)
	assert.NoError(t, err)
	acc, err := configstore.NewAccount("main", cfg)
	assert.Nil(t, err)
	acci := acc.(*configstore.Account)
	acci.IdentityID = defaultDID[:]
	ctxh := contextutil.WithAccount(context.Background(), acci)
	dm := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	signs, signatureErrors, err := client.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	assert.Error(t, signatureErrors[0], "[5]signature invalid with err: no contract code at given address")
	assert.Equal(t, 0, len(signs))
}

func TestClient_SendAnchoredDocument(t *testing.T) {
	tc, cid, err := createLocalCollaborator(t, false)
	assert.NoError(t, err)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	cd, err := dm.PackCoreDocument()
	assert.NoError(t, err)
	_, err = client.SendAnchoredDocument(ctxh, cid, &p2ppb.AnchorDocumentRequest{Document: &cd})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unable to find anchor")
}

func createLocalCollaborator(t *testing.T, corruptID bool) (*configstore.Account, identity.DID, error) {
	did, err := idFactory.NextIdentityAddress()
	assert.NoError(t, err)
	tc, err := configstore.TempAccount("main", cfg)
	assert.NoError(t, err)
	tcr := tc.(*configstore.Account)
	tcr.IdentityID = did[:]
	cdid, err := testingidentity.CreateAccountIDWithKeys(
		cfg.GetEthereumContextWaitTimeout(), tcr, idService, idFactory, ethClient, dispatcher)
	assert.NoError(t, err)
	if !cdid.Equal(did) {
		assert.True(t, false, "Race condition identified when creating accounts")
	}
	tcr.IdentityID = did[:]
	if corruptID {
		tcr.IdentityID = utils.RandomSlice(common.AddressLength)
	}
	tc, err = cfgStore.CreateAccount(tcr)
	assert.NoError(t, err)
	return tcr, did, err
}

func prepareDocumentForP2PHandler(t *testing.T, collaborators [][]byte) documents.Document {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	accCfg, err := contextutil.Account(ctx)
	assert.NoError(t, err)
	acc := accCfg.(*configstore.Account)
	acc.IdentityID = defaultDID[:]
	accKeys, err := acc.GetKeys()
	assert.NoError(t, err)
	payload := generic.CreateGenericPayload(t, nil)
	dids, err := identity.BytesToDIDs(collaborators...)
	assert.NoError(t, err)
	var cs []identity.DID
	for _, did := range dids {
		cs = append(cs, *did)
	}
	payload.Collaborators.ReadWriteCollaborators = cs
	g := generic.InitGeneric(t, defaultDID, payload)
	g.SetUsedAnchorRepoAddress(cfg.GetContractAddress(config.AnchorRepo))
	err = g.AddUpdateLog(defaultDID)
	assert.NoError(t, err)
	sr, err := g.CalculateSigningRoot()
	assert.NoError(t, err)
	s, err := crypto.SignMessage(accKeys[identity.KeyPurposeSigning.Name].PrivateKey, documents.ConsensusSignaturePayload(sr, true), crypto.CurveEd25519)
	assert.NoError(t, err)
	sig := &coredocumentpb.Signature{
		SignatureId:         append(defaultDID[:], accKeys[identity.KeyPurposeSigning.Name].PublicKey...),
		SignerId:            defaultDID[:],
		PublicKey:           accKeys[identity.KeyPurposeSigning.Name].PublicKey,
		Signature:           s,
		TransitionValidated: true,
	}
	g.AppendSignatures(sig)
	_, err = g.CalculateDocumentRoot()
	assert.NoError(t, err)
	return g
}
