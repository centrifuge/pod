//go:build integration || testworld

package bootstrap

// TestBootstrapper must be implemented by all packages that needs bootstrapping at the start of testing suite
type TestBootstrapper interface {

	// TestBootstrap initializes a module for testing
	TestBootstrap(context map[string]interface{}) error

	// TestTearDown tears down a module after testing
	TestTearDown() error
}

func RunTestBootstrappers(bootstrappers []TestBootstrapper, ctx map[string]any) map[string]any {
	if ctx == nil {
		ctx = make(map[string]any)
	}

	for _, b := range bootstrappers {
		err := b.TestBootstrap(ctx)
		if err != nil {
			panic(err)
		}
	}

	return ctx
}

func RunTestTeardown(bootstrappers []TestBootstrapper) {
	for _, b := range bootstrappers {
		err := b.TestTearDown()
		if err != nil {
			panic(err)
		}
	}
}
