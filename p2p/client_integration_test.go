// +build integration

package p2p_test

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"

	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/ethereum/go-ethereum/common"

	"github.com/centrifuge/go-centrifuge/testingutils/documents"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var (
	client     documents.Client
	cfg        config.Configuration
	idService  identity.ServiceDID
	idFactory  identity.Factory
	cfgStore   config.Service
	docService documents.Service
	defaultDID identity.DID
)

func TestMain(m *testing.M) {
	flag.Parse()
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgStore = ctx[config.BootstrappedConfigStorage].(config.Service)
	idService = ctx[identity.BootstrappedDIDService].(identity.ServiceDID)
	idFactory = ctx[identity.BootstrappedDIDFactory].(identity.Factory)
	client = ctx[bootstrap.BootstrappedPeer].(documents.Client)
	docService = ctx[documents.BootstrappedDocumentService].(documents.Service)
	tc, _ := configstore.TempAccount("", cfg)
	didAddr, err := idFactory.CalculateIdentityAddress(context.Background())
	assert.NoError(&testing.T{}, err)
	acc := tc.(*configstore.Account)
	acc.IdentityID = didAddr.Bytes()
	did, err := testingidentity.CreateAccountIDWithKeys(cfg.GetEthereumContextWaitTimeout(), acc, idService, idFactory)
	assert.NoError(&testing.T{}, err)
	defaultDID = did
	result := m.Run()
	testingbootstrap.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestClient_GetSignaturesForDocument(t *testing.T) {
	tc, _, err := createLocalCollaborator(t, false)
	acc, err := configstore.NewAccount("", cfg)
	assert.Nil(t, err)
	acci := acc.(*configstore.Account)
	acci.IdentityID = defaultDID[:]
	ctxh, err := contextutil.New(context.Background(), acci)
	assert.Nil(t, err)
	dm := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID}, defaultDID)
	err = client.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(dm.Document.Signatures))
}

func TestClient_GetSignaturesForDocumentValidationCheck(t *testing.T) {
	tc, _, err := createLocalCollaborator(t, true)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID}, defaultDID)
	err = client.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	// one signature would be missing
	assert.Equal(t, 1, len(dm.Document.Signatures))
}

func TestClient_SendAnchoredDocument(t *testing.T) {
	tc, cid, err := createLocalCollaborator(t, false)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID}, defaultDID)

	_, err = client.SendAnchoredDocument(ctxh, cid, &p2ppb.AnchorDocumentRequest{Document: dm.Document})
	if assert.Error(t, err) {
		assert.Equal(t, "[1]document is invalid: [mismatched document roots]", err.Error())
	}
}

func createLocalCollaborator(t *testing.T, corruptID bool) (*configstore.Account, identity.DID, error) {
	didAddr, err := idFactory.CalculateIdentityAddress(context.Background())
	assert.NoError(t, err)
	did := identity.NewDID(*didAddr)
	tc, err := configstore.TempAccount("", cfg)
	assert.NoError(t, err)
	tcr := tc.(*configstore.Account)
	tcr.IdentityID = did[:]
	cdid, err := testingidentity.CreateAccountIDWithKeys(cfg.GetEthereumContextWaitTimeout(), tcr, idService, idFactory)
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

func prepareDocumentForP2PHandler(t *testing.T, collaborators [][]byte, localID identity.DID) *documents.CoreDocumentModel {
	idConfig, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	idConfig.ID = localID
	dm, err := testingdocuments.GenerateCoreDocumentModelWithCollaborators(collaborators)
	assert.NoError(t, err)
	m, err := docService.DeriveFromCoreDocumentModel(dm)
	assert.Nil(t, err)

	droot, err := m.CalculateDataRoot()
	assert.Nil(t, err)

	dm, err = m.PackCoreDocument()
	assert.NoError(t, err)

	tree, err := dm.GetDocumentSigningTree(droot)
	assert.NoError(t, err)
	dm.Document.SigningRoot = tree.RootHash()

	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, dm.Document.SigningRoot)
	assert.NoError(t, err)
	dm.Document.Signatures = []*coredocumentpb.Signature{sig}

	tree, err = dm.GetDocumentRootTree()
	assert.NoError(t, err)
	dm.Document.DocumentRoot = tree.RootHash()
	return dm
}
