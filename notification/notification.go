package notification

import (
	"context"
	"encoding/json"
	"time"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
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
	JobCompleted    EventType = 2
	Failure         Status    = 0
	Success         Status    = 1
)

// Message is the payload used to send the notifications.
type Message struct {
	EventType    EventType `json:"event_type"`
	Recorded     time.Time `json:"recorded" swaggertype:"primitive,string"`
	DocumentType string    `json:"document_type"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	DocumentID   string    `json:"document_id"`
	AccountID    string    `json:"account_id"` // account_id is the account associated to webhook
	FromID       string    `json:"from_id"`    // from_id if provided, original trigger of the event
	ToID         string    `json:"to_id"`      // to_id if provided, final destination of the event
}

// Sender defines methods that can handle a notification.
type Sender interface {
	Send(ctx context.Context, notification Message) (Status, error)
}

// NewWebhookSender returns an implementation of a Sender that sends notifications through webhooks.
func NewWebhookSender() Sender {
	return webhookSender{}
}

// NewWebhookSender implements Sender.
// Sends notification through a webhook defined.
type webhookSender struct{}

// Send sends notification to the defined webhook.
func (wh webhookSender) Send(ctx context.Context, notification Message) (Status, error) {
	tc, err := contextutil.Account(ctx)
	if err != nil {
		return Failure, err
	}
	url := tc.GetReceiveEventNotificationEndpoint()
	if url == "" {
		log.Warningf("Webhook URL not defined, manually fetch received document")
		return Success, nil
	}

	payload, err := json.Marshal(notification)
	if err != nil {
		return Failure, err
	}

	statusCode, err := utils.SendPOSTRequest(url, "application/json", payload)
	if err != nil {
		return Failure, err
	}

	if !utils.InRange(statusCode, 200, 299) {
		return Failure, errors.New("failed to send webhook: status = %v", statusCode)
	}

	log.Infof("Sent Webhook Notification with Payload [%v] to [%s]", notification, url)

	return Success, nil
}
