package bootstrap

// DO NOT PUT any app logic in this package to avoid any dependency cycles

const (
	BootstrappedConfig         string = "BootstrappedConfig"
	BootstrappedLevelDb        string = "BootstrappedLevelDb"
	BootstrappedEthereumClient string = "BootstrappedEthereumClient"
	BootstrappedAnchorRepository string = "BootstrappedAnchorRepository"
)

// Bootstrapper must be implemented by all packages that needs bootstrapping at application start
type Bootstrapper interface {
	Bootstrap(context map[string]interface{}) error
}
