// +build integration

package p2p_test

import (
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/ethid"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/version"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

var (
	client    p2p.Client
	cfg       config.Configuration
	idService identity.Service
	cfgStore  config.Service
)

func TestMain(m *testing.M) {
	flag.Parse()
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgStore = ctx[configstore.BootstrappedConfigStorage].(config.Service)
	idService = ctx[ethid.BootstrappedIDService].(identity.Service)
	client = ctx[bootstrap.BootstrappedP2PClient].(p2p.Client)
	testingidentity.CreateIdentityWithKeys(cfg, idService)
	result := m.Run()
	testingbootstrap.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestClient_GetSignaturesForDocument(t *testing.T) {
	tc, _, err := createLocalCollaborator(t)
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	err = client.GetSignaturesForDocument(ctxh, idService, doc)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(doc.Signatures))
}

func TestClient_SendAnchoredDocument(t *testing.T) {
	tc, cid, err := createLocalCollaborator(t)
	ctxh := testingconfig.CreateTenantContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	self, err := cfg.GetIdentityID()
	assert.NoError(t, err)
	p2pheader := &p2ppb.CentrifugeHeader{
		SenderCentrifugeId: self,
		CentNodeVersion:    version.GetVersion().String(),
		NetworkIdentifier:  cfg.GetNetworkID(),
	}
	_, err = client.SendAnchoredDocument(ctxh, cid, &p2ppb.AnchorDocumentRequest{Document: doc, Header: p2pheader})
	if assert.Error(t, err) {
		assert.Equal(t, "[1]document is invalid: [mismatched document roots]", err.Error())
	}
}

func createLocalCollaborator(t *testing.T) (*configstore.TenantConfig, identity.Identity, error) {
	tcID := identity.RandomCentID()
	tc, err := configstore.TempTenantConfig("", cfg)
	assert.NoError(t, err)
	tcr := tc.(*configstore.TenantConfig)
	tcr.IdentityID = tcID[:]
	tc, err = cfgStore.CreateTenant(tcr)
	assert.NoError(t, err)
	id := testingidentity.CreateTenantIDWithKeys(cfg.GetEthereumContextWaitTimeout(), tcr, idService)
	return tcr, id, err
}

func prepareDocumentForP2PHandler(t *testing.T, collaborators [][]byte) *coredocumentpb.CoreDocument {
	idConfig, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	identifier := utils.RandomSlice(32)
	salts := &coredocumentpb.CoreDocumentSalts{}
	doc := &coredocumentpb.CoreDocument{
		Collaborators:      collaborators,
		DataRoot:           utils.RandomSlice(32),
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        utils.RandomSlice(32),
		CoredocumentSalts:  salts,
		EmbeddedData: &any.Any{
			TypeUrl: documenttypes.InvoiceDataTypeUrl,
		},
		EmbeddedDataSalts: &any.Any{
			TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
		},
	}
	err = proofs.FillSalts(doc, salts)
	assert.Nil(t, err)
	tree, _ := coredocument.GetDocumentSigningTree(doc)
	doc.SigningRoot = tree.RootHash()
	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	tree, _ = coredocument.GetDocumentRootTree(doc)
	doc.DocumentRoot = tree.RootHash()
	return doc
}
