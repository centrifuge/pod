package rabbitmq

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/queue"
	"log"
)

type TestMessage struct {
	Msg string
}

func (m *TestMessage) MarshalBinary() (data []byte, err error) {
	return []byte(m.Msg), nil
}

func TestQueue_Enqueue(t *testing.T) {
	qu := GetQueue()
	qu.Start()
	w, _ := GetWorkerRegistry().Get("anchor")
	w.AddHandler(func(msg []byte) (queue.HandlerStatus, error) {
		msgStr := string(msg)
		log.Println("Received message ", msgStr)
		return queue.Success, nil
	})
	w.Start(nil)

	qu.Enqueue("anchor",  &queue.MessageWrapper{&queue.Header{}, &TestMessage{"Hello"}})
	forever := make(chan bool)
	<- forever

}