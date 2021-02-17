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

	receivedMsgs     []notification.Message
	receivedMsgsLock sync.RWMutex

	s *http.Server
}

func newWebhookReceiver(port int, endpoint string) *webhookReceiver {
	return &webhookReceiver{
		port:     port,
		endpoint: endpoint,
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
	w.receivedMsgsLock.Lock()
	defer w.receivedMsgsLock.Unlock()
	w.receivedMsgs = append(w.receivedMsgs, msg)
}

func (w *webhookReceiver) getReceivedDocumentMsg(to string, docID string) (msg notification.Message, err error) {
	w.receivedMsgsLock.RLock()
	defer w.receivedMsgsLock.RUnlock()
	for _, msg := range w.receivedMsgs {
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

func (w *webhookReceiver) url() string {
	return "http://localhost:" + strconv.Itoa(w.port) + w.endpoint
}
