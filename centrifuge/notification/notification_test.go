// build +unit

package notification

import (
	"testing"
	"os"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"fmt"
)

var port int
func TestMain(m *testing.M) {
	cc.TestUnitBootstrap()
	p, srv := testingutils.StartWebhookServer()
	port = p
	result := m.Run()
	testingutils.StopWebhookServer(srv)
	cc.TestTearDown()
	os.Exit(result)
}

func TestWebhookSender_Send_200OK(t *testing.T) {
	config.Config.V.Set("notifications.endpoint", fmt.Sprintf("http://localhost:%d/webhook", port))
	var identifier = []byte("1")
	var coredoc = &coredocumentpb.CoreDocument{DocumentIdentifier: identifier}
	notifier := &WebhookSender{}

	statusCode, err := notifier.Send(&Notification{0, "cId", "aId", time.Now().UTC(), coredoc})

	assert.Nil(t, err, "Received error")
	assert.Equal(t, 200, statusCode)
}

func TestWebhookSender_Send_404(t *testing.T) {
	config.Config.V.Set("notifications.endpoint", fmt.Sprintf("http://localhost:%d/webhookish", port))
	var identifier = []byte("1")
	var coredoc = &coredocumentpb.CoreDocument{DocumentIdentifier: identifier}
	notifier := &WebhookSender{}

	statusCode, err := notifier.Send(&Notification{0, "cId", "aId", time.Now().UTC(), coredoc})

	assert.NotNil(t, err, "Should have Received error")
	assert.Equal(t, 404, statusCode)
}

func TestWebhookSender_Send_BADHOST(t *testing.T) {
	config.Config.V.Set("notifications.endpoint", fmt.Sprintf("http://localish:%d/webhook", port))
	var identifier = []byte("1")
	var coredoc = &coredocumentpb.CoreDocument{DocumentIdentifier: identifier}
	notifier := &WebhookSender{}

	_, err := notifier.Send(&Notification{0, "cId", "aId", time.Now().UTC(), coredoc})

	assert.NotNil(t, err, "Should have Received error")
	assert.Contains(t, err.Error(), "no such host")
}
