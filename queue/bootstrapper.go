package queue

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
)

type Bootstrapper struct {
	context map[string]interface{}
}

func (b *Bootstrapper) Bootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}
	cfg := context[config.BootstrappedConfig].(*config.Configuration)
	srv := &Server{config: cfg, taskTypes: []TaskType{}}
	context[bootstrap.BootstrappedQueueServer] = srv
	b.context = context
	return nil
}
