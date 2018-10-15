package bootstrap

// DO NOT PUT any app logic in this package to avoid any dependency cycles

const (
	BootstrappedConfig         string = "BootstrappedConfig"
	BootstrappedLevelDb        string = "BootstrappedLevelDb"
	BootstrappedEthereumClient string = "BootstrappedEthereumClient"
)

// Bootstrapper must be implemented by all packages that needs bootstrapping at application start
type Bootstrapper interface {
	Bootstrap(context map[string]interface{}) error
}

// TestBootstrapper must be implemented by all packages that needs bootstrapping at the start of testing suite
type TestBootstrapper interface {

	// TestBootstrap initializes a module for testing
	TestBootstrap(context map[string]interface{}) error

	// TestTearDown tears down a module after testing
	TestTearDown() error
}
