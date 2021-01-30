package queue

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
)

// Bootstrapper implements bootstrap.Bootstrapper.
type Bootstrapper struct {
	context map[string]interface{}
}

// Bootstrap initiates the queue.
func (b *Bootstrapper) Bootstrap(context map[string]interface{}) error {
	cfg, err := config.RetrieveConfig(false, context)
	if err != nil {
		return err
	}
	srv := &Server{config: cfg, taskTypes: []TaskType{}}
	context[bootstrap.BootstrappedQueueServer] = srv
	b.context = context
	return nil
}
