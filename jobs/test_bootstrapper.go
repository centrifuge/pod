//go:build integration || testworld

package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/errors"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
)

const (
	dispatcherStartTimeout = 3 * time.Second
	dispatcherStopTimeout  = 10 * time.Second
)

func (b *Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	if err := b.Bootstrap(ctx); err != nil {
		return err
	}

	dispatcher := ctx[BootstrappedJobDispatcher].(Dispatcher)

	valueCtx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, testingcommons.CopyServiceContext(ctx))

	b.testDispatcherCtx, b.testDispatcherCtxCanc = context.WithCancel(valueCtx)

	b.testDispatcherWaitGroup.Add(1)

	errChan := make(chan error)

	go dispatcher.Start(b.testDispatcherCtx, &b.testDispatcherWaitGroup, errChan)

	select {
	case err := <-errChan:
		return fmt.Errorf("couldn't start dispatcher: %w", err)
	case <-time.After(dispatcherStartTimeout):
		// Dispatcher started successfully
	}

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	b.testDispatcherCtxCanc()

	dispatcherDone := make(chan struct{})

	go func() {
		b.testDispatcherWaitGroup.Wait()
		close(dispatcherDone)
	}()

	select {
	case <-time.After(dispatcherStopTimeout):
		return errors.New("timeout reached while waiting for dispatcher to stop")
	case <-dispatcherDone:
		// Dispatcher stopped successfully
	}

	return nil
}
