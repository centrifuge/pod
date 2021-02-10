package bootstrap

// DO NOT PUT any app logic in this package to avoid any dependency cycles

// Bootstrap constants are keys to mapped value in bootstrapped context
const (
	BootstrappedConfig    string = "BootstrappedConfig"
	BootstrappedPeer      string = "BootstrappedPeer"
	BootstrappedAPIServer string = "BootstrappedAPIServer"
	NodeObjRegistry       string = "NodeObjRegistry"
	// BootstrappedNFTService is the key to NFT Service in bootstrap context.
	BootstrappedNFTService = "BootstrappedNFTService"
)

// Bootstrapper must be implemented by all packages that needs bootstrapping at application start
type Bootstrapper interface {
	Bootstrap(context map[string]interface{}) error
}
