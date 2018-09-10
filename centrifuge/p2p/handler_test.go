// +build unit

package p2p

import (
	"context"
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/notification"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/stretchr/testify/assert"
)

// MockWebhookSender implements notification.Sender
type MockWebhookSender struct{}

func (wh *MockWebhookSender) Send(notification *notificationpb.NotificationMessage) (status notification.Status, err error) {
	return
}

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	coredocument.InitLevelDBRepository(cc.GetLevelDBStorage())
	identity.SetIdentityService(identity.NewEthereumIdentityService())
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

var coreDoc = testingutils.GenerateCoreDocument()

func TestP2PService(t *testing.T) {
	req := p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID()}
	rpc := Handler{&MockWebhookSender{}}

	res, err := rpc.Post(context.Background(), &req)
	assert.Nil(t, err, "Received error")
	assert.Equal(t, res.Document.DocumentIdentifier, coreDoc.DocumentIdentifier, "Incorrect identifier")

	doc := new(coredocumentpb.CoreDocument)
	err = coredocument.GetRepository().GetByID(coreDoc.DocumentIdentifier, doc)
	assert.Equal(t, doc.DocumentIdentifier, coreDoc.DocumentIdentifier, "Document Identifier doesn't match")
}

func TestP2PService_IncompatibleRequest(t *testing.T) {
	// Test invalid version
	req := p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: "1000.0.0-invalid", NetworkIdentifier: config.Config.GetNetworkID()}
	rpc := Handler{&MockWebhookSender{}}
	res, err := rpc.Post(context.Background(), &req)

	assert.Error(t, err)
	p2perr, _ := errors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.VersionMismatch)
	assert.Nil(t, res)

	// Test invalid network
	req = p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID() + 1}
	res, err = rpc.Post(context.Background(), &req)

	assert.Error(t, err)
	p2perr, _ = errors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.NetworkMismatch)
	assert.Nil(t, res)
}

func TestP2PService_HandleP2PPostNilDocument(t *testing.T) {
	req := p2ppb.P2PMessage{CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID()}
	rpc := Handler{&MockWebhookSender{}}
	res, err := rpc.Post(context.Background(), &req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestHandler_RequestDocumentSignature_nilDocument(t *testing.T) {
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID(),
	}}

	handler := Handler{Notifier: &MockWebhookSender{}}
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.Error(t, err, "must return error")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_RequestDocumentSignature_version_fail(t *testing.T) {
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: "1000.0.1-invalid", NetworkIdentifier: config.Config.GetNetworkID(),
	}}

	handler := Handler{Notifier: &MockWebhookSender{}}
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.Error(t, err, "must return error")
	assert.Contains(t, err.Error(), "Incompatible version")
	assert.Nil(t, resp, "must be nil")
}

func TestHandler_RequestDocumentSignature(t *testing.T) {
	req := &p2ppb.SignatureRequest{
		Header: &p2ppb.CentrifugeHeader{
			CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID(),
		},
		Document: testingutils.GenerateCoreDocument()}

	config.Config.V.Set("keys.signing.publicKey", "../../example/resources/signingKey.pub.pem")
	config.Config.V.Set("keys.signing.privateKey", "../../example/resources/signingKey.key.pem")
	handler := Handler{Notifier: &MockWebhookSender{}}
	wantSig := []byte{0x52, 0x2, 0x7f, 0x51, 0xcc, 0xb, 0x36, 0x1b, 0x52, 0x71, 0xfa, 0xc0, 0x1b, 0x21, 0x34, 0xef, 0xae, 0x76, 0x72, 0x3f, 0xe0, 0x93, 0x5f, 0xe8, 0xc4, 0x15, 0xd0, 0xf3, 0x77, 0x78, 0x1a, 0x2f, 0xf8, 0xa6, 0x42, 0x8b, 0x9, 0xbe, 0x96, 0xa2, 0xb5, 0xc8, 0xce, 0xf2, 0xd9, 0x6a, 0x5f, 0x40, 0x61, 0x52, 0xb5, 0x1e, 0x93, 0x92, 0xf8, 0xc0, 0xad, 0xc2, 0x4e, 0x66, 0xa5, 0xd1, 0x93, 0xa}
	resp, err := handler.RequestDocumentSignature(context.Background(), req)
	assert.Nil(t, err, "must be nil")
	assert.Equal(t, resp.CentNodeVersion, version.GetVersion().String())
	assert.Equal(t, wantSig, resp.Signature.Signature)
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
