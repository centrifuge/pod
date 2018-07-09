package notification

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
	logging "github.com/ipfs/go-log"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/notification"
	"github.com/golang/protobuf/jsonpb"
)

var log = logging.Logger("notification-api")

type NotificationEventType int
const (
	RECEIVED_PAYLOAD NotificationEventType = 1
)

type NotificationStatus int
const (
	FAILURE NotificationStatus = 0
	SUCCESS NotificationStatus = 1
)

type Sender interface {
	Send(notification *notificationpb.NotificationMessage) (NotificationStatus, error)
}

type WebhookSender struct {}

func (wh *WebhookSender) Send(notification *notificationpb.NotificationMessage) (NotificationStatus, error) {
	url := config.Config.GetReceiveEventNotificationEndpoint()
	if url == "" {
		log.Warningf("Webhook URL not defined, manually fetch received document")
		return SUCCESS, nil
	}

	payload, err := wh.constructPayload(notification)
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

func (wh *WebhookSender) constructPayload(notification *notificationpb.NotificationMessage) ([]byte, error) {
	marshaler := jsonpb.Marshaler{}
	payload, err := marshaler.MarshalToString(notification)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return []byte(payload), nil
}
