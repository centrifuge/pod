package bootstrapper

// DO NOT PUT any app logic in this package to avoid any dependency cycles

const (
	BOOTSTRAPPED_CONFIG  string = "BOOTSTRAPPED_CONFIG"
	BOOTSTRAPPED_LEVELDB string = "BOOTSTRAPPED_LEVELDB"
)

type Bootstrapper interface {
	Bootstrap(context map[string]interface{}) error
}

// TODO Vimukthi make integration tests init through this interface
type TestBootstrapper interface {
	TestBootstrap(context map[string]interface{}) error
}
