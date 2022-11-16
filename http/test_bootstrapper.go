//go:build integration || testworld

package http

import (
	"context"
	"errors"
	"fmt"
	"time"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"

	"github.com/centrifuge/go-centrifuge/bootstrap"
)

const (
	httpServerStartTimeout = 3 * time.Second
	httpServerStopTimeout  = 10 * time.Second
)

func (b *Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	if err := b.Bootstrap(ctx); err != nil {
		return err
	}

	httpServer := ctx[bootstrap.BootstrappedAPIServer].(apiServer)

	valueCtx := context.WithValue(context.Background(), bootstrap.NodeObjRegistry, testingcommons.CopyServiceContext(ctx))

	b.testServerCtx, b.testServerCtxCancel = context.WithCancel(valueCtx)

	b.testServerWg.Add(1)

	errChan := make(chan error)

	go httpServer.Start(b.testServerCtx, &b.testServerWg, errChan)

	select {
	case err := <-errChan:
		return fmt.Errorf("couldn't start HTTP server: %w", err)
	case <-time.After(httpServerStartTimeout):
		// HTTP server started successfully
	}

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	b.testServerCtxCancel()

	httpServerDone := make(chan struct{})

	go func() {
		b.testServerWg.Wait()
		close(httpServerDone)
	}()

	select {
	case <-time.After(httpServerStopTimeout):
		return errors.New("timeout reached while waiting for peer to stop")
	case <-httpServerDone:
		// Peer stopped successfully
	}

	return nil
}
