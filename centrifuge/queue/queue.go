package queue

import (
	"github.com/gocelery/gocelery"
	"sync"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
)

var Queue *gocelery.CeleryClient
var queueInit sync.Once

func InitQueue() {
	// TODO do this based on config
	queueInit.Do(func() {
		var err error
		Queue, err = gocelery.NewCeleryClient(gocelery.NewInMemoryBroker(), gocelery.NewInMemoryBackend(), 1)
		Queue.Register(identity.RegistrationConfirmationTaskName, &identity.RegistrationConfirmationTask{})
		if err != nil {
			panic("Could not initialize the queue")
		}
	})
}
