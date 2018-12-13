package queue

import (
	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct {
	context map[string]interface{}
}

// Bootstrap initiates the queue.
func (b *Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[bootstrap.BootstrappedConfig].(config.Configuration)
	srv := &Server{config: cfg, taskTypes: []TaskType{}}
	context[bootstrap.BootstrappedQueueServer] = srv
	b.context = context
	return nil
}
