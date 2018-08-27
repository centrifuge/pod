package bootstrapper

// DO NOT PUT any app logic in this package to avoid any dependency cycles

const (
	BootstrappedConfig  string = "BootstrappedConfig"
	BootstrappedLevelDb string = "BootstrappedLevelDb"
)

// Bootstrapper must be implemented by all packages that needs bootstrapping at application start
type Bootstrapper interface {
	Bootstrap(context map[string]interface{}) error
}

// TestBootstrapper must be implemented by all packages that needs bootstrapping at the start of testing suite
// TODO Vimukthi make integration tests init through this interface
type TestBootstrapper interface {
	TestBootstrap(context map[string]interface{}) error
}
