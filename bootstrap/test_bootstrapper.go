// +build integration unit

package bootstrap

import (
	"github.com/centrifuge/go-centrifuge/queue"
	"sync"
	"context"
)

// TestBootstrapper must be implemented by all packages that needs bootstrapping at the start of testing suite
type TestBootstrapper interface {

	// TestBootstrap initializes a module for testing
	TestBootstrap(context map[string]interface{}) error

	// TestTearDown tears down a module after testing
	TestTearDown() error
}

func RunTestBootstrappers(bootstrappers []TestBootstrapper, ctx map[string]interface{}) {
	if ctx == nil {
		ctx = map[string]interface{}{}
	}
	for _, b := range bootstrappers {
		err := b.TestBootstrap(ctx)
		if err != nil {
			panic(err)
		}
	}

	// handle the special case for running the queue server after task types have been registered (done by node bootstrapper at runtime)
	qs := ctx[BootstrappedQueueServer].(queue.Server)
	childErr := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	go qs.Start(context.Background(), &wg, childErr)
}

func RunTestTeardown(bootstrappers []TestBootstrapper) {
	for _, b := range bootstrappers {
		err := b.TestTearDown()
		if err != nil {
			panic(err)
		}
	}
}
