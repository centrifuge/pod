package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("notification-api")

// EventType is the type of the notification.
type EventType string

// Constants defined for notification delivery.
const (
	EventTypeJob      EventType = "job"
	EventTypeDocument EventType = "document"
)

type JobMessage struct {
	ID         byteutils.HexBytes `json:"id" swaggertype:"primitive,string"`    // job identifier
	Owner      byteutils.HexBytes `json:"owner" swaggertype:"primitive,string"` // job owner
	Desc       string             `json:"desc"`                                 // description of the job
	ValidUntil time.Time          `json:"valid_until"`                          // validity of the job
	FinishedAt time.Time          `json:"finished_at"`                          // job finished at
}

type DocumentMessage struct {
	ID        byteutils.HexBytes `json:"id" swaggertype:"primitive,string"`         // document identifier
	VersionID byteutils.HexBytes `json:"version_id" swaggertype:"primitive,string"` // version identifier
	From      byteutils.HexBytes `json:"from" swaggertype:"primitive,string"`       // document received from
	To        byteutils.HexBytes `json:"to" swaggertype:"primitive,string"`         // document sent to
}

// Message is the payload used to send the notifications.
type Message struct {
	EventType  EventType `json:"event_type" enums:"job,document"`
	RecordedAt time.Time `json:"recorded_at" swaggertype:"primitive,string"`

	// Job contains jobs specific details. Ensure event type is job
	Job *JobMessage `json:"job,omitempty"`

	// Document contains recently received document. Ensure event type is document
	Document *DocumentMessage `json:"document,omitempty"`
}

//go:generate mockery --name Sender --structname SenderMock --filename sender_mock.go --inpackage

// Sender defines methods that can handle a notification.
type Sender interface {
	Send(ctx context.Context, message Message) error
}

// NewWebhookSender returns an implementation of a Sender that sends notifications through webhooks.
func NewWebhookSender() Sender {
	return webhookSender{}
}

// NewWebhookSender implements Sender.
// Sends notification through a webhook defined.
type webhookSender struct{}

// Send sends notification to the defined webhook.
func (wh webhookSender) Send(ctx context.Context, message Message) error {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return err
	}

	url := acc.GetWebhookURL()
	if url == "" {
		log.Warnf("Webhook URL not defined, manually fetch received document")
		return nil
	}

	log.Infof("Sending webhook message with Payload[%v] to [%s]", message, url)
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	statusCode, err := utils.SendPOSTRequest(url, "application/json", payload)
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	if !utils.InRange(statusCode, 200, 299) {
		return errors.New("failed to send webhook: status = %v", statusCode)
	}

	log.Infof("Sent Webhook message with Payload [%v] to [%s]", message, url)
	return nil
}
