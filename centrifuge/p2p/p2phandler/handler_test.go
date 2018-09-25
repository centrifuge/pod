// +build unit

package p2phandler

import (
	"context"
	"math/big"
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/notification"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/centerrors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	cented25519 "github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519keys"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/signatures"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils/commons"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ed25519"
)

var (
	key1Pub = [...]byte{230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	key1    = []byte{102, 109, 71, 239, 130, 229, 128, 189, 37, 96, 223, 5, 189, 91, 210, 47, 89, 4, 165, 6, 188, 53, 49, 250, 109, 151, 234, 139, 57, 205, 231, 253, 230, 49, 10, 12, 200, 149, 43, 184, 145, 87, 163, 252, 114, 31, 91, 163, 24, 237, 36, 51, 165, 8, 34, 104, 97, 49, 114, 85, 255, 15, 195, 199}
	handler = Handler{Notifier: &MockWebhookSender{}}
)

// MockWebhookSender implements notification.Sender
type MockWebhookSender struct{}

func (wh *MockWebhookSender) Send(notification *notificationpb.NotificationMessage) (status notification.NotificationStatus, err error) {
	return
}

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	coredocumentrepository.InitLevelDBRepository(cc.GetLevelDBStorage())
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

var coreDoc = testingutils.GenerateCoreDocument()

func TestP2PService(t *testing.T) {
	req := p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID()}
	res, err := handler.Post(context.Background(), &req)
	assert.Nil(t, err, "Received error")
	assert.Equal(t, res.Document.DocumentIdentifier, coreDoc.DocumentIdentifier, "Incorrect identifier")

	doc := new(coredocumentpb.CoreDocument)
	err = coredocumentrepository.GetRepository().GetByID(coreDoc.DocumentIdentifier, doc)
	assert.Equal(t, doc.DocumentIdentifier, coreDoc.DocumentIdentifier, "Document Identifier doesn't match")
}

func TestP2PService_IncompatibleRequest(t *testing.T) {
	// Test invalid version
	req := p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: "1000.0.0-invalid", NetworkIdentifier: config.Config.GetNetworkID()}
	res, err := handler.Post(context.Background(), &req)

	assert.Error(t, err)
	p2perr, _ := centerrors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.VersionMismatch)
	assert.Nil(t, res)

	// Test invalid network
	req = p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID() + 1}
	res, err = handler.Post(context.Background(), &req)

	assert.Error(t, err)
	p2perr, _ = centerrors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.NetworkMismatch)
	assert.Nil(t, res)
}

func TestP2PService_HandleP2PPostNilDocument(t *testing.T) {
	req := p2ppb.P2PMessage{CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID()}
	res, err := handler.Post(context.Background(), &req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestHandler_RequestDocumentSignature_nilDocument(t *testing.T) {
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID(),
	}}

	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.Error(t, err, "must return error")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_RequestDocumentSignature_version_fail(t *testing.T) {
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: "1000.0.1-invalid", NetworkIdentifier: config.Config.GetNetworkID(),
	}}

	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "Incompatible version")
	assert.Nil(t, resp, "must be nil")
}

func getSignatureRequest() *p2ppb.SignatureRequest {
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID(),
	}, Document: testingutils.GenerateCoreDocument()}

	return req
}

func TestHandler_RequestDocumentSignature_verification_fail(t *testing.T) {
	req := getSignatureRequest()
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.NotNil(t, err, "must be non nil")
	assert.Nil(t, resp, "must be nil")
	assert.Contains(t, err.Error(), "signing_root is missing")
}

func TestHandler_RequestDocumentSignature(t *testing.T) {
	idConfig, _ := cented25519.GetIDConfig()
	sig := &coredocumentpb.Signature{
		EntityId:  idConfig.ID,
		PublicKey: key1Pub[:],
	}
	centID, _ := identity.NewCentID(sig.EntityId)
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
}

func TestSendAnchoredDocument(t *testing.T) {
	header := &p2ppb.CentrifugeHeader{
		CentNodeVersion:   version.GetVersion().String(),
		NetworkIdentifier: config.Config.GetNetworkID(),
	}
	req := p2ppb.AnchDocumentRequest{Document: coreDoc, Header: header}
	res, err := handler.SendAnchoredDocument(context.Background(), &req)
	assert.Nil(t, err, "Received error")
	assert.True(t, res.Accepted, "Document not accepted")

	doc := new(coredocumentpb.CoreDocument)
	err = coredocumentrepository.GetRepository().GetByID(coreDoc.DocumentIdentifier, doc)
	assert.Equal(t, doc.DocumentIdentifier, coreDoc.DocumentIdentifier, "Document Identifier doesn't match")
}

func TestSendAnchoredDocument_IncompatibleRequest(t *testing.T) {
	// Test invalid version
	header := &p2ppb.CentrifugeHeader{
		CentNodeVersion:   "1000.0.0-invalid",
		NetworkIdentifier: config.Config.GetNetworkID(),
	}
	req := p2ppb.AnchDocumentRequest{Document: coreDoc, Header: header}
	res, err := handler.SendAnchoredDocument(context.Background(), &req)
	assert.Error(t, err)
	p2perr, _ := centerrors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.VersionMismatch)
	assert.Nil(t, res)

	// Test invalid network
	header.NetworkIdentifier = config.Config.GetNetworkID() + 1
	header.CentNodeVersion = version.GetVersion().String()
	res, err = handler.SendAnchoredDocument(context.Background(), &req)
	assert.Error(t, err)
	p2perr, _ = centerrors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.NetworkMismatch)
	assert.Nil(t, res)
}

func TestSendAnchoredDocument_NilDocument(t *testing.T) {
	header := &p2ppb.CentrifugeHeader{
		CentNodeVersion:   version.GetVersion().String(),
		NetworkIdentifier: config.Config.GetNetworkID(),
	}
	req := p2ppb.AnchDocumentRequest{Header: header}
	res, err := handler.SendAnchoredDocument(context.Background(), &req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestP2PService_basicChecks(t *testing.T) {
	tests := []struct {
		version   string
		networkID uint32
		err       error
	}{
		{
			version:   "someversion",
			networkID: 12,
			err:       version.IncompatibleVersionError("someversion"),
		},

		{
			version:   "0.0.1",
			networkID: 12,
			err:       incompatibleNetworkError(12),
		},

		{
			version:   version.GetVersion().String(),
			networkID: config.Config.GetNetworkID(),
		},
	}

	for _, c := range tests {
		err := basicChecks(c.version, c.networkID)
		if err != nil {
			if c.err == nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			assert.EqualError(t, err, c.err.Error(), "error mismatch")
		}
	}

}
