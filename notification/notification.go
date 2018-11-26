package notification

import (
	"fmt"
	"net/http"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/golang/protobuf/jsonpb"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("notification-api")

// EventType is the type of the notification.
type EventType int

// Status defines the status of the notification.
type Status int

// Constants defined for notification delivery.
const (
	ReceivedPayload EventType = 1
	Failure         Status    = 0
	Success         Status    = 1
)

// Config defines methods required for this package.
type Config interface {
	GetReceiveEventNotificationEndpoint() string
}

// Sender defines methods that can handle a notification.
type Sender interface {
	Send(notification *notificationpb.NotificationMessage) (Status, error)
}

// NewWebhookSender returns an implementation of a Sender that sends notifications through webhooks.
func NewWebhookSender(config Config) Sender {
	return webhookSender{config}
}

// NewWebhookSender implements Sender.
// Sends notification through a webhook defined.
type webhookSender struct {
	config Config
}

// Send sends notification to the defined webhook.
func (wh webhookSender) Send(notification *notificationpb.NotificationMessage) (Status, error) {
	url := wh.config.GetReceiveEventNotificationEndpoint()
	if url == "" {
		log.Warningf("Webhook URL not defined, manually fetch received document")
		return Success, nil
	}

	payload, err := wh.constructPayload(notification)
	if err != nil {
		return Failure, err
	}

	statusCode, err := utils.SendPOSTRequest(url, "application/json", payload)
	if err != nil {
		return Failure, err
	}

	if statusCode != http.StatusOK {
		return Failure, fmt.Errorf("failed to send webhook: status = %v", statusCode)
	}

	log.Infof("Sent Webhook Notification with Payload [%v] to [%s]", notification, url)

	return Success, nil
}

func (wh webhookSender) constructPayload(notification *notificationpb.NotificationMessage) ([]byte, error) {
	marshaler := jsonpb.Marshaler{}
	payload, err := marshaler.MarshalToString(notification)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return []byte(payload), nil
}
