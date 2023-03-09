//go:build testworld

package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/notification"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("webhook-receiver")
)

type Receiver struct {
	port     int
	endpoint string

	docNotificationRequestChan chan documentNotificationRequest
	notificationChan           chan notification.Message
	jobWaitChan                chan jobWaitRequest

	s *http.Server
}

const (
	defaultChanSize = 100
)

func NewReceiver(port int, endpoint string) *Receiver {
	return &Receiver{
		port:                       port,
		endpoint:                   endpoint,
		docNotificationRequestChan: make(chan documentNotificationRequest, defaultChanSize),
		notificationChan:           make(chan notification.Message, defaultChanSize),
		jobWaitChan:                make(chan jobWaitRequest, defaultChanSize),
	}
}

func (w *Receiver) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer rw.WriteHeader(http.StatusOK)
	defer r.Body.Close()

	var msg notification.Message

	if err := decoder.Decode(&msg); err != nil {
		log.Errorf("Couldn't decode notification message")
		return
	}

	select {
	case <-r.Context().Done():
		log.Errorf("Request context done while attempting to store notification: %s", r.Context().Err())
	case w.notificationChan <- msg:
	}
}

const (
	docNotificationTimeout = 10 * time.Second
)

func (w *Receiver) GetReceivedDocumentMsg(to string, docID string) (msg notification.Message, err error) {
	resChan := make(chan documentNotificationResponse)

	req := documentNotificationRequest{
		to:         to,
		documentID: docID,
		resChan:    resChan,
	}

	select {
	case <-time.After(docNotificationTimeout):
		return msg, errors.New("timeout reached while sending doc notification request")
	case w.docNotificationRequestChan <- req:
	}

	select {
	case <-time.After(docNotificationTimeout):
		return msg, errors.New("timeout reached while waiting for doc notification response")
	case res := <-resChan:
		return res.msg, res.err
	}
}

// WaitForJobCompletion sends bool on channel when the job is complete
func (w *Receiver) WaitForJobCompletion(ctx context.Context, jobID string) error {
	waitChan := make(chan struct{})

	req := jobWaitRequest{
		jobID:    jobID,
		waitChan: waitChan,
	}

	select {
	case <-ctx.Done():
		log.Errorf("Context done while sending job %s: %s", jobID, ctx.Err())

		return ctx.Err()
	case w.jobWaitChan <- req:
	}

	select {
	case <-ctx.Done():
		log.Errorf("Context done while waiting for job %s: %s", jobID, ctx.Err())

		return ctx.Err()
	case <-waitChan:
	}

	return nil
}

func (w *Receiver) GetURL() string {
	return "http://localhost:" + strconv.Itoa(w.port) + w.endpoint
}

type jobWaitRequest struct {
	jobID    string
	waitChan chan struct{}
}

type documentNotificationRequest struct {
	to         string
	documentID string
	resChan    chan documentNotificationResponse
}

type documentNotificationResponse struct {
	msg notification.Message
	err error
}

func (w *Receiver) handleNotifications(ctx context.Context) {
	notificationMap := make(map[string]notification.Message)
	jobWaitMap := make(map[string][]chan struct{})

	for {
		select {
		case <-ctx.Done():
			log.Errorf("Context done while handling notifications: %s", ctx.Err())
			return
		case msg := <-w.notificationChan:
			var key string

			switch msg.EventType {
			case notification.EventTypeJob:
				jobID := strings.ToLower(msg.Job.ID.String())

				if channels, ok := jobWaitMap[jobID]; ok {
					for _, channel := range channels {
						close(channel)
					}
				}

				delete(jobWaitMap, jobID)

				key = jobID
			case notification.EventTypeDocument:
				key = strings.ToLower(msg.Document.ID.String())
			default:
				log.Warnf("Unsupported notification type %s", msg.EventType)
				continue
			}

			notificationMap[key] = msg
		case req := <-w.docNotificationRequestChan:
			docID := strings.ToLower(req.documentID)

			if msg, ok := notificationMap[docID]; ok {
				req.resChan <- documentNotificationResponse{
					msg: msg,
				}

				continue
			}

			req.resChan <- documentNotificationResponse{
				err: errors.New("notification message not found"),
			}
		case req := <-w.jobWaitChan:
			jobID := strings.ToLower(req.jobID)

			if _, ok := notificationMap[jobID]; ok {
				close(req.waitChan)

				continue
			}

			jobWaitMap[jobID] = append(jobWaitMap[jobID], req.waitChan)
		}
	}
}

func (w *Receiver) addr() string {
	return ":" + strconv.Itoa(w.port)
}

func (w *Receiver) Start(ctx context.Context) {
	go w.handleNotifications(ctx)

	w.s = &http.Server{Addr: w.addr(), Handler: w}

	startUpErrOut := make(chan error)

	go func() {
		if err := w.s.ListenAndServe(); err != nil {
			startUpErrOut <- err
		}
	}()

	// listen to context events as well as http server startup errors
	select {
	case err := <-startUpErrOut:
		// this could create an issue if the listeners are blocking.
		// We need to only propagate the error if its an error other than a server closed
		if err != nil && err.Error() != http.ErrServerClosed.Error() {
			log.Fatalf("failed to start webhook receiver %v", err)
		}
		// most probably a graceful shutdown
		log.Error(err)
	case <-ctx.Done():
		ctxn, canc := context.WithTimeout(context.Background(), 1*time.Second)
		defer canc()
		// gracefully shutdown the webhook servstartUpErrInnerer
		log.Info("Shutting down webhook server")
		err := w.s.Shutdown(ctxn)
		if err != nil {
			panic(err)
		}
		log.Info("webhook server stopped")
	}
}
