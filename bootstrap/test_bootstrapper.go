// +build integration unit

package bootstrap

import "fmt"

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
	for i, b := range bootstrappers {
		err := b.TestBootstrap(ctx)
		if err != nil {
			// TODO remove this
			fmt.Println(i)
			panic(err)
		}
	}
}

func RunTestTeardown(bootstrappers []TestBootstrapper) {
	for _, b := range bootstrappers {
		err := b.TestTearDown()
		if err != nil {
			panic(err)
		}
	}
}
