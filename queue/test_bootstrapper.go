// +build integration unit

package queue

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"sync"
	"context"
)

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (b *Bootstrapper) TestTearDown() error {
	srv := b.context[bootstrap.BootstrappedQueueServer].(*Server)
	return srv.Stop()
}

type Starter struct {}

func (Starter) TestBootstrap(ctx map[string]interface{}) error {
	// handle the special case for running the queue server after task types have been registered (done by node bootstrapper at runtime)
	qs := ctx[bootstrap.BootstrappedQueueServer].(Server)
	childErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go qs.Start(context.Background(), &wg, childErr)
	return nil
}

func (Starter) TestTearDown() error {
	return nil
}