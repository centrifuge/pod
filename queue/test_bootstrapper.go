// +build integration unit

package queue

import (
	"context"
	"sync"

	"github.com/centrifuge/go-centrifuge/bootstrap"
)

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

type Starter struct {
	canF context.CancelFunc
}

func (s *Starter) TestBootstrap(ctx map[string]interface{}) error {
	// handle the special case for running the queue server after task types have been registered (done by node bootstrapper at runtime)
	qs := ctx[bootstrap.BootstrappedQueueServer].(*Server)
	childErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	c, canF := context.WithCancel(context.Background())
	s.canF = canF
	go qs.Start(c, &wg, childErr)
	return nil
}

func (s *Starter) TestTearDown() error {
	s.canF()
	return nil
}
