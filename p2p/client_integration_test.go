// +build integration

package p2p_test

import (
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var (
	client    documents.Client
	cfg       config.Configuration
	idService identity.Service
	cfgStore  config.Service
)

func TestMain(m *testing.M) {
	flag.Parse()
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgStore = ctx[config.BootstrappedConfigStorage].(config.Service)
	idService = ctx[identity.BootstrappedIDService].(identity.Service)
	client = ctx[bootstrap.BootstrappedPeer].(documents.Client)
	testingidentity.CreateIdentityWithKeys(cfg, idService)
	result := m.Run()
	testingbootstrap.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestClient_GetSignaturesForDocument(t *testing.T) {
	tc, _, err := createLocalCollaborator(t, false)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	signs, err := client.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	assert.NotNil(t, signs)
}

func TestClient_GetSignaturesForDocumentValidationCheck(t *testing.T) {
	tc, _, err := createLocalCollaborator(t, true)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	signs, err := client.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	// one signature would be missing
	assert.Equal(t, 0, len(signs))
}

func TestClient_SendAnchoredDocument(t *testing.T) {
	tc, cid, err := createLocalCollaborator(t, false)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	dm := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	cd, err := dm.PackCoreDocument()
	assert.NoError(t, err)
	_, err = client.SendAnchoredDocument(ctxh, cid.CentID(), &p2ppb.AnchorDocumentRequest{Document: &cd})
	if assert.Error(t, err) {
		assert.Equal(t, "[1]document is invalid: [mismatched document roots]", err.Error())
	}
}

func createLocalCollaborator(t *testing.T, corruptID bool) (*configstore.Account, identity.Identity, error) {
	tcID := identity.RandomCentID()
	tc, err := configstore.TempAccount("", cfg)
	assert.NoError(t, err)
	tcr := tc.(*configstore.Account)
	tcr.IdentityID = tcID[:]
	id := testingidentity.CreateAccountIDWithKeys(cfg.GetEthereumContextWaitTimeout(), tcr, idService)
	if corruptID {
		tcr.IdentityID = utils.RandomSlice(identity.CentIDLength)
	}
	tc, err = cfgStore.CreateAccount(tcr)
	assert.NoError(t, err)
	return tcr, id, err
}

func prepareDocumentForP2PHandler(t *testing.T, collaborators [][]byte) documents.Model {
	idConfig, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	payalod := testingdocuments.CreatePOPayload()
	var cs []string
	for _, c := range collaborators {
		cs = append(cs, hexutil.Encode(c))
	}
	payalod.Collaborators = cs
	po := new(purchaseorder.PurchaseOrder)
	err = po.InitPurchaseOrderInput(payalod, idConfig.ID.String())
	assert.NoError(t, err)
	_, err = po.DataRoot()
	assert.NoError(t, err)
	sr, err := po.SigningRoot()
	assert.NoError(t, err)
	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, sr)
	po.AppendSignatures(sig)
	_, err = po.DocumentRoot()
	assert.NoError(t, err)
	return po
}
