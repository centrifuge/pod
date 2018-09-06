// build +unit
package grpcservice

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
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/stretchr/testify/assert"
)

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
	rpc := P2PService{&MockWebhookSender{}}

	res, err := rpc.HandleP2PPost(context.Background(), &req)
	assert.Nil(t, err, "Received error")
	assert.Equal(t, res.Document.DocumentIdentifier, coreDoc.DocumentIdentifier, "Incorrect identifier")

	doc := new(coredocumentpb.CoreDocument)
	err = coredocumentrepository.GetRepository().GetByID(coreDoc.DocumentIdentifier, doc)
	assert.Equal(t, doc.DocumentIdentifier, coreDoc.DocumentIdentifier, "Document Identifier doesn't match")
}

func TestP2PService_IncompatibleRequest(t *testing.T) {
	// Test invalid version
	req := p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: "1000.0.0-invalid", NetworkIdentifier: config.Config.GetNetworkID()}
	rpc := P2PService{&MockWebhookSender{}}
	res, err := rpc.HandleP2PPost(context.Background(), &req)

	assert.Error(t, err)
	p2perr, _ := errors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.VersionMismatch)
	assert.Nil(t, res)

	// Test invalid network
	req = p2ppb.P2PMessage{Document: coreDoc, CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID() + 1}
	res, err = rpc.HandleP2PPost(context.Background(), &req)

	assert.Error(t, err)
	p2perr, _ = errors.FromError(err)
	assert.Equal(t, p2perr.Code(), code.NetworkMismatch)
	assert.Nil(t, res)
}

func TestP2PService_HandleP2PPostNilDocument(t *testing.T) {
	req := p2ppb.P2PMessage{CentNodeVersion: version.GetVersion().String(), NetworkIdentifier: config.Config.GetNetworkID()}
	rpc := P2PService{&MockWebhookSender{}}
	res, err := rpc.HandleP2PPost(context.Background(), &req)

	assert.Error(t, err)
	assert.Nil(t, res)
}

// Webhook Notification Mocks //
type MockWebhookSender struct{}

func (wh *MockWebhookSender) Send(notification *notificationpb.NotificationMessage) (status notification.NotificationStatus, err error) {
	return
}
