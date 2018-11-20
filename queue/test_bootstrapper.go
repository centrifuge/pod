// +build integration unit

package queue

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"sync"
	"context"
)

func (b *Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	err := b.Bootstrap(ctx)
	if err != nil {
		return err
	}
	srv := ctx[bootstrap.BootstrappedQueueServer].(*QueueServer)
	childErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go srv.Start(context.Background(), &wg, childErr)
	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	srv := b.context[bootstrap.BootstrappedQueueServer].(*QueueServer)
	return srv.Stop()
}
