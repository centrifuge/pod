// +build testworld

package testworld

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/notification"
)

type webhookReceiver struct {
	port     int
	endpoint string

	messages []notification.Message
	msgMu    sync.RWMutex

	jobSubs map[string]chan<- bool
	subMu   sync.RWMutex

	s *http.Server
}

func newWebhookReceiver(port int, endpoint string) *webhookReceiver {
	return &webhookReceiver{
		port:     port,
		endpoint: endpoint,
		jobSubs:  make(map[string]chan<- bool),
	}
}

func (w *webhookReceiver) start(ctx context.Context) {
	w.s = &http.Server{Addr: ":" + strconv.Itoa(w.port), Handler: w}

	startUpErrOut := make(chan error)
	go func(startUpErrInner chan<- error) {
		err := w.s.ListenAndServe()
		if err != nil {
			startUpErrInner <- err
		}
	}(startUpErrOut)

	// listen to context events as well as http server startup errors
	select {
	case err := <-startUpErrOut:
		// this could create an issue if the listeners are blocking.
		// We need to only propagate the error if its an error other than a server closed
		if err != nil && err.Error() != http.ErrServerClosed.Error() {
			log.Fatalf("failed to start webhook receiver %v", err)
		}
		// most probably a graceful shutdown
		log.Info(err)
	case <-ctx.Done():
		ctxn, canc := context.WithTimeout(context.Background(), 1*time.Second)
		defer canc()
		// gracefully shutdown the webhook server
		log.Info("Shutting down webhook server")
		err := w.s.Shutdown(ctxn)
		if err != nil {
			panic(err)
		}
		log.Info("webhook server stopped")
	}
}

func (w *webhookReceiver) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var msg notification.Message
	err := decoder.Decode(&msg)
	if err != nil {
		log.Error(err)
	}

	// store
	w.msgMu.Lock()
	defer w.msgMu.Unlock()
	w.messages = append(w.messages, msg)

	if msg.EventType != notification.EventTypeJob {
		return
	}

	w.subMu.RLock()
	defer w.subMu.RUnlock()
	ch, ok := w.jobSubs[strings.ToLower(msg.Job.ID.String())]
	if !ok {
		return
	}

	go func() {
		ch <- true
	}()
}

func (w *webhookReceiver) getReceivedDocumentMsg(to string, docID string) (msg notification.Message, err error) {
	w.msgMu.RLock()
	defer w.msgMu.RUnlock()
	for _, msg := range w.messages {
		if msg.EventType != notification.EventTypeDocument {
			continue
		}

		to = strings.ToLower(to)
		if strings.ToLower(msg.Document.To.String()) != to {
			continue
		}

		if strings.ToLower(msg.Document.ID.String()) != docID {
			continue
		}

		return msg, nil
	}

	return msg, errors.New("not found")
}

// waitForJobCompletion sends bool on channel when the job is complete
func (w *webhookReceiver) waitForJobCompletion(jobID string, resp chan<- bool) {
	w.msgMu.RLock()
	defer w.msgMu.RUnlock()
	jobID = strings.ToLower(jobID)
	for _, msg := range w.messages {
		if msg.EventType != notification.EventTypeJob {
			continue
		}

		if strings.ToLower(msg.Job.ID.String()) != jobID {
			continue
		}

		go func() {
			resp <- true
		}()
		return
	}

	w.subMu.Lock()
	defer w.subMu.Unlock()
	w.jobSubs[jobID] = resp
}

func (w *webhookReceiver) url() string {
	return "http://localhost:" + strconv.Itoa(w.port) + w.endpoint
}
