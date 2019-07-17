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

	// receivedMsgs maps accountID+documentID to expected messages
	receivedMsgs     map[string]notification.Message
	receivedMsgsLock sync.RWMutex

	s *http.Server
}

func newWebhookReceiver(port int, endpoint string) *webhookReceiver {
	return &webhookReceiver{
		port:             port,
		endpoint:         endpoint,
		receivedMsgs:     make(map[string]notification.Message),
		receivedMsgsLock: sync.RWMutex{},
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
		return
	case <-ctx.Done():
		ctxn, _ := context.WithTimeout(context.Background(), 1*time.Second)
		// gracefully shutdown the webhook server
		log.Info("Shutting down webhook server")
		err := w.s.Shutdown(ctxn)
		if err != nil {
			panic(err)
		}
		log.Info("webhook server stopped")
		return
	}
}

func (w *webhookReceiver) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var msg notification.Message
	err := decoder.Decode(&msg)
	if err != nil {
		log.Error(err)
	}
	log.Infof("webhook received for received document %s from collaborator %s for node %s", msg.DocumentID, msg.FromID, msg.AccountID)

	// store
	w.receivedMsgsLock.Lock()
	defer w.receivedMsgsLock.Unlock()
	w.receivedMsgs[strings.ToLower(msg.AccountID)+"-"+strconv.Itoa(int(msg.EventType))+"-"+msg.DocumentID] = msg
}

func (w *webhookReceiver) getReceivedMsg(accountID string, eventType int, docID string) (notification.Message, error) {
	w.receivedMsgsLock.RLock()
	defer w.receivedMsgsLock.RUnlock()
	n, ok := w.receivedMsgs[strings.ToLower(accountID)+"-"+strconv.Itoa(eventType)+"-"+docID]
	if !ok {
		return n, errors.New("not found")
	}

	return n, nil
}

func (w *webhookReceiver) url() string {
	return "http://localhost:" + strconv.Itoa(w.port) + w.endpoint
}
