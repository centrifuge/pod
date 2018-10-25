package notification

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/jsonpb"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("notification-api")

type EventType int

type Status int

const (
	ReceivedPayload EventType = 1
	Failure         Status    = 0
	Success         Status    = 1
)

type Sender interface {
	Send(notification *notificationpb.NotificationMessage) (Status, error)
}

type WebhookSender struct{}

func (wh *WebhookSender) Send(notification *notificationpb.NotificationMessage) (Status, error) {
	url := config.Config.GetReceiveEventNotificationEndpoint()
	if url == "" {
		log.Warningf("Webhook URL not defined, manually fetch received document")
		return Success, nil
	}

	payload, err := wh.constructPayload(notification)
	if err != nil {
		log.Error(err)
		return Failure, err
	}

	httpStatusCode, err := utils.SendPOSTRequest(url, "application/json", payload)
	if httpStatusCode != 200 {
		return Failure, err
	}

	log.Infof("Sent Webhook Notification with Payload [%v] to [%s]", notification, url)

	return Success, nil
}

func (wh *WebhookSender) constructPayload(notification *notificationpb.NotificationMessage) ([]byte, error) {
	marshaler := jsonpb.Marshaler{}
	payload, err := marshaler.MarshalToString(notification)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return []byte(payload), nil
}
