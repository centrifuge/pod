// +build integration

package p2phandler_test

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/config"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	cented25519 "github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519keys"
	"github.com/centrifuge/go-centrifuge/centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/centrifuge/p2p/p2phandler"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/centrifuge/version"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

var (
	key1Pub = [...]byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1    = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	handler = p2phandler.Handler{Notifier: &notification.WebhookSender{}}
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	coredocumentrepository.InitLevelDBRepository(cc.GetLevelDBStorage())
	bContext := map[string]interface{}{}
	bContext[bootstrap.BootstrappedLevelDb] = true
	(&invoice.Bootstrapper{}).Bootstrap(bContext)
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

func TestHandler_RequestDocumentSignature_verification_fail(t *testing.T) {
	req := getSignatureRequest()
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.NotNil(t, err, "must be non nil")
	assert.Nil(t, resp, "must be nil")
	assert.Contains(t, err.Error(), "signing root missing")
}

func TestHandler_RequestDocumentSignature(t *testing.T) {
	requestDocumentSignature(t)
}

func TestHandler_SendAnchoredDocument_update_fail(t *testing.T) {
	req := getAnchoredRequest()
	//Document doesn't exist yet
	resp, err := handler.SendAnchoredDocument(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document doesn't exists")
	assert.Nil(t, resp)
}

func TestHandler_SendAnchoredDocument(t *testing.T) {
	doc := requestDocumentSignature(t)
	req := getAnchoredRequest()
	req.Document = doc
	resp, err := handler.SendAnchoredDocument(context.Background(), req)
	assert.Nil(t, err)
	assert.NotNil(t, resp, "must be non nil")
}

func requestDocumentSignature(t *testing.T) *coredocumentpb.CoreDocument {
	idConfig, err := cented25519.GetIDConfig()
	assert.Nil(t, err)
	sig := &coredocumentpb.Signature{
		EntityId:  idConfig.ID,
		PublicKey: key1Pub[:],
	}
	centID, _ := identity.ToCentID(sig.EntityId)
	idkey := &identity.EthereumIdentityKey{
		Key:       key1Pub,
		Purposes:  []*big.Int{big.NewInt(identity.KeyPurposeSigning)},
		RevokedAt: big.NewInt(0),
	}
	id := &testingcommons.MockID{}
	srv := &testingcommons.MockIDService{}
	srv.On("LookupIdentityForID", centID).Return(id, nil).Once()
	id.On("FetchKey", key1Pub[:]).Return(idkey, nil).Once()
	identity.IDService = srv
	doc := testingutils.GenerateCoreDocument()
	tree, _ := coredocument.GetDocumentSigningTree(doc)
	doc.SigningRoot = tree.RootHash()
	sig = signatures.Sign(&config.IdentityConfig{
		ID:         sig.EntityId,
		PublicKey:  key1Pub[:],
		PrivateKey: key1,
	}, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	req := getSignatureRequest()
	req.Document = doc
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	srv.AssertExpectations(t)
	id.AssertExpectations(t)
	assert.Nil(t, err, "must be nil")
	assert.NotNil(t, resp, "must be non nil")
	assert.NotNil(t, resp.Signature.Signature, "must be non nil")
	sig = resp.Signature
	assert.True(t, ed25519.Verify(sig.PublicKey, doc.SigningRoot, sig.Signature), "signature must be valid")

	return doc
}

func getAnchoredRequest() *p2ppb.AnchDocumentRequest {
	return &p2ppb.AnchDocumentRequest{
		Header: &p2ppb.CentrifugeHeader{
			CentNodeVersion:   version.GetVersion().String(),
			NetworkIdentifier: config.Config.GetNetworkID(),
		},
		Document: testingutils.GenerateCoreDocument(),
	}
}

func getSignatureRequest() *p2ppb.SignatureRequest {
	return &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID(),
	}, Document: testingutils.GenerateCoreDocument()}
}
