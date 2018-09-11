// build +unit
package p2phandler

import (
	"context"
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/notification"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/code"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/stretchr/testify/assert"
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
	rpc := Handler{&MockWebhookSender{}}

	res, err := rpc.Post(context.Background(), &req)
	assert.Nil(t, err, "Received error")
	assert.Equal(t, res.Document.DocumentIdentifier, coreDoc.DocumentIdentifier, "Incorrect identifier")

	doc := new(coredocumentpb.CoreDocument)
	err = coredocumentrepository.GetRepository().GetByID(coreDoc.DocumentIdentifier, doc)
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
	req := &p2ppb.SignatureRequest{Header: &p2ppb.CentrifugeHeader{
		CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID(),
	}, Document: testingutils.GenerateCoreDocument()}

	handler := Handler{Notifier: &MockWebhookSender{}}
	wantSig := []byte{0x0, 0x14, 0x36, 0x51, 0xa6, 0xe6, 0x2c, 0xb5, 0xe5, 0x16, 0x8a, 0x7a, 0x18, 0xd8, 0x87, 0xe0, 0xb3, 0x9e, 0xca, 0x9b, 0x2c, 0xa3, 0xeb, 0xd7, 0xbc, 0x86, 0xf2, 0xad, 0xc3, 0x97, 0x11, 0x7f, 0x1e, 0x89, 0x8b, 0x8a, 0xc7, 0xce, 0x4f, 0x71, 0xd5, 0x75, 0xd3, 0xf, 0xe7, 0xae, 0x39, 0x48, 0x16, 0x2f, 0x9d, 0xe5, 0x33, 0x81, 0xef, 0xff, 0xa2, 0x17, 0xc9, 0x34, 0x24, 0x7b, 0x93, 0x8}
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
