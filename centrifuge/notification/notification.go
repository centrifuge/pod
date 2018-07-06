package notification

import (
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	"encoding/json"
	logging "github.com/ipfs/go-log"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"time"
)

var log = logging.Logger("notification-api")

type NotificationStatus int
const (
	FAILURE NotificationStatus = 0
	SUCCESS NotificationStatus = 1
)

type Notification struct {
	EventType int
	CentrifugeID string
	Recorded time.Time
	CoreDocument *coredocumentpb.CoreDocument
}

type Sender interface {
	Send(notification *Notification) (NotificationStatus, error)
}

type WebhookSender struct {}

func (wh *WebhookSender) Send(notification *Notification) (NotificationStatus, error) {
	url := config.Config.GetReceiveEventNotificationEndpoint()
	if url == "" {
		log.Warningf("Webhook URL not defined, manually fetch received document")
		return SUCCESS, nil
	}
	payload, err := json.Marshal(notification)

	if err != nil {
		log.Error(err)
		return FAILURE, err
	}

	httpStatusCode, err := utils.SendPOSTRequest(url, "application/json", payload)
	if httpStatusCode != 200 {
		return FAILURE, err
	}

	return SUCCESS, nil
}
