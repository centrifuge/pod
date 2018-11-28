// Package node provides utilities to control all long running background services on Centrifuge node
package node

import (
	"context"
	"sync"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("node")

// Server interface must be implemented by all background services on Cent Node
type Server interface {

	// Name is the unique name given to the service within the Cent Node
	Name() string

	// Start starts the service, expectation is that this would always be called in a separate go routine.
	// WaitGroup contract should always be honoured by calling `defer wg.Done()`
	Start(ctx context.Context, wg *sync.WaitGroup, startupErr chan<- error)
}

// Node provides utilities to control all background services on Cent Node
type Node struct {
	services []Server
}

// New returns a new Node with given services.
func New(services []Server) *Node {
	return &Node{
		services: services,
	}
}

// Name returns "CentNode".
func (n *Node) Name() string {
	return "CentNode"
}

// Start starts all services that are wired in the Node and waits for further actions depending on errors or commands from upstream.
func (n *Node) Start(ctx context.Context, startupErr chan<- error) {
	ctxCh, cancel := context.WithCancel(ctx)
	defer cancel()
	childErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(len(n.services))
	for _, s := range n.services {
		go s.Start(ctxCh, &wg, childErr)
	}
	for {
		select {
		case errOut := <-childErr:
			log.Error("Node received error from child service, stopping all child services", errOut)
			// if one of the children fails to start all should stop
			cancel()
			// send the error upstream
			startupErr <- errOut
			wg.Wait()
			return
		case <-ctx.Done():
			log.Info("Node received context.done signal, stopping all child services")
			// Note that in this case the children will also receive the done signal via the passed on context
			wg.Wait()
			log.Info("Node stopped all child services")
			// special case to make the caller wait until servers are shutdown
			startupErr <- nil
			return
		}
	}
}
