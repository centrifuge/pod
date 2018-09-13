package notification

import (
	"fmt"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/notification"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"github.com/golang/protobuf/jsonpb"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("notification-api")

// EventType defines the type of the Notification Event
type EventType int

// Status defines the status of the of the Send action
type Status int

const (
	ReceivedPayload EventType = 1
	FAILURE         Status    = 0
	SUCCESS         Status    = 1
)

// Sender implementors sends the notification received as per the implementation logic
// Ex: Webhook
type Sender interface {
	Send(notification *notificationpb.NotificationMessage) (Status, error)
}

// WebhookSender implements Sender
// Makes a POST to defined URL with notification as a Payload
type WebhookSender struct {
	URL string
}

// Send makes a POST call to the URL defined with notification as the payload
func (wh *WebhookSender) Send(notification *notificationpb.NotificationMessage) (Status, error) {
	if wh.URL == "" {
		return FAILURE, fmt.Errorf("webhook URL not defined")
	}

	payload, err := wh.constructPayload(notification)
	if err != nil {
		log.Error(err)
		return FAILURE, err
	}

	httpStatusCode, err := utils.SendPOSTRequest(wh.URL, "application/json", payload)
	if httpStatusCode != 200 {
		return FAILURE, err
	}

	return SUCCESS, nil
}

// constructPayload converts a proto.Message to a JSON string
func (wh *WebhookSender) constructPayload(notification *notificationpb.NotificationMessage) ([]byte, error) {
	marshaler := jsonpb.Marshaler{}
	payload, err := marshaler.MarshalToString(notification)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return []byte(payload), nil
}
