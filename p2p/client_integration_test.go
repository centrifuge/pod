// +build integration

package p2p_test

import (
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/invoice"
	"github.com/golang/protobuf/proto"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
)

var (
	client     documents.Client
	cfg        config.Configuration
	idService  identity.Service
	cfgStore   config.Service
	docService documents.Service
)

func TestMain(m *testing.M) {
	flag.Parse()
	ctx := testingbootstrap.TestFunctionalEthereumBootstrap()
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfgStore = ctx[config.BootstrappedConfigStorage].(config.Service)
	idService = ctx[identity.BootstrappedIDService].(identity.Service)
	client = ctx[bootstrap.BootstrappedPeer].(documents.Client)
	docService = ctx[documents.BootstrappedDocumentService].(documents.Service)
	testingidentity.CreateIdentityWithKeys(cfg, idService)
	result := m.Run()
	testingbootstrap.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestClient_GetSignaturesForDocument(t *testing.T) {
	tc, _, err := createLocalCollaborator(t, false)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	err = client.GetSignaturesForDocument(ctxh, doc)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(doc.Signatures))
}

func TestClient_GetSignaturesForDocumentValidationCheck(t *testing.T) {
	tc, _, err := createLocalCollaborator(t, true)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})
	err = client.GetSignaturesForDocument(ctxh, doc)
	assert.NoError(t, err)
	// one signature would be missing
	assert.Equal(t, 1, len(doc.Signatures))
}

func TestClient_SendAnchoredDocument(t *testing.T) {
	tc, cid, err := createLocalCollaborator(t, false)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	doc := prepareDocumentForP2PHandler(t, [][]byte{tc.IdentityID})

	_, err = client.SendAnchoredDocument(ctxh, cid.CentID(), &p2ppb.AnchorDocumentRequest{Document: doc})
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

func prepareDocumentForP2PHandler(t *testing.T, collaborators [][]byte) *coredocumentpb.CoreDocument {
	idConfig, err := identity.GetIdentityConfig(cfg)
	assert.Nil(t, err)
	identifier := utils.RandomSlice(32)

	dataSalts := &invoicepb.InvoiceDataSalts{}
	invData := &invoicepb.InvoiceData{}
	err = proofs.FillSalts(invData, dataSalts)
	assert.Nil(t, err)

	serializedInv, err := proto.Marshal(invData)
	assert.Nil(t, err)
	serializedInvSalts, err := proto.Marshal(dataSalts)
	assert.Nil(t, err)

	salts := &coredocumentpb.CoreDocumentSalts{}
	doc := &coredocumentpb.CoreDocument{
		Collaborators:      collaborators,
		DocumentIdentifier: identifier,
		CurrentVersion:     identifier,
		NextVersion:        utils.RandomSlice(32),
		CoredocumentSalts:  salts,
		EmbeddedData: &any.Any{
			TypeUrl: documenttypes.InvoiceDataTypeUrl,
			Value:   serializedInv,
		},
		EmbeddedDataSalts: &any.Any{
			TypeUrl: documenttypes.InvoiceSaltsTypeUrl,
			Value:   serializedInvSalts,
		},
	}

	err = proofs.FillSalts(doc, salts)
	assert.Nil(t, err)

	m, err := docService.DeriveFromCoreDocument(doc)
	assert.Nil(t, err)

	droot, err := m.CalculateDataRoot()
	assert.Nil(t, err)

	tree, _ := coredocument.GetDocumentSigningTree(doc, droot)
	doc.SigningRoot = tree.RootHash()
	sig := identity.Sign(idConfig, identity.KeyPurposeSigning, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	tree, _ = coredocument.GetDocumentRootTree(doc)
	doc.DocumentRoot = tree.RootHash()
	return doc
}
