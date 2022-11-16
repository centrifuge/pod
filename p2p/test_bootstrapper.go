//go:build integration || testworld

package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/errors"
)

const (
	peerStartTimeout = 5 * time.Second
	peerStopTimeout  = 10 * time.Second
)

func (b *Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	if err := b.Bootstrap(ctx); err != nil {
		return err
	}

	peer := ctx[bootstrap.BootstrappedPeer].(*p2pPeer)

	b.testPeerCtx, b.testPeerCtxCancel = context.WithCancel(context.Background())

	b.testPeerWg.Add(1)

	errChan := make(chan error)

	go peer.Start(b.testPeerCtx, &b.testPeerWg, errChan)

	select {
	case err := <-errChan:
		return fmt.Errorf("couldn't start p2p peer: %w", err)
	case <-time.After(peerStartTimeout):
		// Peer started successfully
	}

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	b.testPeerCtxCancel()

	peerDone := make(chan struct{})

	go func() {
		b.testPeerWg.Wait()
		close(peerDone)
	}()

	select {
	case <-time.After(peerStopTimeout):
		return errors.New("timeout reached while waiting for peer to stop")
	case <-peerDone:
		// Peer stopped successfully
	}

	return nil
}
