// +build integration

package p2p_test

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/crypto"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var (
	client     documents.Client
	cfg        config.Configuration
	idService  identity.ServiceDID
	idFactory  identity.Factory
	cfgStore   config.Service
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
	dm := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	signs, err := client.GetSignaturesForDocument(ctxh, dm)
	assert.NoError(t, err)
	assert.NotNil(t, signs)
}

func TestClient_GetSignaturesForDocumentValidationCheck(t *testing.T) {
	tc, _, err := createLocalCollaborator(t, true)
	acc, err := configstore.NewAccount("", cfg)
	assert.Nil(t, err)
	acci := acc.(*configstore.Account)
	acci.IdentityID = defaultDID[:]
	ctxh, err := contextutil.New(context.Background(), acci)
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
	_, err = client.SendAnchoredDocument(ctxh, cid, &p2ppb.AnchorDocumentRequest{Document: &cd})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatched document roots")
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

func prepareDocumentForP2PHandler(t *testing.T, collaborators [][]byte) documents.Model {
	idConfig, err := identity.GetIdentityConfig(cfg)
	idConfig.ID = defaultDID
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
	_, err = po.CalculateDataRoot()
	assert.NoError(t, err)
	sr, err := po.CalculateSigningRoot()
	assert.NoError(t, err)
	s, err := crypto.SignMessage(idConfig.Keys[identity.KeyPurposeSigning].PrivateKey, sr, crypto.CurveSecp256K1)
	assert.NoError(t, err)
	sig := &coredocumentpb.Signature{
		EntityId:  idConfig.ID[:],
		PublicKey: idConfig.Keys[identity.KeyPurposeSigning].PublicKey,
		Signature: s,
		Timestamp: utils.ToTimestamp(time.Now().UTC()),
	}
	po.AppendSignatures(sig)
	_, err = po.CalculateDocumentRoot()
	assert.NoError(t, err)
	return po
}
