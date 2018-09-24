// +build unit

package notification

import (
	"os"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

func TestWebhookConstructPayload(t *testing.T) {
	coredoc := &coredocumentpb.CoreDocument{DocumentIdentifier: []byte("1")}
	cid := testingutils.Rand32Bytes()

	ts, err := ptypes.TimestampProto(time.Now().UTC())
	assert.Nil(t, err, "Should not error out")

	notificationMessage := &notificationpb.NotificationMessage{Document: coredoc, CentrifugeId: cid, EventType: uint32(RECEIVED_PAYLOAD), Recorded: ts}

	whs := WebhookSender{}
	bresult, err := whs.constructPayload(notificationMessage)
	assert.Nil(t, err, "Should not error out")

	unmarshaledNotificationMessage := &notificationpb.NotificationMessage{}

	jsonpb.UnmarshalString(string(bresult), unmarshaledNotificationMessage)

	assert.Equal(t, notificationMessage.Recorded, unmarshaledNotificationMessage.Recorded, "Recorder Timestamp should be equal")
	assert.Equal(t, notificationMessage.Document, unmarshaledNotificationMessage.Document, "CoreDocument should be equal")
	assert.Equal(t, notificationMessage.CentrifugeId, unmarshaledNotificationMessage.CentrifugeId, "CentrifugeID should be equal")
	assert.Equal(t, notificationMessage.EventType, unmarshaledNotificationMessage.EventType, "EventType should be equal")
}
