// build +unit
package grpcservice

import (
	"context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/p2p"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/version"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/notification"
)

func TestMain(m *testing.M) {
	cc.TestUnitBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

var identifier = []byte("1")
var coredoc = &coredocumentpb.CoreDocument{DocumentIdentifier: identifier}

func TestP2PService(t *testing.T) {
	req := p2ppb.P2PMessage{Document: coredoc, CentNodeVersion: version.CentrifugeNodeVersion, NetworkIdentifier: config.Config.GetNetworkID()}
	rpc := P2PService{&MockWebhookSender{}}

	res, err := rpc.HandleP2PPost(context.Background(), &req)
	assert.Nil(t, err, "Received error")
	assert.Equal(t, res.Document.DocumentIdentifier, identifier, "Incorrect identifier")

	doc, err := coredocumentrepository.GetCoreDocumentRepository().FindById(identifier)
	assert.Equal(t, doc.DocumentIdentifier, identifier, "Document Identifier doesn't match")
}

func TestP2PService_IncompatibleRequest(t *testing.T) {
	// Test invalid version
	req := p2ppb.P2PMessage{Document: coredoc, CentNodeVersion: "1000.0.0-invalid", NetworkIdentifier: config.Config.GetNetworkID()}
	rpc := P2PService{&MockWebhookSender{}}
	res, err := rpc.HandleP2PPost(context.Background(), &req)

	assert.Error(t, err)
	assert.IsType(t, &IncompatibleVersionError{""}, err)
	assert.Nil(t, res)

	// Test invalid network
	req = p2ppb.P2PMessage{Document: coredoc, CentNodeVersion: version.CentrifugeNodeVersion, NetworkIdentifier: config.Config.GetNetworkID() + 1}
	res, err = rpc.HandleP2PPost(context.Background(), &req)

	assert.Error(t, err)
	assert.IsType(t, &IncompatibleNetworkError{0}, err)
	assert.Nil(t, res)
}

// Mocks //
type MockWebhookSender struct {}
func (wh *MockWebhookSender) Send(notification *notification.Notification) (status notification.NotificationStatus, err error) {return}
