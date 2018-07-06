package notification

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"encoding/json"
	logging "github.com/ipfs/go-log"
	"github.com/CentrifugeInc/go-centrifuge/utils"
	"time"
)

var log = logging.Logger("notification-api")


type Notification struct {
	EventType int
	CentrifugeID string
	AccountID string
	Timestamp time.Time
	CoreDocument *coredocumentpb.CoreDocument
}

type Sender interface {
	Send(notification *Notification) (statusCode int, err error)
}

type WebhookSender struct {}

func (wh *WebhookSender) Send(notification *Notification) (statusCode int, err error) {
	url := config.Config.GetReceiveEventNotificationEndpoint()
	if url == "" {
		log.Warningf("Webhook URL not defined, manually fetch received document")
		return
	}
	payload, err := json.Marshal(notification)

	if err != nil {
		log.Error(err)
		return
	}

	statusCode, err = utils.SendPOSTRequest(url, "application/json", payload)

	return
}
