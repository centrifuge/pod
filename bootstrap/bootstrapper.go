package bootstrap

// DO NOT PUT any app logic in this package to avoid any dependency cycles

const (
	BootstrappedConfigFile       string = "BootstrappedConfigFile"
	BootstrappedConfig           string = "BootstrappedConfig"
	BootstrappedLevelDb          string = "BootstrappedLevelDb"
	BootstrappedEthereumClient   string = "BootstrappedEthereumClient"
	BootstrappedAnchorRepository string = "BootstrappedAnchorRepository"
	BootstrappedP2PClient        string = "BootstrappedP2PClient"
	BootstrappedP2PServer        string = "BootstrappedP2PServer"
	BootstrappedAPIServer        string = "BootstrappedAPIServer"
	BootstrappedQueueServer      string = "BootstrappedQueueServer"
)

// Bootstrapper must be implemented by all packages that needs bootstrapping at application start
type Bootstrapper interface {
	Bootstrap(context map[string]interface{}) error
}
